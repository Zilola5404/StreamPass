import 'dart:async';
import 'package:flutter/material.dart';
import '../theme/app_theme.dart';
import '../widgets/connect_orb.dart';
import '../services/vpn_channel.dart';
import '../services/settings_service.dart';
import 'settings_screen.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});
  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  ConnState _state = ConnState.disconnected;
  String _server = '—';
  int? _pingMs;
  String? _errorMessage;
  bool _autoMode = true;
  StreamSubscription<VpnStatusUpdate>? _sub;

  @override
  void initState() {
    super.initState();
    _sub = VpnChannel.statusStream.listen(_onStatus);
    _maybeAutoConnect();
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
          _server = update.relayName ?? _server;
          _pingMs = update.pingMs;
        case VpnEvent.disconnected:
          _state = ConnState.disconnected;
          _pingMs = null;
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
    } else {
      setState(() {
        _errorMessage = null;
        _state = ConnState.connecting;
      });
      final accepted = await VpnChannel.connect();
      if (!accepted && mounted) {
        setState(() => _state = ConnState.disconnected);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final statusText = switch (_state) {
      ConnState.connected => 'Connected',
      ConnState.connecting => 'Connecting…',
      ConnState.disconnected => 'Disconnected',
      ConnState.error => _errorMessage ?? 'Error',
    };

    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        actions: [
          IconButton(
            icon: const Icon(Icons.settings_outlined),
            onPressed: () => Navigator.of(context).push(
              MaterialPageRoute(builder: (_) => const SettingsScreen()),
            ),
          ),
        ],
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(
              statusText,
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    color: _state == ConnState.error
                        ? AppColors.danger
                        : AppColors.textPrimary,
                  ),
            ),
            const SizedBox(height: 8),
            if (_state != ConnState.error)
              Text(_server, style: Theme.of(context).textTheme.bodyMedium),
            if (_pingMs != null) ...[
              const SizedBox(height: 4),
              Text('Ping $_pingMs ms',
                  style: Theme.of(context).textTheme.bodyMedium),
            ],
            const SizedBox(height: 48),
            ConnectOrb(state: _state, onTap: _toggleConnection),
            const SizedBox(height: 48),
            SwitchListTile(
              value: _autoMode,
              onChanged: (v) => setState(() => _autoMode = v),
              title: const Text('Auto Mode'),
              activeColor: AppColors.cyan,
            ),
          ],
        ),
      ),
    );
  }
}
