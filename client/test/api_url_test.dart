import 'package:flutter_test/flutter_test.dart';
import 'package:streampass/services/auth_service.dart';
import 'package:streampass/services/streampass_api.dart';

void main() {
  test('AuthService uses /api/v1 prefix when base URL is plain host', () {
    final service = AuthService(baseUrl: 'https://example.com');
    expect(service.apiBaseUrl, 'https://example.com/api/v1');
  });

  test('AuthService keeps /api/v1 prefix when already present', () {
    final service = AuthService(baseUrl: 'https://example.com/api/v1');
    expect(service.apiBaseUrl, 'https://example.com/api/v1');
  });

  test('StreamPassApi uses /api/v1 prefix for requests', () {
    final service = AuthService(baseUrl: 'https://example.com');
    final api = StreamPassApi(baseUrl: 'https://example.com', authService: service);
    expect(api.apiBaseUrl, 'https://example.com/api/v1');
  });
}
