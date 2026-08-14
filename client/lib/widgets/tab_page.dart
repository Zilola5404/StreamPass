import 'package:flutter/material.dart';

import '../layout/adaptive.dart';
import '../theme/app_theme.dart';

/// Tab body without a nested [Scaffold]/[AppBar] — those fill the screen grey
/// on Samsung One UI when placed inside the shell Scaffold.
///
/// Must wrap content in [Material]: ListTile/SwitchListTile require it, and
/// IndexedStack on desktop can sit behind a LookupBoundary from the shell.
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
    final wide = isWideLayout(context);
    return Material(
      color: AppColors.bg,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          SafeArea(
            bottom: false,
            child: SizedBox(
              height: wide ? 64 : 56,
              child: Padding(
                padding: EdgeInsets.fromLTRB(wide ? 24 : 12, 0, 8, 0),
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
          Expanded(
            child: LayoutBuilder(
              builder: (context, constraints) {
                final maxW = constraints.maxWidth.isFinite
                    ? constraints.maxWidth
                    : MediaQuery.sizeOf(context).width;
                final maxH = constraints.maxHeight.isFinite
                    ? constraints.maxHeight
                    : MediaQuery.sizeOf(context).height;
                final width = wide ? maxW.clamp(0, 840).toDouble() : maxW;
                return Align(
                  alignment: Alignment.topCenter,
                  child: SizedBox(
                    width: width,
                    height: maxH,
                    child: body,
                  ),
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}
