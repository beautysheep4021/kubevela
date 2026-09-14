const {test} = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');

const example = {namespace:'team',name:'team-numeric-model',source:'numeric-example',status:'ready',version:'v1',modelURI:'configmap://team/example/model.json'};
function app({demo=false,section='models',respond=async()=>({namespace:'team',modelName:example.name,datasetName:'team-numeric-data'})}={}) {
  const nodes=new Map(), listeners={}, calls=[];
  const getNode=id=>{
    if(!nodes.has(id))nodes.set(id,{innerHTML:'',isConnected:true,listeners:{},classList:{add(){},remove(){}},setAttribute(){},insertAdjacentHTML(position,html){this.innerHTML+=html;},addEventListener(type,fn){this.listeners[type]=fn;},querySelector(selector){return this.children?.[selector]||this;},showModal(){this.open=true;this.children={'#modal-form':new FormData(),'[type="submit"]':{isConnected:true},'#modal-error':{innerHTML:''}};},close(){this.open=false;}});
    return nodes.get(id);
  };
  const location={search:demo?'?demo=local':'',hash:'#/'+section};
  const context=vm.createContext({ConsoleState:{example:()=>false,status:()=>({tone:'neutral',label:'待确认'}),escape:value=>String(value??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c])),route:()=>({section})},
    ConsoleAPI:{create:()=>({request:async(...args)=>{calls.push(args);return respond(...args);}})},
    document:{body:{dataset:{accountNamespace:'team',accountTenant:'tenant',accountUsername:'user'}},getElementById:getNode,querySelector:selector=>getNode(selector.replace(/^#/,'').replace(/.*\[data-action="?([^\]"]+).*/, '$1-button')),addEventListener(type,fn){listeners[type]=fn;}},location,window:{addEventListener(){}},
    URLSearchParams,FormData:class extends FormData{constructor(input){super();if(input)for(const [k,v]of input)this.set(k,v);}},setTimeout:()=>0,clearTimeout});
  const source=fs.readFileSync(__dirname+'/app.js','utf8');
  vm.runInContext(fs.readFileSync(__dirname+'/create.js','utf8'),context);
  vm.runInContext(source.replace('  loadRoute();icons();','  globalThis.testApp={assetTable,assetDetail,registerAsset,renderCollection,expired,showModal,loadAssetOptions,renderResult,renderDetail,setDetail:(value,output)=>{detail=value;result=output;route={section:value.kind===\'AIJob\'?\'jobs\':\'services\',detail:true,namespace:value.namespace,name:value.name,tab:\'result\'};routeVersion++;},startWizard:(kind,values={})=>{route.create=true;draft=Object.assign(ConsoleCreate.defaults(kind,identity.namespace,identity.tenant),values);step=1;},getDraft:()=>draft,setItems:value=>{items=value;}};'),context);
  const modal=getNode('modal');
  return {api:context.testApp,calls,getNode,modal,location,
    click:async(action,index=0)=>{const target={dataset:{action},disabled:action==='numeric-predict'?getNode('numeric-predict-button').disabled:false,closest:()=>({querySelector:()=>({dataset:{index}})})};return listeners.click({target:{closest:selector=>selector==='[data-action]'?target:null}});},
    submit:()=>{assert.equal(modal.open,true,'confirmation modal opens before submitting');return modal.listeners.submit({preventDefault(){},target:modal.querySelector('#modal-form')});}};
}
test('asset pages expose example imports and clear existing-registration labels',()=>{
  for(const section of ['models','datasets']){
    const h=app({section});h.api.renderCollection();
    assert.match(h.getNode('page').innerHTML,section==='models'?/使用示例模型/:/使用示例数据/);
    assert.match(h.getNode('page').innerHTML,section==='models'?/登记已有模型/:/登记已有数据集/);
    h.api.registerAsset(section==='models');
    assert.match(h.modal.innerHTML,section==='models'?/模型文件地址/:/数据文件地址/);
    assert.doesNotMatch(h.modal.innerHTML,/存储引用/);
  }
});
test('creation registration forms use explicit file-address labels',async()=>{
  for(const kind of ['model','dataset']){
    const h=app();await h.click('register-create-'+kind);
    assert.match(h.modal.innerHTML,kind==='model'?/登记已有模型/:/登记已有数据集/);
    assert.match(h.modal.innerHTML,kind==='model'?/模型文件地址/:/数据文件地址/);
    assert.doesNotMatch(h.modal.innerHTML,/存储引用/);
  }
});
test('only ready numeric assets offer prediction; other assets remain registered',()=>{
  for(const section of ['models','datasets'])for(const asset of [example,{...example,status:'registered'},{...example,source:'manual'}]){
    const h=app({section});h.api.setItems([asset]);const html=h.api.assetTable();
    if(asset.source==='numeric-example'&&asset.status==='ready'){
      assert.match(html,/可运行/);assert.match(html,/运行预测任务/);
      if(section==='models')assert.match(html,/部署预测服务/);else assert.doesNotMatch(html,/部署预测服务/);
      assert.doesNotMatch(html,/用于训练|asset-eval|asset-deploy/);
    }else{assert.match(html,/仅登记/);assert.doesNotMatch(html,/运行预测任务|部署预测服务|可运行/);}
  }
});
test('both import entries confirm matching assets and send only initialize',async()=>{
  for(const section of ['models','datasets']){
    const h=app({section});await h.click('numeric-initialize');
    assert.match(h.modal.innerHTML,/轻量数值预测示例/);assert.match(h.modal.innerHTML,/模型.*数据/);assert.match(h.modal.innerHTML,/非.*大语言模型/);
    assert.equal(h.calls.length,0);await h.submit();
    assert.equal(h.calls[0][0],'/api/v1/ai/examples/numeric');
    assert.equal(h.calls[0][1].method,'POST');assert.deepEqual(JSON.parse(h.calls[0][1].body),{action:'initialize'});
    assert.match(h.getNode('toast').textContent,/team-numeric-model.*team-numeric-data/);
  }
});
test('generic wizards exclude numeric examples, retain ordinary assets, and choose valid defaults',async()=>{
  const numericData={...example,name:'numeric-data',datasetURI:'configmap://team/example/data.jsonl'};
  for(const kind of ['AIJob','AIService'])for(const onlyExamples of [false,true])for(const selected of [false,true]){
    const model={namespace:'team',name:'ordinary-model',modelURI:'s3://models/ordinary'};
    const dataset={namespace:'team',name:'ordinary-data',datasetURI:'s3://data/ordinary'};
    const h=app({respond:async path=>({items:path.includes('/models')?[example,{...example,status:'failed'},...(onlyExamples?[]:[model]),{...model,namespace:'other',modelURI:'s3://foreign-model'}]:[numericData,...(onlyExamples?[]:[dataset]),{...dataset,namespace:'other',datasetURI:'s3://foreign-data'}]})});
    h.api.startWizard(kind,selected?{modelURI:example.modelURI,modelName:example.name,datasetURI:numericData.datasetURI}:{});
    await h.api.loadAssetOptions();
    const draft=h.api.getDraft(),html=h.getNode('create-fields').innerHTML;
    assert.equal(draft.modelURI,onlyExamples?'':model.modelURI);
    assert.equal(draft.datasetURI,onlyExamples?'':dataset.datasetURI);
    assert.doesNotMatch(html,/configmap:\/\/|numeric-model|numeric-data|foreign-/);
    if(!onlyExamples){assert.ok(html.includes(model.name));if(kind==='AIJob')assert.ok(html.includes(dataset.name));}
  }
});
test('wizard reload preserves ordinary custom references',async()=>{
  const h=app({respond:async()=>({items:[example]})});
  h.api.startWizard('AIJob',{modelURI:'s3://custom/model',datasetURI:'s3://custom/data'});
  await h.api.loadAssetOptions();
  assert.equal(h.api.getDraft().modelURI,'s3://custom/model');
  assert.equal(h.api.getDraft().datasetURI,'s3://custom/data');
});
test('asset lists and details preserve lifecycle status without claiming runtime readiness',()=>{
  for(const section of ['models','datasets'])for(const [status,label]of [['registered','仅登记'],['trained','已登记训练产物'],['published','已发布'],['released','已发布'],['failed','失败'],['ready','仅登记']]){
    const h=app({section}),asset={...example,source:'manual',status};
    h.api.setItems([asset]);assert.ok(h.api.assetTable().includes(label));
    h.api.assetDetail(asset);assert.ok(h.modal.innerHTML.includes(label));
    assert.doesNotMatch(h.api.assetTable(),/可运行/);assert.doesNotMatch(h.modal.innerHTML,/可运行/);
  }
});
test('prediction actions disclose CPU costs and navigate using returned application identity',async()=>{
  for(const [action,section,kind,target] of [['job','datasets','AIJob','jobs'],['job','models','AIJob','jobs'],['service','models','AIService','services']]){
    const h=app({section,respond:async()=>({namespace:'team',modelName:example.name,datasetName:'team-numeric-data',applicationName:'actual-name',kind})});
    h.api.setItems([example]);await h.click('asset-numeric-'+action);
    assert.match(h.modal.innerHTML,/CPU/);assert.match(h.modal.innerHTML,/GPU/);assert.match(h.modal.innerHTML,/非.*大语言模型/);
    for(const value of ['100m','500m','64Mi','128Mi'])assert.ok(h.modal.innerHTML.includes(value),'resource confirmation contains '+value);
    if(action==='service')assert.match(h.modal.innerHTML,/持续/);
    assert.equal(h.calls.length,0);await h.submit();assert.deepEqual(JSON.parse(h.calls[0][1].body),{action});
    assert.equal(h.location.hash,'#/'+target+'/team/actual-name');assert.equal(h.modal.open,false);
  }
});
test('demo rejects imports and launches explicitly without any transport',async()=>{
  for(const action of ['numeric-initialize','asset-numeric-job','asset-numeric-service']){
    const h=app({demo:true});h.api.setItems([example]);await h.click(action);await h.submit();
    assert.match(h.modal.querySelector('#modal-error').innerHTML,/演示环境.*不支持/);
    assert.equal(h.calls.length,0);assert.equal(h.modal.open,true);
  }
});
test('modal blocks duplicate requests and preserves backend error for retry',async()=>{
  let reject;const pending=new Promise((_,no)=>{reject=no;});
  const h=app({respond:()=>pending});await h.click('numeric-initialize');
  const first=h.submit();await h.submit();assert.equal(h.calls.length,1);
  assert.equal(h.modal.querySelector('[type="submit"]').disabled,true);
  reject(new Error('namespace does not exist'));await first;
  assert.equal(h.modal.open,true);assert.match(h.modal.querySelector('#modal-error').innerHTML,/namespace does not exist/);
  assert.equal(h.modal.querySelector('[type="submit"]').disabled,false);
});
test('late example completion after session expiry cannot navigate or show success',async()=>{
  let resolve;const h=app({respond:()=>new Promise(yes=>{resolve=yes;})});
  h.api.setItems([example]);await h.click('asset-numeric-job');const submit=h.submit();
  h.api.expired();h.api.showModal('new','',async()=>{});
  resolve({namespace:'team',applicationName:'late',kind:'AIJob'});await submit;
  assert.equal(h.location.hash,'#/models');assert.equal(h.getNode('toast').textContent,undefined);assert.equal(h.modal.open,true);
});
test('malformed or foreign launch responses cannot navigate',async()=>{
  for(const response of [{namespace:'other',applicationName:'private',kind:'AIJob'},{namespace:'team',kind:'AIJob'},{namespace:'team',applicationName:'wrong',kind:'AIService'}]){
    const h=app({respond:async()=>response});h.api.setItems([example]);await h.click('asset-numeric-job');await h.submit();
    assert.equal(h.location.hash,'#/models');assert.equal(h.modal.open,true);assert.match(h.modal.querySelector('#modal-error').innerHTML,/响应/);
  }
});
test('asset details show file addresses and real numeric parameters without evaluation form',()=>{
  for(const section of ['models','datasets']){
    const h=app({section});h.api.assetDetail({...example,datasetURI:'configmap://team/example/data.jsonl'});
    assert.match(h.modal.innerHTML,section==='models'?/模型文件地址/:/数据文件地址/);
    assert.doesNotMatch(h.modal.innerHTML,/evaluationJobName|存储引用/);
    if(section==='models'){assert.match(h.modal.innerHTML,/y = 2x \+ 1/);assert.match(h.modal.innerHTML,/斜率/);assert.match(h.modal.innerHTML,/截距/);}
    else{
      assert.match(h.modal.innerHTML,/数据预览/);assert.match(h.modal.innerHTML,/<th>x<\/th><th>y<\/th>/);
      // Matches the five immutable data.jsonl records in numeric_example.go.
      for(const [x,y]of [[0,1],[1,3],[2,5],[3,7],[4,9]])assert.ok(h.modal.innerHTML.includes(`<tr><td>${x}</td><td>${y}</td></tr>`));
    }
  }
});
const numericDetail={namespace:'team',name:'numeric-job-team',kind:'AIJob',aiMetadata:{'ai.oam.dev/source':'numeric-example'}};
test('numeric results deploy the real service and hide internal addresses in advanced information',async()=>{
  const h=app();h.api.setDetail(numericDetail,{modelURI:example.modelURI,metrics:{mae:0}});h.api.renderResult();
  const html=h.getNode('result-output').innerHTML;
  assert.match(html,/部署预测服务/);assert.doesNotMatch(html,/data-action="publish-result"|data-action="register-result"/);
  assert.doesNotMatch(html.replace(/<details>[\s\S]*?<\/details>/g,''),/configmap:\/\//);
  assert.match(html,/<details><summary>高级信息<\/summary>/);
  await h.click('numeric-result-service');await h.submit();
  assert.equal(h.calls[0][0],'/api/v1/ai/examples/numeric');assert.deepEqual(JSON.parse(h.calls[0][1].body),{action:'service'});
});
test('ordinary results retain registration and generic publication',()=>{
  const h=app();h.api.setDetail({...numericDetail,aiMetadata:{}},{modelURI:'s3://ordinary'});h.api.renderResult();
  assert.match(h.getNode('result-output').innerHTML,/data-action="publish-result"/);
  assert.match(h.getNode('result-output').innerHTML,/data-action="register-result"/);
});
test('numeric asset addresses only appear inside collapsed advanced information',()=>{
  for(const section of ['models','datasets']){
    const h=app({section});h.api.assetDetail({...example,datasetURI:'configmap://team/example/data.jsonl'});
    assert.match(h.modal.innerHTML,/<details><summary>高级信息<\/summary>/);
    const normal=h.modal.innerHTML.replace(/<details>[\s\S]*?<\/details>/g,'');
    assert.doesNotMatch(normal,/configmap:\/\//);assert.match(normal,section==='models'?/model\.json/:/data\.jsonl/);
  }
});
test('prediction input is only offered on numeric services',()=>{
  for(const kind of ['AIJob','AIService'])for(const numeric of [true,false]){
    const h=app();h.api.setDetail({...numericDetail,kind,aiMetadata:numeric?numericDetail.aiMetadata:{}},null);h.api.renderDetail();
    const html=h.getNode('detail-content').innerHTML;
    if(kind==='AIService'&&numeric){assert.match(html,/id="numeric-x"/);assert.match(html,/data-action="numeric-predict"/);}
    else assert.doesNotMatch(html,/numeric-x|numeric-predict/);
  }
});
test('predict sends finite numeric input once and renders backend output safely',async()=>{
  let resolve;const h=app({respond:()=>new Promise(yes=>{resolve=yes;})});
  h.api.setDetail({...numericDetail,kind:'AIService'},null);h.api.renderDetail();h.getNode('numeric-x').value='2.5';
  const first=h.click('numeric-predict');await h.click('numeric-predict');
  assert.equal(h.calls.length,1);assert.deepEqual(JSON.parse(h.calls[0][1].body),{action:'predict',x:2.5});
  assert.equal(h.getNode('numeric-predict-button').disabled,true);
  resolve({x:2.5,y:6,prediction:6,modelURI:example.modelURI,note:'<img onerror=alert(1)>'});await first;
  const html=h.getNode('numeric-prediction').innerHTML;
  assert.match(html,/6/);assert.doesNotMatch(html,/<img/);assert.doesNotMatch(html.replace(/<details>[\s\S]*?<\/details>/g,''),/configmap:\/\//);
  assert.equal(h.getNode('numeric-predict-button').disabled,false);
});
test('predict rejects demo, nonfinite inputs, ordinary services and jobs without requests',async()=>{
  for(const [demo,kind,numeric,value]of [[true,'AIService',true,'1'],[false,'AIService',true,''],[false,'AIService',true,'Infinity'],[false,'AIService',true,'NaN'],[false,'AIService',false,'1'],[false,'AIJob',true,'1']]){
    const h=app({demo});h.api.setDetail({...numericDetail,kind,aiMetadata:numeric?numericDetail.aiMetadata:{}},null);h.api.renderDetail();h.getNode('numeric-x').value=value;
    await h.click('numeric-predict');assert.equal(h.calls.length,0);
    if(numeric&&kind==='AIService')assert.match(h.getNode('numeric-prediction').innerHTML,demo?/演示环境/:/有限数值/);
  }
});
test('predict errors release busy state and stale results cannot update a new detail',async()=>{
  let reject;const h=app({respond:()=>new Promise((_,no)=>{reject=no;})});
  h.api.setDetail({...numericDetail,kind:'AIService'},null);h.api.renderDetail();h.getNode('numeric-x').value='1';
  const first=h.click('numeric-predict');reject(new Error('service unavailable'));await first;
  assert.match(h.getNode('numeric-prediction').innerHTML,/service unavailable/);assert.equal(h.getNode('numeric-predict-button').disabled,false);
  const next=h.click('numeric-predict');h.api.setDetail({...numericDetail,kind:'AIService',name:'new-service'},null);h.getNode('numeric-prediction').innerHTML='new';
  reject(new Error('late error'));await next;assert.equal(h.getNode('numeric-prediction').innerHTML,'new');
});
