import 'package:flutter_test/flutter_test.dart';
import 'package:streampass/services/streampass_api.dart';

void main() {
  group('RelayServer.fromJson', () {
    test('parses production nl-native-1 payload', () {
      final relay = RelayServer.fromJson({
        'id': 'nl-native-1',
        'region': 'NL',
        'host': '212.43.156.33',
        'port': 443,
        'healthy': true,
        'load_ratio': 0,
        'rtt_ms': 1,
        'connection_config':
            'hysteria2://streampass-secure-auth@212.43.156.33:443/?obfs=salamander&obfs-password=streampass-relay-2024&insecure=1',
      });

      expect(relay.id, 'nl-native-1');
      expect(relay.connectionConfig, startsWith('hysteria2://'));
      expect(relay.connectionConfig, isNotEmpty);
    });

    test('empty connection_config defaults to empty string', () {
      final relay = RelayServer.fromJson({
        'id': 'bad-relay',
        'region': 'NL',
        'host': '10.0.0.1',
        'port': 443,
        'healthy': true,
        'load_ratio': 0,
        'rtt_ms': 0,
      });

      expect(relay.connectionConfig, isEmpty);
    });
  });

  group('SubscriptionInfo.fromJson', () {
    test('ACTIVE status allows VPN connect in UI', () {
      final info = SubscriptionInfo.fromJson({'status': 'ACTIVE'});
      expect(info.isActive, isTrue);
    });

    test('INACTIVE status blocks VPN connect in UI', () {
      final info = SubscriptionInfo.fromJson({'status': 'INACTIVE'});
      expect(info.isActive, isFalse);
    });

    test('future active_until counts as active', () {
      final until = DateTime.now().add(const Duration(days: 7)).toUtc().toIso8601String();
      final info = SubscriptionInfo.fromJson({'status': 'INACTIVE', 'active_until': until});
      expect(info.isActive, isTrue);
    });
  });
}
