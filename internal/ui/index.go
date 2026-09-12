package ui

const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>GB28181 多设备模拟与管控平台</title>
<style>
:root{
  --bg-dark:#080c14;
  --bg:#0e1422;
  --surface:#151d2d;
  --surface-hover:#1c273c;
  --surface-2:#1c263a;
  --surface-3:#27364f;
  --border:rgba(255,255,255,0.08);
  --border-focus:rgba(59,130,246,0.6);
  --text-main:#f1f5f9;
  --text-muted:#94a3b8;
  --text-dim:#64748b;
  --accent:#3b82f6;
  --accent-glow:rgba(59,130,246,0.25);
  --accent-hover:#2563eb;
  --ok:#10b981;
  --ok-glow:rgba(16,185,129,0.25);
  --warn:#f59e0b;
  --err:#f43f5e;
  --purple:#a855f7;
  --cyan:#06b6d4;
  --radius-sm:6px;
  --radius:10px;
  --radius-lg:14px;
  --shadow-sm:0 2px 4px rgba(0,0,0,0.2);
  --shadow:0 8px 24px rgba(0,0,0,0.35);
  --shadow-lg:0 16px 36px rgba(0,0,0,0.5);
  --font-mono:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,"Liberation Mono","Courier New",monospace;
}

*{box-sizing:border-box}
html,body{height:100%;margin:0;padding:0}
body{
  background:var(--bg);
  background-image:
    radial-gradient(at 0% 0%, rgba(30,58,138,0.18) 0px, transparent 50%),
    radial-gradient(at 100% 0%, rgba(15,118,110,0.12) 0px, transparent 50%),
    radial-gradient(at 50% 100%, rgba(15,23,42,0.6) 0px, transparent 100%);
  color:var(--text-main);
  font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"PingFang SC","Microsoft YaHei",sans-serif;
  font-size:13px;
  line-height:1.5;
  -webkit-font-smoothing:antialiased;
}

::-webkit-scrollbar{width:7px;height:7px}
::-webkit-scrollbar-track{background:rgba(0,0,0,0.15)}
::-webkit-scrollbar-thumb{background:var(--surface-3);border-radius:4px}
::-webkit-scrollbar-thumb:hover{background:rgba(255,255,255,0.25)}

/* Header */
.header{
  position:sticky;top:0;z-index:40;
  display:flex;align-items:center;justify-content:space-between;gap:16px;
  padding:0 28px;height:64px;
  background:rgba(14,20,34,0.92);
  backdrop-filter:blur(16px);
  border-bottom:1px solid var(--border);
}
.brand{display:flex;align-items:center;gap:12px;min-width:0}
.brand-icon{
  width:36px;height:36px;border-radius:var(--radius);
  background:linear-gradient(135deg,#2563eb,#38bdf8);
  display:flex;align-items:center;justify-content:center;
  box-shadow:0 0 16px var(--accent-glow);
  color:#fff;flex-shrink:0;
}
.brand-info h1{margin:0;font-size:15px;font-weight:700;letter-spacing:0.02em;color:#fff}
.brand-info .ver{display:inline-block;font-size:11px;color:var(--text-muted);font-weight:400}
.header-actions{display:flex;align-items:center;gap:10px}

/* Badges */
.badge{
  display:inline-flex;align-items:center;gap:6px;
  padding:3px 8px;border-radius:999px;
  font-size:11px;font-weight:500;
  background:var(--surface-2);border:1px solid var(--border);
  color:var(--text-muted);
}
.badge .dot{width:7px;height:7px;border-radius:50%;background:var(--text-dim)}
.badge.on{
  background:rgba(16,185,129,0.12);border-color:rgba(16,185,129,0.3);color:#6ee7b7;
}
.badge.on .dot{
  background:var(--ok);box-shadow:0 0 6px var(--ok);animation:pulse 2s infinite;
}
.badge.off{
  background:rgba(244,63,94,0.1);border-color:rgba(244,63,94,0.25);color:#fda4af;
}
.badge.off .dot{background:var(--err)}
.badge.stopped{
  background:rgba(148,163,184,0.1);border-color:rgba(148,163,184,0.2);color:#94a3b8;
}
.badge.live{
  background:rgba(16,185,129,0.15);border-color:rgba(16,185,129,0.35);color:#a7f3d0;
}
.badge.live .dot{background:var(--ok);box-shadow:0 0 6px var(--ok)}
.badge.warn{
  background:rgba(245,158,11,0.12);border-color:rgba(245,158,11,0.3);color:#fcd34d;
}
.badge.warn .dot{background:var(--warn)}

@keyframes pulse{
  0%,100%{opacity:1;transform:scale(1)}
  50%{opacity:0.5;transform:scale(0.85)}
}

/* Layout */
.container{
  max-width:1440px;margin:0 auto;padding:20px 24px 60px;
  display:flex;flex-direction:column;gap:20px;
}

/* Buttons */
.btn{
  display:inline-flex;align-items:center;justify-content:center;gap:6px;
  background:var(--surface-2);color:var(--text-main);
  border:1px solid var(--border);border-radius:var(--radius-sm);
  padding:6px 12px;font-size:12px;font-weight:500;cursor:pointer;
  transition:all 0.15s ease;user-select:none;
}
.btn:hover{background:var(--surface-3);border-color:rgba(255,255,255,0.15)}
.btn:active{transform:translateY(1px)}
.btn:disabled{opacity:0.5;cursor:not-allowed;transform:none}
.btn-primary{
  background:linear-gradient(135deg,#2563eb,#3b82f6);
  border-color:#3b82f6;color:#fff;
  box-shadow:0 2px 8px var(--accent-glow);
}
.btn-primary:hover{background:linear-gradient(135deg,#1d4ed8,#2563eb);border-color:#2563eb}
.btn-success{
  background:linear-gradient(135deg,#059669,#10b981);
  border-color:#10b981;color:#fff;
}
.btn-success:hover{background:linear-gradient(135deg,#047857,#059669)}
.btn-danger{
  background:rgba(244,63,94,0.12);border-color:rgba(244,63,94,0.3);color:#fda4af;
}
.btn-danger:hover{background:rgba(244,63,94,0.22);border-color:#f43f5e;color:#fff}
.btn-sm{padding:4px 8px;font-size:11px}

/* Form Elements */
input[type=text],input[type=number],input[type=password],select,textarea{
  background:var(--surface);color:var(--text-main);
  border:1px solid var(--border);border-radius:var(--radius-sm);
  padding:7px 10px;font-size:12px;outline:none;
  transition:border-color 0.15s ease,box-shadow 0.15s ease;
  width:100%;
}
input:focus,select:focus,textarea:focus{
  border-color:var(--accent);box-shadow:0 0 0 3px var(--accent-glow);
}
select{cursor:pointer}

/* Stats Cards */
.stats-grid{
  display:grid;grid-template-columns:repeat(4,1fr);gap:14px;
}
@media (max-width:960px){.stats-grid{grid-template-columns:repeat(2,1fr)}}
.stat-card{
  background:var(--surface);border:1px solid var(--border);
  border-radius:var(--radius);padding:12px 16px;
  display:flex;flex-direction:column;gap:4px;
  transition:border-color 0.2s ease;
}
.stat-card:hover{border-color:rgba(255,255,255,0.15)}
.stat-label{
  font-size:11px;font-weight:600;text-transform:uppercase;letter-spacing:0.04em;
  color:var(--text-dim);
}
.stat-val{
  font-size:20px;font-weight:700;color:#fff;
  font-family:var(--font-mono);
}

/* Section Header */
.section-head{
  display:flex;align-items:center;justify-content:space-between;gap:12px;
  margin-bottom:12px;flex-wrap:wrap;
}
.section-title{
  font-size:14px;font-weight:700;letter-spacing:0.03em;
  text-transform:uppercase;color:var(--text-main);
  display:flex;align-items:center;gap:8px;
}

/* Device Card Grid */
.device-grid{
  display:grid;grid-template-columns:repeat(auto-fill,minmax(340px,1fr));gap:14px;
}
.device-card{
  background:var(--surface);border:1px solid var(--border);
  border-radius:var(--radius);padding:14px 16px;
  display:flex;flex-direction:column;gap:12px;position:relative;
  transition:all 0.2s ease;cursor:pointer;
}
.device-card:hover{border-color:rgba(255,255,255,0.2);transform:translateY(-1px)}
.device-card.active{
  border-color:var(--accent);
  box-shadow:0 0 16px var(--accent-glow);
}
.device-card.active::after{
  content:"当前选中";position:absolute;top:12px;right:14px;
  font-size:10px;font-weight:600;color:#93c5fd;background:rgba(59,130,246,0.2);
  padding:2px 6px;border-radius:4px;border:1px solid rgba(59,130,246,0.4);
}
.device-card-head{display:flex;align-items:flex-start;gap:10px}
.device-icon{
  width:34px;height:34px;border-radius:var(--radius-sm);
  background:var(--surface-2);border:1px solid var(--border);
  display:flex;align-items:center;justify-content:center;
  color:#93c5fd;flex-shrink:0;
}
.device-card-title{min-width:0;flex:1}
.device-card-name{font-size:14px;font-weight:700;color:#fff;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.device-card-id{font-size:11px;font-family:var(--font-mono);color:var(--text-dim)}

.device-meta-grid{
  display:grid;grid-template-columns:repeat(2,1fr);gap:8px;
  background:var(--surface-2);border-radius:var(--radius-sm);padding:8px 10px;
  font-size:11px;
}
.meta-item{display:flex;flex-direction:column;gap:2px}
.meta-k{color:var(--text-dim)}
.meta-v{font-weight:600;color:var(--text-main);font-family:var(--font-mono);overflow:hidden;text-overflow:ellipsis;white-space:nowrap}

.device-card-actions{
  display:flex;align-items:center;gap:6px;flex-wrap:wrap;border-top:1px solid var(--border);padding-top:10px;
}

/* Workbench Panel */
.panel{
  background:var(--surface);border:1px solid var(--border);
  border-radius:var(--radius-lg);box-shadow:var(--shadow-sm);
  overflow:hidden;display:flex;flex-direction:column;
}
.panel-head{
  padding:12px 18px;background:rgba(255,255,255,0.02);
  border-bottom:1px solid var(--border);
  display:flex;align-items:center;justify-content:space-between;gap:12px;
  flex-wrap:wrap;
}
.tab-group{display:flex;gap:4px;background:var(--surface-2);padding:3px;border-radius:var(--radius-sm);border:1px solid var(--border)}
.tab-btn{
  padding:5px 12px;font-size:12px;font-weight:500;border-radius:4px;border:none;
  background:transparent;color:var(--text-muted);cursor:pointer;transition:all 0.15s ease;
}
.tab-btn:hover{color:#fff}
.tab-btn.active{background:var(--accent);color:#fff;box-shadow:0 1px 4px rgba(0,0,0,0.3)}

.panel-body{padding:18px}

/* Two-column layout */
.main-grid{
  display:grid;grid-template-columns:1.4fr 1fr;gap:20px;align-items:start;
}
@media (max-width:1080px){.main-grid{grid-template-columns:1fr}}

/* Channel Cards */
.channel-grid{
  display:grid;grid-template-columns:repeat(auto-fill,minmax(300px,1fr));gap:12px;
}
.ch-card{
  background:var(--surface-2);border:1px solid var(--border);
  border-radius:var(--radius);padding:12px 14px;
  display:flex;flex-direction:column;gap:10px;position:relative;
}
.ch-card.live{
  border-color:rgba(16,185,129,0.45);
  box-shadow:0 0 16px rgba(16,185,129,0.1);
}
.ch-header{display:flex;align-items:flex-start;justify-content:space-between;gap:8px}
.ch-name{font-size:13px;font-weight:700;color:#fff}
.ch-id{font-size:11px;font-family:var(--font-mono);color:var(--text-dim)}

/* PTZ Controller Pad */
.ptz-pad-grid {
  display: grid;
  grid-template-columns: repeat(3, 46px);
  grid-template-rows: repeat(3, 46px);
  gap: 8px;
  justify-content: center;
}
.ptz-btn {
  background: var(--surface-3);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  color: var(--text-main);
  font-size: 16px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  user-select: none;
  transition: all 0.12s;
}
.ptz-btn:hover {
  background: var(--accent);
  color: #fff;
  border-color: var(--accent);
  box-shadow: 0 0 12px var(--accent-glow);
}
.ptz-btn:active {
  transform: scale(0.92);
}
.ptz-btn.stop {
  background: rgba(244,63,94,0.18);
  border-color: rgba(244,63,94,0.35);
  color: #fda4af;
}
.ptz-btn.stop:hover {
  background: var(--err);
  color: #fff;
}

/* Session Items */
.session-item{
  background:var(--surface-2);border:1px solid var(--border);
  border-radius:var(--radius);padding:10px 12px;
  display:flex;align-items:center;justify-content:space-between;gap:12px;
  margin-bottom:8px;
}

/* Log Box */
.log-box{
  background:#090d15;border:1px solid var(--border);border-radius:var(--radius);
  padding:10px 12px;height:340px;overflow-y:auto;
  font-family:var(--font-mono);font-size:11px;line-height:1.6;
}
.log-line{display:flex;gap:8px;padding:2px 0;word-break:break-all}
.log-tag{padding:1px 5px;border-radius:3px;font-size:10px;font-weight:600;flex-shrink:0}
.log-tag.sip{background:rgba(59,130,246,0.15);color:#93c5fd}
.log-tag.media{background:rgba(168,85,247,0.15);color:#d8b4fe}
.log-tag.gb{background:rgba(16,185,129,0.15);color:#6ee7b7}
.log-tag.err{background:rgba(244,63,94,0.2);color:#fda4af}

/* Modals */
.modal-mask{
  position:fixed;inset:0;background:rgba(0,0,0,0.7);backdrop-filter:blur(6px);
  display:none;align-items:center;justify-content:center;z-index:100;
  padding:16px;
}
.modal-mask.open{display:flex}
.modal-box{
  background:var(--surface);border:1px solid rgba(255,255,255,0.15);
  border-radius:var(--radius-lg);box-shadow:var(--shadow-lg);
  max-width:680px;width:100%;max-height:90vh;display:flex;flex-direction:column;
  overflow:hidden;animation:modalIn 0.2s cubic-bezier(0.16,1,0.3,1);
}
@keyframes modalIn{
  from{opacity:0;transform:scale(0.96)}
  to{opacity:1;transform:scale(1)}
}
.modal-head{
  padding:14px 20px;border-bottom:1px solid var(--border);
  display:flex;align-items:center;justify-content:space-between;
}
.modal-head h3{margin:0;font-size:15px;font-weight:700;color:#fff}
.modal-body{padding:18px 20px;overflow-y:auto;display:flex;flex-direction:column;gap:14px}
.modal-foot{
  padding:12px 20px;border-top:1px solid var(--border);
  display:flex;align-items:center;justify-content:flex-end;gap:10px;
  background:rgba(0,0,0,0.15);
}

.form-group{display:flex;flex-direction:column;gap:5px}
.form-label{font-size:11px;font-weight:600;color:var(--text-muted);display:flex;align-items:center;justify-content:space-between}
.form-grid-2{display:grid;grid-template-columns:1fr 1fr;gap:10px}
.form-grid-3{display:grid;grid-template-columns:1fr 1fr 1fr;gap:10px}

/* Toast */
.toast-box{
  position:fixed;bottom:24px;right:24px;z-index:110;
  display:flex;flex-direction:column;gap:8px;pointer-events:none;
}
.toast{
  background:var(--surface-3);color:#fff;border:1px solid var(--border);
  padding:9px 14px;border-radius:var(--radius-sm);box-shadow:var(--shadow);
  font-size:12px;font-weight:500;display:flex;align-items:center;gap:8px;
  animation:toastIn 0.2s ease;pointer-events:auto;
}
@keyframes toastIn{from{opacity:0;transform:translateY(8px)}to{opacity:1;transform:translateY(0)}}
.toast.success{border-color:rgba(16,185,129,0.5);background:#064e3b}
.toast.error{border-color:rgba(244,63,94,0.5);background:#881337}
.toast.info{border-color:rgba(59,130,246,0.5);background:#1e3a8a}

/* Video item in library */
.video-card-item{
  display:flex;align-items:center;justify-content:space-between;gap:10px;
  padding:8px 10px;background:var(--surface-2);border-radius:var(--radius-sm);
  margin-bottom:6px;border:1px solid var(--border);
}

/* Records Table */
.rec-table{width:100%;border-collapse:collapse;font-size:12px;text-align:left}
.rec-table th{background:var(--surface-3);color:var(--text-muted);font-weight:600;padding:8px 10px;border-bottom:1px solid var(--border)}
.rec-table td{padding:8px 10px;border-bottom:1px solid var(--border);font-family:var(--font-mono);font-size:11px}
.rec-table tr:hover td{background:rgba(255,255,255,0.03)}
</style>
</head>
<body>

<header class="header">
  <div class="brand">
    <div class="brand-icon">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
    </div>
    <div class="brand-info">
      <h1>GB28181 多设备模拟管控平台</h1>
      <div class="ver">GB/T 28181-2016 模拟器 · 纯 JSON 持久化版</div>
    </div>
  </div>
  <div class="header-actions">
    <button class="btn btn-primary" onclick="openNewDeviceModal()">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg> 新建模拟设备
    </button>
    <button class="btn btn-success btn-sm" onclick="startAllDevices()">
      <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"/></svg> 全部启动
    </button>
    <button class="btn btn-danger btn-sm" onclick="stopAllDevices()">
      <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><rect x="6" y="6" width="12" height="12"/></svg> 全部停止
    </button>
    <button class="btn btn-sm" onclick="openVideosModal()">
      <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="2" width="20" height="20" rx="2.18"/><line x1="7" y1="2" x2="7" y2="22"/><line x1="17" y1="2" x2="17" y2="22"/><line x1="2" y1="12" x2="22" y2="12"/></svg> 公共视频库
    </button>
    <button class="btn btn-sm" onclick="refreshAll(true)" title="刷新状态">
      <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/></svg>
    </button>
  </div>
</header>

<div class="container">

  <!-- Overview Stats -->
  <div class="stats-grid">
    <div class="stat-card">
      <div class="stat-label">模拟设备总数</div>
      <div class="stat-val" id="statTotalDevices">-</div>
    </div>
    <div class="stat-card">
      <div class="stat-label">运行中 / 已注册平台</div>
      <div class="stat-val" id="statOnlineDevices">-</div>
    </div>
    <div class="stat-card">
      <div class="stat-label">总通道数</div>
      <div class="stat-val" id="statTotalChannels">-</div>
    </div>
    <div class="stat-card">
      <div class="stat-label">当前活跃推流会话</div>
      <div class="stat-val" id="statTotalSessions">-</div>
    </div>
  </div>

  <!-- Multi-Device Cards Grid -->
  <div>
    <div class="section-head">
      <div class="section-title">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
        模拟设备列表 (点击卡片切换详情)
      </div>
      <div style="font-size:12px;color:var(--text-dim)">
        存储目录：<code>data/devices/*.json</code>（免安装、原子读写）
      </div>
    </div>
    <div class="device-grid" id="deviceGridContainer">
      <!-- Device Cards rendered here -->
    </div>
  </div>

  <!-- Active Device Workbench -->
  <div class="panel" id="deviceWorkbench">
    <div class="panel-head">
      <div style="display:flex;align-items:center;gap:12px;flex-wrap:wrap">
        <h2 id="workbenchTitle">设备管控工作台</h2>
        <div id="workbenchBadges" style="display:flex;gap:6px"></div>
      </div>
      <div class="tab-group">
        <button class="tab-btn active" onclick="switchWorkbenchTab('channels', this)">通道列表与视频源</button>
        <button class="tab-btn" onclick="switchWorkbenchTab('sessions', this)">实时点播会话</button>
        <button class="tab-btn" onclick="switchWorkbenchTab('logs', this)">设备运行日志</button>
        <button class="tab-btn" onclick="switchWorkbenchTab('records', this)">虚拟录像排程与查询</button>
      </div>
    </div>

    <!-- Tab 1: Channels -->
    <div class="panel-body" id="tabChannels">
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:14px;flex-wrap:wrap;gap:10px">
        <div style="display:flex;align-items:center;gap:8px">
          <span style="font-size:12px;color:var(--text-muted)">媒体分发模式:</span>
          <button class="btn btn-sm" id="btnModeShared" onclick="setDeviceMediaMode('shared')">全通道共用</button>
          <button class="btn btn-sm" id="btnModePerChannel" onclick="setDeviceMediaMode('per_channel')">按通道独立指定</button>
        </div>
        <div style="display:flex;gap:8px">
          <button class="btn btn-sm btn-primary" onclick="openAddChannelModal()">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg> 新增通道
          </button>
          <button class="btn btn-sm" onclick="triggerManualRegister()">立即重新注册</button>
          <button class="btn btn-sm" onclick="triggerManualKeepalive()">发送心跳</button>
          <button class="btn btn-sm btn-danger" onclick="openAlarmModal()">模拟报警</button>
        </div>
      </div>
      <div class="channel-grid" id="channelListContainer">
        <!-- Channel cards rendered here -->
      </div>
    </div>

    <!-- Tab 2: Sessions -->
    <div class="panel-body" id="tabSessions" style="display:none">
      <div id="sessionListContainer">
        <!-- Sessions rendered here -->
      </div>
    </div>

    <!-- Tab 3: Logs -->
    <div class="panel-body" id="tabLogs" style="display:none">
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:10px;gap:10px;flex-wrap:wrap">
        <div style="display:flex;align-items:center;gap:8px;flex:1;max-width:320px">
          <input type="text" id="logKeyword" placeholder="输入关键字过滤日志..." onkeydown="if(event.key==='Enter')loadLogs()"/>
          <button class="btn btn-sm" onclick="loadLogs()">搜索</button>
        </div>
        <div style="display:flex;align-items:center;gap:10px">
          <label style="font-size:11px;color:var(--text-dim);display:flex;align-items:center;gap:4px">
            <input type="checkbox" id="autoScroll" checked/> 自动滚到底部
          </label>
          <button class="btn btn-sm" onclick="exportLogs()">导出日志</button>
        </div>
      </div>
      <div class="log-box" id="logBox">
        <!-- Logs rendered here -->
      </div>
    </div>

    <!-- Tab 4: Records -->
    <div class="panel-body" id="tabRecords" style="display:none">
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:12px;gap:10px;flex-wrap:wrap">
        <div style="display:flex;align-items:center;gap:8px;flex-wrap:wrap;flex:1">
          <label style="font-size:12px;color:var(--text-muted)">通道:</label>
          <select id="recFilterChannel" style="min-width:160px"></select>
          <label style="font-size:12px;color:var(--text-muted)">起始时间:</label>
          <input type="text" id="recFilterStart" placeholder="YYYY-MM-DD HH:mm:ss" style="width:160px"/>
          <label style="font-size:12px;color:var(--text-muted)">结束时间:</label>
          <input type="text" id="recFilterEnd" placeholder="YYYY-MM-DD HH:mm:ss" style="width:160px"/>
          <label style="font-size:12px;color:var(--text-muted)">类型:</label>
          <select id="recFilterType" style="width:90px">
            <option value="all">全部 (all)</option>
            <option value="time">定时 (time)</option>
            <option value="alarm">报警 (alarm)</option>
          </select>
          <button class="btn btn-sm btn-primary" onclick="loadDeviceRecords()">检索录像</button>
        </div>
        <div>
          <button class="btn btn-sm" onclick="openDeviceConfigModal(activeDeviceId, 'record')">⚙️ 修改录像排程配置</button>
        </div>
      </div>
      <div id="recSummaryBar" style="margin-bottom:10px;font-size:12px;color:var(--text-muted);display:flex;gap:16px;align-items:center;padding:8px 12px;background:var(--surface-2);border-radius:var(--radius-sm);flex-wrap:wrap">
        <span>当前排程模式: <b id="recSummaryMode" style="color:var(--text-main)">-</b></span>
        <span>切片时长: <b id="recSummarySlice" style="color:var(--text-main)">-</b></span>
        <span>历史保留: <b id="recSummaryRetain" style="color:var(--text-main)">-</b></span>
        <span>检索结果: <b id="recSummaryCount" style="color:#6ee7b7">0 段</b></span>
      </div>
      <div id="recordListContainer" style="overflow-x:auto">
        <!-- Records table rendered here -->
      </div>
    </div>
  </div>

</div>

<!-- Modal: New Device -->
<div class="modal-mask" id="newDeviceModal">
  <div class="modal-box">
    <div class="modal-head">
      <h3>新建 GB28181 模拟设备</h3>
      <button class="btn btn-sm" onclick="closeModal('newDeviceModal')">✕</button>
    </div>
    <div class="modal-body">
      <div class="form-grid-2">
        <div class="form-group">
          <label class="form-label">设备国标编码 (20位) *</label>
          <input type="text" id="newDevId" placeholder="例如 34020000001180000002" maxlength="20"/>
        </div>
        <div class="form-group">
          <label class="form-label">设备名称 *</label>
          <input type="text" id="newDevName" placeholder="例如 模拟NVR-02"/>
        </div>
      </div>
      <div class="form-grid-2">
        <div class="form-group">
          <label class="form-label">平台 SIP 地址 (WVP IP) *</label>
          <input type="text" id="newDevServerIp" placeholder="192.168.1.100"/>
        </div>
        <div class="form-group">
          <label class="form-label">平台 SIP 端口 *</label>
          <input type="number" id="newDevServerPort" value="5060"/>
        </div>
      </div>
      <div class="form-grid-3">
        <div class="form-group">
          <label class="form-label">本机 SIP IP *</label>
          <input type="text" id="newDevLocalIp" placeholder="192.168.1.20"/>
        </div>
        <div class="form-group">
          <label class="form-label">本机 SIP 端口 (自动推荐)</label>
          <input type="number" id="newDevLocalPort" placeholder="5071"/>
        </div>
        <div class="form-group">
          <label class="form-label">信令传输协议</label>
          <select id="newDevTransport">
            <option value="udp">UDP</option>
            <option value="tcp">TCP</option>
          </select>
        </div>
      </div>
      <div class="form-grid-2">
        <div class="form-group">
          <label class="form-label">SIP 鉴权密码 *</label>
          <input type="text" id="newDevPassword" value="12345678"/>
        </div>
        <div class="form-group">
          <label class="form-label">初始下挂通道数</label>
          <select id="newDevInitChannels">
            <option value="1">生成 1 个初始 IPC 通道</option>
            <option value="2">生成 2 个初始 IPC 通道</option>
            <option value="4">生成 4 个初始 IPC 通道</option>
            <option value="0">暂不生成通道</option>
          </select>
        </div>
      </div>
      <div style="font-size:11px;color:var(--text-dim)">
        新建后将自动生成独立的 <code>data/devices/{设备ID}.json</code>，可随时自定义。
      </div>
    </div>
    <div class="modal-foot">
      <button class="btn" onclick="closeModal('newDeviceModal')">取消</button>
      <button class="btn btn-primary" onclick="submitCreateDevice()">立即创建</button>
    </div>
  </div>
</div>

<!-- Modal: Device Full Configuration (Web 端自定义所有参数) -->
<div class="modal-mask" id="deviceConfigModal">
  <div class="modal-box" style="max-width:760px">
    <div class="modal-head">
      <h3 id="devCfgModalTitle">自定义设备所有信息</h3>
      <button class="btn btn-sm" onclick="closeModal('deviceConfigModal')">✕</button>
    </div>
    <div class="modal-body">
      <div class="tab-group" style="margin-bottom:6px">
        <button class="tab-btn active" onclick="switchCfgTab('sip', this)">1. 平台 SIP 对接参数</button>
        <button class="tab-btn" onclick="switchCfgTab('device', this)">2. 设备国标身份</button>
        <button class="tab-btn" onclick="switchCfgTab('media', this)">3. 默认媒体参数</button>
        <button class="tab-btn" onclick="switchCfgTab('record', this)">4. 虚拟录像排程</button>
      </div>

      <!-- Tab: SIP -->
      <div id="cfgTabSip" class="cfg-tab-pane">
        <div class="form-grid-2">
          <div class="form-group">
            <label class="form-label">平台 SIP 服务端 IP</label>
            <input type="text" id="cfgSipServerIp"/>
          </div>
          <div class="form-group">
            <label class="form-label">平台 SIP 服务端端口</label>
            <input type="number" id="cfgSipServerPort"/>
          </div>
        </div>
        <div class="form-grid-3">
          <div class="form-group">
            <label class="form-label">本机 SIP IP (设备地址)</label>
            <input type="text" id="cfgSipLocalIp"/>
          </div>
          <div class="form-group">
            <label class="form-label">本机 SIP 监听端口</label>
            <input type="number" id="cfgSipLocalPort"/>
          </div>
          <div class="form-group">
            <label class="form-label">传输协议</label>
            <select id="cfgSipTransport">
              <option value="udp">UDP</option>
              <option value="tcp">TCP</option>
            </select>
          </div>
        </div>
        <div class="form-grid-2">
          <div class="form-group">
            <label class="form-label">SIP 注册用户名</label>
            <input type="text" id="cfgSipUsername"/>
          </div>
          <div class="form-group">
            <label class="form-label">SIP 鉴权密码</label>
            <input type="text" id="cfgSipPassword"/>
          </div>
        </div>
        <div class="form-grid-3">
          <div class="form-group">
            <label class="form-label">注册有效期 (秒)</label>
            <input type="number" id="cfgSipExpires"/>
          </div>
          <div class="form-group">
            <label class="form-label">心跳周期 (秒)</label>
            <input type="number" id="cfgSipKeepalive"/>
          </div>
          <div class="form-group">
            <label class="form-label">超时断开次数</label>
            <input type="number" id="cfgSipTimeoutCount"/>
          </div>
        </div>
      </div>

      <!-- Tab: Device -->
      <div id="cfgTabDevice" class="cfg-tab-pane" style="display:none">
        <div class="form-grid-2">
          <div class="form-group">
            <label class="form-label">设备国标编号 (20位编码)</label>
            <input type="text" id="cfgDevId" disabled style="opacity:0.7"/>
          </div>
          <div class="form-group">
            <label class="form-label">设备名称</label>
            <input type="text" id="cfgDevName"/>
          </div>
        </div>
        <div class="form-grid-2">
          <div class="form-group">
            <label class="form-label">行政区划 / 域 (Domain)</label>
            <input type="text" id="cfgDevDomain"/>
          </div>
          <div class="form-group">
            <label class="form-label">设备制造厂商 (Manufacturer)</label>
            <input type="text" id="cfgDevManufacturer"/>
          </div>
        </div>
        <div class="form-grid-2">
          <div class="form-group">
            <label class="form-label">设备型号 (Model)</label>
            <input type="text" id="cfgDevModel"/>
          </div>
          <div class="form-group">
            <label class="form-label">固件版本 (Firmware)</label>
            <input type="text" id="cfgDevFirmware"/>
          </div>
        </div>
      </div>

      <!-- Tab: Media -->
      <div id="cfgTabMedia" class="cfg-tab-pane" style="display:none">
        <div class="form-grid-2">
          <div class="form-group">
            <label class="form-label">媒体模式</label>
            <select id="cfgMediaMode">
              <option value="per_channel">按通道独立指定 (per_channel)</option>
              <option value="shared">全通道共用 (shared)</option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label">全局默认视频源类型</label>
            <select id="cfgMediaSource">
              <option value="mp4">本地 MP4 视频文件 (推荐)</option>
              <option value="ptz">🕹️ PTZ 3D 虚拟全景动态流 (随云台转动实时渲染)</option>
              <option value="file">Annex-B H.264 原生文件</option>
              <option value="synthetic">内置合成彩条流</option>
            </select>
          </div>
        </div>
        <div class="form-grid-2">
          <div class="form-group">
            <label class="form-label">全局默认 MP4 路径</label>
            <input type="text" id="cfgMediaMp4"/>
          </div>
          <div class="form-group">
            <label class="form-label">全局默认 H264 路径</label>
            <input type="text" id="cfgMediaH264"/>
          </div>
        </div>
        <div class="form-grid-3">
          <div class="form-group">
            <label class="form-label">分辨率 宽度</label>
            <input type="number" id="cfgMediaWidth"/>
          </div>
          <div class="form-group">
            <label class="form-label">分辨率 高度</label>
            <input type="number" id="cfgMediaHeight"/>
          </div>
          <div class="form-group">
            <label class="form-label">推流帧率 (FPS)</label>
            <input type="number" id="cfgMediaFps"/>
          </div>
        </div>
        <div class="form-grid-2">
          <div class="form-group">
            <label class="form-label">RTP 负载上限 (Payload Max)</label>
            <input type="number" id="cfgMediaPayloadMax"/>
          </div>
          <div class="form-group">
            <label class="form-label">媒体发送源地址 (Local IP)</label>
            <input type="text" id="cfgMediaLocalIp"/>
          </div>
        </div>
      </div>

      <!-- Tab: Record -->
      <div id="cfgTabRecord" class="cfg-tab-pane" style="display:none">
        <div class="form-grid-2">
          <div class="form-group">
            <label class="form-label">虚拟录像功能开关</label>
            <select id="cfgRecordEnabled">
              <option value="true">启用虚拟录像 (响应 RecordInfo 查询)</option>
              <option value="false">停用 (返回 0 条记录)</option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label">录像排程模式 (Record Mode)</label>
            <select id="cfgRecordMode">
              <option value="continuous">全天连续录像 (24 小时全覆盖)</option>
              <option value="work_hours">工作时段录像 (每日 08:00 ~ 18:00)</option>
              <option value="alarm">报警录像 (整点随机报警模拟)</option>
            </select>
          </div>
        </div>
        <div class="form-grid-2">
          <div class="form-group">
            <label class="form-label">切片时长 (Slice Minutes)</label>
            <select id="cfgRecordSliceMinutes">
              <option value="30">30 分钟 / 段</option>
              <option value="60">60 分钟 / 段 (推荐)</option>
              <option value="120">120 分钟 / 段 (2 小时)</option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label">历史保留天数 (Retain Days)</label>
            <input type="number" id="cfgRecordRetainDays" value="7" min="1" max="365"/>
            <span style="font-size:11px;color:var(--text-dim)">超出保留天数的历史录像将自动模拟被覆盖清除</span>
          </div>
        </div>
        <div class="form-grid-2">
          <div class="form-group">
            <label class="form-label">录像类型标识 (Type)</label>
            <select id="cfgRecordType">
              <option value="time">time (定时录像，通用标准)</option>
              <option value="alarm">alarm (报警录像)</option>
              <option value="all">all (全部录像)</option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label">单段虚拟文件大小 (MB)</label>
            <input type="number" id="cfgRecordFileSizeMB" value="100" min="1"/>
            <span style="font-size:11px;color:var(--text-dim)">返回给平台的虚拟 FileSize (字节数 = MB * 1024 * 1024)</span>
          </div>
        </div>
        <div style="margin-top:12px;padding:10px 14px;background:rgba(59,130,246,0.08);border-radius:var(--radius-sm);border:1px solid rgba(59,130,246,0.2);font-size:12px;color:var(--text-muted)">
          💡 <b>提示：</b>保存后，国标平台（如 WVP、LiveGBS 等）通过 GB/T 28181 向该设备检索录像时，模拟器将根据此排程自动计算切片、按国标规范分页返回；发起历史录像回放 (INVITE Playback) 时亦按此排程对齐时间戳。
        </div>
      </div>

      <input type="hidden" id="cfgActiveDevId"/>
    </div>
    <div class="modal-foot">
      <button class="btn" onclick="closeModal('deviceConfigModal')">取消</button>
      <button class="btn btn-primary" onclick="saveDeviceConfig(false)">保存配置 (存入JSON)</button>
      <button class="btn btn-success" onclick="saveDeviceConfig(true)">保存并重启/重新注册</button>
    </div>
  </div>
</div>

<!-- Modal: Add Channel -->
<div class="modal-mask" id="addChannelModal">
  <div class="modal-box">
    <div class="modal-head">
      <h3>新增下挂 IPC 通道</h3>
      <button class="btn btn-sm" onclick="closeModal('addChannelModal')">✕</button>
    </div>
    <div class="modal-body">
      <div class="form-group">
        <label class="form-label">通道国标编号 (20位编码) *</label>
        <input type="text" id="addChId" placeholder="例如 34020000001320000001" maxlength="20"/>
      </div>
      <div class="form-group">
        <label class="form-label">通道名称</label>
        <input type="text" id="addChName" placeholder="例如 通道01"/>
      </div>
      <div class="form-group">
        <label class="form-label">绑定点播视频源</label>
        <select id="addChVideo">
          <!-- Video options rendered dynamically -->
        </select>
      </div>
    </div>
    <div class="modal-foot">
      <button class="btn" onclick="closeModal('addChannelModal')">取消</button>
      <button class="btn btn-primary" onclick="submitAddChannel()">确认新增</button>
    </div>
  </div>
</div>

<!-- Modal: Videos Library -->
<div class="modal-mask" id="videosModal">
  <div class="modal-box" style="max-width:700px">
    <div class="modal-head">
      <h3>公共视频素材库 (assets)</h3>
      <button class="btn btn-sm" onclick="closeModal('videosModal')">✕</button>
    </div>
    <div class="modal-body">
      <div style="display:flex;gap:10px;align-items:center;background:var(--surface-2);padding:10px 14px;border-radius:var(--radius-sm);border:1px solid var(--border)">
        <input type="file" id="uploadInput" accept=".mp4,.h264" style="display:none" onchange="handleFileSelected()"/>
        <button class="btn btn-sm btn-primary" onclick="document.getElementById('uploadInput').click()">选择本地视频</button>
        <input type="text" id="uploadFileName" readonly placeholder="未选择文件" style="flex:1"/>
        <button class="btn btn-sm btn-success" id="uploadBtn" onclick="uploadVideoFile()">上传视频</button>
      </div>
      <div id="videoListContainer" style="max-height:380px;overflow-y:auto;margin-top:6px">
        <!-- Video list rendered here -->
      </div>
    </div>
    <div class="modal-foot">
      <button class="btn" onclick="closeModal('videosModal')">关闭</button>
    </div>
  </div>
</div>

<!-- Modal: Video Preview -->
<div class="modal-mask" id="videoPreviewModal">
  <div class="modal-box" style="max-width:680px">
    <div class="modal-head">
      <h3 id="videoPreviewTitle">在线视频预览</h3>
      <button class="btn btn-sm" onclick="closeVideoModal()">✕</button>
    </div>
    <div class="modal-body" style="padding:10px;background:#000">
      <video id="previewPlayer" controls style="width:100%;max-height:420px;display:block"></video>
    </div>
    <div class="modal-foot">
      <button class="btn" onclick="closeVideoModal()">关闭</button>
    </div>
  </div>
</div>

<!-- Modal: Alarm -->
<div class="modal-mask" id="alarmModal">
  <div class="modal-box">
    <div class="modal-head">
      <h3>发送国标模拟报警通知</h3>
      <button class="btn btn-sm" onclick="closeModal('alarmModal')">✕</button>
    </div>
    <div class="modal-body">
      <div class="form-group">
        <label class="form-label">报警通道</label>
        <select id="alarmChannelSelect"></select>
      </div>
      <div class="form-grid-2">
        <div class="form-group">
          <label class="form-label">报警方式 (AlarmMethod)</label>
          <select id="alarmMethodSelect">
            <option value="2">2 - 移动侦测报警</option>
            <option value="1">1 - 电话线报警</option>
            <option value="3">3 - 视频丢失报警</option>
            <option value="4">4 - 视频遮挡报警</option>
            <option value="5">5 - 外部探测器报警</option>
          </select>
        </div>
        <div class="form-group">
          <label class="form-label">报警级别 (Priority)</label>
          <select id="alarmPrioritySelect">
            <option value="4">4 - 低级 (默认)</option>
            <option value="3">3 - 中级</option>
            <option value="2">2 - 高级</option>
            <option value="1">1 - 一级 (最高)</option>
          </select>
        </div>
      </div>
      <div class="form-group">
        <label class="form-label">报警描述文本</label>
        <input type="text" id="alarmDescInput" value="Web 控制台触发模拟移动侦测"/>
      </div>
    </div>
    <div class="modal-foot">
      <button class="btn" onclick="closeModal('alarmModal')">取消</button>
      <button class="btn btn-danger" onclick="submitAlarm()">立即发送报警通知</button>
    </div>
  </div>
</div>

<!-- Modal: PTZ Control & Presets -->
<div class="modal-mask" id="ptzModal">
  <div class="modal-box" style="max-width:980px">
    <div class="modal-head">
      <div style="display:flex;align-items:center;gap:10px">
        <span style="font-size:20px">🕹️</span>
        <div>
          <h3 id="ptzModalTitle" style="margin:0;font-size:15px">虚拟云台姿态与预置位操控</h3>
          <div id="ptzModalSubtitle" style="font-size:11px;color:var(--text-dim)">通道ID: -</div>
        </div>
      </div>
      <button class="btn btn-sm" onclick="closePTZModal()">✕</button>
    </div>
    <div class="modal-body">
      <!-- 3D 虚拟全景监控视口 -->
      <div style="background:var(--bg-dark);border-radius:var(--radius);border:1px solid var(--border);position:relative;overflow:hidden;margin-bottom:14px;box-shadow:inset 0 0 40px rgba(0,0,0,0.85)">
        <!-- 视口顶部 OSD 栏 -->
        <div style="position:absolute;top:8px;left:12px;right:12px;display:flex;justify-content:space-between;align-items:center;pointer-events:none;z-index:3">
          <div style="display:flex;align-items:center;gap:8px">
            <span class="badge live" style="background:rgba(16,185,129,0.25);border-color:rgba(16,185,129,0.5);color:#6ee7b7;padding:2px 8px;font-size:10px">
              <span class="dot"></span>LIVE 虚拟全景视口
            </span>
            <span id="ptzCamOsdTag" style="font-size:11px;font-family:var(--font-mono);color:#cbd5e1;background:rgba(0,0,0,0.6);padding:2px 6px;border-radius:4px;border:1px solid rgba(255,255,255,0.1)">
              CAM-01 · 1080P@25FPS
            </span>
            <span id="ptzStreamSourceBadge" class="badge" style="font-size:10px;padding:2px 8px;cursor:pointer;pointer-events:auto" onclick="togglePTZChannelStreamSource()" title="点击一键切换该通道推送给WVP的视频源 (MP4录像 / PTZ虚拟动态流)">
              国标推流: 检测中...
            </span>
          </div>
          <div style="display:flex;align-items:center;gap:6px">
            <button class="btn btn-sm" onclick="snapshotPTZCanvas()" style="pointer-events:auto;background:rgba(30,41,59,0.85);backdrop-filter:blur(4px);font-size:11px;padding:3px 8px" title="抓拍保存当前视口画面">📸 抓拍快照</button>
            <button class="btn btn-sm" onclick="resetPTZView()" style="pointer-events:auto;background:rgba(30,41,59,0.85);backdrop-filter:blur(4px);font-size:11px;padding:3px 8px" title="将视口回正到正北 (0°) 平视">🎯 视口回正</button>
          </div>
        </div>

        <!-- 3D 视口画布 -->
        <canvas id="ptzViewportCanvas" width="940" height="270" style="display:block;width:100%;height:270px;cursor:grab;background:#050914"></canvas>

        <!-- 视口底部提示与目标锁定栏 -->
        <div style="position:absolute;bottom:6px;left:12px;right:12px;display:flex;justify-content:space-between;align-items:center;pointer-events:none;z-index:3;font-size:11px;color:rgba(255,255,255,0.7);font-family:var(--font-mono)">
          <span id="ptzOsdTime">2026-09-12 18:45:00.000</span>
          <span style="color:var(--accent);font-size:10px">🖱️ 画面内支持鼠标拖拽旋转云台 · 滚轮缩放变倍</span>
          <span id="ptzOsdTarget" style="color:#38bdf8;font-weight:600">目标: 园区正门 (0°)</span>
        </div>
      </div>

      <div style="display:grid;grid-template-columns:1.1fr 1fr;gap:18px">
      <!-- 左侧：云台姿态与控制 -->
      <div style="background:var(--surface-2);border-radius:var(--radius);padding:14px;border:1px solid var(--border);display:flex;flex-direction:column;gap:12px">
        <div style="display:flex;justify-content:space-between;align-items:center">
          <span style="font-weight:700;font-size:12px;color:var(--text-main);letter-spacing:0.04em">球机当前姿态</span>
          <span id="ptzStatusBadge" class="badge"><span class="dot"></span><span id="ptzStatusText">检测中...</span></span>
        </div>

        <!-- 仪表显示卡片 -->
        <div style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:8px;text-align:center">
          <div style="background:var(--bg-dark);padding:8px 4px;border-radius:var(--radius-sm);border:1px solid var(--border)">
            <div style="font-size:10px;color:var(--text-dim);margin-bottom:2px">航向角 (Pan)</div>
            <div style="font-size:17px;font-weight:700;color:var(--cyan);font-family:var(--font-mono)" id="ptzValPan">0.0°</div>
            <div style="font-size:10px;color:var(--text-muted);margin-top:2px" id="ptzPanDir">正北 (0°)</div>
          </div>
          <div style="background:var(--bg-dark);padding:8px 4px;border-radius:var(--radius-sm);border:1px solid var(--border)">
            <div style="font-size:10px;color:var(--text-dim);margin-bottom:2px">俯仰角 (Tilt)</div>
            <div style="font-size:17px;font-weight:700;color:var(--ok);font-family:var(--font-mono)" id="ptzValTilt">0.0°</div>
            <div style="font-size:10px;color:var(--text-muted);margin-top:2px" id="ptzTiltDir">水平平视</div>
          </div>
          <div style="background:var(--bg-dark);padding:8px 4px;border-radius:var(--radius-sm);border:1px solid var(--border)">
            <div style="font-size:10px;color:var(--text-dim);margin-bottom:2px">变倍比 (Zoom)</div>
            <div style="font-size:17px;font-weight:700;color:var(--purple);font-family:var(--font-mono)" id="ptzValZoom">1.0x</div>
            <div style="font-size:10px;color:var(--text-muted);margin-top:2px">广角全景</div>
          </div>
        </div>

        <!-- 8 方向物理操控盘 -->
        <div style="display:flex;flex-direction:column;align-items:center;gap:10px;margin-top:2px">
          <div class="ptz-pad-grid">
            <button class="ptz-btn" onmousedown="startPTZ('upleft')" onmouseup="stopPTZ()" title="左上">↖</button>
            <button class="ptz-btn" onmousedown="startPTZ('up')" onmouseup="stopPTZ()" title="向上">▲</button>
            <button class="ptz-btn" onmousedown="startPTZ('upright')" onmouseup="stopPTZ()" title="右上">↗</button>
            <button class="ptz-btn" onmousedown="startPTZ('left')" onmouseup="stopPTZ()" title="向左">◀</button>
            <button class="ptz-btn stop" onclick="stopPTZ()" title="急停">■</button>
            <button class="ptz-btn" onmousedown="startPTZ('right')" onmouseup="stopPTZ()" title="向右">▶</button>
            <button class="ptz-btn" onmousedown="startPTZ('downleft')" onmouseup="stopPTZ()" title="左下">↙</button>
            <button class="ptz-btn" onmousedown="startPTZ('down')" onmouseup="stopPTZ()" title="向下">▼</button>
            <button class="ptz-btn" onmousedown="startPTZ('downright')" onmouseup="stopPTZ()" title="右下">↘</button>
          </div>

          <!-- 镜头变倍按键 -->
          <div style="display:flex;gap:8px;width:100%;justify-content:center">
            <button class="btn btn-sm" onmousedown="startPTZ('zoomin')" onmouseup="stopPTZ()" style="flex:1">🔍 放大 (+)</button>
            <button class="btn btn-sm" onmousedown="startPTZ('zoomout')" onmouseup="stopPTZ()" style="flex:1">🔎 缩小 (-)</button>
          </div>

          <!-- 速度调节滑块 -->
          <div style="width:100%;background:var(--bg-dark);padding:8px 12px;border-radius:var(--radius-sm);border:1px solid var(--border)">
            <div style="display:flex;justify-content:space-between;font-size:11px;margin-bottom:4px">
              <span style="color:var(--text-dim)">手动转动速度 (0~255)</span>
              <span id="ptzSpeedLabel" style="color:var(--accent);font-weight:700">80</span>
            </div>
            <input type="range" id="ptzSpeedSlider" min="10" max="255" value="80" style="width:100%" oninput="document.getElementById('ptzSpeedLabel').innerText = this.value"/>
          </div>
        </div>
      </div>

      <!-- 右侧：预置位管理 (PresetList) -->
      <div style="background:var(--surface-2);border-radius:var(--radius);padding:14px;border:1px solid var(--border);display:flex;flex-direction:column;gap:10px">
        <div style="display:flex;justify-content:space-between;align-items:center">
          <span style="font-weight:700;font-size:12px;color:var(--text-main);letter-spacing:0.04em">预置位列表 (PresetList)</span>
          <button class="btn btn-sm btn-primary" onclick="promptSaveCurrentPreset()">+ 存当前姿态为预置位</button>
        </div>

        <div style="flex:1;min-height:240px;max-height:340px;overflow-y:auto;border:1px solid var(--border);border-radius:var(--radius-sm);background:var(--bg-dark)">
          <table style="width:100%;border-collapse:collapse;font-size:12px" id="ptzPresetTable">
            <thead>
              <tr style="border-bottom:1px solid var(--border);background:var(--surface-3);color:var(--text-dim);text-align:left">
                <th style="padding:6px 8px">ID</th>
                <th style="padding:6px 8px">预置位名称</th>
                <th style="padding:6px 8px">坐标 (P/T/Z)</th>
                <th style="padding:6px 8px;text-align:right">操作</th>
              </tr>
            </thead>
            <tbody id="ptzPresetTbody">
              <tr><td colspan="4" style="padding:20px;text-align:center;color:var(--text-dim)">加载预置位列表中...</td></tr>
            </tbody>
          </table>
        </div>

        <div style="font-size:11px;color:var(--text-dim);line-height:1.5;background:rgba(59,130,246,0.06);padding:8px 10px;border-radius:var(--radius-sm);border:1px solid rgba(59,130,246,0.2)">
          💡 <b>平台联动说明</b>：平台（如 WVP）下发 <code>PresetQuery</code> 会查询此列表；平台下发 <code>0x81</code> 设置、<code>0x82</code> 调用、<code>0x83</code> 删除预置位时，此处将实时响应并自动持久化。
        </div>
      </div>
    </div>
    </div>
    <div class="modal-foot">
      <button class="btn" onclick="closePTZModal()">关闭</button>
    </div>
  </div>
</div>

<div class="toast-box" id="toastBox"></div>

<script>
let devices = [];
let activeDeviceId = '';
let activeDeviceData = null;
let videoList = [];
let rawLogs = [];

function esc(s){
  if(s==null) return '';
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;').replace(/'/g,'&#39;');
}
function fmtSize(b){
  if(!b) return '0 B';
  if(b<1024) return b + ' B';
  if(b<1024*1024) return (b/1024).toFixed(1) + ' KB';
  return (b/(1024*1024)).toFixed(1) + ' MB';
}

function showToast(msg, type){
  const box = document.getElementById('toastBox');
  const d = document.createElement('div');
  d.className = 'toast ' + (type || 'info');
  d.innerHTML = '<span>' + esc(msg) + '</span>';
  box.appendChild(d);
  setTimeout(function(){
    d.style.opacity = '0';
    d.style.transition = 'opacity 0.3s ease';
    setTimeout(function(){ d.remove(); }, 300);
  }, 3200);
}

function openModal(id){ document.getElementById(id).classList.add('open'); }
function closeModal(id){ document.getElementById(id).classList.remove('open'); }

async function api(url, method, body){
  let opts = {method: 'GET'};
  if(typeof method === 'object' && method !== null){
    opts = Object.assign({}, method);
    if(opts.body && typeof opts.body === 'object' && !(opts.body instanceof FormData) && !(opts.body instanceof Blob)){
      opts.headers = Object.assign({'Content-Type': 'application/json'}, opts.headers || {});
      opts.body = JSON.stringify(opts.body);
    }
  } else {
    opts.method = method || 'GET';
    if(body !== undefined && body !== null){
      opts.headers = Object.assign({'Content-Type': 'application/json'}, opts.headers || {});
      opts.body = (typeof body === 'string') ? body : JSON.stringify(body);
    }
  }
  try{
    const r = await fetch(url, opts);
    const j = await r.json();
    if(!r.ok) throw new Error(j.error || ('请求失败 ('+r.status+')'));
    return j;
  }catch(e){
    showToast(e.message, 'error');
    return null;
  }
}

// Switch Workbench Tab
function switchWorkbenchTab(tab, btn){
  document.querySelectorAll('#deviceWorkbench .tab-btn').forEach(function(b){b.classList.remove('active')});
  btn.classList.add('active');
  document.getElementById('tabChannels').style.display = (tab==='channels' ? 'block' : 'none');
  document.getElementById('tabSessions').style.display = (tab==='sessions' ? 'block' : 'none');
  document.getElementById('tabLogs').style.display = (tab==='logs' ? 'block' : 'none');
  document.getElementById('tabRecords').style.display = (tab==='records' ? 'block' : 'none');
  if(tab==='logs') loadLogs();
  if(tab==='records') loadDeviceRecords();
}

// Switch Config Modal Tab
function switchCfgTab(tab, btn){
  document.querySelectorAll('#deviceConfigModal .tab-btn').forEach(function(b){b.classList.remove('active')});
  if(btn) btn.classList.add('active');
  document.getElementById('cfgTabSip').style.display = (tab==='sip' ? 'block' : 'none');
  document.getElementById('cfgTabDevice').style.display = (tab==='device' ? 'block' : 'none');
  document.getElementById('cfgTabMedia').style.display = (tab==='media' ? 'block' : 'none');
  document.getElementById('cfgTabRecord').style.display = (tab==='record' ? 'block' : 'none');
}

// Refresh Everything
async function refreshAll(manual){
  try{
    const j = await api('/api/devices');
    if(!j) return;
    devices = j.devices || [];

    // Overview Stats
    let onlineCount = 0;
    let chCount = 0;
    let sessCount = 0;
    devices.forEach(function(d){
      if(d.running && d.registered) onlineCount++;
      chCount += (d.channelCount || 0);
      sessCount += (d.activeSessions || 0);
    });

    document.getElementById('statTotalDevices').textContent = devices.length;
    document.getElementById('statOnlineDevices').textContent = onlineCount + ' / ' + devices.length;
    document.getElementById('statTotalChannels').textContent = chCount;
    document.getElementById('statTotalSessions').textContent = sessCount;

    // Active Device fallback
    if(!activeDeviceId && devices.length > 0){
      activeDeviceId = devices[0].id;
    } else if(devices.length > 0 && !devices.some(function(d){return d.id === activeDeviceId})){
      activeDeviceId = devices[0].id;
    }

    renderDeviceGrid();
    await loadActiveDeviceDetail();
    await loadVideos();

    if(manual) showToast('设备状态已刷新', 'success');
  }catch(e){
    if(manual) showToast('刷新异常: '+e.message, 'error');
  }
}

// Render Device Cards Grid
function renderDeviceGrid(){
  const container = document.getElementById('deviceGridContainer');
  if(!devices.length){
    container.innerHTML = '<div style="color:var(--text-dim);grid-column:1/-1;padding:24px;text-align:center;background:var(--surface);border-radius:var(--radius);border:1px dashed var(--border)">暂无模拟设备，点击上方「新建模拟设备」创建</div>';
    return;
  }

  container.innerHTML = devices.map(function(d){
    const isActive = (d.id === activeDeviceId);
    let badgeClass = 'stopped';
    let badgeText = '已停止';
    if(d.running){
      if(d.registered){ badgeClass = 'on'; badgeText = '已连接平台'; }
      else { badgeClass = 'warn'; badgeText = '运行中(未注册)'; }
    }
    if(d.error){ badgeClass = 'off'; badgeText = '异常'; }

    return '<div class="device-card ' + (isActive ? 'active' : '') + '" onclick="selectDevice(\'' + esc(d.id) + '\')">' +
      '<div class="device-card-head">' +
        '<div class="device-icon">' +
          '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>' +
        '</div>' +
        '<div class="device-card-title">' +
          '<div class="device-card-name">' + esc(d.name) + '</div>' +
          '<div class="device-card-id">' + esc(d.id) + '</div>' +
        '</div>' +
      '</div>' +
      '<div class="device-meta-grid">' +
        '<div class="meta-item"><span class="meta-k">状态</span><span><span class="badge ' + badgeClass + '"><span class="dot"></span>' + badgeText + '</span></span></div>' +
        '<div class="meta-item"><span class="meta-k">本地 SIP 端口</span><span class="meta-v">' + esc(d.localPort) + ' (' + esc(d.transport) + ')</span></div>' +
        '<div class="meta-item"><span class="meta-k">对接平台</span><span class="meta-v">' + esc(d.server) + '</span></div>' +
        '<div class="meta-item"><span class="meta-k">通道 / 会话</span><span class="meta-v">' + (d.channelCount||0) + ' 通道 · ' + (d.activeSessions||0) + ' 点播</span></div>' +
      '</div>' +
      '<div class="device-card-actions" onclick="event.stopPropagation()">' +
        (d.running ?
          ('<button class="btn btn-sm btn-danger" onclick="stopDevice(\'' + esc(d.id) + '\')">停止</button>' +
           '<button class="btn btn-sm" onclick="restartDevice(\'' + esc(d.id) + '\')">重启</button>') :
          ('<button class="btn btn-sm btn-success" onclick="startDevice(\'' + esc(d.id) + '\')">启动</button>')
        ) +
        '<button class="btn btn-sm btn-primary" onclick="openDeviceConfigModal(\'' + esc(d.id) + '\')">⚙️ 全量配置</button>' +
        '<button class="btn btn-sm" onclick="deleteDevice(\'' + esc(d.id) + '\')">🗑️ 删除</button>' +
      '</div>' +
    '</div>';
  }).join('');
}

function selectDevice(id){
  activeDeviceId = id;
  renderDeviceGrid();
  loadActiveDeviceDetail();
}

// Load Selected Device Detail
async function loadActiveDeviceDetail(){
  if(!activeDeviceId){
    document.getElementById('workbenchTitle').textContent = '未选择设备';
    document.getElementById('channelListContainer').innerHTML = '';
    return;
  }
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId));
  if(!j) return;
  activeDeviceData = j;

  const prof = j.profile || {};
  const devCfg = prof.device || {};
  const sipCfg = prof.sip || {};
  const st = j.status || {};

  document.getElementById('workbenchTitle').textContent = devCfg.name + ' (' + activeDeviceId + ')';
  
  let stBadge = j.running ?
    (st.registered ? '<span class="badge on"><span class="dot"></span>已连接平台</span>' : '<span class="badge warn"><span class="dot"></span>运行中(鉴权中)</span>') :
    '<span class="badge stopped"><span class="dot"></span>已停止</span>';

  document.getElementById('workbenchBadges').innerHTML = stBadge +
    '<span class="badge"><span class="dot"></span>本地端口: ' + esc(sipCfg.local_port) + '</span>' +
    '<span class="badge"><span class="dot"></span>平台: ' + esc(sipCfg.server_ip) + ':' + esc(sipCfg.server_port) + '</span>';

  // Mode buttons
  const isPerChannel = (prof.media && prof.media.mode === 'per_channel');
  document.getElementById('btnModeShared').className = 'btn btn-sm ' + (!isPerChannel ? 'btn-primary' : '');
  document.getElementById('btnModePerChannel').className = 'btn btn-sm ' + (isPerChannel ? 'btn-primary' : '');

  renderChannels(devCfg.channels || [], prof.media || {}, st.sessions || []);
  renderSessions(st.sessions || []);
  updateRecordChannelOptions(devCfg.channels || [], activeDeviceId);
  if(document.getElementById('tabRecords').style.display !== 'none'){
    loadDeviceRecords();
  }
}

// Render Channels for active device
function renderChannels(channels, mediaCfg, sessions){
  const container = document.getElementById('channelListContainer');
  if(!channels.length){
    container.innerHTML = '<div style="color:var(--text-dim);grid-column:1/-1;padding:20px;text-align:center">该设备暂无下挂通道，请点击上方「新增通道」</div>';
    return;
  }

  const liveChannelMap = {};
  (sessions||[]).forEach(function(s){ liveChannelMap[s.channelId] = s; });

  const mediaChannels = (mediaCfg && mediaCfg.channels) || {};

  container.innerHTML = channels.map(function(ch){
    const sess = liveChannelMap[ch.id];
    const isLive = !!sess;
    const chMedia = mediaChannels[ch.id] || {};
    const boundMp4 = chMedia.mp4_file || '';
    const boundH264 = chMedia.h264_file || '';
    const boundSrc = chMedia.source || (mediaCfg.source || 'mp4');

    let boundLabel = '全局默认';
    if(boundSrc === 'ptz') boundLabel = '🕹️ PTZ 虚拟流';
    else if(boundSrc === 'synthetic') boundLabel = '内置彩条流';
    else if(boundMp4) boundLabel = boundMp4.split('/').pop();
    else if(boundH264) boundLabel = boundH264.split('/').pop();

    return '<div class="ch-card ' + (isLive ? 'live' : '') + '">' +
      '<div class="ch-header">' +
        '<div>' +
          '<div class="ch-name">' + esc(ch.name) + '</div>' +
          '<div class="ch-id">' + esc(ch.id) + '</div>' +
        '</div>' +
        '<div>' +
          (isLive ? '<span class="badge live"><span class="dot"></span>推流中</span>' :
            (ch.status === 'ON' ? '<span class="badge on"><span class="dot"></span>在线</span>' : '<span class="badge off"><span class="dot"></span>离线</span>')
          ) +
        '</div>' +
      '</div>' +
      '<div style="font-size:11px;color:var(--text-dim);display:flex;flex-direction:column;gap:4px">' +
        '<div>当前视频源: <b style="color:var(--text-main)">' + esc(boundLabel) + '</b></div>' +
        (isLive ? ('<div>点播目标: <code>' + esc(sess.remoteIp) + ':' + esc(sess.remotePort) + '</code></div>') : '') +
      '</div>' +
      '<div style="display:flex;gap:6px;align-items:center;margin-top:6px;flex-wrap:wrap">' +
        '<select class="channel-bind-select" style="flex:1;min-width:140px">' +
          '<option value="">使用全局默认视频</option>' +
          '<option value="__ptz__" ' + (boundSrc==='ptz'?'selected':'') + '>🕹️ PTZ 虚拟全景流 (随云台转动)</option>' +
          '<option value="__synthetic__" ' + (boundSrc==='synthetic'?'selected':'') + '>内置彩条测试流</option>' +
          videoList.map(function(v){
            const sel = (boundMp4===v.path || boundH264===v.path || boundMp4===v.name) ? 'selected' : '';
            return '<option value="' + esc(v.path) + '" ' + sel + '>' + esc(v.name) + '</option>';
          }).join('') +
        '</select>' +
        '<button class="btn btn-sm btn-primary" onclick="bindChannelMedia(\'' + esc(ch.id) + '\', this)">绑定</button>' +
        '<button class="btn btn-sm" onclick="openPTZModal(\'' + esc(ch.id) + '\',\'' + esc(ch.name) + '\')">🕹️ 云台与预置位</button>' +
        '<button class="btn btn-sm" onclick="toggleChannelStatus(\'' + esc(ch.id) + '\',\'' + (ch.status==='ON'?'OFF':'ON') + '\')">' + (ch.status==='ON'?'设为离线':'设为在线') + '</button>' +
        '<button class="btn btn-sm btn-danger" onclick="removeChannel(\'' + esc(ch.id) + '\')">删除</button>' +
      '</div>' +
    '</div>';
  }).join('');
}

// Render Sessions for active device
function renderSessions(sessions){
  const container = document.getElementById('sessionListContainer');
  if(!sessions || !sessions.length){
    container.innerHTML = '<div style="color:var(--text-dim);padding:24px;text-align:center">当前无实时推流会话。在平台（如 WVP）上点击通道播放后，将在此展示推流状态。</div>';
    return;
  }
  container.innerHTML = sessions.map(function(s){
    return '<div class="session-item">' +
      '<div>' +
        '<div style="font-weight:700;color:#fff;font-size:13px">' + esc(s.channelId) + ' · <span class="badge live"><span class="dot"></span>' + esc(s.streamType||'live') + '</span></div>' +
        '<div style="font-size:11px;font-family:var(--font-mono);color:var(--text-dim);margin-top:4px">' +
          'SSRC: ' + esc(s.ssrc) + ' · 目标: ' + esc(s.remoteIp) + ':' + esc(s.remotePort) + ' (' + (s.isTcp?'TCP':'UDP') + ') · FPS: ' + (s.fps||25) +
        '</div>' +
      '</div>' +
      '<button class="btn btn-sm btn-danger" onclick="stopSession(\'' + esc(s.callId) + '\')">停止推流</button>' +
    '</div>';
  }).join('');
}

// Load Videos in assets
async function loadVideos(){
  const j = await api('/api/videos');
  if(j) videoList = j.videos || [];
  renderVideoList();
}

function renderVideoList(){
  const container = document.getElementById('videoListContainer');
  const addChSel = document.getElementById('addChVideo');
  if(addChSel){
    addChSel.innerHTML = '<option value="">暂不指定 (继承默认)</option>' +
      '<option value="__synthetic__">内置合成彩条流</option>' +
      videoList.map(function(v){
        return '<option value="' + esc(v.path) + '">' + esc(v.name) + ' (' + fmtSize(v.size) + ')</option>';
      }).join('');
  }

  if(!videoList.length){
    container.innerHTML = '<div style="color:var(--text-dim);padding:14px;text-align:center">素材库为空，请选择 MP4 / H264 文件上传</div>';
    return;
  }
  container.innerHTML = videoList.map(function(v){
    const isMp4 = v.name.toLowerCase().endsWith('.mp4');
    return '<div class="video-card-item">' +
      '<div style="min-width:0;flex:1">' +
        '<div style="font-weight:600;font-size:12px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">' + esc(v.name) + '</div>' +
        '<div style="font-size:11px;color:var(--text-dim)">' + fmtSize(v.size) + '</div>' +
      '</div>' +
      '<div style="display:flex;gap:6px">' +
        (isMp4 ? ('<button class="btn btn-sm btn-primary" onclick="previewVideo(\'' + esc(v.path) + '\',\'' + esc(v.name) + '\')">在线预览</button>') : '') +
        '<button class="btn btn-sm btn-danger" onclick="deleteVideoFile(\'' + esc(v.name) + '\')">删除</button>' +
      '</div>' +
    '</div>';
  }).join('');
}

// Records & Virtual Library
function updateRecordChannelOptions(channels, devId){
  const sel = document.getElementById('recFilterChannel');
  if(!sel) return;
  const currentVal = sel.value;
  let opts = [];
  if(channels && channels.length > 0){
    opts = channels.map(function(c){
      return '<option value="' + esc(c.id) + '">' + esc(c.name || c.id) + ' (' + esc(c.id) + ')</option>';
    });
  } else if(devId){
    opts = ['<option value="' + esc(devId) + '">' + esc(devId) + ' (主设备)</option>'];
  }
  sel.innerHTML = opts.join('');
  if(currentVal && opts.some(function(o){ return o.includes('value="' + currentVal + '"'); })){
    sel.value = currentVal;
  }
}

function initRecordFilterTimes(){
  const startEl = document.getElementById('recFilterStart');
  const endEl = document.getElementById('recFilterEnd');
  if(startEl && !startEl.value){
    const d = new Date();
    const y = d.getFullYear();
    const m = String(d.getMonth()+1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    startEl.value = y + '-' + m + '-' + day + ' 00:00:00';
  }
  if(endEl && !endEl.value){
    const d = new Date();
    const y = d.getFullYear();
    const m = String(d.getMonth()+1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    endEl.value = y + '-' + m + '-' + day + ' 23:59:59';
  }
}

async function loadDeviceRecords(){
  if(!activeDeviceId) return;
  initRecordFilterTimes();
  const ch = document.getElementById('recFilterChannel').value;
  const st = document.getElementById('recFilterStart').value.trim();
  const et = document.getElementById('recFilterEnd').value.trim();
  const typ = document.getElementById('recFilterType').value;

  const q = new URLSearchParams();
  if(ch) q.set('channel', ch);
  if(st) q.set('start', st);
  if(et) q.set('end', et);
  if(typ) q.set('type', typ);

  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/records?' + q.toString());
  if(!j) return;

  renderDeviceRecords(j);
}

function renderDeviceRecords(data){
  const cfg = data.config || {};
  const recs = data.records || [];

  const modeLabels = {
    'continuous': '全天连续录像 (24小时)',
    'work_hours': '工作时段 (08:00~18:00)',
    'alarm': '报警录像'
  };

  document.getElementById('recSummaryMode').textContent = (cfg.enabled === false) ? '已停用' : (modeLabels[cfg.mode] || cfg.mode || '全天连续');
  document.getElementById('recSummarySlice').textContent = (cfg.slice_minutes || 60) + ' 分钟/段';
  document.getElementById('recSummaryRetain').textContent = '最近 ' + (cfg.retain_days || 7) + ' 天';
  document.getElementById('recSummaryCount').textContent = recs.length + ' 段录像';

  const container = document.getElementById('recordListContainer');
  if(!recs.length){
    container.innerHTML = '<div style="color:var(--text-dim);padding:24px;text-align:center">当前时间范围与过滤条件下无匹配的录像切片（可能超出了保留天数、未在录像时间段内、或功能已停用）。</div>';
    return;
  }

  let rows = recs.map(function(r, idx){
    let sizeMb = (r.fileSize ? (r.fileSize / (1024*1024)).toFixed(1) : '100.0') + ' MB';
    return '<tr>' +
      '<td style="color:var(--text-dim)">' + (idx + 1) + '</td>' +
      '<td style="color:var(--text-muted)">' + esc(r.deviceID) + '</td>' +
      '<td style="color:#6ee7b7">' + esc(r.startTime) + '</td>' +
      '<td style="color:#93c5fd">' + esc(r.endTime) + '</td>' +
      '<td><span class="badge ' + (r.type==='alarm'?'off':'on') + '" style="font-size:10px">' + esc(r.type || 'time') + '</span></td>' +
      '<td>' + esc(sizeMb) + '</td>' +
      '<td style="color:var(--text-dim);max-width:260px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="' + esc(r.filePath) + '">' + esc(r.filePath) + '</td>' +
    '</tr>';
  }).join('');

  container.innerHTML = '<table class="rec-table">' +
    '<thead>' +
      '<tr>' +
        '<th style="width:40px">#</th>' +
        '<th>通道/设备ID</th>' +
        '<th>起始时间 (StartTime)</th>' +
        '<th>结束时间 (EndTime)</th>' +
        '<th>类型 (Type)</th>' +
        '<th>虚拟大小 (FileSize)</th>' +
        '<th>切片路径 (FilePath)</th>' +
      '</tr>' +
    '</thead>' +
    '<tbody>' + rows + '</tbody>' +
  '</table>';
}

// Open Device Full Config Modal (自定义所有参数)
async function openDeviceConfigModal(id, defaultTab){
  const j = await api('/api/devices/' + encodeURIComponent(id));
  if(!j || !j.profile) return;
  const p = j.profile;
  const sip = p.sip || {};
  const dev = p.device || {};
  const media = p.media || {};
  const rec = p.record || {};

  document.getElementById('cfgActiveDevId').value = id;
  document.getElementById('devCfgModalTitle').textContent = '自定义设备参数 · ' + (dev.name || id);

  // SIP
  document.getElementById('cfgSipServerIp').value = sip.server_ip || '127.0.0.1';
  document.getElementById('cfgSipServerPort').value = sip.server_port || 5060;
  document.getElementById('cfgSipLocalIp').value = sip.local_ip || '127.0.0.1';
  document.getElementById('cfgSipLocalPort').value = sip.local_port || 5070;
  document.getElementById('cfgSipTransport').value = sip.transport || 'udp';
  document.getElementById('cfgSipUsername').value = sip.username || id;
  document.getElementById('cfgSipPassword').value = sip.password || '12345678';
  document.getElementById('cfgSipExpires').value = sip.expires || 3600;
  document.getElementById('cfgSipKeepalive').value = sip.keepalive_interval || 60;
  document.getElementById('cfgSipTimeoutCount').value = sip.keepalive_timeout_count || 3;

  // Device
  document.getElementById('cfgDevId').value = dev.id || id;
  document.getElementById('cfgDevName').value = dev.name || '';
  document.getElementById('cfgDevDomain').value = dev.domain || '';
  document.getElementById('cfgDevManufacturer').value = dev.manufacturer || '';
  document.getElementById('cfgDevModel').value = dev.model || '';
  document.getElementById('cfgDevFirmware').value = dev.firmware || '';

  // Media
  document.getElementById('cfgMediaMode').value = media.mode || 'per_channel';
  document.getElementById('cfgMediaSource').value = media.source || 'mp4';
  document.getElementById('cfgMediaMp4').value = media.mp4_file || '';
  document.getElementById('cfgMediaH264').value = media.h264_file || '';
  document.getElementById('cfgMediaWidth').value = media.width || 1280;
  document.getElementById('cfgMediaHeight').value = media.height || 720;
  document.getElementById('cfgMediaFps').value = media.fps || 25;
  document.getElementById('cfgMediaPayloadMax').value = media.rtp_payload_max || 1400;
  document.getElementById('cfgMediaLocalIp').value = media.local_ip || '';

  // Record
  document.getElementById('cfgRecordEnabled').value = (rec.enabled !== false) ? 'true' : 'false';
  document.getElementById('cfgRecordMode').value = rec.mode || 'continuous';
  document.getElementById('cfgRecordSliceMinutes').value = rec.slice_minutes || 60;
  document.getElementById('cfgRecordRetainDays').value = rec.retain_days || 7;
  document.getElementById('cfgRecordType').value = rec.record_type || 'time';
  document.getElementById('cfgRecordFileSizeMB').value = rec.file_size_mb || 100;

  openModal('deviceConfigModal');
  const targetTab = defaultTab || 'sip';
  const tabIdx = {'sip': 1, 'device': 2, 'media': 3, 'record': 4}[targetTab] || 1;
  const tabBtn = document.querySelector('#deviceConfigModal .tab-btn:nth-child(' + tabIdx + ')');
  switchCfgTab(targetTab, tabBtn);
}

// Save Device Full Config
async function saveDeviceConfig(restart){
  const id = document.getElementById('cfgActiveDevId').value;
  if(!activeDeviceData || !activeDeviceData.profile) return;
  const p = JSON.parse(JSON.stringify(activeDeviceData.profile));

  // Update SIP
  p.sip.server_ip = document.getElementById('cfgSipServerIp').value.trim();
  p.sip.server_port = parseInt(document.getElementById('cfgSipServerPort').value) || 5060;
  p.sip.local_ip = document.getElementById('cfgSipLocalIp').value.trim();
  p.sip.local_port = parseInt(document.getElementById('cfgSipLocalPort').value) || 5070;
  p.sip.transport = document.getElementById('cfgSipTransport').value;
  p.sip.username = document.getElementById('cfgSipUsername').value.trim();
  p.sip.password = document.getElementById('cfgSipPassword').value.trim();
  p.sip.expires = parseInt(document.getElementById('cfgSipExpires').value) || 3600;
  p.sip.keepalive_interval = parseInt(document.getElementById('cfgSipKeepalive').value) || 60;
  p.sip.keepalive_timeout_count = parseInt(document.getElementById('cfgSipTimeoutCount').value) || 3;

  // Update Device
  p.device.name = document.getElementById('cfgDevName').value.trim();
  p.device.domain = document.getElementById('cfgDevDomain').value.trim();
  p.device.manufacturer = document.getElementById('cfgDevManufacturer').value.trim();
  p.device.model = document.getElementById('cfgDevModel').value.trim();
  p.device.firmware = document.getElementById('cfgDevFirmware').value.trim();

  // Update Media
  p.media.mode = document.getElementById('cfgMediaMode').value;
  p.media.source = document.getElementById('cfgMediaSource').value;
  p.media.mp4_file = document.getElementById('cfgMediaMp4').value.trim();
  p.media.h264_file = document.getElementById('cfgMediaH264').value.trim();
  p.media.width = parseInt(document.getElementById('cfgMediaWidth').value) || 1280;
  p.media.height = parseInt(document.getElementById('cfgMediaHeight').value) || 720;
  p.media.fps = parseInt(document.getElementById('cfgMediaFps').value) || 25;
  p.media.rtp_payload_max = parseInt(document.getElementById('cfgMediaPayloadMax').value) || 1400;
  p.media.local_ip = document.getElementById('cfgMediaLocalIp').value.trim();

  // Update Record
  if(!p.record) p.record = {};
  p.record.enabled = (document.getElementById('cfgRecordEnabled').value === 'true');
  p.record.mode = document.getElementById('cfgRecordMode').value;
  p.record.slice_minutes = parseInt(document.getElementById('cfgRecordSliceMinutes').value) || 60;
  p.record.retain_days = parseInt(document.getElementById('cfgRecordRetainDays').value) || 7;
  p.record.record_type = document.getElementById('cfgRecordType').value;
  p.record.file_size_mb = parseInt(document.getElementById('cfgRecordFileSizeMB').value) || 100;

  const url = '/api/devices/' + encodeURIComponent(id) + (restart ? '?restart=true' : '');
  const j = await api(url, 'PUT', p);
  if(j && j.ok){
    showToast('设备配置已保存至 JSON 文件' + (restart ? '，并已重启生效' : ''), 'success');
    closeModal('deviceConfigModal');
    refreshAll();
  }
}

// Open New Device Modal
async function openNewDeviceModal(){
  const res = await api('/api/devices/next-port');
  const port = (res && res.port) ? res.port : 5071;
  document.getElementById('newDevLocalPort').value = port;

  // Default server settings from first device if available
  if(devices.length > 0 && activeDeviceData && activeDeviceData.profile){
    const s = activeDeviceData.profile.sip;
    document.getElementById('newDevServerIp').value = s.server_ip || '127.0.0.1';
    document.getElementById('newDevServerPort').value = s.server_port || 5060;
    document.getElementById('newDevLocalIp').value = s.local_ip || '127.0.0.1';
    document.getElementById('newDevTransport').value = s.transport || 'udp';
  } else {
    document.getElementById('newDevServerIp').value = '127.0.0.1';
    document.getElementById('newDevServerPort').value = 5060;
    document.getElementById('newDevLocalIp').value = '127.0.0.1';
  }

  // Suggest ID based on existing max
  let nextNum = devices.length + 1;
  let idPrefix = '340200000011800000';
  document.getElementById('newDevId').value = idPrefix + (nextNum < 10 ? ('0' + nextNum) : nextNum);
  document.getElementById('newDevName').value = '模拟NVR-' + (nextNum < 10 ? ('0' + nextNum) : nextNum);

  openModal('newDeviceModal');
}

// Submit Create Device
async function submitCreateDevice(){
  const id = document.getElementById('newDevId').value.trim();
  const name = document.getElementById('newDevName').value.trim();
  const sIp = document.getElementById('newDevServerIp').value.trim();
  const sPort = parseInt(document.getElementById('newDevServerPort').value) || 5060;
  const lIp = document.getElementById('newDevLocalIp').value.trim() || '127.0.0.1';
  const lPort = parseInt(document.getElementById('newDevLocalPort').value) || 5070;
  const transport = document.getElementById('newDevTransport').value;
  const pwd = document.getElementById('newDevPassword').value.trim() || '12345678';
  const chCount = parseInt(document.getElementById('newDevInitChannels').value) || 0;

  if(!id || id.length !== 20){
    showToast('设备国标编码必须为20位数字', 'error');
    return;
  }

  const channels = [];
  for(let i=1; i<=chCount; i++){
    const chId = id.substring(0, 10) + '132' + id.substring(13, 18) + (i<10?('0'+i):i);
    channels.push({
      id: chId,
      name: '通道' + i,
      status: 'ON',
      parent_id: id,
      register_way: 1
    });
  }

  const profile = {
    enabled: true,
    sip: {
      server_ip: sIp,
      server_port: sPort,
      local_ip: lIp,
      local_port: lPort,
      transport: transport,
      username: id,
      password: pwd,
      expires: 3600,
      keepalive_interval: 60,
      keepalive_timeout_count: 3
    },
    device: {
      id: id,
      name: name || ('模拟设备' + id.substring(16)),
      domain: id.substring(0, 10),
      manufacturer: 'GB28181-Sim',
      model: 'SIM-NVR-100',
      firmware: 'V1.0.0',
      channels: channels
    },
    media: {
      mode: 'per_channel',
      source: 'mp4',
      mp4_file: 'assets/1.mp4',
      fps: 25,
      width: 1280,
      height: 720
    },
    record: {
      enabled: true,
      mode: 'continuous',
      slice_minutes: 60,
      retain_days: 7,
      record_type: 'time',
      file_size_mb: 100
    }
  };

  const j = await api('/api/devices', 'POST', profile);
  if(j && j.ok){
    showToast('模拟设备已成功创建并保存', 'success');
    closeModal('newDeviceModal');
    activeDeviceId = id;
    refreshAll();
  }
}

// Lifecycle Actions
async function startDevice(id){
  const j = await api('/api/devices/' + encodeURIComponent(id) + '/start', 'POST');
  if(j && j.ok){ showToast('设备已启动', 'success'); refreshAll(); }
}
async function stopDevice(id){
  const j = await api('/api/devices/' + encodeURIComponent(id) + '/stop', 'POST');
  if(j && j.ok){ showToast('设备已停止', 'info'); refreshAll(); }
}
async function restartDevice(id){
  const j = await api('/api/devices/' + encodeURIComponent(id) + '/restart', 'POST');
  if(j && j.ok){ showToast('设备已重启', 'success'); refreshAll(); }
}
async function startAllDevices(){
  const j = await api('/api/devices/start-all', 'POST');
  if(j){ showToast('已发起全部启动', 'success'); refreshAll(); }
}
async function stopAllDevices(){
  const j = await api('/api/devices/stop-all', 'POST');
  if(j){ showToast('已停止所有设备', 'info'); refreshAll(); }
}
async function deleteDevice(id){
  if(!confirm('确定彻底删除该模拟设备及对应的配置文件 ' + id + '.json 吗？')) return;
  const j = await api('/api/devices/' + encodeURIComponent(id), 'DELETE');
  if(j && j.ok){
    showToast('设备已删除', 'info');
    if(activeDeviceId === id) activeDeviceId = '';
    refreshAll();
  }
}

// Channel operations for active device
function openAddChannelModal(){
  if(!activeDeviceId){ showToast('请先选择设备', 'error'); return; }
  let nextIdx = (activeDeviceData && activeDeviceData.profile && activeDeviceData.profile.device && activeDeviceData.profile.device.channels ? activeDeviceData.profile.device.channels.length : 0) + 1;
  let chPrefix = activeDeviceId.substring(0, 10) + '132' + activeDeviceId.substring(13, 18);
  document.getElementById('addChId').value = chPrefix + (nextIdx < 10 ? ('0' + nextIdx) : nextIdx);
  document.getElementById('addChName').value = '通道' + nextIdx;
  openModal('addChannelModal');
}

async function submitAddChannel(){
  const id = document.getElementById('addChId').value.trim();
  const name = document.getElementById('addChName').value.trim();
  const video = document.getElementById('addChVideo').value;
  if(!id || id.length !== 20){ showToast('通道编码必须为20位', 'error'); return; }

  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/channels/add', 'POST', {
    id: id,
    name: name,
    mp4: (video==='__synthetic__'?'':video)
  });
  if(j && j.ok){
    showToast('通道已添加并保存', 'success');
    closeModal('addChannelModal');
    refreshAll();
  }
}

async function toggleChannelStatus(chId, nextStatus){
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/channels/update', 'POST', {
    id: chId,
    status: nextStatus
  });
  if(j && j.ok){
    showToast('通道状态已设为 ' + nextStatus, 'info');
    refreshAll();
  }
}

async function bindChannelMedia(chId, btn){
  const sel = btn.parentElement.querySelector('.channel-bind-select');
  const val = sel.value;
  let body;
  if(val === '__ptz__') body = {channelId: chId, source: 'ptz'};
  else if(val === '__synthetic__') body = {channelId: chId, source: 'synthetic'};
  else if(!val) body = {channelId: chId, source: 'mp4', mp4: ''};
  else if(val.endsWith('.h264') || val.endsWith('.264')) body = {channelId: chId, source: 'file', h264: val};
  else body = {channelId: chId, source: 'mp4', mp4: val};

  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/channels/bind', 'POST', body);
  if(j && j.ok){
    showToast('媒体绑定已更新并存盘', 'success');
    refreshAll();
  }
}

async function removeChannel(chId){
  if(!confirm('确定删除通道 ' + chId + ' 吗？')) return;
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/channels/remove?id=' + encodeURIComponent(chId), 'POST');
  if(j && j.ok){
    showToast('通道已删除', 'info');
    refreshAll();
  }
}

async function setDeviceMediaMode(mode){
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/media/mode', 'POST', {mode: mode});
  if(j && j.ok){
    showToast('媒体模式已更新为: ' + (mode==='per_channel'?'按通道独立':'全通道共用'), 'success');
    refreshAll();
  }
}

// SIP triggers
async function triggerManualRegister(){
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/register', 'POST');
  if(j && j.ok){ showToast('已发送注册请求 (SIP REGISTER)', 'success'); refreshAll(); }
}
async function triggerManualKeepalive(){
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/keepalive', 'POST');
  if(j && j.ok){ showToast('已发送心跳 (Keepalive)', 'success'); }
}
function stopSession(callId){
  api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/session/stop?callId=' + encodeURIComponent(callId), 'POST').then(function(j){
    if(j && j.ok){ showToast('已终止该路推流', 'info'); refreshAll(); }
  });
}

// Alarm
function openAlarmModal(){
  const sel = document.getElementById('alarmChannelSelect');
  const chs = (activeDeviceData && activeDeviceData.profile && activeDeviceData.profile.device && activeDeviceData.profile.device.channels) || [];
  if(chs.length){
    sel.innerHTML = chs.map(function(c){
      return '<option value="' + esc(c.id) + '">' + esc(c.name) + ' (' + esc(c.id) + ')</option>';
    }).join('');
  } else {
    sel.innerHTML = '<option value="' + esc(activeDeviceId) + '">主设备 (' + esc(activeDeviceId) + ')</option>';
  }
  openModal('alarmModal');
}
async function submitAlarm(){
  const chId = document.getElementById('alarmChannelSelect').value;
  const method = document.getElementById('alarmMethodSelect').value;
  const priority = document.getElementById('alarmPrioritySelect').value;
  const desc = document.getElementById('alarmDescInput').value.trim();

  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/alarm', 'POST', {
    channelId: chId,
    alarmMethod: method,
    priority: priority,
    description: desc
  });
  if(j && j.ok){
    showToast('模拟报警通知已成功发出', 'success');
    closeModal('alarmModal');
  }
}

// Video Library Modal & Player
function openVideosModal(){
  loadVideos();
  openModal('videosModal');
}
function handleFileSelected(){
  const f = document.getElementById('uploadInput').files[0];
  if(f) document.getElementById('uploadFileName').value = f.name + ' (' + fmtSize(f.size) + ')';
}
async function uploadVideoFile(){
  const f = document.getElementById('uploadInput').files[0];
  if(!f){ showToast('请先选择要上传的视频文件', 'error'); return; }
  const fd = new FormData();
  fd.append('file', f);
  const btn = document.getElementById('uploadBtn');
  btn.disabled = true; btn.textContent = '上传中...';
  try{
    const r = await fetch('/api/videos/upload', {method:'POST', body:fd});
    const j = await r.json();
    if(!r.ok) showToast(j.error || '上传失败', 'error');
    else{
      showToast('上传成功: ' + j.name, 'success');
      document.getElementById('uploadInput').value = '';
      document.getElementById('uploadFileName').value = '';
      await loadVideos();
    }
  }catch(e){ showToast('上传网络错误: ' + e.message, 'error'); }
  btn.disabled = false; btn.textContent = '上传视频';
}
async function deleteVideoFile(name){
  if(!confirm('确定删除视频文件 ' + name + ' 吗？关联缓存也会被清理。')) return;
  const j = await api('/api/videos/delete?name=' + encodeURIComponent(name), 'POST');
  if(j && j.ok){
    showToast('视频已删除: ' + name, 'info');
    await loadVideos();
  }
}
function previewVideo(path, name){
  const filename = name || path.split('/').pop();
  const player = document.getElementById('previewPlayer');
  document.getElementById('videoPreviewTitle').textContent = '在线预览 · ' + filename;
  player.src = '/assets/' + encodeURIComponent(filename);
  openModal('videoPreviewModal');
  player.play().catch(function(){});
}
function closeVideoModal(){
  const player = document.getElementById('previewPlayer');
  player.pause();
  player.src = '';
  closeModal('videoPreviewModal');
}

// Logs
async function loadLogs(){
  if(!activeDeviceId) return;
  const q = document.getElementById('logKeyword').value.trim();
  const url = '/api/devices/' + encodeURIComponent(activeDeviceId) + '/logs?n=300' + (q ? ('&q=' + encodeURIComponent(q)) : '');
  const j = await api(url);
  if(j) {
    rawLogs = j.lines || [];
    renderLogs();
  }
}
function renderLogs(){
  const box = document.getElementById('logBox');
  if(!rawLogs.length){
    box.innerHTML = '<div style="color:var(--text-dim);padding:8px">（暂无日志）</div>';
    return;
  }
  box.innerHTML = rawLogs.map(function(l){
    let tag = 'other';
    let tagClass = '';
    const lower = l.toLowerCase();
    if(lower.includes('[sip]')) { tag = 'SIP'; tagClass = 'sip'; }
    else if(lower.includes('[media]')) { tag = 'MEDIA'; tagClass = 'media'; }
    else if(lower.includes('[gb]')) { tag = 'GB'; tagClass = 'gb'; }
    else if(lower.includes('error') || lower.includes('failed')) { tagClass = 'err'; }

    return '<div class="log-line">' +
      (tag!=='other' ? ('<span class="log-tag ' + tagClass + '">' + tag + '</span>') : '') +
      '<span>' + esc(l) + '</span>' +
    '</div>';
  }).join('');

  if(document.getElementById('autoScroll').checked){
    box.scrollTop = box.scrollHeight;
  }
}
function exportLogs(){
  const blob = new Blob([rawLogs.join('\n')], {type:'text/plain;charset=utf-8'});
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'gb28181_' + activeDeviceId + '.log';
  a.click();
  URL.revokeObjectURL(url);
}

// ==================== PTZ & Presets ====================
let activePTZChannel = null;
let ptzPollTimer = null;
let ptzFastPollTimer = null;

// Virtual Viewport Engine State
let ptzCanvasAnimId = null;
let ptzCurPan = 0;
let ptzCurTilt = 0;
let ptzCurZoom = 1.0;
let ptzTargetPan = 0;
let ptzTargetTilt = 0;
let ptzTargetZoom = 1.0;
let ptzIsMoving = false;
let ptzPanDir = 0;
let ptzTiltDir = 0;
let ptzZoomDir = 0;
let ptzPanSpeed = 64;
let ptzTiltSpeed = 64;
let ptzZoomSpeed = 4;
let ptzFirstInit = true;

let ptzIsDragging = false;
let ptzDragStartX = 0;
let ptzDragStartY = 0;
let ptzDragStartPan = 0;
let ptzDragStartTilt = 0;
let ptzDragSyncTimer = null;
let ptzListenersAttached = false;

if(typeof CanvasRenderingContext2D !== 'undefined' && !CanvasRenderingContext2D.prototype.roundRect){
  CanvasRenderingContext2D.prototype.roundRect = function(x, y, w, h, r){
    if(typeof r === 'undefined') r = 4;
    if(typeof r === 'number') r = [r, r, r, r];
    const tl = r[0] || 0, tr = r[1] || tl, br = r[2] || tl, bl = r[3] || tr;
    this.beginPath();
    this.moveTo(x + tl, y);
    this.lineTo(x + w - tr, y);
    this.quadraticCurveTo(x + w, y, x + w, y + tr);
    this.lineTo(x + w, y + h - br);
    this.quadraticCurveTo(x + w, y + h, x + w - br, y + h);
    this.lineTo(x + bl, y + h);
    this.quadraticCurveTo(x, y + h, x, y + h - bl);
    this.lineTo(x, y + tl);
    this.quadraticCurveTo(x, y, x + tl, y);
    this.closePath();
    return this;
  };
}

const PTZ_LANDMARKS = [
  { az: 0,   name: '园区正门主出入口', type: 'gate',       dist: 70,  color: '#38bdf8', tag: '01号岗亭·车牌识别' },
  { az: 45,  name: '综合科研大厦 A座',  type: 'tower',      dist: 160, color: '#818cf8', tag: '18层研发楼·顶层天线' },
  { az: 90,  name: '东环主干道绿化带',  type: 'highway',    dist: 120, color: '#34d399', tag: '路灯光网·车流干线' },
  { az: 135, name: '智能生态停车场 P1', type: 'parking',    dist: 85,  color: '#fbbf24', tag: '快充车位·地磁感应' },
  { az: 180, name: '核心数据机房 B座',  type: 'datacenter', dist: 130, color: '#60a5fa', tag: '算力中心·散热冷塔' },
  { az: 225, name: '西周界电子防区',    type: 'fence',      dist: 100, color: '#f87171', tag: '红外脉冲·警戒警报' },
  { az: 270, name: '智慧物流装卸平台',  type: 'logistics',  dist: 140, color: '#fb923c', tag: '智能堆垛·集装货场' },
  { az: 315, name: '变电站与微波铁塔',  type: 'substation', dist: 180, color: '#c084fc', tag: '110KV变电·高耸塔架' }
];

const PTZ_STARS = [];
for(let i=0; i<60; i++){
  PTZ_STARS.push({
    az: (i * 37) % 360,
    elev: 5 + (i * 13) % 75,
    size: 0.8 + ((i * 7) % 3) * 0.6,
    alpha: 0.35 + ((i * 11) % 5) * 0.12
  });
}

function normDeg(d){
  return ((d % 360) + 360) % 360;
}

function degDiff(a, b){
  let d = a - b;
  while(d < -180) d += 360;
  while(d > 180) d -= 360;
  return d;
}

function lerpAngle(cur, target, factor){
  let diff = degDiff(target, cur);
  if(Math.abs(diff) < 0.04) return target;
  return normDeg(cur + diff * factor);
}

function lerpVal(cur, target, factor){
  if(Math.abs(target - cur) < 0.02) return target;
  return cur + (target - cur) * factor;
}

function formatCompassDir(pan){
  const p = ((pan % 360) + 360) % 360;
  if(p >= 337.5 || p < 22.5) return '正北 (N)';
  if(p >= 22.5 && p < 67.5) return '东北 (NE)';
  if(p >= 67.5 && p < 112.5) return '正东 (E)';
  if(p >= 112.5 && p < 157.5) return '东南 (SE)';
  if(p >= 157.5 && p < 202.5) return '正南 (S)';
  if(p >= 202.5 && p < 247.5) return '西南 (SW)';
  if(p >= 247.5 && p < 292.5) return '正西 (W)';
  return '西北 (NW)';
}

function formatTiltDir(tilt){
  if(Math.abs(tilt) < 1.0) return '水平平视 (0°)';
  if(tilt > 0) return '仰视 +' + tilt.toFixed(1) + '°';
  return '俯视 ' + tilt.toFixed(1) + '°';
}

function drawPTZViewport(){
  const canvas = document.getElementById('ptzViewportCanvas');
  if(!canvas) return;
  const ctx = canvas.getContext('2d');
  const rect = canvas.getBoundingClientRect();
  const w = Math.round(rect.width || 940);
  const h = 270;
  const dpr = window.devicePixelRatio || 1;

  if(canvas.width !== Math.round(w * dpr) || canvas.height !== Math.round(h * dpr)){
    canvas.width = Math.round(w * dpr);
    canvas.height = Math.round(h * dpr);
  }

  ctx.save();
  ctx.scale(dpr, dpr);
  ctx.clearRect(0, 0, w, h);

  const baseFovH = 60.0;
  const fovH = baseFovH / Math.max(1, ptzCurZoom);
  const fovV = fovH * (h / w);

  // Horizon line: tilt > 0 is looking up (horizon goes down), tilt < 0 is looking down (horizon goes up)
  const horizonY = (h * 0.5) + (ptzCurTilt / (fovV * 0.5)) * (h * 0.5);

  // 1. Sky Gradient & Twinkling Stars
  const skyH = Math.max(0, Math.min(h, horizonY));
  if(skyH > 0){
    const skyGrad = ctx.createLinearGradient(0, 0, 0, skyH);
    skyGrad.addColorStop(0, '#020510');
    skyGrad.addColorStop(0.6, '#060f24');
    skyGrad.addColorStop(1, '#111d38');
    ctx.fillStyle = skyGrad;
    ctx.fillRect(0, 0, w, skyH);

    ctx.save();
    for(let i=0; i<PTZ_STARS.length; i++){
      const st = PTZ_STARS[i];
      const sDiff = degDiff(st.az, ptzCurPan);
      if(Math.abs(sDiff) < fovH * 0.55){
        const sx = (w * 0.5) + (sDiff / (fovH * 0.5)) * (w * 0.5);
        const sy = horizonY - (st.elev / (fovV * 0.5)) * (h * 0.5);
        if(sy >= 2 && sy < horizonY - 4){
          ctx.fillStyle = 'rgba(224, 242, 254, ' + st.alpha + ')';
          ctx.beginPath();
          ctx.arc(sx, sy, st.size, 0, Math.PI * 2);
          ctx.fill();
        }
      }
    }
    ctx.restore();
  }

  // 2. Ground Plane & Perspective Grid
  if(horizonY < h){
    const gTop = Math.max(0, horizonY);
    const gH = h - gTop;
    const groundGrad = ctx.createLinearGradient(0, gTop, 0, h);
    groundGrad.addColorStop(0, '#070f1e');
    groundGrad.addColorStop(0.3, '#050b16');
    groundGrad.addColorStop(1, '#020409');
    ctx.fillStyle = groundGrad;
    ctx.fillRect(0, gTop, w, gH);

    // Perspective grid rays radiating from vanishing point
    ctx.save();
    ctx.strokeStyle = 'rgba(56, 189, 248, 0.12)';
    ctx.lineWidth = 1;
    const panFloor = Math.floor(ptzCurPan / 10) * 10;
    for(let a = panFloor - 40; a <= panFloor + 40; a += 10){
      const diff = degDiff(a, ptzCurPan);
      if(Math.abs(diff) < fovH * 0.65){
        const startX = (w * 0.5) + (diff / (fovH * 0.5)) * (w * 0.5);
        const bottomX = (w * 0.5) + (diff / (fovH * 0.5)) * (w * 0.5) * 2.6;
        ctx.beginPath();
        ctx.moveTo(startX, horizonY);
        ctx.lineTo(bottomX, h);
        ctx.stroke();
      }
    }

    // Concentric distance rings
    const distRings = [
      { d: 160, label: '160m', ratio: 0.22 },
      { d: 100, label: '100m', ratio: 0.48 },
      { d: 60,  label: '60m',  ratio: 0.78 }
    ];
    for(let r=0; r<distRings.length; r++){
      const ring = distRings[r];
      const ry = horizonY + (h - horizonY) * ring.ratio;
      if(ry > 0 && ry < h){
        ctx.strokeStyle = 'rgba(56, 189, 248, 0.18)';
        ctx.setLineDash([4, 4]);
        ctx.beginPath();
        ctx.moveTo(0, ry);
        ctx.lineTo(w, ry);
        ctx.stroke();
        ctx.setLineDash([]);
        ctx.fillStyle = 'rgba(56, 189, 248, 0.35)';
        ctx.font = '9px monospace';
        ctx.fillText('DIST ' + ring.label, w - 65, ry - 3);
      }
    }
    ctx.restore();
  }

  // Horizon Glowing Line
  if(horizonY >= 0 && horizonY <= h){
    ctx.strokeStyle = 'rgba(56, 189, 248, 0.35)';
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(0, horizonY);
    ctx.lineTo(w, horizonY);
    ctx.stroke();
  }

  // 3. Draw Landmarks (Far to Near)
  const sortedLandmarks = PTZ_LANDMARKS.slice().sort(function(a, b){
    return b.dist - a.dist;
  });

  let lockedLandmark = null;
  let minCenterDiff = 999;

  for(let i=0; i<sortedLandmarks.length; i++){
    const lm = sortedLandmarks[i];
    const diff = degDiff(lm.az, ptzCurPan);
    if(Math.abs(diff) < fovH * 0.65){
      const sx = (w * 0.5) + (diff / (fovH * 0.5)) * (w * 0.5);
      const scale = Math.min(3.2, Math.max(0.4, (80.0 / lm.dist) * (0.6 + ptzCurZoom * 0.35)));
      const groundRatio = Math.min(0.9, Math.max(0.15, 65.0 / lm.dist));
      const baseY = horizonY + Math.max(10, (h - horizonY) * groundRatio);

      if(Math.abs(diff) < minCenterDiff){
        minCenterDiff = Math.abs(diff);
        if(Math.abs(diff) < 5.0){
          lockedLandmark = lm;
        }
      }

      ctx.save();
      ctx.translate(sx, baseY);
      drawSingleLandmark(ctx, lm, scale, (lockedLandmark === lm));
      ctx.restore();
    }
  }

  // Update OSD target text
  const osdTargetEl = document.getElementById('ptzOsdTarget');
  if(osdTargetEl){
    if(lockedLandmark){
      osdTargetEl.innerHTML = '🎯 锁定: <span style="color:' + lockedLandmark.color + ';font-weight:700">' +
        lockedLandmark.name + '</span> (' + lockedLandmark.az + '° · ' + lockedLandmark.dist + 'm)';
    } else {
      osdTargetEl.innerHTML = '🎯 巡航视口: 方位 ' + ptzCurPan.toFixed(1) + '° · 俯仰 ' +
        (ptzCurTilt >= 0 ? '+' : '') + ptzCurTilt.toFixed(1) + '° · ' + ptzCurZoom.toFixed(1) + 'x';
    }
  }

  // 4. Military/Tactical Top Compass Ribbon
  drawCompassRibbon(ctx, w, ptzCurPan, fovH);

  // 5. Tactical Center Crosshair HUD
  drawTacticalHUD(ctx, w, h, ptzCurPan, ptzCurTilt, ptzCurZoom, lockedLandmark);

  ctx.restore();
}

function drawSingleLandmark(ctx, lm, scale, isLocked){
  const s = scale;
  ctx.save();

  if(lm.type === 'gate'){
    // Gate Pillars
    ctx.fillStyle = '#1e293b';
    ctx.strokeStyle = '#475569';
    ctx.lineWidth = 1.5;
    ctx.fillRect(-50*s, -55*s, 16*s, 55*s);
    ctx.strokeRect(-50*s, -55*s, 16*s, 55*s);
    ctx.fillRect(34*s, -55*s, 16*s, 55*s);
    ctx.strokeRect(34*s, -55*s, 16*s, 55*s);
    // Header Beam
    ctx.fillStyle = '#0f172a';
    ctx.fillRect(-56*s, -68*s, 112*s, 15*s);
    ctx.strokeStyle = lm.color;
    ctx.strokeRect(-56*s, -68*s, 112*s, 15*s);
    // Banner Text
    ctx.fillStyle = '#38bdf8';
    ctx.font = 'bold ' + Math.max(7, Math.round(7*s)) + 'px sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText('GB28181 NORTH GATE', 0, -58*s);
    // Barrier Arm
    ctx.strokeStyle = '#ef4444';
    ctx.lineWidth = Math.max(2, 3*s);
    ctx.beginPath();
    ctx.moveTo(-34*s, -12*s);
    ctx.lineTo(26*s, -12*s);
    ctx.stroke();
    // Guard Booth
    ctx.fillStyle = '#334155';
    ctx.fillRect(52*s, -32*s, 22*s, 32*s);
    ctx.fillStyle = '#fef08a';
    ctx.fillRect(56*s, -26*s, 14*s, 12*s);
    // Pass LED
    ctx.fillStyle = '#22c55e';
    ctx.beginPath();
    ctx.arc(38*s, -18*s, 2.5*s, 0, Math.PI*2);
    ctx.fill();

  } else if(lm.type === 'tower'){
    // Skyscraper
    ctx.fillStyle = '#0f172a';
    ctx.strokeStyle = '#3b82f6';
    ctx.lineWidth = 1.2;
    ctx.fillRect(-35*s, -145*s, 70*s, 145*s);
    ctx.strokeRect(-35*s, -145*s, 70*s, 145*s);
    // Tiered top
    ctx.fillRect(-22*s, -170*s, 44*s, 25*s);
    ctx.strokeRect(-22*s, -170*s, 44*s, 25*s);
    // Office Windows
    const cols = 5;
    const rows = 12;
    const winW = 5*s;
    const winH = 4*s;
    for(let r=0; r<rows; r++){
      for(let c=0; c<cols; c++){
        const wx = -28*s + c * 12*s;
        const wy = -135*s + r * 10*s;
        ctx.fillStyle = ((r*5+c) % 3 === 0) ? 'rgba(56, 189, 248, 0.85)' : 'rgba(30, 58, 138, 0.4)';
        ctx.fillRect(wx, wy, winW, winH);
      }
    }
    // Rooftop Spire & Blinking Warning Beacon
    ctx.strokeStyle = '#94a3b8';
    ctx.lineWidth = Math.max(1, 1.5*s);
    ctx.beginPath();
    ctx.moveTo(0, -170*s);
    ctx.lineTo(0, -200*s);
    ctx.stroke();
    const blink = (Date.now() % 800 < 400);
    ctx.fillStyle = blink ? '#ef4444' : 'rgba(239, 68, 68, 0.3)';
    ctx.beginPath();
    ctx.arc(0, -200*s, 3*s, 0, Math.PI*2);
    ctx.fill();

  } else if(lm.type === 'highway'){
    // Road surface
    ctx.fillStyle = '#1e293b';
    ctx.beginPath();
    ctx.moveTo(-65*s, 0);
    ctx.lineTo(65*s, 0);
    ctx.lineTo(45*s, -35*s);
    ctx.lineTo(-45*s, -35*s);
    ctx.closePath();
    ctx.fill();
    ctx.strokeStyle = '#334155';
    ctx.stroke();
    // Center divider dash
    ctx.strokeStyle = '#facc15';
    ctx.lineWidth = Math.max(1, 1.5*s);
    ctx.setLineDash([6*s, 4*s]);
    ctx.beginPath();
    ctx.moveTo(0, 0);
    ctx.lineTo(0, -35*s);
    ctx.stroke();
    ctx.setLineDash([]);
    // Street Lamps
    for(let lx of [-52*s, 52*s]){
      ctx.strokeStyle = '#64748b';
      ctx.lineWidth = 1.2;
      ctx.beginPath();
      ctx.moveTo(lx, 0);
      ctx.lineTo(lx, -42*s);
      ctx.lineTo(lx > 0 ? lx - 10*s : lx + 10*s, -45*s);
      ctx.stroke();
      ctx.fillStyle = 'rgba(253, 224, 71, 0.85)';
      ctx.beginPath();
      ctx.arc(lx > 0 ? lx - 10*s : lx + 10*s, -45*s, 2.5*s, 0, Math.PI*2);
      ctx.fill();
    }
    // Animated car lights
    const tCar = (Date.now() * 0.04) % (80*s);
    ctx.fillStyle = '#ffffff';
    ctx.fillRect(-35*s + tCar, -10*s, 4*s, 2*s);
    ctx.fillStyle = '#ef4444';
    ctx.fillRect(35*s - tCar, -22*s, 4*s, 2*s);

  } else if(lm.type === 'parking'){
    // Ground
    ctx.fillStyle = '#0f172a';
    ctx.fillRect(-55*s, -25*s, 110*s, 25*s);
    ctx.strokeStyle = '#334155';
    ctx.strokeRect(-55*s, -25*s, 110*s, 25*s);
    // Stalls & Cars
    for(let p=-2; p<=2; p++){
      const px = p * 20*s;
      ctx.strokeStyle = 'rgba(255,255,255,0.4)';
      ctx.strokeRect(px - 8*s, -22*s, 16*s, 20*s);
      if(p === -1 || p === 1 || p === 2){
        ctx.fillStyle = (p === 1 ? '#38bdf8' : '#64748b');
        ctx.fillRect(px - 6*s, -18*s, 12*s, 14*s);
      }
    }
    // EV Charger
    ctx.fillStyle = '#22c55e';
    ctx.fillRect(-50*s, -32*s, 5*s, 12*s);
    ctx.beginPath();
    ctx.arc(-47.5*s, -34*s, 2*s, 0, Math.PI*2);
    ctx.fill();
    // Canopy
    ctx.fillStyle = 'rgba(56, 189, 248, 0.2)';
    ctx.strokeStyle = '#38bdf8';
    ctx.lineWidth = 1;
    ctx.fillRect(-58*s, -38*s, 116*s, 5*s);
    ctx.strokeRect(-58*s, -38*s, 116*s, 5*s);

  } else if(lm.type === 'datacenter'){
    // Data Center Cube
    ctx.fillStyle = '#111827';
    ctx.strokeStyle = '#60a5fa';
    ctx.lineWidth = 1.5;
    ctx.fillRect(-55*s, -65*s, 110*s, 65*s);
    ctx.strokeRect(-55*s, -65*s, 110*s, 65*s);
    // Server Rack Glowing Slits
    for(let k=-1; k<=1; k++){
      ctx.fillStyle = 'rgba(56, 189, 248, 0.8)';
      ctx.fillRect(k * 30*s - 3*s, -55*s, 6*s, 45*s);
    }
    // Rooftop Chillers
    for(let ch=-1; ch<=1; ch+=2){
      ctx.fillStyle = '#1f2937';
      ctx.fillRect(ch * 28*s - 14*s, -78*s, 28*s, 13*s);
      ctx.strokeStyle = '#4b5563';
      ctx.strokeRect(ch * 28*s - 14*s, -78*s, 28*s, 13*s);
      const fanAngle = Date.now() * 0.01;
      ctx.save();
      ctx.translate(ch * 28*s, -71.5*s);
      ctx.rotate(fanAngle);
      ctx.strokeStyle = '#9ca3af';
      ctx.lineWidth = 1.5;
      ctx.beginPath();
      ctx.moveTo(-7*s, 0); ctx.lineTo(7*s, 0);
      ctx.moveTo(0, -5*s); ctx.lineTo(0, 5*s);
      ctx.stroke();
      ctx.restore();
    }
    ctx.fillStyle = '#93c5fd';
    ctx.font = 'bold ' + Math.max(7, Math.round(6.5*s)) + 'px sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText('IDC CLOUD B', 0, -7*s);

  } else if(lm.type === 'fence'){
    // Security Fence Posts
    for(let f=-3; f<=3; f++){
      const fx = f * 22*s;
      ctx.fillStyle = '#475569';
      ctx.fillRect(fx - 2*s, -42*s, 4*s, 42*s);
      ctx.fillStyle = '#ef4444';
      ctx.beginPath();
      ctx.arc(fx, -43*s, 2.5*s, 0, Math.PI*2);
      ctx.fill();
    }
    ctx.strokeStyle = 'rgba(148, 163, 184, 0.35)';
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(-66*s, -35*s); ctx.lineTo(66*s, -35*s);
    ctx.moveTo(-66*s, -20*s); ctx.lineTo(66*s, -20*s);
    ctx.moveTo(-66*s, -6*s);  ctx.lineTo(66*s, -6*s);
    ctx.stroke();
    // Pulsing Laser
    const laserAlpha = 0.4 + 0.5 * Math.sin(Date.now() * 0.008);
    ctx.strokeStyle = 'rgba(239, 68, 68, ' + laserAlpha + ')';
    ctx.lineWidth = Math.max(1.5, 2.2*s);
    ctx.beginPath();
    ctx.moveTo(-66*s, -43*s);
    ctx.lineTo(66*s, -43*s);
    ctx.stroke();
    // Hazard Sign
    ctx.fillStyle = '#eab308';
    ctx.fillRect(-12*s, -28*s, 24*s, 10*s);
    ctx.fillStyle = '#000000';
    ctx.font = 'bold ' + Math.max(6, Math.round(5.5*s)) + 'px sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText('DANGER', 0, -20*s);

  } else if(lm.type === 'logistics'){
    // Logistics Hangar
    ctx.fillStyle = '#1e293b';
    ctx.fillRect(-60*s, -50*s, 120*s, 50*s);
    ctx.strokeStyle = '#f97316';
    ctx.lineWidth = 1.2;
    ctx.strokeRect(-60*s, -50*s, 120*s, 50*s);
    for(let d=-1; d<=1; d++){
      const dx = d * 36*s;
      ctx.fillStyle = '#0f172a';
      ctx.fillRect(dx - 12*s, -28*s, 24*s, 28*s);
      ctx.strokeStyle = '#64748b';
      ctx.strokeRect(dx - 12*s, -28*s, 24*s, 28*s);
      ctx.fillStyle = '#cbd5e1';
      ctx.font = 'bold ' + Math.max(6, Math.round(6*s)) + 'px sans-serif';
      ctx.textAlign = 'center';
      ctx.fillText('0' + (d+2), dx, -12*s);
    }
    // Containers
    ctx.fillStyle = '#0284c7';
    ctx.fillRect(-55*s, -14*s, 26*s, 14*s);
    ctx.fillStyle = '#ea580c';
    ctx.fillRect(-52*s, -26*s, 22*s, 12*s);

  } else if(lm.type === 'substation'){
    // Lattice Tower
    ctx.strokeStyle = '#cbd5e1';
    ctx.lineWidth = Math.max(1, 1.3*s);
    ctx.beginPath();
    ctx.moveTo(-24*s, 0);
    ctx.lineTo(-4*s, -150*s);
    ctx.lineTo(4*s, -150*s);
    ctx.lineTo(24*s, 0);
    ctx.stroke();
    const tiers = [ -30*s, -65*s, -100*s, -130*s ];
    for(let tr=0; tr<tiers.length; tr++){
      const ty = tiers[tr];
      ctx.beginPath();
      ctx.moveTo(-20*s * (1 - tr*0.2), ty);
      ctx.lineTo(20*s * (1 - tr*0.2), ty);
      ctx.stroke();
    }
    // Microwave Dish
    ctx.fillStyle = '#94a3b8';
    ctx.beginPath();
    ctx.ellipse(8*s, -110*s, 7*s, 11*s, 0.2, 0, Math.PI*2);
    ctx.fill();
    ctx.stroke();
    // Beacon
    ctx.beginPath();
    ctx.moveTo(0, -150*s);
    ctx.lineTo(0, -175*s);
    ctx.stroke();
    const blink2 = (Date.now() % 600 < 300);
    ctx.fillStyle = blink2 ? '#ef4444' : 'rgba(239, 68, 68, 0.25)';
    ctx.beginPath();
    ctx.arc(0, -175*s, 3*s, 0, Math.PI*2);
    ctx.fill();
  }

  // Landmark Tag & Distance Label
  const topH = (lm.type === 'tower' ? -205*s : (lm.type === 'substation' ? -180*s : -72*s));

  ctx.save();
  ctx.fillStyle = lm.color;
  ctx.beginPath();
  ctx.arc(0, topH, 2.5, 0, Math.PI*2);
  ctx.fill();

  ctx.strokeStyle = 'rgba(255, 255, 255, 0.25)';
  ctx.setLineDash([2, 2]);
  ctx.beginPath();
  ctx.moveTo(0, topH);
  ctx.lineTo(0, topH - 10);
  ctx.stroke();
  ctx.setLineDash([]);

  const text = lm.name + ' · ' + lm.dist + 'm';
  ctx.font = '10px -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif';
  const tw = ctx.measureText(text).width;
  ctx.fillStyle = isLocked ? 'rgba(15, 23, 42, 0.95)' : 'rgba(15, 23, 42, 0.75)';
  ctx.strokeStyle = isLocked ? lm.color : 'rgba(255,255,255,0.2)';
  ctx.lineWidth = isLocked ? 1.5 : 1;
  ctx.beginPath();
  ctx.roundRect(-tw/2 - 6, topH - 24, tw + 12, 16, 4);
  ctx.fill();
  ctx.stroke();

  ctx.fillStyle = isLocked ? lm.color : '#e2e8f0';
  ctx.textAlign = 'center';
  ctx.fillText(text, 0, topH - 12);

  // If Target is Locked, draw animated Target Bracket Box
  if(isLocked){
    const boxW = Math.max(70, 110*s);
    const boxH = Math.max(50, Math.abs(topH) + 15);
    const pulse = 2 * Math.sin(Date.now() * 0.008);
    const bw = boxW/2 + pulse;
    const clen = 8;
    const bTop = -boxH - pulse;

    ctx.strokeStyle = lm.color;
    ctx.lineWidth = 2;
    // Top-Left
    ctx.beginPath(); ctx.moveTo(-bw, bTop + clen); ctx.lineTo(-bw, bTop); ctx.lineTo(-bw + clen, bTop); ctx.stroke();
    // Top-Right
    ctx.beginPath(); ctx.moveTo(bw - clen, bTop); ctx.lineTo(bw, bTop); ctx.lineTo(bw, bTop + clen); ctx.stroke();
    // Bottom-Left
    ctx.beginPath(); ctx.moveTo(-bw, 5 - clen); ctx.lineTo(-bw, 5); ctx.lineTo(-bw + clen, 5); ctx.stroke();
    // Bottom-Right
    ctx.beginPath(); ctx.moveTo(bw - clen, 5); ctx.lineTo(bw, 5); ctx.lineTo(bw, 5 - clen); ctx.stroke();

    ctx.fillStyle = lm.color;
    ctx.font = 'bold 9px monospace';
    ctx.textAlign = 'center';
    ctx.fillText('TARGET LOCKED', 0, bTop - 4);
  }

  ctx.restore();
  ctx.restore();
}

function drawCompassRibbon(ctx, w, pan, fovH){
  ctx.save();
  const ribbonW = 380;
  const ribbonH = 22;
  const ribbonX = (w - ribbonW) / 2;
  const ribbonY = 12;

  ctx.fillStyle = 'rgba(15, 23, 42, 0.7)';
  ctx.strokeStyle = 'rgba(255, 255, 255, 0.15)';
  ctx.lineWidth = 1;
  ctx.beginPath();
  ctx.roundRect(ribbonX, ribbonY, ribbonW, ribbonH, 11);
  ctx.fill();
  ctx.stroke();

  ctx.beginPath();
  ctx.roundRect(ribbonX + 2, ribbonY + 2, ribbonW - 4, ribbonH - 4, 9);
  ctx.clip();

  const halfSpan = 35;
  const startDeg = Math.floor((pan - halfSpan) / 5) * 5;
  const endDeg = Math.ceil((pan + halfSpan) / 5) * 5;

  for(let deg = startDeg; deg <= endDeg; deg += 5){
    const dDiff = degDiff(deg, pan);
    const x = (w * 0.5) + (dDiff / halfSpan) * (ribbonW * 0.46);
    const norm = normDeg(deg);
    const isMajor = (norm % 45 === 0);
    const isMid = (!isMajor && norm % 15 === 0);

    ctx.lineWidth = isMajor ? 1.5 : 1;
    ctx.strokeStyle = isMajor ? '#38bdf8' : (isMid ? 'rgba(255,255,255,0.45)' : 'rgba(255,255,255,0.2)');

    const tickH = isMajor ? 8 : (isMid ? 5 : 3);
    ctx.beginPath();
    ctx.moveTo(x, ribbonY + ribbonH);
    ctx.lineTo(x, ribbonY + ribbonH - tickH);
    ctx.stroke();

    if(isMajor || isMid){
      let label = norm + '°';
      if(norm === 0) label = 'N';
      else if(norm === 45) label = 'NE';
      else if(norm === 90) label = 'E';
      else if(norm === 135) label = 'SE';
      else if(norm === 180) label = 'S';
      else if(norm === 225) label = 'SW';
      else if(norm === 270) label = 'W';
      else if(norm === 315) label = 'NW';

      ctx.fillStyle = isMajor ? (norm === 0 ? '#f87171' : '#38bdf8') : 'rgba(255,255,255,0.6)';
      ctx.font = isMajor ? 'bold 9px monospace' : '8px monospace';
      ctx.textAlign = 'center';
      ctx.fillText(label, x, ribbonY + 10);
    }
  }

  ctx.restore();

  // Yellow pointer triangle pointing down
  ctx.save();
  ctx.fillStyle = '#facc15';
  ctx.beginPath();
  ctx.moveTo(w * 0.5 - 4, ribbonY - 1);
  ctx.lineTo(w * 0.5 + 4, ribbonY - 1);
  ctx.lineTo(w * 0.5, ribbonY + 5);
  ctx.closePath();
  ctx.fill();
  ctx.restore();
}

function drawTacticalHUD(ctx, w, h, pan, tilt, zoom, locked){
  ctx.save();
  const cx = w * 0.5;
  const cy = h * 0.5;

  const hudColor = locked ? '#22c55e' : 'rgba(56, 189, 248, 0.5)';
  ctx.strokeStyle = hudColor;
  ctx.lineWidth = 1;

  // Center crosshair with gap
  ctx.beginPath();
  ctx.moveTo(cx - 45, cy); ctx.lineTo(cx - 8, cy);
  ctx.moveTo(cx + 8, cy);  ctx.lineTo(cx + 45, cy);
  ctx.moveTo(cx, cy - 30); ctx.lineTo(cx, cy - 8);
  ctx.moveTo(cx, cy + 8);  ctx.lineTo(cx, cy + 30);
  ctx.stroke();

  // Mil ticks
  for(let m of [-30, -18, 18, 30]){
    ctx.beginPath();
    ctx.moveTo(cx + m, cy - 3); ctx.lineTo(cx + m, cy + 3);
    ctx.stroke();
  }
  for(let m of [-20, 20]){
    ctx.beginPath();
    ctx.moveTo(cx - 3, cy + m); ctx.lineTo(cx + 3, cy + m);
    ctx.stroke();
  }

  // Center dot
  ctx.beginPath();
  ctx.arc(cx, cy, 2, 0, Math.PI*2);
  ctx.stroke();

  // Corner brackets
  ctx.strokeStyle = 'rgba(255, 255, 255, 0.25)';
  ctx.lineWidth = 1.5;
  const pad = 14;
  const blen = 16;
  ctx.beginPath(); ctx.moveTo(pad, pad + blen); ctx.lineTo(pad, pad); ctx.lineTo(pad + blen, pad); ctx.stroke();
  ctx.beginPath(); ctx.moveTo(w - pad - blen, pad); ctx.lineTo(w - pad, pad); ctx.lineTo(w - pad, pad + blen); ctx.stroke();
  ctx.beginPath(); ctx.moveTo(pad, h - pad - blen); ctx.lineTo(pad, h - pad); ctx.lineTo(pad + blen, h - pad); ctx.stroke();
  ctx.beginPath(); ctx.moveTo(w - pad - blen, h - pad); ctx.lineTo(w - pad, h - pad); ctx.lineTo(w - pad, h - pad - blen); ctx.stroke();

  // Angle Readout in lower-right corner
  ctx.fillStyle = 'rgba(255, 255, 255, 0.7)';
  ctx.font = '10px monospace';
  ctx.textAlign = 'right';
  ctx.fillText('AZ: ' + pan.toFixed(1) + '°  EL: ' + (tilt>=0?'+':'') + tilt.toFixed(1) + '°  MAG: ' + zoom.toFixed(1) + 'x', w - pad - 6, h - pad - 8);

  ctx.restore();
}

function ptzAnimationLoop(){
  if(!activePTZChannel){
    ptzCanvasAnimId = null;
    return;
  }

  // Extrapolate motion smoothly if device is moving
  if(ptzIsMoving && !ptzIsDragging){
    const dt = 1.0 / 60.0;
    if(ptzPanDir !== 0){
      const degPerSec = (ptzPanSpeed / 255.0) * 60.0;
      ptzTargetPan = normDeg(ptzTargetPan + ptzPanDir * degPerSec * dt);
    }
    if(ptzTiltDir !== 0){
      const degPerSec = (ptzTiltSpeed / 255.0) * 30.0;
      ptzTargetTilt = Math.max(-90, Math.min(90, ptzTargetTilt + ptzTiltDir * degPerSec * dt));
    }
    if(ptzZoomDir !== 0){
      const ratePerSec = (ptzZoomSpeed / 15.0) * 2.0;
      ptzTargetZoom = Math.max(1.0, Math.min(30.0, ptzTargetZoom + ptzZoomDir * ratePerSec * dt));
    }
  }

  // Smooth lerp
  ptzCurPan = lerpAngle(ptzCurPan, ptzTargetPan, 0.15);
  ptzCurTilt = lerpVal(ptzCurTilt, ptzTargetTilt, 0.15);
  ptzCurZoom = lerpVal(ptzCurZoom, ptzTargetZoom, 0.15);

  drawPTZViewport();

  // Update real-time timestamp on HUD
  const now = new Date();
  const timeStr = now.getFullYear() + '-' +
    String(now.getMonth()+1).padStart(2, '0') + '-' +
    String(now.getDate()).padStart(2, '0') + ' ' +
    String(now.getHours()).padStart(2, '0') + ':' +
    String(now.getMinutes()).padStart(2, '0') + ':' +
    String(now.getSeconds()).padStart(2, '0') + '.' +
    String(now.getMilliseconds()).padStart(3, '0');
  const tEl = document.getElementById('ptzOsdTime');
  if(tEl) tEl.textContent = timeStr;

  ptzCanvasAnimId = requestAnimationFrame(ptzAnimationLoop);
}

function startPTZCanvasLoop(){
  if(!ptzCanvasAnimId){
    ptzCanvasAnimId = requestAnimationFrame(ptzAnimationLoop);
  }
}

function stopPTZCanvasLoop(){
  if(ptzCanvasAnimId){
    cancelAnimationFrame(ptzCanvasAnimId);
    ptzCanvasAnimId = null;
  }
}

function initPTZCanvas(){
  const canvas = document.getElementById('ptzViewportCanvas');
  if(!canvas) return;

  ptzAttachCanvasEvents();
  startPTZCanvasLoop();
}

function ptzAttachCanvasEvents(){
  if(ptzListenersAttached) return;
  ptzListenersAttached = true;
  const canvas = document.getElementById('ptzViewportCanvas');
  if(!canvas) return;

  canvas.addEventListener('mousedown', function(e){
    ptzIsDragging = true;
    ptzDragStartX = e.clientX;
    ptzDragStartY = e.clientY;
    ptzDragStartPan = ptzCurPan;
    ptzDragStartTilt = ptzCurTilt;
    canvas.style.cursor = 'grabbing';
  });

  window.addEventListener('mousemove', function(e){
    if(!ptzIsDragging) return;
    const dx = e.clientX - ptzDragStartX;
    const dy = e.clientY - ptzDragStartY;
    const baseFovH = 60.0;
    const fovH = baseFovH / Math.max(1, ptzCurZoom);
    const fovV = fovH * (270.0 / (canvas.clientWidth || 940));

    const deltaPan = -(dx / (canvas.clientWidth || 940)) * fovH;
    const deltaTilt = (dy / 270.0) * fovV;

    let newPan = normDeg(ptzDragStartPan + deltaPan);
    let newTilt = Math.max(-90, Math.min(90, ptzDragStartTilt + deltaTilt));

    ptzCurPan = newPan;
    ptzTargetPan = newPan;
    ptzCurTilt = newTilt;
    ptzTargetTilt = newTilt;

    document.getElementById('ptzValPan').innerText = newPan.toFixed(1) + '°';
    document.getElementById('ptzPanDir').innerText = formatCompassDir(newPan);
    document.getElementById('ptzValTilt').innerText = (newTilt >= 0 ? '+' : '') + newTilt.toFixed(1) + '°';
    document.getElementById('ptzTiltDir').innerText = formatTiltDir(newTilt);

    if(ptzDragSyncTimer) clearTimeout(ptzDragSyncTimer);
    ptzDragSyncTimer = setTimeout(function(){
      syncPTZPoseToBackend(newPan, newTilt, ptzCurZoom);
    }, 120);
  });

  window.addEventListener('mouseup', function(){
    if(ptzIsDragging){
      ptzIsDragging = false;
      const cv = document.getElementById('ptzViewportCanvas');
      if(cv) cv.style.cursor = 'grab';
      syncPTZPoseToBackend(ptzCurPan, ptzCurTilt, ptzCurZoom);
    }
  });

  canvas.addEventListener('wheel', function(e){
    e.preventDefault();
    const factor = e.deltaY < 0 ? 1.15 : 0.869;
    let newZoom = Math.max(1.0, Math.min(30.0, ptzTargetZoom * factor));
    newZoom = Math.round(newZoom * 10) / 10;
    ptzTargetZoom = newZoom;
    document.getElementById('ptzValZoom').innerText = newZoom.toFixed(1) + 'x';

    if(ptzDragSyncTimer) clearTimeout(ptzDragSyncTimer);
    ptzDragSyncTimer = setTimeout(function(){
      syncPTZPoseToBackend(ptzCurPan, ptzCurTilt, newZoom);
    }, 150);
  }, { passive: false });
}

async function syncPTZPoseToBackend(pan, tilt, zoom){
  if(!activeDeviceId || !activePTZChannel) return;
  await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/ptz/control', 'POST', {
    channelId: activePTZChannel,
    action: 'set_pose',
    pan: Math.round(pan * 10) / 10,
    tilt: Math.round(tilt * 10) / 10,
    zoom: Math.round(zoom * 10) / 10
  });
}

function snapshotPTZCanvas(){
  const canvas = document.getElementById('ptzViewportCanvas');
  if(!canvas) return;
  try {
    const dataURL = canvas.toDataURL('image/png');
    const a = document.createElement('a');
    const d = new Date();
    const ts = d.getFullYear() +
      String(d.getMonth()+1).padStart(2,'0') +
      String(d.getDate()).padStart(2,'0') + '_' +
      String(d.getHours()).padStart(2,'0') +
      String(d.getMinutes()).padStart(2,'0') +
      String(d.getSeconds()).padStart(2,'0');
    a.download = 'GB28181_Snapshot_' + (activePTZChannel || 'CAM') + '_' + ts + '.png';
    a.href = dataURL;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    showToast('视口快照已成功抓拍并下载', 'success');
  } catch(e) {
    showToast('抓拍快照失败: ' + e.message, 'error');
  }
}

async function resetPTZView(){
  if(!activeDeviceId || !activePTZChannel) return;
  ptzTargetPan = 0;
  ptzTargetTilt = 0;
  ptzTargetZoom = 1.0;
  await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/ptz/control', 'POST', {
    channelId: activePTZChannel,
    action: 'reset'
  });
  showToast('云台已复位回正 (正北 0° 平视)', 'info');
  await refreshPTZStatus();
}

function getActiveProfile(){
  return (activeDeviceData && activeDeviceData.profile) ? activeDeviceData.profile : null;
}

function refreshPTZStreamBadge(channelId){
  const d = getActiveProfile();
  const chMedia = (d && d.media && d.media.channels && d.media.channels[channelId]) || {};
  const curSrc = chMedia.source || (d && d.media && d.media.source) || 'mp4';
  const el = document.getElementById('ptzStreamSourceBadge');
  if(!el) return;
  if(curSrc === 'ptz'){
    el.className = 'badge live';
    el.style.background = 'rgba(56, 189, 248, 0.2)';
    el.style.borderColor = 'rgba(56, 189, 248, 0.6)';
    el.style.color = '#38bdf8';
    el.innerHTML = '⚡ 国标推流: 🕹️ PTZ虚拟流 (点播即动)';
  } else {
    el.className = 'badge';
    el.style.background = 'rgba(100, 116, 139, 0.25)';
    el.style.borderColor = 'rgba(100, 116, 139, 0.5)';
    el.style.color = '#cbd5e1';
    el.innerHTML = '📁 国标推流: ' + (curSrc==='synthetic'?'彩条流':'MP4录像') + ' (点击切为PTZ流)';
  }
}

async function togglePTZChannelStreamSource(){
  if(!activeDeviceId || !activePTZChannel) return;
  const d = getActiveProfile();
  if(!d) return;
  const chMedia = (d.media && d.media.channels && d.media.channels[activePTZChannel]) || {};
  const curSrc = chMedia.source || (d.media && d.media.source) || 'mp4';
  const nextSrc = (curSrc === 'ptz') ? 'mp4' : 'ptz';

  const body = {
    channelId: activePTZChannel,
    source: nextSrc
  };
  if(nextSrc === 'mp4' && d.media && d.media.mp4_file){
    body.mp4 = d.media.mp4_file;
  }
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/channels/bind', 'POST', body);
  if(j && j.ok){
    if(nextSrc === 'ptz'){
      showToast('已切换为【🕹️ PTZ 3D虚拟全景流】，WVP 点播将随云台实时旋转！', 'success');
    } else {
      showToast('已恢复为【MP4 视频文件推流】', 'info');
    }
    await refreshAll();
    refreshPTZStreamBadge(activePTZChannel);
  }
}

async function openPTZModal(channelId, channelName){
  activePTZChannel = channelId;
  ptzFirstInit = true;
  document.getElementById('ptzModalTitle').textContent = '通道云台姿态与预置位 · ' + (channelName || channelId);
  document.getElementById('ptzModalSubtitle').textContent = '设备: ' + activeDeviceId + ' | 通道ID: ' + channelId;
  const osdTag = document.getElementById('ptzCamOsdTag');
  if(osdTag) osdTag.textContent = (channelName || channelId) + ' · 1080P@25FPS';
  if(!activeDeviceData) await loadActiveDeviceDetail();
  refreshPTZStreamBadge(channelId);
  openModal('ptzModal');
  initPTZCanvas();
  await refreshPTZStatus();
  if(ptzPollTimer) clearInterval(ptzPollTimer);
  ptzPollTimer = setInterval(function(){
    refreshPTZStatus();
  }, 1000);
}

function closePTZModal(){
  if(ptzPollTimer){
    clearInterval(ptzPollTimer);
    ptzPollTimer = null;
  }
  if(ptzFastPollTimer){
    clearInterval(ptzFastPollTimer);
    ptzFastPollTimer = null;
  }
  stopPTZCanvasLoop();
  activePTZChannel = null;
  closeModal('ptzModal');
}

async function refreshPTZStatus(){
  if(!activeDeviceId || !activePTZChannel) return;
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/ptz?channel=' + encodeURIComponent(activePTZChannel));
  if(!j) return;

  const pan = j.pan || 0;
  const tilt = j.tilt || 0;
  const zoom = j.zoom || 1.0;

  ptzTargetPan = pan;
  ptzTargetTilt = tilt;
  ptzTargetZoom = zoom;
  ptzIsMoving = !!j.isMoving;
  ptzPanDir = j.panDir || 0;
  ptzTiltDir = j.tiltDir || 0;
  ptzZoomDir = j.zoomDir || 0;
  ptzPanSpeed = j.panSpeed || 64;
  ptzTiltSpeed = j.tiltSpeed || 64;
  ptzZoomSpeed = j.zoomSpeed || 4;

  if(ptzFirstInit){
    ptzCurPan = pan;
    ptzCurTilt = tilt;
    ptzCurZoom = zoom;
    ptzFirstInit = false;
  }

  document.getElementById('ptzValPan').innerText = pan.toFixed(1) + '°';
  document.getElementById('ptzPanDir').innerText = formatCompassDir(pan);

  document.getElementById('ptzValTilt').innerText = (tilt >= 0 ? '+' : '') + tilt.toFixed(1) + '°';
  document.getElementById('ptzTiltDir').innerText = formatTiltDir(tilt);

  document.getElementById('ptzValZoom').innerText = zoom.toFixed(1) + 'x';

  const badge = document.getElementById('ptzStatusBadge');
  const text = document.getElementById('ptzStatusText');
  if(j.isMoving){
    badge.className = 'badge live';
    text.innerText = j.statusDesc || '平滑转动中...';
    if(!ptzFastPollTimer){
      ptzFastPollTimer = setInterval(function(){
        if(!ptzIsMoving){
          clearInterval(ptzFastPollTimer);
          ptzFastPollTimer = null;
        } else {
          refreshPTZStatus();
        }
      }, 250);
    }
  } else if(j.activePresetId > 0){
    badge.className = 'badge on';
    text.innerText = '停留在预置位 #' + j.activePresetId + ' (' + (j.activePreset || '') + ')';
    if(ptzFastPollTimer){
      clearInterval(ptzFastPollTimer);
      ptzFastPollTimer = null;
    }
  } else {
    badge.className = 'badge stopped';
    text.innerText = '定格静止';
    if(ptzFastPollTimer){
      clearInterval(ptzFastPollTimer);
      ptzFastPollTimer = null;
    }
  }

  renderPTZPresets(j.presets || []);
}

function renderPTZPresets(presets){
  const tbody = document.getElementById('ptzPresetTbody');
  if(!presets.length){
    tbody.innerHTML = '<tr><td colspan="4" style="padding:20px;text-align:center;color:var(--text-dim)">暂无预置位。点击上方按钮可将当前姿态添加为预置位。</td></tr>';
    return;
  }
  tbody.innerHTML = presets.map(function(p){
    return '<tr style="border-bottom:1px solid rgba(255,255,255,0.05)">' +
      '<td style="padding:7px 8px;font-weight:700;color:var(--accent);font-family:var(--font-mono)">#' + p.id + '</td>' +
      '<td style="padding:7px 8px;font-weight:600;color:#fff">' + esc(p.name) + '</td>' +
      '<td style="padding:7px 8px;font-family:var(--font-mono);font-size:11px;color:var(--text-dim)">' +
        p.pan.toFixed(1) + '° / ' + (p.tilt>=0?'+':'') + p.tilt.toFixed(1) + '° / ' + p.zoom.toFixed(1) + 'x' +
      '</td>' +
      '<td style="padding:7px 8px;text-align:right;white-space:nowrap">' +
        '<button class="btn btn-sm btn-primary" style="padding:2px 8px;margin-right:6px" onclick="callPTZPreset(' + p.id + ')">调用</button>' +
        '<button class="btn btn-sm btn-danger" style="padding:2px 8px" onclick="deletePTZPreset(' + p.id + ')">删除</button>' +
      '</td>' +
    '</tr>';
  }).join('');
}

async function startPTZ(action){
  if(!activeDeviceId || !activePTZChannel) return;
  const spd = parseInt(document.getElementById('ptzSpeedSlider').value, 10) || 80;
  const zSpd = Math.max(1, Math.min(15, Math.round(spd / 16)));
  await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/ptz/control', 'POST', {
    channelId: activePTZChannel,
    action: action,
    panSpeed: spd,
    tiltSpeed: spd,
    zoomSpeed: zSpd
  });
  refreshPTZStatus();
}

async function stopPTZ(){
  if(!activeDeviceId || !activePTZChannel) return;
  await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/ptz/control', 'POST', {
    channelId: activePTZChannel,
    action: 'stop'
  });
  refreshPTZStatus();
}

async function callPTZPreset(presetId){
  if(!activeDeviceId || !activePTZChannel) return;
  const res = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/ptz/preset/call', 'POST', {
    channelId: activePTZChannel,
    presetId: presetId
  });
  if(res){
    showToast('正在平滑转动至预置位 #' + presetId, 'info');
    refreshPTZStatus();
  }
}

async function promptSaveCurrentPreset(){
  if(!activeDeviceId || !activePTZChannel) return;
  const idStr = prompt('请输入要保存的预置位编号 (1~255):', '1');
  if(!idStr) return;
  const id = parseInt(idStr.trim(), 10);
  if(!id || id <= 0 || id > 255){
    alert('预置位编号必须为 1~255 之间的正整数');
    return;
  }
  const name = prompt('请输入预置位名称 (例如: 大门入口/全景监控):', '预置位' + id);
  if(name === null) return;

  const res = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/ptz/preset', 'POST', {
    channelId: activePTZChannel,
    presetId: id,
    name: name.trim() || ('预置位' + id),
    useCurrent: true
  });
  if(res){
    showToast('预置位 #' + id + ' 已保存当前球机坐标', 'success');
    refreshPTZStatus();
  }
}

async function deletePTZPreset(presetId){
  if(!activeDeviceId || !activePTZChannel) return;
  if(!confirm('确定删除预置位 #' + presetId + ' 吗？')) return;
  const res = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/ptz/preset?channel=' + encodeURIComponent(activePTZChannel) + '&presetId=' + presetId, 'DELETE');
  if(res){
    showToast('预置位 #' + presetId + ' 已删除', 'info');
    refreshPTZStatus();
  }
}

// Bootstrap
refreshAll();
setInterval(function(){ refreshAll(); }, 3500);
</script>
</body>
</html>
`
