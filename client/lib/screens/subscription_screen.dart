import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';

import '../services/streampass_api.dart';
import '../theme/app_theme.dart';

/// E06 — статус, тарифы, оплата, история (BL-048).
class SubscriptionScreen extends StatefulWidget {
  final StreamPassApi api;
  const SubscriptionScreen({super.key, required this.api});

  @override
  State<SubscriptionScreen> createState() => _SubscriptionScreenState();
}

class _SubscriptionScreenState extends State<SubscriptionScreen>
    with WidgetsBindingObserver {
  SubscriptionInfo? _info;
  List<PlanInfo> _plans = const [];
  List<PaymentRecord> _payments = const [];
  String _selectedPlan = 'month';
  bool _loading = true;
  bool _payLoading = false;
  bool _awaitingPaymentReturn = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _load();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      _load();
      if (_awaitingPaymentReturn && mounted) {
        setState(() => _awaitingPaymentReturn = false);
      }
    }
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final info = await widget.api.fetchSubscription();
      List<PlanInfo> plans = const [];
      List<PaymentRecord> payments = const [];
      try {
        plans = await widget.api.fetchPlans();
      } catch (_) {}
      try {
        payments = await widget.api.fetchPayments();
      } catch (_) {}
      if (!mounted) return;
      setState(() {
        _info = info;
        _plans = plans;
        _payments = payments;
        if (_plans.isNotEmpty &&
            !_plans.any((p) => p.code == _selectedPlan)) {
          _selectedPlan = _plans.first.code;
        }
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
      final url = await widget.api.createPayment(planCode: _selectedPlan);
      final uri = Uri.parse(url);
      final opened = await launchUrl(uri, mode: LaunchMode.externalApplication);
      if (!opened) throw Exception('launch failed');
      if (mounted) {
        setState(() => _awaitingPaymentReturn = true);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text(
              'После оплаты вернитесь в приложение — статус обновится автоматически.',
            ),
            duration: Duration(seconds: 5),
          ),
        );
      }
    } catch (e) {
      if (!mounted) return;
      setState(() => _error = 'Не удалось открыть страницу оплаты. Попробуйте позже');
    } finally {
      if (mounted) setState(() => _payLoading = false);
    }
  }

  Future<void> _cancel() async {
    final until = _info?.activeUntil;
    final untilText = until != null
        ? 'Доступ сохранится до ${_fmtDate(until)}.'
        : 'Доступ сохранится до конца оплаченного периода.';
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: AppColors.surface,
        title: const Text('Отменить подписку?'),
        content: Text('Отменить автопродление? $untilText'),
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
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(untilText)),
        );
      }
      await _load();
    } catch (e) {
      if (!mounted) return;
      setState(() => _error = 'Не удалось отменить подписку');
    }
  }

  String _fmtDate(DateTime d) {
    final local = d.toLocal();
    final dd = local.day.toString().padLeft(2, '0');
    final mm = local.month.toString().padLeft(2, '0');
    return '$dd.$mm.${local.year}';
  }

  String _statusLabel(String status) {
    switch (status.toUpperCase()) {
      case 'SUCCEEDED':
        return 'Оплачен';
      case 'PENDING':
        return 'Ожидает';
      case 'FAILED':
        return 'Ошибка';
      default:
        return status;
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
          : RefreshIndicator(
              onRefresh: _load,
              child: ListView(
                padding: const EdgeInsets.all(20),
                children: [
                  _StatusCard(info: _info),
                  if (_error != null) ...[
                    const SizedBox(height: 16),
                    Text(_error!, style: const TextStyle(color: AppColors.danger)),
                  ],
                  if (_info?.isActive != true) ...[
                    const SizedBox(height: 24),
                    Text('Тариф', style: Theme.of(context).textTheme.titleMedium),
                    const SizedBox(height: 12),
                    if (_plans.isEmpty)
                      const Text(
                        'Тарифы временно недоступны',
                        style: TextStyle(color: AppColors.textSecondary),
                      )
                    else
                      ..._plans.map((p) {
                        final selected = p.code == _selectedPlan;
                        return Padding(
                          padding: const EdgeInsets.only(bottom: 8),
                          child: Material(
                            color: selected
                                ? AppColors.cyan.withOpacity(0.12)
                                : AppColors.surface,
                            borderRadius: BorderRadius.circular(14),
                            child: InkWell(
                              borderRadius: BorderRadius.circular(14),
                              onTap: () => setState(() => _selectedPlan = p.code),
                              child: Container(
                                padding: const EdgeInsets.all(16),
                                decoration: BoxDecoration(
                                  borderRadius: BorderRadius.circular(14),
                                  border: Border.all(
                                    color: selected
                                        ? AppColors.cyan
                                        : Colors.white.withOpacity(0.08),
                                  ),
                                ),
                                child: Row(
                                  children: [
                                    Expanded(
                                      child: Column(
                                        crossAxisAlignment: CrossAxisAlignment.start,
                                        children: [
                                          Text(
                                            p.title,
                                            style: Theme.of(context).textTheme.titleMedium,
                                          ),
                                          Text(
                                            '${p.periodDays} дн.',
                                            style: Theme.of(context).textTheme.bodyMedium,
                                          ),
                                        ],
                                      ),
                                    ),
                                    Text(
                                      '${p.amountRub} ₽',
                                      style: Theme.of(context).textTheme.titleMedium,
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          ),
                        );
                      }),
                    const SizedBox(height: 16),
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
                            : const Text('Перейти к оплате'),
                      ),
                    ),
                  ] else ...[
                    const SizedBox(height: 24),
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
                  const SizedBox(height: 32),
                  Text('История платежей', style: Theme.of(context).textTheme.titleMedium),
                  const SizedBox(height: 12),
                  if (_payments.isEmpty)
                    const Text(
                      'Платежей пока нет',
                      style: TextStyle(color: AppColors.textSecondary),
                    )
                  else
                    ..._payments.map((p) {
                      return ListTile(
                        contentPadding: EdgeInsets.zero,
                        title: Text('${p.amountRub} ₽ · ${_statusLabel(p.status)}'),
                        subtitle: Text(
                          p.createdAt != null ? _fmtDate(p.createdAt!) : p.id,
                          style: const TextStyle(color: AppColors.textSecondary),
                        ),
                      );
                    }),
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
