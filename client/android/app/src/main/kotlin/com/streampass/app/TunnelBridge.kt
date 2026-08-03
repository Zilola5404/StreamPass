package com.streampass.app

import android.content.Context
import android.util.Log
import java.lang.reflect.Proxy

/**
 * Bridge between the Android VPN service and the gomobile-generated Go core.
 * gomobile bind ./mobile produces Java package `mobile` with class `Mobile`.
 */
class TunnelBridge(
    private val context: Context?,
    private val onState: (event: String, relay: String?, pingMs: Int?, error: String?) -> Unit,
) {
    companion object {
        private const val TAG = "StreamPassTunnel"
    }

    private fun coreClass(): Class<*>? = listOf(
        "mobile.Mobile",
        "streampasscore.Streampasscore",
    ).firstNotNullOfOrNull { name ->
        try {
            Class.forName(name)
        } catch (_: ClassNotFoundException) {
            null
        }
    }

    private fun callbackClass(core: Class<*>): Class<*>? {
        val pkg = core.`package`?.name ?: return null
        val simple = core.simpleName
        return listOf(
            "$pkg.$simple\$StatusCallback",
            "$pkg.StatusCallback",
        ).firstNotNullOfOrNull { name ->
            try {
                Class.forName(name)
            } catch (_: ClassNotFoundException) {
                null
            }
        }
    }

    fun stopTunnel() {
        ConnectLogger.log(context, "TunnelBridge.stopTunnel")
        try {
            val core = coreClass() ?: return
            core.getMethod("stopTunnel").invoke(null)
        } catch (_: Throwable) {
            // Best-effort shutdown; VPN service still tears down TUN.
        }
    }

    fun prepareRelay(relayHost: String, relayPort: Int, connectionConfig: String): String? {
        return try {
            val coreClass = coreClass() ?: throw ClassNotFoundException("mobile.Mobile")
            val method = coreClass.getMethod(
                "prepareRelay",
                String::class.java,
                Long::class.javaPrimitiveType,
                String::class.java,
            )
            val err = method.invoke(null, relayHost, relayPort.toLong(), connectionConfig) as String
            ConnectLogger.log(context, "PrepareRelay result=${if (err.isEmpty()) "OK" else err}")
            if (err.isEmpty()) null else err
        } catch (t: Throwable) {
            Log.e(TAG, "PrepareRelay failed", t)
            ConnectLogger.log(context, "PrepareRelay failed: ${t.message}")
            t.message ?: "PrepareRelay failed"
        }
    }

    fun startTunnel(fd: Int, relayHost: String, relayPort: Int, connectionConfig: String) {
        try {
            val coreClass = coreClass()
                ?: throw ClassNotFoundException("mobile.Mobile")
            val callbackClass = callbackClass(coreClass)
                ?: throw ClassNotFoundException("StatusCallback")

            val callback = Proxy.newProxyInstance(
                callbackClass.classLoader,
                arrayOf(callbackClass),
            ) { _, method, args ->
                when (method.name) {
                    "onConnecting" -> onState("connecting", null, null, null)
                    "onConnected" -> {
                        val relay = args?.getOrNull(0)?.toString().orEmpty()
                        val pingMs = args?.getOrNull(1)?.toString()?.toIntOrNull()
                        onState("connected", relay.ifBlank { null }, pingMs, null)
                    }
                    "onDisconnected" -> onState("disconnected", null, null, null)
                    "onError" -> onState("error", null, null, args?.getOrNull(0)?.toString())
                }
                null
            }

            val startMethod = coreClass.getMethod(
                "startTunnel",
                Long::class.javaPrimitiveType,
                String::class.java,
                Long::class.javaPrimitiveType,
                String::class.java,
                callbackClass,
            )
            Log.i(TAG, "StartTunnel fd=$fd host=$relayHost port=$relayPort configLen=${connectionConfig.length}")
            ConnectLogger.log(context, "StartTunnel fd=$fd host=$relayHost port=$relayPort")
            startMethod.invoke(null, fd.toLong(), relayHost, relayPort.toLong(), connectionConfig, callback)
        } catch (_: ClassNotFoundException) {
            Log.e(TAG, "gomobile AAR missing (mobile.Mobile)")
            ConnectLogger.log(context, "ERROR: gomobile AAR missing (mobile.Mobile)")
            onState(
                "error",
                null,
                null,
                "Туннельный модуль Go core недоступен. Сначала соберите streampasscore.aar и поместите его в android/app/libs/.",
            )
        } catch (t: Throwable) {
            Log.e(TAG, "StartTunnel failed", t)
            ConnectLogger.log(context, "StartTunnel failed: ${t.message}")
            onState("error", null, null, t.message ?: "Не удалось запустить туннель")
        }
    }
}
