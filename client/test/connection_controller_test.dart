import 'package:flutter_test/flutter_test.dart';
import 'package:streampass/services/connection_controller.dart';
import 'package:streampass/services/streampass_api.dart';
import 'package:streampass/services/vpn_channel.dart';
import 'package:streampass/services/windows_traffic_engine.dart';

void main() {
  tearDown(() async {
    ConnectionController.instance.debugReset();
    await WindowsTrafficEngine.instance.disposeCore();
  });

  test('Android policy: require trafficReady before user-visible connected', () {
    final c = ConnectionController.instance;
    c.policy = ConnectedUiPolicy.requireTrafficReady;
    c.debugApply(VpnStatusUpdate(VpnEvent.connected, relayName: 'nl-native-1'));
    expect(c.showConnected, isFalse);
    expect(c.trafficReady, isFalse);
    c.debugApply(VpnStatusUpdate(
      VpnEvent.connected,
      relayName: 'nl-native-1',
      trafficReadyHint: true,
    ));
    expect(c.showConnected, isTrue);
    expect(c.trafficReady, isTrue);
  });

  test('Legacy policy: connected event is user-visible without extra health flag', () {
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

  test('Windows: traffic_ready before connected status is applied later', () {
    final c = ConnectionController.instance;
    c.policy = ConnectedUiPolicy.requireTrafficReady;
    c.markTrafficReady(); // latch before connected
    expect(c.showConnected, isFalse);
    c.debugApply(VpnStatusUpdate(VpnEvent.connected, relayName: 'pl-1'));
    expect(c.trafficReady, isTrue);
    expect(c.showConnected, isTrue);
  });

  test('disconnect clears trafficReady', () {
    final c = ConnectionController.instance;
    c.policy = ConnectedUiPolicy.requireTrafficReady;
    c.debugApply(VpnStatusUpdate(VpnEvent.connected));
    c.markTrafficReady();
    c.debugApply(VpnStatusUpdate(VpnEvent.disconnected));
    expect(c.showConnected, isFalse);
    expect(c.trafficReady, isFalse);
    expect(c.sessionStartedAt, isNull);
  });

  test('session clock starts once and survives repeat connected events', () async {
    final c = ConnectionController.instance;
    c.policy = ConnectedUiPolicy.requireTrafficReady;
    c.debugApply(VpnStatusUpdate(VpnEvent.connected, relayName: 'nl-1'));
    expect(c.sessionStartedAt, isNull);
    c.markTrafficReady();
    final started = c.sessionStartedAt;
    expect(started, isNotNull);
    await Future<void>.delayed(const Duration(milliseconds: 20));
    c.debugApply(VpnStatusUpdate(
      VpnEvent.connected,
      relayName: 'nl-1',
      pingMs: 42,
    ));
    expect(c.sessionStartedAt, started);
    expect(c.liveSessionDuration.inMilliseconds, greaterThanOrEqualTo(20));
  });

  test('WindowsTrafficEngine starts disconnected', () async {
    final engine = WindowsTrafficEngine.instance;
    expect(engine.last.event, isNot(VpnEvent.connected));
  });

  test('WindowsTrafficEngine coalesce: second connect shares in-flight Future', () async {
    final engine = WindowsTrafficEngine.instance;
    // Without a real core binary this throws; both callers must get the same error
    // completion (not hang / ignore without Future).
    final server = const RelayServer(
      id: 't',
      host: '127.0.0.1',
      port: 443,
      region: 'test',
      healthy: true,
      loadRatio: 0,
      rttMs: 0,
      connectionConfig: 'hysteria2://x@127.0.0.1:443/?insecure=1',
    );
    Object? err1;
    Object? err2;
    final f1 = engine.connect(server).catchError((Object e) {
      err1 = e;
      return false;
    });
    final f2 = engine.connect(server).catchError((Object e) {
      err2 = e;
      return false;
    });
    await Future.wait([f1, f2]);
    expect(err1, isNotNull);
    expect(err2, isNotNull);
    expect(err1.toString(), err2.toString());
    expect(engine.last.event, isNot(VpnEvent.connecting));
  });
}
