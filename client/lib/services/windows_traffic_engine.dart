import 'dart:async';

import 'connection_controller.dart';
import 'connection_log.dart';
import 'native_vpn_log_level.dart';
import 'streampass_api.dart';
import 'vpn_channel.dart';
import 'windows_core_client.dart';

/// Windows traffic engine (TASK-WIN-001): Wintun sidecar + Decision/Hysteria.
///
/// Owns a single connect session: every connect Future completes with
/// success, error, or disconnect — never left pending forever.
class WindowsTrafficEngine implements WindowsVpnAdapter {
  WindowsTrafficEngine._();
  static final WindowsTrafficEngine instance = WindowsTrafficEngine._();

  final _log = ConnectionLog.instance;
  final _controller = StreamController<VpnStatusUpdate>.broadcast();
  VpnStatusUpdate _last = VpnStatusUpdate(VpnEvent.disconnected);
  WindowsCoreClient? _core;

  /// Generation bumped on each connect/disconnect so stale core events are ignored.
  int _session = 0;
  bool _connecting = false;
  Future<bool>? _inFlight;
  Timer? _trafficWatchdog;

  @override
  Stream<VpnStatusUpdate> get statusStream => _controller.stream;
  VpnStatusUpdate get last => _last;

  @override
  void ensureStarted() {
    // Do not re-broadcast disconnected — that floods listeners on every attach.
  }

  @override
  Future<bool> connect(
    RelayServer server, {
    String rulesJson = '',
    String exclusionsJson = '',
    int mtu = 1400,
    String networkMode = 'split',
    bool blockUdp443 = false,
  }) {
    final existing = _inFlight;
    if (existing != null) {
      _log.warn('vpn', 'connect ignored — already in progress', {
        'relayId': server.id,
      });
      // Coalesce: wait for the same Future instead of leaving a second caller
      // with no completion path of its own.
      return existing;
    }

    final op = _connectOnce(
      server,
      rulesJson: rulesJson,
      exclusionsJson: exclusionsJson,
      mtu: mtu,
      networkMode: networkMode,
      blockUdp443: blockUdp443,
    );
    _inFlight = op;
    return op.whenComplete(() {
      if (identical(_inFlight, op)) {
        _inFlight = null;
      }
    });
  }

  Future<bool> _connectOnce(
    RelayServer server, {
    required String rulesJson,
    required String exclusionsJson,
    required int mtu,
    required String networkMode,
    required bool blockUdp443,
  }) async {
    final session = ++_session;
    _connecting = true;
    _cancelTrafficWatchdog();
    ConnectionController.instance.clearTrafficReady();

    _log.info('vpn', 'CONNECT_REQUESTED', {
      'platform': 'windows',
      'session': '$session',
      'relayId': server.id,
      'host': server.host,
      'mtu': '$mtu',
      'networkMode': networkMode,
    });
    _emit(VpnStatusUpdate(VpnEvent.connecting, relayName: server.id));

    try {
      final core = await _ensureCore(session);
      await core.startTunnel(
        server: server,
        rulesJson: rulesJson,
        exclusionsJson: exclusionsJson,
        networkMode: networkMode,
        mtu: mtu,
        blockUdp443: blockUdp443,
      );
      if (session != _session) {
        throw VpnConnectException('connect superseded by newer session');
      }
      _connecting = false;
      // Native should already have emitted connected; ensure terminal state.
      if (_last.event != VpnEvent.connected) {
        _emit(VpnStatusUpdate(
          VpnEvent.connected,
          relayName: server.id,
        ));
      }
      _armTrafficWatchdog(session, server.id);
      return true;
    } on VpnConnectException catch (e) {
      await _failSession(session, server.id, e.message);
      // Timeout / hung start disposed the core — drop soft-reuse handle.
      if (_core != null && !_core!.isAttached) {
        _core = null;
      }
      rethrow;
    } on TimeoutException catch (e) {
      final msg =
          'Подключение не завершилось вовремя (${e.duration ?? const Duration(seconds: 45)}). '
          'Проверьте UAC / Wintun и повторите.';
      await _failSession(session, server.id, msg);
      throw VpnConnectException(msg);
    } catch (e) {
      final msg = e.toString();
      await _failSession(session, server.id, msg);
      throw VpnConnectException(msg);
    } finally {
      if (session == _session) {
        _connecting = false;
      }
    }
  }

  Future<WindowsCoreClient> _ensureCore(int session) async {
    final existing = _core;
    if (existing != null && existing.isAttached) {
      // Soft ping: dead elevated core leaves a stale socket that never replies.
      try {
        await existing.ping().timeout(const Duration(seconds: 2));
        return existing;
      } catch (_) {
        _log.warn('vpn', 'CORE_STALE — respawning after ping fail');
        await existing.dispose();
        if (identical(_core, existing)) {
          _core = null;
        }
      }
    } else {
      await existing?.dispose();
      if (identical(_core, existing)) {
        _core = null;
      }
    }
    if (session != _session) {
      throw VpnConnectException('connect cancelled');
    }
    final core = WindowsCoreClient();
    _core = core;
    await core.ensureStarted(
      onStatus: (update) => _onCoreStatus(session, update),
      onLog: _onNativeLog,
    );
    return core;
  }

  void _onCoreStatus(int session, VpnStatusUpdate update) {
    if (session != _session) return;
    // Tear-down of a previous tunnel must not yank UI out of connecting.
    if (_connecting &&
        (update.event == VpnEvent.disconnected ||
            update.event == VpnEvent.connecting)) {
      return;
    }
    _emit(update);
  }

  Future<void> _failSession(int session, String relayId, String message) async {
    if (session != _session) return;
    _connecting = false;
    _cancelTrafficWatchdog();
    ConnectionController.instance.clearTrafficReady();
    var msg = message;
    if (msg.contains('администратора') ||
        msg.toLowerCase().contains('access is denied')) {
      msg =
          'Нужны права администратора для Wintun (подтвердите UAC). '
          'Если сайты не открывались после прошлого сеанса: '
          'route delete 0.0.0.0 mask 0.0.0.0 10.10.0.2';
    }
    _emit(VpnStatusUpdate(
      VpnEvent.error,
      relayName: relayId,
      errorMessage: msg,
    ));
    // Soft fail: keep elevated core for retry without a second UAC when possible.
    try {
      await _core?.stopTunnel();
    } catch (_) {}
  }

  void _armTrafficWatchdog(int session, String relayId) {
    _cancelTrafficWatchdog();
    _trafficWatchdog = Timer(const Duration(seconds: 20), () {
      if (session != _session) return;
      final c = ConnectionController.instance;
      if (c.event != VpnEvent.connected) return;
      if (c.trafficReady) return;
      // TUN is up but no first_byte — do not leave UI in connecting forever.
      _log.warn('vpn', 'traffic_ready watchdog — marking ready after TUN up', {
        'session': '$session',
        'relayId': relayId,
      });
      c.markTrafficReady();
    });
  }

  void _cancelTrafficWatchdog() {
    _trafficWatchdog?.cancel();
    _trafficWatchdog = null;
  }

  @override
  Future<void> disconnect() async {
    final session = ++_session;
    _connecting = false;
    _inFlight = null;
    _cancelTrafficWatchdog();
    _log.info('vpn', 'DISCONNECTED', {
      'platform': 'windows',
      'session': '$session',
    });
    try {
      await _core?.stopTunnel();
    } finally {
      // Keep process for next connect (one UAC per app session); only stop tunnel.
      ConnectionController.instance.clearTrafficReady();
      _emit(VpnStatusUpdate(VpnEvent.disconnected));
    }
  }

  /// Hard teardown (app exit / tests).
  Future<void> disposeCore() async {
    _session++;
    _connecting = false;
    _inFlight = null;
    _cancelTrafficWatchdog();
    await _core?.dispose();
    _core = null;
    ConnectionController.instance.clearTrafficReady();
    _emit(VpnStatusUpdate(VpnEvent.disconnected));
  }

  @override
  Future<String?> updateRules({
    required String rulesJson,
    required String exclusionsJson,
  }) async {
    final err = await _core?.updateRules(
      rulesJson: rulesJson,
      exclusionsJson: exclusionsJson,
    );
    _log.info('vpn', 'RULES_LOADED', {
      'platform': 'windows',
      'applied': err == null ? 'true' : 'false',
      'bytes': '${rulesJson.length}',
      if (err != null) 'error': err,
    });
    return err;
  }

  void _onNativeLog(String line) {
    final level = classifyNativeVpnLog(line);
    if (level == ConnectionLogLevel.error) {
      _log.error('vpn', line);
    } else if (level == ConnectionLogLevel.warn) {
      _log.warn('vpn', line);
    } else {
      _log.info('vpn', line);
    }
  }

  void _emit(VpnStatusUpdate update) {
    // Collapse duplicate disconnected / identical consecutive events.
    if (_last.event == update.event &&
        _last.relayName == update.relayName &&
        _last.errorMessage == update.errorMessage &&
        (update.event == VpnEvent.disconnected ||
            update.event == VpnEvent.connecting)) {
      return;
    }
    _last = update;
    VpnChannel.lastStatus = update;
    if (!_controller.isClosed) {
      _controller.add(update);
    }
  }
}
