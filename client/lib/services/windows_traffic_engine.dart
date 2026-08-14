import 'dart:async';

import 'connection_controller.dart';
import 'connection_log.dart';
import 'streampass_api.dart';
import 'vpn_channel.dart';
import 'windows_core_client.dart';

/// Windows traffic engine (TASK-WIN-001): Wintun sidecar + Decision/Hysteria.
///
/// [VpnEvent.connected] means TUN is up — not Hysteria handshake.
/// User-visible "Подключено" still waits for [ConnectionController.markTrafficReady]
/// (`[vpn] traffic_ready` from first_byte).
class WindowsTrafficEngine implements WindowsVpnAdapter {
  WindowsTrafficEngine._();
  static final WindowsTrafficEngine instance = WindowsTrafficEngine._();

  final _log = ConnectionLog.instance;
  final _controller = StreamController<VpnStatusUpdate>.broadcast();
  VpnStatusUpdate _last = VpnStatusUpdate(VpnEvent.disconnected);
  WindowsCoreClient? _core;

  @override
  Stream<VpnStatusUpdate> get statusStream => _controller.stream;
  VpnStatusUpdate get last => _last;

  @override
  void ensureStarted() {
    if (_last.event == VpnEvent.disconnected) {
      _emit(VpnStatusUpdate(VpnEvent.disconnected));
    }
  }

  @override
  Future<bool> connect(
    RelayServer server, {
    String rulesJson = '',
    String exclusionsJson = '',
    int mtu = 1400,
    String networkMode = 'split',
    bool blockUdp443 = false,
  }) async {
    _log.info('vpn', 'CONNECT_REQUESTED', {
      'platform': 'windows',
      'relayId': server.id,
      'host': server.host,
      'mtu': '$mtu',
      'networkMode': networkMode,
    });
    _emit(VpnStatusUpdate(VpnEvent.connecting, relayName: server.id));

    await _core?.dispose();
    final core = WindowsCoreClient();
    _core = core;
    try {
      await core.ensureStarted(
        onStatus: _emit,
        onLog: _onNativeLog,
      );
      await core.startTunnel(
        server: server,
        rulesJson: rulesJson,
        exclusionsJson: exclusionsJson,
        networkMode: networkMode,
        mtu: mtu,
        blockUdp443: blockUdp443,
      );
      return true;
    } on VpnConnectException catch (e) {
      var msg = e.message;
      if (msg.contains('администратора') || msg.toLowerCase().contains('access is denied')) {
        msg =
            'Запустите StreamPass от имени администратора (нужно для Wintun). '
            'Если сайты не открывались после прошлого сеанса — отключите адаптер StreamPass '
            'или выполните: route delete 0.0.0.0 mask 0.0.0.0 10.10.0.2';
      }
      _emit(VpnStatusUpdate(
        VpnEvent.error,
        relayName: server.id,
        errorMessage: msg,
      ));
      await core.dispose();
      _core = null;
      throw VpnConnectException(msg);
    } catch (e) {
      final msg = e.toString();
      _emit(VpnStatusUpdate(
        VpnEvent.error,
        relayName: server.id,
        errorMessage: msg,
      ));
      await core.dispose();
      _core = null;
      throw VpnConnectException(msg);
    }
  }

  @override
  Future<void> disconnect() async {
    _log.info('vpn', 'DISCONNECTED', {'platform': 'windows'});
    try {
      await _core?.stopTunnel();
    } finally {
      await _core?.dispose();
      _core = null;
      ConnectionController.instance.clearTrafficReady();
      _emit(VpnStatusUpdate(VpnEvent.disconnected));
    }
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
    final lower = line.toLowerCase();
    final level = lower.contains('fail') || lower.contains('error')
        ? ConnectionLogLevel.error
        : lower.contains('warn')
            ? ConnectionLogLevel.warn
            : ConnectionLogLevel.info;
    if (level == ConnectionLogLevel.error) {
      _log.error('vpn', line);
    } else if (level == ConnectionLogLevel.warn) {
      _log.warn('vpn', line);
    } else {
      _log.info('vpn', line);
    }
  }

  void _emit(VpnStatusUpdate update) {
    _last = update;
    VpnChannel.lastStatus = update;
    if (!_controller.isClosed) {
      _controller.add(update);
    }
  }
}
