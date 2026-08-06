import 'package:flutter_test/flutter_test.dart';
import 'package:streampass/services/connection_log.dart';
import 'package:streampass/services/streampass_api.dart';
import 'package:streampass/services/vpn_channel.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    ConnectionLog.instance.clear();
  });

  test('connect flow logs expected steps when relay is invalid', () async {
    final log = ConnectionLog.instance;
    const relay = RelayServer(
      id: 'bad-relay',
      region: 'XX',
      host: '10.0.0.1',
      port: 443,
      healthy: true,
      loadRatio: 0,
      rttMs: 0,
      connectionConfig: '',
    );

    await expectLater(
      VpnChannel.connect(relay),
      throwsA(isA<VpnConnectException>()),
    );

    expect(log.entries.any((e) => e.tag == 'vpn' && e.level == ConnectionLogLevel.error), isTrue);
    expect(log.entries.any((e) => e.message.contains('empty connection_config')), isTrue);
  });

  test('beginConnectSession records relay host for diagnostics export', () {
    final log = ConnectionLog.instance;
    log.beginConnectSession(relayId: 'nl-native-1', host: '212.43.156.33');
    log.info('vpn', 'MethodChannel connect accepted', {'ok': 'true'});
    log.info('vpn', 'native VPN event', {'event': 'connecting'});
    log.info('vpn', 'native VPN event', {'event': 'connected', 'pingMs': '42'});

    final exported = log.exportText();
    expect(exported, contains('session started'));
    expect(exported, contains('relayId=nl-native-1'));
    expect(exported, contains('event=connected'));
  });
}
