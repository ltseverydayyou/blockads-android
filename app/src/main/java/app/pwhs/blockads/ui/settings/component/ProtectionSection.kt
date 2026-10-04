package app.pwhs.blockads.ui.settings.component

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.expandVertically
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.shrinkVertically
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.material3.FilterChip
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import app.pwhs.blockads.R
import app.pwhs.blockads.data.datastore.AppPreferences

@Composable
fun ProtectionSection(
    autoReconnect: Boolean,
    routingMode: String,
    safeSearchEnabled: Boolean,
    youtubeRestrictedMode: Boolean,
    dnsResponseType: String,
    upstreamDNS: String,
    onSetAutoReconnect: (Boolean) -> Unit,
    onSetRoutingMode: (Boolean) -> Unit,
    onSetSafeSearchEnabled: (Boolean) -> Unit,
    onSetYoutubeRestrictedMode: (Boolean) -> Unit,
    onShowDnsResponseTypeDialog: () -> Unit,
    onNavigateToDNSProvider: () -> Unit,
    onNavigateToWireGuardImport: () -> Unit,
    onNavigateToHttpsFiltering: () -> Unit,
    modifier: Modifier = Modifier
) {
    val dividerColor = MaterialTheme.colorScheme.outline.copy(alpha = 0.08f)

    Column(modifier = modifier) {
        SectionHeader(
            title = stringResource(R.string.settings_category_protection),
            description = stringResource(R.string.settings_category_protection_desc)
        )
        Spacer(modifier = Modifier.height(10.dp))

        SettingsCard {
            Column {
                // 1. DNS Provider
                SettingItem(
                    iconPainter = painterResource(R.drawable.ic_settings_dns),
                    iconTint = Color(0xFF2563EB),
                    title = stringResource(R.string.dns_provider_title),
                    statusValue = upstreamDNS,
                    onClick = onNavigateToDNSProvider
                )

                HorizontalDivider(modifier = Modifier.padding(horizontal = 16.dp), color = dividerColor)

                // 2. DNS Response Type
                SettingItem(
                    iconPainter = painterResource(R.drawable.ic_settings_block_mode),
                    iconTint = Color(0xFFEA580C),
                    title = stringResource(R.string.settings_dns_response_type),
                    desc = when (dnsResponseType) {
                        AppPreferences.DNS_RESPONSE_NXDOMAIN -> stringResource(R.string.dns_response_nxdomain)
                        AppPreferences.DNS_RESPONSE_REFUSED -> stringResource(R.string.dns_response_refused)
                        else -> stringResource(R.string.dns_response_custom_ip)
                    },
                    onClick = onShowDnsResponseTypeDialog
                )

                HorizontalDivider(modifier = Modifier.padding(horizontal = 16.dp), color = dividerColor)

                // 3. HTTPS Filtering
                SettingItem(
                    iconPainter = painterResource(R.drawable.ic_settings_https),
                    iconTint = Color(0xFF7C3AED),
                    title = stringResource(R.string.https_filtering_title) + " (BETA)",
                    desc = stringResource(R.string.https_filtering_settings_desc),
                    onClick = onNavigateToHttpsFiltering
                )

                HorizontalDivider(modifier = Modifier.padding(horizontal = 16.dp), color = dividerColor)

                // 4. WireGuard Import
                SettingItem(
                    iconPainter = painterResource(R.drawable.ic_settings_wireguard),
                    iconTint = Color(0xFF7C3AED),
                    title = stringResource(R.string.wireguard_import_title) + " (BETA)",
                    desc = stringResource(R.string.wireguard_empty_desc),
                    onClick = onNavigateToWireGuardImport
                )

                HorizontalDivider(modifier = Modifier.padding(horizontal = 16.dp), color = dividerColor)

                // 5. Auto-reconnect
                SettingsToggleItem(
                    iconPainter = painterResource(R.drawable.ic_settings_auto_reconnect),
                    iconTint = Color(0xFF2563EB),
                    title = stringResource(R.string.settings_auto_reconnect),
                    subtitle = stringResource(R.string.settings_auto_reconnect_desc),
                    isChecked = autoReconnect,
                    onCheckedChange = onSetAutoReconnect
                )

                HorizontalDivider(modifier = Modifier.padding(horizontal = 16.dp), color = dividerColor)

                // 6. Safe Search
                SettingsToggleItem(
                    iconPainter = painterResource(R.drawable.ic_settings_safesearch),
                    iconTint = Color(0xFF059669),
                    title = stringResource(R.string.settings_safe_search),
                    subtitle = stringResource(R.string.settings_safe_search_desc),
                    isChecked = safeSearchEnabled,
                    onCheckedChange = onSetSafeSearchEnabled
                )

                HorizontalDivider(modifier = Modifier.padding(horizontal = 16.dp), color = dividerColor)

                // 8. YouTube Restricted
                SettingsToggleItem(
                    iconPainter = painterResource(R.drawable.ic_settings_youtube),
                    iconTint = Color(0xFFE11D48),
                    title = stringResource(R.string.settings_youtube_restricted),
                    subtitle = stringResource(R.string.settings_youtube_restricted_desc),
                    isChecked = youtubeRestrictedMode,
                    onCheckedChange = onSetYoutubeRestrictedMode
                )

                HorizontalDivider(modifier = Modifier.padding(horizontal = 16.dp), color = dividerColor)

                // 9. Root Proxy
                SettingsToggleItem(
                    iconPainter = painterResource(R.drawable.ic_settings_root_proxy),
                    iconTint = Color(0xFF7C3AED),
                    title = stringResource(R.string.settings_root_proxy),
                    subtitle = stringResource(R.string.settings_root_proxy_desc),
                    isChecked = routingMode == AppPreferences.ROUTING_MODE_ROOT,
                    onCheckedChange = onSetRoutingMode
                )
            }
        }
    }
}
