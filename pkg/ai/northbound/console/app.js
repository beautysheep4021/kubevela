(function () {
  'use strict';
  const S = ConsoleState, E = S.escape, page = document.getElementById('page');
  const identity = {namespace:document.body.dataset.accountNamespace, tenant:document.body.dataset.accountTenant, username:document.body.dataset.accountUsername};
  const demo = new URLSearchParams(location.search).get('demo') === 'local';
  const modal = document.getElementById('modal');
  const titles = {overview:'工作台',jobs:'任务',services:'在线服务',models:'模型',datasets:'数据集'};
  const purposes = {training:'训练',evaluation:'评测',batch:'批处理','offline-inference':'离线推理',service:'模型服务',agent:'智能体服务'};
  let route = S.route(location.hash), routeVersion = 0, requestVersion = 0, timer, failures = 0;
  let items = [], detail = null, updated = '', dirty = false, draft = null, step = 0, draftError = '', submitMessage = '', submitting = false;
  let modelOptions = [], datasetOptions = [], assetError = '', assetLoading = false, result = null, modalAction = null, lastHash = '', uncertain = false;
  let returningList = {jobs:'#/jobs',services:'#/services'}, modalBusy = false, logVersion = 0, pendingNavigation = '';
  const client = ConsoleAPI.create({...identity,demo,onUnauthorized:expired});
  const request = (path, init) => client.request(path, init);
  const apiRoot = '/api/v1/ai';
  const listPath = entity => `${apiRoot}/${entity}?namespace=${encodeURIComponent(identity.namespace)}`;
  const objectPath = (name = route.name, namespace = route.namespace) => `${apiRoot}/applications/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}`;
  const detailHash = (type, namespace, name) => `#/${type === 'AIJob' ? 'jobs' : 'services'}/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}`;
  const icon = name => `<i data-lucide="${name}" aria-hidden="true"></i>`;
  const button = (action, label, symbol, primary = false) => `<button type="button" class="button${primary ? ' primary' : ''}" data-action="${action}">${symbol ? icon(symbol) : ''}${E(label)}</button>`;
  const iconButton = (action, label, symbol) => `<button type="button" class="icon-button" data-action="${action}" title="${E(label)}" aria-label="${E(label)}">${icon(symbol)}</button>`;
  const badge = (item, type) => { const s = S.status(item,type || S.kind(item)); return `<span class="badge ${s.tone}">${E(s.label)}</span>`; };
  const sample = item => demo || S.example(item) ? '<span class="badge warn">示例</span>' : '';
  const time = value => value && !Number.isNaN(Date.parse(value)) ? new Date(value).toLocaleString('zh-CN',{hour12:false}) : '未提供';
  const json = value => `<pre class="config-output">${E(JSON.stringify(value,null,2))}</pre>`;
  function icons() { if (globalThis.lucide) lucide.createIcons(); }
  function toast(message) { const el = document.getElementById('toast'); el.textContent = message; el.hidden = false; clearTimeout(toast.timer); toast.timer = setTimeout(() => {el.hidden = true;},5000); }
  function expired() {
    clearTimeout(timer); routeVersion++; items=[]; detail=null; draft=null; dirty=false; modelOptions=[]; datasetOptions=[];
    if (modal.open) modal.close();
    page.innerHTML='<div class="empty-state"><h1>会话已失效或账号已切换</h1><a class="button primary" href="/login'+(demo?'?demo=local':'')+'">重新登录</a></div>';
  }
  function errorText(err) {
    if (err.status === 403) return '当前账号无权访问此资源。';
    if (err.status === 404) return '未找到该资源，可能已被删除。';
    if (err.status === 503) return '当前服务未连接所需的集群或资产存储。';
    return err.message || '请求失败，请重试。';
  }
  function heading(title, actions='') { return `<div class="page-heading"><div><div class="eyebrow">${E(identity.tenant)} / ${E(titles[route.section])}</div><h1>${E(title)}</h1></div><div class="actions">${actions}</div></div>`; }
  function empty(title, action='') { return `<div class="empty-state">${icon('inbox')}<h2>${E(title)}</h2>${action}</div>`; }
  function notice(message, retry='refresh') { return `<div class="notice error" role="alert"><span>${E(message)}</span>${retry ? button(retry,'重试','refresh-cw') : ''}</div>`; }
  function purpose(item) { const m = item.aiMetadata || {}; return item.purpose || item.jobKind || m['ai.oam.dev/purpose'] || m['ai.oam.dev/job-kind'] || m['ai.oam.dev/task-type'] || (S.kind(item)==='AIService' ? 'service' : ''); }
  function workTable(rows, fixedType) {
    if (!rows.length) return empty(route.query || route.filter || route.purpose ? '没有符合筛选条件的记录' : '暂无记录',button(fixedType === 'AIService' ? 'new-service' : 'new-job',fixedType === 'AIService' ? '部署服务' : '创建任务','plus',true));
    return `<div class="table-wrap"><table><thead><tr><th>名称</th><th>用途</th><th>状态</th><th>创建时间</th><th>操作</th></tr></thead><tbody>${rows.map(item=>{
      const type = fixedType || S.kind(item), name = E(item.name);
      return `<tr><td><a class="name-link" href="${detailHash(type,item.namespace,item.name)}">${name}</a><div class="subtle">${E(type)} ${sample(item)}</div></td><td>${E(purposes[purpose(item)] || purpose(item) || '未提供')}</td><td>${badge(item,type)}</td><td>${E(time(item.createdAt))}</td><td><a class="icon-button" href="${detailHash(type,item.namespace,item.name)}" title="查看 ${name}" aria-label="查看 ${name}">${icon('arrow-up-right')}</a></td></tr>`;
    }).join('')}</tbody></table></div>`;
  }
  function filters() {
    return `<div class="filters"><label><span class="sr-only">搜索名称</span><input id="search" type="search" placeholder="搜索名称" value="${E(route.query)}"></label><label><span class="sr-only">筛选状态</span><select id="status-filter">${[['','全部状态'],['running','运行中'],['ready','已就绪'],['pending','等待 / 部署中'],['succeeded','已完成'],['failed','失败'],['unknown','状态待确认']].map(([v,l])=>`<option value="${v}"${route.filter===v?' selected':''}>${l}</option>`).join('')}</select></label>${route.section==='jobs'?`<label><span class="sr-only">筛选用途</span><select id="purpose-filter"><option value="">全部用途</option>${['training','evaluation','batch','offline-inference'].map(v=>`<option value="${v}"${route.purpose===v?' selected':''}>${purposes[v]}</option>`).join('')}</select></label>`:''}<span class="updated">${updated ? '更新于 '+E(updated) : ''}</span>${iconButton('refresh','刷新列表','refresh-cw')}</div>`;
  }
  function filtered(rows) {return rows.filter(i=>(i.name || '').toLowerCase().includes(route.query.toLowerCase()) && (!route.filter || S.status(i,S.kind(i)).key===route.filter) && (!route.purpose || purpose(i)===route.purpose));}
  function overview() {
    const jobs=items.filter(i=>S.kind(i)==='AIJob'), services=items.filter(i=>S.kind(i)==='AIService');
    const failed=items.filter(i=>S.status(i,S.kind(i)).key==='failed');
    return `<div class="summary-strip">${[[jobs.length,'任务总数'],[jobs.filter(i=>S.status(i,'AIJob').key==='running').length,'运行中任务'],[services.filter(i=>S.status(i,'AIService').key==='ready').length,'就绪服务'],[failed.length,'失败项']].map(([n,l])=>`<div class="summary-item"><strong>${n}</strong><span>${l}</span></div>`).join('')}</div>${failed.length?`<div class="section-head"><h2>需要处理</h2></div>${workTable(failed)}`:''}<div class="section-head"><h2>最近任务</h2><a href="#/jobs">全部任务 ${icon('arrow-right')}</a></div>${workTable(jobs.slice(0,5),'AIJob')}<div class="section-head"><h2>在线服务</h2><a href="#/services">全部服务 ${icon('arrow-right')}</a></div>${workTable(services.slice(0,5),'AIService')}`;
  }
  function assetTable() {
    const models = route.section==='models';
    if (!items.length) return empty(models?'暂无模型资产':'暂无数据集',button('register',models?'登记模型':'登记数据集','plus',true));
    return `<div class="table-wrap"><table><thead><tr><th>名称</th><th>${models?'版本 / 评测':'格式 / 状态'}</th><th>来源</th><th>操作</th></tr></thead><tbody>${items.map((a,i)=>`<tr><td><strong>${E(a.displayName||a.name)}</strong> ${sample(a)}<div class="subtle">${E(a.namespace)}</div></td><td>${E(models ? a.version||'未提供' : a.format||'未提供')}<div class="subtle">${E(models ? ({passed:'评测通过',failed:'评测未通过',pending:'待评测'}[a.evaluationStatus]||'暂无评测结果') : a.status||'registered')}</div></td><td>${a.jobName?`<a href="${detailHash('AIJob',a.namespace,a.jobName)}">${E(a.jobName)}</a>`:E(a.source||'手动登记')}</td><td><div class="actions">${button('asset-train','用于训练','play')}${button('asset-eval','评测','flask-conical')}${models?button('asset-deploy','部署','rocket'):''}${iconButton('asset-detail','查看资产详情','info')}</div><span hidden data-index="${i}"></span></td></tr>`).join('')}</tbody></table></div>`;
  }
  function renderCollection() {
    const collection=document.getElementById('collection');
    if(collection&&['jobs','services'].includes(route.section)){
      collection.innerHTML=workTable(filtered(items),route.section==='jobs'?'AIJob':'AIService');
      document.getElementById('page-error').innerHTML='';icons();return;
    }
    const actions=route.section==='overview'?button('new-job','创建任务','plus',true)+button('new-service','部署服务','rocket') : route.section==='jobs'?button('new-job','创建任务','plus',true) : route.section==='services'?button('new-service','部署服务','plus',true):button('register',route.section==='models'?'登记模型':'登记数据集','plus',true);
    page.innerHTML=heading(titles[route.section],actions)+'<div id="page-error"></div>'+(route.section==='overview'?`<div class="section-head"><span class="subtle">${updated?'更新于 '+E(updated):''}</span>${iconButton('refresh','刷新工作台','refresh-cw')}</div>`: ['jobs','services'].includes(route.section)?filters():'')+'<section id="collection"></section>';
    document.getElementById('collection').innerHTML=route.section==='overview'?overview():['models','datasets'].includes(route.section)?assetTable():workTable(filtered(items),route.section==='jobs'?'AIJob':'AIService');
    icons();
  }
  async function refresh(initial=false) {
    if (route.create || document.hidden) return;
    clearTimeout(timer);
    const version=routeVersion, seq=++requestVersion;
    if (initial) page.innerHTML=heading(route.detail?route.name:titles[route.section])+'<div class="loading" role="status">正在读取…</div>';
    try {
      if (route.detail) {
        if (route.namespace !== identity.namespace) throw Object.assign(new Error('当前账号无权访问此资源。'),{status:403});
        const loaded=await request(objectPath()+'/status');
        if(version!==routeVersion || seq!==requestVersion)return;
        detail=loaded;
        const expected=route.section==='jobs'?'AIJob':'AIService';
        if(S.kind(detail) && S.kind(detail)!==expected) throw new Error('资源类型与当前页面不匹配。');
        renderDetail();
      } else {
        const entity=['models','datasets'].includes(route.section)?route.section:'applications';
        const data=await request(listPath(entity));
        if(version!==routeVersion || seq!==requestVersion)return;
        items=(data.items || []).filter(i=>i.namespace===identity.namespace);
        if(route.section==='jobs')items=items.filter(i=>S.kind(i)==='AIJob');
        if(route.section==='services')items=items.filter(i=>S.kind(i)==='AIService');
        items.sort((a,b)=>(b.createdAt||'').localeCompare(a.createdAt||'')||(a.name||'').localeCompare(b.name||''));
        updated=new Date().toLocaleTimeString('zh-CN',{hour12:false});renderCollection();
      }
      failures=0;
    } catch(err) {
      if(version!==routeVersion || seq!==requestVersion || err.status===401)return;
      failures++;
      if(err.status===403 || err.status===404){items=[];detail=null;page.innerHTML=heading(route.detail?route.name:titles[route.section])+'<div id="page-error"></div>';}
      if(initial || !document.getElementById('page-error'))page.innerHTML=heading(route.detail?route.name:titles[route.section])+'<div id="page-error"></div>';
      document.getElementById('page-error').innerHTML=notice(errorText(err));icons();
    }
    if(version===routeVersion && (!route.detail || !detail || !S.status(detail,route.section==='jobs'?'AIJob':'AIService').terminal) && failures<4 && (!route.detail || route.tab==='overview')) timer=setTimeout(()=>refresh(),Math.min(10000*2**failures,60000));
  }
  function definition(entries) {return `<dl class="detail-meta">${entries.map(([k,v])=>`<div><dt>${E(k)}</dt><dd>${E(v || '未提供')}</dd></div>`).join('')}</dl>`;}
  function detailOverview() {
    const components=detail.components||[], meta=detail.aiMetadata||{};
    const events=components.flatMap(c=>(c.pods||[]).flatMap(p=>(p.events||[]).map(e=>({...e,pod:p.name}))));
    return definition([['名称',detail.name],['所属空间',detail.namespace],['用途',purposes[purpose(detail)]||purpose(detail)],['运行时',meta['ai.oam.dev/runtime']],['状态说明',detail.message]])+
      (detail.diagnostics||[]).filter(d=>d.reason!=='Completed').map(d=>`<div class="notice ${d.severity==='error'?'error':''}"><strong>${E(d.reason)}</strong><span>${E(d.message)}</span></div>`).join('')+
      (components.length?`<div class="section-head"><h2>运行实例</h2></div><div class="table-wrap"><table><thead><tr><th>组件</th><th>工作负载</th><th>结果 / 副本</th></tr></thead><tbody>${components.map(c=>`<tr><td>${E(c.name)}</td><td>${E(c.workloadKind||c.type)}</td><td>${c.workload && c.workload.completed===true?'已完成':c.workload && c.workload.kind==='Deployment'?`${c.workload.readyReplicas||0} / ${c.workload.desiredReplicas||0} 就绪`:E(c.message||'未提供')}</td></tr>`).join('')}</tbody></table></div>`:'')+
      (events.length?`<div class="section-head"><h2>事件</h2></div>${events.map(e=>`<div class="notice"><strong>${E(e.reason)}</strong><span>${E(e.message)}</span><small>${E(e.pod)}</small></div>`).join('')}`:'');
  }
  function renderDetail() {
    const job=route.section==='jobs';
    const actions=button(job?'rerun':'restart',job?'重新运行':'重启服务',job?'rotate-ccw':'rotate-cw')+iconButton('delete','删除'+(job?'任务':'服务'),'trash-2');
    page.innerHTML=`<a class="back-link" href="${returningList[route.section]}">${icon('arrow-left')} 返回${titles[route.section]}</a>`+heading(route.name,actions)+`<div class="actions detail-status">${badge(detail,job?'AIJob':'AIService')}${sample(detail)}<span class="subtle">${E(route.namespace)}</span>${iconButton('refresh','刷新详情','refresh-cw')}</div><div id="page-error"></div><div class="tabs" role="tablist" aria-label="详情标签">${[['overview','概览'],['logs','日志'],['result',job?'结果与产物':'服务访问'],['config','配置']].map(([v,l])=>`<button type="button" role="tab" aria-selected="${route.tab===v}" data-tab="${v}">${l}</button>`).join('')}</div><section class="detail-body" id="detail-content"></section>`;
    const body=document.getElementById('detail-content');
    if(route.tab==='overview')body.innerHTML=detailOverview();
    if(route.tab==='config')body.innerHTML='<div class="section-head"><h2>已观测配置</h2></div>'+json({aiMetadata:detail.aiMetadata,components:detail.components});
    if(route.tab==='logs'){body.innerHTML='<div class="filters"><label>Pod<select id="log-pod"><option value="">自动选择</option></select></label><label>容器<select id="log-container"><option value="">自动选择</option></select></label><label>日志行数<input id="log-tail" type="number" min="1" max="2000" value="200"></label>'+button('logs','刷新日志','refresh-cw')+'</div><div id="log-error"></div><pre id="log-output" class="log-output" role="status">正在读取日志…</pre>';loadLogs();}
    if(route.tab==='result') {
      body.innerHTML=job?`<div class="section-head"><h2>任务产物</h2>${button('parse-result','读取运行结果','refresh-cw')}</div><div id="result-output">${empty('尚未读取运行结果')}</div>`:`<div class="section-head"><h2>访问检查</h2>${button('probe','检查服务访问','activity',true)}</div><div class="notice">检查服务端点的连通性与响应，不代表模型推理正确性。</div>${definition((detail.components||[]).filter(c=>c.service&&c.service.name).map(c=>['集群内服务',c.service.name+' · '+(c.service.clusterIP||'')]))}<div id="result-output"></div>`;
      if(job && result)renderResult();
    }
    icons();
  }
  async function loadLogs() {
    const seq=++logVersion;
    const v=routeVersion, tab=route.tab, pod=document.getElementById('log-pod').value, container=document.getElementById('log-container').value;
    const tail=document.getElementById('log-tail').value;
    if(!/^\d+$/.test(tail)||Number(tail)<1||Number(tail)>2000){document.getElementById('log-error').innerHTML=notice('日志行数应为 1 到 2000。','');return;}
    try {
      const logs=await request(objectPath()+'/logs?'+new URLSearchParams({pod,container,tailLines:tail}));
      if(v!==routeVersion || route.tab!==tab || seq!==logVersion)return;
      document.getElementById('log-output').textContent=logs.logs || '暂无日志。';document.getElementById('log-error').innerHTML='';
      const pods=logs.pods||[];
      document.getElementById('log-pod').innerHTML='<option value="">自动选择</option>'+pods.map(p=>`<option value="${E(p.name)}"${p.name===pod?' selected':''}>${E(p.name)}</option>`).join('');
      const chosen=pods.find(p=>p.name===(pod||logs.pod));
      document.getElementById('log-container').innerHTML='<option value="">自动选择</option>'+((chosen||{}).containers||[]).map(c=>`<option${c===container?' selected':''}>${E(c)}</option>`).join('');
    }catch(err){if(v===routeVersion && seq===logVersion && document.getElementById('log-error')){if(err.status===403)document.getElementById('log-output').textContent='';document.getElementById('log-error').innerHTML=notice(errorText(err),'logs');icons();}}
  }
  async function detailAction(action) {
    const v=routeVersion, target=objectPath(), name=route.name, namespace=route.namespace;
    try {
      if(action==='parse-result') {
        const data=await request(`${apiRoot}/deliveries/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}/result`);
        if(v!==routeVersion)return;result=data.result;renderResult();
      } else if(action==='probe') {
        const data=await request(target+'/probe',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({path:'/healthz',timeoutSeconds:10})});
        if(v!==routeVersion)return;
        document.getElementById('result-output').innerHTML=`<div class="notice ${data.healthy?'success':'error'}">${data.healthy?'访问检查通过':'访问检查未通过'} · HTTP ${E(data.statusCode || '未返回')}</div>`+definition([['目标',data.url],['错误',data.error]])+`<pre class="log-output">${E(data.body||'无响应内容')}</pre>`;
      }
      icons();
    }catch(err){if(v===routeVersion){document.getElementById('page-error').innerHTML=notice(errorText(err),action);icons();}}
  }
  function renderResult() {
    if(!result || !document.getElementById('result-output'))return;
    document.getElementById('result-output').innerHTML=`${sample(detail)}${definition([['产物',result.modelURI],['摘要',result.summary]])}${result.metrics?`<div class="table-wrap"><table><thead><tr><th>结果指标</th><th>值</th></tr></thead><tbody>${Object.entries(result.metrics).map(([k,v])=>`<tr><td>${E(k)}</td><td>${E(v)}</td></tr>`).join('')}</tbody></table></div>`:''}<div class="actions">${button('register-result','登记模型','box',true)}${button('publish-result','发布为服务','rocket')}</div>`;
    icons();
  }
  function showModal(title, body, action, submitLabel='确认') {
    modalAction=action;modalBusy=false;
    modal.innerHTML=`<form id="modal-form"><div class="dialog-head"><h2 id="modal-title">${E(title)}</h2>${iconButton('close-modal','关闭','x')}</div><div class="dialog-body">${body}<div id="modal-error" role="alert"></div></div><div class="dialog-actions">${button('close-modal','取消')}<button class="button primary" type="submit">${E(submitLabel)}</button></div></form>`;
    modal.showModal();icons();
  }
  function modalField(name,label,value='',type='text') {return `<label class="field">${E(label)}<input name="${name}" type="${type}" value="${E(value)}" required></label>`;}
  function registerAsset(models, supplied) {
    showModal(models?'登记模型':'登记数据集',`<div class="form-grid">${modalField('name','名称',supplied?route.name+'-model':'')}${models?modalField('version','版本','v1'):modalField('format','数据格式','sharegpt-jsonl')}${modalField('uri',models?'模型存储引用':'数据存储引用',supplied?supplied.modelURI:'')}</div>`,async form=>{
      const name=form.get('name').trim();if(!name)throw new Error('请输入名称。');
      const payload={namespace:identity.namespace,name,owner:identity.username,visibility:'private',status:'registered'};
      if(models)Object.assign(payload,{modelURI:form.get('uri'),version:form.get('version'),jobName:supplied?route.name:'',metrics:supplied?supplied.metrics:undefined,summary:supplied?supplied.summary:undefined});
      else Object.assign(payload,{datasetURI:form.get('uri'),format:form.get('format'),displayName:name});
      await request(apiRoot+(models?'/models':'/datasets'),{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)});
      toast('资产已登记');navigate(models?'#/models':'#/datasets');
    },'登记');
  }
  function lifecycle(action) {
    const label={delete:'删除',rerun:'重新运行',restart:'重启服务'}[action], path=objectPath(), name=route.name, section=route.section;
    showModal(label,`<p>${E(label)} <strong>${E(name)}</strong>？</p>`,async()=>{
      await request(path+(action==='delete'?'':'/'+action),{method:action==='delete'?'DELETE':'POST'});
      toast(label+'已提交');if(action==='delete')navigate('#/'+section);else await refresh();
    },label);
  }
  function publishResult() {
    const ns=route.namespace,name=route.name;
    showModal('发布示例服务',`<div class="notice">当前交付接口发布示例 HTTP 运行时，用于验证产物交付链路。</div>${modalField('serviceName','服务名称',name+'-service')}`,async form=>{
      const existing=await request(listPath('applications'));
      if((existing.items||[]).some(item=>item.namespace===ns&&item.name===form.get('serviceName')))throw new Error('同名任务或服务已存在，请更换名称。');
      const data=await request(`${apiRoot}/deliveries/${encodeURIComponent(ns)}/${encodeURIComponent(name)}/publish-service`,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({serviceName:form.get('serviceName'),image:'python:3.11-slim',port:8000,servicePort:80})});
      toast('服务已提交');navigate(detailHash('AIService',ns,data.serviceName));
    },'发布服务');
  }
  function assetDetail(asset) {
    const models=route.section==='models';
    showModal('资产详情',json(asset)+(models?modalField('evaluationJobName','评测任务名称',''):''),async form=>{
      if(!models)return;
      const summary=await request(objectPath(form.get('evaluationJobName'),identity.namespace)+'/status');
      if(purpose(summary)!=='evaluation')throw new Error('请选择用途为评测的任务，不能使用训练任务的结果。');
      await request(`${apiRoot}/models/${encodeURIComponent(identity.namespace)}/${encodeURIComponent(asset.name)}/sync-evaluation`,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({evaluationJobName:form.get('evaluationJobName')})});
      toast('评测结果已同步');await refresh();
    },models?'同步评测结果':'关闭');
  }
  function navigate(hash,force=false) {
    if(modalBusy){pendingNavigation=hash;return;}
    if(submitting && !force){toast('正在提交，请稍候。');return;}
    if(!force && dirty && hash!==location.hash && !confirm('离开将放弃未提交的配置，是否继续？'))return;
    if(dirty){draft=null;dirty=false;}
    if(location.hash===hash)loadRoute();else location.hash=hash;
  }
  function loadRoute() {
    const hash=location.hash || '#/overview';
    if(submitting || modalBusy){history.replaceState(null,'',lastHash);return;}
    if(dirty && lastHash && hash!==lastHash) {
      if(!confirm('离开将放弃未提交的配置，是否继续？')){history.replaceState(null,'',lastHash);return;}
      dirty=false;draft=null;
    }
    routeVersion++;clearTimeout(timer);failures=0;route=S.route(hash);lastHash=hash;items=[];detail=null;result=null;updated='';
    if(modal.open&&!modalBusy)modal.close();
    document.body.classList.remove('nav-open');document.getElementById('nav-toggle').setAttribute('aria-expanded','false');
    document.querySelectorAll('[data-nav]').forEach(el=>{if(el.dataset.nav===route.section)el.setAttribute('aria-current','page');else el.removeAttribute('aria-current');});
    if(!identity.namespace){page.innerHTML=heading('工作空间未绑定')+notice('当前账号尚未绑定命名空间，请联系管理员。','');icons();return;}
    if(route.invalid){page.innerHTML=heading('页面不存在')+empty('无法识别此页面','<a class="button" href="#/overview">返回工作台</a>');icons();return;}
    if(!route.detail&&!route.create&&['jobs','services'].includes(route.section))returningList[route.section]=hash;
    if(route.create){if(!draft)draft=ConsoleCreate.defaults(route.section==='jobs'?'AIJob':'AIService',identity.namespace,identity.tenant);step=0;draftError='';submitMessage='';uncertain=false;renderCreate();loadAssetOptions();}
    else refresh(true);
    page.focus({preventScroll:true});
  }
  window.addEventListener('hashchange',loadRoute);
  window.addEventListener('beforeunload',event=>{if(dirty||submitting){event.preventDefault();event.returnValue='';}});
  document.addEventListener('visibilitychange',()=>{if(document.hidden)clearTimeout(timer);else request(apiRoot+'/session').then(s=>{if(s.namespace!==identity.namespace || s.username!==identity.username)expired();else if(!route.create)refresh();}).catch(err=>{if(err.status===401)expired();else toast('暂时无法检查连接，当前输入已保留。');});});
  document.getElementById('nav-toggle').onclick=()=>{document.body.classList.toggle('nav-open');document.getElementById('nav-toggle').setAttribute('aria-expanded',String(document.body.classList.contains('nav-open')));};
  document.getElementById('environment').textContent=demo?'演示环境':'集群模式';
  document.getElementById('environment').classList.add(demo?'warn':'neutral');
  document.getElementById('logout-form').addEventListener('submit',event=>{if(dirty&&!confirm('退出将放弃未提交的配置，是否继续？')){event.preventDefault();return;}dirty=false;draft=null;clearTimeout(timer);if(demo)event.target.action='/logout?demo=local';});

  // Creation drafts never provide the scope used by lists or selected objects.
  function field(name,label,type='text',options) {
    const value=draft[name] == null ? '' : draft[name];
    const control=options?`<select id="draft-${name}" data-field="${name}">${options.map(([v,l])=>`<option value="${E(v)}"${String(value)===String(v)?' selected':''}>${E(l)}</option>`).join('')}</select>`:type==='textarea'?`<textarea id="draft-${name}" data-field="${name}" rows="5">${E(value)}</textarea>`:`<input id="draft-${name}" data-field="${name}" type="${type}" value="${E(value)}"${type==='number'?' min="0" step="any"':''}>`;
    return `<label class="field${type==='textarea'?' wide':''}" for="draft-${name}">${E(label)}${control}</label>`;
  }
  function chosenOptions(list,fieldName,labelKey) {
    const values=list.map(a=>[a[fieldName],(a.displayName||a.name)+(a.version?' · '+a.version:'')]);
    if(draft[fieldName]&&!values.some(([v])=>v===draft[fieldName]))values.unshift([draft[fieldName],'当前引用（未在资产列表中确认）']);
    return [['','请选择'+labelKey]].concat(values);
  }
  function createInputs() {
    const job=draft.kind==='AIJob';
    const needsModel=job?draft.purpose!=='batch':draft.purpose==='service';
    return `<div class="form-grid">${field('name',job?'任务名称':'服务名称')}${field('example','运行方式','text',draft.purpose==='agent'||draft.purpose==='batch'||draft.purpose==='offline-inference'?[['false','自定义容器']]:[['true','示例运行时'],['false','自定义容器']])}</div>${draft.example?'<div class="notice">示例运行时用于验证纳管流程，不执行真实训练、评测或模型推理。</div>':''}${assetLoading?'<div class="loading" role="status">正在读取资产…</div>':''}${assetError?notice(assetError,'reload-assets'):''}<div class="form-grid">${needsModel?field('modelURI',job?'输入模型':'模型版本','text',chosenOptions(modelOptions,'modelURI','模型')):''}${job&&draft.purpose!=='batch'?field('datasetURI','输入数据集','text',chosenOptions(datasetOptions,'datasetURI','数据集')):''}${field('image','容器镜像')}${field('runtime','运行时','text',[['batch','Batch'],['http','HTTP'],['pytorch','PyTorch'],['vllm','vLLM'],['triton','Triton'],['custom','自定义']])}${!draft.example?field('command','启动命令','textarea'):''}${job&&draft.purpose==='training'?field('epochs','Epoch','number')+field('learningRate','学习率'):''}${job&&draft.purpose==='evaluation'?field('threshold','通过阈值（0–1）','number'):''}</div>${needsModel?`<div class="actions">${button('register-create-model','登记模型','plus')}${job?button('register-create-dataset','登记数据集','plus'):''}</div>`:''}`;
  }
  function renderCreate() {
    const job=draft.kind==='AIJob', steps=['选择用途','运行配置','资源配置','确认提交'];
    page.innerHTML=heading(job?'创建任务':'部署服务',button('cancel-create','取消'))+`<div class="steps" aria-label="创建步骤">${steps.map((label,i)=>`<button type="button" data-step="${i}" class="${i===step?'active':i<step?'done':''}"${i>step?' disabled':''} aria-current="${i===step?'step':'false'}"><span>${i+1}</span>${label}</button>`).join('')}</div><section class="form-layout"><div class="form-section"><h2>${steps[step]}</h2><div id="draft-errors" role="alert">${draftError?notice(draftError,''):''}</div><div id="create-fields"></div></div></section>`;
    let html='';
    if(step===0)html=`<div class="form-grid">${field('purpose','任务用途','text',(job?['training','evaluation','batch','offline-inference']:['service','agent']).map(v=>[v,purposes[v]]))}</div>${definition([['底层类型',draft.kind],['所属租户',identity.tenant],['命名空间',identity.namespace]])}`;
    if(step===1)html=createInputs();
    if(step===2)html=`<div class="form-grid"><label class="field">资源规格<select id="resource-preset"><option value="custom">自定义</option><option value="cpu">CPU · 1 核 / 1 GiB</option><option value="gpu">GPU · 8 核 / 32 GiB / 1 卡</option></select></label>${field('cpu','CPU 核数')}${field('memory','内存')}${field('gpu','GPU 数量','number')}${job?'':field('replicas','副本数','number')+field('port','容器端口','number')}${field('priority','调度优先级','text',[['normal','普通'],['high','高'],['urgent','紧急']])}${field('nodeType','节点类型','text',[['cpu','CPU 节点'],['gpu','GPU 节点'],['any','不限']])}${field('strategy','调度策略','text',[['performance','性能优先'],['cost','成本优先'],['fast-start','快速启动']])}</div><div class="notice">配额由管理员管理；当前页面不提供剩余额度预估。调度策略是否生效取决于集群配置。</div>`;
    if(step===3) {
      let config='';try{config=ConsoleCreate.generate(draft);}catch(err){draftError=err.message;}
      html=definition([['名称',draft.name],['用途',purposes[draft.purpose]],['命名空间',identity.namespace],['运行镜像',draft.image],['资源申请',`${draft.cpu} CPU / ${draft.memory} / ${draft.gpu} GPU`],['运行方式',draft.example?'示例运行时':'自定义容器']])+`${draft.example?'<div class="notice">本次使用示例运行时，结果不代表真实算法运行。</div>':''}<label class="field">执行方式<select id="execution-mode"><option value="dry"${draft.dryRun?' selected':''}>仅校验配置</option><option value="create"${!draft.dryRun?' selected':''}>实际创建${demo?'（演示环境）':''}</option></select></label><details><summary>查看生成配置（只读）</summary><pre class="config-output">${E(config)}</pre></details><div id="submit-result" role="status">${submitMessage?`<div class="notice ${uncertain?'error':'success'}">${E(submitMessage)}</div>`:''}</div>${uncertain?button('check-created','查询创建结果','search'):''}`;
    }
    document.getElementById('create-fields').innerHTML=html;
    page.insertAdjacentHTML('beforeend',`<div class="wizard-footer"><span class="subtle">${step+1} / 4</span><div class="actions">${step>0?button('previous','上一步','arrow-left'):''}${button(step===3?'submit':'next',step===3?(submitting?'提交中…':draft.dryRun?'校验配置':job?'创建任务':'部署服务'):'下一步',step===3?'check':'arrow-right',true)}</div></div>`);
    if(submitting || uncertain)page.querySelectorAll('[data-field],#execution-mode,[data-action="submit"],[data-action="previous"],[data-step]').forEach(el=>el.disabled=true);
    icons();
  }
  async function loadAssetOptions() {
    const v=routeVersion;assetLoading=true;assetError='';
    if(route.create)renderCreate();
    const responses=await Promise.allSettled([request(listPath('models')),request(listPath('datasets'))]);
    if(v!==routeVersion || !route.create)return;
    assetLoading=false;
    const [models,datasets]=responses;
    modelOptions=models.status==='fulfilled'?(models.value.items||[]).filter(a=>a.namespace===identity.namespace):[];
    datasetOptions=datasets.status==='fulfilled'?(datasets.value.items||[]).filter(a=>a.namespace===identity.namespace):[];
    const failed=responses.find(r=>r.status==='rejected');assetError=failed?errorText(failed.reason):'';
    // Defaults select visible assets only; custom references are never fabricated.
    if(draft.modelURI==='inline://models/base' || !draft.modelURI){draft.modelURI=modelOptions[0]?.modelURI||'';draft.modelName=modelOptions[0]?.name||draft.modelName;}
    if(draft.datasetURI==='inline://datasets/customer-sft-demo' || !draft.datasetURI)draft.datasetURI=datasetOptions[0]?.datasetURI||'';
    renderCreate();
  }
  const validationMessages={name:'名称需为 1–63 位小写字母、数字或连字符，首尾不能是连字符。',namespace:'当前账号未绑定有效命名空间。',tenant:'当前账号未绑定有效租户。',image:'请填写有效的容器镜像。',command:'请填写自定义容器的启动命令。',modelURI:'请选择模型。',datasetURI:'请选择数据集。',cpu:'CPU 应为正数，例如 1、0.5 或 500m。',memory:'请填写有效内存，例如 512Mi 或 1Gi。',gpu:'GPU 数量应为非负整数。',replicas:'副本数应为正整数。',port:'端口应为 1–65535 的整数。',epochs:'Epoch 应为正整数。',learningRate:'学习率应为大于 0 的数。',threshold:'评测阈值应在 0 到 1 之间。',runtime:'请为示例服务选择 HTTP；自定义容器可选择对应运行时。'};
  function validateStep(number) {
    const errors=ConsoleCreate.validate(draft,number);
    if(!errors.length){draftError='';return true;}
    draftError=errors.map(e=>validationMessages[e.field]||e.message).join(' ');renderCreate();
    const input=document.getElementById('draft-'+errors[0].field);if(input){input.setAttribute('aria-invalid','true');input.focus();}return false;
  }
  async function submitDraft() {
    if(submitting || uncertain || !validateStep(3))return;
    const data=ConsoleCreate.generate(draft), name=draft.name, type=draft.kind, ns=identity.namespace, v=routeVersion;
    submitting=true;submitMessage='';draftError='';renderCreate();
    try {
      // A successful preflight avoids an accidental apply over an existing object.
      if(!draft.dryRun) {
        const existing=await request(listPath('applications'));
        if((existing.items||[]).some(i=>i.namespace===ns&&i.name===name))throw Object.assign(new Error('同名任务或服务已存在，请更换名称。'),{status:409});
      }
      const response=await request(apiRoot+'/applications?dryRun='+String(draft.dryRun),{method:'POST',headers:{'Content-Type':'application/yaml'},body:data});
      if(v!==routeVersion)return;
      if(response.application?.dryRun){submitMessage='配置校验通过，尚未创建任务或服务。';dirty=true;}
      else {dirty=false;draft=null;submitting=false;toast('已提交，正在读取运行状态');navigate(detailHash(type,ns,response.application?.name||name),true);return;}
    }catch(err){
      if(v!==routeVersion || err.status===401)return;
      if((err.status===0 || err.name==='AbortError')&&!draft.dryRun) {uncertain=true;submitMessage='提交结果暂未确认。请查询创建结果，避免重复创建。';}
      else draftError=errorText(err);
    }finally{submitting=false;if(v===routeVersion&&draft)renderCreate();}
  }
  async function checkCreated() {
    const v=routeVersion;
    try {
      const data=await request(listPath('applications'));
      if(v!==routeVersion)return;
      const found=(data.items||[]).find(i=>i.namespace===identity.namespace&&i.name===draft.name);
      if(found){const hash=detailHash(draft.kind,identity.namespace,draft.name);dirty=false;draft=null;navigate(hash,true);}
      else {submitMessage='当前未查到同名资源，原请求仍可能完成。请稍后再次查询创建结果。';renderCreate();}
    }catch(err){if(v===routeVersion){draftError=errorText(err);renderCreate();}}
  }
  function prefill(asset,action) {
    const service=action==='asset-deploy', d=ConsoleCreate.defaults(service?'AIService':'AIJob',identity.namespace,identity.tenant);
    d.purpose=service?'service':action==='asset-eval'?'evaluation':'training';
    if(asset.modelURI){d.modelURI=asset.modelURI;d.modelName=asset.name;d.modelVersion=asset.version||'v1';}
    if(asset.datasetURI)d.datasetURI=asset.datasetURI;
    draft=d;dirty=false;navigate(service?'#/services/new':'#/jobs/new',true);
  }
  document.addEventListener('click',async event=>{
    const tab=event.target.closest('[data-tab]');
    if(tab){navigate(detailHash(route.section==='jobs'?'AIJob':'AIService',route.namespace,route.name)+'?tab='+tab.dataset.tab);return;}
    const stepButton=event.target.closest('[data-step]');if(stepButton&&!stepButton.disabled){step=Number(stepButton.dataset.step);draftError='';renderCreate();return;}
    const el=event.target.closest('[data-action]');if(!el || el.disabled)return;
    const action=el.dataset.action;
    if(action==='close-modal'){if(!modalBusy)modal.close();return;}
    if(action==='new-job'||action==='new-service'){draft=null;navigate(action==='new-job'?'#/jobs/new':'#/services/new');return;}
    if(action==='refresh'){refresh();return;}
    if(action==='logs'){loadLogs();return;}
    if(action==='register'){registerAsset(route.section==='models');return;}
    if(action==='register-result'){registerAsset(true,result);return;}
    if(action==='publish-result'){publishResult();return;}
    if(['rerun','restart','delete'].includes(action)){lifecycle(action);return;}
    if(['parse-result','probe'].includes(action)) {el.disabled=true;await detailAction(action);if(el.isConnected)el.disabled=false;return;}
    if(action.startsWith('asset-')) {const asset=items[Number(el.closest('td').querySelector('[data-index]').dataset.index)];if(action==='asset-detail')assetDetail(asset);else prefill(asset,action);return;}
    if(action==='cancel-create'){navigate('#/'+route.section);return;}
    if(action==='previous'){step--;draftError='';renderCreate();return;}
    if(action==='next'){if(validateStep(step)){step++;renderCreate();}return;}
    if(action==='submit'){await submitDraft();return;}
    if(action==='check-created'){await checkCreated();return;}
    if(action==='reload-assets'){loadAssetOptions();return;}
    if(action==='register-create-model'||action==='register-create-dataset') {
      const models=action==='register-create-model';
      showModal(models?'登记模型':'登记数据集',modalField('name','名称')+modalField('uri','存储引用'),async form=>{
        await request(apiRoot+(models?'/models':'/datasets'),{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({namespace:identity.namespace,name:form.get('name'),[models?'modelURI':'datasetURI']:form.get('uri'),version:'v1',format:'sharegpt-jsonl',owner:identity.username})});
        draft[models?'modelURI':'datasetURI']=form.get('uri');await loadAssetOptions();toast('已登记');
      },'登记');
    }
  });
  document.addEventListener('input',event=>{
    const key=event.target.dataset.field;
    if(key&&draft){draft[key]=key==='example'?event.target.value==='true':event.target.value;dirty=true;}
  });
  document.addEventListener('change',event=>{
    const el=event.target,key=el.dataset.field;
    if(key&&draft) {
      draft[key]=key==='example'?el.value==='true':el.value;dirty=true;
      if(key==='purpose'&&['agent','batch','offline-inference'].includes(draft.purpose)) {draft.example=false;draft.command='';}
      if(key==='example'&&draft.example){draft.runtime=draft.kind==='AIJob'?'batch':'http';draft.image=draft.kind==='AIJob'?'busybox:1.36':'python:3.11-slim';draft.command='';}
      if(key==='modelURI'){const m=modelOptions.find(m=>m.modelURI===el.value);if(m){draft.modelName=m.name;draft.modelVersion=m.version||'v1';}}
      if(['purpose','example'].includes(key))renderCreate();
    }
    if(el.id==='resource-preset') {if(el.value==='cpu')Object.assign(draft,{cpu:'1',memory:'1Gi',gpu:'0',nodeType:'cpu'});if(el.value==='gpu')Object.assign(draft,{cpu:'8',memory:'32Gi',gpu:'1',nodeType:'gpu'});dirty=true;renderCreate();}
    if(el.id==='execution-mode'){draft.dryRun=el.value==='dry';dirty=true;submitMessage='';renderCreate();}
    if(['search','status-filter','purpose-filter'].includes(el.id)) {
      const params=new URLSearchParams();const q=document.getElementById('search')?.value,s=document.getElementById('status-filter')?.value,p=document.getElementById('purpose-filter')?.value;
      if(q)params.set('q',q);if(s)params.set('status',s);if(p)params.set('purpose',p);navigate('#/'+route.section+(params.size?'?'+params:''));
    }
    if(el.id==='log-pod'){document.getElementById('log-container').value='';loadLogs();}
    if(el.id==='log-container')loadLogs();
  });
  modal.addEventListener('cancel',event=>{if(modalBusy)event.preventDefault();});
  modal.addEventListener('submit',async event=>{
    event.preventDefault();if(modalBusy)return;modalBusy=true;
    const submit=modal.querySelector('[type="submit"]');submit.disabled=true;
    try{await modalAction(new FormData(event.target));modal.close();}catch(err){pendingNavigation='';const error=modal.querySelector('#modal-error');if(error)error.innerHTML=notice(errorText(err),'');}finally{modalBusy=false;if(submit.isConnected)submit.disabled=false;if(pendingNavigation){const next=pendingNavigation;pendingNavigation='';navigate(next);}}
  });
  loadRoute();icons();
})();
