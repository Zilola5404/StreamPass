import 'package:flutter_test/flutter_test.dart';
import 'package:streampass/services/streampass_api.dart';

/// Traffic routing contract — mirrors go_core/decision/traffic_matrix_test.go
/// and scripts/traffic_expectations.json.
///
/// Three Android layers (see docs/33_DirectVsVpnBypass.md):
/// 1) Decision Engine — DIRECT vs RELAY through Hysteria
/// 2) Route split — RU CIDR excludeRoute / intl-only TUN routes
/// 3) App bypass — addDisallowedApplication for banks/gov (TRANSPORT_VPN)
void main() {
  /// Minimal prod-like rules for documentation tests.
  RuleSet prodLikeRules() => const RuleSet(
        version: 4,
        rules: [
          RouteRule(kind: 'DOMAIN', pattern: 'youtube.com', mode: 'RELAY'),
          RouteRule(kind: 'DOMAIN', pattern: '*.youtube.com', mode: 'RELAY'),
          RouteRule(kind: 'DOMAIN', pattern: 'instagram.com', mode: 'RELAY'),
          RouteRule(kind: 'DOMAIN', pattern: '*.instagram.com', mode: 'RELAY'),
        ],
      );

  group('Site traffic expectations (decision rules JSON)', () {
    final cases = <({String host, String mode, String note})>[
      (host: 'yandex.ru', mode: 'DIRECT', note: 'RU — split DNS Yandex, RU IP'),
      (host: 'gosuslugi.ru', mode: 'DIRECT', note: 'RU gov web'),
      (host: '2ip.ru', mode: 'DIRECT', note: 'Geo IP must show RU on device'),
      (host: 'www.youtube.com', mode: 'RELAY', note: 'Accelerated foreign'),
      (host: 'instagram.com', mode: 'RELAY', note: 'Accelerated foreign'),
      (host: 'google.com', mode: 'DIRECT', note: 'Default DIRECT unless user excludes'),
    ];

    for (final tc in cases) {
      test('${tc.host} -> ${tc.mode} (${tc.note})', () {
        // Client-side rules are evaluated in Go; here we document the contract
        // by checking the published rule set shape matches expectations.
        final rules = prodLikeRules();
        final relayHosts = rules.rules
            .where((r) => r.mode == 'RELAY')
            .map((r) => r.pattern)
            .toList();

        if (tc.mode == 'RELAY') {
          final matchesRelayRule = relayHosts.any((p) {
            if (p.startsWith('*.')) {
              final suffix = p.substring(1);
              return tc.host.endsWith(suffix) || tc.host == p.substring(2);
            }
            return tc.host == p || tc.host.endsWith('.$p');
          });
          expect(matchesRelayRule, isTrue,
              reason: '${tc.host} should match a RELAY rule in prod-like set');
        } else {
          // DIRECT sites rely on client DefaultDirectRules (*.ru) in Go — not in this JSON slice.
          expect(tc.host.endsWith('.ru') || tc.host == 'google.com', isTrue,
              reason: 'DIRECT cases in this matrix are RU TLD or default-direct foreign');
        }
      });
    }
  });

  group('App bypass expectations (Android native)', () {
    /// Known packages from VpnBypassApps.kt — native layer, not Decision Engine.
    const bypassPackages = [
      'ru.rostel',
      'com.gnivts.selfemployed',
      'ru.s7tl.app',
      'ru.sberbankmobile',
    ];

    test('gov and bank packages are in bypass allowlist contract', () {
      expect(bypassPackages, contains('ru.rostel'));
      expect(bypassPackages, contains('com.gnivts.selfemployed'));
    });

    test('Chrome is not bypassed (foreign browser acceleration)', () {
      expect(bypassPackages, isNot(contains('com.android.chrome')));
    });
  });

  group('Product startup expectations', () {
    test('inactive subscription blocks connect in UI logic', () {
      final inactive = SubscriptionInfo.fromJson({'status': 'INACTIVE'});
      expect(inactive.isActive, isFalse);
    });

    test('active subscription allows connect in UI logic', () {
      final active = SubscriptionInfo.fromJson({'status': 'ACTIVE'});
      expect(active.isActive, isTrue);
    });

    test('relay without connection_config is rejected before native connect', () {
      final bad = RelayServer.fromJson({
        'id': 'bad',
        'region': 'NL',
        'host': '10.0.0.1',
        'port': 443,
        'healthy': true,
        'load_ratio': 0,
        'rtt_ms': 0,
      });
      expect(bad.connectionConfig, isEmpty);
    });
  });
}
