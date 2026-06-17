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

func TestDiffHunkOffsets(t *testing.T) {
	diff := `diff --git a/a.txt b/a.txt
index 111..222 100644
--- a/a.txt
+++ b/a.txt
@@ -1,1 +1,1 @@
 same
-old
+new
@@ -5,1 +5,1 @@
 tail`

	got := diffHunkOffsets(diff)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0] != 6 {
		t.Fatalf("offsets = %#v, want [6]", got)
	}
}

func TestDiffHunkNavigation(t *testing.T) {
	diff := `diff --git a/a.txt b/a.txt
index 111..222 100644
--- a/a.txt
+++ b/a.txt
@@ -1,1 +1,1 @@
 same
-old
+new
@@ -5,1 +5,1 @@
 tail`

	a := &app{mode: diffMode, side: fullDiff, diff: diff}
	a.handleKey("m")
	if a.diffScroll != 6 {
		t.Fatalf("diffScroll after first m = %d, want 6", a.diffScroll)
	}
	a.handleKey("m")
	if a.diffScroll != 6 {
		t.Fatalf("diffScroll after second m = %d, want 6", a.diffScroll)
	}
	a.handleKey(".")
	if a.diffScroll != 6 {
		t.Fatalf("diffScroll after . = %d, want 6", a.diffScroll)
	}
}

func TestDiffViewScrollKeys(t *testing.T) {
	a := &app{mode: diffMode}

	a.handleKey("j")
	a.handleKey("d")
	if a.diffScroll != 0 {
		t.Fatalf("diffScroll after j/d = %d, want 0", a.diffScroll)
	}

	a.handleKey("k")
	a.handleKey("s")
	if a.diffScroll != 0 {
		t.Fatalf("diffScroll after k/s = %d, want 0", a.diffScroll)
	}

	a.handleKey("s")
	if a.diffScroll != 0 {
		t.Fatalf("diffScroll after s at top = %d, want 0", a.diffScroll)
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

func TestScrollPersistenceAcrossViews(t *testing.T) {
	diff := `diff --git a/a.txt b/a.txt
index 111..222 100644
--- a/a.txt
+++ b/a.txt
@@ -1,3 +1,3 @@
 same
-old
+new
 tail`

	a := &app{mode: diffMode, side: fullDiff, diff: diff}
	a.diffLines = numberedDiffLines(diff)

	// find a diff index that has an oldNo (old file line) of 2
	idxOld2 := -1
	for i, dl := range a.diffLines {
		if dl.oldNo == 2 {
			idxOld2 = i
			break
		}
	}
	if idxOld2 == -1 {
		t.Fatalf("setup failed: couldn't find diff line with oldNo==2")
	}

	// Start in fullDiff with scroll at that diff line and switch to old view
	a.diffScroll = idxOld2
	a.toggleSide(oldSide)
	if a.side != oldSide {
		t.Fatalf("expected side oldSide, got %v", a.side)
	}

	// oldScroll should point to oldNo-1 (zero-based)
	if a.oldScroll != 1 {
		t.Fatalf("oldScroll = %d, want 1", a.oldScroll)
	}

	// simulate scrolling in old view to the 3rd line (index 2)
	a.oldScroll = 2

	// switch to new view; mapping should follow the old view's current line
	a.toggleSide(newSide)
	if a.side != newSide {
		t.Fatalf("expected side newSide, got %v", a.side)
	}

	// newScroll should be mapped to a line near the same content (old line 3 -> new line 3 -> index 2)
	if a.newScroll != 2 {
		t.Fatalf("newScroll = %d, want 2", a.newScroll)
	}

	// now move new view to its first line and switch back to diff
	a.newScroll = 0
	a.toggleSide(fullDiff)
	if a.side != fullDiff {
		t.Fatalf("expected side fullDiff, got %v", a.side)
	}

	// diffScroll should be set to the diff line corresponding to newNo==1
	found := -1
	for i, dl := range a.diffLines {
		if dl.newNo == 1 {
			found = i
			break
		}
	}
	if found == -1 {
		t.Fatalf("setup failed: couldn't find diff line with newNo==1")
	}
	if a.diffScroll != found {
		t.Fatalf("diffScroll = %d, want %d", a.diffScroll, found)
	}
}
