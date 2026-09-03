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
	comment := "BlockAds full-tunnel DNS interception owner=" + pid
	script := "$targetPid=" + pid + ";try{Wait-Process -Id $targetPid -ErrorAction Stop}catch{};$owned=@(Get-DnsClientNrptRule -ErrorAction SilentlyContinue|Where-Object{$_.DisplayName -eq 'BlockAds DNS Filter' -and $_.Comment -eq '" + comment + "'});if($owned.Count -gt 0){foreach($r in $owned){Remove-DnsClientNrptRule -Name $r.Name -Force -ErrorAction SilentlyContinue};$if=Get-NetAdapter -Name 'BlockAds' -ErrorAction SilentlyContinue|Select-Object -First 1;if($null -ne $if){$idx=$if.ifIndex;Get-NetRoute -InterfaceIndex $idx -AddressFamily IPv4 -ErrorAction SilentlyContinue|Where-Object{$_.DestinationPrefix -eq '0.0.0.0/0'}|Remove-NetRoute -Confirm:$false -ErrorAction SilentlyContinue;Get-NetRoute -InterfaceIndex $idx -AddressFamily IPv6 -ErrorAction SilentlyContinue|Where-Object{$_.DestinationPrefix -eq '::/0'}|Remove-NetRoute -Confirm:$false -ErrorAction SilentlyContinue;Set-DnsClientServerAddress -InterfaceIndex $idx -ResetServerAddresses -ErrorAction SilentlyContinue;Get-NetIPAddress -InterfaceIndex $idx -ErrorAction SilentlyContinue|Where-Object{$_.IPAddress -eq '10.254.0.2' -or $_.IPAddress -eq '10.254.0.1' -or $_.IPAddress -eq 'fd00:ad:beef::2' -or $_.IPAddress -eq 'fd00:ad:beef::1'}|Remove-NetIPAddress -Confirm:$false -ErrorAction SilentlyContinue};Clear-DnsClientCache -ErrorAction SilentlyContinue}"
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	if err := cmd.Start(); err == nil {
		_ = cmd.Process.Release()
	}
}
