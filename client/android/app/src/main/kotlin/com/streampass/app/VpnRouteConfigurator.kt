package com.streampass.app

import android.net.IpPrefix
import android.net.VpnService
import android.os.Build
import android.util.Log
import java.net.InetAddress

/**
 * Split-tunnel per ТЗ: Russian IPv4 space bypasses TUN (DIRECT on OS level),
 * foreign traffic is captured by VPN for Decision Engine RELAY/DIRECT.
 *
 * Diagnostic modes (network diagnostics TASK):
 * - split: normal RU exclude / intl capture
 * - full_relay / direct_test: full tunnel (0.0.0.0/0) so all flows hit Go
 */
object VpnRouteConfigurator {
    private const val TAG = "StreamPassVpn"
    private const val RU_CIDRS_ASSET = "ru_ipv4_cidrs.txt"
    private const val INTL_ROUTES_ASSET = "intl_ipv4_routes.txt"

    data class Result(val mode: String, val routeCount: Int, val excludeCount: Int)

    fun apply(
        builder: VpnService.Builder,
        assets: android.content.res.AssetManager,
        networkMode: String = "split",
    ): Result {
        val mode = networkMode.lowercase().ifBlank { "split" }
        if (mode == "full_relay" || mode == "direct_test") {
            builder.addRoute("0.0.0.0", 0)
            Log.i(TAG, "full-tunnel mode=$mode (all traffic via TUN)")
            return Result("full-tunnel-$mode", 1, 0)
        }

        val ruCidrs = loadCidrs(assets, RU_CIDRS_ASSET)
        if (ruCidrs.isEmpty()) {
            Log.w(TAG, "RU CIDR list empty — falling back to full tunnel")
            builder.addRoute("0.0.0.0", 0)
            return Result("fallback-full", 1, 0)
        }

        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            builder.addRoute("0.0.0.0", 0)
            var excluded = 0
            for (cidr in ruCidrs) {
                try {
                    val (host, prefix) = parseCidr(cidr)
                    val addr = InetAddress.getByName(host)
                    builder.excludeRoute(IpPrefix(addr, prefix))
                    excluded++
                } catch (t: Throwable) {
                    Log.w(TAG, "excludeRoute failed for $cidr: ${t.message}")
                }
            }
            Log.i(TAG, "split-tunnel API33+ excludes=$excluded")
            Result("exclude-ru", 1, excluded)
        } else {
            val intlRoutes = loadCidrs(assets, INTL_ROUTES_ASSET)
            if (intlRoutes.isEmpty()) {
                Log.w(TAG, "intl routes missing — falling back to full tunnel")
                builder.addRoute("0.0.0.0", 0)
                return Result("fallback-full", 1, 0)
            }
            var added = 0
            for (cidr in intlRoutes) {
                try {
                    val (host, prefix) = parseCidr(cidr)
                    builder.addRoute(host, prefix)
                    added++
                } catch (t: Throwable) {
                    Log.w(TAG, "addRoute failed for $cidr: ${t.message}")
                }
            }
            Log.i(TAG, "split-tunnel legacy intl routes=$added")
            Result("intl-only", added, ruCidrs.size)
        }
    }

    private fun loadCidrs(assets: android.content.res.AssetManager, name: String): List<String> {
        return try {
            assets.open(name).bufferedReader().useLines { lines ->
                lines.map { it.trim() }
                    .filter { it.isNotEmpty() && !it.startsWith("#") }
                    .toList()
            }
        } catch (t: Throwable) {
            Log.e(TAG, "failed to load asset $name: ${t.message}")
            emptyList()
        }
    }

    private fun parseCidr(cidr: String): Pair<String, Int> {
        val parts = cidr.split("/")
        require(parts.size == 2) { "bad cidr: $cidr" }
        return parts[0] to parts[1].toInt()
    }
}
