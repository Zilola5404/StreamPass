import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'connection_controller.dart';
import 'connection_log.dart';
import 'streampass_api.dart';
import 'vpn_channel.dart';

/// Spawns `streampasscore.exe` and speaks newline JSON on 127.0.0.1.
class WindowsCoreClient {
  WindowsCoreClient();

  final _log = ConnectionLog.instance;
  Process? _proc;
  int? _elevatedPid;
  Socket? _socket;
  String _token = '';
  int _nextId = 1;
  final _pending = <int, Completer<Map<String, dynamic>>>{};
  final _lineBuf = StringBuffer();
  StreamSubscription<List<int>>? _sub;
  bool _closed = false;

  bool get isAttached => _socket != null && !_closed;

  Future<void> ensureStarted({
    required void Function(VpnStatusUpdate update) onStatus,
    required void Function(String line) onLog,
  }) async {
    if (isAttached) return;
    // Stale handle after process death — drop before respawn.
    if (_socket != null) {
      try {
        _socket!.destroy();
      } catch (_) {}
      _socket = null;
    }
    _closed = false;
    _lineBuf.clear();

    final exe = _coreExecutable();
    if (!File(exe).existsSync()) {
      throw VpnConnectException(
        'Не найден streampasscore.exe ($exe). Соберите ядро: '
        r'client\windows\native\build_core.ps1',
      );
    }
    final wintun = File('${File(exe).parent.path}\\wintun.dll');
    if (!wintun.existsSync()) {
      throw VpnConnectException(
        'Не найден wintun.dll рядом с ядром. Запустите build_core.ps1.',
      );
    }

    final portFile = File(
      '${Directory.systemTemp.path}\\streampass-core-${DateTime.now().microsecondsSinceEpoch}.json',
    );
    if (portFile.existsSync()) {
      portFile.deleteSync();
    }

    // Wintun CreateAdapter needs Administrator. Always spawn via UAC RunAs
    // (no prompt if the parent is already elevated).
    final workDir = File(exe).parent.path;
    _log.info('vpn', 'CORE_SPAWN', {'exe': exe, 'elevated': 'RunAs'});
    final ps = StringBuffer()
      ..writeln('\$ErrorActionPreference = "Stop"')
      ..writeln(
        '\$p = Start-Process -FilePath ${_psQuote(exe)} '
        '-ArgumentList @("--port-file", ${_psQuote(portFile.path)}) '
        '-WorkingDirectory ${_psQuote(workDir)} '
        '-Verb RunAs -PassThru -WindowStyle Hidden',
      )
      ..writeln(
        'if (\$null -eq \$p) { throw "UAC cancelled or elevation failed" }',
      )
      ..writeln('Write-Output \$p.Id');
    final elev = await Process.run(
      'powershell',
      ['-NoProfile', '-ExecutionPolicy', 'Bypass', '-Command', ps.toString()],
    );
    if (elev.exitCode != 0) {
      final err = '${elev.stderr}\n${elev.stdout}'.trim();
      throw VpnConnectException(
        'Нужны права администратора для Wintun (UAC). '
        'Подтвердите запрос или запустите StreamPass от администратора.'
        '${err.isEmpty ? '' : '\n$err'}',
      );
    }
    final pid = int.tryParse(
      elev.stdout.toString().trim().split(RegExp(r'\r?\n')).last.trim(),
    );
    if (pid != null && pid > 0) {
      _elevatedPid = pid;
    }

    final meta = await _waitPortFile(portFile);
    _token = meta['token'] as String? ?? '';
    final port = meta['port'] as int? ?? 0;
    if (port <= 0 || _token.isEmpty) {
      await dispose();
      throw VpnConnectException('streampasscore не опубликовал порт/token');
    }

    _socket = await Socket.connect('127.0.0.1', port);
    _sub = _socket!.listen(
      (data) => _onBytes(data, onStatus, onLog),
      onDone: () {
        _sub = null;
        _socket = null;
        for (final c in _pending.values) {
          if (!c.isCompleted) {
            c.completeError(StateError('core socket closed'));
          }
        }
        _pending.clear();
        onStatus(VpnStatusUpdate(VpnEvent.disconnected));
      },
      onError: (Object e) {
        _sub = null;
        _socket = null;
        for (final c in _pending.values) {
          if (!c.isCompleted) {
            c.completeError(e);
          }
        }
        _pending.clear();
        onStatus(VpnStatusUpdate(
          VpnEvent.error,
          errorMessage: e.toString(),
        ));
      },
    );
  }

  Future<void> startTunnel({
    required RelayServer server,
    required String rulesJson,
    required String exclusionsJson,
    required String networkMode,
    required int mtu,
    required bool blockUdp443,
  }) async {
    try {
      final reply = await _rpc({
        'cmd': 'start',
        'relayHost': server.host,
        'relayPort': server.port,
        'connectionConfig': server.connectionConfig,
        'rulesJson': rulesJson,
        'exclusionsJson': exclusionsJson,
        'networkMode': networkMode,
        'mtu': mtu,
        'blockUdp443': blockUdp443,
      });
      if (reply['ok'] != true) {
        throw VpnConnectException(
          reply['error'] as String? ?? 'Не удалось поднять Windows TUN',
        );
      }
    } on VpnConnectException catch (e) {
      // Half-open / hung start: drop core so the next Connect respawns cleanly.
      if (e.message.contains('TimeoutException') ||
          e.message.contains('cmd=start')) {
        await dispose();
      }
      rethrow;
    }
  }

  Future<String?> updateRules({
    required String rulesJson,
    required String exclusionsJson,
  }) async {
    if (_socket == null) return 'no active tunnel';
    final reply = await _rpc({
      'cmd': 'update_rules',
      'rulesJson': rulesJson,
      'exclusionsJson': exclusionsJson,
    });
    if (reply['ok'] == true) return null;
    return reply['error'] as String? ?? 'update_rules failed';
  }

  Future<void> stopTunnel() async {
    if (_socket == null) return;
    try {
      await _rpc({'cmd': 'stop'}).timeout(const Duration(seconds: 4));
    } catch (_) {}
  }

  Future<void> dispose() async {
    _closed = true;
    for (final c in _pending.values) {
      if (!c.isCompleted) {
        c.completeError(StateError('core disposed'));
      }
    }
    _pending.clear();
    await _sub?.cancel();
    _sub = null;
    try {
      _socket?.destroy();
    } catch (_) {}
    _socket = null;
    final pid = _elevatedPid ?? _proc?.pid;
    _elevatedPid = null;
    _proc = null;
    if (pid != null && pid > 0) {
      // Elevated core may outlive unelevated killPid — best-effort.
      Process.killPid(pid);
      unawaited(Process.run('taskkill', ['/F', '/PID', '$pid']));
    }
  }

  String _psQuote(String value) => "'${value.replaceAll("'", "''")}'";

  String _coreExecutable() {
    final dir = File(Platform.resolvedExecutable).parent.path;
    return '$dir\\streampasscore.exe';
  }

  Future<Map<String, dynamic>> _waitPortFile(File file) async {
    // Allow time for the UAC consent dialog.
    final deadline = DateTime.now().add(const Duration(seconds: 90));
    while (DateTime.now().isBefore(deadline)) {
      if (file.existsSync()) {
        try {
          final raw = jsonDecode(file.readAsStringSync());
          if (raw is Map<String, dynamic> && raw['port'] != null) {
            try {
              file.deleteSync();
            } catch (_) {}
            return raw;
          }
        } catch (_) {}
      }
      await Future<void>.delayed(const Duration(milliseconds: 50));
    }
    throw VpnConnectException(
      'streampasscore не запустился (нет port-file). '
      'Подтвердите UAC или запустите StreamPass от администратора.',
    );
  }

  Future<Map<String, dynamic>> _rpc(Map<String, dynamic> body) async {
    final socket = _socket;
    if (socket == null) {
      throw VpnConnectException('ядро TUN не подключено');
    }
    final id = _nextId++;
    final c = Completer<Map<String, dynamic>>();
    _pending[id] = c;
    body['id'] = id;
    body['token'] = _token;
    try {
      socket.add(utf8.encode('${jsonEncode(body)}\n'));
    } catch (e) {
      _pending.remove(id);
      _socket = null;
      throw VpnConnectException('ядро TUN недоступно: $e');
    }
    try {
      return await c.future.timeout(const Duration(seconds: 45));
    } on TimeoutException {
      _pending.remove(id);
      if (!c.isCompleted) {
        c.completeError(StateError('rpc timeout'));
      }
      final cmd = body['cmd'] as String? ?? 'rpc';
      throw VpnConnectException(
        'TimeoutException after 0:00:45.000000: Future not completed '
        '(cmd=$cmd). Подключение прервано — повторите Connect.',
      );
    }
  }

  /// Lightweight liveness check for soft core reuse.
  Future<void> ping() async {
    final reply = await _rpc({'cmd': 'ping'});
    if (reply['ok'] != true) {
      throw StateError('ping failed');
    }
  }

  void _onBytes(
    List<int> data,
    void Function(VpnStatusUpdate update) onStatus,
    void Function(String line) onLog,
  ) {
    if (_closed) return;
    _lineBuf.write(utf8.decode(data, allowMalformed: true));
    var all = _lineBuf.toString();
    final parts = all.split('\n');
    _lineBuf
      ..clear()
      ..write(parts.removeLast());
    for (final line in parts) {
      if (line.trim().isEmpty) continue;
      _handleLine(line, onStatus, onLog);
    }
  }

  void _handleLine(
    String line,
    void Function(VpnStatusUpdate update) onStatus,
    void Function(String line) onLog,
  ) {
    Map<String, dynamic> msg;
    try {
      msg = jsonDecode(line) as Map<String, dynamic>;
    } catch (_) {
      onLog(line);
      return;
    }
    final type = msg['type'] as String? ?? '';
    if (type == 'reply' || msg.containsKey('ok')) {
      final id = msg['id'] as int? ?? 0;
      final c = _pending.remove(id);
      c?.complete(msg);
      return;
    }
    if (type == 'log') {
      final message = msg['message'] as String? ?? '';
      onLog(message);
      if (message.contains('traffic_ready')) {
        ConnectionController.instance.markTrafficReady();
      }
      return;
    }
    if (type == 'status') {
      final eventName = msg['event'] as String? ?? 'disconnected';
      final event = VpnEvent.values.firstWhere(
        (e) => e.name == eventName,
        orElse: () => VpnEvent.disconnected,
      );
      onStatus(VpnStatusUpdate(
        event,
        relayName: msg['relay'] as String?,
        pingMs: msg['pingMs'] as int?,
        errorMessage: msg['error'] as String?,
      ));
    }
  }
}
