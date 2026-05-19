package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/oam-dev/kubevela/pkg/ai/domain"
	domainapply "github.com/oam-dev/kubevela/pkg/ai/domain/apply"
	"github.com/oam-dev/kubevela/pkg/ai/observe"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "translate":
			runTranslate(os.Args[2:])
			return
		case "status":
			runStatus(os.Args[2:])
			return
		case "apply":
			runApply(os.Args[2:])
			return
		}
	}
	runTranslate(os.Args[1:])
}

func runTranslate(args []string) {
	file := flag.String("f", "", "AI domain YAML file to translate")
	flags := flag.NewFlagSet("translate", flag.ExitOnError)
	file = flags.String("f", "", "AI domain YAML file to translate")
	_ = flags.Parse(args)

	if *file == "" {
		fmt.Fprintln(os.Stderr, "-f is required")
		os.Exit(2)
	}

	in, err := os.ReadFile(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", *file, err)
		os.Exit(1)
	}

	out, err := domain.TranslateYAML(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "translate %s: %v\n", *file, err)
		os.Exit(1)
	}

	if _, err := os.Stdout.Write(out); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		os.Exit(1)
	}
}

func runStatus(args []string) {
	flags := flag.NewFlagSet("status", flag.ExitOnError)
	fromFiles := flags.String("from-files", "", "comma-separated Kubernetes object YAML files to summarize")
	namespace := flags.String("n", "default", "Application namespace for live cluster status")
	kubeconfig := flags.String("kubeconfig", "", "kubeconfig path for live cluster status")
	_ = flags.Parse(args)

	var (
		summary *observe.Summary
		err     error
	)
	if *fromFiles != "" {
		var objects []*unstructured.Unstructured
		objects, err = readObjectsFromFiles(*fromFiles)
		if err == nil {
			summary, err = observe.SummarizeObjects(objects)
		}
	} else {
		if flags.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "status requires an application name or --from-files")
			os.Exit(2)
		}
		summary, err = observe.SummarizeApplication(context.Background(), *kubeconfig, *namespace, flags.Arg(0))
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "status failed: %v\n", err)
		os.Exit(1)
	}
	out, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode status output: %v\n", err)
		os.Exit(1)
	}
	_, _ = os.Stdout.Write(append(out, '\n'))
}

func runApply(args []string) {
	flags := flag.NewFlagSet("apply", flag.ExitOnError)
	file := flags.String("f", "", "AI domain YAML file to translate and apply")
	kubeconfig := flags.String("kubeconfig", "", "kubeconfig path")
	dryRun := flags.Bool("dry-run", false, "perform server-side dry-run only")
	yes := flags.Bool("yes", false, "actually apply the generated Application")
	_ = flags.Parse(args)

	if *file == "" {
		fmt.Fprintln(os.Stderr, "-f is required")
		os.Exit(2)
	}
	if !*dryRun && !*yes {
		fmt.Fprintln(os.Stderr, "apply requires --dry-run or --yes")
		os.Exit(2)
	}
	in, err := os.ReadFile(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", *file, err)
		os.Exit(1)
	}
	appYAML, err := domain.TranslateYAML(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "translate %s: %v\n", *file, err)
		os.Exit(1)
	}
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load kubeconfig: %v\n", err)
		os.Exit(1)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create dynamic client: %v\n", err)
		os.Exit(1)
	}
	result, err := domainapply.ApplyYAML(context.Background(), client, appYAML, domainapply.Options{DryRun: *dryRun})
	if err != nil {
		fmt.Fprintf(os.Stderr, "apply failed: %v\n", err)
		os.Exit(1)
	}
	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode apply output: %v\n", err)
		os.Exit(1)
	}
	_, _ = os.Stdout.Write(append(out, '\n'))
}

func readObjectsFromFiles(files string) ([]*unstructured.Unstructured, error) {
	var objects []*unstructured.Unstructured
	for _, file := range strings.Split(files, ",") {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", file, err)
		}
		obj := &unstructured.Unstructured{}
		if err := yaml.Unmarshal(content, obj); err != nil {
			return nil, fmt.Errorf("decode %s: %w", file, err)
		}
		objects = append(objects, obj)
	}
	return objects, nil
}
