import 'dart:math';
import 'package:flutter/material.dart';
import '../theme/app_theme.dart';

enum ConnState { disconnected, connecting, connected, error }

class ConnectOrb extends StatefulWidget {
  final ConnState state;
  final String label;
  final String subtitle;
  final VoidCallback onTap;

  const ConnectOrb({
    super.key,
    required this.state,
    required this.label,
    required this.subtitle,
    required this.onTap,
  });

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
            return SizedBox(
              width: 260,
              height: 260,
              child: Stack(
                alignment: Alignment.center,
                children: [
                  CustomPaint(
                    size: const Size(260, 260),
                    painter:
                        _OrbPainter(t: _controller.value, state: widget.state),
                  ),
                  Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        widget.state == ConnState.connected
                            ? Icons.check_rounded
                            : widget.state == ConnState.error
                                ? Icons.priority_high_rounded
                                : Icons.power_settings_new_rounded,
                        color: Colors.white.withOpacity(0.9),
                        size: 44,
                      ),
                      const SizedBox(height: 10),
                      Text(
                        widget.label,
                        textAlign: TextAlign.center,
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                      const SizedBox(height: 4),
                      Text(
                        widget.subtitle,
                        textAlign: TextAlign.center,
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: widget.state == ConnState.connected
                                  ? AppColors.green
                                  : AppColors.textSecondary,
                            ),
                      ),
                    ],
                  ),
                ],
              ),
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

    _drawOrbit(canvas, center, baseRadius * 1.08, t, glowColor);
    _drawOrbit(canvas, center, baseRadius * 1.22, -t * 0.7, AppColors.violet);

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

  void _drawOrbit(
    Canvas canvas,
    Offset center,
    double radius,
    double turn,
    Color color,
  ) {
    final paint = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1.6
      ..shader = SweepGradient(
        startAngle: 0,
        endAngle: 2 * pi,
        colors: [
          color.withOpacity(0),
          color.withOpacity(0.78),
          color.withOpacity(0),
        ],
        stops: const [0.05, 0.45, 1],
        transform: GradientRotation(turn * 2 * pi),
      ).createShader(Rect.fromCircle(center: center, radius: radius));

    canvas.save();
    canvas.translate(center.dx, center.dy);
    canvas.rotate(-0.34);
    canvas.scale(1.18, 0.42);
    canvas.translate(-center.dx, -center.dy);
    canvas.drawCircle(center, radius, paint);
    canvas.restore();
  }
}
