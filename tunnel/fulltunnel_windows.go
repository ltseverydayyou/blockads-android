//go:build windows

package tunnel

import "io"

// PacketDevice is the Windows full-tunnel packet transport. Wintun implements
// this through a small adapter that exposes its batched packet API as a normal
// io.ReadWriteCloser.
type PacketDevice interface {
	io.ReadWriteCloser
}

// StartFullDevice is the Windows equivalent of StartFull. Android supplies a
// VpnService file descriptor; Windows supplies a Wintun packet device instead.
// Both paths use the same gVisor stack, DNS engine, rules, logs and MITM logic.
func (e *Engine) StartFullDevice(device PacketDevice, protector SocketProtector) {
	if device == nil {
		logf("StartFullDevice: nil packet device")
		return
	}

	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		logf("StartFullDevice: engine already running")
		return
	}
	e.running = true
	e.totalQueries.Store(0)
	e.blockedQueries.Store(0)
	connLogSeen.Range(func(k, _ any) bool { connLogSeen.Delete(k); return true })
	uidPackageCache.Range(func(k, _ any) bool { uidPackageCache.Delete(k); return true })

	var protectFn func(fd int) bool
	if protector != nil {
		protectFn = func(fd int) bool { return protector.Protect(fd) }
	}
	e.protectFn = protectFn
	e.resolver = NewResolver(protectFn)
	e.resolver.Configure(ParseProtocol(e.protocol), e.primaryDNS, e.fallbackDNS, e.dohURL)

	certMgr := e.stackCertMgr
	filter := e.stackMitmFilter
	uidr := e.uidResolver
	done := make(chan struct{})
	e.fullTunnelDone = done
	e.fullTunnelCloser = device
	e.mu.Unlock()

	fail := func(format string, args ...interface{}) {
		logf(format, args...)
		_ = device.Close()
		e.mu.Lock()
		e.running = false
		e.fullTunnelDone = nil
		e.fullTunnelCloser = nil
		e.mu.Unlock()
	}

	mitmActive := certMgr != nil && filter != nil
	stack := NewTcpIpStack()
	stack.SetUIDResolver(uidr)
	if mitmActive {
		stack.SetTcpHandler(newMitmTcpHandler(certMgr, filter, e, uidr, protectFn))
	} else {
		stack.SetTcpHandler(newFullPassthroughTcpHandler(e, uidr, protectFn))
	}
	stack.SetUdpHandler(newFullTunnelUdpHandler(e, filter, uidr, protectFn))
	logf("StartFullDevice: mitm=%t", mitmActive)

	btun := newBufferedTun(device)
	e.mu.Lock()
	e.tcpStack = stack
	e.mu.Unlock()

	if err := stack.Start(btun, uint32(defaultTunMTU)); err != nil {
		btun.halt()
		e.mu.Lock()
		e.tcpStack = nil
		e.mu.Unlock()
		fail("StartFullDevice: stack start failed: %v", err)
		return
	}

	logf("StartFullDevice: full-network stack running through Wintun (mtu=%d)", defaultTunMTU)
	<-done
	btun.halt()
	logf("StartFullDevice: stopped")
}
