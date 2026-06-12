package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type mode int

const (
	statusMode mode = iota
	diffMode
)

type sideMode int

const (
	fullDiff sideMode = iota
	oldSide
	newSide
)

type fileEntry struct {
	Status string
	Path   string
	Old    string
}

type app struct {
	files      []fileEntry
	selected   int
	mode       mode
	side       sideMode
	context    int
	diff       string
	scroll     int
	statusMsg  string
	screenRows int
	screenCols int
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gdv:", err)
		os.Exit(1)
	}
}

func run() error {
	if err := ensureGitRepo(); err != nil {
		return err
	}

	files, err := loadStatus()
	if err != nil {
		return err
	}

	state := &app{
		files:   files,
		context: 3,
	}

	restore, err := enableRawMode()
	if err != nil {
		return err
	}
	defer restore()

	hideCursor()
	defer showCursor()
	defer clearScreen()

	for {
		state.screenRows, state.screenCols = terminalSize()
		state.draw()

		key, err := readKey(os.Stdin)
		if err != nil {
			return err
		}
		if state.handleKey(key) {
			return nil
		}
	}
}

func ensureGitRepo() error {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	if err := cmd.Run(); err != nil {
		return errors.New("not inside a git repository")
	}
	return nil
}

func loadStatus() ([]fileEntry, error) {
	out, err := exec.Command("git", "status", "--porcelain=v1", "-z").Output()
	if err != nil {
		return nil, fmt.Errorf("git status failed: %w", err)
	}
	return parseStatus(out), nil
}

func parseStatus(out []byte) []fileEntry {
	if len(out) == 0 {
		return nil
	}

	parts := bytes.Split(out, []byte{0})
	files := make([]fileEntry, 0, len(parts))
	for i := 0; i < len(parts); i++ {
		part := string(parts[i])
		if len(part) < 4 {
			continue
		}

		status := part[:2]
		path := part[3:]
		entry := fileEntry{Status: status, Path: path}
		if strings.ContainsAny(status, "RC") && i+1 < len(parts) && len(parts[i+1]) > 0 {
			entry.Old = string(parts[i+1])
			i++
		}
		files = append(files, entry)
	}
	return files
}

func (a *app) handleKey(key string) bool {
	if key == "q" || key == "esc" || key == "ctrl-c" {
		return true
	}

	if a.mode == statusMode {
		switch key {
		case "j", "f", "down":
			if a.selected < len(a.files)-1 {
				a.selected++
			}
		case "k", "d", "up":
			if a.selected > 0 {
				a.selected--
			}
		case "enter", "space":
			if len(a.files) > 0 {
				a.side = fullDiff
				a.openDiff()
			}
		}
		return false
	}

	switch key {
	case "enter", "space":
		a.mode = statusMode
		a.statusMsg = ""
	case "h", "s":
		a.toggleSide(oldSide)
		a.scroll = 0
	case "l", "g":
		a.toggleSide(newSide)
		a.scroll = 0
	case "j", "f":
		a.context++
		a.openDiff()
	case "k", "d":
		if a.context > 0 {
			a.context--
			a.openDiff()
		}
	case "down":
		a.scroll++
	case "up":
		if a.scroll > 0 {
			a.scroll--
		}
	}
	return false
}

func (a *app) toggleSide(side sideMode) {
	if a.side == side {
		a.side = fullDiff
		return
	}
	a.side = side
}

func (a *app) openDiff() {
	a.mode = diffMode
	a.scroll = 0
	a.statusMsg = ""

	if a.selected < 0 || a.selected >= len(a.files) {
		a.diff = ""
		return
	}

	entry := a.files[a.selected]
	diff, err := loadDiff(entry.Path, a.context)
	if err != nil {
		a.diff = ""
		a.statusMsg = err.Error()
		return
	}
	if strings.TrimSpace(diff) == "" {
		a.diff = "(no unstaged or committed diff for this file)"
		return
	}
	a.diff = diff
}

func loadDiff(path string, context int) (string, error) {
	arg := "-U" + strconv.Itoa(context)
	out, err := exec.Command("git", "diff", arg, "HEAD", "--", path).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %s", strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func (a *app) draw() {
	clearScreen()
	if a.mode == statusMode {
		a.drawStatus()
		return
	}
	a.drawDiff()
}

func (a *app) drawStatus() {
	fmt.Print(invert(" gdv "))
	fmt.Print(" changed files")
	if len(a.files) > 0 {
		fmt.Printf(" (%d)", len(a.files))
	}
	fmt.Print("\r\n")

	if len(a.files) == 0 {
		fmt.Print("\r\nNo changed files.\r\n")
		fmt.Print("\r\nq/Esc quit\r\n")
		return
	}

	maxRows := a.screenRows - 3
	start := 0
	if maxRows > 0 && a.selected >= maxRows {
		start = a.selected - maxRows + 1
	}
	for i := start; i < len(a.files); i++ {
		if i >= start+maxRows {
			break
		}
		file := a.files[i]
		line := fmt.Sprintf("%s %s", file.Status, file.Path)
		if file.Old != "" {
			line = fmt.Sprintf("%s %s -> %s", file.Status, file.Old, file.Path)
		}
		line = fitLine(line, a.screenCols)
		if i == a.selected {
			fmt.Print(invert(line))
		} else {
			fmt.Print(line)
		}
		fmt.Print("\r\n")
	}
	fmt.Print("\r\nj/f down  k/d up  enter/space diff  q quit\r\n")
}

func (a *app) drawDiff() {
	file := a.files[a.selected]
	title := fmt.Sprintf(" gdv %s context=%d ", file.Path, a.context)
	fmt.Print(invert(fitLine(title, a.screenCols)))
	fmt.Print("\r\n")

	if a.statusMsg != "" {
		fmt.Print(a.statusMsg + "\r\n")
		return
	}

	content := a.diff
	switch a.side {
	case oldSide:
		content = oldSource(a.diff)
	case newSide:
		content = newSource(a.diff)
	}

	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	maxRows := a.screenRows - 3
	if a.scroll > len(lines)-1 {
		a.scroll = max(0, len(lines)-1)
	}
	for i := a.scroll; i < len(lines) && i < a.scroll+maxRows; i++ {
		fmt.Print(colorDiffLine(fitLine(lines[i], a.screenCols)))
		fmt.Print("\r\n")
	}
	fmt.Print("\r\nh/s old  l/g new  j/f more ctx  k/d less ctx  enter/space back  q quit\r\n")
}

func oldSource(diff string) string {
	return sideSource(diff, '-')
}

func newSource(diff string) string {
	return sideSource(diff, '+')
}

func sideSource(diff string, marker byte) string {
	var out []string
	for _, line := range strings.Split(strings.ReplaceAll(diff, "\r\n", "\n"), "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "diff --git") || strings.HasPrefix(line, "index ") {
			continue
		}
		if strings.HasPrefix(line, "@@") {
			out = append(out, line)
			continue
		}
		if line[0] == ' ' {
			out = append(out, line[1:])
			continue
		}
		if line[0] == marker {
			out = append(out, line[1:])
		}
	}
	if len(out) == 0 {
		return "(empty)"
	}
	return strings.Join(out, "\n")
}

func colorDiffLine(line string) string {
	if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
		return "\x1b[32m" + line + "\x1b[0m"
	}
	if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
		return "\x1b[31m" + line + "\x1b[0m"
	}
	if strings.HasPrefix(line, "@@") {
		return "\x1b[36m" + line + "\x1b[0m"
	}
	return line
}

func fitLine(line string, width int) string {
	if width <= 0 {
		return line
	}
	line = strings.ReplaceAll(line, "\t", "    ")
	if len(line) <= width {
		return line
	}
	if width <= 1 {
		return line[:width]
	}
	return line[:width-1] + "~"
}

func clearScreen() {
	fmt.Print("\x1b[2J\x1b[H")
}

func hideCursor() {
	fmt.Print("\x1b[?25l")
}

func showCursor() {
	fmt.Print("\x1b[?25h")
}

func invert(s string) string {
	return "\x1b[7m" + s + "\x1b[0m"
}
