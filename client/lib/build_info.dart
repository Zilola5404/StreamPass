/// Client build metadata — bump when connect flow changes.
class BuildInfo {
  static const version = '0.1.1';
  static const buildNumber = 21;
  static const connectFlow = 'ru-split-tunnel-v1';

  static String get label => 'v$version+$buildNumber ($connectFlow)';
}
