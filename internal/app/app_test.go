package app

import (
	"context"
	"path/filepath"
	"testing"
)

func TestName(t *testing.T) {
	if Name != "ai-pixel-usage-analysis" {
		t.Fatalf("unexpected application name %q", Name)
	}
}

func TestRunRejectsMissingConfig(t *testing.T) {
	err := Run(context.Background(), Options{ConfigPath: filepath.Join(t.TempDir(), "missing.toml")})
	if err == nil {
		t.Fatal("expected missing config error")
	}
}
