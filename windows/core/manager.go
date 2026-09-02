package blockadswin

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tunnel "github.com/nqmgaming/blockads-tunnel"
)

const (
	filterMetadataURL = "https://raw.githubusercontent.com/pass-with-high-score/blockads-default-filter/refs/heads/main/output/filter_lists.json"
	compilerURL       = "https://complier.pwhs.app/api/build"
	controlAddress    = "127.0.0.1:8754"
)

type Manager struct {
	mu                                                                                 sync.RWMutex
	engine                                                                             *tunnel.Engine
	settings                                                                           Settings
	filters                                                                            []FilterList
	rules                                                                              []Rule
	logs                                                                               []LogEntry
	running                                                                            bool
	pausedTrusted                                                                      bool
	dataDir, filtersDir, settingsPath, filtersPath, rulesPath, logsPath, dnsBackupPath string
	client                                                                             *http.Client
	checker                                                                            *ruleChecker
	logCallback                                                                        *dnsLogCallback
	ids                                                                                counters
	autoUpdateStop                                                                     chan struct{}
	systemTunnelCleanup                                                                func() error
}

func defaultSettings() Settings {
	return Settings{
		AutoReconnect: true, NetworkSwitchDelaySec: 30,
		FilterURL:   "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
		UpstreamDNS: "9.9.9.9", FallbackDNS: "94.140.14.14", DNSProtocol: "PLAIN",
		DoHURL: "https://dns.quad9.net/dns-query", DNSProviderID: "quad9",
		ThemeMode: "system", AppLanguage: "system", AutoUpdateEnabled: true,
		AutoUpdateFrequency: "24h", AutoUpdateWiFiOnly: true, AutoUpdateNotification: "silent",
		DNSResponseType: "custom_ip", ProtectionLevel: "STANDARD", AccentColor: "green",
		RecordDNSLogs: true, ShowNavigationLabels: true, RoutingMode: "direct",
		ActiveProfile: "DEFAULT", ListenPort: 53, MinimizeToTray: true,
	}
}

func presetProfiles() []Profile {
	d := []string{
		"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
		"https://easylist.to/easylist/easylist.txt",
		"https://easylist.to/easylist/easyprivacy.txt",
	}
	strict := append(append([]string{}, d...),
		"https://adguardteam.github.io/AdGuardSDNSFilter/Filters/filter.txt",
		"https://pgl.yoyo.org/adservers/serverlist.php?hostformat=adblockplus&showintro=1&mimetype=plaintext",
		"https://filters.adtidy.org/extension/ublock/filters/2.txt",
		"https://filters.adtidy.org/extension/ublock/filters/11.txt")
	family := append(append([]string{}, d...),
		"https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/porn-only/hosts",
		"https://raw.githubusercontent.com/StevenBlack/hosts/master/alternates/gambling-only/hosts")
	gaming := []string{d[0], d[1]}
	seen := map[string]bool{}
	strictFamily := []string{}
	for _, s := range append(strict, family...) {
		if !seen[s] {
			seen[s] = true
			strictFamily = append(strictFamily, s)
		}
	}
	return []Profile{
		{ID: "DEFAULT", Name: "Default", Type: "DEFAULT", EnabledFilterURLs: d},
		{ID: "STRICT", Name: "Strict", Type: "STRICT", EnabledFilterURLs: strict},
		{ID: "FAMILY", Name: "Family", Type: "FAMILY", EnabledFilterURLs: family, SafeSearchEnabled: true, YouTubeRestrictedMode: true},
		{ID: "GAMING", Name: "Gaming", Type: "GAMING", EnabledFilterURLs: gaming},
		{ID: "STRICT_FAMILY", Name: "Strict Family", Type: "STRICT_FAMILY", EnabledFilterURLs: strictFamily, SafeSearchEnabled: true, YouTubeRestrictedMode: true},
	}
}

func newManager() (*Manager, error) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		base = "."
	}
	dataDir := filepath.Join(base, "BlockAds")
	filtersDir := filepath.Join(dataDir, "remote_filters")
	if err := os.MkdirAll(filtersDir, 0755); err != nil {
		return nil, err
	}
	m := &Manager{dataDir: dataDir, filtersDir: filtersDir, settings: defaultSettings(), client: &http.Client{Timeout: 45 * time.Second}}
	m.settingsPath = filepath.Join(dataDir, "settings.json")
	m.filtersPath = filepath.Join(dataDir, "filters.json")
	m.rulesPath = filepath.Join(dataDir, "rules.json")
	m.logsPath = filepath.Join(dataDir, "dns_logs.json")
	m.dnsBackupPath = filepath.Join(dataDir, "dns_backup.json")
	m.checker = &ruleChecker{m: m}
	m.logCallback = &dnsLogCallback{m: m}
	m.loadJSON(m.settingsPath, &m.settings)
	m.loadJSON(m.filtersPath, &m.filters)
	m.loadJSON(m.rulesPath, &m.rules)
	m.loadJSON(m.logsPath, &m.logs)
	for _, r := range m.rules {
		if r.ID > m.ids.rule.Load() {
			m.ids.rule.Store(r.ID)
		}
	}
	for _, e := range m.logs {
		if e.ID > m.ids.log.Load() {
			m.ids.log.Store(e.ID)
		}
	}
	if len(m.filters) == 0 {
		if err := m.syncFilters(); err != nil {
			log.Printf("initial filter sync: %v", err)
		}
		_ = m.activateProfile(m.settings.ActiveProfile, false)
	}
	m.startAutoUpdater()
	return m, nil
}

func (m *Manager) loadJSON(path string, dst any) {
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, dst)
	}
}
func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err = os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
func (m *Manager) saveSettings() error {
	m.mu.RLock()
	v := m.settings
	m.mu.RUnlock()
	return writeJSON(m.settingsPath, v)
}
func (m *Manager) saveFilters() error {
	m.mu.RLock()
	v := append([]FilterList(nil), m.filters...)
	m.mu.RUnlock()
	return writeJSON(m.filtersPath, v)
}
func (m *Manager) saveRules() error {
	m.mu.RLock()
	v := append([]Rule(nil), m.rules...)
	m.mu.RUnlock()
	return writeJSON(m.rulesPath, v)
}
func (m *Manager) saveLogs() error {
	m.mu.RLock()
	v := append([]LogEntry(nil), m.logs...)
	m.mu.RUnlock()
	return writeJSON(m.logsPath, v)
}

func (m *Manager) configureEngine(e *tunnel.Engine) {
	m.mu.RLock()
	s := m.settings
	m.mu.RUnlock()
	e.SetDomainChecker(m.checker)
	e.SetLogCallback(m.logCallback)
	e.SetConnLogEnabled(s.RecordDNSLogs)
	e.SetDNS(strings.ToUpper(s.DNSProtocol), s.UpstreamDNS, s.FallbackDNS, s.DoHURL)
	e.SetBlockResponseType(strings.ToUpper(s.DNSResponseType))
	e.SetSafeSearch(s.SafeSearchEnabled)
	e.SetYouTubeRestricted(s.YouTubeRestrictedMode)
	e.SetSplitDNSZones(s.SplitDNSZones)
	e.SetFilterHttp3(s.FilterHTTP3)
}

func (m *Manager) start(useSystemDNS bool) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	e := tunnel.NewEngine()
	m.engine = e
	m.mu.Unlock()
	m.configureEngine(e)
	if _, err := m.loadEnabledFilters(false); err != nil {
		return err
	}
	var cleanup func() error
	if useSystemDNS {
		// Undo a stale DNS-only takeover from older Windows builds before
		// switching to the real Wintun full-network path.
		_ = m.restoreDNS()
		var err error
		cleanup, err = startWindowsFullTunnel(e)
		if err != nil {
			e.Stop()
			m.mu.Lock()
			m.engine = nil
			m.mu.Unlock()
			return err
		}
		if err = m.takeOverDNS(); err != nil {
			_ = m.restoreDNS()
			e.Stop()
			_ = cleanup()
			m.mu.Lock()
			m.engine = nil
			m.mu.Unlock()
			return err
		}
	} else {
		m.mu.RLock()
		port := m.settings.ListenPort
		m.mu.RUnlock()
		if err := e.StartStandalone(port); err != nil {
			m.mu.Lock()
			m.engine = nil
			m.mu.Unlock()
			return err
		}
	}
	m.mu.Lock()
	m.systemTunnelCleanup = cleanup
	m.running = true
	m.settings.ProtectionEnabled = true
	m.pausedTrusted = false
	m.mu.Unlock()
	_ = m.saveSettings()
	return nil
}

func (m *Manager) stop(restore bool) error {
	m.mu.Lock()
	e := m.engine
	cleanup := m.systemTunnelCleanup
	m.systemTunnelCleanup = nil
	m.engine = nil
	m.running = false
	m.settings.ProtectionEnabled = false
	m.mu.Unlock()
	var err error
	if restore {
		err = m.restoreDNS()
	}
	if e != nil {
		e.Stop()
	}
	if cleanup != nil {
		if cleanupErr := cleanup(); err == nil {
			err = cleanupErr
		}
	}
	_ = m.saveSettings()
	return err
}

func (m *Manager) applySettings(s Settings) error {
	if s.NetworkSwitchDelaySec <= 0 {
		s.NetworkSwitchDelaySec = 30
	}
	if s.ListenPort <= 0 {
		s.ListenPort = 53
	}
	m.mu.Lock()
	m.settings = s
	e := m.engine
	m.mu.Unlock()
	if err := m.saveSettings(); err != nil {
		return err
	}
	if e != nil {
		m.configureEngine(e)
		_, _ = m.loadEnabledFilters(false)
	}
	m.startAutoUpdater()
	return nil
}

func (m *Manager) stats() Stats {
	m.mu.RLock()
	e := m.engine
	m.mu.RUnlock()
	if e == nil {
		return Stats{}
	}
	var s Stats
	_ = json.Unmarshal([]byte(e.GetStats()), &s)
	return s
}

func (m *Manager) startAutoUpdater() {
	m.mu.Lock()
	if m.autoUpdateStop != nil {
		close(m.autoUpdateStop)
	}
	stop := make(chan struct{})
	m.autoUpdateStop = stop
	s := m.settings
	m.mu.Unlock()
	if !s.AutoUpdateEnabled || s.AutoUpdateFrequency == "manual" {
		return
	}
	dur := 24 * time.Hour
	switch s.AutoUpdateFrequency {
	case "6h":
		dur = 6 * time.Hour
	case "12h":
		dur = 12 * time.Hour
	case "48h":
		dur = 48 * time.Hour
	}
	go func() {
		t := time.NewTicker(dur)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				_ = m.syncFilters()
				_, _ = m.loadEnabledFilters(true)
			case <-stop:
				return
			}
		}
	}()
}

func (m *Manager) trustedWatcher(ctx context.Context) {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			m.mu.RLock()
			enabled := m.settings.PauseOnTrusted
			trusted := append([]string(nil), m.settings.TrustedSSIDs...)
			running := m.running
			paused := m.pausedTrusted
			delayOn := m.settings.NetworkSwitchDelayEnabled
			delay := m.settings.NetworkSwitchDelaySec
			m.mu.RUnlock()
			if !enabled {
				continue
			}
			ssid := currentSSID()
			hit := false
			for _, s := range trusted {
				if s == ssid && ssid != "" {
					hit = true
					break
				}
			}
			if hit && running && !paused {
				if m.stop(true) == nil {
					m.mu.Lock()
					m.pausedTrusted = true
					m.mu.Unlock()
				}
			}
			if !hit && paused {
				if delayOn {
					time.Sleep(time.Duration(delay) * time.Second)
				}
				if m.start(true) == nil {
					m.mu.Lock()
					m.pausedTrusted = false
					m.mu.Unlock()
				}
			}
		}
	}
}

func (m *Manager) debugSummary() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return fmt.Sprintf("running=%v filters=%d rules=%d logs=%d", m.running, len(m.filters), len(m.rules), len(m.logs))
}
