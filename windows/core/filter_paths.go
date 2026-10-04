package blockadswin

import (
	"os"
	"path/filepath"
	"strings"
)

func (m *Manager) filterPath(f FilterList, extension string) string {
	url := f.TrieURL
	if extension == ".bloom" {
		url = f.BloomURL
	}
	if strings.HasPrefix(url, "local://") {
		return filepath.Join(m.filtersDir, filepath.Base(strings.TrimPrefix(url, "local://")))
	}
	return filepath.Join(m.filtersDir, f.ID+extension)
}

func (m *Manager) cleanupRetiredFilters(filters []FilterList) {
	keep := map[string]bool{}
	for _, f := range filters {
		keep[m.filterPath(f, ".trie")] = true
		keep[m.filterPath(f, ".bloom")] = true
	}
	paths, _ := filepath.Glob(filepath.Join(m.filtersDir, "*.rev-*"))
	for _, path := range paths {
		if !keep[path] {
			_ = os.Remove(path)
		}
	}
}
