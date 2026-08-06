package com.streampass.app

import android.Manifest
import android.app.Activity
import android.content.Intent
import android.content.pm.PackageManager
import android.net.VpnService
import android.os.Build
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import android.util.Log
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.EventChannel
import io.flutter.plugin.common.MethodChannel

class MainActivity : FlutterActivity() {

    private val METHOD_CHANNEL = "streampass/vpn"
    private val EVENT_CHANNEL = "streampass/vpn/events"
    private val VPN_PERMISSION_REQUEST = 0x51
    private val NOTIFICATION_PERMISSION_REQUEST = 0x52

    // Held while we wait for the user to respond to the system VPN consent
    // dialog, so we know to actually start the service once permission
    // is granted.
    private var pendingConnect = false
    private var pendingArgs: Map<*, *>? = null

    override fun onCreate(savedInstanceState: android.os.Bundle?) {
        super.onCreate(savedInstanceState)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU &&
            ContextCompat.checkSelfPermission(this, Manifest.permission.POST_NOTIFICATIONS)
                != PackageManager.PERMISSION_GRANTED
        ) {
            ActivityCompat.requestPermissions(
                this,
                arrayOf(Manifest.permission.POST_NOTIFICATIONS),
                NOTIFICATION_PERMISSION_REQUEST,
            )
        }
    }

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)

        MethodChannel(flutterEngine.dartExecutor.binaryMessenger, METHOD_CHANNEL)
            .setMethodCallHandler { call, result ->
                when (call.method) {
                    "connect" -> {
                        result.success(requestConnect(call.arguments as? Map<*, *>))
                    }
                    "disconnect" -> {
                        StreamPassVpnService.stop(this)
                        result.success(null)
                    }
                    "getStatus" -> {
                        result.success(StreamPassVpnService.statusSnapshot())
                    }
                    "updateRules" -> {
                        result.success(updateRoutingRules(call.arguments as? Map<*, *>))
                    }
                    else -> result.notImplemented()
                }
            }

        EventChannel(flutterEngine.dartExecutor.binaryMessenger, EVENT_CHANNEL)
            .setStreamHandler(object : EventChannel.StreamHandler {
                override fun onListen(args: Any?, sink: EventChannel.EventSink) {
                    StreamPassVpnService.eventSink = sink
                }

                override fun onCancel(args: Any?) {
                    // Keep eventSink so native events still reach Flutter after brief
                    // detach (tab switch / rotation). onListen replaces sink on resume.
                }
            })

        NativeSettingsChannel.register(this, flutterEngine.dartExecutor.binaryMessenger)
        DiagnosticsChannel.register(this, flutterEngine.dartExecutor.binaryMessenger)
        InstalledAppsChannel.register(this, flutterEngine.dartExecutor.binaryMessenger)
    }

    /**
     * Returns true if the request was accepted for processing (either the
     * service started immediately, or the system permission dialog was
     * shown). The actual "connected" state always arrives later via the
     * event channel — this return value is not a connection guarantee.
     */
    private fun requestConnect(args: Map<*, *>?): Boolean {
        val relayId = args?.get("id") as? String ?: ""
        val host = args?.get("host") as? String ?: ""
        ConnectLogger.log(this, "MainActivity.requestConnect relay=$relayId host=$host")
        val consentIntent = VpnService.prepare(this)
        if (consentIntent != null) {
            ConnectLogger.log(this, "VPN permission dialog shown")
            pendingConnect = true
            pendingArgs = args
            startActivityForResult(consentIntent, VPN_PERMISSION_REQUEST)
            return true
        }

        try {
            ConnectLogger.log(this, "VPN permission already granted, starting service")
            StreamPassVpnService.start(this, args)
            return true
        } catch (t: Throwable) {
            ConnectLogger.log(this, "start service failed: ${t.message}")
            Log.e("StreamPassVpn", "start failed", t)
            StreamPassVpnService.eventSink?.success(
                mapOf(
                    "event" to "error",
                    "error" to (t.message ?: "Не удалось запустить VPN-сервис"),
                )
            )
            return false
        }
    }

    override fun onActivityResult(requestCode: Int, resultCode: Int, data: Intent?) {
        super.onActivityResult(requestCode, resultCode, data)
        if (requestCode != VPN_PERMISSION_REQUEST) return

        if (resultCode == Activity.RESULT_OK && pendingConnect) {
            ConnectLogger.log(this, "VPN permission granted")
            requestConnect(pendingArgs)
        } else {
            ConnectLogger.log(this, "VPN permission denied result=$resultCode")
            StreamPassVpnService.eventSink?.success(
                mapOf("event" to "permissionDenied")
            )
        }
        pendingConnect = false
        pendingArgs = null
    }

    private fun updateRoutingRules(args: Map<*, *>?): String {
        val rulesJson = args?.get("rulesJson") as? String ?: ""
        val exclusionsJson = args?.get("exclusionsJson") as? String ?: "[]"
        ConnectLogger.log(this, "MainActivity.updateRules rulesLen=${rulesJson.length}")
        return StreamPassVpnService.updateRules(rulesJson, exclusionsJson)
    }
}