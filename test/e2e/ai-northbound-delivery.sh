#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:18089}"
NAMESPACE="${NAMESPACE:-sock-shop}"
JOB_NAME="${JOB_NAME:-delivery-e2e-$(date +%m%d%H%M%S)}"
SERVICE_NAME="${SERVICE_NAME:-${JOB_NAME}-service}"
TIMEOUT="${TIMEOUT:-120s}"
LOGIN_ROLE="${AI_NORTHBOUND_ROLE:-user}"
LOGIN_USERNAME="${AI_NORTHBOUND_USERNAME:-admin}"
LOGIN_PASSWORD="${AI_NORTHBOUND_PASSWORD:-shiyong}"

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 1
  }
}

need curl
need kubectl

wait_for_named_resource() {
  local kind="$1"
  local selector="$2"
  local name=""
  for _ in $(seq 1 60); do
    name="$(kubectl get "$kind" -n "$NAMESPACE" -l "$selector" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)"
    if [ -n "$name" ]; then
      echo "$name"
      return 0
    fi
    sleep 2
  done
  echo "timed out waiting for ${kind} with selector ${selector}" >&2
  return 1
}

wait_for_probe() {
  for _ in $(seq 1 30); do
    api_curl -X POST "${BASE_URL}/api/v1/ai/applications/${NAMESPACE}/${SERVICE_NAME}/probe" \
      -H "Content-Type: application/json" \
      -d '{"path":"/healthz","timeoutSeconds":5}' | tee "$tmpdir/probe.json"
    if grep -q '"healthy":true' "$tmpdir/probe.json"; then
      return 0
    fi
    sleep 2
  done
  echo "timed out waiting for successful service probe" >&2
  return 1
}

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT
cookie_file="$tmpdir/cookies.txt"

login() {
  local status
  status="$(curl -sS -o "$tmpdir/login.response" -w '%{http_code}' \
    -c "$cookie_file" \
    -X POST "${BASE_URL}/login" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    --data-urlencode "role=${LOGIN_ROLE}" \
    --data-urlencode "username=${LOGIN_USERNAME}" \
    --data-urlencode "password=${LOGIN_PASSWORD}")"
  if [ "$status" != "302" ]; then
    echo "login failed for ${LOGIN_ROLE}/${LOGIN_USERNAME}: HTTP ${status}" >&2
    sed -n '1,80p' "$tmpdir/login.response" >&2
    return 1
  fi
  if [ ! -s "$cookie_file" ]; then
    echo "login succeeded without a session cookie" >&2
    return 1
  fi
}

api_curl() {
  curl -fsS -b "$cookie_file" "$@"
}

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

echo "== login ${LOGIN_ROLE}/${LOGIN_USERNAME} =="
login

echo "== submit AIJob ${NAMESPACE}/${JOB_NAME} =="
api_curl -X POST "${BASE_URL}/api/v1/ai/applications" \
  -H "Content-Type: application/yaml" \
  --data-binary @"$job_yaml" >"$tmpdir/deploy.json"

echo "== wait Kubernetes Job complete =="
job="$(wait_for_named_resource job "app.oam.dev/name=${JOB_NAME}")"
kubectl wait --for=condition=complete job/"$job" -n "$NAMESPACE" --timeout="$TIMEOUT"

pod="$(kubectl get pod -n "$NAMESPACE" -l "app.oam.dev/name=${JOB_NAME}" -o jsonpath='{.items[0].metadata.name}')"
echo "== pod logs ${pod} =="
kubectl logs -n "$NAMESPACE" "$pod" | tee "$tmpdir/job.log"
grep -q "AI_RESULT_JSON=" "$tmpdir/job.log"

echo "== parse delivery result =="
api_curl "${BASE_URL}/api/v1/ai/deliveries/${NAMESPACE}/${JOB_NAME}/result" | tee "$tmpdir/result.json"
grep -q "inline://models/${JOB_NAME}/v1" "$tmpdir/result.json"

echo "== publish AIService ${SERVICE_NAME} =="
api_curl -X POST "${BASE_URL}/api/v1/ai/deliveries/${NAMESPACE}/${JOB_NAME}/publish-service" \
  -H "Content-Type: application/json" \
  -d "{\"serviceName\":\"${SERVICE_NAME}\",\"image\":\"python:3.11-slim\",\"port\":8080,\"servicePort\":80}" | tee "$tmpdir/publish.json"

echo "== wait AIService deployment ready =="
deployment="$(wait_for_named_resource deployment "app.oam.dev/name=${SERVICE_NAME}")"
kubectl rollout status deployment/"$deployment" -n "$NAMESPACE" --timeout="$TIMEOUT"

echo "== probe AIService =="
wait_for_probe

echo "== verify artifact registry =="
api_curl "${BASE_URL}/api/v1/ai/artifacts?namespace=${NAMESPACE}" | tee "$tmpdir/artifacts.json"
grep -q "inline://models/${JOB_NAME}/v1" "$tmpdir/artifacts.json"

echo "AI northbound delivery E2E passed: ${NAMESPACE}/${JOB_NAME} -> ${SERVICE_NAME}"
