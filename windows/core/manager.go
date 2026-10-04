package blockadswin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	tunnel "github.com/nqmgaming/blockads-tunnel"
)

const (
	filterMetadataURL = "https://complier.pwhs.app/api/filters/default"
	compilerURL       = "https://complier.pwhs.app/api/build"
	controlAddress    = "127.0.0.1:8754"
)

type Manager struct {
	healthQueries                                                                      atomic.Int64
	lifecycleMu                                                                        sync.Mutex
	filterMu                                                                           sync.Mutex
	startError                                                                         string
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
		UpstreamDNS: "", FallbackDNS: "", DNSProtocol: "PLAIN",
		DoHURL: "", DNSProviderID: "system",
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
	dataDir, err := prepareDataDirectory(base)
	if err != nil {
		return nil, err
	}
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
	normalizedSystemDNS := false
	if strings.EqualFold(m.settings.DNSProviderID, "system") {
		normalizedSystemDNS = m.settings.UpstreamDNS != "" || m.settings.FallbackDNS != "" || !strings.EqualFold(m.settings.DNSProtocol, "PLAIN") || m.settings.DoHURL != ""
		m.settings.UpstreamDNS = ""
		m.settings.FallbackDNS = ""
		m.settings.DNSProtocol = "PLAIN"
		m.settings.DoHURL = ""
	}
	m.loadJSON(m.filtersPath, &m.filters)
	m.loadJSON(m.rulesPath, &m.rules)
	m.loadJSON(m.logsPath, &m.logs)
	if normalizedSystemDNS {
		_ = m.saveSettings()
	}
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
	newInstall := len(m.filters) == 0
	if err := m.applyCatalog(mergeCatalog(nil)); err != nil {
		return nil, err
	}
	if newInstall {
		_ = m.activateProfile(m.settings.ActiveProfile, false)
	}
	m.startAutoUpdater()
	return m, nil
}

// filtersNeedSync detects filter metadata written by older Windows builds.
// Those builds could persist enabled built-ins without their compiled trie and
// bloom URLs, which made the app appear active while loading no blocking data.
func (m *Manager) filtersNeedSync() bool {
	for _, f := range m.filters {
		if f.BuiltIn && f.Enabled && (f.TrieURL == "" || f.BloomURL == "") {
			return true
		}
	}
	return false
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

func effectiveDNSSettings(s Settings) (string, string, string, string, error) {
	protocol := strings.ToUpper(strings.TrimSpace(s.DNSProtocol))
	primary := strings.TrimSpace(s.UpstreamDNS)
	fallback := strings.TrimSpace(s.FallbackDNS)
	dohURL := strings.TrimSpace(s.DoHURL)

	if strings.EqualFold(s.DNSProviderID, "system") {
		servers, err := systemDNSServers()
		if err != nil {
			return "", "", "", "", err
		}
		primary = servers[0]
		fallback = ""
		if len(servers) > 1 {
			fallback = servers[1]
		}
		protocol = "PLAIN"
		dohURL = ""
	}
	if strings.EqualFold(s.DNSProviderID, "mullvad") {
		protocol = "DOH"
		fallback = ""
		if dohURL == "" {
			dohURL = "https://dns.mullvad.net/dns-query"
		}
	}

	if primary == "" || primary == "0.0.0.0" || primary == "::" {
		return "", "", "", "", fmt.Errorf("invalid upstream DNS %q", primary)
	}
	if fallback == "0.0.0.0" || fallback == "::" {
		fallback = ""
	}
	return protocol, primary, fallback, dohURL, nil
}

func (m *Manager) configureEngineWithSettings(e *tunnel.Engine, s Settings) error {
	protocol, primary, fallback, dohURL, err := effectiveDNSSettings(s)
	if err != nil {
		return err
	}
	e.SetDomainChecker(m.checker)
	e.SetLogCallback(m.logCallback)
	// The callback honors RecordDNSLogs and separately counts internal health probes.
	e.SetConnLogEnabled(true)
	e.SetDNS(protocol, primary, fallback, dohURL)
	e.SetBlockResponseType(strings.ToUpper(s.DNSResponseType))
	e.SetSafeSearch(s.SafeSearchEnabled)
	e.SetYouTubeRestricted(s.YouTubeRestrictedMode)
	e.SetSplitDNSZones(s.SplitDNSZones)
	e.SetFilterHttp3(s.FilterHTTP3)
	return nil
}

func (m *Manager) configureEngine(e *tunnel.Engine) error {
	m.mu.RLock()
	s := m.settings
	m.mu.RUnlock()
	return m.configureEngineWithSettings(e, s)
}

func (m *Manager) start(useSystemDNS bool) (startErr error) {
	m.lifecycleMu.Lock()
	defer m.lifecycleMu.Unlock()
	defer func() {
		m.mu.Lock()
		if startErr != nil {
			m.startError = startErr.Error()
		} else {
			m.startError = ""
		}
		m.mu.Unlock()
	}()
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	e := tunnel.NewEngine()
	m.healthQueries.Store(0)
	m.engine = e
	m.mu.Unlock()
	if useSystemDNS {
		// Recover any physical-adapter DNS settings left behind by legacy
		// DNS-takeover builds before discovering the real system resolvers.
		_ = m.restoreDNS()
	}
	if err := m.configureEngine(e); err != nil {
		e.Stop()
		m.mu.Lock()
		m.engine = nil
		m.mu.Unlock()
		return err
	}
	if _, err := m.loadEnabledFilters(false); err != nil {
		e.Stop()
		m.mu.Lock()
		m.engine = nil
		m.mu.Unlock()
		return err
	}
	var cleanup func() error
	if useSystemDNS {
		var err error
		cleanup, err = startWindowsFullTunnel(e)
		if err != nil {
			e.Stop()
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
	port := 53
	if !useSystemDNS {
		port = m.settings.ListenPort
	}
	if err := verifyDNS(port); err != nil {
		if cleanup != nil {
			_ = cleanup()
		}
		e.Stop()
		m.mu.Lock()
		m.engine = nil
		m.mu.Unlock()
		return err
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
	return m.stopInternal(restore, true)
}

func (m *Manager) shutdown(restore bool) error {
	return m.stopInternal(restore, false)
}

func (m *Manager) stopInternal(restore bool, disableProtection bool) error {
	m.lifecycleMu.Lock()
	defer m.lifecycleMu.Unlock()
	m.mu.Lock()
	e := m.engine
	cleanup := m.systemTunnelCleanup
	m.systemTunnelCleanup = nil
	m.engine = nil
	m.running = false
	if disableProtection {
		m.settings.ProtectionEnabled = false
	}
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
	m.lifecycleMu.Lock()
	defer m.lifecycleMu.Unlock()
	if s.NetworkSwitchDelaySec <= 0 {
		s.NetworkSwitchDelaySec = 30
	}
	if s.ListenPort <= 0 {
		s.ListenPort = 53
	}
	if strings.EqualFold(s.DNSProviderID, "system") {
		s.UpstreamDNS = ""
		s.FallbackDNS = ""
		s.DoHURL = ""
	}

	m.mu.RLock()
	e := m.engine
	m.mu.RUnlock()
	if e != nil {
		if err := m.configureEngineWithSettings(e, s); err != nil {
			return err
		}
	}

	m.mu.Lock()
	m.settings = s
	m.mu.Unlock()
	if err := m.saveSettings(); err != nil {
		return err
	}
	if e != nil {
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
	health := m.healthQueries.Load()
	s.TotalQueries = max(0, s.TotalQueries-health)
	s.BlockedQueries = max(0, s.BlockedQueries-health)
	return s
}

func (m *Manager) debugSummary() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return fmt.Sprintf("running=%v filters=%d rules=%d logs=%d", m.running, len(m.filters), len(m.rules), len(m.logs))
}
