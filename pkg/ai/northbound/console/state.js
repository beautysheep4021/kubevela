(function (root) {
  'use strict';
  function escape(value) {
    return String(value == null ? '' : value).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
  }
  function kind(item) {
    const metadata = item.aiMetadata || {};
    const types = (item.workloadTypes || []).concat((item.components || []).map(c => c.type || c.workloadKind));
    if (item.kind === 'AIJob' || metadata['ai.oam.dev/workload-kind'] === 'job' || types.some(t => /job|batch/i.test(t))) return 'AIJob';
    if (item.kind === 'AIService' || metadata['ai.oam.dev/workload-kind'] === 'service' || types.some(t => /service|deployment/i.test(t))) return 'AIService';
    return '';
  }
  function status(item, type) {
    const p = (item.phase || '').toLowerCase();
    const workloads = (item.components || []).map(c => c.workload || {});
    const success = /^(succeeded|complete|completed)$/.test(p) || (workloads.length && workloads.every(w => w.completed === true));
    if (type === 'AIJob' && success) return {key:'succeeded',label:'已完成',tone:'ok',terminal:true};
    if (/^(failed|error|errorappconfig)$/.test(p)) return {key:'failed',label:'失败',tone:'bad',terminal:true};
    if (type === 'AIService' && item.healthy) return {key:'ready',label:'已就绪',tone:'ok',terminal:false};
    if (type === 'AIJob' && (item.healthy || p === 'running' || workloads.some(w => w.active > 0))) return {key:'running',label:'运行中',tone:'ok',terminal:false};
    if (p === 'pending' || p === 'rendering' || p === 'starting' || p === 'deploying' || p === 'running') return {key:'pending',label:type === 'AIService' ? '部署中' : '等待中',tone:'warn',terminal:false};
    return {key:'unknown',label:'状态待确认',tone:'neutral',terminal:false};
  }
  function route(hash) {
    try {
      const u = new URL((hash || '#/overview').replace(/^#/, ''), 'http://console');
      const parts = u.pathname.split('/').filter(Boolean).map(decodeURIComponent);
      const section = ['overview','jobs','services','models','datasets'].includes(parts[0]) ? parts[0] : 'overview';
      const objectPage = ['jobs','services'].includes(section);
      const create = objectPage && parts.length === 2 && parts[1] === 'new';
      const detail = objectPage && parts.length === 3;
      return {section,create,detail,namespace:detail ? parts[1] : '',name:detail ? parts[2] : '',
        tab:['overview','logs','result','config'].includes(u.searchParams.get('tab')) ? u.searchParams.get('tab') : 'overview',
        query:u.searchParams.get('q') || '',filter:u.searchParams.get('status') || '',purpose:u.searchParams.get('purpose') || '',
        invalid:parts.length > 1 && !create && !detail};
    } catch (_) { return {section:'overview',invalid:true}; }
  }
  function example(item) {
    const m = item.aiMetadata || {};
    return item.example === true || m['ai.oam.dev/framework'] === 'demo' || m['ai.oam.dev/example'] === 'true' || /^inline:\/\//.test(item.modelURI || item.datasetURI || m['ai.oam.dev/model-uri'] || '');
  }
  root.ConsoleState = {escape,kind,status,route,example};
})(globalThis);
