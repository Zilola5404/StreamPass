import 'package:streampass/services/connection_log.dart';

/// Classifies native go_core / VPN log lines for UI severity.
///
/// Critical: the substring `error=` alone must NOT force ERROR — successful
/// diag lines always include `error=` (often empty). Only a non-empty
/// `error=` value, or explicit fail/timeout/stream_open_no_data markers,
/// are treated as errors.
ConnectionLogLevel classifyNativeVpnLog(String line) {
  final lower = line.toLowerCase();

  final errField = RegExp(r'(?:^|\s)error=([^\s]*)').firstMatch(lower);
  if (errField != null) {
    final v = errField.group(1) ?? '';
    if (v.isNotEmpty && v != '-' && v != 'null') {
      return ConnectionLogLevel.error;
    }
  }

  if (lower.contains('result=fail') ||
      lower.contains('result=timeout') ||
      lower.contains('result=stream_failed') ||
      lower.contains('result=stream_open_no_data') ||
      lower.contains('result=dns_fail') ||
      lower.contains('stream_open_no_data') ||
      lower.contains('relay_blackhole') ||
      lower.contains('quic_no_response') ||
      lower.contains('must-relay fail') ||
      lower.contains('must-relay-udp fail') ||
      RegExp(r'\[tun\].*\bfail\b').hasMatch(lower) ||
      RegExp(r'\[dns\].*\bfail\b').hasMatch(lower)) {
    return ConnectionLogLevel.error;
  }

  if (lower.contains('result=slow') ||
      lower.contains('result=stream_closed_partial') ||
      lower.contains('[warn]') ||
      lower.contains(' fallback') ||
      lower.contains('retry') ||
      lower.contains('blackhole')) {
    return ConnectionLogLevel.warn;
  }

  // DEBUG-ish noise stays INFO in the connection log (no separate debug sink).
  return ConnectionLogLevel.info;
}
