/// Client build metadata — bump when connect flow changes.
class BuildInfo {
  static const version = '0.1.1';
  static const buildNumber = 7;
  static const connectFlow = 'rule-engine-bl006-v1-pingfix';

  static String get label => 'v$version+$buildNumber ($connectFlow)';
}