import 'dart:convert';

import 'package:http/http.dart' as http;

import 'connection_log.dart';
import 'auth_service.dart';

class StreamPassApi {
  final String baseUrl;
  final AuthService authService;
  final http.Client _client;

  String get apiBaseUrl => baseUrl.endsWith('/api/v1') ? baseUrl : '$baseUrl/api/v1';

  StreamPassApi({
    required this.baseUrl,
    required this.authService,
    http.Client? client,
  }) : _client = client ?? http.Client();

  static final _log = ConnectionLog.instance;

  Future<ClientConfig> fetchConfig() async {
    final body = await _get('/config', authenticated: false);
    return ClientConfig.fromJson(body);
  }

  Future<RuleSet> fetchRules() async {
    final body = await _get('/rules', authenticated: false);
    return RuleSet.fromJson(body);
  }

  Future<List<RelayServer>> fetchServers() async {
    final body = await _get('/servers');
    return (body as List<dynamic>)
        .map((item) => RelayServer.fromJson(item as Map<String, dynamic>))
        .toList();
  }

  Future<void> sendTelemetry(TelemetryPayload payload) async {
    await _post('/telemetry', payload.toJson());
  }

  Future<SubscriptionInfo> fetchSubscription() async {
    final body = await _get('/subscription');
    return SubscriptionInfo.fromJson(body as Map<String, dynamic>);
  }

  /// Returns the ЮKassa confirmation URL to open in a browser to complete
  /// payment. Note: backend billing has not been tested against live
  /// ЮKassa credentials yet — this call is only as reliable as that
  /// backend path (see project notes).
  Future<String> createPayment() async {
    final body = await _post('/payments', const {});
    return (body as Map<String, dynamic>)['confirmation_url'] as String;
  }

  Future<void> cancelSubscription() async {
    await _post('/subscription/cancel', const {});
  }

  Future<List<String>> fetchExclusions() async {
    final body = await _get('/exclusions');
    final map = body as Map<String, dynamic>;
    final list = map['domains'] as List<dynamic>? ?? const [];
    return list.map((e) => e.toString()).toList();
  }

  Future<List<String>> putExclusions(List<String> domains) async {
    final body = await _put('/exclusions', {'domains': domains});
    final map = body as Map<String, dynamic>;
    final list = map['domains'] as List<dynamic>? ?? const [];
    return list.map((e) => e.toString()).toList();
  }

  Future<dynamic> _get(String path, {bool authenticated = true, bool retried = false}) async {
    final res = await _client.get(
      Uri.parse('$apiBaseUrl$path'),
      headers: await _headers(authenticated: authenticated),
    );
    return _decode(res, () => _get(path, authenticated: authenticated, retried: true), retried: retried);
  }

  Future<dynamic> _post(String path, Map<String, dynamic> body, {bool retried = false}) async {
    final res = await _client.post(
      Uri.parse('$apiBaseUrl$path'),
      headers: await _headers(),
      body: jsonEncode(body),
    );
    return _decode(res, () => _post(path, body, retried: true), retried: retried);
  }

  Future<dynamic> _put(String path, Map<String, dynamic> body, {bool retried = false}) async {
    final res = await _client.put(
      Uri.parse('$apiBaseUrl$path'),
      headers: await _headers(),
      body: jsonEncode(body),
    );
    return _decode(res, () => _put(path, body, retried: true), retried: retried);
  }

  Future<Map<String, String>> _headers({bool authenticated = true}) async {
    final headers = <String, String>{'Content-Type': 'application/json'};
    if (authenticated) {
      final token = await authService.getValidAccessToken();
      if (token == null) {
        throw SessionExpiredException();
      }
      headers['Authorization'] = 'Bearer $token';
    }
    return headers;
  }

  Future<dynamic> _decode(
    http.Response res,
    Future<dynamic> Function() retry, {
    required bool retried,
  }) async {
    if (res.statusCode == 401 && !retried && _isAuthExpired(res)) {
      _log.warn('auth', 'access token expired, refreshing');
      final refreshed = await authService.refreshSession();
      if (refreshed) {
        _log.info('auth', 'token refreshed, retrying request');
        return retry();
      }
      _log.error('auth', 'refresh failed, session expired');
      throw SessionExpiredException(_errorMessage(res));
    }

    if (res.statusCode >= 200 && res.statusCode < 300) {
      if (res.body.isEmpty) return null;
      return jsonDecode(res.body);
    }

    throw ApiException(res.statusCode, _errorMessage(res));
  }

  bool _isAuthExpired(http.Response res) {
    try {
      final body = jsonDecode(res.body) as Map<String, dynamic>;
      final error = body['error'];
      if (error is Map<String, dynamic>) {
        final code = (error['code'] as String? ?? '').toUpperCase();
        if (code.contains('TOKEN_EXPIRED') || code.contains('UNAUTHORIZED')) {
          return true;
        }
        final message = (error['message'] as String? ?? '').toLowerCase();
        if (message.contains('token expired') || message.contains('missing bearer')) {
          return true;
        }
      }
    } catch (_) {}
    return res.statusCode == 401;
  }

  String _errorMessage(http.Response res) {
    var message = 'Сервис временно недоступен';
    try {
      final body = jsonDecode(res.body) as Map<String, dynamic>;
      final error = body['error'];
      if (error is Map<String, dynamic>) {
        message = error['message'] as String? ?? message;
      }
    } catch (_) {}
    return message;
  }
}

class ApiException implements Exception {
  final int statusCode;
  final String message;

  ApiException(this.statusCode, this.message);

  @override
  String toString() => message;
}

class ClientConfig {
  final int version;
  final String minSupportedClientVersion;
  final bool telemetryEnabled;
  final int rulePollIntervalSec;
  final int relayPollIntervalSec;

  const ClientConfig({
    required this.version,
    required this.minSupportedClientVersion,
    required this.telemetryEnabled,
    required this.rulePollIntervalSec,
    required this.relayPollIntervalSec,
  });

  factory ClientConfig.fromJson(Map<String, dynamic> json) => ClientConfig(
        version: json['version'] as int,
        minSupportedClientVersion:
            json['min_supported_client_version'] as String,
        telemetryEnabled: json['telemetry_enabled'] as bool,
        rulePollIntervalSec: json['rule_poll_interval_sec'] as int,
        relayPollIntervalSec: json['relay_poll_interval_sec'] as int,
      );
}

class RuleSet {
  final int version;
  final List<RouteRule> rules;

  const RuleSet({required this.version, required this.rules});

  factory RuleSet.fromJson(Map<String, dynamic> json) => RuleSet(
        version: json['version'] as int,
        rules: (json['rules'] as List<dynamic>)
            .map((item) => RouteRule.fromJson(item as Map<String, dynamic>))
            .toList(),
      );

  Map<String, dynamic> toJson() => {
        'version': version,
        'rules': rules.map((r) => r.toJson()).toList(),
      };
}

class RouteRule {
  final String kind;
  final String pattern;
  final String mode;

  const RouteRule({
    required this.kind,
    required this.pattern,
    required this.mode,
  });

  factory RouteRule.fromJson(Map<String, dynamic> json) => RouteRule(
        kind: json['kind'] as String,
        pattern: json['pattern'] as String,
        mode: json['mode'] as String,
      );

  Map<String, dynamic> toJson() => {
        'kind': kind,
        'pattern': pattern,
        'mode': mode,
      };
}

class RelayServer {
  final String id;
  final String region;
  final String host;
  final int port;
  final bool healthy;
  final double loadRatio;
  final int rttMs;
  final String connectionConfig;

  const RelayServer({
    required this.id,
    required this.region,
    required this.host,
    required this.port,
    required this.healthy,
    required this.loadRatio,
    required this.rttMs,
    required this.connectionConfig,
  });

  factory RelayServer.fromJson(Map<String, dynamic> json) => RelayServer(
        id: json['id'] as String,
        region: json['region'] as String,
        host: json['host'] as String,
        port: json['port'] as int,
        healthy: json['healthy'] as bool,
        loadRatio: (json['load_ratio'] as num).toDouble(),
        rttMs: json['rtt_ms'] as int,
        connectionConfig: json['connection_config'] as String? ?? '',
      );
}

class TelemetryPayload {
  final int rttMs;
  final double packetLossPct;
  final String relayId;
  final String clientVersion;
  final String os;
  final int connectMillis;
  final String errorCode;

  const TelemetryPayload({
    required this.rttMs,
    required this.packetLossPct,
    required this.relayId,
    required this.clientVersion,
    required this.os,
    required this.connectMillis,
    this.errorCode = '',
  });

  Map<String, dynamic> toJson() => {
        'rtt_ms': rttMs,
        'packet_loss_pct': packetLossPct,
        'relay_id': relayId,
        'client_version': clientVersion,
        'os': os,
        'connect_ms': connectMillis,
        'error_code': errorCode,
      };
}

class SubscriptionInfo {
  final bool isActive;
  final DateTime? activeUntil;

  const SubscriptionInfo({required this.isActive, this.activeUntil});

  /// The exact status string the backend uses is matched tolerantly
  /// (case-insensitive "ACTIVE" substring, or an active_until timestamp
  /// still in the future) rather than a single hardcoded literal — this
  /// degrades safely (reports "inactive") instead of silently misreading
  /// a paid user as unpaid if the exact wording ever changes.
  factory SubscriptionInfo.fromJson(Map<String, dynamic> json) {
    final statusStr = (json['status'] as String? ?? '').toUpperCase();
    final untilRaw = json['active_until'] as String?;
    final until = untilRaw != null ? DateTime.tryParse(untilRaw) : null;
    final active = statusStr == 'ACTIVE' ||
        (until != null && until.isAfter(DateTime.now()));
    return SubscriptionInfo(isActive: active, activeUntil: until);
  }
}