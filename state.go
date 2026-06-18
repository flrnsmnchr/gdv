package main

import (
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

type diffLine struct {
	gutter string
	text   string
	oldNo  int
	newNo  int
}

type displayLine struct {
	gutter string
	text   string
}

type app struct {
	files        []fileEntry
	selected     int
	mode         mode
	side         sideMode
	diff         string
	oldFile      string
	newFile      string
	diffLines    []diffLine
	diffScroll   int
	oldScroll    int
	newScroll    int
	oldScrollSet bool
	newScrollSet bool
	statusMsg    string
	gui          *gocui.Gui
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
		{'m', a.key("m")},
		{',', a.key(",")},
		{'v', a.key("v")},
		{'c', a.key("c")},
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
	case "l", "g":
		a.toggleSide(newSide)
	case "m", "v":
		a.nextHunk()
	case ",", "c":
		a.previousHunk()
	case "j", "f", "down":
		a.incrementScroll()
	case "k", "d", "up":
		a.decrementScroll()
	}
	return false
}

func (a *app) openDiff() {
	a.mode = diffMode
	a.diffScroll = 0
	a.oldScroll = 0
	a.newScroll = 0
	a.oldScrollSet = false
	a.newScrollSet = false
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
	a.diffLines = numberedDiffLines(diff)
	a.oldFile = loadOldFile(entry)
	a.newFile = loadNewFile(entry)
}

func (a *app) currentScroll() int {
	switch a.side {
	case oldSide:
		return a.oldScroll
	case newSide:
		return a.newScroll
	default:
		return a.diffScroll
	}
}

func (a *app) setCurrentScroll(scroll int) {
	switch a.side {
	case oldSide:
		a.oldScroll = scroll
		a.oldScrollSet = true
	case newSide:
		a.newScroll = scroll
		a.newScrollSet = true
	default:
		a.diffScroll = scroll
	}
}

func (a *app) incrementScroll() {
	switch a.side {
	case fullDiff:
		if a.diffScroll < len(a.diffLines)-1 {
			a.diffScroll++
		}
	case oldSide:
		if a.oldScroll < len(a.diffDisplayLines())-1 {
			a.oldScroll++
			a.oldScrollSet = true
		}
	case newSide:
		if a.newScroll < len(a.diffDisplayLines())-1 {
			a.newScroll++
			a.newScrollSet = true
		}
	}
}

func (a *app) decrementScroll() {
	switch a.side {
	case fullDiff:
		if a.diffScroll > 0 {
			a.diffScroll--
		}
	case oldSide:
		if a.oldScroll > 0 {
			a.oldScroll--
			a.oldScrollSet = true
		}
	case newSide:
		if a.newScroll > 0 {
			a.newScroll--
			a.newScrollSet = true
		}
	}
}

func (a *app) diffDisplayLines() []displayLine {
	switch a.side {
	case oldSide:
		return numberedFileLines(a.oldFile, 1)
	case newSide:
		return numberedFileLines(a.newFile, 1)
	default:
		return diffLinesToDisplayLines(a.diffLines)
	}
}

func (a *app) nextHunk() {
	if a.mode != diffMode || a.side != fullDiff || a.diff == "" {
		return
	}
	offsets := diffHunkOffsets(a.diff)
	for _, off := range offsets {
		if off > a.diffScroll {
			a.diffScroll = off
			return
		}
	}
}

func (a *app) previousHunk() {
	if a.mode != diffMode || a.side != fullDiff || a.diff == "" {
		return
	}
	offsets := diffHunkOffsets(a.diff)
	previous := -1
	for _, off := range offsets {
		if off >= a.diffScroll {
			break
		}
		previous = off
	}
	if previous >= 0 {
		a.diffScroll = previous
	}
}

func (a *app) toggleSide(side sideMode) {
	if a.side == side {
		a.switchToSide(fullDiff)
		return
	}
	a.switchToSide(side)
}

func (a *app) switchToSide(side sideMode) {
	prev := a.side
	prevScroll := a.currentScroll()

	switch {
	case prev == fullDiff && side != fullDiff:
		if side == oldSide {
			a.oldScroll = a.diffScrollToSideScroll(oldSide)
			a.oldScrollSet = true
		} else if side == newSide {
			a.newScroll = a.diffScrollToSideScroll(newSide)
			a.newScrollSet = true
		}
	case prev != fullDiff && side == fullDiff:
		a.diffScroll = a.sideScrollToDiffScroll(prev, prevScroll)
	case prev != fullDiff && side != fullDiff:
		if side == oldSide {
			a.oldScroll = a.sideScrollToSideScroll(prev, prevScroll, oldSide)
			a.oldScrollSet = true
		} else if side == newSide {
			a.newScroll = a.sideScrollToSideScroll(prev, prevScroll, newSide)
			a.newScrollSet = true
		}
	}

	a.side = side
}

func (a *app) lineNoForSide(line diffLine, side sideMode) int {
	if side == oldSide {
		return line.oldNo
	}
	return line.newNo
}

func (a *app) diffScrollToSideScroll(side sideMode) int {
	if len(a.diffLines) == 0 {
		return 0
	}

	idx := a.diffScroll
	if idx < 0 {
		idx = 0
	}
	if idx >= len(a.diffLines) {
		idx = len(a.diffLines) - 1
	}

	if no := a.lineNoForSide(a.diffLines[idx], side); no > 0 {
		return no - 1
	}

	for i := idx + 1; i < len(a.diffLines); i++ {
		if no := a.lineNoForSide(a.diffLines[i], side); no > 0 {
			return no - 1
		}
	}
	for i := idx - 1; i >= 0; i-- {
		if no := a.lineNoForSide(a.diffLines[i], side); no > 0 {
			return no - 1
		}
	}
	return 0
}

func (a *app) sideScrollToDiffScroll(side sideMode, scroll int) int {
	target := scroll + 1
	bestAbove := -1

	for i, line := range a.diffLines {
		val := a.lineNoForSide(line, side)
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

	for i := len(a.diffLines) - 1; i >= 0; i-- {
		if val := a.lineNoForSide(a.diffLines[i], side); val > 0 {
			return i
		}
	}
	return 0
}

func (a *app) sideScrollToSideScroll(fromSide sideMode, scroll int, toSide sideMode) int {
	diffIdx := a.sideScrollToDiffScroll(fromSide, scroll)
	if diffIdx < 0 || diffIdx >= len(a.diffLines) {
		return 0
	}
	if no := a.lineNoForSide(a.diffLines[diffIdx], toSide); no > 0 {
		return no - 1
	}
	return 0
}
