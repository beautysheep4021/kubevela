package aiapply

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	domainapply "github.com/oam-dev/kubevela/pkg/ai/domain/apply"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestApplyYAMLDryRunUsesServerDryRun(t *testing.T) {
	applier := &recordingApplier{}

	result, err := domainapply.ApplyYAMLWithApplier(context.Background(), applier, []byte(`
apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
  name: dry-run-app
  namespace: ai-demo
spec:
  components: []
`), domainapply.Options{DryRun: true})
	if err != nil {
		t.Fatalf("ApplyYAMLWithApplier returned error: %v", err)
	}
	if result.Name != "dry-run-app" || result.Namespace != "ai-demo" || !result.DryRun {
		t.Fatalf("unexpected result: %#v", result)
	}
	if applier.namespace != "ai-demo" || applier.name != "dry-run-app" {
		t.Fatalf("unexpected apply target: %#v", applier)
	}
	if len(applier.options.DryRun) != 1 || applier.options.DryRun[0] != metav1.DryRunAll {
		t.Fatalf("expected server dry-run option, got %#v", applier.options.DryRun)
	}
}

func TestApplyYAMLAppliesWhenDryRunDisabled(t *testing.T) {
	applier := &recordingApplier{}

	result, err := domainapply.ApplyYAMLWithApplier(context.Background(), applier, []byte(`
apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
  name: apply-app
  namespace: ai-demo
spec:
  components: []
`), domainapply.Options{})
	if err != nil {
		t.Fatalf("ApplyYAMLWithApplier returned error: %v", err)
	}
	if result.Name != "apply-app" || result.DryRun {
		t.Fatalf("unexpected result: %#v", result)
	}
	if len(applier.options.DryRun) != 0 {
		t.Fatalf("expected no dry-run option, got %#v", applier.options.DryRun)
	}
}

func TestApplyYAMLRejectsNonApplication(t *testing.T) {
	applier := &recordingApplier{}
	_, err := domainapply.ApplyYAMLWithApplier(context.Background(), applier, []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: not-app
`), domainapply.Options{})
	if err == nil {
		t.Fatal("expected non-Application YAML to be rejected")
	}
	if applier.called {
		t.Fatal("expected invalid object to be rejected before apply")
	}
}

func TestApplyCommandRequiresExplicitYesWithoutDryRun(t *testing.T) {
	root := projectRoot(t)
	cmd := exec.Command("go", "run", "./references/cmd/ai-domain", "apply", "-f", "docs/examples/ai-platform/domain/ai-job.yaml")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOCACHE=/tmp/kubevela-go-build-cache")

	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected apply without --dry-run or --yes to fail, output:\n%s", string(out))
	}
	if !strings.Contains(string(out), "requires --dry-run or --yes") {
		t.Fatalf("expected explicit confirmation error, got:\n%s", string(out))
	}
}

type recordingApplier struct {
	called    bool
	namespace string
	name      string
	content   []byte
	options   metav1.PatchOptions
}

func (a *recordingApplier) ApplyApplication(_ context.Context, namespace, name string, content []byte, opts metav1.PatchOptions) (*unstructured.Unstructured, error) {
	a.called = true
	a.namespace = namespace
	a.name = name
	a.content = append([]byte(nil), content...)
	a.options = opts
	return &unstructured.Unstructured{}, nil
}

func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("could not find project root containing go.mod")
		}
		wd = parent
	}
}
