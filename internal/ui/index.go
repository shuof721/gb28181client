package ui

const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>GB28181 模拟设备管理工作台</title>
<style>
:root{
  --bg-dark:#090d14;
  --bg:#0f1522;
  --surface:#151d2d;
  --surface-hover:#1c273c;
  --surface-2:#1e293b;
  --surface-3:#29374e;
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

/* Scrollbar */
::-webkit-scrollbar{width:8px;height:8px}
::-webkit-scrollbar-track{background:rgba(0,0,0,0.15)}
::-webkit-scrollbar-thumb{background:var(--surface-3);border-radius:4px}
::-webkit-scrollbar-thumb:hover{background:rgba(255,255,255,0.25)}

/* Header */
.header{
  position:sticky;top:0;z-index:40;
  display:flex;align-items:center;justify-content:space-between;gap:16px;
  padding:0 28px;height:64px;
  background:rgba(15,21,34,0.88);
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

/* Pills & Badges */
.badge{
  display:inline-flex;align-items:center;gap:6px;
  padding:4px 10px;border-radius:999px;
  font-size:12px;font-weight:500;
  background:var(--surface-2);border:1px solid var(--border);
  color:var(--text-muted);
}
.badge .dot{width:8px;height:8px;border-radius:50%;background:var(--text-dim)}
.badge.on{
  background:rgba(16,185,129,0.12);border-color:rgba(16,185,129,0.3);color:#6ee7b7;
}
.badge.on .dot{
  background:var(--ok);box-shadow:0 0 8px var(--ok);animation:pulse 2s infinite;
}
.badge.off{
  background:rgba(244,63,94,0.1);border-color:rgba(244,63,94,0.25);color:#fda4af;
}
.badge.off .dot{background:var(--err)}
.badge.live{
  background:rgba(16,185,129,0.15);border-color:rgba(16,185,129,0.35);color:#a7f3d0;
}
.badge.live .dot{background:var(--ok);box-shadow:0 0 6px var(--ok)}
.badge.playback{
  background:rgba(99,102,241,0.12);border-color:rgba(99,102,241,0.3);color:#a5b4fc;
}
.badge.playback .dot{background:#818cf8;box-shadow:0 0 6px #818cf8}
.badge.download{
  background:rgba(6,182,212,0.12);border-color:rgba(6,182,212,0.3);color:#67e8f9;
}
.badge.download .dot{background:#22d3ee;box-shadow:0 0 6px #22d3ee}
.badge.paused{
  background:rgba(245,158,11,0.12);border-color:rgba(245,158,11,0.3);color:#fcd34d;
}
.badge.paused .dot{background:#fbbf24}

@keyframes pulse{
  0%,100%{opacity:1;transform:scale(1)}
  50%{opacity:0.5;transform:scale(0.85)}
}

/* Layout */
.container{
  max-width:1380px;margin:0 auto;padding:24px 24px 60px;
  display:flex;flex-direction:column;gap:20px;
}

/* Common Buttons */
.btn{
  display:inline-flex;align-items:center;justify-content:center;gap:6px;
  background:var(--surface-2);color:var(--text-main);
  border:1px solid var(--border);border-radius:var(--radius-sm);
  padding:7px 14px;font-size:12px;font-weight:500;cursor:pointer;
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
.btn-sm{padding:4px 9px;font-size:11px}
.btn-icon{padding:6px;width:30px;height:30px}

/* Inputs & Form Elements */
input[type=text],select{
  background:var(--surface);color:var(--text-main);
  border:1px solid var(--border);border-radius:var(--radius-sm);
  padding:7px 12px;font-size:12px;outline:none;
  transition:border-color 0.15s ease,box-shadow 0.15s ease;
}
input[type=text]:focus,select:focus{
  border-color:var(--accent);box-shadow:0 0 0 3px var(--accent-glow);
}
input[type=text]{width:100%}
select{cursor:pointer}

/* Stats Cards */
.stats-grid{
  display:grid;grid-template-columns:repeat(6,1fr);gap:14px;
}
@media (max-width:1180px){.stats-grid{grid-template-columns:repeat(3,1fr)}}
@media (max-width:680px){.stats-grid{grid-template-columns:repeat(2,1fr)}}
.stat-card{
  background:var(--surface);border:1px solid var(--border);
  border-radius:var(--radius);padding:14px 16px;
  display:flex;flex-direction:column;gap:6px;position:relative;overflow:hidden;
  transition:border-color 0.2s ease,transform 0.2s ease;
}
.stat-card:hover{border-color:rgba(255,255,255,0.15);transform:translateY(-1px)}
.stat-label{
  font-size:11px;font-weight:600;text-transform:uppercase;letter-spacing:0.04em;
  color:var(--text-dim);display:flex;align-items:center;justify-content:space-between;
}
.stat-val{
  font-size:18px;font-weight:700;color:#fff;
  font-family:var(--font-mono);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;
}
.stat-val.sm{font-size:13px;font-weight:500;color:var(--text-muted)}
.stat-sub{font-size:11px;color:var(--text-dim)}

/* Top Quick Actions Bar */
.action-bar{
  background:var(--surface);border:1px solid var(--border);border-radius:var(--radius);
  padding:14px 18px;display:flex;align-items:center;justify-content:space-between;
  flex-wrap:wrap;gap:12px;
}
.action-group{display:flex;align-items:center;gap:8px;flex-wrap:wrap}

/* Panels */
.panel{
  background:var(--surface);border:1px solid var(--border);
  border-radius:var(--radius-lg);box-shadow:var(--shadow-sm);
  overflow:hidden;display:flex;flex-direction:column;
}
.panel-head{
  padding:14px 20px;background:rgba(255,255,255,0.02);
  border-bottom:1px solid var(--border);
  display:flex;align-items:center;justify-content:space-between;gap:12px;
}
.panel-head h2{
  margin:0;font-size:13px;font-weight:700;letter-spacing:0.04em;
  text-transform:uppercase;color:var(--text-main);
  display:flex;align-items:center;gap:8px;
}
.panel-body{padding:18px 20px}
.panel-body.tight{padding:0}

/* Two-column layout */
.main-grid{
  display:grid;grid-template-columns:1.5fr 1fr;gap:20px;align-items:start;
}
@media (max-width:1080px){.main-grid{grid-template-columns:1fr}}

/* Channel Cards Grid */
.channel-grid{
  display:grid;grid-template-columns:repeat(auto-fill,minmax(310px,1fr));gap:14px;
  padding:18px 20px;
}
.ch-card{
  background:var(--surface-2);border:1px solid var(--border);
  border-radius:var(--radius);padding:14px 16px;
  display:flex;flex-direction:column;gap:12px;position:relative;
  transition:all 0.2s ease;
}
.ch-card:hover{border-color:rgba(255,255,255,0.18)}
.ch-card.live{
  border-color:rgba(16,185,129,0.45);
  box-shadow:0 0 16px rgba(16,185,129,0.12);
}
.ch-card.live::before{
  content:"";position:absolute;left:0;top:0;bottom:0;width:3px;
  background:var(--ok);border-radius:var(--radius) 0 0 var(--radius);
}
.ch-header{display:flex;align-items:flex-start;justify-content:space-between;gap:10px}
.ch-title{min-width:0;flex:1}
.ch-name{font-size:14px;font-weight:700;color:#fff;display:flex;align-items:center;gap:6px}
.ch-id{
  font-family:var(--font-mono);font-size:11px;color:var(--text-dim);
  margin-top:2px;display:flex;align-items:center;gap:6px;
}
.copy-btn{
  background:none;border:none;padding:0;color:var(--text-dim);cursor:pointer;
  display:inline-flex;align-items:center;transition:color 0.15s ease;
}
.copy-btn:hover{color:var(--text-main)}
.ch-meta{
  display:grid;grid-template-columns:60px 1fr;gap:6px 10px;
  font-size:12px;background:rgba(0,0,0,0.16);padding:8px 10px;border-radius:var(--radius-sm);
}
.ch-meta span{color:var(--text-dim)}
.ch-meta b{color:var(--text-main);font-weight:500;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.ch-live-bar{
  background:rgba(16,185,129,0.08);border:1px solid rgba(16,185,129,0.25);
  border-radius:var(--radius-sm);padding:8px 10px;
  display:flex;align-items:center;justify-content:space-between;gap:8px;
}
.ch-actions{
  display:flex;align-items:center;justify-content:space-between;gap:8px;
  padding-top:6px;border-top:1px dashed var(--border);
}

/* Tables */
.table-wrap{width:100%;overflow-x:auto}
table.custom-tbl{width:100%;border-collapse:collapse;font-size:12px;text-align:left}
table.custom-tbl th{
  background:rgba(0,0,0,0.25);color:var(--text-dim);
  font-size:11px;font-weight:600;text-transform:uppercase;letter-spacing:0.04em;
  padding:10px 16px;border-bottom:1px solid var(--border);
}
table.custom-tbl td{
  padding:10px 16px;border-bottom:1px solid var(--border);
  color:var(--text-main);vertical-align:middle;
}
table.custom-tbl tr:last-child td{border-bottom:none}
table.custom-tbl tr:hover td{background:rgba(255,255,255,0.02)}
.empty-msg{padding:28px 16px;text-align:center;color:var(--text-dim);font-size:12px}

/* Log Console */
.log-panel{background:var(--bg-dark);border-radius:0 0 var(--radius-lg) var(--radius-lg)}
.log-toolbar{
  padding:12px 18px;background:var(--surface);
  border-bottom:1px solid var(--border);
  display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:10px;
}
.log-tags{display:flex;align-items:center;gap:6px}
.log-tag-btn{
  font-size:11px;padding:3px 8px;border-radius:4px;
  border:1px solid var(--border);background:transparent;
  color:var(--text-muted);cursor:pointer;
}
.log-tag-btn.active{
  background:var(--surface-3);border-color:var(--accent);color:#fff;
}
.log-box{
  margin:0;height:340px;overflow-y:auto;
  padding:12px 16px;font-family:var(--font-mono);font-size:11.5px;line-height:1.6;
  color:#cbd5e1;background:#080c12;white-space:pre-wrap;word-break:break-all;
}
.log-line{display:flex;gap:8px;padding:1px 0}
.log-ts{color:#475569;flex-shrink:0}
.log-tag{
  padding:0 5px;border-radius:3px;font-size:10px;font-weight:600;
  display:inline-block;flex-shrink:0;
}
.log-tag.sip{background:rgba(59,130,246,0.25);color:#93c5fd}
.log-tag.media{background:rgba(168,85,247,0.25);color:#d8b4fe}
.log-tag.gb{background:rgba(16,185,129,0.25);color:#6ee7b7}
.log-tag.device{background:rgba(245,158,11,0.25);color:#fde68a}
.log-tag.ui{background:rgba(6,182,212,0.25);color:#a5f3fc}
.log-tag.err{background:rgba(244,63,94,0.3);color:#fca5a5}

/* Video Items */
.video-card-list{display:flex;flex-direction:column;gap:8px;max-height:260px;overflow-y:auto}
.video-card-item{
  display:flex;align-items:center;justify-content:space-between;gap:12px;
  background:var(--surface-2);border:1px solid var(--border);border-radius:var(--radius-sm);
  padding:8px 12px;transition:border-color 0.15s ease;
}
.video-card-item:hover{border-color:rgba(255,255,255,0.15)}
.video-file-info{min-width:0;flex:1;display:flex;align-items:center;gap:10px}
.video-format-pill{
  font-size:10px;font-weight:700;padding:2px 5px;border-radius:4px;
  text-transform:uppercase;background:rgba(59,130,246,0.18);color:#93c5fd;
}
.video-format-pill.h264{background:rgba(168,85,247,0.18);color:#d8b4fe}

/* Modals */
.modal-overlay{
  position:fixed;inset:0;z-index:90;
  background:rgba(0,0,0,0.72);backdrop-filter:blur(6px);
  display:none;align-items:center;justify-content:center;padding:20px;
}
.modal-overlay.open{display:flex}
.modal-box{
  background:var(--surface);border:1px solid var(--border);border-radius:var(--radius-lg);
  box-shadow:var(--shadow-lg);width:100%;max-width:520px;overflow:hidden;
  animation:modalFade 0.18s ease-out;
}
@keyframes modalFade{
  from{opacity:0;transform:scale(0.96)}
  to{opacity:1;transform:scale(1)}
}
.modal-head{
  padding:16px 20px;background:rgba(255,255,255,0.03);border-bottom:1px solid var(--border);
  display:flex;align-items:center;justify-content:space-between;
}
.modal-head h3{margin:0;font-size:14px;font-weight:700;color:#fff}
.modal-close{background:none;border:none;color:var(--text-muted);cursor:pointer;font-size:18px}
.modal-close:hover{color:#fff}
.modal-body{padding:20px}
.modal-foot{
  padding:14px 20px;background:rgba(0,0,0,0.18);border-top:1px solid var(--border);
  display:flex;justify-content:flex-end;gap:10px;
}
.form-group{margin-bottom:14px}
.form-group label{display:block;font-size:12px;font-weight:600;color:var(--text-muted);margin-bottom:6px}
.form-group .hint{font-size:11px;color:var(--text-dim);margin-top:4px}

/* Toast */
.toast-container{
  position:fixed;top:20px;right:20px;z-index:100;
  display:flex;flex-direction:column;gap:8px;pointer-events:none;
}
.toast{
  pointer-events:auto;min-width:260px;max-width:380px;
  background:var(--surface-2);border:1px solid var(--border);border-radius:var(--radius);
  box-shadow:var(--shadow);padding:10px 14px;display:flex;align-items:center;gap:10px;
  animation:toastSlide 0.2s ease-out;font-size:12px;color:var(--text-main);
}
@keyframes toastSlide{
  from{transform:translateX(50px);opacity:0}
  to{transform:translateX(0);opacity:1}
}
.toast.success{border-color:rgba(16,185,129,0.4);background:#0d261e;color:#6ee7b7}
.toast.error{border-color:rgba(244,63,94,0.4);background:#2c1216;color:#fca5a5}
.toast.info{border-color:rgba(59,130,246,0.4);background:#10223f;color:#93c5fd}

/* Video Player Modal Special Size */
.modal-box.video-modal{max-width:800px}
.video-preview-player{
  width:100%;height:auto;max-height:480px;background:#000;border-radius:var(--radius-sm);
  outline:none;display:block;
}
</style>
</head>
<body>

<!-- Header -->
<header class="header">
  <div class="brand">
    <div class="brand-icon">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M23 7l-7 5 7 5V7z"/><rect x="1" y="5" width="15" height="14" rx="2" ry="2"/>
      </svg>
    </div>
    <div class="brand-info">
      <h1>GB28181 模拟设备工作台</h1>
      <span class="ver">NVR 多通道 · PS/RTP 实时媒体 · 2016 规范</span>
    </div>
  </div>

  <div class="header-actions">
    <div id="modeBadge" class="badge">媒体模式 —</div>
    <div id="liveBadge" class="badge"><span class="dot"></span><span>0 路点播中</span></div>
    <div id="regBadge" class="badge off"><span class="dot"></span><span>检测中</span></div>
  </div>
</header>

<div class="container">
  <!-- Metrics KPI Banner -->
  <section class="stats-grid" id="statsGrid"></section>

  <!-- Quick Actions Bar -->
  <section class="action-bar">
    <div class="action-group">
      <button class="btn btn-primary" onclick="reqRegister()">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
        立即注册
      </button>
      <button class="btn btn-danger" onclick="reqUnregister()">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="4.93" y1="4.93" x2="19.07" y2="19.07"/></svg>
        注销
      </button>
      <button class="btn" onclick="reqKeepalive()">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"/></svg>
        发送心跳
      </button>
      <button class="btn" onclick="openAlarmModal()">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>
        模拟报警上报
      </button>
      <button class="btn" onclick="refreshAll(true)">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M23 4v6h-6"/><path d="M1 20v-6h6"/><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/></svg>
        刷新
      </button>
    </div>
    <div style="font-size:12px;color:var(--text-dim)">
      💡 提示：在 WVP / 国标平台点击通道播放发起 INVITE，设备将向流媒体服务器推送 PS/RTP 流。
    </div>
  </section>

  <!-- Main 2-Column: Channels & Side Tools -->
  <div class="main-grid">
    <!-- Left Column: Channels -->
    <div class="panel">
      <div class="panel-head">
        <h2>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"/><line x1="7" y1="2" x2="7" y2="22"/><line x1="17" y1="2" x2="17" y2="22"/><line x1="2" y1="12" x2="22" y2="12"/></svg>
          通道列表 (<span id="chCount">0</span>)
        </h2>
        <div class="action-group">
          <button class="btn btn-sm btn-primary" onclick="openAddChModal()">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            新增通道
          </button>
        </div>
      </div>
      <div class="channel-grid" id="chGrid">
        <div class="empty-msg">加载中…</div>
      </div>
    </div>

    <!-- Right Column: Media Mode & Video Library -->
    <div style="display:flex;flex-direction:column;gap:20px">
      <!-- Media Mode Card -->
      <div class="panel">
        <div class="panel-head">
          <h2>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M16 12l-4-4-4 4M12 8v8"/></svg>
            推流模式设置
          </h2>
        </div>
        <div class="panel-body" style="display:flex;flex-direction:column;gap:12px">
          <div style="display:flex;gap:10px">
            <button class="btn" id="btnModeShared" onclick="setMediaMode('shared')" style="flex:1">全通道共用媒体</button>
            <button class="btn" id="btnModePer" onclick="setMediaMode('per_channel')" style="flex:1">按通道独立绑定</button>
          </div>
          <div style="font-size:11.5px;color:var(--text-dim);line-height:1.5">
            <b>全通道共用</b>：所有通道点播时均使用全局配置的默认视频。<br/>
            <b>按通道独立</b>：各通道可自由绑定不同视频文件，互不干扰。卡片内绑定视频后会自动切换到此模式。
          </div>
        </div>
      </div>

      <!-- Video Assets Library -->
      <div class="panel">
        <div class="panel-head">
          <h2>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="23 7 16 12 23 17 23 7"/><rect x="1" y="5" width="15" height="14" rx="2" ry="2"/></svg>
            视频资产库
          </h2>
        </div>
        <div class="panel-body" style="display:flex;flex-direction:column;gap:14px">
          <div style="display:flex;gap:8px">
            <input type="file" id="uploadInput" accept=".mp4,.h264,.264" style="display:none" onchange="handleFileSelected()"/>
            <input type="text" id="uploadFileName" placeholder="选择 .mp4 / .h264 视频文件" readonly onclick="document.getElementById('uploadInput').click()" style="cursor:pointer"/>
            <button class="btn btn-primary" id="uploadBtn" onclick="uploadVideoFile()">上传</button>
          </div>
          <div id="videoListContainer" class="video-card-list">
            <div class="empty-msg">暂无视频文件</div>
          </div>
          <div style="font-size:11px;color:var(--text-dim)">
            📁 文件将保存至工作目录 <code>assets/</code>；首次点播 MP4 会自动调用 ffmpeg 抽取无 B 帧 Baseline H.264 缓存。
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Live Sessions Stream Table -->
  <section class="panel">
    <div class="panel-head">
      <h2>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg>
        点播会话与流媒体推流监控
      </h2>
      <span style="font-size:11px;color:var(--text-dim)" id="sessStatsSummary">0 个活动会话</span>
    </div>
    <div class="table-wrap">
      <table class="custom-tbl">
        <thead>
          <tr>
            <th>通道编号</th>
            <th>类型</th>
            <th>SSRC (国标点播)</th>
            <th>对端媒体接收地址</th>
            <th>传输模式</th>
            <th>推流时长</th>
            <th>发包总量</th>
            <th>实时码率</th>
            <th>状态</th>
            <th style="width:80px;text-align:center">操作</th>
          </tr>
        </thead>
        <tbody id="sessTbody">
          <tr><td colspan="10"><div class="empty-msg">当前没有活跃的点播推流会话</div></td></tr>
        </tbody>
      </table>
    </div>
  </section>

  <!-- Log Console -->
  <section class="panel">
    <div class="panel-head">
      <h2>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>
        运行与信令日志
      </h2>
      <div class="action-group">
        <label style="font-size:11px;color:var(--text-dim);display:flex;align-items:center;gap:4px;cursor:pointer">
          <input type="checkbox" id="autoScroll" checked/> 自动滚动
        </label>
        <button class="btn btn-sm" onclick="exportLogFile()">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          导出日志
        </button>
      </div>
    </div>
    <div class="log-panel">
      <div class="log-toolbar">
        <div class="log-tags">
          <button class="log-tag-btn active" onclick="setLogCategory('all')">全部</button>
          <button class="log-tag-btn" onclick="setLogCategory('sip')">SIP 信令</button>
          <button class="log-tag-btn" onclick="setLogCategory('media')">流媒体</button>
          <button class="log-tag-btn" onclick="setLogCategory('gb')">GB28181</button>
          <button class="log-tag-btn" onclick="setLogCategory('err')">错误异常</button>
        </div>
        <div style="display:flex;align-items:center;gap:6px;width:min(320px,100%)">
          <input type="text" id="logKeyword" placeholder="回车筛选日志，如 INVITE / ACK / error" onkeydown="if(event.key==='Enter')loadLogs()"/>
          <button class="btn btn-sm" onclick="clearLogSearch()">清空</button>
        </div>
      </div>
      <div id="logBox" class="log-box">加载日志中…</div>
    </div>
  </section>
</div>

<!-- Video Preview Modal -->
<div class="modal-overlay" id="videoModal">
  <div class="modal-box video-modal">
    <div class="modal-head">
      <h3 id="videoModalTitle">视频在线预览</h3>
      <button class="modal-close" onclick="closeVideoModal()">&times;</button>
    </div>
    <div class="modal-body" style="padding:14px">
      <video id="previewPlayer" class="video-preview-player" controls autoplay loop playsinline></video>
    </div>
    <div class="modal-foot">
      <button class="btn" onclick="closeVideoModal()">关闭</button>
    </div>
  </div>
</div>

<!-- Alarm Modal -->
<div class="modal-overlay" id="alarmModal">
  <div class="modal-box">
    <div class="modal-head">
      <h3>模拟国标报警上报 (Alarm Notify)</h3>
      <button class="modal-close" onclick="closeAlarmModal()">&times;</button>
    </div>
    <div class="modal-body">
      <div class="form-group">
        <label>目标通道</label>
        <select id="alarmChannelSelect" style="width:100%"></select>
      </div>
      <div class="form-group">
        <label>报警方式 (AlarmMethod)</label>
        <select id="alarmMethodSelect" style="width:100%">
          <option value="2">2 - 运动目标检测报警 (移动侦测)</option>
          <option value="1">1 - 人工视频报警</option>
          <option value="3">3 - 遗留物检测报警</option>
          <option value="4">4 - 物体移除检测报警</option>
          <option value="5">5 - 绊线入侵检测报警</option>
          <option value="6">6 - 区域入侵检测报警</option>
        </select>
      </div>
      <div class="form-group">
        <label>报警级别 (AlarmPriority)</label>
        <select id="alarmPrioritySelect" style="width:100%">
          <option value="4">4 - 四级 (低)</option>
          <option value="3">3 - 三级 (中)</option>
          <option value="2">2 - 二级 (高)</option>
          <option value="1">1 - 一级 (紧急)</option>
        </select>
      </div>
      <div class="form-group">
        <label>报警描述 (AlarmDescription)</label>
        <input type="text" id="alarmDescInput" value="检测到运动目标异常触发"/>
      </div>
    </div>
    <div class="modal-foot">
      <button class="btn" onclick="closeAlarmModal()">取消</button>
      <button class="btn btn-primary" onclick="submitAlarm()">立即上报平台</button>
    </div>
  </div>
</div>

<!-- Add Channel Modal -->
<div class="modal-overlay" id="addChModal">
  <div class="modal-box">
    <div class="modal-head">
      <h3>新增模拟通道 (IPC)</h3>
      <button class="modal-close" onclick="closeAddChModal()">&times;</button>
    </div>
    <div class="modal-body">
      <div class="form-group">
        <label>通道国标编号 (20位)</label>
        <input type="text" id="addChId" placeholder="例如 34020000001320000003" maxlength="20"/>
        <div class="hint">前10位为行业编码，11-13位类型通常为 132(网络摄像机) 或 131。</div>
      </div>
      <div class="form-group">
        <label>通道名称</label>
        <input type="text" id="addChName" placeholder="例如：东门高点全景枪机"/>
      </div>
      <div class="form-group">
        <label>绑定初始视频</label>
        <select id="addChVideo" style="width:100%"><option value="">暂不绑定（使用全局默认）</option></select>
      </div>
    </div>
    <div class="modal-foot">
      <button class="btn" onclick="closeAddChModal()">取消</button>
      <button class="btn btn-primary" onclick="submitAddChannel()">确认创建</button>
    </div>
  </div>
</div>

<!-- Edit Channel Modal -->
<div class="modal-overlay" id="editChModal">
  <div class="modal-box">
    <div class="modal-head">
      <h3>编辑通道</h3>
      <button class="modal-close" onclick="closeEditChModal()">&times;</button>
    </div>
    <div class="modal-body">
      <div class="form-group">
        <label>通道国标编号</label>
        <input type="text" id="editChId" readonly style="opacity:0.6;cursor:not-allowed"/>
      </div>
      <div class="form-group">
        <label>通道名称</label>
        <input type="text" id="editChName" placeholder="通道名称"/>
      </div>
      <div class="form-group">
        <label>模拟在线状态</label>
        <select id="editChStatus" style="width:100%">
          <option value="ON">ON - 在线正常</option>
          <option value="OFF">OFF - 离线故障</option>
        </select>
        <div class="hint">切换为 OFF 将在目录通知中标记离线，并自动停止当前活动点播。</div>
      </div>
    </div>
    <div class="modal-foot">
      <button class="btn" onclick="closeEditChModal()">取消</button>
      <button class="btn btn-primary" onclick="submitEditChannel()">保存更新</button>
    </div>
  </div>
</div>

<!-- Confirm Dialog Modal -->
<div class="modal-overlay" id="confirmModal">
  <div class="modal-box" style="max-width:420px">
    <div class="modal-head">
      <h3 id="confirmModalTitle">操作确认</h3>
      <button class="modal-close" onclick="closeConfirmModal()">&times;</button>
    </div>
    <div class="modal-body" id="confirmModalMessage" style="font-size:13px;color:var(--text-main);line-height:1.6">
      确认执行此操作？
    </div>
    <div class="modal-foot">
      <button class="btn" onclick="closeConfirmModal()">取消</button>
      <button class="btn btn-danger" id="confirmModalOkBtn">确认</button>
    </div>
  </div>
</div>

<!-- Toast Container -->
<div class="toast-container" id="toastContainer"></div>

<script>
// State
let globalStatus = {};
let videoList = [];
let logCategory = 'all';
let rawLogs = [];

// Utils
function esc(s){return String(s||'').replace(/[&<>"]/g,function(c){return {'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c]})}
function fmtSize(b){if(b<1024)return b+' B';if(b<1048576)return(b/1024).toFixed(1)+' KB';return(b/1048576).toFixed(1)+' MB'}
function fmtDuration(s){
  if(s<=0)return '00:00';
  const m=Math.floor(s/60),sec=s%60,h=Math.floor(m/60);
  const pad=function(n){return String(n).padStart(2,'0')};
  return h>0 ? (pad(h)+':'+pad(m%60)+':'+pad(sec)) : (pad(m)+':'+pad(sec));
}
function fmtBitrate(kbps){
  if(!kbps||kbps<=0) return '0 kbps';
  if(kbps>=1000) return (kbps/1000).toFixed(2)+' Mbps';
  return Math.round(kbps)+' kbps';
}
function fmtUptime(s){
  const m=Math.floor(s/60),sec=s%60,h=Math.floor(m/60);
  return h ? (h+'小时 '+(m%60)+'分') : (m+'分 '+sec+'秒');
}

// Toast
function showToast(msg, type){
  type = type || 'info';
  const c = document.getElementById('toastContainer');
  const t = document.createElement('div');
  t.className = 'toast '+type;
  t.innerHTML = '<span>'+esc(msg)+'</span>';
  c.appendChild(t);
  setTimeout(function(){
    t.style.opacity='0';
    t.style.transition='opacity 0.25s ease';
    setTimeout(function(){t.remove()}, 250);
  }, 3200);
}

// Custom Confirm Modal
let currentConfirmCb = null;
function showConfirm(title, msg, onConfirm){
  document.getElementById('confirmModalTitle').textContent = title;
  document.getElementById('confirmModalMessage').innerHTML = msg;
  currentConfirmCb = onConfirm;
  document.getElementById('confirmModal').classList.add('open');
}
function closeConfirmModal(){
  document.getElementById('confirmModal').classList.remove('open');
  currentConfirmCb = null;
}
document.getElementById('confirmModalOkBtn').onclick = function(){
  if(currentConfirmCb) currentConfirmCb();
  closeConfirmModal();
};

// Clipboard
function copyText(text, label){
  navigator.clipboard.writeText(text).then(function(){
    showToast('已复制'+(label?' '+label:'')+'到剪贴板', 'success');
  }).catch(function(){
    showToast('复制失败，请手动选择复制', 'error');
  });
}

// API helper
async function postAPI(url, body){
  const opt = {method:'POST'};
  if(body !== undefined){
    opt.headers = {'Content-Type':'application/json'};
    opt.body = JSON.stringify(body);
  }
  try{
    const r = await fetch(url, opt);
    const j = await r.json().catch(function(){return {}});
    if(!r.ok){
      showToast(j.error||('请求失败: HTTP '+r.status), 'error');
      return null;
    }
    return j;
  }catch(e){
    showToast(e.message||'网络错误', 'error');
    return null;
  }
}

// Render Stats KPI Cards
function renderStats(st){
  const sessCount = (st.sessions||[]).length;
  const chs = st.channels||[];
  const onlineCount = chs.filter(function(c){return c.status!=='OFF'}).length;

  const items = [
    {label:'设备编号', val:st.deviceId, copy:true, sub:st.deviceName||'模拟NVR'},
    {label:'平台 SIP 服务', val:st.server, sub:'传输: '+(st.transport||'udp').toUpperCase()},
    {label:'本地监听地址', val:st.local, sub:'SIP 端口'},
    {label:'挂载通道', val:chs.length+' 路', sub:'在线: '+onlineCount+' / 离线: '+(chs.length-onlineCount)},
    {label:'活动推流会话', val:sessCount+' 路', sub:'实时 PS/RTP'},
    {label:'已运行时间', val:fmtUptime(st.uptimeSec||0), sub:'连续运行'}
  ];

  document.getElementById('statsGrid').innerHTML = items.map(function(it){
    return '<div class="stat-card">' +
      '<div class="stat-label">' +
        '<span>' + esc(it.label) + '</span>' +
        (it.copy ? ('<button class="copy-btn" title="复制" onclick="copyText(\'' + esc(it.val) + '\',\'编号\')">' +
          '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>' +
        '</button>') : '') +
      '</div>' +
      '<div class="stat-val' + (it.val.length > 15 ? ' sm' : '') + '" title="' + esc(it.val) + '">' + esc(it.val) + '</div>' +
      '<div class="stat-sub">' + esc(it.sub) + '</div>' +
    '</div>';
  }).join('');

  // Top header badges
  const b = document.getElementById('regBadge');
  b.className = 'badge ' + (st.registered ? 'on' : 'off');
  b.innerHTML = '<span class="dot"></span><span>' + (st.registered ? 'SIP 已注册' : 'SIP 未注册') + '</span>';

  const lb = document.getElementById('liveBadge');
  lb.className = 'badge ' + (sessCount > 0 ? 'live' : '');
  lb.innerHTML = '<span class="dot"></span><span>' + sessCount + ' 路点播中</span>';

  const isPer = (st.mediaMode||'').toLowerCase() === 'per_channel';
  document.getElementById('modeBadge').textContent = isPer ? '模式: 按通道独立' : '模式: 全通道共用';
  document.getElementById('btnModeShared').className = isPer ? 'btn' : 'btn btn-primary';
  document.getElementById('btnModePer').className = isPer ? 'btn btn-primary' : 'btn';
}

// Generate options for video binding selector
function makeVideoOptions(currentVal){
  let opts = '<option value="">-- 全局默认视频 --</option>';
  for(const v of videoList){
    const sel = (currentVal && (currentVal.endsWith(v.name) || currentVal === v.path)) ? ' selected' : '';
    opts += '<option value="' + esc(v.path) + '"' + sel + '>' + esc(v.name) + ' (' + fmtSize(v.size) + ')</option>';
  }
  const synSel = currentVal === '__synthetic__' ? ' selected' : '';
  opts += '<option value="__synthetic__"' + synSel + '>[内置测试图案 I_PCM]</option>';
  return opts;
}

// Render Channel Cards
function renderChannels(st, sessMap){
  const grid = document.getElementById('chGrid');
  const chs = st.channels || [];
  document.getElementById('chCount').textContent = chs.length;

  if(!chs.length){
    grid.innerHTML = '<div class="empty-msg" style="grid-column:1/-1">暂无通道，点击上方「新增通道」进行创建</div>';
    return;
  }

  grid.innerHTML = chs.map(function(c){
    const sess = sessMap[c.id];
    const isLive = !!sess;
    const isOffline = c.status === 'OFF';
    let mediaDesc = c.mp4 || c.h264 || '';
    if(!mediaDesc) mediaDesc = c.source === 'synthetic' ? '内置合成测试帧' : (c.source || '全局默认');

    const isMp4 = (c.mp4 || '').toLowerCase().endsWith('.mp4') || (mediaDesc || '').toLowerCase().endsWith('.mp4');

    return '<div class="ch-card' + (isLive ? ' live' : '') + '">' +
      '<div class="ch-header">' +
        '<div class="ch-title">' +
          '<div class="ch-name">' +
            '<span>' + esc(c.name || '未命名通道') + '</span>' +
            '<span class="badge ' + (isLive ? 'live' : (isOffline ? 'off' : 'on')) + '" style="padding:2px 7px;font-size:10.5px">' +
              '<span class="dot"></span><span>' + (isLive ? '推流中' : (isOffline ? '离线' : '在线')) + '</span>' +
            '</span>' +
          '</div>' +
          '<div class="ch-id">' +
            '<span>' + esc(c.id) + '</span>' +
            '<button class="copy-btn" title="复制国标编号" onclick="copyText(\'' + esc(c.id) + '\',\'通道编号\')">' +
              '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>' +
            '</button>' +
          '</div>' +
        '</div>' +
      '</div>' +

      '<div class="ch-meta">' +
        '<span>当前媒体</span>' +
        '<b title="' + esc(mediaDesc) + '">' + esc(mediaDesc) + '</b>' +
        '<span>视频模式</span>' +
        '<b>' + esc(c.source || 'mp4') + '</b>' +
      '</div>' +

      (isLive ? ('<div class="ch-live-bar">' +
        '<div style="font-size:11.5px;color:#6ee7b7">' +
          '<div><b>' + (sess.streamType === 'playback' ? '录像回放' : (sess.streamType === 'download' ? '录像下载' : '实时点播')) + ' | SSRC:</b> ' + esc(sess.ssrc) + '</div>' +
          '<div><b>对端:</b> ' + esc(sess.remoteIp) + ':' + sess.remotePort + ' (' + (sess.tcp?'TCP':'UDP') + ')</div>' +
        '</div>' +
        '<button class="btn btn-danger btn-sm" onclick="stopSession(\'' + esc(sess.callId) + '\')">断开推流</button>' +
      '</div>') : '') +

      '<div style="display:flex;gap:6px;align-items:center">' +
        '<select class="channel-bind-select" data-ch="' + esc(c.id) + '" style="flex:1;min-width:0">' +
          makeVideoOptions(c.mp4 || c.h264) +
        '</select>' +
        '<button class="btn btn-sm btn-primary" onclick="bindChannelVideo(this)">保存绑定</button>' +
        (isMp4 ? ('<button class="btn btn-sm" title="预览视频" onclick="previewVideoPath(\'' + esc(c.mp4 || mediaDesc) + '\')">' +
          '<svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"/></svg>' +
        '</button>') : '') +
      '</div>' +

      '<div class="ch-actions">' +
        '<button class="btn btn-sm" onclick="toggleChannelOnline(\'' + esc(c.id) + '\',\'' + (isOffline ? 'ON' : 'OFF') + '\')">' +
          (isOffline ? '设为在线' : '模拟掉线') +
        '</button>' +
        '<div style="display:flex;gap:6px">' +
          '<button class="btn btn-sm" onclick="openEditChModal(\'' + esc(c.id) + '\',\'' + esc(c.name) + '\',\'' + esc(c.status) + '\')">编辑</button>' +
          '<button class="btn btn-sm btn-danger" onclick="removeChannel(\'' + esc(c.id) + '\')">删除</button>' +
        '</div>' +
      '</div>' +
    '</div>';
  }).join('');
}

// Render Sessions Table
function renderSessions(st){
  const tbody = document.getElementById('sessTbody');
  const rows = st.sessions || [];
  document.getElementById('sessStatsSummary').textContent = rows.length + ' 个活动推流会话';

  if(!rows.length){
    tbody.innerHTML = '<tr><td colspan="10"><div class="empty-msg">当前没有活跃的点播推流会话。在平台（如 WVP）发起实时点播或录像回放即可连通。</div></td></tr>';
    return;
  }

  tbody.innerHTML = rows.map(function(s){
    const isPlayback = s.streamType === 'playback';
    const isDownload = s.streamType === 'download';
    let typeBadge = '<span class="badge live" style="padding:2px 8px"><span class="dot"></span>实时直播</span>';
    if(isPlayback){
      typeBadge = '<span class="badge playback" style="padding:2px 8px"><span class="dot"></span>录像回放</span>';
    } else if(isDownload){
      typeBadge = '<span class="badge download" style="padding:2px 8px"><span class="dot"></span>录像下载</span>';
    }

    let statusHtml = '';
    if(s.paused){
      statusHtml = '<span class="badge paused"><span class="dot"></span>已暂停</span>';
    } else if(s.sourceReady){
      let scaleText = '';
      if(isPlayback && s.scale && s.scale !== 1){
        scaleText = ' (' + s.scale + 'x)';
      }
      statusHtml = '<span class="badge on"><span class="dot"></span><span>' + (isPlayback ? ('回放中' + scaleText) : (isDownload ? '下载中' : '实时推流中')) + '</span></span>';
    } else {
      statusHtml = '<span class="badge"><span class="dot"></span><span>准备抽流</span></span>';
    }

    return '<tr>' +
      '<td style="font-family:var(--font-mono);font-weight:600">' + esc(s.channelId) + '</td>' +
      '<td>' + typeBadge + '</td>' +
      '<td style="font-family:var(--font-mono);color:var(--text-muted)">' + esc(s.ssrc) + '</td>' +
      '<td>' + esc(s.remoteIp) + ':' + s.remotePort + '</td>' +
      '<td><span class="badge" style="padding:2px 6px">' + (s.tcp ? 'TCP' : 'UDP') + '</span></td>' +
      '<td style="font-family:var(--font-mono)">' + fmtDuration(s.durationSec||0) + '</td>' +
      '<td style="font-family:var(--font-mono)">' + (s.packetsSent||0).toLocaleString() + ' 包 (' + fmtSize(s.bytesSent||0) + ')</td>' +
      '<td style="font-family:var(--font-mono);color:#38bdf8">' + fmtBitrate(s.bitrateKbps||0) + '</td>' +
      '<td>' + statusHtml + '</td>' +
      '<td style="text-align:center">' +
        '<button class="btn btn-danger btn-sm" onclick="stopSession(\'' + esc(s.callId) + '\')">断开</button>' +
      '</td>' +
    '</tr>';
  }).join('');
}

// Render Video Library
function renderVideos(){
  const container = document.getElementById('videoListContainer');
  const modalSel = document.getElementById('addChVideo');

  modalSel.innerHTML = '<option value="">暂不绑定（使用全局默认）</option>' +
    videoList.map(function(v){
      return '<option value="' + esc(v.name) + '">' + esc(v.name) + ' (' + fmtSize(v.size) + ')</option>';
    }).join('');

  if(!videoList.length){
    container.innerHTML = '<div class="empty-msg">视频库为空，请选择文件上传</div>';
    return;
  }

  container.innerHTML = videoList.map(function(v){
    const isMp4 = v.name.toLowerCase().endsWith('.mp4');
    return '<div class="video-card-item">' +
      '<div class="video-file-info">' +
        '<span class="video-format-pill ' + (isMp4 ? 'mp4' : 'h264') + '">' + (isMp4 ? 'MP4' : 'H264') + '</span>' +
        '<div style="min-width:0;flex:1">' +
          '<div style="font-weight:600;font-size:12px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis" title="' + esc(v.name) + '">' + esc(v.name) + '</div>' +
          '<div style="font-size:11px;color:var(--text-dim)">' + fmtSize(v.size) + '</div>' +
        '</div>' +
      '</div>' +
      '<div style="display:flex;gap:6px;align-items:center">' +
        (isMp4 ? ('<button class="btn btn-sm btn-primary" onclick="previewVideoPath(\'' + esc(v.path) + '\',\'' + esc(v.name) + '\')">' +
          '<svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"/></svg> 预览' +
        '</button>') : '') +
        '<button class="btn btn-sm btn-danger" onclick="deleteVideoFile(\'' + esc(v.name) + '\')">' +
          '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>' +
        '</button>' +
      '</div>' +
    '</div>';
  }).join('');
}

// Load Video List
async function loadVideos(){
  try{
    const r = await fetch('/api/videos');
    const j = await r.json();
    videoList = j.videos || [];
  }catch(e){videoList = []}
  renderVideos();
}

// Refresh Everything
async function refreshAll(manual){
  try{
    const st = await fetch('/api/status').then(function(r){return r.json()});
    globalStatus = st;
    const sessMap = {};
    for(const s of (st.sessions||[])) sessMap[s.channelId] = s;

    renderStats(st);
    renderChannels(st, sessMap);
    renderSessions(st);
    await loadLogs();

    if(manual) showToast('状态已刷新', 'success');
  }catch(e){
    if(manual) showToast('刷新状态失败: '+e.message, 'error');
  }
}

// Log Processing & Coloring
function parseLogLine(line){
  let tag = 'other';
  let tagClass = '';
  const lower = line.toLowerCase();

  if(lower.includes('[sip]')) { tag = 'SIP'; tagClass = 'sip'; }
  else if(lower.includes('[media]')) { tag = 'MEDIA'; tagClass = 'media'; }
  else if(lower.includes('[gb]')) { tag = 'GB'; tagClass = 'gb'; }
  else if(lower.includes('[device]')) { tag = 'DEVICE'; tagClass = 'device'; }
  else if(lower.includes('[ui]')) { tag = 'UI'; tagClass = 'ui'; }

  const isErr = lower.includes('error') || lower.includes('failed') || lower.includes('fail:');
  if(isErr) tagClass += ' err';

  return {raw:line, tag:tag, tagClass:tagClass, isErr:isErr};
}

async function loadLogs(){
  const q = document.getElementById('logKeyword').value.trim();
  const url = '/api/logs?n=300' + (q ? ('&q=' + encodeURIComponent(q)) : '');
  try{
    const j = await fetch(url).then(function(r){return r.json()});
    rawLogs = j.lines || [];
    renderLogs();
  }catch(e){}
}

function setLogCategory(cat){
  logCategory = cat;
  document.querySelectorAll('.log-tag-btn').forEach(function(b){
    b.classList.toggle('active', b.textContent.toLowerCase().includes(cat) || (cat==='all' && b.textContent==='全部'));
  });
  renderLogs();
}

function renderLogs(){
  const box = document.getElementById('logBox');
  let lines = rawLogs;

  if(logCategory !== 'all'){
    lines = lines.filter(function(l){
      const lower = l.toLowerCase();
      if(logCategory === 'sip') return lower.includes('[sip]');
      if(logCategory === 'media') return lower.includes('[media]');
      if(logCategory === 'gb') return lower.includes('[gb]');
      if(logCategory === 'err') return lower.includes('error') || lower.includes('failed');
      return true;
    });
  }

  if(!lines.length){
    box.innerHTML = '<div style="color:#64748b;padding:8px 0">（暂无符合条件的日志）</div>';
    return;
  }

  box.innerHTML = lines.map(function(line){
    const p = parseLogLine(line);
    const tagHtml = p.tag !== 'other' ? ('<span class="log-tag ' + p.tagClass + '">' + p.tag + '</span>') : '';
    return '<div class="log-line">' + tagHtml + '<span>' + esc(p.raw) + '</span></div>';
  }).join('');

  if(document.getElementById('autoScroll').checked){
    box.scrollTop = box.scrollHeight;
  }
}

function clearLogSearch(){
  document.getElementById('logKeyword').value = '';
  loadLogs();
}

function exportLogFile(){
  const content = rawLogs.join('\n');
  const blob = new Blob([content], {type:'text/plain;charset=utf-8'});
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'gb28181_device_' + (new Date().toISOString().replace(/[:.]/g,'-')) + '.log';
  a.click();
  URL.revokeObjectURL(url);
  showToast('日志已导出下载', 'success');
}

// SIP Actions
async function reqRegister(){
  const j = await postAPI('/api/register');
  if(j){ showToast('已发送注册请求', 'success'); refreshAll(); }
}
async function reqUnregister(){
  const j = await postAPI('/api/unregister');
  if(j){ showToast('已发送注销请求', 'info'); refreshAll(); }
}
async function reqKeepalive(){
  const j = await postAPI('/api/keepalive');
  if(j){ showToast('心跳通知已发送', 'success'); }
}
async function stopSession(callId){
  showConfirm('挂断推流', '确认断开此路实时流推流会话？', async function(){
    const j = await postAPI('/api/session/stop?callId=' + encodeURIComponent(callId));
    if(j){ showToast('已停止推流', 'info'); refreshAll(); }
  });
}

// Channel Actions
function openAddChModal(){
  document.getElementById('addChId').value = '';
  document.getElementById('addChName').value = '';
  document.getElementById('addChModal').classList.add('open');
  document.getElementById('addChId').focus();
}
function closeAddChModal(){
  document.getElementById('addChModal').classList.remove('open');
}
async function submitAddChannel(){
  const id = document.getElementById('addChId').value.trim();
  const name = document.getElementById('addChName').value.trim();
  const mp4 = document.getElementById('addChVideo').value;
  if(id.length !== 20){
    showToast('通道国标编号必须为 20 位数字', 'error');
    return;
  }
  const j = await postAPI('/api/channels/add', {id:id, name:name, mp4:mp4});
  if(j){
    showToast('通道添加成功', 'success');
    closeAddChModal();
    refreshAll();
  }
}

function openEditChModal(id, name, status){
  document.getElementById('editChId').value = id;
  document.getElementById('editChName').value = name;
  document.getElementById('editChStatus').value = status || 'ON';
  document.getElementById('editChModal').classList.add('open');
}
function closeEditChModal(){
  document.getElementById('editChModal').classList.remove('open');
}
async function submitEditChannel(){
  const id = document.getElementById('editChId').value;
  const name = document.getElementById('editChName').value.trim();
  const status = document.getElementById('editChStatus').value;
  const j = await postAPI('/api/channels/update', {id:id, name:name, status:status});
  if(j){
    showToast('通道已更新', 'success');
    closeEditChModal();
    refreshAll();
  }
}

async function toggleChannelOnline(id, nextStatus){
  const j = await postAPI('/api/channels/update', {id:id, status:nextStatus});
  if(j){
    showToast('通道状态已设为 ' + nextStatus, 'info');
    refreshAll();
  }
}

async function bindChannelVideo(btn){
  const sel = btn.parentElement.querySelector('.channel-bind-select');
  const chId = sel.getAttribute('data-ch');
  const val = sel.value;
  let body;
  if(val === '__synthetic__') body = {channelId:chId, source:'synthetic'};
  else if(!val) body = {channelId:chId, source:'mp4', mp4:''};
  else if(val.endsWith('.h264') || val.endsWith('.264')) body = {channelId:chId, source:'file', h264:val};
  else body = {channelId:chId, source:'mp4', mp4:val};

  const j = await postAPI('/api/channels/bind', body);
  if(j){
    showToast('通道媒体绑定已更新', 'success');
    refreshAll();
  }
}

function removeChannel(id){
  showConfirm('删除通道', '确认删除通道 '+id+'？删除后如在 WVP 中使用，需重新拉取设备目录。', async function(){
    const j = await postAPI('/api/channels/remove?id=' + encodeURIComponent(id));
    if(j){ showToast('通道已删除', 'info'); refreshAll(); }
  });
}

async function setMediaMode(mode){
  const j = await postAPI('/api/media/mode', {mode:mode});
  if(j){
    showToast('已切换媒体模式为: ' + (mode==='per_channel'?'按通道独立':'全通道共用'), 'success');
    refreshAll();
  }
}

// Alarm Modal
function openAlarmModal(){
  const sel = document.getElementById('alarmChannelSelect');
  const chs = (globalStatus.channels || []);
  if(chs.length){
    sel.innerHTML = chs.map(function(c){
      return '<option value="' + esc(c.id) + '">' + esc(c.name) + ' (' + esc(c.id) + ')</option>';
    }).join('');
  } else {
    sel.innerHTML = '<option value="' + esc(globalStatus.deviceId) + '">主设备 (' + esc(globalStatus.deviceId) + ')</option>';
  }
  document.getElementById('alarmModal').classList.add('open');
}
function closeAlarmModal(){
  document.getElementById('alarmModal').classList.remove('open');
}
async function submitAlarm(){
  const channelId = document.getElementById('alarmChannelSelect').value;
  const alarmMethod = document.getElementById('alarmMethodSelect').value;
  const priority = document.getElementById('alarmPrioritySelect').value;
  const desc = document.getElementById('alarmDescInput').value.trim();

  const j = await postAPI('/api/alarm', {
    channelId: channelId,
    alarmMethod: alarmMethod,
    priority: priority,
    description: desc
  });
  if(j){
    showToast('模拟报警通知已发送给平台', 'success');
    closeAlarmModal();
  }
}

// Video Preview & Upload
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
  btn.disabled = true; btn.textContent = '上传中…';

  try{
    const r = await fetch('/api/videos/upload', {method:'POST', body:fd});
    const j = await r.json();
    if(!r.ok) showToast(j.error || '上传失败', 'error');
    else{
      showToast('上传成功: ' + j.name, 'success');
      document.getElementById('uploadInput').value = '';
      document.getElementById('uploadFileName').value = '';
      await loadVideos();
      await refreshAll();
    }
  }catch(e){
    showToast('上传网络错误: ' + e.message, 'error');
  }
  btn.disabled = false; btn.textContent = '上传';
}

function deleteVideoFile(name){
  showConfirm('删除视频', '确认删除视频文件 <b>' + esc(name) + '</b>？关联的抽流缓存也会一并清理。', async function(){
    const j = await postAPI('/api/videos/delete?name=' + encodeURIComponent(name));
    if(j){
      showToast('已删除视频 ' + name, 'info');
      await loadVideos();
      await refreshAll();
    }
  });
}

function previewVideoPath(path, name){
  if(!path){ showToast('无可用视频路径', 'error'); return; }
  let filename = name;
  if(!filename){
    filename = path.replace(/\\/g, '/').split('/').pop();
  }
  if(!filename.toLowerCase().endsWith('.mp4')){
    showToast('仅 MP4 文件支持在浏览器直接预览播放', 'info');
    return;
  }
  const streamUrl = '/assets/' + encodeURIComponent(filename);
  const player = document.getElementById('previewPlayer');
  document.getElementById('videoModalTitle').textContent = '在线预览 · ' + filename;
  player.src = streamUrl;
  document.getElementById('videoModal').classList.add('open');
  player.play().catch(function(){});
}
function closeVideoModal(){
  const player = document.getElementById('previewPlayer');
  player.pause();
  player.src = '';
  document.getElementById('videoModal').classList.remove('open');
}

// Initial bootstrap & interval
loadVideos().then(function(){ refreshAll() });
setInterval(refreshAll, 3500);
</script>
</body>
</html>
`
