import 'package:flutter_test/flutter_test.dart';
import 'package:streampass/services/region_catalog.dart';
import 'package:streampass/services/relay_picker.dart';
import 'package:streampass/services/streampass_api.dart';

RelayServer _srv({
  required String id,
  required String region,
  bool healthy = true,
  double load = 0.5,
  int rtt = 40,
}) {
  return RelayServer(
    id: id,
    region: region,
    regionName: regionLabel(region),
    host: '1.2.3.4',
    port: 443,
    healthy: healthy,
    loadRatio: load,
    rttMs: rtt,
    connectionConfig: 'hysteria2://x',
  );
}

void main() {
  test('normalizeRegionCode maps legacy NL and cities', () {
    expect(normalizeRegionCode('NL'), 'nl');
    expect(normalizeRegionCode('Amsterdam'), 'nl');
    expect(normalizeRegionCode('Frankfurt'), 'de');
    expect(normalizeRegionCode('Warsaw'), 'pl');
    expect(normalizeRegionCode('Helsinki'), 'fi');
  });

  test('regionFlag uses ISO codes', () {
    expect(regionFlag('nl'), 'NL');
    expect(regionFlag('NL'), 'NL');
    expect(regionFlag('de'), 'DE');
  });

  test('pickBestRelay prefers lower load then rtt', () {
    final picked = pickBestRelay([
      _srv(id: 'a', region: 'nl', load: 0.8, rtt: 10),
      _srv(id: 'b', region: 'nl', load: 0.2, rtt: 50),
    ]);
    expect(picked?.id, 'b');
  });

  test('pickBestRelay filters by preferred region', () {
    final picked = pickBestRelay(
      [
        _srv(id: 'nl1', region: 'nl', load: 0.1, rtt: 10),
        _srv(id: 'pl1', region: 'pl', load: 0.9, rtt: 80),
      ],
      preferredRegion: 'pl',
    );
    expect(picked?.id, 'pl1');
  });

  test('pickBestRelay pins server when autoSelect is false', () {
    final picked = pickBestRelay(
      [
        _srv(id: 'nl1', region: 'nl', load: 0.1, rtt: 10),
        _srv(id: 'nl2', region: 'nl', load: 0.9, rtt: 80),
      ],
      preferredServerId: 'nl2',
      autoSelect: false,
    );
    expect(picked?.id, 'nl2');
  });

  test('shouldFailoverRelay when current unhealthy', () {
    final current = _srv(id: 'nl1', region: 'nl', healthy: false);
    final best = _srv(id: 'nl2', region: 'nl');
    final servers = [
      _srv(id: 'nl1', region: 'nl', healthy: false),
      best,
    ];
    expect(
      shouldFailoverRelay(
        current: current,
        servers: servers,
        best: best,
        autoSelect: true,
      ),
      isTrue,
    );
  });

  test('shouldFailoverRelay false when current healthy and best same', () {
    final current = _srv(id: 'nl1', region: 'nl');
    expect(
      shouldFailoverRelay(
        current: current,
        servers: [current],
        best: current,
        autoSelect: true,
      ),
      isFalse,
    );
  });

  test('shouldFailoverRelay respects autoSelect off', () {
    final current = _srv(id: 'nl1', region: 'nl', healthy: false);
    final best = _srv(id: 'nl2', region: 'nl');
    expect(
      shouldFailoverRelay(
        current: current,
        servers: [current, best],
        best: best,
        autoSelect: false,
      ),
      isFalse,
    );
  });
}
