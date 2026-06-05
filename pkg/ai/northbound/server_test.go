package northbound

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oam-dev/kubevela/pkg/ai/domain"
	domainapply "github.com/oam-dev/kubevela/pkg/ai/domain/apply"
	"github.com/oam-dev/kubevela/pkg/ai/observe"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestServerHealthz(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if strings.TrimSpace(recorder.Body.String()) != "ok" {
		t.Fatalf("body = %q, want ok", recorder.Body.String())
	}
}

func TestServerServesConsolePage(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	contentType := recorder.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Fatalf("content-type = %q, want text/html", contentType)
	}
	body := recorder.Body.String()
	for _, expected := range []string{
		"使用方工作台",
		"监测方工作台",
		"算法任务模板",
		"SFT 微调",
		"训练向导",
		"基础模型",
		"训练数据集",
		"待发布模型",
		"评测数据集",
		"训练规格",
		"训练完成后自动发布为服务",
		"高级配置",
		"领域 YAML",
		"治理意图",
		"工作负载意图",
		"全局任务概览",
		"筛选条件",
		"异常任务",
		"AIService 数量",
		"AIJob 数量",
		"任务类型",
		"模型名称",
		"数据集 URI",
		"生成 YAML",
		"提交部署",
		"服务端 DryRun",
		"/api/v1/ai/validate",
		"/api/v1/ai/normalize",
		"/api/v1/ai/applications",
		"任务列表",
		"刷新任务",
		"运行日志",
		"刷新日志",
		"删除任务",
		"重启服务",
		"重新运行 Job",
		"操作审计",
		"应用交付",
		"解析训练结果",
		"发布为服务",
		"测试服务访问",
		"查看模型产物",
		"我的模型",
		"模型资产库",
		"登记模型",
		"设为基础模型",
		"发布此模型",
		"数据资产库",
		"登记数据集",
		"选择数据集",
		"/api/v1/ai/datasets",
		"通过阈值",
		"发起评测",
		"同步评测结果",
		"/api/v1/ai/artifacts",
		"/api/v1/ai/models",
		"/evaluate",
		"/sync-evaluation",
		"/probe",
		"/logs",
		"/api/v1/ai/audits",
		"/api/v1/ai/deliveries",
		"governanceIntent",
		"workloadIntent",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("console page missing %q", expected)
		}
	}
	for _, forbidden := range []string{
		"基础模型 URI",
		"训练数据 URI",
		"待发布模型 URI",
		"评测数据 URI",
		"数据集内部 URI",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("console page still exposes user-facing URI label %q", forbidden)
		}
	}
}

func TestConsoleGeneratedAIJobIncludesDeliveryResultMarker(t *testing.T) {
	for _, expected := range []string{
		"function buildDeliveryJobResultScript",
		"AI_RESULT_JSON=",
		"inline://models/\" + fields.name.value + \"/v1",
		"echo training_size=",
		"line(\"trainingSize\", wizardFields.trainingSize.value, 4)",
		"function looksLikeModelVersion",
	} {
		if !strings.Contains(consoleHTML, expected) {
			t.Fatalf("console AIJob generator missing %q", expected)
		}
	}
}

func TestConsoleKeepsUserAssetPathsAwayFromRawURIs(t *testing.T) {
	for _, forbidden := range []string{
		"datasetURI。</div>",
		"请先填写或选择一个数据集内部 URI。",
		"<span title=\\\"\" + uri + \"\\\">\" + (uri || \"-\") + \"</span>",
		"<span title=\\\"\" + (item.modelURI || \"\") + \"\\\">\" + (item.modelURI || \"-\") + \"</span>",
	} {
		if strings.Contains(consoleHTML, forbidden) {
			t.Fatalf("console user asset path still renders raw URI fragment %q", forbidden)
		}
	}
	for _, expected := range []string{
		"function setBaseModel",
		"function setServiceModel",
		"function setEvaluationDataset",
		"baseModelSelect.onchange",
		"serviceModelSelect.onchange",
		"evaluationDatasetSelect.onchange",
		"资产编号：",
		"数据集编号：",
	} {
		if !strings.Contains(consoleHTML, expected) {
			t.Fatalf("console user asset path missing %q", expected)
		}
	}
}

func TestServerListsApplications(t *testing.T) {
	reader := &recordingApplicationReader{
		items: []observe.ApplicationListItem{
			{
				Name:          "ai-service-northbound-demo",
				Namespace:     "sock-shop",
				Phase:         "running",
				Healthy:       true,
				Message:       "Ready:1/1",
				WorkloadTypes: []string{"service"},
				AIMetadata: map[string]string{
					"ai.oam.dev/tenant": "demo-tenant",
				},
			},
		},
	}
	server := NewServerWithOptions(Options{Reader: reader})
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/applications?namespace=sock-shop", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload ApplicationsResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if reader.listNamespace != "sock-shop" {
		t.Fatalf("namespace = %q, want sock-shop", reader.listNamespace)
	}
	if len(payload.Items) != 1 || payload.Items[0].Name != "ai-service-northbound-demo" || payload.Items[0].WorkloadTypes[0] != "service" {
		t.Fatalf("unexpected application list: %#v", payload)
	}
}

func TestServerReturnsApplicationStatus(t *testing.T) {
	reader := &recordingApplicationReader{
		summaries: map[string]*observe.Summary{
			"sock-shop/ai-service-northbound-demo": {
				Name:      "ai-service-northbound-demo",
				Namespace: "sock-shop",
				Phase:     "running",
				Healthy:   true,
				Message:   "Ready:1/1",
			},
		},
	}
	server := NewServerWithOptions(Options{Reader: reader})
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/applications/sock-shop/ai-service-northbound-demo/status", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var summary observe.Summary
	if err := json.Unmarshal(recorder.Body.Bytes(), &summary); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if summary.Name != "ai-service-northbound-demo" || summary.Namespace != "sock-shop" || !summary.Healthy {
		t.Fatalf("unexpected summary: %#v", summary)
	}
}

func TestServerReturnsApplicationLogs(t *testing.T) {
	reader := &recordingApplicationReader{
		logs: map[string]*observe.Logs{
			"sock-shop/ai-job-demo": {
				Namespace:   "sock-shop",
				Application: "ai-job-demo",
				Pod:         "ai-job-demo-pod",
				Container:   "main",
				TailLines:   80,
				Logs:        "epoch=1 loss=0.42\n",
				Pods: []observe.LogPod{
					{
						Name:       "ai-job-demo-pod",
						Phase:      "Succeeded",
						Containers: []string{"main"},
					},
				},
			},
		},
	}
	server := NewServerWithOptions(Options{Reader: reader})
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/applications/sock-shop/ai-job-demo/logs?pod=ai-job-demo-pod&container=main&tailLines=80", nil)
	server.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload observe.Logs
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if payload.Pod != "ai-job-demo-pod" || payload.Container != "main" || payload.TailLines != 80 || !strings.Contains(payload.Logs, "loss=0.42") {
		t.Fatalf("unexpected logs payload: %#v", payload)
	}
	if reader.logRequest.Namespace != "sock-shop" || reader.logRequest.Name != "ai-job-demo" || reader.logRequest.Pod != "ai-job-demo-pod" || reader.logRequest.Container != "main" || reader.logRequest.TailLines != 80 {
		t.Fatalf("unexpected log request: %#v", reader.logRequest)
	}
}

func TestServerProbesApplicationService(t *testing.T) {
	prober := &recordingApplicationProber{
		result: &observe.ProbeResult{
			Namespace:   "sock-shop",
			Application: "delivery-service",
			Path:        "/healthz",
			StatusCode:  200,
			Healthy:     true,
			Body:        `{"status":"ok"}`,
		},
	}
	server := NewServerWithOptions(Options{Prober: prober})
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/applications/sock-shop/delivery-service/probe", strings.NewReader(`{"path":"/healthz","timeoutSeconds":3}`))
	req.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload observe.ProbeResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if !payload.Healthy || payload.StatusCode != 200 || payload.Body != `{"status":"ok"}` {
		t.Fatalf("unexpected probe payload: %#v", payload)
	}
	if prober.request.Namespace != "sock-shop" || prober.request.Name != "delivery-service" || prober.request.Path != "/healthz" || prober.request.TimeoutSeconds != 3 {
		t.Fatalf("unexpected probe request: %#v", prober.request)
	}
}

func TestServerDeletesApplicationAndRecordsAudit(t *testing.T) {
	manager := &recordingApplicationManager{}
	audits := &recordingAuditStore{}
	server := NewServerWithOptions(Options{Manager: manager, Audits: audits})
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/ai/applications/sock-shop/ai-job-demo", nil)
	req.Header.Set("X-AI-User", "tester")
	server.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload observe.LifecycleResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if payload.Action != "delete" || payload.Namespace != "sock-shop" || payload.Name != "ai-job-demo" {
		t.Fatalf("unexpected lifecycle result: %#v", payload)
	}
	if manager.deleted != "sock-shop/ai-job-demo" {
		t.Fatalf("deleted = %q, want sock-shop/ai-job-demo", manager.deleted)
	}
	if len(audits.events) != 1 || audits.events[0].Action != "delete" || audits.events[0].Actor != "tester" || audits.events[0].Success != true {
		t.Fatalf("unexpected audit events: %#v", audits.events)
	}
}

func TestServerRestartsApplicationAndRecordsAudit(t *testing.T) {
	manager := &recordingApplicationManager{}
	audits := &recordingAuditStore{}
	server := NewServerWithOptions(Options{Manager: manager, Audits: audits})
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/applications/sock-shop/ai-service-demo/restart", nil)
	server.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if manager.restarted != "sock-shop/ai-service-demo" {
		t.Fatalf("restarted = %q, want sock-shop/ai-service-demo", manager.restarted)
	}
	if len(audits.events) != 1 || audits.events[0].Action != "restart" || audits.events[0].Actor != "anonymous" {
		t.Fatalf("unexpected audit events: %#v", audits.events)
	}
}

func TestServerRerunsApplicationAndRecordsAudit(t *testing.T) {
	manager := &recordingApplicationManager{}
	audits := &recordingAuditStore{}
	server := NewServerWithOptions(Options{Manager: manager, Audits: audits})
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/applications/sock-shop/ai-job-demo/rerun", nil)
	server.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if manager.rerun != "sock-shop/ai-job-demo" {
		t.Fatalf("rerun = %q, want sock-shop/ai-job-demo", manager.rerun)
	}
	if len(audits.events) != 1 || audits.events[0].Action != "rerun" {
		t.Fatalf("unexpected audit events: %#v", audits.events)
	}
}

func TestServerListsAuditEvents(t *testing.T) {
	audits := &recordingAuditStore{
		events: []observe.AuditEvent{
			{
				Namespace: "sock-shop",
				Name:      "ai-job-demo",
				Action:    "delete",
				Actor:     "tester",
				Success:   true,
			},
		},
	}
	server := NewServerWithOptions(Options{Audits: audits})
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/audits?namespace=sock-shop", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload observe.AuditList
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if len(payload.Items) != 1 || payload.Items[0].Namespace != "sock-shop" || payload.Items[0].Action != "delete" {
		t.Fatalf("unexpected audits: %#v", payload)
	}
	if audits.listNamespace != "sock-shop" {
		t.Fatalf("list namespace = %q, want sock-shop", audits.listNamespace)
	}
}

func TestServerExtractsDeliveryResultFromAIJobLogs(t *testing.T) {
	reader := &recordingApplicationReader{
		logs: map[string]*observe.Logs{
			"sock-shop/train-demo": {
				Namespace:   "sock-shop",
				Application: "train-demo",
				Pod:         "train-demo-pod",
				Container:   "trainer",
				Logs: strings.Join([]string{
					"epoch=1 loss=0.3",
					`AI_RESULT_JSON={"modelURI":"inline://models/train-demo/v1","metrics":{"loss":0.12,"accuracy":0.98},"summary":"trained"}`,
					"done",
				}, "\n"),
			},
		},
	}
	server := NewServerWithOptions(Options{Reader: reader})
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/deliveries/sock-shop/train-demo/result", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload DeliveryResultResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if payload.Namespace != "sock-shop" || payload.JobName != "train-demo" || payload.Result.ModelURI != "inline://models/train-demo/v1" {
		t.Fatalf("unexpected delivery result: %#v", payload)
	}
	if payload.Result.Metrics["accuracy"] != 0.98 || payload.Result.Summary != "trained" {
		t.Fatalf("unexpected metrics: %#v", payload.Result)
	}
}

func TestServerPublishesDeliveryResultAsAIService(t *testing.T) {
	reader := &recordingApplicationReader{
		logs: map[string]*observe.Logs{
			"sock-shop/train-demo": {
				Logs: `AI_RESULT_JSON={"modelURI":"inline://models/train-demo/v1","metrics":{"loss":0.12}}`,
			},
		},
	}
	applier := &recordingApplicationApplier{}
	audits := &recordingAuditStore{}
	server := NewServerWithOptions(Options{Reader: reader, Applier: applier, Audits: audits})
	body := strings.NewReader(`{"serviceName":"train-demo-service","image":"python:3.11-slim","port":8080,"servicePort":80}`)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/deliveries/sock-shop/train-demo/publish-service", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AI-User", "tester")
	server.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload DeliveryPublishResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if payload.ServiceName != "train-demo-service" || payload.ModelURI != "inline://models/train-demo/v1" {
		t.Fatalf("unexpected publish response: %#v", payload)
	}
	if !applier.called || applier.namespace != "sock-shop" || applier.name != "train-demo-service" {
		t.Fatalf("unexpected applier call: %#v", applier)
	}
	content := string(applier.content)
	for _, expected := range []string{"kind: Application", "type: ai-service", "modelURI: inline://models/train-demo/v1", "name: train-demo-service", "cmd:", "socketserver.TCPServer", "serve model"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected translated AIService Application to contain %q, got:\n%s", expected, content)
		}
	}
	if len(audits.events) != 1 || audits.events[0].Action != "publish-service" || audits.events[0].Actor != "tester" {
		t.Fatalf("unexpected audit events: %#v", audits.events)
	}
}

func TestServerPublishesDeliveryResultAndRecordsArtifact(t *testing.T) {
	reader := &recordingApplicationReader{
		logs: map[string]*observe.Logs{
			"sock-shop/train-demo": {
				Pod:  "train-demo-pod",
				Logs: `AI_RESULT_JSON={"modelURI":"inline://models/train-demo/v1","metrics":{"loss":0.12},"summary":"trained"}`,
			},
		},
	}
	artifacts := &recordingArtifactStore{}
	server := NewServerWithOptions(Options{Reader: reader, Applier: &recordingApplicationApplier{}, Artifacts: artifacts})
	body := strings.NewReader(`{"serviceName":"train-demo-service"}`)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/deliveries/sock-shop/train-demo/publish-service", body))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if len(artifacts.saved) != 1 {
		t.Fatalf("saved artifacts = %d, want 1", len(artifacts.saved))
	}
	artifact := artifacts.saved[0]
	if artifact.Namespace != "sock-shop" || artifact.JobName != "train-demo" || artifact.ModelURI != "inline://models/train-demo/v1" {
		t.Fatalf("unexpected artifact: %#v", artifact)
	}
	if len(artifact.PublishedServices) != 1 || artifact.PublishedServices[0] != "train-demo-service" {
		t.Fatalf("unexpected published services: %#v", artifact.PublishedServices)
	}
	if artifact.Status != "published" || artifact.Visibility != "private" || artifact.EvaluationStatus != "passed" {
		t.Fatalf("unexpected artifact lifecycle fields: %#v", artifact)
	}
}

func TestServerListsArtifacts(t *testing.T) {
	artifacts := &recordingArtifactStore{
		list: []observe.ModelArtifact{
			{
				Namespace: "sock-shop",
				Name:      "train-demo",
				JobName:   "train-demo",
				ModelURI:  "inline://models/train-demo/v1",
			},
		},
	}
	server := NewServerWithOptions(Options{Artifacts: artifacts})
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/artifacts?namespace=sock-shop", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload observe.ArtifactList
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if len(payload.Items) != 1 || payload.Items[0].ModelURI != "inline://models/train-demo/v1" {
		t.Fatalf("unexpected artifacts: %#v", payload)
	}
	if artifacts.listNamespace != "sock-shop" {
		t.Fatalf("namespace = %q, want sock-shop", artifacts.listNamespace)
	}
}

func TestServerListsModelAssets(t *testing.T) {
	artifacts := &recordingArtifactStore{
		list: []observe.ModelArtifact{
			{
				Namespace:        "sock-shop",
				Name:             "customer-sft-demo",
				ModelURI:         "inline://models/customer-sft-demo/v1",
				Status:           "evaluated",
				Visibility:       "private",
				EvaluationStatus: "passed",
			},
		},
	}
	server := NewServerWithOptions(Options{Artifacts: artifacts})
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/models?namespace=sock-shop", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload observe.ArtifactList
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if len(payload.Items) != 1 || payload.Items[0].Name != "customer-sft-demo" || payload.Items[0].Visibility != "private" {
		t.Fatalf("unexpected model assets: %#v", payload)
	}
	if artifacts.listNamespace != "sock-shop" {
		t.Fatalf("namespace = %q, want sock-shop", artifacts.listNamespace)
	}
}

func TestServerRegistersModelAsset(t *testing.T) {
	artifacts := &recordingArtifactStore{}
	server := NewServerWithOptions(Options{Artifacts: artifacts})
	body := strings.NewReader(`{
		"namespace": "sock-shop",
		"name": "customer-sft-demo",
		"jobName": "customer-sft-demo",
		"modelURI": "inline://models/customer-sft-demo/v1",
		"baseModelURI": "modelscope://qwen/Qwen2.5-0.5B",
		"datasetURI": "inline://datasets/customer-sft-demo",
		"status": "trained",
		"visibility": "private",
		"evaluationStatus": "pending"
	}`)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/models", body))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if len(artifacts.saved) != 1 {
		t.Fatalf("saved artifacts = %d, want 1", len(artifacts.saved))
	}
	asset := artifacts.saved[0]
	if asset.Namespace != "sock-shop" || asset.Name != "customer-sft-demo" || asset.ModelURI != "inline://models/customer-sft-demo/v1" {
		t.Fatalf("unexpected saved model asset: %#v", asset)
	}
	if asset.BaseModelURI != "modelscope://qwen/Qwen2.5-0.5B" || asset.DatasetURI != "inline://datasets/customer-sft-demo" {
		t.Fatalf("unexpected model lineage: %#v", asset)
	}
	if asset.Visibility != "private" || asset.Status != "trained" || asset.EvaluationStatus != "pending" {
		t.Fatalf("unexpected model lifecycle fields: %#v", asset)
	}
}

func TestServerRegistersModelAssetDerivesNameFromVersionedURI(t *testing.T) {
	artifacts := &recordingArtifactStore{}
	server := NewServerWithOptions(Options{Artifacts: artifacts})
	body := strings.NewReader(`{
		"namespace": "sock-shop",
		"modelURI": "inline://models/customer-sft-demo/v1"
	}`)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/models", body))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if len(artifacts.saved) != 1 || artifacts.saved[0].Name != "customer-sft-demo" {
		t.Fatalf("unexpected derived model asset: %#v", artifacts.saved)
	}
}

func TestServerListsDatasetAssets(t *testing.T) {
	datasets := &recordingDatasetStore{
		list: []observe.DatasetArtifact{
			{
				Namespace:  "sock-shop",
				Name:       "customer-sft",
				DatasetURI: "dataset://sock-shop/customer-sft/v1",
				Format:     "sharegpt-jsonl",
				Purpose:    "sft",
				Status:     "validated",
			},
		},
	}
	server := NewServerWithOptions(Options{Datasets: datasets})
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/datasets?namespace=sock-shop", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload observe.DatasetList
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if len(payload.Items) != 1 || payload.Items[0].DatasetURI != "dataset://sock-shop/customer-sft/v1" || payload.Items[0].Status != "validated" {
		t.Fatalf("unexpected dataset assets: %#v", payload)
	}
	if datasets.listNamespace != "sock-shop" {
		t.Fatalf("namespace = %q, want sock-shop", datasets.listNamespace)
	}
}

func TestServerRegistersDatasetAsset(t *testing.T) {
	datasets := &recordingDatasetStore{}
	server := NewServerWithOptions(Options{Datasets: datasets})
	body := strings.NewReader(`{
		"namespace": "sock-shop",
		"name": "customer-sft",
		"displayName": "客服问答 SFT 数据集",
		"datasetURI": "oss://datasets/customer-sft/v1/train.jsonl",
		"format": "sharegpt-jsonl",
		"purpose": "sft",
		"status": "validated"
	}`)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/datasets", body)
	req.Header.Set("X-AI-User", "tester")
	server.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if len(datasets.saved) != 1 {
		t.Fatalf("saved datasets = %d, want 1", len(datasets.saved))
	}
	asset := datasets.saved[0]
	if asset.Namespace != "sock-shop" || asset.Name != "customer-sft" || asset.DatasetURI != "oss://datasets/customer-sft/v1/train.jsonl" {
		t.Fatalf("unexpected saved dataset asset: %#v", asset)
	}
	if asset.DisplayName != "客服问答 SFT 数据集" || asset.Format != "sharegpt-jsonl" || asset.Purpose != "sft" {
		t.Fatalf("unexpected dataset metadata: %#v", asset)
	}
	if asset.Status != "validated" || asset.Owner != "tester" {
		t.Fatalf("unexpected dataset lifecycle fields: %#v", asset)
	}
}

func TestServerStartsModelEvaluationJob(t *testing.T) {
	applier := &recordingApplicationApplier{}
	artifacts := &recordingArtifactStore{
		list: []observe.ModelArtifact{
			{
				Namespace: "sock-shop",
				Name:      "customer-sft-demo",
				ModelURI:  "inline://models/customer-sft-demo/v1",
			},
		},
	}
	server := NewServerWithOptions(Options{Applier: applier, Artifacts: artifacts})
	body := strings.NewReader(`{
		"evaluationDatasetURI": "inline://datasets/customer-eval",
		"evaluationType": "accuracy",
		"passThreshold": 0.8
	}`)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/models/sock-shop/customer-sft-demo/evaluate", body))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if !applier.called {
		t.Fatalf("expected evaluation Application to be applied")
	}
	content := string(applier.content)
	for _, expected := range []string{
		"name: customer-sft-demo-eval",
		"type: ai-job",
		"jobKind: evaluation",
		"model_uri=inline://models/customer-sft-demo/v1",
		"evaluation_dataset=inline://datasets/customer-eval",
		"pass_threshold=0.8",
		"AI_RESULT_JSON=",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected evaluation Application to contain %q, got:\n%s", expected, content)
		}
	}
}

func TestServerSyncsModelEvaluationResult(t *testing.T) {
	reader := &recordingApplicationReader{
		logs: map[string]*observe.Logs{
			"sock-shop/customer-sft-demo-eval": {
				Pod:  "customer-sft-demo-eval-pod",
				Logs: `AI_RESULT_JSON={"modelURI":"inline://models/customer-sft-demo/v1","metrics":{"accuracy":0.91},"summary":"evaluation-passed"}`,
			},
		},
	}
	artifacts := &recordingArtifactStore{
		list: []observe.ModelArtifact{
			{
				Namespace:        "sock-shop",
				Name:             "customer-sft-demo",
				ModelURI:         "inline://models/customer-sft-demo/v1",
				EvaluationStatus: "pending",
			},
		},
	}
	server := NewServerWithOptions(Options{Reader: reader, Artifacts: artifacts})
	body := strings.NewReader(`{"evaluationJobName":"customer-sft-demo-eval"}`)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/models/sock-shop/customer-sft-demo/sync-evaluation", body))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if len(artifacts.saved) != 1 {
		t.Fatalf("saved artifacts = %d, want 1", len(artifacts.saved))
	}
	asset := artifacts.saved[0]
	if asset.EvaluationStatus != "passed" || asset.Status != "evaluated" || asset.Metrics["accuracy"] != 0.91 {
		t.Fatalf("unexpected evaluated model asset: %#v", asset)
	}
	if asset.Summary != "evaluation-passed" || asset.SourceApplication != "customer-sft-demo-eval" {
		t.Fatalf("unexpected evaluation lineage: %#v", asset)
	}
}

func TestServerSyncsFailedModelEvaluationResult(t *testing.T) {
	reader := &recordingApplicationReader{
		logs: map[string]*observe.Logs{
			"sock-shop/customer-sft-demo-eval": {
				Pod:  "customer-sft-demo-eval-pod",
				Logs: `AI_RESULT_JSON={"modelURI":"inline://models/customer-sft-demo/v1","metrics":{"accuracy":0.42},"summary":"evaluation-failed"}`,
			},
		},
	}
	artifacts := &recordingArtifactStore{
		list: []observe.ModelArtifact{
			{
				Namespace:        "sock-shop",
				Name:             "customer-sft-demo",
				ModelURI:         "inline://models/customer-sft-demo/v1",
				EvaluationStatus: "pending",
			},
		},
	}
	server := NewServerWithOptions(Options{Reader: reader, Artifacts: artifacts})
	body := strings.NewReader(`{"evaluationJobName":"customer-sft-demo-eval"}`)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/models/sock-shop/customer-sft-demo/sync-evaluation", body))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if len(artifacts.saved) != 1 {
		t.Fatalf("saved artifacts = %d, want 1", len(artifacts.saved))
	}
	asset := artifacts.saved[0]
	if asset.EvaluationStatus != "failed" || asset.Status != "evaluated" || asset.Metrics["accuracy"] != 0.42 {
		t.Fatalf("unexpected failed evaluation asset: %#v", asset)
	}
	if asset.Summary != "evaluation-failed" || asset.SourceApplication != "customer-sft-demo-eval" {
		t.Fatalf("unexpected failed evaluation lineage: %#v", asset)
	}
}

func TestServerListReturnsUnavailableWhenReaderIsNotConfigured(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/applications", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusServiceUnavailable, recorder.Body.String())
	}
}

func TestServerDeploysDomainYAMLWithDryRun(t *testing.T) {
	applier := &recordingApplicationApplier{}
	server := NewServerWithOptions(Options{Applier: applier})
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/applications?dryRun=true", bytes.NewReader([]byte(validAIServiceYAML()))))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var result DeployResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if result.Application.Name != "sentiment-demo" || result.Application.Namespace != "ai-demo" || !result.Application.DryRun {
		t.Fatalf("unexpected application result: %#v", result.Application)
	}
	if result.Normalized.Kind != "AIService" || result.Normalized.GovernanceIntent.Tenant != "demo-tenant" {
		t.Fatalf("unexpected normalized payload: %#v", result.Normalized)
	}
	if !applier.called || applier.namespace != "ai-demo" || applier.name != "sentiment-demo" {
		t.Fatalf("unexpected applier call: %#v", applier)
	}
	if len(applier.options.DryRun) != 1 || applier.options.DryRun[0] != metav1.DryRunAll {
		t.Fatalf("expected server dry-run option, got %#v", applier.options.DryRun)
	}
	if !strings.Contains(string(applier.content), "kind: Application") || !strings.Contains(string(applier.content), "type: ai-service") {
		t.Fatalf("expected translated Application YAML, got:\n%s", string(applier.content))
	}
}

func TestServerDeployReturnsUnavailableWhenApplyIsNotConfigured(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/applications", bytes.NewReader([]byte(validAIServiceYAML()))))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusServiceUnavailable, recorder.Body.String())
	}
	var payload errorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if !strings.Contains(payload.Error, "deployment is not configured") {
		t.Fatalf("unexpected error: %#v", payload)
	}
}

func TestServerValidatesDomainYAML(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/validate", bytes.NewReader([]byte(validAIJobYAML()))))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var result domain.ValidationResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if len(result.Errors) != 0 {
		t.Fatalf("expected no validation errors, got %#v", result.Errors)
	}
}

func TestServerReturnsBadRequestForInvalidDomainYAML(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/validate", bytes.NewReader([]byte(`
apiVersion: ai.oam.dev/v1alpha1
kind: AIJob
metadata:
  name: invalid
spec:
  properties:
    image: busybox:1.36
`))))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	var payload errorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if !strings.Contains(payload.Error, "spec.properties.jobKind is required") {
		t.Fatalf("expected jobKind error, got %#v", payload)
	}
}

func TestServerNormalizesDomainYAML(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/normalize", bytes.NewReader([]byte(validAIServiceYAML()))))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var normalized domain.NormalizedObject
	if err := json.Unmarshal(recorder.Body.Bytes(), &normalized); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if normalized.Kind != "AIService" || normalized.WorkloadType != "service" {
		t.Fatalf("unexpected normalized identity: %#v", normalized)
	}
	if normalized.GovernanceIntent.Tenant != "demo-tenant" || normalized.WorkloadIntent.Service.Model.Name != "sentiment" {
		t.Fatalf("unexpected normalized intent: %#v", normalized)
	}
}

func TestServerRejectsUnsupportedMethod(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/normalize", nil))

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}

func validAIServiceYAML() string {
	return `
apiVersion: ai.oam.dev/v1alpha1
kind: AIService
metadata:
  name: sentiment-demo
  namespace: ai-demo
spec:
  componentName: sentiment-api
  properties:
    image: hashicorp/http-echo:0.2.3
    model:
      name: sentiment
  runtime:
    runtime: http
    tenant: demo-tenant
`
}

func validAIJobYAML() string {
	return `
apiVersion: ai.oam.dev/v1alpha1
kind: AIJob
metadata:
  name: evaluator-demo
spec:
  componentName: batch-evaluator
  properties:
    image: busybox:1.36
    jobKind: evaluation
  runtime:
    runtime: batch
`
}

type recordingApplicationApplier struct {
	called    bool
	namespace string
	name      string
	content   []byte
	options   metav1.PatchOptions
}

func (a *recordingApplicationApplier) ApplyApplication(_ context.Context, namespace, name string, content []byte, opts metav1.PatchOptions) (*unstructured.Unstructured, error) {
	a.called = true
	a.namespace = namespace
	a.name = name
	a.content = append([]byte(nil), content...)
	a.options = opts
	return &unstructured.Unstructured{}, nil
}

var _ domainapply.ApplicationApplier = (*recordingApplicationApplier)(nil)

type recordingApplicationReader struct {
	items         []observe.ApplicationListItem
	summaries     map[string]*observe.Summary
	logs          map[string]*observe.Logs
	listNamespace string
	logRequest    observe.LogOptions
}

func (r *recordingApplicationReader) ListApplications(_ context.Context, namespace string) ([]observe.ApplicationListItem, error) {
	r.listNamespace = namespace
	return r.items, nil
}

func (r *recordingApplicationReader) SummarizeApplication(_ context.Context, namespace, name string) (*observe.Summary, error) {
	key := namespace + "/" + name
	summary := r.summaries[key]
	if summary == nil {
		return nil, fmt.Errorf("not found")
	}
	return summary, nil
}

func (r *recordingApplicationReader) GetApplicationLogs(_ context.Context, options observe.LogOptions) (*observe.Logs, error) {
	r.logRequest = options
	key := options.Namespace + "/" + options.Name
	logs := r.logs[key]
	if logs == nil {
		return nil, fmt.Errorf("not found")
	}
	return logs, nil
}

type recordingApplicationProber struct {
	request observe.ProbeOptions
	result  *observe.ProbeResult
	err     error
}

func (p *recordingApplicationProber) ProbeApplication(_ context.Context, options observe.ProbeOptions) (*observe.ProbeResult, error) {
	p.request = options
	if p.err != nil {
		return nil, p.err
	}
	return p.result, nil
}

type recordingApplicationManager struct {
	deleted   string
	restarted string
	rerun     string
}

func (m *recordingApplicationManager) DeleteApplication(_ context.Context, namespace, name string) (*observe.LifecycleResult, error) {
	m.deleted = namespace + "/" + name
	return &observe.LifecycleResult{Action: "delete", Namespace: namespace, Name: name, Message: "delete requested"}, nil
}

func (m *recordingApplicationManager) RestartApplication(_ context.Context, namespace, name string) (*observe.LifecycleResult, error) {
	m.restarted = namespace + "/" + name
	return &observe.LifecycleResult{Action: "restart", Namespace: namespace, Name: name, Message: "restart requested"}, nil
}

func (m *recordingApplicationManager) RerunApplication(_ context.Context, namespace, name string) (*observe.LifecycleResult, error) {
	m.rerun = namespace + "/" + name
	return &observe.LifecycleResult{Action: "rerun", Namespace: namespace, Name: name, Message: "rerun requested"}, nil
}

type recordingAuditStore struct {
	events        []observe.AuditEvent
	listNamespace string
}

type recordingArtifactStore struct {
	saved         []observe.ModelArtifact
	list          []observe.ModelArtifact
	listNamespace string
}

type recordingDatasetStore struct {
	saved         []observe.DatasetArtifact
	list          []observe.DatasetArtifact
	listNamespace string
}

func (s *recordingArtifactStore) SaveArtifact(_ context.Context, artifact observe.ModelArtifact) error {
	s.saved = append(s.saved, artifact)
	return nil
}

func (s *recordingArtifactStore) ListArtifacts(_ context.Context, namespace string) ([]observe.ModelArtifact, error) {
	s.listNamespace = namespace
	return s.list, nil
}

func (s *recordingDatasetStore) SaveDataset(_ context.Context, dataset observe.DatasetArtifact) error {
	s.saved = append(s.saved, dataset)
	return nil
}

func (s *recordingDatasetStore) ListDatasets(_ context.Context, namespace string) ([]observe.DatasetArtifact, error) {
	s.listNamespace = namespace
	return s.list, nil
}

func (s *recordingAuditStore) RecordAudit(_ context.Context, event observe.AuditEvent) error {
	s.events = append(s.events, event)
	return nil
}

func (s *recordingAuditStore) ListAudits(_ context.Context, namespace string) ([]observe.AuditEvent, error) {
	s.listNamespace = namespace
	return s.events, nil
}
