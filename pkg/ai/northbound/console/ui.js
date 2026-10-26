(function (root) {
  'use strict';
  // The console's page parts as HTML, shared by the tenant and monitor
  // consoles, after the VelaUX console: a page guide, panels, a pager, row
  // links, breadcrumbs and a detail card. Builders are pure; the few document
  // listeners at the end only wire the guide fold, the sidebar and embedding.
  const E = value => String(value == null ? '' : value).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
  const icon = name => `<i data-lucide="${E(name)}" aria-hidden="true"></i>`;
  const store = {
    get(key) { try { return root.localStorage ? root.localStorage.getItem(key) : null; } catch (_) { return null; } },
    set(key, value) { try { if (root.localStorage) root.localStorage.setItem(key, value); } catch (_) { /* Kept for this page only. */ } },
  };
  // basePath is where the console is served, '' at the root; a placeholder
  // the server left unreplaced counts as the root.
  function basePath(value) {
    const v = String(value || '').trim();
    if (!v || /^__.*__$/.test(v)) return '';
    return v.replace(/\/+$/, '').replace(/^(?!\/)/, '/');
  }
  // embedded reports whether the page was opened inside another console.
  function embedded(search) { return new URLSearchParams(search || '').get('embed') === '1'; }
  // paginate is the slice of rows a page shows, the page kept in range.
  function paginate(rows, page, size) {
    const total = rows.length, per = size > 0 ? size : 10, pages = Math.max(1, Math.ceil(total / per));
    const current = Math.min(Math.max(1, page || 1), pages);
    return {rows: rows.slice((current - 1) * per, current * per), total, page: current, pages, size: per};
  }
  // pageNumbers lists the pages a pager shows, gaps as null.
  function pageNumbers(page, pages) {
    if (pages <= 7) return Array.from({length: pages}, (_, i) => i + 1);
    const out = [1], from = Math.max(2, page - 1), to = Math.min(pages - 1, page + 1);
    if (from > 2) out.push(null);
    for (let i = from; i <= to; i++) out.push(i);
    if (to < pages - 1) out.push(null);
    out.push(pages);
    return out;
  }
  function pager(p, sizes = [10, 20, 50]) {
    return `<div class="table-pager"><span class="pager-total">共 ${p.total} 条</span><label class="pager-size">每页<select data-page-size aria-label="每页条数">${sizes.map(s => `<option value="${s}"${s === p.size ? ' selected' : ''}>${s} 条</option>`).join('')}</select></label><div class="pager-pages"><button type="button" class="pager-item" data-page="${p.page - 1}"${p.page <= 1 ? ' disabled' : ''} aria-label="上一页">${icon('chevron-left')}</button>${pageNumbers(p.page, p.pages).map(n => n === null ? '<span class="pager-gap">…</span>' : `<button type="button" class="pager-item${n === p.page ? ' current' : ''}" data-page="${n}"${n === p.page ? ' aria-current="page"' : ''}>${n}</button>`).join('')}<button type="button" class="pager-item" data-page="${p.page + 1}"${p.page >= p.pages ? ' disabled' : ''} aria-label="下一页">${icon('chevron-right')}</button></div></div>`;
  }
  // rowLinks are a row's operations as text links; past visible they fold
  // under 更多, though a lone extra shows in its place.
  function rowLinks(links, visible = 3) {
    const kept = links.filter(l => l && !l.hidden);
    const split = kept.length <= visible + 1 ? kept.length : visible;
    const link = l => l.href
      ? `<a class="row-link${l.danger ? ' danger' : ''}" href="${E(l.href)}"${l.aria ? ` aria-label="${E(l.aria)}"` : ''}>${E(l.label)}</a>`
      : `<button type="button" class="row-link${l.danger ? ' danger' : ''}" data-action="${E(l.action)}"${l.disabled ? ' disabled' : ''}${l.reason ? ` title="${E(l.reason)}"` : ''}${l.aria ? ` aria-label="${E(l.aria)}"` : ''}>${E(l.label)}</button>`;
    const more = kept.slice(split);
    return `<span class="row-links">${kept.slice(0, split).map(link).join('')}${more.length ? `<details class="row-more"><summary class="row-link">更多${icon('chevron-down')}</summary><div class="row-more-menu">${more.map(link).join('')}</div></details>` : ''}</span>`;
  }
  function crumbs(items) {
    return `<nav class="crumbs" aria-label="面包屑">${items.map(([label, href], i) => (i ? '<span class="crumb-sep">/</span>' : '') + (href ? `<a href="${E(href)}">${E(label)}</a>` : `<span aria-current="page">${E(label)}</span>`)).join('')}</nav>`;
  }
  // guide opens a list page with what it is for and its steps, folded away
  // once read; the fold is remembered per guide.
  function guide(id, title, description, steps, extra = '') {
    const folded = store.get('ai-guide-' + id) === '1';
    return `<section class="page-guide${folded ? ' folded' : ''}" data-guide="${E(id)}"><div class="page-guide-head"><div><h1>${E(title)}</h1><p>${E(description)}</p></div><button type="button" class="page-guide-toggle" data-guide-toggle="${E(id)}" aria-expanded="${!folded}">${icon(folded ? 'chevron-down' : 'chevron-up')}<span>${folded ? '展开流程' : '收起流程'}</span></button></div>${extra}<ol class="page-guide-steps">${steps.map(([symbol, name, text], i) => `<li class="page-guide-step"><span class="page-guide-icon">${icon(symbol)}</span><strong><span class="page-guide-index">${i + 1}</span>${E(name)}</strong><span>${E(text)}</span></li>`).join('')}</ol></section>`;
  }
  // header is a page's plain header: title, description and actions.
  function header(title, description = '', actions = '') {
    return `<div class="page-header"><div><h1>${E(title)}</h1>${description ? `<p>${E(description)}</p>` : ''}</div>${actions ? `<div class="actions">${actions}</div>` : ''}</div>`;
  }
  function panel(body, title = '', actions = '', id = '') {
    return `<section class="console-panel"${id ? ` id="${E(id)}"` : ''}>${title || actions ? `<div class="console-panel-head"><h2>${E(title)}</h2><div class="actions">${actions}</div></div>` : ''}${body}</section>`;
  }
  function status(tone, label) { return `<span class="badge ${E(tone)}">${E(label)}</span>`; }
  // tip is an info mark that explains a field on hover and to screen readers.
  function tip(text) { return `<span class="tip" tabindex="0" title="${E(text)}" aria-label="${E(text)}">${icon('info')}</span>`; }

  // announce tells the console embedding this one which page is open.
  function announce(title) {
    if (!embedded(root.location && root.location.search) || root.parent === root) return;
    try { root.parent.postMessage({source: 'ai-console', type: 'route', hash: root.location.hash, title}, '*'); } catch (_) { /* No parent to tell. */ }
  }
  // onNavigate follows the embedding console's requests to open a page.
  function onNavigate(target, go) {
    target.addEventListener('message', event => {
      const m = event.data;
      if (!m || m.source !== 'velaux' || m.type !== 'navigate' || typeof m.hash !== 'string' || !/^#\//.test(m.hash)) return;
      go(m.hash);
    });
  }
  function wire(doc) {
    const html = doc.documentElement;
    if (store.get('ai-nav-collapsed') === '1') html.classList.add('nav-collapsed');
    const avatar = doc.querySelector('.avatar'), user = doc.body && doc.body.dataset.accountUsername;
    if (avatar && user && !/^__.*__$/.test(user)) avatar.textContent = user.slice(0, 1).toUpperCase();
    doc.addEventListener('click', event => {
      const fold = event.target.closest('[data-guide-toggle]');
      if (fold) {
        const id = fold.dataset.guideToggle, section = fold.closest('.page-guide'), folded = !section.classList.contains('folded');
        section.classList.toggle('folded', folded);fold.setAttribute('aria-expanded', String(!folded));
        fold.innerHTML = icon(folded ? 'chevron-down' : 'chevron-up') + `<span>${folded ? '展开流程' : '收起流程'}</span>`;
        store.set('ai-guide-' + id, folded ? '1' : '0');if (root.lucide) root.lucide.createIcons();
        return;
      }
      const group = event.target.closest('[data-nav-group]');
      if (group) { group.parentElement.classList.toggle('folded');group.setAttribute('aria-expanded', String(!group.parentElement.classList.contains('folded')));return; }
      if (event.target.closest('#nav-collapse')) {
        const collapsed = !html.classList.contains('nav-collapsed');
        html.classList.toggle('nav-collapsed', collapsed);store.set('ai-nav-collapsed', collapsed ? '1' : '0');
        return;
      }
      // A row's More menu closes once one of its links is used, or another opens.
      doc.querySelectorAll('details.row-more[open]').forEach(d => { if (!d.contains(event.target) || event.target.closest('.row-more-menu')) d.open = false; });
    });
  }
  const api = {basePath, embedded, paginate, pageNumbers, pager, rowLinks, crumbs, guide, header, panel, status, tip, icon, announce, onNavigate};
  root.ConsoleUI = api;
  if (root.document && root.document.documentElement) wire(root.document);
})(globalThis);
