package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseOptionsRejectsConflictingActions(t *testing.T) {
	_, err := parseOptions([]string{"--enable", "--disable"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "cannot be used together") {
		t.Fatalf("expected conflicting-action error, got %v", err)
	}
}

func TestRunRequiresExplicitActivation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("enabled: true\nverbose: false\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if err := run([]string{"--config", path}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "explicit --enable is required") {
		t.Fatalf("expected explicit-activation state, got %q", stdout.String())
	}
}

func TestRunRejectsInvalidConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("enabled: maybe\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := run([]string{"--config", path}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected invalid config error")
	}
}

func TestParseOptionsHelp(t *testing.T) {
	_, err := parseOptions([]string{"--help"}, &bytes.Buffer{})
	if err != flag.ErrHelp {
		t.Fatalf("expected flag.ErrHelp, got %v", err)
	}
}
