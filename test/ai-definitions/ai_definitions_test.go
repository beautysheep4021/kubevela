package aidefinitions

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/parser"
	"github.com/kubevela/pkg/cue/cuex"
	yamlv3 "go.yaml.in/yaml/v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlv2 "sigs.k8s.io/yaml"

	pkgdef "github.com/oam-dev/kubevela/pkg/definition"
)

const (
	helmDefinitionNamespacePlaceholder = "###HELM_NAMESPACE###"
	helmDefinitionNamespaceTemplate    = "{{ include \"systemDefinitionNamespace\" . }}"
)

var aiDefinitionPairs = map[string]string{
	"vela-templates/definitions/internal/component/ai-service.cue": "charts/vela-core/templates/defwithtemplate/ai-service.yaml",
	"vela-templates/definitions/internal/component/ai-job.cue":     "charts/vela-core/templates/defwithtemplate/ai-job.yaml",
	"vela-templates/definitions/internal/trait/gpu-resource.cue":   "charts/vela-core/templates/defwithtemplate/gpu-resource.yaml",
	"vela-templates/definitions/internal/trait/ai-runtime.cue":     "charts/vela-core/templates/defwithtemplate/ai-runtime.yaml",
}

func TestAIExtensionDefinitionsExist(t *testing.T) {
	root := projectRoot(t)
	paths := []string{
		"docs/examples/ai-platform/ai-service-app.yaml",
		"docs/examples/ai-platform/ai-job-app.yaml",
		"docs/examples/ai-platform/ai-governance-app.yaml",
		"docs/examples/ai-platform/README.md",
	}
	for cuePath, chartPath := range aiDefinitionPairs {
		paths = append(paths, cuePath, chartPath)
	}

	for _, path := range paths {
		path := path
		t.Run(path, func(t *testing.T) {
			if _, err := os.Stat(filepath.Join(root, path)); err != nil {
				t.Fatalf("expected %s to exist: %v", path, err)
			}
		})
	}
}

func TestAIExtensionChartManifestsMatchGeneratedDefinitions(t *testing.T) {
	disableExternalCUEPackagesForTest(t)
	root := projectRoot(t)

	for cuePath, chartPath := range aiDefinitionPairs {
		cuePath := cuePath
		chartPath := chartPath
		t.Run(chartPath, func(t *testing.T) {
			actual := readFile(t, filepath.Join(root, chartPath))
			expected := renderHelmChartDefinition(t, readFile(t, filepath.Join(root, cuePath)), cuePath)
			if actual != expected {
				t.Fatalf("expected %s to match generated definition from %s", chartPath, cuePath)
			}
		})
	}
}

func TestAIExtensionDefinitionsUseKubeVelaNativeModel(t *testing.T) {
	root := projectRoot(t)

	componentChecks := map[string][]string{
		"vela-templates/definitions/internal/component/ai-service.cue": {
			`"ai-service": {`,
			`type: "component"`,
			`kind:       "Deployment"`,
			`model: {`,
			`endpoint: {`,
			`labels?: [string]: string`,
			`annotations?: [string]: string`,
			`replicas: *1 | int & >=1`,
			`strategy?: {...}`,
			`minReadySeconds?: int & >=0`,
			`progressDeadlineSeconds?: int & >=1`,
			`revisionHistoryLimit?: int & >=0`,
			`imagePullPolicy?: "Always" | "Never" | "IfNotPresent"`,
			`serviceAccountName?: string`,
			`imagePullSecrets?: [...{`,
			`priorityClassName?: string`,
			`runtimeClassName?: string`,
			`schedulerName?: string`,
			`valueFrom?: {...}`,
			`envFrom?: [...{...}]`,
			`readinessProbe?: {...}`,
			`livenessProbe?: {...}`,
			`startupProbe?: {...}`,
			`volumes?: [...{...}]`,
			`volumeMounts?: [...{...}]`,
			`securityContext?: {...}`,
			`containerSecurityContext?: {...}`,
			`terminationGracePeriodSeconds?: int & >=0`,
			`automountServiceAccountToken?: bool`,
			`lifecycle?: {...}`,
		},
		"vela-templates/definitions/internal/component/ai-job.cue": {
			`"ai-job": {`,
			`type: "component"`,
			`kind:       "Job"`,
			`jobKind:`,
			`dataset: {`,
			`parallelism: *1 | int & >=1`,
			`completions: *1 | int & >=1`,
			`backoffLimit: *3 | int & >=0`,
			`activeDeadlineSeconds: *3600 | int & >=1`,
			`completionMode?: "NonIndexed" | "Indexed"`,
			`ttlSecondsAfterFinished?: int & >=0`,
			`suspend?: bool`,
			`imagePullPolicy?: "Always" | "Never" | "IfNotPresent"`,
			`serviceAccountName?: string`,
			`imagePullSecrets?: [...{`,
			`priorityClassName?: string`,
			`runtimeClassName?: string`,
			`schedulerName?: string`,
			`valueFrom?: {...}`,
			`envFrom?: [...{...}]`,
			`volumes?: [...{...}]`,
			`volumeMounts?: [...{...}]`,
			`securityContext?: {...}`,
			`containerSecurityContext?: {...}`,
			`terminationGracePeriodSeconds?: int & >=0`,
			`automountServiceAccountToken?: bool`,
			`lifecycle?: {...}`,
		},
	}

	for path, expectedFragments := range componentChecks {
		path := path
		expectedFragments := expectedFragments
		t.Run(path, func(t *testing.T) {
			content := readFile(t, filepath.Join(root, path))
			for _, fragment := range expectedFragments {
				if !strings.Contains(content, fragment) {
					t.Fatalf("expected %s to contain %q", path, fragment)
				}
			}
		})
	}

	trait := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/trait/gpu-resource.cue"))
	for _, fragment := range []string{
		`"gpu-resource": {`,
		`type: "trait"`,
		`appliesToWorkloads: ["deployments.apps", "jobs.batch"]`,
		`"nvidia.com/gpu"`,
		`count: *1 | int & >=1`,
	} {
		if !strings.Contains(trait, fragment) {
			t.Fatalf("expected gpu-resource trait to contain %q", fragment)
		}
	}

	runtimeTrait := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/trait/ai-runtime.cue"))
	for _, fragment := range []string{
		`"ai-runtime": {`,
		`type: "trait"`,
		`appliesToWorkloads: ["deployments.apps", "jobs.batch"]`,
		`"ai.oam.dev/runtime"`,
		`"ai.oam.dev/tenant"`,
		`"ai.oam.dev/project"`,
		`"ai.oam.dev/environment"`,
		`"ai.oam.dev/model-uri"`,
		`"ai.oam.dev/dataset-uri"`,
	} {
		if !strings.Contains(runtimeTrait, fragment) {
			t.Fatalf("expected ai-runtime trait to contain %q", fragment)
		}
	}
}

func TestAIServiceExposesNativeServiceLabelsPassThrough(t *testing.T) {
	root := projectRoot(t)
	service := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/component/ai-service.cue"))

	for _, fragment := range []string{
		`if parameter.endpoint.labels != _|_ {`,
		`parameter.endpoint.labels`,
		`labels?: [string]: string`,
	} {
		if !strings.Contains(service, fragment) {
			t.Fatalf("expected ai-service Service label support to contain %q", fragment)
		}
	}

	job := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/component/ai-job.cue"))
	for _, fragment := range []string{
		`endpoint.labels`,
	} {
		if strings.Contains(job, fragment) {
			t.Fatalf("expected ai-job to avoid Service endpoint labels surface, found %q", fragment)
		}
	}
}

func TestAIServiceExposesNativeServiceAnnotationsPassThrough(t *testing.T) {
	root := projectRoot(t)
	service := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/component/ai-service.cue"))

	for _, fragment := range []string{
		`if parameter.endpoint.annotations != _|_ {`,
		`annotations: parameter.endpoint.annotations`,
		`annotations?: [string]: string`,
	} {
		if !strings.Contains(service, fragment) {
			t.Fatalf("expected ai-service Service annotation support to contain %q", fragment)
		}
	}

	job := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/component/ai-job.cue"))
	for _, fragment := range []string{
		`endpoint.annotations`,
		`annotations: parameter.endpoint.annotations`,
	} {
		if strings.Contains(job, fragment) {
			t.Fatalf("expected ai-job to avoid Service endpoint annotations surface, found %q", fragment)
		}
	}
}

func TestAIJobExposesNativeCompletionModePassThrough(t *testing.T) {
	root := projectRoot(t)
	job := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/component/ai-job.cue"))

	for _, fragment := range []string{
		`if parameter.completionMode != _|_ {`,
		`completionMode: parameter.completionMode`,
		`completionMode?: "NonIndexed" | "Indexed"`,
	} {
		if !strings.Contains(job, fragment) {
			t.Fatalf("expected ai-job completionMode support to contain %q", fragment)
		}
	}

	service := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/component/ai-service.cue"))
	for _, fragment := range []string{
		`completionMode?:`,
		`completionMode: parameter.completionMode`,
	} {
		if strings.Contains(service, fragment) {
			t.Fatalf("expected ai-service to avoid Job completionMode surface, found %q", fragment)
		}
	}
}

func TestAIServiceExposesNativeRevisionHistoryLimitPassThrough(t *testing.T) {
	root := projectRoot(t)
	service := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/component/ai-service.cue"))

	for _, fragment := range []string{
		`if parameter.revisionHistoryLimit != _|_ {`,
		`revisionHistoryLimit: parameter.revisionHistoryLimit`,
		`revisionHistoryLimit?: int & >=0`,
	} {
		if !strings.Contains(service, fragment) {
			t.Fatalf("expected ai-service revisionHistoryLimit support to contain %q", fragment)
		}
	}

	job := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/component/ai-job.cue"))
	for _, fragment := range []string{
		`revisionHistoryLimit?:`,
		`revisionHistoryLimit: parameter.revisionHistoryLimit`,
	} {
		if strings.Contains(job, fragment) {
			t.Fatalf("expected ai-job to avoid Deployment revisionHistoryLimit surface, found %q", fragment)
		}
	}
}

func TestAIServiceExposesNativeProgressDeadlineSecondsPassThrough(t *testing.T) {
	root := projectRoot(t)
	service := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/component/ai-service.cue"))

	for _, fragment := range []string{
		`if parameter.progressDeadlineSeconds != _|_ {`,
		`progressDeadlineSeconds: parameter.progressDeadlineSeconds`,
		`progressDeadlineSeconds?: int & >=1`,
	} {
		if !strings.Contains(service, fragment) {
			t.Fatalf("expected ai-service progressDeadlineSeconds support to contain %q", fragment)
		}
	}

	job := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/component/ai-job.cue"))
	for _, fragment := range []string{
		`progressDeadlineSeconds?:`,
		`progressDeadlineSeconds: parameter.progressDeadlineSeconds`,
	} {
		if strings.Contains(job, fragment) {
			t.Fatalf("expected ai-job to avoid Deployment progressDeadlineSeconds surface, found %q", fragment)
		}
	}
}

func TestAIServiceExposesNativeMinReadySecondsPassThrough(t *testing.T) {
	root := projectRoot(t)
	service := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/component/ai-service.cue"))

	for _, fragment := range []string{
		`if parameter.minReadySeconds != _|_ {`,
		`minReadySeconds: parameter.minReadySeconds`,
		`minReadySeconds?: int & >=0`,
	} {
		if !strings.Contains(service, fragment) {
			t.Fatalf("expected ai-service minReadySeconds support to contain %q", fragment)
		}
	}

	job := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/component/ai-job.cue"))
	for _, fragment := range []string{
		`minReadySeconds?:`,
		`minReadySeconds: parameter.minReadySeconds`,
	} {
		if strings.Contains(job, fragment) {
			t.Fatalf("expected ai-job to avoid Deployment minReadySeconds surface, found %q", fragment)
		}
	}
}

func TestAIServiceExposesNativeDeploymentStrategyPassThrough(t *testing.T) {
	root := projectRoot(t)
	service := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/component/ai-service.cue"))

	for _, fragment := range []string{
		`if parameter.strategy != _|_ {`,
		`strategy: parameter.strategy`,
		`strategy?: {...}`,
	} {
		if !strings.Contains(service, fragment) {
			t.Fatalf("expected ai-service Deployment strategy support to contain %q", fragment)
		}
	}

	job := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/component/ai-job.cue"))
	for _, fragment := range []string{
		`strategy?:`,
		`strategy: parameter.strategy`,
	} {
		if strings.Contains(job, fragment) {
			t.Fatalf("expected ai-job to avoid Deployment strategy surface, found %q", fragment)
		}
	}
}

func TestAIComponentsExposeNativeContainerLifecyclePassThrough(t *testing.T) {
	root := projectRoot(t)

	for _, path := range []string{
		"vela-templates/definitions/internal/component/ai-service.cue",
		"vela-templates/definitions/internal/component/ai-job.cue",
	} {
		path := path
		t.Run(path, func(t *testing.T) {
			content := readFile(t, filepath.Join(root, path))
			for _, fragment := range []string{
				`if parameter.lifecycle != _|_ {`,
				`lifecycle: parameter.lifecycle`,
				`lifecycle?: {...}`,
			} {
				if !strings.Contains(content, fragment) {
					t.Fatalf("expected %s native container lifecycle support to contain %q", path, fragment)
				}
			}
		})
	}
}

func TestAIComponentsExposeNativeServiceAccountTokenMountControl(t *testing.T) {
	root := projectRoot(t)

	for _, path := range []string{
		"vela-templates/definitions/internal/component/ai-service.cue",
		"vela-templates/definitions/internal/component/ai-job.cue",
	} {
		path := path
		t.Run(path, func(t *testing.T) {
			content := readFile(t, filepath.Join(root, path))
			for _, fragment := range []string{
				`if parameter.automountServiceAccountToken != _|_ {`,
				`automountServiceAccountToken: parameter.automountServiceAccountToken`,
				`automountServiceAccountToken?: bool`,
			} {
				if !strings.Contains(content, fragment) {
					t.Fatalf("expected %s native service account token mount support to contain %q", path, fragment)
				}
			}
		})
	}
}

func TestAIComponentsExposeNativeEnvFromPassThrough(t *testing.T) {
	root := projectRoot(t)

	for _, path := range []string{
		"vela-templates/definitions/internal/component/ai-service.cue",
		"vela-templates/definitions/internal/component/ai-job.cue",
	} {
		path := path
		t.Run(path, func(t *testing.T) {
			content := readFile(t, filepath.Join(root, path))
			for _, fragment := range []string{
				`if parameter.envFrom != _|_ {`,
				`envFrom: parameter.envFrom`,
				`envFrom?: [...{...}]`,
			} {
				if !strings.Contains(content, fragment) {
					t.Fatalf("expected %s native envFrom support to contain %q", path, fragment)
				}
			}
		})
	}
}

func TestAIComponentsExposeNativeTerminationGracePeriod(t *testing.T) {
	root := projectRoot(t)

	for _, path := range []string{
		"vela-templates/definitions/internal/component/ai-service.cue",
		"vela-templates/definitions/internal/component/ai-job.cue",
	} {
		path := path
		t.Run(path, func(t *testing.T) {
			content := readFile(t, filepath.Join(root, path))
			for _, fragment := range []string{
				`if parameter.terminationGracePeriodSeconds != _|_ {`,
				`terminationGracePeriodSeconds: parameter.terminationGracePeriodSeconds`,
				`terminationGracePeriodSeconds?: int & >=0`,
			} {
				if !strings.Contains(content, fragment) {
					t.Fatalf("expected %s native termination grace period support to contain %q", path, fragment)
				}
			}
		})
	}
}

func TestAIComponentsExposeNativeSecurityContextPassThrough(t *testing.T) {
	root := projectRoot(t)

	for _, path := range []string{
		"vela-templates/definitions/internal/component/ai-service.cue",
		"vela-templates/definitions/internal/component/ai-job.cue",
	} {
		path := path
		t.Run(path, func(t *testing.T) {
			content := readFile(t, filepath.Join(root, path))
			for _, fragment := range []string{
				`if parameter.securityContext != _|_ {`,
				`securityContext: parameter.securityContext`,
				`if parameter.containerSecurityContext != _|_ {`,
				`securityContext: parameter.containerSecurityContext`,
				`securityContext?: {...}`,
				`containerSecurityContext?: {...}`,
			} {
				if !strings.Contains(content, fragment) {
					t.Fatalf("expected %s native security context support to contain %q", path, fragment)
				}
			}
		})
	}
}

func TestAIComponentsExposeNativeVolumePassThrough(t *testing.T) {
	root := projectRoot(t)

	for _, path := range []string{
		"vela-templates/definitions/internal/component/ai-service.cue",
		"vela-templates/definitions/internal/component/ai-job.cue",
	} {
		path := path
		t.Run(path, func(t *testing.T) {
			content := readFile(t, filepath.Join(root, path))
			for _, fragment := range []string{
				`if parameter.volumes != _|_ {`,
				`volumes: parameter.volumes`,
				`if parameter.volumeMounts != _|_ {`,
				`volumeMounts: parameter.volumeMounts`,
				`volumes?: [...{...}]`,
				`volumeMounts?: [...{...}]`,
			} {
				if !strings.Contains(content, fragment) {
					t.Fatalf("expected %s native volume support to contain %q", path, fragment)
				}
			}
		})
	}
}

func TestAIServiceExposesNativeHealthProbePassThrough(t *testing.T) {
	root := projectRoot(t)
	service := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/component/ai-service.cue"))

	for _, fragment := range []string{
		`if parameter.readinessProbe != _|_ {`,
		`readinessProbe: parameter.readinessProbe`,
		`if parameter.livenessProbe != _|_ {`,
		`livenessProbe: parameter.livenessProbe`,
		`if parameter.startupProbe != _|_ {`,
		`startupProbe: parameter.startupProbe`,
		`readinessProbe?: {...}`,
		`livenessProbe?: {...}`,
		`startupProbe?: {...}`,
	} {
		if !strings.Contains(service, fragment) {
			t.Fatalf("expected ai-service health probe support to contain %q", fragment)
		}
	}

	job := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/component/ai-job.cue"))
	for _, fragment := range []string{
		`readinessProbe?:`,
		`livenessProbe?:`,
		`startupProbe?:`,
	} {
		if strings.Contains(job, fragment) {
			t.Fatalf("expected ai-job to avoid service-oriented health probe surface, found %q", fragment)
		}
	}
}

func TestAIComponentsExposeManagementLabelsOnTopLevelResources(t *testing.T) {
	root := projectRoot(t)

	service := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/component/ai-service.cue"))
	for _, fragment := range []string{
		`metadata: labels: aiManagementLabels`,
		`"ai.oam.dev/workload-kind": "service"`,
		`"ai.oam.dev/model":         parameter.model.name`,
		`metadata: {`,
		`labels: aiManagementLabels`,
	} {
		if !strings.Contains(service, fragment) {
			t.Fatalf("expected ai-service top-level resources to contain %q", fragment)
		}
	}

	job := readFile(t, filepath.Join(root, "vela-templates/definitions/internal/component/ai-job.cue"))
	for _, fragment := range []string{
		`let aiManagementLabels = {`,
		`"ai.oam.dev/workload-kind": "job"`,
		`"ai.oam.dev/job-kind":      parameter.jobKind`,
		`labels: aiManagementLabels`,
	} {
		if !strings.Contains(job, fragment) {
			t.Fatalf("expected ai-job top-level resource to contain %q", fragment)
		}
	}
}

func TestAIExtensionDefinitionsAreValidCUE(t *testing.T) {
	root := projectRoot(t)
	paths := []string{
		"vela-templates/definitions/internal/component/ai-service.cue",
		"vela-templates/definitions/internal/component/ai-job.cue",
		"vela-templates/definitions/internal/trait/gpu-resource.cue",
		"vela-templates/definitions/internal/trait/ai-runtime.cue",
	}

	for _, path := range paths {
		path := path
		t.Run(path, func(t *testing.T) {
			content := readFile(t, filepath.Join(root, path))
			file, err := parser.ParseFile(path, content, parser.ParseComments)
			if err != nil {
				t.Fatalf("expected %s to parse as CUE: %v", path, err)
			}
			if _, err := format.Node(file, format.Simplify()); err != nil {
				t.Fatalf("expected %s to format as CUE: %v", path, err)
			}
		})
	}
}

func TestAIExtensionDefinitionsRenderWithKubeVelaDefinitionParser(t *testing.T) {
	disableExternalCUEPackagesForTest(t)
	root := projectRoot(t)

	for path := range aiDefinitionPairs {
		path := path
		t.Run(path, func(t *testing.T) {
			content := readFile(t, filepath.Join(root, path))
			def := pkgdef.Definition{Unstructured: unstructured.Unstructured{}}
			if err := def.FromCUEString(content, nil); err != nil {
				t.Fatalf("expected %s to render through KubeVela definition parser: %v", path, err)
			}
		})
	}
}

func TestAIExamplesStayOnApplicationAPI(t *testing.T) {
	root := projectRoot(t)
	examples := map[string]string{
		"docs/examples/ai-platform/ai-service-app.yaml": "ai-service",
		"docs/examples/ai-platform/ai-job-app.yaml":     "ai-job",
	}

	for path, componentType := range examples {
		path := path
		componentType := componentType
		t.Run(path, func(t *testing.T) {
			content := readFile(t, filepath.Join(root, path))
			for _, fragment := range []string{
				"apiVersion: core.oam.dev/v1beta1",
				"kind: Application",
				"type: " + componentType,
				"type: gpu-resource",
				"type: ai-runtime",
			} {
				if !strings.Contains(content, fragment) {
					t.Fatalf("expected %s to contain %q", path, fragment)
				}
			}
			if strings.Contains(content, "kind: AIService") || strings.Contains(content, "kind: AIJob") || strings.Contains(content, "kind: AIWorkflow") {
				t.Fatalf("%s must not introduce AI top-level CRDs", path)
			}

			var doc map[string]interface{}
			if err := yamlv2.Unmarshal([]byte(content), &doc); err != nil {
				t.Fatalf("expected %s to be valid YAML: %v", path, err)
			}
		})
	}
}

func TestAIServiceExampleShowsMinimalProductionHooks(t *testing.T) {
	root := projectRoot(t)
	path := "docs/examples/ai-platform/ai-service-app.yaml"
	content := readFile(t, filepath.Join(root, path))

	for _, fragment := range []string{
		"strategy:",
		"imagePullPolicy:",
		"minReadySeconds:",
		"progressDeadlineSeconds:",
		"revisionHistoryLimit:",
		"labels:",
		"annotations:",
		"envFrom:",
		"readinessProbe:",
		"securityContext:",
		"containerSecurityContext:",
		"automountServiceAccountToken:",
		"terminationGracePeriodSeconds:",
	} {
		if !strings.Contains(content, fragment) {
			t.Fatalf("expected %s to demonstrate %q", path, fragment)
		}
	}
	if strings.Contains(content, "kind: AIService") || strings.Contains(content, "kind: AIJob") || strings.Contains(content, "kind: AIWorkflow") {
		t.Fatalf("%s must remain a standard KubeVela Application example", path)
	}
}

func TestAIJobExampleShowsMinimalProductionHooks(t *testing.T) {
	root := projectRoot(t)
	path := "docs/examples/ai-platform/ai-job-app.yaml"
	content := readFile(t, filepath.Join(root, path))

	for _, fragment := range []string{
		"completionMode:",
		"imagePullPolicy:",
		"ttlSecondsAfterFinished:",
		"suspend:",
		"envFrom:",
		"volumes:",
		"volumeMounts:",
		"securityContext:",
		"containerSecurityContext:",
		"automountServiceAccountToken:",
		"terminationGracePeriodSeconds:",
		"lifecycle:",
	} {
		if !strings.Contains(content, fragment) {
			t.Fatalf("expected %s to demonstrate %q", path, fragment)
		}
	}
	if strings.Contains(content, "kind: AIService") || strings.Contains(content, "kind: AIJob") || strings.Contains(content, "kind: AIWorkflow") {
		t.Fatalf("%s must remain a standard KubeVela Application example", path)
	}
}

func TestAIGovernanceExampleUsesNativePolicyControls(t *testing.T) {
	root := projectRoot(t)
	path := "docs/examples/ai-platform/ai-governance-app.yaml"
	content := readFile(t, filepath.Join(root, path))

	for _, fragment := range []string{
		"apiVersion: core.oam.dev/v1beta1",
		"kind: Application",
		"type: ai-service",
		"type: ai-job",
		"type: topology",
		"type: override",
		"type: deploy",
		`policies: ["local-topology", "service-ha-override"]`,
		`policies: ["local-topology", "job-batch-override"]`,
	} {
		if !strings.Contains(content, fragment) {
			t.Fatalf("expected %s to contain %q", path, fragment)
		}
	}
	if strings.Contains(content, "kind: AIService") || strings.Contains(content, "kind: AIJob") || strings.Contains(content, "kind: AIWorkflow") {
		t.Fatalf("%s must not introduce AI top-level CRDs", path)
	}

	var doc map[string]interface{}
	if err := yamlv2.Unmarshal([]byte(content), &doc); err != nil {
		t.Fatalf("expected %s to be valid YAML: %v", path, err)
	}
}

func TestAIScopeDoesNotAddOutOfScopeControlPlaneSurfaces(t *testing.T) {
	root := projectRoot(t)
	paths := []string{
		"vela-templates/definitions/internal/component/ai-service.cue",
		"vela-templates/definitions/internal/component/ai-job.cue",
		"vela-templates/definitions/internal/trait/gpu-resource.cue",
		"vela-templates/definitions/internal/trait/ai-runtime.cue",
		"docs/examples/ai-platform/ai-service-app.yaml",
		"docs/examples/ai-platform/ai-job-app.yaml",
		"docs/examples/ai-platform/ai-governance-app.yaml",
		"charts/vela-core/templates/defwithtemplate/ai-service.yaml",
		"charts/vela-core/templates/defwithtemplate/ai-job.yaml",
		"charts/vela-core/templates/defwithtemplate/gpu-resource.yaml",
		"charts/vela-core/templates/defwithtemplate/ai-runtime.yaml",
	}
	for _, path := range paths {
		path := path
		t.Run(path, func(t *testing.T) {
			content := readFile(t, filepath.Join(root, path))
			for _, forbidden := range []string{
				"WorkflowStepDefinition",
				"Controller",
				"Reconcile",
				"database",
				"Database",
				"SQL",
			} {
				if strings.Contains(content, forbidden) {
					t.Fatalf("expected %s to stay within the first PoC scope, found %q", path, forbidden)
				}
			}
		})
	}
}

func disableExternalCUEPackagesForTest(t *testing.T) {
	t.Helper()
	originalExternalPackageSetting := cuex.EnableExternalPackageForDefaultCompiler
	cuex.EnableExternalPackageForDefaultCompiler = false
	t.Cleanup(func() {
		cuex.EnableExternalPackageForDefaultCompiler = originalExternalPackageSetting
	})
}

func renderHelmChartDefinition(t *testing.T, cueContent string, sourcePath string) string {
	t.Helper()
	def := pkgdef.Definition{Unstructured: unstructured.Unstructured{}}
	if err := def.FromCUEString(cueContent, nil); err != nil {
		t.Fatalf("failed to render %s through KubeVela definition parser: %v", sourcePath, err)
	}
	def.SetNamespace(helmDefinitionNamespacePlaceholder)
	if len(def.GetLabels()) == 0 {
		def.SetLabels(nil)
	}

	rendered, err := prettyYAMLMarshal(def.Object)
	if err != nil {
		t.Fatalf("failed to marshal generated definition for %s: %v", sourcePath, err)
	}
	rendered = strings.ReplaceAll(rendered, "'"+helmDefinitionNamespacePlaceholder+"'", helmDefinitionNamespaceTemplate) + "\n"
	return "# Code generated by KubeVela templates. DO NOT EDIT. Please edit the original cue file.\n" +
		"# Definition source cue file: " + sourcePath + "\n" +
		rendered
}

func prettyYAMLMarshal(obj map[string]interface{}) (string, error) {
	var b bytes.Buffer
	encoder := yamlv3.NewEncoder(&b)
	encoder.SetIndent(2)
	err := encoder.Encode(&obj)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("failed to locate project root")
		}
		wd = parent
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(content)
}
