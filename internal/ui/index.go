package ui

const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>GB28181 模拟设备</title>
<style>
:root{
  --bg:#0b1016;--panel:#141b23;--panel2:#1a2330;--line:#2a3544;
  --text:#e8eef6;--muted:#8fa0b5;--ok:#2fce8a;--warn:#e6b84d;--err:#f07178;
  --accent:#5eb0ff;--accent2:#7c6cff;
}
*{box-sizing:border-box}
body{margin:0;font:14px/1.5 system-ui,-apple-system,"Segoe UI","Microsoft YaHei",sans-serif;background:radial-gradient(1200px 600px at 10% -10%,#152033 0%,var(--bg) 50%);color:var(--text);min-height:100vh}
header{display:flex;align-items:center;justify-content:space-between;gap:12px;padding:14px 22px;border-bottom:1px solid var(--line);backdrop-filter:blur(8px);position:sticky;top:0;background:rgba(11,16,22,.85);z-index:10}
.brand{display:flex;align-items:center;gap:10px}
.logo{width:28px;height:28px;border-radius:8px;background:linear-gradient(135deg,var(--accent),var(--accent2));box-shadow:0 0 18px rgba(94,176,255,.35)}
header h1{margin:0;font-size:15px;font-weight:600}
.badge{display:inline-flex;align-items:center;gap:6px;padding:4px 10px;border-radius:999px;font-size:12px;border:1px solid var(--line);background:var(--panel)}
.badge .dot{width:8px;height:8px;border-radius:50%;background:var(--muted)}
.badge.on .dot{background:var(--ok);box-shadow:0 0 8px var(--ok)}
.badge.off .dot{background:var(--err)}
main{max-width:1200px;margin:0 auto;padding:18px 20px 40px;display:grid;gap:16px}
.stats{display:grid;grid-template-columns:repeat(auto-fit,minmax(160px,1fr));gap:10px}
.stat{background:var(--panel);border:1px solid var(--line);border-radius:14px;padding:12px 14px}
.stat .t{font-size:12px;color:var(--muted)}
.stat .v{font-size:20px;font-weight:600;margin-top:4px;font-variant-numeric:tabular-nums}
.card{background:var(--panel);border:1px solid var(--line);border-radius:14px;padding:14px 16px}
.card h2{margin:0 0 12px;font-size:13px;color:var(--muted);font-weight:600;letter-spacing:.04em;text-transform:uppercase}
.row{display:flex;flex-wrap:wrap;gap:8px;align-items:center}
button,select,input[type=text]{appearance:none;border:1px solid var(--line);background:var(--panel2);color:var(--text);border-radius:8px;padding:7px 12px;font:inherit}
button{cursor:pointer}
button:hover{border-color:var(--accent);color:var(--accent)}
button.primary{background:linear-gradient(180deg,#1d4a78,#173a5e);border-color:#2f6fa8;color:#d7ebff}
button.danger{border-color:#6b3038;color:#ffb4b8}
input[type=text]{min-width:180px}
input[type=file]{color:var(--muted);font-size:12px}
.channels{display:grid;grid-template-columns:repeat(auto-fill,minmax(280px,1fr));gap:12px}
.ch-card{background:linear-gradient(180deg,#182231,#141b23);border:1px solid var(--line);border-radius:14px;padding:14px;display:flex;flex-direction:column;gap:10px}
.ch-card.live{border-color:#2d6a4f;box-shadow:0 0 0 1px rgba(47,206,138,.15),0 8px 24px rgba(0,0,0,.25)}
.ch-head{display:flex;justify-content:space-between;align-items:flex-start;gap:8px}
.ch-name{font-weight:600;font-size:15px}
.ch-id{font:12px ui-monospace,Consolas,monospace;color:var(--muted);margin-top:2px;word-break:break-all}
.ch-meta{display:grid;grid-template-columns:64px 1fr;gap:4px 8px;font-size:12px;color:var(--muted)}
.ch-meta b{color:var(--text);font-weight:500}
.tag{display:inline-flex;align-items:center;padding:2px 8px;border-radius:999px;font-size:11px;background:#243041}
.tag.on{background:#1a3d30;color:#8dffc8}
.tag.live{background:#1a3d5c;color:#9fd0ff}
.bind{display:grid;gap:6px}
.bind select{width:100%}
table{width:100%;border-collapse:collapse;font-size:13px}
th,td{text-align:left;padding:8px 6px;border-bottom:1px solid var(--line)}
th{color:var(--muted);font-weight:500}
.logs{display:flex;flex-direction:column;gap:8px}
.log-tools{display:flex;flex-wrap:wrap;gap:8px;align-items:center}
#logBox{margin:0;height:280px;overflow:auto;background:#0a0e13;border-radius:10px;padding:10px 12px;font:12px/1.5 ui-monospace,Consolas,monospace;color:#9eb0c4;white-space:pre-wrap;border:1px solid #1c2430}
.muted{color:var(--muted)}
.hint{font-size:12px;color:var(--muted);margin:0}
.split{display:grid;grid-template-columns:1.2fr .8fr;gap:16px}
@media (max-width:900px){.split{grid-template-columns:1fr}}
</style>
</head>
<body>
<header>
  <div class="brand"><div class="logo"></div><h1>GB28181 模拟设备控制台</h1></div>
  <div class="row">
    <div id="modeBadge" class="badge"><span>媒体模式：—</span></div>
    <div id="regBadge" class="badge off"><span class="dot"></span><span>检测中…</span></div>
  </div>
</header>
<main>
  <section class="stats" id="stats"></section>

  <section class="card">
    <h2>快捷操作</h2>
    <div class="row">
      <button class="primary" onclick="post('/api/register')">立即注册</button>
      <button onclick="post('/api/keepalive')">发送心跳</button>
      <button onclick="post('/api/alarm')">发送报警</button>
      <button onclick="refresh()">刷新</button>
      <span class="muted">点播请在 WVP 对通道操作；绑定视频后下次点播生效。</span>
    </div>
  </section>

  <section class="split">
    <div class="card">
      <h2>通道卡片</h2>
      <div class="channels" id="chCards"></div>
    </div>
    <div style="display:grid;gap:16px">
      <div class="card">
        <h2>新增通道</h2>
        <div class="row" style="margin-bottom:8px">
          <input id="newChId" type="text" placeholder="20位国标编号" style="flex:1"/>
        </div>
        <div class="row" style="margin-bottom:8px">
          <input id="newChName" type="text" placeholder="通道名称（可选）" style="flex:1"/>
        </div>
        <div class="row">
          <select id="newChVideo" style="flex:1"><option value="">（不绑定视频）</option></select>
          <button class="primary" onclick="addChannel()">添加通道</button>
        </div>
        <p class="hint">编号需 20 位，例如 34020000001320000003。添加后需在 WVP 重新拉目录或刷新设备。</p>
      </div>
      <div class="card">
        <h2>媒体模式</h2>
        <div class="row">
          <button onclick="setMode('shared')">全通道共用</button>
          <button onclick="setMode('per_channel')">按通道绑定</button>
        </div>
        <p class="hint">绑定视频时会自动切到「按通道绑定」。</p>
      </div>
      <div class="card">
        <h2>视频库 / 上传</h2>
        <div class="row" style="margin-bottom:8px">
          <input id="uploadFile" type="file" accept=".mp4,.h264,.264" style="flex:1"/>
          <button class="primary" onclick="uploadVideo()">上传</button>
        </div>
        <div id="videoList" class="muted">加载中…</div>
      </div>
    </div>
  </section>

  <section class="card">
    <h2>点播会话</h2>
    <table>
      <thead><tr><th>通道</th><th>SSRC</th><th>对端</th><th>源</th><th></th></tr></thead>
      <tbody id="sessBody"></tbody>
    </table>
  </section>

  <section class="card logs">
    <h2>日志</h2>
    <div class="log-tools">
      <input id="logQ" type="text" placeholder="过滤关键字，如 media / invite / error" style="min-width:220px"/>
      <button onclick="loadLogs()">筛选</button>
      <button onclick="clearLogFilter()">清除</button>
      <label class="muted"><input type="checkbox" id="autoScroll" checked/> 自动滚动</label>
      <span class="muted" id="logCount"></span>
    </div>
    <pre id="logBox">加载中…</pre>
  </section>
</main>
<script>
let videos=[];
function esc(s){return String(s??'').replace(/[&<>"]/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c]))}
function fmtSize(n){if(n<1024)return n+' B';if(n<1048576)return (n/1024).toFixed(1)+' KB';return (n/1048576).toFixed(1)+' MB'}
function fmtUptime(s){const m=Math.floor(s/60),sec=s%60;const h=Math.floor(m/60);return h? h+' 小时 '+(m%60)+' 分':m+' 分 '+sec+' 秒'}
async function post(url, body, isForm){
  const opt={method:'POST'};
  if(isForm) opt.body=body;
  else if(body!==undefined){opt.headers={'Content-Type':'application/json'};opt.body=JSON.stringify(body)}
  const r=await fetch(url,opt);
  const j=await r.json().catch(()=>({}));
  if(!r.ok){alert(j.error||('HTTP '+r.status));return null}
  return j;
}
function renderStats(st){
  const live=(st.sessions||[]).length;
  document.getElementById('stats').innerHTML=[
    ['设备编号', st.deviceId],
    ['平台', st.server],
    ['本地监听', st.local],
    ['通道数', (st.channels||[]).length],
    ['点播中', live],
    ['运行', fmtUptime(st.uptimeSec||0)],
  ].map(([t,v])=>'<div class="stat"><div class="t">'+t+'</div><div class="v">'+esc(v)+'</div></div>').join('');
  const b=document.getElementById('regBadge');
  b.className='badge '+(st.registered?'on':'off');
  b.innerHTML='<span class="dot"></span><span>'+(st.registered?'已注册':'未注册')+'</span>';
  document.getElementById('modeBadge').innerHTML='<span>媒体模式：'+esc(st.mediaMode||'—')+'</span>';
}
function videoOptions(selected){
  const opts=['<option value="">（继承全局）</option>'];
  for(const v of videos){
    const sel=selected && selected.endsWith(v.name)?' selected':'';
    opts.push('<option value="'+esc(v.path)+'"'+sel+'>'+esc(v.name)+' ('+fmtSize(v.size)+')</option>');
  }
  opts.push('<option value="__synthetic__">内置测试图案 synthetic</option>');
  return opts.join('');
}
function renderChannels(st, sessMap){
  const box=document.getElementById('chCards');
  const chs=st.channels||[];
  if(!chs.length){box.innerHTML='<div class="muted">暂无通道</div>';return}
  box.innerHTML=chs.map(c=>{
    const sess=sessMap[c.id];
    const playing=!!sess;
    const media=(c.source==='mp4'||c.mp4)?(c.mp4||'—'):c.source;
    return '<div class="ch-card'+(playing?' live':'')+'">'+
      '<div class="ch-head"><div><div class="ch-name">'+esc(c.name)+'</div><div class="ch-id">'+esc(c.id)+'</div></div>'+
      '<div>'+(playing?'<span class="tag live">播放中</span>':'<span class="tag">'+esc(c.status||'ON')+'</span>')+'</div></div>'+
      '<div class="ch-meta"><span>源</span><b>'+esc(c.source||'—')+'</b><span>视频</span><b>'+esc(media)+'</b></div>'+
      (playing?('<div class="row"><button class="danger" onclick="stopSess(\''+esc(sess.callId)+'\')">停止会话</button><span class="muted">'+esc(sess.remoteIp)+':'+sess.remotePort+'</span></div>'):'')+
      '<div class="bind"><select data-ch="'+esc(c.id)+'" class="bindSel">'+videoOptions(c.mp4)+'</select>'+
      '<div class="row"><button class="primary" onclick="bindFromCard(this)">绑定视频</button>'+
      '<button class="danger" onclick="removeChannel(\''+esc(c.id)+'\')">删除</button></div></div>'+
      '</div>';
  }).join('');
}
function renderSessions(st){
  const rows=(st.sessions||[]).map(s=>
    '<tr><td>'+esc(s.channelId)+'</td><td class="muted">'+esc(s.ssrc)+'</td><td>'+esc(s.remoteIp)+':'+s.remotePort+(s.tcp?' TCP':' UDP')+'</td><td>'+(s.sourceReady?'就绪':'抽流中')+'</td><td><button class="danger" onclick="stopSess(\''+esc(s.callId)+'\')">停止</button></td></tr>'
  ).join('');
  document.getElementById('sessBody').innerHTML=rows||'<tr><td colspan="5" class="muted">当前无点播会话</td></tr>';
}
function renderVideos(){
  const list=document.getElementById('videoList');
  const sel=document.getElementById('newChVideo');
  if(!videos.length){list.innerHTML='（视频库为空，可上传 mp4）';sel.innerHTML='<option value="">（不绑定视频）</option>';return}
  list.innerHTML='<table><thead><tr><th>文件</th><th>大小</th></tr></thead><tbody>'+
    videos.map(v=>'<tr><td>'+esc(v.name)+'</td><td class="muted">'+fmtSize(v.size)+'</td></tr>').join('')+'</tbody></table>';
  sel.innerHTML='<option value="">（不绑定视频）</option>'+videos.map(v=>'<option value="'+esc(v.name)+'">'+esc(v.name)+'</option>').join('');
}
async function loadVideos(){
  const j=await fetch('/api/videos').then(r=>r.json());
  videos=j.videos||[];
  renderVideos();
}
async function refresh(){
  const st=await fetch('/api/status').then(r=>r.json());
  const sessMap={};
  for(const s of (st.sessions||[])) sessMap[s.channelId]=s;
  renderStats(st);
  renderChannels(st, sessMap);
  renderSessions(st);
  await loadLogs();
}
async function loadLogs(){
  const q=document.getElementById('logQ').value.trim();
  const url='/api/logs?n=200'+(q?('&q='+encodeURIComponent(q)):'');
  const j=await fetch(url).then(r=>r.json());
  const lines=j.lines||[];
  document.getElementById('logCount').textContent=lines.length+' 行';
  const box=document.getElementById('logBox');
  box.textContent=lines.join('\n')||'（无日志）';
  if(document.getElementById('autoScroll').checked) box.scrollTop=box.scrollHeight;
}
function clearLogFilter(){document.getElementById('logQ').value='';loadLogs()}
async function uploadVideo(){
  const f=document.getElementById('uploadFile').files[0];
  if(!f){alert('请选择文件');return}
  const fd=new FormData();
  fd.append('file', f);
  const btn=event.target; btn.disabled=true; btn.textContent='上传中…';
  try{
    const r=await fetch('/api/videos/upload',{method:'POST',body:fd});
    const j=await r.json();
    if(!r.ok) alert(j.error||'上传失败');
    else {alert('已上传 '+j.name); document.getElementById('uploadFile').value=''; await loadVideos(); await refresh()}
  }catch(e){alert(e.message)}
  btn.disabled=false; btn.textContent='上传';
}
async function addChannel(){
  const id=document.getElementById('newChId').value.trim();
  const name=document.getElementById('newChName').value.trim();
  const mp4=document.getElementById('newChVideo').value;
  if(!id){alert('请填写通道编号');return}
  const j=await post('/api/channels/add',{id,name,mp4});
  if(j){document.getElementById('newChId').value='';document.getElementById('newChName').value='';await refresh()}
}
async function bindFromCard(btn){
  const card=btn.closest('.ch-card');
  const sel=card.querySelector('.bindSel');
  const chId=sel.getAttribute('data-ch');
  const val=sel.value;
  let body={channelId:chId, source:'mp4', mp4:''};
  if(val==='__synthetic__'){body={channelId:chId, source:'synthetic'}}
  else if(val){body={channelId:chId, source: val.endsWith('.h264')||val.endsWith('.264')?'file':'mp4', mp4: val.endsWith('.h264')||val.endsWith('.264')?val:'', h264: val.endsWith('.h264')||val.endsWith('.264')?val:''}}
  const j=await post('/api/channels/bind', body);
  if(j) await refresh();
}
async function removeChannel(id){
  if(!confirm('确认删除通道 '+id+' ？')) return;
  const j=await post('/api/channels/remove?id='+encodeURIComponent(id));
  if(j) await refresh();
}
async function setMode(mode){
  const j=await post('/api/media/mode',{mode});
  if(j) await refresh();
}
async function stopSess(id){ if(await post('/api/session/stop?callId='+encodeURIComponent(id))) await refresh(); }
document.getElementById('logQ').addEventListener('keydown',e=>{if(e.key==='Enter')loadLogs()});
loadVideos().then(refresh);
setInterval(refresh, 4000);
</script>
</body>
</html>
`
