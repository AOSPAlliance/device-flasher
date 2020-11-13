//+build windows,GUI

package main

import (
	"syscall"
)

func init() {
	getConsoleWindow := syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleWindow")
	showWindow := syscall.NewLazyDLL("user32.dll").NewProc("ShowWindow")
	hwnd, _, _ := getConsoleWindow.Call()
	if hwnd == 0 {
		return
	}
	_, _, _ = showWindow.Call(hwnd, 0)
}
