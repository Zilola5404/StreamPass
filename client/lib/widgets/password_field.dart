import 'package:flutter/material.dart';

import '../theme/app_theme.dart';

/// Password [TextField] with an animated show/hide toggle.
class PasswordField extends StatefulWidget {
  final TextEditingController controller;
  final String hintText;
  final TextInputAction? textInputAction;
  final ValueChanged<String>? onSubmitted;

  const PasswordField({
    super.key,
    required this.controller,
    required this.hintText,
    this.textInputAction,
    this.onSubmitted,
  });

  @override
  State<PasswordField> createState() => _PasswordFieldState();
}

class _PasswordFieldState extends State<PasswordField> {
  bool _obscure = true;

  void _toggle() => setState(() => _obscure = !_obscure);

  @override
  Widget build(BuildContext context) {
    return TextField(
      controller: widget.controller,
      obscureText: _obscure,
      textInputAction: widget.textInputAction,
      onSubmitted: widget.onSubmitted,
      style: const TextStyle(color: AppColors.textPrimary),
      decoration: InputDecoration(
        hintText: widget.hintText,
        suffixIcon: PasswordVisibilityToggle(
          obscured: _obscure,
          onToggle: _toggle,
        ),
      ),
    );
  }
}

/// Animated eye icon for password fields.
class PasswordVisibilityToggle extends StatelessWidget {
  final bool obscured;
  final VoidCallback onToggle;

  const PasswordVisibilityToggle({
    super.key,
    required this.obscured,
    required this.onToggle,
  });

  @override
  Widget build(BuildContext context) {
    return IconButton(
      tooltip: obscured ? 'Показать пароль' : 'Скрыть пароль',
      onPressed: onToggle,
      icon: AnimatedSwitcher(
        duration: const Duration(milliseconds: 220),
        switchInCurve: Curves.easeOutCubic,
        switchOutCurve: Curves.easeInCubic,
        transitionBuilder: (child, animation) {
          return FadeTransition(
            opacity: animation,
            child: ScaleTransition(
              scale: Tween<double>(begin: 0.82, end: 1.0).animate(animation),
              child: child,
            ),
          );
        },
        child: Icon(
          obscured ? Icons.visibility_outlined : Icons.visibility_off_outlined,
          key: ValueKey<bool>(obscured),
          color: AppColors.textSecondary,
        ),
      ),
    );
  }
}
