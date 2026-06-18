package main

import (
	"fmt"
	"strings"

	"gdv/diff"

	"github.com/jroimartin/gocui"
)

func layout(a *app, g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if maxX < 20 || maxY < 8 {
		return layoutSmall(a, g, maxX, maxY)
	}

	if err := drawTitleView(a, g, maxX); err != nil {
		return err
	}
	if err := drawMainView(a, g, maxX, maxY); err != nil {
		return err
	}
	return drawStatusView(a, g, maxX, maxY)
}

func layoutSmall(a *app, g *gocui.Gui, maxX, maxY int) error {
	view, err := g.SetView("main", 0, 0, maxX-1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	view.Title = " gdv "
	view.Clear()
	fmt.Fprintln(view, "Terminal too small")
	return nil
}

func drawTitleView(a *app, g *gocui.Gui, maxX int) error {
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

func drawMainView(a *app, g *gocui.Gui, maxX, maxY int) error {
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
		writeStatusContent(a, view, maxY-8)
		return nil
	}
	writeDiffContent(a, view)
	return nil
}

func drawStatusView(a *app, g *gocui.Gui, maxX, maxY int) error {
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
	fmt.Fprint(view, "h old  l/g new  m next hunk  . prev hunk  j/d/down scroll down  k/s/up scroll up  enter/space back  q/esc quit")
	return nil
}

func writeStatusContent(a *app, view *gocui.View, maxRows int) {
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

func writeDiffContent(a *app, view *gocui.View) {
	if a.statusMsg != "" {
		fmt.Fprintln(view, a.statusMsg)
		return
	}

	if a.side == fullDiff {
		if a.diffScroll > len(a.diffLines)-1 {
			a.diffScroll = max(0, len(a.diffLines)-1)
		}

		lines := a.diffLines
		if a.diffScroll > 0 {
			lines = lines[a.diffScroll:]
		}

		for _, line := range lines {
			text := colorDiffLine(renderableLine(line.Text))
			fmt.Fprintln(view, line.Gutter+" "+text)
		}
		return
	}

	lines := a.diffDisplayLines()
	scroll := a.currentScroll()
	if scroll > len(lines)-1 {
		scroll = max(0, len(lines)-1)
		a.setCurrentScroll(scroll)
	}
	if scroll > 0 {
		lines = lines[scroll:]
	}
	for _, line := range lines {
		fmt.Fprintln(view, line.gutter+" "+renderableLine(line.text))
	}
}

func diffLinesToDisplayLines(lines []diff.DiffLine) []displayLine {
	out := make([]displayLine, 0, len(lines))
	for _, line := range lines {
		out = append(out, displayLine{gutter: line.Gutter, text: line.Text})
	}
	return out
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

func numberedFileLines(content string, start int) []displayLine {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	width := numberWidth(start + len(lines) - 1)
	out := make([]displayLine, 0, len(lines))
	for i, line := range lines {
		out = append(out, displayLine{
			gutter: fmt.Sprintf("%*d", width, start+i),
			text:   line,
		})
	}
	return out
}

func numberWidth(maxNo int) int {
	if maxNo < 1 {
		return 1
	}
	width := 0
	for maxNo > 0 {
		width++
		maxNo /= 10
	}
	if width < 4 {
		return 4
	}
	return width
}

func blankDiffGutter(width int) string {
	return fmt.Sprintf("%*s | %*s", width, "", width, "")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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
