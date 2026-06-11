//go:build !windows

package main

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func enableRawMode() (func(), error) {
	saved, err := exec.Command("stty", "-g").Output()
	if err != nil {
		return nil, err
	}
	if err := exec.Command("stty", "raw", "-echo").Run(); err != nil {
		return nil, err
	}
	return func() {
		exec.Command("stty", strings.TrimSpace(string(saved))).Run()
	}, nil
}

func platformTerminalSize() (int, int) {
	out, err := exec.Command("stty", "size").Output()
	if err != nil {
		return 24, 80
	}
	fields := strings.Fields(string(out))
	if len(fields) != 2 {
		return 24, 80
	}
	rows, rowErr := strconv.Atoi(fields[0])
	cols, colErr := strconv.Atoi(fields[1])
	if rowErr != nil || colErr != nil {
		return 24, 80
	}
	return rows, cols
}
