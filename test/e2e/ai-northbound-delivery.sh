#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:18088}"
NAMESPACE="${NAMESPACE:-sock-shop}"
JOB_NAME="${JOB_NAME:-delivery-e2e-$(date +%m%d%H%M%S)}"
SERVICE_NAME="${SERVICE_NAME:-${JOB_NAME}-service}"
TIMEOUT="${TIMEOUT:-120s}"

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 1
  }
}

need curl
need kubectl

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

job_yaml="$tmpdir/job.yaml"
cat >"$job_yaml" <<EOF
apiVersion: ai.oam.dev/v1alpha1
kind: AIJob
metadata:
  name: ${JOB_NAME}
  namespace: ${NAMESPACE}
spec:
  componentName: delivery-trainer
  properties:
    image: busybox:1.36
    imagePullPolicy: IfNotPresent
    jobKind: training
    cmd:
      - sh
      - -c
    args:
      - |
        echo train-start
        echo epoch=1 loss=0.30
        echo epoch=2 loss=0.12
        echo 'AI_RESULT_JSON={"modelURI":"inline://models/${JOB_NAME}/v1","metrics":{"loss":0.12,"accuracy":0.98},"summary":"trained-for-delivery"}'
        echo train-complete
    dataset:
      name: dataset
      uri: inline://datasets/tiny-delivery
    output:
      uri: inline://outputs/${JOB_NAME}
    backoffLimit: 0
    ttlSecondsAfterFinished: 3600
  runtime:
    runtime: batch
    framework: shell
    tenant: demo-tenant
    project: delivery-demo
    environment: poc
    owner: ai-platform
    datasetURI: inline://datasets/tiny-delivery
EOF

echo "== healthz =="
curl -fsS "${BASE_URL}/healthz"

echo "== submit AIJob ${NAMESPACE}/${JOB_NAME} =="
curl -fsS -X POST "${BASE_URL}/api/v1/ai/applications" \
  -H "Content-Type: application/yaml" \
  --data-binary @"$job_yaml" >"$tmpdir/deploy.json"

echo "== wait Kubernetes Job complete =="
kubectl wait --for=condition=complete job -n "$NAMESPACE" -l "app.oam.dev/name=${JOB_NAME}" --timeout="$TIMEOUT"

pod="$(kubectl get pod -n "$NAMESPACE" -l "app.oam.dev/name=${JOB_NAME}" -o jsonpath='{.items[0].metadata.name}')"
echo "== pod logs ${pod} =="
kubectl logs -n "$NAMESPACE" "$pod" | tee "$tmpdir/job.log"
grep -q "AI_RESULT_JSON=" "$tmpdir/job.log"

echo "== parse delivery result =="
curl -fsS "${BASE_URL}/api/v1/ai/deliveries/${NAMESPACE}/${JOB_NAME}/result" | tee "$tmpdir/result.json"
grep -q "inline://models/${JOB_NAME}/v1" "$tmpdir/result.json"

echo "== publish AIService ${SERVICE_NAME} =="
curl -fsS -X POST "${BASE_URL}/api/v1/ai/deliveries/${NAMESPACE}/${JOB_NAME}/publish-service" \
  -H "Content-Type: application/json" \
  -d "{\"serviceName\":\"${SERVICE_NAME}\",\"image\":\"python:3.11-slim\",\"port\":8080,\"servicePort\":80}" | tee "$tmpdir/publish.json"

echo "== wait AIService deployment ready =="
kubectl rollout status deployment -n "$NAMESPACE" -l "app.oam.dev/name=${SERVICE_NAME}" --timeout="$TIMEOUT"

echo "== probe AIService =="
curl -fsS -X POST "${BASE_URL}/api/v1/ai/applications/${NAMESPACE}/${SERVICE_NAME}/probe" \
  -H "Content-Type: application/json" \
  -d '{"path":"/healthz","timeoutSeconds":5}' | tee "$tmpdir/probe.json"
grep -q '"healthy":true' "$tmpdir/probe.json"

echo "== verify artifact registry =="
curl -fsS "${BASE_URL}/api/v1/ai/artifacts?namespace=${NAMESPACE}" | tee "$tmpdir/artifacts.json"
grep -q "inline://models/${JOB_NAME}/v1" "$tmpdir/artifacts.json"

echo "AI northbound delivery E2E passed: ${NAMESPACE}/${JOB_NAME} -> ${SERVICE_NAME}"
