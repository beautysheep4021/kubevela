(function () {
  "use strict";
  const S = ConsoleState,
    M = MonitorState,
    E = S.escape;
  const page = document.getElementById("page"),
    modal = document.getElementById("modal");
  const username = document.body.dataset.accountUsername;
  const demo = new URLSearchParams(location.search).get("demo") === "local";
  const root = "/api/v1/ai";
  const titles = {
    overview: "总览",
    jobs: "任务",
    services: "在线服务",
    tenants: "租户资源",
    audits: "操作审计",
    applications: "资源详情",
  };
  const actionLabels = {
    rerun: "重新运行",
    restart: "重启服务",
    delete: "删除资源",
  };
  let route = M.route(location.hash),
    version = 0,
    listSeq = 0,
    quotaSeq = 0,
    auxSeq = 0,
    dead = false;
  let items = [],
    detail = null,
    quotas = [],
    listError = "",
    quotaError = "",
    loading = false,
    quotaLoading = false,
    updated = "";
  let logPods = [],
    modalTarget = null,
    modalBusy = false,
    auxBusy = false;
  const actionStates = new Map();
  const logSelections = new Map();
  const client = MonitorAPI.create({ username, demo, onUnauthorized: expired });
  const request = (path, init) => client.request(path, init);
  const pathFor = (ns, name) =>
    `${root}/applications/${encodeURIComponent(ns)}/${encodeURIComponent(name)}`;
  const active = (v) => !dead && version === v;
  const icon = (name) => `<i data-lucide="${name}" aria-hidden="true"></i>`;
  const icons = () => {
    if (globalThis.lucide) lucide.createIcons();
  };
  const button = (action, label, symbol) =>
    `<button type="button" class="button" data-action="${action}">${icon(symbol)}${E(label)}</button>`;
  const iconButton = (action, label, symbol) =>
    `<button type="button" class="icon-button" data-action="${action}" aria-label="${E(label)}" title="${E(label)}">${icon(symbol)}</button>`;
  const empty = (label) =>
    `<div class="empty-state">${icon("inbox")}<h2>${E(label)}</h2></div>`;
  const busy = () => '<div class="loading" role="status">正在读取…</div>';
  const json = (value) =>
    `<pre class="config-output">${E(JSON.stringify(value, null, 2))}</pre>`;
  const time = (v) =>
    v && !Number.isNaN(Date.parse(v))
      ? new Date(v).toLocaleString("zh-CN", { hour12: false })
      : "未提供";
  const badge = (i) => {
    const s = S.status(i, S.kind(i));
    return `<span class="badge ${s.tone}">${E(s.label)}</span>`;
  };
  const sample = (i) =>
    demo || S.example(i) ? '<span class="badge warn">示例</span>' : "";
  const errorText = (e) =>
    e.status === 403
      ? "当前账号无权访问此资源。"
      : e.status === 404
        ? "未找到该资源，可能已被删除。"
        : e.status === 503
          ? "当前服务未连接所需的集群或存储。"
          : e.message || "请求失败。";
  const notice = (text, retry = "refresh") =>
    `<div class="notice error" role="alert"><span>${E(text)}</span>${retry ? button(retry, "重试", "refresh-cw") : ""}</div>`;
  const definition = (entries) =>
    `<dl class="detail-meta">${entries.map(([k, v]) => `<div><dt>${E(k)}</dt><dd>${E(v ?? "未提供")}</dd></div>`).join("")}</dl>`;
  const heading = (title, actions = "") =>
    `<div class="page-heading"><div><div class="eyebrow">智算纳管 / 监控台</div><h1>${E(title)}</h1></div><div class="actions">${actions}</div></div>`;
  const table = (headers, rows) =>
    rows.length
      ? `<div class="table-wrap"><table class="monitor-table"><thead><tr>${headers.map((h) => `<th scope="col">${E(h)}</th>`).join("")}</tr></thead><tbody>${rows.join("")}</tbody></table></div>`
      : empty("没有符合筛选条件的记录");
  const cell = (v) => `<td>${v}</td>`;
  const targetKey = (ns, name) => JSON.stringify([ns, name]);
  const detailLink = (i) => E(M.detailHash(i, route));
  const returnLink = () =>
    E(
      M.hash({
        ...route,
        detail: false,
        section:
          route.from ||
          (route.section === "applications" ? "audits" : route.section),
      }),
    );
  function updateScope() {
    const scope = document.getElementById("scope-current");
    scope.textContent = route.detail
      ? route.namespace
      : [route.ns || "全部命名空间", route.tenant ? "租户 " + route.tenant : ""]
          .filter(Boolean)
          .join(" / ");
    scope.title = scope.textContent;
  }
  function expired() {
    dead = true;
    version++;
    listSeq++;
    quotaSeq++;
    auxSeq++;
    items = [];
    detail = null;
    quotas = [];
    logPods = [];
    logSelections.clear();
    modalTarget = null;
    actionStates.clear();
    modal.close();
    modal.innerHTML = "";
    document.getElementById("toast").hidden = true;
    page.innerHTML =
      '<div class="empty-state"><h1>会话已失效或账号已切换</h1><a class="button" href="/login' +
      (demo ? "?demo=local" : "") +
      '">重新登录</a></div>';
  }
  function input(key, label, value, type = "text") {
    return `<label>${label}<input id="${key === "q" ? "search" : key + "-filter"}" data-filter="${key}" type="${type}" value="${E(value || "")}" autocomplete="off"></label>`;
  }
  function select(key, label, values, value) {
    return `<label>${label}<select id="${key}-filter" data-filter="${key}">${values.map(([v, l]) => `<option value="${E(v)}"${value === v ? " selected" : ""}>${E(l)}</option>`).join("")}</select></label>`;
  }
  function filters() {
    let html =
      input("q", "搜索名称", route.q, "search") +
      input("ns", "命名空间", route.ns);
    if (route.section === "audits")
      html =
        input("name", "对象名称", route.objectName, "search") +
        input("ns", "命名空间", route.ns) +
        input("actor", "操作人", route.actor) +
        input("action", "操作类型", route.action) +
        select(
          "outcome",
          "操作结果",
          [
            ["", "全部结果"],
            ["success", "成功"],
            ["failure", "失败"],
          ],
          route.outcome,
        );
    else
      html +=
        input("tenant", "租户", route.tenant) +
        select(
          "status",
          "筛选状态",
          [
            ["", "全部状态"],
            ["running", "运行中"],
            ["ready", "已就绪"],
            ["pending", "等待 / 部署中"],
            ["succeeded", "已完成"],
            ["failed", "失败"],
            ["unknown", "状态待确认"],
          ],
          route.status,
        );
    return `<div class="filters monitor-filters">${html}<span class="updated" id="updated"></span>${iconButton("refresh", "刷新列表", "refresh-cw")}</div>`;
  }
  function workTable(rows, evidence = false) {
    const headers = [
      "名称 / 命名空间",
      "租户",
      "状态",
      evidence ? "诊断线索" : "资源请求声明",
      "创建时间",
    ];
    return table(
      headers,
      rows.map((i) => {
        const d = M.diagnosticSummary(i);
        const clues = d
          ? `<a class="monitor-clue" href="${detailLink(i)}"
          title="${E([d.reason, d.message || d.evidence].filter(Boolean).join(": "))}">
          ${E(d.reason || d.message || d.evidence || "状态消息")}
        </a>`
          : "";
        return (
          "<tr>" +
          [
            `<a class="name-link" href="${detailLink(i)}">${E(i.name)}</a>
          <span class="subtle">${E(i.namespace)}</span>${sample(i)}`,
            E(M.tenant(i) || "未提供"),
            badge(i),
            evidence ? clues : resources(i.resourceSummary),
            E(time(i.createdAt)),
          ]
            .map(cell)
            .join("") +
          "</tr>"
        );
      }),
    );
  }
  function resources(r) {
    const parts = M.requestParts(r);
    return parts
      ? `<div class="monitor-resources">${parts.map((p) => `<span>${E(p)}</span>`).join("")}</div>`
      : "未提供";
  }
  function overview(rows) {
    const jobs = rows.filter((i) => S.kind(i) === "AIJob"),
      services = rows.filter((i) => S.kind(i) === "AIService"),
      abnormal = rows.filter((i) =>
        M.diagnostics(i).some(
          (d) => d.severity === "error" || d.severity === "warning",
        ),
      );
    return `<div class="summary-strip">${[
      [jobs.length, "任务总数"],
      [services.length, "服务总数"],
      [
        rows.filter((i) => S.status(i, S.kind(i)).key === "pending").length,
        "等待 / 部署中",
      ],
      [abnormal.length, "待核查项"],
    ]
      .map(
        ([n, l]) =>
          `<div class="summary-item"><strong>${n}</strong><span>${l}</span></div>`,
      )
      .join("")}</div>
      <div class="section-head"><h2>异常与待核查</h2></div>
      ${abnormal.length ? workTable(abnormal, true) : empty("暂无异常证据")}
      <div class="section-head"><h2>最近任务</h2></div>
      ${workTable(jobs.slice(0, 5))}
      <div class="section-head"><h2>在线服务</h2></div>
      ${workTable(services.slice(0, 5))}`;
  }
  function tenantTable(rows) {
    const groups = M.groupResources(rows);
    const totals = (t) => `${t.count} 个对象
      <div>${t.count === t.missing && t.count ? "未提供" : resources(t)}</div>
      ${t.missing ? `<span class="subtle">${t.missing} 个对象未提供声明；汇总不完整</span>` : ""}`;
    return (
      '<div class="section-head"><h2>任务与服务资源请求声明</h2></div>' +
      table(
        ["租户 / 命名空间", "未完成对象声明", "已完成任务声明"],
        groups.map(
          (g) =>
            "<tr>" +
            [
              `${E(g.tenant || "未提供")}<span class="subtle">${E(g.namespace)}</span>`,
              totals(g.other),
              totals(g.completed),
            ]
              .map(cell)
              .join("") +
            "</tr>",
        ),
      )
    );
  }
  function auditTable(rows) {
    return table(
      ["时间 / 操作人", "操作", "对象 / 命名空间", "结果", "消息"],
      rows.map(
        (i) =>
          "<tr>" +
          [
            `${E(time(i.time))}<span class="subtle">${E(i.actor || "未提供")}</span>`,
            E(i.action),
            i.namespace && i.name
              ? `<a class="name-link" href="${detailLink(i)}">${E(i.name)}</a>
        <span class="subtle">${E(i.namespace)}</span>`
              : E(i.name || "未提供"),
            `<span class="badge ${i.success === true ? "ok" : i.success === false ? "bad" : "neutral"}">
        ${i.success === true ? "成功" : i.success === false ? "失败" : "未提供"}</span>`,
            E(i.error || i.message || "未提供"),
          ]
            .map(cell)
            .join("") +
          "</tr>",
      ),
    );
  }
  function renderCollection() {
    const el = document.getElementById("collection");
    if (!el) return;
    document.getElementById("page-error").innerHTML = listError
      ? notice(listError)
      : "";
    document.getElementById("updated").textContent = updated
      ? "更新于 " + updated
      : "";
    el.setAttribute("aria-busy", String(loading));
    if (loading) el.innerHTML = busy();
    else if (listError) el.innerHTML = "";
    else {
      const rows = M.filter(items, route).filter((i) => S.kind(i));
      el.innerHTML =
        route.section === "overview"
          ? overview(rows)
          : route.section === "tenants"
            ? tenantTable(rows)
            : route.section === "audits"
              ? auditTable(
                  M.filterAudits(items, { ...route, name: route.objectName }),
                )
              : workTable(
                  rows.filter(
                    (i) =>
                      S.kind(i) ===
                      (route.section === "jobs" ? "AIJob" : "AIService"),
                  ),
                );
    }
    icons();
  }
  function renderQuotas() {
    const el = document.getElementById("quota-content");
    if (!el) return;
    function quota(q, namespace) {
      const rows = M.quotaRows([{ namespace, quotas: [q] }]);
      const scope = (q.scopes || []).length
        ? `<div class="subtle">适用范围：${E(q.scopes.join(", "))}</div>`
        : "";
      const selector = q.scopeSelector
        ? `<details><summary>范围选择器</summary>${json(q.scopeSelector)}</details>`
        : "";
      return `<div class="monitor-quota"><h3>${E(q.name)}</h3>${scope}${selector}
        ${table(
          ["资源", "期望上限", "已生效上限", "配额已用"],
          rows.map(
            (r) =>
              "<tr>" +
              [
                r.resource,
                r.desiredHard ?? "未提供",
                r.hard ?? "未提供",
                r.used ?? "未提供",
              ]
                .map((v) => cell(E(v)))
                .join("") +
              "</tr>",
          ),
        )}
      </div>`;
    }
    let content = quotaLoading
      ? busy()
      : quotaError
        ? notice(quotaError, "quotas")
        : !quotas.length
          ? empty("暂无命名空间配额数据")
          : quotas
              .map(
                (ns) => `
      <section class="monitor-quota"><h3>${E(ns.namespace)}</h3>
        ${!ns.quotas?.length ? empty("未配置 ResourceQuota") : ns.quotas.map((q) => quota(q, ns.namespace)).join("")}
      </section>`,
              )
              .join("");
    el.innerHTML = `<div class="section-head"><h2>命名空间 ResourceQuota</h2>
      ${iconButton("quotas", "刷新配额", "refresh-cw")}</div>
      <p class="subtle">配额范围：${E(route.ns || "全部命名空间")}（独立于对象名称、租户和状态筛选）</p>${content}`;
    icons();
  }
  async function loadList() {
    const v = version,
      seq = ++listSeq,
      ns = route.ns;
    loading = true;
    listError = "";
    renderCollection();
    try {
      const data = await request(
        `${root}/${route.section === "audits" ? "audits" : "applications"}?` +
          new URLSearchParams({ namespace: ns || "" }),
      );
      if (!active(v) || seq !== listSeq) return;
      items = (data.items || [])
        .slice()
        .sort((a, b) =>
          (b.createdAt || b.time || "").localeCompare(
            a.createdAt || a.time || "",
          ),
        );
      updated = new Date().toLocaleTimeString("zh-CN", { hour12: false });
    } catch (e) {
      if (!active(v) || seq !== listSeq) return;
      if (e.status === 401) {
        expired();
        return;
      }
      items = [];
      listError = errorText(e);
    }
    if (active(v) && seq === listSeq) {
      loading = false;
      renderCollection();
    }
  }
  async function loadQuotas() {
    const v = version,
      seq = ++quotaSeq;
    quotaLoading = true;
    quotaError = "";
    renderQuotas();
    try {
      const data = await request(
        root +
          "/tenant-resources?" +
          new URLSearchParams({ namespace: route.ns || "" }),
      );
      if (!active(v) || seq !== quotaSeq) return;
      quotas = data.items || [];
    } catch (e) {
      if (!active(v) || seq !== quotaSeq) return;
      if (e.status === 401) {
        expired();
        return;
      }
      quotas = [];
      quotaError = errorText(e);
    }
    if (active(v) && seq === quotaSeq) {
      quotaLoading = false;
      renderQuotas();
    }
  }
  function evidenceOverview() {
    const components = detail.components || [],
      events = components.flatMap((c) =>
        (c.pods || []).flatMap((p) =>
          (p.events || []).map((e) => ({
            ...e,
            resource: c.name + "/" + p.name,
          })),
        ),
      );
    return (
      definition([
        ["名称", route.name],
        ["命名空间", route.namespace],
        ["租户", M.tenant(detail) || "未提供"],
        ["阶段", detail.phase || "未提供"],
        ["状态消息", detail.message || "未提供"],
      ]) +
      `<div class="section-head"><h2>诊断证据</h2></div>` +
      (M.diagnostics(detail)
        .map(
          (d) =>
            `<div class="notice monitor-evidence ${d.severity === "error" ? "error" : ""}"><strong>${E(d.reason || "状态消息")}</strong><span>${E(d.message)}</span><small>${E(d.resource)} ${E(d.evidence)}</small></div>`,
        )
        .join("") || empty("暂无异常证据")) +
      '<div class="section-head"><h2>组件与 Pod</h2></div>' +
      table(
        ["组件 / 工作负载", "Pod / 阶段", "容器状态"],
        components.flatMap((c) =>
          (c.pods?.length ? c.pods : [null]).map(
            (p) =>
              "<tr>" +
              cell(
                `${E(c.name)}<span class="subtle">${E(c.workloadKind || c.type)}</span>${c.workload?.completed ? "已完成" : c.workload?.kind === "Deployment" ? `${E(c.workload.readyReplicas || 0)} / ${E(c.workload.desiredReplicas || 0)} 就绪` : E(c.message || "")}`,
              ) +
              cell(
                p
                  ? `${E(p.name)}<span class="subtle">${E(p.phase)} · ${E(p.readyContainers)} / ${E(p.totalContainers)}</span>`
                  : "暂无 Pod",
              ) +
              cell(
                (p?.containers || [])
                  .map(
                    (k) =>
                      `${E(k.name)} · ${E(k.state)} · ${E(k.reason || "")}<span class="subtle">重启 ${E(k.restartCount ?? 0)} · ${E(k.message || "")}</span>`,
                  )
                  .join("") || "未提供",
              ) +
              "</tr>",
          ),
        ),
      ) +
      '<div class="section-head"><h2>Pod 事件</h2></div>' +
      table(
        ["Pod", "类型 / 原因", "消息", "次数 / 时间"],
        events.map(
          (e) =>
            "<tr>" +
            cell(E(e.resource)) +
            cell(`${E(e.type)} / ${E(e.reason)}`) +
            cell(E(e.message)) +
            cell(
              `${E(e.count ?? "未提供")}<span class="subtle">${E(time(e.lastAt))}</span>`,
            ) +
            "</tr>",
        ),
      )
    );
  }
  function renderActionState() {
    const el = document.getElementById("action-status");
    if (!el) return;
    const state = actionStates.get(targetKey(route.namespace, route.name));
    el.innerHTML = state
      ? `<div class="notice monitor-pending" role="status">
      <span>${E(state.message)}</span>
      ${state.checked ? `<span>已重新读取对象状态。提交不代表恢复；再次操作前请核查阶段、诊断和审计。</span>${button("acknowledge", "确认已核查", "check")}` : ""}
    </div>`
      : "";
    page
      .querySelectorAll("[data-lifecycle]")
      .forEach((b) => (b.disabled = !!state || modalBusy));
  }
  function renderDetail() {
    const kind = S.kind(detail),
      job = kind === "AIJob";
    const actions = kind
      ? `<button class="button" data-lifecycle="${job ? "rerun" : "restart"}">${icon("rotate-cw")}${job ? "重新运行" : "重启服务"}</button><button class="icon-button" data-lifecycle="delete" aria-label="删除${job ? "任务" : "服务"}" title="删除${job ? "任务" : "服务"}">${icon("trash-2")}</button>`
      : "";
    const tabs = [
      ["overview", "概览"],
      ["logs", "日志"],
      ...(kind ? [["result", job ? "结果与产物" : "服务访问"]] : []),
      ["config", "配置"],
    ];
    page.innerHTML = `
      <a class="back-link" href="${returnLink()}">${icon("arrow-left")}返回列表</a>
      ${heading(route.name, actions)}
      <div class="actions detail-status">${badge(detail)}${sample(detail)}
        <span class="subtle">${E(route.namespace)}</span>${iconButton("refresh", "刷新详情", "refresh-cw")}
      </div>
      <a class="back-link" href="${E(M.hash({ section: "audits", ns: route.namespace, objectName: route.name }))}">${icon("history")}查看此对象审计</a>
      <div id="page-error"></div><div id="action-status"></div>
      <div class="tabs" role="tablist" aria-label="详情标签">
        ${tabs.map(([t, l]) => `<button type="button" role="tab" aria-selected="${route.tab === t}" data-tab="${t}">${l}</button>`).join("")}
      </div>
      <section class="detail-body" id="detail-content"></section>`;
    const body = document.getElementById("detail-content");
    if (route.tab === "overview") body.innerHTML = evidenceOverview();
    if (route.tab === "config")
      body.innerHTML =
        '<div class="section-head"><h2>已观测配置</h2></div>' +
        json({
          aiMetadata: detail.aiMetadata || {},
          components: detail.components || [],
        });
    if (route.tab === "logs") {
      logPods = (detail.components || []).flatMap((c) =>
        (c.pods || []).map((p) => ({
          name: p.name,
          containers: (p.containers || []).map((k) => k.name),
        })),
      );
      body.innerHTML =
        '<div class="filters monitor-filters monitor-log-filters"><label>Pod<select id="log-pod" aria-label="Pod"></select></label><label>容器<select id="log-container" aria-label="容器"></select></label><label>日志行数<input id="log-tail" type="number" min="1" max="2000" value="200"></label>' +
        iconButton("logs", "刷新日志", "refresh-cw") +
        '</div><div id="aux-error"></div><pre id="log-output" class="log-output" role="status"></pre>';
      const selected = logSelections.get(
        targetKey(route.namespace, route.name),
      ) || { pod: "", container: "", tail: "200" };
      document.getElementById("log-tail").value = selected.tail;
      populateLogs(selected.pod, selected.container);
      loadAux("logs");
    }
    if (route.tab === "result")
      body.innerHTML = kind
        ? `<div class="section-head"><h2>${job ? "任务产物" : "访问检查"}</h2>${button(job ? "result" : "probe", job ? "读取运行结果" : "检查服务访问", job ? "refresh-cw" : "activity")}</div>${job ? "" : `<label class="field">检查路径<input id="probe-path" value="/healthz"></label>`}<div id="aux-error"></div><div id="result-output">${empty(job ? "尚未读取运行结果" : "尚未检查服务访问")}</div>`
        : empty("资源类型未提供");
    renderActionState();
    icons();
  }
  async function loadDetail() {
    const v = version,
      seq = ++listSeq;
    auxSeq++;
    auxBusy = false;
    detail = null;
    page.innerHTML = heading(route.name) + busy();
    try {
      const data = await request(
        pathFor(route.namespace, route.name) + "/status",
      );
      if (!active(v) || seq !== listSeq) return;
      const kind = S.kind(data),
        expected =
          route.section === "jobs"
            ? "AIJob"
            : route.section === "services"
              ? "AIService"
              : "";
      if (expected && kind && kind !== expected)
        throw new Error("资源类型与当前页面不匹配。");
      detail = data;
      const state = actionStates.get(targetKey(route.namespace, route.name));
      if (state && !state.submitting) state.checked = true;
      renderDetail();
    } catch (e) {
      if (!active(v) || seq !== listSeq) return;
      if (e.status === 401) {
        expired();
        return;
      }
      page.innerHTML =
        `<a class="back-link" href="${returnLink()}">返回列表</a>` +
        heading(route.name) +
        notice(errorText(e));
      icons();
    }
  }
  function populateLogs(pod, container) {
    const p = document.getElementById("log-pod"),
      c = document.getElementById("log-container");
    if (!p || !c) return;
    p.innerHTML =
      '<option value="">自动选择</option>' +
      logPods
        .map((x) => `<option value="${E(x.name)}">${E(x.name)}</option>`)
        .join("");
    const chosen = logPods.find((x) => x.name === pod);
    if (pod && !chosen)
      p.insertAdjacentHTML(
        "beforeend",
        `<option value="${E(pod)}">${E(pod)}（已不可用）</option>`,
      );
    p.value = pod;
    c.innerHTML =
      '<option value="">自动选择</option>' +
      (chosen?.containers || [])
        .map((x) => `<option value="${E(x)}">${E(x)}</option>`)
        .join("");
    if (container && !chosen?.containers?.includes(container))
      c.insertAdjacentHTML(
        "beforeend",
        `<option value="${E(container)}">${E(container)}（已不可用）</option>`,
      );
    c.value = container;
  }
  function saveLogSelection() {
    const pod = document.getElementById("log-pod");
    if (!pod) return;
    logSelections.set(targetKey(route.namespace, route.name), {
      pod: pod.value,
      container: document.getElementById("log-container").value,
      tail: document.getElementById("log-tail").value,
    });
  }
  function logSelectionWarning(pod, container) {
    const selected = logPods.find((p) => p.name === pod);
    if (pod && !selected)
      return "所选 Pod 已不可用。请明确选择新的 Pod 后读取日志。";
    if (container && !selected?.containers?.includes(container))
      return "所选容器已不可用。请明确选择新的容器后读取日志。";
    return "";
  }
  async function loadAux(action) {
    if (dead || !detail || (auxBusy && action !== "logs")) return;
    const v = version,
      seq = ++auxSeq,
      ns = route.namespace,
      name = route.name;
    const output = document.getElementById(
        action === "logs" ? "log-output" : "result-output",
      ),
      error = document.getElementById("aux-error");
    if (!output || !error) return;
    let path = pathFor(ns, name),
      init;
    const pod = document.getElementById("log-pod")?.value || "",
      container = document.getElementById("log-container")?.value || "";
    if (action === "logs") {
      saveLogSelection();
      const warning = logSelectionWarning(pod, container);
      if (warning) {
        auxBusy = false;
        output.textContent = "";
        error.innerHTML = notice(warning, "");
        return;
      }
      const tail = document.getElementById("log-tail").value;
      if (!/^\d+$/.test(tail) || Number(tail) < 1 || Number(tail) > 2000) {
        error.innerHTML = notice("日志行数应为 1 到 2000。", "");
        return;
      }
      path +=
        "/logs?" + new URLSearchParams({ pod, container, tailLines: tail });
    } else if (action === "result")
      path = `${root}/deliveries/${encodeURIComponent(ns)}/${encodeURIComponent(name)}/result`;
    else {
      const probePath = document.getElementById("probe-path").value.trim();
      if (!probePath.startsWith("/")) {
        error.innerHTML = notice("检查路径应以 / 开头。", "");
        return;
      }
      path += "/probe";
      init = {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ path: probePath, timeoutSeconds: 10 }),
      };
    }
    auxBusy = true;
    error.innerHTML = "";
    output.textContent = "正在读取…";
    page
      .querySelectorAll('[data-action="result"],[data-action="probe"]')
      .forEach((b) => (b.disabled = true));
    try {
      const data = await request(path, init);
      if (!active(v) || seq !== auxSeq) return;
      if (action === "logs") {
        if (Array.isArray(data.pods)) logPods = data.pods;
        populateLogs(pod || data.pod || "", container || data.container || "");
        saveLogSelection();
        const warning = logSelectionWarning(
          pod || data.pod || "",
          container || data.container || "",
        );
        if (
          warning ||
          (pod && data.pod !== pod) ||
          (container && data.container !== container)
        ) {
          output.textContent = "";
          error.innerHTML = notice(
            warning ||
              "日志响应目标与所选 Pod / 容器不一致，请核查后重新读取。",
            "",
          );
        } else output.textContent = data.logs || "暂无日志。";
      } else if (action === "result")
        output.innerHTML = data.result
          ? json(data.result)
          : empty("暂无运行结果");
      else
        output.innerHTML =
          `<div class="notice ${data.healthy ? "success" : "error"}">${data.healthy ? "访问检查通过" : "访问检查未通过"} · HTTP ${E(data.statusCode || "未返回")}</div>` +
          definition([
            ["目标", data.url || "未提供"],
            ["错误", data.error || "无"],
          ]) +
          `<pre class="log-output">${E(data.body || "无响应内容")}</pre>`;
    } catch (e) {
      if (!active(v) || seq !== auxSeq) return;
      if (e.status === 401) {
        expired();
        return;
      }
      output.textContent = "";
      error.innerHTML = notice(errorText(e), action === "probe" ? "" : action);
    } finally {
      if (active(v) && seq === auxSeq) {
        auxBusy = false;
        page
          .querySelectorAll('[data-action="result"],[data-action="probe"]')
          .forEach((b) => (b.disabled = false));
        icons();
      }
    }
  }
  function confirmAction(action) {
    if (
      !detail ||
      modalBusy ||
      actionStates.has(targetKey(route.namespace, route.name))
    )
      return;
    modalTarget = Object.freeze({
      namespace: route.namespace,
      name: route.name,
      action,
      version,
    });
    modal.innerHTML = `<div class="dialog-head"><h2 id="modal-title">确认${E(actionLabels[action])}</h2></div><div class="dialog-body">${definition(
      [
        ["命名空间", modalTarget.namespace],
        ["名称", modalTarget.name],
      ],
    )}<p>${action === "delete" ? "删除后资源将不可用。" : "操作提交后需等待集群状态更新。"}</p><div id="modal-error"></div></div><div class="dialog-actions">${button("cancel", "取消", "x")}${button("confirm", "确认" + actionLabels[action], "check")}</div>`;
    modal.showModal();
    icons();
    modal.querySelector('[data-action="cancel"]').focus();
  }
  async function submitAction() {
    if (dead || modalBusy || !modalTarget) return;
    const target = modalTarget,
      key = targetKey(target.namespace, target.name);
    if (actionStates.has(key)) return;
    modalBusy = true;
    actionStates.set(key, {
      message: "正在提交操作…",
      submitting: true,
      checked: false,
    });
    renderActionState();
    modal.querySelectorAll("button").forEach((b) => (b.disabled = true));
    try {
      await request(
        pathFor(target.namespace, target.name) +
          (target.action === "delete" ? "" : "/" + target.action),
        { method: target.action === "delete" ? "DELETE" : "POST" },
      );
      if (dead) return;
      actionStates.set(key, {
        message: "已提交，等待状态确认。请刷新详情核查结果。",
        checked: false,
      });
    } catch (e) {
      if (dead) return;
      if (e.status === 401) {
        expired();
        return;
      }
      actionStates.set(key, {
        message:
          "操作结果未确认：" + errorText(e) + " 请刷新详情核查，避免重复提交。",
        checked: false,
      });
    } finally {
      modalBusy = false;
      if (modalTarget === target) {
        modal.close();
        modal.innerHTML = "";
        modalTarget = null;
      }
      if (active(target.version)) renderActionState();
    }
  }
  function closeNavigation() {
    document.body.classList.remove("nav-open");
    document
      .getElementById("nav-toggle")
      .setAttribute("aria-expanded", "false");
  }
  function navigate() {
    if (dead) return;
    version++;
    auxSeq++;
    auxBusy = false;
    route = M.route(location.hash);
    items = [];
    detail = null;
    quotas = [];
    listError = "";
    quotaError = "";
    updated = "";
    logPods = [];
    if (!modalBusy) {
      modal.close();
      modal.innerHTML = "";
      modalTarget = null;
    }
    closeNavigation();
    updateScope();
    document.querySelectorAll("[data-nav]").forEach((a) => {
      const selected = a.dataset.nav === route.section;
      a.classList.toggle("active", selected);
      if (selected) a.setAttribute("aria-current", "page");
      else a.removeAttribute("aria-current");
    });
    if (route.invalid) {
      page.innerHTML =
        heading("无效的资源地址") + empty("请检查命名空间和资源名称");
      icons();
      return;
    }
    if (route.detail) {
      loadDetail();
      return;
    }
    page.innerHTML =
      heading(titles[route.section]) +
      filters() +
      '<div id="page-error"></div><section id="collection"></section>' +
      (route.section === "tenants"
        ? '<section id="quota-content"></section>'
        : "");
    loadList();
    if (route.section === "tenants") loadQuotas();
  }
  function changeFilter(e) {
    const key = e.target.dataset.filter;
    if (!key || dead) return;
    route = { ...route, [key === "name" ? "objectName" : key]: e.target.value };
    history.replaceState(null, "", M.hash(route));
    updateScope();
    if (key === "ns") {
      loadList();
      if (route.section === "tenants") loadQuotas();
    } else renderCollection();
  }
  page.addEventListener("input", (e) => {
    if (e.target.matches("input[data-filter]")) changeFilter(e);
    if (e.target.id === "log-tail") saveLogSelection();
  });
  page.addEventListener("change", (e) => {
    if (e.target.matches("select[data-filter]")) changeFilter(e);
    if (e.target.id === "log-pod") {
      populateLogs(e.target.value, "");
      loadAux("logs");
    }
    if (e.target.id === "log-container") loadAux("logs");
  });
  page.addEventListener("click", (e) => {
    if (dead) return;
    const el = e.target.closest("button");
    if (!el || el.disabled) return;
    if (el.dataset.tab) {
      location.hash = M.hash({ ...route, tab: el.dataset.tab });
      return;
    }
    if (el.dataset.lifecycle) {
      confirmAction(el.dataset.lifecycle);
      return;
    }
    const a = el.dataset.action;
    if (a === "refresh") {
      if (route.detail) loadDetail();
      else {
        loadList();
        if (route.section === "tenants") loadQuotas();
      }
    } else if (a === "quotas") loadQuotas();
    else if (a === "acknowledge") {
      const key = targetKey(route.namespace, route.name),
        state = actionStates.get(key);
      if (state?.checked && !state.submitting) {
        actionStates.delete(key);
        renderActionState();
      }
    } else if (["logs", "result", "probe"].includes(a)) loadAux(a);
  });
  modal.addEventListener("click", (e) => {
    const a = e.target.closest("[data-action]")?.dataset.action;
    if (a === "confirm") submitAction();
    if (a === "cancel" && !modalBusy) {
      modal.close();
      modal.innerHTML = "";
      modalTarget = null;
    }
  });
  modal.addEventListener("cancel", (e) => {
    if (modalBusy) e.preventDefault();
    else {
      modalTarget = null;
      modal.innerHTML = "";
    }
  });
  document.getElementById("nav-toggle").addEventListener("click", () => {
    const open = document.body.classList.toggle("nav-open");
    document
      .getElementById("nav-toggle")
      .setAttribute("aria-expanded", String(open));
  });
  document.querySelector(".main").addEventListener("click", (e) => {
    if (
      document.body.classList.contains("nav-open") &&
      !e.target.closest(".topbar")
    )
      closeNavigation();
  });
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape") closeNavigation();
  });
  document
    .getElementById("logout-form")
    .addEventListener("submit", () => expired());
  document.getElementById("environment").textContent = demo
    ? "本地演示"
    : "集群观测";
  window.addEventListener("hashchange", navigate);
  navigate();
  icons();
})();
