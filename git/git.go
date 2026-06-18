package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type FileEntry struct {
	Status string
	Path   string
	Old    string
}

func EnsureGitRepo() error {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not inside a git repository")
	}
	return nil
}

func LoadStatus() ([]FileEntry, error) {
	out, err := exec.Command("git", "status", "--porcelain=v1", "-z").Output()
	if err != nil {
		return nil, fmt.Errorf("git status failed: %w", err)
	}
	return ParseStatus(out), nil
}

func ParseStatus(out []byte) []FileEntry {
	if len(out) == 0 {
		return nil
	}

	parts := bytes.Split(out, []byte{0})
	files := make([]FileEntry, 0, len(parts))
	for i := 0; i < len(parts); i++ {
		part := string(parts[i])
		if len(part) < 4 {
			continue
		}

		status := part[:2]
		path := part[3:]
		entry := FileEntry{Status: status, Path: path}
		if strings.ContainsAny(status, "RC") && i+1 < len(parts) && len(parts[i+1]) > 0 {
			entry.Old = string(parts[i+1])
			i++
		}
		files = append(files, entry)
	}
	return files
}

func LoadDiff(path string) (string, error) {
	out, err := exec.Command("git", "diff", "--no-color", "--unified=1000000", "HEAD", "--", path).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %s", strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func LoadOldFile(entry FileEntry) string {
	path := entry.Path
	if entry.Old != "" {
		path = entry.Old
	}
	out, err := exec.Command("git", "show", "HEAD:"+NormalizePath(path)).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return "Unable to load old file from HEAD: " + msg
	}
	return string(out)
}

func LoadNewFile(entry FileEntry) string {
	out, err := os.ReadFile(entry.Path)
	if err != nil {
		return "Unable to load file from disk: " + err.Error()
	}
	return string(out)
}

func NormalizePath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
