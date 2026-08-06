package com.streampass.app

import android.content.pm.PackageManager
import android.net.VpnService
import android.os.Build
import android.util.Log

/**
 * OS-level app bypass for Russian gov/bank/airline apps.
 *
 * Domain DIRECT / RU CIDR excludeRoute only keep packets off the relay.
 * Apps that call ConnectivityManager.hasTransport(TRANSPORT_VPN) still
 * see the VPN profile unless they are excluded via
 * [VpnService.Builder.addDisallowedApplication].
 *
 * KindAPP on the backend is optional later; MVP discovers packages on-device.
 */
object VpnBypassApps {
    private const val TAG = "StreamPassVpn"

    /** Verified Play Store / RuStore package ids (keep even if heuristics miss). */
    private val knownPackages: List<String> = listOf(
        // Госуслуги (official Play id=ru.rostel)
        "ru.rostel",
        "ru.gosuslugi.gosuslugi",
        "ru.gosuslugi.pos",
        "ru.gosuslugi.culture",
        "ru.gosuslugi.house",
        // ФНС / налоги
        "ru.fns.lkfl",
        "ru.nalog.lkfl",
        "ru.nalog.lkip",
        "com.gnivts.selfemployed",
        "ru.gnivts.selfemployed",
        // S7
        "ru.s7tl.app",
        "ru.s7.android",
        "com.s7.android",
        // Banks
        "ru.sberbankmobile",
        "com.idamob.tinkoff.android",
        "ru.vtb24.mobilebanking",
        "ru.alfabank.mobile.android",
        "ru.raiffeisennews",
        "ru.mw",
        "ru.gazprombank.android.mobilebank.app",
        "logo.com.mbanking",
        "ru.rshb.dbo",
        "ru.rosbank.android.inbank",
        "ru.otpbank.mobile",
        "ru.sovcombank.halva",
        "ru.yoo.money",
        "ru.yoomoney.app",
        "ru.mts.money",
        "ru.megafon.mlk",
        "ru.beeline.services",
        "ru.tele2.mytele2",
        // Airlines / travel RU
        "ru.aeroflot",
        "ru.utair.android",
        "ru.rzd.pass",
        "ru.russianpost.android",
    )

    /** Package / label fragments → treat as RU critical app (bypass VPN). */
    private val heuristics: List<String> = listOf(
        "gosuslugi",
        "rostel",
        "nalog",
        "fns",
        "gnivts",
        "selfemployed",
        "sberbank",
        "tinkoff",
        "tbank",
        "vtb",
        "alfabank",
        "raiffeisen",
        "gazprombank",
        "rshb",
        "rosbank",
        "s7",
        "aeroflot",
        "utair",
        "rzd",
        "pochta",
        "russianpost",
        "yoomoney",
        "yookassa",
    )

    fun apply(
        builder: VpnService.Builder,
        pm: PackageManager,
        ownPackage: String,
        extraPackages: Collection<String> = emptyList(),
    ): Int {
        var applied = 0
        // StreamPass REST API must not hairpin through its own TUN/Hysteria loop.
        if (ownPackage.isNotBlank()) {
            try {
                builder.addDisallowedApplication(ownPackage)
                applied++
                Log.i(TAG, "VPN app-bypass (self): $ownPackage")
            } catch (t: Throwable) {
                Log.w(TAG, "VPN self-bypass failed: ${t.message}")
            }
        }

        val selected = linkedSetOf<String>()
        selected.addAll(knownPackages)
        selected.addAll(extraPackages.map { it.trim() }.filter { it.isNotEmpty() })
        selected.addAll(discoverHeuristicPackages(pm))

        for (pkg in selected) {
            if (!isInstalled(pm, pkg)) continue
            try {
                builder.addDisallowedApplication(pkg)
                applied++
                Log.i(TAG, "VPN app-bypass: $pkg")
            } catch (t: Throwable) {
                Log.w(TAG, "VPN app-bypass failed for $pkg: ${t.message}")
            }
        }
        Log.i(TAG, "VPN app-bypass applied=$applied candidates=${selected.size}")
        return applied
    }

    private fun discoverHeuristicPackages(pm: PackageManager): Set<String> {
        val out = linkedSetOf<String>()
        val apps = try {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                pm.getInstalledApplications(PackageManager.ApplicationInfoFlags.of(0))
            } else {
                @Suppress("DEPRECATION")
                pm.getInstalledApplications(0)
            }
        } catch (t: Throwable) {
            Log.w(TAG, "list installed apps failed: ${t.message}")
            return out
        }

        for (app in apps) {
            // Do NOT skip FLAG_SYSTEM — many banks/gov apps ship as updated-system
            // apps (FLAG_SYSTEM|FLAG_UPDATED_SYSTEM_APP) and would never be bypassed.
            val pkg = app.packageName.lowercase()
            if (pkg.startsWith("com.android.") ||
                pkg.startsWith("android.") ||
                pkg.startsWith("com.google.android.")
            ) {
                continue
            }
            val label = try {
                pm.getApplicationLabel(app).toString().lowercase()
            } catch (_: Throwable) {
                ""
            }
            if (heuristics.any { pkg.contains(it) || label.contains(it) }) {
                out.add(app.packageName)
            }
        }
        return out
    }

    private fun isInstalled(pm: PackageManager, pkg: String): Boolean {
        return try {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                pm.getPackageInfo(pkg, PackageManager.PackageInfoFlags.of(0))
            } else {
                @Suppress("DEPRECATION")
                pm.getPackageInfo(pkg, 0)
            }
            true
        } catch (_: PackageManager.NameNotFoundException) {
            false
        }
    }
}
