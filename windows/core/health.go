package blockadswin

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
	"time"
)

const healthDomain = "blockads-self-test.invalid"

func verifyDNS(port int) error {
	for _, host := range []string{"127.0.0.1", "::1"} {
		for _, network := range []string{"udp", "tcp"} {
			query := new(dns.Msg)
			query.SetQuestion(dns.Fqdn(healthDomain), dns.TypeA)
			client := &dns.Client{Net: network, Timeout: 2 * time.Second}
			answer, _, err := client.Exchange(query, net.JoinHostPort(host, fmt.Sprint(port)))
			if err != nil {
				return fmt.Errorf("%s DNS health check on %s: %w", network, host, err)
			}
			blocked := answer.Rcode == dns.RcodeNameError || answer.Rcode == dns.RcodeRefused
			for _, rr := range answer.Answer {
				if a, ok := rr.(*dns.A); ok && a.A.IsUnspecified() {
					blocked = true
				}
			}
			if !blocked {
				return fmt.Errorf("%s DNS on %s did not filter the test domain", network, host)
			}
		}
	}
	return nil
}
