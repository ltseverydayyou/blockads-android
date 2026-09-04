package app.pwhs.blockads.desktop

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import java.awt.Desktop
import java.net.URI
import java.net.http.HttpClient
import java.net.http.HttpRequest
import java.net.http.HttpResponse
import java.time.Duration

@Serializable
private data class GitHubRelease(
    val tag_name: String,
    val name: String = "",
    val html_url: String,
    val draft: Boolean = false,
    val prerelease: Boolean = false,
)

data class WindowsUpdate(
    val version: String,
    val name: String,
    val url: String,
)

object UpdateChecker {
    const val currentVersion = "1.3.0"
    private const val releasesUrl = "https://api.github.com/repos/ltseverydayyou/blockads-android/releases?per_page=20"
    private val json = Json { ignoreUnknownKeys = true }
    private val http = HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(5)).build()

    suspend fun check(): WindowsUpdate? = withContext(Dispatchers.IO) {
        runCatching {
            val request = HttpRequest.newBuilder(URI(releasesUrl))
                .timeout(Duration.ofSeconds(10))
                .header("Accept", "application/vnd.github+json")
                .header("User-Agent", "BlockAds-Windows/$currentVersion")
                .GET()
                .build()
            val response = http.send(request, HttpResponse.BodyHandlers.ofString())
            if (response.statusCode() !in 200..299) return@runCatching null
            json.decodeFromString<List<GitHubRelease>>(response.body())
                .asSequence()
                .filter { !it.draft && !it.prerelease }
                .mapNotNull { release ->
                    parseWindowsVersion(release.tag_name)?.let { version -> version to release }
                }
                .filter { (version, _) -> compareVersions(version, currentVersion) > 0 }
                .maxWithOrNull { a, b -> compareVersions(a.first, b.first) }
                ?.let { (version, release) ->
                    WindowsUpdate(version, release.name.ifBlank { "BlockAds for Windows v$version" }, release.html_url)
                }
        }.getOrNull()
    }

    fun open(update: WindowsUpdate) {
        if (!Desktop.isDesktopSupported()) return
        runCatching { Desktop.getDesktop().browse(URI(update.url)) }
    }

    internal fun parseWindowsVersion(tag: String): String? {
        val value = tag.trim().lowercase()
        return (when {
            value.startsWith("windows-v") -> value.removePrefix("windows-v")
            value.startsWith("windows-") -> value.removePrefix("windows-").removePrefix("v")
            else -> null
        })?.takeIf { it.matches(Regex("\\d+(?:\\.\\d+)*")) }
    }

    internal fun compareVersions(left: String, right: String): Int {
        val a = left.split('.').map { it.toIntOrNull() ?: 0 }
        val b = right.split('.').map { it.toIntOrNull() ?: 0 }
        val size = maxOf(a.size, b.size)
        for (i in 0 until size) {
            val av = a.getOrElse(i) { 0 }
            val bv = b.getOrElse(i) { 0 }
            if (av != bv) return av.compareTo(bv)
        }
        return 0
    }
}
