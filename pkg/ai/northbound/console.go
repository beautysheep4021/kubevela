package northbound

const consoleHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>智算纳管北向验证台</title>
    <style>
    :root {
      --ink: #16191f;
      --muted: #5f6b7a;
      --paper: #f2f3f3;
      --panel: #ffffff;
      --line: #d5d9d9;
      --field: #f7f8f8;
      --moss: #16884a;
      --rust: #d13212;
      --blue: #1677ff;
      --steel: #161d26;
      --shadow: 0 8px 28px rgba(22, 25, 31, .14);
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      color: var(--ink);
      font-family: -apple-system, "Segoe UI", system-ui, "PingFang SC", "Microsoft YaHei", Roboto, sans-serif;
      background: var(--paper);
    }
    body:before {
      content: none;
    }
    main {
      width: min(1400px, calc(100vw - 20px));
      margin: 0 auto;
      padding: 14px 0 22px;
      display: grid;
      grid-template-columns: minmax(220px, 250px) minmax(0, 1fr);
      gap: 12px;
      align-items: start;
    }
    .sidebar {
      position: sticky;
      top: 14px;
      display: grid;
      gap: 12px;
      align-content: start;
      padding: 14px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: var(--panel);
      box-shadow: var(--shadow);
    }
    .sidebar-brand {
      display: grid;
      gap: 6px;
    }
    .sidebar-title {
      font-size: 18px;
      font-weight: 900;
      line-height: 1.2;
      color: var(--ink);
    }
    .sidebar-lede {
      color: var(--muted);
      font-size: 12px;
      line-height: 1.5;
    }
    .sidebar-section {
      display: grid;
      gap: 8px;
    }
    .sidebar-section-title {
      color: #5f6b7a;
      font-size: 11px;
      font-weight: 900;
      letter-spacing: .12em;
    }
    .sidebar-nav {
      display: grid;
      gap: 8px;
    }
    .nav-item {
      width: 100%;
      justify-content: flex-start;
      color: var(--ink);
      background: #f7f8f8;
      border: 1px solid var(--line);
      border-radius: 8px;
      text-align: left;
    }
    .nav-item.active {
      color: #0b5cad;
      background: #eaf3ff;
      border-color: #cfe1ff;
    }
    .sidebar-actions {
      display: grid;
      gap: 8px;
    }
    .content {
      display: grid;
      gap: 12px;
      min-width: 0;
    }
    .hero {
      display: grid;
      grid-template-columns: minmax(0, 1fr) minmax(320px, .72fr);
      gap: 12px;
      align-items: center;
      margin-bottom: 12px;
      padding-bottom: 12px;
      border-bottom: 1px solid var(--line);
    }
    h1 {
      margin: 0;
      font-size: clamp(28px, 3vw, 40px);
      line-height: 1.08;
      letter-spacing: 0;
    }
    .lede {
      max-width: 720px;
      margin: 10px 0 0;
      color: var(--muted);
      font-size: 13px;
      line-height: 1.55;
    }
    .status-strip {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 8px;
    }
    .status-card, .panel, .intent-card {
      border: 1px solid var(--line);
      background: var(--panel);
      box-shadow: var(--shadow);
      backdrop-filter: none;
    }
    .status-card {
      padding: 10px 11px;
      border-radius: 8px;
    }
    .tabs {
      display: inline-flex;
      gap: 8px;
      padding: 4px;
      margin: 14px 0 0;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: var(--panel);
      box-shadow: none;
    }
    .tab {
      color: var(--ink);
      background: #f7f8f8;
      border: 1px solid transparent;
      border-radius: 8px;
    }
    .tab.active {
      color: #0b5cad;
      background: #eaf3ff;
    }
    body[data-role="user"] .tabs,
    body[data-role="monitor"] .tabs {
      display: none;
    }
    body[data-role="user"] .monitor-nav,
    body[data-role="monitor"] .user-nav {
      display: none;
    }
    body[data-role="user"] #monitor-view,
    body[data-role="monitor"] #user-view {
      display: none !important;
    }
    .role-badge {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      min-height: 30px;
      padding: 0 10px;
      border: 1px solid var(--line);
      border-radius: 8px;
      color: #0b5cad;
      background: #eef4ff;
      font-size: 12px;
      font-weight: 900;
    }
    .view { display: none; }
    .view.active { display: block; }
    .section-pane { display: none; }
    .section-pane.active { display: block; }
    .internal-field { display: none !important; }
    .label {
      color: var(--muted);
      font-size: 11px;
      font-weight: 800;
      letter-spacing: .12em;
    }
    .metric {
      margin-top: 8px;
      font-size: 16px;
      font-weight: 900;
    }
    .workspace {
      display: grid;
      grid-template-columns: 1fr;
      gap: 12px;
      align-items: start;
    }
    .monitor-grid {
      display: grid;
      grid-template-columns: 1fr;
      gap: 12px;
      align-items: start;
    }
    .panel {
      min-height: 0;
      border-radius: 8px;
      overflow: hidden;
    }
    .panel-head {
      display: flex;
      justify-content: space-between;
      gap: 16px;
      align-items: center;
      padding: 12px 14px;
      border-bottom: 1px solid var(--line);
      background: #fbfcfe;
    }
    .panel-title {
      margin: 0;
      font-size: 15px;
      font-weight: 900;
    }
    .actions {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
    }
    .deploy-options {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 0 14px 12px;
      color: var(--muted);
      font-size: 12px;
      font-weight: 900;
    }
    .template-strip {
      display: grid;
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 8px;
      padding: 10px 14px 0;
    }
    .workflow-summary {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 8px;
      padding: 10px 14px 0;
    }
    .workflow-chip {
      display: grid;
      gap: 4px;
      padding: 10px 12px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: #f7f8f8;
    }
    .workflow-chip strong {
      color: var(--ink);
      font-size: 13px;
      font-weight: 900;
    }
    .workflow-chip span {
      color: var(--muted);
      font-size: 12px;
      line-height: 1.4;
    }
    .template-card {
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 11px;
      color: var(--ink);
      background: var(--panel);
      text-align: left;
      cursor: pointer;
      min-height: 96px;
    }
    .template-card.active {
      border-color: #cfe1ff;
      background: #f5f9ff;
      box-shadow: inset 0 0 0 1px rgba(22, 119, 255, .10);
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
      gap: 6px;
      padding: 10px 14px 0;
    }
    .wizard-step {
      border-radius: 8px;
      padding: 7px 9px;
      color: #5f6b7a;
      background: #f7f8f8;
      border: 1px solid var(--line);
      font-size: 12px;
      font-weight: 900;
      text-align: center;
    }
    .wizard-step.active {
      color: #0b5cad;
      background: #eaf3ff;
    }
    details.fold-section,
    details.advanced-config {
      grid-column: 1 / -1;
      margin: 10px 14px 12px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: var(--panel);
      overflow: hidden;
    }
    details.fold-section > summary,
    details.advanced-config > summary {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
      cursor: pointer;
      padding: 11px 14px;
      list-style: none;
      font-weight: 900;
      color: var(--ink);
    }
    details.fold-section > summary::-webkit-details-marker,
    details.advanced-config > summary::-webkit-details-marker {
      display: none;
    }
    details.fold-section > summary::after,
    details.advanced-config > summary::after {
      content: "▾";
      color: var(--muted);
      font-size: 12px;
    }
    details.fold-section[open] > summary::after,
    details.advanced-config[open] > summary::after {
      content: "▴";
    }
    .fold-body {
      padding: 0 14px 14px;
    }
    details.fold-section .form-grid,
    details.advanced-config .form-grid {
      padding-top: 6px;
    }
    .form-grid {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 12px;
      padding: 14px 14px 4px;
    }
    .filter-grid {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 12px;
      padding: 14px;
    }
    .field {
      display: grid;
      gap: 6px;
    }
    .field.wide { grid-column: 1 / -1; }
    .field label {
      color: #5f6b7a;
      font-size: 12px;
      font-weight: 900;
    }
    input, select {
      width: 100%;
      min-height: 36px;
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 8px 10px;
      color: var(--ink);
      background: var(--field);
      outline: none;
      font: inherit;
    }
    input::placeholder {
      color: rgba(154, 166, 178, .72);
    }
    input[type="checkbox"] {
      width: auto;
      min-height: auto;
      accent-color: var(--moss);
    }
    input:focus, select:focus {
      border-color: #69a7ff;
      box-shadow: 0 0 0 3px rgba(105, 167, 255, .12);
    }
    .subhead {
      grid-column: 1 / -1;
      margin-top: 4px;
      padding-top: 10px;
      border-top: 1px solid var(--line);
      color: #16191f;
      font-size: 14px;
      font-weight: 900;
    }
    .scheduling-explainer {
      grid-column: 1 / -1;
      display: grid;
      gap: 8px;
      padding: 12px 14px;
      border: 1px solid #eaeced;
      border-radius: 8px;
      color: var(--muted);
      background: #f7f8f8;
      font-size: 12px;
      line-height: 1.5;
      white-space: pre-line;
    }
    .scheduling-explainer strong {
      color: var(--ink);
      font-size: 13px;
    }
    .scheduling-explainer.policy-ok {
      border-color: #d8e4f4;
      background: #eef4ff;
    }
    .scheduling-explainer.policy-warn {
      border-color: #ffdcb0;
      background: #fff8ed;
      color: #7c4a16;
    }
    .yaml-preview {
      padding: 0 14px 12px;
    }
    .yaml-preview .label {
      margin: 14px 0 8px;
      display: block;
    }
    button {
      border: 0;
      border-radius: 8px;
      padding: 9px 13px;
      color: #fff;
      background: var(--blue);
      font-weight: 900;
      letter-spacing: .01em;
      cursor: pointer;
      transition: transform .16s ease, box-shadow .16s ease, background .16s ease;
    }
    button:hover {
      transform: translateY(-1px);
      box-shadow: 0 8px 18px rgba(0, 0, 0, .24);
    }
    button.secondary {
      color: #0b5cad;
      background: #eaf3ff;
      border: 1px solid #cfe1ff;
    }
    button.ghost {
      color: var(--ink);
      background: #f7f8f8;
      border: 1px solid var(--line);
    }
    textarea {
      width: 100%;
      min-height: 200px;
      display: block;
      padding: 16px;
      border: 1px solid var(--line);
      resize: vertical;
      outline: none;
      color: #16191f;
      background: #f8fafc;
      font-family: "SFMono-Regular", "Cascadia Code", "PingFang SC", monospace;
      font-size: 12px;
      line-height: 1.5;
    }
    .results {
      padding: 14px;
    }
    .result-grid {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 10px;
      margin-bottom: 12px;
    }
    .monitor-kpi-grid {
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }
    .monitor-chart-grid {
      display: grid;
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 10px;
      margin-bottom: 14px;
    }
    .chart-card {
      display: grid;
      min-height: 232px;
      border: 1px solid var(--line);
      border-radius: 8px;
      overflow: hidden;
      background: var(--panel);
      box-shadow: var(--shadow);
    }
    .chart-head {
      display: flex;
      justify-content: space-between;
      gap: 10px;
      align-items: center;
      padding: 11px 14px;
      border-bottom: 1px solid var(--line);
      background: #fbfcfe;
      font-size: 13px;
      font-weight: 900;
    }
    .chart-note {
      color: var(--muted);
      font-size: 12px;
      font-weight: 700;
      white-space: nowrap;
    }
    .chart-body {
      display: grid;
      gap: 10px;
      min-height: 178px;
      padding: 12px 14px 14px;
    }
    .chart-legend {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
      color: var(--muted);
      font-size: 12px;
    }
    .legend-item {
      display: inline-flex;
      align-items: center;
      gap: 6px;
    }
    .legend-swatch {
      width: 10px;
      height: 10px;
      border-radius: 2px;
      flex: 0 0 auto;
    }
    .chart-surface {
      position: relative;
      min-height: 146px;
    }
    .chart-empty {
      display: grid;
      place-items: center;
      min-height: 146px;
      padding: 16px;
      color: var(--muted);
      text-align: center;
      border: 1px dashed var(--line);
      border-radius: 8px;
      background: #f8fafc;
    }
    .trend-svg {
      width: 100%;
      height: 146px;
      display: block;
    }
    .donut-wrap {
      display: grid;
      grid-template-columns: 112px minmax(0, 1fr);
      gap: 14px;
      align-items: center;
    }
    .donut-ring {
      position: relative;
      width: 112px;
      height: 112px;
      border-radius: 50%;
      flex: 0 0 auto;
    }
    .donut-ring:before {
      content: "";
      position: absolute;
      inset: 26px;
      border-radius: 50%;
      background: var(--panel);
      box-shadow: inset 0 0 0 1px var(--line);
    }
    .donut-center {
      position: absolute;
      inset: 0;
      display: grid;
      place-items: center;
      text-align: center;
      font-weight: 900;
      pointer-events: none;
    }
    .donut-center strong {
      display: block;
      font-size: 22px;
      line-height: 1;
    }
    .donut-center span {
      display: block;
      margin-top: 4px;
      color: var(--muted);
      font-size: 11px;
      line-height: 1;
      font-weight: 700;
    }
    .donut-list {
      display: grid;
      gap: 8px;
    }
    .donut-row {
      display: grid;
      grid-template-columns: 1fr auto;
      gap: 10px;
      align-items: center;
    }
    .donut-row strong {
      font-size: 12px;
    }
    .donut-row span {
      color: var(--muted);
      font-size: 12px;
      font-weight: 700;
    }
    .bar-list {
      display: grid;
      gap: 10px;
    }
    .bar-row {
      display: grid;
      gap: 6px;
    }
    .bar-meta {
      display: flex;
      justify-content: space-between;
      gap: 12px;
      color: var(--muted);
      font-size: 12px;
    }
    .bar-track {
      height: 10px;
      overflow: hidden;
      border-radius: 999px;
      background: #f7f8f8;
      box-shadow: inset 0 0 0 1px rgba(22, 25, 31, .08);
    }
    .bar-fill {
      height: 100%;
      border-radius: 999px;
      background: linear-gradient(90deg, #1677ff, #69a7ff);
    }
    .task-list {
      display: grid;
      gap: 8px;
      margin: 12px 0;
    }
    .task-list.flush { margin: 0; }
    .task-row {
      width: 100%;
      display: grid;
      grid-template-columns: 1.15fr .7fr .7fr .7fr auto;
      gap: 10px;
      align-items: center;
      padding: 10px 12px;
      border: 1px solid var(--line);
      border-radius: 8px;
      color: var(--ink);
      background: var(--panel);
      text-align: left;
      cursor: pointer;
    }
    .task-row:hover {
      transform: translateY(-1px);
      box-shadow: 0 8px 18px rgba(0, 0, 0, .24);
    }
    .task-row strong, .task-row span {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .resource-grid {
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }
    .tenant-list {
      display: grid;
      gap: 10px;
    }
    .tenant-row {
      width: 100%;
      display: grid;
      grid-template-columns: 1.15fr .75fr .7fr .7fr auto;
      gap: 10px;
      align-items: center;
      padding: 10px 12px;
      border: 1px solid var(--line);
      border-radius: 8px;
      color: var(--ink);
      background: var(--panel);
      text-align: left;
    }
    .tenant-row strong, .tenant-row span {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .pill {
      display: inline-flex;
      justify-content: center;
      border-radius: 999px;
      padding: 5px 9px;
      color: #0b5cad;
      background: #eef4ff;
      font-size: 12px;
      font-weight: 900;
    }
    .intent-card {
      padding: 13px;
      border-radius: 8px;
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
      color: #0b5cad;
      background: #eef4ff;
    }
    .warn {
      color: #7c4a16;
      background: #fff8ed;
    }
    pre {
      min-height: 300px;
      margin: 0;
      padding: 14px;
      overflow: auto;
      border-radius: 8px;
      color: #16191f;
      background: #f8fafc;
      font-family: "SFMono-Regular", "Cascadia Code", "PingFang SC", monospace;
      font-size: 12px;
      line-height: 1.5;
      box-shadow: inset 0 0 0 1px rgba(22,25,31,.08);
    }
    .empty {
      color: var(--muted);
      min-height: 220px;
      display: grid;
      place-items: center;
      padding: 26px;
      text-align: center;
      border: 1px dashed rgba(22,25,31,.16);
      border-radius: 8px;
      background: #fff;
    }
    .log-toolbar {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
      align-items: center;
      margin: 12px 0 8px;
    }
    .log-toolbar select, .log-toolbar input {
      width: auto;
      min-width: 160px;
    }
    .log-box {
      min-height: 200px;
      max-height: 380px;
      white-space: pre-wrap;
      color: #16191f;
      background: #f8fafc;
    }
    .hint {
      color: var(--muted);
      font-size: 12px;
      line-height: 1.5;
    }
    .danger {
      color: #fff;
      background: #d13212;
    }
    .audit-list {
      display: grid;
      gap: 8px;
      margin: 8px 0 0;
    }
    .audit-row {
      padding: 10px 12px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: #fff;
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
      main { width: min(100vw - 14px, 760px); padding-top: 12px; }
      main { grid-template-columns: 1fr; }
      .sidebar { position: static; }
      .hero { grid-template-columns: 1fr; }
      .status-strip { grid-template-columns: 1fr; }
      .template-strip { grid-template-columns: 1fr; }
      .workflow-summary { grid-template-columns: 1fr; }
      .wizard-steps { grid-template-columns: 1fr 1fr; }
      .form-grid, .filter-grid { grid-template-columns: 1fr; }
      .panel { min-height: auto; border-radius: 8px; }
      textarea { min-height: 360px; }
      .result-grid { grid-template-columns: 1fr; }
      .monitor-kpi-grid,
      .monitor-chart-grid { grid-template-columns: 1fr; }
      .donut-wrap { grid-template-columns: 1fr; justify-items: center; }
      .task-row { grid-template-columns: 1fr 1fr; }
      .tenant-row { grid-template-columns: 1fr 1fr; }
    }
  </style>
</head>
  <body data-role="__PAGE_ROLE__" data-account-tenant="__ACCOUNT_TENANT__" data-account-namespace="__ACCOUNT_NAMESPACE__">
  <main>
    <aside class="sidebar">
      <div class="sidebar-brand">
        <div class="label">__PAGE_LABEL__</div>
        <div class="sidebar-title">__PAGE_TITLE__</div>
        <div class="sidebar-lede">__PAGE_LEDE__</div>
        <span class="role-badge">__PAGE_ROLE_LABEL__</span>
      </div>
      <div class="sidebar-section user-nav">
        <div class="sidebar-section-title">使用方</div>
        <div class="sidebar-nav">
          <button class="nav-item active" id="nav-training" type="button" data-section="training">训练向导</button>
          <button class="nav-item" id="nav-platform" type="button" data-section="platform">平台视图</button>
        </div>
      </div>
      <div class="sidebar-section monitor-nav">
        <div class="sidebar-section-title">K8s 管理员</div>
        <div class="sidebar-nav">
          <button class="nav-item active" id="nav-monitor-overview" type="button" data-section="overview">全局概览</button>
          <button class="nav-item" id="nav-monitor-tasks" type="button" data-section="tasks">任务与审计</button>
          <button class="nav-item" id="nav-monitor-audits" type="button" data-section="audits">操作审计</button>
        </div>
      </div>
      <div class="sidebar-actions">
        <form action="/logout" method="post">
          <button class="ghost" type="submit">退出登录</button>
        </form>
      </div>
    </aside>
    <section class="content">
      <section class="hero">
        <div>
          <div class="label">__PAGE_LABEL__</div>
          <h1>__PAGE_TITLE__</h1>
          <p class="lede">__PAGE_LEDE__</p>
        </div>
        <div class="status-strip">
          <div class="status-card">
            <div class="label">接口健康</div>
            <div class="metric" id="health">检查中</div>
          </div>
          <div class="status-card">
            <div class="label">当前账号范围</div>
            <div class="metric" id="account-scope">未限定</div>
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
      <section class="panel section-pane active" data-section="training">
        <div class="panel-head">
          <h2 class="panel-title">训练向导</h2>
          <div class="actions">
            <button class="ghost" id="load-service">载入服务画像</button>
            <button class="ghost" id="load-job">载入训练模板</button>
            <button class="ghost" id="generate">生成 YAML</button>
            <button class="secondary" id="validate">校验</button>
            <button id="normalize">归一化</button>
            <button id="deploy">提交部署</button>
          </div>
        </div>
        <div class="template-strip" aria-label="算法任务模板">
          <button class="template-card active" id="template-sft" type="button">
            <strong>训练</strong>
            <span>基础模型、训练数据、资源和产出集中在同一条链路里。</span>
          </button>
          <button class="template-card" id="template-eval" type="button">
            <strong>评测</strong>
            <span>评测模型、评测数据集和阈值单独配置，结果再同步回模型资产。</span>
          </button>
        </div>
        <div class="wizard-steps">
          <div class="wizard-step active">1 选择模式</div>
          <div class="wizard-step">2 配置模型</div>
          <div class="wizard-step">3 配置数据</div>
          <div class="wizard-step">4 资源与提交</div>
        </div>
        <div class="workflow-summary">
          <div class="workflow-chip">
            <strong>训练路径</strong>
            <span>模型 → 训练数据 → 资源与隔离 → 产出登记</span>
          </div>
          <div class="workflow-chip">
            <strong>评测路径</strong>
            <span>模型 → 评测数据集 → 阈值 → 结果同步</span>
          </div>
        </div>
        <div class="form-grid">
          <div class="subhead">通用入口</div>
          <div class="field">
            <label for="taskDisplayName">任务名称</label>
            <input id="taskDisplayName" value="customer-sft-demo">
          </div>
          <div class="field wide shared-model-field">
            <label for="baseModelSelect" id="baseModelLabel">基础模型</label>
            <select id="baseModelSelect">
              <option value="model://platform/qwen2.5-0.5b-demo">Qwen2.5-0.5B 平台示例模型</option>
            </select>
            <div class="hint" id="workflow-model-hint">训练模式会把这个模型作为微调起点，评测模式会把它作为待评测模型。</div>
          </div>
          <div class="field wide internal-field">
            <label for="algorithmTemplate">算法类型</label>
            <select id="algorithmTemplate">
              <option value="sft">SFT 微调</option>
              <option value="evaluation">模型评测</option>
            </select>
          </div>
          <div class="field wide train-field internal-field">
            <label for="baseModelURI">内部基础模型引用</label>
            <input id="baseModelURI" value="modelscope://qwen/Qwen2.5-0.5B">
          </div>
          <details class="fold-section train-band" id="training-band" open>
            <summary>训练配置</summary>
            <div class="fold-body">
              <div class="form-grid">
                <div class="field wide train-field">
                  <label for="trainingDatasetSelect">训练数据集</label>
                  <select id="trainingDatasetSelect">
                    <option value="inline://datasets/customer-sft-demo">客服问答 SFT 数据集（示例）</option>
                  </select>
                </div>
                <div class="field wide train-field internal-field">
                  <label for="trainingDataURI">内部数据集引用</label>
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
                    <option value="true">训练完成后登记模型产物</option>
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
              </div>
            </div>
          </details>
          <details class="fold-section evaluation-band" id="evaluation-band">
            <summary>评测配置</summary>
            <div class="fold-body">
              <div class="form-grid">
                <div class="field wide evaluation-field">
                  <label for="evaluationDatasetSelect">评测数据集</label>
                  <select id="evaluationDatasetSelect" title="评测数据集">
                    <option value="inline://datasets/customer-eval">默认评测数据集</option>
                  </select>
                </div>
                <div class="field wide evaluation-field internal-field">
                  <label for="evaluation-dataset-uri">内部评测数据引用</label>
                  <input class="internal-field" id="evaluation-dataset-uri" value="inline://datasets/customer-eval" title="内部评测数据引用">
                </div>
                <div class="field evaluation-field">
                  <label for="evaluation-threshold">通过阈值</label>
                  <input id="evaluation-threshold" value="0.8" title="通过阈值">
                </div>
                <div class="scheduling-explainer evaluation-field">
                  <strong>评测流程说明</strong>
                  评测会复用当前模型和资源策略，输出结果后可同步到模型资产库。
                </div>
              </div>
            </div>
          </details>
          <div class="field service-template-field wide">
            <label for="serviceModelSelect">待发布模型</label>
            <select id="serviceModelSelect">
              <option value="inline://models/customer-sft-demo/v1">customer-sft-demo 示例模型</option>
            </select>
          </div>
          <div class="field service-template-field wide internal-field">
            <label for="serviceModelURI">内部模型引用</label>
            <input id="serviceModelURI" value="inline://models/customer-sft-demo/v1">
          </div>
          <details class="fold-section" open>
            <summary>资源与调度（以 K8s 调度为准）</summary>
            <div class="fold-body">
              <div class="form-grid">
                <div class="field wide">
                  <label for="schedulingPolicyTemplate">调度策略模板</label>
                  <select id="schedulingPolicyTemplate">
                    <option value="performance">性能优先</option>
                    <option value="cost-efficient">成本优先</option>
                    <option value="fast-start">快速启动</option>
                    <option value="gpu-dedicated">GPU 专用</option>
                  </select>
                </div>
                <div class="scheduling-explainer" id="schedulingPolicyExplanation">
                  <strong>当前调度映射（以 K8s 调度为准）</strong>
                  策略说明：性能优先模板会优先选择 GPU 节点并提高任务优先级，但最终由 Kubernetes 调度决定。
                  Kubernetes 调度：source of truth
                  集群能力边界：CPU Manager static / CPU pinning、GPU device plugin、Topology Manager
                  resources.requests/limits：CPU 8 / 内存 32Gi / GPU 1
                  priorityClassName：ai-high
                  nodeSelector：ai.oam.dev/node-type=gpu
                  ai-runtime.schedulingStrategy：performance
                </div>
                <div class="field">
                  <label for="resourceProfile">资源规格</label>
                  <select id="resourceProfile">
                    <option value="medium-gpu">中规格 单 GPU</option>
                    <option value="small-cpu">小规格 CPU</option>
                    <option value="large-gpu">大规格 多 GPU</option>
                    <option value="custom">自定义</option>
                  </select>
                </div>
                <div class="field">
                  <label for="resourceCPU">CPU 核数</label>
                  <input id="resourceCPU" value="8">
                </div>
                <div class="field">
                  <label for="resourceMemory">内存</label>
                  <input id="resourceMemory" value="32Gi">
                </div>
                <div class="field">
                  <label for="resourceGPU">GPU 数量</label>
                  <input id="resourceGPU" value="1">
                </div>
                <div class="field">
                  <label for="schedulingPriority">调度优先级</label>
                  <select id="schedulingPriority">
                    <option value="high">高优先级</option>
                    <option value="normal">普通优先级</option>
                    <option value="urgent">紧急优先级</option>
                  </select>
                </div>
                <div class="field">
                  <label for="schedulingNodeType">节点类型</label>
                  <select id="schedulingNodeType">
                    <option value="gpu">GPU 节点</option>
                    <option value="cpu">CPU 节点</option>
                    <option value="any">不限节点</option>
                  </select>
                </div>
                <div class="field">
                  <label for="schedulingStrategy">调度策略</label>
                  <select id="schedulingStrategy">
                    <option value="performance">性能优先</option>
                    <option value="cost">成本优先</option>
                    <option value="fast-start">快速启动</option>
                  </select>
                </div>
              </div>
            </div>
          </details>
          <details class="fold-section">
            <summary>租户资源隔离</summary>
            <div class="fold-body">
              <div class="form-grid">
                <div class="field">
                  <label for="tenantNamespace">租户命名空间</label>
                  <input id="tenantNamespace" value="sock-shop">
                </div>
                <div class="field">
                  <label for="quotaCPU">CPU 配额</label>
                  <input id="quotaCPU" value="32">
                </div>
                <div class="field">
                  <label for="quotaMemory">内存配额</label>
                  <input id="quotaMemory" value="128Gi">
                </div>
                <div class="field">
                  <label for="quotaGPU">GPU 配额</label>
                  <input id="quotaGPU" value="4">
                </div>
                <div class="field">
                  <label for="maxTaskCPU">单任务最大 CPU</label>
                  <input id="maxTaskCPU" value="8">
                </div>
                <div class="field">
                  <label for="maxTaskMemory">单任务最大内存</label>
                  <input id="maxTaskMemory" value="32Gi">
                </div>
                <div class="field">
                  <label for="maxTaskGPU">单任务最大 GPU</label>
                  <input id="maxTaskGPU" value="1">
                </div>
                <div class="scheduling-explainer policy-ok" id="isolationPolicyPreview">
                  <strong>隔离策略预览</strong>
                  Namespace：sock-shop
                  ResourceQuota：CPU 32 / 内存 128Gi / GPU 4
                  LimitRange：单任务 CPU 8 / 内存 32Gi / GPU 1
                  配额校验：当前申请资源在租户配额和单任务上限内。
                </div>
              </div>
            </div>
          </details>
        </div>
        <details class="advanced-config fold-section">
          <summary>高级配置：查看和调整底层 AIService / AIJob 字段</summary>
          <div class="fold-body">
          <div class="form-grid">
          <div class="field">
            <label for="kind">任务类型（AIService / AIJob）</label>
            <select id="kind">
              <option value="AIService">AIService</option>
              <option value="AIJob">AIJob</option>
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

      <section class="panel section-pane" data-section="platform">
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
          <details class="fold-section" open>
            <summary>应用交付</summary>
            <div class="fold-body">
              <div class="log-toolbar">
                <input id="delivery-service-name" placeholder="服务名，默认 job-service">
                <input id="delivery-service-image" value="python:3.11-slim" title="服务镜像">
                <button class="ghost" id="parse-delivery-result">解析训练结果</button>
                <button class="secondary" id="publish-delivery-service">发布为服务</button>
                <button class="ghost" id="probe-delivery-service">测试服务访问</button>
                <button class="ghost" id="refresh-artifacts">查看模型产物</button>
              </div>
              <pre id="delivery-output" class="log-box">选择 AIJob 后，可解析 AI_RESULT_JSON 并发布为 AIService。发布成功后可测试服务 /healthz。</pre>
            </div>
          </details>
          <details class="fold-section">
            <summary>我的模型 · 模型资产库</summary>
            <div class="fold-body">
              <div class="log-toolbar">
                <button class="ghost" id="refresh-models">刷新我的模型</button>
                <button class="secondary" id="register-model">登记模型</button>
              </div>
              <div id="model-list" class="task-list">
                <div class="empty">训练完成后可登记模型资产；已登记模型可以“设为基础模型”继续训练，或“发布此模型”生成服务模板。</div>
              </div>
            </div>
          </details>
          <details class="fold-section">
            <summary>数据资产库</summary>
            <div class="fold-body">
              <div class="log-toolbar">
                <input id="dataset-name" value="customer-sft-demo" title="数据集名称">
                <input class="internal-field" id="dataset-uri" value="dataset://ai-demo/customer-sft-demo/v1" title="平台内部数据集引用">
                <select id="dataset-format" title="数据格式">
                  <option value="sharegpt-jsonl">ShareGPT JSONL</option>
                  <option value="alpaca-jsonl">Alpaca JSONL</option>
                  <option value="custom-jsonl">自定义 JSONL</option>
                </select>
                <button class="ghost" id="refresh-datasets">刷新数据集</button>
                <button class="secondary" id="register-dataset">登记数据集</button>
              </div>
              <div id="dataset-list" class="task-list">
                <div class="empty">用户上传或登记数据集后，可在训练向导中“选择数据集”，平台内部再映射为运行时数据引用。</div>
              </div>
            </div>
          </details>
          <details class="fold-section">
            <summary>生命周期操作</summary>
            <div class="fold-body">
              <div class="log-toolbar">
                <button class="ghost" id="rerun-task">重新运行 Job</button>
                <button class="ghost" id="restart-task">重启服务</button>
                <button class="danger" id="delete-task">删除任务</button>
              </div>
            </div>
          </details>
          <details class="fold-section">
            <summary>运行日志</summary>
            <div class="fold-body">
              <div class="log-toolbar">
                <select id="log-pod"><option value="">自动选择 Pod</option></select>
                <select id="log-container"><option value="">自动选择容器</option></select>
                <input id="log-tail" value="200" title="日志行数">
                <button class="ghost" id="refresh-logs">刷新日志</button>
              </div>
              <pre id="logs-output" class="log-box">选择任务后，可查看最近日志。AIJob 用于查看训练输出，AIService 用于查看启动和请求日志。</pre>
            </div>
          </details>
        </div>
      </section>
    </section>
    </section>

    <section class="view" id="monitor-view">
      <section class="monitor-grid">
        <section class="panel section-pane active" data-section="overview">
          <div class="panel-head">
            <h2 class="panel-title">全局任务概览</h2>
            <button id="monitor-refresh">刷新监测</button>
          </div>
          <div class="results">
            <div class="label">核心指标</div>
            <div class="result-grid monitor-kpi-grid" id="monitor-kpi-grid">
              <div class="intent-card ok"><div class="label">AIService 数量</div><strong>0</strong></div>
              <div class="intent-card ok"><div class="label">AIJob 数量</div><strong>0</strong></div>
              <div class="intent-card"><div class="label">Running</div><strong>0</strong></div>
              <div class="intent-card ok"><div class="label">健康率</div><strong>0%</strong></div>
              <div class="intent-card warn"><div class="label">异常任务</div><strong>0</strong></div>
              <div class="intent-card"><div class="label">活跃租户</div><strong>0</strong></div>
            </div>
            <div class="monitor-chart-grid">
              <section class="chart-card">
                <div class="chart-head">
                  <span>最近刷新趋势</span>
                  <span class="chart-note" id="monitor-trend-note">最近 1 次刷新</span>
                </div>
                <div class="chart-body" id="monitor-trend-chart"></div>
              </section>
              <section class="chart-card">
                <div class="chart-head">
                  <span>任务健康分布</span>
                  <span class="chart-note" id="monitor-status-note">按 phase 汇总</span>
                </div>
                <div class="chart-body" id="monitor-status-chart"></div>
              </section>
              <section class="chart-card">
                <div class="chart-head">
                  <span>租户任务 Top5</span>
                  <span class="chart-note" id="monitor-tenant-note">按任务数</span>
                </div>
                <div class="chart-body" id="monitor-tenant-chart"></div>
              </section>
            </div>
          </div>
          <div class="subhead" style="margin:0 14px;">算力资源总览</div>
          <div class="results">
            <div class="result-grid resource-grid" id="monitor-resource-cards">
              <div class="intent-card ok"><div class="label">CPU 申请总量</div><strong>0m</strong></div>
              <div class="intent-card ok"><div class="label">内存 申请总量</div><strong>0Mi</strong></div>
              <div class="intent-card ok"><div class="label">GPU 申请总量</div><strong>0 GPU</strong></div>
            </div>
          </div>
          <div class="subhead" style="margin:0 14px;">租户资源视图</div>
          <div class="results">
            <div class="label">按租户汇总资源申请</div>
            <div id="monitor-tenant-list" class="tenant-list">
              <div class="empty">点击“刷新监测”，按租户查看资源申请。</div>
            </div>
          </div>
          <details class="fold-section">
            <summary>筛选条件</summary>
            <div class="fold-body">
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
            </div>
          </details>
        </section>

        <section class="panel section-pane" data-section="tasks">
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
          </div>
        </section>

        <section class="panel section-pane" data-section="audits">
          <div class="panel-head">
            <h2 class="panel-title">操作审计</h2>
            <div class="actions">
              <button class="ghost" id="refresh-audits">刷新审计</button>
            </div>
          </div>
          <div class="results">
            <div class="label">最近操作</div>
            <div id="audit-list" class="audit-list">
              <div class="empty">点击“刷新审计”，查看最近生命周期操作。</div>
            </div>
          </div>
        </section>
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
    var baseModelSelect = document.getElementById("baseModelSelect");
    var baseModelLabel = document.getElementById("baseModelLabel");
    var workflowModelHint = document.getElementById("workflow-model-hint");
    var serviceModelSelect = document.getElementById("serviceModelSelect");
    var trainingBand = document.getElementById("training-band");
    var evaluationBand = document.getElementById("evaluation-band");
    var evaluationDatasetSelect = document.getElementById("evaluationDatasetSelect");
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
    var monitorCards = document.getElementById("monitor-kpi-grid");
    var monitorResourceCards = document.getElementById("monitor-resource-cards");
    var monitorTenantList = document.getElementById("monitor-tenant-list");
    var monitorTaskList = document.getElementById("monitor-task-list");
    var monitorAlertList = document.getElementById("monitor-alert-list");
    var monitorOutput = document.getElementById("monitor-output");
    var monitorAction = document.getElementById("monitor-action");
    var monitorTrendChart = document.getElementById("monitor-trend-chart");
    var monitorStatusChart = document.getElementById("monitor-status-chart");
    var monitorTenantChart = document.getElementById("monitor-tenant-chart");
    var monitorTrendNote = document.getElementById("monitor-trend-note");
    var monitorStatusNote = document.getElementById("monitor-status-note");
    var monitorTenantNote = document.getElementById("monitor-tenant-note");
    var auditList = document.getElementById("audit-list");
    var schedulingPolicyExplanation = document.getElementById("schedulingPolicyExplanation");
    var isolationPolicyPreview = document.getElementById("isolationPolicyPreview");
    var monitorHistory = [];
    var monitorFields = {};
    ["monitorNamespace", "monitorType", "monitorTenant", "monitorProject", "monitorEnvironment", "monitorHealth"].forEach(function(id) {
      monitorFields[id] = document.getElementById(id);
    });
    var wizardFields = {};
    ["algorithmTemplate", "taskDisplayName", "baseModelURI", "trainingDataURI", "trainingSize", "publishAfterTrain", "epochs", "learningRate", "serviceModelURI", "schedulingPolicyTemplate", "resourceProfile", "resourceCPU", "resourceMemory", "resourceGPU", "schedulingPriority", "schedulingNodeType", "schedulingStrategy", "tenantNamespace", "quotaCPU", "quotaMemory", "quotaGPU", "maxTaskCPU", "maxTaskMemory", "maxTaskGPU"].forEach(function(id) {
      wizardFields[id] = document.getElementById(id);
    });
    var fields = {};
    ["kind", "name", "namespace", "component", "tenant", "project", "environment", "owner", "runtime", "image", "modelName", "modelVersion", "modelURI", "replicas", "port", "jobKind", "ttl", "datasetURI", "outputURI"].forEach(function(id) {
      fields[id] = document.getElementById(id);
    });
    var accountTenant = (document.body.getAttribute("data-account-tenant") || "").trim();
    var accountNamespace = (document.body.getAttribute("data-account-namespace") || "").trim();
    var demoMode = /(?:\?|&)demo=local(?:&|$)/.test(window.location.search);
    var demoState = {
      applications: [
        {
          namespace: "ai-tenant-a",
          name: "tenant-a-training-job",
          workloadTypes: ["job"],
          phase: "succeeded",
          healthy: true,
          aiMetadata: {
            "ai.oam.dev/tenant": "tenant-a",
            "ai.oam.dev/project": "tenant-a-training",
            "ai.oam.dev/environment": "test",
            "ai.oam.dev/owner": "tenant-a-user"
          },
          resourceSummary: {cpuMilli: 8000, memoryMi: 32768, gpu: 1},
          components: [{name: "tenant-a-training-job-trainer"}],
          pods: [{name: "tenant-a-training-job-pod-0", phase: "Succeeded", containers: ["trainer"]}],
          pod: "tenant-a-training-job-pod-0",
          container: "trainer",
          logs: "train-start\nepoch=1 loss=0.30\nepoch=2 loss=0.12\nAI_RESULT_JSON={\"modelURI\":\"inline://models/tenant-a-training-job/v1\",\"metrics\":{\"loss\":0.12,\"accuracy\":0.98},\"summary\":\"tenant-a-training-complete\"}\ntrain-complete",
          deliveryResult: {
            modelURI: "inline://models/tenant-a-training-job/v1",
            metrics: {loss: 0.12, accuracy: 0.98},
            summary: "tenant-a-training-complete"
          }
        },
        {
          namespace: "ai-tenant-a",
          name: "tenant-a-inference-service",
          workloadTypes: ["service"],
          phase: "running",
          healthy: true,
          aiMetadata: {
            "ai.oam.dev/tenant": "tenant-a",
            "ai.oam.dev/project": "tenant-a-inference",
            "ai.oam.dev/environment": "test",
            "ai.oam.dev/owner": "tenant-a-user"
          },
          resourceSummary: {cpuMilli: 2000, memoryMi: 4096, gpu: 0},
          components: [{name: "tenant-a-inference-service"}],
          pods: [{name: "tenant-a-inference-service-pod-0", phase: "Running", containers: ["http"]}],
          pod: "tenant-a-inference-service-pod-0",
          container: "http",
          logs: "service booted\nmodelURI=inline://models/tenant-a-training-job/v1\nhealthz ok",
          modelURI: "inline://models/tenant-a-training-job/v1"
        },
        {
          namespace: "ai-tenant-b",
          name: "tenant-b-evaluation-job",
          workloadTypes: ["job"],
          phase: "succeeded",
          healthy: true,
          aiMetadata: {
            "ai.oam.dev/tenant": "tenant-b",
            "ai.oam.dev/project": "tenant-b-evaluation",
            "ai.oam.dev/environment": "test",
            "ai.oam.dev/owner": "tenant-b-user"
          },
          resourceSummary: {cpuMilli: 4000, memoryMi: 8192, gpu: 0},
          components: [{name: "tenant-b-evaluation-job-runner"}],
          pods: [{name: "tenant-b-evaluation-job-pod-0", phase: "Succeeded", containers: ["eval"]}],
          pod: "tenant-b-evaluation-job-pod-0",
          container: "eval",
          logs: "eval-start\nAI_RESULT_JSON={\"modelURI\":\"inline://models/tenant-b-evaluation-job/v1\",\"metrics\":{\"accuracy\":0.94},\"summary\":\"tenant-b-evaluation-passed\"}\neval-complete",
          deliveryResult: {
            modelURI: "inline://models/tenant-b-evaluation-job/v1",
            metrics: {accuracy: 0.94},
            summary: "tenant-b-evaluation-passed"
          }
        },
        {
          namespace: "ai-tenant-b",
          name: "tenant-b-chat-service",
          workloadTypes: ["service"],
          phase: "running",
          healthy: true,
          aiMetadata: {
            "ai.oam.dev/tenant": "tenant-b",
            "ai.oam.dev/project": "tenant-b-serving",
            "ai.oam.dev/environment": "test",
            "ai.oam.dev/owner": "tenant-b-user"
          },
          resourceSummary: {cpuMilli: 4000, memoryMi: 8192, gpu: 1},
          components: [{name: "tenant-b-chat-service"}],
          pods: [{name: "tenant-b-chat-service-pod-0", phase: "Running", containers: ["http"]}],
          pod: "tenant-b-chat-service-pod-0",
          container: "http",
          logs: "service booted\nmodelURI=oss://models/tenant-b-chat/v1\nhealthz ok",
          modelURI: "oss://models/tenant-b-chat/v1"
        }
      ],
      models: [
        {
          namespace: "ai-tenant-a",
          name: "tenant-a-training-job",
          modelURI: "inline://models/tenant-a-training-job/v1",
          status: "trained",
          evaluationStatus: "passed",
          visibility: "private",
          metrics: {loss: 0.12, accuracy: 0.98},
          summary: "tenant-a-training-complete"
        },
        {
          namespace: "ai-tenant-b",
          name: "tenant-b-chat",
          modelURI: "oss://models/tenant-b-chat/v1",
          status: "released",
          evaluationStatus: "passed",
          visibility: "private",
          metrics: {accuracy: 0.94},
          summary: "tenant-b-chat-ready"
        }
      ],
      datasets: [
        {
          namespace: "ai-tenant-a",
          name: "tenant-a-training-data",
          displayName: "Tenant A 训练数据集",
          datasetURI: "inline://datasets/tenant-a-training-data",
          format: "sharegpt-jsonl",
          purpose: "sft",
          status: "validated",
          visibility: "private"
        },
        {
          namespace: "ai-tenant-b",
          name: "tenant-b-evaluation-data",
          displayName: "Tenant B 评测数据集",
          datasetURI: "inline://datasets/tenant-b-evaluation-data",
          format: "alpaca-jsonl",
          purpose: "evaluation",
          status: "validated",
          visibility: "private"
        }
      ],
      artifacts: [
        {
          namespace: "ai-tenant-a",
          name: "tenant-a-training-job-v1",
          modelURI: "inline://models/tenant-a-training-job/v1",
          status: "ready",
          framework: "demo",
          source: "AIJob",
          owner: "tenant-a-user"
        },
        {
          namespace: "ai-tenant-b",
          name: "tenant-b-evaluation-job-v1",
          modelURI: "inline://models/tenant-b-evaluation-job/v1",
          status: "ready",
          framework: "demo",
          source: "AIJob",
          owner: "tenant-b-user"
        }
      ],
      audits: [
        {
          time: "2026-09-10 09:00:00",
          action: "deploy",
          namespace: "ai-tenant-a",
          name: "tenant-a-training-job",
          actor: "admin",
          success: true,
          message: "Tenant A AIJob 已提交"
        },
        {
          time: "2026-09-10 09:12:00",
          action: "publish-service",
          namespace: "ai-tenant-a",
          name: "tenant-a-inference-service",
          actor: "admin",
          success: true,
          message: "Tenant A AIService 已发布"
        },
        {
          time: "2026-09-10 10:00:00",
          action: "deploy",
          namespace: "ai-tenant-b",
          name: "tenant-b-evaluation-job",
          actor: "tenant-b",
          success: true,
          message: "Tenant B AIJob 已提交"
        },
        {
          time: "2026-09-10 10:12:00",
          action: "publish-service",
          namespace: "ai-tenant-b",
          name: "tenant-b-chat-service",
          actor: "tenant-b",
          success: true,
          message: "Tenant B AIService 已发布"
        }
      ]
    };
    function demoError(message) {
      return {__demoError: message};
    }
    function demoNamespaceError(namespace) {
      var requested = (namespace || "").trim();
      if (!accountNamespace) {
        return "";
      }
      if (!requested) {
        return "namespace is required for account scope \"" + accountNamespace + "\"";
      }
      if (requested !== accountNamespace) {
        return "namespace \"" + requested + "\" is outside account scope \"" + accountNamespace + "\"";
      }
      return "";
    }
    function applyAccountScopeDefaults() {
      if (accountNamespace) {
        setWizardValue("tenantNamespace", accountNamespace);
        setValue("namespace", accountNamespace);
      }
      if (accountTenant) {
        setValue("tenant", accountTenant);
      }
      if (accountNamespace && monitorFields.monitorNamespace) {
        monitorFields.monitorNamespace.value = accountNamespace;
      }
      var accountScope = document.getElementById("account-scope");
      if (accountScope) {
        accountScope.textContent = accountNamespace
          ? (accountTenant ? accountTenant + " / " : "") + accountNamespace
          : "未限定";
      }
    }
    function demoClone(value) {
      return JSON.parse(JSON.stringify(value));
    }
    function demoBody(init) {
      if (!init || !init.body) {
        return {};
      }
      if (typeof init.body !== "string") {
        return init.body;
      }
      try {
        return JSON.parse(init.body);
      } catch (err) {
        return {};
      }
    }
    function demoFindApplication(namespace, name) {
      return (demoState.applications || []).filter(function(item) {
        return item.namespace === namespace && item.name === name;
      })[0] || null;
    }
    function demoUpsertApplication(app) {
      var key = app.namespace + "/" + app.name;
      demoState.applications = (demoState.applications || []).filter(function(item) {
        return item.namespace + "/" + item.name !== key;
      });
      demoState.applications.push(demoClone(app));
    }
    function demoRemoveApplication(namespace, name) {
      demoState.applications = (demoState.applications || []).filter(function(item) {
        return item.namespace !== namespace || item.name !== name;
      });
    }
    function demoAppendAudit(action, namespace, name, message, success) {
      demoState.audits = demoState.audits || [];
      demoState.audits.unshift({
        time: "2026-08-24 " + String(new Date().getHours()).padStart(2, "0") + ":" + String(new Date().getMinutes()).padStart(2, "0") + ":00",
        action: action,
        namespace: namespace,
        name: name,
        actor: "console",
        success: success !== false,
        message: message
      });
      if (demoState.audits.length > 20) {
        demoState.audits.length = 20;
      }
    }
    function demoList(namespace, items) {
      return (items || []).filter(function(item) {
        return !namespace || item.namespace === namespace;
      }).map(demoClone);
    }
    function demoCurrentResources() {
      var cpuMilli = Math.max(1000, Math.round(parseComputeQuantity(wizardFields.resourceCPU.value || "1") * 1000));
      var memoryRaw = (wizardFields.resourceMemory.value || "4Gi").trim().toLowerCase();
      var memoryValue = parseComputeQuantity(wizardFields.resourceMemory.value || "4Gi") || 4;
      var memoryMi = memoryRaw.indexOf("gi") !== -1 ? Math.round(memoryValue * 1024) : Math.round(memoryValue);
      var gpu = parseInt(wizardFields.resourceGPU.value || "0", 10);
      if (isNaN(gpu)) {
        gpu = 0;
      }
      return {cpuMilli: cpuMilli, memoryMi: memoryMi, gpu: gpu};
    }
    function demoNormalizedObject() {
      return {
        kind: fields.kind.value,
        workloadType: fields.kind.value === "AIService" ? "service" : "job",
        runtime: fields.runtime.value,
        image: fields.image.value,
        governanceIntent: {
          tenant: fields.tenant.value,
          project: fields.project.value,
          environment: fields.environment.value,
          owner: fields.owner.value,
          namespace: fields.namespace.value
        }
      };
    }
    function demoApplicationFromCurrentForm() {
      var isService = fields.kind.value === "AIService";
      var namespace = fields.namespace.value || accountNamespace || "demo";
      var name = fields.name.value || "demo-task";
      var component = fields.component.value || name;
      var container = isService ? "http" : "trainer";
      var app = {
        namespace: namespace,
        name: name,
        workloadTypes: [isService ? "service" : "job"],
        phase: isService ? "running" : "succeeded",
        healthy: true,
        aiMetadata: {
          "ai.oam.dev/tenant": fields.tenant.value || "demo-tenant",
          "ai.oam.dev/project": fields.project.value || "demo-project",
          "ai.oam.dev/environment": fields.environment.value || "poc",
          "ai.oam.dev/owner": fields.owner.value || "ai-platform"
        },
        resourceSummary: demoCurrentResources(),
        components: [{name: component}],
        pods: [{name: name + "-pod-0", phase: isService ? "Running" : "Succeeded", containers: [container]}],
        pod: name + "-pod-0",
        container: container,
        logs: isService ? "service booted\nhealthz ok" : "train-start\nepoch=1 loss=0.30\nepoch=2 loss=0.12\nAI_RESULT_JSON={\"modelURI\":\"inline://models/" + name + "/v1\",\"metrics\":{\"loss\":0.12,\"accuracy\":0.98},\"summary\":\"sft-fine-tuned\"}\ntrain-complete"
      };
      if (!isService) {
        app.deliveryResult = {
          modelURI: "inline://models/" + name + "/v1",
          metrics: {loss: 0.12, accuracy: 0.98},
          summary: "sft-fine-tuned"
        };
      }
      return app;
    }
    function demoResponse(path, init) {
      var url = new URL(path, window.location.origin);
      var pathname = url.pathname;
      var method = ((init && init.method) || "GET").toUpperCase();
      var requestedNamespace = url.searchParams.get("namespace") || "";
      var namespace = requestedNamespace || accountNamespace || "";
      var appMatch = pathname.match(/^\/api\/v1\/ai\/applications\/([^/]+)\/([^/]+)(?:\/(status|logs|probe|delete|restart|rerun))?$/);
      var modelMatch = pathname.match(/^\/api\/v1\/ai\/models\/([^/]+)\/([^/]+)(?:\/(evaluate|sync-evaluation))?$/);
      var deliveryMatch = pathname.match(/^\/api\/v1\/ai\/deliveries\/([^/]+)\/([^/]+)\/(result|publish-service)$/);
      var listNamespaceError = demoNamespaceError(requestedNamespace || accountNamespace);
      if (listNamespaceError && (pathname === "/api/v1/ai/applications" || pathname === "/api/v1/ai/artifacts" || pathname === "/api/v1/ai/models" || pathname === "/api/v1/ai/datasets" || pathname === "/api/v1/ai/audits")) {
        return demoError(listNamespaceError);
      }
      var pathNamespace = appMatch
        ? decodeURIComponent(appMatch[1])
        : modelMatch
          ? decodeURIComponent(modelMatch[1])
          : deliveryMatch
            ? decodeURIComponent(deliveryMatch[1])
            : "";
      if (pathNamespace) {
        var pathNamespaceError = demoNamespaceError(pathNamespace);
        if (pathNamespaceError) {
          return demoError(pathNamespaceError);
        }
      }
      if (method === "GET" && pathname === "/api/v1/ai/applications") {
        return {items: demoList(namespace, demoState.applications)};
      }
      if (method === "GET" && pathname === "/api/v1/ai/artifacts") {
        return {items: demoList(namespace, demoState.artifacts)};
      }
      if (method === "GET" && pathname === "/api/v1/ai/models") {
        return {items: demoList(namespace, demoState.models)};
      }
      if (method === "GET" && pathname === "/api/v1/ai/datasets") {
        return {items: demoList(namespace, demoState.datasets)};
      }
      if (method === "GET" && pathname === "/api/v1/ai/audits") {
        return {items: demoList(namespace, demoState.audits)};
      }
      if (method === "POST" && pathname === "/api/v1/ai/validate") {
        return {valid: true, demo: true};
      }
      if (method === "POST" && pathname === "/api/v1/ai/normalize") {
        return demoNormalizedObject();
      }
      if (method === "POST" && pathname === "/api/v1/ai/applications") {
        var app = demoApplicationFromCurrentForm();
        var appNamespaceError = demoNamespaceError(app.namespace);
        if (appNamespaceError) {
          return demoError(appNamespaceError);
        }
        var dryRun = url.searchParams.get("dryRun") === "true";
        if (!dryRun) {
          demoUpsertApplication(app);
          demoAppendAudit("deploy", app.namespace, app.name, "Demo deployment 已提交", true);
        }
        return {
          normalized: demoNormalizedObject(),
          application: {
            namespace: app.namespace,
            name: app.name,
            dryRun: dryRun
          }
        };
      }
      if (appMatch && method === "GET" && appMatch[3] === "status") {
        var appStatus = demoFindApplication(appMatch[1], appMatch[2]);
        return appStatus ? demoClone(appStatus) : {namespace: appMatch[1], name: appMatch[2], phase: "unknown", healthy: false, components: []};
      }
      if (appMatch && method === "GET" && appMatch[3] === "logs") {
        var appLogs = demoFindApplication(appMatch[1], appMatch[2]);
        if (!appLogs) {
          return {logs: "", pods: []};
        }
        return {
          logs: appLogs.logs || "",
          pods: demoClone(appLogs.pods || []),
          pod: appLogs.pod || "",
          container: appLogs.container || ""
        };
      }
      if (appMatch && method === "POST" && appMatch[3] === "probe") {
        var probed = demoFindApplication(appMatch[1], appMatch[2]);
        return {
          namespace: appMatch[1],
          name: appMatch[2],
          healthy: probed ? probed.healthy !== false : false,
          message: probed && probed.healthy !== false ? "demo probe ok" : "demo probe failed",
          path: "/healthz"
        };
      }
      if (appMatch && ((appMatch[3] === "delete" && (method === "DELETE" || method === "POST")) || (method === "DELETE" && !appMatch[3]))) {
        demoRemoveApplication(appMatch[1], appMatch[2]);
        demoAppendAudit("delete", appMatch[1], appMatch[2], "Demo 删除任务", true);
        return {message: "Demo 任务已删除"};
      }
      if (appMatch && appMatch[3] === "restart" && method === "POST") {
        var restarted = demoFindApplication(appMatch[1], appMatch[2]);
        if (restarted) {
          restarted.phase = "running";
          restarted.healthy = true;
          demoUpsertApplication(restarted);
        }
        demoAppendAudit("restart", appMatch[1], appMatch[2], "Demo 重启任务", true);
        return {message: "Demo 任务已重启"};
      }
      if (appMatch && appMatch[3] === "rerun" && method === "POST") {
        var rerun = demoFindApplication(appMatch[1], appMatch[2]);
        if (rerun) {
          rerun.phase = rerun.workloadTypes && rerun.workloadTypes.indexOf("job") !== -1 ? "running" : "running";
          rerun.healthy = true;
          demoUpsertApplication(rerun);
        }
        demoAppendAudit("rerun", appMatch[1], appMatch[2], "Demo 重新运行任务", true);
        return {message: "Demo 任务已重新运行"};
      }
      if (deliveryMatch && method === "GET" && deliveryMatch[3] === "result") {
        var deliveryApp = demoFindApplication(deliveryMatch[1], deliveryMatch[2]);
        if (deliveryApp && deliveryApp.deliveryResult) {
          return {
            namespace: deliveryMatch[1],
            jobName: deliveryMatch[2],
            result: demoClone(deliveryApp.deliveryResult)
          };
        }
        return {
          namespace: deliveryMatch[1],
          jobName: deliveryMatch[2],
          result: {
            modelURI: "inline://models/" + deliveryMatch[2] + "/v1",
            metrics: {loss: 0.12, accuracy: 0.98},
            summary: "demo-result"
          }
        };
      }
      if (deliveryMatch && method === "POST" && deliveryMatch[3] === "publish-service") {
        var publishPayload = demoBody(init);
        var sourceApp = demoFindApplication(deliveryMatch[1], deliveryMatch[2]);
        var serviceName = publishPayload.serviceName || deliveryMatch[2] + "-service";
        var serviceApp = {
          namespace: deliveryMatch[1],
          name: serviceName,
          workloadTypes: ["service"],
          phase: "running",
          healthy: true,
          aiMetadata: sourceApp ? demoClone(sourceApp.aiMetadata || {}) : {},
          resourceSummary: {cpuMilli: 2000, memoryMi: 4096, gpu: 0},
          components: [{name: serviceName + "-component"}],
          pods: [{name: serviceName + "-pod-0", phase: "Running", containers: ["http"]}],
          pod: serviceName + "-pod-0",
          container: "http",
          logs: "service booted\nmodelURI=" + ((sourceApp && sourceApp.deliveryResult && sourceApp.deliveryResult.modelURI) || "inline://models/" + deliveryMatch[2] + "/v1") + "\nhealthz ok",
          modelURI: (sourceApp && sourceApp.deliveryResult && sourceApp.deliveryResult.modelURI) || "inline://models/" + deliveryMatch[2] + "/v1"
        };
        demoUpsertApplication(serviceApp);
        demoAppendAudit("publish-service", deliveryMatch[1], serviceName, "Demo AIService 已发布", true);
        return {
          namespace: deliveryMatch[1],
          jobName: deliveryMatch[2],
          serviceName: serviceName,
          modelURI: serviceApp.modelURI,
          application: {
            namespace: deliveryMatch[1],
            name: serviceName,
            dryRun: false
          }
        };
      }
      if (modelMatch && method === "POST" && modelMatch[3] === "evaluate") {
        var evaluatePayload = demoBody(init);
        var jobName = modelMatch[2] + "-eval";
        var evaluationJob = {
          namespace: modelMatch[1],
          name: jobName,
          workloadTypes: ["job"],
          phase: "running",
          healthy: true,
          aiMetadata: {
            "ai.oam.dev/tenant": "demo-tenant",
            "ai.oam.dev/project": "model-evaluation",
            "ai.oam.dev/environment": "poc",
            "ai.oam.dev/owner": "ai-platform"
          },
          resourceSummary: {cpuMilli: 4000, memoryMi: 8192, gpu: 0},
          components: [{name: jobName + "-runner"}],
          pods: [{name: jobName + "-pod-0", phase: "Running", containers: ["eval"]}],
          pod: jobName + "-pod-0",
          container: "eval",
          logs: "eval-start\nAI_RESULT_JSON={\"modelURI\":\"" + (demoFindApplication(modelMatch[1], modelMatch[2]) && demoFindApplication(modelMatch[1], modelMatch[2]).modelURI || "inline://models/" + modelMatch[2] + "/v1") + "\",\"metrics\":{\"accuracy\":0.98},\"summary\":\"evaluation\"}\neval-complete"
        };
        demoUpsertApplication(evaluationJob);
        demoAppendAudit("evaluate", modelMatch[1], jobName, "Demo 评测已提交", true);
        return {
          namespace: modelMatch[1],
          modelName: modelMatch[2],
          jobName: jobName,
          modelURI: demoFindApplication(modelMatch[1], modelMatch[2]) && demoFindApplication(modelMatch[1], modelMatch[2]).modelURI || "inline://models/" + modelMatch[2] + "/v1",
          application: {
            namespace: modelMatch[1],
            name: jobName,
            dryRun: false
          }
        };
      }
      if (modelMatch && method === "POST" && modelMatch[3] === "sync-evaluation") {
        var syncPayload = demoBody(init);
        var modelItem = (demoState.models || []).filter(function(item) {
          return item.namespace === modelMatch[1] && item.name === modelMatch[2];
        })[0];
        if (modelItem) {
          modelItem.evaluationStatus = "passed";
          demoAppendAudit("sync-evaluation", modelMatch[1], modelMatch[2], "Demo 评测结果已同步", true);
        }
        return {
          namespace: modelMatch[1],
          modelName: modelMatch[2],
          jobName: syncPayload.evaluationJobName || modelMatch[2] + "-eval",
          result: {
            modelURI: modelItem && modelItem.modelURI || "inline://models/" + modelMatch[2] + "/v1",
            metrics: {accuracy: 0.98},
            summary: "evaluation"
          },
          asset: demoClone(modelItem || {namespace: modelMatch[1], name: modelMatch[2], evaluationStatus: "passed"})
        };
      }
      if (pathname === "/api/v1/ai/models" && method === "POST") {
        var modelPayload = demoBody(init);
        var modelNamespace = modelPayload.namespace || fields.namespace.value || "default";
        var modelNamespaceError = demoNamespaceError(modelNamespace);
        if (modelNamespaceError) {
          return demoError(modelNamespaceError);
        }
        var modelRecord = {
          namespace: modelNamespace,
          name: modelPayload.name || modelNameFromURI(modelPayload.modelURI || "model"),
          modelURI: modelPayload.modelURI || "inline://models/" + (modelPayload.name || "model") + "/v1",
          status: modelPayload.status || "trained",
          evaluationStatus: modelPayload.evaluationStatus || "pending",
          visibility: modelPayload.visibility || "private",
          metrics: modelPayload.metrics || {},
          summary: modelPayload.summary || ""
        };
        demoState.models = (demoState.models || []).filter(function(item) {
          return item.namespace !== modelRecord.namespace || item.name !== modelRecord.name;
        });
        demoState.models.unshift(modelRecord);
        demoAppendAudit("register-model", modelRecord.namespace, modelRecord.name, "Demo 模型已登记", true);
        return demoClone(modelRecord);
      }
      if (pathname === "/api/v1/ai/datasets" && method === "POST") {
        var datasetPayload = demoBody(init);
        var datasetNamespace = datasetPayload.namespace || fields.namespace.value || "default";
        var datasetNamespaceError = demoNamespaceError(datasetNamespace);
        if (datasetNamespaceError) {
          return demoError(datasetNamespaceError);
        }
        var datasetRecord = {
          namespace: datasetNamespace,
          name: datasetPayload.name || slug(datasetPayload.displayName || "dataset"),
          displayName: datasetPayload.displayName || datasetPayload.name || "Demo 数据集",
          datasetURI: datasetPayload.datasetURI || "inline://datasets/" + (datasetPayload.name || "dataset"),
          format: datasetPayload.format || "sharegpt-jsonl",
          purpose: datasetPayload.purpose || "sft",
          status: datasetPayload.status || "validated",
          visibility: datasetPayload.visibility || "private"
        };
        demoState.datasets = (demoState.datasets || []).filter(function(item) {
          return item.namespace !== datasetRecord.namespace || item.name !== datasetRecord.name;
        });
        demoState.datasets.unshift(datasetRecord);
        demoAppendAudit("register-dataset", datasetRecord.namespace, datasetRecord.name, "Demo 数据集已登记", true);
        return demoClone(datasetRecord);
      }
      return null;
    }
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
    function applySchedulingPreset(preset) {
      setWizardValue("resourceCPU", preset.cpu);
      setWizardValue("resourceMemory", preset.memory);
      setWizardValue("resourceGPU", preset.gpu);
      setWizardValue("schedulingPriority", preset.priority);
      setWizardValue("schedulingNodeType", preset.nodeType);
      setWizardValue("schedulingStrategy", preset.strategy);
      setWizardValue("maxTaskCPU", preset.maxCPU);
      setWizardValue("maxTaskMemory", preset.maxMemory);
      setWizardValue("maxTaskGPU", preset.maxGPU);
    }
    function resourceProfilePresets() {
      return {
        "small-cpu": {cpu: "2", memory: "4Gi", gpu: "0", priority: "normal", nodeType: "cpu", strategy: "cost", maxCPU: "2", maxMemory: "4Gi", maxGPU: "0"},
        "medium-gpu": {cpu: "8", memory: "32Gi", gpu: "1", priority: "high", nodeType: "gpu", strategy: "performance", maxCPU: "8", maxMemory: "32Gi", maxGPU: "1"},
        "large-gpu": {cpu: "16", memory: "64Gi", gpu: "4", priority: "urgent", nodeType: "gpu", strategy: "performance", maxCPU: "16", maxMemory: "64Gi", maxGPU: "4"}
      };
    }
    function applyResourceProfile(profile) {
      var preset = resourceProfilePresets()[profile];
      if (!preset) return;
      applySchedulingPreset(preset);
    }
    function applySchedulingPolicyTemplate(template) {
      var templates = {
        "cost-efficient": {profile: "small-cpu", cpu: "2", memory: "4Gi", gpu: "0", priority: "normal", nodeType: "cpu", strategy: "cost", maxCPU: "2", maxMemory: "4Gi", maxGPU: "0"},
        "performance": {profile: "medium-gpu", cpu: "8", memory: "32Gi", gpu: "1", priority: "high", nodeType: "gpu", strategy: "performance", maxCPU: "8", maxMemory: "32Gi", maxGPU: "1"},
        "fast-start": {profile: "custom", cpu: "4", memory: "8Gi", gpu: "0", priority: "high", nodeType: "any", strategy: "fast-start", maxCPU: "4", maxMemory: "8Gi", maxGPU: "0"},
        "gpu-dedicated": {profile: "large-gpu", cpu: "16", memory: "64Gi", gpu: "4", priority: "urgent", nodeType: "gpu", strategy: "performance", maxCPU: "16", maxMemory: "64Gi", maxGPU: "4"}
      };
      var preset = templates[template];
      if (!preset) return;
      setWizardValue("resourceProfile", preset.profile);
      applySchedulingPreset(preset);
    }
    function schedulingPriorityClassName(priority) {
      var value = (priority || "").trim().toLowerCase();
      if (value === "normal") return "ai-normal";
      if (value === "high") return "ai-high";
      if (value === "urgent") return "ai-urgent";
      return priority || "-";
    }
    function schedulingPolicyDescription(template) {
      var descriptions = {
        "cost-efficient": "成本优先模板会选择 CPU 节点和普通优先级，适合可排队、对 GPU 不敏感的训练或评测。",
        "performance": "性能优先模板会优先选择 GPU 节点并提高任务优先级，适合常规微调训练。",
        "fast-start": "快速启动模板不限制节点类型并提高优先级，适合希望尽快进入运行态的轻量任务。",
        "gpu-dedicated": "GPU 专用模板会使用多卡资源和紧急优先级，适合高吞吐训练或关键模型交付。"
      };
      return descriptions[template] || "按当前资源和调度字段生成映射。";
    }
    function renderSchedulingPolicyExplanation() {
      if (!schedulingPolicyExplanation) return;
      var nodeType = wizardFields.schedulingNodeType.value;
      var nodeSelector = nodeType && nodeType !== "any" ? "ai.oam.dev/node-type=" + nodeType : "不限节点";
      schedulingPolicyExplanation.textContent = [
        "当前调度映射（以 K8s 调度为准）",
        "策略说明：" + schedulingPolicyDescription(wizardFields.schedulingPolicyTemplate.value),
        "Kubernetes 调度：source of truth",
        "集群能力边界：CPU Manager static / CPU pinning、GPU device plugin、Topology Manager",
        "resources.requests/limits：CPU " + wizardFields.resourceCPU.value + " / 内存 " + wizardFields.resourceMemory.value + " / GPU " + wizardFields.resourceGPU.value,
        "priorityClassName：" + schedulingPriorityClassName(wizardFields.schedulingPriority.value),
        "nodeSelector：" + nodeSelector,
        "ai-runtime.schedulingStrategy：" + wizardFields.schedulingStrategy.value
      ].join("\n");
    }
    function parseComputeQuantity(value) {
      var raw = (value || "").trim();
      if (!raw) return null;
      var match = raw.match(/^([0-9]+(?:\.[0-9]+)?)(m|mi|gi|ti)?$/i);
      if (!match) return null;
      var amount = parseFloat(match[1]);
      if (isNaN(amount)) return null;
      var unit = (match[2] || "").toLowerCase();
      if (unit === "m") return amount / 1000;
      if (unit === "mi") return amount / 1024;
      if (unit === "gi" || unit === "") return amount;
      if (unit === "ti") return amount * 1024;
      return null;
    }
    function resourceLimitViolation(label, requested, limit) {
      var requestedValue = parseComputeQuantity(requested);
      var limitValue = parseComputeQuantity(limit);
      if (requestedValue === null || limitValue === null || !limit) {
        return "";
      }
      if (requestedValue > limitValue) {
        return label + " 申请 " + requested + " 超过上限 " + limit;
      }
      return "";
    }
    function isolationPolicyIssues() {
      var issues = [
        resourceLimitViolation("CPU", wizardFields.resourceCPU.value, wizardFields.maxTaskCPU.value),
        resourceLimitViolation("内存", wizardFields.resourceMemory.value, wizardFields.maxTaskMemory.value),
        resourceLimitViolation("GPU", wizardFields.resourceGPU.value, wizardFields.maxTaskGPU.value),
        resourceLimitViolation("CPU", wizardFields.resourceCPU.value, wizardFields.quotaCPU.value),
        resourceLimitViolation("内存", wizardFields.resourceMemory.value, wizardFields.quotaMemory.value),
        resourceLimitViolation("GPU", wizardFields.resourceGPU.value, wizardFields.quotaGPU.value)
      ];
      return issues.filter(function(item) { return !!item; });
    }
    function renderIsolationPolicyPreview() {
      if (!isolationPolicyPreview) return;
      var issues = isolationPolicyIssues();
      isolationPolicyPreview.className = "scheduling-explainer " + (issues.length ? "policy-warn" : "policy-ok");
      isolationPolicyPreview.textContent = [
        "隔离策略预览",
        "Namespace：" + (wizardFields.tenantNamespace.value || fields.namespace.value || "-"),
        "ResourceQuota：CPU " + wizardFields.quotaCPU.value + " / 内存 " + wizardFields.quotaMemory.value + " / GPU " + wizardFields.quotaGPU.value,
        "LimitRange：单任务 CPU " + wizardFields.maxTaskCPU.value + " / 内存 " + wizardFields.maxTaskMemory.value + " / GPU " + wizardFields.maxTaskGPU.value,
        "配额校验：" + (issues.length ? issues.join("；") : "当前申请资源在租户配额和单任务上限内。")
      ].join("\n");
    }
    function validateIsolationPolicy() {
      renderIsolationPolicyPreview();
      var issues = isolationPolicyIssues();
      if (!issues.length) return true;
      setError("资源隔离校验未通过：" + issues.join("；"));
      cards.appendChild(card("资源隔离校验", "未通过", "warn"));
      return false;
    }
    function setBaseModel(uri) {
      setWizardValue("baseModelURI", uri || "");
      if (baseModelSelect && uri) {
        baseModelSelect.value = uri;
      }
    }
    function setServiceModel(uri) {
      setWizardValue("serviceModelURI", uri || "");
      if (serviceModelSelect && uri) {
        serviceModelSelect.value = uri;
      }
    }
    function setEvaluationDataset(uri) {
      if (evaluationDatasetURI) {
        evaluationDatasetURI.value = uri || "";
      }
      if (evaluationDatasetSelect && uri) {
        evaluationDatasetSelect.value = uri;
      }
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
    function currentJobDatasetURI(template) {
      var mode = template || (wizardFields.algorithmTemplate ? wizardFields.algorithmTemplate.value : "sft");
      if (mode === "evaluation") {
        return (evaluationDatasetURI && evaluationDatasetURI.value) || fields.datasetURI.value || "";
      }
      return (wizardFields.trainingDataURI && wizardFields.trainingDataURI.value) || fields.datasetURI.value || "";
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
    function syncWorkflowMode(template) {
      var isTraining = template === "sft";
      var isEvaluation = template === "evaluation";
      var isService = template === "service";
      if (baseModelLabel) {
        baseModelLabel.textContent = isEvaluation ? "评测模型" : isService ? "服务画像基座" : "基础模型";
      }
      if (workflowModelHint) {
        workflowModelHint.textContent = isEvaluation
          ? "选择待评测模型，再在评测配置里补充数据集和阈值。"
          : isService
            ? "服务画像会把当前模型引用整理为 AIService 配置，不新增运行时类型。"
            : "训练模式会把这个模型作为微调起点，训练数据和超参在下方配置。";
      }
      if (evaluationBand) {
        evaluationBand.style.display = isEvaluation ? "block" : "none";
        evaluationBand.open = isEvaluation;
      }
      if (trainingBand) {
        trainingBand.style.display = isTraining ? "block" : "none";
        trainingBand.open = isTraining;
      }
      Array.prototype.forEach.call(document.querySelectorAll(".workflow-summary"), function(node) {
        node.style.display = isService ? "none" : "grid";
      });
      Array.prototype.forEach.call(document.querySelectorAll(".wizard-steps"), function(node) {
        node.style.display = isService ? "none" : "grid";
      });
      Array.prototype.forEach.call(document.querySelectorAll(".shared-model-field"), function(node) {
        node.style.display = isService ? "none" : "grid";
      });
      Array.prototype.forEach.call(document.querySelectorAll(".train-field"), function(node) {
        node.style.display = isTraining ? "grid" : "none";
      });
      Array.prototype.forEach.call(document.querySelectorAll(".service-template-field"), function(node) {
        node.style.display = isService ? "grid" : "none";
      });
    }
    function setActiveTemplate(template) {
      Array.prototype.forEach.call(document.querySelectorAll(".template-card"), function(node) {
        node.classList.remove("active");
      });
      var active = document.getElementById("template-" + (template === "evaluation" ? "eval" : template === "service" ? "service" : "sft"));
      if (active) active.classList.add("active");
      syncWorkflowMode(template);
    }
    function applyWizardToAdvanced() {
      var template = wizardFields.algorithmTemplate.value;
      var taskName = slug(wizardFields.taskDisplayName.value);
      setActiveTemplate(template);
      var namespace = wizardFields.tenantNamespace.value || fields.namespace.value || accountNamespace || "sock-shop";
      if (template === "service") {
        setValue("kind", "AIService");
        setValue("name", taskName || "model-service-demo");
        setValue("namespace", namespace);
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
      setValue("namespace", namespace);
      setValue("component", taskName ? taskName + (template === "evaluation" ? "-evaluator" : "-trainer") : (template === "evaluation" ? "evaluation-runner" : "sft-trainer"));
      setValue("tenant", fields.tenant.value || "demo-tenant");
      setValue("project", project);
      setValue("environment", fields.environment.value || "poc");
      setValue("owner", fields.owner.value || "ai-platform");
      setValue("runtime", "batch");
      setValue("image", "busybox:1.36");
      setValue("jobKind", template === "evaluation" ? "evaluation" : "training");
      setValue("ttl", "3600");
      setValue("datasetURI", currentJobDatasetURI(template));
      setValue("outputURI", "inline://outputs/" + (taskName || "customer-sft-demo"));
      deliveryServiceName.value = (taskName || "customer-sft-demo") + "-service";
    }
    function deliveryResultJSONString() {
      var template = wizardFields.algorithmTemplate ? wizardFields.algorithmTemplate.value : "sft";
      return JSON.stringify({
        modelURI: "inline://models/" + fields.name.value + "/v1",
        metrics: { loss: 0.12, accuracy: 0.98 },
        summary: template === "evaluation" ? "evaluation-passed" : template === "service" ? "service-ready" : "sft-fine-tuned"
      });
    }
    function buildDeliveryJobResultScript() {
      var template = wizardFields.algorithmTemplate ? wizardFields.algorithmTemplate.value : "sft";
      var dataset = currentJobDatasetURI(template);
      var resultLines = [
        "    imagePullPolicy: IfNotPresent",
        "    cmd:",
        "      - sh",
        "      - -c",
        "    args:",
        "      - |",
        "        echo algorithm=" + template,
        "        echo base_model=" + (wizardFields.baseModelURI ? wizardFields.baseModelURI.value : "inline://models/base"),
        "        echo dataset=" + dataset
      ];
      if (template === "evaluation") {
        resultLines = resultLines.concat([
          "        echo evaluation_threshold=" + (evaluationThreshold ? evaluationThreshold.value : "0.8"),
          "        echo eval-start",
          "        echo 'AI_RESULT_JSON=" + deliveryResultJSONString() + "'",
          "        echo eval-complete"
        ]);
      } else {
        resultLines = resultLines.concat([
          "        echo training_size=" + (wizardFields.trainingSize ? wizardFields.trainingSize.value : "small"),
          "        echo epochs=" + (wizardFields.epochs ? wizardFields.epochs.value : "1") + " learning_rate=" + (wizardFields.learningRate ? wizardFields.learningRate.value : "2e-5"),
          "        echo train-start",
          "        echo epoch=1 loss=0.30",
          "        echo epoch=2 loss=0.12",
          "        echo 'AI_RESULT_JSON=" + deliveryResultJSONString() + "'",
          "        echo train-complete"
        ]);
      }
      resultLines.push("    backoffLimit: 0");
      return resultLines.join("\n");
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
          line("uri", currentJobDatasetURI(), 6),
          "    output:",
          line("uri", fields.outputURI.value, 6),
          line("ttlSecondsAfterFinished", fields.ttl.value, 4)
        ]);
      }
      common = common.concat([
        line("resources", "", 2),
        line("cpu", wizardFields.resourceCPU.value, 4),
        line("memory", wizardFields.resourceMemory.value, 4),
        line("gpu", wizardFields.resourceGPU.value, 4),
        line("scheduling", "", 2),
        line("priority", wizardFields.schedulingPriority.value, 4),
        line("nodeType", wizardFields.schedulingNodeType.value, 4),
        line("strategy", wizardFields.schedulingStrategy.value, 4),
        line("isolation", "", 2),
        line("tenantNamespace", wizardFields.tenantNamespace.value || fields.namespace.value, 4),
        "    resourceQuota:",
        line("cpu", wizardFields.quotaCPU.value, 6),
        line("memory", wizardFields.quotaMemory.value, 6),
        line("gpu", wizardFields.quotaGPU.value, 6),
        "    limitRange:",
        line("maxCpuPerTask", wizardFields.maxTaskCPU.value, 6),
        line("maxMemoryPerTask", wizardFields.maxTaskMemory.value, 6),
        line("maxGpuPerTask", wizardFields.maxTaskGPU.value, 6)
      ]);
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
        common.push(line("datasetURI", currentJobDatasetURI(), 4));
        if (wizardFields.algorithmTemplate.value !== "evaluation") {
          common.push(line("trainingSize", wizardFields.trainingSize.value, 4));
        }
      }
      return common.join("\n") + "\n";
    }
    function generateYAML(applyWizard) {
      if (applyWizard !== false) {
        applyWizardToAdvanced();
      }
      syncVisibility();
      yaml.value = buildYAML();
      renderSchedulingPolicyExplanation();
      renderIsolationPolicyPreview();
    }
    function fillServiceForm() {
      setWizardValue("algorithmTemplate", "service");
      setWizardValue("taskDisplayName", "sentiment-demo");
      setServiceModel("oss://models/sentiment/v1");
      setWizardValue("schedulingPolicyTemplate", "cost-efficient");
      setWizardValue("resourceProfile", "small-cpu");
      applyResourceProfile("small-cpu");
      setWizardValue("tenantNamespace", accountNamespace || "ai-demo");
      setValue("kind", "AIService");
      setValue("name", "sentiment-demo");
      setValue("namespace", accountNamespace || "ai-demo");
      setValue("component", "sentiment-api");
      setValue("tenant", accountTenant || "demo-tenant");
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
      setBaseModel("modelscope://qwen/Qwen2.5-0.5B");
      setTrainingDataset("inline://datasets/customer-sft-demo");
      setWizardValue("publishAfterTrain", "true");
      setWizardValue("schedulingPolicyTemplate", "performance");
      setWizardValue("resourceProfile", "medium-gpu");
      applyResourceProfile("medium-gpu");
      setWizardValue("tenantNamespace", accountNamespace || "sock-shop");
      setValue("kind", "AIJob");
      setValue("name", "customer-sft-demo");
      setValue("namespace", accountNamespace || "sock-shop");
      setValue("component", "customer-sft-demo-trainer");
      setValue("tenant", accountTenant || "demo-tenant");
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
      return fetchJSON(deliveryBase(target) + "/publish-service", {
        method: "POST",
        headers: {"Content-Type": "application/json", "X-AI-User": "console"},
        body: JSON.stringify(payload)
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
      return fetchJSON("/api/v1/ai/applications/" + encodeURIComponent(target.namespace) + "/" + encodeURIComponent(target.name) + "/probe", {
        method: "POST",
        headers: {"Content-Type": "application/json", "X-AI-User": "console"},
        body: JSON.stringify({path: "/healthz", timeoutSeconds: 5})
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
        var display = item.displayName || item.name || "未命名数据集";
        var assetKey = (item.namespace || fields.namespace.value || "-") + "/" + (item.name || display || "-");
        var option = document.createElement("option");
        option.value = uri;
        option.textContent = display + " · " + (item.format || "unknown") + " · " + (item.status || "registered");
        trainingDatasetSelect.appendChild(option);
        var row = document.createElement("div");
        row.className = "task-row";
        row.innerHTML = [
          "<strong>" + display + "</strong>",
          "<span>数据集编号：" + assetKey + "</span>",
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
        deliveryOutput.textContent = "请先选择或登记一个数据集。";
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
      return fetchJSON("/api/v1/ai/datasets", {
        method: "POST",
        headers: {"Content-Type": "application/json", "X-AI-User": "console"},
        body: JSON.stringify(payload)
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
      serviceModelSelect.innerHTML = "";
      if (!items || !items.length) {
        var option = document.createElement("option");
        option.value = wizardFields.serviceModelURI.value || "inline://models/customer-sft-demo/v1";
        option.textContent = "customer-sft-demo 示例模型";
        serviceModelSelect.appendChild(option);
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
        var option = document.createElement("option");
        option.value = item.modelURI || "";
        option.textContent = (item.name || "未命名模型") + " · " + (item.status || "trained") + " · " + (item.evaluationStatus || "pending");
        serviceModelSelect.appendChild(option);
        row.innerHTML = [
          "<strong>" + name + "</strong>",
          "<span>资产编号：" + name + "</span>",
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
          setBaseModel(item.modelURI || "");
          setWizardValue("taskDisplayName", (item.name || "model") + "-sft");
          generateYAML(true);
          lastAction.textContent = "已将模型设为基础模型";
        };
        row.querySelector(".serve-model").onclick = function() {
          setWizardValue("algorithmTemplate", "service");
          setServiceModel(item.modelURI || "");
          setWizardValue("taskDisplayName", (item.name || "model") + "-service");
          generateYAML(true);
          lastAction.textContent = "已载入发布此模型模板";
        };
        modelList.appendChild(row);
      });
      if (wizardFields.serviceModelURI.value) {
        serviceModelSelect.value = wizardFields.serviceModelURI.value;
      } else if (items[0] && items[0].modelURI) {
        setServiceModel(items[0].modelURI);
      }
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
      return fetchJSON("/api/v1/ai/models", {
        method: "POST",
        headers: {"Content-Type": "application/json", "X-AI-User": "console"},
        body: JSON.stringify(payload)
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
      return fetchJSON(modelActionBase(item) + "/evaluate", {
        method: "POST",
        headers: {"Content-Type": "application/json", "X-AI-User": "console"},
        body: JSON.stringify(payload)
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
      return fetchJSON(modelActionBase(item) + "/sync-evaluation", {
        method: "POST",
        headers: {"Content-Type": "application/json", "X-AI-User": "console"},
        body: JSON.stringify({evaluationJobName: jobName})
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
      return fetchJSON(lifecyclePath(target, action), {
        method: action === "delete" ? "DELETE" : "POST",
        headers: {"X-AI-User": "console"}
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
    function defaultSectionForRole(role) {
      if (role === "monitor") {
        return "overview";
      }
      return demoMode ? "platform" : "training";
    }
    function activeView() {
      return document.body.getAttribute("data-role") === "monitor" ? document.getElementById("monitor-view") : document.getElementById("user-view");
    }
    function setActiveSection(name) {
      var view = activeView();
      if (!view) {
        return;
      }
      var matched = false;
      Array.prototype.forEach.call(view.querySelectorAll(".section-pane"), function(pane) {
        var active = pane.getAttribute("data-section") === name;
        pane.classList.toggle("active", active);
        if (active) {
          matched = true;
        }
      });
      if (!matched) {
        var fallback = view.querySelector(".section-pane");
        if (fallback) {
          fallback.classList.add("active");
        }
      }
      Array.prototype.forEach.call(document.querySelectorAll(".sidebar-nav .nav-item[data-section]"), function(button) {
        button.classList.toggle("active", button.getAttribute("data-section") === name);
      });
      if (document.body.getAttribute("data-role") === "user" && name === "platform") {
        refreshTasks();
      }
      if (document.body.getAttribute("data-role") === "monitor") {
        refreshMonitor();
      }
    }
    function setActiveView(name) {
      var userActive = name === "user";
      document.getElementById("user-view").className = userActive ? "view active" : "view";
      document.getElementById("monitor-view").className = userActive ? "view" : "view active";
      setActiveSection(defaultSectionForRole(name));
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
    function fetchJSON(path, init) {
      var demo = demoResponse(path, init);
      if (demoMode && demo !== null) {
        if (demo.__demoError) {
          return Promise.reject(new Error(demo.__demoError));
        }
        return Promise.resolve(demoClone(demo));
      }
      return fetch(path, init).then(function(res) {
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
      return fetchJSON(path, {
        method: "POST",
        headers: {"Content-Type": "application/yaml"},
        body: yaml.value
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
    function uniqueFieldCount(items, getter) {
      var values = {};
      items.forEach(function(item) {
        var value = getter(item);
        if (value) {
          values[value] = true;
        }
      });
      return Object.keys(values).length;
    }
    function aggregatePhaseCounts(items) {
      var summary = {
        running: 0,
        succeeded: 0,
        pending: 0,
        failed: 0,
        unknown: 0
      };
      items.forEach(function(item) {
        var phase = String(item.phase || "unknown").toLowerCase();
        if (summary.hasOwnProperty(phase)) {
          summary[phase] += 1;
        } else {
          summary.unknown += 1;
        }
      });
      return summary;
    }
    function buildMonitorSummary(items) {
      var serviceCount = 0;
      var jobCount = 0;
      var runningCount = 0;
      var unhealthyCount = 0;
      var healthyCount = 0;
      items.forEach(function(item) {
        var types = item.workloadTypes || [];
        if (types.indexOf("service") !== -1) serviceCount++;
        if (types.indexOf("job") !== -1) jobCount++;
        if (item.phase === "running") runningCount++;
        if (!item.healthy) unhealthyCount++;
        if (item.healthy) healthyCount++;
      });
      var tenantRows = aggregateTenantResources(items);
      return {
        items: items,
        taskCount: items.length,
        serviceCount: serviceCount,
        jobCount: jobCount,
        runningCount: runningCount,
        unhealthyCount: unhealthyCount,
        healthyCount: healthyCount,
        healthyRate: items.length ? Math.round(healthyCount * 100 / items.length) : 0,
        tenantCount: tenantRows.length,
        namespaceCount: uniqueFieldCount(items, function(item) { return item.namespace; }),
        projectCount: uniqueFieldCount(items, function(item) { return metadata(item, "project"); }),
        phaseCounts: aggregatePhaseCounts(items),
        resourceSummary: aggregateMonitorResources(items),
        tenantRows: tenantRows
      };
    }
    function renderMonitorCards(summary) {
      monitorCards.innerHTML = "";
      monitorCards.appendChild(card("任务总数", summary.taskCount, ""));
      monitorCards.appendChild(card("AIService 数量", summary.serviceCount, "ok"));
      monitorCards.appendChild(card("AIJob 数量", summary.jobCount, "ok"));
      monitorCards.appendChild(card("健康率", summary.healthyRate + "%", summary.healthyRate >= 80 ? "ok" : "warn"));
      monitorCards.appendChild(card("异常任务", summary.unhealthyCount, summary.unhealthyCount ? "warn" : "ok"));
      monitorCards.appendChild(card("活跃租户", summary.tenantCount, ""));
    }
    function aggregateMonitorResources(items) {
      var summary = {cpuMilli: 0, memoryMi: 0, gpu: 0};
      items.forEach(function(item) {
        var resourceSummary = item.resourceSummary || {};
        summary.cpuMilli += resourceSummary.cpuMilli || 0;
        summary.memoryMi += resourceSummary.memoryMi || 0;
        summary.gpu += resourceSummary.gpu || 0;
      });
      return summary;
    }
    function formatCpuMilli(value) {
      if (!value) {
        return "0m";
      }
      if (value < 1000) {
        return value + "m";
      }
      var cores = Math.round((value / 1000) * 10) / 10;
      return (Number.isInteger(cores) ? cores : cores.toFixed(1)) + "核";
    }
    function formatMemoryMi(value) {
      if (!value) {
        return "0Mi";
      }
      if (value < 1024) {
        return value + "Mi";
      }
      var gib = Math.round((value / 1024) * 10) / 10;
      return (Number.isInteger(gib) ? gib : gib.toFixed(1)) + "Gi";
    }
    function renderMonitorResourceCards(summary) {
      var resourceSummary = summary.resourceSummary || aggregateMonitorResources(summary.items || []);
      monitorResourceCards.innerHTML = "";
      monitorResourceCards.appendChild(card("CPU 申请总量", formatCpuMilli(resourceSummary.cpuMilli), "ok"));
      monitorResourceCards.appendChild(card("内存 申请总量", formatMemoryMi(resourceSummary.memoryMi), "ok"));
      monitorResourceCards.appendChild(card("GPU 申请总量", resourceSummary.gpu + " GPU", "ok"));
    }
    function pushMonitorSnapshot(summary) {
      monitorHistory.push({
        taskCount: summary.taskCount,
        runningCount: summary.runningCount,
        unhealthyCount: summary.unhealthyCount,
        healthyRate: summary.healthyRate,
        serviceCount: summary.serviceCount,
        jobCount: summary.jobCount
      });
      while (monitorHistory.length > 6) {
        monitorHistory.shift();
      }
    }
    function clearNode(node) {
      if (!node) {
        return;
      }
      while (node.firstChild) {
        node.removeChild(node.firstChild);
      }
    }
    function createSvgElement(tag, attrs) {
      var node = document.createElementNS("http://www.w3.org/2000/svg", tag);
      Object.keys(attrs || {}).forEach(function(key) {
        node.setAttribute(key, attrs[key]);
      });
      return node;
    }
    function renderMonitorTrendChart() {
      clearNode(monitorTrendChart);
      if (!monitorTrendChart) {
        return;
      }
      var history = monitorHistory.slice();
      if (!history.length) {
        var empty = document.createElement("div");
        empty.className = "chart-empty";
        empty.textContent = "暂无刷新记录，点击“刷新监测”后开始累计。";
        monitorTrendChart.appendChild(empty);
        return;
      }
      if (monitorTrendNote) {
        monitorTrendNote.textContent = history.length + " 次刷新";
      }
      var surface = document.createElement("div");
      surface.className = "chart-surface";
      var svg = createSvgElement("svg", {
        class: "trend-svg",
        viewBox: "0 0 320 146",
        preserveAspectRatio: "none"
      });
      var padding = {left: 30, right: 14, top: 14, bottom: 18};
      var width = 320 - padding.left - padding.right;
      var height = 146 - padding.top - padding.bottom;
      var series = [
        {key: "taskCount", label: "任务数", color: "#1677ff"},
        {key: "runningCount", label: "Running", color: "#16884a"},
        {key: "unhealthyCount", label: "异常", color: "#d13212"}
      ];
      var plot = history.length === 1 ? [history[0], history[0]] : history;
      var maxValue = 1;
      plot.forEach(function(point) {
        series.forEach(function(def) {
          maxValue = Math.max(maxValue, point[def.key] || 0);
        });
      });
      for (var i = 0; i < 4; i++) {
        var y = padding.top + (height / 3) * i;
        svg.appendChild(createSvgElement("line", {
          x1: padding.left,
          x2: padding.left + width,
          y1: y,
          y2: y,
          stroke: i === 3 ? "#d5d9d9" : "#eaeced",
          "stroke-width": "1"
        }));
      }
      series.forEach(function(def) {
        var points = [];
        plot.forEach(function(point, index) {
          var x = plot.length === 1 ? padding.left + width / 2 : padding.left + (width * index / (plot.length - 1));
          var value = point[def.key] || 0;
          var y = padding.top + height - (value / maxValue) * height;
          points.push([x, y]);
        });
        if (!points.length) {
          return;
        }
        if (points.length === 1) {
          points.push([points[0][0], points[0][1]]);
        }
        svg.appendChild(createSvgElement("polyline", {
          points: points.map(function(pair) { return pair.join(","); }).join(" "),
          fill: "none",
          stroke: def.color,
          "stroke-width": "2.5",
          "stroke-linecap": "round",
          "stroke-linejoin": "round"
        }));
        points.forEach(function(pair) {
          svg.appendChild(createSvgElement("circle", {
            cx: pair[0],
            cy: pair[1],
            r: "3.2",
            fill: def.color,
            stroke: "#fff",
            "stroke-width": "1.5"
          }));
        });
      });
      surface.appendChild(svg);
      var legend = document.createElement("div");
      legend.className = "chart-legend";
      var latest = history[history.length - 1];
      series.forEach(function(def) {
        var item = document.createElement("span");
        item.className = "legend-item";
        item.innerHTML = '<span class="legend-swatch" style="background:' + def.color + '"></span>' + def.label + " " + (latest[def.key] || 0);
        legend.appendChild(item);
      });
      surface.appendChild(legend);
      monitorTrendChart.appendChild(surface);
    }
    function renderMonitorStatusChart(summary) {
      clearNode(monitorStatusChart);
      if (!monitorStatusChart) {
        return;
      }
      var total = summary.taskCount || 0;
      if (!total) {
        var empty = document.createElement("div");
        empty.className = "chart-empty";
        empty.textContent = "暂无任务数据。";
        monitorStatusChart.appendChild(empty);
        return;
      }
      if (monitorStatusNote) {
        monitorStatusNote.textContent = summary.healthyRate + "% 健康率";
      }
      var phases = [
        {key: "running", label: "Running", color: "#1677ff"},
        {key: "succeeded", label: "Succeeded", color: "#16884a"},
        {key: "pending", label: "Pending", color: "#ff9900"},
        {key: "failed", label: "失败", color: "#d13212"},
        {key: "unknown", label: "其他", color: "#9aa6b2"}
      ];
      var wrap = document.createElement("div");
      wrap.className = "donut-wrap";
      var ring = document.createElement("div");
      ring.className = "donut-ring";
      var stops = [];
      var start = 0;
      phases.forEach(function(phase) {
        var count = summary.phaseCounts[phase.key] || 0;
        if (!count) {
          return;
        }
        var slice = count * 100 / total;
        stops.push(phase.color + " " + start + "% " + (start + slice) + "%");
        start += slice;
      });
      ring.style.background = stops.length ? "conic-gradient(" + stops.join(", ") + ")" : "#eef4ff";
      var center = document.createElement("div");
      center.className = "donut-center";
      center.innerHTML = "<div><strong>" + total + "</strong><span>任务</span></div>";
      ring.appendChild(center);
      wrap.appendChild(ring);
      var list = document.createElement("div");
      list.className = "donut-list";
      phases.forEach(function(phase) {
        var count = summary.phaseCounts[phase.key] || 0;
        var row = document.createElement("div");
        row.className = "donut-row";
        row.innerHTML = "<strong><span class='legend-swatch' style='background:" + phase.color + "'></span>" + phase.label + "</strong><span>" + count + " / " + Math.round(count * 100 / total) + "%</span>";
        list.appendChild(row);
      });
      wrap.appendChild(list);
      monitorStatusChart.appendChild(wrap);
    }
    function renderMonitorTenantChart(items) {
      clearNode(monitorTenantChart);
      if (!monitorTenantChart) {
        return;
      }
      var rows = aggregateTenantResources(items).slice().sort(function(a, b) {
        if (b.taskCount !== a.taskCount) {
          return b.taskCount - a.taskCount;
        }
        if (b.cpuMilli !== a.cpuMilli) {
          return b.cpuMilli - a.cpuMilli;
        }
        return a.tenant.localeCompare(b.tenant);
      }).slice(0, 5);
      if (!rows.length) {
        var empty = document.createElement("div");
        empty.className = "chart-empty";
        empty.textContent = "当前筛选条件下暂无租户排行。";
        monitorTenantChart.appendChild(empty);
        return;
      }
      if (monitorTenantNote) {
        monitorTenantNote.textContent = "Top " + rows.length + " · 按任务数";
      }
      var maxTasks = rows[0].taskCount || 1;
      var list = document.createElement("div");
      list.className = "bar-list";
      rows.forEach(function(row, index) {
        var item = document.createElement("div");
        item.className = "bar-row";
        item.innerHTML = [
          "<div class=\"bar-meta\"><strong>" + row.tenant + "</strong><span>" + row.taskCount + " 任务 · " + row.serviceCount + " 服务 / " + row.jobCount + " 训练</span></div>",
          "<div class=\"bar-track\"><div class=\"bar-fill\" style=\"width:" + Math.max(8, Math.round(row.taskCount * 100 / maxTasks)) + "%;background:" + ["#1677ff", "#13a8a8", "#ff9900", "#d13212", "#69a7ff"][index % 5] + "\"></div></div>",
          "<div class=\"bar-meta\"><span>CPU " + formatCpuMilli(row.cpuMilli) + "</span><span>内存 " + formatMemoryMi(row.memoryMi) + "</span><span>" + row.gpu + " GPU</span></div>"
        ].join("");
        list.appendChild(item);
      });
      monitorTenantChart.appendChild(list);
    }
    function renderMonitorDashboard(items) {
      var summary = buildMonitorSummary(items);
      pushMonitorSnapshot(summary);
      renderMonitorCards(summary);
      renderMonitorResourceCards(summary);
      renderMonitorTrendChart();
      renderMonitorStatusChart(summary);
      renderMonitorTenantChart(items);
      return summary;
    }
    function aggregateTenantResources(items) {
      var groups = {};
      items.forEach(function(item) {
        var tenant = metadata(item, "tenant") || "未标记租户";
        if (!groups[tenant]) {
          groups[tenant] = {
            tenant: tenant,
            cpuMilli: 0,
            memoryMi: 0,
            gpu: 0,
            taskCount: 0,
            serviceCount: 0,
            jobCount: 0
          };
        }
        var summary = item.resourceSummary || {};
        var group = groups[tenant];
        group.cpuMilli += summary.cpuMilli || 0;
        group.memoryMi += summary.memoryMi || 0;
        group.gpu += summary.gpu || 0;
        group.taskCount += 1;
        var workloadTypes = item.workloadTypes || [];
        if (workloadTypes.indexOf("service") !== -1) {
          group.serviceCount += 1;
        }
        if (workloadTypes.indexOf("job") !== -1) {
          group.jobCount += 1;
        }
      });
      return Object.keys(groups).map(function(key) {
        return groups[key];
      }).sort(function(a, b) {
        if (b.cpuMilli !== a.cpuMilli) {
          return b.cpuMilli - a.cpuMilli;
        }
        return a.tenant.localeCompare(b.tenant);
      });
    }
    function renderTenantResourceRows(container, items) {
      var rows = aggregateTenantResources(items);
      container.innerHTML = "";
      if (!rows.length) {
        var empty = document.createElement("div");
        empty.className = "empty";
        empty.textContent = "当前筛选条件下暂无租户资源记录。";
        container.appendChild(empty);
        return;
      }
      rows.forEach(function(row) {
        var node = document.createElement("div");
        node.className = "tenant-row";
        node.innerHTML = [
          "<strong>" + row.tenant + "</strong>",
          "<span>" + row.taskCount + " 个任务 · " + row.serviceCount + " 服务 / " + row.jobCount + " 训练</span>",
          "<span>CPU " + formatCpuMilli(row.cpuMilli) + "</span>",
          "<span>内存 " + formatMemoryMi(row.memoryMi) + "</span>",
          "<span class=\"pill\">" + row.gpu + " GPU</span>"
        ].join("");
        container.appendChild(node);
      });
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
        renderMonitorDashboard(items);
        renderTenantResourceRows(monitorTenantList, items);
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
      lastAction.textContent = "已载入服务画像";
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
    document.getElementById("generate").onclick = function() {
      generateYAML(true);
      lastAction.textContent = "已生成 YAML";
    };
    Object.keys(wizardFields).forEach(function(id) {
      wizardFields[id].addEventListener("input", function() { generateYAML(true); });
      wizardFields[id].addEventListener("change", function() {
        if (id === "schedulingPolicyTemplate") {
          applySchedulingPolicyTemplate(wizardFields.schedulingPolicyTemplate.value);
        }
        if (id === "resourceProfile") {
          applyResourceProfile(wizardFields.resourceProfile.value);
        }
        generateYAML(true);
      });
    });
    Object.keys(fields).forEach(function(id) {
      fields[id].addEventListener("input", function() { generateYAML(false); });
      fields[id].addEventListener("change", function() { generateYAML(false); });
    });
    document.getElementById("validate").onclick = function() {
      if (!validateIsolationPolicy()) return;
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
      if (!validateIsolationPolicy()) return;
      post("/api/v1/ai/normalize").then(function(data) {
        renderCards(data);
        setJSON(data, "已归一化");
      }).catch(function(err) {
        setError(err.message);
      });
    };
    document.getElementById("deploy").onclick = function() {
      if (!validateIsolationPolicy()) return;
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
    baseModelSelect.onchange = function() {
      setBaseModel(baseModelSelect.value);
      generateYAML(true);
    };
    serviceModelSelect.onchange = function() {
      setServiceModel(serviceModelSelect.value);
      generateYAML(true);
    };
    evaluationDatasetSelect.onchange = function() {
      setEvaluationDataset(evaluationDatasetSelect.value);
      generateYAML(true);
    };
    if (evaluationThreshold) {
      evaluationThreshold.oninput = function() {
        generateYAML(true);
      };
      evaluationThreshold.onchange = function() {
        generateYAML(true);
      };
    }
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
    document.getElementById("nav-training").onclick = function() { setActiveSection("training"); };
    document.getElementById("nav-platform").onclick = function() { setActiveSection("platform"); };
    document.getElementById("nav-monitor-overview").onclick = function() { setActiveSection("overview"); };
    document.getElementById("nav-monitor-tasks").onclick = function() { setActiveSection("tasks"); };
    document.getElementById("nav-monitor-audits").onclick = function() { setActiveSection("audits"); };
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
    fillJobForm();
    setActiveView(document.body.getAttribute("data-role") || "user");
    Object.keys(monitorFields).forEach(function(id) {
      monitorFields[id].addEventListener("change", refreshMonitor);
    });
    if (demoMode) {
      document.getElementById("health").textContent = "本地演示";
    } else {
      fetch("/healthz").then(function(res) {
        return res.text();
      }).then(function(text) {
        document.getElementById("health").textContent = text.trim() ? "正常" : "正常";
      }).catch(function() {
        document.getElementById("health").textContent = "离线";
      });
    }
  </script>
</body>
</html>`
