# StreamPass — Architecture Decision Records (ADR)

> Формат: Decision ID | Дата | Проблема | Решение | Причина | Последствия

---

## ADR-001: Go Monolith вместо микросервисов

| | |
|---|---|
| **Дата** | 2026 (MVP start) |
| **Проблема** | Как организовать backend для MVP |
| **Решение** | Single Go monolith с Clean Architecture |
| **Причина** | ТЗ §11: «Go Monolith, без микросервисов». MVP scale не требует распределения |
| **Последствия** | Простой deploy (один binary), но scaling только vertical |

---

## ADR-002: Dependency-free infrastructure

| | |
|---|---|
| **Дата** | 2026 (MVP start) |
| **Проблема** | Песочница разработки без доступа к proxy.golang.org |
| **Решение** | Vendored deps (crypto, sys, pq) + custom JWT, Redis, YAML |
| **Причина** | Минимизировать внешние зависимости, KISS/YAGNI |
| **Последствия** | Меньше supply chain risk, но custom code требует maintenance |

---

## ADR-003: stdlib http.ServeMux вместо chi/gin

| | |
|---|---|
| **Дата** | 2026 |
| **Проблема** | Выбор HTTP router |
| **Решение** | Go 1.22+ native ServeMux с method patterns |
| **Причина** | Нет необходимости в third-party router |
| **Последствия** | Меньше deps, но меньше middleware ecosystem |

---

## ADR-004: Admin API Key вместо Admin Panel

| | |
|---|---|
| **Дата** | 2026 |
| **Проблема** | Как защитить admin endpoints до построения UI |
| **Решение** | Static `X-Admin-Key` header, constant-time compare |
| **Причина** | MVP scope — admin UI не входит |
| **Последствия** | Операторы используют curl/Postman; key rotation manual. **Update 2026-08-04:** BL-020 добавил static Admin UI на `/admin/` с тем же `X-Admin-Key` (без RBAC). |

---

## ADR-005: JSONB для rules вместо normalized tables

| | |
|---|---|
| **Дата** | 2026 |
| **Проблема** | Как хранить rule sets |
| **Решение** | JSONB array в `rule_sets.rules`, versioned rows |
| **Причина** | Atomic publish per version, MVP scale ~ hundreds rules |
| **Последствия** | Простота, но нет SQL-level rule queries |

---

## ADR-006: Telemetry без FK на users

| | |
|---|---|
| **Дата** | 2026 |
| **Проблема** | Связь telemetry_events с users |
| **Решение** | user_id TEXT без FK constraint |
| **Причина** | Performance при high-volume inserts, user deletion не удаляет telemetry |
| **Последствия** | Orphan events possible, GDPR cleanup manual |

---

## ADR-007: Flutter + Android VPNService для mobile MVP

| | |
|---|---|
| **Дата** | 2026 |
| **Проблема** | Mobile client platform |
| **Решение** | Flutter UI + native Android VPNService + Go core via gomobile |
| **Причина** | ТЗ §3: Android 10+ в первой версии |
| **Последствия** | iOS/Windows/macOS — отдельные adapters позже |

---

## ADR-008: Hysteria2 как transport (planned)

| | |
|---|---|
| **Дата** | 2026 |
| **Проблема** | VPN transport protocol |
| **Решение** | Hysteria2 (готовая реализация, не свой протокол) |
| **Причина** | ТЗ §8: «Использовать готовую реализацию» |
| **Последствия** | Зависимость от Hysteria2 ecosystem; **NOT YET IMPLEMENTED in go_core** |

---

## ADR-009: Caddy для TLS termination

| | |
|---|---|
| **Дата** | 2026 |
| **Проблема** | HTTPS для API |
| **Решение** | Caddy 2 reverse proxy с auto-HTTPS |
| **Причина** | Automatic Let's Encrypt, simple config |
| **Последствия** | nip.io domain for MVP; production needs real domain |

---

## ADR-010: Subscription cancel = immediate access revocation

| | |
|---|---|
| **Дата** | 2026 |
| **Проблема** | Поведение POST /subscription/cancel |
| **Решение** | ActiveUntil = now (немедленное прекращение) |
| **Причина** | ТЗ не детализировало auto-renewal |
| **Последствия** | Нужно уточнить с продуктом перед production |

---

## ADR-011: Hysteria2 core + sing-tun для Android VPN

| | |
|---|---|
| **Дата** | 2026-08-03 |
| **Проблема** | BL-001: нужен клиентский Hysteria2 transport в go_core для Android TUN |
| **Решение** | `github.com/apernet/hysteria/core/v2` + `extras/v2/obfs` (salamander), мост TUN через `github.com/sagernet/sing-tun` system stack с `FileDescriptor` от VpnService |
| **Причина** | ТЗ §8 запрещает писать свой транспорт; sing-tun поддерживает Android fd и актуальный Handler API |
| **Последствия** | Новые Go-зависимости; сборка AAR через gomobile + JDK; self-signed relay TLS по умолчанию `insecure` без pinSHA256 |

---

## ADR-010: On-device connect diagnostics log

| | |
|---|---|
| **Дата** | 2026-08-04 |
| **Проблема** | Экран Diagnostics расширен step-by-step connect log с копированием в clipboard; прежний комментарий в коде ограничивал scope полями ТЗ §14 |
| **Решение** | Лог хранится только на устройстве; экспорт — по явному действию пользователя. Разрешены: relay id/host/port, HTTP status, auth error codes, build label, native/Dart event names. Запрещены: URL сайтов, payload трафика, содержимое `connection_config` |
| **Причина** | Connect log нужен для отладки VPN-flow на устройстве без отправки данных на сервер; это не server telemetry |
| **Последствия** | Обновлён `docs/28_SecurityChecklist.md`; комментарий в `diagnostics_screen.dart` |

---

## ADR-011: Prepare relay before TUN on Android

| | |
|---|---|
| **Дата** | 2026-08-04 |
| **Проблема** | QUIC handshake к relay таймаутил, если default route через TUN уже активен — пакеты зацикливались в пустой туннель |
| **Решение** | Разделить `PrepareRelay()` (Hysteria connect до TUN) и `StartTunnel()` (attach fd к готовой сессии) |
| **Причина** | Android VpnService поднимает full tunnel до старта Go core; без pre-connect relay недостижим |
| **Последствия** | Состояние `prepared`/`active` в `mobile/tunnel.go`; RTT handshake сохраняется в `pingMs` |

---

## Шаблон для новых ADR

```
## ADR-XXX: [Title]
| **Дата** | |
| **Проблема** | |
| **Решение** | |
| **Причина** | |
| **Последствия** | |
```
