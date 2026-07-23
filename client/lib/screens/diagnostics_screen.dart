import 'dart:io' show Platform;
import 'package:flutter/material.dart';
import '../theme/app_theme.dart';
import '../services/vpn_channel.dart';

/// Shows exactly the telemetry categories allowed by ТЗ §14:
/// RTT, Packet Loss, Relay, client version, OS, connection time, error code.
/// Deliberately does NOT show and never will show: site history,
/// traffic content, URLs, or personal data — matching the "Не собирать" list.
class DiagnosticsScreen extends StatefulWidget {
  const DiagnosticsScreen({super.key});

  @override
  State<DiagnosticsScreen> createState() => _DiagnosticsScreenState();
}

class _DiagnosticsScreenState extends State<DiagnosticsScreen> {
  VpnStatusUpdate? _last;
  DateTime? _connectedSince;

  @override
  void initState() {
    super.initState();
    VpnChannel.statusStream.listen((update) {
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
      _DiagRow('Версия клиента', '0.1.0'),
      _DiagRow('ОС', Platform.operatingSystem),
    ];

    return Scaffold(
      appBar: AppBar(
        title: const Text('Диагностика'),
        backgroundColor: Colors.transparent,
        elevation: 0,
      ),
      body: ListView(
        padding: const EdgeInsets.symmetric(vertical: 8),
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
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Text(
              'StreamPass передаёт только эти технические параметры. '
              'История сайтов, содержимое трафика и персональные данные '
              'не собираются и не отправляются.',
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

class _DiagRow {
  final String label;
  final String value;
  _DiagRow(this.label, this.value);
}
