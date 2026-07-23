import 'package:flutter/material.dart';
import '../theme/app_theme.dart';
import '../services/settings_service.dart';
import 'exclusions_screen.dart';
import 'diagnostics_screen.dart';

class SettingsScreen extends StatefulWidget {
  const SettingsScreen({super.key});

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  final _service = SettingsService();
  AppSettings _settings = const AppSettings();
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _service.load().then((s) {
      if (!mounted) return;
      setState(() {
        _settings = s;
        _loading = false;
      });
    });
  }

  Future<void> _updateAutostart(bool value) async {
    setState(() => _settings = _settings.copyWith(autostart: value));
    await _service.setAutostart(value);
    // Native side: a BOOT_COMPLETED BroadcastReceiver reads this flag
    // (shared via SharedPreferences / DataStore) to decide whether to
    // start StreamPassVpnService on device boot. See android README note.
  }

  Future<void> _updateAutoConnect(bool value) async {
    setState(() => _settings = _settings.copyWith(autoConnect: value));
    await _service.setAutoConnect(value);
    // "Автоподключение" = connect automatically whenever the app is opened
    // or the network becomes available, without a manual tap on the orb.
  }

  Future<void> _updateAutoRelay(bool value) async {
    setState(() => _settings = _settings.copyWith(autoSelectRelay: value));
    await _service.setAutoSelectRelay(value);
    // When off, the user is expected to be able to pick a relay manually
    // from GET /api/v1/servers — that picker screen isn't built yet.
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
              _settings.exclusions.isEmpty
                  ? 'Нет исключений'
                  : '${_settings.exclusions.length} домен(ов) всегда напрямую',
            ),
            trailing: const Icon(Icons.chevron_right, color: AppColors.textSecondary),
            onTap: () async {
              final updated = await Navigator.of(context).push<List<String>>(
                MaterialPageRoute(
                  builder: (_) => ExclusionsScreen(initial: _settings.exclusions),
                ),
              );
              if (updated != null) {
                setState(() => _settings = _settings.copyWith(exclusions: updated));
                await _service.setExclusions(updated);
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
      padding: const EdgeInsets.fromLTRB(16, 20, 16, 8),
      child: Text(
        text.toUpperCase(),
        style: Theme.of(context).textTheme.bodySmall?.copyWith(
              letterSpacing: 1.1,
              color: AppColors.textSecondary,
            ),
      ),
    );
  }
}
