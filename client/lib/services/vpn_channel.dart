import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'dart:async';
import 'connection_log.dart';
import 'streampass_api.dart';

enum VpnEvent { connecting, connected, disconnected, permissionDenied, error }

class VpnStatusUpdate {
  final VpnEvent event;
  final String? relayName;
  final int? pingMs;
  final String? errorMessage;
  VpnStatusUpdate(this.event, {this.relayName, this.pingMs, this.errorMessage});
}

/// Thrown when the native VPN layer rejects the connect request.
class VpnConnectException implements Exception {
  final String message;
  VpnConnectException(this.message);

  @override
  String toString() => message;
}

/// Bridges Dart <-> native VpnService.
///
/// Android side flow:
/// 1. `connect(server)` calls MethodChannel "connect" with the real relay's
///    id/host/port/connection config — the native side no longer hardcodes
///    a fake relay IP, it uses whatever GET /servers actually returned.
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
  static final _log = ConnectionLog.instance;

  static Stream<VpnStatusUpdate>? _statusStream;
  static StreamSubscription<VpnStatusUpdate>? _keepalive;

  /// Last VPN event seen by Flutter (Diagnostics can show this without waiting
  /// for a new EventChannel push).
  static VpnStatusUpdate? lastStatus;

  /// Ensure the EventChannel is subscribed early so [eventSink] is set before
  /// native connect completes.
  static void ensureListening() {
    _keepalive ??= statusStream.listen((_) {});
  }

  static VpnStatusUpdate? _parseStatusMap(Map<dynamic, dynamic>? raw) {
    if (raw == null) return null;
    final map = Map<String, dynamic>.from(raw);
    final eventName = map['event'] as String? ?? 'disconnected';
    final event = VpnEvent.values.firstWhere(
      (e) => e.name == eventName,
      orElse: () => VpnEvent.disconnected,
    );
    return VpnStatusUpdate(
      event,
      relayName: map['relay'] as String?,
      pingMs: map['pingMs'] as int?,
      errorMessage: map['error'] as String?,
    );
  }

  /// Query native VpnService for the real current status (AUDIT-003 BUG-004).
  static Future<VpnStatusUpdate?> fetchNativeStatus() async {
    try {
      final raw = await _method.invokeMethod<Map<dynamic, dynamic>>('getStatus');
      final update = _parseStatusMap(raw);
      if (update != null) {
        lastStatus = update;
      }
      return update;
    } on PlatformException {
      return lastStatus;
    } on MissingPluginException {
      return lastStatus;
    }
  }

  /// Test helper: record a status as if it arrived from the EventChannel.
  @visibleForTesting
  static void debugSetLastStatus(VpnStatusUpdate? update) {
    lastStatus = update;
  }

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
    }).map((update) {
      lastStatus = update;
      _logVpnEvent(update);
      return update;
    }).asBroadcastStream();
    return _statusStream!;
  }

  static void _logVpnEvent(VpnStatusUpdate update) {
    final details = <String, String>{
      'event': update.event.name,
      if (update.relayName != null) 'relay': update.relayName!,
      if (update.pingMs != null) 'pingMs': '${update.pingMs}',
      if (update.errorMessage != null) 'error': update.errorMessage!,
    };
    if (update.event == VpnEvent.error) {
      _log.error('vpn', 'native VPN event', details);
    } else {
      _log.info('vpn', 'native VPN event', details);
    }
  }

  /// Returns true if the connect request was accepted (does not guarantee
  /// tunnel is up yet — listen to [statusStream] for the actual state).
  static Future<bool> connect(
    RelayServer server, {
    String rulesJson = '',
    String exclusionsJson = '',
    String bypassPackagesJson = '[]',
  }) async {
    ensureListening();
    if (server.connectionConfig.isEmpty) {
      _log.error('vpn', 'connect blocked: empty connection_config', {'relayId': server.id});
      throw VpnConnectException(
        'У relay нет connection_config. Проверьте настройки сервера в backend.',
      );
    }
    _log.beginConnectSession(relayId: server.id, host: server.host);
    _log.info('vpn', 'MethodChannel connect', {
      'relayId': server.id,
      'host': server.host,
      'port': '${server.port}',
      'configScheme': server.connectionConfig.split(':').first,
    });
    try {
      final ok = await _method.invokeMethod<bool>('connect', {
        'id': server.id,
        'host': server.host,
        'port': server.port,
        'displayName': server.region,
        'connectionConfig': server.connectionConfig,
        'rulesJson': rulesJson,
        'exclusionsJson': exclusionsJson,
        'bypassPackagesJson': bypassPackagesJson,
      });
      _log.info('vpn', 'MethodChannel connect accepted', {'ok': '${ok ?? false}'});
      return ok ?? false;
    } on PlatformException catch (e) {
      _log.error('vpn', 'MethodChannel PlatformException', {'message': e.message ?? 'unknown'});
      throw VpnConnectException(e.message ?? 'Не удалось запустить VPN');
    }
  }

  static Future<void> disconnect() async {
    _log.info('vpn', 'disconnect requested');
    try {
      await _method.invokeMethod('disconnect');
    } on PlatformException {
      // no-op — UI treats absence of further events as disconnected
    }
  }

  /// Hot-reload routing rules on the active tunnel (BL-006).
  static Future<String?> updateRules({
    required String rulesJson,
    required String exclusionsJson,
  }) async {
    try {
      final err = await _method.invokeMethod<String>('updateRules', {
        'rulesJson': rulesJson,
        'exclusionsJson': exclusionsJson,
      });
      if (err == null || err.isEmpty) {
        _log.info('rules', 'native updateRules OK');
        return null;
      }
      return err;
    } on PlatformException catch (e) {
      _log.warn('rules', 'updateRules PlatformException', {'message': e.message ?? 'unknown'});
      return e.message;
    } on MissingPluginException {
      return 'native updateRules unavailable';
    }
  }
}
