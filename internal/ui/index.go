package ui

const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>GB28181 模拟设备控制台</title>
<style>
:root{
  --bg:#0e1116;
  --surface:#161b22;
  --surface-2:#1c232d;
  --line:#2a313c;
  --line-soft:#222933;
  --text:#e6edf3;
  --text-2:#9aa7b5;
  --text-3:#6e7a88;
  --ok:#3dd68c;
  --warn:#e3b341;
  --err:#f47067;
  --info:#58a6ff;
  --accent:#388bfd;
  --accent-ink:#041220;
  --radius:10px;
  --shadow:0 1px 0 rgba(255,255,255,.03) inset, 0 8px 24px rgba(0,0,0,.28);
}
*{box-sizing:border-box}
html,body{height:100%}
body{
  margin:0;
  font:14px/1.55 "Segoe UI",system-ui,-apple-system,"Microsoft YaHei",sans-serif;
  color:var(--text);
  background:
    linear-gradient(180deg,#12171e 0%, var(--bg) 220px);
}
a{color:var(--info)}
button,select,input[type=text],input[type=file]{
  font:inherit;color:var(--text);
}
button{
  appearance:none;border:1px solid var(--line);
  background:linear-gradient(180deg,#232b36,#1b222b);
  border-radius:8px;padding:7px 12px;cursor:pointer;
  transition:border-color .12s ease, background .12s ease, color .12s ease;
}
button:hover{border-color:#3d4a5a;background:linear-gradient(180deg,#283240,#202833)}
button:active{transform:translateY(1px)}
button:focus-visible,select:focus-visible,input:focus-visible{
  outline:2px solid rgba(56,139,253,.55);outline-offset:2px;
}
button.primary{
  background:linear-gradient(180deg,#3d97ff,#2f7fdf);
  border-color:#2f7fdf;color:#fff;
}
button.primary:hover{background:linear-gradient(180deg,#57a4ff,#3d8ef0)}
button.danger{color:#ffb1ad;border-color:#5a3036;background:linear-gradient(180deg,#2a1e21,#22181b)}
button.danger:hover{border-color:#8a454d}
button:disabled{opacity:.55;cursor:not-allowed;transform:none}
select,input[type=text]{
  background:var(--surface-2);border:1px solid var(--line);border-radius:8px;padding:7px 10px;
}
input[type=text]{min-width:0;width:100%}
input[type=file]{font-size:12px;color:var(--text-2)}

/* header */
.top{
  position:sticky;top:0;z-index:20;
  display:flex;align-items:center;justify-content:space-between;gap:16px;
  padding:0 24px;height:56px;
  background:rgba(14,17,22,.92);
  border-bottom:1px solid var(--line);
  backdrop-filter:blur(10px);
}
.brand{display:flex;align-items:center;gap:12px;min-width:0}
.mark{
  width:28px;height:28px;border-radius:8px;flex:0 0 auto;
  background:
    linear-gradient(135deg,#1f6feb 0%, #388bfd 45%, #79c0ff 100%);
  box-shadow:0 0 0 1px rgba(255,255,255,.08) inset;
  position:relative;
}
.mark:after{
  content:"";position:absolute;inset:8px;border-radius:3px;
  border:1.5px solid rgba(255,255,255,.85);
  border-right-color:transparent;border-bottom-color:transparent;
  transform:rotate(45deg);
}
.brand h1{margin:0;font-size:14px;font-weight:600;letter-spacing:.01em;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.brand .sub{display:block;font-size:11px;color:var(--text-3);font-weight:400;margin-top:1px}
.pills{display:flex;gap:8px;align-items:center;flex-wrap:wrap;justify-content:flex-end}
.pill{
  display:inline-flex;align-items:center;gap:7px;
  height:28px;padding:0 10px;border-radius:999px;
  border:1px solid var(--line);background:var(--surface);
  font-size:12px;color:var(--text-2);
}
.pill .dot{width:7px;height:7px;border-radius:50%;background:var(--text-3)}
.pill.on{color:#c8f5df;border-color:#245c43;background:#12241c}
.pill.on .dot{background:var(--ok);box-shadow:0 0 0 3px rgba(61,214,140,.12)}
.pill.off{color:#ffc9c5;border-color:#5a3036;background:#241416}
.pill.off .dot{background:var(--err)}
.pill.mode{color:var(--text-2)}

.wrap{
  max-width:1180px;margin:0 auto;padding:20px 20px 48px;
  display:grid;gap:16px;
}

/* stats */
.stats{
  display:grid;grid-template-columns:repeat(6,1fr);gap:10px;
}
@media (max-width:1100px){.stats{grid-template-columns:repeat(3,1fr)}}
@media (max-width:640px){.stats{grid-template-columns:repeat(2,1fr)}}
.stat{
  background:var(--surface);border:1px solid var(--line-soft);border-radius:var(--radius);
  padding:12px 14px;box-shadow:var(--shadow);min-width:0;
}
.stat .k{font-size:11px;color:var(--text-3);letter-spacing:.04em;text-transform:uppercase}
.stat .v{
  margin-top:6px;font-size:16px;font-weight:600;
  font-variant-numeric:tabular-nums;
  white-space:nowrap;overflow:hidden;text-overflow:ellipsis;
}
.stat .v.sm{font-size:13px;font-weight:500;color:var(--text-2)}

/* panels */
.panel{
  background:var(--surface);border:1px solid var(--line);border-radius:12px;
  box-shadow:var(--shadow);overflow:hidden;
}
.panel-hd{
  display:flex;align-items:center;justify-content:space-between;gap:12px;
  padding:12px 16px;border-bottom:1px solid var(--line-soft);
  background:linear-gradient(180deg,#1a2029,#161b22);
}
.panel-hd h2{
  margin:0;font-size:12px;font-weight:600;letter-spacing:.08em;
  text-transform:uppercase;color:var(--text-2);
}
.panel-bd{padding:14px 16px}
.panel-bd.tight{padding:0}

.row{display:flex;flex-wrap:wrap;gap:8px;align-items:center}
.muted{color:var(--text-3);font-size:12px}
.hint{margin:10px 0 0;color:var(--text-3);font-size:12px;line-height:1.5}

/* channel cards */
.grid2{display:grid;grid-template-columns:1.35fr .9fr;gap:16px;align-items:start}
@media (max-width:980px){.grid2{grid-template-columns:1fr}}
.channels{
  display:grid;grid-template-columns:repeat(auto-fill,minmax(270px,1fr));gap:12px;
  padding:14px 16px 16px;
}
.ch{
  background:var(--surface-2);border:1px solid var(--line);border-radius:12px;
  padding:14px;display:flex;flex-direction:column;gap:12px;position:relative;
  overflow:hidden;
}
.ch:before{
  content:"";position:absolute;left:0;top:0;bottom:0;width:3px;background:transparent;
}
.ch.live:before{background:var(--ok)}
.ch.live{border-color:#2a4a3a}
.ch-top{display:flex;justify-content:space-between;gap:10px;align-items:flex-start}
.ch-name{font-size:14px;font-weight:600}
.ch-id{margin-top:3px;font:11px/1.4 ui-monospace,SFMono-Regular,Consolas,monospace;color:var(--text-3);word-break:break-all}
.status{
  flex:0 0 auto;font-size:11px;padding:2px 8px;border-radius:999px;
  border:1px solid var(--line);color:var(--text-2);background:#141a22;
}
.status.live{border-color:#245c43;color:#8dffc8;background:#12241c}
.meta{
  display:grid;grid-template-columns:52px 1fr;gap:6px 8px;
  font-size:12px;color:var(--text-3);
}
.meta b{color:var(--text-2);font-weight:500;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.bind{display:grid;gap:8px;padding-top:4px;border-top:1px dashed var(--line-soft)}
.bind select{width:100%}

/* forms stack */
.stack{display:grid;gap:10px}
.field label{display:block;font-size:12px;color:var(--text-3);margin-bottom:6px}
.side{display:grid;gap:16px}

/* table */
.tbl{width:100%;border-collapse:collapse;font-size:13px}
.tbl th,.tbl td{padding:10px 14px;text-align:left;border-bottom:1px solid var(--line-soft);vertical-align:middle}
.tbl th{
  font-size:11px;font-weight:600;letter-spacing:.05em;text-transform:uppercase;color:var(--text-3);
  background:#141920;
}
.tbl tr:last-child td{border-bottom:none}
.tbl tr:hover td{background:rgba(255,255,255,.015)}
.empty{padding:18px 16px;color:var(--text-3);font-size:13px}

/* logs */
.log-tools{display:flex;flex-wrap:wrap;gap:8px;align-items:center}
.log-tools input[type=text]{width:min(280px,100%)}
#logBox{
  margin:0;height:300px;overflow:auto;
  background:#0b0f14;border-top:1px solid var(--line-soft);
  padding:12px 14px;
  font:12px/1.55 ui-monospace,SFMono-Regular,Consolas,monospace;
  color:#9eafc2;white-space:pre-wrap;word-break:break-word;
}
#logBox::-webkit-scrollbar{width:10px;height:10px}
#logBox::-webkit-scrollbar-thumb{background:#2a3340;border-radius:8px;border:2px solid #0b0f14}

.videos{max-height:180px;overflow:auto}
</style>
</head>
<body>
<header class="top">
  <div class="brand">
    <div class="mark" aria-hidden="true"></div>
    <div>
      <h1>GB28181 模拟设备</h1>
      <span class="sub">设备接入 · 通道媒体 · 运行日志</span>
    </div>
  </div>
  <div class="pills">
    <div id="modeBadge" class="pill mode">媒体模式 —</div>
    <div id="regBadge" class="pill off"><span class="dot"></span><span>检测中</span></div>
  </div>
</header>

<div class="wrap">
  <section class="stats" id="stats"></section>

  <section class="panel">
    <div class="panel-hd"><h2>快捷操作</h2></div>
    <div class="panel-bd">
      <div class="row">
        <button class="primary" onclick="post('/api/register')">立即注册</button>
        <button onclick="post('/api/keepalive')">发送心跳</button>
        <button onclick="post('/api/alarm')">发送报警</button>
        <button onclick="refresh()">刷新状态</button>
        <span class="muted">点播在 WVP 发起；改绑定后需重新点播。</span>
      </div>
    </div>
  </section>

  <section class="grid2">
    <div class="panel">
      <div class="panel-hd">
        <h2>通道</h2>
        <span class="muted" id="chCount"></span>
      </div>
      <div class="channels" id="chCards"></div>
    </div>

    <div class="side">
      <div class="panel">
        <div class="panel-hd"><h2>新增通道</h2></div>
        <div class="panel-bd stack">
          <div class="field">
            <label for="newChId">国标编号（20 位）</label>
            <input id="newChId" type="text" placeholder="例如 34020000001320000003" autocomplete="off"/>
          </div>
          <div class="field">
            <label for="newChName">名称（可选）</label>
            <input id="newChName" type="text" placeholder="通道名称" autocomplete="off"/>
          </div>
          <div class="field">
            <label for="newChVideo">初始视频（可选）</label>
            <select id="newChVideo"><option value="">暂不绑定</option></select>
          </div>
          <div class="row">
            <button class="primary" onclick="addChannel()">添加通道</button>
          </div>
          <p class="hint">添加后请在 WVP 重新同步设备目录。</p>
        </div>
      </div>

      <div class="panel">
        <div class="panel-hd"><h2>媒体模式</h2></div>
        <div class="panel-bd">
          <div class="row">
            <button onclick="setMode('shared')">全通道共用</button>
            <button onclick="setMode('per_channel')">按通道绑定</button>
          </div>
          <p class="hint">在通道卡上绑定视频时，会自动切到「按通道绑定」。</p>
        </div>
      </div>

      <div class="panel">
        <div class="panel-hd"><h2>视频库</h2></div>
        <div class="panel-bd stack">
          <div class="row">
            <input id="uploadFile" type="file" accept=".mp4,.h264,.264" style="flex:1;min-width:0"/>
            <button class="primary" id="uploadBtn" onclick="uploadVideo()">上传</button>
          </div>
          <div id="videoList" class="muted videos">加载中…</div>
          <p class="hint">文件保存到工作目录 <code>assets/</code>。</p>
        </div>
      </div>
    </div>
  </section>

  <section class="panel">
    <div class="panel-hd"><h2>点播会话</h2></div>
    <div class="panel-bd tight">
      <table class="tbl">
        <thead>
          <tr><th>通道</th><th>SSRC</th><th>对端地址</th><th>源状态</th><th style="width:88px">操作</th></tr>
        </thead>
        <tbody id="sessBody"></tbody>
      </table>
    </div>
  </section>

  <section class="panel">
    <div class="panel-hd">
      <h2>运行日志</h2>
      <span class="muted" id="logCount"></span>
    </div>
    <div class="panel-bd">
      <div class="log-tools">
        <input id="logQ" type="text" placeholder="过滤关键字，如 media / invite / error"/>
        <button onclick="loadLogs()">筛选</button>
        <button onclick="clearLogFilter()">清除</button>
        <label class="muted"><input type="checkbox" id="autoScroll" checked/> 自动滚动</label>
      </div>
    </div>
    <pre id="logBox">加载中…</pre>
  </section>
</div>

<script>
let videos=[];
function esc(s){return String(s??'').replace(/[&<>"]/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c]))}
function fmtSize(n){if(n<1024)return n+' B';if(n<1048576)return (n/1024).toFixed(1)+' KB';return (n/1048576).toFixed(1)+' MB'}
function fmtUptime(s){const m=Math.floor(s/60),sec=s%60,h=Math.floor(m/60);return h?(h+' 小时 '+(m%60)+' 分'):(m+' 分 '+sec+' 秒')}
async function post(url, body){
  const opt={method:'POST'};
  if(body!==undefined){opt.headers={'Content-Type':'application/json'};opt.body=JSON.stringify(body)}
  const r=await fetch(url,opt);
  const j=await r.json().catch(()=>({}));
  if(!r.ok){alert(j.error||('HTTP '+r.status));return null}
  return j;
}
function renderStats(st){
  const items=[
    ['设备编号', st.deviceId, true],
    ['平台', st.server, true],
    ['本地监听', st.local, true],
    ['通道', String((st.channels||[]).length), false],
    ['点播中', String((st.sessions||[]).length), false],
    ['运行', fmtUptime(st.uptimeSec||0), false],
  ];
  document.getElementById('stats').innerHTML=items.map(([k,v,sm])=>
    '<div class="stat"><div class="k">'+esc(k)+'</div><div class="v'+(sm?' sm':'')+'" title="'+esc(v)+'">'+esc(v)+'</div></div>'
  ).join('');
  const b=document.getElementById('regBadge');
  b.className='pill '+(st.registered?'on':'off');
  b.innerHTML='<span class="dot"></span><span>'+(st.registered?'已注册':'未注册')+'</span>';
  document.getElementById('modeBadge').textContent='媒体模式 '+(st.mediaMode||'—');
}
function videoOptions(selected){
  const opts=['<option value="">继承全局</option>'];
  for(const v of videos){
    const sel=(selected && (selected.endsWith(v.name)||selected===v.path))?' selected':'';
    opts.push('<option value="'+esc(v.path)+'"'+sel+'>'+esc(v.name)+' · '+fmtSize(v.size)+'</option>');
  }
  opts.push('<option value="__synthetic__">内置测试图案</option>');
  return opts.join('');
}
function isH264(p){return /\.(h264|264)$/i.test(p||'')}
function renderChannels(st, sessMap){
  const box=document.getElementById('chCards');
  const chs=st.channels||[];
  document.getElementById('chCount').textContent=chs.length? (chs.length+' 路') : '';
  if(!chs.length){box.innerHTML='<div class="empty">还没有通道。可在右侧「新增通道」创建。</div>';return}
  box.innerHTML=chs.map(c=>{
    const sess=sessMap[c.id];
    const playing=!!sess;
    let media=c.mp4||c.h264||'';
    if(!media) media=c.source==='synthetic'?'内置测试图案':(c.source||'—');
    return '<article class="ch'+(playing?' live':'')+'">'+
      '<div class="ch-top"><div style="min-width:0">'+
        '<div class="ch-name">'+esc(c.name||'未命名')+'</div>'+
        '<div class="ch-id">'+esc(c.id)+'</div>'+
      '</div><span class="status'+(playing?' live':'')+'">'+(playing?'播放中':esc(c.status||'ON'))+'</span></div>'+
      '<div class="meta">'+
        '<span>源</span><b>'+esc(c.source||'—')+'</b>'+
        '<span>视频</span><b title="'+esc(media)+'">'+esc(media)+'</b>'+
      '</div>'+
      (playing?('<div class="row"><button class="danger" onclick="stopSess(\''+esc(sess.callId)+'\')">停止</button><span class="muted">'+esc(sess.remoteIp)+':'+sess.remotePort+(sess.tcp?' TCP':' UDP')+'</span></div>'):'')+
      '<div class="bind">'+
        '<select data-ch="'+esc(c.id)+'" class="bindSel">'+videoOptions(c.mp4||c.h264)+'</select>'+
        '<div class="row">'+
          '<button class="primary" onclick="bindFromCard(this)">绑定视频</button>'+
          '<button class="danger" onclick="removeChannel(\''+esc(c.id)+'\')">删除</button>'+
        '</div>'+
      '</div></article>';
  }).join('');
}
function renderSessions(st){
  const body=document.getElementById('sessBody');
  const rows=(st.sessions||[]);
  if(!rows.length){body.innerHTML='<tr><td colspan="5"><div class="empty">当前没有点播会话。在 WVP 打开通道预览后会出现在这里。</div></td></tr>';return}
  body.innerHTML=rows.map(s=>
    '<tr><td>'+esc(s.channelId)+'</td>'+
    '<td class="muted" style="font-family:ui-monospace,Consolas,monospace">'+esc(s.ssrc)+'</td>'+
    '<td>'+esc(s.remoteIp)+':'+s.remotePort+(s.tcp?' TCP':' UDP')+'</td>'+
    '<td>'+(s.sourceReady?'<span class="status live">就绪</span>':'<span class="status">抽流中</span>')+'</td>'+
    '<td><button class="danger" onclick="stopSess(\''+esc(s.callId)+'\')">停止</button></td></tr>'
  ).join('');
}
function renderVideos(){
  const list=document.getElementById('videoList');
  const sel=document.getElementById('newChVideo');
  if(!videos.length){
    list.textContent='视频库为空，可上传 mp4 / h264。';
    sel.innerHTML='<option value="">暂不绑定</option>';
    return;
  }
  list.innerHTML='<table class="tbl videos"><thead><tr><th>文件</th><th style="width:88px">大小</th></tr></thead><tbody>'+
    videos.map(v=>'<tr><td>'+esc(v.name)+'</td><td class="muted">'+fmtSize(v.size)+'</td></tr>').join('')+
    '</tbody></table>';
  sel.innerHTML='<option value="">暂不绑定</option>'+videos.map(v=>'<option value="'+esc(v.name)+'">'+esc(v.name)+'</option>').join('');
}
async function loadVideos(){
  try{
    const j=await fetch('/api/videos').then(r=>r.json());
    videos=j.videos||[];
  }catch(e){videos=[]}
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
  const j=await fetch(url).then(r=>r.json()).catch(()=>({lines:[]}));
  const lines=j.lines||[];
  document.getElementById('logCount').textContent=lines.length+' 行';
  const box=document.getElementById('logBox');
  box.textContent=lines.join('\n')||'（暂无日志）';
  if(document.getElementById('autoScroll').checked) box.scrollTop=box.scrollHeight;
}
function clearLogFilter(){document.getElementById('logQ').value='';loadLogs()}
async function uploadVideo(){
  const f=document.getElementById('uploadFile').files[0];
  if(!f){alert('请选择文件');return}
  const fd=new FormData();
  fd.append('file', f);
  const btn=document.getElementById('uploadBtn');
  btn.disabled=true; btn.textContent='上传中…';
  try{
    const r=await fetch('/api/videos/upload',{method:'POST',body:fd});
    const j=await r.json();
    if(!r.ok) alert(j.error||'上传失败');
    else {
      alert('已上传 '+j.name);
      document.getElementById('uploadFile').value='';
      await loadVideos(); await refresh();
    }
  }catch(e){alert(e.message)}
  btn.disabled=false; btn.textContent='上传';
}
async function addChannel(){
  const id=document.getElementById('newChId').value.trim();
  const name=document.getElementById('newChName').value.trim();
  const mp4=document.getElementById('newChVideo').value;
  if(!id){alert('请填写 20 位国标编号');document.getElementById('newChId').focus();return}
  const j=await post('/api/channels/add',{id,name,mp4});
  if(j){
    document.getElementById('newChId').value='';
    document.getElementById('newChName').value='';
    await refresh();
  }
}
async function bindFromCard(btn){
  const card=btn.closest('.ch');
  const sel=card.querySelector('.bindSel');
  const chId=sel.getAttribute('data-ch');
  const val=sel.value;
  let body;
  if(val==='__synthetic__') body={channelId:chId, source:'synthetic'};
  else if(!val) body={channelId:chId, source:'mp4', mp4:''};
  else if(isH264(val)) body={channelId:chId, source:'file', h264:val};
  else body={channelId:chId, source:'mp4', mp4:val};
  const j=await post('/api/channels/bind', body);
  if(j) await refresh();
}
async function removeChannel(id){
  if(!confirm('确认删除通道 '+id+' ？删除后需在 WVP 重新同步目录。')) return;
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
