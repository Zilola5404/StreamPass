package com.streampass.app

import android.content.Intent
import android.content.pm.PackageManager
import android.os.Build
import io.flutter.plugin.common.BinaryMessenger
import io.flutter.plugin.common.MethodChannel

/**
 * Lists launchable apps for the Flutter "App bypass" picker and does not
 * require QUERY_ALL_PACKAGES (uses CATEGORY_LAUNCHER).
 */
object InstalledAppsChannel {
    private const val CHANNEL = "streampass/installed_apps"

    fun register(activity: MainActivity, messenger: BinaryMessenger) {
        MethodChannel(messenger, CHANNEL).setMethodCallHandler { call, result ->
            when (call.method) {
                "listLaunchable" -> {
                    try {
                        result.success(listLaunchable(activity))
                    } catch (t: Throwable) {
                        result.error("list_failed", t.message, null)
                    }
                }
                else -> result.notImplemented()
            }
        }
    }

    private fun listLaunchable(activity: MainActivity): List<Map<String, String>> {
        val pm = activity.packageManager
        val intent = Intent(Intent.ACTION_MAIN, null).addCategory(Intent.CATEGORY_LAUNCHER)
        val resolved = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            pm.queryIntentActivities(intent, PackageManager.ResolveInfoFlags.of(0))
        } else {
            @Suppress("DEPRECATION")
            pm.queryIntentActivities(intent, 0)
        }

        val self = activity.packageName
        val out = LinkedHashMap<String, Map<String, String>>()
        for (info in resolved) {
            val pkg = info.activityInfo?.packageName ?: continue
            if (pkg == self) continue
            if (out.containsKey(pkg)) continue
            val label = try {
                info.loadLabel(pm)?.toString() ?: pkg
            } catch (_: Throwable) {
                pkg
            }
            out[pkg] = mapOf(
                "packageName" to pkg,
                "label" to label,
            )
        }
        return out.values.sortedBy { it["label"]?.lowercase() ?: "" }
    }
}
