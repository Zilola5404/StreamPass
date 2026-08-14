import 'package:flutter/material.dart';

import '../services/auth_service.dart';
import '../services/streampass_api.dart';
import '../widgets/app_bottom_nav.dart';
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
  late final PageController _pageController;
  final _homeKey = GlobalKey<HomeScreenState>();
  final _statsKey = GlobalKey<StatisticsScreenState>();

  @override
  void initState() {
    super.initState();
    _pageController = PageController(initialPage: 0);
  }

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  void _goToTab(int index) {
    if (index == _index) return;
    _pageController.animateToPage(
      index,
      duration: const Duration(milliseconds: 280),
      curve: Curves.easeOutCubic,
    );
  }

  void _onPageChanged(int index) {
    setState(() => _index = index);
    if (index == 0) {
      _homeKey.currentState?.refreshRelayDisplay();
    } else if (index == 1) {
      _statsKey.currentState?.refresh();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: PageView(
        controller: _pageController,
        onPageChanged: _onPageChanged,
        physics: const BouncingScrollPhysics(),
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
      ),
      bottomNavigationBar: AppBottomNav(
        currentIndex: _index,
        onTap: (i) {
          if (i == _index) return;
          _pageController.animateToPage(
            i,
            duration: const Duration(milliseconds: 280),
            curve: Curves.easeOutCubic,
          );
        },
      ),
    );
  }
}
