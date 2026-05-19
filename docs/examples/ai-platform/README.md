# AI Platform Native Definitions

This directory documents the first PoC boundary for the AI management platform extension. The implementation keeps KubeVela's native `Application` model as the runtime API and adds AI-specific meaning through ComponentDefinitions and TraitDefinitions.

The first phase is intentionally conservative:

- Keep the control plane on standard KubeVela `Application` resources.
- Express AI semantics through CUE definitions, not new controllers.
- Reuse native Kubernetes `Deployment`, `Job`, and `Service` resources.
- Reuse KubeVela native `topology`, `override`, and `deploy` policy/workflow behavior.
- Add only small Kubernetes-native pass-through hooks that are needed for production validation.

The first PoC intentionally does not introduce `AIService`, `AIJob`, or `AIWorkflow` CRDs. Product surfaces can still present those names, but the control plane stores and reconciles standard KubeVela `Application` resources.

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
```

The test suite verifies that:

- The four AI definitions and three examples exist.
- The generated chart manifests match the CUE definitions rendered through KubeVela's definition parser.
- The definitions parse and format as valid CUE.
- The examples remain standard `core.oam.dev/v1beta1` `Application` resources.
- The examples use native `topology`, `override`, and `deploy` controls for governance.
- The first PoC scope does not add controller, database, or custom workflow-step surfaces.

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
