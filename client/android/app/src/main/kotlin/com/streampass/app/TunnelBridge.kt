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

    fun setSocketProtector(protectFn: (Int) -> Boolean) {
        ConnectLogger.log(context, "TunnelBridge.setSocketProtector")
        try {
            val core = coreClass() ?: return
            val protectorClass = callbackClassNamed(core, "SocketProtector")
                ?: throw ClassNotFoundException("SocketProtector")
            val proxy = Proxy.newProxyInstance(
                protectorClass.classLoader,
                arrayOf(protectorClass),
            ) { _, method, args ->
                when (method.name) {
                    "protect" -> {
                        val fd = (args?.getOrNull(0) as? Number)?.toInt() ?: -1
                        protectFn(fd)
                    }
                    else -> null
                }
            }
            core.getMethod("setSocketProtector", protectorClass).invoke(null, proxy)
        } catch (t: Throwable) {
            Log.e(TAG, "setSocketProtector failed", t)
            ConnectLogger.log(context, "setSocketProtector failed: ${t.message}")
        }
    }

    fun setEventLogger(logFn: (String) -> Unit) {
        try {
            val core = coreClass() ?: return
            val loggerClass = callbackClassNamed(core, "EventLogger")
                ?: throw ClassNotFoundException("EventLogger")
            val proxy = Proxy.newProxyInstance(
                loggerClass.classLoader,
                arrayOf(loggerClass),
            ) { _, method, args ->
                when (method.name) {
                    "log" -> {
                        val msg = args?.getOrNull(0)?.toString().orEmpty()
                        if (msg.isNotEmpty()) logFn(msg)
                    }
                    else -> null
                }
                null
            }
            core.getMethod("setEventLogger", loggerClass).invoke(null, proxy)
            ConnectLogger.log(context, "TunnelBridge.setEventLogger")
        } catch (t: Throwable) {
            Log.e(TAG, "setEventLogger failed", t)
            ConnectLogger.log(context, "setEventLogger failed: ${t.message}")
        }
    }

    fun clearSocketProtector() {
        try {
            val core = coreClass() ?: return
            val protectorClass = callbackClassNamed(core, "SocketProtector") ?: return
            // Explicit null argument array avoids Kotlin/Java overload ambiguity.
            protectorClass.let { cls ->
                core.getMethod("setSocketProtector", cls).invoke(null, *arrayOf<Any?>(null))
            }
            ConnectLogger.log(context, "TunnelBridge.clearSocketProtector")
        } catch (t: Throwable) {
            Log.e(TAG, "clearSocketProtector failed", t)
        }
    }

    private fun callbackClassNamed(core: Class<*>, simpleName: String): Class<*>? {
        val pkg = core.`package`?.name ?: return null
        val coreSimple = core.simpleName
        return listOf(
            "$pkg.$coreSimple\$$simpleName",
            "$pkg.$simpleName",
        ).firstNotNullOfOrNull { name ->
            try {
                Class.forName(name)
            } catch (_: ClassNotFoundException) {
                null
            }
        }
    }

    private fun callbackClass(core: Class<*>): Class<*>? = callbackClassNamed(core, "StatusCallback")

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

    fun startTunnel(
        fd: Int,
        relayHost: String,
        relayPort: Int,
        connectionConfig: String,
        rulesJson: String,
        exclusionsJson: String,
        optionsJson: String = "",
    ) {
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

            val startMethod = findStartTunnelMethod(coreClass, callbackClass)
            Log.i(TAG, "StartTunnel fd=$fd host=$relayHost port=$relayPort configLen=${connectionConfig.length} rulesLen=${rulesJson.length} optionsLen=${optionsJson.length}")
            ConnectLogger.log(context, "StartTunnel fd=$fd host=$relayHost port=$relayPort rulesLen=${rulesJson.length} options=$optionsJson")
            when (startMethod.parameterCount) {
                8 -> startMethod.invoke(
                    null,
                    fd.toLong(),
                    relayHost,
                    relayPort.toLong(),
                    connectionConfig,
                    rulesJson,
                    exclusionsJson,
                    optionsJson,
                    callback,
                )
                7 -> startMethod.invoke(
                    null,
                    fd.toLong(),
                    relayHost,
                    relayPort.toLong(),
                    connectionConfig,
                    rulesJson,
                    exclusionsJson,
                    callback,
                )
                else -> startMethod.invoke(
                    null,
                    fd.toLong(),
                    relayHost,
                    relayPort.toLong(),
                    connectionConfig,
                    callback,
                )
            }
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

    private fun findStartTunnelMethod(coreClass: Class<*>, callbackClass: Class<*>): java.lang.reflect.Method {
        return try {
            coreClass.getMethod(
                "startTunnel",
                Long::class.javaPrimitiveType,
                String::class.java,
                Long::class.javaPrimitiveType,
                String::class.java,
                String::class.java,
                String::class.java,
                String::class.java,
                callbackClass,
            )
        } catch (_: NoSuchMethodException) {
            try {
                coreClass.getMethod(
                    "startTunnel",
                    Long::class.javaPrimitiveType,
                    String::class.java,
                    Long::class.javaPrimitiveType,
                    String::class.java,
                    String::class.java,
                    String::class.java,
                    callbackClass,
                )
            } catch (_: NoSuchMethodException) {
                coreClass.getMethod(
                    "startTunnel",
                    Long::class.javaPrimitiveType,
                    String::class.java,
                    Long::class.javaPrimitiveType,
                    String::class.java,
                    callbackClass,
                )
            }
        }
    }

    fun updateRules(rulesJson: String, exclusionsJson: String): String {
        return try {
            val coreClass = coreClass() ?: return "gomobile AAR missing"
            val method = coreClass.getMethod(
                "updateRules",
                String::class.java,
                String::class.java,
            )
            val err = method.invoke(null, rulesJson, exclusionsJson) as String
            ConnectLogger.log(context, "UpdateRules result=${if (err.isEmpty()) "OK" else err}")
            err
        } catch (t: Throwable) {
            Log.e(TAG, "UpdateRules failed", t)
            ConnectLogger.log(context, "UpdateRules failed: ${t.message}")
            t.message ?: "UpdateRules failed"
        }
    }
}
