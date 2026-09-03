import 'dart:async';

import 'package:flutter/foundation.dart';

import 'vpn_channel.dart';

/// Canonical connection state for all screens (TASK-WIN-001 §6.2).
///
/// Native EventChannel / WindowsTrafficEngine feed this object.
/// UI must not treat Hysteria handshake as [VpnEvent.connected] until
/// [trafficReady] is set by a data-plane health check.
enum ConnectedUiPolicy {
  /// Legacy: native connected event is treated as user-visible ready.
  nativeConnectedMeansReady,

  /// Handshake/TUN up is not enough — wait for [markTrafficReady] (first_byte).
  requireTrafficReady,
}

class ConnectionController extends ChangeNotifier {
  ConnectionController._();
  static final ConnectionController instance = ConnectionController._();

  VpnStatusUpdate _status = VpnStatusUpdate(VpnEvent.disconnected);
  bool _trafficReady = false;
  bool _pendingTrafficReady = false;
  bool _attached = false;
  StreamSubscription<VpnStatusUpdate>? _sub;
  ConnectedUiPolicy policy = ConnectedUiPolicy.nativeConnectedMeansReady;

  /// Wall-clock start of the user-visible connected session.
  /// Set once when [showConnected] becomes true; cleared only on real disconnect.
  DateTime? _sessionStartedAt;

  VpnStatusUpdate get status => _status;
  VpnEvent get event => _status.event;
  bool get trafficReady => _trafficReady;

  /// Shared by Home orb timer and Statistics «Сейчас подключено».
  DateTime? get sessionStartedAt => _sessionStartedAt;

  Duration get liveSessionDuration {
    final since = _sessionStartedAt;
    if (since == null || !showConnected) return Duration.zero;
    return DateTime.now().difference(since);
  }

  /// User-visible "Подключено" only when tunnel is up **and** data path proven.
  bool get showConnected {
    if (_status.event != VpnEvent.connected) return false;
    if (policy == ConnectedUiPolicy.requireTrafficReady) return _trafficReady;
    return true;
  }

  void attach() {
    if (_attached) return;
    _attached = true;
    // RELEASE-NETWORK-001: CONNECTED ≠ TRAFFIC_READY on Android and Windows.
    if (!kIsWeb &&
        (defaultTargetPlatform == TargetPlatform.windows ||
            defaultTargetPlatform == TargetPlatform.android)) {
      policy = ConnectedUiPolicy.requireTrafficReady;
    }
    VpnChannel.ensureListening();
    // Single subscription — avoid duplicate handlers if attach() is called again.
    _sub ??= VpnChannel.statusStream.listen(_onUpdate);
  }

  void _onUpdate(VpnStatusUpdate update) {
    _status = update;
    if (update.trafficReadyHint) {
      if (update.event == VpnEvent.connected) {
        _trafficReady = true;
        _pendingTrafficReady = false;
      } else {
        _pendingTrafficReady = true;
      }
    } else if (update.event != VpnEvent.connected) {
      _trafficReady = false;
      _pendingTrafficReady = false;
    } else if (policy == ConnectedUiPolicy.nativeConnectedMeansReady) {
      _trafficReady = true;
    } else if (_pendingTrafficReady) {
      _trafficReady = true;
      _pendingTrafficReady = false;
    }
    _syncSessionClock();
    notifyListeners();
  }

  /// Called by the traffic engine after first_byte / health test — never from handshake alone.
  void markTrafficReady() {
    if (_trafficReady) return;
    if (_status.event != VpnEvent.connected) {
      // first_byte may arrive before status=connected is applied — latch it.
      _pendingTrafficReady = true;
      return;
    }
    _trafficReady = true;
    _pendingTrafficReady = false;
    _syncSessionClock();
    notifyListeners();
  }

  void clearTrafficReady() {
    if (!_trafficReady) return;
    _trafficReady = false;
    // Keep [_sessionStartedAt] while native event is still connected — only
    // a real disconnect clears the session clock.
    notifyListeners();
  }

  /// Starts the shared session clock at most once per connected period.
  void _syncSessionClock() {
    if (showConnected) {
      _sessionStartedAt ??= DateTime.now();
      return;
    }
    switch (_status.event) {
      case VpnEvent.disconnected:
      case VpnEvent.error:
      case VpnEvent.permissionDenied:
      case VpnEvent.connecting:
        _sessionStartedAt = null;
      case VpnEvent.connected:
        // Connected but not yet traffic-ready (Windows): wait.
        break;
    }
  }

  @visibleForTesting
  void debugReset() {
    _status = VpnStatusUpdate(VpnEvent.disconnected);
    _trafficReady = false;
    _pendingTrafficReady = false;
    _sessionStartedAt = null;
    _attached = false;
    unawaited(_sub?.cancel());
    _sub = null;
    policy = ConnectedUiPolicy.nativeConnectedMeansReady;
  }

  @visibleForTesting
  void debugApply(VpnStatusUpdate update) => _onUpdate(update);
}
