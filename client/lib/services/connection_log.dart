import 'dart:async';
import 'dart:collection';

import 'package:flutter/foundation.dart';

enum ConnectionLogLevel { info, warn, error }

/// One connection-diagnostic event (API, auth, VPN native bridge).
class ConnectionLogEntry {
  final DateTime at;
  final ConnectionLogLevel level;
  final String tag;
  final String message;
  final Map<String, String>? details;

  const ConnectionLogEntry({
    required this.at,
    required this.level,
    required this.tag,
    required this.message,
    this.details,
  });

  String get levelLabel => level.name.toUpperCase();

  String formatLine() {
    final ts = at.toIso8601String();
    final extra = details == null || details!.isEmpty
        ? ''
        : ' ${details!.entries.map((e) => '${e.key}=${e.value}').join(' ')}';
    return '[$ts] [$levelLabel] [$tag] $message$extra';
  }

  Map<String, dynamic> toJson() => {
        'at': at.toIso8601String(),
        'level': level.name,
        'tag': tag,
        'message': message,
        if (details != null) 'details': details,
      };

  factory ConnectionLogEntry.fromJson(Map<String, dynamic> json) {
    return ConnectionLogEntry(
      at: DateTime.parse(json['at'] as String),
      level: ConnectionLogLevel.values.firstWhere(
        (l) => l.name == json['level'],
        orElse: () => ConnectionLogLevel.info,
      ),
      tag: json['tag'] as String? ?? 'unknown',
      message: json['message'] as String? ?? '',
      details: json['details'] != null
          ? Map<String, String>.from(json['details'] as Map)
          : null,
    );
  }
}

/// In-memory ring buffer of connection events for on-device diagnostics.
class ConnectionLog {
  ConnectionLog._();
  static final ConnectionLog instance = ConnectionLog._();

  static const maxEntries = 500;
  final Queue<ConnectionLogEntry> _entries = Queue<ConnectionLogEntry>();
  final StreamController<ConnectionLogEntry> _controller =
      StreamController<ConnectionLogEntry>.broadcast();

  Stream<ConnectionLogEntry> get stream => _controller.stream;
  List<ConnectionLogEntry> get entries => List.unmodifiable(_entries);

  void info(String tag, String message, [Map<String, String>? details]) =>
      _add(ConnectionLogLevel.info, tag, message, details);

  void warn(String tag, String message, [Map<String, String>? details]) =>
      _add(ConnectionLogLevel.warn, tag, message, details);

  void error(String tag, String message, [Map<String, String>? details]) =>
      _add(ConnectionLogLevel.error, tag, message, details);

  void _add(
    ConnectionLogLevel level,
    String tag,
    String message,
    Map<String, String>? details,
  ) {
    final entry = ConnectionLogEntry(
      at: DateTime.now().toUtc(),
      level: level,
      tag: tag,
      message: message,
      details: details,
    );
    _entries.add(entry);
    while (_entries.length > maxEntries) {
      _entries.removeFirst();
    }
    debugPrint('[StreamPassConnect] ${entry.formatLine()}');
    if (!_controller.isClosed) {
      _controller.add(entry);
    }
  }

  void clear() => _entries.clear();

  void importNativeLines(Iterable<String> lines, {bool replaceExisting = true}) {
    if (replaceExisting) {
      final kept = _entries.where((e) => e.tag != 'native').toList();
      _entries
        ..clear()
        ..addAll(kept);
    }
    for (final line in lines) {
      final trimmed = line.trim();
      if (trimmed.isEmpty) continue;
      info('native', trimmed);
    }
  }

  String exportText() => _entries.map((e) => e.formatLine()).join('\n');

  /// Starts a new connect attempt — clears previous session log.
  void beginConnectSession({required String relayId, required String host}) {
    clear();
    info('connect', 'session started', {
      'relayId': relayId,
      'host': host,
    });
  }
}
