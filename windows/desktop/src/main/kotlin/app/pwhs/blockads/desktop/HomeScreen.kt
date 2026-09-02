package app.pwhs.blockads.desktop

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.IntrinsicSize
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Block
import androidx.compose.material.icons.filled.DataSaverOn
import androidx.compose.material.icons.filled.GppGood
import androidx.compose.material.icons.filled.QueryStats
import androidx.compose.material.icons.filled.Shield
import androidx.compose.material.icons.filled.Timer
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.launch
import java.awt.Desktop
import java.net.URI

private fun browse(url: String) = runCatching { if (Desktop.isDesktopSupported()) Desktop.getDesktop().browse(URI(url)) }

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(
    state: DesktopState,
    padding: PaddingValues,
    onLogs: () -> Unit,
    onStatistics: () -> Unit,
    onProfiles: () -> Unit,
) {
    val scope = rememberCoroutineScope()
    val running = state.status.running
    val activeProfile = state.profiles.firstOrNull { it.id == state.settings.activeProfile }
    val total = state.status.stats.totalQueries
    val blocked = state.status.stats.blockedQueries
    val securityBlocked = state.logs.count { it.blocked && it.blockedBy.contains("security", ignoreCase = true) }.toLong()
    val enabledRules = state.filters.filter { it.enabled }.sumOf { it.ruleCount.toLong() }
    val blockRate = if (total > 0) blocked * 100.0 / total else 0.0
    val dataSavedKb = blocked * 120L

    Column(Modifier.fillMaxSize().padding(padding).background(MaterialTheme.colorScheme.background)) {
        TopAppBar(
            navigationIcon = {
                Row {
                    IconButton(onClick = { browse("https://blockads.pwhs.app/test") }) {
                        Icon(BlockAdsIcons.Bug, "Test blocking", tint = MaterialTheme.colorScheme.primary, modifier = Modifier.size(24.dp))
                    }
                    IconButton(onClick = { browse("https://t.me/blockads_app") }) {
                        Icon(BlockAdsIcons.Telegram, "Telegram", tint = MaterialTheme.colorScheme.primary, modifier = Modifier.size(24.dp))
                    }
                }
            },
            title = {
                if (state.busy) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        CircularProgressIndicator(Modifier.size(16.dp), strokeWidth = 2.dp)
                        Spacer(Modifier.width(8.dp))
                        Text("Working…", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                    }
                }
            },
            actions = {
                IconButton(onClick = onStatistics) { Icon(BlockAdsIcons.Chart, "Statistics", tint = MaterialTheme.colorScheme.primary, modifier = Modifier.size(24.dp)) }
                IconButton(onClick = onLogs) { Icon(BlockAdsIcons.History, "Logs", tint = MaterialTheme.colorScheme.primary, modifier = Modifier.size(24.dp)) }
            },
            colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background),
        )

        Column(
            Modifier.fillMaxSize().verticalScroll(rememberScrollState()).padding(horizontal = 24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Spacer(Modifier.height(16.dp))
            Text(
                if (running) "Protected" else if (state.status.pausedTrusted) "Paused" else "Unprotected",
                style = MaterialTheme.typography.headlineMedium,
                color = if (running) MaterialTheme.colorScheme.primary else if (state.status.pausedTrusted) SecurityOrange else DangerRed,
                fontWeight = FontWeight.Bold,
            )
            Spacer(Modifier.height(8.dp))
            Text(
                if (running) "Ads, trackers, and malicious domains are being filtered." else "Turn on protection to block ads and trackers system-wide.",
                style = MaterialTheme.typography.bodyMedium,
                color = TextSecondary,
                textAlign = TextAlign.Center,
            )
            Spacer(Modifier.height(12.dp))
            Box(
                Modifier.clip(RoundedCornerShape(12.dp)).background(MaterialTheme.colorScheme.primaryContainer).padding(horizontal = 10.dp, vertical = 4.dp)
            ) {
                Text(
                    when (state.settings.routingMode) {
                        "wireguard" -> "WireGuard Mode"
                        "root" -> "Root Proxy Mode"
                        else -> "Wintun VPN Mode"
                    },
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onPrimaryContainer,
                    fontWeight = FontWeight.SemiBold,
                )
            }
            Spacer(Modifier.height(12.dp))
            Button(
                onClick = onProfiles,
                colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.primary.copy(alpha = 0.15f)),
                shape = RoundedCornerShape(8.dp),
            ) {
                Icon(BlockAdsIcons.Shield, null, tint = MaterialTheme.colorScheme.primary, modifier = Modifier.size(16.dp))
                Spacer(Modifier.width(8.dp))
                Text(activeProfile?.name ?: "Default", color = MaterialTheme.colorScheme.primary)
                Spacer(Modifier.width(8.dp))
                Icon(BlockAdsIcons.Edit, null, tint = MaterialTheme.colorScheme.primary, modifier = Modifier.size(16.dp))
            }
            Spacer(Modifier.height(12.dp))
            PowerButton(
                isActive = running,
                isConnecting = state.busy,
                onClick = { scope.launch { state.toggleProtection() } },
            )
            Spacer(Modifier.height(36.dp))

            Row(Modifier.fillMaxWidth().height(IntrinsicSize.Min), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                StatCard(Icons.Default.QueryStats, "Total queries", formatCount(total), MaterialTheme.colorScheme.secondary, Modifier.weight(1f).fillMaxHeight())
                StatCard(Icons.Default.Block, "Blocked queries", formatCount(blocked), DangerRed, Modifier.weight(1f).fillMaxHeight())
                StatCard(Icons.Default.GppGood, "Security threats", formatCount(securityBlocked), SecurityOrange, Modifier.weight(1f).fillMaxHeight())
            }
            Spacer(Modifier.height(12.dp))
            Card(
                Modifier.fillMaxWidth(),
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
                shape = RoundedCornerShape(16.dp),
            ) {
                Row(Modifier.fillMaxWidth().padding(20.dp), verticalAlignment = Alignment.CenterVertically) {
                    Icon(Icons.Default.Shield, "Block rate", tint = MaterialTheme.colorScheme.primary, modifier = Modifier.size(32.dp))
                    Spacer(Modifier.width(16.dp))
                    Column(Modifier.weight(1f)) {
                        Text("Block rate", style = MaterialTheme.typography.bodyMedium, color = TextSecondary)
                        Text(String.format("%.1f%%", blockRate), style = MaterialTheme.typography.headlineSmall, fontWeight = FontWeight.Bold)
                    }
                    Column(horizontalAlignment = Alignment.End) {
                        Text("Filter rules", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                        Text(formatCount(enabledRules), style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold, color = MaterialTheme.colorScheme.primary)
                    }
                }
            }
            Spacer(Modifier.height(12.dp))
            Row(Modifier.fillMaxWidth().height(IntrinsicSize.Min), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                StatCard(Icons.Default.DataSaverOn, "Data saved", if (dataSavedKb < 1024) "${dataSavedKb} KB" else String.format("%.1f MB", dataSavedKb / 1024.0), MaterialTheme.colorScheme.primary, Modifier.weight(1f).fillMaxHeight())
                StatCard(Icons.Default.Timer, "DNS provider", state.settings.dnsProviderId.replaceFirstChar { it.uppercase() }, AccentBlue, Modifier.weight(1f).fillMaxHeight())
            }

            val recent = state.logs.filter { it.blocked }.take(5)
            if (recent.isNotEmpty()) {
                Spacer(Modifier.height(20.dp))
                SectionHeader("Recently blocked")
                SectionCard {
                    recent.forEachIndexed { index, log ->
                        Row(Modifier.fillMaxWidth().padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
                            Icon(Icons.Default.Block, null, tint = DangerRed, modifier = Modifier.size(20.dp))
                            Spacer(Modifier.width(12.dp))
                            Column(Modifier.weight(1f)) {
                                Text(log.domain, style = MaterialTheme.typography.titleSmall)
                                Text(log.blockedBy.ifBlank { "BlockAds filter" }, style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                            }
                            Text("${log.responseTimeMs} ms", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                        }
                        if (index != recent.lastIndex) DividerInset()
                    }
                }
            }
            Spacer(Modifier.height(32.dp))
        }
    }
}
