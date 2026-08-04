/// Semver-ish comparison for StreamPass client versions (e.g. 0.1.1).
/// Build numbers (+N) are ignored; only the dotted version is compared.
int compareClientVersions(String a, String b) {
  final pa = _parts(a);
  final pb = _parts(b);
  final n = pa.length > pb.length ? pa.length : pb.length;
  for (var i = 0; i < n; i++) {
    final x = i < pa.length ? pa[i] : 0;
    final y = i < pb.length ? pb[i] : 0;
    if (x != y) return x < y ? -1 : 1;
  }
  return 0;
}

List<int> _parts(String raw) {
  final core = raw.trim().split('+').first.split('-').first;
  if (core.isEmpty) return const [0];
  return core
      .split('.')
      .map((p) => int.tryParse(p.replaceAll(RegExp(r'[^0-9]'), '')) ?? 0)
      .toList();
}

enum UpdateUrgency { none, optional, required }

class UpdateCheckResult {
  final UpdateUrgency urgency;
  final String currentVersion;
  final String? latestVersion;
  final String? minSupportedVersion;
  final String? downloadUrl;

  const UpdateCheckResult({
    required this.urgency,
    required this.currentVersion,
    this.latestVersion,
    this.minSupportedVersion,
    this.downloadUrl,
  });

  bool get hasUpdate => urgency != UpdateUrgency.none;
}

UpdateCheckResult evaluateClientUpdate({
  required String currentVersion,
  required String minSupportedVersion,
  String latestVersion = '',
  String downloadUrl = '',
}) {
  final min = minSupportedVersion.trim();
  final latest = latestVersion.trim();
  final url = downloadUrl.trim();

  if (min.isNotEmpty && compareClientVersions(currentVersion, min) < 0) {
    return UpdateCheckResult(
      urgency: UpdateUrgency.required,
      currentVersion: currentVersion,
      latestVersion: latest.isEmpty ? null : latest,
      minSupportedVersion: min,
      downloadUrl: url.isEmpty ? null : url,
    );
  }

  if (latest.isNotEmpty &&
      url.isNotEmpty &&
      compareClientVersions(currentVersion, latest) < 0) {
    return UpdateCheckResult(
      urgency: UpdateUrgency.optional,
      currentVersion: currentVersion,
      latestVersion: latest,
      minSupportedVersion: min.isEmpty ? null : min,
      downloadUrl: url,
    );
  }

  return UpdateCheckResult(
    urgency: UpdateUrgency.none,
    currentVersion: currentVersion,
    latestVersion: latest.isEmpty ? null : latest,
    minSupportedVersion: min.isEmpty ? null : min,
    downloadUrl: url.isEmpty ? null : url,
  );
}
