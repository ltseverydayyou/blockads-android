package blockadswin

import (
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

//go:embed assets/default_filters.json
var bundledCatalog []byte

func mergeCatalog(remote []remoteFilter) []remoteFilter {
	var bundled []remoteFilter
	_ = json.Unmarshal(bundledCatalog, &bundled)
	byURL := make(map[string]int)
	result := make([]remoteFilter, 0, len(bundled)+len(remote))
	for _, filter := range append(bundled, remote...) {
		filter.OriginalURL = strings.TrimSpace(filter.OriginalURL)
		if filter.OriginalURL == "" || filter.Name == "" {
			continue
		}
		if filter.ID == "" {
			filter.ID = fmt.Sprintf("filter_%x", sha256.Sum256([]byte(filter.OriginalURL)))
		}
		if index, ok := byURL[filter.OriginalURL]; ok {
			result[index] = filter
		} else {
			byURL[filter.OriginalURL] = len(result)
			result = append(result, filter)
		}
	}
	return result
}

func (m *Manager) applyCatalog(remote []remoteFilter) error {
	m.mu.Lock()
	old := map[string]FilterList{}
	oldByURL := map[string]FilterList{}
	custom := []FilterList{}
	for _, f := range m.filters {
		if f.BuiltIn {
			old[f.ID] = f
			oldByURL[f.OriginalURL] = f
		} else {
			custom = append(custom, f)
		}
	}
	next := make([]FilterList, 0, len(remote)+len(custom))
	now := time.Now().UnixMilli()
	for _, rf := range remote {
		enabled := rf.IsEnabled
		if o, ok := old[rf.ID]; ok {
			enabled = o.Enabled
		} else if o, ok := oldByURL[rf.OriginalURL]; ok {
			enabled = o.Enabled
		}
		cat := "AD"
		if strings.EqualFold(rf.Category, "security") {
			cat = "SECURITY"
		}
		entry := FilterList{ID: rf.ID, Name: rf.Name, URL: rf.OriginalURL, Description: rf.Description, Enabled: enabled, BuiltIn: true, Category: cat, RuleCount: rf.RuleCount, BloomURL: rf.BloomURL, TrieURL: rf.TrieURL, CSSURL: rf.CSSURL, ScriptletsURL: rf.ScriptletsURL, OriginalURL: rf.OriginalURL, LastUpdated: now}
		if previous, ok := oldByURL[rf.OriginalURL]; ok && strings.HasPrefix(previous.TrieURL, "local://") {
			entry.TrieURL, entry.BloomURL = previous.TrieURL, previous.BloomURL
			entry.RuleCount, entry.LastUpdated = previous.RuleCount, previous.LastUpdated
		}
		next = append(next, entry)
	}
	next = append(next, custom...)
	m.filters = next
	m.mu.Unlock()
	return m.saveFilters()
}
