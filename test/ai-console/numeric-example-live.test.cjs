const {test}=require('node:test');
const assert=require('node:assert/strict');
const {chromium}=require('playwright');
const fs=require('node:fs');
const path=require('node:path');

// Read-only verification against an explicitly selected, already initialized server.
const base=process.env.AI_NUMERIC_LIVE_URL;
test('real numeric assets render on desktop/mobile and remain tenant-scoped',{skip:!base},async()=>{
  const browser=await chromium.launch();
  try {
    for(const [username,password,namespace,foreign] of [
      ['admin','shiyong','ai-tenant-a','ai-tenant-b'],
      ['tenant-b','tenant-b-123456','ai-tenant-b','ai-tenant-a'],
    ]) {
      const context=await browser.newContext({viewport:{width:1440,height:900}});
      try {
        const page=await context.newPage(),errors=[];
        page.on('pageerror',e=>errors.push(e.message));
        await page.goto(base+'/login');
        await page.locator('#username').fill(username);
        await page.locator('#password').fill(password);
        await page.getByRole('button',{name:'登录',exact:true}).click();
        await page.waitForURL(/\/user/);
        for(const collection of ['models','datasets']) {
          const response=await context.request.get(base+'/api/v1/ai/'+collection);
          assert.equal(response.status(),200);
          const items=(await response.json()).items;
          const example=items.find(item=>item.source==='numeric-example');
          assert.ok(example,'missing initialized '+collection);
          assert.equal(example.namespace,namespace);
          assert.ok(items.every(item=>item.namespace===namespace));
          const denied=await context.request.get(base+'/api/v1/ai/'+collection+'?namespace='+foreign);
          assert.equal(denied.status(),403);
          await page.locator('[data-nav='+collection+']').click();
          await page.locator('#collection').getByText(example.displayName||example.name,{exact:true}).waitFor();
          assert.doesNotMatch(await page.locator('#collection').innerText(),new RegExp(foreign));
          await page.getByRole('button',{name:'运行预测任务',exact:true}).waitFor();
          fs.mkdirSync(path.join(__dirname,'screenshots'),{recursive:true});
          for(const [width,height] of [[1440,900],[390,844]]) {
            await page.setViewportSize({width,height});
            assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);
            await page.screenshot({path:path.join(__dirname,'screenshots',namespace+'-'+collection+'-'+width+'.png'),fullPage:true});
          }
          await page.setViewportSize({width:1440,height:900});
        }
        await page.goto(base+'/user#/services/'+namespace+'/numeric-service-'+namespace+'?tab=result');
        await page.getByRole('tab',{name:'服务访问',exact:true}).click();
        await page.locator('#numeric-x').fill('4');
        await page.getByRole('button',{name:'预测',exact:true}).click();
        await page.locator('#numeric-prediction dd').filter({hasText:/^9$/}).waitFor();
        await page.locator('#numeric-x').fill('0');
        await page.getByRole('button',{name:'预测',exact:true}).click();
        await page.locator('#numeric-prediction dd').filter({hasText:/^0$/}).waitFor();
        await page.locator('#numeric-prediction dd').filter({hasText:/^1$/}).waitFor();
        await page.setViewportSize({width:390,height:844});
        assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);
        await page.screenshot({path:path.join(__dirname,'screenshots',namespace+'-prediction-mobile.png'),fullPage:true});
        assert.deepEqual(errors,[]);
      } finally {await context.close();}
    }
  } finally {await browser.close();}
});
