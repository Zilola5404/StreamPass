import 'package:flutter/material.dart';

import '../main.dart' show navigateToLogin;
import '../services/auth_service.dart';
import '../services/streampass_api.dart';
import '../services/vpn_channel.dart';
import '../theme/app_theme.dart';
import 'subscription_screen.dart';

/// E10 — профиль: email, смена пароля, удаление аккаунта (BL-043).
class ProfileScreen extends StatefulWidget {
  final AuthService authService;
  final StreamPassApi api;

  const ProfileScreen({
    super.key,
    required this.authService,
    required this.api,
  });

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  UserProfile? _profile;
  bool _loading = true;
  String? _error;

  final _currentPass = TextEditingController();
  final _newPass = TextEditingController();
  final _confirmPass = TextEditingController();
  bool _changingPass = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _currentPass.dispose();
    _newPass.dispose();
    _confirmPass.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final p = await widget.authService.fetchProfile();
      if (!mounted) return;
      setState(() {
        _profile = p;
        _loading = false;
        if (p == null) _error = 'Не удалось загрузить профиль';
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _error = 'Не удалось загрузить профиль';
        _loading = false;
      });
    }
  }

  Future<void> _changePassword() async {
    final current = _currentPass.text;
    final next = _newPass.text;
    if (next.length < 8) {
      setState(() => _error = 'Пароль должен содержать минимум 8 символов.');
      return;
    }
    if (next != _confirmPass.text) {
      setState(() => _error = 'Новый пароль и подтверждение не совпадают.');
      return;
    }
    setState(() {
      _changingPass = true;
      _error = null;
    });
    final result = await widget.authService.changePassword(current, next);
    if (!mounted) return;
    setState(() => _changingPass = false);
    if (result.success) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Пароль изменён. Войдите снова.')),
      );
      navigateToLogin(context, widget.authService, widget.api);
    } else {
      setState(() => _error = result.error ?? 'Не удалось сменить пароль');
    }
  }

  Future<void> _deleteAccount() async {
    final first = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: AppColors.surface,
        title: const Text('Удалить аккаунт?'),
        content: const Text(
          'Все данные аккаунта будут удалены. Это действие нельзя отменить.',
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Отмена')),
          TextButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Продолжить', style: TextStyle(color: AppColors.danger)),
          ),
        ],
      ),
    );
    if (first != true || !mounted) return;

    final second = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: AppColors.surface,
        title: const Text('Подтвердите удаление'),
        content: const Text('Удалить аккаунт безвозвратно?'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Отмена')),
          FilledButton(
            style: FilledButton.styleFrom(backgroundColor: AppColors.danger),
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Удалить'),
          ),
        ],
      ),
    );
    if (second != true || !mounted) return;

    try {
      await VpnChannel.disconnect();
    } catch (_) {}
    final result = await widget.authService.deleteAccount();
    if (!mounted) return;
    if (result.success) {
      navigateToLogin(context, widget.authService, widget.api);
    } else {
      setState(() => _error = result.error ?? 'Не удалось удалить аккаунт');
    }
  }

  String _fmt(DateTime? d) {
    if (d == null) return '—';
    final local = d.toLocal();
    return '${local.day.toString().padLeft(2, '0')}.'
        '${local.month.toString().padLeft(2, '0')}.'
        '${local.year}';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Профиль'),
        backgroundColor: Colors.transparent,
        elevation: 0,
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: const EdgeInsets.all(20),
              children: [
                Text('Email', style: Theme.of(context).textTheme.labelLarge),
                const SizedBox(height: 6),
                Text(
                  _profile?.email ?? '—',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: 8),
                Text(
                  'Создан: ${_fmt(_profile?.createdAt)}',
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
                if (_profile?.subscriptionActiveUntil != null) ...[
                  const SizedBox(height: 4),
                  Text(
                    'Подписка до: ${_fmt(_profile!.subscriptionActiveUntil)}',
                    style: Theme.of(context).textTheme.bodyMedium,
                  ),
                ],
                const SizedBox(height: 8),
                TextButton(
                  onPressed: () => Navigator.of(context).push(
                    MaterialPageRoute(
                      builder: (_) => SubscriptionScreen(api: widget.api),
                    ),
                  ),
                  child: const Text('Управление подпиской'),
                ),
                const Divider(height: 36),
                Text('Смена пароля', style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 12),
                TextField(
                  controller: _currentPass,
                  obscureText: true,
                  decoration: const InputDecoration(hintText: 'Текущий пароль'),
                  style: const TextStyle(color: AppColors.textPrimary),
                ),
                const SizedBox(height: 10),
                TextField(
                  controller: _newPass,
                  obscureText: true,
                  decoration: const InputDecoration(hintText: 'Новый пароль'),
                  style: const TextStyle(color: AppColors.textPrimary),
                ),
                const SizedBox(height: 10),
                TextField(
                  controller: _confirmPass,
                  obscureText: true,
                  decoration: const InputDecoration(hintText: 'Повтор нового пароля'),
                  style: const TextStyle(color: AppColors.textPrimary),
                ),
                const SizedBox(height: 16),
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: _changingPass ? null : _changePassword,
                    child: _changingPass
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: AppColors.bg,
                            ),
                          )
                        : const Text('Сменить пароль'),
                  ),
                ),
                if (_error != null) ...[
                  const SizedBox(height: 16),
                  Text(_error!, style: const TextStyle(color: AppColors.danger)),
                ],
                const Divider(height: 48),
                OutlinedButton(
                  onPressed: _deleteAccount,
                  style: OutlinedButton.styleFrom(
                    foregroundColor: AppColors.danger,
                    side: const BorderSide(color: AppColors.danger),
                    padding: const EdgeInsets.symmetric(vertical: 14),
                  ),
                  child: const Text('Удалить аккаунт'),
                ),
              ],
            ),
    );
  }
}
