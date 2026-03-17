package com.goldbox.rpg

import android.content.Intent
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
    private lateinit var btnOpen: Button
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
        progressBar = findViewById(R.id.progressBar)
        tvStatus = findViewById(R.id.tvStatus)
        tvUrl = findViewById(R.id.tvUrl)
        tvLanUrl = findViewById(R.id.tvLanUrl)
        tvLogs = findViewById(R.id.tvLogs)
        scrollLogs = findViewById(R.id.scrollLogs)
        tvUrl.text = "Local: http://127.0.0.1:$port"
        updateLanDisplay()
        copyBinaryFromAssets()
        btnToggle.setOnClickListener { if (isRunning) stopService() else startService() }
        btnOpen.setOnClickListener {
            startActivity(Intent(Intent.ACTION_VIEW, Uri.parse("http://127.0.0.1:$port")))
        }
    }

    private fun copyBinaryFromAssets() {
        val bin = File(filesDir, "webservice")
        if (bin.exists()) return
        try {
            assets.open("webservice").use { i -> bin.outputStream().use { o -> i.copyTo(o) } }
            bin.setExecutable(true)
            appendLog("Binary extracted to ${bin.absolutePath}")
        } catch (e: Exception) {
            appendLog("ERROR: Failed to extract binary: ${e.message}")
        }
    }

    private fun startService() {
        val binary = File(filesDir, "webservice")
        if (!binary.exists() || !binary.canExecute()) {
            appendLog("ERROR: Service binary not found or not executable"); return
        }
        try {
            val pb = ProcessBuilder(binary.absolutePath)
            pb.redirectErrorStream(true)
            pb.environment()["HOME"] = filesDir.absolutePath
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
        serviceProcess?.let { p ->
            p.destroy()
            try { p.waitFor() } catch (_: InterruptedException) { p.destroyForcibly() }
        }
        serviceProcess = null
        isRunning = false; updateUI(); appendLog("Service stopped.")
    }

    private fun readProcessOutput() {
        Thread {
            try {
                val reader = BufferedReader(InputStreamReader(serviceProcess?.inputStream ?: return@Thread))
                var line: String?
                while (reader.readLine().also { line = it } != null) {
                    val text = line ?: continue
                    handler.post { appendLog(text) }
                }
            } catch (_: Exception) { }
            handler.post {
                if (isRunning) { isRunning = false; updateUI(); appendLog("Service process exited.") }
            }
        }.start()
    }

    private fun updateUI() {
        btnToggle.text = if (isRunning) "Stop Service" else "Start Service"
        progressBar.visibility = if (isRunning) View.VISIBLE else View.GONE
        btnOpen.isEnabled = isRunning
        tvStatus.text = if (isRunning) "Running" else "Stopped"
        tvStatus.setTextColor(if (isRunning) 0xFF008800.toInt() else 0xFFCC0000.toInt())
        if (isRunning) updateLanDisplay()
    }

    private fun appendLog(msg: String) {
        tvLogs.append("$msg\n")
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
        } catch (_: Exception) { }
        return null
    }

    override fun onDestroy() { super.onDestroy(); stopService() }
}
