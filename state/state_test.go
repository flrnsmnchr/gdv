package state

import (
	difflib "gdv/diff"
	"testing"
)

func TestSideViewKeysToggleBackToFullDiff(t *testing.T) {
	a := &App{Mode: DiffMode}

	a.HandleKey("h")
	if a.Side != OldSide {
		t.Fatalf("side after h = %v, want OldSide", a.Side)
	}
	a.HandleKey("h")
	if a.Side != FullDiff {
		t.Fatalf("side after second h = %v, want FullDiff", a.Side)
	}

	a.HandleKey("l")
	if a.Side != NewSide {
		t.Fatalf("side after l = %v, want NewSide", a.Side)
	}
	a.HandleKey("l")
	if a.Side != FullDiff {
		t.Fatalf("side after second l = %v, want FullDiff", a.Side)
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

	a := &App{Mode: DiffMode, Side: FullDiff, Diff: diff}
	a.DiffLines = difflib.NumberedDiffLines(diff)
	a.HandleKey("m")
	if a.DiffScroll != 6 {
		t.Fatalf("diffScroll after first m = %d, want 6", a.DiffScroll)
	}
	a.HandleKey("m")
	if a.DiffScroll != 6 {
		t.Fatalf("diffScroll after second m = %d, want 6", a.DiffScroll)
	}
	a.HandleKey(".")
	if a.DiffScroll != 6 {
		t.Fatalf("diffScroll after . = %d, want 6", a.DiffScroll)
	}
}

func TestDiffViewScrollKeys(t *testing.T) {
	a := &App{Mode: DiffMode}

	a.DiffLines = difflib.NumberedDiffLines("diff --git a/a.txt b/a.txt\nindex 111..222 100644\n--- a/a.txt\n+++ b/a.txt\n@@ -1,1 +1,1 @@\n same\n-old\n+new")
	a.HandleKey("j")
	a.HandleKey("d")
	if a.DiffScroll != 0 {
		t.Fatalf("diffScroll after j/d = %d, want 0", a.DiffScroll)
	}

	a.HandleKey("k")
	a.HandleKey("d")
	if a.DiffScroll != 0 {
		t.Fatalf("diffScroll after k/d = %d, want 0", a.DiffScroll)
	}

	a.HandleKey("k")
	if a.DiffScroll != 0 {
		t.Fatalf("diffScroll after k at top = %d, want 0", a.DiffScroll)
	}
}

func TestNumberedFileLines(t *testing.T) {
	got := NumberedFileLines("alpha\nbeta", 1)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Gutter != "   1" || got[0].Text != "alpha" {
		t.Fatalf("first line = %#v", got[0])
	}
	if got[1].Gutter != "   2" || got[1].Text != "beta" {
		t.Fatalf("second line = %#v", got[1])
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

	a := &App{Mode: DiffMode, Side: FullDiff, Diff: diff}
	a.DiffLines = difflib.NumberedDiffLines(diff)

	idxOld2 := -1
	for i, dl := range a.DiffLines {
		if dl.OldNo == 2 {
			idxOld2 = i
			break
		}
	}
	if idxOld2 == -1 {
		t.Fatalf("setup failed: couldn't find diff line with oldNo==2")
	}

	a.DiffScroll = idxOld2
	a.ToggleSide(OldSide)
	if a.Side != OldSide {
		t.Fatalf("expected side OldSide, got %v", a.Side)
	}

	if a.OldScroll != 1 {
		t.Fatalf("oldScroll = %d, want 1", a.OldScroll)
	}

	a.OldScroll = 2
	a.ToggleSide(NewSide)
	if a.Side != NewSide {
		t.Fatalf("expected side NewSide, got %v", a.Side)
	}

	if a.NewScroll != 2 {
		t.Fatalf("newScroll = %d, want 2", a.NewScroll)
	}

	a.NewScroll = 0
	a.ToggleSide(FullDiff)
	if a.Side != FullDiff {
		t.Fatalf("expected side FullDiff, got %v", a.Side)
	}

	found := -1
	for i, dl := range a.DiffLines {
		if dl.NewNo == 1 {
			found = i
			break
		}
	}
	if found == -1 {
		t.Fatalf("setup failed: couldn't find diff line with newNo==1")
	}
	if a.DiffScroll != found {
		t.Fatalf("diffScroll = %d, want %d", a.DiffScroll, found)
	}
}
