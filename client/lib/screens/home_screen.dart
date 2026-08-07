import 'dart:async';
import 'dart:convert';
import 'dart:io' show SocketException;

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

import '../main.dart' show navigateToLogin;
import '../build_info.dart';
import '../services/connection_log.dart';
import '../services/auth_errors.dart';
import '../services/auth_service.dart';
import '../services/rule_engine_service.dart';
import '../services/settings_service.dart';
import '../services/streampass_api.dart';
import '../services/vpn_channel.dart';
import '../services/client_update.dart';
import '../services/diag_uploader.dart';
import '../services/connection_duration.dart';
import '../services/relay_picker.dart';
import '../services/region_catalog.dart';
import '../services/session_stats.dart';
import '../theme/app_theme.dart';
import '../widgets/connect_orb.dart';
import 'settings_screen.dart';
import 'subscription_screen.dart';
import 'servers_screen.dart';
import 'statistics_screen.dart';
import 'package:url_launcher/url_launcher.dart';

class HomeScreen extends StatefulWidget {
  final StreamPassApi api;
  final AuthService authService;

  const HomeScreen({super.key, required this.api, required this.authService});

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
  SubscriptionInfo? _subscription;
  bool _loadingSubscription = true;
  bool _subscriptionCheckFailed = false;
  StreamSubscription<VpnStatusUpdate>? _sub;
  final _connectLog = ConnectionLog.instance;
  late final RuleEngineService _ruleEngine = RuleEngineService(api: widget.api);
  int _pendingRulesVersion = 0;
  DateTime? _connectedSince;
  Timer? _durationTimer;
  Timer? _healthTimer;
  String _durationLabel = 'Smart routing';
  final _sessionStats = SessionStatsService();
  bool _failoverInFlight = false;
  DiagUploader? _diagUploader;

  @override
  void initState() {
    super.initState();
    _sub = VpnChannel.statusStream.listen(_onStatus);
    _diagUploader = DiagUploader(api: widget.api)..start();
    _bootstrap();
  }

  Future<void> _bootstrap() async {
    await _checkClientUpdate();
    await _loadSubscription();
    await _loadStartupData();
    await _restoreVpnState();
  }

  Future<void> _restoreVpnState() async {
    try {
      final native = await VpnChannel.fetchNativeStatus();
      if (!mounted || native == null) return;
      if (native.event != VpnEvent.connected) return;
      setState(() {
        _state = ConnState.connected;
        _connectedSince = DateTime.now();
        _durationLabel = formatConnectionDuration(Duration.zero);
        _pingMs = _effectivePing(native.pingMs);
      });
      _startDurationTimer();
      unawaited(_ruleEngine.start(initialVersion: _pendingRulesVersion));
    } catch (_) {
      // Non-fatal — UI stays on disconnected until user connects.
    }
  }

  Future<void> _checkClientUpdate() async {
    try {
      final cfg = await widget.api.fetchConfig();
      final result = evaluateClientUpdate(
        currentVersion: BuildInfo.version,
        minSupportedVersion: cfg.minSupportedClientVersion,
        latestVersion: cfg.latestClientVersion,
        downloadUrl: cfg.clientDownloadUrl,
      );
      if (!mounted || !result.hasUpdate) return;
      _connectLog.info('update', 'client update check', {
        'urgency': result.urgency.name,
        'current': result.currentVersion,
        'latest': result.latestVersion ?? '',
        'min': result.minSupportedVersion ?? '',
      });
      await _showUpdateDialog(result);
    } catch (e) {
      _connectLog.warn('update', 'config fetch for update check failed', {'error': '$e'});
    }
  }

  Future<void> _showUpdateDialog(UpdateCheckResult result) async {
    final required = result.urgency == UpdateUrgency.required;
    final url = result.downloadUrl;
    await showDialog<void>(
      context: context,
      barrierDismissible: !required,
      builder: (ctx) => AlertDialog(
        title: Text(required ? 'Требуется обновление' : 'Доступно обновление'),
        content: Text(
          required
              ? 'Эта версия клиента (${result.currentVersion}) больше не поддерживается. '
                  'Минимум: ${result.minSupportedVersion ?? "—"}. '
                  'Установите новую версию, чтобы продолжить.'
              : 'Доступна версия ${result.latestVersion}. Сейчас установлена ${result.currentVersion}.',
        ),
        actions: [
          if (!required)
            TextButton(
              onPressed: () => Navigator.of(ctx).pop(),
              child: const Text('Позже'),
            ),
          if (url != null && url.isNotEmpty)
            FilledButton(
              onPressed: () async {
                final uri = Uri.tryParse(url);
                if (uri != null) {
                  await launchUrl(uri, mode: LaunchMode.externalApplication);
                }
                if (!required && ctx.mounted) Navigator.of(ctx).pop();
              },
              child: const Text('Скачать'),
            )
          else if (required)
            TextButton(
              onPressed: () => Navigator.of(ctx).pop(),
              child: const Text('Понятно'),
            ),
        ],
      ),
    );
  }

  Future<void> _loadSubscription() async {
    try {
      final info = await widget.api.fetchSubscription();
      if (!mounted) return;
      setState(() {
        _subscription = info;
        _loadingSubscription = false;
      });
      _connectLog.info('api', 'subscription loaded', {'active': '${info.isActive}'});
    } on SessionExpiredException {
      _connectLog.error('auth', 'subscription fetch: session expired');
      if (!mounted) return;
      navigateToLogin(context, widget.authService, widget.api);
    } catch (e) {
      if (!mounted) return;
      _connectLog.error('api', 'subscription fetch failed', {'error': '$e'});
      setState(() {
        _subscriptionCheckFailed = _isNetworkError(e);
        _subscription = _subscriptionCheckFailed
            ? null
            : const SubscriptionInfo(isActive: false);
        _loadingSubscription = false;
      });
    }
  }

  bool _isNetworkError(Object e) {
    if (e is SocketException) return true;
    if (e is http.ClientException) return true;
    if (e is ApiException && e.statusCode >= 500) return true;
    final msg = e.toString().toLowerCase();
    return msg.contains('failed host lookup') ||
        msg.contains('connection timed out') ||
        msg.contains('connection refused') ||
        msg.contains('network is unreachable');
  }

  Future<void> _openSubscriptionScreen() async {
    await Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => SubscriptionScreen(api: widget.api)),
    );
    _loadSubscription();
  }

  Future<void> _loadStartupData({bool allowAutoConnect = true}) async {
    try {
      final settings = await SettingsService().load();
      final servers = await widget.api.fetchServers();
      final selected = pickBestRelay(
        servers,
        preferredRegion: settings.preferredRegion,
        preferredServerId: settings.preferredServerId,
        autoSelect: settings.autoSelectRelay,
      );
      if (!mounted) return;
      setState(() {
        _selectedRelay = selected;
        _pingMs = _selectedRelay?.rttMs;
        _loadingRelay = false;
        _autoMode = settings.autoSelectRelay;
      });
      _connectLog.info('api', 'servers loaded', {
        'count': '${servers.length}',
        'selected': _selectedRelay?.id ?? 'none',
        'region': _selectedRelay?.region ?? '',
      });
      if (allowAutoConnect) {
        await _maybeAutoConnect();
      }
    } on SessionExpiredException {
      _connectLog.error('auth', 'fetchServers: session expired');
      if (!mounted) return;
      navigateToLogin(context, widget.authService, widget.api);
    } catch (e) {
      if (!mounted) return;
      _connectLog.error('api', 'fetchServers failed', {'error': '$e'});
      final network = _isNetworkError(e);
      setState(() {
        _loadingRelay = false;
        if (network) {
          _subscriptionCheckFailed = true;
          _errorMessage = null;
          _state = ConnState.disconnected;
        } else {
          _errorMessage = e is ApiException ? e.message : 'Не удалось загрузить список серверов';
          _state = ConnState.error;
        }
      });
    }
  }

  Future<void> _openServersPicker() async {
    final wasConnected = _state == ConnState.connected;
    final result = await Navigator.of(context).push<RelayPickResult>(
      MaterialPageRoute(
        builder: (_) => ServersScreen(
          api: widget.api,
          authService: widget.authService,
          selectedServerId: _selectedRelay?.id,
        ),
      ),
    );
    if (!mounted) return;
    if (result == null) return;

    if (wasConnected) {
      _connectLog.info('connect', 'reconnect after server change');
      try {
        await VpnChannel.disconnect();
      } catch (_) {}
      if (mounted) {
        _stopDurationTimer();
        setState(() {
          _state = ConnState.disconnected;
        });
      }
      unawaited(_sessionStats.recordReconnect());
    }

    setState(() => _loadingRelay = true);
    await _loadStartupData(allowAutoConnect: !wasConnected);

    if (wasConnected &&
        mounted &&
        _subscription?.isActive == true &&
        _selectedRelay != null &&
        _state == ConnState.disconnected) {
      await _toggleConnection();
    }
  }

  Future<void> _setAutoMode(bool value) async {
    setState(() => _autoMode = value);
    await SettingsService().setAutoSelectRelay(value);
    if (value) {
      await SettingsService().setPreferredServerId('');
    }
    if (!mounted) return;
    setState(() => _loadingRelay = true);
    final wasConnected = _state == ConnState.connected;
    if (wasConnected) {
      try {
        await VpnChannel.disconnect();
      } catch (_) {}
      if (mounted) {
        setState(() {
          _state = ConnState.disconnected;
          _stopDurationTimer();
        });
      }
    }
    await _loadStartupData(allowAutoConnect: !wasConnected);
    if (wasConnected &&
        mounted &&
        _subscription?.isActive == true &&
        _selectedRelay != null &&
        _state == ConnState.disconnected) {
      await _toggleConnection();
    }
  }

  Future<void> _maybeAutoConnect() async {
    if (_subscription?.isActive != true) return;
    final settings = await SettingsService().load();
    if (settings.autoConnect && _state == ConnState.disconnected) {
      await _toggleConnection();
    }
  }

  @override
  void dispose() {
    _diagUploader?.stop();
    _durationTimer?.cancel();
    _healthTimer?.cancel();
    _ruleEngine.stop();
    _sub?.cancel();
    super.dispose();
  }

  void _startDurationTimer() {
    _durationTimer?.cancel();
    _durationTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      final since = _connectedSince;
      if (!mounted || since == null) return;
      setState(() {
        _durationLabel = formatConnectionDuration(DateTime.now().difference(since));
      });
    });
    _startHealthPoll();
  }

  void _stopDurationTimer() {
    final since = _connectedSince;
    if (since != null) {
      final sec = DateTime.now().difference(since).inSeconds;
      if (sec > 0) {
        unawaited(_sessionStats.addOnlineSeconds(sec));
      }
    }
    _durationTimer?.cancel();
    _durationTimer = null;
    _connectedSince = null;
    _durationLabel = 'Smart routing';
    _stopHealthPoll();
  }

  void _startHealthPoll() {
    _healthTimer?.cancel();
    _healthTimer = Timer.periodic(const Duration(seconds: 60), (_) {
      unawaited(_checkRelayHealth());
    });
  }

  void _stopHealthPoll() {
    _healthTimer?.cancel();
    _healthTimer = null;
  }

  Future<void> _checkRelayHealth() async {
    if (!mounted ||
        _state != ConnState.connected ||
        !_autoMode ||
        _failoverInFlight) {
      return;
    }
    try {
      final settings = await SettingsService().load();
      if (!settings.autoSelectRelay) return;
      final servers = await widget.api.fetchServers();
      final best = pickBestRelay(
        servers,
        preferredRegion: settings.preferredRegion,
        autoSelect: true,
      );
      if (!shouldFailoverRelay(
        current: _selectedRelay,
        servers: servers,
        best: best,
        autoSelect: true,
      )) {
        return;
      }
      await _failoverTo(best!);
    } catch (e) {
      _connectLog.warn('connect', 'relay health check failed', {'error': '$e'});
    }
  }

  Future<void> _failoverTo(RelayServer next, {String reason = 'degraded'}) async {
    if (_failoverInFlight) return;
    _failoverInFlight = true;
    try {
      _connectLog.info('connect', 'auto failover', {
        'from': _selectedRelay?.id ?? '',
        'to': next.id,
        'reason': reason,
      });
      try {
        await VpnChannel.disconnect();
      } catch (_) {}
      if (!mounted) return;
      _stopDurationTimer();
      setState(() {
        _selectedRelay = next;
        _pingMs = next.rttMs > 0 ? next.rttMs : _pingMs;
        _state = ConnState.disconnected;
      });
      unawaited(_sessionStats.recordReconnect());
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Переключили сервер для стабильности'),
            duration: Duration(seconds: 3),
          ),
        );
      }
      if (mounted &&
          _subscription?.isActive == true &&
          _state == ConnState.disconnected) {
        await _toggleConnection();
      }
    } finally {
      _failoverInFlight = false;
    }
  }

  int? _effectivePing(int? nativePing) {
    if (nativePing != null && nativePing > 0) return nativePing;
    final relayPing = _selectedRelay?.rttMs;
    if (relayPing != null && relayPing > 0) return relayPing;
    return null;
  }

  void _onStatus(VpnStatusUpdate update) {
    if (!mounted) return;

    try {
      if (update.event == VpnEvent.connected) {
        unawaited(_ruleEngine.start(initialVersion: _pendingRulesVersion));
      } else if (update.event == VpnEvent.disconnected) {
        _ruleEngine.stop();
      }
    } catch (_) {
      // Rule engine must never crash the UI / process.
    }

    if (!mounted) return;
    setState(() {
      switch (update.event) {
        case VpnEvent.connecting:
          _state = ConnState.connecting;
          _stopDurationTimer();
        case VpnEvent.connected:
          _state = ConnState.connected;
          _connectedSince = DateTime.now();
          _durationLabel = formatConnectionDuration(Duration.zero);
          _startDurationTimer();
          _pingMs = _effectivePing(update.pingMs);
          final ping = _pingMs;
          if (ping != null && ping > 0) {
            unawaited(_sessionStats.recordRtt(ping));
          }
        case VpnEvent.disconnected:
          _state = ConnState.disconnected;
          _stopDurationTimer();
          _pingMs = _effectivePing(_selectedRelay?.rttMs);
        case VpnEvent.permissionDenied:
          _state = ConnState.error;
          _stopDurationTimer();
          _errorMessage = 'Нужно разрешение на VPN-соединение';
        case VpnEvent.error:
          _state = ConnState.error;
          _stopDurationTimer();
          final msg = update.errorMessage ?? 'Ошибка подключения';
          if (AuthErrorCodes.isExpiredMessage(msg)) {
            _errorMessage = 'Сессия истекла. Войдите снова';
            navigateToLogin(context, widget.authService, widget.api);
          } else {
            _errorMessage = msg;
            if (_autoMode) {
              unawaited(_tryFailoverAfterError());
            }
          }
      }
    });
  }

  Future<void> _tryFailoverAfterError() async {
    if (_failoverInFlight || !_autoMode) return;
    try {
      final settings = await SettingsService().load();
      if (!settings.autoSelectRelay) return;
      final servers = await widget.api.fetchServers();
      final others = servers
          .where((s) => s.id != _selectedRelay?.id)
          .toList();
      final best = pickBestRelay(
        others,
        preferredRegion: settings.preferredRegion,
        autoSelect: true,
      );
      if (best != null && best.healthy) {
        await _failoverTo(best, reason: 'error');
      }
    } catch (_) {}
  }

  Future<void> _toggleConnection() async {
    if (_state == ConnState.connected || _state == ConnState.connecting) {
      setState(() {
        _state = ConnState.disconnected;
        _stopDurationTimer();
      });
      try {
        await VpnChannel.disconnect();
      } catch (_) {
        if (mounted) {
          setState(() => _state = ConnState.disconnected);
        }
      }
      return;
    }

    if (_subscription?.isActive != true) {
      _connectLog.warn('connect', 'blocked: subscription inactive');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Для подключения нужна активная подписка'),
            duration: Duration(seconds: 3),
          ),
        );
      }
      await _openSubscriptionScreen();
      return;
    }

    final relay = _selectedRelay;
    if (relay == null) {
      _connectLog.error('connect', 'blocked: no relay selected');
      setState(() {
        _state = ConnState.error;
        _errorMessage = 'Нет доступных серверов. Попробуйте позже';
      });
      return;
    }

    _connectLog.beginConnectSession(relayId: relay.id, host: relay.host);
    _connectLog.info('app', 'build ${BuildInfo.label}');
    setState(() {
      _errorMessage = null;
      _state = ConnState.connecting;
    });
    try {
      var rulesJson = '';
      var exclusionsJson = '[]';
      var bypassPackagesJson = '[]';
      try {
        final ruleSet = await widget.api.fetchRules();
        rulesJson = jsonEncode(ruleSet.toJson());
        final settings = await SettingsService().load();
        exclusionsJson = jsonEncode(settings.exclusions);
        bypassPackagesJson = jsonEncode(settings.bypassPackages);
        _pendingRulesVersion = ruleSet.version;
        _connectLog.info('decision', 'rules loaded', {
          'version': '${ruleSet.version}',
          'count': '${ruleSet.rules.length}',
          'exclusions': '${settings.exclusions.length}',
          'bypassApps': '${settings.bypassPackages.length}',
        });
      } catch (e) {
        _connectLog.warn('decision', 'rules load failed, using defaults', {'error': '$e'});
      }
      final accepted = await VpnChannel.connect(
        relay,
        rulesJson: rulesJson,
        exclusionsJson: exclusionsJson,
        bypassPackagesJson: bypassPackagesJson,
      );
      if (!accepted && mounted) {
        setState(() => _state = ConnState.disconnected);
      }
    } on VpnConnectException catch (e) {
      _connectLog.error('connect', 'VpnConnectException', {'message': e.message});
      if (!mounted) return;
      setState(() {
        _state = ConnState.error;
        _errorMessage = e.message;
      });
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
                    MaterialPageRoute(
                      builder: (_) => SettingsScreen(
                        api: widget.api,
                        authService: widget.authService,
                      ),
                    ),
                  ),
                  onSubscription: _openSubscriptionScreen,
                  subscriptionActive: _subscription?.isActive ?? false,
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
                              ? _durationLabel
                              : 'Smart routing',
                          onTap: _toggleConnection,
                        ),
                      ),
                      const SizedBox(height: 24),
                      if (!_loadingSubscription && _subscriptionCheckFailed) ...[
                        _BackendUnreachableBanner(onRetry: () {
                          setState(() {
                            _loadingSubscription = true;
                            _loadingRelay = true;
                            _subscriptionCheckFailed = false;
                            _errorMessage = null;
                            _state = ConnState.disconnected;
                          });
                          _bootstrap();
                        }),
                        const SizedBox(height: 14),
                      ] else if (!_loadingSubscription && _subscription?.isActive != true) ...[
                        _SubscriptionBanner(onTap: _openSubscriptionScreen),
                        const SizedBox(height: 14),
                      ],
                      _RelayCard(
                        relay: _selectedRelay,
                        pingMs: _pingMs,
                        loading: _loadingRelay,
                        onTap: _openServersPicker,
                      ),
                      const SizedBox(height: 14),
                      _RouteCard(
                        state: _state,
                        autoMode: _autoMode,
                        onAutoModeChanged: _setAutoMode,
                      ),
                    ],
                  ),
                ),
                _BottomNav(
                  onTapStatistics: () => Navigator.of(context).push(
                    MaterialPageRoute(builder: (_) => const StatisticsScreen()),
                  ),
                  onTapServers: _openServersPicker,
                  onTapSettings: () => Navigator.of(context).push(
                    MaterialPageRoute(
                      builder: (_) => SettingsScreen(
                        api: widget.api,
                        authService: widget.authService,
                      ),
                    ),
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

class _TopBar extends StatelessWidget {
  final VoidCallback onSettings;
  final VoidCallback onSubscription;
  final bool subscriptionActive;

  const _TopBar({
    required this.onSettings,
    required this.onSubscription,
    required this.subscriptionActive,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(18, 10, 18, 8),
      child: Row(
        children: [
          IconButton(
            onPressed: onSettings,
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
            onPressed: onSubscription,
            icon: Icon(
              Icons.workspace_premium_rounded,
              color: subscriptionActive ? AppColors.amber : AppColors.textSecondary,
            ),
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

class _BackendUnreachableBanner extends StatelessWidget {
  final VoidCallback onRetry;
  const _BackendUnreachableBanner({required this.onRetry});

  @override
  Widget build(BuildContext context) {
    return _GlassCard(
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Icon(Icons.cloud_off_rounded, color: AppColors.danger),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('Нет связи с сервером',
                    style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 4),
                Text(
                  'Список серверов и подписка не загрузились. '
                  'Проверьте интернет или отключите VPN и нажмите «Повторить».',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
                const SizedBox(height: 10),
                TextButton(onPressed: onRetry, child: const Text('Повторить')),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _SubscriptionBanner extends StatelessWidget {
  final VoidCallback onTap;
  const _SubscriptionBanner({required this.onTap});

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: _GlassCard(
        child: Row(
          children: [
            const Icon(Icons.workspace_premium_rounded, color: AppColors.amber),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('Подписка не активна',
                      style: Theme.of(context).textTheme.titleMedium),
                  const SizedBox(height: 4),
                  Text('Оформите подписку, чтобы подключиться к VPN',
                      style: Theme.of(context).textTheme.bodySmall),
                ],
              ),
            ),
            const Icon(Icons.chevron_right_rounded, color: AppColors.textSecondary),
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
  final VoidCallback? onTap;

  const _RelayCard({
    required this.relay,
    required this.pingMs,
    required this.loading,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final title = loading
        ? 'Ищем лучший relay'
        : relay?.displayRegion ?? 'Relay не найден';
    final subtitle = relay == null
        ? 'Проверьте backend и список серверов'
        : relay!.healthy
            ? relay!.id
            : 'Резервный · ${relay!.id}';

    return GestureDetector(
      onTap: onTap,
      child: _GlassCard(
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
            if (pingMs != null && pingMs! > 0)
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
            if (onTap != null) ...[
              const SizedBox(width: 4),
              const Icon(Icons.chevron_right_rounded, color: AppColors.textSecondary),
            ],
          ],
        ),
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
  final VoidCallback onTapStatistics;
  final VoidCallback onTapServers;
  final VoidCallback onTapSettings;

  const _BottomNav({
    required this.onTapStatistics,
    required this.onTapServers,
    required this.onTapSettings,
  });

  @override
  Widget build(BuildContext context) {
    final items = [
      (Icons.home_rounded, 'Главная', true, null),
      (Icons.bar_chart_rounded, 'Статистика', false, onTapStatistics),
      (Icons.public_rounded, 'Серверы', false, onTapServers),
      (Icons.settings_rounded, 'Настройки', false, onTapSettings),
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
              child: InkWell(
                borderRadius: BorderRadius.circular(16),
                onTap: item.$4,
                child: Padding(
                  padding: const EdgeInsets.symmetric(vertical: 4),
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

  String _flagLabel(String? region) => regionFlag(region);
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