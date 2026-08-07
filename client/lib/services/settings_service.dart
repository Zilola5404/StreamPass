import 'dart:convert';
import 'package:flutter/services.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Persists the toggles from ТЗ §20 ("Настройки"):
/// Автозапуск / Автоподключение / Автовыбор Relay / Исключения /
/// preferred region & server (BL-026) + network diagnostics modes.
class AppSettings {
  final bool autostart;
  final bool autoConnect;
  final bool autoSelectRelay;
  final List<String> exclusions;
  /// Canonical region code (`de`/`nl`/`pl`/`fi`) or empty for any.
  final String preferredRegion;
  /// Concrete relay id when auto-select is off.
  final String preferredServerId;
  /// Extra Android packages excluded from VPN (OS bypass).
  final List<String> bypassPackages;
  /// Upload routing diagnostics to backend (TASK-01).
  final bool diagnosticsEnabled;
  /// split | full_relay | direct_test | tcp_only
  final String networkMode;
  /// TUN/Hysteria MTU: 1280 | 1350 | 1400
  final int mtu;
  /// Drop UDP/443 so apps fall back to TCP (QUIC off).
  final bool blockUdp443;

  const AppSettings({
    this.autostart = false,
    this.autoConnect = false,
    this.autoSelectRelay = true,
    this.exclusions = const [],
    this.preferredRegion = '',
    this.preferredServerId = '',
    this.bypassPackages = const [],
    this.diagnosticsEnabled = true,
    this.networkMode = 'split',
    this.mtu = 1400,
    this.blockUdp443 = false,
  });

  AppSettings copyWith({
    bool? autostart,
    bool? autoConnect,
    bool? autoSelectRelay,
    List<String>? exclusions,
    String? preferredRegion,
    String? preferredServerId,
    List<String>? bypassPackages,
    bool? diagnosticsEnabled,
    String? networkMode,
    int? mtu,
    bool? blockUdp443,
  }) {
    return AppSettings(
      autostart: autostart ?? this.autostart,
      autoConnect: autoConnect ?? this.autoConnect,
      autoSelectRelay: autoSelectRelay ?? this.autoSelectRelay,
      exclusions: exclusions ?? this.exclusions,
      preferredRegion: preferredRegion ?? this.preferredRegion,
      preferredServerId: preferredServerId ?? this.preferredServerId,
      bypassPackages: bypassPackages ?? this.bypassPackages,
      diagnosticsEnabled: diagnosticsEnabled ?? this.diagnosticsEnabled,
      networkMode: networkMode ?? this.networkMode,
      mtu: mtu ?? this.mtu,
      blockUdp443: blockUdp443 ?? this.blockUdp443,
    );
  }

  String get optionsJson => jsonEncode({
        'networkMode': networkMode,
        'mtu': mtu,
        'blockUdp443': blockUdp443 || networkMode == 'tcp_only',
      });
}

class SettingsService {
  static const _kAutostart = 'sp_autostart';
  static const _kAutoConnect = 'sp_auto_connect';
  static const _kAutoRelay = 'sp_auto_relay';
  static const _kExclusions = 'sp_exclusions';
  static const _kPreferredRegion = 'sp_preferred_region';
  static const _kPreferredServer = 'sp_preferred_server';
  static const _kBypassPackages = 'sp_bypass_packages';
  static const _kDiagnostics = 'sp_diagnostics_enabled';
  static const _kNetworkMode = 'sp_network_mode';
  static const _kMtu = 'sp_mtu';
  static const _kBlockUdp443 = 'sp_block_udp443';

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
    final bypassRaw = prefs.getString(_kBypassPackages);
    final bypassPackages = bypassRaw != null
        ? List<String>.from(jsonDecode(bypassRaw) as List)
        : <String>[];
    final mode = prefs.getString(_kNetworkMode) ?? 'split';
    final mtu = prefs.getInt(_kMtu) ?? 1400;

    return AppSettings(
      autostart: prefs.getBool(_kAutostart) ?? false,
      autoConnect: prefs.getBool(_kAutoConnect) ?? false,
      autoSelectRelay: prefs.getBool(_kAutoRelay) ?? true,
      exclusions: exclusions,
      preferredRegion: prefs.getString(_kPreferredRegion) ?? '',
      preferredServerId: prefs.getString(_kPreferredServer) ?? '',
      bypassPackages: bypassPackages,
      diagnosticsEnabled: prefs.getBool(_kDiagnostics) ?? true,
      networkMode: mode,
      mtu: mtu,
      blockUdp443: prefs.getBool(_kBlockUdp443) ?? false,
    );
  }

  Future<void> setAutostart(bool value) async {
    (await SharedPreferences.getInstance()).setBool(_kAutostart, value);
    try {
      await _nativeChannel.invokeMethod('setAutostart', value);
    } on PlatformException {
      // Native channel unavailable in tests / non-Android.
    } on MissingPluginException {
      // ignore
    }
  }

  Future<void> setAutoConnect(bool value) async {
    (await SharedPreferences.getInstance()).setBool(_kAutoConnect, value);
    try {
      await _nativeChannel.invokeMethod('setAutoConnect', value);
    } on PlatformException {
      // ignore
    } on MissingPluginException {
      // ignore
    }
  }

  Future<void> setAutoSelectRelay(bool value) async =>
      (await SharedPreferences.getInstance()).setBool(_kAutoRelay, value);

  Future<void> setPreferredRegion(String code) async =>
      (await SharedPreferences.getInstance()).setString(_kPreferredRegion, code);

  Future<void> setPreferredServerId(String id) async =>
      (await SharedPreferences.getInstance()).setString(_kPreferredServer, id);

  Future<void> setExclusions(List<String> domains) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_kExclusions, jsonEncode(domains));
  }

  Future<void> setBypassPackages(List<String> packages) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_kBypassPackages, jsonEncode(packages));
  }

  Future<void> setDiagnosticsEnabled(bool value) async =>
      (await SharedPreferences.getInstance()).setBool(_kDiagnostics, value);

  Future<void> setNetworkMode(String mode) async =>
      (await SharedPreferences.getInstance()).setString(_kNetworkMode, mode);

  Future<void> setMtu(int mtu) async =>
      (await SharedPreferences.getInstance()).setInt(_kMtu, mtu);

  Future<void> setBlockUdp443(bool value) async =>
      (await SharedPreferences.getInstance()).setBool(_kBlockUdp443, value);
}
