import 'dart:io' show Platform;

import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'theme/app_theme.dart';
import 'services/api_timeouts.dart';
import 'services/auth_service.dart';
import 'services/settings_service.dart';
import 'services/streampass_api.dart';
import 'services/vpn_channel.dart';
import 'services/connection_controller.dart';
import 'services/windows_traffic_engine.dart';
import 'screens/onboarding_screen.dart';
import 'screens/main_shell.dart';

const _apiBaseUrl = String.fromEnvironment(
  'STREAMPASS_API_URL',
  defaultValue: 'https://212-43-156-33.nip.io/api/v1',
);

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  // Avoid blocking/crashing on first frame when fonts cannot be downloaded.
  GoogleFonts.config.allowRuntimeFetching = false;
  if (!kIsWeb && Platform.isWindows) {
    VpnChannel.windowsImpl = WindowsTrafficEngine.instance;
  }
  // Subscribe EventChannel / Windows engine before any screen listens.
  VpnChannel.ensureListening();
  ConnectionController.instance.attach();

  final authService = AuthService(baseUrl: _apiBaseUrl);
  final api = StreamPassApi(baseUrl: _apiBaseUrl, authService: authService);
  runApp(StreamPassApp(authService: authService, api: api));
}

class StreamPassApp extends StatefulWidget {
  final AuthService authService;
  final StreamPassApi api;
  const StreamPassApp({
    super.key,
    required this.authService,
    required this.api,
  });

  @override
  State<StreamPassApp> createState() => _StreamPassAppState();

  static _StreamPassAppState? _of(BuildContext context) =>
      context.findAncestorStateOfType<_StreamPassAppState>();
}

/// Updates [MaterialApp.themeMode] from Settings (BL-052). No MaterialApp.builder overlays.
void applyAppAppearance(BuildContext context, {required ThemeMode themeMode}) {
  StreamPassApp._of(context)?.setThemeMode(themeMode);
}

class _StreamPassAppState extends State<StreamPassApp> {
  ThemeMode _themeMode = ThemeMode.dark;

  @override
  void initState() {
    super.initState();
    _loadTheme();
  }

  Future<void> _loadTheme() async {
    final s = await SettingsService().load();
    if (!mounted) return;
    setState(() => _themeMode = themeModeFromSetting(s.themeMode));
  }

  void setThemeMode(ThemeMode mode) {
    if (_themeMode == mode) return;
    setState(() => _themeMode = mode);
  }

  @override
  Widget build(BuildContext context) {
    // Intentionally no MaterialApp.builder Stack overlays (caused grey AppBar bugs).
    return MaterialApp(
      title: 'StreamPass',
      debugShowCheckedModeBanner: false,
      theme: buildLightAppTheme(),
      darkTheme: buildAppTheme(),
      themeMode: _themeMode,
      home: FutureBuilder<bool>(
        // Issue #4: never block splash forever on hung /refresh.
        future: _resolveLoggedIn(widget.authService),
        builder: (context, snapshot) {
          if (snapshot.hasError) {
            return OnboardingScreen(authService: widget.authService, api: widget.api);
          }
          if (!snapshot.hasData) {
            return const Scaffold(
              body: Center(child: CircularProgressIndicator()),
            );
          }
          return snapshot.data!
              ? MainShell(api: widget.api, authService: widget.authService)
              : OnboardingScreen(authService: widget.authService, api: widget.api);
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

/// Bound auth gate: local session wins over hung refresh (Issue #4).
Future<bool> _resolveLoggedIn(AuthService auth) async {
  try {
    return await auth.isLoggedIn.timeout(ApiTimeouts.authGate);
  } catch (_) {
    try {
      return await auth.hasSession;
    } catch (_) {
      return false;
    }
  }
}
