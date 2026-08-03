package com.streampass.app

import java.lang.reflect.Proxy

/**
 * Bridge between the Android VPN service and a future gomobile-generated
 * Go-core binding. This keeps the service code clean and lets the app
 * use a real tunnel implementation once streampasscore.aar is built and
 * placed in android/app/libs/.
 */
class TunnelBridge(
    private val onState: (event: String, relay: String?, pingMs: Int?, error: String?) -> Unit,
) {
    fun stopTunnel() {
        try {
            val coreClass = Class.forName("streampasscore.Streampasscore")
            coreClass.getMethod("stopTunnel").invoke(null)
        } catch (_: ClassNotFoundException) {
            // Go core not packaged — nothing to stop.
        } catch (_: Throwable) {
            // Best-effort shutdown; VPN service still tears down TUN.
        }
    }

    fun startTunnel(fd: Int, relayHost: String, relayPort: Int, connectionConfig: String) {
        try {
            val coreClass = Class.forName("streampasscore.Streampasscore")
            val callbackClass = Class.forName("streampasscore.Streampasscore\$StatusCallback")
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
            startMethod.invoke(null, fd.toLong(), relayHost, relayPort.toLong(), connectionConfig, callback)
        } catch (_: ClassNotFoundException) {
            onState(
                "error",
                null,
                null,
                "Туннельный модуль Go core недоступен. Сначала соберите streampasscore.aar и поместите его в android/app/libs/.",
            )
        } catch (t: Throwable) {
            onState("error", null, null, t.message ?: "Не удалось запустить туннель")
        }
    }
}
