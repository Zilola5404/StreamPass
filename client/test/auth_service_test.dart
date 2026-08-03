import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import 'package:streampass/services/auth_service.dart';

class _ThrowingClient extends http.BaseClient {
  @override
  Future<http.StreamedResponse> send(http.BaseRequest request) {
    throw http.ClientException('offline');
  }
}

class _MockClient extends http.BaseClient {
  _MockClient(this.handler);
  final Future<http.Response> Function(http.BaseRequest request) handler;

  @override
  Future<http.StreamedResponse> send(http.BaseRequest request) async {
    final res = await handler(request);
    return http.StreamedResponse(
      Stream.value(res.bodyBytes),
      res.statusCode,
      headers: res.headers,
      request: request,
    );
  }
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await SharedPreferences.getInstance();
  });

  test('login returns a friendly error when the backend is unreachable', () async {
    final service = AuthService(baseUrl: 'https://example.com', client: _ThrowingClient());

    final result = await service.login('test@example.com', 'secret');

    expect(result.success, isFalse);
    expect(result.error, contains('Сервис временно недоступен'));
  });

  test('refreshSession stores a new access token', () async {
    SharedPreferences.setMockInitialValues({
      'sp_refresh_token': 'refresh-abc',
      'sp_access_token': 'expired-access',
    });

    final client = _MockClient((request) async {
      expect(request.url.path, endsWith('/refresh'));
      return http.Response(
        jsonEncode({
          'access_token': 'new-access',
          'access_expires_at': '2099-01-01T00:00:00Z',
        }),
        200,
        headers: {'content-type': 'application/json'},
      );
    });

    final service = AuthService(baseUrl: 'https://example.com/api/v1', client: client);
    final ok = await service.refreshSession();

    expect(ok, isTrue);
    expect(await service.storedToken, 'new-access');
  });

  test('ensureValidSession refreshes when access token expiry passed', () async {
    SharedPreferences.setMockInitialValues({
      'sp_refresh_token': 'refresh-abc',
      'sp_access_token': 'old-access',
      'sp_access_expires_at': '2020-01-01T00:00:00Z',
    });

    final client = _MockClient((request) async {
      return http.Response(
        jsonEncode({
          'access_token': 'fresh-access',
          'access_expires_at': '2099-01-01T00:00:00Z',
        }),
        200,
        headers: {'content-type': 'application/json'},
      );
    });

    final service = AuthService(baseUrl: 'https://example.com/api/v1', client: client);
    final ok = await service.ensureValidSession();

    expect(ok, isTrue);
    expect(await service.storedToken, 'fresh-access');
  });
}
