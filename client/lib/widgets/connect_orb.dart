import 'dart:math';
import 'package:flutter/material.dart';
import '../theme/app_theme.dart';

enum ConnState { disconnected, connecting, connected, error }

class ConnectOrb extends StatefulWidget {
  final ConnState state;
  final VoidCallback onTap;

  const ConnectOrb({super.key, required this.state, required this.onTap});

  @override
  State<ConnectOrb> createState() => _ConnectOrbState();
}

class _ConnectOrbState extends State<ConnectOrb>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller = AnimationController(
    vsync: this,
    duration: const Duration(seconds: 3),
  )..repeat();

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Semantics(
      button: true,
      label: switch (widget.state) {
        ConnState.connected => 'Disconnect',
        ConnState.connecting => 'Connecting',
        ConnState.disconnected => 'Connect',
        ConnState.error => 'Retry connection',
      },
      child: GestureDetector(
        onTap: widget.onTap,
        child: AnimatedBuilder(
          animation: _controller,
          builder: (context, _) {
            return CustomPaint(
              size: const Size(220, 220),
              painter: _OrbPainter(t: _controller.value, state: widget.state),
            );
          },
        ),
      ),
    );
  }
}

class _OrbPainter extends CustomPainter {
  final double t;
  final ConnState state;
  _OrbPainter({required this.t, required this.state});

  @override
  void paint(Canvas canvas, Size size) {
    final center = size.center(Offset.zero);
    final baseRadius = size.width / 2.6;

    final pulse = state == ConnState.connecting
        ? 1 + 0.06 * sin(t * 2 * pi)
        : state == ConnState.connected
            ? 1 + 0.015 * sin(t * 2 * pi)
            : 1.0;

    final glowColor = switch (state) {
      ConnState.connected => AppColors.green,
      ConnState.connecting => AppColors.amber,
      ConnState.error => AppColors.danger,
      ConnState.disconnected => AppColors.cyan,
    };

    final glowPaint = Paint()
      ..shader = RadialGradient(
        colors: [glowColor.withOpacity(0.35), glowColor.withOpacity(0.0)],
      ).createShader(Rect.fromCircle(center: center, radius: baseRadius * 1.8))
      ..maskFilter = const MaskFilter.blur(BlurStyle.normal, 40);
    canvas.drawCircle(center, baseRadius * 1.8 * pulse, glowPaint);

    final corePaint = Paint()
      ..shader = RadialGradient(
        center: const Alignment(-0.35, -0.4),
        radius: 1.1,
        colors: [
          Color.lerp(AppColors.cyan, Colors.white, 0.35)!,
          AppColors.violet,
          AppColors.bg,
        ],
        stops: const [0.0, 0.55, 1.0],
      ).createShader(Rect.fromCircle(center: center, radius: baseRadius));
    canvas.drawCircle(center, baseRadius * pulse, corePaint);

    final ringPaint = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = 2
      ..color = glowColor.withOpacity(0.5);
    canvas.drawCircle(center, baseRadius * pulse + 6, ringPaint);

    final highlight = Paint()
      ..color = Colors.white.withOpacity(0.25)
      ..maskFilter = const MaskFilter.blur(BlurStyle.normal, 12);
    canvas.drawCircle(
      center.translate(-baseRadius * 0.35, -baseRadius * 0.4),
      baseRadius * 0.22,
      highlight,
    );
  }

  @override
  bool shouldRepaint(covariant _OrbPainter old) =>
      old.t != t || old.state != state;
}
