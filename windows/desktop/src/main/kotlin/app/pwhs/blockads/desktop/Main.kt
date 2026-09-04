package app.pwhs.blockads.desktop

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.graphics.vector.rememberVectorPainter
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Window
import androidx.compose.ui.window.application
import androidx.compose.ui.window.rememberWindowState
import kotlinx.coroutines.runBlocking

fun main(args: Array<String>) {
    if (args.any { it.equals("--toggle", ignoreCase = true) }) {
        runBlocking {
            runCatching {
                BackendClient.ensureRunning()
                val current = BackendClient.status()
                if (current.running) BackendClient.stopProtection() else BackendClient.startProtection()
            }
        }
        return
    }

    application {
        val state = remember { DesktopState() }
        val windowState = rememberWindowState(width = 1080.dp, height = 820.dp)
        var windowVisible by remember { mutableStateOf(true) }

        fun openWindow() {
            windowVisible = true
            windowState.isMinimized = false
        }

        LaunchedEffect(windowState.isMinimized, state.settings.minimizeToTray) {
            if (windowState.isMinimized && state.settings.minimizeToTray) {
                windowState.isMinimized = false
                windowVisible = false
            }
        }

        LaunchedEffect(Unit) {
            runCatching { JumpListManager.install() }.onFailure { error ->
                val base = System.getenv("LOCALAPPDATA") ?: "."
                java.io.File(base, "BlockAds/jumplist.log").apply { parentFile?.mkdirs(); writeText(error.stackTraceToString()) }
            }
            runCatching {
                TrayManager.install(
                    onOpen = { openWindow() },
                    onExit = {
                        TrayManager.remove()
                        exitApplication()
                    },
                )
            }.onFailure { error ->
                val base = System.getenv("LOCALAPPDATA") ?: "."
                java.io.File(base, "BlockAds/tray.log").apply { parentFile?.mkdirs(); writeText(error.stackTraceToString()) }
            }
        }

        Window(
            onCloseRequest = {
                if (state.settings.minimizeToTray) windowVisible = false
                else {
                    TrayManager.remove()
                    exitApplication()
                }
            },
            visible = windowVisible,
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
}
