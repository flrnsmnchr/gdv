package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jroimartin/gocui"
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
	files     []fileEntry
	selected  int
	mode      mode
	side      sideMode
	diff      string
	scroll    int
	statusMsg string
	gui       *gocui.Gui
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

	state := &app{files: files}

	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	defer gui.Close()

	state.gui = gui
	gui.SetManagerFunc(state.layout)

	if err := state.setKeybindings(); err != nil {
		return err
	}
	if err := gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
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

func (a *app) setKeybindings() error {
	bindings := []struct {
		key     interface{}
		handler func(*gocui.Gui, *gocui.View) error
	}{
		{gocui.KeyCtrlC, a.quit},
		{gocui.KeyEsc, a.quit},
		{'q', a.key("q")},
		{'j', a.key("j")},
		{'f', a.key("f")},
		{'k', a.key("k")},
		{'d', a.key("d")},
		{'h', a.key("h")},
		{'s', a.key("s")},
		{'l', a.key("l")},
		{'g', a.key("g")},
		{gocui.KeyArrowDown, a.key("down")},
		{gocui.KeyArrowUp, a.key("up")},
		{gocui.KeyEnter, a.key("enter")},
		{gocui.KeySpace, a.key("space")},
	}

	for _, binding := range bindings {
		if err := a.gui.SetKeybinding("", binding.key, gocui.ModNone, binding.handler); err != nil {
			return err
		}
	}
	return nil
}

func (a *app) key(key string) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if a.handleKey(key) {
			return gocui.ErrQuit
		}
		return nil
	}
}

func (a *app) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
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
	case "h":
		a.toggleSide(oldSide)
		a.scroll = 0
	case "l", "g":
		a.toggleSide(newSide)
		a.scroll = 0
	case "j", "d", "down":
		a.scroll++
	case "k", "s", "up":
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
	diff, err := loadDiff(entry.Path)
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

func loadDiff(path string) (string, error) {
	out, err := exec.Command("git", "diff", "--no-color", "--unified=1000000", "HEAD", "--", path).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %s", strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func (a *app) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if maxX < 20 || maxY < 8 {
		return a.layoutSmall(g, maxX, maxY)
	}

	if err := a.drawTitleView(g, maxX); err != nil {
		return err
	}
	if err := a.drawMainView(g, maxX, maxY); err != nil {
		return err
	}
	return a.drawStatusView(g, maxX, maxY)
}

func (a *app) layoutSmall(g *gocui.Gui, maxX, maxY int) error {
	view, err := g.SetView("main", 0, 0, maxX-1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	view.Title = " gdv "
	view.Clear()
	fmt.Fprintln(view, "Terminal too small")
	return nil
}

func (a *app) drawTitleView(g *gocui.Gui, maxX int) error {
	view, err := g.SetView("title", 0, 0, maxX-1, 2)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	view.Title = " gdv "
	view.Clear()
	if a.mode == statusMode {
		fmt.Fprintf(view, "Changed files")
		if len(a.files) > 0 {
			fmt.Fprintf(view, " (%d)", len(a.files))
		}
		return nil
	}

	if a.selected >= 0 && a.selected < len(a.files) {
		fmt.Fprint(view, a.files[a.selected].Path)
	}
	return nil
}

func (a *app) drawMainView(g *gocui.Gui, maxX, maxY int) error {
	view, err := g.SetView("main", 0, 3, maxX-1, maxY-4)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	view.Title = " files "
	if a.mode == diffMode {
		view.Title = " diff "
	}
	view.Wrap = false
	view.Clear()

	if a.mode == statusMode {
		a.writeStatusContent(view, maxY-8)
		return nil
	}
	a.writeDiffContent(view)
	return nil
}

func (a *app) drawStatusView(g *gocui.Gui, maxX, maxY int) error {
	view, err := g.SetView("status", 0, maxY-3, maxX-1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	view.Title = " keys "
	view.Clear()
	if a.mode == statusMode {
		fmt.Fprint(view, "j/f/down move down  k/d/up move up  enter/space open diff  q/esc quit")
		return nil
	}
	fmt.Fprint(view, "h old  l/g new  j/d/down scroll down  k/s/up scroll up  enter/space back  q/esc quit")
	return nil
}

func (a *app) writeStatusContent(view *gocui.View, maxRows int) {
	if len(a.files) == 0 {
		fmt.Fprintln(view, "No changed files.")
		return
	}

	start := 0
	if maxRows > 0 && a.selected >= maxRows {
		start = a.selected - maxRows + 1
	}
	for i := start; i < len(a.files); i++ {
		if maxRows > 0 && i >= start+maxRows {
			break
		}
		file := a.files[i]
		prefix := "  "
		if i == a.selected {
			prefix = "> "
		}
		line := fmt.Sprintf("%s%s %s", prefix, file.Status, file.Path)
		if file.Old != "" {
			line = fmt.Sprintf("%s%s %s -> %s", prefix, file.Status, file.Old, file.Path)
		}
		fmt.Fprintln(view, line)
	}
}

func (a *app) writeDiffContent(view *gocui.View) {
	if a.statusMsg != "" {
		fmt.Fprintln(view, a.statusMsg)
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
	if a.scroll > len(lines)-1 {
		a.scroll = max(0, len(lines)-1)
	}
	if a.scroll > 0 {
		lines = lines[a.scroll:]
	}
	for _, line := range lines {
		fmt.Fprintln(view, colorDiffLine(renderableLine(line)))
	}
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

func renderableLine(line string) string {
	return strings.ReplaceAll(line, "\t", "    ")
}

func colorDiffLine(line string) string {
	switch {
	case strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
		return "\x1b[32m" + line + "\x1b[0m"
	case strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---"):
		return "\x1b[31m" + line + "\x1b[0m"
	case strings.HasPrefix(line, "@@"):
		return "\x1b[36m" + line + "\x1b[0m"
	case strings.HasPrefix(line, "diff --git"), strings.HasPrefix(line, "index "), strings.HasPrefix(line, "---"), strings.HasPrefix(line, "+++"):
		return "\x1b[36m" + line + "\x1b[0m"
	default:
		return line
	}
}
