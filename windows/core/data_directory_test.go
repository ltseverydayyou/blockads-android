package blockadswin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSavedDataSurvivesInstallerDirectoryRemoval(t *testing.T) {
	base := t.TempDir()
	legacy := filepath.Join(base, "BlockAds")
	if err := os.MkdirAll(filepath.Join(legacy, "remote_filters"), 0700); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"settings.json":              `{"protectionEnabled":true,"dnsProviderId":"custom"}`,
		"filters.json":               `[{"id":"custom","enabled":true}]`,
		"remote_filters/custom.trie": "cached filter bytes",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(legacy, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	dir, err := prepareDataDirectory(base)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(legacy); err != nil {
		t.Fatal(err)
	}
	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || string(got) != want {
			t.Fatalf("saved %s = %q, %v; want %q", name, got, err, want)
		}
	}
	if _, err := prepareDataDirectory(base); err != nil {
		t.Fatal(err)
	}
}

func TestMigrationKeepsNewerSavedSettings(t *testing.T) {
	base := t.TempDir()
	legacy := filepath.Join(base, "BlockAds")
	current := filepath.Join(base, "BlockAdsData")
	for _, dir := range []string{legacy, current} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(legacy, "settings.json"), []byte("old settings"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(current, "settings.json"), []byte("new settings"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := prepareDataDirectory(base); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(current, "settings.json"))
	if err != nil || string(got) != "new settings" {
		t.Fatalf("new settings overwritten: %q, %v", got, err)
	}
}
