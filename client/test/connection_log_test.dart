import 'package:flutter_test/flutter_test.dart';
import 'package:streampass/services/connection_log.dart';

void main() {
  setUp(() {
    ConnectionLog.instance.clear();
  });

  test('beginConnectSession clears previous entries and logs start', () {
    final log = ConnectionLog.instance;
    log.info('old', 'should be removed');
    log.beginConnectSession(relayId: 'nl-native-1', host: '212.43.156.33');

    expect(log.entries.length, 1);
    expect(log.entries.first.tag, 'connect');
    expect(log.entries.first.details?['relayId'], 'nl-native-1');
  });

  test('ring buffer keeps at most maxEntries', () {
    final log = ConnectionLog.instance;
    for (var i = 0; i < ConnectionLog.maxEntries + 10; i++) {
      log.info('test', 'line $i');
    }
    expect(log.entries.length, ConnectionLog.maxEntries);
    expect(log.entries.first.message, contains('line 10'));
  });

  test('exportText joins all lines', () {
    final log = ConnectionLog.instance;
    log.info('api', 'servers loaded', {'count': '1'});
    log.error('vpn', 'tunnel failed', {'error': 'timeout'});

    final text = log.exportText();
    expect(text, contains('[api] servers loaded'));
    expect(text, contains('[vpn] tunnel failed'));
    expect(text, contains('count=1'));
  });

  test('error level is recorded separately from info', () {
    final log = ConnectionLog.instance;
    log.error('connect', 'blocked: no relay');

    expect(log.entries.single.level, ConnectionLogLevel.error);
  });
}
