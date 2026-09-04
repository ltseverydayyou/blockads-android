package blockadswin

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func NewManager() (*Manager, error)               { return newManager() }
func (m *Manager) Start(useSystemDNS bool) error  { return m.start(useSystemDNS) }
func (m *Manager) Stop(restoreDNS bool) error     { return m.stop(restoreDNS) }
func (m *Manager) Shutdown(restoreDNS bool) error { return m.shutdown(restoreDNS) }
func (m *Manager) ApplySettings(s Settings) error { return m.applySettings(s) }
func (m *Manager) Stats() Stats                   { return m.stats() }
func (m *Manager) DebugSummary() string           { return m.debugSummary() }

func (m *Manager) Settings() Settings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := m.settings
	s.TrustedSSIDs = append([]string{}, s.TrustedSSIDs...)
	s.WireGuardProfiles = append([]WireGuardProfile{}, s.WireGuardProfiles...)
	return s
}

func (m *Manager) Status() Status {
	m.mu.RLock()
	running := m.running
	paused := m.pausedTrusted
	filterCount := len(m.filters)
	ruleCount := len(m.rules)
	checkSSID := m.settings.PauseOnTrusted
	m.mu.RUnlock()
	ssid := ""
	if checkSSID {
		ssid = currentSSID()
	}
	return Status{Running: running, PausedTrusted: paused, Stats: m.stats(), FilterCount: filterCount, RuleCount: ruleCount, CurrentSSID: ssid, Admin: isAdmin(), Version: "1.3.0"}
}

func (m *Manager) Filters() []FilterList {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]FilterList{}, m.filters...)
}

func (m *Manager) SetFilterEnabled(id string, enabled bool) error {
	m.mu.Lock()
	found := false
	for i := range m.filters {
		if m.filters[i].ID == id {
			m.filters[i].Enabled = enabled
			found = true
			break
		}
	}
	if found {
		m.settings.ActiveProfile = "CUSTOM"
	}
	running := m.engine != nil
	m.mu.Unlock()
	if !found {
		return errors.New("filter not found")
	}
	if err := m.saveFilters(); err != nil {
		return err
	}
	_ = m.saveSettings()
	if running {
		_, _ = m.loadEnabledFilters(false)
	}
	return nil
}

func (m *Manager) UpdateFilters() (int, error) {
	if err := m.syncFilters(); err != nil {
		return 0, err
	}
	return m.loadEnabledFilters(true)
}

func (m *Manager) AddCustomFilter(name, url string) (FilterList, error) {
	return m.addCustomFilter(name, url)
}

func (m *Manager) RemoveCustomFilter(id string) error {
	m.mu.Lock()
	idx := -1
	for i := range m.filters {
		if m.filters[i].ID == id {
			if m.filters[i].BuiltIn {
				m.mu.Unlock()
				return errors.New("built-in filters cannot be removed")
			}
			idx = i
			break
		}
	}
	if idx < 0 {
		m.mu.Unlock()
		return errors.New("filter not found")
	}
	m.filters = append(m.filters[:idx], m.filters[idx+1:]...)
	running := m.engine != nil
	m.mu.Unlock()
	for _, ext := range []string{".trie", ".bloom", ".css", ".scriptlets"} {
		_ = os.Remove(filepath.Join(m.filtersDir, id+ext))
	}
	if err := m.saveFilters(); err != nil {
		return err
	}
	if running {
		_, _ = m.loadEnabledFilters(false)
	}
	return nil
}

func (m *Manager) Profiles() []Profile             { return presetProfiles() }
func (m *Manager) ActivateProfile(id string) error { return m.activateProfile(id, true) }

func (m *Manager) Rules() []Rule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]Rule{}, m.rules...)
}

func (m *Manager) AddRule(text string) (Rule, error) {
	id := m.ids.rule.Add(1)
	r, err := parseRuleText(text, id)
	if err != nil {
		return Rule{}, err
	}
	m.mu.Lock()
	for _, x := range m.rules {
		if r.RuleType == "COMMENT" {
			if x.RuleType == "COMMENT" && x.Rule == r.Rule {
				m.mu.Unlock()
				return Rule{}, errors.New("duplicate rule")
			}
		} else if x.RuleType == r.RuleType && x.Domain == r.Domain {
			m.mu.Unlock()
			return Rule{}, errors.New("duplicate rule")
		}
	}
	m.rules = append(m.rules, r)
	m.mu.Unlock()
	if err := m.saveRules(); err != nil {
		return Rule{}, err
	}
	return r, nil
}

func (m *Manager) DeleteRule(id int64) error {
	m.mu.Lock()
	idx := -1
	for i := range m.rules {
		if m.rules[i].ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		m.mu.Unlock()
		return errors.New("rule not found")
	}
	m.rules = append(m.rules[:idx], m.rules[idx+1:]...)
	m.mu.Unlock()
	return m.saveRules()
}

func (m *Manager) SetRuleEnabled(id int64, enabled bool) error {
	m.mu.Lock()
	found := false
	for i := range m.rules {
		if m.rules[i].ID == id {
			m.rules[i].Enabled = enabled
			found = true
			break
		}
	}
	m.mu.Unlock()
	if !found {
		return errors.New("rule not found")
	}
	return m.saveRules()
}

func (m *Manager) Logs() []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]LogEntry{}, m.logs...)
}
func (m *Manager) ClearLogs() error {
	m.mu.Lock()
	m.logs = nil
	m.mu.Unlock()
	return m.saveLogs()
}

func (m *Manager) DataDir() string { return m.dataDir }
func (m *Manager) ExportState(path string) error {
	if path == "" {
		return errors.New("empty export path")
	}
	state := struct {
		Settings Settings     `json:"settings"`
		Filters  []FilterList `json:"filters"`
		Rules    []Rule       `json:"rules"`
	}{m.Settings(), m.Filters(), m.Rules()}
	return writeJSON(path, state)
}

func (m *Manager) Describe() string {
	st := m.Status()
	return fmt.Sprintf("running=%v admin=%v filters=%d rules=%d queries=%d blocked=%d", st.Running, st.Admin, st.FilterCount, st.RuleCount, st.Stats.TotalQueries, st.Stats.BlockedQueries)
}
