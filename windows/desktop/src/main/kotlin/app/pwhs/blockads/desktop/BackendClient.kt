package app.pwhs.blockads.desktop

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.delay
import kotlinx.coroutines.withContext
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.serializer
import java.io.File
import java.net.URI
import java.net.URLEncoder
import java.net.http.HttpClient
import java.net.http.HttpRequest
import java.net.http.HttpResponse
import java.nio.charset.StandardCharsets
import java.time.Duration

class BackendException(message: String) : RuntimeException(message)

object BackendClient {
    private const val base = "http://127.0.0.1:8754"
    private val json = Json { ignoreUnknownKeys = true; encodeDefaults = true }
    private val http = HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(2)).build()
    private var startedProcess: Process? = null

    suspend fun ensureRunning() = withContext(Dispatchers.IO) {
        if (ping()) return@withContext
        val runtimeDir = File(System.getenv("LOCALAPPDATA") ?: ".", "BlockAds/runtime").apply { mkdirs() }
        val exe = File(runtimeDir, "BlockAdsCore.exe")
        BackendClient::class.java.getResourceAsStream("/backend/BlockAdsCore.exe")?.use { input ->
            exe.outputStream().use { input.copyTo(it) }
        } ?: throw BackendException("Bundled BlockAdsCore.exe is missing")
        val log = File(runtimeDir, "core.log")
        startedProcess = ProcessBuilder(exe.absolutePath)
            .directory(runtimeDir)
            .redirectOutput(ProcessBuilder.Redirect.appendTo(log))
            .redirectError(ProcessBuilder.Redirect.appendTo(log))
            .start()
        repeat(50) {
            if (ping()) return@withContext
            delay(100)
        }
        throw BackendException("BlockAds Windows core did not start. See ${log.absolutePath}")
    }

    private fun ping(): Boolean = try {
        val request = HttpRequest.newBuilder(URI("$base/health")).timeout(Duration.ofSeconds(1)).GET().build()
        http.send(request, HttpResponse.BodyHandlers.discarding()).statusCode() == 200
    } catch (_: Exception) { false }

    private suspend inline fun <reified T> get(path: String): T = request("GET", path, null)
    private suspend inline fun <reified T> delete(path: String): T = request("DELETE", path, null)

    private suspend inline fun <reified B, reified T> send(method: String, path: String, body: B): T =
        request(method, path, json.encodeToString(serializer<B>(), body))

    private suspend inline fun <reified T> request(method: String, path: String, body: String?): T = withContext(Dispatchers.IO) {
        val builder = HttpRequest.newBuilder(URI("$base$path"))
            .timeout(Duration.ofSeconds(90))
            .header("Accept", "application/json")
        if (body == null) builder.method(method, HttpRequest.BodyPublishers.noBody())
        else builder.header("Content-Type", "application/json").method(method, HttpRequest.BodyPublishers.ofString(body))
        val response = http.send(builder.build(), HttpResponse.BodyHandlers.ofString())
        if (response.statusCode() !in 200..299) {
            val msg = runCatching { json.parseToJsonElement(response.body()).jsonObject["error"]?.toString()?.trim('"') }.getOrNull()
            throw BackendException(msg ?: "Backend HTTP ${response.statusCode()}")
        }
        json.decodeFromString(serializer<T>(), response.body())
    }

    suspend fun status(): Status = get("/status")
    suspend fun settings(): Settings = get("/settings")
    suspend fun saveSettings(settings: Settings): Settings = send("PUT", "/settings", settings)
    suspend fun startProtection(): Status = send("POST", "/protection/start", mapOf("systemDns" to true))
    suspend fun stopProtection(): Status = send("POST", "/protection/stop", emptyMap<String, String>())
    suspend fun filters(): List<FilterList> = get("/filters")
    suspend fun updateFilters() { request<Map<String, kotlinx.serialization.json.JsonElement>>("POST", "/filters/update", null) }
    suspend fun setFilter(id: String, enabled: Boolean): List<FilterList> =
        send("PUT", "/filters/${enc(id)}", mapOf("enabled" to enabled))
    suspend fun addCustomFilter(name: String, url: String): FilterList =
        send("POST", "/filters/custom", mapOf("name" to name, "url" to url))
    suspend fun removeCustomFilter(id: String): List<FilterList> = delete("/filters/${enc(id)}")
    suspend fun profiles(): List<Profile> = get("/profiles")
    suspend fun activateProfile(id: String) { request<Map<String, kotlinx.serialization.json.JsonElement>>("POST", "/profiles/${enc(id)}/activate", null) }
    suspend fun rules(): List<Rule> = get("/rules")
    suspend fun addRule(rule: String): Rule = send("POST", "/rules", mapOf("rule" to rule))
    suspend fun setRule(id: Long, enabled: Boolean): List<Rule> = send("PUT", "/rules/$id", mapOf("enabled" to enabled))
    suspend fun deleteRule(id: Long): List<Rule> = delete("/rules/$id")
    suspend fun logs(): List<LogEntry> = get("/logs")
    suspend fun clearLogs(): List<LogEntry> = delete("/logs")

    private fun enc(value: String): String = URLEncoder.encode(value, StandardCharsets.UTF_8).replace("+", "%20")
}
