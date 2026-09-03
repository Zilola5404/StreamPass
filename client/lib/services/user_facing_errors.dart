import 'dart:async';
import 'dart:io' show SocketException;

import 'package:http/http.dart' as http;

import 'streampass_api.dart';

/// Short Russian strings for the Home orb / connect UX (Issue #4).
class UserFacingErrors {
  UserFacingErrors._();

  static const connectCta = 'Подключить';
  static const connecting = 'Подключение…';
  static const connected = 'Подключено';
  static const disconnecting = 'Отключение…';
  static const serverUnavailable =
      'Сервер временно недоступен. Попробуйте позже.';
  static const noInternet = 'Нет подключения к интернету';
  static const connectFailed = 'Не удалось подключиться';
  static const permissionDenied = 'Нужно разрешение на VPN-соединение';
  static const sessionExpired = 'Сессия истекла. Войдите снова';

  /// Map any raw error / VpnEvent message to a short user string.
  static String map(Object? error, {String? fallback}) {
    if (error == null) return fallback ?? connectFailed;
    if (error is ApiException) {
      if (error.statusCode >= 500 || error.statusCode == 0) {
        return serverUnavailable;
      }
      return mapMessage(error.message, fallback: fallback ?? serverUnavailable);
    }
    if (error is String) return mapMessage(error, fallback: fallback);
    return mapMessage('$error', fallback: fallback);
  }

  static String mapMessage(String raw, {String? fallback}) {
    final msg = raw.trim();
    if (msg.isEmpty) return fallback ?? connectFailed;
    final lower = msg.toLowerCase();

    if (_looksLikeNoInternet(lower)) return noInternet;
    // Handshake / transport failures before generic timeout classification.
    if (_looksLikeHandshakeOrTransport(lower)) {
      return fallback ?? connectFailed;
    }
    if (_looksLikeServerUnavailable(lower) || _looksLikeNoRelay(lower)) {
      return serverUnavailable;
    }
    if (_looksLikePermission(lower)) return permissionDenied;
    if (_looksLikeSession(lower)) return sessionExpired;

    // Already one of our short strings.
    if (msg == serverUnavailable ||
        msg == noInternet ||
        msg == connectFailed ||
        msg == permissionDenied ||
        msg == sessionExpired ||
        msg == connecting ||
        msg == connected ||
        msg == disconnecting ||
        msg == connectCta) {
      return msg;
    }

    // Never leak SocketException / TimeoutException / HTTP codes to the orb.
    if (_looksTechnical(lower)) return fallback ?? connectFailed;
    // Unknown short Russian-ish message — keep if short; else generic.
    if (msg.length <= 48 && !_looksTechnical(lower) && !msg.contains('Exception')) {
      return msg;
    }
    return fallback ?? connectFailed;
  }

  static bool isServerUnavailableMessage(String? msg) {
    if (msg == null) return false;
    return mapMessage(msg) == serverUnavailable;
  }

  static bool isNetworkOrServer(Object e) {
    if (e is TimeoutException) return true;
    if (e is SocketException) return true;
    if (e is http.ClientException) return true;
    if (e is ApiException && e.statusCode >= 500) return true;
    final lower = e.toString().toLowerCase();
    return _looksLikeNoInternet(lower) || _looksLikeServerUnavailable(lower);
  }

  static bool _looksLikeNoInternet(String lower) {
    return lower.contains('socketexception') ||
        lower.contains('network is unreachable') ||
        lower.contains('no address associated') ||
        lower.contains('failed host lookup') ||
        lower.contains('network_unreachable') ||
        lower.contains('offline');
  }

  static bool _looksLikeHandshakeOrTransport(String lower) {
    return lower.contains('hysteria') ||
        lower.contains('handshake') ||
        lower.contains('quic') ||
        lower.contains('wintun') ||
        lower.contains('gvisor') ||
        lower.contains('tun bridge') ||
        lower.contains('traffic_ready') ||
        lower.contains('relay_reconnect');
  }

  static bool _looksLikeServerUnavailable(String lower) {
    return lower.contains('connection refused') ||
        lower.contains('connection timed out') ||
        lower.contains('timed out') ||
        lower.contains('timeout') ||
        lower.contains('clientexception') ||
        lower.contains('connection reset') ||
        lower.contains('http 5') ||
        lower.contains('statuscode: 5') ||
        lower.contains('server unavailable') ||
        lower.contains('сервер недоступен') ||
        lower.contains('relay_unavailable') ||
        lower.contains('relay_reconnect_failed') ||
        RegExp(r'\b5\d\d\b').hasMatch(lower);
  }

  static bool _looksLikeNoRelay(String lower) {
    return lower.contains('no healthy') ||
        lower.contains('no relay') ||
        lower.contains('нет доступных сервер') ||
        lower.contains('нет активных сервер') ||
        lower.contains('empty server');
  }

  static bool _looksLikePermission(String lower) {
    return lower.contains('permission') || lower.contains('vpn-соединение');
  }

  static bool _looksLikeSession(String lower) {
    return lower.contains('session expired') ||
        lower.contains('unauthorized') ||
        lower.contains('сессия истекла');
  }

  static bool _looksTechnical(String lower) {
    return lower.contains('exception') ||
        lower.contains('error:') ||
        lower.contains('stack') ||
        lower.contains('socket') ||
        lower.contains('errno') ||
        lower.contains('hysteria') ||
        lower.contains('wintun') ||
        lower.contains('gvisor') ||
        lower.contains('quic') ||
        lower.contains('rpc') ||
        lower.contains('traffic_ready') ||
        lower.contains('tunnel') ||
        RegExp(r'\b\d{3}\b').hasMatch(lower);
  }
}
