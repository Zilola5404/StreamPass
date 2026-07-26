import 'dart:async';

import 'package:flutter/material.dart';

import '../services/settings_service.dart';
import '../services/streampass_api.dart';
import '../services/vpn_channel.dart';
import '../theme/app_theme.dart';
import '../widgets/connect_orb.dart';
import 'settings_screen.dart';

class HomeScreen extends StatefulWidget {
  final StreamPassApi api;

  const HomeScreen({super.key, required this.api});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  ConnState _state = ConnState.disconnected;
  RelayServer? _selectedRelay;
  int? _pingMs;
  String? _errorMessage;
  bool _autoMode = true;
  bool _loadingRelay = true;
  StreamSubscription<VpnStatusUpdate>? _sub;

  @override
  void initState() {
    super.initState();
    _sub = VpnChannel.statusStream.listen(_onStatus);
    _loadStartupData();
  }

  Future<void> _loadStartupData() async {
    try {
      final servers = await widget.api.fetchServers();
      final healthy = servers.where((server) => server.healthy).toList()
        ..sort((a, b) {
          final load = a.loadRatio.compareTo(b.loadRatio);
          if (load != 0) return load;
          return a.rttMs.compareTo(b.rttMs);
        });
      if (!mounted) return;
      setState(() {
        _selectedRelay = healthy.isNotEmpty
            ? healthy.first
            : servers.isNotEmpty
                ? servers.first
                : null;
        _pingMs = _selectedRelay?.rttMs;
        _loadingRelay = false;
      });
      await _maybeAutoConnect();
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _loadingRelay = false;
        _errorMessage = e.toString();
      });
    }
  }

  Future<void> _maybeAutoConnect() async {
    final settings = await SettingsService().load();
    if (settings.autoConnect && _state == ConnState.disconnected) {
      await _toggleConnection();
    }
  }

  @override
  void dispose() {
    _sub?.cancel();
    super.dispose();
  }

  void _onStatus(VpnStatusUpdate update) {
    setState(() {
      switch (update.event) {
        case VpnEvent.connecting:
          _state = ConnState.connecting;
        case VpnEvent.connected:
          _state = ConnState.connected;
          _pingMs = update.pingMs ?? _selectedRelay?.rttMs;
        case VpnEvent.disconnected:
          _state = ConnState.disconnected;
          _pingMs = _selectedRelay?.rttMs;
        case VpnEvent.permissionDenied:
          _state = ConnState.error;
          _errorMessage = 'Нужно разрешение на VPN-соединение';
        case VpnEvent.error:
          _state = ConnState.error;
          _errorMessage = update.errorMessage ?? 'Ошибка подключения';
      }
    });
  }

  Future<void> _toggleConnection() async {
    if (_state == ConnState.connected) {
      await VpnChannel.disconnect();
      return;
    }

    setState(() {
      _errorMessage = null;
      _state = ConnState.connecting;
    });
    final accepted = await VpnChannel.connect();
    if (!accepted && mounted) {
      setState(() => _state = ConnState.disconnected);
    }
  }

  @override
  Widget build(BuildContext context) {
    final statusText = switch (_state) {
      ConnState.connected => 'Подключено',
      ConnState.connecting => 'Подключение...',
      ConnState.disconnected => 'Готов к подключению',
      ConnState.error => _errorMessage ?? 'Ошибка',
    };

    return Scaffold(
      body: Stack(
        children: [
          const _AmbientBackground(),
          SafeArea(
            child: Column(
              children: [
                _TopBar(
                  onSettings: () => Navigator.of(context).push(
                    MaterialPageRoute(builder: (_) => const SettingsScreen()),
                  ),
                ),
                Expanded(
                  child: ListView(
                    padding: const EdgeInsets.fromLTRB(22, 8, 22, 24),
                    children: [
                      Center(
                        child: _StatusChip(
                          label: _state == ConnState.connected
                              ? 'Система активна'
                              : 'Авто-маршрут готов',
                          active: _state != ConnState.error,
                        ),
                      ),
                      const SizedBox(height: 18),
                      Center(
                        child: ConnectOrb(
                          state: _state,
                          label: statusText,
                          subtitle: _state == ConnState.connected
                              ? '00:03:47'
                              : 'Smart routing',
                          onTap: _toggleConnection,
                        ),
                      ),
                      const SizedBox(height: 24),
                      _RelayCard(
                        relay: _selectedRelay,
                        pingMs: _pingMs,
                        loading: _loadingRelay,
                      ),
                      const SizedBox(height: 14),
                      _RouteCard(
                        state: _state,
                        autoMode: _autoMode,
                        onAutoModeChanged: (value) =>
                            setState(() => _autoMode = value),
                      ),
                    ],
                  ),
                ),
                const _BottomNav(),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _TopBar extends StatelessWidget {
  final VoidCallback onSettings;

  const _TopBar({required this.onSettings});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(18, 10, 18, 8),
      child: Row(
        children: [
          IconButton(
            onPressed: () {},
            icon: const Icon(Icons.menu_rounded),
          ),
          Expanded(
            child: Text(
              'StreamPass',
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.titleMedium,
            ),
          ),
          IconButton(
            onPressed: onSettings,
            icon: const Icon(Icons.workspace_premium_rounded,
                color: AppColors.amber),
          ),
        ],
      ),
    );
  }
}

class _StatusChip extends StatelessWidget {
  final String label;
  final bool active;

  const _StatusChip({required this.label, required this.active});

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        color: AppColors.surface.withOpacity(0.78),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: AppColors.cyan.withOpacity(0.18)),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 9),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 9,
              height: 9,
              decoration: BoxDecoration(
                color: active ? AppColors.green : AppColors.danger,
                shape: BoxShape.circle,
                boxShadow: [
                  BoxShadow(
                    color: (active ? AppColors.green : AppColors.danger)
                        .withOpacity(0.5),
                    blurRadius: 12,
                  ),
                ],
              ),
            ),
            const SizedBox(width: 9),
            Text(label, style: Theme.of(context).textTheme.bodySmall),
          ],
        ),
      ),
    );
  }
}

class _RelayCard extends StatelessWidget {
  final RelayServer? relay;
  final int? pingMs;
  final bool loading;

  const _RelayCard({
    required this.relay,
    required this.pingMs,
    required this.loading,
  });

  @override
  Widget build(BuildContext context) {
    final title = loading
        ? 'Ищем лучший relay'
        : relay?.region ?? 'Relay не найден';
    final subtitle = relay == null
        ? 'Проверьте backend и список серверов'
        : relay!.healthy
            ? 'Основной relay'
            : 'Резервный relay';

    return _GlassCard(
      child: Row(
        children: [
          _FlagBadge(region: relay?.region),
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
          if (pingMs != null)
            Text(
              '$pingMs ms',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    color: AppColors.green,
                    fontSize: 15,
                  ),
            ),
          const SizedBox(width: 10),
          const Icon(Icons.signal_cellular_alt_rounded,
              color: AppColors.green, size: 20),
        ],
      ),
    );
  }
}

class _RouteCard extends StatelessWidget {
  final ConnState state;
  final bool autoMode;
  final ValueChanged<bool> onAutoModeChanged;

  const _RouteCard({
    required this.state,
    required this.autoMode,
    required this.onAutoModeChanged,
  });

  @override
  Widget build(BuildContext context) {
    return _GlassCard(
      child: Row(
        children: [
          Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(
              gradient: AppColors.orbGradient,
              borderRadius: BorderRadius.circular(16),
              boxShadow: [
                BoxShadow(
                  color: AppColors.violet.withOpacity(0.32),
                  blurRadius: 20,
                ),
              ],
            ),
            child: const Icon(Icons.hub_rounded, color: Colors.white),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  state == ConnState.connected
                      ? 'Маршрут оптимален'
                      : 'Автоматический маршрут',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: 4),
                Text(
                  'DIRECT для локальных сервисов, RELAY для зарубежных',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ],
            ),
          ),
          Switch(
            value: autoMode,
            activeColor: AppColors.green,
            onChanged: onAutoModeChanged,
          ),
        ],
      ),
    );
  }
}

class _BottomNav extends StatelessWidget {
  const _BottomNav();

  @override
  Widget build(BuildContext context) {
    final items = [
      (Icons.home_rounded, 'Главная', true),
      (Icons.bar_chart_rounded, 'Статистика', false),
      (Icons.public_rounded, 'Серверы', false),
      (Icons.settings_rounded, 'Настройки', false),
    ];

    return Container(
      margin: const EdgeInsets.fromLTRB(18, 0, 18, 14),
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 10),
      decoration: BoxDecoration(
        color: AppColors.surface.withOpacity(0.82),
        borderRadius: BorderRadius.circular(26),
        border: Border.all(color: Colors.white.withOpacity(0.08)),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceAround,
        children: [
          for (final item in items)
            Expanded(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    item.$1,
                    color: item.$3 ? AppColors.cyan : AppColors.textSecondary,
                  ),
                  const SizedBox(height: 4),
                  Text(
                    item.$2,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: item.$3
                              ? AppColors.cyan
                              : AppColors.textSecondary,
                          fontSize: 11,
                        ),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }
}

class _GlassCard extends StatelessWidget {
  final Widget child;

  const _GlassCard({required this.child});

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        color: AppColors.surface.withOpacity(0.78),
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: Colors.white.withOpacity(0.08)),
        boxShadow: [
          BoxShadow(
            color: AppColors.cyan.withOpacity(0.08),
            blurRadius: 24,
            offset: const Offset(0, 14),
          ),
        ],
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: child,
      ),
    );
  }
}

class _FlagBadge extends StatelessWidget {
  final String? region;

  const _FlagBadge({required this.region});

  @override
  Widget build(BuildContext context) {
    final flag = _flagLabel(region);

    return Container(
      width: 48,
      height: 48,
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: Colors.black.withOpacity(0.35),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: Colors.white.withOpacity(0.08)),
      ),
      child: Text(
        flag,
        style: Theme.of(context).textTheme.titleMedium?.copyWith(fontSize: 13),
      ),
    );
  }

  String _flagLabel(String? region) {
    final value = region?.toLowerCase() ?? '';
    if (value.contains('germany') || value.contains('frankfurt')) return 'DE';
    if (value.contains('netherlands') || value.contains('amsterdam')) {
      return 'NL';
    }
    if (value.contains('warsaw') || value.contains('poland')) return 'PL';
    if (value.contains('finland') || value.contains('helsinki')) return 'FI';
    return 'GL';
  }
}

class _AmbientBackground extends StatelessWidget {
  const _AmbientBackground();

  @override
  Widget build(BuildContext context) {
    return Stack(
      children: [
        const DecoratedBox(
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [Color(0xFF020612), Color(0xFF07111F)],
            ),
          ),
          child: SizedBox.expand(),
        ),
        Positioned(
          top: -150,
          right: -130,
          child: _GlowOrb(color: AppColors.violet.withOpacity(0.28), size: 320),
        ),
        Positioned(
          bottom: 160,
          left: -170,
          child: _GlowOrb(color: AppColors.cyan.withOpacity(0.18), size: 300),
        ),
      ],
    );
  }
}

class _GlowOrb extends StatelessWidget {
  final Color color;
  final double size;

  const _GlowOrb({required this.color, required this.size});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        gradient: RadialGradient(
          colors: [color, color.withOpacity(0)],
        ),
      ),
    );
  }
}
