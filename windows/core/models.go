package blockadswin

import "sync/atomic"

type Settings struct {
	ProtectionEnabled             bool               `json:"protectionEnabled"`
	AutoReconnect                 bool               `json:"autoReconnect"`
	NetworkSwitchDelayEnabled     bool               `json:"networkSwitchDelayEnabled"`
	NetworkSwitchDelaySec         int                `json:"networkSwitchDelaySec"`
	FilterURL                     string             `json:"filterUrl"`
	UpstreamDNS                   string             `json:"upstreamDns"`
	FallbackDNS                   string             `json:"fallbackDns"`
	DNSProtocol                   string             `json:"dnsProtocol"`
	DoHURL                        string             `json:"dohUrl"`
	DNSProviderID                 string             `json:"dnsProviderId"`
	ThemeMode                     string             `json:"themeMode"`
	AppLanguage                   string             `json:"appLanguage"`
	AutoUpdateEnabled             bool               `json:"autoUpdateEnabled"`
	AutoUpdateFrequency           string             `json:"autoUpdateFrequency"`
	AutoUpdateWiFiOnly            bool               `json:"autoUpdateWifiOnly"`
	AutoUpdateNotification        string             `json:"autoUpdateNotification"`
	DNSResponseType               string             `json:"dnsResponseType"`
	ProtectionLevel               string             `json:"protectionLevel"`
	SafeSearchEnabled             bool               `json:"safeSearchEnabled"`
	YouTubeRestrictedMode         bool               `json:"youtubeRestrictedMode"`
	DailySummaryEnabled           bool               `json:"dailySummaryEnabled"`
	MilestoneNotificationsEnabled bool               `json:"milestoneNotificationsEnabled"`
	AccentColor                   string             `json:"accentColor"`
	RecordDNSLogs                 bool               `json:"recordDnsLogs"`
	FirewallEnabled               bool               `json:"firewallEnabled"`
	ShowNavigationLabels          bool               `json:"showNavigationLabels"`
	RoutingMode                   string             `json:"routingMode"`
	WireGuardProfiles             []WireGuardProfile `json:"wireGuardProfiles"`
	ActiveWireGuardProfileID      string             `json:"activeWireGuardProfileId"`
	HTTPSFilteringEnabled         bool               `json:"httpsFilteringEnabled"`
	FilterHTTP3                   bool               `json:"filterHttp3"`
	CrashReportingEnabled         bool               `json:"crashReportingEnabled"`
	HideFromRecents               bool               `json:"hideFromRecents"`
	SplitDNSZones                 string             `json:"splitDnsZones"`
	ExcludeLAN                    bool               `json:"excludeLan"`
	TrustedSSIDs                  []string           `json:"trustedSsids"`
	PauseOnTrusted                bool               `json:"pauseOnTrusted"`
	ActiveProfile                 string             `json:"activeProfile"`
	ListenPort                    int                `json:"listenPort"`
	StartWithWindows              bool               `json:"startWithWindows"`
	MinimizeToTray                bool               `json:"minimizeToTray"`
}

type WireGuardProfile struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Config string `json:"config"`
}

type FilterList struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	URL           string `json:"url"`
	Description   string `json:"description"`
	Enabled       bool   `json:"isEnabled"`
	BuiltIn       bool   `json:"isBuiltIn"`
	Category      string `json:"category"`
	RuleCount     int    `json:"ruleCount"`
	BloomURL      string `json:"bloomUrl"`
	TrieURL       string `json:"trieUrl"`
	CSSURL        string `json:"cssUrl"`
	ScriptletsURL string `json:"scriptletsUrl"`
	OriginalURL   string `json:"originalUrl"`
	LastUpdated   int64  `json:"lastUpdated"`
}

type Rule struct {
	ID             int64  `json:"id"`
	Rule           string `json:"rule"`
	RuleType       string `json:"ruleType"`
	Domain         string `json:"domain"`
	Enabled        bool   `json:"isEnabled"`
	AddedTimestamp int64  `json:"addedTimestamp"`
}
type LogEntry struct {
	ID             int64  `json:"id"`
	Timestamp      int64  `json:"timestamp"`
	Domain         string `json:"domain"`
	Blocked        bool   `json:"blocked"`
	QueryType      int    `json:"queryType"`
	ResponseTimeMS int64  `json:"responseTimeMs"`
	AppName        string `json:"appName"`
	ResolvedIP     string `json:"resolvedIp"`
	BlockedBy      string `json:"blockedBy"`
}
type Profile struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	Type                  string   `json:"profileType"`
	EnabledFilterURLs     []string `json:"enabledFilterUrls"`
	SafeSearchEnabled     bool     `json:"safeSearchEnabled"`
	YouTubeRestrictedMode bool     `json:"youtubeRestrictedMode"`
}
type Stats struct {
	TotalQueries   int64 `json:"total"`
	BlockedQueries int64 `json:"blocked"`
}
type Status struct {
	Running       bool   `json:"running"`
	PausedTrusted bool   `json:"pausedTrusted"`
	Stats         Stats  `json:"stats"`
	FilterCount   int    `json:"filterCount"`
	RuleCount     int    `json:"ruleCount"`
	CurrentSSID   string `json:"currentSsid"`
	Admin         bool   `json:"admin"`
	Version       string `json:"version"`
}

type remoteFilter struct {
	Name          string `json:"name"`
	ID            string `json:"id"`
	Description   string `json:"description"`
	IsEnabled     bool   `json:"isEnabled"`
	IsBuiltIn     bool   `json:"isBuiltIn"`
	Category      string `json:"category"`
	RuleCount     int    `json:"ruleCount"`
	BloomURL      string `json:"bloomUrl"`
	TrieURL       string `json:"trieUrl"`
	CSSURL        string `json:"cssUrl"`
	ScriptletsURL string `json:"scriptletsUrl"`
	OriginalURL   string `json:"originalUrl"`
}
type compilerResponse struct {
	Status      string `json:"status"`
	DownloadURL string `json:"downloadUrl"`
	RuleCount   int    `json:"ruleCount"`
	FileSize    int64  `json:"fileSize"`
}
type dnsBackup struct {
	Index int      `json:"Index"`
	Alias string   `json:"Alias"`
	DHCP  bool     `json:"DHCP"`
	V4    []string `json:"V4"`
	V6    []string `json:"V6"`
}

type counters struct {
	rule atomic.Int64
	log  atomic.Int64
}
