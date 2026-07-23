import 'package:flutter/material.dart';
import 'theme/app_theme.dart';
import 'services/auth_service.dart';
import 'screens/onboarding_screen.dart';
import 'screens/home_screen.dart';

const _apiBaseUrl = String.fromEnvironment(
  'STREAMPASS_API_URL',
  defaultValue: 'https://api.streampass.com/api/v1',
);

void main() {
  runApp(StreamPassApp(authService: AuthService(baseUrl: _apiBaseUrl)));
}

class StreamPassApp extends StatelessWidget {
  final AuthService authService;
  const StreamPassApp({super.key, required this.authService});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'StreamPass',
      debugShowCheckedModeBanner: false,
      theme: buildAppTheme(),
      home: FutureBuilder<bool>(
        future: authService.isLoggedIn,
        builder: (context, snapshot) {
          if (!snapshot.hasData) {
            return const Scaffold(
              body: Center(child: CircularProgressIndicator()),
            );
          }
          return snapshot.data!
              ? const HomeScreen()
              : OnboardingScreen(authService: authService);
        },
      ),
    );
  }
}
