# AI Northbound Kubernetes Deployment

This directory packages the Phase 2 northbound API as a Kubernetes workload so it no longer has to run as a manually uploaded Linux binary.

## Build Image

```bash
docker build \
  -f references/deploy/ai-northbound/Dockerfile \
  -t ghcr.io/beautysheep4021/kubevela-ai-northbound:dev .
```

Push the image to a registry reachable from the target cluster:

```bash
docker push ghcr.io/beautysheep4021/kubevela-ai-northbound:dev
```

If the cluster cannot pull from GitHub Container Registry, retag the image for the internal registry and update `deployment.yaml`.

## Deploy

```bash
kubectl apply -f references/deploy/ai-northbound/deployment.yaml
kubectl -n ai-platform rollout status deployment/ai-northbound
kubectl -n ai-platform port-forward svc/ai-northbound 18088:80
```

Then open:

```text
http://127.0.0.1:18088/
```

## Verify

```bash
curl -fsS http://127.0.0.1:18088/healthz
BASE_URL=http://127.0.0.1:18088 test/e2e/ai-northbound-delivery.sh
```

## Permissions

The included RBAC is intentionally broad enough for the current MVP:

- read and apply KubeVela `Application`
- read Deployments, Jobs, Services, Pods, Events and Pod logs
- update Deployments for restart
- create/update ConfigMaps for audit and model artifact records

Before production use, restrict namespaces and bind permissions per tenant.
