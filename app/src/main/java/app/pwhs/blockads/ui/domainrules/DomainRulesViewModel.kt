package app.pwhs.blockads.ui.domainrules

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import app.pwhs.blockads.R
import app.pwhs.blockads.data.dao.CustomDnsRuleDao
import app.pwhs.blockads.data.dao.WhitelistDomainDao
import app.pwhs.blockads.data.entities.CustomDnsRule
import app.pwhs.blockads.data.entities.RuleType
import app.pwhs.blockads.data.entities.WhitelistDomain
import app.pwhs.blockads.service.AdBlockVpnService
import app.pwhs.blockads.service.ServiceController
import app.pwhs.blockads.ui.event.UiEvent
import app.pwhs.blockads.ui.event.toast
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

class DomainRulesViewModel(
    private val whitelistDomainDao: WhitelistDomainDao,
    private val customDnsRuleDao: CustomDnsRuleDao,
    application: Application
) : AndroidViewModel(application) {

    val whitelistDomains: StateFlow<List<WhitelistDomain>> = whitelistDomainDao.getAll()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    val blocklistDomains: StateFlow<List<CustomDnsRule>> = customDnsRuleDao.getAllFlow()
        .map { rules -> rules.filter { it.ruleType == RuleType.BLOCK } }
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), emptyList())

    private val _events = MutableSharedFlow<UiEvent>(extraBufferCapacity = 1)
    val events: SharedFlow<UiEvent> = _events.asSharedFlow()

    // ── Whitelist ────────────────────────────────────────────

    fun addWhitelistDomain(domain: String) {
        viewModelScope.launch {
            val cleanDomain = domain.trim().lowercase()
            if (cleanDomain.isNotBlank()) {
                val exists = whitelistDomainDao.exists(cleanDomain)
                if (exists == 0) {
                    whitelistDomainDao.insert(WhitelistDomain(domain = cleanDomain))
                    _events.toast(R.string.whitelist_domain_added, listOf(cleanDomain))
                    requestVpnRestart()
                } else {
                    _events.toast(R.string.filter_domain_already_whitelisted)
                }
            }
        }
    }

    fun updateWhitelistDomain(oldDomain: WhitelistDomain, newDomain: String) {
        viewModelScope.launch {
            val clean = newDomain.trim().lowercase()
            if (clean.isNotBlank() && clean != oldDomain.domain.lowercase()) {
                val exists = whitelistDomainDao.exists(clean)
                if (exists == 0) {
                    whitelistDomainDao.update(oldDomain.copy(domain = clean))
                    _events.toast(R.string.whitelist_domain_added, listOf(clean))
                    requestVpnRestart()
                } else {
                    _events.toast(R.string.filter_domain_already_whitelisted)
                }
            }
        }
    }

    fun removeWhitelistDomain(domain: WhitelistDomain) {
        viewModelScope.launch {
            whitelistDomainDao.delete(domain)
            _events.toast(R.string.whitelist_domain_removed)
            requestVpnRestart()
        }
    }

    // ── Blocklist ────────────────────────────────────────────

    fun addBlocklistDomain(domain: String) {
        viewModelScope.launch {
            val cleanDomain = domain.trim().lowercase()
            if (cleanDomain.isNotBlank()) {
                val allRules = customDnsRuleDao.getAll()
                val exists = allRules.any {
                    it.ruleType == RuleType.BLOCK && it.domain.equals(cleanDomain, ignoreCase = true)
                }
                if (!exists) {
                    customDnsRuleDao.insert(
                        CustomDnsRule(
                            rule = "||$cleanDomain^",
                            ruleType = RuleType.BLOCK,
                            domain = cleanDomain
                        )
                    )
                    _events.toast(R.string.blocklist_domain_added, listOf(cleanDomain))
                    requestVpnRestart()
                } else {
                    _events.toast(R.string.blocklist_domain_already_exists)
                }
            }
        }
    }

    fun updateBlocklistDomain(oldRule: CustomDnsRule, newDomain: String) {
        viewModelScope.launch {
            val clean = newDomain.trim().lowercase()
            if (clean.isNotBlank() && clean != oldRule.domain.lowercase()) {
                val allRules = customDnsRuleDao.getAll()
                val exists = allRules.any {
                    it.ruleType == RuleType.BLOCK && it.domain.equals(clean, ignoreCase = true) && it.id != oldRule.id
                }
                if (!exists) {
                    customDnsRuleDao.update(
                        oldRule.copy(
                            domain = clean,
                            rule = "||$clean^"
                        )
                    )
                    _events.toast(R.string.blocklist_domain_added, listOf(clean))
                    requestVpnRestart()
                } else {
                    _events.toast(R.string.blocklist_domain_already_exists)
                }
            }
        }
    }

    fun removeBlocklistDomain(rule: CustomDnsRule) {
        viewModelScope.launch {
            customDnsRuleDao.delete(rule)
            _events.toast(R.string.blocklist_domain_removed)
            requestVpnRestart()
        }
    }

    // ── Bulk Import ──────────────────────────────────────────

    fun importDomains(domains: List<String>, isAllow: Boolean) {
        if (domains.isEmpty()) return
        viewModelScope.launch {
            if (isAllow) {
                val existing = whitelistDomainDao.getAllDomains().map { it.lowercase() }.toSet()
                val newDomains = domains.map { it.trim().lowercase() }
                    .filter { it.isNotBlank() && !existing.contains(it) }
                    .distinct()
                if (newDomains.isNotEmpty()) {
                    whitelistDomainDao.insertAll(newDomains.map { WhitelistDomain(domain = it) })
                    _events.toast(R.string.wireguard_imported, listOf("${newDomains.size} domains"))
                    requestVpnRestart()
                }
            } else {
                val existing = customDnsRuleDao.getAll()
                    .filter { it.ruleType == RuleType.BLOCK }
                    .map { it.domain.lowercase() }
                    .toSet()
                val newDomains = domains.map { it.trim().lowercase() }
                    .filter { it.isNotBlank() && !existing.contains(it) }
                    .distinct()
                if (newDomains.isNotEmpty()) {
                    customDnsRuleDao.insertAll(
                        newDomains.map {
                            CustomDnsRule(
                                rule = "||$it^",
                                ruleType = RuleType.BLOCK,
                                domain = it
                            )
                        }
                    )
                    _events.toast(R.string.wireguard_imported, listOf("${newDomains.size} domains"))
                    requestVpnRestart()
                }
            }
        }
    }

    private fun requestVpnRestart() {
        ServiceController.requestRestart(getApplication<Application>().applicationContext)
    }
}
