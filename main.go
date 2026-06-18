package main

import (
	"fmt"
	"os"

	"gdv/git"

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

	fileEntries := make([]fileEntry, len(files))
	for i, f := range files {
		fileEntries[i] = fileEntry{
			Status: f.Status,
			Path:   f.Path,
			Old:    f.Old,
		}
	}

	state := &app{files: fileEntries}

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
