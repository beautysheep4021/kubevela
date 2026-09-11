# 智算纳管平台启动与部署

本文描述本功能分支的使用方式，不代表服务器当前已部署或已验证。除特别说明外，命令均在仓库根目录执行。

## 架构与运行方式

```text
浏览器
  -> ai-northbound（同一个 Go 进程提供前端、登录和管理 API）
  -> Kubernetes API Server（提交、查询 KubeVela Application）
  -> KubeVela 控制器 + AI 定义
       AIJob     -> Job -> 任务容器
       AIService -> Deployment + Service -> 服务容器
```

前端 HTML/CSS/JavaScript 内嵌在 Go 二进制中，浏览器以同源请求调用后端，不需要另起前端服务器。
平台镜像负责管理任务；训练、推理或智能体镜像负责执行任务。部署平台不会自动安装模型、数据或推理引擎。

## 本地预览

需要 Go（版本要求见 `go.mod`），不需要 Docker 或本地 Kubernetes：

```bash
go run ./references/cmd/ai-northbound -addr 127.0.0.1:18095
```

打开 <http://127.0.0.1:18095/login?demo=local>。

| 角色 | 用户名 | 密码 | 使用方范围 |
| --- | --- | --- | --- |
| 使用方 A | admin | shiyong | tenant-a / ai-tenant-a |
| 使用方 B | tenant-b | tenant-b-123456 | tenant-b / ai-tenant-b |
| 监测方 | admin | jiankong | 管理员视图 |

以上是默认测试账号，不应直接用于公网或生产环境。演示数据仅保存在页面内存，刷新重置；每个使用方有自己的 Job 和 Service，不会创建真实工作负载。
示例指标和访问探测不代表真实训练或模型推理成功。

未找到 kubeconfig 时，后端仍能提供页面、登录和配置转换，但集群相关接口不可用。
真实模式接口失败不会自动回退到演示数据。如果本机已有默认 kubeconfig，程序可能自动连接该集群，操作前应确认目标环境。

## 本地连接真实集群

不必先部署平台镜像，可以直接指定 kubeconfig：

```bash
go run ./references/cmd/ai-northbound \
  -addr 127.0.0.1:18095 \
  -kubeconfig /实际路径/kubeconfig
```

打开 <http://127.0.0.1:18095/login>，不要带 `?demo=local`。
本机必须能访问 kubeconfig 中的 API Server，凭据也必须有相应权限。仅复制配置文件不能解决网络不可达。

程序按以下顺序加载配置，见 [main.go](../../cmd/ai-northbound/main.go)：

1. 显式 `-kubeconfig`；指定的配置错误时直接退出。
2. Pod 内的 ServiceAccount，即 in-cluster config。
3. 默认 kubeconfig 加载规则，包括 `KUBECONFIG` 和 `~/.kube/config`。

本地进程可能无法访问集群内部的 Service 地址，因此任务查询可用不代表服务探测可用；完整访问验证优先在集群内运行平台。

## 服务器首次部署

### 准备集群

- Kubernetes 可用，`kubectl` 指向目标集群。
- KubeVela 控制器及 Application CRD 已安装。
- 已安装本分支的 `ai-job`、`ai-service` ComponentDefinition 和 `ai-runtime` TraitDefinition。
- 租户 Namespace 已建立，按需配置 ResourceQuota、LimitRange、NetworkPolicy 等。
- 节点能拉取平台和任务镜像，任务能访问模型与数据；GPU 任务还需驱动及设备插件等运行条件。

AI 定义位于 `charts/vela-core/templates/defwithtemplate/`，包含 Helm 模板，不能对源文件直接执行 `kubectl apply`。
首次安装参考 [vela-core chart](../../../charts/vela-core/README.md)；已有集群需保留现有 Helm values，并审核升级影响。
不要为更新前端而无条件重装整个 KubeVela。

默认测试租户可在不存在时创建：

```bash
kubectl create namespace ai-tenant-a --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace ai-tenant-b --dry-run=client -o yaml | kubectl apply -f -
```

平台部署清单只创建 `ai-platform`，不会自动创建上述租户及隔离策略。

### 构建平台镜像

以下命令在有 Docker 的构建机或 CI 上执行，不要求开发电脑或 K8s 节点启动 Docker。
镜像仓库需有推送权限，节点需有拉取权限；示例地址不代表镜像已经发布。

```bash
IMAGE="ghcr.io/beautysheep4021/kubevela-ai-northbound:$(git rev-parse --short HEAD)"
docker build -f references/deploy/ai-northbound/Dockerfile -t "$IMAGE" .
docker push "$IMAGE"
```

默认目标为 Linux amd64；其他节点架构需调整构建平台和 `TARGETARCH`。
可以替换为内部镜像仓库；私有仓库还需配置拉取凭据。离线导入应导入节点实际使用的容器运行时，仅存在于 Docker 的镜像不一定能被 containerd 使用。

### 启动平台

把 [deployment.yaml](deployment.yaml) 的容器 `image` 改为实际发布的完整镜像地址，再执行：

```bash
kubectl apply -f references/deploy/ai-northbound/deployment.yaml
kubectl -n ai-platform rollout status deployment/ai-northbound
kubectl -n ai-platform get pods -l app.kubernetes.io/name=ai-northbound
```

清单包含 Namespace、ServiceAccount、RBAC、Deployment 和 ClusterIP Service。
进程监听 `0.0.0.0:8088`，Service 使用 80 端口。Pod 自动使用 ServiceAccount，通常不需要挂载 kubeconfig。

### 访问页面

在能连接集群的机器上运行，保持该终端打开：

```bash
kubectl -n ai-platform port-forward svc/ai-northbound 18089:80
```

若在电脑执行，直接打开 <http://127.0.0.1:18089/login>。
若在服务器执行，默认只监听服务器回环地址；在电脑另开终端建立隧道：

```bash
ssh -N -L 18089:127.0.0.1:18089 用户名@服务器地址
```

再打开电脑的同一地址。端口冲突时更换本地端口。长期访问应配置 HTTPS 和访问控制，例如使用 Ingress，不要直接暴露默认测试账号。

## 验证

```bash
curl -fsS http://127.0.0.1:18089/healthz
```

`healthz` 只确认 HTTP 进程存活，不证明 K8s、KubeVela 或任务正常。
随后在不带演示参数的页面登录，检查本租户任务列表，创建示例 Job，查看完成状态和日志，再验证服务发布与访问。

已有 E2E 脚本会登录并创建真实资源，不是只读检查。执行机器需要 `curl`、`kubectl`，且 kubectl 与平台必须连接同一集群：

```bash
BASE_URL=http://127.0.0.1:18089 \
NAMESPACE=ai-tenant-a \
AI_NORTHBOUND_ROLE=user \
AI_NORTHBOUND_USERNAME=admin \
AI_NORTHBOUND_PASSWORD=shiyong \
TIMEOUT=180s \
bash test/e2e/ai-northbound-delivery.sh
```

显式指定账号对应的 Namespace，避免脚本默认 `sock-shop` 与租户范围不一致。
脚本验证示例交付链路，不等于真实大模型推理验收；资源保留与清理以脚本行为为准。

## 更新版本

```text
推送 GitHub -> 构建并发布新镜像 -> 更新 Deployment -> rollout -> 登录及任务验证
```

使用不同的镜像版本标签，避免 `IfNotPresent` 重用旧镜像：

```bash
kubectl -n ai-platform set image deployment/ai-northbound \
  ai-northbound=你的镜像仓库/kubevela-ai-northbound:新版本标签
kubectl -n ai-platform rollout status deployment/ai-northbound
```

同时更新部署清单或 GitOps 配置中的镜像版本，避免下次 apply 回退。前端内嵌于二进制，前端修改也需要重新构建部署。
Pod 替换后，临时 port-forward 可能需要重启，页面可能需要重新登录。

## 权限边界与排查

### 监测方工作台

登录页选择监测方角色，使用测试账号 `admin / jiankong`。新版监测方与使用方共用样式和状态规则，导航分为总览、任务、在线服务、租户资源、操作审计。

- 总览和列表支持命名空间、租户、名称及状态筛选；异常详情关联诊断、Pod 事件和日志。
- 重跑、重启和删除需要确认目标命名空间与名称。操作提交不代表恢复成功，应刷新详情检查运行状态并查看审计。
- 租户资源页区分任务资源声明与 ResourceQuota。配额分别展示期望上限、已生效上限和控制器记录的已用量，保留配额适用范围；这些都不是实时 CPU/GPU 利用率。
- 未配置配额、尚未更新配额状态和读取失败是不同状态，不把缺失数据显示为零。
- 新增只读 `GET /api/v1/ai/tenant-resources`，仅监测方可访问。升级已有部署时，除了更新镜像，还需应用新版 RBAC，为 ServiceAccount 增加 namespaces/resourcequotas 的 get/list 权限。
- 本地演示数据仍为各页面内存中的样例，不在使用方与监测方浏览器之间实时同步。真实模式两端读取同一集群资源。

当前平台使用共享 ServiceAccount 访问 K8s，由后端按登录会话校验租户范围，并非每个网站账号都对应独立 Kubernetes 用户。
ClusterRole 面向 MVP，允许 Application 管理、工作负载与日志读取、Deployment 更新、ConfigMap 读写。
生产化需收紧授权、替换默认认证并完善网络与资源隔离，不能只依赖前端隐藏内容。

| 现象 | 优先检查 |
| --- | --- |
| 页面正常，任务接口不可用 | kubeconfig、API Server 网络、后端启动日志 |
| 403 / 无权访问 | 登录账号与 Namespace 是否匹配 |
| Application 有了，没有 Job/Deployment | KubeVela 控制器、AI 定义、Application 状态 |
| Pod Pending / ImagePullBackOff | 资源、GPU 条件、镜像地址和凭据 |
| 服务有了，访问失败 | 就绪状态、监听端口、Service、网络策略和路由 |
| GitHub 已更新，页面还是旧版 | Deployment 镜像、rollout、访问端口及浏览器缓存 |

推送成功不等于部署成功；页面可访问不等于集群可用；示例健康检查成功不等于模型推理正确。
