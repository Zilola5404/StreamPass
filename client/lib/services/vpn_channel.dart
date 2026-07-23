import 'package:flutter/services.dart';

enum VpnEvent { connecting, connected, disconnected, permissionDenied, error }

class VpnStatusUpdate {
  final VpnEvent event;
  final String? relayName;
  final int? pingMs;
  final String? errorMessage;
  VpnStatusUpdate(this.event, {this.relayName, this.pingMs, this.errorMessage});
}

/// Bridges Dart <-> native VpnService.
///
/// Android side flow:
/// 1. `connect()` calls MethodChannel "connect".
/// 2. Native MainActivity checks VpnService.prepare(context):
///    - null  -> permission already granted, start service directly.
///    - Intent -> must be launched via startActivityForResult; result comes
///      back through onActivityResult, which then starts the service.
/// 3. Native service pushes state changes through an EventChannel so the UI
///    reacts to state changes that happen outside of a direct user tap
///    (e.g. relay switch on degradation, per ТЗ section 5).
class VpnChannel {
  static const _method = MethodChannel('streampass/vpn');
  static const _events = EventChannel('streampass/vpn/events');

  static Stream<VpnStatusUpdate>? _statusStream;

  static Stream<VpnStatusUpdate> get statusStream {
    _statusStream ??= _events.receiveBroadcastStream().map((raw) {
      final map = Map<String, dynamic>.from(raw as Map);
      final event = VpnEvent.values.firstWhere(
        (e) => e.name == map['event'],
        orElse: () => VpnEvent.error,
      );
      return VpnStatusUpdate(
        event,
        relayName: map['relay'] as String?,
        pingMs: map['pingMs'] as int?,
        errorMessage: map['error'] as String?,
      );
    });
    return _statusStream!;
  }

  /// Returns true if the connect request was accepted (does not guarantee
  /// tunnel is up yet — listen to [statusStream] for the actual state).
  static Future<bool> connect() async {
    try {
      final ok = await _method.invokeMethod<bool>('connect');
      return ok ?? false;
    } on PlatformException {
      return false;
    }
  }

  static Future<void> disconnect() async {
    try {
      await _method.invokeMethod('disconnect');
    } on PlatformException {
      // no-op — UI treats absence of further events as disconnected
    }
  }
}
