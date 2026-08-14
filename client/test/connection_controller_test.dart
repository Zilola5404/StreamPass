import 'package:flutter_test/flutter_test.dart';
import 'package:streampass/services/connection_controller.dart';
import 'package:streampass/services/vpn_channel.dart';
import 'package:streampass/services/windows_traffic_engine.dart';

void main() {
  tearDown(() {
    ConnectionController.instance.debugReset();
  });

  test('Android policy: connected event is user-visible without extra health flag', () {
    final c = ConnectionController.instance;
    c.policy = ConnectedUiPolicy.nativeConnectedMeansReady;
    c.debugApply(VpnStatusUpdate(VpnEvent.connected, relayName: 'nl-native-1'));
    expect(c.showConnected, isTrue);
    expect(c.trafficReady, isTrue);
  });

  test('Windows policy: handshake connected is not user-visible until trafficReady', () {
    final c = ConnectionController.instance;
    c.policy = ConnectedUiPolicy.requireTrafficReady;
    c.debugApply(VpnStatusUpdate(VpnEvent.connected, relayName: 'nl-native-1'));
    expect(c.event, VpnEvent.connected);
    expect(c.showConnected, isFalse);
    expect(c.trafficReady, isFalse);

    c.markTrafficReady();
    expect(c.showConnected, isTrue);
    expect(c.trafficReady, isTrue);
  });

  test('disconnect clears trafficReady', () {
    final c = ConnectionController.instance;
    c.policy = ConnectedUiPolicy.requireTrafficReady;
    c.debugApply(VpnStatusUpdate(VpnEvent.connected));
    c.markTrafficReady();
    c.debugApply(VpnStatusUpdate(VpnEvent.disconnected));
    expect(c.showConnected, isFalse);
    expect(c.trafficReady, isFalse);
  });

  test('WindowsTrafficEngine starts disconnected', () async {
    final engine = WindowsTrafficEngine.instance;
    expect(engine.last.event, isNot(VpnEvent.connected));
  });
}
