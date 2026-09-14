# Real CPU Prediction Example

This example uses a small linear model, `y = 2*x + 1`, and matching JSONL input data. It reads real files and computes predictions; it is not a language model or a training benchmark.

## Start The Console

Build the current northbound binary, then start it with both cluster and upload storage configured:

```sh
go build -o /tmp/ai-northbound ./references/cmd/ai-northbound
/tmp/ai-northbound -addr 127.0.0.1:18097 \
  -kubeconfig /path/to/cluster-kubeconfig \
  -dataset-dir /path/to/persistent-datasets
```

The cluster must already have KubeVela and the `ai-job` / `ai-service` ComponentDefinitions. An administrator must provision each account's namespace before importing assets. Do not use production cluster-admin credentials with a publicly reachable PoC console.

## Use The Example

1. Log into the user console without `?demo=local`.
2. In the model or dataset library, choose the example import action. It creates the matching model and dataset in the logged-in account's namespace.
3. Run a prediction task from the example asset. The resulting KubeVela Application owns a Kubernetes Job.
4. View task logs for computed predictions and the `AI_RESULT_JSON` summary.
5. Deploy a prediction service from the example model. Its KubeVela Application owns the Deployment and Service.

The service supports `/predict?x=4`, which returns a computed prediction of `9`, and a health endpoint. The service is cluster-internal, not an automatically published Internet endpoint.

## Limits

- ConfigMap file delivery is intentionally limited to these tiny bundled example assets. It is not a large-model storage backend.
- Imported assets and workload launches are separate operations; importing does not consume a running Pod.
- Local dataset uploads remain locally persisted assets and are not automatically mounted into arbitrary cluster tasks.
- A model file address entered manually is only a registration record, not proof that a runtime can load it.
- Tenant-scoped API checks and namespaces do not replace production OIDC, Kubernetes RBAC, network policy, or storage authorization.
- No local Docker build is required. Cluster nodes still need the Python container image used by the workload.
