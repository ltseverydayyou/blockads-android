package blockadswin

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()
	return port
}

func TestDNSProtectionFiltersAndRestoresAllTransports(t *testing.T) {
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("||ads.example.test^\n||workers.dev^$domain=example.test\n"))
	}))
	defer source.Close()
	upstreamPacket, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	upstream := &dns.Server{PacketConn: upstreamPacket, Handler: dns.HandlerFunc(func(w dns.ResponseWriter, query *dns.Msg) {
		reply := new(dns.Msg)
		reply.SetReply(query)
		rr, _ := dns.NewRR(query.Question[0].Name + " 60 IN A 192.0.2.42")
		reply.Answer = []dns.RR{rr}
		_ = w.WriteMsg(reply)
	})}
	go func() { _ = upstream.ActivateAndServe() }()
	defer upstream.Shutdown()
	dir := t.TempDir()
	m := &Manager{dataDir: dir, filtersDir: filepath.Join(dir, "filters"), client: source.Client(), settings: defaultSettings()}
	_ = os.MkdirAll(m.filtersDir, 0700)
	m.settings.DNSProviderID = "custom"
	m.settings.UpstreamDNS = upstreamPacket.LocalAddr().String()
	m.settings.FallbackDNS = upstreamPacket.LocalAddr().String()
	m.settings.ListenPort = testPort(t)
	m.settings.AutoUpdateEnabled = false
	m.settings.RecordDNSLogs = false
	m.settingsPath = filepath.Join(dir, "settings.json")
	m.filtersPath = filepath.Join(dir, "filters.json")
	m.logsPath = filepath.Join(dir, "logs.json")
	m.checker = &ruleChecker{m: m}
	m.logCallback = &dnsLogCallback{m: m}
	m.filters = []FilterList{{ID: "test", Name: "Test", BuiltIn: true, Enabled: true, URL: source.URL, OriginalURL: source.URL}}
	if err := m.Start(false); err != nil {
		t.Fatal(err)
	}
	defer m.Stop(false)
	if stats := m.Stats(); stats.TotalQueries != 0 || stats.BlockedQueries != 0 {
		t.Fatalf("health probes appeared in user statistics with logging disabled: %+v", stats)
	}
	if !m.Status().Running {
		t.Fatal("ready DNS protection reported stopped")
	}
	for _, host := range []string{"127.0.0.1", "::1"} {
		for _, network := range []string{"udp", "tcp"} {
			client := &dns.Client{Net: network, Timeout: time.Second}
			for domain, want := range map[string]string{"ads.example.test": "0.0.0.0", "allowed.example.test": "192.0.2.42", "app.workers.dev": "192.0.2.42"} {
				query := new(dns.Msg)
				query.SetQuestion(dns.Fqdn(domain), dns.TypeA)
				reply, _, err := client.Exchange(query, net.JoinHostPort(host, fmt.Sprint(m.settings.ListenPort)))
				if err != nil {
					t.Fatal(err)
				}
				if len(reply.Answer) == 0 {
					t.Fatalf("%s/%s/%s: no answer", host, network, domain)
				}
				if got := reply.Answer[0].(*dns.A).A.String(); got != want {
					t.Fatalf("%s/%s/%s = %s, want %s", host, network, domain, got, want)
				}
			}
		}
	}
	if _, err := m.loadEnabledFilters(true); err != nil {
		t.Fatalf("refresh active memory-mapped filters: %v", err)
	}
	query := new(dns.Msg)
	query.SetQuestion("ads.example.test.", dns.TypeA)
	client := &dns.Client{Net: "udp", Timeout: time.Second}
	reply, _, err := client.Exchange(query, net.JoinHostPort("127.0.0.1", fmt.Sprint(m.settings.ListenPort)))
	if err != nil || len(reply.Answer) != 1 || !reply.Answer[0].(*dns.A).A.IsUnspecified() {
		t.Fatalf("filtering lost after refresh: %v, %v", reply, err)
	}
	if err := m.Stop(false); err != nil {
		t.Fatal(err)
	}
	if m.Status().Running {
		t.Fatal("stopped DNS reported protected")
	}
	if err := verifyDNS(m.settings.ListenPort); err == nil {
		t.Fatal("health check passed after DNS was stopped")
	}
}

func TestStartupDoesNotReportProtectedOnPortConflict(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	dir := t.TempDir()
	m := &Manager{settings: defaultSettings(), settingsPath: filepath.Join(dir, "settings.json"), filtersPath: filepath.Join(dir, "filters.json")}
	m.settings.DNSProviderID = "custom"
	m.settings.UpstreamDNS = "127.0.0.1:9"
	m.settings.ListenPort = listener.Addr().(*net.TCPAddr).Port
	m.checker = &ruleChecker{m: m}
	m.logCallback = &dnsLogCallback{m: m}
	if err := m.Start(false); err == nil {
		t.Fatal("start succeeded with an occupied DNS TCP port")
	}
	if status := m.Status(); status.Running || status.Starting || status.Error == "" {
		t.Fatalf("bad failure status: %+v", status)
	}
}

func TestCatalogPreservesEnabledSourcesAcrossIDChange(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{filtersPath: filepath.Join(dir, "filters.json"), filters: []FilterList{{ID: "old", OriginalURL: "https://example.test/filter", Enabled: true, BuiltIn: true}}}
	remote := []remoteFilter{{ID: "new", Name: "Same source", OriginalURL: "https://example.test/filter"}}
	if err := m.applyCatalog(remote); err != nil {
		t.Fatal(err)
	}
	if len(m.filters) != 1 || m.filters[0].ID != "new" || !m.filters[0].Enabled {
		t.Fatalf("selection lost: %+v", m.filters)
	}
}
