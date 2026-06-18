package main

import (
	"fmt"
	"os"

	"github.com/jroimartin/gocui"
)

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
