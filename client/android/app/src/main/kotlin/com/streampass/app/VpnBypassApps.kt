package com.streampass.app

import android.content.pm.PackageManager
import android.net.VpnService
import android.util.Log

/**
 * Russian gov/bank/airline apps that must bypass the TUN interface entirely.
 * Domain DIRECT rules are not enough — while VpnService captures 0.0.0.0/0,
 * these apps still see TRANSPORT_VPN and block login (Gosuslugi, FNS, S7…).
 */
object VpnBypassApps {
    private const val TAG = "StreamPassVpn"

    /** Installed package names excluded from the VPN (split-tunnel at OS level). */
    val packages: List<String> = listOf(
        // Gosuslugi
        "ru.gosuslugi.gosuslugi",
        "ru.gosuslugi.culture",
        // FNS / My Taxes
        "ru.nalog.lkfl",
        "ru.nalog.lkip",
        "ru.fns.lkfl",
        "ru.gnivts.selfemployed",
        // Airlines
        "ru.s7.android",
        // Major banks (TZ direct list)
        "ru.sberbankmobile",
        "com.idamob.tinkoff.android",
        "ru.vtb24.mobilebanking",
        "ru.alfabank.mobile.android",
        "ru.raiffeisennews",
        "ru.mw",
    )

    fun apply(builder: VpnService.Builder, pm: PackageManager): Int {
        var applied = 0
        for (pkg in packages) {
            try {
                pm.getPackageInfo(pkg, 0)
                builder.addDisallowedApplication(pkg)
                applied++
                Log.i(TAG, "VPN bypass: $pkg")
            } catch (_: PackageManager.NameNotFoundException) {
                // App not installed on this device — skip.
            } catch (t: Throwable) {
                Log.w(TAG, "VPN bypass failed for $pkg: ${t.message}")
            }
        }
        return applied
    }
}
