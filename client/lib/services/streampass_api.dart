import 'dart:convert';

import 'package:http/http.dart' as http;

import 'connection_log.dart';
import 'auth_errors.dart';
import 'auth_service.dart';
import 'region_catalog.dart';

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

  Future<List<RelayServer>> fetchServers({String? region}) async {
    final q = (region != null && region.isNotEmpty)
        ? '?region=${Uri.encodeQueryComponent(region)}'
        : '';
    final body = await _get('/servers$q');
    return (body as List<dynamic>)
        .map((item) => RelayServer.fromJson(item as Map<String, dynamic>))
        .toList();
  }

  Future<List<RegionInfo>> fetchRegions() async {
    try {
      final body = await _get('/regions', authenticated: false);
      return (body as List<dynamic>)
          .map((item) => RegionInfo.fromJson(item as Map<String, dynamic>))
          .toList();
    } catch (_) {
      return List<RegionInfo>.from(fallbackRegions);
    }
  }

  Future<void> sendTelemetry(TelemetryPayload payload) async {
    await _post('/telemetry', payload.toJson());
  }

  /// Operator routing diagnostics (hostname/IP/mode/latency — no full URLs).
  Future<void> uploadDiag(List<DiagEvent> events) async {
    if (events.isEmpty) return;
    await _post('/diag', {
      'events': events.map((e) => e.toJson()).toList(),
    });
  }

  Future<SubscriptionInfo> fetchSubscription() async {
    final body = await _get('/subscription');
    return SubscriptionInfo.fromJson(body as Map<String, dynamic>);
  }

  /// Returns the ЮKassa confirmation URL to open in a browser to complete
  /// payment. Note: backend billing has not been tested against live
  /// ЮKassa credentials yet — this call is only as reliable as that
  /// backend path (see project notes).
  Future<String> createPayment({String planCode = 'month'}) async {
    final body = await _post('/payments', {'plan_code': planCode});
    return (body as Map<String, dynamic>)['confirmation_url'] as String;
  }

  Future<void> cancelSubscription() async {
    await _post('/subscription/cancel', const {});
  }

  Future<List<PlanInfo>> fetchPlans() async {
    final body = await _get('/plans');
    return (body as List<dynamic>)
        .map((e) => PlanInfo.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<PaymentRecord>> fetchPayments() async {
    final body = await _get('/payments');
    return (body as List<dynamic>)
        .map((e) => PaymentRecord.fromJson(e as Map<String, dynamic>))
        .toList();
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
    if (res.statusCode == 401 && !retried && AuthErrorCodes.isExpiredResponse(res)) {
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
  final String latestClientVersion;
  final String clientDownloadUrl;
  final bool telemetryEnabled;
  final int rulePollIntervalSec;
  final int relayPollIntervalSec;

  const ClientConfig({
    required this.version,
    required this.minSupportedClientVersion,
    this.latestClientVersion = '',
    this.clientDownloadUrl = '',
    required this.telemetryEnabled,
    required this.rulePollIntervalSec,
    required this.relayPollIntervalSec,
  });

  factory ClientConfig.fromJson(Map<String, dynamic> json) => ClientConfig(
        version: json['version'] as int,
        minSupportedClientVersion:
            json['min_supported_client_version'] as String,
        latestClientVersion: (json['latest_client_version'] as String?) ?? '',
        clientDownloadUrl: (json['client_download_url'] as String?) ?? '',
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
  final String regionName;
  final String host;
  final int port;
  final bool healthy;
  final double loadRatio;
  final int rttMs;
  final String connectionConfig;

  const RelayServer({
    required this.id,
    required this.region,
    this.regionName = '',
    required this.host,
    required this.port,
    required this.healthy,
    required this.loadRatio,
    required this.rttMs,
    required this.connectionConfig,
  });

  String get displayRegion =>
      regionName.isNotEmpty ? regionName : regionLabel(region);

  factory RelayServer.fromJson(Map<String, dynamic> json) => RelayServer(
        id: json['id'] as String,
        region: json['region'] as String,
        regionName: json['region_name'] as String? ?? '',
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

class PlanInfo {
  final String code;
  final String title;
  final int amountRub;
  final int periodDays;

  const PlanInfo({
    required this.code,
    required this.title,
    required this.amountRub,
    required this.periodDays,
  });

  factory PlanInfo.fromJson(Map<String, dynamic> json) => PlanInfo(
        code: json['code'] as String? ?? '',
        title: json['title'] as String? ?? '',
        amountRub: (json['amount_rub'] as num?)?.toInt() ?? 0,
        periodDays: (json['period_days'] as num?)?.toInt() ?? 0,
      );
}

class PaymentRecord {
  final String id;
  final int amountRub;
  final int periodDays;
  final String status;
  final DateTime? createdAt;

  const PaymentRecord({
    required this.id,
    required this.amountRub,
    required this.periodDays,
    required this.status,
    this.createdAt,
  });

  factory PaymentRecord.fromJson(Map<String, dynamic> json) => PaymentRecord(
        id: json['id'] as String? ?? '',
        amountRub: (json['amount_rub'] as num?)?.toInt() ?? 0,
        periodDays: (json['period_days'] as num?)?.toInt() ?? 0,
        status: json['status'] as String? ?? '',
        createdAt: DateTime.tryParse(json['created_at'] as String? ?? ''),
      );
}

/// One routing diagnostic sample uploaded to POST /diag.
class DiagEvent {
  final String proto;
  final String host;
  final String destIp;
  final int destPort;
  final String mode;
  final String result;
  final int latencyMs;
  final String errorCode;
  final String relayId;
  final String clientVersion;
  final DateTime? recordedAt;

  const DiagEvent({
    required this.proto,
    required this.host,
    required this.destIp,
    required this.destPort,
    required this.mode,
    required this.result,
    required this.latencyMs,
    this.errorCode = '',
    this.relayId = '',
    this.clientVersion = '',
    this.recordedAt,
  });

  DiagEvent copyWith({
    String? clientVersion,
    DateTime? recordedAt,
    String? relayId,
  }) =>
      DiagEvent(
        proto: proto,
        host: host,
        destIp: destIp,
        destPort: destPort,
        mode: mode,
        result: result,
        latencyMs: latencyMs,
        errorCode: errorCode,
        relayId: relayId ?? this.relayId,
        clientVersion: clientVersion ?? this.clientVersion,
        recordedAt: recordedAt ?? this.recordedAt,
      );

  Map<String, dynamic> toJson() => {
        'proto': proto,
        'host': host,
        'dest_ip': destIp,
        'dest_port': destPort,
        'mode': mode,
        'result': result,
        'latency_ms': latencyMs,
        'error_code': errorCode,
        'relay_id': relayId,
        'client_version': clientVersion,
        if (recordedAt != null) 'recorded_at': recordedAt!.toUtc().toIso8601String(),
      };
}