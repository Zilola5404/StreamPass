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
}
