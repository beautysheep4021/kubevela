# AI Readonly Observe Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the minimal readonly observability layer for AIService and AIJob workloads managed through KubeVela Applications.

**Architecture:** Observation is a client-side readonly query and aggregation path. It reads KubeVela `Application` status and selected Kubernetes workload objects, then emits a compact summary without adding controllers, databases, Prometheus, UI, or any write/reconcile behavior.

**Tech Stack:** Go, Kubernetes dynamic client, unstructured objects, KubeVela `Application` status, JSON output, existing `ai-domain` CLI.

---

## Chunk 1: Readonly Summary Core

### Task 1: Add failing summary tests

**Files:**
- Create: `pkg/ai/observe/summary.go`
- Create: `test/ai-observe/ai_observe_test.go`

- [x] **Step 1: Write a failing AIService summary test**

The test should build an in-memory KubeVela `Application` object plus Deployment, Service, and Pod objects. It should call the observe package and assert:
- application name and namespace are preserved
- component name and component type are read from Application status/services
- phase, health, and message are read from Application status
- Deployment ready replicas and desired replicas are summarized
- Service cluster IP and port are summarized
- Pod phase and ready container count are summarized
- `ai.oam.dev/*` labels and annotations are collected as AI metadata

- [x] **Step 2: Write a failing AIJob summary test**

The test should build an Application plus Job and Pod objects. It should assert:
- Job active/succeeded/failed counts are summarized
- TTL is exposed when present
- Completed Pod status is summarized
- empty pod owner references are surfaced as a warning because this was observed on the server

- [x] **Step 3: Verify RED**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-observe -count=1`

Expected: FAIL because `pkg/ai/observe` does not exist.

### Task 2: Implement summary aggregation

**Files:**
- Create: `pkg/ai/observe/summary.go`
- Test: `test/ai-observe/ai_observe_test.go`

- [x] **Step 1: Define summary structs**

Add JSON-friendly structs:
- `Summary`
- `ComponentSummary`
- `WorkloadSummary`
- `PodSummary`
- `Warning`

- [x] **Step 2: Implement object-based aggregation**

Implement a pure function that accepts unstructured objects and returns a `Summary`. This keeps most behavior testable without a live cluster.

- [x] **Step 3: Collect AI metadata**

Collect labels and annotations with prefix `ai.oam.dev/` from Application, workload, pod template, and Pods.

- [x] **Step 4: Detect known orphan Job Pod symptom**

When a completed or failed Job Pod has no owner references, add a warning explaining that the pod may not be garbage-collected with the Job.

- [x] **Step 5: Verify GREEN**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-observe -count=1`

Expected: PASS.

## Chunk 2: CLI Status Command

### Task 3: Add status subcommand

**Files:**
- Modify: `references/cmd/ai-domain/main.go`
- Create: `pkg/ai/observe/client.go`
- Modify: `test/ai-observe/ai_observe_test.go`

- [x] **Step 1: Write failing CLI test with file inputs**

The test should run:

```bash
go run ./references/cmd/ai-domain status --from-files <fixtures>
```

Expected before implementation: FAIL because the status subcommand does not exist.

- [x] **Step 2: Add `ai-domain status` CLI**

Support:
- `ai-domain translate -f <domain.yaml>`
- legacy `ai-domain -f <domain.yaml>` for compatibility
- `ai-domain status -n <namespace> <application-name>` for live cluster query
- `ai-domain status --from-files <file1,file2,...>` for local fixture testing

- [x] **Step 3: Implement live readonly client**

Use Kubernetes client-go dynamic client to read:
- `applications.core.oam.dev/v1beta1`
- `deployments.apps/v1`
- `jobs.batch/v1`
- `services.v1`
- `pods.v1`

Only use GET/LIST verbs.

- [x] **Step 4: Verify CLI test**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-observe -count=1`

Expected: PASS.

## Chunk 3: Docs and Verification

### Task 4: Document readonly observe usage

**Files:**
- Modify: `docs/examples/ai-platform/README.md`

- [x] **Step 1: Add local status example**

Document `ai-domain status --from-files ...` for local fixture-based checks.

- [x] **Step 2: Add live cluster status example**

Document `ai-domain status -n sock-shop ai-service-domain-demo`.

- [x] **Step 3: State the boundary**

README must state the observe layer is readonly and does not add Prometheus, database, controller, reconcile loop, or UI.

### Task 5: Final verification and upload

**Files:**
- All changed files.

- [ ] **Step 1: Run observe tests**

Run: `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-observe -count=1`

Expected: PASS.

- [ ] **Step 2: Run existing AI tests**

Run:
- `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-domain -count=1`
- `env GOCACHE=/tmp/kubevela-go-build-cache go test ./test/ai-definitions -count=1`

Expected: PASS.

- [ ] **Step 3: Commit and push**

Run:

```bash
git add pkg/ai/observe references/cmd/ai-domain test/ai-observe docs/examples/ai-platform docs/superpowers/plans/2026-05-19-ai-observe-readonly.md
git commit -m "Add readonly AI workload observe summary"
git push fork feature/zhisuan-naguan-platform
```
