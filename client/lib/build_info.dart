/// Client build metadata — bump when connect flow changes.
class BuildInfo {
  static const version = '0.1.1';
  static const buildNumber = 20;
  static const connectFlow = 'traffic-accelerator-v2';

  static String get label => 'v$version+$buildNumber ($connectFlow)';
}
