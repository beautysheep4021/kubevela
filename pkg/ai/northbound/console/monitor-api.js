(function (root) {
  'use strict';
  const base = '/api/v1/ai';
  const copy = value => JSON.parse(JSON.stringify(value));
  function error(status, message) { return Object.assign(new Error(message), {status}); }
  function create(options = {}) {
    const transport = root.ConsoleAPI.create({...options, role: 'monitor', demo: false});
    let apps, audits = [];
    function seed() {
      apps = [];
      for (const letter of ['a', 'b']) {
        for (const service of [false, true]) {
          const namespace = 'ai-tenant-' + letter;
          const name = 'tenant-' + letter + (service ? '-inference-service' : '-training-job');
          const workloadKind = service ? 'Deployment' : 'Job';
          apps.push({name, namespace, kind: service ? 'AIService' : 'AIJob',
            phase: service ? 'running' : 'succeeded', healthy: true, demo: true,
            createdAt: new Date().toISOString(), workloadTypes: [service ? 'ai-service' : 'ai-job'],
            aiMetadata: {'ai.oam.dev/tenant': 'tenant-' + letter, 'ai.oam.dev/job-kind': service ? '' : 'training',
              'ai.oam.dev/example': 'true', 'ai.oam.dev/runtime': service ? 'http' : 'batch'},
            resourceSummary: {cpuMilli: 1000, memoryMi: 1024, gpu: 0},
            components: [{name: name + '-component', type: service ? 'ai-service' : 'ai-job', workloadKind, healthy: true,
              workload: {kind: workloadKind, completed: !service, succeeded: service ? 0 : 1,
                desiredReplicas: service ? 1 : 0, readyReplicas: service ? 1 : 0},
              service: service ? {name: name, clusterIP: '10.96.0.10', ports: [{port: 80, targetPort: 8080}]} : {},
              pods: [{name: name + '-pod', phase: service ? 'Running' : 'Succeeded', readyContainers: service ? 1 : 0,
                totalContainers: 1, containers: [{name: 'main', ready: service}], events: []}]}], diagnostics: []});
        }
      }
      const waiting = copy(apps[0]);
      waiting.name = 'tenant-a-waiting-job'; waiting.phase = 'pending'; waiting.healthy = false;
      waiting.message = 'FailedScheduling: Insufficient cpu';
      waiting.components = [{name: 'waiting', type: 'ai-job', workloadKind: 'Job', healthy: false,
        workload: {kind: 'Job', completed: false, active: 0},
        pods: [{name: 'tenant-a-waiting-job-pod', phase: 'Pending', containers: [{name: 'main', ready: false}],
          events: [{type: 'Warning', reason: 'FailedScheduling', message: 'Insufficient cpu', count: 1}]}]}];
      waiting.diagnostics = [{severity: 'warning', reason: 'FailedScheduling', message: 'Insufficient cpu', resource: waiting.name}];
      apps.push(waiting);
    }
    function record(app, action) {
      audits.unshift({id: String(Date.now()) + '-' + audits.length, time: new Date().toISOString(), namespace: app.namespace,
        name: app.name, action, actor: options.username, success: true, message: '演示操作已提交，运行状态待确认', demo: true});
    }
    function demoRequest(url, init) {
      const method = (init.method || 'GET').toUpperCase();
      const parts = url.pathname.slice(base.length + 1).split('/').map(decodeURIComponent);
      const [entity, namespace, name, action] = parts;
      const ns = url.searchParams.get('namespace');
      if (parts.length === 1 && method === 'GET') {
        if (entity === 'applications') return {items: apps.filter(item => !ns || item.namespace === ns)};
        if (entity === 'audits') return {items: audits.filter(item => !ns || item.namespace === ns)};
        if (entity === 'tenant-resources') return {items: ['ai-tenant-a', 'ai-tenant-b'].filter(n => !ns || n === ns).map(namespace => ({namespace,
          quotas: [{name: 'compute', desiredHard: {'requests.cpu': '8', 'requests.memory': '16Gi', 'requests.nvidia.com/gpu': '1'},
            hard: {'requests.cpu': '8', 'requests.memory': '16Gi', 'requests.nvidia.com/gpu': '1'}, scopes: [],
            used: {'requests.cpu': '1', 'requests.memory': '1Gi', 'requests.nvidia.com/gpu': '0'}}]}))};
      }
      const app = apps.find(item => item.namespace === namespace && item.name === name);
      if (!app) throw error(404, '未找到该资源');
      if (entity === 'deliveries' && action === 'result' && method === 'GET' && app.kind === 'AIJob') {
        if (app.phase !== 'succeeded') throw error(409, '任务尚未完成，暂无产物');
        return {namespace, jobName: name, result: {modelURI: 'inline://models/' + name, metrics: {accuracy: 0.98}, summary: '示例结果', demo: true}};
      }
      if (entity !== 'applications') throw error(404, '未找到该接口');
      if (action === 'status' && method === 'GET') return app;
      if (action === 'logs' && method === 'GET') {
        const pods = app.components.flatMap(c => c.pods || []).map(p => ({name: p.name, phase: p.phase, containers: p.containers.map(c => c.name)}));
        const pod = url.searchParams.get('pod') || pods[0]?.name;
        const selected = pods.find(p => p.name === pod);
        const container = url.searchParams.get('container') || selected?.containers[0];
        if (!selected || !selected.containers.includes(container)) throw error(404, '未找到 Pod 或容器');
        return {namespace, application: name, pod, container, pods,
          logs: app.phase === 'pending' ? '等待容器启动，暂无运行日志。' : '[demo] ' + name + '\n' + (app.kind === 'AIJob' ? 'run-complete' : 'service ready')};
      }
      if (action === 'probe' && method === 'POST' && app.kind === 'AIService') {
        let input;
        try { input = JSON.parse(init.body || '{}'); } catch (_) { throw error(400, '无效请求内容'); }
        const path = input?.path || '/healthz';
        if (typeof path !== 'string' || !path.startsWith('/') || path.startsWith('//')) throw error(400, '检查路径必须是相对服务路径');
        return {namespace, application: name, serviceName: name, path, healthy: app.healthy,
          statusCode: app.healthy ? 200 : 503, body: '[demo] ' + path, demo: true};
      }
      if ((!action && method === 'DELETE') || (['rerun', 'restart'].includes(action) && method === 'POST')) {
        if (action === 'rerun' && app.kind !== 'AIJob' || action === 'restart' && app.kind !== 'AIService') throw error(400, '操作与工作负载类型不匹配');
        if (!action) apps = apps.filter(item => item !== app);
        else {
          app.phase = action === 'rerun' ? 'pending' : 'deploying'; app.healthy = false;
          app.message = '操作已提交，等待运行状态'; app.diagnostics = [];
          for (const component of app.components) {
            component.healthy = false; component.workload.completed = false; component.workload.succeeded = 0; component.workload.readyReplicas = 0;
            for (const pod of component.pods || []) {pod.phase = 'Pending'; pod.events = []; pod.readyContainers = 0;}
          }
        }
        record(app, action || 'delete');
        return {namespace, name, action: action || 'delete', message: '操作已提交', demo: true};
      }
      throw error(405, '不支持此操作');
    }
    return {async request(path, init = {}) {
      if (!options.demo) return transport.request(path, init);
      // Authenticate the real browser session even when the data is simulated.
      await transport.request(base + '/session', {signal: init.signal});
      if (init.signal?.aborted) throw error(0, '请求已取消');
      if (typeof path !== 'string' || !path.startsWith(base + '/') || /[\\#\s]/.test(path)) throw error(400, '无效 API 路径');
      const url = new URL(path, 'http://console');
      if (url.pathname !== path.split('?')[0]) throw error(400, '无效 API 路径');
      if (!apps) seed();
      if (url.pathname === base + '/session') return {role: 'monitor', username: options.username};
      return copy(demoRequest(url, init));
    }};
  }
  root.MonitorAPI = {create};
})(globalThis);
