const {test} = require('node:test');
const assert = require('node:assert/strict');
require('./state.js');
const S = globalThis.ConsoleState;
test('completed Job is successful even when no container is ready', () => {
  assert.equal(S.status({healthy:false,phase:'running',components:[{type:'ai-job',workload:{succeeded:1,completed:true}}]}, 'AIJob').key, 'succeeded');
  assert.notEqual(S.status({phase:'running',components:[{type:'ai-job',workload:{succeeded:1,active:1}}]}, 'AIJob').key, 'succeeded');
});
test('unhealthy without evidence is not a failed job', () => {
  assert.notEqual(S.status({healthy:false,phase:'running'}, 'AIJob').key, 'failed');
  assert.notEqual(S.status({healthy:false,components:[{workload:{failed:1,active:1}}]}, 'AIJob').key, 'failed');
});
test('service ready and job running remain different', () => {
  assert.equal(S.status({healthy:true},'AIService').key, 'ready');
  assert.equal(S.status({healthy:true},'AIJob').key, 'running');
});
test('route handles detail paths, tabs and filters independently', () => {
  const r = S.route('#/jobs/ai-tenant-b/test?tab=logs');
  assert.equal(r.name,'test'); assert.equal(r.namespace,'ai-tenant-b'); assert.equal(r.tab,'logs');
  assert.equal(S.route('#/jobs/new').create,true);
  assert.equal(S.route('#/unexpected').section,'overview');
  assert.equal(S.route('#/jobs/%E0%A4%A/test').invalid,true);
});
test('workload type derives from existing backend list contract', () => {
  assert.equal(S.kind({workloadTypes:['ai-job']}),'AIJob');
  assert.equal(S.kind({components:[{type:'ai-service'}]}),'AIService');
  assert.equal(S.kind({aiMetadata:{'ai.oam.dev/workload-kind':'job'}}),'AIJob');
});
test('escaping never renders supplied markup', () => {
  assert.equal(S.escape('<img onerror="x">'), '&lt;img onerror=&quot;x&quot;&gt;');
});
