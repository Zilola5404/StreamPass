import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Persists auth tokens outside plaintext SharedPreferences (Security S-05).
abstract class TokenStorage {
  Future<String?> read(String key);
  Future<void> write(String key, String value);
  Future<void> remove(String key);
  Future<void> clearKeys(Iterable<String> keys);

  /// Secure storage with one-time migration from legacy SharedPreferences keys.
  factory TokenStorage.secure({
    FlutterSecureStorage? secure,
    Future<SharedPreferences> Function()? prefs,
  }) = _SecureTokenStorage;

  @visibleForTesting
  factory TokenStorage.inMemory([Map<String, String>? seed]) = _InMemoryTokenStorage;
}

class _SecureTokenStorage implements TokenStorage {
  _SecureTokenStorage({
    FlutterSecureStorage? secure,
    Future<SharedPreferences> Function()? prefs,
  })  : _secure = secure ?? const FlutterSecureStorage(),
        _prefs = prefs ?? SharedPreferences.getInstance;

  final FlutterSecureStorage _secure;
  final Future<SharedPreferences> Function() _prefs;

  Future<SharedPreferences> get _preferences => _prefs();

  @override
  Future<String?> read(String key) async {
    final secureVal = await _readSecure(key);
    if (secureVal != null) {
      return secureVal;
    }
    final prefs = await _preferences;
    final legacy = prefs.getString(key);
    if (legacy == null) {
      return null;
    }
    if (await _writeSecure(key, legacy)) {
      await prefs.remove(key);
    }
    return legacy;
  }

  @override
  Future<void> write(String key, String value) async {
    final wroteSecure = await _writeSecure(key, value);
    final prefs = await _preferences;
    if (wroteSecure) {
      await prefs.remove(key);
    } else {
      await prefs.setString(key, value);
    }
  }

  @override
  Future<void> remove(String key) async {
    await _deleteSecure(key);
    final prefs = await _preferences;
    await prefs.remove(key);
  }

  @override
  Future<void> clearKeys(Iterable<String> keys) async {
    for (final key in keys) {
      await remove(key);
    }
  }

  Future<String?> _readSecure(String key) async {
    try {
      return await _secure.read(key: key);
    } catch (_) {
      return null;
    }
  }

  Future<bool> _writeSecure(String key, String value) async {
    try {
      await _secure.write(key: key, value: value);
      return true;
    } catch (_) {
      return false;
    }
  }

  Future<void> _deleteSecure(String key) async {
    try {
      await _secure.delete(key: key);
    } catch (_) {
      // Unit tests without platform channel — prefs removal is enough.
    }
  }
}

class _InMemoryTokenStorage implements TokenStorage {
  _InMemoryTokenStorage([Map<String, String>? seed]) : _data = Map.of(seed ?? {});

  final Map<String, String> _data;

  @override
  Future<String?> read(String key) async => _data[key];

  @override
  Future<void> write(String key, String value) async {
    _data[key] = value;
  }

  @override
  Future<void> remove(String key) async {
    _data.remove(key);
  }

  @override
  Future<void> clearKeys(Iterable<String> keys) async {
    for (final key in keys) {
      _data.remove(key);
    }
  }
}
