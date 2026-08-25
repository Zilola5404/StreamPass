import 'package:http/http.dart' as http;

/// Finite timeouts for startup/connect HTTP (Issue #4).
class ApiTimeouts {
  ApiTimeouts._();

  /// Default API / auth request budget.
  static const Duration http = Duration(seconds: 12);

  /// Splash / FutureBuilder must never spin forever.
  static const Duration authGate = Duration(seconds: 5);

  /// Windows relay traffic_ready watchdog after ENGINE_STARTED.
  static const Duration trafficReady = Duration(seconds: 35);
}

/// [http.Client] that applies [ApiTimeouts.http] to every request.
class TimedHttpClient extends http.BaseClient {
  TimedHttpClient({http.Client? inner, this.timeout = ApiTimeouts.http})
      : _inner = inner ?? http.Client();

  final http.Client _inner;
  final Duration timeout;

  @override
  Future<http.StreamedResponse> send(http.BaseRequest request) {
    return _inner.send(request).timeout(timeout);
  }

  @override
  void close() => _inner.close();
}
