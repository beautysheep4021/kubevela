package northbound

const consoleHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>智算纳管北向验证台</title>
  <style>
    :root {
      --ink: #1d2522;
      --muted: #61706b;
      --paper: #fff8e8;
      --line: rgba(29, 37, 34, .14);
      --field: rgba(255, 255, 255, .72);
      --moss: #375b4c;
      --rust: #b75d35;
      --gold: #d9a73e;
      --blue: #2f5f74;
      --shadow: 0 22px 80px rgba(29, 37, 34, .16);
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      color: var(--ink);
      font-family: "Noto Serif SC", "Songti SC", "STSong", "PingFang SC", serif;
      background:
        radial-gradient(circle at top left, rgba(217, 167, 62, .42), transparent 32rem),
        radial-gradient(circle at 86% 14%, rgba(47, 95, 116, .2), transparent 28rem),
        linear-gradient(135deg, #f6ecd8 0%, #efe1c5 42%, #dbe8dd 100%);
    }
    body:before {
      content: "";
      position: fixed;
      inset: 0;
      pointer-events: none;
      background-image:
        linear-gradient(rgba(24, 33, 31, .04) 1px, transparent 1px),
        linear-gradient(90deg, rgba(24, 33, 31, .04) 1px, transparent 1px);
      background-size: 34px 34px;
      mask-image: linear-gradient(to bottom, rgba(0,0,0,.65), transparent);
    }
    main {
      width: min(1480px, calc(100vw - 32px));
      margin: 0 auto;
      padding: 32px 0 48px;
    }
    .hero {
      display: grid;
      grid-template-columns: 1.1fr .9fr;
      gap: 24px;
      align-items: end;
      margin-bottom: 24px;
    }
    h1 {
      margin: 0;
      font-family: "Noto Serif SC", "Songti SC", Georgia, serif;
      font-size: clamp(42px, 7vw, 92px);
      line-height: .96;
      letter-spacing: -0.08em;
    }
    .lede {
      max-width: 760px;
      margin: 18px 0 0;
      color: var(--muted);
      font-size: 17px;
      line-height: 1.55;
    }
    .status-strip {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 12px;
    }
    .status-card, .panel, .intent-card {
      border: 1px solid var(--line);
      background: rgba(255, 250, 240, .74);
      box-shadow: var(--shadow);
      backdrop-filter: blur(18px);
    }
    .status-card {
      padding: 18px;
      border-radius: 24px;
    }
    .label {
      color: var(--muted);
      font-size: 12px;
      font-weight: 800;
      letter-spacing: .12em;
    }
    .metric {
      margin-top: 8px;
      font-size: 26px;
      font-weight: 900;
    }
    .workspace {
      display: grid;
      grid-template-columns: minmax(420px, .9fr) minmax(480px, 1.1fr);
      gap: 20px;
    }
    .panel {
      min-height: 640px;
      border-radius: 32px;
      overflow: hidden;
    }
    .panel-head {
      display: flex;
      justify-content: space-between;
      gap: 16px;
      align-items: center;
      padding: 20px 22px;
      border-bottom: 1px solid var(--line);
    }
    .panel-title {
      margin: 0;
      font-size: 18px;
      font-weight: 900;
    }
    .actions {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
    }
    .deploy-options {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 0 22px 18px;
      color: var(--muted);
      font-size: 13px;
      font-weight: 900;
    }
    .form-grid {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 14px;
      padding: 20px 22px 8px;
    }
    .field {
      display: grid;
      gap: 7px;
    }
    .field.wide { grid-column: 1 / -1; }
    .field label {
      color: var(--muted);
      font-size: 13px;
      font-weight: 900;
    }
    input, select {
      width: 100%;
      min-height: 42px;
      border: 1px solid var(--line);
      border-radius: 16px;
      padding: 10px 12px;
      color: var(--ink);
      background: rgba(255,255,255,.7);
      outline: none;
      font: inherit;
    }
    input[type="checkbox"] {
      width: auto;
      min-height: auto;
      accent-color: var(--moss);
    }
    input:focus, select:focus {
      border-color: rgba(47, 95, 116, .48);
      box-shadow: 0 0 0 4px rgba(47, 95, 116, .12);
    }
    .subhead {
      grid-column: 1 / -1;
      margin-top: 8px;
      padding-top: 14px;
      border-top: 1px solid var(--line);
      color: var(--ink);
      font-size: 15px;
      font-weight: 900;
    }
    .yaml-preview {
      padding: 0 22px 22px;
    }
    .yaml-preview .label {
      margin: 14px 0 8px;
      display: block;
    }
    button {
      border: 0;
      border-radius: 999px;
      padding: 11px 16px;
      color: #fffaf0;
      background: var(--ink);
      font-weight: 900;
      letter-spacing: .01em;
      cursor: pointer;
      transition: transform .16s ease, box-shadow .16s ease, background .16s ease;
    }
    button:hover {
      transform: translateY(-2px);
      box-shadow: 0 10px 22px rgba(24, 33, 31, .22);
    }
    button.secondary { background: var(--moss); }
    button.ghost {
      color: var(--ink);
      background: rgba(255,255,255,.62);
      border: 1px solid var(--line);
    }
    textarea {
      width: 100%;
      min-height: 268px;
      display: block;
      padding: 22px;
      border: 0;
      resize: vertical;
      outline: none;
      color: #17211e;
      background: var(--field);
      font-family: "SFMono-Regular", "Cascadia Code", "PingFang SC", monospace;
      font-size: 14px;
      line-height: 1.54;
    }
    .results {
      padding: 22px;
    }
    .result-grid {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 14px;
      margin-bottom: 16px;
    }
    .task-list {
      display: grid;
      gap: 10px;
      margin: 18px 0;
    }
    .task-row {
      width: 100%;
      display: grid;
      grid-template-columns: 1.15fr .7fr .7fr .7fr auto;
      gap: 10px;
      align-items: center;
      padding: 13px 14px;
      border: 1px solid var(--line);
      border-radius: 18px;
      color: var(--ink);
      background: rgba(255,255,255,.58);
      text-align: left;
      cursor: pointer;
    }
    .task-row:hover {
      transform: translateY(-1px);
      box-shadow: 0 8px 18px rgba(24, 33, 31, .12);
    }
    .task-row strong, .task-row span {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .pill {
      display: inline-flex;
      justify-content: center;
      border-radius: 999px;
      padding: 5px 9px;
      color: #f7fff7;
      background: var(--moss);
      font-size: 12px;
      font-weight: 900;
    }
    .intent-card {
      padding: 16px;
      border-radius: 22px;
      box-shadow: none;
      animation: lift .32s ease both;
    }
    .intent-card strong {
      display: block;
      margin-top: 5px;
      font-size: 22px;
      word-break: break-word;
    }
    .ok {
      color: #f7fff7;
      background: linear-gradient(135deg, var(--moss), #244437);
    }
    .warn {
      color: #fff8e5;
      background: linear-gradient(135deg, var(--rust), #743823);
    }
    pre {
      min-height: 330px;
      margin: 0;
      padding: 18px;
      overflow: auto;
      border-radius: 24px;
      color: #f8f0da;
      background: #19211f;
      font-family: "SFMono-Regular", "Cascadia Code", "PingFang SC", monospace;
      font-size: 13px;
      line-height: 1.5;
      box-shadow: inset 0 0 0 1px rgba(255,255,255,.08);
    }
    .empty {
      color: var(--muted);
      min-height: 330px;
      display: grid;
      place-items: center;
      padding: 26px;
      text-align: center;
      border: 1px dashed rgba(24,33,31,.22);
      border-radius: 24px;
      background: rgba(255,255,255,.38);
    }
    @keyframes lift {
      from { opacity: 0; transform: translateY(10px); }
      to { opacity: 1; transform: translateY(0); }
    }
    @media (max-width: 980px) {
      main { width: min(100vw - 20px, 760px); padding-top: 18px; }
      .hero, .workspace { grid-template-columns: 1fr; }
      .status-strip { grid-template-columns: 1fr; }
      .form-grid { grid-template-columns: 1fr; }
      .panel { min-height: auto; border-radius: 24px; }
      textarea { min-height: 430px; }
      .result-grid { grid-template-columns: 1fr; }
      .task-row { grid-template-columns: 1fr 1fr; }
    }
  </style>
</head>
<body>
  <main>
    <section class="hero">
      <div>
        <div class="label">智算纳管 · 使用方工作台 PoC</div>
        <h1>使用方工作台</h1>
        <p class="lede">面向业务用户的提交意图入口。用户先通过表单描述模型服务或批任务，页面生成领域 YAML，并调用北向 API 完成校验、语义归一化和提交部署。</p>
      </div>
      <div class="status-strip">
        <div class="status-card">
          <div class="label">接口健康</div>
          <div class="metric" id="health">检查中</div>
        </div>
        <div class="status-card">
          <div class="label">校验入口</div>
          <div class="metric">/api/v1/ai/validate</div>
        </div>
        <div class="status-card">
          <div class="label">部署入口</div>
          <div class="metric">/api/v1/ai/applications</div>
        </div>
      </div>
    </section>

    <section class="workspace">
      <section class="panel">
        <div class="panel-head">
          <h2 class="panel-title">提交意图</h2>
          <div class="actions">
            <button class="ghost" id="load-service">载入 AIService</button>
            <button class="ghost" id="load-job">载入 AIJob</button>
            <button class="ghost" id="generate">生成 YAML</button>
            <button class="secondary" id="validate">校验</button>
            <button id="normalize">归一化</button>
            <button id="deploy">提交部署</button>
          </div>
        </div>
        <div class="form-grid">
          <div class="field">
            <label for="kind">任务类型</label>
            <select id="kind">
              <option value="AIService">AIService 模型服务</option>
              <option value="AIJob">AIJob 批任务</option>
            </select>
          </div>
          <div class="field">
            <label for="name">名称</label>
            <input id="name" value="sentiment-demo">
          </div>
          <div class="field">
            <label for="namespace">命名空间</label>
            <input id="namespace" value="ai-demo">
          </div>
          <div class="field">
            <label for="component">组件名称</label>
            <input id="component" value="sentiment-api">
          </div>
          <div class="field">
            <label for="tenant">租户</label>
            <input id="tenant" value="demo-tenant">
          </div>
          <div class="field">
            <label for="project">项目</label>
            <input id="project" value="sentiment">
          </div>
          <div class="field">
            <label for="environment">环境</label>
            <select id="environment">
              <option value="poc">poc</option>
              <option value="dev">dev</option>
              <option value="test">test</option>
              <option value="prod">prod</option>
            </select>
          </div>
          <div class="field">
            <label for="owner">负责人</label>
            <input id="owner" value="ai-platform">
          </div>
          <div class="subhead">运行配置</div>
          <div class="field">
            <label for="runtime">运行时</label>
            <select id="runtime">
              <option value="http">http</option>
              <option value="batch">batch</option>
              <option value="triton">triton</option>
              <option value="vllm">vllm</option>
              <option value="pytorch">pytorch</option>
              <option value="custom">custom</option>
            </select>
          </div>
          <div class="field">
            <label for="image">镜像</label>
            <input id="image" value="hashicorp/http-echo:0.2.3">
          </div>
          <div class="field service-field">
            <label for="modelName">模型名称</label>
            <input id="modelName" value="sentiment">
          </div>
          <div class="field service-field">
            <label for="modelVersion">模型版本</label>
            <input id="modelVersion" value="v1">
          </div>
          <div class="field wide service-field">
            <label for="modelURI">模型 URI</label>
            <input id="modelURI" value="oss://models/sentiment/v1">
          </div>
          <div class="field service-field">
            <label for="replicas">副本数</label>
            <input id="replicas" value="1">
          </div>
          <div class="field service-field">
            <label for="port">服务端口</label>
            <input id="port" value="5678">
          </div>
          <div class="field job-field">
            <label for="jobKind">Job 类型</label>
            <select id="jobKind">
              <option value="evaluation">evaluation</option>
              <option value="training">training</option>
              <option value="batch">batch</option>
              <option value="offline-inference">offline-inference</option>
            </select>
          </div>
          <div class="field job-field">
            <label for="ttl">TTL 秒数</label>
            <input id="ttl" value="300">
          </div>
          <div class="field wide job-field">
            <label for="datasetURI">数据集 URI</label>
            <input id="datasetURI" value="oss://datasets/eval-set/v1">
          </div>
          <div class="field wide job-field">
            <label for="outputURI">输出 URI</label>
            <input id="outputURI" value="oss://outputs/eval-run/v1">
          </div>
        </div>
        <div class="yaml-preview">
          <span class="label">领域 YAML</span>
          <textarea id="yaml" spellcheck="false"></textarea>
        </div>
        <label class="deploy-options">
          <input id="dry-run" type="checkbox" checked>
          服务端 DryRun：只验证 Kubernetes 接收能力，不实际落集群；取消勾选后才会真实创建或更新 Application。
        </label>
      </section>

      <section class="panel">
        <div class="panel-head">
          <h2 class="panel-title">平台视图预览</h2>
          <div class="actions">
            <button class="ghost" id="refresh-tasks">刷新任务</button>
            <div class="label" id="last-action">待操作</div>
          </div>
        </div>
        <div class="results">
          <div class="result-grid" id="cards"></div>
          <div class="label">任务列表</div>
          <div id="task-list" class="task-list">
            <div class="empty">点击“刷新任务”，查看当前命名空间下的 AIService / AIJob。</div>
          </div>
          <div class="label">任务状态详情</div>
          <div id="output" class="empty">点击“校验”或“归一化”，查看治理意图 governanceIntent 与工作负载意图 workloadIntent 的解析结果。</div>
        </div>
      </section>
    </section>
  </main>

  <script>
    var serviceSample = [
      "apiVersion: ai.oam.dev/v1alpha1",
      "kind: AIService",
      "metadata:",
      "  name: sentiment-demo",
      "  namespace: ai-demo",
      "spec:",
      "  componentName: sentiment-api",
      "  properties:",
      "    image: hashicorp/http-echo:0.2.3",
      "    replicas: 1",
      "    model:",
      "      name: sentiment",
      "      version: v1",
      "      uri: oss://models/sentiment/v1",
      "    endpoint:",
      "      port: 5678",
      "      servicePort: 80",
      "      type: ClusterIP",
      "  runtime:",
      "    runtime: http",
      "    framework: demo",
      "    tenant: demo-tenant",
      "    project: sentiment",
      "    environment: poc",
      "    owner: ai-platform",
      "    modelURI: oss://models/sentiment/v1",
      "  placement:",
      "    namespace: ai-demo",
      "    clusters:",
      "    - local"
    ].join("\n");
    var jobSample = [
      "apiVersion: ai.oam.dev/v1alpha1",
      "kind: AIJob",
      "metadata:",
      "  name: evaluator-demo",
      "  namespace: ai-demo",
      "spec:",
      "  componentName: batch-evaluator",
      "  properties:",
      "    image: busybox:1.36",
      "    jobKind: evaluation",
      "    dataset:",
      "      name: eval-set",
      "      uri: oss://datasets/eval-set/v1",
      "    output:",
      "      uri: oss://outputs/eval-run/v1",
      "    ttlSecondsAfterFinished: 300",
      "  runtime:",
      "    runtime: batch",
      "    framework: shell",
      "    tenant: demo-tenant",
      "    project: evaluation",
      "    environment: poc",
      "    owner: ai-platform",
      "    datasetURI: oss://datasets/eval-set/v1"
    ].join("\n");

    var yaml = document.getElementById("yaml");
    var output = document.getElementById("output");
    var cards = document.getElementById("cards");
    var taskList = document.getElementById("task-list");
    var lastAction = document.getElementById("last-action");
    var fields = {};
    ["kind", "name", "namespace", "component", "tenant", "project", "environment", "owner", "runtime", "image", "modelName", "modelVersion", "modelURI", "replicas", "port", "jobKind", "ttl", "datasetURI", "outputURI"].forEach(function(id) {
      fields[id] = document.getElementById(id);
    });
    yaml.value = serviceSample;

    function line(key, value, indent) {
      return Array((indent || 0) + 1).join(" ") + key + ": " + value;
    }
    function setValue(id, value) {
      fields[id].value = value;
    }
    function syncVisibility() {
      var isService = fields.kind.value === "AIService";
      Array.prototype.forEach.call(document.querySelectorAll(".service-field"), function(node) {
        node.style.display = isService ? "grid" : "none";
      });
      Array.prototype.forEach.call(document.querySelectorAll(".job-field"), function(node) {
        node.style.display = isService ? "none" : "grid";
      });
    }
    function buildYAML() {
      var common = [
        "apiVersion: ai.oam.dev/v1alpha1",
        line("kind", fields.kind.value),
        "metadata:",
        line("name", fields.name.value, 2),
        line("namespace", fields.namespace.value, 2),
        "spec:",
        line("componentName", fields.component.value, 2),
        "  properties:",
        line("image", fields.image.value, 4)
      ];
      if (fields.kind.value === "AIService") {
        common = common.concat([
          line("replicas", fields.replicas.value, 4),
          "    model:",
          line("name", fields.modelName.value, 6),
          line("version", fields.modelVersion.value, 6),
          line("uri", fields.modelURI.value, 6),
          "    endpoint:",
          line("port", fields.port.value, 6),
          "      servicePort: 80",
          "      type: ClusterIP"
        ]);
      } else {
        common = common.concat([
          line("jobKind", fields.jobKind.value, 4),
          "    dataset:",
          "      name: dataset",
          line("uri", fields.datasetURI.value, 6),
          "    output:",
          line("uri", fields.outputURI.value, 6),
          line("ttlSecondsAfterFinished", fields.ttl.value, 4)
        ]);
      }
      common = common.concat([
        "  runtime:",
        line("runtime", fields.runtime.value, 4),
        "    framework: demo",
        line("tenant", fields.tenant.value, 4),
        line("project", fields.project.value, 4),
        line("environment", fields.environment.value, 4),
        line("owner", fields.owner.value, 4)
      ]);
      if (fields.kind.value === "AIService") {
        common.push(line("modelURI", fields.modelURI.value, 4));
      } else {
        common.push(line("datasetURI", fields.datasetURI.value, 4));
      }
      return common.join("\n") + "\n";
    }
    function generateYAML() {
      syncVisibility();
      yaml.value = buildYAML();
    }
    function fillServiceForm() {
      setValue("kind", "AIService");
      setValue("name", "sentiment-demo");
      setValue("namespace", "ai-demo");
      setValue("component", "sentiment-api");
      setValue("tenant", "demo-tenant");
      setValue("project", "sentiment");
      setValue("environment", "poc");
      setValue("owner", "ai-platform");
      setValue("runtime", "http");
      setValue("image", "hashicorp/http-echo:0.2.3");
      setValue("modelName", "sentiment");
      setValue("modelVersion", "v1");
      setValue("modelURI", "oss://models/sentiment/v1");
      setValue("replicas", "1");
      setValue("port", "5678");
      generateYAML();
    }
    function fillJobForm() {
      setValue("kind", "AIJob");
      setValue("name", "evaluator-demo");
      setValue("namespace", "ai-demo");
      setValue("component", "batch-evaluator");
      setValue("tenant", "demo-tenant");
      setValue("project", "evaluation");
      setValue("environment", "poc");
      setValue("owner", "ai-platform");
      setValue("runtime", "batch");
      setValue("image", "busybox:1.36");
      setValue("jobKind", "evaluation");
      setValue("ttl", "300");
      setValue("datasetURI", "oss://datasets/eval-set/v1");
      setValue("outputURI", "oss://outputs/eval-run/v1");
      generateYAML();
    }

    function setJSON(payload, tone) {
      output.className = "";
      output.innerHTML = "";
      var pre = document.createElement("pre");
      pre.textContent = JSON.stringify(payload, null, 2);
      output.appendChild(pre);
      lastAction.textContent = tone;
    }
    function setError(message) {
      cards.innerHTML = "";
      output.className = "empty";
      output.textContent = message;
      lastAction.textContent = "错误";
    }
    function card(label, value, mode) {
      var node = document.createElement("div");
      node.className = "intent-card " + (mode || "");
      node.innerHTML = "<div class=\"label\">" + label + "</div><strong>" + (value || "-") + "</strong>";
      return node;
    }
    function renderCards(data) {
      cards.innerHTML = "";
      cards.appendChild(card("对象类型", data.kind, "ok"));
      cards.appendChild(card("负载类型", data.workloadType, "ok"));
      cards.appendChild(card("租户", data.governanceIntent && data.governanceIntent.tenant, ""));
      cards.appendChild(card("项目", data.governanceIntent && data.governanceIntent.project, ""));
      cards.appendChild(card("运行时", data.runtime, ""));
      cards.appendChild(card("镜像", data.image, ""));
    }
    function fetchJSON(path) {
      return fetch(path).then(function(res) {
        return res.json().then(function(data) {
          if (!res.ok) {
            throw new Error(data.error || ("HTTP " + res.status));
          }
          return data;
        });
      });
    }
    function post(path) {
      lastAction.textContent = "处理中";
      generateYAML();
      return fetch(path, {
        method: "POST",
        headers: {"Content-Type": "application/yaml"},
        body: yaml.value
      }).then(function(res) {
        return res.json().then(function(data) {
          if (!res.ok) {
            throw new Error(data.error || ("HTTP " + res.status));
          }
          return data;
        });
      });
    }
    function deploy() {
      var dryRun = document.getElementById("dry-run").checked;
      return post("/api/v1/ai/applications?dryRun=" + dryRun);
    }
    function renderTasks(items) {
      taskList.innerHTML = "";
      if (!items || !items.length) {
        var empty = document.createElement("div");
        empty.className = "empty";
        empty.textContent = "当前命名空间暂无 AIService / AIJob。";
        taskList.appendChild(empty);
        return;
      }
      items.forEach(function(item) {
        var row = document.createElement("button");
        row.className = "task-row";
        row.innerHTML = [
          "<strong>" + item.namespace + "/" + item.name + "</strong>",
          "<span>" + ((item.workloadTypes || []).join(",") || "-") + "</span>",
          "<span>" + (item.aiMetadata && item.aiMetadata["ai.oam.dev/tenant"] || "-") + "</span>",
          "<span>" + (item.phase || "-") + "</span>",
          "<span class=\"pill\">" + (item.healthy ? "健康" : "未就绪") + "</span>"
        ].join("");
        row.onclick = function() {
          loadTaskDetail(item.namespace, item.name);
        };
        taskList.appendChild(row);
      });
    }
    function refreshTasks() {
      lastAction.textContent = "刷新任务中";
      var ns = encodeURIComponent(fields.namespace.value || "");
      return fetchJSON("/api/v1/ai/applications?namespace=" + ns).then(function(data) {
        renderTasks(data.items || []);
        lastAction.textContent = "任务列表已刷新";
      }).catch(function(err) {
        setError(err.message);
      });
    }
    function loadTaskDetail(namespace, name) {
      lastAction.textContent = "读取任务详情";
      return fetchJSON("/api/v1/ai/applications/" + encodeURIComponent(namespace) + "/" + encodeURIComponent(name) + "/status").then(function(data) {
        cards.innerHTML = "";
        cards.appendChild(card("任务", data.namespace + "/" + data.name, "ok"));
        cards.appendChild(card("状态", data.phase || "-", data.healthy ? "ok" : "warn"));
        cards.appendChild(card("健康", data.healthy ? "健康" : "未就绪", data.healthy ? "ok" : "warn"));
        cards.appendChild(card("组件数", data.components ? data.components.length : 0, ""));
        setJSON(data, "任务详情已更新");
      }).catch(function(err) {
        setError(err.message);
      });
    }
    document.getElementById("load-service").onclick = function() {
      fillServiceForm();
      lastAction.textContent = "已载入服务样例";
    };
    document.getElementById("load-job").onclick = function() {
      fillJobForm();
      lastAction.textContent = "已载入任务样例";
    };
    document.getElementById("generate").onclick = function() {
      generateYAML();
      lastAction.textContent = "已生成 YAML";
    };
    Object.keys(fields).forEach(function(id) {
      fields[id].addEventListener("input", generateYAML);
      fields[id].addEventListener("change", generateYAML);
    });
    document.getElementById("validate").onclick = function() {
      post("/api/v1/ai/validate").then(function(data) {
        cards.innerHTML = "";
        cards.appendChild(card("校验结果", "通过", "ok"));
        setJSON(data, "已校验");
      }).catch(function(err) {
        cards.innerHTML = "";
        cards.appendChild(card("校验结果", "未通过", "warn"));
        setError(err.message);
      });
    };
    document.getElementById("normalize").onclick = function() {
      post("/api/v1/ai/normalize").then(function(data) {
        renderCards(data);
        setJSON(data, "已归一化");
      }).catch(function(err) {
        setError(err.message);
      });
    };
    document.getElementById("deploy").onclick = function() {
      deploy().then(function(data) {
        renderCards(data.normalized || {});
        cards.appendChild(card("部署结果", data.application && data.application.dryRun ? "DryRun 通过" : "已提交", "ok"));
        cards.appendChild(card("Application", data.application && ((data.application.namespace || "default") + "/" + data.application.name), "ok"));
        setJSON(data, data.application && data.application.dryRun ? "已 DryRun" : "已提交部署");
        if (data.application && !data.application.dryRun) {
          refreshTasks();
        }
      }).catch(function(err) {
        setError(err.message);
      });
    };
    document.getElementById("refresh-tasks").onclick = refreshTasks;
    fetch("/healthz").then(function(res) {
      return res.text();
    }).then(function(text) {
      document.getElementById("health").textContent = text.trim() ? "正常" : "正常";
    }).catch(function() {
      document.getElementById("health").textContent = "离线";
    });
    fillServiceForm();
  </script>
</body>
</html>`
