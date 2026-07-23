import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

/// Wraps POST /register, POST /login, POST /logout and local token storage.
/// Endpoints per StreamPass API spec (section 13).
class AuthService {
  final String baseUrl; // e.g. https://api.streampass.com/api/v1
  AuthService({required this.baseUrl});

  static const _tokenKey = 'sp_jwt_token';
  static const _refreshKey = 'sp_refresh_token';

  Future<String?> get storedToken async =>
      (await SharedPreferences.getInstance()).getString(_tokenKey);

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
    final token = await storedToken;
    if (token != null) {
      await http.post(
        Uri.parse('$baseUrl/logout'),
        headers: {'Authorization': 'Bearer $token'},
      );
    }
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_tokenKey);
    await prefs.remove(_refreshKey);
  }

  Future<AuthResult> _handleAuthResponse(http.Response res) async {
    if (res.statusCode != 200 && res.statusCode != 201) {
      String message = 'Не удалось выполнить вход';
      try {
        final body = jsonDecode(res.body) as Map<String, dynamic>;
        message = body['message'] ?? message;
      } catch (_) {}
      return AuthResult(success: false, error: message);
    }

    final body = jsonDecode(res.body) as Map<String, dynamic>;
    final token = body['token'] as String?;
    final refresh = body['refresh_token'] as String?;

    if (token == null) {
      return AuthResult(success: false, error: 'Некорректный ответ сервера');
    }

    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_tokenKey, token);
    if (refresh != null) await prefs.setString(_refreshKey, refresh);

    return AuthResult(success: true, token: token);
  }
}

class AuthResult {
  final bool success;
  final String? token;
  final String? error;
  AuthResult({required this.success, this.token, this.error});
}
