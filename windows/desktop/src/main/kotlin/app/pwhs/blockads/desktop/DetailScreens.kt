package app.pwhs.blockads.desktop

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Block
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.DeleteSweep
import androidx.compose.material.icons.filled.QueryStats
import androidx.compose.material.icons.filled.Security
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun LogsScreen(state: DesktopState, padding: PaddingValues, onBack: () -> Unit) {
    val scope = rememberCoroutineScope()
    Column(Modifier.fillMaxSize().padding(padding)) {
        TopAppBar(
            navigationIcon = { IconButton(onClick = onBack) { Icon(Icons.AutoMirrored.Filled.ArrowBack, "Back") } },
            title = { Text("DNS logs", fontWeight = FontWeight.Bold) },
            actions = {
                TextButton(onClick = { scope.launch { state.clearLogs() } }, enabled = state.logs.isNotEmpty()) {
                    Icon(Icons.Default.DeleteSweep, null, Modifier.size(18.dp)); Spacer(Modifier.width(6.dp)); Text("Clear")
                }
            },
            colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background),
        )
        if (state.logs.isEmpty()) {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Icon(BlockAdsIcons.History, null, tint = TextSecondary.copy(alpha = 0.5f), modifier = Modifier.size(64.dp))
                    Spacer(Modifier.height(16.dp))
                    Text("No DNS queries yet", style = MaterialTheme.typography.titleMedium)
                    Text("Queries appear here while logging is enabled.", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                }
            }
        } else {
            LazyColumn(Modifier.fillMaxSize().padding(horizontal = 16.dp), contentPadding = PaddingValues(bottom = 24.dp)) {
                item { SectionHeader("Recent activity", "${state.logs.size} locally stored queries") }
                item {
                    Card(shape = RoundedCornerShape(16.dp), colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)) {
                        state.logs.forEachIndexed { index, log ->
                            Row(Modifier.fillMaxWidth().padding(14.dp), verticalAlignment = Alignment.CenterVertically) {
                                Icon(
                                    if (log.blocked) Icons.Default.Block else Icons.Default.CheckCircle,
                                    null,
                                    tint = if (log.blocked) DangerRed else MaterialTheme.colorScheme.primary,
                                    modifier = Modifier.size(20.dp),
                                )
                                Spacer(Modifier.width(12.dp))
                                Column(Modifier.weight(1f)) {
                                    Text(log.domain, style = MaterialTheme.typography.titleSmall)
                                    Text(
                                        listOfNotNull(formatDate(log.timestamp).takeIf { it.isNotBlank() }, log.blockedBy.takeIf { it.isNotBlank() }).joinToString(" · "),
                                        style = MaterialTheme.typography.bodySmall,
                                        color = TextSecondary,
                                    )
                                }
                                Text("${log.responseTimeMs} ms", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                            }
                            if (index != state.logs.lastIndex) HorizontalDivider(Modifier.padding(horizontal = 16.dp), color = MaterialTheme.colorScheme.outline.copy(alpha = 0.1f))
                        }
                    }
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun StatisticsScreen(state: DesktopState, padding: PaddingValues, onBack: () -> Unit) {
    val total = state.status.stats.totalQueries
    val blocked = state.status.stats.blockedQueries
    val security = state.logs.count { it.blocked && it.blockedBy.contains("security", true) }.toLong()
    val rate = if (total > 0) blocked * 100.0 / total else 0.0
    Column(Modifier.fillMaxSize().padding(padding)) {
        TopAppBar(
            navigationIcon = { IconButton(onClick = onBack) { Icon(Icons.AutoMirrored.Filled.ArrowBack, "Back") } },
            title = { Text("Statistics", fontWeight = FontWeight.Bold) },
            colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background),
        )
        LazyColumn(
            Modifier.fillMaxSize().padding(horizontal = 20.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
            contentPadding = PaddingValues(bottom = 24.dp),
        ) {
            item {
                Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                    StatCard(Icons.Default.QueryStats, "Total queries", formatCount(total), MaterialTheme.colorScheme.secondary, Modifier.weight(1f))
                    StatCard(Icons.Default.Block, "Blocked", formatCount(blocked), DangerRed, Modifier.weight(1f))
                    StatCard(Icons.Default.Security, "Security", formatCount(security), SecurityOrange, Modifier.weight(1f))
                }
            }
            item {
                SectionCard {
                    Row(Modifier.fillMaxWidth().padding(20.dp), verticalAlignment = Alignment.CenterVertically) {
                        Icon(BlockAdsIcons.Shield, null, tint = MaterialTheme.colorScheme.primary, modifier = Modifier.size(32.dp))
                        Spacer(Modifier.width(16.dp))
                        Column(Modifier.weight(1f)) {
                            Text("Block rate", color = TextSecondary, style = MaterialTheme.typography.bodyMedium)
                            Text(String.format("%.1f%%", rate), style = MaterialTheme.typography.headlineSmall, fontWeight = FontWeight.Bold)
                        }
                        Column(horizontalAlignment = Alignment.End) {
                            Text("Enabled filter rules", color = TextSecondary, style = MaterialTheme.typography.bodySmall)
                            Text(formatCount(state.filters.filter { it.enabled }.sumOf { it.ruleCount.toLong() }), color = MaterialTheme.colorScheme.primary, fontWeight = FontWeight.Bold)
                        }
                    }
                }
            }
            item { SectionHeader("Top blocked domains") }
            val top = state.logs.filter { it.blocked }.groupingBy { it.domain }.eachCount().entries.sortedByDescending { it.value }.take(10)
            if (top.isEmpty()) item { Text("No blocking data yet.", color = TextSecondary, modifier = Modifier.padding(16.dp)) }
            else item {
                SectionCard {
                    top.forEachIndexed { i, e ->
                        Row(Modifier.fillMaxWidth().padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
                            Text("${i + 1}", color = TextSecondary, modifier = Modifier.width(28.dp))
                            Text(e.key, modifier = Modifier.weight(1f), style = MaterialTheme.typography.titleSmall)
                            Text(formatCount(e.value), color = DangerRed, fontWeight = FontWeight.SemiBold)
                        }
                        if (i != top.lastIndex) DividerInset()
                    }
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ProfilesScreen(state: DesktopState, padding: PaddingValues, onBack: () -> Unit) {
    val scope = rememberCoroutineScope()
    Column(Modifier.fillMaxSize().padding(padding)) {
        TopAppBar(
            navigationIcon = { IconButton(onClick = onBack) { Icon(Icons.AutoMirrored.Filled.ArrowBack, "Back") } },
            title = { Text("Protection profiles", fontWeight = FontWeight.Bold) },
            colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background),
        )
        LazyColumn(Modifier.fillMaxSize().padding(horizontal = 16.dp), contentPadding = PaddingValues(bottom = 24.dp)) {
            item { SectionHeader("Profiles", "Switch filter sets, SafeSearch, and YouTube Restricted Mode together.") }
            items(state.profiles) { profile ->
                val active = state.settings.activeProfile == profile.id
                Card(
                    modifier = Modifier.fillMaxWidth().padding(bottom = 10.dp),
                    shape = RoundedCornerShape(16.dp),
                    colors = CardDefaults.cardColors(
                        containerColor = if (active) MaterialTheme.colorScheme.primary.copy(alpha = 0.12f) else MaterialTheme.colorScheme.surface,
                    ),
                    onClick = { scope.launch { state.activateProfile(profile) } },
                ) {
                    Row(Modifier.fillMaxWidth().padding(18.dp), verticalAlignment = Alignment.CenterVertically) {
                        Icon(BlockAdsIcons.Shield, null, tint = if (active) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.secondary, modifier = Modifier.size(28.dp))
                        Spacer(Modifier.width(14.dp))
                        Column(Modifier.weight(1f)) {
                            Text(profile.name, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
                            Text(
                                "${profile.enabledFilterUrls.size} filter sources" + if (profile.safeSearchEnabled || profile.youtubeRestrictedMode) " · family protections" else "",
                                style = MaterialTheme.typography.bodySmall,
                                color = TextSecondary,
                            )
                        }
                        if (active) Text("Active", color = MaterialTheme.colorScheme.primary, fontWeight = FontWeight.SemiBold)
                    }
                }
            }
        }
    }
}
