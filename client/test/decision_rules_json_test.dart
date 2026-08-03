import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:streampass/services/streampass_api.dart';

void main() {
  test('RuleSet round-trips JSON for Go decision engine', () {
    const set = RuleSet(
      version: 3,
      rules: [
        RouteRule(kind: 'DOMAIN', pattern: '*.ru', mode: 'DIRECT'),
        RouteRule(kind: 'CIDR', pattern: '178.248.232.0/21', mode: 'DIRECT'),
        RouteRule(kind: 'DOMAIN', pattern: 'youtube.com', mode: 'RELAY'),
      ],
    );

    final encoded = jsonEncode(set.toJson());
    final decoded = RuleSet.fromJson(jsonDecode(encoded) as Map<String, dynamic>);

    expect(decoded.version, 3);
    expect(decoded.rules.length, 3);
    expect(decoded.rules.first.pattern, '*.ru');
    expect(decoded.rules.last.mode, 'RELAY');
  });
}
