package ui

// 轻量单页控制台，无外部依赖。
const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>GB28181 模拟设备</title>
<style>
:root{
  --bg:#0f1419;--panel:#171d24;--line:#2a3340;--text:#e7edf5;--muted:#8b9bb0;
  --ok:#3ecf8e;--warn:#f0b429;--err:#f07178;--accent:#4da3ff;
}
*{box-sizing:border-box}
body{margin:0;font:14px/1.5 system-ui,-apple-system,"Segoe UI",sans-serif;background:var(--bg);color:var(--text)}
header{display:flex;align-items:center;justify-content:space-between;gap:12px;padding:14px 20px;border-bottom:1px solid var(--line);background:var(--panel);position:sticky;top:0}
header h1{margin:0;font-size:16px;font-weight:600;letter-spacing:.02em}
.badge{display:inline-flex;align-items:center;gap:6px;padding:4px 10px;border-radius:999px;font-size:12px;border:1px solid var(--line)}
.badge .dot{width:8px;height:8px;border-radius:50%;background:var(--muted)}
.badge.on .dot{background:var(--ok);box-shadow:0 0 8px var(--ok)}
.badge.off .dot{background:var(--err)}
main{max-width:1100px;margin:0 auto;padding:20px;display:grid;gap:16px}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(240px,1fr));gap:12px}
.card{background:var(--panel);border:1px solid var(--line);border-radius:12px;padding:14px 16px}
.card h2{margin:0 0 10px;font-size:13px;color:var(--muted);font-weight:600;text-transform:uppercase;letter-spacing:.06em}
.kv{display:grid;grid-template-columns: 88px 1fr;gap:6px 10px;font-size:13px}
.kv dt{color:var(--muted)}
.kv dd{margin:0;font-family:ui-monospace,Consolas,monospace;word-break:break-all}
table{width:100%;border-collapse:collapse;font-size:13px}
th,td{text-align:left;padding:8px 6px;border-bottom:1px solid var(--line)}
th{color:var(--muted);font-weight:500}
tr:last-child td{border-bottom:none}
.actions{display:flex;flex-wrap:wrap;gap:8px}
button{appearance:none;border:1px solid var(--line);background:#1f2833;color:var(--text);border-radius:8px;padding:7px 12px;font:inherit;cursor:pointer}
button:hover{border-color:var(--accent);color:var(--accent)}
button.primary{background:#1a3a5c;border-color:#2d6da3;color:#cfe6ff}
button.danger{border-color:#6b3038;color:#ffb4b8}
#logs{margin:0;max-height:280px;overflow:auto;background:#0b0f14;border-radius:8px;padding:10px;font:12px/1.45 ui-monospace,Consolas,monospace;color:#a8b8c8;white-space:pre-wrap}
.muted{color:var(--muted)}
.tag{display:inline-block;padding:1px 6px;border-radius:4px;background:#243041;font-size:11px}
</style>
</head>
<body>
<header>
  <h1>GB28181 模拟设备控制台</h1>
  <div id="regBadge" class="badge off"><span class="dot"></span><span>检测中…</span></div>
</header>
<main>
  <section class="card">
    <h2>操作</h2>
    <div class="actions">
      <button class="primary" onclick="post('/api/register')">立即注册</button>
      <button onclick="post('/api/keepalive')">发送心跳</button>
      <button onclick="post('/api/alarm')">发送报警</button>
      <button onclick="refresh()">刷新状态</button>
    </div>
    <p class="muted" style="margin:10px 0 0">点播请在 WVP 页面对通道操作；此页用于查看状态与维护信令。</p>
  </section>

  <section class="grid">
    <div class="card">
      <h2>设备</h2>
      <dl class="kv" id="devInfo"></dl>
    </div>
    <div class="card">
      <h2>信令 / 媒体</h2>
      <dl class="kv" id="sipInfo"></dl>
    </div>
  </section>

  <section class="card">
    <h2>通道</h2>
    <table>
      <thead><tr><th>名称</th><th>国标编号</th><th>状态</th><th>媒体源</th><th>视频</th></tr></thead>
      <tbody id="chBody"></tbody>
    </table>
  </section>

  <section class="card">
    <h2>点播会话</h2>
    <table>
      <thead><tr><th>通道</th><th>SSRC</th><th>对端</th><th>源就绪</th><th></th></tr></thead>
      <tbody id="sessBody"></tbody>
    </table>
  </section>

  <section class="card">
    <h2>最近日志</h2>
    <pre id="logs">加载中…</pre>
  </section>
</main>
<script>
async function post(url){
  try{
    const r=await fetch(url,{method:'POST'});
    const j=await r.json();
    if(!r.ok) alert(j.error||'操作失败');
    await refresh();
  }catch(e){ alert(e.message); }
}
function el(id){return document.getElementById(id)}
function kv(node, items){
  node.innerHTML = items.map(([k,v])=>'<dt>'+k+'</dt><dd>'+esc(v)+'</dd>').join('');
}
function esc(s){return String(s??'').replace(/[&<>"]/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c]))}
async function refresh(){
  const st=await fetch('/api/status').then(r=>r.json());
  const b=el('regBadge');
  b.className='badge '+(st.registered?'on':'off');
  b.innerHTML='<span class="dot"></span><span>'+(st.registered?'已注册':'未注册')+'</span>';
  kv(el('devInfo'),[
    ['设备号', st.deviceId],
    ['名称', st.deviceName],
    ['运行', Math.floor(st.uptimeSec/60)+' 分 '+(st.uptimeSec%60)+' 秒'],
  ]);
  kv(el('sipInfo'),[
    ['平台', st.server],
    ['本地', st.local],
    ['传输', st.transport],
    ['媒体模式', st.mediaMode],
    ['默认源', st.mediaSource],
  ]);
  el('chBody').innerHTML=(st.channels||[]).map(c=>
    '<tr><td>'+esc(c.name)+'</td><td class="muted">'+esc(c.id)+'</td><td><span class="tag">'+esc(c.status)+'</span></td><td>'+esc(c.source)+'</td><td class="muted">'+esc(c.mp4||'—')+'</td></tr>'
  ).join('') || '<tr><td colspan="5" class="muted">无通道</td></tr>';
  el('sessBody').innerHTML=(st.sessions||[]).map(s=>
    '<tr><td>'+esc(s.channelId)+'</td><td class="muted">'+esc(s.ssrc)+'</td><td>'+esc(s.remoteIp)+':'+s.remotePort+(s.tcp?' TCP':' UDP')+'</td><td>'+(s.sourceReady?'是':'抽流中')+'</td><td><button class="danger" onclick="stopSess(\''+esc(s.callId)+'\')">停止</button></td></tr>'
  ).join('') || '<tr><td colspan="5" class="muted">当前无点播</td></tr>';
  const logs=await fetch('/api/logs?n=150').then(r=>r.json());
  el('logs').textContent=(logs.lines||[]).join('\n')||'（暂无日志）';
  el('logs').scrollTop=el('logs').scrollHeight;
}
async function stopSess(id){ await post('/api/session/stop?callId='+encodeURIComponent(id)); }
refresh();
setInterval(refresh, 3000);
</script>
</body>
</html>
`
