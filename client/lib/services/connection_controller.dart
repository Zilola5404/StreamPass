import 'dart:async';

import 'package:flutter/foundation.dart';

import 'vpn_channel.dart';

/// Canonical connection state for all screens (TASK-WIN-001 §6.2).
///
/// Native EventChannel / WindowsTrafficEngine feed this object.
/// UI must not treat Hysteria handshake as [VpnEvent.connected] until
/// [trafficReady] is set by a data-plane health check.
enum ConnectedUiPolicy {
  /// Android: native already requires a working tunnel before emitting connected.
  nativeConnectedMeansReady,

  /// Windows: handshake/TUN up is not enough — wait for [markTrafficReady].
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

  VpnStatusUpdate get status => _status;
  VpnEvent get event => _status.event;
  bool get trafficReady => _trafficReady;

  /// User-visible "Подключено" only when tunnel is up **and** data path proven.
  bool get showConnected {
    if (_status.event != VpnEvent.connected) return false;
    if (policy == ConnectedUiPolicy.requireTrafficReady) return _trafficReady;
    return true;
  }

  void attach() {
    if (_attached) return;
    _attached = true;
    if (!kIsWeb && defaultTargetPlatform == TargetPlatform.windows) {
      policy = ConnectedUiPolicy.requireTrafficReady;
    }
    VpnChannel.ensureListening();
    // Single subscription — avoid duplicate handlers if attach() is called again.
    _sub ??= VpnChannel.statusStream.listen(_onUpdate);
  }

  void _onUpdate(VpnStatusUpdate update) {
    _status = update;
    if (update.event != VpnEvent.connected) {
      _trafficReady = false;
      _pendingTrafficReady = false;
    } else if (policy == ConnectedUiPolicy.nativeConnectedMeansReady) {
      _trafficReady = true;
    } else if (_pendingTrafficReady) {
      _trafficReady = true;
      _pendingTrafficReady = false;
    }
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
    notifyListeners();
  }

  void clearTrafficReady() {
    if (!_trafficReady) return;
    _trafficReady = false;
    notifyListeners();
  }

  @visibleForTesting
  void debugReset() {
    _status = VpnStatusUpdate(VpnEvent.disconnected);
    _trafficReady = false;
    _pendingTrafficReady = false;
    _attached = false;
    unawaited(_sub?.cancel());
    _sub = null;
    policy = ConnectedUiPolicy.nativeConnectedMeansReady;
  }

  @visibleForTesting
  void debugApply(VpnStatusUpdate update) => _onUpdate(update);
}
