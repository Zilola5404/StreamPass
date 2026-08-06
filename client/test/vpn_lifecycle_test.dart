import 'package:flutter_test/flutter_test.dart';
import 'package:streampass/services/connection_log.dart';
import 'package:streampass/services/vpn_channel.dart';

/// Documents and verifies VPN lifecycle: connect → connected → disconnect.
/// Native side must emit disconnected without killing the Flutter process
/// (StreamPassVpnService.onDestroy must not call stopSelf()).
void main() {
  tearDown(() {
    VpnChannel.debugSetLastStatus(null);
    ConnectionLog.instance.clear();
  });

  group('VPN lifecycle state', () {
    test('connected then disconnected leaves lastStatus disconnected', () {
      VpnChannel.debugSetLastStatus(
        VpnStatusUpdate(VpnEvent.connected, relayName: 'nl-native-1', pingMs: 42),
      );
      expect(VpnChannel.lastStatus!.event, VpnEvent.connected);

      VpnChannel.debugSetLastStatus(
        VpnStatusUpdate(VpnEvent.disconnected),
      );
      expect(VpnChannel.lastStatus!.event, VpnEvent.disconnected);
      expect(VpnChannel.lastStatus!.relayName, isNull);
    });

    test('disconnect after error resets to error then disconnected', () {
      VpnChannel.debugSetLastStatus(
        VpnStatusUpdate(VpnEvent.error, errorMessage: 'hysteria connect: timeout'),
      );
      expect(VpnChannel.lastStatus!.event, VpnEvent.error);

      VpnChannel.debugSetLastStatus(
        VpnStatusUpdate(VpnEvent.disconnected),
      );
      expect(VpnChannel.lastStatus!.event, VpnEvent.disconnected);
    });

    test('permissionDenied is surfaced as distinct event', () {
      VpnChannel.debugSetLastStatus(
        VpnStatusUpdate(VpnEvent.permissionDenied),
      );
      expect(VpnChannel.lastStatus!.event, VpnEvent.permissionDenied);
    });
  });

  group('Disconnect logging contract', () {
    test('disconnect request is logged for diagnostics export', () async {
      final log = ConnectionLog.instance;
      log.info('vpn', 'disconnect requested');
      log.info('vpn', 'native VPN event', {'event': 'disconnected'});

      final text = log.exportText();
      expect(text, contains('disconnect requested'));
      expect(text, contains('event=disconnected'));
    });

    test('connect session ends cleanly after disconnect sequence', () {
      final log = ConnectionLog.instance;
      log.beginConnectSession(relayId: 'nl-native-1', host: '212.43.156.33');
      log.info('vpn', 'native VPN event', {'event': 'connecting'});
      log.info('vpn', 'native VPN event', {'event': 'connected', 'pingMs': '55'});
      log.info('vpn', 'disconnect requested');
      log.info('vpn', 'native VPN event', {'event': 'disconnected'});

      final events = log.entries.where((e) => e.tag == 'vpn').toList();
      expect(events.any((e) => e.details?['event'] == 'connecting'), isTrue);
      expect(events.any((e) => e.details?['event'] == 'connected'), isTrue);
      expect(events.any((e) => e.message.contains('disconnect requested')), isTrue);
      expect(events.any((e) => e.details?['event'] == 'disconnected'), isTrue);
    });
  });

  group('Why app must stay open on disconnect', () {
    test('disconnected is not treated as fatal error event', () {
      final update = VpnStatusUpdate(VpnEvent.disconnected);
      expect(update.errorMessage, isNull);
      expect(update.event, isNot(VpnEvent.error));
    });
  });
}
