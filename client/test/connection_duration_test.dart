import 'package:flutter_test/flutter_test.dart';
import 'package:streampass/services/connection_duration.dart';

void main() {
  test('formatConnectionDuration zero', () {
    expect(formatConnectionDuration(Duration.zero), '00:00:00');
  });

  test('formatConnectionDuration hours minutes seconds', () {
    expect(
      formatConnectionDuration(const Duration(hours: 1, minutes: 2, seconds: 3)),
      '01:02:03',
    );
  });

  test('formatConnectionDuration rolls minutes and hours', () {
    expect(
      formatConnectionDuration(const Duration(minutes: 3, seconds: 47)),
      '00:03:47',
    );
  });
}
