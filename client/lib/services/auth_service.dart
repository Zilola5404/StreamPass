import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

/// Wraps POST /register, POST /login, POST /logout and local token storage.
/// Endpoints per StreamPass API spec section 13.
class AuthService {
  final String baseUrl; // e.g. https://api.streampass.com/api/v1
  AuthService({required this.baseUrl});

  static const _tokenKey = 'sp_access_token';
  static const _legacyTokenKey = 'sp_jwt_token';
  static const _refreshKey = 'sp_refresh_token';

  Future<String?> get storedToken async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_tokenKey) ?? prefs.getString(_legacyTokenKey);
  }

  Future<String?> get storedRefreshToken async =>
      (await SharedPreferences.getInstance()).getString(_refreshKey);

  Future<bool> get isLoggedIn async => (await storedToken) != null;

  Future<AuthResult> login(String email, String password) async {
    final res = await http.post(
      Uri.parse('$baseUrl/login'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'email': email, 'password': password}),
    );
    return _handleAuthResponse(res);
  }

  Future<AuthResult> register(String email, String password) async {
    final res = await http.post(
      Uri.parse('$baseUrl/register'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'email': email, 'password': password}),
    );
    return _handleAuthResponse(res);
  }

  Future<void> logout() async {
    final prefs = await SharedPreferences.getInstance();
    final refresh = prefs.getString(_refreshKey);
    if (refresh != null) {
      await http.post(
        Uri.parse('$baseUrl/logout'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'refresh_token': refresh}),
      );
    }
    await prefs.remove(_tokenKey);
    await prefs.remove(_legacyTokenKey);
    await prefs.remove(_refreshKey);
  }

  Future<AuthResult> _handleAuthResponse(http.Response res) async {
    if (res.statusCode != 200 && res.statusCode != 201) {
      return AuthResult(success: false, error: _errorMessage(res));
    }

    final body = jsonDecode(res.body) as Map<String, dynamic>;
    final token = body['access_token'] as String? ?? body['token'] as String?;
    final refresh = body['refresh_token'] as String?;

    if (token == null) {
      return AuthResult(success: false, error: 'Некорректный ответ сервера');
    }

    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_tokenKey, token);
    await prefs.remove(_legacyTokenKey);
    if (refresh != null) await prefs.setString(_refreshKey, refresh);

    return AuthResult(success: true, token: token, refreshToken: refresh);
  }

  String _errorMessage(http.Response res) {
    var message = 'Не удалось выполнить вход';
    try {
      final body = jsonDecode(res.body) as Map<String, dynamic>;
      final error = body['error'];
      if (error is Map<String, dynamic>) {
        message = error['message'] as String? ?? message;
      } else {
        message = body['message'] as String? ?? message;
      }
    } catch (_) {}
    return message;
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
