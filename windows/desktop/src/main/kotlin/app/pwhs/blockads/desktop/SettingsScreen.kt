package app.pwhs.blockads.desktop

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material.icons.filled.Cached
import androidx.compose.material.icons.filled.ColorLens
import androidx.compose.material.icons.filled.Dns
import androidx.compose.material.icons.filled.Download
import androidx.compose.material.icons.filled.FilterList
import androidx.compose.material.icons.filled.History
import androidx.compose.material.icons.filled.Http
import androidx.compose.material.icons.filled.Language
import androidx.compose.material.icons.filled.Lock
import androidx.compose.material.icons.filled.NetworkCheck
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.Power
import androidx.compose.material.icons.filled.Save
import androidx.compose.material.icons.filled.Security
import androidx.compose.material.icons.filled.SettingsEthernet
import androidx.compose.material.icons.filled.Shield
import androidx.compose.material.icons.filled.Speed
import androidx.compose.material.icons.filled.Wifi
import androidx.compose.material.icons.filled.PlayCircle
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.RadioButton
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.launch

private data class Choice(val value: String, val label: String)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingsScreen(state: DesktopState, padding: PaddingValues, onOpenFilters: () -> Unit) {
    val scope = rememberCoroutineScope()
    var chooserTitle by remember { mutableStateOf<String?>(null) }
    var chooserChoices by remember { mutableStateOf<List<Choice>>(emptyList()) }
    var chooserSelected by remember { mutableStateOf("") }
    var chooserApply by remember { mutableStateOf<(suspend (String) -> Unit)?>(null) }

    fun openChoice(title: String, selected: String, choices: List<Choice>, apply: suspend (String) -> Unit) {
        chooserTitle = title
        chooserSelected = selected
        chooserChoices = choices
        chooserApply = apply
    }

    Column(Modifier.fillMaxSize().padding(padding)) {
        TopAppBar(
            title = { Text("Settings", fontWeight = FontWeight.Bold) },
            colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background),
        )
        Column(
            Modifier.fillMaxSize().verticalScroll(rememberScrollState()).padding(horizontal = 16.dp),
        ) {
            SectionHeader("Protection", "Configure how BlockAds protects this PC.")
            SectionCard {
                SettingsToggleItem(Icons.Default.Cached, "Auto reconnect", "Restart protection after network changes.", state.settings.autoReconnect) {
                    scope.launch { state.updateSettings { current -> current.copy(autoReconnect = it) } }
                }
                DividerInset()
                SettingItem(Icons.Default.SettingsEthernet, "Routing mode", routingLabel(state.settings.routingMode)) {
                    openChoice(
                        "Routing mode",
                        state.settings.routingMode,
                        listOf(Choice("direct", "Direct / Local DNS"), Choice("wireguard", "WireGuard"), Choice("root", "Root Proxy")),
                    ) { v -> state.updateSettings { current -> current.copy(routingMode = v) } }
                }
                DividerInset()
                SettingsToggleItem(Icons.Default.Speed, "Network switch delay", "Wait before restarting protection after a network change.", state.settings.networkSwitchDelayEnabled) {
                    scope.launch { state.updateSettings { current -> current.copy(networkSwitchDelayEnabled = it) } }
                }
                DividerInset()
                SettingsToggleItem(Icons.Default.Security, "SafeSearch", "Force supported search engines into SafeSearch.", state.settings.safeSearchEnabled) {
                    scope.launch { state.updateSettings { current -> current.copy(safeSearchEnabled = it) } }
                }
                DividerInset()
                SettingsToggleItem(Icons.Default.PlayCircle, "YouTube Restricted Mode", "Enforce YouTube Restricted Mode through DNS.", state.settings.youtubeRestrictedMode) {
                    scope.launch { state.updateSettings { current -> current.copy(youtubeRestrictedMode = it) } }
                }
                DividerInset()
                SettingItem(Icons.Default.Dns, "DNS provider", dnsProviderLabel(state.settings.dnsProviderId)) {
                    openChoice("DNS provider", state.settings.dnsProviderId, dnsProviders()) { id ->
                        val p = dnsProviderValues(id)
                        state.updateSettings { current ->
                            current.copy(
                                dnsProviderId = id,
                                upstreamDns = p.primary,
                                fallbackDns = p.fallback,
                                dohUrl = p.dohUrl,
                                dnsProtocol = when (id) {
                                    "system" -> "PLAIN"
                                    "mullvad" -> "DOH"
                                    else -> current.dnsProtocol
                                },
                            )
                        }
                    }
                }
                DividerInset()
                SettingItem(Icons.Default.NetworkCheck, "DNS protocol", state.settings.dnsProtocol) {
                    openChoice("DNS protocol", state.settings.dnsProtocol, listOf(Choice("PLAIN", "Plain DNS"), Choice("DOH", "DNS-over-HTTPS"), Choice("DOT", "DNS-over-TLS"), Choice("DOQ", "DNS-over-QUIC"))) { v ->
                        state.updateSettings { current -> current.copy(dnsProtocol = v) }
                    }
                }
                DividerInset()
                SettingItem(Icons.Default.Lock, "Blocked response", responseLabel(state.settings.dnsResponseType)) {
                    openChoice("Blocked response", state.settings.dnsResponseType, listOf(Choice("custom_ip", "0.0.0.0"), Choice("nxdomain", "NXDOMAIN"), Choice("refused", "REFUSED"))) { v ->
                        state.updateSettings { current -> current.copy(dnsResponseType = v) }
                    }
                }
            }

            Spacer(Modifier.height(24.dp))
            SectionHeader("Interface", "Appearance and navigation.")
            SectionCard {
                SettingItem(Icons.Default.ColorLens, "Theme", state.settings.themeMode.replaceFirstChar { it.uppercase() }) {
                    openChoice("Theme", state.settings.themeMode, listOf(Choice("system", "System"), Choice("dark", "Dark"), Choice("light", "Light"))) { v -> state.updateSettings { current -> current.copy(themeMode = v) } }
                }
                DividerInset()
                SettingItem(Icons.Default.AutoAwesome, "Accent color", state.settings.accentColor.replaceFirstChar { it.uppercase() }) {
                    openChoice("Accent color", state.settings.accentColor, listOf("green", "blue", "purple", "orange", "pink", "teal", "grey").map { Choice(it, it.replaceFirstChar(Char::uppercase)) }) { v -> state.updateSettings { current -> current.copy(accentColor = v) } }
                }
                DividerInset()
                SettingsToggleItem(Icons.Default.Language, "Navigation labels", "Show labels under the bottom navigation icons.", state.settings.showNavigationLabels) {
                    scope.launch { state.updateSettings { current -> current.copy(showNavigationLabels = it) } }
                }
            }

            Spacer(Modifier.height(24.dp))
            SectionHeader("Applications", "Windows equivalents of the Android app-management options.")
            SectionCard {
                SettingsToggleItem(Icons.Default.Shield, "Firewall", "Enable per-application firewall policy.", state.settings.firewallEnabled) {
                    scope.launch { state.updateSettings { current -> current.copy(firewallEnabled = it) } }
                }
                DividerInset()
                SettingsToggleItem(Icons.Default.Wifi, "Pause on trusted Wi-Fi", "Pause protection on trusted wireless networks.", state.settings.pauseOnTrusted) {
                    scope.launch { state.updateSettings { current -> current.copy(pauseOnTrusted = it) } }
                }
                DividerInset()
                SettingsToggleItem(Icons.Default.Http, "HTTPS filtering", "Enable BlockAds HTTPS filtering when the Windows tunnel layer is available.", state.settings.httpsFilteringEnabled) {
                    scope.launch { state.updateSettings { current -> current.copy(httpsFilteringEnabled = it) } }
                }
                DividerInset()
                SettingsToggleItem(Icons.Default.Power, "Filter HTTP/3", "Filter QUIC/HTTP3 traffic when full-tunnel mode is active.", state.settings.filterHttp3) {
                    scope.launch { state.updateSettings { current -> current.copy(filterHttp3 = it) } }
                }
            }

            Spacer(Modifier.height(24.dp))
            SectionHeader("Filters", "Manage filter lists and automatic updates.")
            SectionCard {
                SettingItem(Icons.Default.FilterList, "Filter setup", "${state.filters.count { it.enabled }} active lists") { onOpenFilters() }
                DividerInset()
                SettingsToggleItem(Icons.Default.Download, "Automatic updates", "Keep enabled filter lists up to date.", state.settings.autoUpdateEnabled) {
                    scope.launch { state.updateSettings { current -> current.copy(autoUpdateEnabled = it) } }
                }
                DividerInset()
                SettingItem(Icons.Default.History, "Update frequency", state.settings.autoUpdateFrequency) {
                    openChoice("Update frequency", state.settings.autoUpdateFrequency, listOf("6h", "12h", "24h", "48h", "manual").map { Choice(it, if (it == "manual") "Manual" else "Every $it") }) { v -> state.updateSettings { current -> current.copy(autoUpdateFrequency = v) } }
                }
                DividerInset()
                SettingsToggleItem(Icons.Default.Wifi, "Update on Wi-Fi only", "Avoid automatic filter downloads on metered/mobile connections.", state.settings.autoUpdateWifiOnly) {
                    scope.launch { state.updateSettings { current -> current.copy(autoUpdateWifiOnly = it) } }
                }
            }

            Spacer(Modifier.height(24.dp))
            SectionHeader("Logs & privacy")
            SectionCard {
                SettingsToggleItem(Icons.Default.History, "Record DNS logs", "Store DNS query history locally on this PC.", state.settings.recordDnsLogs) {
                    scope.launch { state.updateSettings { current -> current.copy(recordDnsLogs = it) } }
                }
                DividerInset()
                SettingsToggleItem(Icons.Default.Notifications, "Daily summary", "Daily protection summary notification.", state.settings.dailySummaryEnabled) {
                    scope.launch { state.updateSettings { current -> current.copy(dailySummaryEnabled = it) } }
                }
                DividerInset()
                SettingsToggleItem(Icons.Default.Notifications, "Milestone notifications", "Notify when BlockAds reaches blocking milestones.", state.settings.milestoneNotificationsEnabled) {
                    scope.launch { state.updateSettings { current -> current.copy(milestoneNotificationsEnabled = it) } }
                }
            }

            Spacer(Modifier.height(24.dp))
            SectionHeader("Windows")
            SectionCard {
                SettingsToggleItem(Icons.Default.Power, "Start with Windows", "Launch BlockAds when you sign in.", state.settings.startWithWindows) {
                    scope.launch { state.updateSettings { current -> current.copy(startWithWindows = it) } }
                }
                DividerInset()
                SettingsToggleItem(Icons.Default.Save, "Minimize to tray", "Keep BlockAds running when its window is closed/minimized.", state.settings.minimizeToTray) {
                    scope.launch { state.updateSettings { current -> current.copy(minimizeToTray = it) } }
                }
            }

            Spacer(Modifier.height(24.dp))
            Text("BlockAds for Windows", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
            Text("Desktop port using the original BlockAds DNS/filter engine and the Android app's Material 3 visual system.", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
            Spacer(Modifier.height(120.dp))
        }
    }

    if (chooserTitle != null) {
        ChoiceDialog(
            title = chooserTitle!!,
            selected = chooserSelected,
            choices = chooserChoices,
            onDismiss = { chooserTitle = null },
            onSelect = { v ->
                chooserTitle = null
                chooserApply?.let { apply -> scope.launch { apply(v) } }
            },
        )
    }
}

@Composable
private fun ChoiceDialog(title: String, selected: String, choices: List<Choice>, onDismiss: () -> Unit, onSelect: (String) -> Unit) {
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(title) },
        text = {
            Column {
                choices.forEach { choice ->
                    Row(Modifier.fillMaxWidth().padding(vertical = 4.dp), verticalAlignment = Alignment.CenterVertically) {
                        RadioButton(selected = choice.value == selected, onClick = { onSelect(choice.value) })
                        Spacer(Modifier.width(8.dp))
                        Text(choice.label)
                    }
                }
            }
        },
        confirmButton = { TextButton(onClick = onDismiss) { Text("Cancel") } },
    )
}

private fun routingLabel(value: String) = when (value) { "wireguard" -> "WireGuard Mode"; "root" -> "Root Proxy Mode"; else -> "Direct / Local DNS" }
private fun responseLabel(value: String) = when (value.lowercase()) { "nxdomain" -> "NXDOMAIN"; "refused" -> "REFUSED"; else -> "0.0.0.0" }
private fun dnsProviderLabel(id: String) = dnsProviders().firstOrNull { it.value == id }?.label ?: id
private fun dnsProviders() = listOf(
    Choice("system", "System Default"), Choice("adguard", "AdGuard DNS"), Choice("cloudflare", "Cloudflare DNS"), Choice("cloudflare_family", "Cloudflare Family"),
    Choice("google", "Google DNS"), Choice("mullvad", "Mullvad DNS"), Choice("opendns", "OpenDNS"), Choice("opendns_family", "OpenDNS Family Shield"), Choice("quad9", "Quad9"),
)
private data class DnsProviderValue(val primary: String, val fallback: String, val dohUrl: String)

private fun dnsProviderValues(id: String): DnsProviderValue = when (id) {
    "adguard" -> DnsProviderValue("94.140.14.14", "94.140.15.15", "https://dns.adguard-dns.com/dns-query")
    "cloudflare" -> DnsProviderValue("1.1.1.1", "1.0.0.1", "https://cloudflare-dns.com/dns-query")
    "cloudflare_family" -> DnsProviderValue("1.1.1.3", "1.0.0.3", "https://family.cloudflare-dns.com/dns-query")
    "google" -> DnsProviderValue("8.8.8.8", "8.8.4.4", "https://dns.google/dns-query")
    "mullvad" -> DnsProviderValue("194.242.2.2", "", "https://dns.mullvad.net/dns-query")
    "opendns" -> DnsProviderValue("208.67.222.222", "208.67.220.220", "")
    "opendns_family" -> DnsProviderValue("208.67.222.123", "208.67.220.123", "")
    "quad9" -> DnsProviderValue("9.9.9.9", "149.112.112.112", "https://dns.quad9.net/dns-query")
    else -> DnsProviderValue("", "", "")
}
