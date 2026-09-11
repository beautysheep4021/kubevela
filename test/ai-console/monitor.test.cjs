const {test,before,after}=require('node:test');
const assert=require('node:assert/strict');
const {chromium}=require('playwright');
const fs=require('node:fs');
const path=require('node:path');
const base=process.env.AI_CONSOLE_URL||'http://127.0.0.1:18096';
let browser;
before(async()=>{browser=await chromium.launch();fs.mkdirSync(path.join(__dirname,'screenshots'),{recursive:true});});
after(async()=>{await browser?.close();});
async function login(demo=true){
  const context=await browser.newContext({viewport:{width:1440,height:960}});
  const page=await context.newPage();page.errors=[];page.on('pageerror',error=>page.errors.push(error.message));
  await page.goto(base+'/login'+(demo?'?demo=local':''));
  await page.locator('#role').selectOption('monitor');
  await page.locator('#username').fill('admin');await page.locator('#password').fill('jiankong');
  await page.getByRole('button',{name:'登录',exact:true}).click();
  await page.waitForURL(/\/monitor/);
  return {page,context};
}
test('monitor lists both tenants, filters scope and renders responsive pages',async()=>{
  const {page,context}=await login();
  try{
    await page.locator('a[href*="tenant-a-training-job"]').first().waitFor();
    assert.match(await page.locator('#page').innerText(),/tenant-b/);
    for(const width of [1440,1024,390]){
      await page.setViewportSize({width,height:900});
      assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);
      await page.screenshot({path:path.join(__dirname,'screenshots','monitor-'+width+'.png'),fullPage:true});
    }
    await page.goto(base+'/monitor?demo=local#/jobs?ns=ai-tenant-b');
    await page.locator('a[href*="tenant-b-training-job"]').first().waitFor();
    assert.equal(await page.locator('a[href*="tenant-a-training-job"]').count(),0);
    await page.goto(base+'/monitor?demo=local#/tenants');
    await page.getByText('requests.cpu',{exact:true}).first().waitFor();
    assert.match(await page.locator('#page').innerText(),/16Gi/);
    await page.screenshot({path:path.join(__dirname,'screenshots','monitor-quotas-mobile.png'),fullPage:true});
    assert.deepEqual(page.errors,[]);
  }finally{await context.close();}
});
test('monitor detail exposes scheduling evidence, service probe and job results',async()=>{
  const {page,context}=await login();
  try{
    await page.goto(base+'/monitor?demo=local#/jobs/ai-tenant-a/tenant-a-waiting-job');
    await page.getByText(/Insufficient cpu/).first().waitFor();
    await page.goto(base+'/monitor?demo=local#/jobs/ai-tenant-b/tenant-b-training-job?tab=logs');
    await page.getByText(/run-complete/).first().waitFor();
    await page.locator('#log-pod').selectOption('tenant-b-training-job-pod');
    await page.locator('#log-container').selectOption('main');
    await page.locator('#log-tail').fill('37');
    await page.getByRole('button',{name:'刷新详情',exact:true}).click();
    await page.getByText(/run-complete/).first().waitFor();
    assert.equal(await page.locator('#log-pod').inputValue(),'tenant-b-training-job-pod');
    assert.equal(await page.locator('#log-container').inputValue(),'main');
    assert.equal(await page.locator('#log-tail').inputValue(),'37');
    await page.screenshot({path:path.join(__dirname,'screenshots','monitor-logs.png'),fullPage:true});
    await page.getByRole('tab',{name:'结果与产物'}).click();
    await page.getByRole('button',{name:'读取运行结果'}).click();await page.getByText(/inline:\/\/models/).waitFor();
    await page.goto(base+'/monitor?demo=local#/services/ai-tenant-b/tenant-b-inference-service?tab=result');
    await page.getByRole('button',{name:'检查服务访问'}).click();await page.getByText(/访问检查通过/).waitFor();
    assert.deepEqual(page.errors,[]);
  }finally{await context.close();}
});

test('monitor confirms one lifecycle action and links audit to captured resource',async()=>{
  const {page,context}=await login();
  try{
    await page.goto(base+'/monitor?demo=local#/jobs/ai-tenant-b/tenant-b-training-job');
    await page.getByRole('button',{name:'删除任务',exact:true}).click();
    await page.locator('dialog').getByRole('button',{name:'取消',exact:true}).click();
    await page.getByRole('heading',{name:'tenant-b-training-job',exact:true}).waitFor();
    await page.getByRole('button',{name:'重新运行',exact:true}).click();
    await page.screenshot({path:path.join(__dirname,'screenshots','monitor-confirm.png'),fullPage:true});
    await page.locator('dialog').getByRole('button',{name:'确认重新运行',exact:true}).evaluate(button=>{button.click();button.click();});
    await page.getByText(/已提交，等待状态确认/).waitFor();
    await page.getByRole('button',{name:'刷新详情',exact:true}).click();
    await page.locator('.detail-status').getByText('等待中',{exact:true}).waitFor();
    assert.equal(await page.getByRole('button',{name:'重新运行',exact:true}).isDisabled(),true);
    await page.getByRole('button',{name:'确认已核查',exact:true}).click();
    assert.equal(await page.getByRole('button',{name:'重新运行',exact:true}).isEnabled(),true);
    await page.locator('[data-nav=audits]').click();
    await page.locator('.name-link').filter({hasText:'tenant-b-training-job'}).waitFor();
    assert.equal(await page.locator('.name-link').filter({hasText:'tenant-b-training-job'}).count(),1);
    assert.match(await page.locator('#page').innerText(),/admin/);
    await page.locator('.name-link').filter({hasText:'tenant-b-training-job'}).click();
    await page.getByRole('heading',{name:'tenant-b-training-job',exact:true}).waitFor();
    assert.match(page.url(),/ai-tenant-b/);
    assert.deepEqual(page.errors,[]);
  }finally{await context.close();}
});

test('late detail response cannot replace a newly selected list',async()=>{
  const {page,context}=await login(false);
  let release,started;
  const held=new Promise(resolve=>{release=resolve;});
  const requested=new Promise(resolve=>{started=resolve;});
  try{
    await page.route('**/applications/ai-tenant-b/slow-job/status',async route=>{
      started();await held;
      await route.fulfill({json:{name:'slow-job',namespace:'ai-tenant-b',kind:'AIJob',phase:'running',components:[]}});
    });
    await page.goto(base+'/monitor#/jobs/ai-tenant-b/slow-job');
    await requested;
    await page.locator('[data-nav=services]').click();
    await page.getByRole('heading',{name:'在线服务',exact:true}).waitFor();
    const response=page.waitForResponse('**/applications/ai-tenant-b/slow-job/status');
    release();await response;
    await page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))));
    assert.equal(await page.getByRole('heading',{name:'slow-job',exact:true}).count(),0);
    assert.equal(await page.getByRole('heading',{name:'在线服务',exact:true}).count(),1);
    assert.deepEqual(page.errors,[]);
  }finally{release();await context.close();}
});
test('missing detail returns to the original filtered list',async()=>{
  const {page,context}=await login();
  try{
    await page.goto(base+'/monitor?demo=local#/applications/ai-tenant-b/deleted-job?from=audits&ns=ai-tenant-b&actor=admin&action=delete&name=deleted-job');
    await page.getByRole('alert').waitFor();
    const href=await page.locator('.back-link').first().getAttribute('href');
    assert.match(href,/^#\/audits\?/);
    for(const value of ['ns=ai-tenant-b','actor=admin','action=delete','name=deleted-job'])assert.ok(href.includes(value));
    await page.locator('.back-link').first().click();
    await page.getByText('没有符合筛选条件的记录',{exact:true}).waitFor();
    assert.deepEqual(page.errors,[]);
  }finally{await context.close();}
});
test('live monitor errors do not become demo and user cannot access monitor quota API',async()=>{
  const {page,context}=await login(false);
  try{
    await page.getByRole('alert').first().waitFor();
    assert.equal(await page.locator('a[href*="tenant-a-training-job"]').count(),0);
    const response=await context.request.post(base+'/login',{form:{role:'user',username:'admin',password:'shiyong'},maxRedirects:0});
    assert.equal(response.status(),302);
    const quota=await context.request.get(base+'/api/v1/ai/tenant-resources');assert.equal(quota.status(),403);
    await page.reload();await page.waitForURL(/\/user/);
  }finally{await context.close();}
});

test('live detail preserves missing log target and does not mislabel upstream failures',async()=>{
  const {page,context}=await login(false);
  let removed=false,failed=false,logCalls=0;
  const app={name:'observed-job',namespace:'ai-tenant-b',kind:'AIJob',phase:'running',components:[{type:'ai-job',pods:[{name:'observed-pod',containers:[{name:'main'}]}]}]};
  try{
    await page.route('**/applications/ai-tenant-b/observed-job/status',route=>route.fulfill({status:failed?502:200,json:failed?{error:'pods "observed-pod" not found'}:{...app,components:removed?[]:app.components}}));
    await page.route('**/applications/ai-tenant-b/observed-job/logs?*',route=>{
      logCalls++;
      return route.fulfill({json:{pod:'observed-pod',container:'main',pods:[{name:'observed-pod',containers:['main']}],logs:'captured logs'}});
    });
    await page.goto(base+'/monitor#/jobs/ai-tenant-b/observed-job?tab=logs');
    await page.getByText('captured logs',{exact:true}).waitFor();
    removed=true;
    await page.getByRole('button',{name:'刷新详情',exact:true}).click();
    await page.getByText(/所选 Pod 已不可用/).waitFor();
    assert.equal(logCalls,1);
    assert.equal(await page.locator('#log-pod').inputValue(),'observed-pod');
    failed=true;
    await page.getByRole('button',{name:'刷新详情',exact:true}).click();
    await page.getByRole('alert').waitFor();
    assert.match(await page.getByRole('alert').innerText(),/observed-pod/);
    assert.doesNotMatch(await page.getByRole('alert').innerText(),/可能已被删除/);
    assert.deepEqual(page.errors,[]);
  }finally{await context.close();}
});
