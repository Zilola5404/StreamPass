import 'package:flutter/services.dart';

import 'connection_log.dart';

/// Reads native connect.log written by Android ConnectLogger.
class NativeConnectLog {
  static const _channel = MethodChannel('streampass/diagnostics');
  static final _log = ConnectionLog.instance;

  static Future<void> pullFromNative() async {
    try {
      final lines = await _channel.invokeMethod<List<dynamic>>('readConnectLog', {
        'maxLines': 300,
      });
      if (lines == null || lines.isEmpty) return;
      _log.importNativeLines(lines.cast<String>());
    } on PlatformException catch (e) {
      _log.warn('native-log', 'readConnectLog failed', {'error': e.message ?? 'unknown'});
    } on MissingPluginException {
      // Desktop / iOS stub — ignore.
    }
  }

  static Future<void> clearNative() async {
    try {
      await _channel.invokeMethod('clearConnectLog');
    } on PlatformException {
      // ignore
    } on MissingPluginException {
      // ignore
    }
  }
}
