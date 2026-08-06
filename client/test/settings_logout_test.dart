import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:streampass/screens/settings_screen.dart';
import 'package:streampass/services/auth_service.dart';
import 'package:streampass/services/streampass_api.dart';
import 'package:streampass/services/token_storage.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('Settings shows logout when auth is provided', (tester) async {
    SharedPreferences.setMockInitialValues({});
    await SharedPreferences.getInstance();

    final auth = AuthService(
      baseUrl: 'https://example.invalid/api/v1',
      tokenStorage: TokenStorage.inMemory({
        'sp_refresh_token': 'r',
        'sp_access_token': 'a',
        'sp_access_expires_at': '2099-01-01T00:00:00Z',
      }),
    );
    final api = StreamPassApi(baseUrl: 'https://example.invalid/api/v1', authService: auth);

    await tester.pumpWidget(
      MaterialApp(
        home: SettingsScreen(api: api, authService: auth),
      ),
    );
    for (var i = 0; i < 30; i++) {
      await tester.pump(const Duration(milliseconds: 50));
      if (find.text('Автозапуск').evaluate().isNotEmpty) break;
    }

    expect(find.text('Автозапуск'), findsOneWidget);
    await tester.drag(find.byType(ListView), const Offset(0, -400));
    await tester.pump();
    expect(find.text('Выйти'), findsOneWidget);
  });

  test('AuthService.logout clears local session', () async {
    final auth = AuthService(
      baseUrl: 'https://example.invalid/api/v1',
      tokenStorage: TokenStorage.inMemory({
        'sp_refresh_token': 'r',
        'sp_access_token': 'a',
        'sp_access_expires_at': '2099-01-01T00:00:00Z',
      }),
    );

    expect(await auth.hasSession, isTrue);
    await auth.logout();
    expect(await auth.hasSession, isFalse);
    expect(await auth.storedToken, isNull);
  });
}
