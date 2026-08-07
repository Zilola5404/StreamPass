import 'dart:async';
import 'dart:collection';

import '../build_info.dart';
import 'connection_log.dart';
import 'native_connect_log.dart';
import 'streampass_api.dart';

/// Parses Go-core `[diag]` / `[route]` lines from the on-device connect log
/// and uploads batches to `POST /api/v1/diag`.
///
/// Stores hostname / https://host origin only — never full URL paths (ТЗ §14).
class DiagUploader {
  DiagUploader({required this.api, this.interval = const Duration(seconds: 20)});

  final StreamPassApi api;
  final Duration interval;

  static final _log = ConnectionLog.instance;
  static final RegExp _diagRe = RegExp(r'\[diag\]\s+(.+)$');
  static final RegExp _kvRe = RegExp(r'(\w+)=([^\s]+)');
  static final RegExp _dnsRe = RegExp(
    r'\[dns\]\s+query\s+(\S+)\s+via=(\S+)(?:\s+rtt=(\d+)ms)?(?:\s+err=(.+))?$',
  );

  Timer? _timer;
  final LinkedHashSet<String> _uploaded = LinkedHashSet<String>();
  bool _inFlight = false;

  void start() {
    _timer?.cancel();
    _timer = Timer.periodic(interval, (_) => flush());
    unawaited(Future<void>.delayed(const Duration(seconds: 3), flush));
  }

  void stop() {
    _timer?.cancel();
    _timer = null;
  }

  Future<void> flush() async {
    if (_inFlight) return;
    _inFlight = true;
    try {
      await NativeConnectLog.pullFromNative();
      final events = <DiagEvent>[];
      for (final entry in _log.entries) {
        final parsed = parseLine(entry.message);
        if (parsed == null) continue;
        final key = _fingerprint(parsed);
        if (_uploaded.contains(key)) continue;
        events.add(parsed.copyWith(
          clientVersion: BuildInfo.label,
          recordedAt: entry.at.toUtc(),
        ));
        _uploaded.add(key);
      }
      if (events.isEmpty) return;
      for (var i = 0; i < events.length; i += 100) {
        final end = (i + 100 > events.length) ? events.length : i + 100;
        await api.uploadDiag(events.sublist(i, end));
      }
      _log.info('diag', 'uploaded ${events.length} routing events');
      while (_uploaded.length > 2000) {
        _uploaded.remove(_uploaded.first);
      }
    } catch (e) {
      _log.warn('diag', 'upload failed', {'error': e.toString()});
    } finally {
      _inFlight = false;
    }
  }

  /// Exposed for unit tests.
  static DiagEvent? parseLine(String raw) {
    final line = raw.trim();
    final diagIdx = line.indexOf('[diag]');
    final dnsIdx = line.indexOf('[dns]');
    final slice = diagIdx >= 0
        ? line.substring(diagIdx)
        : (dnsIdx >= 0 ? line.substring(dnsIdx) : line);
    final diagMatch = _diagRe.firstMatch(slice);
    if (diagMatch != null) {
      final map = <String, String>{};
      for (final m in _kvRe.allMatches(diagMatch.group(1)!)) {
        map[m.group(1)!] = m.group(2)!;
      }
      final port = int.tryParse(map['dest_port'] ?? '') ?? 0;
      final latency = int.tryParse(map['latency_ms'] ?? '') ?? 0;
      final slow = map['slow'] == '1' || map['slow'] == 'true';
      final err = map['error'] ?? '';
      final host = map['host'] ?? '';
      var site = map['site'] ?? '';
      if (site.isEmpty && host.isNotEmpty) {
        site = 'https://$host';
      }
      final reason = map['reason'] ?? '';
      final decision = map['decision'] ?? map['decision_reason'] ?? '';
      final speed = int.tryParse(map['speed_kbps'] ?? '') ?? 0;
      return DiagEvent(
        proto: map['proto'] ?? 'tcp',
        site: site,
        host: host,
        destIp: map['dest_ip'] ?? '',
        destPort: port,
        mode: map['mode'] ?? map['via'] ?? '',
        result: map['result'] ?? 'unknown',
        latencyMs: latency,
        slow: slow || (map['result'] == 'slow'),
        speedKbps: speed,
        reason: reason,
        rule: map['rule'] ?? '',
        decisionReason: decision,
        errorCode: err == '_' || err.isEmpty ? '' : err,
        relayId: map['relay_id'] ?? '',
      );
    }
    final dnsMatch = _dnsRe.firstMatch(slice);
    if (dnsMatch != null) {
      final via = dnsMatch.group(2) ?? '';
      final rtt = int.tryParse(dnsMatch.group(3) ?? '') ?? 0;
      final err = (dnsMatch.group(4) ?? '').trim();
      final fail = via == 'fail' || err.isNotEmpty;
      final host = dnsMatch.group(1) ?? '';
      return DiagEvent(
        proto: 'dns',
        site: host.isEmpty ? '' : 'https://$host',
        host: host,
        destIp: '',
        destPort: 53,
        mode: via.toUpperCase(),
        result: fail ? 'fail' : 'ok',
        latencyMs: rtt,
        slow: rtt >= 1500,
        reason: fail ? 'dns_fail' : 'dns_$via',
        errorCode: err.replaceAll(' ', '_'),
      );
    }
    return null;
  }

  static String _fingerprint(DiagEvent e) {
    return '${e.proto}|${e.site}|${e.host}|${e.destIp}|${e.destPort}|${e.mode}|'
        '${e.result}|${e.latencyMs}|${e.speedKbps}|${e.reason}|${e.rule}|${e.errorCode}';
  }
}
