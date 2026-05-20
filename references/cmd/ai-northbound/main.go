package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/oam-dev/kubevela/pkg/ai/northbound"
)

type args struct {
	addr string
}

func main() {
	parsed, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	server := &http.Server{
		Addr:    parsed.addr,
		Handler: northbound.NewServer(),
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
	if err := flags.Parse(raw); err != nil {
		return args{}, err
	}
	return args{addr: *addr}, nil
}
