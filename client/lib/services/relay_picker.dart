import '../services/streampass_api.dart';
import 'region_catalog.dart';

/// Relays eligible for connect (healthy + non-empty hysteria2 config).
List<RelayServer> connectableRelays(List<RelayServer> servers) => servers
    .where((s) => s.healthy && s.connectionConfig.isNotEmpty)
    .toList();

/// Picks the best healthy relay, optionally constrained to a preferred
/// region and/or a pinned server id (when auto-select is off).
/// Returns null when no healthy relay is available (never falls back to unhealthy).
RelayServer? pickBestRelay(
  List<RelayServer> servers, {
  String? preferredRegion,
  String? preferredServerId,
  bool autoSelect = true,
}) {
  if (servers.isEmpty) return null;

  final wantRegion = normalizeRegionCode(preferredRegion);
  final pool = connectableRelays(servers);
  if (pool.isEmpty) return null;

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

/// Whether auto-mode should leave [current] for [best] (BL-047).
/// Failover when current is missing from catalog or marked unhealthy.
bool shouldFailoverRelay({
  required RelayServer? current,
  required List<RelayServer> servers,
  required RelayServer? best,
  required bool autoSelect,
}) {
  if (!autoSelect || best == null) return false;
  if (current == null) return true;
  if (current.id == best.id) return false;

  RelayServer? fresh;
  for (final s in servers) {
    if (s.id == current.id) {
      fresh = s;
      break;
    }
  }
  if (fresh == null || !fresh.healthy) return true;
  return false;
}
