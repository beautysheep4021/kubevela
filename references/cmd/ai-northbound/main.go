package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	domainapply "github.com/oam-dev/kubevela/pkg/ai/domain/apply"
	"github.com/oam-dev/kubevela/pkg/ai/northbound"
	"github.com/oam-dev/kubevela/pkg/ai/observe"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type args struct {
	addr       string
	kubeconfig string
}

func main() {
	parsed, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	options, err := buildServerOptions(parsed)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	server := &http.Server{
		Addr:    parsed.addr,
		Handler: northbound.NewServerWithOptions(options),
	}
	fmt.Fprintf(os.Stderr, "ai-northbound listening on %s\n", parsed.addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "serve ai-northbound: %v\n", err)
		os.Exit(1)
	}
}

func parseArgs(raw []string) (args, error) {
	flags := flag.NewFlagSet("ai-northbound", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	addr := flags.String("addr", "127.0.0.1:8088", "HTTP listen address")
	kubeconfig := flags.String("kubeconfig", "", "path to kubeconfig used by the deployment endpoint")
	if err := flags.Parse(raw); err != nil {
		return args{}, err
	}
	return args{addr: *addr, kubeconfig: *kubeconfig}, nil
}

func buildServerOptions(parsed args) (northbound.Options, error) {
	config, err := loadRestConfig(parsed.kubeconfig)
	if err != nil {
		if parsed.kubeconfig != "" {
			return northbound.Options{}, err
		}
		fmt.Fprintf(os.Stderr, "ai-northbound deployment endpoint disabled: %v\n", err)
		return northbound.Options{}, nil
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return northbound.Options{}, fmt.Errorf("create dynamic client: %w", err)
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return northbound.Options{}, fmt.Errorf("create kubernetes client: %w", err)
	}
	reader := observe.NewClientWithKube(client, kubeClient)
	return northbound.Options{
		Applier:   domainapply.DynamicApplicationApplier{Client: client},
		Reader:    reader,
		Manager:   reader,
		Prober:    reader,
		Artifacts: reader,
		Audits:    reader,
	}, nil
}

func loadRestConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("load kubeconfig %s: %w", kubeconfig, err)
		}
		return config, nil
	}
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}
	config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}
	return config, nil
}
