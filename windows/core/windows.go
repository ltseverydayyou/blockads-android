package blockadswin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"golang.org/x/sys/windows"
)

func currentSSID() string {
	if runtime.GOOS != "windows" {
		return ""
	}
	out, err := hiddenCommand("netsh", "wlan", "show", "interfaces").Output()
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
	return windows.GetCurrentProcessToken().IsElevated()
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
		if out, e := hiddenCommand("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", cmd).CombinedOutput(); e != nil {
			errs = append(errs, string(out))
		}
	}
	_, _ = hiddenCommand("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", "Clear-DnsClientCache").CombinedOutput()
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	_ = os.Remove(m.dnsBackupPath)
	return nil
}

