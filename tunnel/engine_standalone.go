package tunnel

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
)

func (e *Engine) standaloneBlock(w dns.ResponseWriter, r *dns.Msg, blockedBy, appName string, startTime time.Time) {
	m := new(dns.Msg)
	m.SetReply(r)

	switch e.responseType {
	case ResponseNXDomain:
		m.Rcode = dns.RcodeNameError
	case ResponseRefused:
		m.Rcode = dns.RcodeRefused
	default:
		m.Rcode = dns.RcodeSuccess
		if r.Question[0].Qtype == dns.TypeA {
			rr, _ := dns.NewRR(fmt.Sprintf("%s 300 IN A 0.0.0.0", r.Question[0].Name))
			m.Answer = append(m.Answer, rr)
		} else if r.Question[0].Qtype == dns.TypeAAAA {
			rr, _ := dns.NewRR(fmt.Sprintf("%s 300 IN AAAA ::", r.Question[0].Name))
			m.Answer = append(m.Answer, rr)
		}
	}

	_ = w.WriteMsg(m)
	e.totalQueries.Add(1)
	e.blockedQueries.Add(1)
	elapsed := time.Since(startTime).Milliseconds()
	e.notifyLog(strings.TrimSuffix(r.Question[0].Name, "."), true, r.Question[0].Qtype, elapsed, appName, "", blockedBy)
}

func (e *Engine) standaloneForward(w dns.ResponseWriter, r *dns.Msg, appName string, startTime time.Time) {
	raw, err := r.Pack()
	if err != nil {
		dns.HandleFailed(w, r)
		return
	}

	e.mu.Lock()
	resolver := e.resolver
	e.mu.Unlock()
	if resolver == nil {
		dns.HandleFailed(w, r)
		return
	}

	respRaw, err := resolver.Resolve(raw)
	if err != nil {
		logf("DNS resolve failed standalone %s: %v", r.Question[0].Name, err)
		dns.HandleFailed(w, r)
		e.totalQueries.Add(1)
		elapsed := time.Since(startTime).Milliseconds()
		e.notifyLog(strings.TrimSuffix(r.Question[0].Name, "."), false, r.Question[0].Qtype, elapsed, appName, "", "")
		return
	}

	var respMsg dns.Msg
	if err := respMsg.Unpack(respRaw); err != nil {
		dns.HandleFailed(w, r)
		return
	}

	if isUpstreamBlocked(respRaw) {
		e.totalQueries.Add(1)
		e.blockedQueries.Add(1)
		elapsed := time.Since(startTime).Milliseconds()
		e.notifyLog(strings.TrimSuffix(r.Question[0].Name, "."), true, r.Question[0].Qtype, elapsed, appName, "", "upstream_dns")
	} else {
		e.totalQueries.Add(1)
		elapsed := time.Since(startTime).Milliseconds()
		e.notifyLog(strings.TrimSuffix(r.Question[0].Name, "."), false, r.Question[0].Qtype, elapsed, appName, "", "")
	}

	respMsg.Id = r.Id
	_ = w.WriteMsg(&respMsg)
}

func (e *Engine) standaloneRedirect(w dns.ResponseWriter, r *dns.Msg, redirectDomain, appName string, startTime time.Time) bool {
	ip := e.safeSearch.GetCachedIP(redirectDomain)
	if ip == nil {
		e.mu.Lock()
		resolver := e.resolver
		e.mu.Unlock()
		if resolver == nil {
			return false
		}

		var err error
		ip, err = resolver.ResolveARecord(redirectDomain, e.primaryDNS)
		if err != nil {
			return false
		}
		e.safeSearch.CacheIP(redirectDomain, ip)
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Rcode = dns.RcodeSuccess

	if r.Question[0].Qtype == dns.TypeA {
		rr, _ := dns.NewRR(fmt.Sprintf("%s 300 IN A %s", r.Question[0].Name, ip.String()))
		m.Answer = append(m.Answer, rr)
	}

	_ = w.WriteMsg(m)
	e.totalQueries.Add(1)
	elapsed := time.Since(startTime).Milliseconds()
	e.notifyLog(strings.TrimSuffix(r.Question[0].Name, "."), false, r.Question[0].Qtype, elapsed, appName, ip.String(), "")
	return true
}

// StartStandalone starts the engine in DNS-only standalone mode on 127.0.0.1:port
// It bypasses TUN and directly serves incoming UDP/TCP DNS queries.
func (e *Engine) StartStandalone(port int) error {
	e.mu.Lock()

	var oldUdp, oldTcp, oldUdp6, oldTcp6 *dns.Server
	var oldResolver *Resolver

	if e.running {
		oldUdp = e.standaloneUdp
		e.standaloneUdp = nil
		oldTcp = e.standaloneTcp
		e.standaloneTcp = nil
		oldUdp6 = e.standaloneUdp6
		e.standaloneUdp6 = nil
		oldTcp6 = e.standaloneTcp6
		e.standaloneTcp6 = nil
		oldResolver = e.resolver
		e.resolver = nil
		e.running = false
	}

	e.running = true
	e.totalQueries.Store(0)
	e.blockedQueries.Store(0)

	e.resolver = NewResolver(nil)
	e.resolver.Configure(ParseProtocol(e.protocol), e.primaryDNS, e.fallbackDNS, e.dohURL)
	e.resolver.SetODoHRelay(e.odohRelayURL)
	e.mu.Unlock()

	if oldUdp != nil {
		oldUdp.Shutdown()
	}
	if oldTcp != nil {
		oldTcp.Shutdown()
	}
	if oldUdp6 != nil {
		oldUdp6.Shutdown()
	}
	if oldTcp6 != nil {
		oldTcp6.Shutdown()
	}
	if oldResolver != nil {
		oldResolver.Shutdown()
	}

	addr4 := fmt.Sprintf("127.0.0.1:%d", port)
	addr6 := fmt.Sprintf("[::1]:%d", port)

	servers := make([]*dns.Server, 0, 4)
	closeSockets := func() {
		for _, server := range servers {
			if server.Listener != nil {
				_ = server.Listener.Close()
			}
			if server.PacketConn != nil {
				_ = server.PacketConn.Close()
			}
		}
	}
	for _, endpoint := range []struct{ address, network string }{
		{addr4, "udp4"}, {addr4, "tcp4"}, {addr6, "udp6"}, {addr6, "tcp6"},
	} {
		server := &dns.Server{Addr: endpoint.address, Net: endpoint.network, Handler: dns.HandlerFunc(e.ServeDNS)}
		var err error
		if strings.HasPrefix(endpoint.network, "udp") {
			server.PacketConn, err = net.ListenPacket(endpoint.network, endpoint.address)
		} else {
			server.Listener, err = net.Listen(endpoint.network, endpoint.address)
		}
		if err != nil {
			closeSockets()
			e.mu.Lock()
			e.running = false
			resolver := e.resolver
			e.resolver = nil
			e.mu.Unlock()
			if resolver != nil {
				resolver.Shutdown()
			}
			return fmt.Errorf("bind %s DNS on %s: %w", endpoint.network, endpoint.address, err)
		}
		servers = append(servers, server)
	}
	e.mu.Lock()
	e.standaloneUdp, e.standaloneTcp = servers[0], servers[1]
	e.standaloneUdp6, e.standaloneTcp6 = servers[2], servers[3]
	e.mu.Unlock()
	for _, server := range servers {
		go func(server *dns.Server) {
			if err := server.ActivateAndServe(); err != nil {
				logf("Standalone %s stopped: %v", server.Net, err)
				e.mu.Lock()
				if e.standaloneUdp == servers[0] {
					e.running = false
				}
				e.mu.Unlock()
			}
		}(server)
	}
	logf("Engine started in STANDALONE mode on %s and %s", addr4, addr6)
	return nil
}
