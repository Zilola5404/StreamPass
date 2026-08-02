import 'package:flutter/material.dart';

import '../theme/app_theme.dart';

/// Placeholder rather than a fake dashboard: the backend does not track
/// per-session traffic, uptime, or relay-switch counts yet (ТЗ §18
/// monitoring covers server-side metrics, not a per-user client view) —
/// showing invented numbers here would be actively misleading.
class StatisticsScreen extends StatelessWidget {
  const StatisticsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Статистика'),
        backgroundColor: Colors.transparent,
        elevation: 0,
      ),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 32),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.bar_chart_rounded, size: 56, color: AppColors.textSecondary),
              const SizedBox(height: 20),
              Text('Раздел в разработке', style: Theme.of(context).textTheme.titleMedium),
              const SizedBox(height: 8),
              Text(
                'Статистика трафика и подключений появится здесь позже.',
                style: Theme.of(context).textTheme.bodyMedium,
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ),
      ),
    );
  }
}