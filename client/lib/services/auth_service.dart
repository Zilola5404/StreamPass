import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

import 'auth_errors.dart';
import 'connection_log.dart';
import 'token_storage.dart';

/// Wraps POST /register, POST /login, POST /logout, POST /refresh and local token storage.
/// Endpoints per StreamPass API spec section 13.
class AuthService {
  final String baseUrl; // e.g. https://api.streampass.com/api/v1
  final http.Client _client;
  final TokenStorage _tokens;

  AuthService({
    required this.baseUrl,
    http.Client? client,
    TokenStorage? tokenStorage,
  })  : _client = client ?? http.Client(),
        _tokens = tokenStorage ?? TokenStorage.secure();

  static final _log = ConnectionLog.instance;

  String get apiBaseUrl => baseUrl.endsWith('/api/v1') ? baseUrl : '$baseUrl/api/v1';

  static const _tokenKey = 'sp_access_token';
  static const _legacyTokenKey = 'sp_jwt_token';
  static const _refreshKey = 'sp_refresh_token';
  static const _accessExpiresKey = 'sp_access_expires_at';

  static const _sessionKeys = [_tokenKey, _legacyTokenKey, _refreshKey, _accessExpiresKey];

  Future<String?> get storedToken async =>
      await _tokens.read(_tokenKey) ?? await _tokens.read(_legacyTokenKey);

  Future<String?> get storedRefreshToken async => _tokens.read(_refreshKey);

  Future<DateTime?> get storedAccessExpiresAt async {
    final raw = await _tokens.read(_accessExpiresKey);
    return raw != null ? DateTime.tryParse(raw) : null;
  }

  /// True when a refresh token exists locally (access token may still be expired).
  Future<bool> get hasSession async => (await storedRefreshToken) != null;

  Future<bool> get isLoggedIn async => await ensureValidSession();

  /// Returns a usable access token, refreshing proactively when near expiry.
  Future<String?> getValidAccessToken() async {
    if (!await ensureValidSession()) return null;
    return storedToken;
  }

  Future<bool>? _refreshInFlight;

  /// Ensures the access token is valid; refreshes using the refresh token when needed.
  Future<bool> ensureValidSession() async {
    final refresh = await storedRefreshToken;
    if (refresh == null) return false;

    final access = await storedToken;
    final expiresAt = await storedAccessExpiresAt;
    final stillValid = access != null &&
        expiresAt != null &&
        expiresAt.isAfter(DateTime.now().toUtc().add(const Duration(minutes: 1)));

    if (stillValid) return true;
    _log.info('auth', 'access token stale, calling POST /refresh');
    return _refreshSessionDeduped();
  }

  Future<bool> _refreshSessionDeduped() {
    final inFlight = _refreshInFlight;
    if (inFlight != null) return inFlight;

    final future = _refreshSessionImpl();
    _refreshInFlight = future;
    return future.whenComplete(() {
      if (identical(_refreshInFlight, future)) {
        _refreshInFlight = null;
      }
    });
  }

  Future<bool> refreshSession() => _refreshSessionDeduped();

  Future<bool> _refreshSessionImpl() async {
    final refresh = await storedRefreshToken;
    if (refresh == null) {
      _log.warn('auth', 'refresh skipped: no refresh token');
      await clearSession();
      return false;
    }

    try {
      final res = await _client.post(
        Uri.parse('${apiBaseUrl}/refresh'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'refresh_token': refresh}),
      );

      if (res.statusCode != 200) {
        _log.error('auth', 'POST /refresh failed', {'status': '${res.statusCode}'});
        await clearSession();
        return false;
      }

      final body = _decodeBody(res.body) as Map<String, dynamic>?;
      final token = body?['access_token'] as String?;
      if (token == null) {
        await clearSession();
        return false;
      }

      await _tokens.write(_tokenKey, token);
      await _tokens.remove(_legacyTokenKey);
      final expiresRaw = body?['access_expires_at'] as String?;
      if (expiresRaw != null) {
        await _tokens.write(_accessExpiresKey, expiresRaw);
      }
      _log.info('auth', 'access token refreshed');
      return true;
    } catch (e) {
      _log.warn('auth', 'refresh network error', {'error': '$e'});
      // Network failure — keep session so a retry can succeed later.
      return (await storedToken) != null;
    }
  }

  Future<void> clearSession() async {
    await _tokens.clearKeys(_sessionKeys);
  }

  Future<AuthResult> login(String email, String password) async {
    try {
      final res = await _client.post(
        Uri.parse('${apiBaseUrl}/login'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'email': email, 'password': password}),
      );
      return _handleAuthResponse(res);
    } catch (_) {
      return AuthResult(success: false, error: 'Сервис временно недоступен. Проверьте подключение к серверу.');
    }
  }

  Future<AuthResult> register(String email, String password) async {
    try {
      final res = await _client.post(
        Uri.parse('${apiBaseUrl}/register'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'email': email, 'password': password}),
      );
      return _handleAuthResponse(res);
    } catch (_) {
      return AuthResult(success: false, error: 'Сервис временно недоступен. Проверьте подключение к серверу.');
    }
  }

  Future<void> logout() async {
    final refresh = await storedRefreshToken;
    if (refresh != null) {
      try {
        await _client.post(
          Uri.parse('${apiBaseUrl}/logout'),
          headers: {'Content-Type': 'application/json'},
          body: jsonEncode({'refresh_token': refresh}),
        );
      } catch (_) {
        // ignore logout errors and clear local state anyway
      }
    }
    await clearSession();
  }

  Future<AuthResult> _handleAuthResponse(http.Response res) async {
    if (res.statusCode != 200 && res.statusCode != 201) {
      return AuthResult(success: false, error: _errorMessage(res));
    }

    final body = _decodeBody(res.body) as Map<String, dynamic>?;
    if (body == null) {
      return AuthResult(success: false, error: 'Некорректный ответ сервера');
    }

    final token = body['access_token'] as String? ?? body['token'] as String?;
    final refresh = body['refresh_token'] as String?;

    if (token == null) {
      return AuthResult(success: false, error: 'Некорректный ответ сервера');
    }

    await _tokens.write(_tokenKey, token);
    await _tokens.remove(_legacyTokenKey);
    if (refresh != null) {
      await _tokens.write(_refreshKey, refresh);
    }
    final expiresRaw = body['access_expires_at'] as String?;
    if (expiresRaw != null) {
      await _tokens.write(_accessExpiresKey, expiresRaw);
    }

    return AuthResult(success: true, token: token, refreshToken: refresh);
  }

  String _errorMessage(http.Response res) {
    var message = 'Не удалось выполнить вход';
    try {
      final body = _decodeBody(res.body);
      if (body is Map<String, dynamic>) {
        final error = body['error'];
        if (error is Map<String, dynamic>) {
          message = error['message'] as String? ?? message;
        } else {
          message = body['message'] as String? ?? message;
        }
      }
    } catch (_) {}
    return message;
  }

  dynamic _decodeBody(String body) {
    if (body.trim().isEmpty) return null;
    return jsonDecode(body);
  }
}

class AuthResult {
  final bool success;
  final String? token;
  final String? refreshToken;
  final String? error;
  AuthResult({
    required this.success,
    this.token,
    this.refreshToken,
    this.error,
  });
}

class SessionExpiredException implements Exception {
  final String message;
  SessionExpiredException([this.message = 'Сессия истекла. Войдите снова.']);

  @override
  String toString() => message;
}
