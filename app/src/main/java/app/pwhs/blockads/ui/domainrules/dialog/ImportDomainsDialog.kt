package app.pwhs.blockads.ui.domainrules.dialog

import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Block
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.FileUpload
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Button
import androidx.compose.material3.FilterChip
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import app.pwhs.blockads.R
import app.pwhs.blockads.ui.domainrules.DomainRuleParser
import app.pwhs.blockads.ui.theme.TextSecondary
import timber.log.Timber

@Composable
fun ImportDomainsDialog(
    initialIsAllow: Boolean,
    onDismiss: () -> Unit,
    onImport: (domains: List<String>, isAllow: Boolean) -> Unit,
    modifier: Modifier = Modifier
) {
    val context = LocalContext.current
    var isAllow by remember { mutableStateOf(initialIsAllow) }
    var rawContent by remember { mutableStateOf("") }

    val filePickerLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.GetContent()
    ) { uri: Uri? ->
        uri?.let {
            try {
                context.contentResolver.openInputStream(it)?.use { stream ->
                    val text = stream.bufferedReader().readText()
                    rawContent = text
                }
            } catch (e: Exception) {
                Timber.e(e, "Failed to read file from URI: $it")
            }
        }
    }

    val parsedDomains = remember(rawContent) {
        DomainRuleParser.parseDomains(rawContent)
    }

    AlertDialog(
        modifier = modifier,
        containerColor = MaterialTheme.colorScheme.background,
        onDismissRequest = onDismiss,
        title = { Text(stringResource(R.string.settings_import)) },
        text = {
            Column(
                modifier = Modifier.fillMaxWidth(),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                // Rule type selection
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    FilterChip(
                        selected = isAllow,
                        onClick = { isAllow = true },
                        label = { Text(stringResource(R.string.domain_rules_tab_whitelist)) },
                        leadingIcon = {
                            Icon(
                                Icons.Default.CheckCircle,
                                contentDescription = null,
                                modifier = Modifier.size(16.dp)
                            )
                        }
                    )
                    FilterChip(
                        selected = !isAllow,
                        onClick = { isAllow = false },
                        label = { Text(stringResource(R.string.domain_rules_tab_blocklist)) },
                        leadingIcon = {
                            Icon(
                                Icons.Default.Block,
                                contentDescription = null,
                                modifier = Modifier.size(16.dp)
                            )
                        }
                    )
                }

                // File picker button
                OutlinedButton(
                    onClick = { filePickerLauncher.launch("*/*") },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Icon(
                        Icons.Default.FileUpload,
                        contentDescription = null,
                        modifier = Modifier.size(18.dp)
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(stringResource(R.string.import_from_file))
                }

                // Raw content text field
                OutlinedTextField(
                    value = rawContent,
                    onValueChange = { rawContent = it },
                    label = { Text(stringResource(R.string.settings_add_domain_hint)) },
                    placeholder = {
                        Text(
                            text = "example.com\n127.0.0.1 ad.example.com\n||tracker.com^",
                            style = MaterialTheme.typography.bodySmall
                        )
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(140.dp),
                    shape = RoundedCornerShape(12.dp),
                    maxLines = 6
                )

                if (parsedDomains.isNotEmpty()) {
                    Text(
                        text = "${parsedDomains.size} domains",
                        style = MaterialTheme.typography.labelMedium,
                        color = MaterialTheme.colorScheme.primary,
                        modifier = Modifier.align(Alignment.End)
                    )
                }
            }
        },
        confirmButton = {
            Button(
                onClick = {
                    if (parsedDomains.isNotEmpty()) {
                        onImport(parsedDomains, isAllow)
                    }
                },
                enabled = parsedDomains.isNotEmpty()
            ) {
                Text(
                    if (parsedDomains.isNotEmpty()) {
                        "${stringResource(R.string.settings_import)} (${parsedDomains.size})"
                    } else {
                        stringResource(R.string.settings_import)
                    }
                )
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) { Text(stringResource(R.string.settings_cancel)) }
        }
    )
}
