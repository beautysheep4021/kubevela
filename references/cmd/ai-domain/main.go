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
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "translate":
			runTranslate(os.Args[2:])
			return
		case "validate":
			runValidate(os.Args[2:])
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

func runValidate(args []string) {
	flags := flag.NewFlagSet("validate", flag.ExitOnError)
	file := flags.String("f", "", "AI domain YAML file to validate")
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
	result, err := domain.ValidateYAML(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "validate %s: %v\n", *file, err)
		os.Exit(1)
	}
	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode validation output: %v\n", err)
		os.Exit(1)
	}
	_, _ = os.Stdout.Write(append(out, '\n'))
	if len(result.Errors) > 0 {
		os.Exit(1)
	}
}

func runStatus(args []string) {
	parsed, err := parseStatusArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	var (
		summary *observe.Summary
	)
	if parsed.fromFiles != "" {
		var objects []*unstructured.Unstructured
		objects, err = readObjectsFromFiles(parsed.fromFiles)
		if err == nil {
			summary, err = observe.SummarizeObjects(objects)
		}
	} else {
		summary, err = observe.SummarizeApplication(context.Background(), parsed.kubeconfig, parsed.namespace, parsed.appName)
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
	parsed, err := parseApplyArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	in, err := os.ReadFile(parsed.file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", parsed.file, err)
		os.Exit(1)
	}
	appYAML, err := domain.TranslateYAML(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "translate %s: %v\n", parsed.file, err)
		os.Exit(1)
	}
	if parsed.namespace != "" {
		appYAML, err = setApplicationNamespace(appYAML, parsed.namespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "set namespace: %v\n", err)
			os.Exit(1)
		}
	}
	config, err := loadRestConfig(parsed.kubeconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create dynamic client: %v\n", err)
		os.Exit(1)
	}
	result, err := domainapply.ApplyYAML(context.Background(), client, appYAML, domainapply.Options{DryRun: parsed.dryRun})
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

type statusArgs struct {
	fromFiles  string
	namespace  string
	kubeconfig string
	appName    string
}

func parseStatusArgs(args []string) (statusArgs, error) {
	parsed := statusArgs{namespace: "default"}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--from-files":
			value, next, err := nextFlagValue(args, i, arg)
			if err != nil {
				return statusArgs{}, err
			}
			parsed.fromFiles = value
			i = next
		case strings.HasPrefix(arg, "--from-files="):
			parsed.fromFiles = strings.TrimPrefix(arg, "--from-files=")
		case arg == "--kubeconfig":
			value, next, err := nextFlagValue(args, i, arg)
			if err != nil {
				return statusArgs{}, err
			}
			parsed.kubeconfig = value
			i = next
		case strings.HasPrefix(arg, "--kubeconfig="):
			parsed.kubeconfig = strings.TrimPrefix(arg, "--kubeconfig=")
		case arg == "-n":
			value, next, err := nextFlagValue(args, i, arg)
			if err != nil {
				return statusArgs{}, err
			}
			parsed.namespace = value
			i = next
		case strings.HasPrefix(arg, "-n="):
			parsed.namespace = strings.TrimPrefix(arg, "-n=")
		case strings.HasPrefix(arg, "-"):
			return statusArgs{}, fmt.Errorf("unknown status flag %s", arg)
		default:
			if parsed.appName != "" {
				return statusArgs{}, fmt.Errorf("status accepts only one application name")
			}
			parsed.appName = arg
		}
	}
	if parsed.fromFiles != "" {
		if parsed.appName != "" {
			return statusArgs{}, fmt.Errorf("status --from-files does not accept an application name")
		}
		return parsed, nil
	}
	if parsed.appName == "" {
		return statusArgs{}, fmt.Errorf("status requires an application name or --from-files")
	}
	return parsed, nil
}

type applyArgs struct {
	file       string
	namespace  string
	kubeconfig string
	dryRun     bool
	yes        bool
}

func parseApplyArgs(args []string) (applyArgs, error) {
	flags := flag.NewFlagSet("apply", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	file := flags.String("f", "", "AI domain YAML file to translate and apply")
	namespace := flags.String("n", "", "Application namespace override")
	kubeconfig := flags.String("kubeconfig", "", "kubeconfig path")
	dryRun := flags.Bool("dry-run", false, "perform server-side dry-run only")
	yes := flags.Bool("yes", false, "actually apply the generated Application")
	if err := flags.Parse(args); err != nil {
		return applyArgs{}, err
	}
	if *file == "" {
		return applyArgs{}, fmt.Errorf("-f is required")
	}
	if !*dryRun && !*yes {
		return applyArgs{}, fmt.Errorf("apply requires --dry-run or --yes")
	}
	return applyArgs{file: *file, namespace: *namespace, kubeconfig: *kubeconfig, dryRun: *dryRun, yes: *yes}, nil
}

func nextFlagValue(args []string, index int, flagName string) (string, int, error) {
	next := index + 1
	if next >= len(args) || args[next] == "" || strings.HasPrefix(args[next], "-") {
		return "", index, fmt.Errorf("%s requires a value", flagName)
	}
	return args[next], next, nil
}

func setApplicationNamespace(content []byte, namespace string) ([]byte, error) {
	obj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(content, obj); err != nil {
		return nil, fmt.Errorf("decode Application YAML: %w", err)
	}
	obj.SetNamespace(namespace)
	out, err := yaml.Marshal(obj.Object)
	if err != nil {
		return nil, fmt.Errorf("encode Application YAML: %w", err)
	}
	return out, nil
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
