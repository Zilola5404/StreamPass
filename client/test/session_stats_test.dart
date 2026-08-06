import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:streampass/services/session_stats.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await SharedPreferences.getInstance();
  });

  test('SessionStats accumulates online time and reconnects', () async {
    final svc = SessionStatsService();
    await svc.addOnlineSeconds(120);
    await svc.recordReconnect();
    await svc.recordReconnect();
    await svc.recordRtt(40);
    await svc.recordRtt(60);

    final snap = await svc.load();
    expect(snap.onlineToday.inSeconds, 120);
    expect(snap.reconnectsToday, 2);
    expect(snap.averageRttMs, 50);
    expect(snap.hasData, isTrue);
  });

  test('empty stats hasData is false', () async {
    final snap = await SessionStatsService().load();
    expect(snap.hasData, isFalse);
  });
}
