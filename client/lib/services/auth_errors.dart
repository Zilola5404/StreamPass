import 'dart:convert';

import 'package:http/http.dart' as http;

/// Stable auth error codes aligned with `shared/errors` on the backend.
class AuthErrorCodes {
  AuthErrorCodes._();

  static const tokenExpired = 'AUTH_TOKEN_EXPIRED';
  static const tokenInvalid = 'AUTH_TOKEN_INVALID';
  static const tokenRevoked = 'AUTH_TOKEN_REVOKED';
  static const unauthorized = 'UNAUTHORIZED';
  static const invalidCredentials = 'AUTH_INVALID_CREDENTIALS';

  static const expiredCodes = {
    tokenExpired,
    tokenInvalid,
    tokenRevoked,
    unauthorized,
  };

  /// True when an HTTP response indicates the access token is no longer valid.
  static bool isExpiredResponse(http.Response res) {
    if (res.statusCode == 401) {
      final code = _errorCode(res);
      if (code != null && expiredCodes.contains(code)) {
        return true;
      }
      if (code == null) {
        return true;
      }
    }
    final message = _errorMessage(res).toLowerCase();
    return message.contains('token expired') || message.contains('missing bearer');
  }

  /// True for native/API error strings shown in the VPN UI layer.
  static bool isExpiredMessage(String message) {
    final lower = message.toLowerCase();
    if (lower.contains('token expired')) {
      return true;
    }
    return expiredCodes.any(message.contains);
  }

  static String? _errorCode(http.Response res) {
    try {
      final body = jsonDecode(res.body) as Map<String, dynamic>;
      final error = body['error'];
      if (error is Map<String, dynamic>) {
        return (error['code'] as String?)?.toUpperCase();
      }
    } catch (_) {}
    return null;
  }

  static String _errorMessage(http.Response res) {
    try {
      final body = jsonDecode(res.body) as Map<String, dynamic>;
      final error = body['error'];
      if (error is Map<String, dynamic>) {
        return error['message'] as String? ?? '';
      }
    } catch (_) {}
    return '';
  }
}
