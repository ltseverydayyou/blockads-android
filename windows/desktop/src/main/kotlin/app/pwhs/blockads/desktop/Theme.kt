package app.pwhs.blockads.desktop

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Typography
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.sp

val DarkBackground = Color(0xFF0D1117)
val DarkSurface = Color(0xFF161B22)
val DarkSurfaceVariant = Color(0xFF21262D)
val NeonGreen = Color(0xFF39D353)
val NeonGreenDim = Color(0xFF238636)
val DangerRed = Color(0xFFFF6B6B)
val DangerRedDim = Color(0xFFDA3633)
val AccentBlue = Color(0xFF58A6FF)
val AccentBlueDim = Color(0xFF388BFD)
val AccentGreen = Color(0xFF39D353)
val AccentGreenDim = Color(0xFF238636)
val AccentBluePreset = Color(0xFF4285F4)
val AccentBluePresetDim = Color(0xFF1A73E8)
val AccentPurple = Color(0xFFA855F7)
val AccentPurpleDim = Color(0xFF7C3AED)
val AccentOrange = Color(0xFFF97316)
val AccentOrangeDim = Color(0xFFEA580C)
val AccentPink = Color(0xFFEC4899)
val AccentPinkDim = Color(0xFFDB2777)
val AccentTeal = Color(0xFF14B8A6)
val AccentTealDim = Color(0xFF0D9488)
val AccentGrey = Color(0xFFBDBDBD)
val AccentGreyDim = Color(0xFF9E9E9E)
val TextPrimary = Color(0xFFF0F6FC)
val TextSecondary = Color(0xFF8B949E)
val TextTertiary = Color(0xFF6E7681)
val LightBackground = Color(0xFFF6F8FA)
val LightSurface = Color(0xFFFFFFFF)
val LightSurfaceVariant = Color(0xFFEFF2F5)
val LightTextPrimary = Color(0xFF1F2328)
val LightTextSecondary = Color(0xFF656D76)
val SecurityOrange = Color(0xFFFF9800)
val WhitelistAmber = Color(0xFFFFC107)
val UpstreamDnsPurple = Color(0xFFA855F7)

private val DarkColorScheme = darkColorScheme(
    primary = NeonGreen,
    onPrimary = Color.Black,
    primaryContainer = NeonGreenDim,
    onPrimaryContainer = Color.White,
    secondary = AccentBlue,
    onSecondary = Color.Black,
    secondaryContainer = AccentBlueDim,
    onSecondaryContainer = Color.White,
    tertiary = DangerRed,
    background = DarkBackground,
    onBackground = TextPrimary,
    surface = DarkSurface,
    onSurface = TextPrimary,
    surfaceVariant = DarkSurfaceVariant,
    onSurfaceVariant = TextSecondary,
    outline = TextTertiary,
    error = DangerRed,
    onError = Color.White,
)

private val LightColorScheme = lightColorScheme(
    primary = NeonGreenDim,
    onPrimary = Color.White,
    primaryContainer = NeonGreen,
    onPrimaryContainer = Color.Black,
    secondary = AccentBlueDim,
    onSecondary = Color.White,
    secondaryContainer = AccentBlue,
    onSecondaryContainer = Color.Black,
    tertiary = DangerRedDim,
    background = LightBackground,
    onBackground = LightTextPrimary,
    surface = LightSurface,
    onSurface = LightTextPrimary,
    surfaceVariant = LightSurfaceVariant,
    onSurfaceVariant = LightTextSecondary,
    outline = LightTextSecondary,
    error = DangerRedDim,
    onError = Color.White,
)

private fun accentPair(key: String): Pair<Color, Color> = when (key) {
    "blue" -> AccentBluePreset to AccentBluePresetDim
    "purple" -> AccentPurple to AccentPurpleDim
    "orange" -> AccentOrange to AccentOrangeDim
    "pink" -> AccentPink to AccentPinkDim
    "teal" -> AccentTeal to AccentTealDim
    "grey" -> AccentGrey to AccentGreyDim
    else -> AccentGreen to AccentGreenDim
}

val BlockAdsTypography = Typography(
    displayLarge = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Bold, fontSize = 57.sp, lineHeight = 64.sp, letterSpacing = (-0.25).sp),
    displayMedium = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Bold, fontSize = 45.sp, lineHeight = 52.sp),
    displaySmall = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Bold, fontSize = 36.sp, lineHeight = 44.sp),
    headlineLarge = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Bold, fontSize = 32.sp, lineHeight = 40.sp),
    headlineMedium = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.SemiBold, fontSize = 28.sp, lineHeight = 36.sp),
    headlineSmall = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.SemiBold, fontSize = 24.sp, lineHeight = 32.sp),
    titleLarge = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.SemiBold, fontSize = 22.sp, lineHeight = 28.sp),
    titleMedium = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.SemiBold, fontSize = 16.sp, lineHeight = 24.sp, letterSpacing = 0.15.sp),
    titleSmall = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Medium, fontSize = 14.sp, lineHeight = 20.sp, letterSpacing = 0.1.sp),
    bodyLarge = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Normal, fontSize = 16.sp, lineHeight = 24.sp, letterSpacing = 0.5.sp),
    bodyMedium = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Normal, fontSize = 14.sp, lineHeight = 20.sp, letterSpacing = 0.25.sp),
    bodySmall = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Normal, fontSize = 12.sp, lineHeight = 16.sp, letterSpacing = 0.4.sp),
    labelLarge = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Medium, fontSize = 14.sp, lineHeight = 20.sp, letterSpacing = 0.1.sp),
    labelMedium = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Medium, fontSize = 12.sp, lineHeight = 16.sp, letterSpacing = 0.5.sp),
    labelSmall = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Medium, fontSize = 11.sp, lineHeight = 16.sp, letterSpacing = 0.5.sp),
)
@Composable
fun BlockadsTheme(themeMode: String, accentColor: String, content: @Composable () -> Unit) {
    val darkTheme = when (themeMode) {
        "dark" -> true
        "light" -> false
        else -> isSystemInDarkTheme()
    }
    val (primary, dim) = accentPair(accentColor)
    val scheme = if (darkTheme) {
        DarkColorScheme.copy(primary = primary, primaryContainer = dim)
    } else {
        LightColorScheme.copy(primary = dim, primaryContainer = primary)
    }
    MaterialTheme(colorScheme = scheme, typography = BlockAdsTypography, content = content)
}
