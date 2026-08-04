import 'package:flutter/material.dart';

import '../services/auth_service.dart';
import '../services/region_catalog.dart';
import '../services/settings_service.dart';
import '../services/streampass_api.dart';
import '../main.dart' show navigateToLogin;
import '../theme/app_theme.dart';

/// Region / relay picker (BL-026). Groups healthy relays by region and
/// persists the user's preferred region + server id.
class ServersScreen extends StatefulWidget {
  final StreamPassApi api;
  final AuthService? authService;
  final String? selectedServerId;

  const ServersScreen({
    super.key,
    required this.api,
    this.authService,
    this.selectedServerId,
  });

  @override
  State<ServersScreen> createState() => _ServersScreenState();
}

class _ServersScreenState extends State<ServersScreen> {
  final _settings = SettingsService();
  List<RelayServer>? _servers;
  AppSettings _prefs = const AppSettings();
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() => _error = null);
    try {
      final prefs = await _settings.load();
      final servers = await widget.api.fetchServers();
      if (!mounted) return;
      setState(() {
        _prefs = prefs;
        _servers = servers;
      });
    } on SessionExpiredException catch (e) {
      if (!mounted) return;
      if (widget.authService != null) {
        navigateToLogin(context, widget.authService!, widget.api);
      } else {
        setState(() => _error = e.message);
      }
    } catch (e) {
      if (!mounted) return;
      setState(() =>
          _error = e is ApiException ? e.message : 'Не удалось загрузить список серверов');
    }
  }

  Future<void> _selectAuto() async {
    await _settings.setPreferredRegion('');
    await _settings.setPreferredServerId('');
    await _settings.setAutoSelectRelay(true);
    if (!mounted) return;
    Navigator.of(context).pop(const RelayPickResult(auto: true));
  }

  Future<void> _selectServer(RelayServer server) async {
    final code = normalizeRegionCode(server.region);
    await _settings.setPreferredRegion(code);
    await _settings.setPreferredServerId(server.id);
    await _settings.setAutoSelectRelay(false);
    if (!mounted) return;
    Navigator.of(context).pop(RelayPickResult(server: server));
  }

  Future<void> _selectRegion(String code) async {
    await _settings.setPreferredRegion(code);
    await _settings.setPreferredServerId('');
    await _settings.setAutoSelectRelay(true);
    if (!mounted) return;
    Navigator.of(context).pop(RelayPickResult(regionCode: code));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Регионы'),
        backgroundColor: Colors.transparent,
        elevation: 0,
      ),
      body: RefreshIndicator(
        onRefresh: _load,
        child: _error != null
            ? ListView(
                children: [
                  const SizedBox(height: 60),
                  Center(
                    child: Text(_error!, style: const TextStyle(color: AppColors.danger)),
                  ),
                ],
              )
            : _servers == null
                ? const Center(child: CircularProgressIndicator())
                : _buildList(),
      ),
    );
  }

  Widget _buildList() {
    final servers = _servers!;
    if (servers.isEmpty) {
      return ListView(
        children: const [
          SizedBox(height: 60),
          Center(child: Text('Нет доступных серверов')),
        ],
      );
    }

    final byRegion = <String, List<RelayServer>>{};
    for (final s in servers) {
      final code = normalizeRegionCode(s.region);
      byRegion.putIfAbsent(code, () => []).add(s);
    }

    final orderedCodes = [
      for (final info in fallbackRegions)
        if (byRegion.containsKey(info.code)) info.code,
      ...byRegion.keys.where((c) => !fallbackRegions.any((i) => i.code == c)),
    ];

    final selectedId = widget.selectedServerId ?? _prefs.preferredServerId;

    return ListView(
      padding: const EdgeInsets.all(18),
      children: [
        _ChoiceTile(
          title: 'Автовыбор',
          subtitle: 'Лучший relay по нагрузке и RTT',
          selected: _prefs.autoSelectRelay && _prefs.preferredRegion.isEmpty,
          onTap: _selectAuto,
        ),
        const SizedBox(height: 16),
        for (final code in orderedCodes) ...[
          InkWell(
            onTap: () => _selectRegion(code),
            borderRadius: BorderRadius.circular(12),
            child: Padding(
              padding: const EdgeInsets.symmetric(vertical: 6),
              child: Row(
                children: [
                  Text(
                    regionFlag(code),
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Text(
                      regionLabel(code),
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                  ),
                  if (_prefs.autoSelectRelay && _prefs.preferredRegion == code)
                    const Icon(Icons.check_rounded, color: AppColors.green, size: 20),
                ],
              ),
            ),
          ),
          const SizedBox(height: 8),
          for (final server in byRegion[code]!) ...[
            _ServerTile(
              server: server,
              selected: selectedId == server.id && !_prefs.autoSelectRelay,
              onTap: () => _selectServer(server),
            ),
            const SizedBox(height: 10),
          ],
          const SizedBox(height: 8),
        ],
      ],
    );
  }
}

class RelayPickResult {
  final bool auto;
  final String? regionCode;
  final RelayServer? server;

  const RelayPickResult({this.auto = false, this.regionCode, this.server});
}

class _ChoiceTile extends StatelessWidget {
  final String title;
  final String subtitle;
  final bool selected;
  final VoidCallback onTap;

  const _ChoiceTile({
    required this.title,
    required this.subtitle,
    required this.selected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Material(
      color: AppColors.surface,
      borderRadius: BorderRadius.circular(18),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(18),
        child: Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(18),
            border: Border.all(
              color: selected ? AppColors.green.withOpacity(0.5) : Colors.white.withOpacity(0.08),
            ),
          ),
          child: Row(
            children: [
              Icon(
                Icons.auto_awesome_rounded,
                color: selected ? AppColors.green : AppColors.textSecondary,
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(title, style: Theme.of(context).textTheme.titleMedium),
                    const SizedBox(height: 4),
                    Text(subtitle, style: Theme.of(context).textTheme.bodySmall),
                  ],
                ),
              ),
              if (selected) const Icon(Icons.check_circle_rounded, color: AppColors.green),
            ],
          ),
        ),
      ),
    );
  }
}

class _ServerTile extends StatelessWidget {
  final RelayServer server;
  final bool selected;
  final VoidCallback onTap;

  const _ServerTile({
    required this.server,
    required this.selected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Material(
      color: AppColors.surface,
      borderRadius: BorderRadius.circular(18),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(18),
        child: Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(18),
            border: Border.all(
              color: selected ? AppColors.green.withOpacity(0.5) : Colors.white.withOpacity(0.08),
            ),
          ),
          child: Row(
            children: [
              Icon(
                server.healthy ? Icons.check_circle_rounded : Icons.error_rounded,
                color: server.healthy ? AppColors.green : AppColors.danger,
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(server.id, style: Theme.of(context).textTheme.titleMedium),
                    const SizedBox(height: 4),
                    Text(
                      server.healthy
                          ? 'Доступен · ${server.rttMs} ms'
                          : 'Недоступен',
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ],
                ),
              ),
              if (selected) const Icon(Icons.check_circle_rounded, color: AppColors.green),
            ],
          ),
        ),
      ),
    );
  }
}
