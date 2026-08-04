/// Client build metadata — bump when connect flow changes.
class BuildInfo {
  static const version = '0.1.1';
  static const buildNumber = 15;
  static const connectFlow = 'audit003-status-udp-v1';

  static String get label => 'v$version+$buildNumber ($connectFlow)';
}
