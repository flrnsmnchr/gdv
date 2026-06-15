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

func TestGitPathNormalizesWindowsSeparators(t *testing.T) {
	got := gitPath(`dir\file.txt`)
	if got != "dir/file.txt" {
		t.Fatalf("gitPath = %q, want %q", got, "dir/file.txt")
	}
}

func TestNumberedFileLines(t *testing.T) {
	got := numberedFileLines("alpha\nbeta", 1)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].gutter != "   1" || got[0].text != "alpha" {
		t.Fatalf("first line = %#v", got[0])
	}
	if got[1].gutter != "   2" || got[1].text != "beta" {
		t.Fatalf("second line = %#v", got[1])
	}
}

func TestNumberedDiffLines(t *testing.T) {
	diff := `diff --git a/a.txt b/a.txt
index 111..222 100644
--- a/a.txt
+++ b/a.txt
@@ -1,3 +1,3 @@
 same
-old
+new
 tail`

	got := numberedDiffLines(diff)
	if len(got) != 9 {
		t.Fatalf("len = %d, want 9", len(got))
	}
	if got[4].text != "@@ -1,3 +1,3 @@" {
		t.Fatalf("hunk header = %#v", got[4])
	}
	if got[5].gutter != "   1 |    1" || got[5].text != " same" {
		t.Fatalf("context line = %#v", got[5])
	}
	if got[6].gutter != "   2 |     " || got[6].text != "-old" {
		t.Fatalf("deleted line = %#v", got[6])
	}
	if got[7].gutter != "     |    2" || got[7].text != "+new" {
		t.Fatalf("added line = %#v", got[7])
	}
	if got[8].gutter != "   3 |    3" || got[8].text != " tail" {
		t.Fatalf("tail line = %#v", got[8])
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
