import 'package:flutter/material.dart';
import '../theme/app_theme.dart';
import '../services/auth_service.dart';
import '../services/streampass_api.dart';
import 'home_screen.dart';
import 'forgot_password_screen.dart';
import 'legal_document_screen.dart';

class OnboardingScreen extends StatefulWidget {
  final AuthService authService;
  final StreamPassApi api;
  const OnboardingScreen({
    super.key,
    required this.authService,
    required this.api,
  });

  @override
  State<OnboardingScreen> createState() => _OnboardingScreenState();
}

class _OnboardingScreenState extends State<OnboardingScreen> {
  final _emailCtrl = TextEditingController();
  final _passCtrl = TextEditingController();
  bool _isRegisterMode = false;
  bool _loading = false;
  String? _error;

  Future<void> _submit() async {
    setState(() {
      _loading = true;
      _error = null;
    });

    final result = _isRegisterMode
        ? await widget.authService.register(_emailCtrl.text.trim(), _passCtrl.text)
        : await widget.authService.login(_emailCtrl.text.trim(), _passCtrl.text);

    if (!mounted) return;
    setState(() => _loading = false);

    if (result.success) {
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(
          builder: (_) => HomeScreen(api: widget.api, authService: widget.authService),
        ),
      );
    } else {
      setState(() => _error = result.error);
    }
  }

  @override
  void dispose() {
    _emailCtrl.dispose();
    _passCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 28),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Spacer(),
              // Signature moment: same gradient language as the connect orb,
              // so the very first screen already teaches the visual identity.
              ShaderMask(
                shaderCallback: (bounds) =>
                    AppColors.orbGradient.createShader(bounds),
                child: Text(
                  'StreamPass',
                  style: Theme.of(context)
                      .textTheme
                      .displayLarge
                      ?.copyWith(color: Colors.white),
                ),
              ),
              const SizedBox(height: 12),
              Text(
                'Автоматически выбирает самый надёжный маршрут\n'
                'для каждого соединения.',
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const SizedBox(height: 40),
              TextField(
                controller: _emailCtrl,
                keyboardType: TextInputType.emailAddress,
                style: const TextStyle(color: AppColors.textPrimary),
                decoration: const InputDecoration(hintText: 'Email'),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: _passCtrl,
                obscureText: true,
                style: const TextStyle(color: AppColors.textPrimary),
                decoration: const InputDecoration(hintText: 'Пароль'),
              ),
              if (_error != null) ...[
                const SizedBox(height: 12),
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: AppColors.danger.withOpacity(0.12),
                    borderRadius: BorderRadius.circular(10),
                    border: Border.all(color: AppColors.danger.withOpacity(0.35)),
                  ),
                  child: Text(
                    _error!,
                    style: const TextStyle(color: AppColors.danger, fontSize: 13),
                  ),
                ),
              ],
              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: _loading ? null : _submit,
                  child: _loading
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            color: AppColors.bg,
                          ),
                        )
                      : Text(_isRegisterMode ? 'Создать аккаунт' : 'Войти'),
                ),
              ),
              const SizedBox(height: 16),
              if (!_isRegisterMode)
                Center(
                  child: TextButton(
                    onPressed: () {
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (_) => ForgotPasswordScreen(
                            authService: widget.authService,
                          ),
                        ),
                      );
                    },
                    child: const Text(
                      'Забыли пароль?',
                      style: TextStyle(color: AppColors.textSecondary),
                    ),
                  ),
                ),
              Center(
                child: TextButton(
                  onPressed: () =>
                      setState(() => _isRegisterMode = !_isRegisterMode),
                  child: Text(
                    _isRegisterMode
                        ? 'Уже есть аккаунт? Войти'
                        : 'Нет аккаунта? Зарегистрироваться',
                    style: const TextStyle(color: AppColors.textSecondary),
                  ),
                ),
              ),
              if (_isRegisterMode) ...[
                const SizedBox(height: 8),
                _LegalAcceptLine(
                  onTerms: () {
                    Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) => LegalDocumentScreen.terms(),
                      ),
                    );
                  },
                  onPrivacy: () {
                    Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) => LegalDocumentScreen.privacy(),
                      ),
                    );
                  },
                ),
              ],
              const Spacer(flex: 2),
            ],
          ),
        ),
      ),
    );
  }
}

class _LegalAcceptLine extends StatelessWidget {
  final VoidCallback onTerms;
  final VoidCallback onPrivacy;

  const _LegalAcceptLine({required this.onTerms, required this.onPrivacy});

  @override
  Widget build(BuildContext context) {
    const base = TextStyle(
      color: AppColors.textSecondary,
      fontSize: 12,
      height: 1.35,
    );
    const link = TextStyle(
      color: AppColors.cyan,
      fontSize: 12,
      height: 1.35,
      decoration: TextDecoration.underline,
    );
    return Text.rich(
      TextSpan(
        style: base,
        children: [
          const TextSpan(text: 'Создавая аккаунт, вы принимаете '),
          WidgetSpan(
            alignment: PlaceholderAlignment.baseline,
            baseline: TextBaseline.alphabetic,
            child: GestureDetector(
              onTap: onTerms,
              child: const Text('Условия', style: link),
            ),
          ),
          const TextSpan(text: ' и '),
          WidgetSpan(
            alignment: PlaceholderAlignment.baseline,
            baseline: TextBaseline.alphabetic,
            child: GestureDetector(
              onTap: onPrivacy,
              child: const Text('Политику конфиденциальности', style: link),
            ),
          ),
          const TextSpan(text: '.'),
        ],
      ),
      textAlign: TextAlign.center,
    );
  }
}
