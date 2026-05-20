package main

import "testing"

func TestParseArgsDefaultsToLocalhost8088(t *testing.T) {
	args, err := parseArgs(nil)
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}
	if args.addr != "127.0.0.1:8088" {
		t.Fatalf("addr = %q, want 127.0.0.1:8088", args.addr)
	}
}

func TestParseArgsAcceptsAddrOverride(t *testing.T) {
	args, err := parseArgs([]string{"--addr", "0.0.0.0:18088"})
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}
	if args.addr != "0.0.0.0:18088" {
		t.Fatalf("addr = %q, want 0.0.0.0:18088", args.addr)
	}
}

func TestParseArgsAcceptsKubeconfigOverride(t *testing.T) {
	args, err := parseArgs([]string{"--addr", "0.0.0.0:18088", "--kubeconfig", "/root/.kube/config"})
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}
	if args.addr != "0.0.0.0:18088" {
		t.Fatalf("addr = %q, want 0.0.0.0:18088", args.addr)
	}
	if args.kubeconfig != "/root/.kube/config" {
		t.Fatalf("kubeconfig = %q, want /root/.kube/config", args.kubeconfig)
	}
}
