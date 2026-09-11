const {test} = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const path = require('node:path');

const prefix = '/api/v1/ai';
const json = (value, status = 200) => new Response(JSON.stringify(value), {
  status, headers: {'Content-Type': 'application/json'}
});
const post = body => ({method: 'POST', body: JSON.stringify(body), headers: {'Content-Type': 'application/json'}});
const ownIdentity = {namespace: 'ai-tenant-a', tenant: 'tenant-a', username: 'user-a'};
test('monitor transport accepts empty scope and invalidates when role changes', async () => {
  let identity = {role:'monitor',username:'admin',namespace:'',tenant:''};
  const client = harness(async () => json(identity)).create({role:'monitor',username:'admin'});
  assert.equal((await client.request(prefix+'/session')).role,'monitor');
  identity = {role:'user',username:'admin',namespace:'ai-tenant-a',tenant:'tenant-a'};
  await assert.rejects(client.request(prefix+'/session'),{status:401});
});
function harness(fetcher, globals = {}) {
  const sandbox = {URL, Headers, AbortController, setTimeout, clearTimeout, fetch: fetcher, ...globals};
  const file = path.join(__dirname, 'api.js');
  if (fs.existsSync(file)) vm.runInNewContext(fs.readFileSync(file, 'utf8'), sandbox);
  assert.equal(typeof sandbox.ConsoleAPI?.create, 'function', 'plain script exposes ConsoleAPI.create');
  return sandbox.ConsoleAPI;
}
function demoHarness(letter = 'a') {
  let identity = {namespace: 'ai-tenant-' + letter, tenant: 'tenant-' + letter, username: 'user-' + letter, role: 'user'};
  let normalized;
  const calls = [];
  const unauthorized = [];
  const api = harness(async (url, init) => {
    calls.push({url, init});
    if (url === prefix + '/session') return json(identity);
    if (url === prefix + '/normalize') return json(normalized);
    if (url === prefix + '/validate') return json({valid: true});
    throw new Error('Unexpected real demo request: ' + url);
  });
  const options = {namespace: identity.namespace, tenant: identity.tenant, demo: true, onUnauthorized: e => unauthorized.push(e)};
  return {client: api.create(options), fresh: () => api.create(options), calls, unauthorized,
    identity: value => {identity = value;}, normalized: value => {normalized = value;}};
}
test('real requests return payload and force same-origin credentials', async () => {
  let options;
  const calls = [];
  const client = harness(async (url, init) => {
    calls.push(url);
    if (url === prefix + '/session') return json(ownIdentity);
    options = init; assert.equal(url, prefix + '/models'); return json({items: []});
  }).create(ownIdentity);
  assert.equal((await client.request(prefix + '/models', {credentials: 'omit'})).items.length, 0);
  assert.equal(options.credentials, 'same-origin');
  assert.equal(options.redirect, 'manual');
  assert.deepEqual(calls, [prefix + '/session', prefix + '/models']);
});
test('real errors never fall back to demo', async () => {
  for (const status of [400, 403, 404, 503]) {
    let calls = 0;
    const client = harness(async url => {
      if (url === prefix + '/session') return json(ownIdentity);
      calls++; return json({error: 'backend unavailable'}, status);
    }).create({...ownIdentity, demo: false});
    await assert.rejects(client.request(prefix + '/applications'), e => e.status === status && e.message === 'backend unavailable');
    assert.equal(calls, 1);
  }
  const client = harness(async () => {throw new TypeError('offline');}).create({});
  await assert.rejects(client.request(prefix + '/applications'), e => e.status === 0 && /offline/.test(e.message));
});
test('first session enforces expected username in live and demo modes, including direct session reads', async () => {
  for (const demo of [false, true]) {
    for (const endpoint of ['/session', '/applications']) {
      let calls = 0;
      const notices = [];
      const client = harness(async url => {
        calls++;
        assert.equal(url, prefix + '/session');
        return json({...ownIdentity, username: 'another-user'});
      }).create({...ownIdentity, demo, onUnauthorized: e => notices.push(e)});
      await assert.rejects(client.request(prefix + endpoint), e => e.status === 401);
      assert.equal(calls, 1);
      assert.equal(notices.length, 1);
    }
  }
});
test('direct session read verifies identity with exactly one GET', async () => {
  for (const demo of [false, true]) {
    const calls = [];
    const client = harness(async (url, init) => {calls.push({url, init}); return json(ownIdentity);}).create({...ownIdentity, demo});
    assert.equal((await client.request(prefix + '/session')).username, ownIdentity.username);
    assert.equal(calls.length, 1);
    assert.equal(calls[0].init.method || 'GET', 'GET');
    assert.equal(calls[0].init.cache, 'no-store');
  }
});
test('live identity change blocks write before dispatch and permanently invalidates client', async () => {
  let identity = ownIdentity;
  const calls = [];
  const client = harness(async url => {
    calls.push(url);
    return json(url === prefix + '/session' ? identity : {items: []});
  }).create(ownIdentity);
  await client.request(prefix + '/models');
  identity = {...ownIdentity, username: 'another-user'};
  await assert.rejects(client.request(prefix + '/applications', post({})), e => e.status === 401);
  identity = ownIdentity;
  await assert.rejects(client.request(prefix + '/models'), e => e.status === 401);
  assert.deepEqual(calls, [prefix + '/session', prefix + '/models', prefix + '/session']);
});
function controlledTimers() {
  let fire;
  let delay;
  let cleared = 0;
  return {globals: {setTimeout: (fn, ms) => {fire = fn; delay = ms; return 1;}, clearTimeout: () => {cleared++;}},
    fire: () => {assert.equal(typeof fire, 'function'); fire();}, delay: () => delay, cleared: () => cleared};
}
test('20-second deadline bounds session checks and prevents dispatch after a late response', async () => {
  const timer = controlledTimers();
  let signal;
  let release;
  const calls = [];
  const client = harness(async (url, init) => {
    calls.push(url); signal = init.signal;
    return new Promise(resolve => {release = resolve;});
  }, timer.globals).create(ownIdentity);
  const pending = client.request(prefix + '/applications', post({}));
  assert.equal(timer.delay(), 20000);
  timer.fire();
  await assert.rejects(pending, e => e.status === 0 && e.name === 'TimeoutError' && /timed out/i.test(e.message));
  assert.equal(signal.aborted, true);
  release(json(ownIdentity));
  await new Promise(resolve => setImmediate(resolve));
  assert.deepEqual(calls, [prefix + '/session']);
  assert.equal(timer.cleared(), 1);
});
test('write and body-read timeouts report status 0 without retry or demo fallback', async () => {
  for (const bodyRead of [false, true]) {
    const timer = controlledTimers();
    const calls = [];
    let signal;
    let started;
    const ready = new Promise(resolve => {started = resolve;});
    const client = harness(async (url, init) => {
      calls.push(url);
      if (url === prefix + '/session') return json(ownIdentity);
      signal = init.signal;
      if (!bodyRead) {started(); return new Promise(() => {});}
      return {status: 200, ok: true, text: () => {started(); return new Promise(() => {});}};
    }, timer.globals).create(ownIdentity);
    const pending = client.request(prefix + '/applications', post({}));
    await ready;
    timer.fire();
    await assert.rejects(pending, e => e.status === 0 && e.name === 'TimeoutError');
    assert.equal(signal.aborted, true);
    assert.deepEqual(calls, [prefix + '/session', prefix + '/applications']);
    assert.equal(timer.cleared(), 1);
  }
});
test('caller cancellation remains AbortError and cleans up timer and listener', async () => {
  const timer = controlledTimers();
  const controller = new AbortController();
  let removed = 0;
  const original = controller.signal.removeEventListener.bind(controller.signal);
  controller.signal.removeEventListener = (...args) => {removed++; return original(...args);};
  let combined;
  const client = harness(async (url, init) => {combined = init.signal; return new Promise(() => {});}, timer.globals).create(ownIdentity);
  const pending = client.request(prefix + '/models', {signal: controller.signal});
  controller.abort();
  await assert.rejects(pending, e => e.status === 0 && e.name === 'AbortError');
  assert.notEqual(combined, controller.signal);
  assert.equal(combined.aborted, true);
  assert.equal(timer.cleared(), 1);
  assert.equal(removed, 1);
});
test('successful requests clear deadline and permit configured timeout', async () => {
  const timer = controlledTimers();
  const client = harness(async () => json(ownIdentity), timer.globals).create({...ownIdentity, timeoutMs: 50});
  await client.request(prefix + '/session');
  assert.equal(timer.delay(), 50);
  assert.equal(timer.cleared(), 1);
});
test('401 and redirects notify unauthorized; HTML and malformed JSON produce useful errors', async () => {
  for (const response of [json({error: 'expired'}, 401), new Response(null, {status: 302}),
    {status: 0, type: 'opaqueredirect'}, {status: 200, redirected: true}]) {
    const notices = [];
    const client = harness(async () => response).create({onUnauthorized: e => notices.push(e)});
    await assert.rejects(client.request(prefix + '/models'), e => e.status === 401);
    assert.equal(notices.length, 1);
  }
  for (const response of [new Response('<html>login</html>', {headers: {'Content-Type': 'text/html'}}),
    new Response('{oops', {headers: {'Content-Type': 'application/json'}}),
    new Response('gateway down', {status: 503})]) {
    const client = harness(async () => response).create({});
    await assert.rejects(client.request(prefix + '/models'), e => typeof e.status === 'number' && !/Unexpected token/.test(e.message));
  }
});
test('rejects external, escaped traversal and non API paths before fetching', async () => {
  const client = harness(async () => {assert.fail('must not fetch');}).create({});
  for (const url of ['https://example.com/api/v1/ai/models', '//example.com/api/v1/ai/models', '/login',
    '/api/v1/ai/../models', '/api/v1/ai/%2e%2e/models', '/api/v1/ai/models#fragment']) {
    await assert.rejects(client.request(url), e => e.status === 400);
  }
});
test('demo seeds exactly one job and service per tenant and returns detached objects', async () => {
  for (const letter of ['a', 'b']) {
    const h = demoHarness(letter);
    const apps = (await h.client.request(prefix + '/applications')).items;
    assert.equal(apps.length, 2);
    assert.equal(apps.filter(a => a.workloadTypes.includes('job')).length, 1);
    assert.equal(apps.filter(a => a.workloadTypes.includes('service')).length, 1);
    for (const collection of ['applications', 'models', 'datasets', 'artifacts', 'audits']) {
      const {items} = await h.client.request(prefix + '/' + collection);
      assert.ok(items.length);
      assert.ok(items.every(i => i.namespace === 'ai-tenant-' + letter && i.demo === true));
    }
    apps[0].name = 'tampered';
    assert.notEqual((await h.client.request(prefix + '/applications')).items[0].name, 'tampered');
    assert.ok(h.calls.every(c => c.url === prefix + '/session'));
  }
});
test('demo rejects cross-scope paths, bodies and asset references', async () => {
  const {client} = demoHarness();
  for (const url of ['/applications?namespace=ai-tenant-b', '/applications/ai-tenant-b/x/status',
    '/models/ai-tenant-b/x/evaluate', '/deliveries/ai-tenant-b/x/result',
    '/applications?namespace=ai-tenant-a&namespace=ai-tenant-b']) {
    await assert.rejects(client.request(prefix + url), e => e.status === 403);
  }
  for (const body of [{namespace: 'ai-tenant-b', modelURI: 's3://bucket/model'},
    {namespace: 'ai-tenant-a', modelURI: 'inline://models/tenant-b-evaluation-job/v1'},
    {namespace: 'ai-tenant-a', modelURI: 'model://ai-tenant-b/private/v1'},
    {namespace: 'ai-tenant-a', modelURI: 's3://bucket/m', sourceApplication: 'tenant-b-evaluation-job'}]) {
    await assert.rejects(client.request(prefix + '/models', post(body)), e => e.status === 403 || e.status === 404);
  }
});
test('demo stale scope and same-scope identity changes invalidate the old client permanently', async () => {
  for (const identity of [{namespace: 'ai-tenant-b', tenant: 'tenant-b'},
    {namespace: 'ai-tenant-a', tenant: 'tenant-a', username: 'another-user', role: 'user'}]) {
    const h = demoHarness();
    await h.client.request(prefix + '/applications');
    h.identity(identity);
    await assert.rejects(h.client.request(prefix + '/applications'), e => e.status === 401);
    h.identity({namespace: 'ai-tenant-a', tenant: 'tenant-a', username: 'user-a', role: 'user'});
    await assert.rejects(h.client.request(prefix + '/applications'), e => e.status === 401);
    assert.equal(h.unauthorized.length, 1);
  }
});
test('demo session failures fail closed and aborted actions do not mutate', async () => {
  for (const status of [401, 404, 503]) {
    const client = harness(async () => json({error: 'session unavailable'}, status)).create({namespace: 'ai-tenant-a', tenant: 'tenant-a', demo: true});
    await assert.rejects(client.request(prefix + '/applications'), e => e.status === status);
  }
  const h = demoHarness();
  const controller = new AbortController();
  controller.abort();
  await assert.rejects(h.client.request(prefix + '/applications/ai-tenant-a/tenant-a-training-job', {method: 'DELETE', signal: controller.signal}), e => e.name === 'AbortError');
  assert.equal((await h.client.request(prefix + '/applications')).items.length, 2);
});
test('demo status, logs, lifecycle and missing objects', async () => {
  const {client} = demoHarness();
  const job = prefix + '/applications/ai-tenant-a/tenant-a-training-job';
  const service = prefix + '/applications/ai-tenant-a/tenant-a-inference-service';
  assert.equal((await client.request(job + '/status')).phase, 'succeeded');
  assert.match((await client.request(job + '/logs')).logs, /AI_RESULT_JSON=/);
  await assert.rejects(client.request(job + '/restart', post({})), e => e.status === 400);
  await assert.rejects(client.request(service + '/rerun', post({})), e => e.status === 400);
  await client.request(job + '/rerun', post({}));
  assert.equal((await client.request(job)).phase, 'running');
  await client.request(service + '/restart', post({}));
  assert.equal((await client.request(service + '/probe', post({path: '/ready'}))).path, '/ready');
  assert.equal((await client.request(service + '/probe', post({}))).application, 'tenant-a-inference-service');
  await client.request(job, {method: 'DELETE'});
  await assert.rejects(client.request(job + '/status'), e => e.status === 404);
  await assert.rejects(client.request(prefix + '/unknown'), e => e.status === 404);
});
test('demo model, dataset, evaluation, result and publishing workflow', async () => {
  const h = demoHarness();
  const {client} = h;
  const dataset = await client.request(prefix + '/datasets', post({namespace: 'ai-tenant-a', name: 'data', datasetURI: 's3://my-bucket/data'}));
  const model = await client.request(prefix + '/models', post({namespace: 'ai-tenant-a', name: 'model', modelURI: 's3://my-bucket/model/v1'}));
  const base = prefix + '/models/ai-tenant-a/' + model.name;
  const evaluation = await client.request(base + '/evaluate', post({jobName: 'custom-eval', evaluationDatasetURI: dataset.datasetURI}));
  assert.equal(evaluation.modelURI, model.modelURI);
  assert.equal(evaluation.jobName, 'custom-eval');
  const synced = await client.request(base + '/sync-evaluation', post({evaluationJobName: evaluation.jobName}));
  assert.equal(synced.asset.evaluationStatus, 'passed');
  assert.equal(synced.asset.status, 'evaluated');
  assert.equal(synced.result.modelURI, model.modelURI);
  const result = await client.request(prefix + '/deliveries/ai-tenant-a/custom-eval/result');
  assert.equal(result.result.modelURI, model.modelURI);
  const published = await client.request(prefix + '/deliveries/ai-tenant-a/custom-eval/publish-service', post({serviceName: 'published'}));
  assert.equal(published.application.dryRun, false);
  assert.equal((await client.request(prefix + '/applications/ai-tenant-a/published/status')).modelURI, model.modelURI);
  assert.ok((await client.request(prefix + '/artifacts')).items.some(a => a.modelURI === model.modelURI));
  assert.equal((await h.fresh().request(prefix + '/applications')).items.length, 2);
});
test('demo YAML creation uses backend normalization, honors dryRun and scope', async () => {
  const h = demoHarness();
  const normalized = {kind: 'AIJob', name: 'created', namespace: 'ai-tenant-a', componentName: 'worker',
    workloadType: 'job', image: 'busybox', governanceIntent: {tenant: 'tenant-a', placement: {namespace: 'ai-tenant-a'}},
    workloadIntent: {job: {jobKind: 'batch', resources: {cpu: '500m', memory: '1Gi', gpu: '0'}}}};
  h.normalized(normalized);
  const init = {method: 'POST', headers: {'Content-Type': 'application/yaml'}, body: 'kind: AIJob\nmetadata: {}'};
  const dry = await h.client.request(prefix + '/applications?dryRun=true', init);
  assert.equal(dry.application.dryRun, true);
  assert.equal((await h.client.request(prefix + '/applications')).items.length, 2);
  await h.client.request(prefix + '/applications?dryRun=false', init);
  const app = await h.client.request(prefix + '/applications/ai-tenant-a/created/status');
  assert.equal(app.resourceSummary.cpuMilli, 500);
  assert.equal(app.resourceSummary.memoryMi, 1024);
  assert.equal(app.jobKind, 'batch');
  assert.ok(h.calls.every(c => c.url !== prefix + '/applications?dryRun=false'));
  h.normalized({...normalized, namespace: 'ai-tenant-b'});
  await assert.rejects(h.client.request(prefix + '/applications', init), e => e.status === 403);
});
test('late session response cannot resurrect invalidated demo state', async () => {
  const own = {namespace: 'ai-tenant-a', tenant: 'tenant-a', username: 'a'};
  let count = 0;
  let release;
  const client = harness(async () => {
    count++;
    if (count === 1) return json(own);
    if (count === 2) return new Promise(resolve => {release = resolve;});
    return json({namespace: 'ai-tenant-b', tenant: 'tenant-b', username: 'b'});
  }).create({...own, demo: true});
  await client.request(prefix + '/applications');
  const late = client.request(prefix + '/applications/ai-tenant-a/tenant-a-training-job', {method: 'DELETE'});
  await assert.rejects(client.request(prefix + '/applications'), e => e.status === 401);
  release(json(own));
  await assert.rejects(late, e => e.status === 401);
});
test('cancellation during session fetch leaves the demo job intact', async () => {
  let release;
  let count = 0;
  const own = {namespace: 'ai-tenant-a', tenant: 'tenant-a'};
  const client = harness(async () => {
    if (++count === 1) return new Promise(resolve => {release = resolve;});
    return json(own);
  }).create({...own, demo: true});
  const controller = new AbortController();
  const pending = client.request(prefix + '/applications/ai-tenant-a/tenant-a-training-job', {method: 'DELETE', signal: controller.signal});
  controller.abort();
  release(json(own));
  await assert.rejects(pending, e => e.name === 'AbortError' && e.status === 0);
  assert.equal((await client.request(prefix + '/applications')).items.length, 2);
});
test('seed result metrics and registration source references retain tenant isolation', async () => {
  const b = demoHarness('b');
  const result = await b.client.request(prefix + '/deliveries/ai-tenant-b/tenant-b-evaluation-job/result');
  assert.equal(result.result.metrics.accuracy, 0.94);
  const {client} = demoHarness();
  await assert.rejects(client.request(prefix + '/models', post({namespace: 'ai-tenant-a', name: 'bad-source',
    modelURI: 's3://my-bucket/model', jobName: 'tenant-b-evaluation-job'})), e => e.status === 403 || e.status === 404);
  await assert.rejects(client.request(prefix + '/datasets', {method: 'POST', body: '{broken'}), e => e.status === 400);
  assert.equal((await client.request(prefix + '/models')).items.length, 1);
});
