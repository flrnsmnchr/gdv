package main

import "testing"

func TestParseStatus(t *testing.T) {
	input := []byte(" M main.go\x00A  new.go\x00R  renamed.go\x00old.go\x00")
	got := parseStatus(input)

	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].Status != " M" || got[0].Path != "main.go" {
		t.Fatalf("first entry = %#v", got[0])
	}
	if got[1].Status != "A " || got[1].Path != "new.go" {
		t.Fatalf("second entry = %#v", got[1])
	}
	if got[2].Status != "R " || got[2].Path != "renamed.go" || got[2].Old != "old.go" {
		t.Fatalf("rename entry = %#v", got[2])
	}
}

func TestSideSources(t *testing.T) {
	diff := `diff --git a/a.txt b/a.txt
index 111..222 100644
--- a/a.txt
+++ b/a.txt
@@ -1,3 +1,3 @@
 same
-old
+new
 tail`

	old := oldSource(diff)
	if old != "same\nold\ntail" {
		t.Fatalf("oldSource = %q", old)
	}

	new := newSource(diff)
	if new != "same\nnew\ntail" {
		t.Fatalf("newSource = %q", new)
	}
}

func TestSideViewKeysToggleBackToFullDiff(t *testing.T) {
	a := &app{mode: diffMode}

	a.handleKey("h")
	if a.side != oldSide {
		t.Fatalf("side after h = %v, want oldSide", a.side)
	}
	a.handleKey("h")
	if a.side != fullDiff {
		t.Fatalf("side after second h = %v, want fullDiff", a.side)
	}

	a.handleKey("l")
	if a.side != newSide {
		t.Fatalf("side after l = %v, want newSide", a.side)
	}
	a.handleKey("l")
	if a.side != fullDiff {
		t.Fatalf("side after second l = %v, want fullDiff", a.side)
	}
}

func TestDiffViewScrollKeys(t *testing.T) {
	a := &app{mode: diffMode}

	a.handleKey("j")
	a.handleKey("d")
	if a.scroll != 2 {
		t.Fatalf("scroll after j/d = %d, want 2", a.scroll)
	}

	a.handleKey("k")
	a.handleKey("s")
	if a.scroll != 0 {
		t.Fatalf("scroll after k/s = %d, want 0", a.scroll)
	}

	a.handleKey("s")
	if a.scroll != 0 {
		t.Fatalf("scroll after s at top = %d, want 0", a.scroll)
	}
}

func TestRenderableLineExpandsTabs(t *testing.T) {
	got := renderableLine("+\tvalue")
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
		if got := colorDiffLine(input); got != want {
			t.Fatalf("colorDiffLine(%q) = %q, want %q", input, got, want)
		}
	}
}
