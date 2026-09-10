(function () {
  'use strict';

  const purposes = {
    AIJob: ['training', 'evaluation', 'batch', 'offline-inference'],
    AIService: ['service', 'agent']
  };
  const schedulingOptions = {
    priority: ['normal', 'high', 'urgent'],
    nodeType: ['cpu', 'gpu', 'any'],
    strategy: ['cost', 'performance', 'fast-start']
  };
  const dnsLabel = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;
  const labelValue = /^[a-zA-Z0-9](?:[a-zA-Z0-9_.-]{0,61}[a-zA-Z0-9])?$/;
  const text = value => typeof value === 'string' && value.trim().length > 0;
  const numericText = value => typeof value === 'number' || typeof value === 'string' ? String(value) : '';
  const decimal = value => /^(?:\d+(?:\.\d+)?|\.\d+)(?:[eE][+-]?\d+)?$/.test(numericText(value)) && Number.isFinite(Number(value));
  const integer = (value, min, max = 2147483647) => /^\d+$/.test(numericText(value)) && Number.isSafeInteger(Number(value)) && Number(value) >= min && Number(value) <= max;

  // Scope is supplied by the authenticated caller; this module cannot authorize it.
  function defaults(kind = 'AIJob', namespace = '', tenant = '') {
    if (!Object.hasOwn(purposes, kind)) throw new TypeError('Unsupported kind');
    const service = kind === 'AIService';
    return {
      kind, name: '', namespace, tenant, purpose: service ? 'service' : 'training',
      modelURI: service ? '' : 'inline://models/base',
      modelName: service ? 'example-model' : 'base', modelVersion: 'v1',
      datasetURI: service ? '' : 'inline://datasets/customer-sft-demo',
      image: service ? 'python:3.11-slim' : 'busybox:1.36', command: '',
      cpu: '1', memory: '1Gi', gpu: '0', replicas: '1', port: '8080',
      priority: 'normal', nodeType: 'cpu', strategy: 'performance',
      runtime: service ? 'http' : 'batch', epochs: '1', learningRate: '2e-5',
      threshold: '0.8', dryRun: true, example: true
    };
  }

  function exampleJob(draft) {
    return draft.kind === 'AIJob' && draft.example === true && ['training', 'evaluation'].includes(draft.purpose);
  }

  function exampleService(draft) {
    return draft.kind === 'AIService' && draft.purpose === 'service' && draft.example === true;
  }

  // Steps validate only their own fields. Summary (or omitted step) validates all.
  // No coercion is written back, so previous steps and inactive fields retain input.
  function validate(draft, step = 3) {
    const errors = [];
    const error = (field, message) => errors.push({ field, message });
    if (!draft || typeof draft !== 'object' || Array.isArray(draft)) return [{ field: 'draft', message: 'A draft is required.' }];
    if (!Number.isInteger(step) || step < 0 || step > 3) return [{ field: 'step', message: 'Step must be 0, 1, 2 or 3.' }];
    const d = draft;
    const service = d.kind === 'AIService';
    if (step === 0 || step === 3) {
      if (!Object.hasOwn(purposes, d.kind)) error('kind', 'Choose AIJob or AIService.');
      else if (!purposes[d.kind].includes(d.purpose)) error('purpose', 'Choose a purpose supported by this kind.');
    }
    if (step === 1 || step === 3) {
      for (const field of ['name', 'namespace']) {
        if (typeof d[field] !== 'string' || !dnsLabel.test(d[field])) error(field, 'Use 1-63 lowercase letters, digits or hyphens, starting and ending with a letter or digit.');
      }
      if (typeof d.tenant !== 'string' || !labelValue.test(d.tenant)) error('tenant', 'A valid tenant label from the current session is required.');
      if (!text(d.image) || /\s/.test(d.image)) error('image', 'A container image without whitespace is required.');
      if (!text(d.runtime) || !labelValue.test(d.runtime)) error('runtime', 'A valid runtime label is required.');
      if (typeof d.example !== 'boolean') error('example', 'Example must be a boolean.');
      if (d.purpose === 'agent' && d.example === true) error('example', 'An agent requires a custom container, not the example runtime.');
      const needsModel = (!service && ['training', 'evaluation', 'offline-inference'].includes(d.purpose)) || exampleService(d);
      if (needsModel && !text(d.modelURI)) error('modelURI', 'Select a model.');
      if (!service && ['training', 'evaluation', 'offline-inference'].includes(d.purpose) && !text(d.datasetURI)) error('datasetURI', 'Select a dataset.');
      if (service && text(d.modelURI) && !text(d.modelName)) error('modelName', 'A model name is required for the selected model.');
      if (service && text(d.modelName) && !labelValue.test(d.modelName)) error('modelName', 'The model name must be a valid Kubernetes label value.');
      for (const field of ['modelURI', 'modelName', 'modelVersion', 'datasetURI', 'command']) {
        if (d[field] !== undefined && typeof d[field] !== 'string') error(field, 'Use a string value.');
      }
      if (!exampleJob(d) && !exampleService(d) && !text(d.command)) error('command', 'A real container command is required.');
      if (exampleService(d) && d.runtime !== 'http') error('runtime', 'The example model service supports only the HTTP example runtime.');
      if (!service && d.purpose === 'training') {
        if (!integer(d.epochs, 1)) error('epochs', 'Epochs must be a positive integer.');
        if (!decimal(d.learningRate) || Number(d.learningRate) <= 0) error('learningRate', 'Learning rate must be a positive finite number.');
      }
      if (!service && d.purpose === 'evaluation' && (!decimal(d.threshold) || Number(d.threshold) < 0 || Number(d.threshold) > 1)) error('threshold', 'Threshold must be between 0 and 1.');
    }
    if (step === 2 || step === 3) {
      for (const [field, allowed] of Object.entries(schedulingOptions)) {
        if (d[field] !== undefined && d[field] !== '' && !allowed.includes(d[field])) error(field, 'Choose ' + allowed.join(', ') + '.');
      }
      const cpu = numericText(d.cpu);
      const cores = cpu.endsWith('m') ? Number(cpu.slice(0, -1)) / 1000 : Number(cpu);
      if (!/^(?:\d+m|\d+(?:\.\d{1,3})?|\.\d{1,3})$/.test(cpu) || !Number.isFinite(cores) || cores <= 0) error('cpu', 'CPU must be positive cores (up to 3 decimal places) or whole millicores.');
      const memory = numericText(d.memory).match(/^(\d+(?:\.\d+)?|\.\d+)([KMGTPE]i|[kMGTPE])?$/);
      if (!memory || !Number.isFinite(Number(memory[1])) || Number(memory[1]) <= 0) error('memory', 'Memory must be positive bytes or a quantity such as 512Mi or 1Gi.');
      if (!integer(d.gpu, 0)) error('gpu', 'GPU count must be a nonnegative integer.');
      if (service) {
        if (!integer(d.replicas, 1)) error('replicas', 'Replicas must be a positive integer.');
        if (!integer(d.port, 1, 65535)) error('port', 'Port must be an integer from 1 to 65535.');
      }
    }
    if (step === 3 && typeof d.dryRun !== 'boolean') error('dryRun', 'Dry run must be a boolean.');
    return errors;
  }

  function demoJobCommand(purpose) {
    const lines = [
      'echo "EXAMPLE: fixed demo results; no real training or evaluation is performed"',
      'printf "base_model=%s\\ndataset=%s\\n" "$MODEL_URI" "$INPUT_DATASET_URI"'
    ];
    if (purpose === 'evaluation') {
      lines.push('printf "evaluation_threshold=%s\\n" "$EVALUATION_THRESHOLD"', 'echo eval-start');
    } else {
      lines.push('printf "epochs=%s learning_rate=%s\\n" "$EPOCHS" "$LEARNING_RATE"', 'echo train-start', 'echo epoch=1 loss=0.30', 'echo epoch=2 loss=0.12');
    }
    lines.push('printf "AI_RESULT_JSON=%s\\n" "$AI_EXAMPLE_RESULT"', purpose === 'evaluation' ? 'echo eval-complete' : 'echo train-complete');
    return lines.join('\n');
  }

  function generate(draft) {
    const errors = validate(draft);
    if (errors.length) {
      const error = new TypeError('Invalid draft: ' + errors.map(item => item.field + ': ' + item.message).join('; '));
      error.errors = errors;
      throw error;
    }
    const d = draft;
    const service = d.kind === 'AIService';
    const example = exampleJob(d) || exampleService(d);
    const env = [];
    const addEnv = (name, value) => {
      if (value !== undefined && value !== '') env.push({ name, value: String(value) });
    };
    const properties = { image: d.image, imagePullPolicy: 'IfNotPresent' };
    const runtime = { runtime: d.runtime, tenant: d.tenant, project: d.purpose };
    if (example) runtime.framework = 'demo';
    if (text(d.modelURI)) runtime.modelURI = d.modelURI;
    if (!service && text(d.datasetURI)) runtime.datasetURI = d.datasetURI;
    addEnv('PORT', service ? d.port : undefined);
    // ai-service injects MODEL_URI itself; ai-job does not.
    if (!service) addEnv('MODEL_URI', d.modelURI);
    if (!service) addEnv('INPUT_DATASET_URI', d.datasetURI);
    let command = d.command;
    if (service) {
      properties.replicas = Number(d.replicas);
      properties.model = { name: text(d.modelName) ? d.modelName : d.name };
      if (text(d.modelVersion)) properties.model.version = d.modelVersion;
      if (text(d.modelURI)) properties.model.uri = d.modelURI;
      properties.endpoint = { port: Number(d.port), servicePort: 80, type: 'ClusterIP' };
      if (exampleService(d) && !text(command)) {
        command = 'mkdir -p /tmp/console-example\nprintf "%s\\n" "EXAMPLE HTTP service; no model inference is performed" > /tmp/console-example/index.html\nprintf "%s\\n" "EXAMPLE health check OK; no model inference is performed" > /tmp/console-example/healthz\nexec python -m http.server "$PORT" --bind 0.0.0.0 --directory /tmp/console-example';
      }
    } else {
      properties.jobKind = d.purpose;
      properties.backoffLimit = 0;
      properties.ttlSecondsAfterFinished = 3600;
      if (text(d.datasetURI)) properties.dataset = { name: 'dataset', uri: d.datasetURI };
      if (d.purpose === 'training') {
        addEnv('EPOCHS', d.epochs);
        addEnv('LEARNING_RATE', d.learningRate);
      }
      if (d.purpose === 'evaluation') addEnv('EVALUATION_THRESHOLD', d.threshold);
      if (exampleJob(d) && !text(command)) {
        properties.output = { uri: 'inline://outputs/' + d.name };
        addEnv('AI_EXAMPLE_RESULT', JSON.stringify({
          modelURI: d.purpose === 'training' ? 'inline://models/' + d.name + '/v1' : d.modelURI,
          metrics: { loss: 0.12, accuracy: 0.98 },
          summary: 'example-' + d.purpose + '-fixed-results', example: true
        }));
        command = demoJobCommand(d.purpose);
      }
    }
    properties.cmd = ['sh', '-c'];
    properties.args = [command];
    properties.env = env;
    properties.annotations = { 'ai.oam.dev/purpose': d.purpose, 'ai.oam.dev/example': String(example) };
    const document = {
      apiVersion: 'ai.oam.dev/v1alpha1', kind: d.kind,
      metadata: { name: d.name, namespace: d.namespace },
      spec: {
        componentName: d.name, properties,
        resources: { cpu: String(d.cpu), memory: String(d.memory), gpu: String(Number(d.gpu)) },
        runtime
      }
    };
    const scheduling = {};
    for (const field of Object.keys(schedulingOptions)) {
      if (d[field] !== undefined && d[field] !== '') scheduling[field] = d[field];
    }
    if (Object.keys(scheduling).length) document.spec.scheduling = scheduling;
    // JSON is a YAML subset accepted by the domain translator. Escape Unicode
    // line separators too, which older YAML parsers can interpret as line breaks.
    return JSON.stringify(document, null, 2).replace(/\u0085/g, '\\u0085').replace(/\u2028/g, '\\u2028').replace(/\u2029/g, '\\u2029') + '\n';
  }

  globalThis.ConsoleCreate = { defaults, validate, generate };
})();
