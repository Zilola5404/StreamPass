import 'package:flutter/material.dart';

import '../services/connection_duration.dart';
import '../services/session_stats.dart';
import '../theme/app_theme.dart';

/// Client-local connection statistics (BL-044). No URLs or browsing history.
class StatisticsScreen extends StatefulWidget {
  const StatisticsScreen({super.key});

  @override
  State<StatisticsScreen> createState() => _StatisticsScreenState();
}

class _StatisticsScreenState extends State<StatisticsScreen> {
  final _service = SessionStatsService();
  SessionStatsSnapshot? _stats;
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _refresh();
  }

  Future<void> _refresh() async {
    setState(() => _loading = true);
    final snap = await _service.load();
    if (!mounted) return;
    setState(() {
      _stats = snap;
      _loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    final stats = _stats;
    return Scaffold(
      appBar: AppBar(
        title: const Text('Статистика'),
        backgroundColor: Colors.transparent,
        elevation: 0,
        actions: [
          IconButton(
            tooltip: 'Обновить',
            onPressed: _loading ? null : _refresh,
            icon: const Icon(Icons.refresh_rounded),
          ),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : stats == null || !stats.hasData
              ? Center(
                  child: Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 32),
                    child: Text(
                      'Пока недостаточно данных. Подключитесь — и здесь появится статистика.',
                      style: Theme.of(context).textTheme.bodyMedium,
                      textAlign: TextAlign.center,
                    ),
                  ),
                )
              : ListView(
                  padding: const EdgeInsets.fromLTRB(20, 8, 20, 24),
                  children: [
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

  const _StatTile({required this.title, required this.value});

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 10),
      child: ListTile(
        title: Text(title),
        trailing: Text(
          value,
          style: Theme.of(context).textTheme.titleMedium?.copyWith(
                color: AppColors.cyan,
              ),
        ),
      ),
    );
  }
}
