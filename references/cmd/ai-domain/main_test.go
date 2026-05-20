package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestParseStatusArgsAllowsFlagsAfterApplicationName(t *testing.T) {
	args, err := parseStatusArgs([]string{"-n", "ai-demo", "ai-job-demo", "--kubeconfig", "/root/.kube/config"})
	if err != nil {
		t.Fatalf("parseStatusArgs returned error: %v", err)
	}
	if args.namespace != "ai-demo" {
		t.Fatalf("namespace = %q, want ai-demo", args.namespace)
	}
	if args.appName != "ai-job-demo" {
		t.Fatalf("appName = %q, want ai-job-demo", args.appName)
	}
	if args.kubeconfig != "/root/.kube/config" {
		t.Fatalf("kubeconfig = %q, want /root/.kube/config", args.kubeconfig)
	}
}

func TestParseStatusArgsSupportsEqualsSyntax(t *testing.T) {
	args, err := parseStatusArgs([]string{"--kubeconfig=/root/.kube/config", "-n=ai-demo", "ai-job-demo"})
	if err != nil {
		t.Fatalf("parseStatusArgs returned error: %v", err)
	}
	if args.namespace != "ai-demo" || args.appName != "ai-job-demo" || args.kubeconfig != "/root/.kube/config" {
		t.Fatalf("unexpected parsed args: %#v", args)
	}
}

func TestParseApplyArgsSupportsNamespaceOverride(t *testing.T) {
	args, err := parseApplyArgs([]string{"-f", "domain.yaml", "-n", "ai-demo", "--dry-run", "--kubeconfig", "/root/.kube/config"})
	if err != nil {
		t.Fatalf("parseApplyArgs returned error: %v", err)
	}
	if args.file != "domain.yaml" {
		t.Fatalf("file = %q, want domain.yaml", args.file)
	}
	if args.namespace != "ai-demo" {
		t.Fatalf("namespace = %q, want ai-demo", args.namespace)
	}
	if !args.dryRun {
		t.Fatal("expected dryRun to be true")
	}
	if args.kubeconfig != "/root/.kube/config" {
		t.Fatalf("kubeconfig = %q, want /root/.kube/config", args.kubeconfig)
	}
}

func TestSetApplicationNamespaceOverridesTranslatedNamespace(t *testing.T) {
	out, err := setApplicationNamespace([]byte(`
apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
  name: ai-demo
  namespace: old
spec:
  components: []
`), "ai-demo")
	if err != nil {
		t.Fatalf("setApplicationNamespace returned error: %v", err)
	}
	if !containsYAMLLine(string(out), "namespace: ai-demo") {
		t.Fatalf("expected namespace override in YAML, got:\n%s", string(out))
	}
	if containsYAMLLine(string(out), "namespace: old") {
		t.Fatalf("expected old namespace to be replaced, got:\n%s", string(out))
	}
}

func TestLoadRestConfigUsesDefaultKubeconfigWhenPathIsEmpty(t *testing.T) {
	path := writeTestKubeconfig(t, "https://127.0.0.1:6443")
	t.Setenv("KUBECONFIG", path)
	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	t.Setenv("KUBERNETES_SERVICE_PORT", "")

	config, err := loadRestConfig("")
	if err != nil {
		t.Fatalf("loadRestConfig returned error: %v", err)
	}
	if config.Host != "https://127.0.0.1:6443" {
		t.Fatalf("host = %q, want https://127.0.0.1:6443", config.Host)
	}
}

func writeTestKubeconfig(t *testing.T, server string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config")
	cfg := clientcmdapi.NewConfig()
	cfg.Clusters["local"] = &clientcmdapi.Cluster{Server: server}
	cfg.AuthInfos["user"] = &clientcmdapi.AuthInfo{Token: "token"}
	cfg.Contexts["ctx"] = &clientcmdapi.Context{Cluster: "local", AuthInfo: "user"}
	cfg.CurrentContext = "ctx"
	if err := clientcmd.WriteToFile(*cfg, path); err != nil {
		t.Fatalf("write kubeconfig: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat kubeconfig: %v", err)
	}
	return path
}

func containsYAMLLine(content, line string) bool {
	for _, candidate := range strings.Split(content, "\n") {
		if strings.TrimSpace(candidate) == line {
			return true
		}
	}
	return false
}
