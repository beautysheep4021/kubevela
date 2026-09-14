const {test,before,after}=require('node:test');
const assert=require('node:assert/strict');
const {chromium}=require('playwright');
const {spawn}=require('node:child_process');
const fs=require('node:fs');
const os=require('node:os');
const path=require('node:path');
const base='http://127.0.0.1:18098';
let server,browser,dir;
before(async()=>{
  dir=fs.mkdtempSync(path.join(os.tmpdir(),'dataset-upload-e2e-'));
  server=spawn(process.env.AI_NORTHBOUND_BINARY||'/private/tmp/ai-northbound-assets',['-addr','127.0.0.1:18098','-dataset-dir',dir],{stdio:['ignore','pipe','pipe']});
  await new Promise((resolve,reject)=>{
    const timer=setTimeout(()=>reject(new Error('upload test server did not start')),10000);
    server.once('error',error=>{clearTimeout(timer);reject(error);});
    server.once('exit',code=>{clearTimeout(timer);reject(new Error('server exited '+code));});
    server.stderr.on('data',data=>{if(data.toString().includes('listening')){clearTimeout(timer);resolve();}});
  });
  browser=await chromium.launch();
  fs.mkdirSync(path.join(__dirname,'screenshots'),{recursive:true});
});
after(async()=>{
  await browser?.close();
  if(server && server.exitCode===null){await new Promise(resolve=>{server.once('exit',resolve);server.kill('SIGTERM');});}
  if(dir)fs.rmSync(dir,{recursive:true,force:true});
});
async function login(username='admin',password='shiyong',demo=false){
  const context=await browser.newContext({viewport:{width:1440,height:900}});
  const page=await context.newPage();page.errors=[];page.on('pageerror',e=>page.errors.push(e.message));
  await page.goto(base+'/login'+(demo?'?demo=local':''));
  await page.locator('#username').fill(username);await page.locator('#password').fill(password);
  await page.getByRole('button',{name:'登录',exact:true}).click();await page.waitForURL(/\/user/);
  await page.locator('[data-nav=datasets]').click();
  await page.getByRole('heading',{name:'本地上传数据集',exact:true}).waitFor();
  return {page,context};
}
async function upload(page,name,content){
  await page.getByRole('button',{name:'上传数据集',exact:true}).click();
  await page.locator('dialog [name=name]').fill(name);
  await page.locator('dialog [name=file]').setInputFiles({name:'data.jsonl',mimeType:'application/octet-stream',buffer:Buffer.from(content)});
  await page.locator('dialog button[type=submit]').click();
}
test('actual JSONL files persist across reload and remain isolated between tenants',async()=>{
  const a=await login();
  let version;
  try{
    await upload(a.page,'customer-a','{"text":"private sample A"}\n{"text":"second record"}\n');
    await a.page.locator('dialog').waitFor({state:'hidden'});
    await a.page.locator('#uploaded-datasets').getByText('customer-a',{exact:true}).waitFor();
    await a.page.reload();
    await a.page.locator('#uploaded-datasets').getByText('customer-a',{exact:true}).waitFor();
    assert.match(await a.page.locator('#uploaded-datasets').innerText(),/2 行/);
    await a.page.screenshot({path:path.join(__dirname,'screenshots','dataset-uploads-desktop.png'),fullPage:true});
    await a.page.getByRole('button',{name:'查看数据预览',exact:true}).click();
    await a.page.locator('dialog').getByText('private sample A',{exact:true}).waitFor();
    await a.page.setViewportSize({width:390,height:844});
    assert.equal(await a.page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);
    await a.page.screenshot({path:path.join(__dirname,'screenshots','dataset-preview-mobile.png'),fullPage:true});
    const list=await a.context.request.get(base+'/api/v1/ai/dataset-uploads');
    const asset=(await list.json()).items[0];version=asset.version;
    assert.equal(asset.runtimeReady,false);assert.equal(asset.owner,'admin');
    assert.equal(fs.readFileSync(path.join(dir,'ai-tenant-a','customer-a',version,'data.jsonl'),'utf8'),'{"text":"private sample A"}\n{"text":"second record"}\n');
    const changed=await a.context.request.post(base+'/api/v1/ai/dataset-uploads',{multipart:{namespace:'ai-tenant-a',name:'customer-a',file:{name:'data.jsonl',mimeType:'application/octet-stream',buffer:Buffer.from('{"text":"revised A"}\n')}}});
    assert.equal(changed.status(),201);assert.notEqual((await changed.json()).version,version);
    const old=await a.context.request.get(base+'/api/v1/ai/dataset-uploads/ai-tenant-a/customer-a/'+version);
    assert.equal((await old.json()).preview[0].text,'private sample A');
    assert.deepEqual(a.page.errors,[]);
  }finally{await a.context.close();}
  const b=await login('tenant-b','tenant-b-123456');
  try{
    await b.page.getByText('暂无本地上传数据集',{exact:true}).waitFor();
    assert.doesNotMatch(await b.page.locator('#uploaded-datasets').innerText(),/customer-a/);
    const foreign=await b.context.request.get(base+'/api/v1/ai/dataset-uploads/ai-tenant-a/customer-a/'+version);assert.equal(foreign.status(),403);
    await upload(b.page,'customer-b','{"text":"private sample B"}\n');
    await b.page.locator('dialog').waitFor({state:'hidden'});
    await b.page.locator('#uploaded-datasets').getByText('customer-b',{exact:true}).waitFor();
    assert.deepEqual(b.page.errors,[]);
  }finally{await b.context.close();}
});
test('invalid JSON is rejected and demo never claims to save files',async()=>{
  const a=await login();
  try{
    await upload(a.page,'invalid-file','not json\n');
    await a.page.locator('#modal-error').getByText(/invalid dataset/).waitFor();
    const list=await a.context.request.get(base+'/api/v1/ai/dataset-uploads');
    assert.ok(!(await list.json()).items.some(a=>a.name==='invalid-file'));
    assert.deepEqual(a.page.errors,[]);
  }finally{await a.context.close();}
  const demo=await login('admin','shiyong',true);
  try{
    await demo.page.getByRole('button',{name:'上传数据集',exact:true}).click();
    assert.equal(await demo.page.locator('dialog [name=file]').count(),0);
    assert.match(await demo.page.locator('#uploaded-datasets').innerText(),/演示环境不支持真实上传/);
  }finally{await demo.context.close();}
});
