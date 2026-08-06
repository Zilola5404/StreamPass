import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

import 'package:streampass/screens/home_screen.dart';
import 'package:streampass/services/auth_service.dart';
import 'package:streampass/services/streampass_api.dart';
import 'package:streampass/services/token_storage.dart';

/// HomeScreen must stay alive when the backend is unreachable on startup.
/// A crash here matches the user report "app closes on launch".
void main() {
  testWidgets('HomeScreen survives fetchServers network failure', (tester) async {
    final client = MockClient((request) async {
      if (request.url.path.endsWith('/subscription')) {
        return http.Response(
          jsonEncode({'status': 'ACTIVE', 'active_until': '2099-01-01T00:00:00Z'}),
          200,
        );
      }
      if (request.url.path.endsWith('/servers')) {
        return http.Response(
          'ClientException: Failed host lookup',
          503,
        );
      }
      return http.Response('not found', 404);
    });

    final authService = AuthService(
      baseUrl: 'https://example.invalid/api/v1',
      client: client,
      tokenStorage: TokenStorage.inMemory({
        'sp_access_token': 'test-access',
        'sp_refresh_token': 'test-refresh',
        'sp_access_expires_at':
            DateTime.now().toUtc().add(const Duration(hours: 1)).toIso8601String(),
      }),
    );
    final api = StreamPassApi(
      baseUrl: 'https://example.invalid/api/v1',
      authService: authService,
      client: client,
    );

    await tester.pumpWidget(
      MaterialApp(
        home: HomeScreen(api: api, authService: authService),
      ),
    );

    await tester.pump();
    await tester.pump(const Duration(seconds: 2));

    expect(find.text('StreamPass'), findsOneWidget);
    expect(find.byType(CircularProgressIndicator), findsNothing);
  });

  testWidgets('HomeScreen shows relay error instead of crashing when servers empty',
      (tester) async {
    final client = MockClient((request) async {
      if (request.url.path.endsWith('/subscription')) {
        return http.Response(
          jsonEncode({'status': 'ACTIVE', 'active_until': '2099-01-01T00:00:00Z'}),
          200,
        );
      }
      if (request.url.path.endsWith('/servers')) {
        return http.Response(jsonEncode([]), 200);
      }
      return http.Response('not found', 404);
    });

    final authService = AuthService(
      baseUrl: 'https://example.invalid/api/v1',
      client: client,
      tokenStorage: TokenStorage.inMemory({
        'sp_access_token': 'test-access',
        'sp_refresh_token': 'test-refresh',
        'sp_access_expires_at':
            DateTime.now().toUtc().add(const Duration(hours: 1)).toIso8601String(),
      }),
    );
    final api = StreamPassApi(
      baseUrl: 'https://example.invalid/api/v1',
      authService: authService,
      client: client,
    );

    await tester.pumpWidget(
      MaterialApp(
        home: HomeScreen(api: api, authService: authService),
      ),
    );

    await tester.pump();
    await tester.pump(const Duration(seconds: 2));

    expect(find.text('StreamPass'), findsOneWidget);
  });
}
