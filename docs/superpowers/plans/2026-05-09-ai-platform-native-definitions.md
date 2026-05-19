# AI Platform Native Definitions Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first AI platform PoC on KubeVela's native Application, Definition, Policy, and Workflow extension model instead of introducing `AIService`, `AIJob`, or `AIWorkflow` top-level CRDs.

**Architecture:** AI workload semantics are expressed as ComponentDefinitions and TraitDefinitions. End users still submit standard KubeVela `Application` resources, so existing revision, rollback, topology, override, workflow, and resource tracking behavior remains reusable.

**Tech Stack:** KubeVela `Application`, CUE-based `ComponentDefinition` and `TraitDefinition`, Kubernetes `Deployment`, Kubernetes `Job`, Go tests, YAML examples.

---

## Chunk 1: Native AI Workload Definitions

### Task 1: Add AI workload components

**Files:**
- Create: `vela-templates/definitions/internal/component/ai-service.cue`
- Create: `vela-templates/definitions/internal/component/ai-job.cue`
- Test: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write failing tests for expected AI definitions**

The test requires `ai-service` and `ai-job` to exist as native KubeVela component definitions.

- [x] **Step 2: Verify the tests fail before implementation**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions`

Expected: FAIL because the AI definition files do not exist.

- [x] **Step 3: Add minimal component definitions**

`ai-service` renders a Kubernetes `Deployment` and optional `Service`.

`ai-job` renders a Kubernetes `Job`.

Both stay inside `Application.spec.components[*].type` instead of introducing new CRDs.

- [x] **Step 4: Verify the tests pass**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions`

Expected: PASS.

### Task 2: Add GPU trait

**Files:**
- Create: `vela-templates/definitions/internal/trait/gpu-resource.cue`
- Test: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write failing tests for GPU trait**

The test requires `gpu-resource` to exist as a native KubeVela trait and apply to `Deployment` and `Job` workloads.

- [x] **Step 2: Add minimal GPU trait**

The trait patches `nvidia.com/gpu` requests and limits into pod template containers, with optional node selectors and tolerations.

- [x] **Step 3: Verify the tests pass**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions`

Expected: PASS.

### Task 3: Add Application examples

**Files:**
- Create: `docs/examples/ai-platform/ai-service-app.yaml`
- Create: `docs/examples/ai-platform/ai-job-app.yaml`
- Create: `docs/examples/ai-platform/README.md`
- Test: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write tests that examples remain standard Applications**

The test rejects `kind: AIService`, `kind: AIJob`, and `kind: AIWorkflow`.

- [x] **Step 2: Add examples**

Examples use `apiVersion: core.oam.dev/v1beta1`, `kind: Application`, `type: ai-service`, `type: ai-job`, and `type: gpu-resource`.

- [x] **Step 3: Verify example YAML parses**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions`

Expected: PASS.

## Chunk 2: Minimal Management Metadata

### Task 4: Add AI runtime metadata trait

**Files:**
- Create: `vela-templates/definitions/internal/trait/ai-runtime.cue`
- Create: `charts/vela-core/templates/defwithtemplate/ai-runtime.yaml`
- Modify: `docs/examples/ai-platform/ai-service-app.yaml`
- Modify: `docs/examples/ai-platform/ai-job-app.yaml`
- Modify: `docs/examples/ai-platform/README.md`
- Test: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write failing tests for runtime metadata**

The test requires an `ai-runtime` trait and requires examples to use it while staying standard KubeVela `Application` resources.

- [x] **Step 2: Verify the tests fail before implementation**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions`

Expected: FAIL because `ai-runtime.cue` does not exist and examples do not reference `type: ai-runtime`.

- [x] **Step 3: Add minimal runtime trait**

The trait only patches pod template labels and annotations for runtime, framework, tenant, project, environment, model URI, dataset URI, and owner metadata. It does not change scheduling, workflow execution, or controller behavior.

- [x] **Step 4: Verify tests and scope**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions`

Expected: PASS.

Run: `rg -n "kind: AIService|kind: AIJob|kind: AIWorkflow|type AIService|type AIJob|type AIWorkflow|Controller|Reconcile" vela-templates/definitions docs/examples/ai-platform test/ai-definitions`

Expected: no implementation matches outside the guard assertion in `test/ai-definitions/ai_definitions_test.go`.

- [x] **Step 5: Add minimal ownership metadata for management boundaries**

The `ai-runtime` trait supports optional `tenant`, `project`, and `environment` fields. These are emitted as pod template labels only, so they can support audit, filtering, and later policy integration without changing the control plane shape.

## Chunk 3: Definition Validation

### Task 5: Validate AI definition syntax with Go tests

**Files:**
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Add CUE parser and formatter checks**

The test uses the CUE Go parser and formatter on each AI definition file. This keeps validation local and stable without depending on a cluster or the external `cue` binary.

- [x] **Step 2: Run the validation test**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions`

Expected: PASS.

## Chunk 4: Helm Chart Manifests

### Task 6: Generate installable chart manifests

**Files:**
- Modify: `test/ai-definitions/ai_definitions_test.go`
- Create: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Create: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Create: `charts/vela-core/templates/defwithtemplate/gpu-resource.yaml`
- Create: `charts/vela-core/templates/defwithtemplate/ai-runtime.yaml`

- [x] **Step 1: Attempt generation through the local CLI path**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go run ./references/cmd/cli def render vela-templates/definitions/internal/trait/ai-runtime.cue --format yaml -o /tmp/ai-runtime.yaml`

Observed: command exits with status 1 in the current environment and does not emit actionable stderr. Root cause investigation showed `FromCUEString` initializes the default CUE compiler, which attempts to load external CUE packages through `singleton.DynamicClient`; without kubeconfig this path can call `config.GetConfigOrDie()` and exit the process. The external `cue` binary is also unavailable.

- [x] **Step 2: Add parser-level validation that does not require kubeconfig**

The test disables external CUE package loading for this validation path and verifies all four AI definitions render through `pkg/definition.FromCUEString`. This is valid for these definitions because they do not import external CUE providers.

- [x] **Step 3: Generate chart manifests from the same parser path**

The four chart manifests are generated from the CUE definitions with the Helm namespace placeholder `{{ include "systemDefinitionNamespace" . }}`.

- [x] **Step 4: Replace the missing-manifest guard with content validation**

The test now requires each chart manifest to exactly match the YAML produced from the corresponding CUE definition by the KubeVela definition parser. This prevents stale generated manifests.

- [x] **Step 5: Verify generated manifests and existing definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions`

Expected: PASS.

## Chunk 5: Native Policy Governance Example

### Task 7: Add minimal topology and override governance example

**Files:**
- Create: `docs/examples/ai-platform/ai-governance-app.yaml`
- Modify: `docs/examples/ai-platform/README.md`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for native policy governance**

The test requires a standard `Application` example that uses `ai-service`, `ai-job`, native `topology`, native `override`, and native `deploy` workflow steps. The test rejects top-level AI CRDs.

- [x] **Step 2: Add the governance example**

`ai-governance-app.yaml` demonstrates environment placement and environment-specific changes with `topology + override + deploy`, without adding a new controller, CRD, or workflow type.

- [x] **Step 3: Verify the policy governance example**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -run TestAIGovernanceExampleUsesNativePolicyControls -count=1 -v`

Expected: PASS.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 6: Minimal Runtime Parameter Guards

### Task 8: Add low-risk numeric constraints to AI definitions

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `vela-templates/definitions/internal/component/ai-job.cue`
- Modify: `vela-templates/definitions/internal/trait/gpu-resource.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Modify: `charts/vela-core/templates/defwithtemplate/gpu-resource.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write failing tests for unsafe numeric defaults**

The test requires `ai-service.replicas`, `ai-job.parallelism`, `ai-job.completions`, `ai-job.activeDeadlineSeconds`, and `gpu-resource.count` to be positive integers. It also allows `ai-job.backoffLimit` to be zero or greater.

- [x] **Step 2: Add minimal CUE constraints**

The definitions constrain only AI-specific numeric parameters. No upstream generic component, trait, controller, workflow, or CRD is changed.

- [x] **Step 3: Regenerate chart manifests**

The generated chart manifests for `ai-service`, `ai-job`, and `gpu-resource` are updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1 -v`

Expected: PASS.

## Chunk 7: Management Labels On Workload Resources

### Task 9: Add top-level labels for query, audit, and policy selection

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `vela-templates/definitions/internal/component/ai-job.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for top-level management labels**

The test requires AI management labels to be present on top-level Kubernetes workload resources, not only on pod templates.

- [x] **Step 2: Add shared AI management label blocks**

`ai-service` adds labels to the generated `Deployment` and optional `Service`. `ai-job` adds labels to the generated `Job`. The existing pod template labels continue to include the same management identity.

- [x] **Step 3: Regenerate chart manifests**

The generated chart manifests for `ai-service` and `ai-job` are updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 8: Minimal Image Pull Governance

### Task 10: Add optional image pull policy to AI components

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `vela-templates/definitions/internal/component/ai-job.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for image pull policy support**

The test requires `ai-service` and `ai-job` to expose an optional `imagePullPolicy` parameter with Kubernetes-supported values.

- [x] **Step 2: Add optional container image pull policy**

Both components pass `imagePullPolicy` into the generated container only when the parameter is set. Default behavior is unchanged.

- [x] **Step 3: Regenerate chart manifests**

The generated chart manifests for `ai-service` and `ai-job` are updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 9: Minimal Runtime Identity Hook

### Task 11: Add optional service account to AI components

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `vela-templates/definitions/internal/component/ai-job.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for pod service account support**

The test requires `ai-service` and `ai-job` to expose an optional `serviceAccountName` parameter.

- [x] **Step 2: Add optional pod service account wiring**

Both components pass `serviceAccountName` into the generated pod spec only when the parameter is set. Default behavior is unchanged.

- [x] **Step 3: Regenerate chart manifests**

The generated chart manifests for `ai-service` and `ai-job` are updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 10: Minimal Private Registry Hook

### Task 12: Add optional image pull secrets to AI components

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `vela-templates/definitions/internal/component/ai-job.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for image pull secret support**

The test requires `ai-service` and `ai-job` to expose an optional `imagePullSecrets` parameter.

- [x] **Step 2: Add optional pod image pull secrets**

Both components pass `imagePullSecrets` into the generated pod spec only when the parameter is set. Default behavior is unchanged.

- [x] **Step 3: Regenerate chart manifests**

The generated chart manifests for `ai-service` and `ai-job` are updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 11: Minimal Job Cleanup Hook

### Task 13: Add optional TTL cleanup to AI jobs

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-job.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for finished Job cleanup**

The test requires `ai-job` to expose an optional `ttlSecondsAfterFinished` parameter constrained to zero or greater.

- [x] **Step 2: Add optional Kubernetes Job TTL wiring**

`ai-job` passes `ttlSecondsAfterFinished` into the generated Job spec only when the parameter is set. Default behavior is unchanged.

- [x] **Step 3: Regenerate chart manifest**

The generated chart manifest for `ai-job` is updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 12: Minimal Job Start Control Hook

### Task 14: Add optional suspend control to AI jobs

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-job.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for suspended Job creation**

The test requires `ai-job` to expose an optional `suspend` parameter.

- [x] **Step 2: Add optional Kubernetes Job suspend wiring**

`ai-job` passes `suspend` into the generated Job spec only when the parameter is set. Default behavior is unchanged.

- [x] **Step 3: Regenerate chart manifest**

The generated chart manifest for `ai-job` is updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 13: Minimal Priority Scheduling Hook

### Task 15: Add optional priority class to AI components

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `vela-templates/definitions/internal/component/ai-job.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for pod priority class support**

The test requires `ai-service` and `ai-job` to expose an optional `priorityClassName` parameter.

- [x] **Step 2: Add optional pod priority class wiring**

Both components pass `priorityClassName` into the generated pod spec only when the parameter is set. Default behavior is unchanged.

- [x] **Step 3: Regenerate chart manifests**

The generated chart manifests for `ai-service` and `ai-job` are updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 14: Minimal Runtime Class Hook

### Task 16: Add optional runtime class to AI components

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `vela-templates/definitions/internal/component/ai-job.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for pod runtime class support**

The test requires `ai-service` and `ai-job` to expose an optional `runtimeClassName` parameter.

- [x] **Step 2: Add optional pod runtime class wiring**

Both components pass `runtimeClassName` into the generated pod spec only when the parameter is set. Default behavior is unchanged.

- [x] **Step 3: Regenerate chart manifests**

The generated chart manifests for `ai-service` and `ai-job` are updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 15: Minimal External Scheduler Hook

### Task 17: Add optional scheduler name to AI components

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `vela-templates/definitions/internal/component/ai-job.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for pod scheduler name support**

The test requires `ai-service` and `ai-job` to expose an optional `schedulerName` parameter.

- [x] **Step 2: Add optional pod scheduler name wiring**

Both components pass `schedulerName` into the generated pod spec only when the parameter is set. This preserves the default Kubernetes scheduler behavior unless explicitly overridden.

- [x] **Step 3: Regenerate chart manifests**

The generated chart manifests for `ai-service` and `ai-job` are updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 16: Minimal Secret-backed Environment Hook

### Task 18: Allow native Kubernetes env.valueFrom in AI components

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `vela-templates/definitions/internal/component/ai-job.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for `env.valueFrom` support**

The test requires `ai-service` and `ai-job` to expose Kubernetes-native `valueFrom` on custom container environment variables.

- [x] **Step 2: Add minimal env schema support**

Both components keep the existing environment passthrough behavior, make `value` optional, and allow open `valueFrom` objects. This supports Secret and ConfigMap references without modeling new platform storage, database, controller, or workflow behavior.

- [x] **Step 3: Regenerate chart manifests**

The generated chart manifests for `ai-service` and `ai-job` are updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 17: Minimal Service Health Probe Hook

### Task 19: Allow native Kubernetes health probes on AI services

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for service health probe support**

The test requires `ai-service` to expose and pass through `readinessProbe`, `livenessProbe`, and `startupProbe`, while keeping `ai-job` free from service-oriented health probe surface.

- [x] **Step 2: Add minimal probe passthrough**

`ai-service` passes optional Kubernetes-native probe objects directly to the generated container only when set. This supports service health observation without introducing a custom anomaly engine, controller, workflow step, or AI-specific probe model.

- [x] **Step 3: Regenerate chart manifest**

The generated chart manifest for `ai-service` is updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 18: Minimal Native Volume Hook

### Task 20: Allow native Kubernetes volumes and volume mounts on AI components

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `vela-templates/definitions/internal/component/ai-job.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for native volume support**

The test requires `ai-service` and `ai-job` to expose and pass through Kubernetes-native `volumes` and `volumeMounts`.

- [x] **Step 2: Add minimal volume passthrough**

Both components pass optional pod `volumes` and container `volumeMounts` only when set. This supports model files, datasets, config, and secret mounts without introducing a storage subsystem, database, controller, or AI-specific volume abstraction.

- [x] **Step 3: Regenerate chart manifests**

The generated chart manifests for `ai-service` and `ai-job` are updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 19: Minimal Native Security Context Hook

### Task 21: Allow native Kubernetes pod and container security contexts on AI components

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `vela-templates/definitions/internal/component/ai-job.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for native security context support**

The test requires `ai-service` and `ai-job` to expose pod-level `securityContext` and container-level `containerSecurityContext`, and verifies they are written to the correct Kubernetes resource levels.

- [x] **Step 2: Add minimal security context passthrough**

Both components pass optional pod and container security contexts only when set. This enables non-root execution, read-only root filesystems, capability controls, and similar baseline security hardening without imposing defaults or introducing a new policy engine.

- [x] **Step 3: Regenerate chart manifests**

The generated chart manifests for `ai-service` and `ai-job` are updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 20: Minimal Native Termination Grace Hook

### Task 22: Allow native Kubernetes termination grace period on AI components

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `vela-templates/definitions/internal/component/ai-job.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for native termination grace period support**

The test requires `ai-service` and `ai-job` to expose optional `terminationGracePeriodSeconds` and constrain it to a non-negative integer.

- [x] **Step 2: Add minimal pod termination grace passthrough**

Both components pass optional `terminationGracePeriodSeconds` to the generated pod spec only when set. This supports graceful model server shutdown and task cleanup without adding lifecycle controllers or platform-specific policy behavior.

- [x] **Step 3: Regenerate chart manifests**

The generated chart manifests for `ai-service` and `ai-job` are updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 21: Minimal Native EnvFrom Hook

### Task 23: Allow native Kubernetes envFrom on AI components

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `vela-templates/definitions/internal/component/ai-job.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for native envFrom support**

The test requires `ai-service` and `ai-job` to expose optional Kubernetes-native `envFrom` and pass it to the generated container.

- [x] **Step 2: Add minimal envFrom passthrough**

Both components pass optional `envFrom` only when set. This supports ConfigMap and Secret backed environment variable sets without introducing a configuration service, database, or controller.

- [x] **Step 3: Regenerate chart manifests**

The generated chart manifests for `ai-service` and `ai-job` are updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 22: Minimal Service Account Token Mount Control

### Task 24: Allow native Kubernetes service account token automount control on AI components

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `vela-templates/definitions/internal/component/ai-job.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for service account token mount control**

The test requires `ai-service` and `ai-job` to expose optional Kubernetes-native `automountServiceAccountToken` and pass it to the generated pod spec.

- [x] **Step 2: Add minimal automount passthrough**

Both components pass optional `automountServiceAccountToken` only when set. This supports least-privilege runtime hardening without imposing defaults or introducing a policy engine.

- [x] **Step 3: Regenerate chart manifests**

The generated chart manifests for `ai-service` and `ai-job` are updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 23: Minimal Native Container Lifecycle Hook

### Task 25: Allow native Kubernetes container lifecycle hooks on AI components

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `vela-templates/definitions/internal/component/ai-job.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for native container lifecycle support**

The test requires `ai-service` and `ai-job` to expose optional Kubernetes-native `lifecycle` and pass it to the generated container.

- [x] **Step 2: Add minimal lifecycle passthrough**

Both components pass optional container `lifecycle` only when set. This supports native `preStop` and `postStart` hooks for graceful model server draining and task cleanup without adding platform-specific lifecycle controllers.

- [x] **Step 3: Regenerate chart manifests**

The generated chart manifests for `ai-service` and `ai-job` are updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 24: Minimal Native Deployment Strategy Hook

### Task 26: Allow native Kubernetes Deployment strategy on AI services

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for native Deployment strategy support**

The test requires `ai-service` to expose optional Kubernetes-native `strategy` and pass it to the generated Deployment spec, while keeping `ai-job` free from Deployment strategy surface.

- [x] **Step 2: Add minimal strategy passthrough**

`ai-service` passes optional Deployment `strategy` only when set. This supports controlled rolling update parameters without introducing rollout controllers, service meshes, or platform-specific release logic.

- [x] **Step 3: Regenerate chart manifest**

The generated chart manifest for `ai-service` is updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 25: Minimal Native Min Ready Seconds Hook

### Task 27: Allow native Kubernetes minReadySeconds on AI services

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for native minReadySeconds support**

The test requires `ai-service` to expose optional Kubernetes-native `minReadySeconds` and pass it to the generated Deployment spec, while keeping `ai-job` free from Deployment availability surface.

- [x] **Step 2: Add minimal minReadySeconds passthrough**

`ai-service` passes optional Deployment `minReadySeconds` only when set. This supports conservative service availability and rollout control without introducing rollout controllers or service meshes.

- [x] **Step 3: Regenerate chart manifest**

The generated chart manifest for `ai-service` is updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 26: Minimal Native Progress Deadline Hook

### Task 28: Allow native Kubernetes progressDeadlineSeconds on AI services

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for native progressDeadlineSeconds support**

The test requires `ai-service` to expose optional Kubernetes-native `progressDeadlineSeconds` and pass it to the generated Deployment spec, while keeping `ai-job` free from Deployment rollout surface.

- [x] **Step 2: Add minimal progress deadline passthrough**

`ai-service` passes optional Deployment `progressDeadlineSeconds` only when set and constrains it to a positive integer. This lets Kubernetes detect rollout progress stalls without adding platform-specific rollout controllers.

- [x] **Step 3: Regenerate chart manifest**

The generated chart manifest for `ai-service` is updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 27: Minimal Native Revision History Hook

### Task 29: Allow native Kubernetes revisionHistoryLimit on AI services

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for native revisionHistoryLimit support**

The test requires `ai-service` to expose optional Kubernetes-native `revisionHistoryLimit` and pass it to the generated Deployment spec, while keeping `ai-job` free from Deployment history surface.

- [x] **Step 2: Add minimal revision history passthrough**

`ai-service` passes optional Deployment `revisionHistoryLimit` only when set and constrains it to a non-negative integer. This controls Kubernetes ReplicaSet history without changing KubeVela ApplicationRevision behavior.

- [x] **Step 3: Regenerate chart manifest**

The generated chart manifest for `ai-service` is updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 28: Minimal Native Job Completion Mode Hook

### Task 30: Allow native Kubernetes completionMode on AI jobs

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-job.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-job.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for native completionMode support**

The test requires `ai-job` to expose optional Kubernetes-native `completionMode` and pass it to the generated Job spec, while keeping `ai-service` free from Job completion surface.

- [x] **Step 2: Add minimal completionMode passthrough**

`ai-job` passes optional Job `completionMode` only when set and constrains it to Kubernetes-supported `NonIndexed` or `Indexed` values. This supports indexed batch jobs without introducing schedulers or workflow engines.

- [x] **Step 3: Regenerate chart manifest**

The generated chart manifest for `ai-job` is updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 29: Minimal Native Service Annotation Hook

### Task 31: Allow native Kubernetes Service annotations on AI services

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for native Service annotation support**

The test requires `ai-service` to expose optional `endpoint.annotations` and pass it to the generated Kubernetes Service metadata, while keeping `ai-job` free from Service endpoint annotation surface.

- [x] **Step 2: Add minimal Service annotation passthrough**

`ai-service` passes optional Service annotations only when the Service is generated. This supports cloud load balancer, discovery, and monitoring integrations without introducing Ingress, Gateway API, or service mesh resources.

- [x] **Step 3: Regenerate chart manifest**

The generated chart manifest for `ai-service` is updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Chunk 30: Minimal Native Service Label Hook

### Task 32: Allow native Kubernetes Service labels on AI services

**Files:**
- Modify: `vela-templates/definitions/internal/component/ai-service.cue`
- Modify: `charts/vela-core/templates/defwithtemplate/ai-service.yaml`
- Modify: `test/ai-definitions/ai_definitions_test.go`

- [x] **Step 1: Write a failing test for native Service label support**

The test requires `ai-service` to expose optional `endpoint.labels` and merge it into the generated Kubernetes Service metadata labels, while keeping `ai-job` free from Service endpoint label surface.

- [x] **Step 2: Add minimal Service label passthrough**

`ai-service` merges optional Service labels with the built-in AI management labels only when the Service is generated. This supports service discovery, monitoring selection, and platform filtering without changing pod labels or introducing new network resources.

- [x] **Step 3: Regenerate chart manifest**

The generated chart manifest for `ai-service` is updated from the same CUE parser path used by the tests.

- [x] **Step 4: Verify all AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

## Remaining Work

- Validate these definitions against a live KubeVela control plane when a kubeconfig-backed local cluster is available.
- Keep AI remediation workflow steps out of the first PoC unless the phase boundary changes.
