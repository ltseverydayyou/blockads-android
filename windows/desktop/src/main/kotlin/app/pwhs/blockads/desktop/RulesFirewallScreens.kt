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
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Block
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.DesktopWindows
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.Lock
import androidx.compose.material.icons.filled.Public
import androidx.compose.material.icons.filled.Shield
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Switch
import androidx.compose.material3.SwitchDefaults
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FirewallScreen(state: DesktopState, padding: PaddingValues) {
    val scope = rememberCoroutineScope()
    Column(Modifier.fillMaxSize().padding(padding)) {
        TopAppBar(
            title = { Text("Firewall", fontWeight = FontWeight.Bold) },
            colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background),
        )
        LazyColumn(
            Modifier.fillMaxSize().padding(horizontal = 16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
            contentPadding = PaddingValues(bottom = 24.dp),
        ) {
            item {
                Card(
                    shape = RoundedCornerShape(16.dp),
                    colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
                ) {
                    SettingsToggleItem(
                        Icons.Default.Shield,
                        "Firewall",
                        "Control network access per application.",
                        state.settings.firewallEnabled,
                    ) { enabled -> scope.launch { state.saveSettings(state.settings.copy(firewallEnabled = enabled)) } }
                }
            }
            item {
                Card(
                    shape = RoundedCornerShape(16.dp),
                    colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.secondaryContainer.copy(alpha = 0.45f)),
                ) {
                    Row(Modifier.fillMaxWidth().padding(16.dp), verticalAlignment = Alignment.Top) {
                        Icon(Icons.Default.Info, null, tint = MaterialTheme.colorScheme.secondary, modifier = Modifier.size(22.dp))
                        Spacer(Modifier.width(12.dp))
                        Column {
                            Text("Windows app filtering", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
                            Text(
                                "The Android app maps connections to package UIDs through VpnService. The equivalent Windows WFP/Wintun process attribution layer is still being wired in, so this switch is retained but per-process rules are not active yet.",
                                style = MaterialTheme.typography.bodySmall,
                                color = TextSecondary,
                            )
                        }
                    }
                }
            }
            item { SectionHeader("Network access", "Same BlockAds firewall categories, adapted for Windows processes.") }
            item {
                SectionCard {
                    FirewallPlaceholder(Icons.Default.Public, "Allow internet access", "Applications use the normal BlockAds route unless explicitly blocked.")
                    DividerInset()
                    FirewallPlaceholder(Icons.Default.Lock, "Block selected apps", "Per-app Windows process rules will appear here once WFP attribution is active.")
                    DividerInset()
                    FirewallPlaceholder(Icons.Default.DesktopWindows, "System processes", "Windows services and system processes are kept separate from user applications.")
                }
            }
        }
    }
}

@Composable
private fun FirewallPlaceholder(icon: androidx.compose.ui.graphics.vector.ImageVector, title: String, subtitle: String) {
    Row(Modifier.fillMaxWidth().padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
        Icon(icon, null, tint = MaterialTheme.colorScheme.secondary, modifier = Modifier.size(20.dp))
        Spacer(Modifier.width(12.dp))
        Column {
            Text(title, style = MaterialTheme.typography.titleSmall)
            Text(subtitle, style = MaterialTheme.typography.bodySmall, color = TextSecondary)
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DomainRulesScreen(state: DesktopState, padding: PaddingValues) {
    val scope = rememberCoroutineScope()
    var showAdd by remember { mutableStateOf(false) }
    LaunchedEffect(Unit) { state.refreshRules() }
    Column(Modifier.fillMaxSize().padding(padding)) {
        TopAppBar(
            title = { Text("Domain rules", fontWeight = FontWeight.Bold) },
            actions = {
                IconButton(onClick = { showAdd = true }) { Icon(Icons.Default.Add, "Add rule", tint = MaterialTheme.colorScheme.primary) }
            },
            colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background),
        )
        if (state.rules.isEmpty()) {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Icon(BlockAdsIcons.Crown, null, tint = TextSecondary.copy(alpha = 0.5f), modifier = Modifier.size(64.dp))
                    Spacer(Modifier.height(16.dp))
                    Text("No custom domain rules", style = MaterialTheme.typography.titleMedium)
                    Text("Add block or allow rules just like on Android.", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                    Spacer(Modifier.height(16.dp))
                    OutlinedButton(onClick = { showAdd = true }) { Icon(Icons.Default.Add, null); Spacer(Modifier.width(8.dp)); Text("Add rule") }
                }
            }
        } else {
            LazyColumn(
                Modifier.fillMaxSize().padding(horizontal = 16.dp),
                contentPadding = PaddingValues(bottom = 24.dp),
            ) {
                val blocks = state.rules.filter { it.ruleType == "BLOCK" }
                val allows = state.rules.filter { it.ruleType == "ALLOW" }
                val comments = state.rules.filter { it.ruleType == "COMMENT" }
                if (blocks.isNotEmpty()) {
                    item { RuleHeader("Blocked domains", blocks.size, DangerRed) }
                    item { RuleCard(blocks, state) }
                }
                if (allows.isNotEmpty()) {
                    item { Spacer(Modifier.height(12.dp)); RuleHeader("Allowed domains", allows.size, MaterialTheme.colorScheme.primary) }
                    item { RuleCard(allows, state) }
                }
                if (comments.isNotEmpty()) {
                    item { Spacer(Modifier.height(12.dp)); RuleHeader("Comments", comments.size, TextSecondary) }
                    item { RuleCard(comments, state) }
                }
                item {
                    Spacer(Modifier.height(16.dp))
                    OutlinedButton(onClick = { showAdd = true }, modifier = Modifier.fillMaxWidth(), shape = RoundedCornerShape(12.dp)) {
                        Icon(Icons.Default.Add, null); Spacer(Modifier.width(8.dp)); Text("Add domain rule")
                    }
                }
            }
        }
    }
    if (showAdd) AddRuleDialog(
        busy = state.busy,
        onDismiss = { showAdd = false },
        onAdd = { text -> scope.launch { state.addRule(text); if (state.message == null) showAdd = false } },
    )
}

@Composable
private fun RuleHeader(title: String, count: Int, color: androidx.compose.ui.graphics.Color) {
    Row(Modifier.fillMaxWidth().padding(start = 4.dp, top = 12.dp, bottom = 8.dp), verticalAlignment = Alignment.CenterVertically) {
        Text(title, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold, modifier = Modifier.weight(1f))
        Text("$count", style = MaterialTheme.typography.labelMedium, color = color)
    }
}

@Composable
private fun RuleCard(rules: List<Rule>, state: DesktopState) {
    val scope = rememberCoroutineScope()
    Card(shape = RoundedCornerShape(16.dp), colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)) {
        rules.forEachIndexed { index, rule ->
            Row(Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 12.dp), verticalAlignment = Alignment.CenterVertically) {
                Icon(
                    when (rule.ruleType) { "ALLOW" -> Icons.Default.CheckCircle; "BLOCK" -> Icons.Default.Block; else -> Icons.Default.Info },
                    null,
                    tint = when (rule.ruleType) { "ALLOW" -> MaterialTheme.colorScheme.primary; "BLOCK" -> DangerRed; else -> TextSecondary },
                    modifier = Modifier.size(20.dp),
                )
                Spacer(Modifier.width(12.dp))
                Column(Modifier.weight(1f)) {
                    Text(rule.rule, style = MaterialTheme.typography.titleSmall, maxLines = 1, overflow = TextOverflow.Ellipsis)
                    if (rule.domain.isNotBlank()) Text(rule.domain, style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                }
                IconButton(onClick = { scope.launch { state.deleteRule(rule) } }, modifier = Modifier.size(34.dp)) {
                    Icon(Icons.Default.Delete, "Delete", tint = TextSecondary.copy(alpha = 0.6f), modifier = Modifier.size(17.dp))
                }
                if (rule.ruleType != "COMMENT") {
                    Switch(
                        checked = rule.enabled,
                        onCheckedChange = { scope.launch { state.setRule(rule, it) } },
                        colors = SwitchDefaults.colors(checkedThumbColor = MaterialTheme.colorScheme.onPrimary, checkedTrackColor = MaterialTheme.colorScheme.primary),
                    )
                }
            }
            if (index != rules.lastIndex) HorizontalDivider(Modifier.padding(horizontal = 16.dp), color = MaterialTheme.colorScheme.outline.copy(alpha = 0.1f))
        }
    }
}

@Composable
private fun AddRuleDialog(busy: Boolean, onDismiss: () -> Unit, onAdd: (String) -> Unit) {
    var text by remember { mutableStateOf("") }
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Add domain rule") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                Text("Supported formats", style = MaterialTheme.typography.titleSmall)
                Text("Block:  ||example.com^  or  example.com\nAllow:  @@||example.com^\nWildcard:  ||*.ads.example.com^", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                OutlinedTextField(text, { text = it }, label = { Text("Rule") }, singleLine = true, modifier = Modifier.fillMaxWidth())
            }
        },
        confirmButton = { Button(onClick = { onAdd(text) }, enabled = text.isNotBlank() && !busy) { Text("Add") } },
        dismissButton = { TextButton(onClick = onDismiss) { Text("Cancel") } },
    )
}
