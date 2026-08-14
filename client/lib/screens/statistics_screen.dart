import 'dart:async';

import 'package:flutter/material.dart';

import '../services/connection_duration.dart';
import '../services/session_stats.dart';
import '../services/vpn_channel.dart';
import '../theme/app_theme.dart';

/// Client-local connection statistics (BL-044). No URLs or browsing history.
class StatisticsScreen extends StatefulWidget {
  const StatisticsScreen({super.key});

  @override
  State<StatisticsScreen> createState() => StatisticsScreenState();
}

class StatisticsScreenState extends State<StatisticsScreen>
    with AutomaticKeepAliveClientMixin {
  final _service = SessionStatsService();
  SessionStatsSnapshot? _stats;
  bool _loading = true;
  bool _vpnConnected = false;
  Duration _liveSession = Duration.zero;
  int? _liveRttMs;
  StreamSubscription<VpnStatusUpdate>? _vpnSub;
  Timer? _refreshTimer;
  DateTime? _connectedSince;

  @override
  bool get wantKeepAlive => true;

  @override
  void initState() {
    super.initState();
    _vpnSub = VpnChannel.statusStream.listen(_onVpnStatus);
    _refreshTimer = Timer.periodic(const Duration(seconds: 4), (_) {
      if (mounted) _refresh(silent: true);
    });
    unawaited(_refresh());
    unawaited(_syncNativeVpnState());
  }

  Future<void> _syncNativeVpnState() async {
    try {
      final native = await VpnChannel.fetchNativeStatus();
      if (!mounted || native == null) return;
      if (native.event == VpnEvent.connected) {
        setState(() {
          _vpnConnected = true;
          _connectedSince ??= DateTime.now();
          _liveRttMs = native.pingMs;
        });
      }
    } catch (_) {}
  }

  void _onVpnStatus(VpnStatusUpdate update) {
    if (!mounted) return;
    setState(() {
      switch (update.event) {
        case VpnEvent.connected:
          _vpnConnected = true;
          _connectedSince = DateTime.now();
          _liveRttMs = update.pingMs;
        case VpnEvent.connecting:
          _vpnConnected = false;
          _connectedSince = null;
          _liveSession = Duration.zero;
        case VpnEvent.disconnected:
        case VpnEvent.error:
        case VpnEvent.permissionDenied:
          _vpnConnected = false;
          _connectedSince = null;
          _liveSession = Duration.zero;
          _liveRttMs = null;
      }
    });
    if (update.event == VpnEvent.disconnected) {
      unawaited(_refresh(silent: true));
    }
  }

  Future<void> refresh() => _refresh();

  Future<void> _refresh({bool silent = false}) async {
    if (!silent) setState(() => _loading = true);
    final snap = await _service.load();
    if (!mounted) return;
    Duration live = Duration.zero;
    if (_vpnConnected && _connectedSince != null) {
      live = DateTime.now().difference(_connectedSince!);
    }
    setState(() {
      _stats = snap;
      _liveSession = live;
      _loading = false;
    });
  }

  @override
  void dispose() {
    _vpnSub?.cancel();
    _refreshTimer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    super.build(context);
    final stats = _stats;
    return Scaffold(
      appBar: AppBar(
        title: const Text('Статистика'),
        backgroundColor: Colors.transparent,
        elevation: 0,
        actions: [
          IconButton(
            tooltip: 'Обновить',
            onPressed: _loading ? null : () => _refresh(),
            icon: const Icon(Icons.refresh_rounded),
          ),
        ],
      ),
      body: _loading && stats == null
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: const EdgeInsets.fromLTRB(20, 8, 20, 24),
              children: [
                if (_vpnConnected) ...[
                  _StatTile(
                    title: 'Сейчас подключено',
                    value: formatConnectionDuration(_liveSession),
                    highlight: true,
                  ),
                  if (_liveRttMs != null && _liveRttMs! > 0)
                    _StatTile(
                      title: 'Текущий отклик',
                      value: '$_liveRttMs мс',
                    ),
                  const SizedBox(height: 8),
                ],
                if (stats == null || !stats.hasData)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 16),
                    child: Text(
                      _vpnConnected
                          ? 'Сессия идёт — накопленная статистика обновится после отключения.'
                          : 'Подключитесь хотя бы раз — здесь появятся данные с этого устройства.',
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                  )
                else ...[
                  Text(
                    'На этом устройстве',
                    style: Theme.of(context).textTheme.labelLarge?.copyWith(
                          color: AppColors.textSecondary,
                        ),
                  ),
                  const SizedBox(height: 12),
                  _StatTile(
                    title: 'В сети сегодня',
                    value: formatConnectionDuration(stats.onlineToday),
                  ),
                  _StatTile(
                    title: 'В сети за 7 дней',
                    value: formatConnectionDuration(stats.onlineLast7Days),
                  ),
                  _StatTile(
                    title: 'Переподключений сегодня',
                    value: '${stats.reconnectsToday}',
                  ),
                  _StatTile(
                    title: 'Переподключений за 7 дней',
                    value: '${stats.reconnectsLast7Days}',
                  ),
                  _StatTile(
                    title: 'Средний отклик сегодня',
                    value: stats.averageRttMs != null
                        ? '${stats.averageRttMs} мс'
                        : '—',
                  ),
                ],
                const SizedBox(height: 16),
                Text(
                  'Мы не сохраняем адреса сайтов и историю просмотров — только технические параметры соединения.',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: AppColors.textSecondary,
                      ),
                ),
              ],
            ),
    );
  }
}

class _StatTile extends StatelessWidget {
  final String title;
  final String value;
  final bool highlight;

  const _StatTile({
    required this.title,
    required this.value,
    this.highlight = false,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 10),
      color: highlight ? AppColors.cyan.withOpacity(0.08) : null,
      child: ListTile(
        title: Text(title),
        trailing: Text(
          value,
          style: Theme.of(context).textTheme.titleMedium?.copyWith(
                color: highlight ? AppColors.cyan : AppColors.cyan,
              ),
        ),
      ),
    );
  }
}
