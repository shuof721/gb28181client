package media

import "encoding/binary"

// GB28181 视频：H.264 AU → PES → PS → RTP(PT=96)
// 结构与常见国标设备/ ZLMediaKit 兼容。

// PackVideoPES 将一帧 H.264（Annex-B AU）打成纯视频 PS 包。
// ts 为 90kHz。
func PackVideoPES(es []byte, ts uint64) []byte {
	return PackCompoundPES(es, nil, ts)
}

// PackCompoundPES 将一帧视频和一帧音频打成 PS 复合流包。
// 若 audioES 为空，则生成符合国标规范的单视频 PS 包。
// ts 为 90kHz 时钟基准。
func PackCompoundPES(videoES []byte, audioES []byte, ts uint64) []byte {
	hasAudio := len(audioES) > 0
	totalCap := 64 + len(videoES) + len(audioES)
	out := make([]byte, 0, totalCap)
	out = append(out, buildPackHeader(ts)...)
	out = append(out, buildSystemHeader(hasAudio)...)
	out = append(out, buildPSM(hasAudio)...)
	if hasAudio {
		// 音频 PES 置于视频 PES 之前：音频帧长明确（320字节），不会因视频 I 帧超出 64KB 触发 pesLen=0 而被解析器吞噬
		out = append(out, buildPES(0xC0, audioES, ts)...)
	}
	out = append(out, buildPES(0xE0, videoES, ts)...)
	return out
}

func buildPackHeader(scr uint64) []byte {
	// 00 00 01 BA
	// '01' SCR[32..30] 1 SCR[29..15] 1 SCR[14..0] 1 SCR_ext[8..0] 1
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
	return hdr
}

func buildSystemHeader(hasAudio bool) []byte {
	// ISO 13818-1 system header，length=6
	// audio_bound: 6 bit, fixed: 1 bit, CSPS: 1 bit
	audioBound := byte(0x00)
	if hasAudio {
		audioBound = 0x01
	}
	b3 := (audioBound << 2) | 0x01
	return []byte{
		0x00, 0x00, 0x01, 0xBB,
		0x00, 0x06,
		0x80, 0x00, 0x01, // rate_bound + markers
		b3,               // audio_bound, fixed=0, CSPS=1
		0xFF,             // video_bound + markers
		0xFC,             // packet_rate_restriction
	}
}

func buildPSM(hasAudio bool) []byte {
	// Program Stream Map (0x000001BC)
	// stream_type=0x1B H.264 (stream_id=0xE0)
	// stream_type=0x90 G.711A (stream_id=0xC0)
	var esMap []byte
	if hasAudio {
		esMap = []byte{
			0x1B, 0xE0, 0x00, 0x00, // Video: H.264, stream_id 0xE0
			0x90, 0xC0, 0x00, 0x00, // Audio: G.711A, stream_id 0xC0
		}
	} else {
		esMap = []byte{
			0x1B, 0xE0, 0x00, 0x00, // Video: H.264, stream_id 0xE0
		}
	}
	esMapLen := len(esMap)
	body := make([]byte, 6+esMapLen+4)
	body[0] = 0xE0 // current_next=1, reserved, version=0
	body[1] = 0xFF // reserved + marker
	body[2] = 0x00 // program_stream_info_length = 0
	body[3] = 0x00
	body[4] = byte(esMapLen >> 8) // elementary_stream_map_length
	body[5] = byte(esMapLen)
	copy(body[6:], esMap)

	// MPEG-2 CRC32 (polynomial: 0x04C11DB7, init: 0xFFFFFFFF)
	crc := crc32MPEG2(body[:6+esMapLen])
	binary.BigEndian.PutUint32(body[6+esMapLen:], crc)

	out := make([]byte, 6+len(body))
	out[0], out[1], out[2], out[3] = 0x00, 0x00, 0x01, 0xBC
	binary.BigEndian.PutUint16(out[4:6], uint16(len(body)))
	copy(out[6:], body)
	return out
}

// crc32MPEG2 计算 ISO/IEC 13818-1 规定的 32 位 CRC
func crc32MPEG2(data []byte) uint32 {
	crc := uint32(0xFFFFFFFF)
	for _, b := range data {
		crc ^= uint32(b) << 24
		for i := 0; i < 8; i++ {
			if crc&0x80000000 != 0 {
				crc = (crc << 1) ^ 0x04C11DB7
			} else {
				crc <<= 1
			}
		}
	}
	return crc
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
