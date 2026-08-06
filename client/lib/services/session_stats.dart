import 'dart:convert';

import 'package:shared_preferences/shared_preferences.dart';

/// On-device session aggregates for E03 Statistics (BL-044).
/// No URLs / browsing history — only durations, reconnects, RTT samples.
class SessionStatsSnapshot {
  final Duration onlineToday;
  final Duration onlineLast7Days;
  final int reconnectsToday;
  final int reconnectsLast7Days;
  final int? averageRttMs;
  final int rttSampleCount;

  const SessionStatsSnapshot({
    required this.onlineToday,
    required this.onlineLast7Days,
    required this.reconnectsToday,
    required this.reconnectsLast7Days,
    this.averageRttMs,
    this.rttSampleCount = 0,
  });

  bool get hasData =>
      onlineLast7Days > Duration.zero ||
      reconnectsLast7Days > 0 ||
      rttSampleCount > 0;
}

class SessionStatsService {
  static const _kOnline = 'sp_stats_online_sec';
  static const _kReconnects = 'sp_stats_reconnects';
  static const _kRttSum = 'sp_stats_rtt_sum';
  static const _kRttCount = 'sp_stats_rtt_count';
  static const _kRttDay = 'sp_stats_rtt_day';

  /// Day key YYYY-MM-DD in local time.
  static String dayKey([DateTime? now]) {
    final d = (now ?? DateTime.now()).toLocal();
    final mm = d.month.toString().padLeft(2, '0');
    final dd = d.day.toString().padLeft(2, '0');
    return '${d.year}-$mm-$dd';
  }

  Future<Map<String, int>> _readIntMap(String key) async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(key);
    if (raw == null || raw.isEmpty) return {};
    try {
      final map = jsonDecode(raw) as Map<String, dynamic>;
      return map.map((k, v) => MapEntry(k, (v as num).toInt()));
    } catch (_) {
      return {};
    }
  }

  Future<void> _writeIntMap(String key, Map<String, int> map) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(key, jsonEncode(map));
  }

  Future<void> _prune(Map<String, int> map, {int keepDays = 14}) async {
    final cutoff = DateTime.now().toLocal().subtract(Duration(days: keepDays));
    map.removeWhere((k, _) {
      final parts = k.split('-');
      if (parts.length != 3) return true;
      final dt = DateTime.tryParse(k);
      return dt == null || dt.isBefore(cutoff);
    });
  }

  /// Add [seconds] of connected time to today.
  Future<void> addOnlineSeconds(int seconds) async {
    if (seconds <= 0) return;
    final day = dayKey();
    final map = await _readIntMap(_kOnline);
    map[day] = (map[day] ?? 0) + seconds;
    await _prune(map);
    await _writeIntMap(_kOnline, map);
  }

  Future<void> recordReconnect() async {
    final day = dayKey();
    final map = await _readIntMap(_kReconnects);
    map[day] = (map[day] ?? 0) + 1;
    await _prune(map);
    await _writeIntMap(_kReconnects, map);
  }

  Future<void> recordRtt(int rttMs) async {
    if (rttMs <= 0) return;
    final prefs = await SharedPreferences.getInstance();
    final day = dayKey();
    final storedDay = prefs.getString(_kRttDay);
    var sum = prefs.getInt(_kRttSum) ?? 0;
    var count = prefs.getInt(_kRttCount) ?? 0;
    if (storedDay != day) {
      sum = 0;
      count = 0;
    }
    sum += rttMs;
    count += 1;
    await prefs.setString(_kRttDay, day);
    await prefs.setInt(_kRttSum, sum);
    await prefs.setInt(_kRttCount, count);
  }

  Future<SessionStatsSnapshot> load() async {
    final online = await _readIntMap(_kOnline);
    final reconnects = await _readIntMap(_kReconnects);
    final today = dayKey();
    final last7 = <String>{};
    for (var i = 0; i < 7; i++) {
      last7.add(dayKey(DateTime.now().toLocal().subtract(Duration(days: i))));
    }

    int sumOnline(Iterable<String> days) =>
        days.fold(0, (acc, d) => acc + (online[d] ?? 0));
    int sumReconnects(Iterable<String> days) =>
        days.fold(0, (acc, d) => acc + (reconnects[d] ?? 0));

    final prefs = await SharedPreferences.getInstance();
    final rttDay = prefs.getString(_kRttDay);
    final rttSum = prefs.getInt(_kRttSum) ?? 0;
    final rttCount = prefs.getInt(_kRttCount) ?? 0;
    final avg = (rttDay == today && rttCount > 0) ? (rttSum ~/ rttCount) : null;

    return SessionStatsSnapshot(
      onlineToday: Duration(seconds: online[today] ?? 0),
      onlineLast7Days: Duration(seconds: sumOnline(last7)),
      reconnectsToday: reconnects[today] ?? 0,
      reconnectsLast7Days: sumReconnects(last7),
      averageRttMs: avg,
      rttSampleCount: (rttDay == today) ? rttCount : 0,
    );
  }
}
