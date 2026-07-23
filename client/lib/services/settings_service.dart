import 'dart:convert';
import 'package:flutter/services.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Persists the toggles from ТЗ §20 ("Настройки"):
/// Автозапуск / Автоподключение / Автовыбор Relay / Исключения.
class AppSettings {
  final bool autostart;
  final bool autoConnect;
  final bool autoSelectRelay;
  final List<String> exclusions;

  const AppSettings({
    this.autostart = false,
    this.autoConnect = false,
    this.autoSelectRelay = true,
    this.exclusions = const [],
  });

  AppSettings copyWith({
    bool? autostart,
    bool? autoConnect,
    bool? autoSelectRelay,
    List<String>? exclusions,
  }) {
    return AppSettings(
      autostart: autostart ?? this.autostart,
      autoConnect: autoConnect ?? this.autoConnect,
      autoSelectRelay: autoSelectRelay ?? this.autoSelectRelay,
      exclusions: exclusions ?? this.exclusions,
    );
  }
}

class SettingsService {
  static const _kAutostart = 'sp_autostart';
  static const _kAutoConnect = 'sp_auto_connect';
  static const _kAutoRelay = 'sp_auto_relay';
  static const _kExclusions = 'sp_exclusions';

  // Mirrors autostart/autoConnect into native SharedPreferences so
  // BootReceiver (which runs outside the Flutter engine) can read them
  // without depending on the shared_preferences plugin's internal format.
  static const _nativeChannel = MethodChannel('streampass/settings');

  Future<AppSettings> load() async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(_kExclusions);
    final exclusions = raw != null
        ? List<String>.from(jsonDecode(raw) as List)
        : <String>[];

    return AppSettings(
      autostart: prefs.getBool(_kAutostart) ?? false,
      autoConnect: prefs.getBool(_kAutoConnect) ?? false,
      autoSelectRelay: prefs.getBool(_kAutoRelay) ?? true,
      exclusions: exclusions,
    );
  }

  Future<void> setAutostart(bool value) async {
    (await SharedPreferences.getInstance()).setBool(_kAutostart, value);
    await _nativeChannel.invokeMethod('setAutostart', value);
  }

  Future<void> setAutoConnect(bool value) async {
    (await SharedPreferences.getInstance()).setBool(_kAutoConnect, value);
    await _nativeChannel.invokeMethod('setAutoConnect', value);
  }

  Future<void> setAutoSelectRelay(bool value) async =>
      (await SharedPreferences.getInstance()).setBool(_kAutoRelay, value);

  Future<void> setExclusions(List<String> domains) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_kExclusions, jsonEncode(domains));
  }
}
