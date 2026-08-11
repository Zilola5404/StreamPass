import 'dart:math';

import 'package:flutter/foundation.dart';

import 'token_storage.dart';

/// Stable per-install device identity for multi-device limit (BL-049).
class DeviceIdentity {
  DeviceIdentity({TokenStorage? storage}) : _tokens = storage ?? TokenStorage.secure();

  final TokenStorage _tokens;

  static const _idKey = 'sp_device_id';
  static const _nameKey = 'sp_device_name';

  /// Returns a persistent opaque device id (created once per install).
  Future<String> deviceId() async {
    final existing = await _tokens.read(_idKey);
    if (existing != null && existing.isNotEmpty) return existing;
    final id = _generateId();
    await _tokens.write(_idKey, id);
    return id;
  }

  /// Human-readable label shown in E10 device list.
  Future<String> deviceName() async {
    final existing = await _tokens.read(_nameKey);
    if (existing != null && existing.isNotEmpty) return existing;
    final name = defaultTargetPlatform == TargetPlatform.android
        ? 'Android'
        : defaultTargetPlatform == TargetPlatform.iOS
            ? 'iPhone'
            : 'Устройство';
    await _tokens.write(_nameKey, name);
    return name;
  }

  static String _generateId() {
    final rnd = Random.secure();
    final bytes = List<int>.generate(16, (_) => rnd.nextInt(256));
    return bytes.map((b) => b.toRadixString(16).padLeft(2, '0')).join();
  }
}
