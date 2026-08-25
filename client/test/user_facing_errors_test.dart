import 'dart:async';
import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;

import 'package:streampass/services/streampass_api.dart';
import 'package:streampass/services/user_facing_errors.dart';

void main() {
  group('UserFacingErrors', () {
    test('maps network failures to short Russian strings', () {
      expect(
        UserFacingErrors.map(const SocketException('Failed host lookup')),
        UserFacingErrors.noInternet,
      );
      expect(
        UserFacingErrors.map(TimeoutException('x')),
        UserFacingErrors.serverUnavailable,
      );
      expect(
        UserFacingErrors.map(http.ClientException('Connection refused')),
        UserFacingErrors.serverUnavailable,
      );
      expect(
        UserFacingErrors.map(ApiException(503, 'unavailable')),
        UserFacingErrors.serverUnavailable,
      );
      expect(
        UserFacingErrors.mapMessage('relay_unavailable'),
        UserFacingErrors.serverUnavailable,
      );
      expect(
        UserFacingErrors.mapMessage('Нет доступных серверов. Попробуйте позже'),
        UserFacingErrors.serverUnavailable,
      );
    });

    test('hides technical strings behind connectFailed', () {
      expect(
        UserFacingErrors.mapMessage(
          'Exception: hysteria QUIC dial timed out errno=110',
        ),
        UserFacingErrors.connectFailed,
      );
      expect(
        UserFacingErrors.mapMessage('traffic_ready still pending'),
        UserFacingErrors.connectFailed,
      );
    });

    test('CTA constants match Issue #4', () {
      expect(UserFacingErrors.connectCta, 'Подключить');
      expect(UserFacingErrors.connecting, 'Подключение…');
      expect(UserFacingErrors.connected, 'Подключено');
      expect(UserFacingErrors.disconnecting, 'Отключение…');
      expect(UserFacingErrors.serverUnavailable, 'Сервер недоступен');
      expect(UserFacingErrors.connectFailed, 'Не удалось подключиться');
    });
  });
}
