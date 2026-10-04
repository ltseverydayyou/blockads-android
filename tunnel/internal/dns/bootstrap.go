package dns

import (
	"context"
	"fmt"
	"net"
	"strings"
	"syscall"
)

var endpointBootstrapServers = []string{"1.1.1.1:53", "8.8.8.8:53"}

func newEndpointResolver(protectFn func(fd int) bool) *net.Resolver {
	return &net.Resolver{PreferGo: true, Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
		var lastErr error
		for _, server := range endpointBootstrapServers {
			dialer := &net.Dialer{Timeout: connectTimeout, Control: protectedControl(protectFn)}
			conn, err := dialer.DialContext(ctx, network, server)
			if err == nil {
				return conn, nil
			}
			lastErr = err
		}
		return nil, fmt.Errorf("endpoint bootstrap DNS: %w", lastErr)
	}}
}

func (r *Resolver) resolveEndpointHost(host string) ([]string, error) {
	host = strings.Trim(host, "[]")
	if ip := net.ParseIP(host); ip != nil {
		return []string{ip.String()}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()
	return r.endpointResolver.LookupHost(ctx, host)
}

func dnsServerEndpoint(server, defaultPort string) string {
	server = strings.TrimSpace(server)
	if server == "" {
		return ""
	}
	if _, _, err := net.SplitHostPort(server); err == nil {
		return server
	}
	return net.JoinHostPort(strings.Trim(server, "[]"), defaultPort)
}

func protectedControl(protectFn func(fd int) bool) func(string, string, syscall.RawConn) error {
	if protectFn == nil {
		return nil
	}
	return func(network, address string, conn syscall.RawConn) error {
		return conn.Control(func(fd uintptr) { protectFn(int(fd)) })
	}
}
