package main

import (
	"io"
	"os"
)

func readKey(r io.Reader) (string, error) {
	var b [3]byte
	n, err := r.Read(b[:1])
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "", nil
	}

	switch b[0] {
	case 3:
		return "ctrl-c", nil
	case 13, 10:
		return "enter", nil
	case 27:
		os.Stdin.Read(b[1:2])
		if b[1] == '[' {
			os.Stdin.Read(b[2:3])
			switch b[2] {
			case 'A':
				return "up", nil
			case 'B':
				return "down", nil
			}
		}
		return "esc", nil
	case 32:
		return "space", nil
	default:
		return string(b[0]), nil
	}
}

func terminalSize() (int, int) {
	return platformTerminalSize()
}
