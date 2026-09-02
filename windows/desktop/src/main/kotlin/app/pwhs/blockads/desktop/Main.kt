package app.pwhs.blockads.desktop

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.runtime.remember
import androidx.compose.ui.graphics.vector.rememberVectorPainter
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Window
import androidx.compose.ui.window.application
import androidx.compose.ui.window.rememberWindowState

fun main() = application {
    val state = remember { DesktopState() }
    val windowState = rememberWindowState(width = 1080.dp, height = 820.dp)

    Window(
        onCloseRequest = ::exitApplication,
        title = "BlockAds",
        icon = rememberVectorPainter(BlockAdsIcons.Shield),
        state = windowState,
    ) {
        BlockadsTheme(state.settings.themeMode, state.settings.accentColor) {
            Surface(color = MaterialTheme.colorScheme.background) {
                BlockAdsDesktopApp(state)
            }
        }
    }
}
