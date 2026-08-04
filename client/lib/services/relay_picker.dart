import '../services/streampass_api.dart';
import 'region_catalog.dart';

/// Picks the best healthy relay, optionally constrained to a preferred
/// region and/or a pinned server id (when auto-select is off).
RelayServer? pickBestRelay(
  List<RelayServer> servers, {
  String? preferredRegion,
  String? preferredServerId,
  bool autoSelect = true,
}) {
  if (servers.isEmpty) return null;

  final wantRegion = normalizeRegionCode(preferredRegion);
  final healthy = servers.where((s) => s.healthy).toList();
  final pool = healthy.isNotEmpty ? healthy : servers;

  if (!autoSelect && preferredServerId != null && preferredServerId.isNotEmpty) {
    for (final s in pool) {
      if (s.id == preferredServerId) return s;
    }
  }

  var candidates = pool;
  if (wantRegion.isNotEmpty) {
    final filtered = pool
        .where((s) => normalizeRegionCode(s.region) == wantRegion)
        .toList();
    if (filtered.isNotEmpty) {
      candidates = filtered;
    }
  }

  candidates = List<RelayServer>.from(candidates)
    ..sort((a, b) {
      final load = a.loadRatio.compareTo(b.loadRatio);
      if (load != 0) return load;
      return a.rttMs.compareTo(b.rttMs);
    });

  return candidates.isNotEmpty ? candidates.first : null;
}
