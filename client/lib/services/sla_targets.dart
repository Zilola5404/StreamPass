/// Client SLA targets from ТЗ §22 / FS §5 (BL-053).
class SlaTargets {
  /// Cold start / session check before first interactive screen.
  static const Duration coldStartMax = Duration(seconds: 2);

  /// Connect to Connected (VPN + tunnel ready).
  static const Duration connectMax = Duration(seconds: 5);

  /// Recover after network loss / failover.
  static const Duration recoverMax = Duration(seconds: 10);

  static bool meetsColdStart(Duration measured) =>
      !measured.isNegative && measured <= coldStartMax;

  static bool meetsConnect(Duration measured) =>
      !measured.isNegative && measured <= connectMax;

  static bool meetsRecover(Duration measured) =>
      !measured.isNegative && measured <= recoverMax;
}
