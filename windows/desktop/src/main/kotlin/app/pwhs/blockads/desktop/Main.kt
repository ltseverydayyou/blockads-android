package app.pwhs.blockads.desktop

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.runtime.remember
import androidx.compose.ui.graphics.vector.rememberVectorPainter
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Window
import androidx.compose.ui.window.WindowState
import androidx.compose.ui.window.application

fun main() = application {
    val state = remember { DesktopState() }
    BlockadsTheme(state.settings.themeMode, state.settings.accentColor) {
        Window(
            onCloseRequest = ::exitApplication,
            title = "BlockAds",
            icon = rememberVectorPainter(BlockAdsIcons.Shield),
            state = WindowState(width = 1080.dp, height = 820.dp),
        ) {
            Surface(color = MaterialTheme.colorScheme.background) {
                BlockAdsDesktopApp(state)
            }
        }
    }
}
