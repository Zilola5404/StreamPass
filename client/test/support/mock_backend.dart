import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

/// In-memory mock of StreamPass `/api/v1` for Flutter E2E widget tests (BL-031).
class MockStreamPassBackend {
  MockStreamPassBackend({
    this.email = 'user@example.com',
    this.password = 'secret123',
    List<Map<String, dynamic>>? servers,
  }) : servers = servers ??
            [
              {
                'id': 'nl-native-1',
                'region': 'nl',
                'region_name': 'Amsterdam (NL)',
                'host': '212.43.156.33',
                'port': 443,
                'healthy': true,
                'load_ratio': 0.1,
                'rtt_ms': 18,
                'connection_config': 'hysteria2://test@host:443/',
              },
              {
                'id': 'pl-warsaw-1',
                'region': 'pl',
                'region_name': 'Warsaw (PL)',
                'host': '10.0.0.2',
                'port': 443,
                'healthy': true,
                'load_ratio': 0.4,
                'rtt_ms': 40,
                'connection_config': 'hysteria2://test@host:443/',
              },
            ];

  final String email;
  final String password;
  final List<Map<String, dynamic>> servers;

  int loginCalls = 0;
  int registerCalls = 0;
  int serversCalls = 0;
  int refreshCalls = 0;

  MockClient get client => MockClient(_handle);

  Future<http.Response> _handle(http.Request request) async {
    final path = request.url.path;
    final method = request.method.toUpperCase();

    if (method == 'POST' && path.endsWith('/login')) {
      loginCalls++;
      final body = jsonDecode(request.body) as Map<String, dynamic>;
      if (body['email'] == email && body['password'] == password) {
        return _json(200, _tokens());
      }
      return _json(401, {
        'error': {'code': 'invalid_credentials', 'message': 'invalid email or password'},
      });
    }

    if (method == 'POST' && path.endsWith('/register')) {
      registerCalls++;
      return _json(201, _tokens());
    }

    if (method == 'POST' && path.endsWith('/refresh')) {
      refreshCalls++;
      return _json(200, {
        'access_token': 'access-refreshed',
        'access_expires_at': '2099-01-01T00:00:00Z',
      });
    }

    if (method == 'GET' && path.endsWith('/config')) {
      return _json(200, {
        'version': 1,
        'min_supported_client_version': '0.1.0',
        'latest_client_version': '0.1.1',
        'client_download_url': '',
        'telemetry_enabled': true,
        'rule_poll_interval_sec': 300,
        'relay_poll_interval_sec': 60,
      });
    }

    if (method == 'GET' && path.endsWith('/rules')) {
      return _json(200, {
        'version': 1,
        'updated_at': '2026-08-04T00:00:00Z',
        'rules': [
          {'kind': 'domain', 'pattern': '*.ru', 'mode': 'DIRECT'},
        ],
      });
    }

    if (method == 'GET' && path.endsWith('/regions')) {
      return _json(200, [
        {'code': 'de', 'city': 'Frankfurt', 'country': 'Germany', 'label': 'Frankfurt (DE)'},
        {'code': 'nl', 'city': 'Amsterdam', 'country': 'Netherlands', 'label': 'Amsterdam (NL)'},
        {'code': 'pl', 'city': 'Warsaw', 'country': 'Poland', 'label': 'Warsaw (PL)'},
        {'code': 'fi', 'city': 'Helsinki', 'country': 'Finland', 'label': 'Helsinki (FI)'},
      ]);
    }

    if (method == 'GET' && path.contains('/servers')) {
      serversCalls++;
      final region = request.url.queryParameters['region'];
      var list = servers;
      if (region != null && region.isNotEmpty) {
        list = servers.where((s) => s['region'] == region).toList();
      }
      return _json(200, list);
    }

    if (method == 'GET' && path.endsWith('/subscription')) {
      return _json(200, {
        'status': 'ACTIVE',
        'active_until': '2099-01-01T00:00:00Z',
      });
    }

    if (method == 'GET' && path.endsWith('/exclusions')) {
      return _json(200, {'domains': <String>[]});
    }

    if (method == 'PUT' && path.endsWith('/exclusions')) {
      final body = jsonDecode(request.body) as Map<String, dynamic>;
      return _json(200, body);
    }

    if (method == 'POST' && path.endsWith('/telemetry')) {
      return http.Response('', 204);
    }

    return _json(404, {
      'error': {'code': 'not_found', 'message': 'no mock for $method $path'},
    });
  }

  Map<String, dynamic> _tokens() => {
        'access_token': 'access-mock',
        'refresh_token': 'refresh-mock',
        'access_expires_at': '2099-01-01T00:00:00Z',
        'token_type': 'Bearer',
      };

  http.Response _json(int status, Object body) => http.Response(
        jsonEncode(body),
        status,
        headers: {'content-type': 'application/json'},
      );
}
