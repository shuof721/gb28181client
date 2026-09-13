# GB28181 模拟设备客户端

用 Go 实现的 GB/T 28181-2016 **设备端模拟器**：模拟一台 **NVR + 多通道 IPC**，可注册到国标平台，并完成目录/状态/控制等信令交互，以及实时点播的 **PS/RTP** 媒体发送。

对接目标：**WVP-GB28181-pro + ZLMediaKit**（也附带本地 mock 平台便于无平台自测）。

## 功能

| 类别 | 能力 |
|------|------|
| 注册 | SIP REGISTER + Digest 鉴权，到期自动续注册 |
| 心跳 | MANSCDP Keepalive，失败重试与超时重注册 |
| 目录 | 响应 Catalog 查询（NVR + 多通道） |
| 设备信息 | DeviceInfo / DeviceStatus |
| 控制 | DeviceControl（虚拟云台 PTZ 8方向/变倍/速度平滑运动状态机，预置位 0x81/0x82/0x83 设置/调用/删除） |
| 查询 | RecordInfo 虚拟录像排程查询、PresetQuery 预置位列表查询 |
| 报警 | 可主动发送 Alarm Notify |
| 媒体 | INVITE 实时点播与历史录像回放 (含 MANSRTSP 拖动 Seek 与倍速) |
| 传输 | SIP UDP/TCP；媒体 RTP/UDP，以及 TCP 长度前缀 |

## 目录结构

```
cmd/gb28181-device/     设备模拟器入口
cmd/mock-platform/      本地简易 SIP 平台（联调用）
configs/config.yaml     示例配置
internal/config/        配置加载
internal/sip/           SIP 消息/传输/Digest/UA
internal/gb28181/       MANSCDP XML
internal/media/         H264 源、PS 封装、RTP 会话
internal/device/        设备业务状态机
assets/test.h264        测试用 Annex-B H.264
```

## 快速开始

### 1. 编译

```powershell
go build -o bin/gb28181-device.exe ./cmd/gb28181-device
go build -o bin/mock-platform.exe ./cmd/mock-platform
```

### 2. 配置 WVP 侧

在 WVP 中「添加国标设备」：

- **设备编号**：与 `device.id` 一致，例如 `34020000001180000001`
- **密码**：与 `sip.password` 一致
- **信令地址**：设备可达的 IP，端口为 WVP 的 SIP 端口（默认 5060）
- 媒体收流由 ZLM 负责（WVP 配置里 `media` 相关地址）

### 3. 配置本程序

编辑 `configs/config.yaml`：

```yaml
sip:
  server_ip: "192.168.1.10"    # WVP 所在 IP
  server_port: 5060
  local_ip: "192.168.1.20"     # 本机 IP（WVP 必须能连回来）
  local_port: 5070
  transport: "udp"
  username: "34020000001180000001"
  password: "12345678"

device:
  id: "34020000001180000001"
  channels:
    - id: "34020000001320000001"
      name: "通道1"
      # ...

media:
  source: "file"                 # 推荐 file；synthetic 为内置测试帧
  h264_file: "assets/test.h264"
  fps: 25
```

### 4. 运行

```powershell
.\bin\gb28181-device.exe -config configs\config.yaml
```

启动后浏览器打开轻量控制台（默认）：

```text
http://127.0.0.1:7080
```

控制台能力：

- 注册/心跳/报警、通道卡片、点播会话
- 上传 mp4/h264 到 `assets/`
- Web 端新增/删除通道，并绑定各通道视频
- 日志关键字过滤

```yaml
ui:
  enabled: true
  listen: "127.0.0.1:7080"
```

成功注册后日志类似：

```
[sip] REGISTER OK expires=3600
[device] registered as 34020000001180000001
[gb] recv Catalog SN=...
[gb] catalog response sent, channels=2
```

在 WVP 页面对通道点「播放」，设备收到 INVITE 后开始向 ZLM 推 PS 流。

### 无平台本地自测

```powershell
# 终端 1：mock 平台
.\bin\mock-platform.exe -listen-ip 127.0.0.1 -listen-port 5060 -device-id 34020000001180000001

# 终端 2：把 config 里 server/local 都改成 127.0.0.1 后启动设备
.\bin\gb28181-device.exe -config configs\config.yaml
```

mock 平台会在注册成功后自动下发 Catalog / DeviceInfo 查询。

## 媒体源说明

| source | 说明 |
|--------|------|
| `mp4` | **推荐**。本地 MP4 循环推流。首次点播时用 ffmpeg 抽 H.264 到同目录 `*.h264.cache`，之后复用。需本机有 ffmpeg。 |
| `file` | 循环读取 Annex-B H.264（`h264_file`）。可 `ffmpeg -f lavfi -i testsrc2=size=1280x720:rate=25 -t 10 -c:v libx264 -profile:v baseline -bf 0 -f h264 assets/test.h264` |
| `synthetic` | 程序内生成 I_PCM 宏块 IDR，无需外部文件，码率大，仅连通性验证。 |

```yaml
media:
  source: "mp4"
  mp4_file: "assets/demo.mp4"   # 改成你的本地视频
  fps: 25
```

## 国标编号建议

- 设备（NVR）：`34020000001180000001`（11=编码设备，18=类型可按规范调整）
- 通道（IPC）：`34020000001320000001`（13=摄像机）
- 须满足 20 位；平台侧添加的设备号必须与配置完全一致。

## 已实现协议细节

- SIP 事务：REGISTER / MESSAGE / INVITE / BYE / INFO / OPTIONS / SUBSCRIBE
- Digest：MD5 / MD5-sess / SHA-256，支持 qop=auth
- MANSCDP：Catalog、DeviceInfo、DeviceStatus、DeviceControl、PresetQuery、RecordInfo、Keepalive、Alarm
- 实时点播与回放：解析 `c=`/`m=`/`y=`，回 `sendonly` SDP（含 `y=` SSRC），支持 MANSRTSP 拖动/倍速与结束通知 (121)
- PS 复合流推流：支持音视频复合流 (H.264 + G.711A)，System Header 动态声明 `audio_bound`，PSM 注册视频 (0x1B) 与音频 (0x90) 双轨并计算 MPEG-2 CRC32，支持蜂鸣/正弦/环境底噪/静音等音源与 Web 实时跳动 VU 电平表
- 虚拟云台与预置位：PTZ 8方向/变倍平滑运动仿真，支持 PresetQuery、Set(0x81)、Call(0x82)、Delete(0x83) 与 Web 可视化操控
- RTP：PT=96，SSRC 使用平台 `y=` 中的值

## 尚未覆盖 / 后续可做

- 云台巡航组 (Cruise / Patrol) 与自动线扫 (Auto Scan) 状态机推进
- 历史录像文件下载 (s=Download) 与极速/倍速推流
- H.265 (HEVC, stream_type=0x24) 视频编码推流支持
- TCP 媒体的 `0x24` interleaved 模式（当前为 RFC 4571 2 字节长度前缀）
- 更严格的 SIP 事务层（重传定时器、CANCEL 等）

## 测试

```powershell
go test ./...
```

## 常见问题

1. **注册不上**  
   检查防火墙是否放行 UDP/TCP 5070 与 5060；`local_ip` 必须是平台能路由到的地址。

2. **有注册无画面**  
   - ZLM 收流端口是否可达  
   - `media.source=file` 时路径是否正确  
   - WVP 是否把 INVITE 的 SDP 媒体 IP 配成了 ZLM 可监听地址  

3. **目录为空**  
   看设备日志是否收到 Catalog；响应 MESSAGE 是否发往平台 SIP 地址。

4. **鉴权失败**  
   确认 `sip.password` 与 WVP 添加设备时密码一致；必要时用 `-auth` 测试 mock 平台。
