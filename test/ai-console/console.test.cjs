const {test,before,after} = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {chromium} = require('playwright');
const base = process.env.AI_CONSOLE_URL || 'http://127.0.0.1:18095';
let browser;
before(async()=>{browser=await chromium.launch({headless:true});fs.mkdirSync(path.join(__dirname,'screenshots'),{recursive:true});});
after(async()=>{await browser?.close();});
async function login(username='admin',password='shiyong',demo=true,width=1440) {
  const context=await browser.newContext({viewport:{width,height:960}});
  const page=await context.newPage();
  page.errors=[];page.on('pageerror',e=>page.errors.push(e.message));
  await page.goto(base+'/login'+(demo?'?demo=local':''));
  await page.locator('#username').fill(username);await page.locator('#password').fill(password);
  await page.getByRole('button',{name:'登录',exact:true}).click();
  await page.waitForURL(/\/user/);await page.getByRole('heading',{name:'工作台',exact:true}).waitFor();
  return {page,context};
}
async function next(page){await page.getByRole('button',{name:'下一步',exact:true}).click();}
async function prepareJob(page,name) {
  await page.getByRole('button',{name:'创建任务',exact:true}).first().click();
  await next(page);
  await page.locator('#draft-modelURI option').filter({hasText:'当前引用'}).count();
  await page.waitForFunction(()=>document.querySelector('#draft-modelURI')?.options.length>1);
  await page.locator('#draft-name').fill(name);
  await next(page);await next(page);
  await page.locator('#execution-mode').waitFor();
}
test('tenant lists, independent details, assets and responsive layout',async()=>{
  for(const [username,password,prefix,other] of [['admin','shiyong','tenant-a','tenant-b'],['tenant-b','tenant-b-123456','tenant-b','tenant-a']]) {
    const {page,context}=await login(username,password);
    try {
      await page.locator('.name-link').first().waitFor();
      assert.match(await page.locator('#page').innerText(),new RegExp(prefix));
      assert.doesNotMatch(await page.locator('#page').innerText(),new RegExp(other+'-(training|evaluation|inference|chat)'));
      await page.locator('[data-nav=jobs]').click();await page.locator('.name-link').first().click();
      await page.getByRole('tab',{name:'日志',exact:true}).click();await page.locator('#log-output').filter({hasText:'AI_RESULT_JSON'}).waitFor();
      await page.getByRole('tab',{name:'结果与产物'}).click();await page.getByRole('button',{name:'读取运行结果'}).click();
      await page.getByRole('button',{name:'登记模型',exact:true}).waitFor();
      await page.locator('[data-nav=models]').click();await page.locator('tbody tr').first().waitFor();
      await page.locator('[data-nav=datasets]').click();await page.locator('tbody tr').first().waitFor();
      await page.locator('[data-nav=services]').click();await page.locator('.name-link').first().click();
      await page.getByRole('tab',{name:'服务访问'}).click();await page.getByRole('button',{name:'检查服务访问'}).click();
      await page.getByText(/访问检查通过/).waitFor();
      await page.locator('[data-nav=overview]').click();await page.locator('.name-link').first().waitFor();
      await page.screenshot({path:path.join(__dirname,'screenshots',prefix+'-desktop.png'),fullPage:true});
      for(const width of [1024,390]) {
        await page.setViewportSize({width,height:844});
        assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);
        await page.screenshot({path:path.join(__dirname,'screenshots',prefix+'-'+width+'.png'),fullPage:true});
      }
      assert.deepEqual(page.errors,[]);
    }finally{await context.close();}
  }
});
test('four-step job preserves inputs, dry run creates nothing, actual submit creates once',async()=>{
  const {page,context}=await login();
  let writes=0;page.on('request',r=>{if(r.url().includes('/normalize')&&r.method()==='POST')writes++;});
  try {
    await prepareJob(page,'console-p1-training');
    await page.setViewportSize({width:390,height:844});
    assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);
    await page.screenshot({path:path.join(__dirname,'screenshots','create-mobile.png'),fullPage:true});
    await page.setViewportSize({width:1440,height:960});
    await page.getByRole('button',{name:'上一步'}).click();
    assert.equal(await page.locator('#draft-cpu').inputValue(),'1');await next(page);
    await page.getByRole('button',{name:'校验配置'}).click();
    await page.getByText('配置校验通过，尚未创建任务或服务。').waitFor();
    assert.match(page.url(),/jobs\/new/);
    await page.locator('#execution-mode').selectOption('create');
    const before=writes;await page.getByRole('button',{name:'创建任务',exact:true}).click();
    await page.waitForURL(/jobs\/ai-tenant-a\/console-p1-training/);
    await page.getByRole('heading',{name:'console-p1-training',exact:true}).waitFor();
    assert.equal(writes-before,1);
    await page.getByRole('tab',{name:'结果与产物'}).click();await page.getByRole('button',{name:'读取运行结果'}).click();
    await page.getByRole('button',{name:'发布为服务'}).click();
    await page.locator('dialog [name=serviceName]').fill('tenant-a-inference-service');
    await page.getByRole('button',{name:'发布服务',exact:true}).click();
    await page.getByText('同名任务或服务已存在，请更换名称。',{exact:true}).waitFor();
    await page.locator('dialog [name=serviceName]').fill('console-p1-training-service');
    await page.getByRole('button',{name:'发布服务',exact:true}).click();
    await page.waitForURL(/services\/ai-tenant-a\/console-p1-training-service/);
    await page.getByRole('heading',{name:'console-p1-training-service',exact:true}).waitFor();
    await page.screenshot({path:path.join(__dirname,'screenshots','service-detail.png'),fullPage:true});
    assert.deepEqual(page.errors,[]);
  }finally{await context.close();}
});
test('filter history, validation focus, keyboard navigation and delete confirmation',async()=>{
  const {page,context}=await login();
  try {
    await page.locator('[data-nav=jobs]').click();await page.locator('#search').fill('training');
    await page.locator('#search').press('Enter');await page.locator('#search').press('Tab');
    await page.waitForURL(/q=training/);
    await page.locator('.name-link').first().click();await page.goBack();await page.locator('#search').waitFor();
    assert.equal(await page.locator('#search').inputValue(),'training');
    await page.getByRole('button',{name:'创建任务',exact:true}).click();await next(page);await next(page);
    await page.locator('#draft-name').waitFor();
    assert.equal(await page.locator('#draft-name').evaluate(el=>el===document.activeElement),true);
    await page.locator('#draft-name').press('Tab');
    assert.notEqual(await page.evaluate(()=>document.activeElement.tagName),'BODY');
    page.on('dialog',dialog=>dialog.accept());
    await page.locator('[data-nav=jobs]').click();await page.locator('.name-link').first().click();
    await page.getByRole('button',{name:'删除任务',exact:true}).click();await page.locator('dialog').waitFor();
    await page.keyboard.press('Escape');
    assert.equal(await page.locator('dialog').isVisible(),false);
    await page.getByRole('button',{name:'删除任务',exact:true}).click();
    await page.locator('dialog').getByRole('button',{name:'删除',exact:true}).click();
    await page.getByText('暂无记录',{exact:true}).waitFor();
  }finally{await context.close();}
});
test('polling preserves search focus and unconfirmed writes remain queryable',async()=>{
  const {page,context}=await login();
  try {
    await page.clock.install();
    await page.locator('[data-nav=jobs]').click();await page.locator('#search').waitFor();
    await page.locator('#search').fill('unfinished-search');
    await page.clock.fastForward(11000);
    await page.waitForLoadState('networkidle');
    assert.equal(await page.locator('#search').inputValue(),'unfinished-search');
    assert.equal(await page.locator('#search').evaluate(el=>el===document.activeElement),true);
    await prepareJob(page,'unconfirmed-job');
    await page.locator('#execution-mode').selectOption('create');
    await page.route('**/normalize',route=>route.abort());
    await page.getByRole('button',{name:'创建任务',exact:true}).click();
    await page.getByRole('button',{name:'查询创建结果'}).waitFor();
    for(let i=0;i<2;i++){
      await page.getByRole('button',{name:'查询创建结果'}).click();
      await page.getByText('当前未查到同名资源，原请求仍可能完成。请稍后再次查询创建结果。',{exact:true}).waitFor();
      assert.equal(await page.getByRole('button',{name:'创建任务',exact:true}).isDisabled(),true);
    }
  }finally{await context.close();}
});
test('service workflow, direct foreign detail denial and asset registration',async()=>{
  const {page,context}=await login('tenant-b','tenant-b-123456');
  try {
    await page.getByRole('button',{name:'部署服务',exact:true}).first().click();await next(page);
    await page.waitForFunction(()=>document.querySelector('#draft-modelURI')?.options.length>1);
    await page.locator('#draft-name').fill('tenant-b-new-service');await next(page);await next(page);
    await page.locator('#execution-mode').selectOption('create');await page.getByRole('button',{name:'部署服务',exact:true}).click();
    await page.waitForURL(/services\/ai-tenant-b\/tenant-b-new-service/);
    await page.locator('[data-nav=datasets]').click();await page.getByRole('button',{name:'登记数据集',exact:true}).click();
    await page.locator('dialog [name=name]').fill('tenant-b-extra-data');await page.locator('dialog [name=uri]').fill('dataset://ai-tenant-b/extra/v1');
    await page.getByRole('button',{name:'登记',exact:true}).click();await page.getByText('tenant-b-extra-data',{exact:true}).waitFor();
    await page.goto(base+'/user?demo=local#/jobs/ai-tenant-a/tenant-a-training-job');
    await page.getByText('当前账号无权访问此资源。',{exact:true}).waitFor();
    assert.equal(await page.locator('.detail-body').count(),0);assert.deepEqual(page.errors,[]);
  }finally{await context.close();}
});
test('live API errors show retry instead of mock data; identity change clears content',async()=>{
  const {page,context}=await login('admin','shiyong',false);
  try {
    await page.getByText('当前服务未连接所需的集群或资产存储。',{exact:true}).waitFor();
    assert.equal(await page.locator('.name-link').count(),0);
    await page.getByRole('button',{name:'重试',exact:true}).click();
    await page.getByText('当前服务未连接所需的集群或资产存储。',{exact:true}).waitFor();
    await page.goto(base+'/user?demo=local');await page.locator('.name-link').first().waitFor();
    await context.request.post(base+'/logout');
    await context.request.post(base+'/login',{form:{role:'user',username:'tenant-b',password:'tenant-b-123456'}});
    await page.getByRole('button',{name:'刷新工作台',exact:true}).click();
    await page.getByRole('heading',{name:'会话已失效或账号已切换'}).waitFor();
    assert.equal(await page.locator('.name-link').count(),0);assert.deepEqual(page.errors,[]);
  }finally{await context.close();}
});
