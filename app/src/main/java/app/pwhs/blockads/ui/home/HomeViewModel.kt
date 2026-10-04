package app.pwhs.blockads.ui.home

import android.content.Context
import android.content.Intent
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import app.pwhs.blockads.data.dao.DnsLogDao
import app.pwhs.blockads.data.dao.FilterListDao
import app.pwhs.blockads.data.dao.ProtectionProfileDao
import app.pwhs.blockads.data.entities.DailyStat
import app.pwhs.blockads.data.entities.DnsLogEntry
import app.pwhs.blockads.data.entities.FilterList
import app.pwhs.blockads.data.entities.HourlyStat
import app.pwhs.blockads.data.entities.ProtectionProfile
import app.pwhs.blockads.data.entities.TopBlockedDomain
import app.pwhs.blockads.data.repository.FilterListRepository
import app.pwhs.blockads.service.AdBlockVpnService
import app.pwhs.blockads.service.VpnState
import app.pwhs.blockads.service.RootProxyService
import app.pwhs.blockads.data.datastore.AppPreferences
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import app.pwhs.blockads.service.NotificationHelper
import app.pwhs.blockads.R
import app.pwhs.blockads.data.dao.CustomDnsRuleDao
import app.pwhs.blockads.data.dao.WhitelistDomainDao
import app.pwhs.blockads.data.entities.WhitelistDomain
import app.pwhs.blockads.utils.CustomRuleParser
import app.pwhs.blockads.ui.event.UiEvent
import app.pwhs.blockads.ui.event.toast
import app.pwhs.blockads.ui.home.data.RecentLogFilter
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import timber.log.Timber

class HomeViewModel(
    private val appPrefs: AppPreferences,
    dnsLogDao: DnsLogDao,
    private val filterRepo: FilterListRepository,
    profileDao: ProtectionProfileDao,
    filterListDao: FilterListDao,
    private val whitelistDomainDao: WhitelistDomainDao,
    private val customDnsRuleDao: CustomDnsRuleDao,
) : ViewModel() {

    val whitelistedDomains: StateFlow<Set<String>> = whitelistDomainDao.getAll()
        .map { list -> list.map { it.domain.lowercase() }.toSet() }
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptySet())

    private val _events = MutableSharedFlow<UiEvent>()
    val events: SharedFlow<UiEvent> = _events.asSharedFlow()

    val routingMode: StateFlow<String> = appPrefs.routingMode
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), "local")

    // Trusted networks (#197): true when BlockAds was auto-paused on a
    // trusted Wi-Fi, so Home can show a distinct state instead of "Unprotected".
    val pausedByTrusted: StateFlow<Boolean> = appPrefs.pausedByTrusted
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), false)

    val pausedTrustedSsid: StateFlow<String> = appPrefs.pausedTrustedSsid
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), "")

    // ── Reactive VPN state (derived from the single source of truth) ──
    val vpnEnabled: StateFlow<Boolean> = combine(
        AdBlockVpnService.state,
        RootProxyService.state
    ) { state1, state2 ->
        state1 == VpnState.RUNNING || state1 == VpnState.STOPPING ||
        state2 == VpnState.RUNNING || state2 == VpnState.STOPPING
    }.stateIn(
        viewModelScope,
        SharingStarted.WhileSubscribed(5000),
        AdBlockVpnService.isRunning || AdBlockVpnService.isStopping || RootProxyService.isRunning
    )

    val vpnConnecting: StateFlow<Boolean> = combine(
        AdBlockVpnService.state,
        RootProxyService.state
    ) { state1, state2 ->
        state1 == VpnState.STARTING || state1 == VpnState.RESTARTING ||
        state2 == VpnState.STARTING || state2 == VpnState.RESTARTING
    }.stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), AdBlockVpnService.isConnecting)

    val vpnStopping: StateFlow<Boolean> = combine(
        AdBlockVpnService.state,
        RootProxyService.state
    ) { state1, state2 ->
        state1 == VpnState.STOPPING || state2 == VpnState.STOPPING
    }.stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), false)

    val blockedCount: StateFlow<Int> = dnsLogDao.getBlockedCount()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), 0)

    val totalCount: StateFlow<Int> = dnsLogDao.getTotalCount()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), 0)

    private val _recentFilter = MutableStateFlow(RecentLogFilter.BLOCKED)
    val recentFilter: StateFlow<RecentLogFilter> = _recentFilter.asStateFlow()

    val recentLogs: StateFlow<List<DnsLogEntry>> = combine(
        _recentFilter,
        dnsLogDao.getRecentBlocked(5),
        dnsLogDao.getRecentLogs(5)
    ) { filter, blocked, all ->
        if (filter == RecentLogFilter.BLOCKED) blocked else all
    }.stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    val recentBlocked: StateFlow<List<DnsLogEntry>> = recentLogs

    fun setRecentFilter(filter: RecentLogFilter) {
        _recentFilter.value = filter
    }

    val hourlyStats: StateFlow<List<HourlyStat>> = dnsLogDao.getHourlyStats()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    val dailyStats: StateFlow<List<DailyStat>> = dnsLogDao.getDailyStats()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    val topBlockedDomains: StateFlow<List<TopBlockedDomain>> = dnsLogDao.getTopBlockedDomains()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    val activeProfile: StateFlow<ProtectionProfile?> = profileDao.getActiveFlow()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), null)

    val milestoneReached: StateFlow<Long?> = combine(
        blockedCount,
        appPrefs.lastSeenMilestoneDialog,
        appPrefs.milestoneNotificationsEnabled
    ) { blocked, lastSeen, enabled ->
        NotificationHelper.unseenMilestone(blocked.toLong(), lastSeen, enabled)
    }.stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), null)

    fun dismissMilestoneDialog(milestone: Long) {
        viewModelScope.launch {
            appPrefs.setLastSeenMilestoneDialog(milestone)
        }
    }

    val securityFilterIds: StateFlow<Set<String>> = filterListDao.getAll()
        .map { list -> list.filter { it.category == FilterList.CATEGORY_SECURITY }.map { it.id.toString() }.toSet() }
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptySet())

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    private val _filterLoadFailed = MutableStateFlow(false)
    val filterLoadFailed: StateFlow<Boolean> = _filterLoadFailed.asStateFlow()

    private val _protectionUptimeMs = MutableStateFlow(0L)
    val protectionUptimeMs: StateFlow<Long> = _protectionUptimeMs.asStateFlow()

    val domainCount: StateFlow<Int> = filterRepo.domainCountFlow
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), filterRepo.domainCount)

    // Warn when protection is on in VPN mode but Android Private DNS (Strict
    // DoT) is active — it bypasses BlockAds' DNS interception, so filtering
    // silently doesn't apply. Root Proxy mode disables Private DNS itself, so
    // the warning is VPN-mode only (#145).
    val privateDnsWarning: StateFlow<Boolean> = combine(
        vpnEnabled,
        routingMode,
        AdBlockVpnService.privateDnsStrict
    ) { enabled, mode, strict ->
        enabled && mode != AppPreferences.ROUTING_MODE_ROOT && strict
    }.stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), false)

    init {
        // Uptime ticker — only ticks while VPN or Root Proxy is RUNNING
        viewModelScope.launch {
            while (isActive) {
                var uptime = 0L
                if (AdBlockVpnService.isRunning && AdBlockVpnService.startTimestamp > 0) {
                    uptime = System.currentTimeMillis() - AdBlockVpnService.startTimestamp
                } else if (RootProxyService.isRunning && RootProxyService.startTimestamp > 0) {
                    uptime = System.currentTimeMillis() - RootProxyService.startTimestamp
                }
                _protectionUptimeMs.value = uptime
                delay(1000)
            }
        }
    }

    fun stopVpn(context: Context) {
        if (RootProxyService.isRunning) {
            RootProxyService.stop(context)
        }
        if (AdBlockVpnService.isRunning) {
            val intent = Intent(context, AdBlockVpnService::class.java).apply {
                action = AdBlockVpnService.ACTION_STOP
            }
            context.startService(intent)
        }
    }

    fun preloadFilter() {
        if (_isLoading.value || filterRepo.domainCount > 0) return // Already loaded or loading
        viewModelScope.launch {
            _isLoading.value = true
            _filterLoadFailed.value = false
            try {
                filterRepo.seedDefaultsIfNeeded()
                filterRepo.loadAllEnabledFilters()
                _filterLoadFailed.value = false
            } catch (e: Exception) {
                Timber.e(e)
                _filterLoadFailed.value = true
            } finally {
                _isLoading.value = false
            }
        }
    }

    fun retryLoadFilter() {
        viewModelScope.launch {
            _isLoading.value = true
            _filterLoadFailed.value = false
            try {
                filterRepo.seedDefaultsIfNeeded()
                filterRepo.loadAllEnabledFilters()
                _filterLoadFailed.value = false
            } catch (e: Exception) {
                _filterLoadFailed.value = true
            } finally {
                _isLoading.value = false
            }
        }
    }

    fun addToWhitelist(domain: String) {
        viewModelScope.launch {
            val cleanDomain = domain.trim().lowercase()
            val exists = whitelistDomainDao.exists(cleanDomain)
            if (exists == 0) {
                whitelistDomainDao.insert(WhitelistDomain(domain = cleanDomain))
                filterRepo.loadWhitelist()
                _events.toast(R.string.log_whitelisted, listOf(": $cleanDomain"))
            } else {
                _events.toast(R.string.log_already_whitelisted)
            }
        }
    }

    fun removeFromWhitelist(domain: String) {
        viewModelScope.launch {
            val cleanDomain = domain.trim().lowercase()
            whitelistDomainDao.deleteByDomain(cleanDomain)
            filterRepo.loadWhitelist()
            _events.toast(R.string.whitelist_domain_removed)
        }
    }

    fun addWildcardWhitelist(domain: String) {
        viewModelScope.launch {
            val cleanDomain = domain.trim().lowercase()
            val domainRuleText = "@@||$cleanDomain^"
            val wildcardRuleText = "@@||*.$cleanDomain^"

            var addedAny = false
            if (customDnsRuleDao.exists(domainRuleText) == 0) {
                val domainRule = CustomRuleParser.parseRule(domainRuleText)
                if (domainRule != null) {
                    customDnsRuleDao.insert(domainRule)
                    addedAny = true
                }
            }
            if (customDnsRuleDao.exists(wildcardRuleText) == 0) {
                val wildcardRule = CustomRuleParser.parseRule(wildcardRuleText)
                if (wildcardRule != null) {
                    customDnsRuleDao.insert(wildcardRule)
                    addedAny = true
                }
            }
            if (addedAny) {
                filterRepo.loadCustomRules()
            }
            _events.toast(R.string.log_wildcard_whitelisted, listOf(cleanDomain))
        }
    }

    fun addToCustomBlockRules(domain: String) {
        viewModelScope.launch {
            val cleanDomain = domain.trim().lowercase()
            val ruleText = CustomRuleParser.formatBlockRule(cleanDomain)
            val rule = CustomRuleParser.parseRule(ruleText)
            if (rule != null) {
                if (customDnsRuleDao.exists(rule.rule) == 0) {
                    customDnsRuleDao.insert(rule)
                    filterRepo.loadCustomRules()
                }
                _events.toast(R.string.rule_added)
            }
        }
    }
}
