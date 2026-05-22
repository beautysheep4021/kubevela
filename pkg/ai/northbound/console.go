package northbound

const consoleHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>智算纳管北向验证台</title>
  <style>
    :root {
      --ink: #17202a;
      --muted: #627084;
      --paper: #f7f9fc;
      --panel: rgba(255, 255, 255, .86);
      --line: rgba(23, 32, 42, .12);
      --field: rgba(248, 250, 252, .92);
      --moss: #256b5f;
      --rust: #b4493f;
      --blue: #245b91;
      --steel: #31465f;
      --shadow: 0 18px 54px rgba(28, 39, 52, .12);
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      color: var(--ink);
      font-family: "Avenir Next", "Noto Sans SC", "PingFang SC", "Microsoft YaHei", sans-serif;
      background:
        radial-gradient(circle at 4% 0%, rgba(36, 91, 145, .16), transparent 30rem),
        radial-gradient(circle at 86% 10%, rgba(37, 107, 95, .14), transparent 28rem),
        linear-gradient(135deg, #f8fafc 0%, #edf3f8 48%, #e7edf3 100%);
    }
    body:before {
      content: "";
      position: fixed;
      inset: 0;
      pointer-events: none;
      background-image:
        linear-gradient(rgba(23, 32, 42, .035) 1px, transparent 1px),
        linear-gradient(90deg, rgba(23, 32, 42, .035) 1px, transparent 1px);
      background-size: 28px 28px;
      mask-image: linear-gradient(to bottom, rgba(0,0,0,.5), transparent);
    }
    main {
      width: min(1480px, calc(100vw - 32px));
      margin: 0 auto;
      padding: 24px 0 42px;
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
      font-size: clamp(30px, 5vw, 64px);
      line-height: 1.02;
      letter-spacing: -.055em;
    }
    .lede {
      max-width: 760px;
      margin: 18px 0 0;
      color: var(--muted);
      font-size: 14px;
      line-height: 1.55;
    }
    .status-strip {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 12px;
    }
    .status-card, .panel, .intent-card {
      border: 1px solid var(--line);
      background: var(--panel);
      box-shadow: var(--shadow);
      backdrop-filter: blur(18px);
    }
    .status-card {
      padding: 14px;
      border-radius: 18px;
    }
    .tabs {
      display: inline-flex;
      gap: 8px;
      padding: 5px;
      margin: 18px 0 0;
      border: 1px solid var(--line);
      border-radius: 999px;
      background: rgba(255,255,255,.66);
      box-shadow: 0 10px 26px rgba(28,39,52,.08);
    }
    .tab {
      color: var(--ink);
      background: transparent;
      border: 1px solid transparent;
    }
    .tab.active {
      color: #fffaf0;
      background: var(--steel);
    }
    .view { display: none; }
    .view.active { display: block; }
    .label {
      color: var(--muted);
      font-size: 11px;
      font-weight: 800;
      letter-spacing: .12em;
    }
    .metric {
      margin-top: 8px;
      font-size: 18px;
      font-weight: 900;
    }
    .workspace {
      display: grid;
      grid-template-columns: minmax(420px, .9fr) minmax(480px, 1.1fr);
      gap: 20px;
    }
    .monitor-grid {
      display: grid;
      grid-template-columns: minmax(340px, .58fr) minmax(520px, 1fr);
      gap: 20px;
    }
    .panel {
      min-height: 640px;
      border-radius: 24px;
      overflow: hidden;
    }
    .panel-head {
      display: flex;
      justify-content: space-between;
      gap: 16px;
      align-items: center;
      padding: 16px 18px;
      border-bottom: 1px solid var(--line);
    }
    .panel-title {
      margin: 0;
      font-size: 16px;
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
      padding: 0 18px 14px;
      color: var(--muted);
      font-size: 12px;
      font-weight: 900;
    }
    .form-grid {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 14px;
      padding: 16px 18px 6px;
    }
    .filter-grid {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 14px;
      padding: 16px 18px;
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
      min-height: 36px;
      border: 1px solid var(--line);
      border-radius: 12px;
      padding: 8px 10px;
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
      padding: 0 18px 18px;
    }
    .yaml-preview .label {
      margin: 14px 0 8px;
      display: block;
    }
    button {
      border: 0;
      border-radius: 999px;
      padding: 9px 13px;
      color: #fffaf0;
      background: var(--steel);
      font-weight: 900;
      letter-spacing: .01em;
      cursor: pointer;
      transition: transform .16s ease, box-shadow .16s ease, background .16s ease;
    }
    button:hover {
      transform: translateY(-2px);
      box-shadow: 0 8px 18px rgba(28, 39, 52, .16);
    }
    button.secondary { background: var(--moss); }
    button.ghost {
      color: var(--ink);
      background: rgba(255,255,255,.74);
      border: 1px solid var(--line);
    }
    textarea {
      width: 100%;
      min-height: 248px;
      display: block;
      padding: 16px;
      border: 0;
      resize: vertical;
      outline: none;
      color: #17211e;
      background: var(--field);
      font-family: "SFMono-Regular", "Cascadia Code", "PingFang SC", monospace;
      font-size: 12px;
      line-height: 1.5;
    }
    .results {
      padding: 18px;
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
    .task-list.flush { margin: 0; }
    .task-row {
      width: 100%;
      display: grid;
      grid-template-columns: 1.15fr .7fr .7fr .7fr auto;
      gap: 10px;
      align-items: center;
      padding: 11px 12px;
      border: 1px solid var(--line);
      border-radius: 14px;
      color: var(--ink);
      background: rgba(255,255,255,.72);
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
      padding: 13px;
      border-radius: 16px;
      box-shadow: none;
      animation: lift .32s ease both;
    }
    .intent-card strong {
      display: block;
      margin-top: 5px;
      font-size: 18px;
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
      min-height: 300px;
      margin: 0;
      padding: 14px;
      overflow: auto;
      border-radius: 18px;
      color: #dce7f3;
      background: #111821;
      font-family: "SFMono-Regular", "Cascadia Code", "PingFang SC", monospace;
      font-size: 12px;
      line-height: 1.5;
      box-shadow: inset 0 0 0 1px rgba(255,255,255,.08);
    }
    .empty {
      color: var(--muted);
      min-height: 300px;
      display: grid;
      place-items: center;
      padding: 26px;
      text-align: center;
      border: 1px dashed rgba(24,33,31,.22);
      border-radius: 18px;
      background: rgba(255,255,255,.54);
    }
    .log-toolbar {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
      align-items: center;
      margin: 18px 0 10px;
    }
    .log-toolbar select, .log-toolbar input {
      width: auto;
      min-width: 160px;
    }
    .log-box {
      min-height: 260px;
      max-height: 460px;
      white-space: pre-wrap;
      color: #dce7f3;
      background: #0f1720;
    }
    .hint {
      color: var(--muted);
      font-size: 12px;
      line-height: 1.5;
    }
    .danger {
      color: #fff;
      background: var(--rust);
    }
    .audit-list {
      display: grid;
      gap: 10px;
      margin: 12px 0 0;
    }
    .audit-row {
      padding: 11px 12px;
      border: 1px solid var(--line);
      border-radius: 14px;
      background: rgba(255,255,255,.7);
      font-size: 12px;
      line-height: 1.5;
    }
    .audit-row strong {
      display: block;
      color: var(--ink);
      font-size: 13px;
    }
    @keyframes lift {
      from { opacity: 0; transform: translateY(10px); }
      to { opacity: 1; transform: translateY(0); }
    }
    @media (max-width: 980px) {
      main { width: min(100vw - 20px, 760px); padding-top: 18px; }
      .hero, .workspace { grid-template-columns: 1fr; }
      .monitor-grid { grid-template-columns: 1fr; }
      .status-strip { grid-template-columns: 1fr; }
      .form-grid, .filter-grid { grid-template-columns: 1fr; }
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
        <p class="lede">面向业务用户和平台监测方的同页 PoC。使用方提交模型服务或批任务，监测方查看全局任务概览、筛选条件和异常任务。</p>
        <div class="tabs">
          <button class="tab active" id="user-tab">使用方工作台</button>
          <button class="tab" id="monitor-tab">监测方工作台</button>
        </div>
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

    <section class="view active" id="user-view">
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
          <div class="log-toolbar">
            <div class="label">应用交付</div>
            <input id="delivery-service-name" placeholder="发布服务名，默认 job-service">
            <input id="delivery-service-image" value="python:3.11-slim" title="服务镜像">
            <button class="ghost" id="parse-delivery-result">解析训练结果</button>
            <button class="secondary" id="publish-delivery-service">发布为服务</button>
          </div>
          <pre id="delivery-output" class="log-box">选择 AIJob 后，可解析 AI_RESULT_JSON 并发布为 AIService。接口：/api/v1/ai/deliveries</pre>
          <div class="log-toolbar">
            <div class="label">生命周期操作</div>
            <button class="ghost" id="rerun-task">重新运行 Job</button>
            <button class="ghost" id="restart-task">重启服务</button>
            <button class="danger" id="delete-task">删除任务</button>
          </div>
          <div class="log-toolbar">
            <div class="label">运行日志</div>
            <select id="log-pod"><option value="">自动选择 Pod</option></select>
            <select id="log-container"><option value="">自动选择容器</option></select>
            <input id="log-tail" value="200" title="日志行数">
            <button class="ghost" id="refresh-logs">刷新日志</button>
          </div>
          <pre id="logs-output" class="log-box">选择任务后，可查看最近日志。AIJob 用于查看训练输出，AIService 用于查看启动和请求日志。</pre>
        </div>
      </section>
    </section>
    </section>

    <section class="view" id="monitor-view">
      <section class="monitor-grid">
        <section class="panel">
          <div class="panel-head">
            <h2 class="panel-title">全局任务概览</h2>
            <button id="monitor-refresh">刷新监测</button>
          </div>
          <div class="results">
            <div class="result-grid" id="monitor-cards">
              <div class="intent-card ok"><div class="label">AIService 数量</div><strong>0</strong></div>
              <div class="intent-card ok"><div class="label">AIJob 数量</div><strong>0</strong></div>
              <div class="intent-card"><div class="label">Running</div><strong>0</strong></div>
              <div class="intent-card warn"><div class="label">异常任务</div><strong>0</strong></div>
            </div>
          </div>
          <div class="subhead" style="margin:0 22px;">筛选条件</div>
          <div class="filter-grid">
            <div class="field">
              <label for="monitorNamespace">命名空间</label>
              <input id="monitorNamespace" value="sock-shop" placeholder="留空表示全部">
            </div>
            <div class="field">
              <label for="monitorType">任务类型</label>
              <select id="monitorType">
                <option value="">全部</option>
                <option value="service">AIService</option>
                <option value="job">AIJob</option>
              </select>
            </div>
            <div class="field">
              <label for="monitorTenant">租户</label>
              <input id="monitorTenant" placeholder="tenant">
            </div>
            <div class="field">
              <label for="monitorProject">项目</label>
              <input id="monitorProject" placeholder="project">
            </div>
            <div class="field">
              <label for="monitorEnvironment">环境</label>
              <input id="monitorEnvironment" placeholder="environment">
            </div>
            <div class="field">
              <label for="monitorHealth">健康状态</label>
              <select id="monitorHealth">
                <option value="">全部</option>
                <option value="healthy">健康</option>
                <option value="unhealthy">未就绪</option>
              </select>
            </div>
          </div>
        </section>

        <section class="panel">
          <div class="panel-head">
            <h2 class="panel-title">监测任务列表</h2>
            <div class="label" id="monitor-action">待刷新</div>
          </div>
          <div class="results">
            <div class="label">任务列表</div>
            <div id="monitor-task-list" class="task-list flush">
              <div class="empty">点击“刷新监测”，查看平台任务。</div>
            </div>
            <div class="label">异常任务</div>
            <div id="monitor-alert-list" class="task-list">
              <div class="empty">暂无异常任务。</div>
            </div>
            <div class="label">任务状态详情</div>
            <div id="monitor-output" class="empty">点击监测任务，查看单任务状态详情。</div>
            <div class="log-toolbar">
              <div class="label">生命周期操作</div>
              <button class="ghost" id="monitor-rerun-task">重新运行 Job</button>
              <button class="ghost" id="monitor-restart-task">重启服务</button>
              <button class="danger" id="monitor-delete-task">删除任务</button>
            </div>
            <div class="log-toolbar">
              <div class="label">运行日志</div>
              <select id="monitor-log-pod"><option value="">自动选择 Pod</option></select>
              <select id="monitor-log-container"><option value="">自动选择容器</option></select>
              <input id="monitor-log-tail" value="200" title="日志行数">
              <button class="ghost" id="monitor-refresh-logs">刷新日志</button>
            </div>
            <pre id="monitor-logs-output" class="log-box">选择监测任务后，可查看最近日志。</pre>
            <div class="log-toolbar">
              <div class="label">操作审计</div>
              <button class="ghost" id="refresh-audits">刷新审计</button>
            </div>
            <div id="audit-list" class="audit-list">
              <div class="empty">点击“刷新审计”，查看最近生命周期操作。</div>
            </div>
          </div>
        </section>
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
      "  name: delivery-train-demo",
      "  namespace: sock-shop",
      "spec:",
      "  componentName: delivery-trainer",
      "  properties:",
      "    image: busybox:1.36",
      "    imagePullPolicy: IfNotPresent",
      "    jobKind: training",
      "    cmd:",
      "      - sh",
      "      - -c",
      "    args:",
      "      - |",
      "        echo train-start",
      "        echo epoch=1 loss=0.30",
      "        echo epoch=2 loss=0.12",
      "        echo 'AI_RESULT_JSON={\"modelURI\":\"inline://models/delivery-demo/v1\",\"metrics\":{\"loss\":0.12,\"accuracy\":0.98},\"summary\":\"trained-for-delivery\"}'",
      "        echo train-complete",
      "    dataset:",
      "      name: tiny-delivery",
      "      uri: inline://datasets/tiny-delivery",
      "    output:",
      "      uri: inline://outputs/delivery-demo",
      "    backoffLimit: 0",
      "    ttlSecondsAfterFinished: 3600",
      "  runtime:",
      "    runtime: batch",
      "    framework: shell",
      "    tenant: demo-tenant",
      "    project: delivery-demo",
      "    environment: poc",
      "    owner: ai-platform",
      "    datasetURI: inline://datasets/tiny-delivery",
      "  placement:",
      "    namespace: sock-shop",
      "    clusters:",
      "    - local"
    ].join("\n");

    var yaml = document.getElementById("yaml");
    var output = document.getElementById("output");
    var cards = document.getElementById("cards");
    var taskList = document.getElementById("task-list");
    var lastAction = document.getElementById("last-action");
    var selectedTask = null;
    var selectedMonitorTask = null;
    var logPod = document.getElementById("log-pod");
    var logContainer = document.getElementById("log-container");
    var logTail = document.getElementById("log-tail");
    var logsOutput = document.getElementById("logs-output");
    var deliveryOutput = document.getElementById("delivery-output");
    var deliveryServiceName = document.getElementById("delivery-service-name");
    var deliveryServiceImage = document.getElementById("delivery-service-image");
    var monitorLogPod = document.getElementById("monitor-log-pod");
    var monitorLogContainer = document.getElementById("monitor-log-container");
    var monitorLogTail = document.getElementById("monitor-log-tail");
    var monitorLogsOutput = document.getElementById("monitor-logs-output");
    var monitorCards = document.getElementById("monitor-cards");
    var monitorTaskList = document.getElementById("monitor-task-list");
    var monitorAlertList = document.getElementById("monitor-alert-list");
    var monitorOutput = document.getElementById("monitor-output");
    var monitorAction = document.getElementById("monitor-action");
    var auditList = document.getElementById("audit-list");
    var monitorFields = {};
    ["monitorNamespace", "monitorType", "monitorTenant", "monitorProject", "monitorEnvironment", "monitorHealth"].forEach(function(id) {
      monitorFields[id] = document.getElementById(id);
    });
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
      setValue("name", "delivery-train-demo");
      setValue("namespace", "sock-shop");
      setValue("component", "delivery-trainer");
      setValue("tenant", "demo-tenant");
      setValue("project", "delivery-demo");
      setValue("environment", "poc");
      setValue("owner", "ai-platform");
      setValue("runtime", "batch");
      setValue("image", "busybox:1.36");
      setValue("jobKind", "training");
      setValue("ttl", "3600");
      setValue("datasetURI", "inline://datasets/tiny-delivery");
      setValue("outputURI", "inline://outputs/delivery-demo");
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
    function setMonitorError(message) {
      monitorOutput.className = "empty";
      monitorOutput.textContent = message;
      monitorAction.textContent = "错误";
    }
    function resetLogSelectors(podSelect, containerSelect) {
      podSelect.innerHTML = "<option value=\"\">自动选择 Pod</option>";
      containerSelect.innerHTML = "<option value=\"\">自动选择容器</option>";
    }
    function fillLogSelectors(payload, podSelect, containerSelect) {
      var currentPod = podSelect.value;
      var currentContainer = containerSelect.value;
      resetLogSelectors(podSelect, containerSelect);
      (payload.pods || []).forEach(function(pod) {
        var option = document.createElement("option");
        option.value = pod.name;
        option.textContent = pod.name + (pod.phase ? " · " + pod.phase : "");
        podSelect.appendChild(option);
      });
      if (payload.pod) {
        podSelect.value = payload.pod;
      } else if (currentPod) {
        podSelect.value = currentPod;
      }
      var selectedPod = (payload.pods || []).filter(function(pod) { return pod.name === podSelect.value; })[0] || (payload.pods || [])[0];
      ((selectedPod && selectedPod.containers) || []).forEach(function(name) {
        var option = document.createElement("option");
        option.value = name;
        option.textContent = name;
        containerSelect.appendChild(option);
      });
      if (payload.container) {
        containerSelect.value = payload.container;
      } else if (currentContainer) {
        containerSelect.value = currentContainer;
      }
    }
    function loadLogs(target, controls) {
      if (!target) {
        controls.output.textContent = "请先选择一个任务。";
        return Promise.resolve();
      }
      controls.output.textContent = "正在读取日志...";
      var query = [
        "tailLines=" + encodeURIComponent(controls.tail.value || "200")
      ];
      if (controls.pod.value) query.push("pod=" + encodeURIComponent(controls.pod.value));
      if (controls.container.value) query.push("container=" + encodeURIComponent(controls.container.value));
      var path = "/api/v1/ai/applications/" + encodeURIComponent(target.namespace) + "/" + encodeURIComponent(target.name) + "/logs?" + query.join("&");
      return fetchJSON(path).then(function(data) {
        fillLogSelectors(data, controls.pod, controls.container);
        controls.output.textContent = data.logs || "该容器暂无日志输出。";
      }).catch(function(err) {
        controls.output.textContent = "日志读取失败：" + err.message;
      });
    }
    function deliveryBase(target) {
      return "/api/v1/ai/deliveries/" + encodeURIComponent(target.namespace) + "/" + encodeURIComponent(target.name);
    }
    function deliveryTargetFromSelectionOrForm() {
      if (selectedTask) {
        return selectedTask;
      }
      if (fields.kind.value === "AIJob" && fields.name.value && fields.namespace.value) {
        return {namespace: fields.namespace.value, name: fields.name.value};
      }
      return null;
    }
    function parseDeliveryResult() {
      var target = deliveryTargetFromSelectionOrForm();
      if (!target) {
        deliveryOutput.textContent = "请先选择一个已提交并完成的 AIJob，或在左侧表单载入 AIJob 后先提交部署。";
        return Promise.resolve();
      }
      deliveryOutput.textContent = "正在解析训练结果...";
      return fetchJSON(deliveryBase(target) + "/result").then(function(data) {
        deliveryOutput.textContent = JSON.stringify(data, null, 2);
        if (!deliveryServiceName.value) {
          deliveryServiceName.value = target.name + "-service";
        }
      }).catch(function(err) {
        deliveryOutput.textContent = "训练结果解析失败：" + err.message + "\\n\\n请确认：1. 这个 AIJob 已经提交部署；2. Job 已完成；3. 日志里包含 AI_RESULT_JSON=...。";
      });
    }
    function publishDeliveryService() {
      var target = deliveryTargetFromSelectionOrForm();
      if (!target) {
        deliveryOutput.textContent = "请先选择一个已提交并完成的 AIJob，或在左侧表单载入 AIJob 后先提交部署。";
        return Promise.resolve();
      }
      var payload = {
        serviceName: deliveryServiceName.value || (target.name + "-service"),
        image: deliveryServiceImage.value || "python:3.11-slim",
        port: 8080,
        servicePort: 80
      };
      deliveryOutput.textContent = "正在发布 AIService...";
      return fetch(deliveryBase(target) + "/publish-service", {
        method: "POST",
        headers: {"Content-Type": "application/json", "X-AI-User": "console"},
        body: JSON.stringify(payload)
      }).then(function(res) {
        return res.json().then(function(data) {
          if (!res.ok) {
            throw new Error(data.error || ("HTTP " + res.status));
          }
          return data;
        });
      }).then(function(data) {
        deliveryOutput.textContent = JSON.stringify(data, null, 2);
        refreshTasks();
        refreshAudits();
      }).catch(function(err) {
        deliveryOutput.textContent = "发布服务失败：" + err.message + "\\n\\n请先解析训练结果，确认该 AIJob 日志中存在 modelURI。";
      });
    }
    function lifecyclePath(target, action) {
      var base = "/api/v1/ai/applications/" + encodeURIComponent(target.namespace) + "/" + encodeURIComponent(target.name);
      if (action === "delete") return base;
      return base + "/" + action;
    }
    function runLifecycle(target, action, statusNode, after) {
      if (!target) {
        statusNode.textContent = "请先选择一个任务。";
        return Promise.resolve();
      }
      if (action === "delete" && !window.confirm("确认删除任务 " + target.namespace + "/" + target.name + "？")) {
        return Promise.resolve();
      }
      statusNode.textContent = "正在执行生命周期操作";
      return fetch(lifecyclePath(target, action), {
        method: action === "delete" ? "DELETE" : "POST",
        headers: {"X-AI-User": "console"}
      }).then(function(res) {
        return res.json().then(function(data) {
          if (!res.ok) {
            throw new Error(data.error || ("HTTP " + res.status));
          }
          return data;
        });
      }).then(function(data) {
        statusNode.textContent = data.message || "操作已提交";
        if (after) after(data);
        return refreshAudits();
      }).catch(function(err) {
        statusNode.textContent = "操作失败：" + err.message;
      });
    }
    function refreshAudits() {
      var ns = encodeURIComponent(monitorFields.monitorNamespace.value || fields.namespace.value || "");
      return fetchJSON("/api/v1/ai/audits?namespace=" + ns).then(function(data) {
        renderAudits(data.items || []);
      }).catch(function(err) {
        auditList.innerHTML = "";
        var empty = document.createElement("div");
        empty.className = "empty";
        empty.textContent = "审计读取失败：" + err.message;
        auditList.appendChild(empty);
      });
    }
    function renderAudits(items) {
      auditList.innerHTML = "";
      if (!items.length) {
        var empty = document.createElement("div");
        empty.className = "empty";
        empty.textContent = "暂无审计记录。";
        auditList.appendChild(empty);
        return;
      }
      items.slice(0, 20).forEach(function(item) {
        var row = document.createElement("div");
        row.className = "audit-row";
        row.innerHTML = [
          "<strong>" + (item.action || "-") + " · " + (item.namespace || "-") + "/" + (item.name || "-") + "</strong>",
          "<span>" + (item.time || "-") + " · " + (item.actor || "anonymous") + " · " + (item.success ? "成功" : "失败") + "</span>",
          "<div class=\"hint\">" + (item.error || item.message || "-") + "</div>"
        ].join("");
        auditList.appendChild(row);
      });
    }
    function card(label, value, mode) {
      var node = document.createElement("div");
      node.className = "intent-card " + (mode || "");
      node.innerHTML = "<div class=\"label\">" + label + "</div><strong>" + (value || "-") + "</strong>";
      return node;
    }
    function setActiveView(name) {
      var userActive = name === "user";
      document.getElementById("user-view").className = userActive ? "view active" : "view";
      document.getElementById("monitor-view").className = userActive ? "view" : "view active";
      document.getElementById("user-tab").className = userActive ? "tab active" : "tab";
      document.getElementById("monitor-tab").className = userActive ? "tab" : "tab active";
      if (!userActive) {
        refreshMonitor();
      }
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
    function metadata(item, key) {
      return item.aiMetadata && (item.aiMetadata["ai.oam.dev/" + key] || item.aiMetadata[key]) || "";
    }
    function matchesMonitorFilters(item) {
      var workloadTypes = item.workloadTypes || [];
      var type = monitorFields.monitorType.value;
      if (type && workloadTypes.indexOf(type) === -1) {
        return false;
      }
      if (monitorFields.monitorTenant.value && metadata(item, "tenant") !== monitorFields.monitorTenant.value) {
        return false;
      }
      if (monitorFields.monitorProject.value && metadata(item, "project") !== monitorFields.monitorProject.value) {
        return false;
      }
      if (monitorFields.monitorEnvironment.value && metadata(item, "environment") !== monitorFields.monitorEnvironment.value) {
        return false;
      }
      if (monitorFields.monitorHealth.value === "healthy" && !item.healthy) {
        return false;
      }
      if (monitorFields.monitorHealth.value === "unhealthy" && item.healthy) {
        return false;
      }
      return true;
    }
    function renderMonitorCards(items) {
      var serviceCount = 0;
      var jobCount = 0;
      var runningCount = 0;
      var unhealthyCount = 0;
      items.forEach(function(item) {
        var types = item.workloadTypes || [];
        if (types.indexOf("service") !== -1) serviceCount++;
        if (types.indexOf("job") !== -1) jobCount++;
        if (item.phase === "running") runningCount++;
        if (!item.healthy) unhealthyCount++;
      });
      monitorCards.innerHTML = "";
      monitorCards.appendChild(card("AIService 数量", serviceCount, "ok"));
      monitorCards.appendChild(card("AIJob 数量", jobCount, "ok"));
      monitorCards.appendChild(card("Running", runningCount, ""));
      monitorCards.appendChild(card("异常任务", unhealthyCount, unhealthyCount ? "warn" : "ok"));
    }
    function renderTaskRows(container, items, emptyText) {
      container.innerHTML = "";
      if (!items.length) {
        var empty = document.createElement("div");
        empty.className = "empty";
        empty.textContent = emptyText;
        container.appendChild(empty);
        return;
      }
      items.forEach(function(item) {
        var row = document.createElement("button");
        row.className = "task-row";
        row.innerHTML = [
          "<strong>" + item.namespace + "/" + item.name + "</strong>",
          "<span>" + ((item.workloadTypes || []).join(",") || "-") + "</span>",
          "<span>" + (metadata(item, "tenant") || "-") + "</span>",
          "<span>" + (metadata(item, "environment") || item.phase || "-") + "</span>",
          "<span class=\"pill\">" + (item.healthy ? "健康" : "未就绪") + "</span>"
        ].join("");
        row.onclick = function() {
          loadMonitorDetail(item.namespace, item.name);
        };
        container.appendChild(row);
      });
    }
    function refreshMonitor() {
      monitorAction.textContent = "刷新监测中";
      var ns = encodeURIComponent(monitorFields.monitorNamespace.value || "");
      return fetchJSON("/api/v1/ai/applications?namespace=" + ns).then(function(data) {
        var items = (data.items || []).filter(matchesMonitorFilters);
        renderMonitorCards(items);
        renderTaskRows(monitorTaskList, items, "当前筛选条件下暂无任务。");
        renderTaskRows(monitorAlertList, items.filter(function(item) { return !item.healthy; }), "暂无异常任务。");
        monitorAction.textContent = "监测视图已刷新";
        refreshAudits();
      }).catch(function(err) {
        setMonitorError(err.message);
      });
    }
    function loadMonitorDetail(namespace, name) {
      monitorAction.textContent = "读取任务详情";
      selectedMonitorTask = {namespace: namespace, name: name};
      return fetchJSON("/api/v1/ai/applications/" + encodeURIComponent(namespace) + "/" + encodeURIComponent(name) + "/status").then(function(data) {
        monitorOutput.className = "";
        monitorOutput.innerHTML = "";
        var pre = document.createElement("pre");
        pre.textContent = JSON.stringify(data, null, 2);
        monitorOutput.appendChild(pre);
        monitorAction.textContent = "任务详情已更新";
        loadLogs(selectedMonitorTask, {
          pod: monitorLogPod,
          container: monitorLogContainer,
          tail: monitorLogTail,
          output: monitorLogsOutput
        });
      }).catch(function(err) {
        setMonitorError(err.message);
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
      selectedTask = {namespace: namespace, name: name};
      return fetchJSON("/api/v1/ai/applications/" + encodeURIComponent(namespace) + "/" + encodeURIComponent(name) + "/status").then(function(data) {
        cards.innerHTML = "";
        cards.appendChild(card("任务", data.namespace + "/" + data.name, "ok"));
        cards.appendChild(card("状态", data.phase || "-", data.healthy ? "ok" : "warn"));
        cards.appendChild(card("健康", data.healthy ? "健康" : "未就绪", data.healthy ? "ok" : "warn"));
        cards.appendChild(card("组件数", data.components ? data.components.length : 0, ""));
        setJSON(data, "任务详情已更新");
        loadLogs(selectedTask, {
          pod: logPod,
          container: logContainer,
          tail: logTail,
          output: logsOutput
        });
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
      selectedTask = null;
      deliveryServiceName.value = "";
      deliveryOutput.textContent = "已载入可交付 AIJob 样例。下一步：取消 DryRun 后点击“提交部署”，等待 Job 完成，再点击“解析训练结果”。";
      lastAction.textContent = "已载入可交付任务样例";
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
    document.getElementById("parse-delivery-result").onclick = parseDeliveryResult;
    document.getElementById("publish-delivery-service").onclick = publishDeliveryService;
    document.getElementById("delete-task").onclick = function() {
      runLifecycle(selectedTask, "delete", lastAction, function() {
        selectedTask = null;
        refreshTasks();
      });
    };
    document.getElementById("restart-task").onclick = function() {
      runLifecycle(selectedTask, "restart", lastAction, function() {
        if (selectedTask) loadTaskDetail(selectedTask.namespace, selectedTask.name);
      });
    };
    document.getElementById("rerun-task").onclick = function() {
      runLifecycle(selectedTask, "rerun", lastAction, refreshTasks);
    };
    document.getElementById("refresh-logs").onclick = function() {
      loadLogs(selectedTask, {
        pod: logPod,
        container: logContainer,
        tail: logTail,
        output: logsOutput
      });
    };
    document.getElementById("user-tab").onclick = function() { setActiveView("user"); };
    document.getElementById("monitor-tab").onclick = function() { setActiveView("monitor"); };
    document.getElementById("monitor-refresh").onclick = refreshMonitor;
    document.getElementById("monitor-delete-task").onclick = function() {
      runLifecycle(selectedMonitorTask, "delete", monitorAction, function() {
        selectedMonitorTask = null;
        refreshMonitor();
      });
    };
    document.getElementById("monitor-restart-task").onclick = function() {
      runLifecycle(selectedMonitorTask, "restart", monitorAction, function() {
        if (selectedMonitorTask) loadMonitorDetail(selectedMonitorTask.namespace, selectedMonitorTask.name);
      });
    };
    document.getElementById("monitor-rerun-task").onclick = function() {
      runLifecycle(selectedMonitorTask, "rerun", monitorAction, refreshMonitor);
    };
    document.getElementById("monitor-refresh-logs").onclick = function() {
      loadLogs(selectedMonitorTask, {
        pod: monitorLogPod,
        container: monitorLogContainer,
        tail: monitorLogTail,
        output: monitorLogsOutput
      });
    };
    document.getElementById("refresh-audits").onclick = refreshAudits;
    Object.keys(monitorFields).forEach(function(id) {
      monitorFields[id].addEventListener("change", refreshMonitor);
    });
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
