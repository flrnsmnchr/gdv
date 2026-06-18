package state

import (
	"fmt"
	"strings"

	"gdv/diff"
	"gdv/git"

	"github.com/jroimartin/gocui"
)

type Mode int

type SideMode int

type FileEntry struct {
	Status string
	Path   string
	Old    string
}

type DisplayLine struct {
	Gutter string
	Text   string
}

type App struct {
	Files        []FileEntry
	Selected     int
	Mode         Mode
	Side         SideMode
	Diff         string
	OldFile      string
	NewFile      string
	DiffLines    []diff.DiffLine
	DiffScroll   int
	OldScroll    int
	NewScroll    int
	OldScrollSet bool
	NewScrollSet bool
	StatusMsg    string
	Gui          *gocui.Gui
}

const (
	StatusMode Mode = iota
	DiffMode
)

const (
	FullDiff SideMode = iota
	OldSide
	NewSide
)

func (a *App) SetKeybindings() error {
	bindings := []struct {
		key     interface{}
		handler func(*gocui.Gui, *gocui.View) error
	}{
		{gocui.KeyCtrlC, a.Quit},
		{gocui.KeyEsc, a.Quit},
		{'q', a.Key("q")},
		{'j', a.Key("j")},
		{'f', a.Key("f")},
		{'k', a.Key("k")},
		{'d', a.Key("d")},
		{'h', a.Key("h")},
		{'s', a.Key("s")},
		{'l', a.Key("l")},
		{'g', a.Key("g")},
		{'m', a.Key("m")},
		{',', a.Key(",")},
		{'v', a.Key("v")},
		{'c', a.Key("c")},
		{gocui.KeyArrowDown, a.Key("down")},
		{gocui.KeyArrowUp, a.Key("up")},
		{gocui.KeyEnter, a.Key("enter")},
		{gocui.KeySpace, a.Key("space")},
	}

	for _, binding := range bindings {
		if err := a.Gui.SetKeybinding("", binding.key, gocui.ModNone, binding.handler); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) Key(key string) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		if a.HandleKey(key) {
			return gocui.ErrQuit
		}
		return nil
	}
}

func (a *App) Quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (a *App) HandleKey(key string) bool {
	if key == "q" || key == "esc" || key == "ctrl-c" {
		return true
	}

	if a.Mode == StatusMode {
		switch key {
		case "j", "f", "down":
			if a.Selected < len(a.Files)-1 {
				a.Selected++
			}
		case "k", "d", "up":
			if a.Selected > 0 {
				a.Selected--
			}
		case "enter", "space":
			if len(a.Files) > 0 {
				a.Side = FullDiff
				a.OpenDiff()
			}
		}
		return false
	}

	switch key {
	case "enter", "space":
		a.Mode = StatusMode
		a.StatusMsg = ""
	case "h", "s":
		a.ToggleSide(OldSide)
	case "l", "g":
		a.ToggleSide(NewSide)
	case "m", "v":
		a.NextHunk()
	case ",", "c":
		a.PreviousHunk()
	case "j", "f", "down":
		a.IncrementScroll()
	case "k", "d", "up":
		a.DecrementScroll()
	}
	return false
}

func (a *App) OpenDiff() {
	a.Mode = DiffMode
	a.DiffScroll = 0
	a.OldScroll = 0
	a.NewScroll = 0
	a.OldScrollSet = false
	a.NewScrollSet = false
	a.StatusMsg = ""

	if a.Selected < 0 || a.Selected >= len(a.Files) {
		a.Diff = ""
		return
	}

	entry := a.Files[a.Selected]
	diffData, err := git.LoadDiff(entry.Path)
	if err != nil {
		a.Diff = ""
		a.StatusMsg = err.Error()
		return
	}
	if strings.TrimSpace(diffData) == "" {
		a.Diff = "(no unstaged or committed diff for this file)"
		return
	}
	a.Diff = diffData
	a.DiffLines = diff.NumberedDiffLines(a.Diff)
	gitEntry := git.FileEntry{
		Status: entry.Status,
		Path:   entry.Path,
		Old:    entry.Old,
	}
	a.OldFile = git.LoadOldFile(gitEntry)
	a.NewFile = git.LoadNewFile(gitEntry)
}

func (a *App) CurrentScroll() int {
	switch a.Side {
	case OldSide:
		return a.OldScroll
	case NewSide:
		return a.NewScroll
	default:
		return a.DiffScroll
	}
}

func (a *App) SetCurrentScroll(scroll int) {
	switch a.Side {
	case OldSide:
		a.OldScroll = scroll
		a.OldScrollSet = true
	case NewSide:
		a.NewScroll = scroll
		a.NewScrollSet = true
	default:
		a.DiffScroll = scroll
	}
}

func (a *App) IncrementScroll() {
	switch a.Side {
	case FullDiff:
		if a.DiffScroll < len(a.DiffLines)-1 {
			a.DiffScroll++
		}
	case OldSide:
		if a.OldScroll < len(a.DiffDisplayLines())-1 {
			a.OldScroll++
			a.OldScrollSet = true
		}
	case NewSide:
		if a.NewScroll < len(a.DiffDisplayLines())-1 {
			a.NewScroll++
			a.NewScrollSet = true
		}
	}
}

func (a *App) DecrementScroll() {
	switch a.Side {
	case FullDiff:
		if a.DiffScroll > 0 {
			a.DiffScroll--
		}
	case OldSide:
		if a.OldScroll > 0 {
			a.OldScroll--
			a.OldScrollSet = true
		}
	case NewSide:
		if a.NewScroll > 0 {
			a.NewScroll--
			a.NewScrollSet = true
		}
	}
}

func (a *App) DiffDisplayLines() []DisplayLine {
	switch a.Side {
	case OldSide:
		return NumberedFileLines(a.OldFile, 1)
	case NewSide:
		return NumberedFileLines(a.NewFile, 1)
	default:
		return diffLinesToDisplayLines(a.DiffLines)
	}
}

func diffLinesToDisplayLines(lines []diff.DiffLine) []DisplayLine {
	out := make([]DisplayLine, 0, len(lines))
	for _, line := range lines {
		out = append(out, DisplayLine{Gutter: line.Gutter, Text: line.Text})
	}
	return out
}

func NumberedFileLines(content string, start int) []DisplayLine {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	width := numberWidth(start + len(lines) - 1)
	out := make([]DisplayLine, 0, len(lines))
	for i, line := range lines {
		out = append(out, DisplayLine{
			Gutter: fmt.Sprintf("%*d", width, start+i),
			Text:   line,
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

func (a *App) NextHunk() {
	if a.Mode != DiffMode || a.Side != FullDiff || a.Diff == "" {
		return
	}
	offsets := diff.DiffHunkOffsets(a.Diff)
	for _, off := range offsets {
		if off > a.DiffScroll {
			a.DiffScroll = off
			return
		}
	}
}

func (a *App) PreviousHunk() {
	if a.Mode != DiffMode || a.Side != FullDiff || a.Diff == "" {
		return
	}
	offsets := diff.DiffHunkOffsets(a.Diff)
	previous := -1
	for _, off := range offsets {
		if off >= a.DiffScroll {
			break
		}
		previous = off
	}
	if previous >= 0 {
		a.DiffScroll = previous
	}
}

func (a *App) ToggleSide(side SideMode) {
	if a.Side == side {
		a.SwitchToSide(FullDiff)
		return
	}
	a.SwitchToSide(side)
}

func (a *App) SwitchToSide(side SideMode) {
	prev := a.Side
	prevScroll := a.CurrentScroll()

	switch {
	case prev == FullDiff && side != FullDiff:
		if side == OldSide {
			a.OldScroll = a.DiffScrollToSideScroll(OldSide)
			a.OldScrollSet = true
		} else if side == NewSide {
			a.NewScroll = a.DiffScrollToSideScroll(NewSide)
			a.NewScrollSet = true
		}
	case prev != FullDiff && side == FullDiff:
		a.DiffScroll = a.SideScrollToDiffScroll(prev, prevScroll)
	case prev != FullDiff && side != FullDiff:
		if side == OldSide {
			a.OldScroll = a.SideScrollToSideScroll(prev, prevScroll, OldSide)
			a.OldScrollSet = true
		} else if side == NewSide {
			a.NewScroll = a.SideScrollToSideScroll(prev, prevScroll, NewSide)
			a.NewScrollSet = true
		}
	}

	a.Side = side
}

func (a *App) LineNoForSide(line diff.DiffLine, side SideMode) int {
	if side == OldSide {
		return line.OldNo
	}
	return line.NewNo
}

func (a *App) DiffScrollToSideScroll(side SideMode) int {
	if len(a.DiffLines) == 0 {
		return 0
	}

	idx := a.DiffScroll
	if idx < 0 {
		idx = 0
	}
	if idx >= len(a.DiffLines) {
		idx = len(a.DiffLines) - 1
	}

	if no := a.LineNoForSide(a.DiffLines[idx], side); no > 0 {
		return no - 1
	}

	for i := idx + 1; i < len(a.DiffLines); i++ {
		if no := a.LineNoForSide(a.DiffLines[i], side); no > 0 {
			return no - 1
		}
	}
	for i := idx - 1; i >= 0; i-- {
		if no := a.LineNoForSide(a.DiffLines[i], side); no > 0 {
			return no - 1
		}
	}
	return 0
}

func (a *App) SideScrollToDiffScroll(side SideMode, scroll int) int {
	target := scroll + 1
	bestAbove := -1

	for i, line := range a.DiffLines {
		val := a.LineNoForSide(line, side)
		if val == 0 {
			continue
		}
		if val == target {
			return i
		}
		if val > target {
			bestAbove = i
			break
		}
	}
	if bestAbove != -1 {
		return bestAbove
	}

	for i := len(a.DiffLines) - 1; i >= 0; i-- {
		if val := a.LineNoForSide(a.DiffLines[i], side); val > 0 {
			return i
		}
	}
	return 0
}

func (a *App) SideScrollToSideScroll(fromSide SideMode, scroll int, toSide SideMode) int {
	diffIdx := a.SideScrollToDiffScroll(fromSide, scroll)
	if diffIdx < 0 || diffIdx >= len(a.DiffLines) {
		return 0
	}
	if no := a.LineNoForSide(a.DiffLines[diffIdx], toSide); no > 0 {
		return no - 1
	}
	return 0
}
