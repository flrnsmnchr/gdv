package ui

import "testing"

func TestRenderableLineExpandsTabs(t *testing.T) {
	got := RenderableLine("+\tvalue")
	if got != "+    value" {
		t.Fatalf("renderableLine = %q, want %q", got, "+    value")
	}
}

func TestColorDiffLine(t *testing.T) {
	tests := map[string]string{
		"+added":  "\x1b[32m+added\x1b[0m",
		"-gone":   "\x1b[31m-gone\x1b[0m",
		"@@ hunk": "\x1b[36m@@ hunk\x1b[0m",
		"+++ b/x": "\x1b[36m+++ b/x\x1b[0m",
		"same":    "same",
	}

	for input, want := range tests {
		if got := ColorDiffLine(input); got != want {
			t.Fatalf("colorDiffLine(%q) = %q, want %q", input, got, want)
		}
	}
}
