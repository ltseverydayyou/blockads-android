package app.pwhs.blockads.desktop

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
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
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.CloudDownload
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.FilterList
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Shield
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
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
import androidx.compose.material3.TextField
import androidx.compose.material3.TextFieldDefaults
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
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FiltersScreen(state: DesktopState, padding: PaddingValues) {
    val scope = rememberCoroutineScope()
    var searchVisible by remember { mutableStateOf(false) }
    var search by remember { mutableStateOf("") }
    var showAdd by remember { mutableStateOf(false) }

    LaunchedEffect(Unit) { state.refreshFilters() }

    Column(Modifier.fillMaxSize().padding(padding)) {
        TopAppBar(
            title = { Text("Filter setup", fontWeight = FontWeight.Bold) },
            actions = {
                IconButton(onClick = { searchVisible = !searchVisible; if (!searchVisible) search = "" }) {
                    Icon(if (searchVisible) Icons.Default.Close else Icons.Default.Search, "Search", tint = TextSecondary)
                }
                TextButton(onClick = { scope.launch { state.updateFilters() } }, enabled = !state.busy) {
                    if (state.busy) CircularProgressIndicator(Modifier.size(18.dp), strokeWidth = 2.dp)
                    else Icon(Icons.Default.CloudDownload, null, Modifier.size(18.dp))
                    Spacer(Modifier.width(8.dp))
                    Text(if (state.busy) "Updating" else "Update all")
                }
            },
            colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background),
        )
        AnimatedVisibility(searchVisible, enter = fadeIn(), exit = fadeOut()) {
            TextField(
                value = search,
                onValueChange = { search = it },
                placeholder = { Text("Search filter lists", color = TextSecondary) },
                leadingIcon = { Icon(Icons.Default.Search, null, tint = TextSecondary) },
                trailingIcon = if (search.isNotEmpty()) {{ IconButton(onClick = { search = "" }) { Icon(Icons.Default.Close, "Clear") } }} else null,
                modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 4.dp),
                shape = RoundedCornerShape(16.dp),
                colors = TextFieldDefaults.colors(
                    focusedContainerColor = MaterialTheme.colorScheme.surfaceVariant,
                    unfocusedContainerColor = MaterialTheme.colorScheme.surfaceVariant,
                    focusedIndicatorColor = Color.Transparent,
                    unfocusedIndicatorColor = Color.Transparent,
                ),
                singleLine = true,
            )
        }

        val shown = state.filters.filter { search.isBlank() || it.name.contains(search, true) || it.description.contains(search, true) || it.url.contains(search, true) }
        val ad = shown.filter { it.builtIn && it.category != "SECURITY" }
        val security = shown.filter { it.builtIn && it.category == "SECURITY" }
        val custom = shown.filter { !it.builtIn }

        if (search.isNotBlank() && shown.isEmpty()) {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Icon(Icons.Default.FilterList, null, Modifier.size(64.dp), tint = TextSecondary.copy(alpha = 0.5f))
                    Spacer(Modifier.height(16.dp))
                    Text("No filters found for “$search”", color = TextSecondary)
                }
            }
        } else {
            LazyColumn(
                Modifier.fillMaxSize().padding(horizontal = 16.dp),
                verticalArrangement = Arrangement.spacedBy(0.dp),
                contentPadding = PaddingValues(bottom = 24.dp),
            ) {
                if (ad.isNotEmpty()) {
                    item { FilterSectionHeader("Ad & tracking filters", ad.count { it.enabled }) }
                    item { FilterCard(ad, state) }
                }
                if (security.isNotEmpty()) {
                    item { Spacer(Modifier.height(12.dp)); FilterSectionHeader("Security filters", security.count { it.enabled }) }
                    item { FilterCard(security, state) }
                }
                item { Spacer(Modifier.height(12.dp)); FilterSectionHeader("Custom filters", custom.count { it.enabled }) }
                if (custom.isNotEmpty()) item { FilterCard(custom, state) }
                item {
                    Spacer(Modifier.height(12.dp))
                    OutlinedButton(onClick = { showAdd = true }, modifier = Modifier.fillMaxWidth(), shape = RoundedCornerShape(12.dp)) {
                        Icon(Icons.Default.Add, null, Modifier.size(18.dp)); Spacer(Modifier.width(8.dp)); Text("Add filter list")
                    }
                }
            }
        }
    }

    if (showAdd) AddFilterDialog(
        busy = state.busy,
        onDismiss = { showAdd = false },
        onAdd = { name, url -> scope.launch { state.addFilter(name, url); if (state.message == null) showAdd = false } },
    )
}

@Composable
private fun FilterSectionHeader(title: String, activeCount: Int) {
    Row(Modifier.fillMaxWidth().padding(start = 4.dp, top = 12.dp, bottom = 8.dp), verticalAlignment = Alignment.CenterVertically) {
        Text(title, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold, modifier = Modifier.weight(1f))
        Text("$activeCount active", style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.primary)
    }
}

@Composable
private fun FilterCard(filters: List<FilterList>, state: DesktopState) {
    val scope = rememberCoroutineScope()
    Card(
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        shape = RoundedCornerShape(16.dp),
    ) {
        filters.forEachIndexed { index, filter ->
            Row(Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 12.dp), verticalAlignment = Alignment.CenterVertically) {
                Column(Modifier.weight(1f)) {
                    Text(
                        filter.name,
                        style = MaterialTheme.typography.titleSmall,
                        fontWeight = FontWeight.Medium,
                        color = if (filter.enabled) MaterialTheme.colorScheme.onBackground else TextSecondary,
                    )
                    if (filter.builtIn) {
                        Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(3.dp)) {
                            Icon(Icons.Default.Shield, null, Modifier.size(10.dp), tint = MaterialTheme.colorScheme.primary)
                            Text("Built-in", style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.primary, fontWeight = FontWeight.SemiBold)
                        }
                    }
                    if (filter.description.isNotBlank()) Text(filter.description, style = MaterialTheme.typography.bodySmall, color = TextSecondary, maxLines = 2, overflow = TextOverflow.Ellipsis)
                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        if (filter.ruleCount > 0) Text("${formatCount(filter.ruleCount)} rules", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                        if (filter.lastUpdated > 0) Text(formatDate(filter.lastUpdated), style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                    }
                    Text(filter.url, style = MaterialTheme.typography.labelSmall, color = TextSecondary.copy(alpha = 0.6f), maxLines = 1, overflow = TextOverflow.Ellipsis)
                }
                if (!filter.builtIn) {
                    IconButton(onClick = { scope.launch { state.removeFilter(filter) } }, modifier = Modifier.size(32.dp)) {
                        Icon(Icons.Default.Delete, "Delete", tint = TextSecondary.copy(alpha = 0.6f), modifier = Modifier.size(16.dp))
                    }
                }
                Switch(
                    checked = filter.enabled,
                    onCheckedChange = { scope.launch { state.setFilter(filter, it) } },
                    colors = SwitchDefaults.colors(checkedThumbColor = MaterialTheme.colorScheme.onPrimary, checkedTrackColor = MaterialTheme.colorScheme.primary),
                )
            }
            if (index != filters.lastIndex) HorizontalDivider(Modifier.padding(horizontal = 16.dp), color = MaterialTheme.colorScheme.outline.copy(alpha = 0.1f))
        }
    }
}

@Composable
private fun AddFilterDialog(busy: Boolean, onDismiss: () -> Unit, onAdd: (String, String) -> Unit) {
    var name by remember { mutableStateOf("") }
    var url by remember { mutableStateOf("") }
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Add filter list") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                Text("BlockAds will compile the same filter format used by the Android app.", color = TextSecondary, style = MaterialTheme.typography.bodySmall)
                OutlinedTextField(name, { name = it }, label = { Text("Name") }, singleLine = true, modifier = Modifier.fillMaxWidth())
                OutlinedTextField(url, { url = it }, label = { Text("Filter URL") }, singleLine = true, modifier = Modifier.fillMaxWidth())
            }
        },
        confirmButton = { Button(onClick = { onAdd(name, url) }, enabled = name.isNotBlank() && url.isNotBlank() && !busy) { Text("Add") } },
        dismissButton = { TextButton(onClick = onDismiss) { Text("Cancel") } },
    )
}
