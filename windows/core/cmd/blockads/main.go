package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	core "github.com/ltseverydayyou/blockads-windows-core"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

var (
	mgr                                                                                                *core.Manager
	mw                                                                                                 *walk.MainWindow
	statusLabel, statsLabel, ssidLabel                                                                 *walk.Label
	startButton                                                                                        *walk.PushButton
	filterList, ruleList, profileList, logList                                                         *walk.ListBox
	topTabs, settingsTabs                                                                              *walk.TabWidget
	customName, customURL, ruleEntry                                                                   *walk.LineEdit
	protocolBox, responseBox, providerBox, updateFreqBox, themeBox, accentBox, routingBox              *walk.ComboBox
	upstreamEdit, fallbackEdit, dohEdit, delayEdit, splitEdit, trustedEdit                             *walk.LineEdit
	safeSearchCheck, youtubeCheck, autoReconnectCheck, autoUpdateCheck, wifiOnlyCheck, recordLogsCheck *walk.CheckBox
	pauseTrustedCheck, networkDelayCheck, firewallCheck, httpsCheck, http3Check, excludeLANCheck       *walk.CheckBox
	startWindowsCheck, minimizeTrayCheck                                                               *walk.CheckBox
	filters                                                                                            []core.FilterList
	rules                                                                                              []core.Rule
	profiles                                                                                           []core.Profile
	logs                                                                                               []core.LogEntry
)

func main() {
	var err error
	mgr, err = core.NewManager()
	if err != nil {
		walk.MsgBox(nil, "BlockAds", err.Error(), walk.MsgBoxIconError)
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--diagnostics" {
		fmt.Println(mgr.Describe())
		return
	}
	defer func() { _ = mgr.Stop(true) }()

	s := mgr.Settings()
	go func() {
		for i := 0; i < 100; i++ {
			time.Sleep(50 * time.Millisecond)
			if mw != nil {
				mw.Synchronize(initUIState)
				return
			}
		}
	}()
	if _, err = (MainWindow{
		AssignTo: &mw,
		Title:    "BlockAds for Windows",
		MinSize:  Size{Width: 980, Height: 700},
		Size:     Size{Width: 1100, Height: 760},
		Layout:   VBox{MarginsZero: true},
		Children: []Widget{
			Composite{Layout: HBox{Margins: Margins{Left: 12, Top: 10, Right: 12, Bottom: 8}}, Children: []Widget{
				Label{Text: "BlockAds", Font: Font{PointSize: 15, Bold: true}},
				HSpacer{},
				Label{AssignTo: &statusLabel, Text: "Stopped"},
				PushButton{AssignTo: &startButton, Text: "Start Protection", OnClicked: toggleProtection},
				PushButton{Text: "Update Filters", OnClicked: updateFilters},
			}},
			TabWidget{AssignTo: &topTabs, Pages: []TabPage{
				{Title: "Home", Layout: VBox{Margins: Margins{Left: 16, Top: 16, Right: 16, Bottom: 16}}, Children: []Widget{
					GroupBox{Title: "Protection", Layout: VBox{}, Children: []Widget{
						Label{AssignTo: &statsLabel, Text: "Queries: 0   Blocked: 0"},
						Label{AssignTo: &ssidLabel, Text: "Network: -"},
						PushButton{Text: "Open BlockAds data folder", OnClicked: openDataFolder},
					}},
					GroupBox{Title: "Quick controls", Layout: Grid{Columns: 2}, Children: []Widget{
						CheckBox{AssignTo: &safeSearchCheck, Text: "SafeSearch", Checked: s.SafeSearchEnabled},
						CheckBox{AssignTo: &youtubeCheck, Text: "YouTube Restricted Mode", Checked: s.YouTubeRestrictedMode},
						PushButton{Text: "Apply", OnClicked: saveSettings},
					}},
					VSpacer{},
				}},
				{Title: "Filters", Layout: VBox{Margins: Margins{Left: 12, Top: 12, Right: 12, Bottom: 12}}, Children: []Widget{
					Label{Text: "Uses the same compiled .trie/.bloom filter files as the Android app."},
					ListBox{AssignTo: &filterList, MinSize: Size{Height: 360}},
					Composite{Layout: HBox{}, Children: []Widget{
						PushButton{Text: "Enable", OnClicked: func() { setSelectedFilter(true) }},
						PushButton{Text: "Disable", OnClicked: func() { setSelectedFilter(false) }},
						PushButton{Text: "Remove custom", OnClicked: removeSelectedFilter},
						HSpacer{},
						PushButton{Text: "Refresh", OnClicked: refreshAll},
					}},
					GroupBox{Title: "Add custom filter", Layout: Grid{Columns: 3}, Children: []Widget{
						Label{Text: "Name"}, LineEdit{AssignTo: &customName}, HSpacer{},
						Label{Text: "URL"}, LineEdit{AssignTo: &customURL}, PushButton{Text: "Add", OnClicked: addCustomFilter},
					}},
				}},
				{Title: "Rules", Layout: VBox{Margins: Margins{Left: 12, Top: 12, Right: 12, Bottom: 12}}, Children: []Widget{
					Label{Text: "Android syntax: ||example.com^, @@||example.com^, *.ads.example.com, ! comment"},
					ListBox{AssignTo: &ruleList, MinSize: Size{Height: 390}},
					Composite{Layout: HBox{}, Children: []Widget{
						LineEdit{AssignTo: &ruleEntry},
						PushButton{Text: "Add", OnClicked: addRule},
						PushButton{Text: "Enable", OnClicked: func() { setSelectedRule(true) }},
						PushButton{Text: "Disable", OnClicked: func() { setSelectedRule(false) }},
						PushButton{Text: "Delete", OnClicked: deleteSelectedRule},
					}},
				}},
				{Title: "DNS", Layout: VBox{Margins: Margins{Left: 16, Top: 16, Right: 16, Bottom: 16}}, Children: []Widget{
					GroupBox{Title: "Resolver", Layout: Grid{Columns: 2}, Children: []Widget{
						Label{Text: "Provider"}, ComboBox{AssignTo: &providerBox, Model: []string{"System Default", "AdGuard DNS", "Cloudflare DNS", "Cloudflare Family", "Google DNS", "Mullvad DNS", "OpenDNS", "OpenDNS Family Shield", "Quad9"}},
						Label{Text: "Protocol"}, ComboBox{AssignTo: &protocolBox, Model: []string{"PLAIN", "DOH", "DOT", "DOQ"}},
						Label{Text: "Upstream DNS"}, LineEdit{AssignTo: &upstreamEdit, Text: s.UpstreamDNS},
						Label{Text: "Fallback DNS"}, LineEdit{AssignTo: &fallbackEdit, Text: s.FallbackDNS},
						Label{Text: "DoH / DoQ URL"}, LineEdit{AssignTo: &dohEdit, Text: s.DoHURL},
						Label{Text: "Blocked response"}, ComboBox{AssignTo: &responseBox, Model: []string{"custom_ip", "nxdomain", "refused"}},
					}},
					PushButton{Text: "Save DNS settings", OnClicked: saveSettings},
					VSpacer{},
				}},
				{Title: "Profiles", Layout: VBox{Margins: Margins{Left: 12, Top: 12, Right: 12, Bottom: 12}}, Children: []Widget{
					Label{Text: "Presets use the same filter URL sets and SafeSearch behavior as Android."},
					ListBox{AssignTo: &profileList, MinSize: Size{Height: 430}},
					PushButton{Text: "Activate selected profile", OnClicked: activateProfile},
				}},
				{Title: "Logs", Layout: VBox{Margins: Margins{Left: 12, Top: 12, Right: 12, Bottom: 12}}, Children: []Widget{
					ListBox{AssignTo: &logList, MinSize: Size{Height: 450}},
					Composite{Layout: HBox{}, Children: []Widget{
						PushButton{Text: "Refresh", OnClicked: refreshAll},
						PushButton{Text: "Clear logs", OnClicked: clearLogs},
					}},
				}},
				{Title: "Settings", Layout: VBox{Margins: Margins{Left: 14, Top: 14, Right: 14, Bottom: 14}}, Children: []Widget{
					TabWidget{AssignTo: &settingsTabs, Pages: []TabPage{
						{Title: "General", Layout: Grid{Columns: 2}, Children: []Widget{
							CheckBox{AssignTo: &autoReconnectCheck, Text: "Auto reconnect", Checked: s.AutoReconnect},
							CheckBox{AssignTo: &recordLogsCheck, Text: "Record DNS logs", Checked: s.RecordDNSLogs},
							CheckBox{AssignTo: &autoUpdateCheck, Text: "Auto-update filters", Checked: s.AutoUpdateEnabled},
							CheckBox{AssignTo: &wifiOnlyCheck, Text: "Update on Wi-Fi only", Checked: s.AutoUpdateWiFiOnly},
							Label{Text: "Update frequency"}, ComboBox{AssignTo: &updateFreqBox, Model: []string{"6h", "12h", "24h", "48h", "manual"}},
							Label{Text: "Theme"}, ComboBox{AssignTo: &themeBox, Model: []string{"system", "dark", "light"}},
							Label{Text: "Accent"}, ComboBox{AssignTo: &accentBox, Model: []string{"green", "blue", "purple", "orange", "pink", "teal", "grey", "dynamic"}},
							CheckBox{AssignTo: &startWindowsCheck, Text: "Start with Windows", Checked: s.StartWithWindows},
							CheckBox{AssignTo: &minimizeTrayCheck, Text: "Minimize to tray", Checked: s.MinimizeToTray},
						}},
						{Title: "Network", Layout: Grid{Columns: 2}, Children: []Widget{
							Label{Text: "Routing mode"}, ComboBox{AssignTo: &routingBox, Model: []string{"direct", "wireguard", "root"}},
							CheckBox{AssignTo: &excludeLANCheck, Text: "Exclude LAN", Checked: s.ExcludeLAN}, HSpacer{},
							CheckBox{AssignTo: &networkDelayCheck, Text: "Network switch delay", Checked: s.NetworkSwitchDelayEnabled},
							LineEdit{AssignTo: &delayEdit, Text: strconv.Itoa(s.NetworkSwitchDelaySec)},
							Label{Text: "Split DNS zones"}, LineEdit{AssignTo: &splitEdit, Text: s.SplitDNSZones},
							CheckBox{AssignTo: &pauseTrustedCheck, Text: "Pause on trusted Wi-Fi", Checked: s.PauseOnTrusted},
							LineEdit{AssignTo: &trustedEdit, Text: strings.Join(s.TrustedSSIDs, ",")},
						}},
						{Title: "Advanced", Layout: Grid{Columns: 2}, Children: []Widget{
							CheckBox{AssignTo: &firewallCheck, Text: "Per-app firewall", Checked: s.FirewallEnabled}, HSpacer{},
							CheckBox{AssignTo: &httpsCheck, Text: "HTTPS filtering", Checked: s.HTTPSFilteringEnabled},
							CheckBox{AssignTo: &http3Check, Text: "Filter HTTP/3", Checked: s.FilterHTTP3},
							Label{Text: "Note"}, Label{Text: "Wintun/WFP process attribution is the remaining Windows-specific layer for per-app/full HTTPS mode."},
						}},
					}},
					PushButton{Text: "Save settings", OnClicked: saveSettings},
				}},
			}},
		},
	}).Run(); err != nil {
		walk.MsgBox(nil, "BlockAds", err.Error(), walk.MsgBoxIconError)
	}
}

func showErr(err error) {
	if err != nil {
		walk.MsgBox(mw, "BlockAds", err.Error(), walk.MsgBoxIconError)
	}
}
func selected(box *walk.ComboBox, values []string, fallback string) string {
	i := box.CurrentIndex()
	if i >= 0 && i < len(values) {
		return values[i]
	}
	return fallback
}

var providerIDs = []string{"system", "adguard", "cloudflare", "cloudflare_family", "google", "mullvad", "opendns", "opendns_family", "quad9"}

func setComboByID(box *walk.ComboBox, id string) {
	for i, v := range providerIDs {
		if strings.EqualFold(v, id) {
			_ = box.SetCurrentIndex(i)
			return
		}
	}
}

func setCombo(box *walk.ComboBox, values []string, value string) {
	for i, v := range values {
		if strings.EqualFold(v, value) {
			_ = box.SetCurrentIndex(i)
			return
		}
	}
}

func refreshAll() {
	st := mgr.Status()
	filters = mgr.Filters()
	rules = mgr.Rules()
	profiles = mgr.Profiles()
	logs = mgr.Logs()
	if st.Running {
		statusLabel.SetText("Protection active")
		startButton.SetText("Stop Protection")
	} else {
		statusLabel.SetText("Protection stopped")
		startButton.SetText("Start Protection")
	}
	statsLabel.SetText(fmt.Sprintf("Queries: %d   Blocked: %d   Filters: %d   Rules: %d", st.Stats.TotalQueries, st.Stats.BlockedQueries, st.FilterCount, st.RuleCount))
	ssidLabel.SetText(fmt.Sprintf("Network: %s   Administrator: %v", st.CurrentSSID, st.Admin))
	fr := make([]string, len(filters))
	for i, f := range filters {
		mark := "OFF"
		if f.Enabled {
			mark = "ON"
		}
		fr[i] = fmt.Sprintf("[%s] %s  |  %s  |  %d rules", mark, f.Name, f.Category, f.RuleCount)
	}
	_ = filterList.SetModel(fr)
	rr := make([]string, len(rules))
	for i, r := range rules {
		mark := "OFF"
		if r.Enabled {
			mark = "ON"
		}
		rr[i] = fmt.Sprintf("[%s] %s  %s", mark, r.RuleType, r.Rule)
	}
	_ = ruleList.SetModel(rr)
	pr := make([]string, len(profiles))
	for i, p := range profiles {
		pr[i] = p.Name + " (" + p.Type + ")"
	}
	_ = profileList.SetModel(pr)
	lr := make([]string, 0, len(logs))
	start := 0
	if len(logs) > 500 {
		start = len(logs) - 500
	}
	for i := len(logs) - 1; i >= start; i-- {
		e := logs[i]
		state := "Allowed"
		if e.Blocked {
			state = "Blocked"
		}
		lr = append(lr, fmt.Sprintf("%s  %-7s  %s  %dms  %s", time.UnixMilli(e.Timestamp).Format("15:04:05"), state, e.Domain, e.ResponseTimeMS, e.BlockedBy))
	}
	_ = logList.SetModel(lr)
}

func toggleProtection() {
	st := mgr.Status()
	if st.Running {
		showErr(mgr.Stop(true))
		refreshAll()
		return
	}
	if !st.Admin {
		walk.MsgBox(mw, "BlockAds", "System-wide DNS protection needs Administrator rights. Close BlockAds and run BlockAds.exe as Administrator.", walk.MsgBoxIconWarning)
		return
	}
	showErr(mgr.Start(true))
	refreshAll()
}
func updateFilters() { _, err := mgr.UpdateFilters(); showErr(err); refreshAll() }
func setSelectedFilter(enabled bool) {
	i := filterList.CurrentIndex()
	if i < 0 || i >= len(filters) {
		return
	}
	showErr(mgr.SetFilterEnabled(filters[i].ID, enabled))
	refreshAll()
}
func removeSelectedFilter() {
	i := filterList.CurrentIndex()
	if i < 0 || i >= len(filters) {
		return
	}
	showErr(mgr.RemoveCustomFilter(filters[i].ID))
	refreshAll()
}
func addCustomFilter() {
	_, err := mgr.AddCustomFilter(customName.Text(), customURL.Text())
	showErr(err)
	if err == nil {
		customName.SetText("")
		customURL.SetText("")
	}
	refreshAll()
}
func addRule() {
	_, err := mgr.AddRule(ruleEntry.Text())
	showErr(err)
	if err == nil {
		ruleEntry.SetText("")
	}
	refreshAll()
}
func setSelectedRule(enabled bool) {
	i := ruleList.CurrentIndex()
	if i < 0 || i >= len(rules) {
		return
	}
	showErr(mgr.SetRuleEnabled(rules[i].ID, enabled))
	refreshAll()
}
func deleteSelectedRule() {
	i := ruleList.CurrentIndex()
	if i < 0 || i >= len(rules) {
		return
	}
	showErr(mgr.DeleteRule(rules[i].ID))
	refreshAll()
}
func activateProfile() {
	i := profileList.CurrentIndex()
	if i < 0 || i >= len(profiles) {
		return
	}
	showErr(mgr.ActivateProfile(profiles[i].ID))
	refreshAll()
}
func clearLogs()      { showErr(mgr.ClearLogs()); refreshAll() }
func openDataFolder() { showErr(exec.Command("explorer.exe", mgr.DataDir()).Start()) }

func saveSettings() {
	s := mgr.Settings()
	s.SafeSearchEnabled = safeSearchCheck.Checked()
	s.YouTubeRestrictedMode = youtubeCheck.Checked()
	s.AutoReconnect = autoReconnectCheck.Checked()
	s.AutoUpdateEnabled = autoUpdateCheck.Checked()
	s.AutoUpdateWiFiOnly = wifiOnlyCheck.Checked()
	s.RecordDNSLogs = recordLogsCheck.Checked()
	s.PauseOnTrusted = pauseTrustedCheck.Checked()
	s.NetworkSwitchDelayEnabled = networkDelayCheck.Checked()
	s.FirewallEnabled = firewallCheck.Checked()
	s.HTTPSFilteringEnabled = httpsCheck.Checked()
	s.FilterHTTP3 = http3Check.Checked()
	s.ExcludeLAN = excludeLANCheck.Checked()
	s.StartWithWindows = startWindowsCheck.Checked()
	s.MinimizeToTray = minimizeTrayCheck.Checked()
	if i := providerBox.CurrentIndex(); i >= 0 && i < len(providerIDs) {
		s.DNSProviderID = providerIDs[i]
	}
	s.DNSProtocol = selected(protocolBox, []string{"PLAIN", "DOH", "DOT", "DOQ"}, s.DNSProtocol)
	s.DNSResponseType = selected(responseBox, []string{"custom_ip", "nxdomain", "refused"}, s.DNSResponseType)
	s.AutoUpdateFrequency = selected(updateFreqBox, []string{"6h", "12h", "24h", "48h", "manual"}, s.AutoUpdateFrequency)
	s.ThemeMode = selected(themeBox, []string{"system", "dark", "light"}, s.ThemeMode)
	s.AccentColor = selected(accentBox, []string{"green", "blue", "purple", "orange", "pink", "teal", "grey", "dynamic"}, s.AccentColor)
	s.RoutingMode = selected(routingBox, []string{"direct", "wireguard", "root"}, s.RoutingMode)
	s.UpstreamDNS = strings.TrimSpace(upstreamEdit.Text())
	s.FallbackDNS = strings.TrimSpace(fallbackEdit.Text())
	s.DoHURL = strings.TrimSpace(dohEdit.Text())
	s.SplitDNSZones = strings.TrimSpace(splitEdit.Text())
	if n, err := strconv.Atoi(strings.TrimSpace(delayEdit.Text())); err == nil && n > 0 {
		s.NetworkSwitchDelaySec = n
	}
	ss := []string{}
	for _, x := range strings.Split(trustedEdit.Text(), ",") {
		if x = strings.TrimSpace(x); x != "" {
			ss = append(ss, x)
		}
	}
	s.TrustedSSIDs = ss
	showErr(mgr.ApplySettings(s))
	refreshAll()
}

func initUIState() {
	if topTabs != nil {
		_ = topTabs.SetCurrentIndex(0)
	}
	if settingsTabs != nil {
		_ = settingsTabs.SetCurrentIndex(0)
	}
	s := mgr.Settings()
	setCombo(protocolBox, []string{"PLAIN", "DOH", "DOT", "DOQ"}, s.DNSProtocol)
	setCombo(responseBox, []string{"custom_ip", "nxdomain", "refused"}, s.DNSResponseType)
	setCombo(updateFreqBox, []string{"6h", "12h", "24h", "48h", "manual"}, s.AutoUpdateFrequency)
	setCombo(themeBox, []string{"system", "dark", "light"}, s.ThemeMode)
	setCombo(accentBox, []string{"green", "blue", "purple", "orange", "pink", "teal", "grey", "dynamic"}, s.AccentColor)
	setCombo(routingBox, []string{"direct", "wireguard", "root"}, s.RoutingMode)
	setComboByID(providerBox, s.DNSProviderID)
	refreshAll()
}
