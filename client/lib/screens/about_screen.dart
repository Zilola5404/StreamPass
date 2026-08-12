import 'package:flutter/material.dart';

import '../build_info.dart';
import '../theme/app_theme.dart';

/// E05 «О приложении» (BL-052). Uses stock [AppBar] (no SpAppBar wrappers).
class AboutScreen extends StatelessWidget {
  const AboutScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('О приложении'),
        backgroundColor: Colors.transparent,
        elevation: 0,
      ),
      body: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          Text('StreamPass', style: Theme.of(context).textTheme.headlineSmall),
          const SizedBox(height: 8),
          Text(
            'Интеллектуальная маршрутизация трафика',
            style: Theme.of(context).textTheme.bodyMedium,
          ),
          const SizedBox(height: 24),
          _row(context, 'Версия', BuildInfo.version),
          _row(context, 'Сборка', '${BuildInfo.buildNumber}'),
          _row(context, 'Connect flow', BuildInfo.connectFlow),
          _row(context, 'Канал', 'Android MVP'),
          const SizedBox(height: 24),
          Text(
            'Оплата: Telegram Stars / USDT (TRC20). ЮKassa live отложена.',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: AppColors.textSecondary,
                ),
          ),
        ],
      ),
    );
  }

  Widget _row(BuildContext context, String k, String v) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        children: [
          Expanded(
            child: Text(k, style: const TextStyle(color: AppColors.textSecondary)),
          ),
          Text(v, style: Theme.of(context).textTheme.titleMedium),
        ],
      ),
    );
  }
}
