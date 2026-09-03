package tunnel

import "testing"

func TestDNSServerEndpoint(t *testing.T) {
	tests := []struct {
		name   string
		server string
		port   string
		want   string
	}{
		{name: "ipv4", server: "1.1.1.1", port: "53", want: "1.1.1.1:53"},
		{name: "ipv4 with port", server: "1.1.1.1:5353", port: "53", want: "1.1.1.1:5353"},
		{name: "ipv6", server: "2606:4700:4700::1111", port: "53", want: "[2606:4700:4700::1111]:53"},
		{name: "bracketed ipv6", server: "[2606:4700:4700::1111]", port: "53", want: "[2606:4700:4700::1111]:53"},
		{name: "ipv6 with port", server: "[2606:4700:4700::1111]:5353", port: "53", want: "[2606:4700:4700::1111]:5353"},
		{name: "hostname", server: "dns.example", port: "53", want: "dns.example:53"},
		{name: "empty", server: "", port: "53", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dnsServerEndpoint(tt.server, tt.port); got != tt.want {
				t.Fatalf("dnsServerEndpoint(%q, %q) = %q, want %q", tt.server, tt.port, got, tt.want)
			}
		})
	}
}
