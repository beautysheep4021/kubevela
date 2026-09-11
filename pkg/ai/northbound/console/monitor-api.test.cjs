const {test}=require('node:test');
const assert=require('node:assert/strict');
const fs=require('node:fs');
const vm=require('node:vm');
const path=require('node:path');
function setup(demo=true){
  let identity={role:'monitor',username:'admin',namespace:'',tenant:''};
  const calls=[];
  const scope={URL,AbortController,setTimeout,clearTimeout,fetch:async(url)=>{
    calls.push(url);
    return new Response(JSON.stringify(url.endsWith('/session')?identity:{error:'Unavailable'}),{status:url.endsWith('/session')?200:503,headers:{'Content-Type':'application/json'}});
  }};
  for(const file of ['api.js','monitor-api.js'])if(fs.existsSync(path.join(__dirname,file)))vm.runInNewContext(fs.readFileSync(path.join(__dirname,file),'utf8'),scope);
  assert.equal(typeof scope.MonitorAPI?.create,'function');
  return {client:scope.MonitorAPI.create({username:'admin',demo}),calls,setIdentity:value=>identity=value};
}
test('monitor demo lists both namespaces and filters without mixing objects',async()=>{
  const {client}=setup();
  const items=(await client.request('/api/v1/ai/applications')).items;
  assert.ok(items.some(i=>i.namespace==='ai-tenant-a'));
  assert.ok(items.some(i=>i.namespace==='ai-tenant-b'));
  assert.ok((await client.request('/api/v1/ai/applications?namespace=ai-tenant-b')).items.every(i=>i.namespace==='ai-tenant-b'));
  await assert.rejects(client.request('/api/v1/ai/applications/ai-tenant-a/tenant-b-training-job/status'),{status:404});
});
test('demo lifecycle records submission, not recovery, with captured namespace',async()=>{
  const {client}=setup();
  await client.request('/api/v1/ai/applications/ai-tenant-b/tenant-b-training-job/rerun',{method:'POST'});
  const item=await client.request('/api/v1/ai/applications/ai-tenant-b/tenant-b-training-job/status');
  assert.equal(item.phase,'pending');
  const audit=(await client.request('/api/v1/ai/audits?namespace=ai-tenant-b')).items[0];
  assert.equal(audit.actor,'admin');assert.equal(audit.namespace,'ai-tenant-b');assert.equal(audit.action,'rerun');
  const quotas=await client.request('/api/v1/ai/tenant-resources');
  assert.equal(quotas.items.length,2);
});
test('changed role invalidates demo before mutation; live failures never use mock',async()=>{
  const test=setup();test.setIdentity({role:'user',username:'admin',namespace:'ai-tenant-a',tenant:'tenant-a'});
  await assert.rejects(test.client.request('/api/v1/ai/applications/ai-tenant-a/tenant-a-training-job',{method:'DELETE'}),{status:401});
  const live=setup(false);await assert.rejects(live.client.request('/api/v1/ai/applications'),{status:503});
});

test('demo service probe preserves requested path and rejects invalid paths',async()=>{
  const {client}=setup();
  const path='/api/v1/ai/applications/ai-tenant-b/tenant-b-inference-service/probe';
  const result=await client.request(path,{method:'POST',body:JSON.stringify({path:'/ready'})});
  assert.equal(result.path,'/ready');
  assert.equal(result.demo,true);
  await assert.rejects(client.request(path,{method:'POST',body:JSON.stringify({path:'https://example.com'})}),{status:400});
});

test('canceled operations do not alter data; demo responses are detached and invalid paths rejected',async()=>{
  const {client}=setup();
  const original=await client.request('/api/v1/ai/applications');
  original.items[0].name='changed-by-caller';
  const controller=new AbortController();controller.abort();
  await assert.rejects(client.request('/api/v1/ai/applications/ai-tenant-a/tenant-a-training-job',{method:'DELETE',signal:controller.signal}));
  const summary=await client.request('/api/v1/ai/applications/ai-tenant-a/tenant-a-training-job/status');
  assert.equal(summary.name,'tenant-a-training-job');
  assert.equal((await client.request('/api/v1/ai/audits')).items.length,0);
  await assert.rejects(client.request('https://example.com/api/v1/ai/applications'),{status:400});
  await assert.rejects(client.request('/api/v1/ai/../session'),{status:400});
});
