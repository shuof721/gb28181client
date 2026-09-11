package media

import "encoding/binary"

// GB28181 视频：H.264 AU → PES → PS → RTP(PT=96)
// 结构与常见国标设备/ ZLMediaKit 兼容。

// PackVideoPES 将一帧 H.264（Annex-B AU）打成完整 PS 包。
// ts 为 90kHz。
func PackVideoPES(es []byte, ts uint64) []byte {
	// Pack(14) + System(18) + PSM(~26) + PES(14+es)
	out := make([]byte, 0, 64+32+len(es))
	out = append(out, buildPackHeader(ts)...)
	out = append(out, buildSystemHeader()...)
	out = append(out, buildPSM()...)
	out = append(out, buildPES(0xE0, es, ts)...)
	return out
}

func buildPackHeader(scr uint64) []byte {
	// 00 00 01 BA
	// '01' SCR[32..30] 1 SCR[29..15] 1 SCR[14..0] 1 SCR_ext[8..0] 1
	b0 := byte(0x44) | byte((scr>>30)&0x07)<<3 | byte((scr>>28)&0x03)
	// 重新按标准位拼（48bit）
	// bit7-6: 01
	// bit5-3: SCR32-30
	// bit2:   marker
	// bit1-0: SCR29-28
	b0 = 0x40 | byte((scr>>30)&0x07)<<3 | 0x04 | byte((scr>>28)&0x03)
	b1 := byte((scr >> 20) & 0xFF)
	// 实际 SCR29-15 是连续 15bit，跨 b0 低 2bit + b1 8bit + b2 高 5bit
	// 用位缓冲更稳：
	var bits uint64
	bits |= 0x2 << 46
	bits |= ((scr >> 30) & 0x7) << 43
	bits |= 1 << 42
	bits |= ((scr >> 15) & 0x7FFF) << 27
	bits |= 1 << 26
	bits |= (scr & 0x7FFF) << 11
	bits |= 1 << 10
	bits |= 1 // marker after ext=0

	hdr := make([]byte, 14)
	hdr[0], hdr[1], hdr[2], hdr[3] = 0x00, 0x00, 0x01, 0xBA
	hdr[4] = byte(bits >> 40)
	hdr[5] = byte(bits >> 32)
	hdr[6] = byte(bits >> 24)
	hdr[7] = byte(bits >> 16)
	hdr[8] = byte(bits >> 8)
	hdr[9] = byte(bits)
	// program_mux_rate: 22bit + marker. 填 0x061F 之类常见值
	// 00 04 00 为常见
	hdr[10], hdr[11], hdr[12] = 0x00, 0x04, 0x00
	// reserved 5bit=11111, pack_stuffing_length=0
	hdr[13] = 0xF8
	_ = b0
	_ = b1
	return hdr
}

func buildSystemHeader() []byte {
	// ISO 13818-1 system header，length=6，仅 rate/video bound
	return []byte{
		0x00, 0x00, 0x01, 0xBB,
		0x00, 0x06,
		0x80, 0x00, 0x01, // rate_bound + markers
		0x01,             // audio_bound=0, fixed=0, CSPS=1
		0xFF,             // video_bound + markers
		0xFC,             // packet_rate_restriction
	}
}

func buildPSM() []byte {
	// stream_type=0x1B H.264, elementary_stream_id=0xE0
	// after start+length:
	//   E0 FF          current_next=1, reserved, version=0
	//   00 00          program_stream_info_length=0
	//   00 04          elementary_stream_map_length=4
	//   1B E0 00 00    type, id, es_info_len
	//   CRC32 (4)
	body := []byte{
		0xE0, 0xFF,
		0x00, 0x00,
		0x00, 0x04,
		0x1B, 0xE0,
		0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, // CRC 占位，多数 demux 不校验
	}
	out := []byte{0x00, 0x00, 0x01, 0xBC, byte(len(body) >> 8), byte(len(body))}
	return append(out, body...)
}

func buildPES(streamID byte, es []byte, pts uint64) []byte {
	// PES_packet_length = 3(flags+headerlen) + 5(PTS) + len(es)，视频可写 0
	pesLen := 3 + 5 + len(es)
	if pesLen > 0xFFFF {
		pesLen = 0
	}
	out := []byte{
		0x00, 0x00, 0x01, streamID,
		byte(pesLen >> 8), byte(pesLen),
		0x80, // '10' + scrambling/priority/alignment/copyright/original
		0x80, // PTS_DTS_flag = '10' (PTS only)
		0x05, // PES_header_data_length
	}
	out = append(out, encodePTS(pts)...)
	out = append(out, es...)
	return out
}

func encodePTS(ts uint64) []byte {
	// '0010' PTS[32:30] 1 PTS[29:15] 1 PTS[14:0] 1
	b0 := byte(0x20) | byte(((ts>>30)&0x7)<<1) | 0x01
	b1 := byte((ts >> 22) & 0xFF)
	b2 := byte(((ts>>15)&0x7F)<<1) | 0x01
	b3 := byte((ts >> 7) & 0xFF)
	b4 := byte((ts&0x7F)<<1) | 0x01
	return []byte{b0, b1, b2, b3, b4}
}

// RTPPacketizePS 将 PS 按 maxPayload 切成 RTP 包。
func RTPPacketizePS(ps []byte, ssrc uint32, seq *uint16, ts uint32, payloadType byte, maxPayload int) [][]byte {
	if maxPayload <= 0 {
		maxPayload = 1400
	}
	var pkts [][]byte
	if len(ps) == 0 {
		return pkts
	}
	for off := 0; off < len(ps); off += maxPayload {
		end := off + maxPayload
		if end > len(ps) {
			end = len(ps)
		}
		marker := byte(0)
		if end >= len(ps) {
			marker = 1
		}
		pkt := make([]byte, 12+end-off)
		pkt[0] = 0x80
		pkt[1] = payloadType | (marker << 7)
		binary.BigEndian.PutUint16(pkt[2:], *seq)
		binary.BigEndian.PutUint32(pkt[4:], ts)
		binary.BigEndian.PutUint32(pkt[8:], ssrc)
		copy(pkt[12:], ps[off:end])
		pkts = append(pkts, pkt)
		*seq++
	}
	return pkts
}
