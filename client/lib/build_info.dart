/// Client build metadata — bump when connect flow changes.
class BuildInfo {
  static const version = '0.1.1';
  static const buildNumber = 9;
  static const connectFlow = 'disconnect-crashfix-v1';

  static String get label => 'v$version+$buildNumber ($connectFlow)';
}