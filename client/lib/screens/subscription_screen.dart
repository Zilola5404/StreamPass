import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';

import '../services/streampass_api.dart';
import '../theme/app_theme.dart';

/// Did not exist anywhere in this codebase — POST /payments and
/// GET /subscription were fully built and tested on the backend, but
/// nothing on the client ever called them. This screen closes that gap.
class SubscriptionScreen extends StatefulWidget {
  final StreamPassApi api;
  const SubscriptionScreen({super.key, required this.api});

  @override
  State<SubscriptionScreen> createState() => _SubscriptionScreenState();
}

class _SubscriptionScreenState extends State<SubscriptionScreen> {
  SubscriptionInfo? _info;
  bool _loading = true;
  bool _payLoading = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final info = await widget.api.fetchSubscription();
      if (!mounted) return;
      setState(() {
        _info = info;
        _loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = 'Не удалось загрузить статус подписки';
        _loading = false;
      });
    }
  }

  Future<void> _pay() async {
    setState(() {
      _payLoading = true;
      _error = null;
    });
    try {
      final url = await widget.api.createPayment();
      final uri = Uri.parse(url);
      final opened = await launchUrl(uri, mode: LaunchMode.externalApplication);
      if (!opened) throw Exception('launch failed');
    } catch (e) {
      if (!mounted) return;
      setState(() => _error = 'Не удалось открыть страницу оплаты. Попробуйте позже');
    } finally {
      if (mounted) setState(() => _payLoading = false);
    }
  }

  Future<void> _cancel() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: AppColors.surface,
        title: const Text('Отменить подписку?'),
        content: const Text('Доступ к VPN будет остановлен немедленно.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text('Не отменять'),
          ),
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            child: const Text('Отменить', style: TextStyle(color: AppColors.danger)),
          ),
        ],
      ),
    );
    if (confirmed != true) return;

    try {
      await widget.api.cancelSubscription();
      await _load();
    } catch (e) {
      if (!mounted) return;
      setState(() => _error = 'Не удалось отменить подписку');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Подписка'),
        backgroundColor: Colors.transparent,
        elevation: 0,
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : Padding(
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _StatusCard(info: _info),
                  if (_error != null) ...[
                    const SizedBox(height: 16),
                    Text(_error!, style: const TextStyle(color: AppColors.danger)),
                  ],
                  const SizedBox(height: 24),
                  if (_info?.isActive != true)
                    SizedBox(
                      width: double.infinity,
                      child: ElevatedButton(
                        onPressed: _payLoading ? null : _pay,
                        child: _payLoading
                            ? const SizedBox(
                                height: 20,
                                width: 20,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                  color: AppColors.bg,
                                ),
                              )
                            : const Text('Оплатить подписку'),
                      ),
                    )
                  else
                    SizedBox(
                      width: double.infinity,
                      child: OutlinedButton(
                        onPressed: _cancel,
                        style: OutlinedButton.styleFrom(
                          foregroundColor: AppColors.danger,
                          side: const BorderSide(color: AppColors.danger),
                          padding: const EdgeInsets.symmetric(vertical: 16),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(18),
                          ),
                        ),
                        child: const Text('Отменить подписку'),
                      ),
                    ),
                ],
              ),
            ),
    );
  }
}

class _StatusCard extends StatelessWidget {
  final SubscriptionInfo? info;
  const _StatusCard({required this.info});

  @override
  Widget build(BuildContext context) {
    final active = info?.isActive ?? false;
    final until = info?.activeUntil;

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.white.withOpacity(0.08)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.workspace_premium_rounded,
                  color: active ? AppColors.amber : AppColors.textSecondary),
              const SizedBox(width: 10),
              Text(
                active ? 'Premium активна' : 'Подписка не активна',
                style: Theme.of(context).textTheme.titleMedium,
              ),
            ],
          ),
          if (active && until != null) ...[
            const SizedBox(height: 10),
            Text(
              'Действует до ${_formatDate(until)}',
              style: Theme.of(context).textTheme.bodyMedium,
            ),
          ],
          if (!active) ...[
            const SizedBox(height: 10),
            Text(
              'Оформите подписку, чтобы подключаться к relay-серверам StreamPass.',
              style: Theme.of(context).textTheme.bodyMedium,
            ),
          ],
        ],
      ),
    );
  }

  String _formatDate(DateTime d) =>
      '${d.day.toString().padLeft(2, '0')}.${d.month.toString().padLeft(2, '0')}.${d.year}';
}