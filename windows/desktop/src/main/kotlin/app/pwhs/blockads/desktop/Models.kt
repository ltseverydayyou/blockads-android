package app.pwhs.blockads.desktop

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class Stats(
    @SerialName("total") val totalQueries: Long = 0,
    @SerialName("blocked") val blockedQueries: Long = 0,
)

@Serializable
data class Status(
    val running: Boolean = false,
    val pausedTrusted: Boolean = false,
    val stats: Stats = Stats(),
    val filterCount: Int = 0,
    val ruleCount: Int = 0,
    val currentSsid: String = "",
    val admin: Boolean = false,
    val version: String = "",
)

@Serializable
data class WireGuardProfile(val id: String = "", val name: String = "", val config: String = "")

@Serializable
data class Settings(
    val protectionEnabled: Boolean = false,
    val autoReconnect: Boolean = true,
    val networkSwitchDelayEnabled: Boolean = false,
    val networkSwitchDelaySec: Int = 30,
    val filterUrl: String = "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
    val upstreamDns: String = "9.9.9.9",
    val fallbackDns: String = "94.140.14.14",
    val dnsProtocol: String = "PLAIN",
    val dohUrl: String = "https://dns.quad9.net/dns-query",
    val dnsProviderId: String = "quad9",
    val themeMode: String = "system",
    val appLanguage: String = "system",
    val autoUpdateEnabled: Boolean = true,
    val autoUpdateFrequency: String = "24h",
    val autoUpdateWifiOnly: Boolean = true,
    val autoUpdateNotification: String = "silent",
    val dnsResponseType: String = "custom_ip",
    val protectionLevel: String = "STANDARD",
    val safeSearchEnabled: Boolean = false,
    val youtubeRestrictedMode: Boolean = false,
    val dailySummaryEnabled: Boolean = false,
    val milestoneNotificationsEnabled: Boolean = false,
    val accentColor: String = "green",
    val recordDnsLogs: Boolean = true,
    val firewallEnabled: Boolean = false,
    val showNavigationLabels: Boolean = true,
    val routingMode: String = "direct",
    val wireGuardProfiles: List<WireGuardProfile> = emptyList(),
    val activeWireGuardProfileId: String = "",
    val httpsFilteringEnabled: Boolean = false,
    val filterHttp3: Boolean = false,
    val crashReportingEnabled: Boolean = false,
    val hideFromRecents: Boolean = false,
    val splitDnsZones: String = "",
    val excludeLan: Boolean = false,
    val trustedSsids: List<String> = emptyList(),
    val pauseOnTrusted: Boolean = false,
    val activeProfile: String = "DEFAULT",
    val listenPort: Int = 53,
    val startWithWindows: Boolean = false,
    val minimizeToTray: Boolean = true,
)

@Serializable
data class FilterList(
    val id: String,
    val name: String,
    val url: String = "",
    val description: String = "",
    @SerialName("isEnabled") val enabled: Boolean = true,
    @SerialName("isBuiltIn") val builtIn: Boolean = false,
    val category: String = "AD",
    val ruleCount: Int = 0,
    val bloomUrl: String = "",
    val trieUrl: String = "",
    val cssUrl: String = "",
    val scriptletsUrl: String = "",
    val originalUrl: String = "",
    val lastUpdated: Long = 0,
)

@Serializable
data class Rule(
    val id: Long,
    val rule: String,
    val ruleType: String,
    val domain: String = "",
    @SerialName("isEnabled") val enabled: Boolean = true,
    val addedTimestamp: Long = 0,
)

@Serializable
data class LogEntry(
    val id: Long,
    val timestamp: Long,
    val domain: String,
    val blocked: Boolean,
    val queryType: Int,
    val responseTimeMs: Long,
    val appName: String = "",
    val resolvedIp: String = "",
    val blockedBy: String = "",
)

@Serializable
data class Profile(
    val id: String,
    val name: String,
    val profileType: String,
    val enabledFilterUrls: List<String> = emptyList(),
    val safeSearchEnabled: Boolean = false,
    val youtubeRestrictedMode: Boolean = false,
)
