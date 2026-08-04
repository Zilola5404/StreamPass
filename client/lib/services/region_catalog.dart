/// Canonical relay regions (ТЗ Этап 6 / BL-026).
library;

class RegionInfo {
  final String code;
  final String city;
  final String country;
  final String label;

  const RegionInfo({
    required this.code,
    required this.city,
    required this.country,
    required this.label,
  });

  factory RegionInfo.fromJson(Map<String, dynamic> json) => RegionInfo(
        code: json['code'] as String,
        city: json['city'] as String? ?? '',
        country: json['country'] as String? ?? '',
        label: json['label'] as String? ?? json['code'] as String,
      );
}

const fallbackRegions = <RegionInfo>[
  RegionInfo(code: 'de', city: 'Frankfurt', country: 'Germany', label: 'Frankfurt (DE)'),
  RegionInfo(code: 'nl', city: 'Amsterdam', country: 'Netherlands', label: 'Amsterdam (NL)'),
  RegionInfo(code: 'pl', city: 'Warsaw', country: 'Poland', label: 'Warsaw (PL)'),
  RegionInfo(code: 'fi', city: 'Helsinki', country: 'Finland', label: 'Helsinki (FI)'),
];

String normalizeRegionCode(String? raw) {
  final s = (raw ?? '').trim().toLowerCase();
  if (s.isEmpty) return '';
  for (final info in fallbackRegions) {
    if (info.code == s) return info.code;
  }
  if (s.contains('frankfurt') || s.contains('germany') || s == 'de') return 'de';
  if (s.contains('amsterdam') || s.contains('netherlands') || s.contains('holland') || s == 'nl') {
    return 'nl';
  }
  if (s.contains('warsaw') || s.contains('poland') || s.contains('warszawa') || s == 'pl') {
    return 'pl';
  }
  if (s.contains('helsinki') || s.contains('finland') || s == 'fi') return 'fi';
  return s;
}

String regionLabel(String? raw) {
  final code = normalizeRegionCode(raw);
  for (final info in fallbackRegions) {
    if (info.code == code) return info.label;
  }
  return code.isEmpty ? (raw ?? '') : code;
}

String regionFlag(String? raw) {
  switch (normalizeRegionCode(raw)) {
    case 'de':
      return 'DE';
    case 'nl':
      return 'NL';
    case 'pl':
      return 'PL';
    case 'fi':
      return 'FI';
    default:
      return 'GL';
  }
}
