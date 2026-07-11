package app

import "testing"

func TestName(t *testing.T) {
	if Name != "ai-pixel-usage-analysis" {
		t.Fatalf("unexpected application name %q", Name)
	}
}
