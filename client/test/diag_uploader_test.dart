import 'package:flutter_test/flutter_test.dart';

import 'package:streampass/services/diag_uploader.dart';

void main() {
  test('parseLine extracts structured [diag] fields', () {
    const line =
        '[diag] proto=tcp host=ya.ru dest_ip=77.88.8.8 dest_port=443 mode=DIRECT result=ok latency_ms=42 error=';
    final e = DiagUploader.parseLine(line)!;
    expect(e.proto, 'tcp');
    expect(e.host, 'ya.ru');
    expect(e.destIp, '77.88.8.8');
    expect(e.destPort, 443);
    expect(e.mode, 'DIRECT');
    expect(e.result, 'ok');
    expect(e.latencyMs, 42);
  });

  test('parseLine maps DNS query lines', () {
    const line = '[dns] query google.com via=doh rtt=55ms';
    final e = DiagUploader.parseLine(line)!;
    expect(e.proto, 'dns');
    expect(e.host, 'google.com');
    expect(e.mode, 'DOH');
    expect(e.result, 'ok');
    expect(e.latencyMs, 55);
  });

  test('parseLine ignores unrelated log lines', () {
    expect(DiagUploader.parseLine('[tun] tcp mode=DIRECT dest=1.2.3.4:443'), isNull);
  });
}
