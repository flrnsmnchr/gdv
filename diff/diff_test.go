package diff

import "testing"

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

	old := OldSource(diff)
	if old != "same\nold\ntail" {
		t.Fatalf("OldSource = %q", old)
	}

	new := NewSource(diff)
	if new != "same\nnew\ntail" {
		t.Fatalf("NewSource = %q", new)
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

	got := DiffHunkOffsets(diff)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0] != 6 {
		t.Fatalf("offsets = %#v, want [6]", got)
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

	got := NumberedDiffLines(diff)
	if len(got) != 9 {
		t.Fatalf("len = %d, want 9", len(got))
	}
	if got[4].Text != "@@ -1,3 +1,3 @@" {
		t.Fatalf("hunk header = %#v", got[4])
	}
	if got[5].Gutter != "   1 |    1" || got[5].Text != " same" {
		t.Fatalf("context line = %#v", got[5])
	}
	if got[6].Gutter != "   2 |     " || got[6].Text != "-old" {
		t.Fatalf("deleted line = %#v", got[6])
	}
	if got[7].Gutter != "     |    2" || got[7].Text != "+new" {
		t.Fatalf("added line = %#v", got[7])
	}
	if got[8].Gutter != "   3 |    3" || got[8].Text != " tail" {
		t.Fatalf("tail line = %#v", got[8])
	}
}
