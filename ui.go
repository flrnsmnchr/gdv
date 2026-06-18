package main

import (
	"fmt"
	"strings"

	"gdv/state"

	"github.com/jroimartin/gocui"
)

func layout(a *state.App, g *gocui.Gui) error {
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

func layoutSmall(a *state.App, g *gocui.Gui, maxX, maxY int) error {
	view, err := g.SetView("main", 0, 0, maxX-1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}
	view.Title = " gdv "
	view.Clear()
	fmt.Fprintln(view, "Terminal too small")
	return nil
}

func drawTitleView(a *state.App, g *gocui.Gui, maxX int) error {
	view, err := g.SetView("title", 0, 0, maxX-1, 2)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	view.Title = " gdv "
	view.Clear()
	if a.Mode == state.StatusMode {
		fmt.Fprintf(view, "Changed files")
		if len(a.Files) > 0 {
			fmt.Fprintf(view, " (%d)", len(a.Files))
		}
		return nil
	}

	if a.Selected >= 0 && a.Selected < len(a.Files) {
		fmt.Fprint(view, a.Files[a.Selected].Path)
	}
	return nil
}

func drawMainView(a *state.App, g *gocui.Gui, maxX, maxY int) error {
	view, err := g.SetView("main", 0, 3, maxX-1, maxY-4)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	view.Title = " files "
	if a.Mode == state.DiffMode {
		view.Title = " diff "
	}
	view.Wrap = false
	view.Clear()

	if a.Mode == state.StatusMode {
		writeStatusContent(a, view, maxY-8)
		return nil
	}
	writeDiffContent(a, view)
	return nil
}

func drawStatusView(a *state.App, g *gocui.Gui, maxX, maxY int) error {
	view, err := g.SetView("status", 0, maxY-3, maxX-1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	view.Title = " keys "
	view.Clear()
	if a.Mode == state.StatusMode {
		fmt.Fprint(view, "j/f/down move down  k/d/up move up  enter/space open diff  q/esc quit")
		return nil
	}
	fmt.Fprint(view, "h old  l/g new  m next hunk  . prev hunk  j/d/down scroll down  k/s/up scroll up  enter/space back  q/esc quit")
	return nil
}

func writeStatusContent(a *state.App, view *gocui.View, maxRows int) {
	if len(a.Files) == 0 {
		fmt.Fprintln(view, "No changed files.")
		return
	}

	start := 0
	if maxRows > 0 && a.Selected >= maxRows {
		start = a.Selected - maxRows + 1
	}
	for i := start; i < len(a.Files); i++ {
		if maxRows > 0 && i >= start+maxRows {
			break
		}
		file := a.Files[i]
		prefix := "  "
		if i == a.Selected {
			prefix = "> "
		}
		line := fmt.Sprintf("%s%s %s", prefix, file.Status, file.Path)
		if file.Old != "" {
			line = fmt.Sprintf("%s%s %s -> %s", prefix, file.Status, file.Old, file.Path)
		}
		fmt.Fprintln(view, line)
	}
}

func writeDiffContent(a *state.App, view *gocui.View) {
	if a.StatusMsg != "" {
		fmt.Fprintln(view, a.StatusMsg)
		return
	}

	if a.Side == state.FullDiff {
		if a.DiffScroll > len(a.DiffLines)-1 {
			a.DiffScroll = max(0, len(a.DiffLines)-1)
		}

		lines := a.DiffLines
		if a.DiffScroll > 0 {
			lines = lines[a.DiffScroll:]
		}

		for _, line := range lines {
			text := colorDiffLine(renderableLine(line.Text))
			fmt.Fprintln(view, line.Gutter+" "+text)
		}
		return
	}

	lines := a.DiffDisplayLines()
	scroll := a.CurrentScroll()
	if scroll > len(lines)-1 {
		scroll = max(0, len(lines)-1)
		a.SetCurrentScroll(scroll)
	}
	if scroll > 0 {
		lines = lines[scroll:]
	}
	for _, line := range lines {
		fmt.Fprintln(view, line.Gutter+" "+renderableLine(line.Text))
	}
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

func blankDiffGutter(width int) string {
	return fmt.Sprintf("%*s | %*s", width, "", width, "")
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
