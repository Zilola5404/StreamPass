import 'package:flutter_test/flutter_test.dart';

import 'package:streampass/services/diag_uploader.dart';

void main() {
  test('parseLine extracts site/via/reason from enriched [diag]', () {
    const line =
        '[diag] proto=tcp site=https://ya.ru host=ya.ru dest_ip=77.88.8.8 dest_port=443 mode=DIRECT via=DIRECT result=ok latency_ms=42 slow=0 reason=ok_direct error=';
    final e = DiagUploader.parseLine(line)!;
    expect(e.proto, 'tcp');
    expect(e.site, 'https://ya.ru');
    expect(e.host, 'ya.ru');
    expect(e.destIp, '77.88.8.8');
    expect(e.destPort, 443);
    expect(e.mode, 'DIRECT');
    expect(e.result, 'ok');
    expect(e.latencyMs, 42);
    expect(e.slow, isFalse);
    expect(e.reason, 'ok_direct');
  });

  test('parseLine marks slow dials', () {
    const line =
        '[diag] proto=tcp site=https://youtube.com host=youtube.com dest_ip=142.1.2.3 dest_port=443 mode=RELAY via=RELAY result=slow latency_ms=2100 slow=1 reason=slow_dial_relay error=';
    final e = DiagUploader.parseLine(line)!;
    expect(e.result, 'slow');
    expect(e.slow, isTrue);
    expect(e.reason, 'slow_dial_relay');
    expect(e.site, 'https://youtube.com');
  });

  test('parseLine maps DNS query lines', () {
    const line = '[dns] query google.com via=doh rtt=55ms';
    final e = DiagUploader.parseLine(line)!;
    expect(e.proto, 'dns');
    expect(e.host, 'google.com');
    expect(e.site, 'https://google.com');
    expect(e.mode, 'DOH');
    expect(e.result, 'ok');
    expect(e.latencyMs, 55);
  });

  test('parseLine ignores unrelated log lines', () {
    expect(DiagUploader.parseLine('[tun] tcp mode=DIRECT dest=1.2.3.4:443'), isNull);
  });
}
