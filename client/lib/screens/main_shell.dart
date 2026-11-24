import 'package:flutter/material.dart';

import '../services/auth_service.dart';
import '../services/streampass_api.dart';
import '../theme/app_theme.dart';
import '../widgets/app_bottom_nav.dart';
import '../layout/adaptive.dart';
import 'home_screen.dart';
import 'servers_screen.dart';
import 'settings_screen.dart';
import 'statistics_screen.dart';

/// Root authenticated shell with persistent bottom navigation (all tabs).
class MainShell extends StatefulWidget {
  final StreamPassApi api;
  final AuthService authService;

  const MainShell({
    super.key,
    required this.api,
    required this.authService,
  });

  @override
  State<MainShell> createState() => _MainShellState();
}

class _MainShellState extends State<MainShell> {
  int _index = 0;
  final _homeKey = GlobalKey<HomeScreenState>();
  final _statsKey = GlobalKey<StatisticsScreenState>();

  void _goToTab(int index) {
    if (index == _index) return;
    setState(() => _index = index);
    _afterTab(index);
  }

  void _afterTab(int index) {
    if (index == 0) {
      _homeKey.currentState?.refreshRelayDisplay();
    } else if (index == 1) {
      _statsKey.currentState?.refresh();
    }
  }

  @override
  Widget build(BuildContext context) {
    final wide = isWideLayout(context);
    final stack = IndexedStack(
      index: _index,
      sizing: StackFit.expand,
      children: [
        HomeScreen(
          key: _homeKey,
          api: widget.api,
          authService: widget.authService,
          onNavigateTab: _goToTab,
        ),
        StatisticsScreen(key: _statsKey),
        ServersScreen(
          api: widget.api,
          authService: widget.authService,
          mode: ServersScreenMode.tab,
          onSelectionChanged: () =>
              _homeKey.currentState?.refreshRelayDisplay(),
        ),
        SettingsScreen(
          api: widget.api,
          authService: widget.authService,
        ),
      ],
    );

    return Scaffold(
      backgroundColor: AppColors.bg,
      body: wide
          ? Row(
              children: [
                AppSideRail(currentIndex: _index, onTap: _goToTab),
                VerticalDivider(
                  width: 1,
                  thickness: 1,
                  color: Colors.white.withOpacity(0.06),
                ),
                Expanded(child: stack),
              ],
            )
          : stack,
      bottomNavigationBar: wide
          ? null
          : AppBottomNav(
              currentIndex: _index,
              onTap: _goToTab,
            ),
    );
  }
}
