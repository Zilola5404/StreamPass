package com.streampass.app

import android.Manifest
import android.app.Activity
import android.content.Intent
import android.content.pm.PackageManager
import android.net.VpnService
import android.os.Build
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
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
                    else -> result.notImplemented()
                }
            }

        EventChannel(flutterEngine.dartExecutor.binaryMessenger, EVENT_CHANNEL)
            .setStreamHandler(object : EventChannel.StreamHandler {
                override fun onListen(args: Any?, sink: EventChannel.EventSink) {
                    StreamPassVpnService.eventSink = sink
                }

                override fun onCancel(args: Any?) {
                    StreamPassVpnService.eventSink = null
                }
            })

        NativeSettingsChannel.register(this, flutterEngine.dartExecutor.binaryMessenger)
    }

    /**
     * Returns true if the request was accepted for processing (either the
     * service started immediately, or the system permission dialog was
     * shown). The actual "connected" state always arrives later via the
     * event channel — this return value is not a connection guarantee.
     */
    private fun requestConnect(args: Map<*, *>?): Boolean {
        val consentIntent = VpnService.prepare(this)
        if (consentIntent != null) {
            pendingConnect = true
            pendingArgs = args
            startActivityForResult(consentIntent, VPN_PERMISSION_REQUEST)
            return true
        }

        try {
            StreamPassVpnService.start(this, args)
            return true
        } catch (t: Throwable) {
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
            requestConnect(pendingArgs)
        } else {
            StreamPassVpnService.eventSink?.success(
                mapOf("event" to "permissionDenied")
            )
        }
        pendingConnect = false
        pendingArgs = null
    }
}