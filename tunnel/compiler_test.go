package tunnel

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDomainLineDNSContextSafety(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"||ads.example.com^", "ads.example.com"},
		{"||dust-0001.delorazahnow.workers.dev^", "dust-0001.delorazahnow.workers.dev"},
		{"||workers.dev^$domain=chessgames.com|player.vidzee.wtf", ""},
		{"||ads.example.com^$third-party", ""},
		{"||ads.example.com^$script", ""},
		{"||ads.example.com/path", ""},
		{"example.com##.ad-slot", ""},
		{"example.com#@#.ad-slot", ""},
		{"example.com", "example.com"},
		{"0.0.0.0 tracker.example.com", "tracker.example.com"},
		{"@@||ads.example.com^", ""},
	}

	for _, tt := range tests {
		if got := parseDomainLine(tt.line); got != tt.want {
			t.Fatalf("parseDomainLine(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestCompileFilterListDoesNotGlobalizeScopedWorkersDevRule(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "easylist.txt")
	triePath := filepath.Join(dir, "easylist.trie")
	bloomPath := filepath.Join(dir, "easylist.bloom")

	data := "||workers.dev^$domain=chessgames.com|player.vidzee.wtf\n" +
		"||dust-0001.delorazahnow.workers.dev^\n" +
		"||ads.example.com^\n"
	if err := os.WriteFile(input, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := CompileFilterList(input, triePath, bloomPath); err != nil {
		t.Fatal(err)
	}

	trie, err := LoadMmapTrie(triePath)
	if err != nil {
		t.Fatal(err)
	}
	defer trie.Close()

	if trie.ContainsOrParent("aryssucksass.ltseverydayyou.workers.dev") {
		t.Fatal("scoped EasyList workers.dev rule was incorrectly compiled as a global DNS block")
	}
	if !trie.ContainsOrParent("dust-0001.delorazahnow.workers.dev") {
		t.Fatal("context-free workers.dev ad hostname should remain blocked")
	}
	if !trie.ContainsOrParent("cdn.ads.example.com") {
		t.Fatal("context-free parent-domain blocking should remain functional")
	}
}
