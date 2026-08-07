import 'dart:async';
import 'dart:io' show Platform;
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../build_info.dart';
import '../theme/app_theme.dart';
import '../services/vpn_channel.dart';
import '../services/connection_log.dart';
import '../services/native_connect_log.dart';
import '../services/settings_service.dart';

/// Connection diagnostics: live VPN status + on-device connect log (E09).
///
/// Network Mode / MTU live here (debug), not on E05 Settings — docs/07.4 §9.
class DiagnosticsScreen extends StatefulWidget {
  const DiagnosticsScreen({super.key});

  @override
  State<DiagnosticsScreen> createState() => _DiagnosticsScreenState();
}

class _DiagnosticsScreenState extends State<DiagnosticsScreen> {
  VpnStatusUpdate? _last;
  DateTime? _connectedSince;
  final _log = ConnectionLog.instance;
  final _settingsService = SettingsService();
  AppSettings _settings = const AppSettings();
  StreamSubscription<ConnectionLogEntry>? _logSub;

  StreamSubscription<VpnStatusUpdate>? _statusSub;

  @override
  void initState() {
    super.initState();
    _refreshLogs();
    unawaited(_loadSettings());
    _last = VpnChannel.lastStatus;
    if (_last?.event == VpnEvent.connected) {
      _connectedSince ??= DateTime.now();
    }
    _logSub = _log.stream.listen((_) {
      if (mounted) setState(() {});
    });
    _statusSub = VpnChannel.statusStream.listen((update) {
      if (!mounted) return;
      setState(() {
        _last = update;
        if (update.event == VpnEvent.connected) {
          _connectedSince ??= DateTime.now();
        }
        if (update.event == VpnEvent.disconnected) {
          _connectedSince = null;
        }
      });
    });
    unawaited(_syncNativeStatus());
  }

  Future<void> _loadSettings() async {
    final s = await _settingsService.load();
    if (!mounted) return;
    setState(() => _settings = s);
  }

  Future<void> _updateNetworkMode(String mode) async {
    await _settingsService.setNetworkMode(mode);
    setState(() => _settings = _settings.copyWith(networkMode: mode));
    _needReconnectHint();
  }

  void _needReconnectHint() {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Переподключите VPN, чтобы применить Network Mode / MTU'),
        duration: Duration(seconds: 3),
      ),
    );
  }

  Future<void> _syncNativeStatus() async {
    final native = await VpnChannel.fetchNativeStatus();
    if (!mounted || native == null) return;
    setState(() {
      _last = native;
      if (native.event == VpnEvent.connected) {
        _connectedSince ??= DateTime.now();
      }
      if (native.event == VpnEvent.disconnected) {
        _connectedSince = null;
      }
    });
  }

  Future<void> _refreshLogs() async {
    await NativeConnectLog.pullFromNative();
    if (mounted) setState(() {});
  }

  Future<void> _copyLogs() async {
    await _refreshLogs();
    await Clipboard.setData(ClipboardData(text: _log.exportText()));
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Лог скопирован в буфер обмена')),
    );
  }

  Future<void> _clearLogs() async {
    _log.clear();
    await NativeConnectLog.clearNative();
    if (mounted) setState(() {});
  }

  @override
  void dispose() {
    _logSub?.cancel();
    _statusSub?.cancel();
    super.dispose();
  }

  String get _uptime {
    if (_connectedSince == null) return '—';
    final d = DateTime.now().difference(_connectedSince!);
    final m = d.inMinutes.remainder(60).toString().padLeft(2, '0');
    final h = d.inHours.toString().padLeft(2, '0');
    return '$h:$m';
  }

  @override
  Widget build(BuildContext context) {
    final rows = <_DiagRow>[
      _DiagRow('Статус', _statusLabel(_last?.event)),
      _DiagRow('Relay', _last?.relayName ?? '—'),
      _DiagRow('RTT', _last?.pingMs != null ? '${_last!.pingMs} ms' : '—'),
      _DiagRow('Время соединения', _uptime),
      _DiagRow('Код ошибки', _last?.errorMessage ?? '—'),
      _DiagRow('Версия клиента', BuildInfo.label),
      _DiagRow('ОС', Platform.operatingSystem),
    ];

    final entries = _log.entries.reversed.toList();

    return Scaffold(
      appBar: AppBar(
        title: const Text('Диагностика'),
        backgroundColor: Colors.transparent,
        elevation: 0,
        actions: [
          IconButton(
            tooltip: 'Обновить',
            onPressed: _refreshLogs,
            icon: const Icon(Icons.refresh_rounded),
          ),
          IconButton(
            tooltip: 'Копировать лог',
            onPressed: _copyLogs,
            icon: const Icon(Icons.copy_rounded),
          ),
          IconButton(
            tooltip: 'Очистить',
            onPressed: _clearLogs,
            icon: const Icon(Icons.delete_outline_rounded),
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.only(bottom: 24),
        children: [
          for (final row in rows)
            ListTile(
              title: Text(row.label,
                  style: const TextStyle(color: AppColors.textSecondary)),
              trailing: Text(row.value,
                  style: const TextStyle(
                      color: AppColors.textPrimary, fontWeight: FontWeight.w600)),
            ),
          const Divider(height: 32),
          const Padding(
            padding: EdgeInsets.symmetric(horizontal: 16),
            child: Text(
              'Network Mode (debug / E09)',
              style: TextStyle(fontWeight: FontWeight.w600),
            ),
          ),
          RadioListTile<String>(
            title: const Text('Split Tunnel'),
            subtitle: const Text('Product default'),
            value: 'split',
            groupValue: _settings.networkMode,
            activeColor: AppColors.cyan,
            onChanged: (v) => _updateNetworkMode(v!),
          ),
          RadioListTile<String>(
            title: const Text('Full Relay'),
            subtitle: const Text('Force RELAY — isolate relay path'),
            value: 'full_relay',
            groupValue: _settings.networkMode,
            activeColor: AppColors.cyan,
            onChanged: (v) => _updateNetworkMode(v!),
          ),
          RadioListTile<String>(
            title: const Text('Direct Test'),
            subtitle: const Text('Force DIRECT — isolate ISP path'),
            value: 'direct_test',
            groupValue: _settings.networkMode,
            activeColor: AppColors.cyan,
            onChanged: (v) => _updateNetworkMode(v!),
          ),
          RadioListTile<String>(
            title: const Text('TCP only'),
            subtitle: const Text('Drop UDP/443 (QUIC off)'),
            value: 'tcp_only',
            groupValue: _settings.networkMode,
            activeColor: AppColors.cyan,
            onChanged: (v) => _updateNetworkMode(v!),
          ),
          SwitchListTile(
            title: const Text('UDP 443 blocked'),
            subtitle: const Text('Extra QUIC drop even in Split'),
            value: _settings.blockUdp443,
            activeColor: AppColors.cyan,
            onChanged: (v) async {
              await _settingsService.setBlockUdp443(v);
              setState(() => _settings = _settings.copyWith(blockUdp443: v));
              _needReconnectHint();
            },
          ),
          ListTile(
            title: const Text('MTU'),
            subtitle: Text('Сейчас ${_settings.mtu}'),
            trailing: DropdownButton<int>(
              value: _settings.mtu,
              items: const [
                DropdownMenuItem(value: 1280, child: Text('1280')),
                DropdownMenuItem(value: 1350, child: Text('1350')),
                DropdownMenuItem(value: 1400, child: Text('1400')),
              ],
              onChanged: (v) async {
                if (v == null) return;
                await _settingsService.setMtu(v);
                setState(() => _settings = _settings.copyWith(mtu: v));
                _needReconnectHint();
              },
            ),
          ),
          const Divider(height: 32),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Text(
              'Лог подключения (${entries.length})',
              style: Theme.of(context).textTheme.titleSmall,
            ),
          ),
          const SizedBox(height: 8),
          if (entries.isEmpty)
            const Padding(
              padding: EdgeInsets.all(16),
              child: Text(
                'Лог пуст. Нажмите Connect на главном экране — шаги подключения появятся здесь.',
                style: TextStyle(color: AppColors.textSecondary),
              ),
            )
          else
            ...entries.map((e) => _LogTile(entry: e)),
          const Divider(height: 32),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Text(
              'StreamPass передаёт только технические параметры. '
              'Пароли и содержимое connection_config в лог не пишутся.',
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ),
        ],
      ),
    );
  }

  String _statusLabel(VpnEvent? e) => switch (e) {
        VpnEvent.connected => 'Connected',
        VpnEvent.connecting => 'Connecting…',
        VpnEvent.disconnected || null => 'Disconnected',
        VpnEvent.permissionDenied => 'Permission denied',
        VpnEvent.error => 'Error',
      };
}

class _LogTile extends StatelessWidget {
  final ConnectionLogEntry entry;
  const _LogTile({required this.entry});

  Color get _color => switch (entry.level) {
        ConnectionLogLevel.error => AppColors.danger,
        ConnectionLogLevel.warn => AppColors.amber,
        ConnectionLogLevel.info => AppColors.textSecondary,
      };

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      child: SelectableText(
        entry.formatLine(),
        style: TextStyle(fontSize: 11, color: _color, fontFamily: 'monospace'),
      ),
    );
  }
}

class _DiagRow {
  final String label;
  final String value;
  _DiagRow(this.label, this.value);
}
