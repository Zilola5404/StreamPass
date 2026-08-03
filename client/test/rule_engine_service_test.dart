import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import 'package:streampass/services/auth_service.dart';
import 'package:streampass/services/connection_log.dart';
import 'package:streampass/services/rule_engine_service.dart';
import 'package:streampass/services/streampass_api.dart';

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

StreamPassApi _apiWithRulesVersion(int version) {
  final client = _MockClient((request) async {
    final path = request.url.path;
    if (path.endsWith('/config')) {
      return http.Response(
        jsonEncode({
          'version': 1,
          'min_supported_client_version': '0.1.0',
          'telemetry_enabled': false,
          'rule_poll_interval_sec': 3600,
          'relay_poll_interval_sec': 60,
        }),
        200,
      );
    }
    if (path.endsWith('/rules')) {
      return http.Response(
        jsonEncode({
          'version': version,
          'rules': [
            {'kind': 'DOMAIN', 'pattern': '*.ru', 'mode': 'DIRECT'},
          ],
        }),
        200,
      );
    }
    return http.Response('{}', 404);
  });
  final auth = AuthService(baseUrl: 'https://example.com', client: client);
  return StreamPassApi(baseUrl: 'https://example.com', authService: auth, client: client);
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await SharedPreferences.getInstance();
    ConnectionLog.instance.clear();
  });

  test('stop is safe when polling was never started', () {
    final engine = RuleEngineService(api: _apiWithRulesVersion(1));
    engine.stop();
    expect(engine.knownVersion, 0);
  });

  test('syncOnce skips push when version unchanged', () async {
    final engine = RuleEngineService(api: _apiWithRulesVersion(42));
    await engine.start(initialVersion: 42);
    engine.stop();

    final applied = await engine.syncOnce(force: false);
    expect(applied, isFalse);
    expect(engine.knownVersion, 42);
  });
}
