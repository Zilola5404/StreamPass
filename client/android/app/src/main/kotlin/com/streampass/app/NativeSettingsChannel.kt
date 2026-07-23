package com.streampass.app

import android.content.Context
import io.flutter.plugin.common.MethodCall
import io.flutter.plugin.common.MethodChannel

/**
 * The Flutter `shared_preferences` plugin stores values under an
 * implementation-specific file/format that can change between plugin
 * versions. BootReceiver runs outside of the Flutter engine and needs a
 * stable, native-only source of truth for the two flags it cares about —
 * so settings_service.dart mirrors "autostart" and "autoConnect" here via
 * this dedicated channel, in addition to shared_preferences (which the UI
 * reads back on next launch).
 */
object NativeSettingsChannel {
    const val PREFS_NAME = "streampass_native_prefs"
    const val KEY_AUTOSTART = "autostart"
    const val KEY_AUTO_CONNECT = "auto_connect"

    fun register(context: Context, messenger: io.flutter.plugin.common.BinaryMessenger) {
        MethodChannel(messenger, "streampass/settings").setMethodCallHandler { call, result ->
            when (call.method) {
                "setAutostart" -> {
                    write(context, KEY_AUTOSTART, call.arguments as Boolean)
                    result.success(null)
                }
                "setAutoConnect" -> {
                    write(context, KEY_AUTO_CONNECT, call.arguments as Boolean)
                    result.success(null)
                }
                else -> result.notImplemented()
            }
        }
    }

    private fun write(context: Context, key: String, value: Boolean) {
        context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
            .edit()
            .putBoolean(key, value)
            .apply()
    }
}
