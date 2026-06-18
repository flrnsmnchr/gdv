package main

import (
	"fmt"
	"os"

	"gdv/git"
	"gdv/state"
	"gdv/ui"

	"github.com/jroimartin/gocui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gdv:", err)
		os.Exit(1)
	}
}

func run() error {
	if err := git.EnsureGitRepo(); err != nil {
		return err
	}

	files, err := git.LoadStatus()
	if err != nil {
		return err
	}

	fileEntries := make([]state.FileEntry, len(files))
	for i, f := range files {
		fileEntries[i] = state.FileEntry{
			Status: f.Status,
			Path:   f.Path,
			Old:    f.Old,
		}
	}

	appState := &state.App{Files: fileEntries}

	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	defer gui.Close()

	appState.Gui = gui
	gui.SetManagerFunc(func(g *gocui.Gui) error { return ui.Layout(appState, g) })

	if err := appState.SetKeybindings(); err != nil {
		return err
	}
	if err := gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}
