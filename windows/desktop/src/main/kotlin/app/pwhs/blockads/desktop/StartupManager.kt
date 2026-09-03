package app.pwhs.blockads.desktop

import java.nio.file.Path

object StartupManager {
    private const val RUN_KEY = "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run"
    private const val VALUE_NAME = "BlockAds"

    fun sync(enabled: Boolean) {
        if (enabled) enable() else disable()
    }

    private fun enable() {
        val executable = applicationExecutable()
        val command = "\"$executable\""
        runReg("add", RUN_KEY, "/v", VALUE_NAME, "/t", "REG_SZ", "/d", command, "/f")
    }

    private fun disable() {
        val process = ProcessBuilder("reg.exe", "delete", RUN_KEY, "/v", VALUE_NAME, "/f")
            .redirectErrorStream(true)
            .start()
        val output = process.inputStream.bufferedReader().use { it.readText() }
        val exit = process.waitFor()
        if (exit != 0 && !output.contains("unable to find", ignoreCase = true)) {
            error(output.trim().ifEmpty { "Failed to remove BlockAds from Windows startup (exit $exit)" })
        }
    }

    private fun applicationExecutable(): String {
        val packaged = System.getProperty("jpackage.app-path")
            ?.trim()
            ?.takeIf { it.isNotEmpty() }
        val processCommand = ProcessHandle.current().info().command().orElse(null)
            ?.trim()
            ?.takeIf { it.isNotEmpty() }
        val candidate = packaged ?: processCommand ?: error("Unable to determine the BlockAds executable path")
        val absolute = Path.of(candidate).toAbsolutePath().normalize().toString()
        if (!absolute.endsWith(".exe", ignoreCase = true)) {
            error("Windows startup registration requires the packaged BlockAds executable")
        }
        return absolute
    }

    private fun runReg(vararg args: String) {
        val process = ProcessBuilder(listOf("reg.exe") + args)
            .redirectErrorStream(true)
            .start()
        val output = process.inputStream.bufferedReader().use { it.readText() }
        val exit = process.waitFor()
        if (exit != 0) {
            error(output.trim().ifEmpty { "Windows startup registration failed (exit $exit)" })
        }
    }
}
