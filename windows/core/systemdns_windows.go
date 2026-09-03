//go:build windows

package blockadswin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
)

type systemDNSResult struct {
	Servers []string `json:"Servers"`
}

func systemDNSServers() ([]string, error) {
	physical, err := activePhysicalRoute()
	if err != nil {
		return nil, err
	}

	script := fmt.Sprintf(`$ErrorActionPreference='Stop';$servers=@(Get-DnsClientServerAddress -InterfaceIndex %d -ErrorAction Stop|ForEach-Object{$_.ServerAddresses}|Where-Object{$_ -and $_ -ne '0.0.0.0' -and $_ -ne '::'});[pscustomobject]@{Servers=$servers}|ConvertTo-Json -Compress`, physical.InterfaceIndex)
	out, err := hiddenCommand("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script).Output()
	if err != nil {
		return nil, fmt.Errorf("read system DNS for %s: %w", physical.InterfaceAlias, err)
	}

	var result systemDNSResult
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("decode system DNS: %w", err)
	}

	v4 := make([]string, 0, len(result.Servers))
	v6 := make([]string, 0, len(result.Servers))
	seen := map[string]bool{}
	for _, raw := range result.Servers {
		host := strings.Trim(strings.TrimSpace(raw), "[]")
		ip := net.ParseIP(host)
		if ip == nil || ip.IsUnspecified() || ip.IsLoopback() || host == blockAdsDNS || host == blockAdsDNSIPv6 || seen[host] {
			continue
		}
		seen[host] = true
		if ip.To4() != nil {
			v4 = append(v4, host)
		} else {
			v6 = append(v6, host)
		}
	}
	servers := append(v4, v6...)
	if len(servers) == 0 {
		return nil, errors.New("active physical interface has no usable DNS servers")
	}
	return servers, nil
}
