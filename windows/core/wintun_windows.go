//go:build windows

package blockadswin

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	tunnel "github.com/nqmgaming/blockads-tunnel"
)

const (
	blockAdsDNS     = "10.254.0.1"
	blockAdsDNSIPv6 = "fd00:ad:beef::1"
)

type physicalRoute struct {
	InterfaceIndex uint32 `json:"InterfaceIndex"`
	InterfaceAlias string `json:"InterfaceAlias"`
	NextHop        string `json:"NextHop"`
	Metric         int    `json:"Metric"`
}

func activePhysicalRoute() (physicalRoute, error) {
	const script = `$ErrorActionPreference='Stop';$r=Get-NetRoute -AddressFamily IPv4 -DestinationPrefix '0.0.0.0/0'|Where-Object{$_.InterfaceAlias -ne 'BlockAds' -and $_.NextHop -ne '0.0.0.0'}|ForEach-Object{$i=Get-NetIPInterface -AddressFamily IPv4 -InterfaceIndex $_.InterfaceIndex -ErrorAction SilentlyContinue;[pscustomobject]@{InterfaceIndex=[uint32]$_.InterfaceIndex;InterfaceAlias=$_.InterfaceAlias;NextHop=$_.NextHop;Metric=[int]($_.RouteMetric+$i.InterfaceMetric)}}|Sort-Object Metric|Select-Object -First 1;if($null -eq $r){throw 'No physical IPv4 default route found'};$r|ConvertTo-Json -Compress`
	out, err := hiddenCommand("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script).Output()
	if err != nil {
		return physicalRoute{}, fmt.Errorf("detect physical route: %w", err)
	}
	var r physicalRoute
	if err := json.Unmarshal(out, &r); err != nil {
		return physicalRoute{}, fmt.Errorf("decode physical route: %w", err)
	}
	if r.InterfaceIndex == 0 {
		return physicalRoute{}, errors.New("physical interface index is zero")
	}
	return r, nil
}

func configureDNSRedirect() error {
	script := fmt.Sprintf(`$ErrorActionPreference='Stop';$old=@(Get-DnsClientNrptRule -ErrorAction SilentlyContinue|Where-Object{$_.DisplayName -eq 'BlockAds DNS Filter'});foreach($r in $old){Remove-DnsClientNrptRule -Name $r.Name -Force -ErrorAction SilentlyContinue};Add-DnsClientNrptRule -Namespace '.' -NameServers @('127.0.0.1','::1') -DisplayName 'BlockAds DNS Filter' -Comment 'BlockAds DNS-only interception owner=%d'|Out-Null;Clear-DnsClientCache`, os.Getpid())
	out, err := hiddenCommand("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script).CombinedOutput()
	if err != nil {
		return fmt.Errorf("configure BlockAds DNS redirect: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func cleanupDNSRedirect() error {
	script := `$old=@(Get-DnsClientNrptRule -ErrorAction SilentlyContinue|Where-Object{$_.DisplayName -eq 'BlockAds DNS Filter'});foreach($r in $old){Remove-DnsClientNrptRule -Name $r.Name -Force -ErrorAction SilentlyContinue};Clear-DnsClientCache`
	out, err := hiddenCommand("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cleanup BlockAds DNS redirect: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func startWindowsFullTunnel(e *tunnel.Engine) (func() error, error) {
	if err := e.StartStandalone(53); err != nil {
		return nil, fmt.Errorf("start BlockAds loopback DNS server: %w", err)
	}
	if err := configureDNSRedirect(); err != nil {
		e.Stop()
		_ = cleanupDNSRedirect()
		return nil, err
	}
	return cleanupDNSRedirect, nil
}
