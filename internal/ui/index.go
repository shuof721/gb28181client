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
        <button class="tab-btn" id="tabBtnSessions" onclick="switchWorkbenchTab('sessions', this)">实时点播与对讲</button>
        <button class="tab-btn" id="tabBtnAlarms" onclick="switchWorkbenchTab('alarms', this)">🚨 报警与布防联动</button>
        <button class="tab-btn" id="tabBtnGPS" onclick="switchWorkbenchTab('gps', this)">🛰️ 移动位置与轨迹模拟</button>
        <button class="tab-btn" id="tabBtnSubs" onclick="switchWorkbenchTab('subs', this)">📡 目录订阅与增量通知</button>
        <button class="tab-btn" id="tabBtnConfigCtrl" onclick="switchWorkbenchTab('configctrl', this)">⚙️ 远程配置与控制</button>
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
      <div id="talkSessionContainer" style="margin-bottom:16px"></div>
      <div style="font-size:12px;font-weight:700;color:var(--text-muted);margin-bottom:8px;display:flex;align-items:center;gap:6px">
        <span>📹 实时视频点播会话 (Video Sessions)</span>
      </div>
      <div id="sessionListContainer">
        <!-- Sessions rendered here -->
      </div>
    </div>

    <!-- Tab: Alarms & Guard Control -->
    <div class="panel-body" id="tabAlarms" style="display:none">
      <div style="display:grid;grid-template-columns:1fr 1.6fr;gap:16px;margin-bottom:16px">
        <!-- Guard & Auto-Alarm Control Card -->
        <div style="background:var(--surface-2);border:1px solid var(--border);border-radius:var(--radius-sm);padding:14px">
          <div style="font-weight:700;font-size:13px;color:var(--text-main);margin-bottom:10px;display:flex;align-items:center;justify-content:space-between">
            <span>🛡️ 设备整机布撤防 (Global Guard)</span>
            <span id="alarmGuardOverallBadge" class="badge">初始状态</span>
          </div>
          <div style="font-size:11px;color:var(--text-dim);line-height:1.5;margin-bottom:12px">
            支持 GB/T 28181 附录 A.2.3 平台布撤防指令 (<code>SetGuard</code> / <code>ResetGuard</code> / <code>ResetAlarm</code>)。整机设防将一键同步下属所有通道。
          </div>
          <div style="display:flex;gap:8px;margin-bottom:14px;flex-wrap:wrap">
            <button class="btn btn-sm btn-primary" onclick="setDeviceGuardState('SetGuard')">🛡️ 全局整机布防 (SetGuard)</button>
            <button class="btn btn-sm" onclick="setDeviceGuardState('ResetGuard')">🔓 全局整机撤防 (ResetGuard)</button>
            <button class="btn btn-sm btn-danger" onclick="setDeviceGuardState('ResetAlarm')">🔕 全局复位报警 (ResetAlarm)</button>
          </div>
          
          <div style="border-top:1px dashed var(--border);padding-top:12px">
            <div style="font-weight:700;font-size:12px;color:var(--text-main);margin-bottom:6px;display:flex;align-items:center;justify-content:space-between">
              <span>⏱️ 周期性自动报警生成器</span>
              <span id="autoAlarmBadge" class="badge off">未启用</span>
            </div>
            <div style="font-size:11px;color:var(--text-dim);margin-bottom:8px">自动按设定的时间间隔轮询<b>已布防</b>通道向平台发送国标报警通知。</div>
            <div style="display:flex;align-items:center;gap:8px">
              <label style="font-size:12px;color:var(--text-muted)">间隔(秒):</label>
              <input type="number" id="autoAlarmInterval" value="15" min="3" max="3600" style="width:70px"/>
              <button class="btn btn-sm" id="btnToggleAutoAlarm" onclick="toggleAutoAlarm()">开启自动报警</button>
            </div>
          </div>
        </div>

        <!-- Quick Trigger Toolbox -->
        <div style="background:var(--surface-2);border:1px solid var(--border);border-radius:var(--radius-sm);padding:14px">
          <div style="font-weight:700;font-size:13px;color:var(--text-main);margin-bottom:10px;display:flex;align-items:center;justify-content:space-between">
            <span>⚡ 快捷报警模拟工具箱 (Quick Trigger)</span>
            <button class="btn btn-sm btn-primary" onclick="openAlarmModal()">自定义高级报警...</button>
          </div>
          <div style="font-size:11px;color:var(--text-dim);margin-bottom:10px">
            向国标平台发送符合 <code>GB/T 28181-2016 附录 A.2.2.1 表 A.2</code> 规范的 <code>Notify &gt; CmdType Alarm</code> 报文。字符集采用 GB2312 对齐 WVP。<br/>
            ⚠️ 规范注意：移动侦测属于<b>视频报警 (Method 5 / Type 2)</b>；若误用设备报警 (Method 2 / Type 2) 则在国标中为<b>设备防拆报警</b>。
          </div>
          <div style="display:flex;gap:10px;align-items:center;margin-bottom:10px;flex-wrap:wrap">
            <label style="font-size:12px;color:var(--text-muted);font-weight:600">目标通道:</label>
            <select id="quickAlarmChannelSelect" style="flex:1;min-width:220px;font-size:12px;font-weight:600;color:var(--primary);background:var(--surface-3);border:1px solid var(--border);border-radius:4px;padding:4px 8px" onchange="updateQuickAlarmChannelBadge()"></select>
            <span id="quickAlarmChannelBadge" class="badge on" style="font-size:11px">默认第1通道</span>
            <label style="font-size:11px;color:var(--text-dim);display:inline-flex;align-items:center;gap:4px">
              <input type="checkbox" id="quickAlarmForce"/> 强制测试发送 (忽略撤防门禁)
            </label>
          </div>
          <div style="display:grid;grid-template-columns:repeat(auto-fill, minmax(170px, 1fr));gap:8px">
            <button class="btn btn-sm" style="border-color:#f59e0b;color:#f59e0b" onclick="triggerQuickAlarm('motion')">🏃 移动侦测 (Method 5 / Type 2)</button>
            <button class="btn btn-sm" style="border-color:#ef4444;color:#ef4444" onclick="triggerQuickAlarm('intrusion')">🚨 周界入侵 (Method 5 / Type 6)</button>
            <button class="btn btn-sm" style="border-color:#eab308;color:#eab308" onclick="triggerQuickAlarm('tamper')">🖐️ 视频遮挡 (Method 5 / Type 11)</button>
            <button class="btn btn-sm" style="border-color:#ec4899;color:#ec4899" onclick="triggerQuickAlarm('videoloss')">📹 视频丢失 (Method 5 / Type 1)</button>
            <button class="btn btn-sm" style="border-color:#dc2626;color:#dc2626" onclick="triggerQuickAlarm('sos')">🆘 紧急求助 (Method 2 / 1级豁免)</button>
            <button class="btn btn-sm" style="border-color:#6366f1;color:#6366f1" onclick="triggerQuickAlarm('diskfault')">💾 存储故障 (Method 6 / Type 21)</button>
          </div>
        </div>
      </div>

      <!-- Channel Guard Management Card -->
      <div style="background:var(--surface-2);border:1px solid var(--border);border-radius:var(--radius-sm);padding:14px;margin-bottom:16px">
        <div style="font-weight:700;font-size:13px;color:var(--text-main);margin-bottom:6px;display:flex;align-items:center;justify-content:space-between">
          <span>🎯 各通道独立防区布撤防与状态管理 (Channel Duty & Guard Status)</span>
          <span style="font-size:11px;font-weight:normal;color:var(--text-dim)">对应国标 <code>DeviceStatus &gt; Alarmstatus</code> 各通道汇报</span>
        </div>
        <div id="channelGuardTableContainer" style="overflow-x:auto"></div>
      </div>

      <!-- Real-time Alarm History Table -->
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:8px">
        <div style="font-weight:700;font-size:12px;color:var(--text-main)">📜 最近报警上报历史记录 (实时存储近100条)</div>
        <div style="display:flex;gap:8px">
          <button class="btn btn-sm btn-danger" onclick="clearDeviceAlarms()">🗑️ 清空历史</button>
          <button class="btn btn-sm" onclick="loadDeviceAlarms()">🔄 刷新记录</button>
        </div>
      </div>
      <div id="alarmHistoryContainer" style="overflow-x:auto;max-height:360px;overflow-y:auto;border:1px solid var(--border);border-radius:var(--radius-sm)">
        <!-- Alarm history table rendered here -->
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

    <!-- Tab 5: GPS / MobilePosition -->
    <div class="panel-body" id="tabGPS" style="display:none">
      <!-- 通道切换与全局控制栏 -->
      <div style="background:var(--surface-2);border-radius:var(--radius);padding:12px 16px;border:1px solid var(--border);margin-bottom:16px;display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:12px">
        <div style="display:flex;align-items:center;gap:8px;flex-wrap:wrap">
          <span style="font-weight:700;font-size:13px;color:var(--text-main);margin-right:6px">🎯 当前配置通道:</span>
          <div id="gpsChannelTabs" style="display:flex;gap:6px;flex-wrap:wrap">
            <!-- 动态填充通道切换药丸按钮 -->
          </div>
        </div>
        <div style="display:flex;align-items:center;gap:8px">
          <button class="btn btn-sm" onclick="syncAllChannelsGPS(true)" title="让所有通道跟随当前主车轨迹，模拟同一车辆上的多摄像头">🚗 一键全通道跟随主车</button>
          <button class="btn btn-sm" onclick="syncAllChannelsGPS(false)" title="将当前通道参数完整克隆给所有通道">📋 一键克隆至所有通道</button>
        </div>
      </div>

      <div style="display:grid;grid-template-columns:1fr 1fr;gap:18px;margin-bottom:16px">
        <!-- 实时位置与仪表卡片 -->
        <div style="background:var(--surface-2);border-radius:var(--radius);padding:16px;border:1px solid var(--border);display:flex;flex-direction:column;gap:12px">
          <div style="display:flex;justify-content:space-between;align-items:center">
            <span style="font-weight:700;font-size:13px;color:var(--text-main)" id="gpsGaugeTitle">🛰️ 实时移动位置仪表 (GB/T 28181 附录 A.2.5)</span>
            <span id="gpsActiveBadge" class="badge"><span class="dot"></span>检测中...</span>
          </div>
          <div style="display:grid;grid-template-columns:1fr 1fr;gap:10px">
            <div style="background:var(--bg-dark);padding:10px;border-radius:var(--radius-sm);border:1px solid var(--border)">
              <div style="font-size:11px;color:var(--text-dim)">当前经度 (Longitude)</div>
              <div style="font-size:18px;font-weight:700;color:var(--cyan);font-family:var(--font-mono)" id="gpsValLon">116.397428</div>
            </div>
            <div style="background:var(--bg-dark);padding:10px;border-radius:var(--radius-sm);border:1px solid var(--border)">
              <div style="font-size:11px;color:var(--text-dim)">当前纬度 (Latitude)</div>
              <div style="font-size:18px;font-weight:700;color:var(--ok);font-family:var(--font-mono)" id="gpsValLat">39.909230</div>
            </div>
          </div>
          <div style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:10px;text-align:center">
            <div style="background:var(--bg-dark);padding:8px;border-radius:var(--radius-sm);border:1px solid var(--border)">
              <div style="font-size:10px;color:var(--text-dim)">移动速度 (Speed)</div>
              <div style="font-size:15px;font-weight:700;color:var(--warn);font-family:var(--font-mono)" id="gpsValSpeed">30.0 km/h</div>
            </div>
            <div style="background:var(--bg-dark);padding:8px;border-radius:var(--radius-sm);border:1px solid var(--border)">
              <div style="font-size:10px;color:var(--text-dim)">航向角 (Direction)</div>
              <div style="font-size:15px;font-weight:700;color:var(--purple);font-family:var(--font-mono)" id="gpsValDir">90.0°</div>
            </div>
            <div style="background:var(--bg-dark);padding:8px;border-radius:var(--radius-sm);border:1px solid var(--border)">
              <div style="font-size:10px;color:var(--text-dim)">海拔高度 (Altitude)</div>
              <div style="font-size:15px;font-weight:700;color:#38bdf8;font-family:var(--font-mono)" id="gpsValAlt">50.0 m</div>
            </div>
          </div>
          <div style="display:flex;justify-content:space-between;align-items:center;font-size:11px;color:var(--text-dim)">
            <span>更新时间: <span id="gpsValTime" style="color:var(--text-main);font-family:var(--font-mono)">-</span></span>
            <button class="btn btn-sm btn-primary" onclick="triggerManualGPSReport()">🚀 立即为当前通道上报位置</button>
          </div>
        </div>

        <!-- 轨迹参数配置表单 -->
        <div style="background:var(--surface-2);border-radius:var(--radius);padding:16px;border:1px solid var(--border);display:flex;flex-direction:column;gap:10px">
          <div style="display:flex;justify-content:space-between;align-items:center">
            <span style="font-weight:700;font-size:13px;color:var(--text-main)" id="gpsCfgFormTitle">⚙️ 运动轨迹参数与仿真规则</span>
            <span id="gpsTargetBadge" class="badge" style="font-size:11px">主设备</span>
          </div>
          <div class="form-grid-2">
            <div class="form-group">
              <label class="form-label">位置上报开关</label>
              <select id="gpsCfgEnabled">
                <option value="true">开启该通道位置模拟</option>
                <option value="false">关闭 (停用)</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">上报触发模式</label>
              <select id="gpsCfgMode">
                <option value="both">两者兼顾 (订阅NOTIFY + 主动MESSAGE)</option>
                <option value="subscribe">仅响应平台订阅 (SUBSCRIBE -&gt; NOTIFY)</option>
                <option value="active">主动定时上报 (SIP MESSAGE)</option>
              </select>
            </div>
          </div>
          <div class="form-group">
            <label class="form-label">轨迹生成模式 (Pattern)</label>
            <select id="gpsCfgPattern" onchange="onGPSPatternChange()">
              <option value="follow">🚗 跟随设备主GPS (车载同车模式，共享主车经纬度)</option>
              <option value="circle">🔄 圆周巡逻 (以中心点圆周环形巡航)</option>
              <option value="linear">↔️ 线性往返 (沿航向折返往返移动)</option>
              <option value="sine">〰️ 正弦波动 (曲线蛇形摆动行进)</option>
              <option value="static">📍 静态驻留 (固定经纬度定点)</option>
            </select>
          </div>
          <div id="gpsParamFields">
            <div class="form-grid-3">
              <div class="form-group">
                <label class="form-label">基准经度 (Longitude)</label>
                <input type="number" step="0.000001" id="gpsCfgLon"/>
              </div>
              <div class="form-group">
                <label class="form-label">基准纬度 (Latitude)</label>
                <input type="number" step="0.000001" id="gpsCfgLat"/>
              </div>
              <div class="form-group">
                <label class="form-label">基准海拔 (Altitude/m)</label>
                <input type="number" id="gpsCfgAlt"/>
              </div>
            </div>
            <div class="form-grid-3">
              <div class="form-group">
                <label class="form-label">巡航速度 (Speed km/h)</label>
                <input type="number" id="gpsCfgSpeed"/>
              </div>
              <div class="form-group">
                <label class="form-label">运动半径/范围 (m)</label>
                <input type="number" id="gpsCfgRadius"/>
              </div>
              <div class="form-group">
                <label class="form-label">上报周期 (Interval 秒)</label>
                <input type="number" id="gpsCfgInterval"/>
              </div>
            </div>
          </div>
          <div id="gpsFollowHint" style="display:none;padding:10px;background:rgba(6,182,212,0.1);border:1px solid #06b6d4;border-radius:var(--radius-sm);font-size:12px;color:#67e8f9">
            ℹ️ <b>车载同车跟随模式已开启</b>：该通道已绑定为主车随动摄像头，其实时经纬度、移动速度与航向角将全自动跟随主设备 GPS 同步演算，无需重复设定独立运动轨迹。
          </div>
          <div style="display:flex;justify-content:flex-end;gap:10px;margin-top:4px">
            <button class="btn btn-sm btn-success" onclick="saveGPSConfig()">💾 保存并生效当前通道配置</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Tab 6: Catalog Subscriptions & Incremental Notify -->
    <div class="panel-body" id="tabSubs" style="display:none">
      <div style="margin-bottom:16px">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:10px">
          <div style="font-weight:700;font-size:13px;color:var(--text-main)">📡 当前活动 SIP 订阅会话 (RFC 3265 / RFC 6665)</div>
          <button class="btn btn-sm" onclick="loadSubscriptions()">🔄 刷新订阅列表</button>
        </div>
        <div id="subscriptionTableContainer" style="overflow-x:auto;border:1px solid var(--border);border-radius:var(--radius-sm)">
          <!-- Subscriptions table rendered here -->
        </div>
      </div>

      <!-- 目录增量通知调试面板 -->
      <div style="background:var(--surface-2);border-radius:var(--radius);padding:16px;border:1px solid var(--border)">
        <div style="font-weight:700;font-size:13px;color:var(--text-main);margin-bottom:6px">📢 国标目录增量通知模拟调试工具 (GB/T 28181 附录 A.2.2)</div>
        <div style="font-size:12px;color:var(--text-dim);margin-bottom:12px">
          当设备产生通道状态变化、更名或上下线时，向所有活跃 Catalog 订阅者推送增量 NOTIFY 报文，平台端即可实时感知更新而无需每次耗时轮询拉取完整全量目录。
        </div>
        <div style="display:flex;align-items:center;gap:12px;flex-wrap:wrap">
          <label style="font-size:12px;color:var(--text-muted)">目标通道:</label>
          <select id="catNotifyChannelSelect" style="min-width:200px"></select>
          <label style="font-size:12px;color:var(--text-muted)">事件类型 (&lt;Event&gt;):</label>
          <select id="catNotifyEventSelect" style="width:160px">
            <option value="ON">ON - 通道上线</option>
            <option value="OFF">OFF - 通道离线</option>
            <option value="UPDATE">UPDATE - 属性更新</option>
            <option value="ADD">ADD - 增加通道</option>
            <option value="DEL">DEL - 删除通道</option>
            <option value="VLOST">VLOST - 视频丢失</option>
            <option value="DEFECT">DEFECT - 硬件故障</option>
          </select>
          <button class="btn btn-sm btn-primary" onclick="submitCatalogNotify()">📢 广播发送目录增量通知</button>
        </div>
      </div>
    </div>

    <!-- Tab 7: Remote Device Configuration & Control -->
    <div class="panel-body" id="tabConfigCtrl" style="display:none">
      <!-- Row 1: Top Status & Capabilities Cards (2 columns) -->
      <div style="display:grid;grid-template-columns:1.2fr 1fr;gap:16px;margin-bottom:16px">
        <!-- Card 1: 虚拟时钟与设备控制操作台 -->
        <div style="background:var(--surface-2);border:1px solid var(--border);border-radius:var(--radius-sm);padding:14px">
          <div style="font-weight:700;font-size:13px;color:var(--text-main);margin-bottom:10px;display:flex;align-items:center;justify-content:space-between">
            <span>⏱️ 设备虚拟时钟与运行状态</span>
            <div style="display:flex;gap:6px">
              <span id="cfgCtrlRecordBadge" class="badge">录像状态</span>
              <span id="cfgCtrlRebootBadge" class="badge">运行正常</span>
            </div>
          </div>
          <div style="font-size:11px;color:var(--text-dim);line-height:1.5;margin-bottom:12px">
            支持响应平台校时指令 (<code>DeviceConfig &gt; Time / Date</code>)，自动维护虚拟时钟偏移量；支持响应远程重启 (<code>TeleBoot</code>)、强制关键帧 (<code>IFrameCmd</code>) 及手动录像控制。
          </div>
          
          <div style="background:var(--surface-3);border:1px solid var(--border);border-radius:var(--radius-sm);padding:10px;margin-bottom:12px">
            <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:6px">
              <span style="font-size:12px;color:var(--text-muted)">当前设备虚拟时钟:</span>
              <span id="cfgCtrlDeviceTime" style="font-family:var(--font-mono);font-size:14px;font-weight:700;color:var(--primary)">-</span>
            </div>
            <div style="display:flex;align-items:center;justify-content:space-between;font-size:11px">
              <span style="color:var(--text-dim)">时钟偏移量 (Offset):</span>
              <span id="cfgCtrlTimeOffset" style="font-family:var(--font-mono);color:var(--text-muted)">0s (与宿主机严格同步)</span>
            </div>
          </div>

          <div style="display:flex;gap:8px;flex-wrap:wrap;margin-bottom:14px">
            <button class="btn btn-sm" onclick="triggerControlResetTime()">🔄 还原对齐宿主机时间</button>
            <button class="btn btn-sm" onclick="promptControlCustomTime()">🕒 手动模拟时间校准...</button>
            <button class="btn btn-sm" id="btnToggleRecord" onclick="triggerControlToggleRecord()">⏺️ 切换录像 (Record ON/OFF)</button>
            <button class="btn btn-sm btn-danger" onclick="triggerControlReboot()">⚠️ 模拟远程重启 (TeleBoot)</button>
          </div>

          <div style="border-top:1px dashed var(--border);padding-top:10px">
            <div style="font-weight:700;font-size:12px;color:var(--text-main);margin-bottom:8px">⚡ 强制请求关键帧 (IFrameCmd / IFCDCmd)</div>
            <div style="display:flex;align-items:center;gap:8px;flex-wrap:wrap">
              <label style="font-size:12px;color:var(--text-muted)">目标通道:</label>
              <select id="cfgCtrlIFrameChannel" style="min-width:180px"></select>
              <button class="btn btn-sm btn-primary" onclick="triggerControlIFrame()">⚡ 立即注入 I 帧 (SPS/PPS+IDR)</button>
            </div>
            <div style="font-size:11px;color:var(--text-dim);margin-top:6px">
              当平台或播放器请求强制关键帧时，模拟器会在推流循环中立即无缝注入 SPS/PPS 参数集及 IDR 关键帧，使画面快速起流或恢复解码。
            </div>
          </div>
        </div>

        <!-- Card 2: 云台守望位 (Home Position) -->
        <div style="background:var(--surface-2);border:1px solid var(--border);border-radius:var(--radius-sm);padding:14px">
          <div style="font-weight:700;font-size:13px;color:var(--text-main);margin-bottom:10px;display:flex;align-items:center;justify-content:space-between">
            <span>🎯 云台看守位 (PTZ HomePosition)</span>
            <span id="homePositionBadge" class="badge">未启用</span>
          </div>
          <div style="font-size:11px;color:var(--text-dim);line-height:1.5;margin-bottom:10px">
            符合 <code>GB/T 28181 DeviceControl &gt; HomePosition</code> 规范。启用后，若云台停止转动超过设定空闲时间，将自动平滑转回看守预置位。
          </div>

          <div class="form-group" style="margin-bottom:10px">
            <label class="form-label" style="font-size:12px">选择通道</label>
            <select id="homePosChannelSelect" onchange="onHomePosChannelChange()"></select>
          </div>

          <div class="form-grid-2" style="margin-bottom:10px">
            <div class="form-group">
              <label class="form-label" style="font-size:12px">归位预置位号 (1-255)</label>
              <input type="number" id="homePosPreset" min="1" max="255" value="1"/>
            </div>
            <div class="form-group">
              <label class="form-label" style="font-size:12px">空闲复位时间 (秒, 1-3600)</label>
              <input type="number" id="homePosResetSec" min="1" max="3600" value="30"/>
            </div>
          </div>

          <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:12px">
            <label style="font-size:12px;color:var(--text-muted);display:inline-flex;align-items:center;gap:6px;cursor:pointer">
              <input type="checkbox" id="homePosEnabled"/> 启用该通道守望位联动
            </label>
            <span id="homePosCountdownInfo" style="font-size:11px;color:var(--text-dim)">-</span>
          </div>

          <div style="display:flex;justify-content:flex-end;gap:8px">
            <button class="btn btn-sm btn-success" onclick="saveHomePositionConfig()">💾 保存守望位配置</button>
          </div>

          <!-- 国标参数能力快速提示卡 -->
          <div style="border-top:1px dashed var(--border);margin-top:12px;padding-top:10px">
            <div style="font-weight:700;font-size:12px;color:var(--text-main);margin-bottom:6px">📋 国标配置查询能力 (ConfigDownload)</div>
            <div style="font-size:11px;color:var(--text-dim);line-height:1.6">
              • <b>BasicParam</b>: 心跳周期 <code>60s</code> / 超时 <code>3次</code> / 定位能力 <code>1 (GPS/北斗)</code><br/>
              • <b>VideoParamOpt</b>: 分辨率 <code>1080P/720P/D1/CIF/QCIF</code> / 支持倍速 <code>1/2/4/8/16/0.5/0.25</code><br/>
              • <b>AudioParamOpt</b>: 音频编码 <code>G.711A (PCMA) / G.711U / AAC</code> / 采样率 <code>8000Hz</code>
            </div>
          </div>
        </div>
      </div>

      <!-- Row 2: 审计流水 (Audit Log) -->
      <div style="background:var(--surface-2);border-radius:var(--radius);padding:16px;border:1px solid var(--border)">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:10px">
          <div>
            <div style="font-weight:700;font-size:13px;color:var(--text-main)">📜 远程控制与配置审计流水日志 (Audit Log - 最近 50 条)</div>
            <div style="font-size:11px;color:var(--text-dim);margin-top:2px">
              记录所有来自上级 SIP 平台信令及 Web 控制台的配置修改、校时、强制关键帧、远程重启及云台守望位指令。
            </div>
          </div>
          <button class="btn btn-sm" onclick="loadConfigCtrlData(true)">🔄 刷新流水</button>
        </div>
        <div id="cfgCtrlAuditContainer" style="overflow-x:auto;max-height:360px;overflow-y:auto;border:1px solid var(--border);border-radius:var(--radius-sm)">
          <!-- Audit logs rendered here -->
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
        <div class="form-grid-2">
          <div class="form-group">
            <label class="form-label">平台国标编号 (Server ID)</label>
            <input type="text" id="cfgSipServerId" placeholder="例如 34020000002000000001 (可自动学习)"/>
          </div>
          <div class="form-group">
            <label class="form-label">XML 字符编码 (Charset)</label>
            <select id="cfgSipCharset">
              <option value="GB2312">GB2312 (推荐对接 WVP / 海康 / 大华等)</option>
              <option value="UTF-8">UTF-8 (部分现代云平台)</option>
            </select>
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
      <h3>发送国标模拟报警通知 (GB/T 28181 附录 D)</h3>
      <button class="btn btn-sm" onclick="closeModal('alarmModal')">✕</button>
    </div>
    <div class="modal-body">
      <div class="form-group">
        <label class="form-label">报警通道或设备编码 (DeviceID)</label>
        <select id="alarmChannelSelect"></select>
      </div>
      <div class="form-grid-2">
        <div class="form-group">
          <label class="form-label">报警方式 (AlarmMethod)</label>
          <select id="alarmMethodSelect">
            <option value="5">5 - 视频报警 (移动侦测/周界/遮挡/丢失等)</option>
            <option value="2">2 - 设备报警 (探头/门磁/红外/人工求助)</option>
            <option value="6">6 - 设备故障报警 (硬盘故障/满/掉电/断网)</option>
            <option value="1">1 - 电话报警</option>
            <option value="3">3 - 短信报警</option>
            <option value="4">4 - GPS报警</option>
            <option value="7">7 - 其他报警</option>
          </select>
        </div>
        <div class="form-group">
          <label class="form-label">报警事件子类型 (AlarmType)</label>
          <select id="alarmTypeSelect">
            <option value="2">2 - 移动侦测报警 (视频)</option>
            <option value="6">6 - 周界防区入侵报警 (视频)</option>
            <option value="5">5 - 警戒绊线越界报警 (视频) / 人工求助 (设备)</option>
            <option value="11">11 - 视频遮挡或镜头篡改 (视频)</option>
            <option value="1">1 - 视频信号丢失 (视频) / 门磁开关 (设备)</option>
            <option value="21">21 - 存储设备/硬盘满或故障 (故障)</option>
            <option value="22">22 - 网络通信链路断开 (故障)</option>
            <option value="51">51 - 区域违章停车/违规停留 (视频)</option>
            <option value="0">0 - 未指定子类型</option>
          </select>
        </div>
      </div>
      <div class="form-grid-2">
        <div class="form-group">
          <label class="form-label">报警级别 (Priority)</label>
          <select id="alarmPrioritySelect">
            <option value="3">3 - 三级警情 (普通日常)</option>
            <option value="2">2 - 二级警情 (重要严重)</option>
            <option value="1">1 - 一级警情 (最高紧急/24h豁免)</option>
            <option value="4">4 - 四级警情 (轻微/提示)</option>
          </select>
        </div>
        <div class="form-group">
          <label class="form-label">报警描述文本 (AlarmDescription)</label>
          <input type="text" id="alarmDescInput" value="Web 控制台触发模拟移动侦测"/>
        </div>
      </div>
      <div class="form-grid-2">
        <div class="form-group">
          <label class="form-label">发生经度 (Longitude, 可选)</label>
          <input type="number" step="0.000001" id="alarmLonInput" placeholder="例如 116.397458"/>
        </div>
        <div class="form-group">
          <label class="form-label">发生纬度 (Latitude, 可选)</label>
          <input type="number" step="0.000001" id="alarmLatInput" placeholder="例如 39.909187"/>
        </div>
      </div>
      <div style="margin-top:8px">
        <label style="font-size:12px;color:var(--text-muted);display:inline-flex;align-items:center;gap:6px">
          <input type="checkbox" id="alarmForceCheck"/> 强制测试发送 (即使通道处于撤防状态也向平台发送 Notify 报文)
        </label>
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
  document.getElementById('tabAlarms').style.display = (tab==='alarms' ? 'block' : 'none');
  document.getElementById('tabGPS').style.display = (tab==='gps' ? 'block' : 'none');
  document.getElementById('tabSubs').style.display = (tab==='subs' ? 'block' : 'none');
  document.getElementById('tabConfigCtrl').style.display = (tab==='configctrl' ? 'block' : 'none');
  document.getElementById('tabLogs').style.display = (tab==='logs' ? 'block' : 'none');
  document.getElementById('tabRecords').style.display = (tab==='records' ? 'block' : 'none');
  if(tab==='alarms') loadDeviceAlarms();
  if(tab==='gps') loadGPSStatus(true);
  if(tab==='subs') loadSubscriptions();
  if(tab==='configctrl') loadConfigCtrlData();
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

  let guardBadge = (st.guardStatus === 'SetGuard') ?
    '<span class="badge on"><span class="dot"></span>🛡️ 已布防</span>' :
    (st.guardStatus === 'ResetGuard' ? '<span class="badge off"><span class="dot"></span>🔓 已撤防</span>' : '');

  let gpsBadge = (st.gps && st.gps.enabled) ?
    '<span class="badge" style="background:#0e7490;color:#fff"><span class="dot"></span>🛰️ 坐标: ' + st.gps.longitude.toFixed(4) + ',' + st.gps.latitude.toFixed(4) + '</span>' : '';
  let subBadge = (st.subscribers && st.subscribers > 0) ?
    '<span class="badge" style="background:#6d28d9;color:#fff"><span class="dot"></span>📡 订阅: ' + st.subscribers + '</span>' : '';

  document.getElementById('workbenchBadges').innerHTML = stBadge +
    guardBadge +
    gpsBadge +
    subBadge +
    '<span class="badge"><span class="dot"></span>本地端口: ' + esc(sipCfg.local_port) + '</span>' +
    '<span class="badge"><span class="dot"></span>平台: ' + esc(sipCfg.server_ip) + ':' + esc(sipCfg.server_port) + '</span>';

  // Mode buttons
  const isPerChannel = (prof.media && prof.media.mode === 'per_channel');
  document.getElementById('btnModeShared').className = 'btn btn-sm ' + (!isPerChannel ? 'btn-primary' : '');
  document.getElementById('btnModePerChannel').className = 'btn btn-sm ' + (isPerChannel ? 'btn-primary' : '');

  renderChannels(devCfg.channels || [], prof.media || {}, st.sessions || []);
  renderSessions(st.sessions || []);
  renderTalkSessions(st.talkSessions || []);
  const tabBtn = document.getElementById('tabBtnSessions');
  if(tabBtn){
    if(st.talkSessions && st.talkSessions.length > 0){
      tabBtn.innerHTML = '实时点播与对讲 <span style="background:#e11d48;color:#fff;border-radius:10px;padding:1px 6px;font-size:10px;font-weight:700">🎙️ 对讲中 (' + st.talkSessions.length + ')</span>';
    } else {
      tabBtn.innerHTML = '实时点播与对讲';
    }
  }
  const tabBtnAlarms = document.getElementById('tabBtnAlarms');
  if(tabBtnAlarms){
    if(st.autoAlarm && st.autoAlarm.enabled){
      tabBtnAlarms.innerHTML = '🚨 报警与布防 <span style="background:#ef4444;color:#fff;border-radius:10px;padding:1px 6px;font-size:10px;font-weight:700">自动上报中</span>';
    } else if(st.guardStatus === 'SetGuard'){
      tabBtnAlarms.innerHTML = '🚨 报警与布防 <span style="background:#10b981;color:#fff;border-radius:10px;padding:1px 6px;font-size:10px;font-weight:700">已布防</span>';
    } else {
      tabBtnAlarms.innerHTML = '🚨 报警与布防联动';
    }
  }
  const tabBtnGPS = document.getElementById('tabBtnGPS');
  if(tabBtnGPS){
    if(st.gps && st.gps.enabled){
      tabBtnGPS.innerHTML = '🛰️ 移动位置与轨迹 <span style="background:#06b6d4;color:#fff;border-radius:10px;padding:1px 6px;font-size:10px;font-weight:700">巡航中</span>';
    } else {
      tabBtnGPS.innerHTML = '🛰️ 移动位置与轨迹模拟';
    }
  }
  const tabBtnSubs = document.getElementById('tabBtnSubs');
  if(tabBtnSubs){
    if(st.subscribers && st.subscribers > 0){
      tabBtnSubs.innerHTML = '📡 目录与事件订阅 <span style="background:#8b5cf6;color:#fff;border-radius:10px;padding:1px 6px;font-size:10px;font-weight:700">' + st.subscribers + ' 个活跃</span>';
    } else {
      tabBtnSubs.innerHTML = '📡 目录订阅与增量通知';
    }
  }
  renderAlarmsTab(st);
  updateRecordChannelOptions(devCfg.channels || [], activeDeviceId);
  if(document.getElementById('tabRecords').style.display !== 'none'){
    loadDeviceRecords();
  }
  if(document.getElementById('tabAlarms').style.display !== 'none'){
    loadDeviceAlarms();
  }
  if(document.getElementById('tabGPS').style.display !== 'none'){
    loadGPSStatus(false);
  }
  if(document.getElementById('tabSubs').style.display !== 'none'){
    loadSubscriptions();
  }
  if(document.getElementById('tabConfigCtrl') && document.getElementById('tabConfigCtrl').style.display !== 'none'){
    loadConfigCtrlData(false);
  }
}

// Render Channels for active device
function renderChannels(channels, mediaCfg, sessions){
  const container = document.getElementById('channelListContainer');
  if(!channels.length){
    container.innerHTML = '<div style="color:var(--text-dim);grid-column:1/-1;padding:20px;text-align:center">该设备暂无下挂通道，请点击上方「新增通道」</div>';
    return;
  }

  // 若用户当前正打开/聚焦通道卡片中的下拉框或输入框，跳过本次定时刷新重绘，避免打断操作或重置选项
  const activeEl = document.activeElement;
  if(activeEl && (activeEl.tagName === 'SELECT' || activeEl.tagName === 'INPUT') && container.contains(activeEl)){
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

    let dutyBadge = '';
    if(ch.dutyStatus === 'ALARM' || ch.isAlarming){
      dutyBadge = '<span class="badge" style="background:#dc2626;color:#fff"><span class="dot"></span>🚨 报警中</span>';
    } else if(ch.guardStatus === 'SetGuard' || ch.dutyStatus === 'ONDUTY'){
      dutyBadge = '<span class="badge on"><span class="dot"></span>🛡️ 已布防</span>';
    } else {
      dutyBadge = '<span class="badge off"><span class="dot"></span>🔓 已撤防</span>';
    }

    const audioEnabled = (chMedia.audio_enabled !== undefined) ? chMedia.audio_enabled : ((mediaCfg && mediaCfg.audio_enabled !== undefined) ? mediaCfg.audio_enabled : true);
    const audioSrc = chMedia.audio_source || (mediaCfg && mediaCfg.audio_source) || 'mp4';
    let audioSrcLabel = 'MP4原声';
    if(audioSrc === 'beep') audioSrcLabel = '安防蜂鸣';
    else if(audioSrc === 'sine') audioSrcLabel = '正弦波(1kHz)';
    else if(audioSrc === 'ambient' || audioSrc === 'noise') audioSrcLabel = '环境底噪';
    else if(audioSrc === 'silence') audioSrcLabel = '静音轨';
    else if(audioSrc === 'file') audioSrcLabel = '本地音频';

    let audioBadge = '';
    if(audioEnabled){
      audioBadge = '<span class="badge" style="background:#059669;color:#fff"><span class="dot"></span>🔊 伴音: ' + esc(audioSrcLabel) + '</span>';
    } else {
      audioBadge = '<span class="badge off"><span class="dot"></span>🔇 纯视频</span>';
    }

    return '<div class="ch-card ' + (isLive ? 'live' : '') + '">' +
      '<div class="ch-header">' +
        '<div>' +
          '<div class="ch-name">' + esc(ch.name) + '</div>' +
          '<div class="ch-id">' + esc(ch.id) + '</div>' +
        '</div>' +
        '<div style="display:flex;gap:4px;align-items:center;flex-wrap:wrap">' +
          (isLive ? '<span class="badge live"><span class="dot"></span>推流中</span>' : '') +
          dutyBadge +
          audioBadge +
          (ch.status === 'ON' ? '<span class="badge on"><span class="dot"></span>在线</span>' : '<span class="badge off"><span class="dot"></span>离线</span>') +
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
        '<button class="btn btn-sm btn-primary" onclick="bindChannelMedia(\'' + esc(ch.id) + '\', this)">绑定视频</button>' +
        '<button class="btn btn-sm" onclick="openPTZModal(\'' + esc(ch.id) + '\',\'' + esc(ch.name) + '\')">🕹️ 云台与预置位</button>' +
        (ch.dutyStatus === 'ALARM' || ch.isAlarming ?
          '<button class="btn btn-sm btn-danger" onclick="setChannelGuardState(\'' + esc(ch.id) + '\',\'ResetAlarm\')">🔕 复位报警</button>' : ''
        ) +
        (ch.guardStatus === 'SetGuard' ?
          '<button class="btn btn-sm" onclick="setChannelGuardState(\'' + esc(ch.id) + '\',\'ResetGuard\')">🔓 撤防</button>' :
          '<button class="btn btn-sm btn-primary" onclick="setChannelGuardState(\'' + esc(ch.id) + '\',\'SetGuard\')">🛡️ 布防</button>'
        ) +
        '<button class="btn btn-sm" onclick="toggleChannelStatus(\'' + esc(ch.id) + '\',\'' + (ch.status==='ON'?'OFF':'ON') + '\')">' + (ch.status==='ON'?'设为离线':'设为在线') + '</button>' +
        '<button class="btn btn-sm" onclick="quickCatalogNotify(\'' + esc(ch.id) + '\')">📢 增量通知</button>' +
        '<button class="btn btn-sm btn-danger" onclick="removeChannel(\'' + esc(ch.id) + '\')">删除</button>' +
      '</div>' +
      '<div style="display:flex;gap:6px;align-items:center;margin-top:6px;padding:6px 8px;background:rgba(255,255,255,0.03);border:1px dashed var(--border);border-radius:var(--radius-sm);flex-wrap:wrap">' +
        '<span style="font-size:11px;color:var(--text-muted);display:flex;align-items:center;gap:4px">🔊 复合伴音:</span>' +
        '<select class="channel-audio-select" onchange="applyChannelAudioSource(\'' + esc(ch.id) + '\', this)" style="padding:2px 6px;border-radius:var(--radius-sm);background:var(--surface-3);color:var(--text-main);border:1px solid var(--border);font-size:11px">' +
          '<option value="mp4" ' + (audioSrc==='mp4'||!audioSrc?'selected':'') + '>🎬 跟随 MP4 原始声音 (如有)</option>' +
          '<option value="beep" ' + (audioSrc==='beep'?'selected':'') + '>🔊 安防蜂鸣提示音 (1000/1400Hz)</option>' +
          '<option value="sine" ' + (audioSrc==='sine'?'selected':'') + '>🎵 1000Hz 纯正弦波测试音</option>' +
          '<option value="ambient" ' + (audioSrc==='ambient'?'selected':'') + '>🍃 摄像头真实环境底噪 (逼真自然)</option>' +
          '<option value="silence" ' + (audioSrc==='silence'?'selected':'') + '>🔇 静音帧 (G.711A 0xD5 满足检测)</option>' +
        '</select>' +
        '<button class="btn btn-sm ' + (audioEnabled?'':'btn-primary') + '" onclick="toggleChannelAudio(\'' + esc(ch.id) + '\', ' + (!audioEnabled) + ', this)">' + (audioEnabled?'🔇 停用伴音':'🔊 开启伴音') + '</button>' +
        '<button class="btn btn-sm" onclick="applyChannelAudioSource(\'' + esc(ch.id) + '\', this)">应用音源</button>' +
      '</div>' +
    '</div>';
  }).join('');
}

let sessFastPollTimer = null;

function updateLiveSessionsVU(sessions){
  if(!sessions) return;
  sessions.forEach(function(s){
    const safeId = (s.callId || '').replace(/[^a-zA-Z0-9_-]/g, '_');
    if(s.audioEnabled){
      applyVULevel('sessVUBar_' + safeId, 'sessVUTxt_' + safeId, s.audioLevel || 0);
    }
  });
}

// Render Sessions for active device
function renderSessions(sessions){
  const container = document.getElementById('sessionListContainer');
  if(!sessions || !sessions.length){
    if(sessFastPollTimer){
      clearInterval(sessFastPollTimer);
      sessFastPollTimer = null;
    }
    container.innerHTML = '<div style="color:var(--text-dim);padding:24px;text-align:center">当前无实时推流会话。在平台（如 WVP）上点击通道播放后，将在此展示推流状态。</div>';
    return;
  }

  // 启动 300ms 高频轻量轮询，使推流伴音电平表如对讲般敏锐跳动
  if(!sessFastPollTimer){
    sessFastPollTimer = setInterval(function(){
      if(activeDeviceId){
        api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/sessions').then(function(res){
          if(res && res.sessions && res.sessions.length > 0){
            updateLiveSessionsVU(res.sessions);
          } else if(res && res.sessions && res.sessions.length === 0){
            renderSessions([]);
          }
        });
      }
    }, 300);
  }

  // 检查是否已有相同会话卡片结构，若已存在则直接增量平滑刷新 VU，避免整个 DOM 被重建闪烁
  const existingItems = container.querySelectorAll('.session-item[data-call-id]');
  if(existingItems.length === sessions.length){
    let allMatch = true;
    for(let i=0; i<sessions.length; i++){
      if(existingItems[i].getAttribute('data-call-id') !== sessions[i].callId){
        allMatch = false;
        break;
      }
    }
    if(allMatch){
      updateLiveSessionsVU(sessions);
      return;
    }
  }

  container.innerHTML = sessions.map(function(s){
    const safeId = (s.callId || '').replace(/[^a-zA-Z0-9_-]/g, '_');
    let audioHtml = '';
    if(s.audioEnabled){
      const lvl = Math.round(s.audioLevel || 0);
      audioHtml = '<div style="margin-top:6px;padding:4px 8px;background:rgba(5,150,105,0.12);border:1px solid rgba(5,150,105,0.3);border-radius:4px;display:flex;align-items:center;gap:8px">' +
        '<span style="font-size:11px;font-weight:700;color:#34d399">🔊 复合音频 (' + esc(s.audioCodec||'G.711A') + ' 8kHz · ' + esc(s.audioSource||'beep') + '):</span>' +
        '<div style="flex:1;height:8px;background:rgba(0,0,0,0.5);border-radius:4px;overflow:hidden;position:relative">' +
          '<div id="sessVUBar_' + safeId + '" style="width:' + lvl + '%;height:100%;background:linear-gradient(90deg,#10b981,#34d399);border-radius:4px;transition:width 0.12s ease-out"></div>' +
        '</div>' +
        '<span id="sessVUTxt_' + safeId + '" style="font-size:10px;font-family:var(--font-mono);color:#34d399;min-width:32px;text-align:right">' + lvl + '%</span>' +
      '</div>';
    } else {
      audioHtml = '<div style="margin-top:4px;font-size:11px;color:var(--text-dim)">🔇 纯视频推流 (未启用伴音)</div>';
    }

    return '<div class="session-item" data-call-id="' + esc(s.callId) + '">' +
      '<div style="flex:1">' +
        '<div style="font-weight:700;color:#fff;font-size:13px">' + esc(s.channelId) + ' · <span class="badge live"><span class="dot"></span>' + esc(s.streamType||'live') + '</span></div>' +
        '<div style="font-size:11px;font-family:var(--font-mono);color:var(--text-dim);margin-top:4px">' +
          'SSRC: ' + esc(s.ssrc) + ' · 目标: ' + esc(s.remoteIp) + ':' + esc(s.remotePort) + ' (' + (s.isTcp?'TCP':'UDP') + ') · FPS: ' + (s.fps||25) +
        '</div>' +
        audioHtml +
      '</div>' +
      '<button class="btn btn-sm btn-danger" onclick="stopSession(\'' + esc(s.callId) + '\')">停止推流</button>' +
    '</div>';
  }).join('');

  updateLiveSessionsVU(sessions);
}

async function toggleChannelAudio(channelId, enable, btn){
  if(!activeDeviceId) return;
  const card = btn ? btn.closest('.ch-card') : null;
  const sel = card ? card.querySelector('.channel-audio-select') : null;
  const src = sel ? sel.value : 'mp4';
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/channels/audio', 'POST', {
    channelId: channelId,
    audioEnabled: enable,
    audioSource: src
  });
  if(j && j.ok){
    showToast('通道伴音已' + (enable ? '启用 (复合流推流中即时生效)' : '停用 (降级为纯视频)'), 'success');
    if(document.activeElement) document.activeElement.blur();
    refreshAll();
  }
}

async function applyChannelAudioSource(channelId, btnOrSel){
  if(!activeDeviceId) return;
  const card = btnOrSel ? btnOrSel.closest('.ch-card') : null;
  const sel = card ? card.querySelector('.channel-audio-select') : null;
  const src = sel ? sel.value : 'mp4';
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/channels/audio', 'POST', {
    channelId: channelId,
    audioEnabled: true,
    audioSource: src
  });
  if(j && j.ok){
    let label = '🎬 MP4原声';
    if(src === 'beep') label = '🔊 安防蜂鸣';
    else if(src === 'sine') label = '🎵 1000Hz正弦波';
    else if(src === 'ambient') label = '🍃 环境底噪';
    else if(src === 'silence') label = '🔇 静音帧';
    showToast('伴音音源已即时生效为: ' + label, 'success');
    if(document.activeElement) document.activeElement.blur();
    refreshAll();
  }
}

// Voice Intercom / Broadcast State & Functions
let talkWS = null;
let talkWSSessionCallId = '';
let talkAudioCtx = null;
let talkNextPlayTime = 0;
let talkSpeakerEnabled = false;
let talkMicStream = null;
let talkMicProcessor = null;
let isTalkingMic = false;
let talkFastPollTimer = null;
let vuSmoothLevels = { rxVUBar: 0, txVUBar: 0 };

function applyVULevel(barId, txtId, targetLvl){
  const bar = document.getElementById(barId);
  const txt = document.getElementById(txtId);
  if(!bar) return;
  targetLvl = Math.max(0, Math.min(100, targetLvl));

  let current = vuSmoothLevels[barId] || 0;
  if(targetLvl >= current){
    current = targetLvl;
  } else {
    current = current * 0.72 + targetLvl * 0.28;
    if(current < 0.5) current = 0;
  }
  vuSmoothLevels[barId] = current;

  const rounded = Math.round(current);
  bar.style.width = rounded + '%';

  if(rounded < 50){
    bar.style.background = 'linear-gradient(90deg, #10b981, #34d399)';
    bar.style.boxShadow = 'none';
  } else if(rounded < 80){
    bar.style.background = 'linear-gradient(90deg, #10b981 0%, #f59e0b 100%)';
    bar.style.boxShadow = 'none';
  } else {
    bar.style.background = 'linear-gradient(90deg, #f59e0b 0%, #ef4444 100%)';
    bar.style.boxShadow = '0 0 10px rgba(239, 68, 68, 0.65)';
  }

  if(txt){
    txt.textContent = rounded + '%';
    if(rounded >= 80){
      txt.style.color = '#f87171';
    } else if(rounded >= 50){
      txt.style.color = '#fbbf24';
    } else {
      txt.style.color = 'var(--text-main)';
    }
  }
}

function updateRxVUFromPCM(arrayBuffer){
  if(!arrayBuffer || arrayBuffer.byteLength < 2) return;
  const pcm16 = new Int16Array(arrayBuffer);
  let sumSq = 0;
  for(let i = 0; i < pcm16.length; i++){
    sumSq += pcm16[i] * pcm16[i];
  }
  const rms = Math.sqrt(sumSq / pcm16.length);
  if(rms < 12.0){
    applyVULevel('rxVUBar', 'rxVUTxt', 0);
    return;
  }
  const db = 20.0 * Math.log10(rms / 32768.0);
  const minDB = -52.0;
  const maxDB = -2.0;
  let lvl = 0;
  if(db > minDB){
    lvl = ((db - minDB) / (maxDB - minDB)) * 100.0;
    if(lvl > 100) lvl = 100;
  }
  applyVULevel('rxVUBar', 'rxVUTxt', lvl);
}

function renderTalkSessions(talkSessions){
  const container = document.getElementById('talkSessionContainer');
  if(!container) return;
  if(!talkSessions || !talkSessions.length){
    if(talkFastPollTimer){
      clearInterval(talkFastPollTimer);
      talkFastPollTimer = null;
    }
    vuSmoothLevels = { rxVUBar: 0, txVUBar: 0 };
    if(talkWS){
      talkWS.close();
      talkWS = null;
      talkWSSessionCallId = '';
    }
    stopTalkMic();
    container.innerHTML = '<div style="background:rgba(255,255,255,0.02);border:1px dashed var(--border);border-radius:8px;padding:14px 16px;display:flex;align-items:center;justify-content:space-between;gap:12px;flex-wrap:wrap">' +
      '<div style="display:flex;align-items:center;gap:10px">' +
        '<div style="width:32px;height:32px;border-radius:50%;background:rgba(59,130,246,0.1);display:flex;align-items:center;justify-content:center;font-size:16px">🎙️</div>' +
        '<div>' +
          '<div style="font-weight:700;color:var(--text-main);font-size:12px">语音对讲与广播通道已就绪 (GB/T 28181 附录 B/C)</div>' +
          '<div style="font-size:11px;color:var(--text-dim);margin-top:2px">支持 G.711A (PCMA/8000Hz) 双向对讲与下行广播。在平台（如 WVP）上点击通道对讲时将在此处实时联动。</div>' +
        '</div>' +
      '</div>' +
    '</div>';
    return;
  }

  const s = talkSessions[0];
  ensureTalkWS(s.callId);

  if(!talkFastPollTimer){
    talkFastPollTimer = setInterval(function(){
      if(activeDeviceId){
        api('/api/devices/' + encodeURIComponent(activeDeviceId)).then(function(dev){
          const ts = (dev && dev.status && dev.status.talkSessions) || (dev && dev.state && dev.state.talkSessions);
          if(ts){
            renderTalkSessions(ts);
          }
        });
      }
    }, 400);
  }

  const existingCard = document.getElementById('talkSessionCard');
  if(existingCard &&
     existingCard.getAttribute('data-call-id') === s.callId &&
     existingCard.getAttribute('data-uplink-mode') === (s.uplinkMode || 'synthetic') &&
     existingCard.getAttribute('data-speaker') === String(talkSpeakerEnabled)){
    const rxPkts = document.getElementById('talkRxPackets');
    if(rxPkts) rxPkts.textContent = '接收包数: ' + s.rxPackets + ' · 流量: ' + fmtSize(s.rxBytes);
    const txPkts = document.getElementById('talkTxPackets');
    if(txPkts) txPkts.textContent = '推流包数: ' + s.txPackets + ' · 流量: ' + fmtSize(s.txBytes);
    if(!isTalkingMic){
      applyVULevel('txVUBar', 'txVUTxt', s.txVolume || 0);
    }
    if(!talkWS || talkWS.readyState !== WebSocket.OPEN){
      applyVULevel('rxVUBar', 'rxVUTxt', s.rxVolume || 0);
    }
    return;
  }

  const rxPct = Math.min(100, Math.max(0, s.rxVolume || 0));
  const txPct = Math.min(100, Math.max(0, s.txVolume || 0));

  container.innerHTML = '<div id="talkSessionCard" data-call-id="' + esc(s.callId) + '" data-uplink-mode="' + esc(s.uplinkMode || 'synthetic') + '" data-speaker="' + String(talkSpeakerEnabled) + '" style="background:linear-gradient(135deg, rgba(225,29,72,0.08) 0%, rgba(30,41,59,0.7) 100%);border:1px solid rgba(225,29,72,0.3);border-radius:10px;padding:16px;box-shadow:0 4px 20px rgba(0,0,0,0.3)">' +
    '<div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:12px;flex-wrap:wrap;gap:8px">' +
      '<div style="display:flex;align-items:center;gap:8px">' +
        '<span class="badge" style="background:#e11d48;color:#fff;font-size:11px;padding:2px 8px;font-weight:700"><span class="dot" style="background:#fff"></span>🎙️ ' + esc(s.streamType==='broadcast'?'语音广播中':'双向对讲中') + '</span>' +
        '<span style="font-weight:700;color:#fff;font-size:13px">' + esc(s.channelId) + '</span>' +
        '<span style="font-size:11px;color:var(--text-dim);font-family:var(--font-mono)">' + esc(s.audioCodec||'PCMA') + ' / 8000Hz · SSRC: ' + esc(s.ssrc) + '</span>' +
      '</div>' +
      '<div style="display:flex;gap:6px">' +
        '<button class="btn btn-sm ' + (talkSpeakerEnabled ? 'btn-primary' : '') + '" onclick="toggleTalkSpeaker(\'' + esc(s.callId) + '\')">' + (talkSpeakerEnabled ? '🔊 扬声器已开启 (点击静音)' : '🔈 开启扬声器收听') + '</button>' +
        '<button class="btn btn-sm btn-danger" onclick="stopTalkSession(\'' + esc(s.callId) + '\')">挂断对讲</button>' +
      '</div>' +
    '</div>' +

    // VU Meters
    '<div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;background:rgba(0,0,0,0.25);border-radius:8px;padding:12px;margin-bottom:12px">' +
      '<div>' +
        '<div style="display:flex;justify-content:space-between;font-size:11px;color:var(--text-dim);margin-bottom:4px">' +
          '<span>下行平台收音 (平台 → 设备)</span>' +
          '<span id="rxVUTxt" style="font-family:var(--font-mono);font-weight:700">' + Math.round(rxPct) + '%</span>' +
        '</div>' +
        '<div style="height:12px;background:#0b1120;border:1px solid rgba(255,255,255,0.08);border-radius:6px;overflow:hidden;padding:1px">' +
          '<div id="rxVUBar" style="width:' + rxPct + '%;height:100%;background:linear-gradient(90deg, #10b981, #f59e0b);transition:width 0.08s ease-out, background 0.15s ease;border-radius:4px"></div>' +
        '</div>' +
        '<div id="talkRxPackets" style="font-size:10px;font-family:var(--font-mono);color:var(--text-dim);margin-top:4px">' +
          '接收包数: ' + s.rxPackets + ' · 流量: ' + fmtSize(s.rxBytes) +
        '</div>' +
      '</div>' +
      '<div>' +
        '<div style="display:flex;justify-content:space-between;font-size:11px;color:var(--text-dim);margin-bottom:4px">' +
          '<span>上行设备发音 (设备 → 平台)</span>' +
          '<span id="txVUTxt" style="font-family:var(--font-mono);font-weight:700">' + Math.round(txPct) + '%</span>' +
        '</div>' +
        '<div style="height:12px;background:#0b1120;border:1px solid rgba(255,255,255,0.08);border-radius:6px;overflow:hidden;padding:1px">' +
          '<div id="txVUBar" style="width:' + txPct + '%;height:100%;background:linear-gradient(90deg, #3b82f6, #8b5cf6);transition:width 0.08s ease-out, background 0.15s ease;border-radius:4px"></div>' +
        '</div>' +
        '<div id="talkTxPackets" style="font-size:10px;font-family:var(--font-mono);color:var(--text-dim);margin-top:4px">' +
          '推流包数: ' + s.txPackets + ' · 流量: ' + fmtSize(s.txBytes) +
        '</div>' +
      '</div>' +
    '</div>' +

    // Control bar for talking
    '<div style="display:flex;align-items:center;justify-content:space-between;gap:10px;flex-wrap:wrap">' +
      '<div style="display:flex;align-items:center;gap:8px">' +
        '<span style="font-size:12px;color:var(--text-muted)">向平台推流声音源:</span>' +
        '<select class="channel-bind-select" style="width:180px" onchange="setTalkUplinkMode(\'' + esc(s.callId) + '\', this.value)">' +
          '<option value="synthetic" ' + (s.uplinkMode==='synthetic'?'selected':'') + '>🎵 内置蜂鸣提示音</option>' +
          '<option value="mic" ' + (s.uplinkMode==='mic'?'selected':'') + '>🎤 浏览器麦克风喊话</option>' +
        '</select>' +
      '</div>' +
      '<div style="display:flex;align-items:center;gap:8px">' +
        (s.uplinkMode==='mic' ?
          ('<button class="btn btn-sm btn-primary" id="btnTalkMic" style="background:#e11d48;border-color:#e11d48;font-weight:700;padding:6px 14px" ' +
            'onmousedown="startTalkMic(\'' + esc(s.callId) + '\')" onmouseup="stopTalkMic()" onmouseleave="stopTalkMic()" ' +
            'ontouchstart="startTalkMic(\'' + esc(s.callId) + '\')" ontouchend="stopTalkMic()">' +
            '🎙️ 按住对讲 (Push-to-Talk)' +
          '</button>') :
          '<span style="font-size:11px;color:var(--text-dim)">平台当前正在收听内置模拟蜂鸣提示音</span>'
        ) +
      '</div>' +
    '</div>' +
  '</div>';
}

function ensureTalkWS(callId){
  if(talkWS && talkWS.readyState === WebSocket.OPEN && talkWSSessionCallId === callId) return;
  if(talkWS) talkWS.close();
  talkWSSessionCallId = callId;
  const proto = (location.protocol === 'https:') ? 'wss://' : 'ws://';
  const url = proto + location.host + '/api/devices/' + encodeURIComponent(activeDeviceId) + '/talk/ws?callId=' + encodeURIComponent(callId);
  try {
    talkWS = new WebSocket(url);
    talkWS.binaryType = 'arraybuffer';
    talkWS.onmessage = function(ev){
      if(ev.data instanceof ArrayBuffer){
        if(talkSpeakerEnabled){
          playPCMFrames(ev.data);
        }
        updateRxVUFromPCM(ev.data);
      }
    };
    talkWS.onclose = function(){ talkWS = null; };
    talkWS.onerror = function(){ talkWS = null; };
  } catch(e) {
    console.error('talk ws error:', e);
  }
}

function playPCMFrames(arrayBuffer){
  if(!talkSpeakerEnabled) return;
  if(!talkAudioCtx){
    talkAudioCtx = new (window.AudioContext || window.webkitAudioContext)({ sampleRate: 8000 });
  }
  if(talkAudioCtx.state === 'suspended'){
    talkAudioCtx.resume();
  }
  const int16 = new Int16Array(arrayBuffer);
  const float32 = new Float32Array(int16.length);
  for(let i = 0; i < int16.length; i++){
    float32[i] = int16[i] / 32768.0;
  }
  const buffer = talkAudioCtx.createBuffer(1, float32.length, 8000);
  buffer.getChannelData(0).set(float32);

  const src = talkAudioCtx.createBufferSource();
  src.buffer = buffer;
  src.connect(talkAudioCtx.destination);

  const now = talkAudioCtx.currentTime;
  if(talkNextPlayTime < now){
    talkNextPlayTime = now;
  }
  src.start(talkNextPlayTime);
  talkNextPlayTime += buffer.duration;
}

function toggleTalkSpeaker(callId){
  talkSpeakerEnabled = !talkSpeakerEnabled;
  if(talkSpeakerEnabled){
    if(talkAudioCtx && talkAudioCtx.state === 'suspended') talkAudioCtx.resume();
    ensureTalkWS(callId);
    showToast('扬声器已开启，正在收听平台声音', 'success');
  } else {
    talkNextPlayTime = 0;
    showToast('扬声器已静音', 'info');
  }
  refreshAll();
}

let talkMicFilter = null;
let talkMicBuffer = [];
let talkMicPhase = 0;

async function startTalkMic(callId){
  isTalkingMic = true;
  const btn = document.getElementById('btnTalkMic');
  if(btn){
    btn.style.background = '#9f1239';
    btn.innerText = '🔴 正在说话... (松开停止)';
  }
  try {
    ensureTalkWS(callId);
    if(!talkMicStream){
      talkMicStream = await navigator.mediaDevices.getUserMedia({
        audio: {
          channelCount: 1,
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true
        }
      });
    }
    const ctx = new (window.AudioContext || window.webkitAudioContext)();
    const micSrc = ctx.createMediaStreamSource(talkMicStream);

    // 硬件级抗混叠低通滤波器 (Anti-Aliasing Filter)：
    // 切除 3400Hz 以上高频与杂音，彻底消除降采样至 8000Hz 时的混叠折叠失真与沙哑毛刺感
    const lowpass = ctx.createBiquadFilter();
    lowpass.type = 'lowpass';
    lowpass.frequency.value = 3400;
    lowpass.Q.value = 0.707;
    micSrc.connect(lowpass);
    talkMicFilter = lowpass;

    // 采用 2048 样本缓冲区，减少音频线程上下文切换开销
    const proc = ctx.createScriptProcessor(2048, 1, 1);
    lowpass.connect(proc);
    proc.connect(ctx.destination);
    talkMicProcessor = proc;

    talkMicBuffer = [];
    talkMicPhase = 0;
    const inRate = ctx.sampleRate;
    const outRate = 8000;
    const ratio = inRate / outRate;

    proc.onaudioprocess = function(e){
      if(!isTalkingMic) return;
      const input = e.inputBuffer.getChannelData(0);

      // 高保真线性插值重采样 (带相位累加，杜绝样本丢失与相位截断杂音)
      while(talkMicPhase < input.length){
        const i0 = Math.floor(talkMicPhase);
        const i1 = Math.min(i0 + 1, input.length - 1);
        const frac = talkMicPhase - i0;
        let s = input[i0] * (1 - frac) + input[i1] * frac;

        // 软动态余量 (0.92)，防止大声说话时突发 G.711 硬截断破音
        s = s * 0.92;
        if(s < -1) s = -1;
        if(s > 1) s = 1;
        const v = Math.round(s < 0 ? s * 32768 : s * 32767);
        talkMicBuffer.push(v);
        talkMicPhase += ratio;
      }
      talkMicPhase -= input.length;

      // 严格按 160 采样 (20ms/8000Hz 国标单帧标准) 切分成帧推送 WebSocket
      while(talkMicBuffer.length >= 160){
        const frame = talkMicBuffer.splice(0, 160);
        const pcm16 = new Int16Array(frame);

        // 实时音量计算驱动上行 VU 表
        let sumSq = 0;
        for(let j = 0; j < 160; j++){
          sumSq += pcm16[j] * pcm16[j];
        }
        const rms = Math.sqrt(sumSq / 160);
        let lvl = 0;
        if(rms >= 12.0){
          const db = 20.0 * Math.log10(rms / 32768.0);
          if(db > -52.0){
            lvl = Math.min(100, ((db + 52.0) / 50.0) * 100.0);
          }
        }
        applyVULevel('txVUBar', 'txVUTxt', Math.round(lvl));

        if(talkWS && talkWS.readyState === WebSocket.OPEN){
          talkWS.send(pcm16.buffer);
        }
      }
    };
  } catch(err){
    console.error('mic error:', err);
    showToast('无法启用麦克风: ' + err.message, 'error');
  }
}

function stopTalkMic(){
  isTalkingMic = false;
  const btn = document.getElementById('btnTalkMic');
  if(btn){
    btn.style.background = '#e11d48';
    btn.innerText = '🎙️ 按住对讲 (Push-to-Talk)';
  }
  if(talkMicProcessor){
    talkMicProcessor.disconnect();
    talkMicProcessor = null;
  }
  if(talkMicFilter){
    talkMicFilter.disconnect();
    talkMicFilter = null;
  }
  if(talkMicBuffer && talkMicBuffer.length > 0){
    const frame = new Int16Array(160);
    for(let i = 0; i < talkMicBuffer.length; i++){
      frame[i] = talkMicBuffer[i];
    }
    talkMicBuffer = [];
    if(talkWS && talkWS.readyState === WebSocket.OPEN){
      talkWS.send(frame.buffer);
    }
  }
  talkMicPhase = 0;
  applyVULevel('txVUBar', 'txVUTxt', 0);
}

async function setTalkUplinkMode(callId, mode){
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/talk/mode?callId=' + encodeURIComponent(callId) + '&mode=' + encodeURIComponent(mode), 'POST');
  if(j && j.ok){
    showToast('上行推流模式已切换为: ' + (mode==='mic'?'麦克风喊话':'模拟提示音'), 'success');
    refreshAll();
  }
}

async function stopTalkSession(callId){
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/talk/stop?callId=' + encodeURIComponent(callId), 'POST');
  if(j && j.ok){
    showToast('对讲会话已终止', 'info');
    if(talkFastPollTimer){
      clearInterval(talkFastPollTimer);
      talkFastPollTimer = null;
    }
    if(talkWS) talkWS.close();
    talkSpeakerEnabled = false;
    stopTalkMic();
    refreshAll();
  }
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
  document.getElementById('cfgSipServerId').value = sip.server_id || '';
  document.getElementById('cfgSipCharset').value = sip.charset || 'GB2312';
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
  p.sip.server_id = document.getElementById('cfgSipServerId').value.trim();
  p.sip.charset = document.getElementById('cfgSipCharset').value;
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
      server_id: id.substring(0, 10) + '2000000001',
      server_ip: sIp,
      server_port: sPort,
      local_ip: lIp,
      local_port: lPort,
      transport: transport,
      charset: 'GB2312',
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
  else if(!val) body = {channelId: chId, source: 'mp4', mp4: '', audioEnabled: true, audioSource: 'mp4'};
  else if(val.endsWith('.h264') || val.endsWith('.264')) body = {channelId: chId, source: 'file', h264: val};
  else body = {channelId: chId, source: 'mp4', mp4: val, audioEnabled: true, audioSource: 'mp4'};

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

// Alarm & Guard Control
function openAlarmModal(){
  const sel = document.getElementById('alarmChannelSelect');
  const chs = (activeDeviceData && activeDeviceData.profile && activeDeviceData.profile.device && activeDeviceData.profile.device.channels) || [];
  if(chs.length){
    sel.innerHTML = chs.map(function(c){
      return '<option value="' + esc(c.id) + '">' + esc(c.name) + ' (' + esc(c.id) + ')</option>';
    }).join('') + '<option value="' + esc(activeDeviceId) + '">主设备编码 (' + esc(activeDeviceId) + ')</option>';
  } else {
    sel.innerHTML = '<option value="' + esc(activeDeviceId) + '">主设备 (' + esc(activeDeviceId) + ')</option>';
  }
  openModal('alarmModal');
}

async function submitAlarm(){
  const chId = document.getElementById('alarmChannelSelect').value;
  const method = document.getElementById('alarmMethodSelect').value;
  const type = document.getElementById('alarmTypeSelect').value;
  const priority = document.getElementById('alarmPrioritySelect').value;
  const desc = document.getElementById('alarmDescInput').value.trim();
  const lon = parseFloat(document.getElementById('alarmLonInput').value) || 0;
  const lat = parseFloat(document.getElementById('alarmLatInput').value) || 0;
  const force = document.getElementById('alarmForceCheck') ? document.getElementById('alarmForceCheck').checked : false;

  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/alarm', 'POST', {
    channelId: chId,
    alarmMethod: method,
    alarmType: type,
    priority: priority,
    description: desc,
    longitude: lon,
    latitude: lat,
    force: force
  });
  if(j && j.ok){
    showToast('模拟报警通知已成功发出 (' + (j.record && j.record.latencyMs || 0) + 'ms) · 通道已切换为 ALARM', 'success');
    closeModal('alarmModal');
    refreshAll();
  } else if(j && j.suppressed){
    showToast('⚠️ 撤防门禁拦截: ' + j.error, 'warn');
    closeModal('alarmModal');
    refreshAll();
  } else if(j && j.error){
    showToast('报警上报失败: ' + j.error, 'error');
  }
}

async function loadDeviceAlarms(){
  if(!activeDeviceId) return;
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/alarms?limit=100');
  if(j && j.records){
    renderAlarmHistory(j.records);
  }
}

async function clearDeviceAlarms(){
  if(!activeDeviceId) return;
  if(!confirm('确定要清空当前设备的所有报警历史记录吗？')) return;
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/alarms/clear', 'POST');
  if(j && j.ok){
    showToast('已清空所有本地报警历史记录', 'success');
    loadDeviceAlarms();
  }
}

function renderAlarmHistory(records){
  const container = document.getElementById('alarmHistoryContainer');
  if(!container) return;
  if(!records || !records.length){
    container.innerHTML = '<div style="color:var(--text-dim);padding:24px;text-align:center">暂无报警上报记录。可点击上方快捷按钮或自定义高级报警模拟发送事件。</div>';
    return;
  }

  const methodNames = {
    '1': '1 - 电话报警',
    '2': '2 - 设备报警 (探头/SOS)',
    '3': '3 - 短信报警',
    '4': '4 - GPS/位置',
    '5': '5 - 视频报警',
    '6': '6 - 设备故障报警',
    '7': '7 - 其他报警'
  };
  const typeNames = {
    '1': '1 - 视频丢失',
    '2': '2 - 移动侦测',
    '5': '5 - 绊线越界 / SOS求助',
    '6': '6 - 周界入侵',
    '11': '11 - 视频遮挡/镜头篡改',
    '21': '21 - 存储设备故障/满',
    '22': '22 - 通信断开',
    '51': '51 - 违章停车'
  };

  let rows = records.map(function(r, idx){
    const mStr = methodNames[r.alarmMethod] || (r.alarmMethod ? (r.alarmMethod + '号方式') : '-');
    const tStr = typeNames[r.alarmType] || (r.alarmType && r.alarmType !== '0' ? (r.alarmType + '号类型') : '-');
    const priBadge = (r.priority === '1' || r.priority === 1) ? '<span class="badge off" style="font-size:10px">1级(最高紧急)</span>' :
      (r.priority === '2' || r.priority === 2) ? '<span class="badge warn" style="font-size:10px">2级(重要)</span>' :
      '<span class="badge" style="font-size:10px">' + (r.priority || 4) + '级</span>';

    let statusBadge = '';
    if(r.status === 'confirmed'){
      statusBadge = '<span class="badge on" style="font-size:10px">200 OK (' + (r.latencyMs || 0) + 'ms)</span>';
    } else if(r.status === 'suppressed'){
      statusBadge = '<span class="badge warn" style="font-size:10px" title="撤防状态拦截上报">⚠️ 撤防门禁拦截</span>';
    } else {
      statusBadge = '<span class="badge off" style="font-size:10px">失败</span>';
    }

    return '<tr>' +
      '<td style="color:var(--text-dim)">' + (idx + 1) + '</td>' +
      '<td style="font-family:var(--font-mono);color:#93c5fd">' + esc(r.time) + '</td>' +
      '<td style="font-family:var(--font-mono);font-size:11px">' + esc(r.channelId) + '</td>' +
      '<td>' + esc(mStr) + '</td>' +
      '<td>' + esc(tStr) + '</td>' +
      '<td>' + priBadge + '</td>' +
      '<td>' + statusBadge + '</td>' +
      '<td style="color:var(--text-main);max-width:260px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="' + esc(r.description) + '">' + esc(r.description || '-') + '</td>' +
    '</tr>';
  }).join('');

  container.innerHTML = '<table class="rec-table">' +
    '<thead>' +
      '<tr>' +
        '<th style="width:36px">#</th>' +
        '<th>上报时间</th>' +
        '<th>报警通道/设备</th>' +
        '<th>报警方式</th>' +
        '<th>事件类型</th>' +
        '<th>级别</th>' +
        '<th>平台响应</th>' +
        '<th>描述文本</th>' +
      '</tr>' +
    '</thead>' +
    '<tbody>' + rows + '</tbody>' +
  '</table>';
}

function renderChannelGuardTable(channels){
  const container = document.getElementById('channelGuardTableContainer');
  if(!container) return;
  if(!channels || !channels.length){
    container.innerHTML = '<div style="color:var(--text-dim);padding:14px;text-align:center">该设备暂无配置通道。可在上方配置下挂通道。</div>';
    return;
  }

  let rows = channels.map(function(ch){
    let dutyBadge = '<span class="badge off"><span class="dot"></span>OFFDUTY 已撤防</span>';
    if(ch.dutyStatus === 'ALARM' || ch.isAlarming){
      dutyBadge = '<span class="badge" style="background:#dc2626;color:#fff"><span class="dot"></span>🚨 ALARM 报警中</span>';
    } else if(ch.dutyStatus === 'ONDUTY' || ch.guardStatus === 'SetGuard'){
      dutyBadge = '<span class="badge on"><span class="dot"></span>🛡️ ONDUTY 已布防</span>';
    }

    return '<tr>' +
      '<td style="font-weight:700;color:#fff">' + esc(ch.name) + '</td>' +
      '<td style="font-family:var(--font-mono);font-size:11px">' + esc(ch.id) + '</td>' +
      '<td>' + dutyBadge + '</td>' +
      '<td>' +
        '<div style="display:flex;gap:6px;align-items:center">' +
          (ch.guardStatus === 'SetGuard' ?
            '<button class="btn btn-sm" onclick="setChannelGuardState(\'' + esc(ch.id) + '\',\'ResetGuard\')">🔓 单独撤防</button>' :
            '<button class="btn btn-sm btn-primary" onclick="setChannelGuardState(\'' + esc(ch.id) + '\',\'SetGuard\')">🛡️ 单独布防</button>'
          ) +
          (ch.dutyStatus === 'ALARM' || ch.isAlarming ?
            '<button class="btn btn-sm btn-danger" onclick="setChannelGuardState(\'' + esc(ch.id) + '\',\'ResetAlarm\')">🔕 报警复位</button>' :
            '<button class="btn btn-sm" onclick="setChannelGuardState(\'' + esc(ch.id) + '\',\'ResetAlarm\')">🔕 复位</button>'
          ) +
        '</div>' +
      '</td>' +
    '</tr>';
  }).join('');

  container.innerHTML = '<table class="rec-table">' +
    '<thead>' +
      '<tr>' +
        '<th>通道名称</th>' +
        '<th>国标编码 (DeviceID)</th>' +
        '<th>国标防区状态 (DutyStatus)</th>' +
        '<th>通道独立设防控制</th>' +
      '</tr>' +
    '</thead>' +
    '<tbody>' + rows + '</tbody>' +
  '</table>';
}

function renderAlarmsTab(st){
  const overallBadge = document.getElementById('alarmGuardOverallBadge');
  if(overallBadge){
    if(st.guardStatus === 'SetGuard' || st.dutyStatus === 'ONDUTY'){
      overallBadge.className = 'badge on';
      overallBadge.innerHTML = '<span class="dot"></span>已布防 (SetGuard / ONDUTY)';
    } else if(st.guardStatus === 'ResetGuard' || st.dutyStatus === 'OFFDUTY'){
      overallBadge.className = 'badge off';
      overallBadge.innerHTML = '<span class="dot"></span>已撤防 (ResetGuard / OFFDUTY)';
    } else {
      overallBadge.className = 'badge';
      overallBadge.innerHTML = '初始未设防';
    }
  }

  const quickChSel = document.getElementById('quickAlarmChannelSelect');
  if(quickChSel && activeDeviceData && activeDeviceData.profile && activeDeviceData.profile.device){
    const chs = activeDeviceData.profile.device.channels || [];
    const prev = quickChSel.value;
    quickChSel.innerHTML = chs.map(function(c, idx){
      return '<option value="' + esc(c.id) + '">' + (idx === 0 ? '【默认第1通道】' : '【通道】') + esc(c.name) + ' (' + esc(c.id) + ')</option>';
    }).join('') + '<option value="' + esc(activeDeviceId) + '">【主设备根节点】' + esc(activeDeviceId) + '</option>';
    if(prev) quickChSel.value = prev;
    updateQuickAlarmChannelBadge();
  }

  renderChannelGuardTable(st.channels || []);

  const autoBadge = document.getElementById('autoAlarmBadge');
  const btnToggle = document.getElementById('btnToggleAutoAlarm');
  const intervalInput = document.getElementById('autoAlarmInterval');
  if(autoBadge && btnToggle){
    const auto = st.autoAlarm || {};
    if(auto.enabled){
      autoBadge.className = 'badge on';
      autoBadge.innerHTML = '<span class="dot"></span>运行中 (每 ' + (auto.intervalSec || 15) + ' 秒)';
      btnToggle.className = 'btn btn-sm btn-danger';
      btnToggle.textContent = '停止自动报警';
    } else {
      autoBadge.className = 'badge off';
      autoBadge.innerHTML = '<span class="dot"></span>已停止';
      btnToggle.className = 'btn btn-sm btn-primary';
      btnToggle.textContent = '开启自动报警';
    }
    if(intervalInput && auto.intervalSec) intervalInput.value = auto.intervalSec;
  }

  loadDeviceAlarms();
}

function updateQuickAlarmChannelBadge(){
  const sel = document.getElementById('quickAlarmChannelSelect');
  const badge = document.getElementById('quickAlarmChannelBadge');
  if(!sel || !badge) return;
  const opt = sel.options[sel.selectedIndex];
  if(opt){
    const nameOnly = opt.text.replace(/【.*?】/, '').split('(')[0].trim();
    badge.textContent = '当前目标: ' + nameOnly;
    badge.className = 'badge on';
  }
}

async function setDeviceGuardState(cmd){
  if(!activeDeviceId) return;
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/guard', 'POST', {
    guardCmd: cmd
  });
  if(j && j.ok){
    showToast('整机布防控制指令已生效: ' + cmd, 'success');
    refreshAll();
  }
}

async function setChannelGuardState(channelId, cmd){
  if(!activeDeviceId) return;
  let body = { channelId: channelId };
  if(cmd === 'ResetAlarm'){
    body.alarmCmd = 'ResetAlarm';
  } else {
    body.guardCmd = cmd;
  }
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/guard', 'POST', body);
  if(j && j.ok){
    showToast('通道 [' + channelId + '] 防区指令已生效: ' + cmd, 'success');
    refreshAll();
  }
}

async function toggleAutoAlarm(){
  if(!activeDeviceId) return;
  const st = (activeDeviceData && activeDeviceData.status) || {};
  const currentAuto = st.autoAlarm || {};
  const nextEnabled = !currentAuto.enabled;
  const interval = parseInt(document.getElementById('autoAlarmInterval').value) || 15;

  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/alarm/auto', 'POST', {
    enabled: nextEnabled,
    intervalSec: interval
  });
  if(j && j.ok){
    showToast(nextEnabled ? ('已开启周期自动报警 (间隔 ' + interval + ' 秒，仅针对已布防通道)') : '已停止周期自动报警', nextEnabled ? 'success' : 'info');
    refreshAll();
  }
}

async function triggerQuickAlarm(kind){
  if(!activeDeviceId) return;
  const chs = (activeDeviceData && activeDeviceData.profile && activeDeviceData.profile.device && activeDeviceData.profile.device.channels) || [];
  let sel = document.getElementById('quickAlarmChannelSelect');
  let chId = (sel && sel.value) ? sel.value : (chs.length > 0 ? chs[0].id : activeDeviceId);
  let selOpt = sel ? sel.options[sel.selectedIndex] : null;
  let chLabel = selOpt ? selOpt.text.replace(/【.*?】/, '').split('(')[0].trim() : chId;
  let force = document.getElementById('quickAlarmForce') ? document.getElementById('quickAlarmForce').checked : false;

  let body = { channelId: chId, priority: '3', alarmMethod: '5', alarmType: '2', description: '通道模拟报警', force: force };
  switch(kind){
    case 'motion':
      body.alarmMethod = '5'; // 视频报警 (Method 5)
      body.alarmType = '2';   // 运动目标检测 (Type 2，国标 Table A.2: 移动侦测报警)
      body.priority = '3';
      body.description = 'Web控制台触发：画面移动侦测报警';
      break;
    case 'intrusion':
      body.alarmMethod = '5'; // 视频报警
      body.alarmType = '6';   // 周界入侵
      body.priority = '2';
      body.description = 'Web控制台触发：防区周界入侵报警';
      break;
    case 'tamper':
      body.alarmMethod = '5'; // 视频报警
      body.alarmType = '11';  // 视频遮挡
      body.priority = '3';
      body.description = 'Web控制台触发：摄像机镜头被遮挡或篡改';
      break;
    case 'videoloss':
      body.alarmMethod = '5'; // 视频报警
      body.alarmType = '1';   // 视频丢失
      body.priority = '2';
      body.description = 'Web控制台触发：视频信号丢失报警';
      break;
    case 'sos':
      body.alarmMethod = '2'; // 设备报警 (Method 2)
      body.alarmType = '5';   // 人工求助 (SOS)
      body.priority = '1';    // 一级最高紧急 (具有24h豁免权)
      body.description = 'Web控制台触发：人工紧急报警/求助触发';
      break;
    case 'diskfault':
      body.alarmMethod = '6'; // 设备故障 (Method 6)
      body.alarmType = '21';  // 存储介质故障
      body.priority = '2';
      body.description = 'Web控制台触发：存储介质故障/硬盘满';
      break;
  }

  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/alarm', 'POST', body);
  if(j && j.ok){
    showToast('【' + chLabel + '】报警上报成功 (200 OK) · 通道已切换为 ALARM', 'success');
    refreshAll();
  } else if(j && j.suppressed){
    showToast('⚠️ 撤防门禁拦截【' + chLabel + '】: ' + j.error, 'warn');
    refreshAll();
  } else if(j && j.error){
    showToast('【' + chLabel + '】报警上报失败: ' + j.error, 'error');
  }
}

// ==================== GPS & MobilePosition Control ====================
let selectedGPSChannelId = null; // null or "" means first channel / master
let gpsConfigLoadedKey = null;

function onGPSPatternChange(){
  const pat = document.getElementById('gpsCfgPattern').value;
  const paramFields = document.getElementById('gpsParamFields');
  const followHint = document.getElementById('gpsFollowHint');
  if(pat === 'follow'){
    if(paramFields) paramFields.style.display = 'none';
    if(followHint) followHint.style.display = 'block';
  } else {
    if(paramFields) paramFields.style.display = 'block';
    if(followHint) followHint.style.display = 'none';
  }
}

function selectGPSChannel(channelId){
  selectedGPSChannelId = channelId;
  loadGPSStatus(true);
}

async function loadGPSStatus(loadConfigToo){
  if(!activeDeviceId) return;
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/gps');
  if(!j) return;

  const masterSt = j.gps || {};
  const masterCfg = j.config || {};
  const chStatuses = j.channelStatuses || {};
  const chConfigs = j.channelConfigs || {};

  const devChs = (activeDeviceData && activeDeviceData.profile && activeDeviceData.profile.device && activeDeviceData.profile.device.channels) || [];

  // Default selected channel to the first channel if not set
  if(selectedGPSChannelId === null){
    if(devChs.length > 0){
      selectedGPSChannelId = devChs[0].id;
    } else {
      selectedGPSChannelId = '__master__';
    }
  }

  // Render channel selection tabs / pills
  const tabContainer = document.getElementById('gpsChannelTabs');
  if(tabContainer){
    let tabsHtml = '';
    devChs.forEach(function(c, idx){
      const isSel = (selectedGPSChannelId === c.id);
      const cSt = chStatuses[c.id] || {};
      const isRunning = cSt.enabled;
      const dotColor = isRunning ? 'var(--ok)' : 'var(--text-dim)';
      const btnClass = isSel ? 'btn btn-sm btn-primary' : 'btn btn-sm';
      tabsHtml += '<button class="' + btnClass + '" style="display:flex;align-items:center;gap:6px" onclick="selectGPSChannel(\'' + esc(c.id) + '\')">' +
        '<span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:' + dotColor + '"></span>' +
        '通道 ' + (idx + 1) + ': ' + esc(c.name) +
        '</button>';
    });

    // Master vehicle option
    const isMasterSel = (selectedGPSChannelId === '__master__' || selectedGPSChannelId === activeDeviceId);
    const mRunning = masterSt.enabled;
    const mDotColor = mRunning ? 'var(--cyan)' : 'var(--text-dim)';
    const mBtnClass = isMasterSel ? 'btn btn-sm btn-primary' : 'btn btn-sm';
    tabsHtml += '<button class="' + mBtnClass + '" style="display:flex;align-items:center;gap:6px" onclick="selectGPSChannel(\'__master__\')">' +
      '<span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:' + mDotColor + '"></span>' +
      '★ 设备/车体主GPS' +
      '</button>';

    tabContainer.innerHTML = tabsHtml;
  }

  // Determine current active target's status and config
  let curSt, curCfg, curTitle, curTargetLabel;
  if(selectedGPSChannelId === '__master__' || selectedGPSChannelId === activeDeviceId){
    curSt = masterSt;
    curCfg = masterCfg;
    curTitle = '🛰️ 【车体/设备主GPS】实时移动位置仪表';
    curTargetLabel = '设备主GPS (' + activeDeviceId + ')';
  } else {
    curSt = chStatuses[selectedGPSChannelId] || masterSt;
    curCfg = chConfigs[selectedGPSChannelId] || masterCfg;
    let foundName = selectedGPSChannelId;
    for(let i=0; i<devChs.length; i++){
      if(devChs[i].id === selectedGPSChannelId){
        foundName = devChs[i].name;
        break;
      }
    }
    curTitle = '🛰️ 【通道: ' + esc(foundName) + '】实时移动位置仪表';
    curTargetLabel = '通道: ' + esc(foundName) + ' (' + selectedGPSChannelId + ')';
  }

  const gaugeTitle = document.getElementById('gpsGaugeTitle');
  if(gaugeTitle) gaugeTitle.innerHTML = curTitle;

  const targetBadge = document.getElementById('gpsTargetBadge');
  if(targetBadge) targetBadge.textContent = curTargetLabel;

  const badge = document.getElementById('gpsActiveBadge');
  if(badge){
    if(curSt.enabled){
      badge.className = 'badge on';
      let patName = curSt.pattern || curCfg.pattern || 'circle';
      if(patName === 'follow') patName = '跟随主车';
      badge.innerHTML = '<span class="dot"></span>正在仿真巡航 (' + esc(patName) + ')';
    } else {
      badge.className = 'badge stopped';
      badge.innerHTML = '<span class="dot"></span>已停用上报';
    }
  }

  const lonEl = document.getElementById('gpsValLon');
  const latEl = document.getElementById('gpsValLat');
  const spdEl = document.getElementById('gpsValSpeed');
  const dirEl = document.getElementById('gpsValDir');
  const altEl = document.getElementById('gpsValAlt');
  const timeEl = document.getElementById('gpsValTime');

  if(lonEl) lonEl.textContent = (curSt.longitude !== undefined ? curSt.longitude.toFixed(6) : '116.397428');
  if(latEl) latEl.textContent = (curSt.latitude !== undefined ? curSt.latitude.toFixed(6) : '39.909230');
  if(spdEl) spdEl.textContent = (curSt.speed !== undefined ? curSt.speed.toFixed(1) : '0.0') + ' km/h';
  if(dirEl) dirEl.textContent = (curSt.direction !== undefined ? curSt.direction.toFixed(1) : '0.0') + '°';
  if(altEl) altEl.textContent = (curSt.altitude !== undefined ? curSt.altitude.toFixed(1) : '0.0') + ' m';
  if(timeEl) timeEl.textContent = curSt.time || '-';

  // Fill the form if explicitly requested or channel switched
  const curKey = activeDeviceId + '_' + selectedGPSChannelId;
  if(loadConfigToo || gpsConfigLoadedKey !== curKey){
    gpsConfigLoadedKey = curKey;

    const enSel = document.getElementById('gpsCfgEnabled');
    if(enSel) enSel.value = curCfg.enabled ? 'true' : 'false';
    const modeSel = document.getElementById('gpsCfgMode');
    if(modeSel) modeSel.value = curCfg.mode || 'both';
    const patSel = document.getElementById('gpsCfgPattern');
    if(patSel) patSel.value = curCfg.pattern || (selectedGPSChannelId==='__master__'?'circle':'follow');

    const cfgLon = document.getElementById('gpsCfgLon');
    if(cfgLon) cfgLon.value = curCfg.longitude !== undefined ? curCfg.longitude : 116.397428;
    const cfgLat = document.getElementById('gpsCfgLat');
    if(cfgLat) cfgLat.value = curCfg.latitude !== undefined ? curCfg.latitude : 39.909230;
    const cfgAlt = document.getElementById('gpsCfgAlt');
    if(cfgAlt) cfgAlt.value = curCfg.altitude !== undefined ? curCfg.altitude : 50;
    const cfgSpeed = document.getElementById('gpsCfgSpeed');
    if(cfgSpeed) cfgSpeed.value = curCfg.speed !== undefined ? curCfg.speed : 30;
    const cfgRadius = document.getElementById('gpsCfgRadius');
    if(cfgRadius) cfgRadius.value = curCfg.radius !== undefined ? curCfg.radius : 500;
    const cfgInterval = document.getElementById('gpsCfgInterval');
    if(cfgInterval) cfgInterval.value = curCfg.interval !== undefined ? curCfg.interval : 5;

    onGPSPatternChange();
  }
}

async function saveGPSConfig(){
  if(!activeDeviceId) return;
  const targetCh = (selectedGPSChannelId === '__master__') ? '' : selectedGPSChannelId;
  const body = {
    enabled: document.getElementById('gpsCfgEnabled').value === 'true',
    mode: document.getElementById('gpsCfgMode').value,
    pattern: document.getElementById('gpsCfgPattern').value,
    channel_id: targetCh,
    longitude: parseFloat(document.getElementById('gpsCfgLon').value) || 0,
    latitude: parseFloat(document.getElementById('gpsCfgLat').value) || 0,
    altitude: parseFloat(document.getElementById('gpsCfgAlt').value) || 0,
    speed: parseFloat(document.getElementById('gpsCfgSpeed').value) || 0,
    radius: parseFloat(document.getElementById('gpsCfgRadius').value) || 0,
    interval: parseInt(document.getElementById('gpsCfgInterval').value, 10) || 5
  };

  const q = targetCh ? ('?channelId=' + encodeURIComponent(targetCh)) : '';
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/gps/config' + q, 'POST', body);
  if(j && j.ok){
    showToast('当前通道 GPS 轨迹参数已保存并生效', 'success');
    loadGPSStatus(false);
    refreshAll();
  }
}

async function triggerManualGPSReport(){
  if(!activeDeviceId) return;
  const targetCh = (selectedGPSChannelId === '__master__') ? '' : selectedGPSChannelId;
  const q = targetCh ? ('?channelId=' + encodeURIComponent(targetCh)) : '';
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/gps/report' + q, 'POST');
  if(j && j.ok){
    const st = j.gps || {};
    showToast('已向平台发送当前通道位置上报 (经度:' + (st.longitude||0).toFixed(5) + ', 纬度:' + (st.latitude||0).toFixed(5) + ')', 'success');
    loadGPSStatus(false);
  }
}

async function syncAllChannelsGPS(followMode){
  if(!activeDeviceId) return;
  const promptMsg = followMode ?
    '确定将所有下挂通道设为【跟随主车】模式吗？\n所有通道将共享设备主轨迹，模拟车载同一载具上的多摄像头。' :
    '确定将当前配置完整【克隆】给所有下挂通道吗？\n所有通道将复制相同的经纬度与轨迹算法独立演进。';

  if(!confirm(promptMsg)) return;

  const targetCh = (selectedGPSChannelId === '__master__') ? '' : selectedGPSChannelId;
  const body = {
    enabled: document.getElementById('gpsCfgEnabled').value === 'true',
    mode: document.getElementById('gpsCfgMode').value,
    pattern: followMode ? 'follow' : document.getElementById('gpsCfgPattern').value,
    channel_id: targetCh,
    longitude: parseFloat(document.getElementById('gpsCfgLon').value) || 0,
    latitude: parseFloat(document.getElementById('gpsCfgLat').value) || 0,
    altitude: parseFloat(document.getElementById('gpsCfgAlt').value) || 0,
    speed: parseFloat(document.getElementById('gpsCfgSpeed').value) || 0,
    radius: parseFloat(document.getElementById('gpsCfgRadius').value) || 0,
    interval: parseInt(document.getElementById('gpsCfgInterval').value, 10) || 5
  };

  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/gps/sync', 'POST', {
    config: body,
    followMode: followMode
  });
  if(j && j.ok){
    showToast(followMode ? '已将所有通道设置为跟随主车轨迹模式' : '已将配置参数克隆到所有通道', 'success');
    loadGPSStatus(true);
    refreshAll();
  }
}

// ==================== Subscriptions & Incremental Notify ====================
async function loadSubscriptions(){
  if(!activeDeviceId) return;
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/subscriptions');
  if(!j) return;

  const subs = j.subscribers || [];
  const container = document.getElementById('subscriptionTableContainer');
  if(!container) return;

  // Also populate catalog notify channel options
  const catChSel = document.getElementById('catNotifyChannelSelect');
  if(catChSel && activeDeviceData && activeDeviceData.profile && activeDeviceData.profile.device){
    const chs = activeDeviceData.profile.device.channels || [];
    const prev = catChSel.value;
    catChSel.innerHTML = '<option value="' + esc(activeDeviceId) + '">【主设备自身】' + esc(activeDeviceId) + '</option>' +
      chs.map(function(c){
        return '<option value="' + esc(c.id) + '">' + esc(c.name) + ' (' + esc(c.id) + ')</option>';
      }).join('');
    if(prev) catChSel.value = prev;
  }

  if(!subs.length){
    container.innerHTML = '<div style="color:var(--text-dim);padding:24px;text-align:center">当前无活动订阅会话。<br/><span style="font-size:12px;color:var(--text-muted)">当上级平台（如 WVP 或国标联网平台）发起目录订阅 (Catalog) 或位置订阅 (presence/MobilePosition) 并建立成功后，将在此展示订阅对话、Contact URI 与剩余租期。</span></div>';
    return;
  }

  let rows = subs.map(function(s, idx){
    let evBadge = '<span class="badge" style="font-size:10px">' + esc(s.event) + '</span>';
    if(s.event === 'Catalog') evBadge = '<span class="badge on" style="font-size:10px">📁 目录订阅 (Catalog)</span>';
    else if(s.event === 'presence') evBadge = '<span class="badge warn" style="font-size:10px">🛰️ 状态订阅 (presence)</span>';
    else if(s.event === 'MobilePosition') evBadge = '<span class="badge" style="background:#06b6d4;color:#fff;font-size:10px">📍 位置订阅 (MobilePosition)</span>';

    const subTime = s.subscribedAt ? s.subscribedAt.replace('T', ' ').substring(0, 19) : '-';
    let remSec = s.expiresSec || 0;
    if(s.expiresAt){
      const diff = Math.round((new Date(s.expiresAt).getTime() - Date.now()) / 1000);
      remSec = diff > 0 ? diff : 0;
    }

    return '<tr>' +
      '<td style="color:var(--text-dim)">' + (idx + 1) + '</td>' +
      '<td>' + evBadge + '</td>' +
      '<td style="font-family:var(--font-mono);color:#93c5fd">' + esc(s.platformId || '-') + '</td>' +
      '<td style="font-family:var(--font-mono);font-size:11px;color:var(--text-dim)">' + esc(s.contactUri || '-') + '</td>' +
      '<td style="font-family:var(--font-mono);font-size:11px;color:var(--text-muted);max-width:180px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="' + esc(s.callId) + '">' + esc(s.callId) + '</td>' +
      '<td style="font-size:11px;color:var(--text-dim)">' + esc(subTime) + '</td>' +
      '<td><span class="badge ' + (remSec > 0 ? 'on' : 'off') + '" style="font-size:10px">' + remSec + ' 秒 / 租约 ' + (s.expiresSec||0) + 's</span></td>' +
    '</tr>';
  }).join('');

  container.innerHTML = '<table class="rec-table">' +
    '<thead>' +
      '<tr>' +
        '<th style="width:36px">#</th>' +
        '<th>事件类型 (Event)</th>' +
        '<th>平台编码 (PlatformID)</th>' +
        '<th>目标 Contact URI</th>' +
        '<th>对话 Call-ID</th>' +
        '<th>订阅时间</th>' +
        '<th>剩余有效时间 (Expires)</th>' +
      '</tr>' +
    '</thead>' +
    '<tbody>' + rows + '</tbody>' +
  '</table>';
}

async function submitCatalogNotify(){
  if(!activeDeviceId) return;
  const chSel = document.getElementById('catNotifyChannelSelect');
  const evSel = document.getElementById('catNotifyEventSelect');
  if(!chSel || !evSel) return;

  const chId = chSel.value;
  const ev = evSel.value;

  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/catalog/notify', 'POST', {
    channelId: chId,
    event: ev
  });
  if(j && j.ok){
    showToast('📢 已成功向活跃订阅者广播目录增量通知 (Event: ' + ev + ')', 'success');
  }
}

function quickCatalogNotify(chId){
  const tabBtn = document.getElementById('tabBtnSubs');
  if(tabBtn) switchWorkbenchTab('subs', tabBtn);
  setTimeout(function(){
    const chSel = document.getElementById('catNotifyChannelSelect');
    if(chSel) chSel.value = chId;
  }, 100);
}

// ==================== Remote Device Config & Control ====================
let configCtrlActiveChannel = '';

async function loadConfigCtrlData(manual){
  if(!activeDeviceId) return;
  const dev = (devices || []).find(function(d){ return d.id === activeDeviceId; });
  const st = (activeDeviceData && activeDeviceData.status) ? activeDeviceData.status : (dev || {});

  // Update Recording badge & toggle button
  const recBadge = document.getElementById('cfgCtrlRecordBadge');
  const btnToggleRec = document.getElementById('btnToggleRecord');
  const isRec = !!st.recording;
  if(recBadge){
    recBadge.className = 'badge ' + (isRec ? 'on' : 'off');
    recBadge.textContent = isRec ? '⏺️ 录像中 (ON)' : '⏹️ 未录像 (OFF)';
  }
  if(btnToggleRec){
    btnToggleRec.textContent = isRec ? '⏹️ 停止录像 (Record OFF)' : '⏺️ 开启录像 (Record ON)';
  }

  // Update Reboot badge
  const rebootBadge = document.getElementById('cfgCtrlRebootBadge');
  if(rebootBadge){
    if(st.rebooting){
      rebootBadge.className = 'badge warn';
      rebootBadge.textContent = '⚠️ 正在重启中...';
    } else {
      rebootBadge.className = 'badge on';
      rebootBadge.textContent = '🟢 运行正常';
    }
  }

  // Update Device Time & Offset
  const devTimeEl = document.getElementById('cfgCtrlDeviceTime');
  const devOffsetEl = document.getElementById('cfgCtrlTimeOffset');
  const offsetSec = st.timeOffsetSec || 0;
  if(devTimeEl){
    if(st.deviceTime){
      devTimeEl.textContent = st.deviceTime.replace('T', ' ').substring(0, 19);
    } else {
      const now = new Date(Date.now() + offsetSec * 1000);
      devTimeEl.textContent = now.toISOString().replace('T', ' ').substring(0, 19);
    }
  }
  if(devOffsetEl){
    if(offsetSec === 0){
      devOffsetEl.textContent = '0s (与宿主机严格同步)';
      devOffsetEl.style.color = 'var(--text-muted)';
    } else {
      devOffsetEl.textContent = (offsetSec > 0 ? '+' : '') + offsetSec + 's (平台校时偏移)';
      devOffsetEl.style.color = '#38bdf8';
    }
  }

  // Populate channel dropdowns
  const iframeChSel = document.getElementById('cfgCtrlIFrameChannel');
  const homeChSel = document.getElementById('homePosChannelSelect');
  const channels = (activeDeviceData && activeDeviceData.profile && activeDeviceData.profile.device && activeDeviceData.profile.device.channels) || [];
  
  if(channels.length > 0 && !configCtrlActiveChannel){
    configCtrlActiveChannel = channels[0].id;
  }

  if(iframeChSel && iframeChSel.options.length !== channels.length){
    const prev = iframeChSel.value;
    iframeChSel.innerHTML = channels.map(function(c){
      return '<option value="' + esc(c.id) + '">' + esc(c.name) + ' (' + esc(c.id) + ')</option>';
    }).join('');
    if(prev) iframeChSel.value = prev;
    else if(channels.length > 0) iframeChSel.value = channels[0].id;
  }

  if(homeChSel && homeChSel.options.length !== channels.length){
    const prev = homeChSel.value;
    homeChSel.innerHTML = channels.map(function(c){
      return '<option value="' + esc(c.id) + '">' + esc(c.name) + ' (' + esc(c.id) + ')</option>';
    }).join('');
    if(prev) homeChSel.value = prev;
    else if(channels.length > 0) homeChSel.value = channels[0].id;
    configCtrlActiveChannel = homeChSel.value;
  }

  // Fetch PTZ Home Position info for current selected channel
  const targetChannel = homeChSel ? homeChSel.value : configCtrlActiveChannel;
  if(targetChannel){
    try{
      const ptzRes = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/ptz?channel=' + encodeURIComponent(targetChannel));
      if(ptzRes && ptzRes.ptz){
        renderHomePositionStatus(ptzRes.ptz);
      }
    }catch(e){}
  }

  // Fetch Audit Logs
  try{
    const auditRes = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/control/events');
    if(auditRes && auditRes.events){
      renderAuditLogs(auditRes.events);
    }
  }catch(e){}

  if(manual){
    showToast('已刷新设备控制状态与审计流水', 'info');
  }
}

function onHomePosChannelChange(){
  const homeChSel = document.getElementById('homePosChannelSelect');
  if(homeChSel){
    configCtrlActiveChannel = homeChSel.value;
    loadConfigCtrlData(false);
  }
}

function renderHomePositionStatus(ptz){
  const badge = document.getElementById('homePositionBadge');
  const enCheck = document.getElementById('homePosEnabled');
  const presetIn = document.getElementById('homePosPreset');
  const resetIn = document.getElementById('homePosResetSec');
  const cdInfo = document.getElementById('homePosCountdownInfo');

  if(enCheck) enCheck.checked = !!ptz.homePositionEnabled;
  if(presetIn && ptz.homePositionPreset) presetIn.value = ptz.homePositionPreset;
  if(resetIn && ptz.homePositionResetSec) resetIn.value = ptz.homePositionResetSec;

  if(badge){
    if(ptz.homePositionEnabled){
      badge.className = 'badge on';
      badge.textContent = '已启用 (预置位 #' + (ptz.homePositionPreset || 1) + ')';
    } else {
      badge.className = 'badge off';
      badge.textContent = '未启用';
    }
  }

  if(cdInfo){
    if(ptz.homePositionEnabled){
      const cd = ptz.homeCountdownSec || 0;
      if(cd > 0){
        cdInfo.innerHTML = '⏱️ 空闲中：距自动归位还剩 <b style="color:var(--primary)">' + cd + 's</b>';
      } else {
        cdInfo.innerHTML = '✅ 当前处于守望位或转动中';
      }
    } else {
      cdInfo.textContent = '尚未配置守望位';
    }
  }
}

function renderAuditLogs(events){
  const container = document.getElementById('cfgCtrlAuditContainer');
  if(!container) return;

  if(!events || events.length === 0){
    container.innerHTML = '<div style="color:var(--text-dim);padding:24px;text-align:center">暂无远程配置或控制操作流水。<br/><span style="font-size:12px;color:var(--text-muted)">当上级平台下发 ConfigDownload / DeviceConfig / DeviceControl，或通过此工作台执行操作时，审计记录将自动呈现在此。</span></div>';
    return;
  }

  // Newest first
  const rows = events.map(function(ev, idx){
    const t = ev.time ? ev.time.replace('T', ' ').substring(0, 19) : '-';
    let cmdBadge = '<span class="badge" style="font-size:10px">' + esc(ev.cmdType) + '</span>';
    if(ev.cmdType === 'ConfigDownload') cmdBadge = '<span class="badge" style="background:#0284c7;color:#fff;font-size:10px">📥 ConfigDownload</span>';
    else if(ev.cmdType === 'DeviceConfig') cmdBadge = '<span class="badge" style="background:#8b5cf6;color:#fff;font-size:10px">⚙️ DeviceConfig</span>';
    else if(ev.cmdType === 'DeviceControl') cmdBadge = '<span class="badge" style="background:#f59e0b;color:#fff;font-size:10px">🎮 DeviceControl</span>';

    let srcBadge = '<span style="font-size:11px;color:var(--text-muted)">' + esc(ev.source) + '</span>';
    if(ev.source && ev.source.indexOf('SIP') >= 0) srcBadge = '<span style="color:#38bdf8;font-weight:600">🌐 ' + esc(ev.source) + '</span>';
    else if(ev.source && ev.source.indexOf('UI') >= 0) srcBadge = '<span style="color:#a78bfa;font-weight:600">🖥️ ' + esc(ev.source) + '</span>';

    let statusBadge = '<span class="badge on" style="font-size:10px">SUCCESS</span>';
    if(ev.status && ev.status.toLowerCase().indexOf('fail') >= 0){
      statusBadge = '<span class="badge danger" style="font-size:10px">' + esc(ev.status) + '</span>';
    } else if(ev.status){
      statusBadge = '<span class="badge on" style="font-size:10px">' + esc(ev.status) + '</span>';
    }

    return '<tr>' +
      '<td style="color:var(--text-dim)">' + (idx + 1) + '</td>' +
      '<td style="font-size:11px;color:var(--text-dim)">' + esc(t) + '</td>' +
      '<td>' + cmdBadge + '</td>' +
      '<td>' + srcBadge + '</td>' +
      '<td style="font-family:var(--font-mono);font-size:11px;color:#93c5fd">' + esc(ev.channelId || '-') + '</td>' +
      '<td style="font-size:12px;color:var(--text-main)">' + esc(ev.detail) + '</td>' +
      '<td>' + statusBadge + '</td>' +
    '</tr>';
  }).join('');

  container.innerHTML = '<table class="rec-table">' +
    '<thead>' +
      '<tr>' +
        '<th style="width:36px">#</th>' +
        '<th style="width:140px">时间戳</th>' +
        '<th style="width:140px">指令类型</th>' +
        '<th style="width:130px">操作来源</th>' +
        '<th style="width:170px">通道/对象</th>' +
        '<th>事件详情</th>' +
        '<th style="width:90px">执行状态</th>' +
      '</tr>' +
    '</thead>' +
    '<tbody>' + rows + '</tbody>' +
  '</table>';
}

async function triggerControlIFrame(){
  if(!activeDeviceId) return;
  const sel = document.getElementById('cfgCtrlIFrameChannel');
  const chId = sel ? sel.value : '';
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/control/iframe', 'POST', {
    channelId: chId
  });
  if(j && j.ok){
    showToast(j.applied ? '⚡ 强制关键帧指令已下发并注入推流会话 (SPS/PPS+IDR)' : '⚡ 关键帧请求已记录（当前通道尚未处于活跃推流中）', 'success');
    loadConfigCtrlData(false);
  }
}

async function triggerControlReboot(){
  if(!activeDeviceId) return;
  if(!confirm('确定要模拟远程重启该设备吗？\n设备将注销 SIP 注册、停止推流，3秒后自动重新启动并重新注册。')) return;
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/control/reboot', 'POST', {});
  if(j && j.ok){
    showToast('⚠️ 设备远程重启流程已启动，3秒后将自动重新注册', 'warning');
    setTimeout(function(){ refreshAll(); }, 500);
    setTimeout(function(){ refreshAll(); }, 3500);
  }
}

async function triggerControlToggleRecord(){
  if(!activeDeviceId) return;
  const dev = (devices || []).find(function(d){ return d.id === activeDeviceId; });
  const st = (activeDeviceData && activeDeviceData.status) ? activeDeviceData.status : (dev || {});
  const nextRec = !st.recording;
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/control/record', 'POST', {
    recording: nextRec
  });
  if(j && j.ok){
    showToast('录像状态已切换为: ' + (nextRec ? 'ON (正在录像)' : 'OFF (未录像)'), 'success');
    refreshAll();
    loadConfigCtrlData(false);
  }
}

async function triggerControlResetTime(){
  if(!activeDeviceId) return;
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/control/time', 'POST', {
    reset: true
  });
  if(j && j.ok){
    showToast('设备时钟已重置，与当前宿主机系统时间对齐', 'success');
    refreshAll();
    loadConfigCtrlData(false);
  }
}

async function promptControlCustomTime(){
  if(!activeDeviceId) return;
  const currentVal = (new Date()).toISOString().replace('T', ' ').substring(0, 19);
  const val = prompt('请输入要模拟设置的设备时间 (格式: YYYY-MM-DD HH:mm:ss 或 HH:mm:ss):', currentVal);
  if(!val) return;
  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/control/time', 'POST', {
    time: val.trim()
  });
  if(j && j.ok){
    showToast('设备虚拟时钟校准成功: ' + j.deviceTime, 'success');
    refreshAll();
    loadConfigCtrlData(false);
  }
}

async function saveHomePositionConfig(){
  if(!activeDeviceId) return;
  const chSel = document.getElementById('homePosChannelSelect');
  const enCheck = document.getElementById('homePosEnabled');
  const presetIn = document.getElementById('homePosPreset');
  const resetIn = document.getElementById('homePosResetSec');
  if(!chSel) return;

  const chId = chSel.value;
  const enabled = !!(enCheck && enCheck.checked);
  const preset = parseInt(presetIn ? presetIn.value : '1', 10) || 1;
  const resetSec = parseInt(resetIn ? resetIn.value : '30', 10) || 30;

  const j = await api('/api/devices/' + encodeURIComponent(activeDeviceId) + '/control/home-position', 'POST', {
    channelId: chId,
    enabled: enabled,
    presetIndex: preset,
    resetSec: resetSec
  });
  if(j && j.ok){
    showToast('💾 通道 ' + chId + ' 守望位配置已保存并生效', 'success');
    if(j.ptz){
      renderHomePositionStatus(j.ptz);
    }
    loadConfigCtrlData(false);
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
