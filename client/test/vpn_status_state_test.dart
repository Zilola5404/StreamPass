import 'package:flutter_test/flutter_test.dart';
import 'package:streampass/services/vpn_channel.dart';

void main() {
  tearDown(() {
    VpnChannel.debugSetLastStatus(null);
  });

  test('VpnChannel retains last known status for late listeners', () {
    VpnChannel.debugSetLastStatus(
      VpnStatusUpdate(VpnEvent.connected, relayName: 'Relay NL-01', pingMs: 42),
    );

    final current = VpnChannel.lastStatus;
    expect(current, isNotNull);
    expect(current!.event, equals(VpnEvent.connected));
    expect(current.relayName, equals('Relay NL-01'));
    expect(current.pingMs, equals(42));
  });

  test('late Diagnostics-style read sees Connected not Disconnected', () {
    // Simulate EventChannel push before Settings/Diagnostics opens.
    VpnChannel.debugSetLastStatus(
      VpnStatusUpdate(VpnEvent.connected, relayName: 'nl-native-1', pingMs: 363),
    );

    // Late screen init: only lastStatus is available (no new event yet).
    final lateRead = VpnChannel.lastStatus?.event ?? VpnEvent.disconnected;
    expect(lateRead, equals(VpnEvent.connected));
  });
}
