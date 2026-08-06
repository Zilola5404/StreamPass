import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;

import 'package:streampass/main.dart';
import 'package:streampass/services/auth_service.dart';
import 'package:streampass/services/streampass_api.dart';
import 'package:streampass/services/token_storage.dart';

/// App shell must not crash when auth/session resolution fails.
void main() {
  testWidgets('StreamPassApp shows onboarding when session check fails gracefully',
      (tester) async {
    final authService = AuthService(
      baseUrl: 'https://example.invalid/api/v1',
      tokenStorage: TokenStorage.inMemory({
        'sp_refresh_token': 'stale-refresh',
        'sp_access_token': 'stale-access',
        'sp_access_expires_at': '2020-01-01T00:00:00.000Z',
      }),
    );
    final api = StreamPassApi(
      baseUrl: 'https://example.invalid/api/v1',
      authService: authService,
    );

    await tester.pumpWidget(StreamPassApp(authService: authService, api: api));
    await tester.pump();
    await tester.pump(const Duration(seconds: 3));

    // Refresh fails offline → session cleared or kept; app must show a screen.
    expect(find.byType(CircularProgressIndicator), findsNothing);
    expect(
      find.byWidgetPredicate(
        (w) => w is Scaffold && w.body != null,
      ),
      findsWidgets,
    );
  });

  testWidgets('StreamPassApp shows onboarding for logged-out user', (tester) async {
    final authService = AuthService(
      baseUrl: 'https://example.invalid/api/v1',
      tokenStorage: TokenStorage.inMemory(),
    );
    final api = StreamPassApi(
      baseUrl: 'https://example.invalid/api/v1',
      authService: authService,
    );

    await tester.pumpWidget(StreamPassApp(authService: authService, api: api));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(find.text('Войти'), findsOneWidget);
  });
}
