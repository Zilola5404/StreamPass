import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:streampass/main.dart';
import 'package:streampass/screens/home_screen.dart';
import 'package:streampass/screens/onboarding_screen.dart';
import 'package:streampass/screens/servers_screen.dart';
import 'package:streampass/services/auth_service.dart';
import 'package:streampass/services/streampass_api.dart';
import 'package:streampass/services/token_storage.dart';

import 'support/mock_backend.dart';

/// Home has ambient animations — avoid pumpAndSettle (never settles).
Future<void> pumpFrames(WidgetTester tester, {int frames = 20}) async {
  for (var i = 0; i < frames; i++) {
    await tester.pump(const Duration(milliseconds: 50));
  }
}

/// BL-031: end-to-end style Flutter tests against an in-memory mock API.
void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  late MockStreamPassBackend backend;
  late AuthService auth;
  late StreamPassApi api;

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    await SharedPreferences.getInstance();
    backend = MockStreamPassBackend();
    auth = AuthService(
      baseUrl: 'https://mock.test/api/v1',
      client: backend.client,
      tokenStorage: TokenStorage.inMemory(),
    );
    api = StreamPassApi(baseUrl: 'https://mock.test/api/v1', authService: auth, client: backend.client);
  });

  testWidgets('E2E: login navigates to Home with Amsterdam relay', (tester) async {
    await tester.pumpWidget(StreamPassApp(authService: auth, api: api));
    await pumpFrames(tester);

    expect(find.byType(OnboardingScreen), findsOneWidget);

    await tester.enterText(find.byType(TextField).at(0), 'user@example.com');
    await tester.enterText(find.byType(TextField).at(1), 'secret123');
    await tester.tap(find.text('Войти'));
    await pumpFrames(tester, frames: 60);

    expect(backend.loginCalls, 1);
    expect(find.byType(HomeScreen), findsOneWidget);
    expect(find.textContaining('Amsterdam'), findsWidgets);
  });

  testWidgets('E2E: bad password shows error and stays on onboarding', (tester) async {
    await tester.pumpWidget(StreamPassApp(authService: auth, api: api));
    await pumpFrames(tester);

    await tester.enterText(find.byType(TextField).at(0), 'user@example.com');
    await tester.enterText(find.byType(TextField).at(1), 'wrong');
    await tester.tap(find.text('Войти'));
    await pumpFrames(tester);

    expect(find.byType(OnboardingScreen), findsOneWidget);
    expect(find.byType(HomeScreen), findsNothing);
    expect(find.textContaining('invalid'), findsOneWidget);
  });

  testWidgets('E2E: logged-in session opens Home and Servers picker', (tester) async {
    auth = AuthService(
      baseUrl: 'https://mock.test/api/v1',
      client: backend.client,
      tokenStorage: TokenStorage.inMemory({
        'sp_access_token': 'access-mock',
        'sp_refresh_token': 'refresh-mock',
        'sp_access_expires_at': '2099-01-01T00:00:00Z',
      }),
    );
    api = StreamPassApi(baseUrl: 'https://mock.test/api/v1', authService: auth, client: backend.client);

    await tester.pumpWidget(StreamPassApp(authService: auth, api: api));
    await pumpFrames(tester, frames: 60);

    expect(find.byType(HomeScreen), findsOneWidget);
    expect(backend.serversCalls, greaterThan(0));

    await tester.tap(find.textContaining('Amsterdam').first);
    await pumpFrames(tester, frames: 30);

    expect(find.byType(ServersScreen), findsOneWidget);
    expect(find.text('Автовыбор'), findsOneWidget);
    expect(find.textContaining('Warsaw'), findsWidgets);
    expect(find.text('nl-native-1'), findsOneWidget);
  });

  testWidgets('E2E: register mode creates account then Home', (tester) async {
    await tester.pumpWidget(StreamPassApp(authService: auth, api: api));
    await pumpFrames(tester);

    await tester.tap(find.textContaining('Зарегистрироваться'));
    await pumpFrames(tester);

    await tester.enterText(find.byType(TextField).at(0), 'new@example.com');
    await tester.enterText(find.byType(TextField).at(1), 'password');
    await tester.tap(find.text('Создать аккаунт'));
    await pumpFrames(tester, frames: 60);

    expect(backend.registerCalls, 1);
    expect(find.byType(HomeScreen), findsOneWidget);
  });

  test('E2E API: fetchServers?region=pl filters mock catalog', () async {
    auth = AuthService(
      baseUrl: 'https://mock.test/api/v1',
      client: backend.client,
      tokenStorage: TokenStorage.inMemory({
        'sp_access_token': 'access-mock',
        'sp_refresh_token': 'refresh-mock',
        'sp_access_expires_at': '2099-01-01T00:00:00Z',
      }),
    );
    api = StreamPassApi(baseUrl: 'https://mock.test/api/v1', authService: auth, client: backend.client);

    final list = await api.fetchServers(region: 'pl');
    expect(list, hasLength(1));
    expect(list.first.id, 'pl-warsaw-1');
    expect(list.first.displayRegion, contains('Warsaw'));

    final regions = await api.fetchRegions();
    expect(regions.map((r) => r.code), containsAll(['de', 'nl', 'pl', 'fi']));
  });
}
