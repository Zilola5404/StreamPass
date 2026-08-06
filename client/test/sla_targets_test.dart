import 'package:flutter_test/flutter_test.dart';
import 'package:streampass/services/sla_targets.dart';

void main() {
  group('SlaTargets', () {
    test('cold start ≤ 2s', () {
      expect(SlaTargets.meetsColdStart(const Duration(milliseconds: 1999)), isTrue);
      expect(SlaTargets.meetsColdStart(const Duration(seconds: 2)), isTrue);
      expect(SlaTargets.meetsColdStart(const Duration(milliseconds: 2001)), isFalse);
    });

    test('connect ≤ 5s', () {
      expect(SlaTargets.meetsConnect(const Duration(seconds: 5)), isTrue);
      expect(SlaTargets.meetsConnect(const Duration(milliseconds: 5001)), isFalse);
    });

    test('recover ≤ 10s', () {
      expect(SlaTargets.meetsRecover(const Duration(seconds: 10)), isTrue);
      expect(SlaTargets.meetsRecover(const Duration(seconds: 11)), isFalse);
    });
  });
}
