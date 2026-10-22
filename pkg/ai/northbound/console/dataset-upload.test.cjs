const {test} = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');

function app(demo = false, respond = async () => ({items: []})) {
  const node = {dataset: {}, innerHTML:'', isConnected:true, classList: {add() {},remove() {}}, addEventListener(type, fn) {this.listeners[type]=fn;}, setAttribute() {}, insertAdjacentHTML(position, html) {this.innerHTML+=html;}, querySelector(selector) {return this.children?.[selector] || this;}, showModal() {this.open=true;this.children={'#modal-form':new FormData(),'[type="submit"]':{isConnected:true},'#modal-error':{innerHTML:''}};}, close() {this.open=false;}};
  const nodes = new Map();
  const getNode = id => {if(!nodes.has(id))nodes.set(id,{...node,listeners:{}});return nodes.get(id);};
  const calls = [];
  const context = vm.createContext({
    ConsoleState: {example: () => false, escape: value => String(value).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c])), route: () => ({section:'datasets'})},
    ConsoleAPI: {create: () => ({request: async (...args) => {calls.push(args); return respond(...args);}})},
    document: {body: {...node, dataset: {accountNamespace:'team',accountTenant:'tenant',accountUsername:'user'}}, getElementById: getNode, addEventListener() {}},
    location: {search: demo ? '?demo=local' : '',hash:'#/datasets'}, window: {addEventListener() {}},
    URLSearchParams, FormData: class extends FormData {constructor(input) {super();if(input)for(const [k,v] of input)this.set(k,v);}}, Blob, AbortController, setTimeout: () => 0, clearTimeout,
  });
  const source = fs.readFileSync(__dirname + '/app.js', 'utf8');
  vm.runInContext(fs.readFileSync(__dirname + '/ui.js', 'utf8'), context);
  vm.runInContext(source.replace('  loadRoute();icons();', '  globalThis.uploadTest = {uploadBody, uploadRequest, uploadedTable, refresh, loadUploads, showModal, expired, uploadDataset, previewUpload, state: () => ({modalBusy,uploadController}), setAssets: value => {uploadedAssets=value;}};'), context);
  return {api: context.uploadTest, calls, getNode};
}
function form(name = 'training-data', filename = 'data.jsonl', bytes = '{}\n') {
  const data = new FormData();
  data.set('name', name); data.set('file', new Blob([bytes]), filename);
  return data;
}
test('multipart contains only account namespace, trimmed name and original file', () => {
  const {api} = app(); const input = form(' training-data '); const body = api.uploadBody(input);
  assert.ok(body instanceof FormData);
  assert.deepEqual([...body.keys()], ['namespace','name','file']);
  assert.equal(body.get('namespace'), 'team'); assert.equal(body.get('name'), 'training-data');
  assert.equal(body.get('file'), input.get('file'));
});
test('rejects invalid names, empty files, wrong extensions and oversized files', () => {
  const {api} = app();
  for (const input of [form(''), form('../bad'), form('ok','data.csv'), form('ok','data.jsonl',''), form('ok','data.jsonl','x'.repeat(10*1024*1024+1))]) assert.throws(() => api.uploadBody(input));
  assert.doesNotThrow(() => api.uploadBody(form('ok','data.jsonl','x'.repeat(10*1024*1024))));
});
test('real upload passes FormData and cancellation signal without a content-type header', async () => {
  const {api,calls} = app(); const body = api.uploadBody(form()); const signal = new AbortController().signal;
  await api.uploadRequest('/api/v1/ai/dataset-uploads', {method:'POST',body,signal});
  assert.equal(calls[0][1].body,body); assert.equal(calls[0][1].signal,signal); assert.equal(calls[0][1].headers,undefined);
});
test('demo refuses upload and list requests without transport', async () => {
  const {api,calls} = app(true);
  await assert.rejects(api.uploadRequest('/api/v1/ai/dataset-uploads'), /演示/);
  await assert.rejects(api.uploadRequest('/api/v1/ai/dataset-uploads',{method:'POST',body:form()}), /演示/);
  assert.equal(calls.length,0);
});
test('503 reports storage configuration specifically', async () => {
  const {api} = app(false, async () => {throw Object.assign(new Error('disabled'),{status:503});});
  await assert.rejects(api.uploadRequest('/api/v1/ai/dataset-uploads'), /上传存储未配置/);
});
test('uploaded rows escape metadata, abbreviate versions and never offer execution', () => {
  const {api} = app(); const version = 'a'.repeat(64);
  api.setAssets([{name:'<unsafe>',namespace:'team',version,rows:0,bytes:3,format:'jsonl',status:'validated',runtimeReady:false}]);
  const html = api.uploadedTable();
  assert.match(html,/&lt;unsafe&gt;/); assert.ok(html.includes('title="'+version+'"'));
  assert.match(html,/未就绪/); assert.match(html,/本地持久化/); assert.match(html,/upload-preview/);
  assert.doesNotMatch(html,/asset-train|asset-eval|datasetURI/);
});
test('registered datasets survive disabled upload storage', async () => {
  const {api,getNode} = app(false, async path => {
    if(path.includes('/dataset-uploads'))throw Object.assign(new Error('disabled'),{status:503});
    return {items:[{namespace:'team',name:'external-data'}]};
  });
  await api.refresh(true); await api.loadUploads();
  assert.match(getNode('collection').innerHTML,/external-data/);
  assert.match(getNode('uploaded-datasets').innerHTML,/上传存储未配置/);
});
test('local uploads survive registered datasets being disabled without kubeconfig', async () => {
  const {api,getNode} = app(false, async path => {
    if(path.includes('/dataset-uploads'))return {items:[{namespace:'team',name:'local-data',version:'a'.repeat(64)}, {namespace:'other',name:'private-other'}]};
    throw Object.assign(new Error('no kubeconfig'),{status:503});
  });
  await api.refresh(true); await api.loadUploads();
  assert.match(getNode('uploaded-datasets').innerHTML,/local-data/);
  assert.doesNotMatch(getNode('uploaded-datasets').innerHTML,/private-other/);
  assert.match(getNode('page').innerHTML,/upload-dataset/);
  assert.match(getNode('page-error').innerHTML,/集群/);
  const html=getNode('page').innerHTML;
  assert.ok(html.indexOf('id="uploaded-datasets"')<html.indexOf('id="page-error"'));
  assert.ok(html.indexOf('id="uploaded-datasets"')<html.indexOf('id="collection"'));
});
function deferred() {
  let resolve, reject;
  const promise=new Promise((yes,no)=>{resolve=yes;reject=no;});
  return {promise,resolve,reject};
}
test('preview uses sample labels and explains omitted records only for nonempty data', async () => {
  for(const [rows,preview] of [[6,[{text:'sample'}]],[6,[]],[0,[]]]) {
    const asset={namespace:'team',name:'data',version:'a'.repeat(64),rows,preview};
    const {api,getNode}=app(false,async()=>asset);
    await api.previewUpload(asset);
    const html=getNode('upload-preview-content').innerHTML;
    assert.doesNotMatch(html,/第 1 行/);
    if(preview.length)assert.match(html,/样本 1/);
    else if(rows>0)assert.match(html,/前 5 条记录超出预览大小限制/);
    else {assert.match(html,/暂无预览数据/);assert.doesNotMatch(html,/超出预览大小限制/);}
  }
});
function submitModal(modal) {
  return modal.listeners.submit({preventDefault() {},target:modal.querySelector('#modal-form')});
}
for(const outcome of ['resolve','reject'])test(`late modal ${outcome} leaves newer submitting modal untouched`, async () => {
  const {api,getNode}=app(), old=deferred(), next=deferred(), modal=getNode('modal');
  api.showModal('old','',()=>old.promise);
  const first=submitModal(modal);
  api.showModal('new','',()=>next.promise);
  const second=submitModal(modal);
  old[outcome](new Error('old error'));
  await first;
  assert.equal(modal.open,true);
  assert.equal(api.state().modalBusy,true);
  assert.equal(modal.querySelector('[type="submit"]').disabled,true);
  assert.equal(modal.querySelector('#modal-error').innerHTML,'');
  next.resolve();await second;
  assert.equal(modal.open,false);
  assert.equal(api.state().modalBusy,false);
});
test('session expiry aborts upload and late settlement cannot refresh or close a newer modal', async () => {
  const pending=deferred();
  const {api,getNode,calls}=app(false,()=>pending.promise), modal=getNode('modal');
  api.uploadDataset();
  for(const [k,v] of form())modal.querySelector('#modal-form').set(k,v);
  const submitted=submitModal(modal);
  const signal=calls[0][1].signal;
  api.expired();
  assert.equal(signal.aborted,true);
  api.showModal('new','',async()=>{});
  pending.reject(Object.assign(new Error('aborted'),{name:'AbortError',status:0}));
  await submitted;
  assert.equal(calls.length,1);
  assert.equal(modal.open,true);
  assert.equal(modal.querySelector('#modal-error').innerHTML,'');
});
test('replaced upload cannot clear the new upload cancellation controller', async () => {
  const old=deferred(), next=deferred();let posts=0;
  const {api,getNode,calls}=app(false,async (path,init)=>init?.method==='POST'?(++posts===1?old.promise:next.promise):{items:[]});
  const modal=getNode('modal');
  function start() {
    api.uploadDataset();for(const [k,v] of form())modal.querySelector('#modal-form').set(k,v);
    return submitModal(modal);
  }
  const first=start(), second=start(), controller=api.state().uploadController;
  assert.equal(calls[0][1].signal.aborted,true);
  old.reject(Object.assign(new Error('aborted'),{name:'AbortError',status:0}));
  await first;
  assert.equal(api.state().uploadController,controller);
  assert.equal(api.state().modalBusy,true);
  assert.equal(modal.open,true);
  assert.equal(calls.length,2);
  next.resolve({});await second;
  assert.equal(modal.open,false);
  assert.equal(api.state().uploadController,null);
});
