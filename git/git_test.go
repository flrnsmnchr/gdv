package git

import "testing"

func TestParseStatus(t *testing.T) {
	input := []byte(" M main.go\x00A  new.go\x00R  renamed.go\x00old.go\x00")
	got := ParseStatus(input)

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

func TestGitPathNormalizesWindowsSeparators(t *testing.T) {
	got := NormalizePath(`dir\file.txt`)
	if got != "dir/file.txt" {
		t.Fatalf("NormalizePath = %q, want %q", got, "dir/file.txt")
	}
}
