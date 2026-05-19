package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/oam-dev/kubevela/pkg/ai/domain"
)

func main() {
	file := flag.String("f", "", "AI domain YAML file to translate")
	flag.Parse()

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
