# AI Platform Native Definitions

This directory documents the first PoC boundary for the AI management platform extension. The implementation keeps KubeVela's native `Application` model as the runtime API and adds AI-specific meaning through ComponentDefinitions and TraitDefinitions.

The first phase is intentionally conservative:

- Keep the control plane on standard KubeVela `Application` resources.
- Express AI semantics through CUE definitions, not new controllers.
- Reuse native Kubernetes `Deployment`, `Job`, and `Service` resources.
- Reuse KubeVela native `topology`, `override`, and `deploy` policy/workflow behavior.
- Add only small Kubernetes-native pass-through hooks that are needed for production validation.

The first PoC intentionally does not introduce `AIService`, `AIJob`, or `AIWorkflow` CRDs. Product surfaces can still present those names, but the control plane stores and reconciles standard KubeVela `Application` resources.

## Minimal Domain Control Layer

The minimal domain-control improvement is an offline domain adapter for AI domain YAML. It lets a northbound surface accept `AIService` and `AIJob` shaped documents without installing new Kubernetes CRDs or adding a new reconciliation loop.

The supported domain inputs are:

- `docs/examples/ai-platform/domain/ai-service.yaml`
- `docs/examples/ai-platform/domain/ai-job.yaml`

The `translate` path maps:

- `kind: AIService` to a KubeVela component with `type: ai-service`
- `kind: AIJob` to a KubeVela component with `type: ai-job`
- `spec.runtime` to an `ai-runtime` trait
- `spec.placement` to native KubeVela `topology` policy and `deploy` workflow

The `normalize` path extracts a stable JSON contract from the same domain YAML:

```bash
ai-domain normalize -f docs/examples/ai-platform/domain/ai-service.yaml
ai-domain normalize -f docs/examples/ai-platform/domain/ai-job.yaml
```

The normalized output separates:

- identity fields: `kind`, `name`, `namespace`, `componentName`, `workloadType`, `image`, and `runtime`.
- `governanceIntent`: tenant, project, environment, owner, model or dataset URI, and placement.
- `workloadIntent`: service model and endpoint intent, or job kind, dataset, output, retry, and TTL intent.

This gives the domain-control layer a small semantic contract for later northbound APIs, policy checks, and governance projection while still keeping KubeVela `Application` as the runtime object.

This domain layer is intentionally not a runtime API server extension. The generated output is still `apiVersion: core.oam.dev/v1beta1`, `kind: Application`, so KubeVela remains the only reconciler for these workloads.

`AIWorkflow` remains out of first PoC scope and is rejected by the translator. Workflow-level orchestration should be added only after the service and job management boundaries are validated in a real cluster.

## Minimal Northbound API Layer

The minimal northbound API is a small HTTP adapter over the existing domain-control functions. It does not add storage, authentication, a database, a controller, or any write path to the cluster. Its first purpose is to provide a stable platform-facing API for validation and semantic normalization.

Start the API locally:

```bash
ai-northbound --addr 127.0.0.1:8088
```

Available endpoints:

- `GET /healthz` returns `ok`.
- `POST /api/v1/ai/validate` accepts an `AIService` or `AIJob` YAML document and returns validation JSON.
- `POST /api/v1/ai/normalize` accepts the same YAML document and returns the normalized identity, governance intent, and workload intent JSON.

Example:

```bash
curl -sS --data-binary @docs/examples/ai-platform/domain/ai-service.yaml \
  http://127.0.0.1:8088/api/v1/ai/normalize
```

This layer is intentionally read-only with respect to Kubernetes. Creation and update remain explicit through `ai-domain apply` until the northbound API contract is validated.

## Minimal Readonly Observe Layer

The minimal observe improvement is a readonly status summary in the `ai-domain` command. It does not install Prometheus, write to a database, add a controller, add a UI, or reconcile cluster state. It only reads existing KubeVela and Kubernetes objects and aggregates them into JSON.

For live cluster inspection:

```bash
ai-domain status -n sock-shop ai-service-domain-demo
ai-domain status -n sock-shop ai-job-domain-demo
```

The command reads:

- KubeVela `Application` status for phase, health, service message, and workload kind.
- Kubernetes `Deployment`, `Job`, `Service`, and `Pod` objects selected by `app.oam.dev/name`.
- AI metadata from labels and annotations with the `ai.oam.dev/` prefix.

For local or CI checks without a cluster, use fixture files:

```bash
ai-domain status --from-files app.yaml,deploy.yaml,svc.yaml,pod.yaml
```

The summary also surfaces a warning when a completed or failed Job Pod has empty `ownerReferences`, because this was observed during server validation and can leave completed Pods behind after Application deletion.

## Definitions

- `ai-service` describes long-running inference or model serving workloads.
- `ai-job` describes batch, training, evaluation, and offline inference jobs.
- `ai-runtime` attaches runtime, framework, tenant, project, environment, owner, model, and dataset metadata to pod templates.
- `gpu-resource` attaches GPU requests, node selectors, and tolerations to supported workloads.

## Phase 1 Improvements

The phase 1 changes focus on the minimum production hooks needed to validate AI workload management without changing KubeVela's control plane shape.

`ai-service` renders a Kubernetes `Deployment` and, when `endpoint.port` is configured, an optional `Service`. It supports the following production-oriented pass-through fields:

- Deployment rollout settings: `strategy`, `minReadySeconds`, `progressDeadlineSeconds`, and `revisionHistoryLimit`.
- Pod identity and scheduling: `serviceAccountName`, `automountServiceAccountToken`, `imagePullSecrets`, `priorityClassName`, `runtimeClassName`, and `schedulerName`.
- Runtime hardening: pod-level `securityContext`, container-level `containerSecurityContext`, and `terminationGracePeriodSeconds`.
- Container runtime configuration: `imagePullPolicy`, `cmd`, `args`, `env`, `env.valueFrom`, `envFrom`, `resources`, `volumes`, `volumeMounts`, and `lifecycle`.
- Service health and exposure integration: `readinessProbe`, `livenessProbe`, `startupProbe`, `endpoint.labels`, and `endpoint.annotations`.

`ai-job` renders a Kubernetes `Job`. It supports the following production-oriented pass-through fields:

- Job execution controls: `parallelism`, `completions`, `backoffLimit`, `activeDeadlineSeconds`, `completionMode`, `ttlSecondsAfterFinished`, `suspend`, and `restartPolicy`.
- Pod identity and scheduling: `serviceAccountName`, `automountServiceAccountToken`, `imagePullSecrets`, `priorityClassName`, `runtimeClassName`, and `schedulerName`.
- Runtime hardening: pod-level `securityContext`, container-level `containerSecurityContext`, and `terminationGracePeriodSeconds`.
- Container runtime configuration: `imagePullPolicy`, `cmd`, `args`, `env`, `env.valueFrom`, `envFrom`, `resources`, `volumes`, `volumeMounts`, and `lifecycle`.

`gpu-resource` patches `nvidia.com/gpu` requests and limits into supported workloads and optionally applies GPU node selectors and tolerations. It applies only to `deployments.apps` and `jobs.batch`.

`ai-runtime` only patches pod template labels and annotations for metadata such as runtime, framework, tenant, project, environment, owner, model URI, and dataset URI. It does not change scheduling, workflow execution, or reconciliation behavior.

## Examples

`ai-governance-app.yaml` shows the first management boundary for the PoC: reuse native `topology`, `override`, and `deploy` policies to place AI workloads and apply environment-specific changes without adding a new controller or workflow type.

`ai-service-app.yaml` also shows optional Kubernetes-native production hooks such as Deployment rollout settings, Service labels and annotations, readiness probes, ConfigMap-backed environment sources, and security contexts. These fields are pass-through parameters on the native component definition, not new platform controllers or custom workflow behavior.

`ai-job-app.yaml` shows the same principle for batch workloads: the example stays on the native Application API while demonstrating Job completion mode, TTL cleanup, `envFrom`, volume mounts, security contexts, and lifecycle hooks as pass-through parameters on the native component definition.

## Local Static Verification

These checks do not require a Kubernetes cluster. They validate the definition files, generated chart manifests, examples, and first-phase scope boundary.

```bash
env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1 -v
env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-domain -count=1 -v
env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-observe -count=1 -v
```

The test suite verifies that:

- The four AI definitions and three examples exist.
- The generated chart manifests match the CUE definitions rendered through KubeVela's definition parser.
- The definitions parse and format as valid CUE.
- The examples remain standard `core.oam.dev/v1beta1` `Application` resources.
- The examples use native `topology`, `override`, and `deploy` controls for governance.
- The first PoC scope does not add controller, database, or custom workflow-step surfaces.
- The domain examples translate into native KubeVela `Application` resources without adding CRDs or controllers.
- The observe summary reads existing Application, workload, Service, and Pod objects without adding a runtime control-plane component.

You can also run a direct scope scan:

```bash
rg -n "WorkflowStepDefinition|Controller|Reconcile|database|Database|SQL" \
  vela-templates/definitions/internal/component/ai-service.cue \
  vela-templates/definitions/internal/component/ai-job.cue \
  vela-templates/definitions/internal/trait/gpu-resource.cue \
  vela-templates/definitions/internal/trait/ai-runtime.cue \
  docs/examples/ai-platform/ai-service-app.yaml \
  docs/examples/ai-platform/ai-job-app.yaml \
  docs/examples/ai-platform/ai-governance-app.yaml \
  charts/vela-core/templates/defwithtemplate/ai-service.yaml \
  charts/vela-core/templates/defwithtemplate/ai-job.yaml \
  charts/vela-core/templates/defwithtemplate/gpu-resource.yaml \
  charts/vela-core/templates/defwithtemplate/ai-runtime.yaml
```

No matches are expected.

## End-to-End Verification

True end-to-end verification requires access to a Kubernetes API server because KubeVela must reconcile `Application` resources into Kubernetes workloads and read back workload status. Docker is not required if you test on a server or remote cluster that already has Kubernetes and a valid kubeconfig.

Recommended server-side flow:

```bash
kubectl config current-context
kubectl get nodes

vela def apply vela-templates/definitions/internal/component/ai-service.cue
vela def apply vela-templates/definitions/internal/component/ai-job.cue
vela def apply vela-templates/definitions/internal/trait/gpu-resource.cue
vela def apply vela-templates/definitions/internal/trait/ai-runtime.cue

kubectl apply -f docs/examples/ai-platform/ai-service-app.yaml
kubectl apply -f docs/examples/ai-platform/ai-job-app.yaml
kubectl apply -f docs/examples/ai-platform/ai-governance-app.yaml
```

After applying the examples, inspect both KubeVela and Kubernetes views:

```bash
vela status ai-service-demo
vela status ai-job-demo
vela status ai-governance-demo

kubectl get applications.core.oam.dev
kubectl get deploy,job,svc
kubectl get pods --show-labels
```

For minimal validation, the important result is not that the placeholder images become production-ready workloads. The important result is that KubeVela accepts the definitions, reconciles standard `Application` resources, produces native Kubernetes `Deployment`, `Job`, and `Service` objects, and preserves the AI management labels and annotations on generated resources.

## Boundary Notes

- Do not add `AIService`, `AIJob`, or `AIWorkflow` top-level CRDs in the first PoC.
- Do not add a new workflow step type for AI remediation in the first PoC.
- Do not add a database, storage subsystem, scheduler, rollout controller, or service mesh abstraction in the first PoC.
- Prefer native KubeVela policies and Kubernetes pass-through fields unless a later phase proves that a new control-plane surface is required.
