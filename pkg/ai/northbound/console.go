package northbound

const consoleHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>AI Northbound Console</title>
  <style>
    :root {
      --ink: #18211f;
      --muted: #5d6b66;
      --paper: #fffaf0;
      --line: rgba(24, 33, 31, .14);
      --field: rgba(255, 255, 255, .72);
      --moss: #3f5f4c;
      --rust: #b85f37;
      --gold: #e4b95b;
      --blue: #315f7c;
      --shadow: 0 22px 80px rgba(24, 33, 31, .16);
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      color: var(--ink);
      font-family: "Avenir Next", "Gill Sans", "Trebuchet MS", sans-serif;
      background:
        radial-gradient(circle at top left, rgba(228, 185, 91, .45), transparent 32rem),
        radial-gradient(circle at 86% 14%, rgba(49, 95, 124, .18), transparent 28rem),
        linear-gradient(135deg, #f7efe0 0%, #efe4cf 44%, #dfe9df 100%);
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
      font-family: Georgia, "Times New Roman", serif;
      font-size: clamp(42px, 7vw, 92px);
      line-height: .88;
      letter-spacing: -0.06em;
    }
    .lede {
      max-width: 760px;
      margin: 18px 0 0;
      color: var(--muted);
      font-size: 18px;
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
      letter-spacing: .16em;
      text-transform: uppercase;
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
      min-height: 568px;
      display: block;
      padding: 22px;
      border: 0;
      resize: vertical;
      outline: none;
      color: #17211e;
      background: var(--field);
      font-family: "SFMono-Regular", "Cascadia Code", "Liberation Mono", monospace;
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
      font-family: "SFMono-Regular", "Cascadia Code", "Liberation Mono", monospace;
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
      .panel { min-height: auto; border-radius: 24px; }
      textarea { min-height: 430px; }
      .result-grid { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
  <main>
    <section class="hero">
      <div>
        <div class="label">AI management northbound PoC</div>
        <h1>AI Northbound Console</h1>
        <p class="lede">A minimal verification surface for AIService and AIJob domain YAML. Validate the document, normalize semantic intent, and inspect governance signals before any deployment path is exposed.</p>
      </div>
      <div class="status-strip">
        <div class="status-card">
          <div class="label">API health</div>
          <div class="metric" id="health">checking</div>
        </div>
        <div class="status-card">
          <div class="label">Validate endpoint</div>
          <div class="metric">/api/v1/ai/validate</div>
        </div>
        <div class="status-card">
          <div class="label">Normalize endpoint</div>
          <div class="metric">/api/v1/ai/normalize</div>
        </div>
      </div>
    </section>

    <section class="workspace">
      <section class="panel">
        <div class="panel-head">
          <h2 class="panel-title">Domain YAML</h2>
          <div class="actions">
            <button class="ghost" id="load-service">AIService sample</button>
            <button class="ghost" id="load-job">AIJob sample</button>
            <button class="secondary" id="validate">Validate</button>
            <button id="normalize">Normalize</button>
          </div>
        </div>
        <textarea id="yaml" spellcheck="false"></textarea>
      </section>

      <section class="panel">
        <div class="panel-head">
          <h2 class="panel-title">Platform Preview</h2>
          <div class="label" id="last-action">ready</div>
        </div>
        <div class="results">
          <div class="result-grid" id="cards"></div>
          <div id="output" class="empty">Run validate or normalize to inspect governanceIntent and workloadIntent.</div>
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
    var lastAction = document.getElementById("last-action");
    yaml.value = serviceSample;

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
      lastAction.textContent = "error";
    }
    function card(label, value, mode) {
      var node = document.createElement("div");
      node.className = "intent-card " + (mode || "");
      node.innerHTML = "<div class=\"label\">" + label + "</div><strong>" + (value || "-") + "</strong>";
      return node;
    }
    function renderCards(data) {
      cards.innerHTML = "";
      cards.appendChild(card("kind", data.kind, "ok"));
      cards.appendChild(card("workload", data.workloadType, "ok"));
      cards.appendChild(card("tenant", data.governanceIntent && data.governanceIntent.tenant, ""));
      cards.appendChild(card("project", data.governanceIntent && data.governanceIntent.project, ""));
      cards.appendChild(card("runtime", data.runtime, ""));
      cards.appendChild(card("image", data.image, ""));
    }
    function post(path) {
      lastAction.textContent = "running";
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
    document.getElementById("load-service").onclick = function() {
      yaml.value = serviceSample;
      lastAction.textContent = "service sample";
    };
    document.getElementById("load-job").onclick = function() {
      yaml.value = jobSample;
      lastAction.textContent = "job sample";
    };
    document.getElementById("validate").onclick = function() {
      post("/api/v1/ai/validate").then(function(data) {
        cards.innerHTML = "";
        cards.appendChild(card("validation", "passed", "ok"));
        setJSON(data, "validated");
      }).catch(function(err) {
        cards.innerHTML = "";
        cards.appendChild(card("validation", "failed", "warn"));
        setError(err.message);
      });
    };
    document.getElementById("normalize").onclick = function() {
      post("/api/v1/ai/normalize").then(function(data) {
        renderCards(data);
        setJSON(data, "normalized");
      }).catch(function(err) {
        setError(err.message);
      });
    };
    fetch("/healthz").then(function(res) {
      return res.text();
    }).then(function(text) {
      document.getElementById("health").textContent = text.trim() || "ok";
    }).catch(function() {
      document.getElementById("health").textContent = "offline";
    });
  </script>
</body>
</html>`
