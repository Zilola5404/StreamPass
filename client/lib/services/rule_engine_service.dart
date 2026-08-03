import 'dart:async';
import 'dart:convert';

import 'connection_log.dart';
import 'settings_service.dart';
import 'streampass_api.dart';
import 'vpn_channel.dart';

/// Loads routing rules from the backend, polls for updates, and pushes
/// hot-reloads to the native Decision Engine (BL-006).
class RuleEngineService {
  RuleEngineService({
    required this.api,
    SettingsService? settings,
  }) : _settings = settings ?? SettingsService();

  final StreamPassApi api;
  final SettingsService _settings;
  static final _log = ConnectionLog.instance;

  Timer? _timer;
  int _knownVersion = 0;
  int _pollIntervalSec = 3600;

  int get knownVersion => _knownVersion;
  int get pollIntervalSec => _pollIntervalSec;

  /// Reads client config and starts periodic rule polling while VPN is up.
  Future<void> start({int initialVersion = 0}) async {
    _knownVersion = initialVersion;
    await _loadPollInterval();
    await syncOnce(force: true);
    _timer?.cancel();
    _timer = Timer.periodic(Duration(seconds: _pollIntervalSec), (_) {
      syncOnce();
    });
    _log.info('rules', 'rule engine polling started', {
      'intervalSec': '$_pollIntervalSec',
      'version': '$_knownVersion',
    });
  }

  void stop() {
    _timer?.cancel();
    _timer = null;
    _log.info('rules', 'rule engine polling stopped');
  }

  Future<void> _loadPollInterval() async {
    try {
      final config = await api.fetchConfig();
      if (config.rulePollIntervalSec > 0) {
        _pollIntervalSec = config.rulePollIntervalSec;
      }
    } catch (e) {
      _log.warn('rules', 'config load failed, using default poll interval', {
        'intervalSec': '$_pollIntervalSec',
        'error': '$e',
      });
    }
  }

  /// Fetches rules + exclusions and pushes to native when version changes.
  Future<bool> syncOnce({bool force = false}) async {
    try {
      final ruleSet = await api.fetchRules();
      if (!force && ruleSet.version == _knownVersion) {
        return false;
      }

      final settings = await _settings.load();
      final rulesJson = jsonEncode(ruleSet.toJson());
      final exclusionsJson = jsonEncode(settings.exclusions);

      final err = await VpnChannel.updateRules(
        rulesJson: rulesJson,
        exclusionsJson: exclusionsJson,
      );
      if (err != null && err.isNotEmpty) {
        _log.warn('rules', 'native updateRules skipped', {'error': err});
        return false;
      }

      _knownVersion = ruleSet.version;
      _log.info('rules', 'rule set applied', {
        'version': '${ruleSet.version}',
        'count': '${ruleSet.rules.length}',
        'exclusions': '${settings.exclusions.length}',
      });
      return true;
    } catch (e) {
      _log.warn('rules', 'sync failed', {'error': '$e'});
      return false;
    }
  }
}
