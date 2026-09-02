//go:build windows

package main

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

var shellExecuteW = windows.NewLazySystemDLL("shell32.dll").NewProc("ShellExecuteW")

func ensureElevated() bool {
	if windows.GetCurrentProcessToken().IsElevated() {
		return true
	}

	exe, err := os.Executable()
	if err != nil {
		return false
	}
	verb, _ := windows.UTF16PtrFromString("runas")
	file, _ := windows.UTF16PtrFromString(exe)
	cwd, _ := os.Getwd()
	dir, _ := windows.UTF16PtrFromString(cwd)

	_, _, _ = shellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		0,
		uintptr(unsafe.Pointer(dir)),
		0,
	)
	return false
}
