package blockadswin

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	tunnel "github.com/nqmgaming/blockads-tunnel"
)

func TestEnsureSafeLocalDNSFilterMigratesScopedRules(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte(
			"||workers.dev^$domain=chessgames.com|player.vidzee.wtf\n" +
				"||adinplay-venatus.workers.dev^\n" +
				"||ads.example.com^\n",
		))
	}))
	defer server.Close()

	dataDir := t.TempDir()
	filtersDir := filepath.Join(dataDir, "remote_filters")
	if err := os.MkdirAll(filtersDir, 0755); err != nil {
		t.Fatal(err)
	}
	m := &Manager{dataDir: dataDir, filtersDir: filtersDir, client: server.Client()}
	filter := FilterList{ID: "easylist", URL: server.URL, OriginalURL: server.URL, BuiltIn: true}

	if err := m.ensureSafeLocalDNSFilter(&filter, true); err != nil {
		t.Fatal(err)
	}
	triePath := filepath.Join(filtersDir, "easylist.trie")
	if tunnel.CheckDomainInTrieFile(triePath, "aryssucksass.ltseverydayyou.workers.dev") {
		t.Fatal("scoped workers.dev rule became a global DNS block")
	}
	if !tunnel.CheckDomainInTrieFile(triePath, "adinplay-venatus.workers.dev") {
		t.Fatal("context-free workers.dev block should remain active")
	}
	if filter.TrieURL != "local://easylist.trie" || filter.BloomURL != "local://easylist.bloom" {
		t.Fatalf("unexpected local filter URLs: %q %q", filter.TrieURL, filter.BloomURL)
	}
	marker, err := os.ReadFile(filepath.Join(filtersDir, "easylist.dns-compiler"))
	if err != nil {
		t.Fatal(err)
	}
	if string(marker) != dnsFilterCompilerVersion+"\n" {
		t.Fatalf("unexpected compiler marker %q", string(marker))
	}

	before := requests.Load()
	if err := m.ensureSafeLocalDNSFilter(&filter, false); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != before {
		t.Fatal("current compiler version should reuse the local trie without downloading again")
	}
}
