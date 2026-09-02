package app.pwhs.blockads.desktop

import androidx.compose.foundation.layout.size
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.NavigationBarItemDefaults
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.delay

enum class RootTab(val label: String, val icon: ImageVector) {
    Home("Home", BlockAdsIcons.Home),
    Filters("Filters", BlockAdsIcons.Shield),
    Firewall("Firewall", BlockAdsIcons.Fire),
    DomainRules("Rules", BlockAdsIcons.Crown),
    Settings("Settings", BlockAdsIcons.Settings),
}

enum class DetailPage { Logs, Statistics, Profiles }

@Composable
fun BlockAdsDesktopApp(state: DesktopState) {
    var currentTab by remember { mutableStateOf(RootTab.Home) }
    var detail by remember { mutableStateOf<DetailPage?>(null) }
    val snackbar = remember { SnackbarHostState() }

    LaunchedEffect(Unit) { state.initialize() }
    LaunchedEffect(Unit) {
        while (true) {
            delay(1500)
            if (!state.loading && !state.busy) runCatching {
                state.refreshStatus()
                if (currentTab == RootTab.Home || detail == DetailPage.Logs || detail == DetailPage.Statistics) state.refreshLogs()
            }
        }
    }
    LaunchedEffect(state.message) {
        state.message?.let { snackbar.showSnackbar(it); state.message = null }
    }

    if (state.loading) {
        androidx.compose.foundation.layout.Box(
            modifier = Modifier,
            contentAlignment = androidx.compose.ui.Alignment.Center,
        ) { CircularProgressIndicator(color = MaterialTheme.colorScheme.primary) }
        return
    }

    Scaffold(
        snackbarHost = { SnackbarHost(snackbar) },
        bottomBar = {
            if (detail == null) {
                NavigationBar(containerColor = MaterialTheme.colorScheme.surface) {
                    RootTab.entries.forEach { tab ->
                        NavigationBarItem(
                            selected = currentTab == tab,
                            onClick = { currentTab = tab },
                            icon = { Icon(tab.icon, tab.label, Modifier.size(24.dp)) },
                            label = if (state.settings.showNavigationLabels) { { Text(tab.label) } } else null,
                            alwaysShowLabel = state.settings.showNavigationLabels,
                            colors = NavigationBarItemDefaults.colors(
                                selectedIconColor = MaterialTheme.colorScheme.primary,
                                selectedTextColor = MaterialTheme.colorScheme.primary,
                                indicatorColor = MaterialTheme.colorScheme.primary.copy(alpha = 0.12f),
                            ),
                        )
                    }
                }
            }
        },
    ) { padding ->
        when (detail) {
            DetailPage.Logs -> LogsScreen(state, padding, onBack = { detail = null })
            DetailPage.Statistics -> StatisticsScreen(state, padding, onBack = { detail = null })
            DetailPage.Profiles -> ProfilesScreen(state, padding, onBack = { detail = null })
            null -> when (currentTab) {
                RootTab.Home -> HomeScreen(
                    state = state,
                    padding = padding,
                    onLogs = { detail = DetailPage.Logs },
                    onStatistics = { detail = DetailPage.Statistics },
                    onProfiles = { detail = DetailPage.Profiles },
                )
                RootTab.Filters -> FiltersScreen(state, padding)
                RootTab.Firewall -> FirewallScreen(state, padding)
                RootTab.DomainRules -> DomainRulesScreen(state, padding)
                RootTab.Settings -> SettingsScreen(state, padding, onOpenFilters = { currentTab = RootTab.Filters })
            }
        }
    }
}
