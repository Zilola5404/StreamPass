import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../services/auth_service.dart';
import '../theme/app_theme.dart';

/// E01b — восстановление пароля (BL-042).
class ForgotPasswordScreen extends StatefulWidget {
  final AuthService authService;
  const ForgotPasswordScreen({super.key, required this.authService});

  @override
  State<ForgotPasswordScreen> createState() => _ForgotPasswordScreenState();
}

class _ForgotPasswordScreenState extends State<ForgotPasswordScreen> {
  final _emailCtrl = TextEditingController();
  final _tokenCtrl = TextEditingController();
  final _passCtrl = TextEditingController();
  bool _sent = false;
  bool _loading = false;
  String? _message;
  String? _error;
  String? _debugToken;

  Future<void> _request() async {
    final email = _emailCtrl.text.trim();
    if (email.isEmpty || !email.contains('@')) {
      setState(() => _error = 'Введите корректный адрес электронной почты.');
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
      _message = null;
    });
    final result = await widget.authService.forgotPassword(email);
    if (!mounted) return;
    setState(() {
      _loading = false;
      if (!result.success) {
        _error = result.message;
        return;
      }
      _sent = true;
      _message = result.message;
      _debugToken = result.resetToken;
      if (result.resetToken != null && result.resetToken!.isNotEmpty) {
        _tokenCtrl.text = result.resetToken!;
      }
    });
  }

  Future<void> _reset() async {
    final token = _tokenCtrl.text.trim();
    final pass = _passCtrl.text;
    if (token.isEmpty) {
      setState(() => _error = 'Введите код восстановления.');
      return;
    }
    if (pass.length < 8) {
      setState(() => _error = 'Пароль должен содержать минимум 8 символов.');
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    final result = await widget.authService.resetPassword(token, pass);
    if (!mounted) return;
    setState(() => _loading = false);
    if (result.success) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Пароль обновлён. Войдите с новым паролем.')),
      );
      Navigator.of(context).pop();
    } else {
      setState(() => _error = result.error ?? 'Не удалось сбросить пароль');
    }
  }

  @override
  void dispose() {
    _emailCtrl.dispose();
    _tokenCtrl.dispose();
    _passCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Восстановление пароля'),
        backgroundColor: Colors.transparent,
        elevation: 0,
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 28),
          child: ListView(
            children: [
              const SizedBox(height: 12),
              Text(
                _sent
                    ? 'Введите код из письма и новый пароль.'
                    : 'Укажите email — мы отправим инструкции, если аккаунт существует.',
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const SizedBox(height: 24),
              if (!_sent) ...[
                TextField(
                  controller: _emailCtrl,
                  keyboardType: TextInputType.emailAddress,
                  style: const TextStyle(color: AppColors.textPrimary),
                  decoration: const InputDecoration(hintText: 'Электронная почта'),
                ),
                const SizedBox(height: 20),
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: _loading ? null : _request,
                    child: _loading
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: AppColors.bg,
                            ),
                          )
                        : const Text('Отправить'),
                  ),
                ),
              ] else ...[
                if (_message != null)
                  Text(_message!, style: Theme.of(context).textTheme.bodyMedium),
                if (_debugToken != null && _debugToken!.isNotEmpty) ...[
                  const SizedBox(height: 12),
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: AppColors.surface,
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'Код (временный режим без email):',
                          style: TextStyle(color: AppColors.textSecondary, fontSize: 12),
                        ),
                        const SizedBox(height: 6),
                        SelectableText(
                          _debugToken!,
                          style: const TextStyle(fontSize: 11, color: AppColors.textPrimary),
                        ),
                        TextButton(
                          onPressed: () {
                            Clipboard.setData(ClipboardData(text: _debugToken!));
                          },
                          child: const Text('Скопировать'),
                        ),
                      ],
                    ),
                  ),
                ],
                const SizedBox(height: 16),
                TextField(
                  controller: _tokenCtrl,
                  style: const TextStyle(color: AppColors.textPrimary),
                  decoration: const InputDecoration(hintText: 'Код восстановления'),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: _passCtrl,
                  obscureText: true,
                  style: const TextStyle(color: AppColors.textPrimary),
                  decoration: const InputDecoration(hintText: 'Новый пароль'),
                ),
                const SizedBox(height: 20),
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: _loading ? null : _reset,
                    child: _loading
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: AppColors.bg,
                            ),
                          )
                        : const Text('Сохранить пароль'),
                  ),
                ),
                TextButton(
                  onPressed: _loading
                      ? null
                      : () => setState(() {
                            _sent = false;
                            _debugToken = null;
                            _message = null;
                          }),
                  child: const Text('Отправить код снова'),
                ),
              ],
              if (_error != null) ...[
                const SizedBox(height: 16),
                Text(_error!, style: const TextStyle(color: AppColors.danger)),
              ],
            ],
          ),
        ),
      ),
    );
  }
}
