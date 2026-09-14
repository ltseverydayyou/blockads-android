//go:build windows

package main

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"
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

func startDNSCleanupWatchdog() {
	pid := strconv.Itoa(os.Getpid())
	comment := "BlockAds DNS-only interception owner=" + pid
	script := "$targetPid=" + pid + ";try{Wait-Process -Id $targetPid -ErrorAction Stop}catch{};$owned=@(Get-DnsClientNrptRule -ErrorAction SilentlyContinue|Where-Object{$_.DisplayName -eq 'BlockAds DNS Filter' -and $_.Comment -eq '" + comment + "'});if($owned.Count -gt 0){foreach($r in $owned){Remove-DnsClientNrptRule -Name $r.Name -Force -ErrorAction SilentlyContinue};Clear-DnsClientCache -ErrorAction SilentlyContinue}"
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	if err := cmd.Start(); err == nil {
		_ = cmd.Process.Release()
	}
}
