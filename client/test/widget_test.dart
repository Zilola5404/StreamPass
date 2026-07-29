// Minimal smoke test: verifies the app boots to the onboarding screen for
// a logged-out user, without needing a real backend connection.
//
// This replaces the original file, which was still the unmodified Flutter
// "counter app" template (referencing a MyApp class and a '+' counter
// button that don't exist anywhere in this project) — it had clearly
// never been adapted after `flutter create` and could not compile.
//
// AuthService.isLoggedIn only reads local SharedPreferences (no network
// call), so mocking SharedPreferences to empty is enough to test the
// logged-out path deterministically and offline. StreamPassApi is
// constructed with a placeholder base URL since nothing in this test
// triggers a network request.
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:streampass/main.dart';
import 'package:streampass/services/auth_service.dart';
import 'package:streampass/services/streampass_api.dart';

void main() {
  testWidgets('shows the onboarding screen when logged out', (tester) async {
    SharedPreferences.setMockInitialValues({});

    final authService = AuthService(baseUrl: 'https://example.invalid/api/v1');
    final api = StreamPassApi(baseUrl: 'https://example.invalid/api/v1', authService: authService);

    await tester.pumpWidget(StreamPassApp(authService: authService, api: api));
    await tester.pumpAndSettle();

    expect(find.text('Войти'), findsOneWidget);
  });
}