package app.pwhs.blockads.desktop

import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock

class DesktopState {
    var status by mutableStateOf(Status())
    var settings by mutableStateOf(Settings())
    var filters by mutableStateOf<List<FilterList>>(emptyList())
    var rules by mutableStateOf<List<Rule>>(emptyList())
    var profiles by mutableStateOf<List<Profile>>(emptyList())
    var logs by mutableStateOf<List<LogEntry>>(emptyList())
    var loading by mutableStateOf(true)
    var busy by mutableStateOf(false)
    var message by mutableStateOf<String?>(null)

    private val settingsMutex = Mutex()

    suspend fun initialize() {
        loading = true
        runCatching {
            BackendClient.ensureRunning()
            refreshAll()
        }.onFailure { message = it.message ?: it.toString() }
        loading = false
    }

    suspend fun refreshAll() {
        status = BackendClient.status()
        settings = BackendClient.settings()
        filters = BackendClient.filters()
        rules = BackendClient.rules()
        profiles = BackendClient.profiles()
        logs = BackendClient.logs().sortedByDescending { it.timestamp }
    }

    suspend fun refreshStatus() { status = BackendClient.status() }
    suspend fun refreshFilters() { filters = BackendClient.filters() }
    suspend fun refreshRules() { rules = BackendClient.rules() }
    suspend fun refreshLogs() { logs = BackendClient.logs().sortedByDescending { it.timestamp } }

    suspend fun toggleProtection() = action {
        status = if (status.running) BackendClient.stopProtection() else BackendClient.startProtection()
        settings = BackendClient.settings()
    }

    suspend fun updateFilters() = action {
        BackendClient.updateFilters()
        filters = BackendClient.filters()
        status = BackendClient.status()
    }

    suspend fun setFilter(filter: FilterList, enabled: Boolean) = action {
        filters = BackendClient.setFilter(filter.id, enabled)
        settings = BackendClient.settings()
    }

    suspend fun addFilter(name: String, url: String) = action {
        BackendClient.addCustomFilter(name, url)
        filters = BackendClient.filters()
    }

    suspend fun removeFilter(filter: FilterList) = action {
        filters = BackendClient.removeCustomFilter(filter.id)
    }

    suspend fun activateProfile(profile: Profile) = action {
        BackendClient.activateProfile(profile.id)
        settings = BackendClient.settings()
        filters = BackendClient.filters()
    }

    suspend fun addRule(text: String) = action {
        BackendClient.addRule(text)
        rules = BackendClient.rules()
        status = BackendClient.status()
    }

    suspend fun setRule(rule: Rule, enabled: Boolean) = action {
        rules = BackendClient.setRule(rule.id, enabled)
    }

    suspend fun deleteRule(rule: Rule) = action {
        rules = BackendClient.deleteRule(rule.id)
        status = BackendClient.status()
    }

    suspend fun clearLogs() = action { logs = BackendClient.clearLogs() }

    suspend fun updateSettings(transform: (Settings) -> Settings) {
        settingsMutex.withLock {
            val previous = settings
            val next = transform(previous)
            if (next == previous) return

            settings = next
            busy = true
            message = null
            try {
                settings = BackendClient.saveSettings(next)
                status = BackendClient.status()
            } catch (t: Throwable) {
                settings = previous
                message = t.message ?: t.toString()
            } finally {
                busy = false
            }
        }
    }

    private suspend fun action(block: suspend () -> Unit) {
        busy = true
        message = null
        runCatching { block() }.onFailure { message = it.message ?: it.toString() }
        busy = false
    }
}
