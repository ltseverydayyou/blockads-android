package app.pwhs.blockads.desktop

import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Switch
import androidx.compose.material3.SwitchDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import java.text.NumberFormat
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

fun formatCount(value: Long): String = NumberFormat.getIntegerInstance().format(value)
fun formatCount(value: Int): String = NumberFormat.getIntegerInstance().format(value)
fun formatDate(timestamp: Long): String = if (timestamp <= 0) "" else SimpleDateFormat("MMM d, HH:mm", Locale.getDefault()).format(Date(timestamp))

@Composable
fun PowerButton(
    isActive: Boolean,
    isConnecting: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val buttonColor by animateColorAsState(
        targetValue = when {
            isConnecting -> AccentBlue
            isActive -> MaterialTheme.colorScheme.primary
            else -> DangerRed
        },
        animationSpec = tween(500),
        label = "buttonColor",
    )
    val scale by animateFloatAsState(
        targetValue = if (isActive || isConnecting) 1f else 0.95f,
        animationSpec = tween(300),
        label = "scale",
    )
    val glowAlpha by animateFloatAsState(
        targetValue = when {
            isConnecting -> 0.3f
            isActive -> 0.4f
            else -> 0.2f
        },
        animationSpec = tween(500),
        label = "glow",
    )
    val transition = rememberInfiniteTransition(label = "pulse")
    val pulseScale by transition.animateFloat(
        initialValue = 1f,
        targetValue = if (isActive && !isConnecting) 1.08f else 1f,
        animationSpec = infiniteRepeatable(tween(1500, easing = FastOutSlowInEasing), RepeatMode.Reverse),
        label = "pulseScale",
    )
    Box(contentAlignment = Alignment.Center, modifier = modifier.size(180.dp)) {
        Box(
            Modifier.size(180.dp)
                .graphicsLayer { scaleX = pulseScale; scaleY = pulseScale }
                .clip(CircleShape)
                .background(Brush.radialGradient(listOf(buttonColor.copy(alpha = glowAlpha), Color.Transparent)))
        )
        Box(
            contentAlignment = Alignment.Center,
            modifier = Modifier.size(140.dp)
                .graphicsLayer { scaleX = scale; scaleY = scale }
                .shadow(if (isActive || isConnecting) 20.dp else 8.dp, CircleShape)
                .clip(CircleShape)
                .background(Brush.radialGradient(listOf(buttonColor.copy(alpha = 0.2f), MaterialTheme.colorScheme.surface)))
                .border(3.dp, Brush.linearGradient(listOf(buttonColor, buttonColor.copy(alpha = 0.5f))), CircleShape)
                .clickable(enabled = !isConnecting, onClick = onClick)
        ) {
            if (isConnecting) {
                CircularProgressIndicator(color = buttonColor, modifier = Modifier.size(56.dp), strokeWidth = 3.dp)
            } else {
                Icon(BlockAdsIcons.Power, null, tint = buttonColor, modifier = Modifier.size(64.dp))
            }
        }
    }
}

@Composable
fun StatCard(icon: ImageVector, label: String, value: String, color: Color, modifier: Modifier = Modifier) {
    Card(
        modifier = modifier,
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        shape = RoundedCornerShape(16.dp),
    ) {
        Column(Modifier.padding(20.dp)) {
            Icon(icon, null, tint = color, modifier = Modifier.size(24.dp))
            Spacer(Modifier.height(12.dp))
            Text(value, style = MaterialTheme.typography.headlineSmall, fontWeight = FontWeight.Bold, color = MaterialTheme.colorScheme.onBackground)
            Text(label, style = MaterialTheme.typography.bodySmall, color = TextSecondary)
        }
    }
}

@Composable
fun SectionCard(content: @Composable () -> Unit) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
    ) { content() }
}

@Composable
fun SectionHeader(title: String, description: String? = null) {
    Column(Modifier.fillMaxWidth().padding(start = 4.dp, top = 8.dp, bottom = 8.dp)) {
        Text(title, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
        if (!description.isNullOrBlank()) Text(description, style = MaterialTheme.typography.bodySmall, color = TextSecondary)
    }
}

@Composable
fun SettingsToggleItem(
    icon: ImageVector,
    title: String,
    subtitle: String,
    checked: Boolean,
    onCheckedChange: (Boolean) -> Unit,
) {
    Row(
        Modifier.fillMaxWidth().clickable { onCheckedChange(!checked) }.padding(16.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(icon, null, tint = MaterialTheme.colorScheme.secondary, modifier = Modifier.size(20.dp))
        Spacer(Modifier.width(12.dp))
        Column(Modifier.weight(1f)) {
            Text(title, style = MaterialTheme.typography.titleSmall)
            Text(subtitle, style = MaterialTheme.typography.bodySmall, color = TextSecondary)
        }
        Switch(
            checked = checked,
            onCheckedChange = null,
            colors = SwitchDefaults.colors(checkedThumbColor = MaterialTheme.colorScheme.onPrimary, checkedTrackColor = MaterialTheme.colorScheme.primary),
        )
    }
}

@Composable
fun SettingItem(
    icon: ImageVector,
    title: String,
    subtitle: String,
    trailing: String = "›",
    onClick: () -> Unit,
) {
    Row(
        Modifier.fillMaxWidth().clickable(onClick = onClick).padding(16.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(icon, null, tint = MaterialTheme.colorScheme.secondary, modifier = Modifier.size(20.dp))
        Spacer(Modifier.width(12.dp))
        Column(Modifier.weight(1f)) {
            Text(title, style = MaterialTheme.typography.titleSmall)
            Text(subtitle, style = MaterialTheme.typography.bodySmall, color = TextSecondary, maxLines = 2, overflow = TextOverflow.Ellipsis)
        }
        Text(trailing, style = MaterialTheme.typography.titleLarge, color = TextSecondary)
    }
}

@Composable
fun DividerInset() {
    HorizontalDivider(Modifier.padding(horizontal = 16.dp), color = MaterialTheme.colorScheme.outline.copy(alpha = 0.1f))
}
