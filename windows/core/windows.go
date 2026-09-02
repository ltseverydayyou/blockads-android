package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func currentSSID() string {
	if runtime.GOOS != "windows" {
		return ""
	}
	out, err := exec.Command("netsh", "wlan", "show", "interfaces").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		p := strings.SplitN(line, ":", 2)
		if len(p) == 2 && strings.TrimSpace(p[0]) == "SSID" {
			return strings.TrimSpace(p[1])
		}
	}
	return ""
}
func isAdmin() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", "$p=New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent());if($p.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)){'true'}else{'false'}")
	out, err := cmd.Output()
	return err == nil && strings.Contains(strings.ToLower(string(out)), "true")
}

func (m *Manager) takeOverDNS() error {
	if runtime.GOOS != "windows" {
		return nil
	}
	if !isAdmin() {
		return errors.New("administrator privileges are required to change Windows DNS")
	}
	capture := `$a=Get-NetAdapter|? Status -eq 'Up';$o=@();foreach($x in $a){$c=Get-CimInstance Win32_NetworkAdapterConfiguration|? InterfaceIndex -eq $x.InterfaceIndex|select -First 1;$v4=(Get-DnsClientServerAddress -InterfaceIndex $x.InterfaceIndex -AddressFamily IPv4).ServerAddresses;$v6=(Get-DnsClientServerAddress -InterfaceIndex $x.InterfaceIndex -AddressFamily IPv6).ServerAddresses;$o+=[pscustomobject]@{Index=$x.InterfaceIndex;Alias=$x.Name;DHCP=[bool]$c.DHCPEnabled;V4=@($v4);V6=@($v6)}};$o|ConvertTo-Json -Compress -Depth 5`
	out, err := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", capture).Output()
	if err != nil {
		return fmt.Errorf("capture DNS: %w", err)
	}
	if err = os.WriteFile(m.dnsBackupPath, out, 0600); err != nil {
		return err
	}
	setcmd := `Get-NetAdapter|? Status -eq 'Up'|%{Set-DnsClientServerAddress -InterfaceIndex $_.InterfaceIndex -ServerAddresses @('127.0.0.1','::1') -ErrorAction Stop};Clear-DnsClientCache`
	b, err := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", setcmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("set DNS: %v: %s", err, string(b))
	}
	return nil
}

func (m *Manager) restoreDNS() error {
	if runtime.GOOS != "windows" {
		return nil
	}
	b, err := os.ReadFile(m.dnsBackupPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	arr := []dnsBackup{}
	trim := bytes.TrimSpace(b)
	if len(trim) > 0 && trim[0] == '{' {
		var one dnsBackup
		if err = json.Unmarshal(b, &one); err == nil {
			arr = []dnsBackup{one}
		}
	} else {
		_ = json.Unmarshal(b, &arr)
	}
	errs := []string{}
	for _, x := range arr {
		var cmd string
		if x.DHCP {
			cmd = fmt.Sprintf("Set-DnsClientServerAddress -InterfaceIndex %d -ResetServerAddresses -ErrorAction Stop", x.Index)
		} else {
			all := append(append([]string{}, x.V4...), x.V6...)
			parts := make([]string, len(all))
			for i, s := range all {
				parts[i] = "'" + strings.ReplaceAll(s, "'", "''") + "'"
			}
			cmd = fmt.Sprintf("Set-DnsClientServerAddress -InterfaceIndex %d -ServerAddresses @(%s) -ErrorAction Stop", x.Index, strings.Join(parts, ","))
		}
		if out, e := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", cmd).CombinedOutput(); e != nil {
			errs = append(errs, string(out))
		}
	}
	_, _ = exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", "Clear-DnsClientCache").CombinedOutput()
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	_ = os.Remove(m.dnsBackupPath)
	return nil
}
