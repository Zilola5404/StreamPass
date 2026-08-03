package com.streampass.app

import android.content.Context
import android.util.Log
import java.io.File
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import java.util.TimeZone

/**
 * Append-only connect log on device (filesDir/connect.log) + logcat.
 * Flutter reads the file via MethodChannel for the diagnostics screen.
 */
object ConnectLogger {
    private const val TAG = "StreamPassConnect"
    private const val FILE_NAME = "connect.log"
    private const val MAX_BYTES = 256 * 1024

    private val tsFormat = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss.SSS'Z'", Locale.US).apply {
        timeZone = TimeZone.getTimeZone("UTC")
    }

    fun log(context: Context?, message: String) {
        val line = "${tsFormat.format(Date())} $message"
        Log.i(TAG, message)
        if (context == null) return
        try {
            val file = File(context.filesDir, FILE_NAME)
            file.appendText("$line\n")
            trimIfNeeded(file)
        } catch (t: Throwable) {
            Log.w(TAG, "failed to write connect.log", t)
        }
    }

    fun clear(context: Context?) {
        if (context == null) return
        try {
            File(context.filesDir, FILE_NAME).writeText("")
        } catch (_: Throwable) {
        }
    }

    fun readTail(context: Context?, maxLines: Int = 300): List<String> {
        if (context == null) return emptyList()
        return try {
            val file = File(context.filesDir, FILE_NAME)
            if (!file.exists()) return emptyList()
            file.readLines().takeLast(maxLines)
        } catch (_: Throwable) {
            emptyList()
        }
    }

    private fun trimIfNeeded(file: File) {
        if (file.length() <= MAX_BYTES) return
        val lines = file.readLines()
        val keep = lines.takeLast(400)
        file.writeText(keep.joinToString("\n", postfix = "\n"))
    }
}
