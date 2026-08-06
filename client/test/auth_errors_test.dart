import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:streampass/services/auth_errors.dart';

void main() {
  test('isExpiredResponse matches stable error codes', () {
    final res = http.Response(
      '{"error":{"code":"AUTH_TOKEN_EXPIRED","message":"token expired"}}',
      401,
    );
    expect(AuthErrorCodes.isExpiredResponse(res), isTrue);
  });

  test('isExpiredMessage matches native VPN errors', () {
    expect(AuthErrorCodes.isExpiredMessage('AUTH_TOKEN_EXPIRED'), isTrue);
    expect(AuthErrorCodes.isExpiredMessage('network ok'), isFalse);
  });
}
