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
    .template-strip {
      display: grid;
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 12px;
      padding: 16px 18px 2px;
    }
    .template-card {
      border: 1px solid var(--line);
      border-radius: 18px;
      padding: 14px;
      color: var(--ink);
      background: rgba(255,255,255,.72);
      text-align: left;
      cursor: pointer;
      min-height: 112px;
    }
    .template-card.active {
      border-color: rgba(47, 95, 116, .62);
      background: linear-gradient(135deg, rgba(47, 95, 116, .14), rgba(232, 236, 223, .82));
      box-shadow: inset 0 0 0 1px rgba(47, 95, 116, .18);
    }
    .template-card strong {
      display: block;
      margin-bottom: 8px;
      font-size: 15px;
    }
    .template-card span {
      display: block;
      color: var(--muted);
      font-size: 12px;
      line-height: 1.45;
    }
    .wizard-steps {
      display: grid;
      grid-template-columns: repeat(4, minmax(0, 1fr));
      gap: 8px;
      padding: 14px 18px 0;
    }
    .wizard-step {
      border-radius: 999px;
      padding: 8px 10px;
      color: var(--muted);
      background: rgba(255,255,255,.62);
      border: 1px solid var(--line);
      font-size: 12px;
      font-weight: 900;
      text-align: center;
    }
    .wizard-step.active {
      color: #fffaf0;
      background: var(--steel);
    }
    details.advanced-config {
      margin: 10px 18px 14px;
      border: 1px solid var(--line);
      border-radius: 18px;
      background: rgba(255,255,255,.42);
      overflow: hidden;
    }
    details.advanced-config summary {
      cursor: pointer;
      padding: 13px 14px;
      font-weight: 900;
      color: var(--ink);
    }
    details.advanced-config .form-grid {
      padding-top: 6px;
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
      .template-strip { grid-template-columns: 1fr; }
      .wizard-steps { grid-template-columns: 1fr 1fr; }
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
          <h2 class="panel-title">训练向导</h2>
          <div class="actions">
            <button class="ghost" id="load-service">发布已有模型</button>
            <button class="ghost" id="load-job">载入训练模板</button>
            <button class="ghost" id="generate">生成 YAML</button>
            <button class="secondary" id="validate">校验</button>
            <button id="normalize">归一化</button>
            <button id="deploy">提交部署</button>
          </div>
        </div>
        <div class="template-strip" aria-label="算法任务模板">
          <button class="template-card active" id="template-sft" type="button">
            <strong>SFT 微调</strong>
            <span>选择基础模型和训练数据，提交一次监督微调任务，完成后可发布为服务。</span>
          </button>
          <button class="template-card" id="template-eval" type="button">
            <strong>模型评测</strong>
            <span>用评测集验证模型效果，输出指标和评估结果，适合上线前检查。</span>
          </button>
          <button class="template-card" id="template-service" type="button">
            <strong>发布服务</strong>
            <span>已有 modelURI 时直接发布为在线服务，并通过 /healthz 做访问测试。</span>
          </button>
        </div>
        <div class="wizard-steps">
          <div class="wizard-step active">1 选择算法</div>
          <div class="wizard-step active">2 填写模型与数据</div>
          <div class="wizard-step">3 提交训练</div>
          <div class="wizard-step">4 发布与测试</div>
        </div>
        <div class="form-grid">
          <div class="subhead">算法任务模板</div>
          <div class="field">
            <label for="algorithmTemplate">算法类型</label>
            <select id="algorithmTemplate">
              <option value="sft">SFT 微调</option>
              <option value="evaluation">模型评测</option>
              <option value="service">发布已有模型</option>
            </select>
          </div>
          <div class="field">
            <label for="taskDisplayName">任务名称</label>
            <input id="taskDisplayName" value="customer-sft-demo">
          </div>
          <div class="field wide train-field">
            <label for="baseModelURI">基础模型 URI</label>
            <input id="baseModelURI" value="modelscope://qwen/Qwen2.5-0.5B">
          </div>
          <div class="field wide train-field">
            <label for="trainingDatasetSelect">训练数据集</label>
            <select id="trainingDatasetSelect">
              <option value="inline://datasets/customer-sft-demo">客服问答 SFT 数据集（示例）</option>
            </select>
          </div>
          <div class="field wide train-field">
            <label for="trainingDataURI">数据集内部 URI</label>
            <input id="trainingDataURI" value="inline://datasets/customer-sft-demo">
          </div>
          <div class="field train-field">
            <label for="trainingSize">训练规格</label>
            <select id="trainingSize">
              <option value="small">小规格 CPU PoC</option>
              <option value="medium">中规格 单卡</option>
              <option value="large">大规格 多卡</option>
            </select>
          </div>
          <div class="field train-field">
            <label for="publishAfterTrain">训练完成后</label>
            <select id="publishAfterTrain">
              <option value="true">训练完成后自动发布为服务</option>
              <option value="false">仅保存模型产物</option>
            </select>
          </div>
          <div class="field train-field">
            <label for="epochs">Epoch</label>
            <input id="epochs" value="1">
          </div>
          <div class="field train-field">
            <label for="learningRate">Learning Rate</label>
            <input id="learningRate" value="2e-5">
          </div>
          <div class="field service-template-field wide">
            <label for="serviceModelURI">待发布模型 URI</label>
            <input id="serviceModelURI" value="inline://models/customer-sft-demo/v1">
          </div>
        </div>
        <details class="advanced-config">
          <summary>高级配置：查看和调整底层 AIJob / AIService 字段</summary>
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
        </details>
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
            <button class="ghost" id="probe-delivery-service">测试服务访问</button>
            <button class="ghost" id="refresh-artifacts">查看模型产物</button>
          </div>
          <pre id="delivery-output" class="log-box">选择 AIJob 后，可解析 AI_RESULT_JSON 并发布为 AIService。发布成功后可测试服务 /healthz。</pre>
          <div class="log-toolbar">
            <div class="label">我的模型 · 模型资产库</div>
            <input id="evaluation-dataset-uri" value="inline://datasets/customer-eval" title="评测数据 URI">
            <input id="evaluation-threshold" value="0.8" title="通过阈值">
            <button class="ghost" id="refresh-models">刷新我的模型</button>
            <button class="secondary" id="register-model">登记模型</button>
          </div>
          <div id="model-list" class="task-list">
            <div class="empty">训练完成后可登记模型资产；已登记模型可以“设为基础模型”继续训练，或“发布此模型”生成服务模板。</div>
          </div>
          <div class="log-toolbar">
            <div class="label">数据资产库</div>
            <input id="dataset-name" value="customer-sft-demo" title="数据集名称">
            <input id="dataset-uri" value="inline://datasets/customer-sft-demo" title="数据集内部 URI">
            <select id="dataset-format" title="数据格式">
              <option value="sharegpt-jsonl">ShareGPT JSONL</option>
              <option value="alpaca-jsonl">Alpaca JSONL</option>
              <option value="custom-jsonl">自定义 JSONL</option>
            </select>
            <button class="ghost" id="refresh-datasets">刷新数据集</button>
            <button class="secondary" id="register-dataset">登记数据集</button>
          </div>
          <div id="dataset-list" class="task-list">
            <div class="empty">用户上传或登记数据集后，可在训练向导中“选择数据集”，平台内部再映射为 datasetURI。</div>
          </div>
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
    var modelList = document.getElementById("model-list");
    var datasetList = document.getElementById("dataset-list");
    var lastAction = document.getElementById("last-action");
    var selectedTask = null;
    var selectedMonitorTask = null;
    var lastPublishedService = null;
    var lastDeliveryResult = null;
    var logPod = document.getElementById("log-pod");
    var logContainer = document.getElementById("log-container");
    var logTail = document.getElementById("log-tail");
    var logsOutput = document.getElementById("logs-output");
    var deliveryOutput = document.getElementById("delivery-output");
    var deliveryServiceName = document.getElementById("delivery-service-name");
    var deliveryServiceImage = document.getElementById("delivery-service-image");
    var evaluationDatasetURI = document.getElementById("evaluation-dataset-uri");
    var evaluationThreshold = document.getElementById("evaluation-threshold");
    var trainingDatasetSelect = document.getElementById("trainingDatasetSelect");
    var datasetName = document.getElementById("dataset-name");
    var datasetURI = document.getElementById("dataset-uri");
    var datasetFormat = document.getElementById("dataset-format");
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
    var wizardFields = {};
    ["algorithmTemplate", "taskDisplayName", "baseModelURI", "trainingDataURI", "trainingSize", "publishAfterTrain", "epochs", "learningRate", "serviceModelURI"].forEach(function(id) {
      wizardFields[id] = document.getElementById(id);
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
    function setWizardValue(id, value) {
      wizardFields[id].value = value;
    }
    function setTrainingDataset(uri) {
      setWizardValue("trainingDataURI", uri || "");
      if (trainingDatasetSelect && uri) {
        trainingDatasetSelect.value = uri;
      }
      if (datasetURI && uri) {
        datasetURI.value = uri;
      }
    }
    function slug(value) {
      return (value || "ai-task").toLowerCase().replace(/[^a-z0-9-]+/g, "-").replace(/^-+|-+$/g, "").slice(0, 50) || "ai-task";
    }
    function modelNameFromURI(uri) {
      var clean = (uri || "model").split("?")[0].replace(/\/+$/g, "");
      var parts = clean.split("/");
      var name = parts[parts.length - 1] || "model";
      if (looksLikeModelVersion(name) && parts.length > 1) {
        name = parts[parts.length - 2] || name;
      }
      return slug(name);
    }
    function looksLikeModelVersion(value) {
      value = (value || "").toLowerCase();
      return value === "latest" || /^v[0-9]+$/.test(value);
    }
    function setActiveTemplate(template) {
      Array.prototype.forEach.call(document.querySelectorAll(".template-card"), function(node) {
        node.classList.remove("active");
      });
      var active = document.getElementById("template-" + (template === "evaluation" ? "eval" : template === "service" ? "service" : "sft"));
      if (active) active.classList.add("active");
      Array.prototype.forEach.call(document.querySelectorAll(".train-field"), function(node) {
        node.style.display = template === "service" ? "none" : "grid";
      });
      Array.prototype.forEach.call(document.querySelectorAll(".service-template-field"), function(node) {
        node.style.display = template === "service" ? "grid" : "none";
      });
    }
    function applyWizardToAdvanced() {
      var template = wizardFields.algorithmTemplate.value;
      var taskName = slug(wizardFields.taskDisplayName.value);
      setActiveTemplate(template);
      if (template === "service") {
        setValue("kind", "AIService");
        setValue("name", taskName || "model-service-demo");
        setValue("namespace", fields.namespace.value || "sock-shop");
        setValue("component", taskName || "model-service");
        setValue("tenant", fields.tenant.value || "demo-tenant");
        setValue("project", fields.project.value || "model-serving");
        setValue("environment", fields.environment.value || "poc");
        setValue("owner", fields.owner.value || "ai-platform");
        setValue("runtime", "http");
        setValue("image", "python:3.11-slim");
        setValue("modelName", modelNameFromURI(wizardFields.serviceModelURI.value));
        setValue("modelVersion", "v1");
        setValue("modelURI", wizardFields.serviceModelURI.value);
        setValue("replicas", "1");
        setValue("port", "8080");
        return;
      }
      var project = template === "evaluation" ? "model-evaluation" : "sft-training";
      setValue("kind", "AIJob");
      setValue("name", taskName || "customer-sft-demo");
      setValue("namespace", fields.namespace.value || "sock-shop");
      setValue("component", taskName ? taskName + "-trainer" : "sft-trainer");
      setValue("tenant", fields.tenant.value || "demo-tenant");
      setValue("project", project);
      setValue("environment", fields.environment.value || "poc");
      setValue("owner", fields.owner.value || "ai-platform");
      setValue("runtime", "batch");
      setValue("image", "busybox:1.36");
      setValue("jobKind", template === "evaluation" ? "evaluation" : "training");
      setValue("ttl", "3600");
      setValue("datasetURI", wizardFields.trainingDataURI.value);
      setValue("outputURI", "inline://outputs/" + (taskName || "customer-sft-demo"));
      deliveryServiceName.value = (taskName || "customer-sft-demo") + "-service";
    }
    function deliveryResultJSONString() {
      return JSON.stringify({
        modelURI: "inline://models/" + fields.name.value + "/v1",
        metrics: { loss: 0.12, accuracy: 0.98 },
        summary: (wizardFields.algorithmTemplate && wizardFields.algorithmTemplate.value === "sft") ? "sft-fine-tuned" : "trained-for-delivery"
      });
    }
    function buildDeliveryJobResultScript() {
      return [
        "    imagePullPolicy: IfNotPresent",
        "    cmd:",
        "      - sh",
        "      - -c",
        "    args:",
        "      - |",
        "        echo algorithm=" + (wizardFields.algorithmTemplate ? wizardFields.algorithmTemplate.value : "training"),
        "        echo base_model=" + (wizardFields.baseModelURI ? wizardFields.baseModelURI.value : "inline://models/base"),
        "        echo dataset=" + fields.datasetURI.value,
        "        echo training_size=" + (wizardFields.trainingSize ? wizardFields.trainingSize.value : "small"),
        "        echo epochs=" + (wizardFields.epochs ? wizardFields.epochs.value : "1") + " learning_rate=" + (wizardFields.learningRate ? wizardFields.learningRate.value : "2e-5"),
        "        echo train-start",
        "        echo epoch=1 loss=0.30",
        "        echo epoch=2 loss=0.12",
        "        echo 'AI_RESULT_JSON=" + deliveryResultJSONString() + "'",
        "        echo train-complete",
        "    backoffLimit: 0"
      ];
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
        ]);
        common = common.concat(buildDeliveryJobResultScript());
        common = common.concat([
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
        common.push(line("trainingSize", wizardFields.trainingSize.value, 4));
      }
      return common.join("\n") + "\n";
    }
    function generateYAML(applyWizard) {
      if (applyWizard !== false) {
        applyWizardToAdvanced();
      }
      syncVisibility();
      yaml.value = buildYAML();
    }
    function fillServiceForm() {
      setWizardValue("algorithmTemplate", "service");
      setWizardValue("taskDisplayName", "sentiment-demo");
      setWizardValue("serviceModelURI", "oss://models/sentiment/v1");
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
      setWizardValue("algorithmTemplate", "sft");
      setWizardValue("taskDisplayName", "customer-sft-demo");
      setWizardValue("baseModelURI", "modelscope://qwen/Qwen2.5-0.5B");
      setWizardValue("trainingDataURI", "inline://datasets/customer-sft-demo");
      setWizardValue("publishAfterTrain", "true");
      setValue("kind", "AIJob");
      setValue("name", "customer-sft-demo");
      setValue("namespace", "sock-shop");
      setValue("component", "customer-sft-demo-trainer");
      setValue("tenant", "demo-tenant");
      setValue("project", "sft-training");
      setValue("environment", "poc");
      setValue("owner", "ai-platform");
      setValue("runtime", "batch");
      setValue("image", "busybox:1.36");
      setValue("jobKind", "training");
      setValue("ttl", "3600");
      setValue("datasetURI", "inline://datasets/customer-sft-demo");
      setValue("outputURI", "inline://outputs/customer-sft-demo");
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
        lastDeliveryResult = data.result || null;
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
        lastPublishedService = {namespace: data.namespace, name: data.serviceName};
        selectedTask = lastPublishedService;
        deliveryOutput.textContent = JSON.stringify(data, null, 2) + "\\n\\n已创建 AIService：" + data.namespace + "/" + data.serviceName + "\\n下一步：点击“测试服务访问”，验证 /healthz 是否可达。";
        refreshTasks().then(function() {
          return loadTaskDetail(data.namespace, data.serviceName);
        });
        refreshAudits();
      }).catch(function(err) {
        deliveryOutput.textContent = "发布服务失败：" + err.message + "\\n\\n请先解析训练结果，确认该 AIJob 日志中存在 modelURI。";
      });
    }
    function serviceTargetFromSelectionOrPublished() {
      if (lastPublishedService) {
        return lastPublishedService;
      }
      if (selectedTask) {
        return selectedTask;
      }
      if (deliveryServiceName.value && fields.namespace.value) {
        return {namespace: fields.namespace.value, name: deliveryServiceName.value};
      }
      return null;
    }
    function probeDeliveryService() {
      var target = serviceTargetFromSelectionOrPublished();
      if (!target) {
        deliveryOutput.textContent = "请先发布一个 AIService，或在任务列表中选择一个 AIService。";
        return Promise.resolve();
      }
      deliveryOutput.textContent = "正在测试服务访问 /healthz...";
      return fetch("/api/v1/ai/applications/" + encodeURIComponent(target.namespace) + "/" + encodeURIComponent(target.name) + "/probe", {
        method: "POST",
        headers: {"Content-Type": "application/json", "X-AI-User": "console"},
        body: JSON.stringify({path: "/healthz", timeoutSeconds: 5})
      }).then(function(res) {
        return res.json().then(function(data) {
          if (!res.ok) {
            throw new Error(data.error || ("HTTP " + res.status));
          }
          return data;
        });
      }).then(function(data) {
        deliveryOutput.textContent = JSON.stringify(data, null, 2);
        if (data.healthy) {
          lastAction.textContent = "服务访问正常";
        } else {
          lastAction.textContent = "服务访问异常";
        }
      }).catch(function(err) {
        deliveryOutput.textContent = "服务访问测试失败：" + err.message + "\\n\\n请确认 AIService 已 Ready，且服务容器暴露 /healthz。";
      });
    }
    function refreshArtifacts() {
      var ns = encodeURIComponent(fields.namespace.value || "");
      deliveryOutput.textContent = "正在读取模型产物...";
      return fetchJSON("/api/v1/ai/artifacts?namespace=" + ns).then(function(data) {
        deliveryOutput.textContent = JSON.stringify(data, null, 2);
      }).catch(function(err) {
        deliveryOutput.textContent = "模型产物读取失败：" + err.message;
      });
    }
    function renderDatasets(items) {
      datasetList.innerHTML = "";
      trainingDatasetSelect.innerHTML = "";
      if (!items || !items.length) {
        var option = document.createElement("option");
        option.value = wizardFields.trainingDataURI.value || "inline://datasets/customer-sft-demo";
        option.textContent = "客服问答 SFT 数据集（示例）";
        trainingDatasetSelect.appendChild(option);
        var empty = document.createElement("div");
        empty.className = "empty";
        empty.textContent = "暂无数据资产。可以先登记一个数据集，再在训练向导中选择数据集。";
        datasetList.appendChild(empty);
        return;
      }
      items.forEach(function(item) {
        var uri = item.datasetURI || "";
        var option = document.createElement("option");
        option.value = uri;
        option.textContent = (item.displayName || item.name || uri) + " · " + (item.format || "unknown") + " · " + (item.status || "registered");
        trainingDatasetSelect.appendChild(option);
        var row = document.createElement("div");
        row.className = "task-row";
        row.innerHTML = [
          "<strong>" + ((item.namespace || fields.namespace.value || "-") + "/" + (item.displayName || item.name || "-")) + "</strong>",
          "<span title=\"" + uri + "\">" + (uri || "-") + "</span>",
          "<span>" + (item.purpose || "sft") + " · " + (item.format || "unknown") + "</span>",
          "<span>" + (item.status || "registered") + " · " + (item.visibility || "private") + "</span>",
          "<span><button class=\"ghost use-dataset\" type=\"button\">选择数据集</button></span>"
        ].join("");
        row.querySelector(".use-dataset").onclick = function() {
          setTrainingDataset(uri);
          generateYAML(true);
          lastAction.textContent = "已选择数据集";
        };
        datasetList.appendChild(row);
      });
      if (wizardFields.trainingDataURI.value) {
        trainingDatasetSelect.value = wizardFields.trainingDataURI.value;
      } else if (items[0] && items[0].datasetURI) {
        setTrainingDataset(items[0].datasetURI);
      }
    }
    function refreshDatasets() {
      var ns = encodeURIComponent(fields.namespace.value || "");
      datasetList.innerHTML = "<div class=\"empty\">正在读取数据资产库...</div>";
      return fetchJSON("/api/v1/ai/datasets?namespace=" + ns).then(function(data) {
        renderDatasets(data.items || []);
      }).catch(function(err) {
        datasetList.innerHTML = "<div class=\"empty\">数据资产读取失败：" + err.message + "</div>";
      });
    }
    function registerDatasetAsset() {
      var uri = datasetURI.value || wizardFields.trainingDataURI.value || "";
      if (!uri) {
        deliveryOutput.textContent = "请先填写或选择一个数据集内部 URI。";
        return Promise.resolve();
      }
      var name = slug(datasetName.value || uri);
      var payload = {
        namespace: fields.namespace.value || "default",
        name: name,
        displayName: datasetName.value || name,
        datasetURI: uri,
        format: datasetFormat.value || "sharegpt-jsonl",
        purpose: wizardFields.algorithmTemplate.value === "evaluation" ? "evaluation" : "sft",
        status: "validated",
        visibility: "private",
        source: "console"
      };
      deliveryOutput.textContent = "正在登记数据集...";
      return fetch("/api/v1/ai/datasets", {
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
        setTrainingDataset(data.datasetURI || uri);
        deliveryOutput.textContent = JSON.stringify(data, null, 2) + "\\n\\n数据资产已登记，可在训练向导中选择。";
        return refreshDatasets();
      }).catch(function(err) {
        deliveryOutput.textContent = "数据集登记失败：" + err.message;
      });
    }
    function renderModels(items) {
      modelList.innerHTML = "";
      if (!items || !items.length) {
        var empty = document.createElement("div");
        empty.className = "empty";
        empty.textContent = "暂无模型资产。训练完成后点击“解析训练结果”，再点击“登记模型”。";
        modelList.appendChild(empty);
        return;
      }
      items.forEach(function(item) {
        var row = document.createElement("div");
        row.className = "task-row";
        var name = (item.namespace || fields.namespace.value || "-") + "/" + (item.name || "-");
        row.innerHTML = [
          "<strong>" + name + "</strong>",
          "<span title=\"" + (item.modelURI || "") + "\">" + (item.modelURI || "-") + "</span>",
          "<span>" + (item.evaluationStatus || "pending") + "</span>",
          "<span>" + (item.status || "trained") + " · " + (item.visibility || "private") + "</span>",
          "<span><button class=\"ghost eval-model\" type=\"button\">发起评测</button> <button class=\"ghost sync-eval\" type=\"button\">同步评测结果</button> <button class=\"ghost use-model\" type=\"button\">设为基础模型</button> <button class=\"secondary serve-model\" type=\"button\">发布此模型</button></span>"
        ].join("");
        row.querySelector(".eval-model").onclick = function() {
          startModelEvaluation(item);
        };
        row.querySelector(".sync-eval").onclick = function() {
          syncModelEvaluation(item);
        };
        row.querySelector(".use-model").onclick = function() {
          setWizardValue("algorithmTemplate", "sft");
          setWizardValue("baseModelURI", item.modelURI || "");
          setWizardValue("taskDisplayName", (item.name || "model") + "-sft");
          generateYAML(true);
          lastAction.textContent = "已将模型设为基础模型";
        };
        row.querySelector(".serve-model").onclick = function() {
          setWizardValue("algorithmTemplate", "service");
          setWizardValue("serviceModelURI", item.modelURI || "");
          setWizardValue("taskDisplayName", (item.name || "model") + "-service");
          generateYAML(true);
          lastAction.textContent = "已载入发布此模型模板";
        };
        modelList.appendChild(row);
      });
    }
    function refreshModels() {
      var ns = encodeURIComponent(fields.namespace.value || "");
      modelList.innerHTML = "<div class=\"empty\">正在读取模型资产库...</div>";
      return fetchJSON("/api/v1/ai/models?namespace=" + ns).then(function(data) {
        renderModels(data.items || []);
      }).catch(function(err) {
        modelList.innerHTML = "<div class=\"empty\">模型资产读取失败：" + err.message + "</div>";
      });
    }
    function registerModelAsset() {
      var target = deliveryTargetFromSelectionOrForm();
      var result = lastDeliveryResult || {};
      var modelURI = result.modelURI || "";
      if (!modelURI && fields.kind.value === "AIService") {
        modelURI = fields.modelURI.value || wizardFields.serviceModelURI.value;
      }
      if (!modelURI) {
        deliveryOutput.textContent = "请先解析训练结果，或在高级配置中填写 modelURI 后再登记模型。";
        return Promise.resolve();
      }
      var payload = {
        namespace: fields.namespace.value || "default",
        name: modelNameFromURI(modelURI),
        jobName: target ? target.name : modelNameFromURI(modelURI),
        modelURI: modelURI,
        baseModelURI: wizardFields.baseModelURI.value || "",
        datasetURI: wizardFields.trainingDataURI.value || fields.datasetURI.value || "",
        status: "trained",
        visibility: "private",
        evaluationStatus: result.metrics ? "passed" : "pending",
        metrics: result.metrics || undefined,
        summary: result.summary || ""
      };
      return fetch("/api/v1/ai/models", {
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
        deliveryOutput.textContent = JSON.stringify(data, null, 2) + "\\n\\n模型资产已登记，可在“我的模型”中继续训练或发布。";
        return refreshModels();
      }).catch(function(err) {
        deliveryOutput.textContent = "模型登记失败：" + err.message;
      });
    }
    function modelActionBase(item) {
      return "/api/v1/ai/models/" + encodeURIComponent(item.namespace || fields.namespace.value || "default") + "/" + encodeURIComponent(item.name || modelNameFromURI(item.modelURI || ""));
    }
    function startModelEvaluation(item) {
      var threshold = parseFloat(evaluationThreshold.value || "0.8");
      if (isNaN(threshold)) {
        threshold = 0.8;
      }
      var payload = {
        evaluationDatasetURI: evaluationDatasetURI.value || "inline://datasets/customer-eval",
        evaluationType: "accuracy",
        passThreshold: threshold
      };
      deliveryOutput.textContent = "正在发起模型评测...";
      return fetch(modelActionBase(item) + "/evaluate", {
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
        selectedTask = {namespace: data.namespace, name: data.jobName};
        deliveryOutput.textContent = JSON.stringify(data, null, 2) + "\\n\\n评测 Job 已提交。等待完成后点击“同步评测结果”。";
        refreshTasks();
      }).catch(function(err) {
        deliveryOutput.textContent = "发起评测失败：" + err.message;
      });
    }
    function syncModelEvaluation(item) {
      var jobName = (item.name || modelNameFromURI(item.modelURI || "")) + "-eval";
      deliveryOutput.textContent = "正在同步评测结果...";
      return fetch(modelActionBase(item) + "/sync-evaluation", {
        method: "POST",
        headers: {"Content-Type": "application/json", "X-AI-User": "console"},
        body: JSON.stringify({evaluationJobName: jobName})
      }).then(function(res) {
        return res.json().then(function(data) {
          if (!res.ok) {
            throw new Error(data.error || ("HTTP " + res.status));
          }
          return data;
        });
      }).then(function(data) {
        deliveryOutput.textContent = JSON.stringify(data, null, 2) + "\\n\\n评测结果已同步到模型资产。";
        return refreshModels();
      }).catch(function(err) {
        deliveryOutput.textContent = "同步评测结果失败：" + err.message + "\\n\\n请确认评测 Job 已完成，日志中包含 AI_RESULT_JSON=...。";
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
      generateYAML(false);
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
      deliveryOutput.textContent = "已载入 SFT 微调模板。下一步：确认基础模型和训练数据，取消 DryRun 后点击“提交部署”。";
      lastAction.textContent = "已载入 SFT 微调模板";
    };
    document.getElementById("template-sft").onclick = function() {
      setWizardValue("algorithmTemplate", "sft");
      if (!wizardFields.taskDisplayName.value || wizardFields.taskDisplayName.value === "sentiment-demo") {
        setWizardValue("taskDisplayName", "customer-sft-demo");
      }
      generateYAML();
      lastAction.textContent = "已选择 SFT 微调";
    };
    document.getElementById("template-eval").onclick = function() {
      setWizardValue("algorithmTemplate", "evaluation");
      if (!wizardFields.taskDisplayName.value || wizardFields.taskDisplayName.value === "customer-sft-demo") {
        setWizardValue("taskDisplayName", "model-eval-demo");
      }
      generateYAML();
      lastAction.textContent = "已选择模型评测";
    };
    document.getElementById("template-service").onclick = function() {
      setWizardValue("algorithmTemplate", "service");
      if (!wizardFields.taskDisplayName.value || wizardFields.taskDisplayName.value === "customer-sft-demo") {
        setWizardValue("taskDisplayName", "model-service-demo");
      }
      generateYAML();
      lastAction.textContent = "已选择发布服务";
    };
    document.getElementById("generate").onclick = function() {
      generateYAML(true);
      lastAction.textContent = "已生成 YAML";
    };
    Object.keys(wizardFields).forEach(function(id) {
      wizardFields[id].addEventListener("input", function() { generateYAML(true); });
      wizardFields[id].addEventListener("change", function() { generateYAML(true); });
    });
    Object.keys(fields).forEach(function(id) {
      fields[id].addEventListener("input", function() { generateYAML(false); });
      fields[id].addEventListener("change", function() { generateYAML(false); });
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
    document.getElementById("probe-delivery-service").onclick = probeDeliveryService;
    document.getElementById("refresh-artifacts").onclick = refreshArtifacts;
    document.getElementById("refresh-models").onclick = refreshModels;
    document.getElementById("register-model").onclick = registerModelAsset;
    document.getElementById("refresh-datasets").onclick = refreshDatasets;
    document.getElementById("register-dataset").onclick = registerDatasetAsset;
    trainingDatasetSelect.onchange = function() {
      setTrainingDataset(trainingDatasetSelect.value);
      generateYAML(true);
    };
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
