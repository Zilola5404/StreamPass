package com.streampass.app

import android.content.Context
import io.flutter.plugin.common.BinaryMessenger
import io.flutter.plugin.common.MethodChannel

object DiagnosticsChannel {
    private const val CHANNEL = "streampass/diagnostics"

    fun register(context: Context, messenger: BinaryMessenger) {
        MethodChannel(messenger, CHANNEL).setMethodCallHandler { call, result ->
            when (call.method) {
                "readConnectLog" -> {
                    val maxLines = (call.argument<Int>("maxLines") ?: 300).coerceIn(1, 1000)
                    result.success(ConnectLogger.readTail(context, maxLines))
                }
                "clearConnectLog" -> {
                    ConnectLogger.clear(context)
                    result.success(null)
                }
                else -> result.notImplemented()
            }
        }
    }
}
