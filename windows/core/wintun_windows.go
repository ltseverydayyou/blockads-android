//go:build windows

package blockadswin

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
	"unsafe"

	tunnel "github.com/nqmgaming/blockads-tunnel"
	"golang.org/x/sys/windows"
	wgtun "golang.zx2c4.com/wireguard/tun"
)

const (
	blockAdsInterfaceName = "BlockAds"
	blockAdsIPv4          = "10.254.0.2"
	blockAdsDNS           = "10.254.0.1"
	blockAdsIPv6          = "fd00:ad:beef::2"
	blockAdsDNSIPv6       = "fd00:ad:beef::1"
	blockAdsMTU           = 1500
	ipUnicastIf           = 31
	ipv6UnicastIf         = 31
)

type physicalRoute struct {
	InterfaceIndex uint32 `json:"InterfaceIndex"`
	InterfaceAlias string `json:"InterfaceAlias"`
	NextHop        string `json:"NextHop"`
	Metric         int    `json:"Metric"`
}

type wintunPacketDevice struct {
	dev       wgtun.Device
	closeOnce sync.Once
}

func (w *wintunPacketDevice) Read(p []byte) (int, error) {
	bufs := [][]byte{p}
	sizes := []int{0}
	for {
		n, err := w.dev.Read(bufs, sizes, 0)
		if err != nil {
			return 0, err
		}
		if n > 0 && sizes[0] > 0 {
			return sizes[0], nil
		}
	}
}

func (w *wintunPacketDevice) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	n, err := w.dev.Write([][]byte{p}, 0)
	if err != nil {
		return 0, err
	}
	if n < 1 {
		return 0, io.ErrShortWrite
	}
	return len(p), nil
}

func (w *wintunPacketDevice) Close() error {
	var err error
	w.closeOnce.Do(func() { err = w.dev.Close() })
	return err
}

type physicalInterfaceProtector struct{ index uint32 }

func (p *physicalInterfaceProtector) Protect(fd int) bool {
	if p == nil || p.index == 0 || fd <= 0 {
		return false
	}
	h := windows.Handle(uintptr(fd))
	err4 := bindSocketToInterface4(h, p.index)
	err6 := windows.SetsockoptInt(h, windows.IPPROTO_IPV6, ipv6UnicastIf, int(p.index))
	return err4 == nil || err6 == nil
}

func bindSocketToInterface4(handle windows.Handle, index uint32) error {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], index)
	networkOrder := *(*uint32)(unsafe.Pointer(&b[0]))
	return windows.SetsockoptInt(handle, windows.IPPROTO_IP, ipUnicastIf, int(networkOrder))
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

func waitForInterface(name string) (*net.Interface, error) {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		ifc, err := net.InterfaceByName(name)
		if err == nil && ifc.Index > 0 {
			return ifc, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil, fmt.Errorf("Windows interface %q did not appear", name)
}

func configureWintun(index int) error {
	// Keep physical adapter DNS settings untouched. The virtual adapter is
	// assigned an on-link resolver address; DNS packets sent there enter the
	// full tunnel and are handled by the engine's UDP/TCP port-53 handlers.
	script := fmt.Sprintf(`$ErrorActionPreference='Stop';$idx=%d;Get-NetRoute -InterfaceIndex $idx -AddressFamily IPv4 -ErrorAction SilentlyContinue|Where-Object{$_.DestinationPrefix -eq '0.0.0.0/0'}|Remove-NetRoute -Confirm:$false -ErrorAction SilentlyContinue;Get-NetRoute -InterfaceIndex $idx -AddressFamily IPv6 -ErrorAction SilentlyContinue|Where-Object{$_.DestinationPrefix -eq '::/0'}|Remove-NetRoute -Confirm:$false -ErrorAction SilentlyContinue;Get-NetIPAddress -InterfaceIndex $idx -AddressFamily IPv4 -ErrorAction SilentlyContinue|Where-Object{$_.IPAddress -eq '%s' -or $_.IPAddress -eq '%s'}|Remove-NetIPAddress -Confirm:$false -ErrorAction SilentlyContinue;Get-NetIPAddress -InterfaceIndex $idx -AddressFamily IPv6 -ErrorAction SilentlyContinue|Where-Object{$_.IPAddress -eq '%s' -or $_.IPAddress -eq '%s'}|Remove-NetIPAddress -Confirm:$false -ErrorAction SilentlyContinue;New-NetIPAddress -InterfaceIndex $idx -IPAddress '%s' -PrefixLength 24 -AddressFamily IPv4 -SkipAsSource $false -PolicyStore ActiveStore|Out-Null;New-NetIPAddress -InterfaceIndex $idx -IPAddress '%s' -PrefixLength 64 -AddressFamily IPv6 -SkipAsSource $false -PolicyStore ActiveStore|Out-Null;Set-NetIPInterface -InterfaceIndex $idx -AddressFamily IPv4 -InterfaceMetric 1 -NlMtuBytes %d;Set-NetIPInterface -InterfaceIndex $idx -AddressFamily IPv6 -InterfaceMetric 1 -NlMtuBytes %d;Set-DnsClientServerAddress -InterfaceIndex $idx -ServerAddresses @('%s','%s');New-NetRoute -InterfaceIndex $idx -AddressFamily IPv4 -DestinationPrefix '0.0.0.0/0' -NextHop '0.0.0.0' -RouteMetric 0 -PolicyStore ActiveStore|Out-Null;New-NetRoute -InterfaceIndex $idx -AddressFamily IPv6 -DestinationPrefix '::/0' -NextHop '::' -RouteMetric 0 -PolicyStore ActiveStore|Out-Null;Clear-DnsClientCache`, index, blockAdsIPv4, blockAdsDNS, blockAdsIPv6, blockAdsDNSIPv6, blockAdsIPv4, blockAdsIPv6, blockAdsMTU, blockAdsMTU, blockAdsDNS, blockAdsDNSIPv6)
	out, err := hiddenCommand("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script).CombinedOutput()
	if err != nil {
		return fmt.Errorf("configure BlockAds Wintun: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func cleanupWintun(index int) error {
	script := fmt.Sprintf(`$idx=%d;Get-NetRoute -InterfaceIndex $idx -AddressFamily IPv4 -ErrorAction SilentlyContinue|Where-Object{$_.DestinationPrefix -eq '0.0.0.0/0'}|Remove-NetRoute -Confirm:$false -ErrorAction SilentlyContinue;Get-NetRoute -InterfaceIndex $idx -AddressFamily IPv6 -ErrorAction SilentlyContinue|Where-Object{$_.DestinationPrefix -eq '::/0'}|Remove-NetRoute -Confirm:$false -ErrorAction SilentlyContinue;Set-DnsClientServerAddress -InterfaceIndex $idx -ResetServerAddresses -ErrorAction SilentlyContinue;Get-NetIPAddress -InterfaceIndex $idx -AddressFamily IPv4 -ErrorAction SilentlyContinue|Where-Object{$_.IPAddress -eq '%s' -or $_.IPAddress -eq '%s'}|Remove-NetIPAddress -Confirm:$false -ErrorAction SilentlyContinue;Get-NetIPAddress -InterfaceIndex $idx -AddressFamily IPv6 -ErrorAction SilentlyContinue|Where-Object{$_.IPAddress -eq '%s' -or $_.IPAddress -eq '%s'}|Remove-NetIPAddress -Confirm:$false -ErrorAction SilentlyContinue;Clear-DnsClientCache`, index, blockAdsIPv4, blockAdsDNS, blockAdsIPv6, blockAdsDNSIPv6)
	out, err := hiddenCommand("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cleanup BlockAds Wintun: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func startWindowsFullTunnel(e *tunnel.Engine) (func() error, error) {
	physical, err := activePhysicalRoute()
	if err != nil {
		return nil, err
	}

	dev, err := wgtun.CreateTUN(blockAdsInterfaceName, blockAdsMTU)
	if err != nil {
		return nil, fmt.Errorf("create Wintun adapter: %w", err)
	}
	packetDevice := &wintunPacketDevice{dev: dev}
	ifc, err := waitForInterface(blockAdsInterfaceName)
	if err != nil {
		_ = packetDevice.Close()
		return nil, err
	}

	protector := &physicalInterfaceProtector{index: physical.InterfaceIndex}
	go e.StartFullDevice(packetDevice, protector)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if e.IsFullTunnelReady() {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !e.IsFullTunnelReady() {
		_ = packetDevice.Close()
		return nil, errors.New("BlockAds full-tunnel engine did not start")
	}

	if err := configureWintun(ifc.Index); err != nil {
		e.Stop()
		_ = cleanupWintun(ifc.Index)
		return nil, err
	}

	return func() error {
		return cleanupWintun(ifc.Index)
	}, nil
}

