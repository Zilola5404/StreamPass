package com.streampass.app

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent

/**
 * Reads the "Автозапуск" flag (ТЗ §20) written natively via
 * NativeSettingsChannel and starts the VPN service on boot if enabled.
 *
 * Note for Android 14+ (API 34): BOOT_COMPLETED is one of the few contexts
 * still allowed to start a foreground service from the background, but the
 * service's declared foregroundServiceType must match what's permitted for
 * that trigger. "specialUse" requires a justification string submitted via
 * Play Console's Permissions Declaration Form before release — do this
 * before submitting for review, not after a rejection.
 */
class BootReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action != Intent.ACTION_BOOT_COMPLETED) return

        val prefs = context.getSharedPreferences(
            NativeSettingsChannel.PREFS_NAME,
            Context.MODE_PRIVATE,
        )
        val autostartEnabled = prefs.getBoolean(NativeSettingsChannel.KEY_AUTOSTART, false)
        val autoConnectEnabled = prefs.getBoolean(NativeSettingsChannel.KEY_AUTO_CONNECT, false)

        if (autostartEnabled && autoConnectEnabled) {
            StreamPassVpnService.start(context)
        }
        // If autostart is on but auto-connect is off, we deliberately do NOT
        // start the tunnel — autostart alone should only mean "have the app
        // ready", not "connect without the user's current session choice".
    }
}
