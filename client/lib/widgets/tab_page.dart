import 'package:flutter/material.dart';

import '../theme/app_theme.dart';

/// Tab body without a nested [Scaffold]/[AppBar] — those fill the screen grey
/// on Samsung One UI when placed inside the shell Scaffold.
class TabPage extends StatelessWidget {
  final String title;
  final Widget body;
  final List<Widget>? actions;

  const TabPage({
    super.key,
    required this.title,
    required this.body,
    this.actions,
  });

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      color: AppColors.bg,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          SafeArea(
            bottom: false,
            child: SizedBox(
              height: 56,
              child: Padding(
                padding: const EdgeInsets.fromLTRB(12, 0, 4, 0),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        title,
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                    ),
                    if (actions != null) ...actions!,
                  ],
                ),
              ),
            ),
          ),
          Expanded(child: body),
        ],
      ),
    );
  }
}
