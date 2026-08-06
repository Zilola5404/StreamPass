import 'dart:convert';

import 'package:flutter/material.dart';
import '../theme/app_theme.dart';
import '../services/auth_service.dart';
import '../services/settings_service.dart';
import '../services/streampass_api.dart';
import '../services/vpn_channel.dart';
import '../main.dart' show navigateToLogin;
import 'exclusions_screen.dart';
import 'app_bypass_screen.dart';
import 'diagnostics_screen.dart';
import 'profile_screen.dart';

class SettingsScreen extends StatefulWidget {
  final StreamPassApi? api;
  final AuthService? authService;

  const SettingsScreen({super.key, this.api, this.authService});

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  final _service = SettingsService();
  AppSettings _settings = const AppSettings();
  bool _loading = true;
  String? _syncHint;

  @override
  void initState() {
    super.initState();
    _bootstrap();
  }

  Future<void> _bootstrap() async {
    final local = await _service.load();
    if (!mounted) return;
    setState(() {
      _settings = local;
      _loading = false;
    });
    await _pullExclusions();
  }

  Future<void> _pullExclusions() async {
    final api = widget.api;
    if (api == null) return;
    try {
      final remote = await api.fetchExclusions();
      await _service.setExclusions(remote);
      if (!mounted) return;
      setState(() {
        _settings = _settings.copyWith(exclusions: remote);
        _syncHint = null;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() => _syncHint = 'Локальный список (синхронизация недоступна)');
    }
  }

  Future<void> _saveExclusions(List<String> updated) async {
    setState(() => _settings = _settings.copyWith(exclusions: updated));
    await _service.setExclusions(updated);

    final api = widget.api;
    if (api != null) {
      try {
        final saved = await api.putExclusions(updated);
        await _service.setExclusions(saved);
        if (mounted) {
          setState(() {
            _settings = _settings.copyWith(exclusions: saved);
            _syncHint = null;
          });
        }
      } catch (_) {
        if (mounted) {
          setState(() => _syncHint = 'Сохранено локально, сервер недоступен');
        }
      }
    }

    // Hot-reload Decision Engine if VPN is up.
    try {
      final api = widget.api;
      if (api != null) {
        final ruleSet = await api.fetchRules();
        await VpnChannel.updateRules(
          rulesJson: jsonEncode(ruleSet.toJson()),
          exclusionsJson: jsonEncode(updated),
        );
      }
    } catch (_) {
      // Best-effort; connect will reload exclusions next time.
    }
  }

  Future<void> _updateAutostart(bool value) async {
    setState(() => _settings = _settings.copyWith(autostart: value));
    await _service.setAutostart(value);
  }

  Future<void> _updateAutoConnect(bool value) async {
    setState(() => _settings = _settings.copyWith(autoConnect: value));
    await _service.setAutoConnect(value);
  }

  Future<void> _updateAutoRelay(bool value) async {
    setState(() => _settings = _settings.copyWith(autoSelectRelay: value));
    await _service.setAutoSelectRelay(value);
  }

  Future<void> _confirmLogout() async {
    final auth = widget.authService;
    final api = widget.api;
    if (auth == null || api == null) return;

    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Выйти из аккаунта?'),
        content: const Text(
          'Ускоритель будет отключён, и потребуется войти снова.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text('Отмена'),
          ),
          FilledButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            child: const Text('Выйти'),
          ),
        ],
      ),
    );
    if (ok != true || !mounted) return;

    try {
      await VpnChannel.disconnect();
    } catch (_) {}
    await auth.logout();
    if (!mounted) return;
    navigateToLogin(context, auth, api);
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return const Scaffold(
        body: Center(child: CircularProgressIndicator()),
      );
    }

    return Scaffold(
      appBar: AppBar(
        title: const Text('Настройки'),
        backgroundColor: Colors.transparent,
        elevation: 0,
      ),
      body: ListView(
        children: [
          _SectionLabel('Подключение'),
          SwitchListTile(
            title: const Text('Автозапуск'),
            subtitle: const Text('Запускать StreamPass при включении устройства'),
            value: _settings.autostart,
            activeColor: AppColors.cyan,
            onChanged: _updateAutostart,
          ),
          SwitchListTile(
            title: const Text('Автоподключение'),
            subtitle: const Text('Подключаться автоматически при запуске приложения'),
            value: _settings.autoConnect,
            activeColor: AppColors.cyan,
            onChanged: _updateAutoConnect,
          ),
          SwitchListTile(
            title: const Text('Автовыбор Relay'),
            subtitle: const Text('Выбирать сервер автоматически по RTT и нагрузке'),
            value: _settings.autoSelectRelay,
            activeColor: AppColors.cyan,
            onChanged: _updateAutoRelay,
          ),
          const Divider(height: 32),
          _SectionLabel('Маршрутизация'),
          ListTile(
            title: const Text('Исключения'),
            subtitle: Text(
              _syncHint ??
                  (_settings.exclusions.isEmpty
                      ? 'Нет исключений'
                      : '${_settings.exclusions.length} домен(ов) всегда напрямую'),
            ),
            trailing: const Icon(Icons.chevron_right, color: AppColors.textSecondary),
            onTap: () async {
              final updated = await Navigator.of(context).push<List<String>>(
                MaterialPageRoute(
                  builder: (_) => ExclusionsScreen(initial: _settings.exclusions),
                ),
              );
              if (updated != null) {
                await _saveExclusions(updated);
              }
            },
          ),
          ListTile(
            title: const Text('Приложения без VPN'),
            subtitle: Text(
              _settings.bypassPackages.isEmpty
                  ? 'Дополнительно к встроенному списку (Госуслуги, банки…)'
                  : '${_settings.bypassPackages.length} приложений обходят VPN',
            ),
            trailing: const Icon(Icons.chevron_right, color: AppColors.textSecondary),
            onTap: () async {
              final updated = await Navigator.of(context).push<List<String>>(
                MaterialPageRoute(
                  builder: (_) =>
                      AppBypassScreen(initialSelected: _settings.bypassPackages),
                ),
              );
              if (updated != null) {
                setState(() =>
                    _settings = _settings.copyWith(bypassPackages: updated));
                await _service.setBypassPackages(updated);
                if (mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(
                      content: Text(
                        'Список сохранён. Переподключите StreamPass, чтобы применить.',
                      ),
                    ),
                  );
                }
              }
            },
          ),
          const Divider(height: 32),
          _SectionLabel('Поддержка'),
          ListTile(
            title: const Text('Диагностика'),
            subtitle: const Text('RTT, потери пакетов, статус relay, версия клиента'),
            trailing: const Icon(Icons.chevron_right, color: AppColors.textSecondary),
            onTap: () => Navigator.of(context).push(
              MaterialPageRoute(builder: (_) => const DiagnosticsScreen()),
            ),
          ),
          if (widget.authService != null && widget.api != null) ...[
            const Divider(height: 32),
            _SectionLabel('Аккаунт'),
            ListTile(
              title: const Text('Профиль'),
              subtitle: const Text('Email, смена пароля, удаление аккаунта'),
              leading: const Icon(Icons.person_outline, color: AppColors.textSecondary),
              trailing: const Icon(Icons.chevron_right, color: AppColors.textSecondary),
              onTap: () => Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (_) => ProfileScreen(
                    authService: widget.authService!,
                    api: widget.api!,
                  ),
                ),
              ),
            ),
            ListTile(
              title: const Text('Выйти'),
              subtitle: const Text('Завершить сеанс на этом устройстве'),
              leading: const Icon(Icons.logout, color: AppColors.textSecondary),
              onTap: _confirmLogout,
            ),
          ],
        ],
      ),
    );
  }
}

class _SectionLabel extends StatelessWidget {
  final String text;
  const _SectionLabel(this.text);

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 4),
      child: Text(
        text,
        style: Theme.of(context).textTheme.labelLarge?.copyWith(
              color: AppColors.textSecondary,
            ),
      ),
    );
  }
}
