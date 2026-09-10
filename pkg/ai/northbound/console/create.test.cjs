const { test } = require('node:test');
const assert = require('node:assert/strict');
const { readFileSync, mkdtempSync, rmSync } = require('node:fs');
const { tmpdir } = require('node:os');
const { join } = require('node:path');
const vm = require('node:vm');
const { spawnSync } = require('node:child_process');

const context = vm.createContext({});
vm.runInContext(readFileSync(__dirname + '/create.js', 'utf8'), context);
const { defaults, validate, generate } = context.ConsoleCreate;
const draft = (kind = 'AIJob', changes = {}) => ({
  ...defaults(kind, 'tenant-a', 'tenant-a'), name: 'test-workload', ...changes
});
const document = d => JSON.parse(generate(d));
const fields = (d, step) => Array.from(validate(d, step), error => error.field);

test('defaults are fresh flat scoped drafts with dry-run and explicit example flags', () => {
  const job = defaults('AIJob', 'tenant-a', 'a');
  assert.equal(job.kind, 'AIJob');
  assert.equal(job.namespace, 'tenant-a');
  assert.equal(job.tenant, 'a');
  assert.equal(job.purpose, 'training');
  assert.equal(job.example, true);
  assert.equal(job.dryRun, true);
  assert.equal(job.name, '');
  assert.equal(job.cpu, '1');
  assert.equal(job.memory, '1Gi');
  assert.equal(job.gpu, '0');
  assert.equal(job.isolation, undefined);
  assert.equal(job.quota, undefined);
  assert.ok(Object.values(job).every(value => value === null || typeof value !== 'object'));
  job.cpu = '8';
  assert.equal(defaults('AIJob', 'tenant-b', 'b').cpu, '1');
  assert.equal(defaults('AIService', 'tenant-a', 'a').purpose, 'service');
  assert.throws(() => defaults('Unknown', 'tenant-a', 'a'));
});

test('steps validate their own fields; summary rechecks everything without changing retained values', () => {
  const d = draft('AIJob', { name: '', image: '', cpu: '-1' });
  const before = JSON.stringify(d);
  assert.deepEqual(fields(d, 0), []);
  assert.ok(fields(d, 1).includes('name'));
  assert.ok(fields(d, 1).includes('image'));
  assert.ok(!fields(d, 1).includes('cpu'));
  assert.deepEqual(fields(d, 2), ['cpu']);
  assert.ok(fields(d, 3).includes('cpu'));
  assert.ok(fields(d, 3).includes('name'));
  assert.equal(JSON.stringify(d), before);
  assert.ok(fields(d, 4).includes('step'));
  assert.ok(fields({ ...d, purpose: 'service' }, 0).includes('purpose'));
});

test('example training and evaluation preserve the demo result protocol and label fixed results', () => {
  for (const purpose of ['training', 'evaluation']) {
    const d = draft('AIJob', { purpose });
    const out = document(d);
    assert.equal(out.apiVersion, 'ai.oam.dev/v1alpha1');
    assert.equal(out.kind, 'AIJob');
    assert.equal(out.metadata.namespace, 'tenant-a');
    assert.equal(out.spec.properties.jobKind, purpose);
    assert.equal(out.spec.properties.dataset.uri, d.datasetURI);
    assert.deepEqual(out.spec.properties.cmd, ['sh', '-c']);
    const run = spawnSync('sh', ['-c', out.spec.properties.args[0]], {
      encoding: 'utf8', env: Object.fromEntries(out.spec.properties.env.map(e => [e.name, e.value]))
    });
    assert.equal(run.status, 0, run.stderr);
    assert.match(run.stdout, /EXAMPLE/);
    const result = JSON.parse(run.stdout.split('\n').find(line => line.startsWith('AI_RESULT_JSON=')).slice(15));
    assert.equal(result.example, true);
    assert.match(result.summary, /example/);
    assert.equal(out.spec.properties.annotations['ai.oam.dev/example'], 'true');
    assert.equal(out.spec.properties.replicas, undefined);
    assert.equal(out.spec.properties.endpoint, undefined);
    assert.equal(out.spec.isolation, undefined);
    assert.equal(out.dryRun, undefined);
  }
});

test('custom batch and offline inference require inputs and preserve exact multiline shell command', () => {
  for (const purpose of ['batch', 'offline-inference']) {
    const d = draft('AIJob', { purpose, example: false, image: 'registry.local/worker:v1', command: 'printf start\nexec /worker --run' });
    assert.equal(document(d).spec.properties.args[0], d.command);
    assert.ok(fields({ ...d, command: '' }).includes('command'));
    assert.ok(fields({ ...d, image: '' }).includes('image'));
    assert.equal(document(d).spec.properties.output, undefined);
    if (purpose === 'offline-inference') {
      assert.ok(fields({ ...d, modelURI: '' }).includes('modelURI'));
      assert.ok(fields({ ...d, datasetURI: '' }).includes('datasetURI'));
    } else {
      assert.deepEqual(fields({ ...d, modelURI: '', datasetURI: '' }), []);
    }
  }
});

test('example model service uses only HTTP; custom service and agent use supplied container', () => {
  const d = draft('AIService', { modelURI: 's3://models/test', modelName: 'test-model', port: '9090', replicas: '2' });
  const out = document(d);
  assert.equal(out.kind, 'AIService');
  assert.equal(out.spec.properties.model.uri, d.modelURI);
  assert.equal(out.spec.properties.replicas, 2);
  assert.equal(out.spec.properties.endpoint.port, 9090);
  assert.equal(out.spec.properties.jobKind, undefined);
  assert.equal(out.spec.properties.dataset, undefined);
  assert.match(out.spec.properties.args[0], /http\.server/);
  assert.doesNotMatch(out.spec.properties.args[0], /vllm/);
  assert.ok(fields({ ...d, runtime: 'vllm' }).includes('runtime'));
  assert.ok(fields({ ...d, modelURI: '' }).includes('modelURI'));
  for (const purpose of ['service', 'agent']) {
    const custom = { ...d, purpose, example: false, modelURI: '', modelName: '', command: 'exec /app', image: 'registry.local/app:v1' };
    const generated = document(custom);
    assert.equal(generated.spec.properties.model.name, custom.name);
    assert.equal(generated.spec.properties.args[0], custom.command);
    assert.equal(generated.spec.runtime.project, purpose);
    assert.ok(fields({ ...custom, image: '' }).includes('image'));
    assert.ok(fields({ ...custom, command: '' }).includes('command'));
  }
});

test('rejects invalid identifiers, required assets, booleans and resources before generation', () => {
  for (const [field, values] of Object.entries({
    name: ['', 'Bad Name', '-bad', 'x'.repeat(64), 'bad\nname'],
    namespace: ['', 'Bad', 'a.b'], tenant: ['', 'bad\ntenant'],
    modelURI: ['', '   '], datasetURI: ['', '   '],
    cpu: ['', 0, -1, 'NaN', 'Infinity', '1e3', '0.0001', '0.1m', '1\ngpu: 99', true],
    memory: ['', 0, '-1Gi', '1GB', 'Infinity', '1Gi\nfoo: bar'],
    gpu: ['', -1, 0.5, '2Gi', 'Infinity', false],
    epochs: ['', 0, 1.2], learningRate: ['', 0, -1, 'Infinity'],
    dryRun: ['false', null], example: ['true', null]
  })) {
    for (const value of values) {
      const d = draft('AIJob', { [field]: value });
      assert.ok(fields(d).includes(field), `${field}=${String(value)}`);
      assert.throws(() => generate(d), /Invalid draft/);
    }
  }
  for (const threshold of ['', -0.1, 1.1, 'NaN']) {
    assert.ok(fields(draft('AIJob', { purpose: 'evaluation', threshold })).includes('threshold'));
  }
  for (const field of ['port', 'replicas']) {
    for (const value of ['', 0, -1, 1.5, Infinity, '999999999999999999999']) {
      assert.ok(fields(draft('AIService', { [field]: value })).includes(field));
    }
  }
  assert.ok(fields(draft('AIService', { port: 65536 })).includes('port'));
});

test('example HTTP service creates a clearly labeled healthz file in its served directory', () => {
  const command = document(draft('AIService', { modelURI: 's3://models/test' })).spec.properties.args[0];
  const directory = mkdtempSync(join(tmpdir(), 'console-create-healthz-'));
  try {
    const [setup, server] = command.split('\nexec ');
    assert.match(server, /^python -m http\.server "\$PORT" --bind 0\.0\.0\.0 --directory \/tmp\/console-example$/);
    const run = spawnSync('sh', ['-c', setup.replaceAll('/tmp/console-example', "'" + directory.replaceAll("'", "'\\''") + "'")], { encoding: 'utf8' });
    assert.equal(run.status, 0, run.stderr);
    assert.match(readFileSync(join(directory, 'index.html'), 'utf8'), /EXAMPLE/);
    assert.match(readFileSync(join(directory, 'healthz'), 'utf8'), /EXAMPLE.*no model inference/);
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
});

test('scheduling defaults and every supported value round-trip for jobs and services', () => {
  const allowed = {
    priority: ['normal', 'high', 'urgent'],
    nodeType: ['cpu', 'gpu', 'any'],
    strategy: ['cost', 'performance', 'fast-start']
  };
  for (const kind of ['AIJob', 'AIService']) {
    const d = draft(kind, { modelURI: 's3://models/test' });
    const expected = { priority: 'normal', nodeType: 'cpu', strategy: 'performance' };
    for (const [field, value] of Object.entries(expected)) assert.equal(d[field], value);
    assert.deepEqual(document(d).spec.scheduling, expected);
    for (const [field, values] of Object.entries(allowed)) {
      for (const value of values) {
        const changed = { ...d, [field]: value };
        assert.deepEqual(fields(changed), []);
        assert.equal(document(changed).spec.scheduling[field], value);
        assert.equal(document(changed).spec.isolation, undefined);
      }
      for (const value of ['unsupported', 'normal\nnodeType: gpu', null, 1, false]) {
        const changed = { ...d, [field]: value };
        assert.deepEqual(fields(changed, 1), []);
        assert.deepEqual(fields(changed, 2), [field]);
        assert.deepEqual(fields(changed, 3), [field]);
        assert.throws(() => generate(changed), /Invalid draft/);
      }
    }
    const optional = { ...d, priority: undefined, nodeType: '', strategy: undefined };
    assert.deepEqual(fields(optional), []);
    assert.equal(document(optional).spec.scheduling, undefined);
    assert.deepEqual(document({ ...optional, nodeType: 'gpu' }).spec.scheduling, { nodeType: 'gpu' });
  }
});

test('accepts CPU millicores and quantity memory; ignores retained fields for other purposes', () => {
  const d = draft('AIJob', { cpu: '250m', memory: '512Mi', gpu: 2, replicas: 'bad', threshold: 'bad' });
  assert.deepEqual(fields(d), []);
  assert.deepEqual(document(d).spec.resources, { cpu: '250m', memory: '512Mi', gpu: '2' });
  assert.deepEqual(fields(draft('AIJob', { cpu: '0.125', memory: '1.5Gi' })), []);
  assert.deepEqual(fields(draft('AIJob', { purpose: 'evaluation', epochs: 'bad', learningRate: 'bad' })), []);
});

test('untrusted string scalars cannot inject YAML structure or shell interpolation into demo scripts', () => {
  const payload = "s3://bucket/a'\n---\nkind: AIService\n$(echo INJECTED)\n\"\\\u0085\u2028\u2029";
  const d = draft('AIJob', { modelURI: payload, datasetURI: payload });
  assert.doesNotMatch(generate(d), /[\u0085\u2028\u2029]/);
  const out = document(d);
  assert.equal(out.kind, 'AIJob');
  assert.equal(out.spec.properties.dataset.uri, payload);
  assert.equal(out.spec.runtime.modelURI, payload);
  assert.equal(out.spec.runtime.datasetURI, payload);
  assert.ok(!out.spec.properties.args[0].includes(payload));
  assert.ok(out.spec.properties.env.some(e => e.name === 'MODEL_URI' && e.value === payload));
  const custom = draft('AIJob', { purpose: 'batch', example: false, image: 'worker:v1', command: payload, isolation: { resourceQuota: { cpu: '999' } } });
  assert.equal(document(custom).spec.properties.args[0], payload);
  assert.equal(document(custom).spec.isolation, undefined);
  const before = JSON.stringify(custom);
  generate(custom);
  assert.equal(JSON.stringify(custom), before);
});
