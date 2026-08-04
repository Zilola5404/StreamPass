import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'theme/app_theme.dart';
import 'services/auth_service.dart';
import 'services/streampass_api.dart';
import 'services/vpn_channel.dart';
import 'screens/onboarding_screen.dart';
import 'screens/home_screen.dart';

const _apiBaseUrl = String.fromEnvironment(
  'STREAMPASS_API_URL',
  defaultValue: 'https://212-43-156-33.nip.io/api/v1',
);

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  // Avoid blocking/crashing on first frame when fonts cannot be downloaded.
  GoogleFonts.config.allowRuntimeFetching = false;
  // Subscribe EventChannel early so native VPN events are not dropped.
  VpnChannel.ensureListening();

  final authService = AuthService(baseUrl: _apiBaseUrl);
  final api = StreamPassApi(baseUrl: _apiBaseUrl, authService: authService);
  runApp(StreamPassApp(authService: authService, api: api));
}

class StreamPassApp extends StatelessWidget {
  final AuthService authService;
  final StreamPassApi api;
  const StreamPassApp({
    super.key,
    required this.authService,
    required this.api,
  });

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'StreamPass',
      debugShowCheckedModeBanner: false,
      theme: buildAppTheme(),
      home: FutureBuilder<bool>(
        future: authService.isLoggedIn.catchError((_) => false),
        builder: (context, snapshot) {
          if (snapshot.hasError) {
            return OnboardingScreen(authService: authService, api: api);
          }
          if (!snapshot.hasData) {
            return const Scaffold(
              body: Center(child: CircularProgressIndicator()),
            );
          }
          return snapshot.data!
              ? HomeScreen(api: api, authService: authService)
              : OnboardingScreen(authService: authService, api: api);
        },
      ),
    );
  }
}

/// Sends the user back to login when refresh fails.
void navigateToLogin(BuildContext context, AuthService authService, StreamPassApi api) {
  Navigator.of(context).pushAndRemoveUntil(
    MaterialPageRoute(builder: (_) => OnboardingScreen(authService: authService, api: api)),
    (_) => false,
  );
}
