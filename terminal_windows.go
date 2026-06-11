//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	enableProcessedInput = 0x0001
	enableLineInput      = 0x0002
	enableEchoInput      = 0x0004
	enableVirtualInput   = 0x0200
)

var kernel32 = syscall.NewLazyDLL("kernel32.dll")
var getConsoleMode = kernel32.NewProc("GetConsoleMode")
var setConsoleMode = kernel32.NewProc("SetConsoleMode")
var getConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")

type coord struct {
	x int16
	y int16
}

type smallRect struct {
	left   int16
	top    int16
	right  int16
	bottom int16
}

type consoleScreenBufferInfo struct {
	size              coord
	cursorPosition    coord
	attributes        uint16
	window            smallRect
	maximumWindowSize coord
}

func enableRawMode() (func(), error) {
	handle := syscall.Handle(os.Stdin.Fd())
	var original uint32
	if r, _, err := getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&original))); r == 0 {
		return nil, err
	}

	raw := original
	raw &^= enableLineInput | enableEchoInput
	raw |= enableProcessedInput | enableVirtualInput
	if r, _, err := setConsoleMode.Call(uintptr(handle), uintptr(raw)); r == 0 {
		return nil, err
	}

	return func() {
		setConsoleMode.Call(uintptr(handle), uintptr(original))
	}, nil
}

func platformTerminalSize() (int, int) {
	handle := syscall.Handle(os.Stdout.Fd())
	var info consoleScreenBufferInfo
	if r, _, _ := getConsoleScreenBufferInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&info))); r == 0 {
		return 24, 80
	}
	rows := int(info.window.bottom - info.window.top + 1)
	cols := int(info.window.right - info.window.left + 1)
	if rows <= 0 || cols <= 0 {
		return 24, 80
	}
	return rows, cols
}
