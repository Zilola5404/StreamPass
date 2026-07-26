import 'dart:convert';

import 'package:http/http.dart' as http;

import 'auth_service.dart';

class StreamPassApi {
  final String baseUrl;
  final AuthService authService;
  final http.Client _client;

  StreamPassApi({
    required this.baseUrl,
    required this.authService,
    http.Client? client,
  }) : _client = client ?? http.Client();

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

  Future<dynamic> _get(String path, {bool authenticated = true}) async {
    final res = await _client.get(
      Uri.parse('$baseUrl$path'),
      headers: await _headers(authenticated: authenticated),
    );
    return _decode(res);
  }

  Future<dynamic> _post(String path, Map<String, dynamic> body) async {
    final res = await _client.post(
      Uri.parse('$baseUrl$path'),
      headers: await _headers(),
      body: jsonEncode(body),
    );
    return _decode(res);
  }

  Future<Map<String, String>> _headers({bool authenticated = true}) async {
    final headers = <String, String>{'Content-Type': 'application/json'};
    if (authenticated) {
      final token = await authService.storedToken;
      if (token != null) headers['Authorization'] = 'Bearer $token';
    }
    return headers;
  }

  dynamic _decode(http.Response res) {
    if (res.statusCode >= 200 && res.statusCode < 300) {
      if (res.body.isEmpty) return null;
      return jsonDecode(res.body);
    }

    var message = 'Сервер временно недоступен';
    try {
      final body = jsonDecode(res.body) as Map<String, dynamic>;
      final error = body['error'];
      if (error is Map<String, dynamic>) {
        message = error['message'] as String? ?? message;
      }
    } catch (_) {}
    throw ApiException(res.statusCode, message);
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
