(function (root) {
  'use strict';
  var BASE = '/api/v1/ai';
  function failure(status, message) {
    var error = new Error(message);
    error.status = status;
    return error;
  }
  function clone(value) { return JSON.parse(JSON.stringify(value)); }
  function abort(signal) {
    if (signal && signal.aborted) {
      var error = failure(0, 'Request aborted');
      error.name = 'AbortError';
      throw error;
    }
  }
  function parsePath(path) {
    if (typeof path !== 'string' || !path.startsWith(BASE + '/') || /[\\#\s]/.test(path)) {
      throw failure(400, 'Expected a relative /api/v1/ai path');
    }
    var url = new URL(path, 'https://console.invalid');
    if (url.pathname !== path.split('?')[0]) throw failure(400, 'Invalid API path');
    return url;
  }
  function bodyObject(init) {
    try {
      var body = init.body ? (typeof init.body === 'string' ? JSON.parse(init.body) : clone(init.body)) : {};
      if (!body || typeof body !== 'object' || Array.isArray(body)) throw new Error();
      return body;
    } catch (_) { throw failure(400, 'Expected a JSON object request body'); }
  }
  function create(options) {
    options = options || {};
    var namespace = options.namespace || '';
    var tenant = options.tenant || '';
    var username = options.username;
    var timeoutMs = Number.isFinite(options.timeoutMs) && options.timeoutMs > 0 && options.timeoutMs <= 2147483647
      ? options.timeoutMs : 20000;
    var demo = options.demo === true;
    var identity;
    var invalid;
    var state;
    var onUnauthorized = options.onUnauthorized;
    function unauthorized(message) {
      if (!invalid) {
        invalid = failure(401, message || 'Session expired. Please sign in again.');
        state = null;
        if (typeof onUnauthorized === 'function') {
          try { onUnauthorized(invalid); } catch (_) { /* Preserve the request error. */ }
        }
      }
      return invalid;
    }
    async function transport(path, init) {
      abort(init.signal);
      var response;
      try {
        response = await root.fetch(path, Object.assign({}, init, {credentials: 'same-origin', redirect: 'manual'}));
      } catch (error) {
        if (error.name === 'AbortError') {
          var cancelled = failure(0, 'Request aborted');
          cancelled.name = 'AbortError';
          throw cancelled;
        }
        throw failure(0, error.message || 'Network request failed');
      }
      abort(init.signal);
      if (invalid) throw invalid;
      if (response.status === 401 || response.redirected || response.type === 'opaqueredirect' ||
          (response.status >= 300 && response.status < 400)) throw unauthorized();
      if (response.status === 204) return null;
      var text;
      try { text = await response.text(); }
      catch (error) { throw failure(response.status || 0, 'Unable to read API response'); }
      abort(init.signal);
      if (invalid) throw invalid;
      var type = response.headers.get('content-type') || '';
      if (/text\/html/i.test(type) || /^\s*</.test(text)) {
        // A successful HTML response from an API route is normally the login page.
        if (response.ok) throw unauthorized();
        throw failure(response.status, 'HTTP ' + response.status + ': API returned HTML instead of JSON');
      }
      var payload;
      try { payload = JSON.parse(text); }
      catch (_) { throw failure(response.status, 'HTTP ' + response.status + ': API returned a non-JSON response'); }
      if (!response.ok) throw failure(response.status, payload && (payload.error || payload.message) || 'HTTP ' + response.status);
      return payload;
    }
    function scope(value) {
      if (value && value !== namespace) throw failure(403, 'Namespace is outside account scope');
    }
    async function session(init) {
      var current = await transport(BASE + '/session', {signal: init.signal, cache: 'no-store'});
      if (!current || !current.namespace || !current.tenant || current.namespace !== namespace || current.tenant !== tenant) {
        throw unauthorized('Account scope changed. Please sign in again.');
      }
      if (username !== undefined && current.username !== username) {
        throw unauthorized('Account identity changed. Please sign in again.');
      }
      var key = JSON.stringify([current.namespace, current.tenant, current.username || '', current.role || '']);
      if (identity && identity !== key) throw unauthorized('Account identity changed. Please sign in again.');
      identity = key;
      return current;
    }
    function requireName(value) {
      if (typeof value !== 'string' || !/^[a-z0-9](?:[-a-z0-9.]*[a-z0-9])?$/.test(value) || value.length > 253) {
        throw failure(400, 'A valid resource name is required');
      }
      return value;
    }
    function find(collection, name) {
      var item = state[collection].find(function (entry) { return entry.name === name; });
      if (!item) throw failure(404, collection + ' object not found: ' + name);
      return item;
    }
    function upsert(collection, item) {
      state[collection] = state[collection].filter(function (entry) { return entry.name !== item.name; });
      state[collection].unshift(item);
      return item;
    }
    function audit(action, name) {
      state.audits.unshift({namespace: namespace, name: name, action: action, actor: 'console',
        success: true, message: 'Demo ' + action, time: new Date().toISOString(), demo: true});
      state.audits.length = Math.min(state.audits.length, 20);
    }
    function application(name, kind, modelURI, jobKind) {
      requireName(name);
      var service = kind === 'service';
      var result = {modelURI: modelURI || 'inline://models/' + name + '/v1',
        metrics: {loss: 0.12, accuracy: 0.98}, summary: 'demo-result', demo: true};
      var container = service ? 'http' : (jobKind === 'evaluation' ? 'eval' : 'trainer');
      var app = {namespace: namespace, name: name, kind: service ? 'AIService' : 'AIJob',
        workloadTypes: [kind], jobKind: service ? undefined : jobKind || 'training',
        phase: service ? 'running' : 'succeeded', healthy: true, demo: true, createdAt: new Date().toISOString(),
        aiMetadata: {'ai.oam.dev/tenant': tenant, 'ai.oam.dev/project': name, 'ai.oam.dev/environment': 'test'},
        resourceSummary: {cpuMilli: service ? 2000 : 4000, memoryMi: service ? 4096 : 8192, gpu: 0},
        components: [{name: name + '-component'}], pod: name + '-pod-0', container: container,
        pods: [{name: name + '-pod-0', phase: service ? 'Running' : 'Succeeded', containers: [container]}],
        logs: service ? 'service booted\nhealthz ok' : 'run-start\nAI_RESULT_JSON=' + JSON.stringify(result) + '\nrun-complete'};
      if (service) app.modelURI = modelURI || '';
      else app.deliveryResult = result;
      return app;
    }
    function seed() {
      state = {applications: [], models: [], datasets: [], artifacts: [], audits: []};
      var letter = tenant === 'tenant-a' && namespace === 'ai-tenant-a' ? 'a' :
        tenant === 'tenant-b' && namespace === 'ai-tenant-b' ? 'b' : '';
      if (!letter) return;
      var a = letter === 'a';
      var jobName = a ? 'tenant-a-training-job' : 'tenant-b-evaluation-job';
      var serviceName = a ? 'tenant-a-inference-service' : 'tenant-b-chat-service';
      var modelName = a ? jobName : 'tenant-b-chat';
      var modelURI = a ? 'inline://models/' + jobName + '/v1' : 'oss://models/tenant-b-chat/v1';
      var job = application(jobName, 'job', '', a ? 'training' : 'evaluation');
      job.deliveryResult.metrics = a ? {loss: 0.12, accuracy: 0.98} : {accuracy: 0.94};
      job.deliveryResult.summary = a ? 'tenant-a-training-complete' : 'tenant-b-evaluation-passed';
      job.logs = 'run-start\nAI_RESULT_JSON=' + JSON.stringify(job.deliveryResult) + '\nrun-complete';
      job.resourceSummary = a ? {cpuMilli: 8000, memoryMi: 32768, gpu: 1} : {cpuMilli: 4000, memoryMi: 8192, gpu: 0};
      state.applications = [job, application(serviceName, 'service', modelURI)];
      if (!a) state.applications[1].resourceSummary = {cpuMilli: 4000, memoryMi: 8192, gpu: 1};
      state.models = [{namespace: namespace, name: modelName, modelURI: modelURI, status: a ? 'trained' : 'released',
        evaluationStatus: 'passed', visibility: 'private', metrics: a ? {loss: 0.12, accuracy: 0.98} : {accuracy: 0.94}, demo: true}];
      var datasetName = a ? 'tenant-a-training-data' : 'tenant-b-evaluation-data';
      state.datasets = [{namespace: namespace, name: datasetName, displayName: datasetName,
        datasetURI: 'inline://datasets/' + datasetName, format: a ? 'sharegpt-jsonl' : 'alpaca-jsonl',
        purpose: a ? 'sft' : 'evaluation', status: 'validated', visibility: 'private', demo: true}];
      state.artifacts = [{namespace: namespace, name: jobName + '-v1', modelURI: job.deliveryResult.modelURI,
        status: 'ready', framework: 'demo', source: 'AIJob', demo: true}];
      audit('deploy', jobName);
      audit('publish-service', serviceName);
    }
    function references(value, key) {
      if (Array.isArray(value)) { value.forEach(function (v) { references(v, key); }); return; }
      if (value && typeof value === 'object') {
        Object.keys(value).forEach(function (k) { references(value[k], k); }); return;
      }
      if (typeof value !== 'string' || !value) return;
      if (key === 'namespace' || key === 'tenantNamespace') scope(value);
      if (key === 'tenant' && value !== tenant) throw failure(403, 'Tenant is outside account scope');
      if (/URI$|^uri$/i.test(key || '')) {
        var uri;
        try { uri = new URL(value); } catch (_) { throw failure(400, 'Invalid asset URI'); }
        var decoded;
        try { decoded = decodeURIComponent(uri.hostname + uri.pathname); } catch (_) { throw failure(400, 'Invalid asset URI'); }
        var other = tenant === 'tenant-a' ? 'tenant-b' : tenant === 'tenant-b' ? 'tenant-a' : '';
        if (other && decoded.split('/').some(function (part) { return part === 'ai-' + other || part === other || part.startsWith(other + '-'); })) {
          throw failure(403, 'Asset reference is outside account scope');
        }
        if ((uri.protocol === 'model:' || uri.protocol === 'dataset:') && uri.hostname !== namespace) {
          throw failure(403, 'Asset reference is outside account scope');
        }
      }
      if (key === 'sourceApplication' || key === 'jobName' || key === 'evaluationJobName') requireName(value);
      if (key === 'sourceApplication') find('applications', value);
    }
    function assetName(uri, model) {
      var pieces = uri.split('?')[0].replace(/\/+$/, '').split('/');
      var name = pieces.pop();
      if (model && /^(v\d+|latest)$/i.test(name)) name = pieces.pop();
      return requireName((name || (model ? 'model' : 'dataset')).toLowerCase().replace(/[_.:/]/g, '-').replace(/^-+|-+$/g, '').slice(0, 50));
    }
    function resultFor(job) {
      if (!job.workloadTypes.includes('job') || !job.deliveryResult) throw failure(400, 'No job result is available');
      return job.deliveryResult;
    }
    async function normalize(init) {
      var normalized = await transport(BASE + '/normalize', init);
      references(normalized);
      scope(normalized.namespace);
      requireName(normalized.name);
      return normalized;
    }
    async function demoRequest(url, init) {
      var method = (init.method || 'GET').toUpperCase();
      var parts;
      try { parts = url.pathname.slice(BASE.length + 1).split('/').map(decodeURIComponent); }
      catch (_) { throw failure(400, 'Invalid encoded API path'); }
      url.searchParams.getAll('namespace').forEach(scope);
      var collection = parts[0];
      if (parts.length > 1) scope(parts[1]);
      if (parts.length === 1 && method === 'POST' && ['validate', 'normalize', 'applications'].includes(collection)) {
        var normalized = await normalize(init);
        if (collection === 'normalize') return Object.assign({}, normalized, {demo: true});
        if (collection === 'validate') return Object.assign({}, await transport(BASE + '/validate', init), {demo: true});
        // Recheck identity after normalization before changing in-memory state.
        await session(init);
        var dryRun = url.searchParams.get('dryRun') === 'true';
        if (!dryRun) {
          var intent = normalized.workloadIntent || {};
          var job = intent.job || {};
          var service = intent.service || {};
          var app = application(normalized.name, normalized.workloadType, service.model && service.model.uri, job.jobKind);
          app.normalized = clone(normalized);
          app.components = [{name: normalized.componentName || normalized.name}];
          var resources = (intent.job || intent.service || {}).resources || {};
          var cpu = resources.cpu || '1';
          var memory = resources.memory || '1Gi';
          app.resourceSummary = {cpuMilli: parseFloat(cpu) * (/m$/.test(cpu) ? 1 : 1000),
            memoryMi: parseFloat(memory) * (/Gi$/.test(memory) ? 1024 : /Ki$/.test(memory) ? 1 / 1024 : /Mi$/.test(memory) ? 1 : 1 / 1048576),
            gpu: Number(resources.gpu || 0)};
          upsert('applications', app);
          audit('deploy', app.name);
        }
        return {normalized: normalized, application: {namespace: namespace, name: normalized.name, dryRun: dryRun}, demo: true};
      }
      if (parts.length === 1 && Object.prototype.hasOwnProperty.call(state, collection)) {
        if (method === 'GET') return {items: state[collection], demo: true};
        if (method === 'POST' && (collection === 'models' || collection === 'datasets')) {
          var record = bodyObject(init);
          references(record);
          if (!record.namespace) throw failure(400, 'namespace is required');
          var isModel = collection === 'models';
          var uri = record[isModel ? 'modelURI' : 'datasetURI'];
          if (!uri) throw failure(400, (isModel ? 'modelURI' : 'datasetURI') + ' is required');
          record.name = requireName(record.name || assetName(uri, isModel));
          if (record.jobName) find('applications', record.jobName);
          record.status = record.status || (isModel ? 'trained' : 'registered');
          record.visibility = record.visibility || 'private';
          record.demo = true;
          record.createdAt = record.createdAt || new Date().toISOString();
          if (isModel) record.evaluationStatus = record.evaluationStatus || 'pending';
          else record.displayName = record.displayName || record.name;
          upsert(collection, record);
          if (isModel) upsert('artifacts', clone(record));
          audit(isModel ? 'register-model' : 'register-dataset', record.name);
          return record;
        }
        throw failure(405, 'Method not allowed');
      }
      if (parts.length < 3 || parts.length > 4) throw failure(404, 'API route not found');
      var name = requireName(parts[2]);
      var action = parts[3] || '';
      if (collection === 'applications') {
        var target = find('applications', name);
        if (method === 'GET' && (!action || action === 'status')) return target;
        if (method === 'GET' && action === 'logs') return {namespace: namespace, name: name, logs: target.logs,
          pods: target.pods, pod: target.pod, container: target.container, demo: true};
        if ((method === 'DELETE' && !action) || (action === 'delete' && ['DELETE', 'POST'].includes(method))) {
          state.applications = state.applications.filter(function (app) { return app.name !== name; });
          audit('delete', name);
          return {namespace: namespace, name: name, action: 'delete', message: 'Demo application deleted', demo: true};
        }
        if (method === 'POST' && ['restart', 'rerun', 'probe'].includes(action)) {
          var payload = bodyObject(init);
          references(payload);
          var isService = target.workloadTypes.includes('service');
          if ((action === 'rerun' && isService) || (action !== 'rerun' && !isService)) throw failure(400, 'Action does not apply to this workload kind');
          if (action === 'probe') return {namespace: namespace, name: name, application: name, healthy: target.healthy,
            message: 'Demo access check', path: payload.path || '/healthz', statusCode: target.healthy ? 200 : 503, demo: true};
          target.phase = 'running';
          target.healthy = true;
          target.pods.forEach(function (pod) { pod.phase = 'Running'; });
          audit(action, name);
          return {namespace: namespace, name: name, action: action, message: 'Demo ' + action + ' submitted', demo: true};
        }
      }
      if (collection === 'deliveries') {
        var source = find('applications', name);
        var result = resultFor(source);
        if (method === 'GET' && action === 'result') return {namespace: namespace, jobName: name, result: result, demo: true};
        if (method === 'POST' && action === 'publish-service') {
          var publish = bodyObject(init);
          references(publish);
          var serviceName = requireName(publish.serviceName || name + '-service');
          upsert('applications', application(serviceName, 'service', result.modelURI));
          var artifact = {namespace: namespace, name: name, jobName: name, modelURI: result.modelURI,
            status: 'published', evaluationStatus: 'passed', metrics: result.metrics, summary: result.summary,
            publishedServices: [serviceName], visibility: 'private', demo: true};
          upsert('artifacts', artifact);
          upsert('models', clone(artifact));
          audit('publish-service', serviceName);
          return {namespace: namespace, jobName: name, serviceName: serviceName, modelURI: result.modelURI,
            application: {namespace: namespace, name: serviceName, dryRun: false}, demo: true};
        }
      }
      if (collection === 'models' && method === 'POST') {
        var model = find('models', name);
        var request = bodyObject(init);
        references(request);
        if (action === 'evaluate') {
          if (!request.evaluationDatasetURI) throw failure(400, 'evaluationDatasetURI is required');
          var evaluationName = requireName(request.jobName || name + '-eval');
          var evaluation = application(evaluationName, 'job', model.modelURI, 'evaluation');
          evaluation.deliveryResult.metrics = {accuracy: 0.98};
          evaluation.deliveryResult.summary = 'evaluation';
          evaluation.logs = 'eval-start\nAI_RESULT_JSON=' + JSON.stringify(evaluation.deliveryResult) + '\neval-complete';
          upsert('applications', evaluation);
          audit('evaluate', evaluationName);
          return {namespace: namespace, modelName: name, jobName: evaluationName, modelURI: model.modelURI,
            application: {namespace: namespace, name: evaluationName, dryRun: false}, demo: true};
        }
        if (action === 'sync-evaluation') {
          var syncName = requireName(request.evaluationJobName || name + '-eval');
          var syncResult = resultFor(find('applications', syncName));
          if (syncResult.modelURI !== model.modelURI) throw failure(400, 'Evaluation result belongs to another model');
          Object.assign(model, {metrics: clone(syncResult.metrics), summary: syncResult.summary,
            status: 'evaluated', evaluationStatus: 'passed', sourceApplication: syncName});
          upsert('artifacts', clone(model));
          audit('sync-evaluation', name);
          return {namespace: namespace, modelName: name, jobName: syncName, result: syncResult, asset: model, demo: true};
        }
      }
      throw failure(404, 'API route or action not found');
    }
    return {
      request: async function (path, init) {
        init = init || {};
        if (invalid) throw invalid;
        abort(init.signal);
        var url = parsePath(path);
        var callerSignal = init.signal;
        var controller = new AbortController();
        var requestInit = Object.assign({}, init, {signal: controller.signal});
        var timer;
        var cancel;
        // One deadline covers authentication, the operation, and response-body reads.
        var interrupted = new Promise(function (_, reject) {
          function stop(error) {
            if (controller.signal.aborted) return;
            reject(error);
            controller.abort();
          }
          cancel = function () {
            var error = failure(0, 'Request aborted');
            error.name = 'AbortError';
            stop(error);
          };
          if (callerSignal) callerSignal.addEventListener('abort', cancel, {once: true});
          timer = root.setTimeout(function () {
            var error = failure(0, 'Request timed out. The operation outcome may be unknown.');
            error.name = 'TimeoutError';
            stop(error);
          }, timeoutMs);
        });
        async function execute() {
          if (url.pathname === BASE + '/session') {
            if ((init.method || 'GET').toUpperCase() !== 'GET') throw failure(405, 'Method not allowed');
            return session(requestInit);
          }
          await session(requestInit);
          abort(controller.signal);
          if (invalid) throw invalid;
          if (!demo) return transport(path, requestInit);
          if (!state) seed();
          var payload = await demoRequest(url, requestInit);
          abort(controller.signal);
          if (invalid) throw invalid;
          return clone(payload);
        }
        try {
          return await Promise.race([interrupted, execute()]);
        } finally {
          root.clearTimeout(timer);
          if (callerSignal) callerSignal.removeEventListener('abort', cancel);
        }
      }
    };
  }
  root.ConsoleAPI = {create: create};
})(globalThis);
