const {test} = require('node:test');
const assert = require('node:assert/strict');
require('./ui.js');
const U = globalThis.ConsoleUI;

test('base path is the served prefix, an unreplaced placeholder counting as the root', () => {
  assert.equal(U.basePath(''), '');
  assert.equal(U.basePath('__BASE_PATH__'), '');
  assert.equal(U.basePath('/ai-console/'), '/ai-console');
  assert.equal(U.basePath('ai-console'), '/ai-console');
});
test('embed mode is only an explicit embed=1', () => {
  assert.equal(U.embedded('?embed=1'), true);
  assert.equal(U.embedded('?demo=local&embed=1'), true);
  assert.equal(U.embedded('?embed=0'), false);
  assert.equal(U.embedded(''), false);
});
test('paginate keeps the page in range and slices its rows', () => {
  const rows = Array.from({length: 23}, (_, i) => i);
  assert.deepEqual(U.paginate(rows, 3, 10), {rows: [20, 21, 22], total: 23, page: 3, pages: 3, size: 10});
  assert.equal(U.paginate(rows, 9, 10).page, 3);
  assert.equal(U.paginate(rows, 0, 10).page, 1);
  assert.deepEqual(U.paginate([], 4, 10), {rows: [], total: 0, page: 1, pages: 1, size: 10});
});
test('page numbers show the ends and the neighbours, gaps between', () => {
  assert.deepEqual(U.pageNumbers(1, 5), [1, 2, 3, 4, 5]);
  assert.deepEqual(U.pageNumbers(6, 12), [1, null, 5, 6, 7, null, 12]);
  assert.deepEqual(U.pageNumbers(2, 12), [1, 2, 3, null, 12]);
});
test('pager states the total and disables moves past either end', () => {
  const html = U.pager(U.paginate(Array.from({length: 15}), 1, 10));
  assert.match(html, /共 15 条/);
  assert.match(html, /data-page="0" disabled/);
  assert.match(html, /data-page="2"/);
  assert.match(html, /<option value="10" selected>/);
});
test('row links fold past the visible few, but never fold a lone extra', () => {
  const links = n => Array.from({length: n}, (_, i) => ({action: 'a' + i, label: 'L' + i}));
  assert.doesNotMatch(U.rowLinks(links(4), 3), /更多/);
  const folded = U.rowLinks(links(5), 3);
  assert.match(folded, /更多/);
  assert.ok(folded.indexOf('data-action="a2"') < folded.indexOf('更多'));
  assert.ok(folded.indexOf('data-action="a3"') > folded.indexOf('更多'));
  assert.doesNotMatch(U.rowLinks([{action: 'x', label: 'X', hidden: true}, false, {action: 'y', label: 'Y'}]), /data-action="x"/);
});
test('row links and crumbs escape what they show', () => {
  assert.match(U.rowLinks([{href: '#/jobs/a"b', label: '<x>'}]), /href="#\/jobs\/a&quot;b">&lt;x&gt;/);
  assert.match(U.crumbs([['<t>', '#/'], ['n']]), /&lt;t&gt;/);
  assert.match(U.crumbs([['a', '#/'], ['b']]), /<span aria-current="page">b<\/span>/);
});
test('the console tells its embedder which page is open, only when embedded', () => {
  const sent = [];
  const parent = {postMessage: (m, origin) => sent.push([m, origin])};
  const prior = {location: globalThis.location, parent: globalThis.parent};
  globalThis.location = {search: '?embed=1', hash: '#/jobs'};
  globalThis.parent = parent;
  try {
    U.announce('任务');
    globalThis.location = {search: '', hash: '#/jobs'};
    U.announce('任务');
  } finally {
    globalThis.location = prior.location; globalThis.parent = prior.parent;
  }
  assert.deepEqual(sent, [[{source: 'ai-console', type: 'route', hash: '#/jobs', title: '任务'}, '*']]);
});
test('the console follows only its embedder\'s navigate messages to hash routes', () => {
  let handler; const target = {addEventListener: (type, fn) => {handler = fn;}};
  const went = [];
  U.onNavigate(target, hash => went.push(hash));
  handler({data: {source: 'velaux', type: 'navigate', hash: '#/models'}});
  handler({data: {source: 'other', type: 'navigate', hash: '#/jobs'}});
  handler({data: {source: 'velaux', type: 'navigate', hash: 'javascript:alert(1)'}});
  handler({data: null});
  assert.deepEqual(went, ['#/models']);
});
