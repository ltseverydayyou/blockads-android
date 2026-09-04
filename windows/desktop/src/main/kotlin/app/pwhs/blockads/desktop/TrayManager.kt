package app.pwhs.blockads.desktop

import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import java.awt.EventQueue
import java.awt.MenuItem
import java.awt.PopupMenu
import java.awt.SystemTray
import java.awt.Toolkit
import java.awt.TrayIcon

object TrayManager {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private var trayIcon: TrayIcon? = null
    private var statusItem: MenuItem? = null
    private var toggleItem: MenuItem? = null
    private var pollingJob: Job? = null

    fun install(onOpen: () -> Unit, onExit: () -> Unit) {
        if (!SystemTray.isSupported() || trayIcon != null) return

        val imageUrl = TrayManager::class.java.getResource("/tray.png")
            ?: error("Bundled tray.png is missing")
        val image = Toolkit.getDefaultToolkit().getImage(imageUrl)
        val popup = PopupMenu()

        val status = MenuItem("Status: Checking...").apply { isEnabled = false }
        val toggle = MenuItem("Toggle Ad Blocking")
        val open = MenuItem("Open BlockAds")
        val exit = MenuItem("Exit")

        popup.add(status)
        popup.add(toggle)
        popup.addSeparator()
        popup.add(open)
        popup.add(exit)

        val icon = TrayIcon(image, "BlockAds", popup).apply {
            isImageAutoSize = true
            addActionListener { onOpen() }
        }

        toggle.addActionListener {
            scope.launch {
                runCatching {
                    BackendClient.ensureRunning()
                    val current = BackendClient.status()
                    if (current.running) BackendClient.stopProtection() else BackendClient.startProtection()
                    refreshStatus()
                }
            }
        }
        open.addActionListener { onOpen() }
        exit.addActionListener {
            remove()
            onExit()
        }

        if (EventQueue.isDispatchThread()) SystemTray.getSystemTray().add(icon)
        else EventQueue.invokeAndWait { SystemTray.getSystemTray().add(icon) }
        trayIcon = icon
        statusItem = status
        toggleItem = toggle

        pollingJob?.cancel()
        pollingJob = scope.launch {
            while (isActive) {
                refreshStatus()
                delay(1500)
            }
        }
    }

    private suspend fun refreshStatus() {
        val status = runCatching { BackendClient.status() }.getOrNull()
        val running = status?.running == true
        val known = status != null
        EventQueue.invokeLater {
            statusItem?.label = when {
                !known -> "Status: Core unavailable"
                running -> "Status: Protected"
                else -> "Status: Unprotected"
            }
            toggleItem?.label = if (running) "Turn Off Ad Blocking" else "Turn On Ad Blocking"
            trayIcon?.toolTip = when {
                !known -> "BlockAds - Core unavailable"
                running -> "BlockAds - Protected"
                else -> "BlockAds - Unprotected"
            }
        }
    }

    fun remove() {
        pollingJob?.cancel()
        pollingJob = null
        trayIcon?.let { icon ->
            EventQueue.invokeLater { SystemTray.getSystemTray().remove(icon) }
        }
        trayIcon = null
        statusItem = null
        toggleItem = null
    }
}
