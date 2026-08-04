import 'package:flutter_test/flutter_test.dart';
import 'package:streampass/services/client_update.dart';

void main() {
  group('compareClientVersions', () {
    test('orders dotted versions', () {
      expect(compareClientVersions('0.1.0', '0.1.1'), lessThan(0));
      expect(compareClientVersions('0.1.1', '0.1.1'), 0);
      expect(compareClientVersions('0.2.0', '0.1.9'), greaterThan(0));
    });

    test('ignores build metadata', () {
      expect(compareClientVersions('0.1.1+15', '0.1.1'), 0);
      expect(compareClientVersions('0.1.1+15', '0.1.2'), lessThan(0));
    });
  });

  group('evaluateClientUpdate', () {
    test('requires update below min supported', () {
      final r = evaluateClientUpdate(
        currentVersion: '0.1.0',
        minSupportedVersion: '0.1.1',
        latestVersion: '0.1.2',
        downloadUrl: 'https://example.com/a.apk',
      );
      expect(r.urgency, UpdateUrgency.required);
    });

    test('optional update when newer latest available', () {
      final r = evaluateClientUpdate(
        currentVersion: '0.1.1',
        minSupportedVersion: '0.1.0',
        latestVersion: '0.1.2',
        downloadUrl: 'https://example.com/a.apk',
      );
      expect(r.urgency, UpdateUrgency.optional);
      expect(r.downloadUrl, 'https://example.com/a.apk');
    });

    test('none when current is latest', () {
      final r = evaluateClientUpdate(
        currentVersion: '0.1.2',
        minSupportedVersion: '0.1.0',
        latestVersion: '0.1.2',
        downloadUrl: 'https://example.com/a.apk',
      );
      expect(r.urgency, UpdateUrgency.none);
    });

    test('no optional without download url', () {
      final r = evaluateClientUpdate(
        currentVersion: '0.1.1',
        minSupportedVersion: '0.1.0',
        latestVersion: '0.1.2',
        downloadUrl: '',
      );
      expect(r.urgency, UpdateUrgency.none);
    });
  });
}
