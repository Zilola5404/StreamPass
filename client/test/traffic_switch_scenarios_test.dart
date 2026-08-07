import 'dart:convert';
import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

/// Validates scripts/traffic_expectations.json switch_scenarios for QA matrix.
void main() {
  final jsonFile = File('${Directory.current.path}/../scripts/traffic_expectations.json');

  test('switch_scenarios JSON loads and has required fields', () {
    expect(jsonFile.existsSync(), isTrue, reason: 'traffic_expectations.json missing');
    final map = jsonDecode(jsonFile.readAsStringSync()) as Map<String, dynamic>;
    final scenarios = map['switch_scenarios'] as List<dynamic>?;
    expect(scenarios, isNotNull);
    expect(scenarios!.length, greaterThanOrEqualTo(9));

    final ids = <String>{};
    for (final raw in scenarios) {
      final sc = raw as Map<String, dynamic>;
      expect(sc['id'], isNotEmpty);
      expect(sc['label'], isNotEmpty);
      expect(sc['action'], isNotEmpty);
      expect(sc['type'], isIn(['site', 'app', 'lifecycle']));
      expect(ids.add(sc['id'] as String), isTrue, reason: 'duplicate id ${sc['id']}');

      if (sc['type'] == 'site') {
        expect(sc['url'], isNotEmpty);
        expect(sc['decision'], isIn(['DIRECT', 'RELAY']));
      }
      if (sc['type'] == 'app') {
        expect(sc['package'], isNotEmpty);
      }
      expect(sc['manual_checks'], isA<List<dynamic>>());
      expect(sc['failure_signs'], isA<List<dynamic>>());
    }
  });

  test('every site scenario maps to decision matrix expectation', () {
    final map = jsonDecode(jsonFile.readAsStringSync()) as Map<String, dynamic>;
    final scenarios = map['switch_scenarios'] as List<dynamic>;
    final sites = map['sites'] as List<dynamic>;
    final siteByHost = {for (final s in sites) s['host'] as String: s['decision'] as String};

    for (final raw in scenarios) {
      final sc = raw as Map<String, dynamic>;
      if (sc['type'] != 'site') continue;
      final url = sc['url'] as String;
      final host = Uri.parse(url).host.replaceFirst(RegExp(r'^www\.'), '');
      final matrixHost = siteByHost.keys.cast<String?>().firstWhere(
            (h) => h == host || host.endsWith('.$h') || h == host.replaceFirst('www.', ''),
            orElse: () => null,
          );
      if (matrixHost != null) {
        expect(sc['decision'], siteByHost[matrixHost],
            reason: '${sc['id']}: decision must match sites matrix');
      }
    }
  });
}
