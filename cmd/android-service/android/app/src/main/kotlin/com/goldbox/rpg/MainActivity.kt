package com.goldbox.rpg

import android.content.Intent
import android.content.res.AssetManager
import android.net.Uri
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.view.View
import android.widget.Button
import android.widget.ProgressBar
import android.widget.ScrollView
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import java.io.BufferedReader
import java.io.File
import java.io.InputStreamReader
import java.net.Inet4Address
import java.net.NetworkInterface

class MainActivity : AppCompatActivity() {
    private lateinit var btnToggle: Button
    private val maxLogLines = 500
    private val logBuffer = ArrayDeque<String>(maxLogLines)
    private lateinit var btnOpen: Button
    private lateinit var btnShareLan: Button
    private lateinit var progressBar: ProgressBar
    private lateinit var tvStatus: TextView
    private lateinit var tvUrl: TextView
    private lateinit var tvLanUrl: TextView
    private lateinit var tvLogs: TextView
    private lateinit var scrollLogs: ScrollView
    private var serviceProcess: Process? = null
    private var isRunning = false
    private val handler = Handler(Looper.getMainLooper())
    private val port = 8080

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)
        btnToggle = findViewById(R.id.btnToggle)
        btnOpen = findViewById(R.id.btnOpen)
        btnShareLan = findViewById(R.id.btnShareLan)
        progressBar = findViewById(R.id.progressBar)
        tvStatus = findViewById(R.id.tvStatus)
        tvUrl = findViewById(R.id.tvUrl)
        tvLanUrl = findViewById(R.id.tvLanUrl)
        tvLogs = findViewById(R.id.tvLogs)
        scrollLogs = findViewById(R.id.scrollLogs)
        tvUrl.text = "Local: http://127.0.0.1:$port"
        updateLanDisplay()
        btnToggle.setOnClickListener { if (isRunning) stopService() else startService() }
        btnOpen.setOnClickListener {
            startActivity(Intent(Intent.ACTION_VIEW, Uri.parse("http://127.0.0.1:$port")))
        }
        btnShareLan.setOnClickListener {
            val ip = getLanIpAddress() ?: return@setOnClickListener
            val url = "http://$ip:$port"
            val sendIntent = Intent(Intent.ACTION_SEND).apply {
                type = "text/plain"
                putExtra(Intent.EXTRA_TEXT, url)
            }
            if (sendIntent.resolveActivity(packageManager) != null) {
                val shareIntent = Intent.createChooser(
                    sendIntent,
                    getString(R.string.share)
                )
                startActivity(shareIntent)
            }
        }
    }

    private fun serviceBinaryPath(): String =
        "${applicationInfo.nativeLibraryDir}/libwebservice.so"

    /** Directory inside internal storage where bundled assets are extracted. */
    private fun extractedAssetsDir(): File = File(filesDir, "goldbox")

    /**
     * Extract bundled assets (web UI and game data) from the APK into internal
     * storage so the Go process can serve them from the filesystem.  A simple
     * version stamp avoids re-extracting on every launch.
     */
    private fun extractAssetsIfNeeded() {
        val destRoot = extractedAssetsDir()
        val stamp = File(destRoot, ".version")
        val currentVersion = packageManager.getPackageInfo(packageName, 0).versionCode.toString()
        if (stamp.exists() && stamp.readText().trim() == currentVersion) return

        handler.post { appendLog("Extracting game assets...") }
        // Only clear bundled asset directories so that any runtime data the
        // Go server writes under destRoot (e.g. saves, sessions) is preserved
        // across app updates.
        val webDir = File(destRoot, "web")
        if (webDir.exists()) {
            webDir.deleteRecursively()
        }
        val dataDir = File(destRoot, "data")
        if (dataDir.exists()) {
            dataDir.deleteRecursively()
        }
        copyAssetDir(assets, "web", File(destRoot, "web"))
        copyAssetDir(assets, "data", File(destRoot, "data"))
        stamp.parentFile?.mkdirs()
        stamp.writeText(currentVersion)
        handler.post { appendLog("Asset extraction complete.") }
    }

    /** Recursively copy an asset directory to the filesystem. */
    private fun copyAssetDir(am: AssetManager, assetPath: String, dest: File) {
        val children = am.list(assetPath) ?: return
        if (children.isNotEmpty()) {
            // It is a directory – recurse into children.
            dest.mkdirs()
            for (child in children) {
                copyAssetDir(am, "$assetPath/$child", File(dest, child))
            }
        } else {
            // It is a file – copy it.
            dest.parentFile?.mkdirs()
            try {
                am.open(assetPath).use { input ->
                    dest.outputStream().use { output -> input.copyTo(output) }
                }
            } catch (e: java.io.FileNotFoundException) {
                // Empty directory or inaccessible asset – skip silently.
            }
        }
    }

    private fun startService() {
        val binary = File(serviceBinaryPath())
        if (!binary.exists()) {
            appendLog("ERROR: Service binary not found at ${binary.absolutePath}"); return
        }
        // Run potentially slow asset extraction on a background thread to
        // avoid UI jank / ANRs on slower devices or large asset sets.
        btnToggle.isEnabled = false
        Thread {
            try {
                extractAssetsIfNeeded()
                handler.post {
                    btnToggle.isEnabled = true
                    launchServiceProcess(binary)
                }
            } catch (e: Exception) {
                handler.post {
                    btnToggle.isEnabled = true
                    appendLog("ERROR: Asset extraction failed: ${e.message}")
                }
            }
        }.start()
    }

    private fun launchServiceProcess(binary: File) {
        try {
            val root = extractedAssetsDir()
            val pb = ProcessBuilder(binary.absolutePath)
            pb.redirectErrorStream(true)
            pb.directory(root)
            pb.environment()["HOME"] = filesDir.absolutePath
            pb.environment()["WEB_DIR"] = File(root, "web").absolutePath
            pb.environment()["DATA_DIR"] = File(root, "data").absolutePath
            pb.environment()["SERVER_PORT"] = port.toString()
            serviceProcess = pb.start()
            isRunning = true
            updateUI()
            appendLog("Service starting on port $port...")
            readProcessOutput()
        } catch (e: Exception) {
            appendLog("ERROR: Failed to start service: ${e.message}")
            isRunning = false; updateUI()
        }
    }

    private fun stopService() {
        val process = serviceProcess
        if (process == null) {
            isRunning = false
            updateUI()
            appendLog("Service stopped.")
            return
        }

        Thread {
            process.destroy()
            logThread?.interrupt()
            try {
                process.waitFor()
            } catch (_: InterruptedException) {
                Thread.currentThread().interrupt()
                process.destroyForcibly()
            }
            handler.post {
                logThread = null
                if (serviceProcess === process) {
                    serviceProcess = null
                }
                if (isRunning) {
                    isRunning = false
                    updateUI()
                }
                appendLog("Service stopped.")
            }
        }.start()
    }

    private var logThread: Thread? = null

    private fun readProcessOutput() {
        logThread = Thread {
            try {
                val stream = serviceProcess?.inputStream ?: return@Thread
                BufferedReader(InputStreamReader(stream)).use { reader ->
                    var line: String?
                    while (reader.readLine().also { line = it } != null) {
                        val text = line ?: continue
                        handler.post { appendLog(text) }
                    }
                }
            } catch (e: Exception) { handler.post { appendLog(getString(R.string.error_log_reader, e.message ?: "")) } }
            handler.post {
                if (isRunning) { isRunning = false; updateUI(); appendLog(getString(R.string.log_service_exited)) }
            }
        }.also { it.start() }
    }

    private fun updateUI() {
        btnToggle.text = getString(if (isRunning) R.string.btn_toggle_stop else R.string.btn_toggle_start)
        progressBar.visibility = if (isRunning) View.VISIBLE else View.GONE
        btnOpen.isEnabled = isRunning
        btnShareLan.isEnabled = isRunning && getLanIpAddress() != null
        tvStatus.text = getString(if (isRunning) R.string.status_running else R.string.status_stopped)
        tvStatus.setTextColor(if (isRunning) 0xFF008800.toInt() else 0xFFCC0000.toInt())
        if (isRunning) updateLanDisplay()
    }

    private fun appendLog(msg: String) {
        // Add new message to the ring buffer
        logBuffer.addLast(msg)
        if (logBuffer.size > maxLogLines) {
            logBuffer.removeFirst()
        }

        // Rebuild the TextView contents from the bounded buffer
        tvLogs.text = buildString {
            val iterator = logBuffer.iterator()
            while (iterator.hasNext()) {
                append(iterator.next())
                append('\n')
            }
        }

        // Keep the scroll view pinned to the bottom
        scrollLogs.post { scrollLogs.fullScroll(View.FOCUS_DOWN) }
    }

    private fun updateLanDisplay() {
        val ip = getLanIpAddress()
        tvLanUrl.text = if (ip != null) "LAN: http://$ip:$port" else "LAN: unavailable"
    }

    private fun getLanIpAddress(): String? {
        try {
            for (intf in NetworkInterface.getNetworkInterfaces() ?: return null) {
                if (!intf.isUp || intf.isLoopback) continue
                for (addr in intf.inetAddresses)
                    if (addr is Inet4Address && !addr.isLoopbackAddress) return addr.hostAddress
            }
        } catch (e: Exception) { android.util.Log.w("MainActivity", "LAN IP detection failed", e) }
        return null
    }

    override fun onDestroy() { super.onDestroy(); stopService() }
}
