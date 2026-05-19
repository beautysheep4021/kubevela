# AI Domain Translator Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the minimal domain-control layer for the AI platform PoC by translating AI domain YAML into native KubeVela `Application` YAML.

**Architecture:** The domain layer is an offline translator, not a Kubernetes CRD, controller, database, or reconciliation loop. `AIService` and `AIJob` YAML documents map to existing `ai-service` and `ai-job` ComponentDefinitions, while placement and runtime metadata map to native KubeVela policy, workflow, and trait constructs.

**Tech Stack:** Go, Kubernetes `unstructured` objects, YAML/JSON marshaling, KubeVela `Application`, existing `ai-service`, `ai-job`, and `ai-runtime` definitions.

---

## Chunk 1: Domain Translator Core

### Task 1: Add failing translator tests

**Files:**
- Create: `test/ai-domain/ai_domain_test.go`
- Create: `pkg/ai/domain/translator.go`

- [x] **Step 1: Write a failing test for AIService translation**

The test should call `domain.TranslateYAML` with an `apiVersion: ai.oam.dev/v1alpha1`, `kind: AIService` document and assert the generated YAML is a standard KubeVela `Application`.

Required assertions:
- `apiVersion` is `core.oam.dev/v1beta1`
- `kind` is `Application`
- component `type` is `ai-service`
- component properties preserve `image`, `replicas`, and `model`
- an `ai-runtime` trait is attached when `spec.runtime` is provided
- `topology` policy and `deploy` workflow are generated from `spec.placement`

- [x] **Step 2: Write a failing test for AIJob translation**

The test should call `domain.TranslateYAML` with `kind: AIJob` and assert:
- component `type` is `ai-job`
- component properties preserve `jobKind`, `image`, `dataset`, and `output`
- runtime metadata is attached as an `ai-runtime` trait
- placement is represented through native KubeVela policy/workflow

- [x] **Step 3: Write a failing test for unsupported AIWorkflow**

The translator should reject `kind: AIWorkflow` with a clear unsupported-kind error. This keeps first-phase scope aligned with the decision not to include AI workflow in the first PoC.

- [x] **Step 4: Run tests and verify RED**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-domain -count=1`

Expected: FAIL because `pkg/ai/domain` does not exist yet.

### Task 2: Implement the minimal translator

**Files:**
- Create: `pkg/ai/domain/translator.go`
- Test: `test/ai-domain/ai_domain_test.go`

- [x] **Step 1: Add public translation API**

Implement:

```go
func TranslateYAML(in []byte) ([]byte, error)
```

The function parses one YAML document and emits one KubeVela `Application` YAML document.

- [x] **Step 2: Define the minimal domain schema**

Support:
- `apiVersion`
- `kind`
- `metadata.name`
- `metadata.namespace`
- `spec.componentName`
- `spec.properties`
- `spec.runtime`
- `spec.placement.namespace`
- `spec.placement.clusters`

Unknown fields can pass through inside `spec.properties`; unknown top-level kinds should fail.

- [x] **Step 3: Map domain kinds to native component types**

Map:
- `AIService` -> `ai-service`
- `AIJob` -> `ai-job`

Reject:
- `AIWorkflow`
- any other kind

- [x] **Step 4: Generate native KubeVela composition**

Generate:
- one component with `name`, `type`, `properties`, and optional `traits`
- one `topology` policy when placement exists
- one `deploy` workflow step referencing that topology policy when placement exists

- [x] **Step 5: Run tests and verify GREEN**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-domain -count=1`

Expected: PASS.

## Chunk 2: Examples and Documentation

### Task 3: Add local translation command

**Files:**
- Create: `references/cmd/ai-domain/main.go`
- Modify: `test/ai-domain/ai_domain_test.go`

- [x] **Step 1: Write a failing command test**

The test should run:

```bash
go run ./references/cmd/ai-domain -f docs/examples/ai-platform/domain/ai-service.yaml
```

Expected before implementation: FAIL because the command does not exist.

- [x] **Step 2: Implement the command**

The command should:
- require `-f`
- read the domain YAML file
- call `domain.TranslateYAML`
- write the generated KubeVela `Application` YAML to stdout

- [x] **Step 3: Run the command test**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-domain -count=1`

Expected: PASS.

### Task 4: Add domain examples

**Files:**
- Create: `docs/examples/ai-platform/domain/ai-service.yaml`
- Create: `docs/examples/ai-platform/domain/ai-job.yaml`
- Modify: `docs/examples/ai-platform/README.md`

- [x] **Step 1: Add AIService domain input example**

The example should be runnable through the translator and should use a real pullable image for minimal validation.

- [x] **Step 2: Add AIJob domain input example**

The example should use `busybox` and a short command that can reach `Completed` in a basic Kubernetes cluster.

- [x] **Step 3: Document the boundary**

README must state:
- domain YAML is not installed as a CRD
- the translator emits native KubeVela `Application`
- no new database/controller/reconcile loop is added
- `AIWorkflow` is intentionally out of first PoC scope

### Task 5: Add tests for examples and scope guard

**Files:**
- Modify: `test/ai-domain/ai_domain_test.go`

- [x] **Step 1: Test example files translate successfully**

Read both domain examples and ensure `TranslateYAML` returns standard KubeVela `Application` YAML.

- [x] **Step 2: Add scope guard**

Scan the new domain package and examples for forbidden implementation scope terms that would indicate accidental controller/database expansion:
- `Reconcile`
- `Controller`
- `database`
- `SQL`

- [x] **Step 3: Run tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-domain -count=1`

Expected: PASS.

## Chunk 3: Verification and Upload

### Task 6: Verify related AI platform tests

**Files:**
- Test only.

- [ ] **Step 1: Run new domain tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-domain -count=1`

Expected: PASS.

- [ ] **Step 2: Run existing AI definition tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

- [ ] **Step 3: Check git diff**

Run: `git status --short` and `git diff --stat`

Expected: only domain translator, examples, docs, and this plan changed.

### Task 7: Commit and push

**Files:**
- All changed files from this plan.

- [ ] **Step 1: Commit**

Run:

```bash
git add pkg/ai/domain references/cmd/ai-domain test/ai-domain docs/examples/ai-platform docs/superpowers/plans/2026-05-19-ai-domain-translator.md
git commit -m "Add minimal AI domain translator"
```

- [ ] **Step 2: Push**

Run:

```bash
git push fork feature/zhisuan-naguan-platform
```

- [ ] **Step 3: Report verification evidence**

Include the exact test commands run and the pushed branch in the final response.
