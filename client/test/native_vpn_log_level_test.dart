import 'package:flutter_test/flutter_test.dart';
import 'package:streampass/services/connection_log.dart';
import 'package:streampass/services/native_vpn_log_level.dart';

void main() {
  test('empty error= is not ERROR', () {
    const line =
        '[diag] proto=dns site=https://x host=x result=dns_resolved error=';
    expect(classifyNativeVpnLog(line), ConnectionLogLevel.info);
  });

  test('result=ok with empty error is not ERROR', () {
    const line =
        '[diag] proto=tcp result=ok speed_kbps=0 reason=default_relay_foreign error=';
    expect(classifyNativeVpnLog(line), ConnectionLogLevel.info);
  });

  test('non-empty error= is ERROR', () {
    const line =
        '[diag] proto=tcp result=timeout error=dial_error:_i/o_timeout';
    expect(classifyNativeVpnLog(line), ConnectionLogLevel.error);
  });

  test('stream_open_no_data is ERROR', () {
    const line =
        '[tun] stream_open_no_data (4s) dest=1.2.3.4:443 mode=RELAY bytes_tx=0 bytes_rx=0';
    expect(classifyNativeVpnLog(line), ConnectionLogLevel.error);
  });

  test('transfer_done is INFO', () {
    const line =
        '[diag] result=transfer_done speed_kbps=350 bytes_tx=1000 bytes_rx=2000 error=';
    expect(classifyNativeVpnLog(line), ConnectionLogLevel.info);
  });

  test('conn success with empty error is INFO', () {
    const line =
        '[conn] host=x result=success error=';
    expect(classifyNativeVpnLog(line), ConnectionLogLevel.info);
  });
}
