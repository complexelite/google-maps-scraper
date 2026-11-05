// Run inside the Tools runtime (which supplies Chromium and puppeteer-core).
// GETs use the real scraper; POSTs use fixtures so no scraping jobs are created.
const assert = require('node:assert/strict');
const http = require('node:http');
const {createRequire} = require('node:module');
const puppeteer = createRequire('/app/package.json')('puppeteer-core');

(async () => {
  const upstream = process.env.SCRAPER_URL || 'http://tools-google-maps-scraper:8080';
  let submissions = 0;
  const proxy = http.createServer(async (req, res) => {
    if (req.method === 'POST') {
      req.resume();
      submissions++;
      if (submissions === 1) {
        res.writeHead(422, {'Content-Type': 'text/plain'});
        res.end('Please check the job settings.');
      } else {
        res.writeHead(200, {'Content-Type': 'text/html'});
        res.end('<tbody id="job-tbody"><tr><td>Browser submission check</td><td>pending</td></tr></tbody><div id="pagination-container" hx-swap-oob="true"></div>');
      }
      return;
    }
    const response = await fetch(upstream + req.url);
    res.writeHead(response.status, {'Content-Type': response.headers.get('content-type') || 'text/plain'});
    res.end(await response.text());
  });
  await new Promise(resolve => proxy.listen(0, '127.0.0.1', resolve));
  const browser = await puppeteer.launch({executablePath: '/usr/bin/chromium', headless: true, args: ['--no-sandbox', '--disable-dev-shm-usage']});
  try {
    const page = await browser.newPage();
    await page.goto(`http://127.0.0.1:${proxy.address().port}/google-maps-scraper/`, {waitUntil: 'networkidle0'});
    assert.ok(await page.$('#job-tbody'), 'initial jobs refresh removed the form target');
    await page.type('#name', 'Browser submission check');
    await page.type('#keywords', 'coffee in London');
    await page.click('button[type=submit]');
    await page.waitForFunction(() => document.getElementById('error-container').textContent.includes('Please check'));
    assert.equal(submissions, 1);
    assert.ok(await page.$('#job-tbody'), 'validation error removed the form target');
    await page.click('button[type=submit]');
    await page.waitForFunction(() => document.getElementById('job-tbody')?.textContent.includes('Browser submission check'));
    assert.equal(submissions, 2);
    assert.match(await page.$eval('#submission-status', el => el.textContent), /Job submitted/);
    assert.equal(await page.$eval('#error-container', el => el.textContent), '');
    console.log('PASS: initial refresh retains tbody; Start Scraping submits; errors are visible; retry inserts the new job and confirms submission.');
  } finally {
    await browser.close();
    await new Promise(resolve => proxy.close(resolve));
  }
})().catch(error => { console.error(error); process.exitCode = 1; });
