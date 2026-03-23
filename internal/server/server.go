package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/loyispa/jgc_exporter/internal/health"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	port     int
	registry *prometheus.Registry
	monitor  *health.Monitor
	mux      *http.ServeMux
	srv      *http.Server
}

func New(port int, registry *prometheus.Registry, monitor *health.Monitor) *Server {
	s := &Server{
		port:     port,
		registry: registry,
		monitor:  monitor,
		mux:      http.NewServeMux(),
	}
	s.mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	s.mux.HandleFunc("/ui", s.handleUI)
	s.mux.HandleFunc("/ui/", s.handleUI)
	s.mux.HandleFunc("/api/dashboard", s.handleAPIDashboard)
	s.mux.HandleFunc("/", s.handleRoot)
	return s
}

func (s *Server) Addr() string {
	return fmt.Sprintf(":%d", s.port)
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) ListenAndServe() error {
	addr := s.Addr()
	slog.Info("HTTP server starting", "addr", addr)
	s.srv = &http.Server{Addr: addr, Handler: s.mux}
	return s.srv.ListenAndServe()
}

func (s *Server) ListenAndServeOn(ln net.Listener) error {
	s.srv = &http.Server{Handler: s.mux}
	return s.srv.Serve(ln)
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv != nil {
		return s.srv.Shutdown(ctx)
	}
	return nil
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/ui", http.StatusFound)
}

func (s *Server) handleAPIDashboard(w http.ResponseWriter, r *http.Request) {
	dashboard := s.monitor.GetDashboard()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(uiHTML))
}

const uiHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>JGC Exporter - GC Health Dashboard</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#0f1117;color:#e1e4e8;min-height:100vh}
.header{padding:12px 24px;display:flex;align-items:center;gap:20px;border-bottom:1px solid #21262d;flex-wrap:wrap}
.header h1{font-size:20px;font-weight:600;white-space:nowrap}
.file-select-wrap{position:relative;flex:1;max-width:480px;min-width:200px}
.file-select{width:100%;padding:7px 12px;background:#0d1117;border:1px solid #30363d;border-radius:8px;color:#e1e4e8;font-size:13px;appearance:none;-webkit-appearance:none;cursor:pointer;outline:none;padding-right:28px}
.file-select:focus{border-color:#58a6ff}
.file-select-arrow{position:absolute;right:10px;top:50%;transform:translateY(-50%);pointer-events:none;color:#8b949e;font-size:10px}
.header-right{margin-left:auto;font-size:12px;color:#8b949e;white-space:nowrap}
.header-right a{color:#58a6ff;text-decoration:none}
.header-meta{display:flex;align-items:center;gap:16px}
.header-label{color:#8b949e;font-size:13px;white-space:nowrap}
.mini-badge{padding:3px 10px;border-radius:4px;font-size:12px;font-weight:600;display:inline-block}
.header-gctype{font-size:13px;white-space:nowrap;display:inline-flex;align-items:center;gap:6px}
.header-gctype .header-label{color:#8b949e}
.container{max-width:1400px;margin:0 auto;padding:20px}
.section{background:#161b22;border:1px solid #21262d;border-radius:12px;padding:20px;margin-bottom:20px}
.section-title{font-size:16px;font-weight:600;margin-bottom:16px;padding-bottom:8px;border-bottom:1px solid #21262d}
.dashboard-top{margin-bottom:20px}
.row2-grid{display:grid;grid-template-columns:1fr 1fr;gap:20px;margin-bottom:20px}
@media(max-width:900px){.row2-grid{grid-template-columns:1fr}}
.row2-grid .cards-wrapper{display:grid;grid-template-columns:repeat(2,1fr);gap:12px;margin-bottom:0}
@media(max-width:900px){.row2-grid .cards-wrapper{grid-template-columns:1fr}}
.cards{display:contents}
.card{background:#161b22;border:1px solid #21262d;border-radius:12px;padding:20px}
.row2-grid .card{padding:12px}
.row2-grid .card-label{font-size:11px;color:#8b949e;text-transform:uppercase;letter-spacing:0.5px;margin-bottom:6px}
.row2-grid .card-sublabel{font-size:9px;color:#6e7681;text-transform:none;font-weight:400}
.row2-grid .card-value{font-size:20px;font-weight:700}
.row2-grid .card-unit{font-size:12px;color:#8b949e;margin-left:4px}
.card-label{font-size:12px;color:#8b949e;text-transform:uppercase;letter-spacing:0.5px;margin-bottom:8px}
.card-sublabel{font-size:10px;color:#6e7681;text-transform:none;font-weight:400}
.card-value{font-size:28px;font-weight:700}
.card-unit{font-size:14px;color:#8b949e;margin-left:4px}
.chart-box{background:#161b22;border:1px solid #21262d;border-radius:12px;padding:20px;margin-bottom:20px}
.chart-box .section-title{margin-bottom:12px}
.charts-grid{display:grid;grid-template-columns:1fr 1fr;gap:20px;margin-bottom:20px}
@media(max-width:900px){.charts-grid{grid-template-columns:1fr}}
canvas{width:100%!important;height:220px!important}
.chart-container{position:relative;height:220px}
.chart-container-wide{position:relative;height:260px}
.chart-container-wide canvas{width:100%!important;height:260px!important}
.text-green{color:#3fb950}.text-yellow{color:#d29922}.text-red{color:#f85149}.text-dim{color:#8b949e}
.rec-card{background:#0d1117;border:1px solid #21262d;border-radius:8px;padding:14px 16px;margin-bottom:10px}
.rec-condition{font-weight:600;font-size:13px;margin-bottom:6px;color:#f0883e}
.rec-advice{font-size:13px;color:#c9d1d9;margin-bottom:4px}
@keyframes pulse{0%,100%{opacity:1}50%{opacity:.6}}
.status-critical-anim{animation:pulse 2s infinite}
</style>
</head>
<body>
<div class="header">
  <h1>JGC Exporter</h1>
  <span class="header-label">Path:</span>
  <div class="file-select-wrap">
    <select id="fileSelect" class="file-select"><option value="">No files</option></select>
    <span class="file-select-arrow">&#9662;</span>
  </div>
  <div class="header-meta">
    <span id="headerStatus"></span>
    <span id="headerGCType" class="header-gctype"></span>
  </div>
  <div class="header-right">Auto-refresh: 30s | <a href="/metrics">/metrics</a></div>
</div>
<div class="container">
  <div id="recsSection" class="section">
    <div class="section-title">Recommendation</div>
    <div id="recsContainer"></div>
  </div>

  <div class="dashboard-top">
    <div class="row2-grid">
      <div class="cards-wrapper">
        <div id="cards" class="cards"></div>
      </div>
      <div class="chart-box" style="margin-bottom:0">
        <div class="section-title">Heap Usage Trend</div>
        <div class="chart-container"><canvas id="heapCanvas"></canvas></div>
      </div>
    </div>
  </div>

  <div class="chart-box">
    <div class="section-title">GC Events</div>
    <div class="chart-container-wide"><canvas id="gcCountCanvas"></canvas></div>
  </div>

  <div class="charts-grid">
    <div class="chart-box" style="margin-bottom:0">
      <div class="section-title">Pause Time Trend</div>
      <div class="chart-container"><canvas id="pauseCanvas"></canvas></div>
    </div>
    <div class="chart-box" style="margin-bottom:0">
      <div class="section-title">Allocation Rate</div>
      <div class="chart-container"><canvas id="allocCanvas"></canvas></div>
    </div>
  </div>
</div>
<script src="https://cdn.jsdelivr.net/npm/chart.js@4/dist/chart.umd.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/chartjs-adapter-date-fns@3/dist/chartjs-adapter-date-fns.bundle.min.js"></script>
<script>
let heapChart,gcCountChart,pauseChart,allocChart;
let lastDashboard=null;
let selectedPath='';
function defaultTimeLabels(){const now=Date.now();return [new Date(now-3600000),new Date(now)];}
function defaultZeroData(len){return Array(len).fill(0);}
const catColors=['#58a6ff','#f85149','#d29922','#3fb950','#bc8cff','#f78166','#79c0ff','#a5d6ff','#7ee787','#ffa657'];

function categoryColor(cat,idx){
  const palette=['#58a6ff','#f85149','#d29922','#3fb950','#bc8cff','#f78166','#79c0ff','#a5d6ff','#7ee787','#ffa657','#db6d28','#56d364','#e34c26','#8b949e'];
  return palette[idx%palette.length];
}

function statusColor(s){return s==='critical'?'#f85149':s==='warning'?'#d29922':'#3fb950'}
function fmtMs(sec){if(!sec)return'-';if(sec<0.001)return(sec*1e6).toFixed(0)+'us';if(sec<1)return(sec*1000).toFixed(1)+'ms';return sec.toFixed(3)+'s'}
function escHtml(s){return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')}

const fileSelect=document.getElementById('fileSelect');
fileSelect.addEventListener('change',function(){
  selectedPath=this.value;
  if(lastDashboard)renderAll(lastDashboard);
});

function updateFileSelect(files){
  const prev=selectedPath;
  const opts=files.map(f=>'<option value="'+escHtml(f.path)+'"'+(f.path===prev?' selected':'')+'>'+escHtml(f.path)+'</option>');
  if(!opts.length){fileSelect.innerHTML='<option value="">No files</option>';selectedPath='';return}
  fileSelect.innerHTML=opts.join('');
  if(!prev||!files.find(f=>f.path===prev)){
    selectedPath=files[0].path;
    fileSelect.value=selectedPath;
  }
}

function getSelectedFile(d){
  return (d.files||[]).find(f=>f.path===selectedPath)||null;
}

function gcTypeBadgeColor(t){
  const s=(t||'').toUpperCase();
  if(s==='G1')return'#58a6ff';
  if(s==='ZGC')return'#bc8cff';
  if(s==='CMS')return'#f0883e';
  return'#8b949e';
}
function renderHeaderMeta(f){
  const statusEl=document.getElementById('headerStatus');
  const gcTypeEl=document.getElementById('headerGCType');
  if(!f){statusEl.innerHTML='';gcTypeEl.innerHTML='';return}
  const sc=statusColor(f.status);
  const animCls=f.status==='critical'?' status-critical-anim':'';
  statusEl.innerHTML='<span class="header-label">Status:</span> <span class="mini-badge'+animCls+'" style="background:'+sc+'22;color:'+sc+'">'+f.status+'</span>';
  if(!f.gc_type){gcTypeEl.innerHTML='';return}
  const gcCol=gcTypeBadgeColor(f.gc_type);
  gcTypeEl.innerHTML='<span class="header-label">Garbage Collector:</span> <span class="mini-badge" style="background:'+gcCol+'22;color:'+gcCol+'">'+escHtml(f.gc_type)+'</span>';
}

function fmtVal(v){return v===undefined||v===null||(typeof v==='number'&&isNaN(v))?'-':(typeof v==='number'?v.toFixed(1):v)}
function multiCard(label,stat,unit,fmtFn,colorFn){
  const s=stat||{};
  const v1=fmtFn?fmtFn(s.last_1m):fmtVal(s.last_1m);
  const v5=fmtFn?fmtFn(s.avg_5m):fmtVal(s.avg_5m);
  const vh=fmtFn?fmtFn(s.avg_1h):fmtVal(s.avg_1h);
  const val=(v1==='-'&&v5==='-'&&vh==='-')?'-':(v1+' / '+v5+' / '+vh);
  const color=colorFn?colorFn(s.last_1m):'';
  return '<div class="card"><div class="card-label">'+label+' <span class="card-sublabel">(last-1m / 5m-avg / 1h-avg)</span></div><div class="card-value '+(color?'text-'+color:'')+'">'+val+'<span class="card-unit">'+unit+'</span></div></div>';
}
function renderCards(f){
  const c=document.getElementById('cards');
  if(!f){c.innerHTML='';return}
  const tp=f.throughput||{};
  const p99=f.p99_pause||{};
  const gc=f.gc_rate||{};
  const fgc=f.full_gc_rate||{};
  const alloc=f.alloc_rate||{};
  const heap=f.heap_usage||{};
  c.innerHTML=''+
    multiCard('Throughput',tp,'%',function(v){return v!==undefined?(v*100).toFixed(2):'-'},function(v){return v<0.9?'red':v<0.95?'yellow':'green'})+
    multiCard('P99 Pause',p99,'',function(v){return (v!=null&&v!==undefined&&!isNaN(v)&&v>0)?fmtMs(v):'0.0';},function(v){return v>0.5?'red':v>0.2?'yellow':'green'})+
    multiCard('Heap Usage',heap,'%',function(v){return v!==undefined?(v*100).toFixed(0):'-'},function(v){return v>0.9?'red':v>0.8?'yellow':'green'})+
    multiCard('Full GCs',fgc,'',null,function(v){return v>0?'red':'green'})+
    multiCard('Alloc Rate',alloc,' MB/s',null,null)+
    multiCard('GC Rate',gc,'',null,null);
}
function card(label,value,unit,color){
  const cl=color?'text-'+color:'';
  return '<div class="card"><div class="card-label">'+label+'</div><div class="card-value '+cl+'">'+value+'<span class="card-unit">'+unit+'</span></div></div>';
}

function renderHeapChart(history,path){
  const ctx=document.getElementById('heapCanvas');
  const filtered=(history||[]).filter(h=>h.path===path);
  const timestamps=[];const heapUsed=[];const heapTotal=[];
  const youngUsed=[];const oldUsed=[];const metaUsed=[];
  let hasYoung=false,hasOld=false,hasMeta=false;
  filtered.forEach(h=>{
    timestamps.push(new Date(h.timestamp));
    heapUsed.push(h.heap_used_kb/1024);
    heapTotal.push(h.heap_total_kb/1024);
    youngUsed.push(h.young_used_kb/1024);
    oldUsed.push(h.old_used_kb/1024);
    metaUsed.push(h.meta_used_kb/1024);
    if(h.young_used_kb>0)hasYoung=true;
    if(h.old_used_kb>0)hasOld=true;
    if(h.meta_used_kb>0)hasMeta=true;
  });
  const datasets=[
    {label:'Heap Used',data:heapUsed,borderColor:'#58a6ff',backgroundColor:'#58a6ff22',fill:true,tension:0.3,pointRadius:0},
    {label:'Heap Total',data:heapTotal,borderColor:'#8b949e44',borderDash:[4,4],fill:false,tension:0.3,pointRadius:0}
  ];
  if(hasYoung)datasets.push({label:'Young',data:youngUsed,borderColor:'#3fb950',fill:false,tension:0.3,pointRadius:0,borderWidth:1.5});
  if(hasOld)datasets.push({label:'Old',data:oldUsed,borderColor:'#f0883e',fill:false,tension:0.3,pointRadius:0,borderWidth:1.5});
  if(hasMeta)datasets.push({label:'Metaspace',data:metaUsed,borderColor:'#bc8cff',fill:false,tension:0.3,pointRadius:0,borderWidth:1.5});
  const heapOpts={responsive:true,maintainAspectRatio:false,animation:false,interaction:{mode:'index',intersect:false},scales:{x:{type:'time',time:{unit:'minute'},grid:{color:'#21262d'},ticks:{color:'#8b949e'}},y:{grid:{color:'#21262d'},ticks:{color:'#8b949e',callback:function(v){return v>=1024?(v/1024).toFixed(1)+'G':v+'M'}}}},plugins:{legend:{labels:{color:'#e1e4e8',usePointStyle:true,pointStyle:'circle',font:{size:11}}}}};
  const useDefault=!timestamps.length;
  const labels=useDefault?defaultTimeLabels():timestamps;
  const ds=useDefault?[
    {label:'Heap Used',data:defaultZeroData(2),borderColor:'#58a6ff',backgroundColor:'#58a6ff22',fill:true,tension:0.3,pointRadius:0},
    {label:'Heap Total',data:defaultZeroData(2),borderColor:'#8b949e44',borderDash:[4,4],fill:false,tension:0.3,pointRadius:0}
  ]:datasets;
  if(heapChart){
    heapChart.data.labels=labels;heapChart.data.datasets=ds;
    heapChart.update('none');return;
  }
  if(typeof Chart==='undefined')return;
  heapChart=new Chart(ctx,{type:'line',data:{labels,datasets:ds},options:heapOpts});
}

function renderGCEventsChart(series){
  const ctx=document.getElementById('gcCountCanvas');
  const buckets=series||[];
  const allCats=new Set();
  buckets.forEach(b=>{if(b.counts)Object.keys(b.counts).forEach(c=>allCats.add(c))});
  const cats=Array.from(allCats).sort();
  const labels=buckets.map(b=>new Date(b.timestamp));
  const datasets=cats.map((cat,i)=>({
    label:cat,
    data:buckets.map(b=>(b.counts&&b.counts[cat])||0),
    borderColor:categoryColor(cat,i),
    backgroundColor:categoryColor(cat,i)+'33',
    fill:true,
    tension:0.3,
    pointRadius:0,
    borderWidth:1.5
  }));
  const opts={responsive:true,maintainAspectRatio:false,animation:false,interaction:{mode:'index',intersect:false},scales:{x:{type:'time',time:{unit:'minute'},grid:{color:'#21262d'},ticks:{color:'#8b949e'}},y:{beginAtZero:true,stacked:true,grid:{color:'#21262d'},ticks:{color:'#8b949e',precision:0}}},plugins:{legend:{position:'top',labels:{color:'#e1e4e8',usePointStyle:true,pointStyle:'circle',font:{size:10},padding:8,boxWidth:10},maxWidth:400}}};
  const useDefault=!cats.length;
  const finalLabels=useDefault?defaultTimeLabels():labels;
  const finalDatasets=useDefault?[{label:'GC',data:defaultZeroData(2),borderColor:'#8b949e',backgroundColor:'#8b949e33',fill:true,tension:0.3,pointRadius:0,borderWidth:1.5}]:datasets;
  if(gcCountChart){
    gcCountChart.data.labels=finalLabels;gcCountChart.data.datasets=finalDatasets;
    gcCountChart.options.scales.y.stacked=true;
    gcCountChart.update('none');return;
  }
  if(typeof Chart==='undefined')return;
  gcCountChart=new Chart(ctx,{type:'line',data:{labels:finalLabels,datasets:finalDatasets},options:opts});
}

function renderAllocationRateChart(history,path){
  const ctx=document.getElementById('allocCanvas');
  const points=(history||[]).filter(p=>p.path===path);
  const labels=points.map(p=>new Date(p.timestamp));
  const vals=points.map(p=>p.rate_mb_per_sec);
  const opts={responsive:true,maintainAspectRatio:false,animation:false,interaction:{mode:'index',intersect:false},scales:{x:{type:'time',time:{unit:'minute'},grid:{color:'#21262d'},ticks:{color:'#8b949e'}},y:{beginAtZero:true,grid:{color:'#21262d'},ticks:{color:'#8b949e'}}},plugins:{legend:{labels:{color:'#e1e4e8'}}}};
  const useDefault=!points.length;
  const finalLabels=useDefault?defaultTimeLabels():labels;
  const finalVals=useDefault?defaultZeroData(2):vals;
  if(allocChart){
    allocChart.data.labels=finalLabels;allocChart.data.datasets[0].data=finalVals;
    allocChart.update('none');return;
  }
  if(typeof Chart==='undefined')return;
  allocChart=new Chart(ctx,{type:'line',data:{labels:finalLabels,datasets:[{label:'MB/s',data:finalVals,borderColor:'#3fb950',backgroundColor:'#3fb95022',fill:true,tension:0.3,pointRadius:1,pointHoverRadius:4}]},options:opts});
}

function renderPauseChart(pauseHistory,path){
  const ctx=document.getElementById('pauseCanvas');
  const points=(pauseHistory||[]).filter(p=>p.path===path);
  const labels=points.map(p=>new Date(p.timestamp));
  const vals=points.map(p=>p.duration*1000);
  const pauseOpts={responsive:true,maintainAspectRatio:false,animation:false,interaction:{mode:'index',intersect:false},scales:{x:{type:'time',time:{unit:'minute'},grid:{color:'#21262d'},ticks:{color:'#8b949e'}},y:{beginAtZero:true,grid:{color:'#21262d'},ticks:{color:'#8b949e'}}},plugins:{legend:{labels:{color:'#e1e4e8'}}}};
  const useDefault=!points.length;
  const finalLabels=useDefault?defaultTimeLabels():labels;
  const finalVals=useDefault?defaultZeroData(2):vals;
  if(pauseChart){
    pauseChart.data.labels=finalLabels;pauseChart.data.datasets[0].data=finalVals;
    pauseChart.update('none');return;
  }
  if(typeof Chart==='undefined')return;
  pauseChart=new Chart(ctx,{type:'line',data:{labels:finalLabels,datasets:[{label:'Pause (ms)',data:finalVals,borderColor:'#f0883e',backgroundColor:'#f0883e22',fill:true,tension:0.3,pointRadius:1,pointHoverRadius:4}]},options:pauseOpts});
}

function renderRecommendations(recs){
  const el=document.getElementById('recsContainer');
  if(!recs||!recs.length){el.innerHTML='<div class="text-dim">No recommendations at this time</div>';return}
  el.innerHTML=recs.map(r=>{
    return '<div class="rec-card"><div class="rec-condition">'+escHtml(r.condition)+'</div>'+
      '<div class="rec-advice">'+escHtml(r.advice)+'</div></div>';
  }).join('');
}

function renderAll(d){
  updateFileSelect(d.files||[]);
  const f=getSelectedFile(d);
  renderHeaderMeta(f);
  renderCards(f);
  renderHeapChart(d.heap_history,selectedPath);
  renderGCEventsChart(d.gc_events_timeseries);
  renderPauseChart(d.pause_history,selectedPath);
  renderAllocationRateChart(d.allocation_rate_history,selectedPath);
  renderRecommendations(d.recommendations);
}

async function refresh(){
  try{
    const r=await fetch('/api/dashboard');
    lastDashboard=await r.json();
    renderAll(lastDashboard);
  }catch(e){console.error('refresh failed',e)}
}
refresh();
setInterval(refresh,30000);
</script>
</body>
</html>`
