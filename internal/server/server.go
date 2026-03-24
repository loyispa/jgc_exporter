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
.row2-grid .card-label{font-size:11px;color:#e1e4e8;text-transform:uppercase;letter-spacing:0.5px;margin-bottom:6px}
.row2-grid .card-sublabel{font-size:9px;color:#6e7681;text-transform:none;font-weight:400}
.row2-grid .card-value{font-size:20px;font-weight:700}
.row2-grid .card-unit{font-size:12px;color:#8b949e;margin-left:4px}
.card-label{font-size:12px;color:#e1e4e8;text-transform:uppercase;letter-spacing:0.5px;margin-bottom:8px}
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
.gc-two-col-row{display:grid;grid-template-columns:1fr 1fr;gap:20px;margin-bottom:20px;align-items:stretch}
@media(max-width:1100px){.gc-two-col-row{grid-template-columns:1fr}}
.gc-half-chart{margin-bottom:0}
.gc-half-chart .chart-container-wide{margin-top:0}
.gc-events-head{display:flex;align-items:center;flex-wrap:wrap;gap:12px 16px;margin-bottom:12px;padding-bottom:8px;border-bottom:1px solid #21262d}
.gc-events-head .section-title{margin:0;padding:0;border:none}
.gc-event-filter-wrap{position:relative;flex:1;min-width:180px;max-width:320px}
.gc-event-filter-btn{width:100%;padding:7px 28px 7px 12px;background:#0d1117;border:1px solid #30363d;border-radius:8px;color:#e1e4e8;font-size:13px;text-align:left;cursor:pointer;outline:none;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.gc-event-filter-btn:hover{border-color:#484f58}
.gc-event-filter-btn:focus,.gc-event-filter-btn[aria-expanded="true"]{border-color:#58a6ff}
.gc-event-filter-arrow{position:absolute;right:10px;top:50%;transform:translateY(-50%);pointer-events:none;color:#8b949e;font-size:10px}
.gc-event-filter-panel{display:none;position:absolute;left:0;right:0;top:calc(100% + 4px);z-index:30;background:#161b22;border:1px solid #30363d;border-radius:8px;box-shadow:0 8px 24px rgba(0,0,0,.45);max-height:260px;overflow-y:auto;padding:6px 0}
.gc-event-filter-panel.open{display:block}
.gc-event-filter-option{display:flex;align-items:center;gap:8px;padding:8px 12px;font-size:13px;color:#e1e4e8;cursor:pointer;user-select:none}
.gc-event-filter-option:hover{background:#21262d}
.gc-event-filter-option input{accent-color:#58a6ff;cursor:pointer;flex-shrink:0}
.gc-type-filter-search-wrap{padding:6px 10px 4px;border-bottom:1px solid #21262d}
.gc-type-filter-search{width:100%;box-sizing:border-box;padding:6px 10px;background:#0d1117;border:1px solid #30363d;border-radius:6px;color:#e1e4e8;font-size:12px;outline:none}
.gc-type-filter-search:focus{border-color:#58a6ff}
.gc-type-filter-search::placeholder{color:#6e7681}
.text-green{color:#3fb950}.text-yellow{color:#d29922}.text-red{color:#f85149}.text-dim{color:#8b949e}
.card-value-accent-alloc{color:#7ee787}
.card-value-accent-gc{color:#3fb950}
.recs-empty{font-size:11px;color:#8b949e;line-height:1.4}
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
  <div class="header-right">Auto-refresh: 5s | <a href="/metrics">/metrics</a></div>
</div>
<div class="container">
  <div id="recsSection" class="section">
    <div class="section-title">Tuning Advice</div>
    <div id="recsContainer"></div>
  </div>

  <div class="dashboard-top">
    <div class="row2-grid">
      <div class="cards-wrapper">
        <div id="cards" class="cards"></div>
      </div>
      <div class="chart-box" style="margin-bottom:0">
        <div class="section-title">Heap Usage</div>
        <div class="chart-container"><canvas id="heapCanvas"></canvas></div>
      </div>
    </div>
  </div>

  <div class="gc-two-col-row">
    <div class="chart-box gc-half-chart">
      <div class="gc-events-head">
        <div class="section-title">GC Events</div>
        <span class="header-label">Types</span>
        <div id="gcEventTypeFilter" class="gc-event-filter-wrap">
          <button type="button" id="gcEventFilterBtn" class="gc-event-filter-btn" aria-expanded="false" aria-haspopup="listbox">Total</button>
          <span class="gc-event-filter-arrow" aria-hidden="true">&#9662;</span>
          <div id="gcEventFilterPanel" class="gc-event-filter-panel" role="listbox" aria-multiselectable="true">
            <div class="gc-type-filter-search-wrap">
              <input type="search" id="gcEventFilterSearch" class="gc-type-filter-search" placeholder="Filter types..." autocomplete="off" aria-label="Filter event types"/>
            </div>
            <div id="gcEventFilterList"></div>
          </div>
        </div>
      </div>
      <div class="chart-container-wide"><canvas id="gcCountCanvas"></canvas></div>
    </div>
    <div class="chart-box gc-half-chart">
      <div class="gc-events-head">
        <div class="section-title">GC Duration</div>
        <span class="header-label">Types</span>
        <div id="gcDurationTypeFilter" class="gc-event-filter-wrap">
          <button type="button" id="gcDurationFilterBtn" class="gc-event-filter-btn" aria-expanded="false" aria-haspopup="listbox">Total</button>
          <span class="gc-event-filter-arrow" aria-hidden="true">&#9662;</span>
          <div id="gcDurationFilterPanel" class="gc-event-filter-panel" role="listbox" aria-multiselectable="true">
            <div class="gc-type-filter-search-wrap">
              <input type="search" id="gcDurationFilterSearch" class="gc-type-filter-search" placeholder="Filter types..." autocomplete="off" aria-label="Filter duration types"/>
            </div>
            <div id="gcDurationFilterList"></div>
          </div>
        </div>
      </div>
      <div class="chart-container-wide"><canvas id="gcDurationCanvas"></canvas></div>
    </div>
  </div>

  <div class="charts-grid">
    <div class="chart-box" style="margin-bottom:0">
      <div class="section-title">Throughput</div>
      <div class="chart-container"><canvas id="throughputCanvas"></canvas></div>
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
let heapChart,gcCountChart,gcDurChart,throughputChart,allocChart;
let lastDashboard=null;
let selectedPath='';
function truncateToMinuteMs(ms){
  const d=new Date(ms);
  if(isNaN(d.getTime())){return 0;}
  return Date.UTC(d.getUTCFullYear(),d.getUTCMonth(),d.getUTCDate(),d.getUTCHours(),d.getUTCMinutes(),0,0);
}
function buildMinuteLabelDates(minMs,maxMs){
  const start=truncateToMinuteMs(minMs);
  const end=truncateToMinuteMs(maxMs);
  const labels=[];
  for(let t=start;t<=end;t+=60000){labels.push(new Date(t));}
  if(labels.length===0){labels.push(new Date(end));}
  if(labels.length===1){labels.push(new Date(start+60000));}
  return labels;
}
function defaultMinuteLabelsLastHour(){
  const end=truncateToMinuteMs(Date.now());
  const labels=[];
  for(let t=end-60*60000;t<=end;t+=60000){labels.push(new Date(t));}
  if(labels.length===0){labels.push(new Date(end));}
  if(labels.length===1){labels.push(new Date(end+60000));}
  return labels;
}
function mapLastPerMinute(labels,rows,getTs,getVal){
  const m=new Map();
  rows.forEach(function(r){
    const k=truncateToMinuteMs(new Date(getTs(r)).getTime());
    m.set(k,getVal(r));
  });
  return labels.map(function(l){return m.has(l.getTime())?m.get(l.getTime()):0;});
}
function legendWithIsolate(baseLegend){
  const b=baseLegend||{};
  return Object.assign({},b,{
    onClick:function(e,legendItem,legend){
      const chart=legend.chart;
      const n=chart.data.datasets.length;
      if(n<=1){return;}
      const i=legendItem.datasetIndex;
      let vis=0;
      for(let j=0;j<n;j++){if(chart.isDatasetVisible(j)){vis++;}}
      if(vis===1&&chart.isDatasetVisible(i)){
        for(let j=0;j<n;j++){chart.setDatasetVisibility(j,true);}
      }else{
        for(let j=0;j<n;j++){chart.setDatasetVisibility(j,j===i);}
      }
      chart.update('none');
    }
  });
}
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

const gcEventTypeFilterRoot=document.getElementById('gcEventTypeFilter');
const gcEventFilterBtn=document.getElementById('gcEventFilterBtn');
const gcEventFilterPanel=document.getElementById('gcEventFilterPanel');
const gcEventFilterList=document.getElementById('gcEventFilterList');
function setGCEventFilterPanelOpen(open){
  if(!gcEventFilterPanel||!gcEventFilterBtn){return;}
  if(open){
    gcEventFilterPanel.classList.add('open');
    gcEventFilterBtn.setAttribute('aria-expanded','true');
  }else{
    gcEventFilterPanel.classList.remove('open');
    gcEventFilterBtn.setAttribute('aria-expanded','false');
  }
}
function updateGCEventFilterBtnLabel(){
  if(!gcEventFilterBtn){return;}
  var mode=getGCEventChartFilterMode();
  if(mode.mode==='total'){
    gcEventFilterBtn.textContent='Total';
    return;
  }
  if(mode.types.length===1){
    gcEventFilterBtn.textContent=mode.types[0];
    return;
  }
  gcEventFilterBtn.textContent=String(mode.types.length)+' types';
}
function normalizeGCEventFilterSelection(root){
  if(!root){return;}
  var typeCbs=root.querySelectorAll('input[type=checkbox][data-gc-ev]:not([value="__ALL__"])');
  var allCb=root.querySelector('input[type=checkbox][data-gc-ev][value="__ALL__"]');
  var anyType=false;
  for(var i=0;i<typeCbs.length;i++){if(typeCbs[i].checked){anyType=true;break;}}
  if(anyType){
    if(allCb){allCb.checked=false;}
  }else{
    if(allCb){allCb.checked=true;}
    for(var j=0;j<typeCbs.length;j++){typeCbs[j].checked=false;}
  }
}
function getGCEventChartFilterMode(){
  if(!gcEventTypeFilterRoot){return{mode:'total'};}
  var types=[];
  var boxes=gcEventTypeFilterRoot.querySelectorAll('input[type=checkbox][data-gc-ev]:checked');
  for(var i=0;i<boxes.length;i++){
    var v=boxes[i].value;
    if(v!=='__ALL__'){types.push(v);}
  }
  if(types.length===0){return{mode:'total'};}
  return{mode:'types',types:types};
}
function appendGCEventFilterRow(list,val,label,checked){
  var lab=document.createElement('label');
  lab.className='gc-event-filter-option';
  var inp=document.createElement('input');
  inp.type='checkbox';
  inp.value=val;
  inp.setAttribute('data-gc-ev','1');
  inp.checked=checked;
  lab.appendChild(inp);
  lab.appendChild(document.createTextNode(' '+label));
  list.appendChild(lab);
}
function setGCEventCheckboxChecked(container,val,on){
  var boxes=container.querySelectorAll('input[type=checkbox][data-gc-ev]');
  for(var i=0;i<boxes.length;i++){
    if(boxes[i].value===val){boxes[i].checked=on;return;}
  }
}
function updateGCEventTypeFilter(series){
  if(!gcEventFilterList||!gcEventTypeFilterRoot){return;}
  var buckets=series||[];
  var allCats=new Set();
  buckets.forEach(function(b){
    if(b.counts){Object.keys(b.counts).forEach(function(c){allCats.add(c);});}
  });
  var cats=Array.from(allCats).sort();
  var prev=[];
  var oldBoxes=gcEventFilterList.querySelectorAll('input[type=checkbox][data-gc-ev]');
  for(var p=0;p<oldBoxes.length;p++){
    if(oldBoxes[p].checked){prev.push(oldBoxes[p].value);}
  }
  gcEventFilterList.innerHTML='';
  appendGCEventFilterRow(gcEventFilterList,'__ALL__','Total',false);
  cats.forEach(function(c){
    appendGCEventFilterRow(gcEventFilterList,c,c,false);
  });
  var optSet={};
  optSet['__ALL__']=true;
  for(var cidx=0;cidx<cats.length;cidx++){optSet[cats[cidx]]=true;}
  var any=false;
  for(var j=0;j<prev.length;j++){
    var v=prev[j];
    if(!optSet[v]){continue;}
    setGCEventCheckboxChecked(gcEventFilterList,v,true);
    any=true;
  }
  if(!any){setGCEventCheckboxChecked(gcEventFilterList,'__ALL__',true);}
  normalizeGCEventFilterSelection(gcEventTypeFilterRoot);
  updateGCEventFilterBtnLabel();
  refreshGCTypeSearchFilter('gcEventFilterSearch');
}
if(gcEventFilterBtn&&gcEventFilterPanel){
  gcEventFilterBtn.addEventListener('click',function(e){
    e.stopPropagation();
    var open=!gcEventFilterPanel.classList.contains('open');
    setGCEventFilterPanelOpen(open);
  });
  gcEventFilterPanel.addEventListener('click',function(e){e.stopPropagation();});
}
if(gcEventFilterList){
  gcEventFilterList.addEventListener('change',function(e){
    var t=e.target;
    if(!t||t.getAttribute('data-gc-ev')!=='1'){return;}
    if(t.value==='__ALL__'&&t.checked){
      var others=gcEventTypeFilterRoot.querySelectorAll('input[type=checkbox][data-gc-ev]:not([value="__ALL__"])');
      for(var oi=0;oi<others.length;oi++){others[oi].checked=false;}
      t.checked=true;
    }else{
      normalizeGCEventFilterSelection(gcEventTypeFilterRoot);
    }
    updateGCEventFilterBtnLabel();
    if(lastDashboard){renderGCEventsChart(lastDashboard.gc_events_timeseries);}
  });
}

const gcDurationTypeFilterRoot=document.getElementById('gcDurationTypeFilter');
const gcDurationFilterBtn=document.getElementById('gcDurationFilterBtn');
const gcDurationFilterPanel=document.getElementById('gcDurationFilterPanel');
const gcDurationFilterList=document.getElementById('gcDurationFilterList');
function setGCDurationFilterPanelOpen(open){
  if(!gcDurationFilterPanel||!gcDurationFilterBtn){return;}
  if(open){
    gcDurationFilterPanel.classList.add('open');
    gcDurationFilterBtn.setAttribute('aria-expanded','true');
  }else{
    gcDurationFilterPanel.classList.remove('open');
    gcDurationFilterBtn.setAttribute('aria-expanded','false');
  }
}
function updateGCDurationFilterBtnLabel(){
  if(!gcDurationFilterBtn){return;}
  var mode=getGCDurationChartFilterMode();
  if(mode.mode==='total'){
    gcDurationFilterBtn.textContent='Total';
    return;
  }
  if(mode.types.length===1){
    gcDurationFilterBtn.textContent=mode.types[0];
    return;
  }
  gcDurationFilterBtn.textContent=String(mode.types.length)+' types';
}
function normalizeGCDurationFilterSelection(root){
  if(!root){return;}
  var typeCbs=root.querySelectorAll('input[type=checkbox][data-gc-dur]:not([value="__ALL__"])');
  var allCb=root.querySelector('input[type=checkbox][data-gc-dur][value="__ALL__"]');
  var anyType=false;
  for(var i=0;i<typeCbs.length;i++){if(typeCbs[i].checked){anyType=true;break;}}
  if(anyType){
    if(allCb){allCb.checked=false;}
  }else{
    if(allCb){allCb.checked=true;}
    for(var j=0;j<typeCbs.length;j++){typeCbs[j].checked=false;}
  }
}
function getGCDurationChartFilterMode(){
  if(!gcDurationTypeFilterRoot){return{mode:'total'};}
  var types=[];
  var boxes=gcDurationTypeFilterRoot.querySelectorAll('input[type=checkbox][data-gc-dur]:checked');
  for(var i=0;i<boxes.length;i++){
    var v=boxes[i].value;
    if(v!=='__ALL__'){types.push(v);}
  }
  if(types.length===0){return{mode:'total'};}
  return{mode:'types',types:types};
}
function appendGCDurationFilterRow(list,val,label,checked){
  var lab=document.createElement('label');
  lab.className='gc-event-filter-option';
  var inp=document.createElement('input');
  inp.type='checkbox';
  inp.value=val;
  inp.setAttribute('data-gc-dur','1');
  inp.checked=checked;
  lab.appendChild(inp);
  lab.appendChild(document.createTextNode(' '+label));
  list.appendChild(lab);
}
function setGCDurationCheckboxChecked(container,val,on){
  var boxes=container.querySelectorAll('input[type=checkbox][data-gc-dur]');
  for(var i=0;i<boxes.length;i++){
    if(boxes[i].value===val){boxes[i].checked=on;return;}
  }
}
function updateGCDurationTypeFilter(series){
  if(!gcDurationFilterList||!gcDurationTypeFilterRoot){return;}
  var buckets=series||[];
  var allCats=new Set();
  buckets.forEach(function(b){
    if(b.seconds){Object.keys(b.seconds).forEach(function(c){allCats.add(c);});}
  });
  var cats=Array.from(allCats).sort();
  var prev=[];
  var oldBoxes=gcDurationFilterList.querySelectorAll('input[type=checkbox][data-gc-dur]');
  for(var p=0;p<oldBoxes.length;p++){
    if(oldBoxes[p].checked){prev.push(oldBoxes[p].value);}
  }
  gcDurationFilterList.innerHTML='';
  appendGCDurationFilterRow(gcDurationFilterList,'__ALL__','Total',false);
  cats.forEach(function(c){
    appendGCDurationFilterRow(gcDurationFilterList,c,c,false);
  });
  var optSet={};
  optSet['__ALL__']=true;
  for(var cidx=0;cidx<cats.length;cidx++){optSet[cats[cidx]]=true;}
  var any=false;
  for(var j=0;j<prev.length;j++){
    var v=prev[j];
    if(!optSet[v]){continue;}
    setGCDurationCheckboxChecked(gcDurationFilterList,v,true);
    any=true;
  }
  if(!any){setGCDurationCheckboxChecked(gcDurationFilterList,'__ALL__',true);}
  normalizeGCDurationFilterSelection(gcDurationTypeFilterRoot);
  updateGCDurationFilterBtnLabel();
  refreshGCTypeSearchFilter('gcDurationFilterSearch');
}
if(gcDurationFilterBtn&&gcDurationFilterPanel){
  gcDurationFilterBtn.addEventListener('click',function(e){
    e.stopPropagation();
    var open=!gcDurationFilterPanel.classList.contains('open');
    setGCDurationFilterPanelOpen(open);
  });
  gcDurationFilterPanel.addEventListener('click',function(e){e.stopPropagation();});
}
if(gcDurationFilterList){
  gcDurationFilterList.addEventListener('change',function(e){
    var t=e.target;
    if(!t||t.getAttribute('data-gc-dur')!=='1'){return;}
    if(t.value==='__ALL__'&&t.checked){
      var others=gcDurationTypeFilterRoot.querySelectorAll('input[type=checkbox][data-gc-dur]:not([value="__ALL__"])');
      for(var oi=0;oi<others.length;oi++){others[oi].checked=false;}
      t.checked=true;
    }else{
      normalizeGCDurationFilterSelection(gcDurationTypeFilterRoot);
    }
    updateGCDurationFilterBtnLabel();
    if(lastDashboard){renderGCDurationChart(lastDashboard.gc_duration_timeseries);}
  });
}

document.addEventListener('click',function(e){
  if(!e.target){return;}
  if(gcEventTypeFilterRoot&&!gcEventTypeFilterRoot.contains(e.target)){
    setGCEventFilterPanelOpen(false);
  }
  if(gcDurationTypeFilterRoot&&!gcDurationTypeFilterRoot.contains(e.target)){
    setGCDurationFilterPanelOpen(false);
  }
});
function setupGCTypeFilterSearch(inputId,listId){
  var inp=document.getElementById(inputId);
  var list=document.getElementById(listId);
  if(!inp||!list){return;}
  inp.addEventListener('input',function(){
    var q=String(inp.value||'').trim().toLowerCase();
    var opts=list.querySelectorAll('.gc-event-filter-option');
    for(var i=0;i<opts.length;i++){
      var el=opts[i];
      var text=String(el.textContent||'').toLowerCase().replace(/\s+/g,' ');
      el.style.display=(!q||text.indexOf(q)>=0)?'':'none';
    }
  });
}
setupGCTypeFilterSearch('gcEventFilterSearch','gcEventFilterList');
setupGCTypeFilterSearch('gcDurationFilterSearch','gcDurationFilterList');
function refreshGCTypeSearchFilter(inputId){
  var inp=document.getElementById(inputId);
  if(inp){inp.dispatchEvent(new Event('input',{bubbles:true}));}
}

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
function multiCard(label,stat,unit,fmtFn,colorFn,valueExtraClass,sublabel){
  const s=stat||{};
  const v1=fmtFn?fmtFn(s.last_1m):fmtVal(s.last_1m);
  const v5=fmtFn?fmtFn(s.last_5m):fmtVal(s.last_5m);
  const vh=fmtFn?fmtFn(s.last_1h):fmtVal(s.last_1h);
  const val=(v1==='-'&&v5==='-'&&vh==='-')?'-':(v1+' / '+v5+' / '+vh);
  const color=colorFn?colorFn(s.last_1m):'';
  const extra=(valueExtraClass&&String(valueExtraClass).trim())?(' '+String(valueExtraClass).trim()):'';
  const legend=(sublabel&&String(sublabel).trim())?String(sublabel).trim():'(last-1m / last-5m / last-1h)';
  return '<div class="card"><div class="card-label">'+label+' <span class="card-sublabel">'+escHtml(legend)+'</span></div><div class="card-value '+(color?'text-'+color:'')+extra+'">'+val+'<span class="card-unit">'+unit+'</span></div></div>';
}
function renderCards(f){
  const c=document.getElementById('cards');
  if(!f){c.innerHTML='';return}
  const tp=f.throughput||{};
  const pm=f.pause_max||{};
  const gc=f.gc_rate||{};
  const fgc=f.full_gc_rate||{};
  const alloc=f.alloc_rate||{};
  const heap=f.heap_usage||{};
  c.innerHTML=''+
    multiCard('Throughput',tp,'%',function(v){return v!==undefined?(v*100).toFixed(2):'-'},function(v){return v<0.9?'red':v<0.95?'yellow':'green'})+
    multiCard('Pause Duration',pm,'',function(v){return (v!=null&&v!==undefined&&!isNaN(v)&&v>0)?fmtMs(v):'0.0';},function(v){return v>0.5?'red':v>0.2?'yellow':'green'},null,'(1m-max / 5m-max / 1h-max)')+
    multiCard('Heap Usage',heap,'%',function(v){return v!==undefined?(v*100).toFixed(0):'-'},function(v){return v>0.9?'red':v>0.8?'yellow':'green'})+
    multiCard('Full GCs',fgc,'',null,function(v){return v>0?'red':'green'})+
    multiCard('Alloc Rate',alloc,' MB/s',null,null,'card-value-accent-alloc')+
    multiCard('GC Rate',gc,'',null,null,'card-value-accent-gc');
}
function card(label,value,unit,color){
  const cl=color?'text-'+color:'';
  return '<div class="card"><div class="card-label">'+label+'</div><div class="card-value '+cl+'">'+value+'<span class="card-unit">'+unit+'</span></div></div>';
}

// Heap chart: Heap Used + Heap Total (Java heap); optional Metaspace line (includes PermGen when logged; some GCs omit).
function renderHeapChart(history,path){
  const ctx=document.getElementById('heapCanvas');
  const filtered=(history||[]).filter(function(h){return h.path===path;}).sort(function(a,b){return new Date(a.timestamp)-new Date(b.timestamp);});
  let hasMeta=false;
  filtered.forEach(function(h){
    if(h.meta_used_kb>0)hasMeta=true;
  });
  const heapOpts={responsive:true,maintainAspectRatio:false,animation:false,interaction:{mode:'index',intersect:false},scales:{x:{type:'time',time:{unit:'minute'},grid:{color:'#21262d'},ticks:{color:'#8b949e'}},y:{grid:{color:'#21262d'},ticks:{color:'#8b949e',callback:function(v){return v>=1024?(v/1024).toFixed(1)+'G':v+'M'}}}},plugins:{legend:legendWithIsolate({labels:{color:'#e1e4e8',usePointStyle:true,pointStyle:'circle',font:{size:11}}})}};
  let labels,ds;
  if(!filtered.length){
    labels=defaultMinuteLabelsLastHour();
    const z=labels.map(function(){return 0;});
    ds=[
      {label:'Heap Used',data:z.slice(),borderColor:'#58a6ff',backgroundColor:'#58a6ff22',fill:true,tension:0.3,pointRadius:0},
      {label:'Heap Total',data:z.slice(),borderColor:'#b1bac4',borderDash:[6,4],borderWidth:2,fill:false,tension:0.3,pointRadius:0}
    ];
  }else{
    const t0=new Date(filtered[0].timestamp).getTime();
    const t1=new Date(filtered[filtered.length-1].timestamp).getTime();
    labels=buildMinuteLabelDates(t0,t1);
    const heapUsed=mapLastPerMinute(labels,filtered,function(h){return h.timestamp;},function(h){return h.heap_used_kb/1024;});
    const heapTotal=mapLastPerMinute(labels,filtered,function(h){return h.timestamp;},function(h){return h.heap_total_kb/1024;});
    ds=[
      {label:'Heap Used',data:heapUsed,borderColor:'#58a6ff',backgroundColor:'#58a6ff22',fill:true,tension:0.3,pointRadius:0},
      {label:'Heap Total',data:heapTotal,borderColor:'#b1bac4',borderDash:[6,4],borderWidth:2,fill:false,tension:0.3,pointRadius:0}
    ];
    if(hasMeta){ds.push({label:'Metaspace',data:mapLastPerMinute(labels,filtered,function(h){return h.timestamp;},function(h){return h.meta_used_kb/1024;}),borderColor:'#bc8cff',fill:false,tension:0.3,pointRadius:0,borderWidth:1.5});}
  }
  if(heapChart){
    heapChart.data.labels=labels;heapChart.data.datasets=ds;
    heapChart.update('none');return;
  }
  if(typeof Chart==='undefined')return;
  heapChart=new Chart(ctx,{type:'line',data:{labels:labels,datasets:ds},options:heapOpts});
}

function renderGCEventsChart(series){
  const ctx=document.getElementById('gcCountCanvas');
  const buckets=(series||[]).slice().sort(function(a,b){return new Date(a.timestamp)-new Date(b.timestamp);});
  const allCats=new Set();
  buckets.forEach(function(b){if(b.counts){Object.keys(b.counts).forEach(function(c){allCats.add(c);});}});
  const cats=Array.from(allCats).sort();
  const bucketByMin=new Map();
  buckets.forEach(function(b){
    const k=truncateToMinuteMs(new Date(b.timestamp).getTime());
    var ex=bucketByMin.get(k);
    if(!ex){
      bucketByMin.set(k,{timestamp:b.timestamp,counts:Object.assign({},b.counts||{})});
      return;
    }
    var merged=Object.assign({},ex.counts||{});
    if(b.counts){
      Object.keys(b.counts).forEach(function(ck){
        var v=b.counts[ck];
        var add=typeof v==='number'&&!isNaN(v)?v:0;
        merged[ck]=(merged[ck]||0)+add;
      });
    }
    bucketByMin.set(k,{timestamp:ex.timestamp,counts:merged});
  });
  const filterMode=getGCEventChartFilterMode();
  let finalLabels,finalDatasets,stackedY;
  if(!cats.length){
    finalLabels=defaultMinuteLabelsLastHour();
    const z=finalLabels.map(function(){return 0;});
    finalDatasets=[{label:'Total',data:z,borderColor:'#58a6ff',backgroundColor:'#58a6ff33',fill:true,tension:0.3,pointRadius:0,borderWidth:1.5}];
    stackedY=false;
  }else{
    const t0=new Date(buckets[0].timestamp).getTime();
    const t1=new Date(buckets[buckets.length-1].timestamp).getTime();
    finalLabels=buildMinuteLabelDates(t0,t1);
    if(filterMode.mode==='total'){
      finalDatasets=[{
        label:'Total',
        data:finalLabels.map(function(l){
          const b=bucketByMin.get(l.getTime());
          if(!b||!b.counts){return 0;}
          var s=0;
          Object.keys(b.counts).forEach(function(k){
            var v=b.counts[k];
            if(typeof v==='number'&&!isNaN(v)){s+=v;}
          });
          return s;
        }),
        borderColor:'#58a6ff',
        backgroundColor:'#58a6ff33',
        fill:true,
        tension:0.3,
        pointRadius:0,
        borderWidth:1.5
      }];
      stackedY=false;
    }else{
      var selTypes=filterMode.types;
      finalDatasets=selTypes.map(function(cat,idx){
        var gi=cats.indexOf(cat);
        var ci=gi>=0?gi:idx;
        return{
          label:cat,
          data:finalLabels.map(function(l){
            const b=bucketByMin.get(l.getTime());
            if(!b||!b.counts){return 0;}
            const v=b.counts[cat];
            return v===undefined||v===null?0:v;
          }),
          borderColor:categoryColor(cat,ci),
          backgroundColor:categoryColor(cat,ci)+'33',
          fill:true,
          tension:0.3,
          pointRadius:0,
          borderWidth:1.5
        };
      });
      stackedY=true;
    }
  }
  if(gcCountChart){
    gcCountChart.data.labels=finalLabels;
    gcCountChart.data.datasets=finalDatasets;
    gcCountChart.options.scales.y.stacked=stackedY;
    gcCountChart.update('none');
    return;
  }
  if(typeof Chart==='undefined'){return;}
  const gcOpts={responsive:true,maintainAspectRatio:false,animation:false,interaction:{mode:'index',intersect:false},scales:{x:{type:'time',time:{unit:'minute'},grid:{color:'#21262d'},ticks:{color:'#8b949e'}},y:{beginAtZero:true,stacked:stackedY,grid:{color:'#21262d'},ticks:{color:'#8b949e',precision:0}}},plugins:{legend:legendWithIsolate({position:'top',labels:{color:'#e1e4e8',usePointStyle:true,pointStyle:'circle',font:{size:10},padding:8,boxWidth:10},maxWidth:400})}};
  gcCountChart=new Chart(ctx,{type:'line',data:{labels:finalLabels,datasets:finalDatasets},options:gcOpts});
}

function renderGCDurationChart(series){
  const ctx=document.getElementById('gcDurationCanvas');
  const buckets=(series||[]).slice().sort(function(a,b){return new Date(a.timestamp)-new Date(b.timestamp);});
  const allCats=new Set();
  buckets.forEach(function(b){if(b.seconds){Object.keys(b.seconds).forEach(function(c){allCats.add(c);});}});
  const cats=Array.from(allCats).sort();
  const bucketByMin=new Map();
  buckets.forEach(function(b){
    const k=truncateToMinuteMs(new Date(b.timestamp).getTime());
    var ex=bucketByMin.get(k);
    if(!ex){
      bucketByMin.set(k,{timestamp:b.timestamp,seconds:Object.assign({},b.seconds||{})});
      return;
    }
    var merged=Object.assign({},ex.seconds||{});
    if(b.seconds){
      Object.keys(b.seconds).forEach(function(ck){
        var v=b.seconds[ck];
        var add=typeof v==='number'&&!isNaN(v)?v:0;
        merged[ck]=(merged[ck]||0)+add;
      });
    }
    bucketByMin.set(k,{timestamp:ex.timestamp,seconds:merged});
  });
  const filterMode=getGCDurationChartFilterMode();
  let finalLabels,finalDatasets,stackedY;
  if(!cats.length){
    finalLabels=defaultMinuteLabelsLastHour();
    const z=finalLabels.map(function(){return 0;});
    finalDatasets=[{label:'Total (ms)',data:z,borderColor:'#bc8cff',backgroundColor:'#bc8cff33',fill:true,tension:0.3,pointRadius:0,borderWidth:1.5}];
    stackedY=false;
  }else{
    const t0=new Date(buckets[0].timestamp).getTime();
    const t1=new Date(buckets[buckets.length-1].timestamp).getTime();
    finalLabels=buildMinuteLabelDates(t0,t1);
    if(filterMode.mode==='total'){
      finalDatasets=[{
        label:'Total (ms)',
        data:finalLabels.map(function(l){
          const b=bucketByMin.get(l.getTime());
          if(!b||!b.seconds){return 0;}
          var s=0;
          Object.keys(b.seconds).forEach(function(k){
            var v=b.seconds[k];
            if(typeof v==='number'&&!isNaN(v)){s+=v;}
          });
          return s*1000;
        }),
        borderColor:'#bc8cff',
        backgroundColor:'#bc8cff33',
        fill:true,
        tension:0.3,
        pointRadius:0,
        borderWidth:1.5
      }];
      stackedY=false;
    }else{
      var selTypes=filterMode.types;
      finalDatasets=selTypes.map(function(cat,idx){
        var gi=cats.indexOf(cat);
        var ci=gi>=0?gi:idx;
        return{
          label:cat,
          data:finalLabels.map(function(l){
            const b=bucketByMin.get(l.getTime());
            if(!b||!b.seconds){return 0;}
            const v=b.seconds[cat];
            return v===undefined||v===null||isNaN(v)?0:v*1000;
          }),
          borderColor:categoryColor(cat,ci),
          backgroundColor:categoryColor(cat,ci)+'33',
          fill:true,
          tension:0.3,
          pointRadius:0,
          borderWidth:1.5
        };
      });
      stackedY=true;
    }
  }
  const durOpts={responsive:true,maintainAspectRatio:false,animation:false,interaction:{mode:'index',intersect:false},scales:{x:{type:'time',time:{unit:'minute'},grid:{color:'#21262d'},ticks:{color:'#8b949e'}},y:{beginAtZero:true,stacked:stackedY,grid:{color:'#21262d'},ticks:{color:'#8b949e'}}},plugins:{legend:legendWithIsolate({position:'top',labels:{color:'#e1e4e8',usePointStyle:true,pointStyle:'circle',font:{size:10},padding:8,boxWidth:10},maxWidth:400})}};
  if(gcDurChart){
    gcDurChart.data.labels=finalLabels;
    gcDurChart.data.datasets=finalDatasets;
    gcDurChart.options.scales.y.stacked=stackedY;
    gcDurChart.update('none');
    return;
  }
  if(typeof Chart==='undefined'){return;}
  gcDurChart=new Chart(ctx,{type:'line',data:{labels:finalLabels,datasets:finalDatasets},options:durOpts});
}

function renderAllocationRateChart(history,path){
  const ctx=document.getElementById('allocCanvas');
  const points=(history||[]).filter(function(p){return p.path===path;}).sort(function(a,b){return new Date(a.timestamp)-new Date(b.timestamp);});
  const opts={responsive:true,maintainAspectRatio:false,animation:false,interaction:{mode:'index',intersect:false},scales:{x:{type:'time',time:{unit:'minute'},grid:{color:'#21262d'},ticks:{color:'#8b949e'}},y:{beginAtZero:true,grid:{color:'#21262d'},ticks:{color:'#8b949e'}}},plugins:{legend:{labels:{color:'#e1e4e8'}}}};
  let finalLabels,finalVals;
  if(!points.length){
    finalLabels=defaultMinuteLabelsLastHour();
    finalVals=finalLabels.map(function(){return 0;});
  }else{
    const t0=new Date(points[0].timestamp).getTime();
    const t1=new Date(points[points.length-1].timestamp).getTime();
    finalLabels=buildMinuteLabelDates(t0,t1);
    finalVals=mapLastPerMinute(finalLabels,points,function(p){return p.timestamp;},function(p){return p.rate_mb_per_sec;});
  }
  if(allocChart){
    allocChart.data.labels=finalLabels;allocChart.data.datasets[0].data=finalVals;
    allocChart.update('none');return;
  }
  if(typeof Chart==='undefined')return;
  allocChart=new Chart(ctx,{type:'line',data:{labels:finalLabels,datasets:[{label:'MB/s',data:finalVals,borderColor:'#3fb950',backgroundColor:'#3fb95022',fill:true,tension:0.3,pointRadius:1,pointHoverRadius:4}]},options:opts});
}

function renderThroughputChart(rows,path){
  const ctx=document.getElementById('throughputCanvas');
  const points=(rows||[]).filter(function(p){return p.path===path;}).sort(function(a,b){return new Date(a.timestamp)-new Date(b.timestamp);});
  const opts={responsive:true,maintainAspectRatio:false,animation:false,interaction:{mode:'index',intersect:false},scales:{x:{type:'time',time:{unit:'minute'},grid:{color:'#21262d'},ticks:{color:'#8b949e'}},y:{min:0,max:100,beginAtZero:true,grid:{color:'#21262d'},ticks:{color:'#8b949e',callback:function(v){return v+'%';}}}},plugins:{legend:{labels:{color:'#e1e4e8'}}}};
  let finalLabels,finalVals;
  if(!points.length){
    finalLabels=defaultMinuteLabelsLastHour();
    finalVals=finalLabels.map(function(){return 0;});
  }else{
    finalLabels=points.map(function(p){return new Date(p.timestamp);});
    finalVals=points.map(function(p){
      var r=p.ratio;
      return(typeof r==='number'&&!isNaN(r))?r*100:0;
    });
  }
  if(throughputChart){
    throughputChart.data.labels=finalLabels;
    throughputChart.data.datasets[0].data=finalVals;
    throughputChart.update('none');return;
  }
  if(typeof Chart==='undefined')return;
  throughputChart=new Chart(ctx,{type:'line',data:{labels:finalLabels,datasets:[{label:'Throughput (%)',data:finalVals,borderColor:'#58a6ff',backgroundColor:'#58a6ff22',fill:true,tension:0.3,pointRadius:1,pointHoverRadius:4}]},options:opts});
}

function renderRecommendations(recs){
  const el=document.getElementById('recsContainer');
  if(!recs||!recs.length){el.innerHTML='<div class="recs-empty">Nothing.</div>';return}
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
  updateGCEventTypeFilter(d.gc_events_timeseries);
  renderGCEventsChart(d.gc_events_timeseries);
  updateGCDurationTypeFilter(d.gc_duration_timeseries);
  renderGCDurationChart(d.gc_duration_timeseries);
  renderThroughputChart(d.throughput_history,selectedPath);
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
setInterval(refresh,5000);
</script>
</body>
</html>`
