# Архитектурное решение

**ID:** TASK-02  
**Название:** Traffic Reliability & Network Diagnostics (клиент Android)  
**Роль:** Chief Solution Architect  
**Дата:** 2026-08-07  
**Статус:** ✅ Разрешить разработку **только как REWORK по Routing Policy**  
**Политика маршрутизации (SoT):** [`docs/07.4_RoutingPolicy.md`](../../docs/07.4_RoutingPolicy.md)  
**Источник анализа:** `ai/CurrentTask.md`, WIP, ТЗ §5–7, FS §6, `docs/33`, согласованный разбор слоёв (2026-08-07)

---

## ADDENDUM 2026-08-07 — согласование с Routing Policy

### Вердикт по предложению разработчика

**«QUIC → DIRECT; TCP RELAY с fallback на DIRECT»**

| Как использовать | Вердикт |
|------------------|---------|
| Постоянная архитектура продукта | ❌ **Отклонено** |
| Временная диагностика (`tcp_only` / `direct_test`, не default `split`) | ✅ Допустимо |
| Product path | Rule Engine → Decision Engine → Transport.Execute(Mode) |

**Согласие архитектора с разбором заказчика/ревью:** полное.

Причины отклонения product-варианта:

1. Transport принимает продуктовые решения (порт/протокол → Mode) — нарушение слоёв.  
2. QUIC нельзя делать правилом продукта — foreign QUIC снова упрётся в блокировки.  
3. Fallback должен быть политикой правила (`allow_direct_fallback` / `mode=FALLBACK`), не silent if в tunbridge.  
4. Накопление transport-исключений уничтожит Rule Engine.  
5. Симптом «relay up, data нет» — чинить data-plane (MTU / protect / DNS / Hysteria), не маскировать DIRECT.

### Зафиксированные продуктовые ответы (закрывают B1–B5)

| ID | Решение |
|----|---------|
| B1 DefaultMode | **DIRECT** (FS §6) — WIP `ModeRelay` откатить |
| B2 QUIC→DIRECT | **Запрещён** в product `split`; только diagnostic `tcp_only` |
| B3 RELAY→DIRECT | Только если правило `FALLBACK` / `allow_direct_fallback=true`; must-relay — **нет** тихого DIRECT |
| B4 Network Mode UI | Только E09 / debug; не E05 Must |
| B5 Широкие CIDR | Domain-first; Cloudflare `/12` и аналоги — **удалить** |

### Обязательный REWORK WIP (чеклист разработчика)

1. Удалить из product path `quic_direct_bypass` (всегда UDP/443→DIRECT).  
2. Удалить silent «любой RELAY → DIRECT» из `tunbridge`; transport возвращает Result; Decision применяет §6 политики.  
3. Откатить `DefaultMode` → `ModeDirect`.  
4. DNS-in-TUN `10.10.0.1` + split DNS + `HostForIP` — **оставить**.  
5. Blackhole: dial ok + нет first_byte за 3s → fail вверх в Decision (не Mode в transport).  
6. Must-relay (YouTube и т.п.): `mode=RELAY`, fallback запрещён → error/diag / смена relay.  
7. Diagnostic modes — E09/debug; default всегда `split`.  
8. Узкие domain `DefaultRelayRules` OK; широкие CIDR — вырезать.  
9. Основной список RELAY — backend rules (`publish-accelerator-rules.sh`).  
10. Чинить Hysteria TCP/UDP data-plane отдельно; не подменять политикой.  
11. Следовать `docs/07.4_RoutingPolicy.md` буквально; отклонения — только обновление 07.4 + ADR.

### Пять слоёв (обязательно)

```
Packet → DNS Resolver → Rule Engine → Decision Engine → Route Manager → Transport / Relay
```

Transport / VpnService / Relay Manager **не** знают YouTube/Госуслуги.

---

## Описание задачи

### Что обнаружено

В `ai/CurrentTask.md` **нет формализованной задачи разработки**. Next-пункты — ops (домен, manual E2E, ЮKassa по запросу).

В working tree при этом уже лежит крупный незакоммиченный пакет изменений (~800+ строк), фактически начинающий новую задачу:

| Тема WIP | Суть |
|----------|------|
| Network Mode | UI на E05: `split` / `full_relay` / `direct_test` / `tcp_only` + MTU + blockUdp443 |
| DefaultMode | **WIP: `ModeRelay`** vs **FS/HEAD: `ModeDirect`** |
| DNS | VPN DNS → `10.10.0.1` (Go dnscache) вместо Yandex OS DNS |
| DefaultRelayRules | Расширение domain + широкие CIDR (Google, Meta, Telegram, **Cloudflare /12**) |
| Transport | Silent RELAY→DIRECT fallback; **всегда** UDP/443 RELAY→DIRECT (`quic_direct_bypass`) |
| Observability | `[conn]` лог; QA matrix `traffic_expectations.json` + `VerifyAppSiteSwitch.ps1` |

Это **не** продолжение TASK-01 «как есть»: TASK-01 закрывал операторскую диагностику (`diag_events` / Admin). WIP меняет **продуктовое поведение маршрутизации**.

### Формальная постановка (предложение архитектора)

После ответов PM задача должна быть оформлена как:

> **TASK-02 — Traffic Reliability & Network Diagnostics**  
> Стабилизировать split-tunnel маршрутизацию (DNS-in-TUN, domain match на IP-only flows, серверные/builtin RELAY rules, корректный FALLBACK) и дать **диагностические** режимы сети для поддержки/QA — без превращения продукта в full-tunnel VPN и без противоречия FS §6.

### Проверка соответствия документации (gate)

| Проверка | Результат |
|----------|-----------|
| Соответствует ТЗ §1–2 (одна кнопка, без ручной настройки) | ⚠️ WIP: Network Mode на E05 противоречит |
| Соответствует FS §1.2 / §6 (ускоритель, DefaultMode=DIRECT) | ❌ WIP: `DefaultMode=ModeRelay` прямо противоречит FS §6 п.3 |
| Business Rules (`docs/02.5_BusinessRules.md`) | ❌ Документ **отсутствует** в репозитории |
| Не нарушает архитектуру Clean Architecture / go_core | ✅ Слои клиента соблюдены |
| Не дублирует функционал | ⚠️ Частично пересекается с TASK-01 / E09 Diagnostics |
| Можно ли проще | ✅ Да: серверные rules + DNS-in-TUN + узкий FALLBACK; без user-facing Network Mode |
| Риски | Высокие (см. раздел «Риски») |

**Остановка процесса:** пока PM не зафиксирует ответы на блокеры ниже, разработка/мердж WIP **запрещены**.

---

## Обоснование решения

### Почему не «разрешить как есть»

1. **Конфликт источника истины по DefaultMode.**  
   FS §6 и `docs/33` требуют `DefaultMode = DIRECT`. HEAD в коде — `ModeDirect`. WIP меняет на `ModeRelay` с аргументом «в TUN уже только international». Это продуктовое решение, а не багфикс: без обновления FS + ADR менять нельзя.

2. **Постоянный `quic_direct_bypass` (UDP/443 → DIRECT).**  
   В WIP foreign QUIC всегда идёт protect()-DIRECT, даже при решении RELAY. Это обходит ускоритель для приложений с QUIC-first (Telegram, LinkedIn, Chrome HTTP/3). В ТЗ/FS такого поведения нет. Допустимо только как:
   - временный workaround с явным expiry/ADR, **или**
   - поведение режима `tcp_only` / diagnostic toggle, **не** default `split`.

3. **Silent RELAY→DIRECT на любой fail dial.**  
   ТЗ §5 знает режим `FALLBACK`. Silent fallback без `ModeFallback` / reason / метрик размывает Decision Engine и маскирует поломку relay.

4. **Network Mode на E05.**  
   FS E05 Must не содержит режимов full_relay/direct_test. ТЗ запрещает ручную настройку маршрутизации. Diagnostic modes допустимы только в E09 / debug build / скрытом флаге.

5. **Широкие CIDR (особенно `104.16.0.0/12`).**  
   Cloudflare /12 затягивает массу SaaS/CDN в RELAY (и наоборот рискует ошибочно тащить «лишнее»). Списки CIDR устаревают. Предпочтение: domain rules + reverse DNS; CIDR — точечно и документировано.

6. **Нет BL/TASK id в backlog.**  
   По `00_ProjectRules` / workflow задача должна быть в backlog до разработки.

### Почему DNS-in-TUN и часть reliability — правильное направление

IP-only flows после OS DNS ломают domain rules — это подтверждено `docs/33`. Перенос DNS на `10.10.0.1` + `dnscache.HostForIP` — корректный архитектурный путь для accelerator, **если** внутри Go сохраняется split DNS (`.ru` → Yandex, foreign → DoH) и telemetry без path/query.

### Рекомендуемая продуктовая модель (после ответов PM)

```
Продуктовый default (networkMode=split, единственный для пользователя):
  OS: RU CIDR excludeRoute + app bypass (как сейчас)
  DNS: 10.10.0.1 → Go dnscache (split resolver)
  Decision: exclusions → backend rules → builtin defaults → DefaultMode=DIRECT (если PM не меняет FS)
  RELAY dial fail → ModeFallback semantics (один DIRECT retry, reason, diag)
  UDP/443: НЕ always-DIRECT; см. блокеры PM

Диагностика (не продукт):
  force DIRECT / force RELAY / tcp_only / MTU
  только E09 или debug; reconnect hint; default всегда split
```

---

## Архитектура

### Область воздействия

| Слой | Затрагивается? | Комментарий |
|------|----------------|-------------|
| Backend | Нет (обязательно) | Rules публикуются существующим Admin/API |
| Database | Нет | — |
| Docker / Caddy | Нет | — |
| API | Нет | Опционально позже: remote flags — **out of scope** |
| Flutter | Да | Settings/Diagnostics + VpnChannel options |
| Android VpnService | Да | DNS, MTU, route mode, optionsJson |
| go_core decision | Да | defaults, forceMode, DefaultMode (по решению PM) |
| go_core dnscache | Да | reverse IP→host, DNS-in-TUN |
| go_core tunbridge | Да | Options, FALLBACK, conn log |
| Telemetry / Admin diag | Минимально | Новые `decision_reason` значения; без новых endpoint |
| Monitoring | Нет | — |

### Компонентная схема (целевая)

```
Flutter (E02 Connect / E05 or E09 Diagnostics)
  → SettingsService (networkMode default=split, mtu, blockUdp443)
  → VpnChannel.connect(..., optionsJson)
       ↓
StreamPassVpnService
  → Builder: addr 10.10.0.1/30, DNS 10.10.0.1, MTU from options
  → VpnRouteConfigurator(mode):
        split → RU exclude / intl routes
        full_relay|direct_test → 0.0.0.0/0 (diagnostic only)
  → VpnBypassApps (без изменений контракта)
  → TunnelBridge.StartTunnel(..., optionsJson)
       ↓
go_core mobile.StartTunnel
  → parseTunnelOptions
  → AtomicEngine (+ optional SetForceMode for diagnostics)
  → tunbridge.StartWithOptions(BlockUDP443, …)
       ↓
Decision per flow:
  forceMode? → else DecideDetailed
  DNS reverse HostForIP if host empty
  DIRECT | RELAY | FALLBACK
```

### Принципы (обязательные)

1. **KISS / YAGNI:** не вводить remote feature-flags, новые таблицы, новые библиотеки.
2. **Одна кнопка для пользователя:** продуктовый путь всегда `split`.
3. **DIRECT ≠ app bypass** — не ломать `docs/33`.
4. **Privacy:** `[conn]`/`[diag]` — host only, без path/query/cookies (ТЗ §14, ADR-012).
5. **First-match rules:** `DefaultDirectRules` → backend rules → `DefaultRelayRules` → DefaultMode.
6. **Новые ADR обязательны** для: DNS-in-TUN; DefaultMode (если меняется); QUIC policy; FALLBACK semantics.

---

## Алгоритм работы

### A. Connect (продуктовый split)

1. Пользователь нажимает «Подключить» (подписка ACTIVE, VPN permission).  
   ↓  
2. Flutter загружает rules (`GET /api/v1/rules`) + exclusions + bypassPackages.  
   ↓  
3. Берёт `networkMode=split` (игнорировать user override в release, если PM запретит UI).  
   ↓  
4. `optionsJson = {networkMode, mtu, blockUdp443}`.  
   ↓  
5. Native: `PrepareRelay` (ADR-013) → handshake Hysteria (UDP/TCP underlay BL-017).  
   ↓  
6. VpnService.Builder: TUN `10.10.0.1/30`, DNS `10.10.0.1`, MTU, split routes, app bypass.  
   ↓  
7. `StartTunnel(fd, …, optionsJson)`.  
   ↓  
8. Go: engine из rules+exclusions+defaults; forceMode пустой.  
   ↓  
9. На каждый flow:  
   a. Target(host/ip/port/proto); если host пуст → `HostForIP`.  
   b. DecideDetailed.  
   c. DIRECT → protect dial.  
   d. RELAY → Hysteria dial; при fail → FALLBACK path (один DIRECT retry) + reason.  
   e. FALLBACK rule → RELAY then DIRECT per existing ModeFallback.  
   f. Emit `[decision]` / `[diag]` / `[conn]` без URL path.  
   ↓  
10. UI Connected; telemetry/diag upload по существующему каналу TASK-01.

### B. Diagnostic mode (только после PM-одобрения UX)

1. Пользователь/инженер на E09 (или debug) выбирает mode ≠ split.  
   ↓  
2. SnackBar: «Переподключите VPN».  
   ↓  
3. При следующем Connect:  
   - `full_relay` → forceMode=RELAY + full tunnel routes.  
   - `direct_test` → forceMode=DIRECT + full tunnel routes.  
   - `tcp_only` → split routes + BlockUDP443=true.  
   ↓  
4. Логировать mode в connect log (`networkMode=…`).  
   ↓  
5. Release-сборка: либо скрыто, либо сброс к `split` при старте приложения.

### C. QA switch matrix

1. Connect в `split`.  
   ↓  
2. `VerifyAppSiteSwitch.ps1` читает `traffic_expectations.json`.  
   ↓  
3. Для каждого scenario: launch URL/app → проверить decision/log patterns / manual checks.  
   ↓  
4. Отчёт в `reports/QA/`.

---

## Структура модулей

### Пакеты / сервисы (существующие — расширить, не плодить)

| Модуль | Ответственность |
|--------|-----------------|
| `client/lib/services/settings_service.dart` | Persist diagnostic options; default split/mtu=1400 |
| `client/lib/services/vpn_channel.dart` | Передача `optionsJson` в MethodChannel |
| `client/lib/screens/diagnostics_screen.dart` **или** скрытая секция | UX diagnostic modes (не E05 Must) |
| `VpnRouteConfigurator` | split vs diagnostic full-tunnel routes |
| `TunnelBridge` | Совместимость StartTunnel 7/8 args на время миграции AAR |
| `mobile` (`tunnel.go`) | `tunnelOptions`, apply forceMode/MTU/BlockUDP443 |
| `decision` | defaults, AtomicEngine.SetForceMode, DefaultMode per ADR |
| `dnscache` | in-TUN resolve + HostForIP + LastResolveMS |
| `tunbridge` | Options, FALLBACK, conn log, UDP443 policy per ADR |

### Интерфейсы / структуры данных (контракт)

**TunnelOptions (JSON, Flutter → Native → Go):**

| Поле | Тип | Допустимые значения | Default |
|------|-----|---------------------|---------|
| `networkMode` | string | `split` \| `full_relay` \| `direct_test` \| `tcp_only` | `split` |
| `mtu` | int | `1280` \| `1350` \| `1400` (clamp 1200–1500) | `1400` |
| `blockUdp443` | bool | true/false; true если mode=`tcp_only` | `false` |

**Decision reasons (новые, стабильные строки):**

| reason | Когда |
|--------|-------|
| `network_mode_DIRECT` / `network_mode_RELAY` | forceMode |
| `fallback_after_relay_fail` | RELAY dial fail → DIRECT |
| `udp443_blocked` | BlockUDP443 drop |
| `quic_policy_*` | только если PM утвердит QUIC policy (имя зафиксировать в ADR) |

**Не создавать:** новые Go modules, новые Flutter packages, backend services, Redis keys, DB tables.

---

## Изменяемые файлы

### Разрешено изменять (после ✅ PM)

**Flutter**

- `client/lib/services/settings_service.dart`
- `client/lib/services/vpn_channel.dart`
- `client/lib/screens/home_screen.dart` (передача options)
- `client/lib/screens/diagnostics_screen.dart` **и/или** `settings_screen.dart` — **только по решению PM о UX**
- `client/lib/build_info.dart` / `client/pubspec.yaml` (build bump)
- `client/test/traffic_behavior_test.dart`
- `client/test/traffic_switch_scenarios_test.dart` (новый OK)

**Android**

- `client/android/app/src/main/kotlin/com/streampass/app/StreamPassVpnService.kt`
- `client/android/app/src/main/kotlin/com/streampass/app/TunnelBridge.kt`
- `client/android/app/src/main/kotlin/com/streampass/app/VpnRouteConfigurator.kt`
- `client/android/app/src/main/kotlin/com/streampass/app/VpnBypassApps.kt` (только если package ids)
- `client/android/app/libs/streampasscore.aar` (rebuild; не коммитить extract)

**go_core**

- `client/go_core/mobile/tunnel.go`
- `client/go_core/mobile/tunnel_decision_test.go`
- `client/go_core/mobile/traffic_matrix_test.go`
- `client/go_core/internal/decision/*.go` (+ tests)
- `client/go_core/internal/dnscache/doh.go`, `reverse.go` (+ tests)
- `client/go_core/internal/tunbridge/bridge.go` (+ tests при наличии)

**Scripts / docs / expectations**

- `scripts/traffic_expectations.json`
- `scripts/VerifyAppSiteSwitch.ps1`
- `scripts/publish-accelerator-rules.sh` (публикация **серверных** rules)
- `docs/06_TestPlan.md`
- `docs/33_DirectVsVpnBypass.md`
- `docs/07_Architecture.md` (после реализации)
- `docs/11_Decisions.md` (ADR-015+)
- `docs/04_Backlog.md`, `docs/10_Progress.md`, `docs/17_CHANGELOG.md`
- `ai/CurrentTask.md`, `docs/tasks/TASK-02_*.md` (создать)

### Запрещено изменять

- `backend/**` (для TASK-02) — кроме опциональной публикации rules через уже существующий Admin API скриптом
- `backend/internal/infrastructure/postgres/migrations/**`
- `docs/08_API.md` / `docs/09_Database.md` — без API/DB изменений
- `docker-compose.yml`, `Caddyfile`, monitoring configs
- Секреты: `.env`, `connection_config`, keystore
- `client/android/app/libs/_aar_extract/**` — не коммитить
- `diag.json`, `diag-fresh.json`, `docs.zip` — не коммитить
- Windows/iOS/macOS клиенты (BL-023…025)
- Billing / ЮKassa (BL-040) без явного запроса
- Изменение контракта `POST /api/v1/diag` / схемы `diag_events` без отдельного ADR

---

## Новые файлы

| Файл | Назначение |
|------|------------|
| `docs/tasks/TASK-02_Traffic_Reliability_Network_Diagnostics.md` | Формальная задача (после PM) |
| `reports/Architecture/TASK-02-ArchitectureDecision.md` | Этот документ |
| `client/test/traffic_switch_scenarios_test.dart` | Валидация JSON matrix |
| `client/go_core/internal/dnscache/reverse.go` | IP→host (если ещё нет в HEAD) |
| ADR записи в `docs/11_Decisions.md` | ADR-015 DNS-in-TUN; ADR-016 DefaultMode/QUIC/FALLBACK |

Новые пакеты Go/Dart — **не создавать**.

---

## Используемые библиотеки

**Новые библиотеки: запрещены.**

Использовать только уже принятый стек (`docs/24_Dependencies.md`):

| Компонент | Версия / источник |
|-----------|-------------------|
| Go (go_core) | 1.22.2 |
| Flutter / Dart | Flutter 3.x / Dart ≥3.3 |
| shared_preferences | ^2.2.3 (уже есть) |
| encoding/json (stdlib) | для optionsJson |
| Hysteria2 + sing-tun | ADR-011 (уже в go_core) |
| gomobile AAR | существующий pipeline |

Запрет: GeoIP DB, новые VPN SDK, ML, отдельный QUIC stack, feature-flag SaaS.

---

## API

**Изменения HTTP API не требуются.**

Rules по-прежнему: `GET /api/v1/rules` (JWT).  
Публикация расширенных RELAY domain rules — через существующий Admin Rules publish (`scripts/publish-accelerator-rules.sh`), не через раздувание только client defaults.

Diag upload: существующий `POST /api/v1/diag` (TASK-01). Новые поля в payload **не обязательны**; новые `decision_reason` — строки в уже свободном текстовом поле.

---

## Database

**Миграции не требуются.**  
Таблицы не меняются.

---

## Безопасность

| Контроль | Требование |
|----------|------------|
| JWT / RBAC | Без изменений; rules только после auth как сейчас |
| Secrets | Не логировать `connection_config`, пароли, tokens |
| Telemetry privacy | host / `https://host` only; запрет path, query, cookies, body |
| Validation | `networkMode` whitelist; MTU clamp; неизвестный mode → `split` |
| Rate limit | N/A (клиентская задача) |
| Encryption / TLS | Без изменений; Hysteria TLS как сейчас |
| Diagnostic modes | Не рекламировать как «полный VPN»; в release — ограничить доступ |
| App bypass | Не ослаблять список банков/госуслуг |

---

## Производительность

| Тема | Решение |
|------|---------|
| Decide path | Без аллокаций сверх текущих; forceMode — один RLock |
| Reverse DNS map | Bounded cache (TTL + max entries); запрет unbounded growth |
| `[conn]` log | Sample/успешные flows OK; не писать на каждый пакет UDP |
| DNS-in-TUN | Один resolver path; cache hit must быть hot path |
| MTU 1280 | Допустимо для плохих сетей; default 1400 |
| Broad CIDR | Запретить Cloudflare /12 в builtin — риск лишнего RELAY dial load |

Цели ТЗ §22 не ухудшать: connect ≤5 с, recover ≤10 с (device SLA уже измеряется скриптом).

---

## Масштабируемость

1. **Серверные rules** — основной способ добавлять домены без релиза APK.  
2. **Builtin defaults** — только safety net при sparse/offline rules.  
3. **Diagnostic options** — локальный JSON контракт; позже можно зеркалировать в `/config` без ломания клиента.  
4. **Platform adapters:** логика в go_core; Android только DNS/routes/MTU.  
5. **Не** закладывать Kind=APP на backend в TASK-02 (уже отвергнуто как блокер в `docs/33`).

---

## Риски

| ID | Риск | Severity | Митигация |
|----|------|----------|-----------|
| R1 | DefaultMode=RELAY противоречит FS | Critical | Блокер PM; до ответа — запрет мерджа |
| R2 | QUIC always-DIRECT обходит accelerator | High | Только diagnostic / ADR с expiry |
| R3 | Silent fallback маскирует мёртвый relay | High | ModeFallback + reason + diag |
| R4 | Cloudflare /12 misroute | High | Удалить из defaults; domain-first |
| R5 | DNS-in-TUN регрессия RU geo-DNS | High | Сохранить split resolver; device QA матрица |
| R6 | Network Mode на E05 путает пользователя | Medium | Перенос в E09 / debug |
| R7 | AAR signature mismatch 7 vs 8 args | Medium | TunnelBridge dual invoke |
| R8 | Отсутствие `02.5_BusinessRules.md` | Medium | PM создать или явно сказать «FS+TZ = BR» |
| R9 | WIP смешивает product+debug в одном diff | Medium | Разрезать PR: reliability / diagnostics |

---

## Definition of Done

Задача Done только когда:

1. Есть `docs/tasks/TASK-02_*.md` + запись в `docs/04_Backlog.md`.  
2. PM закрыл блокеры B1–B5 (ниже).  
3. ADR-015+ записаны в `docs/11_Decisions.md`.  
4. FS §6 и код согласованы по DefaultMode (либо код, либо FS обновлены осознанно).  
5. Продуктовый default = `split`; diagnostic modes не ломают one-button UX.  
6. `go test ./...` в `client/go_core` PASS; `flutter test` PASS; `flutter analyze` clean.  
7. AAR пересобран; APK устанавливается; connect log содержит `networkMode=split`, `vpn dns=10.10.0.1` (если DNS-in-TUN approved).  
8. QA matrix: YouTube/IG RELAY; yandex/2ip DIRECT; Госуслуги bypass; нет FATAL на disconnect.  
9. Обновлены: `docs/33`, `docs/06_TestPlan`, `docs/07_Architecture`, `docs/03_CurrentState`, `docs/17_CHANGELOG`, `ai/*`.  
10. В commit нет `_aar_extract`, `diag*.json`, `docs.zip`, секретов.

---

## Чек-лист для разработчика

### Перед кодом

- [ ] Дождаться вердикта PM по B1–B5  
- [ ] Прочитать этот ADR + `docs/33` + FS §6  
- [ ] Создать TASK-02 файл и BL id  
- [ ] Не продолжать «как в WIP» без сверки с решениями PM  

### Реализация (рекомендуемый порядок)

1. [ ] DNS-in-TUN + HostForIP + unit tests (без смены DefaultMode)  
2. [ ] Серверные RELAY domain rules через publish script; client defaults — узкий safety net  
3. [ ] Корректный FALLBACK (reason + один retry), убрать silent «всегда DIRECT» если PM запретит  
4. [ ] QUIC policy строго по ADR  
5. [ ] TunnelOptions JSON end-to-end; default split  
6. [ ] Diagnostic UI **только** в согласованном месте  
7. [ ] MTU clamp + reconnect hint  
8. [ ] Rebuild AAR; bump build number  
9. [ ] Tests: decision matrix, traffic_switch_scenarios, defaults_test  
10. [ ] Обновить docs + ADR  

### Запреты разработчику

- [ ] Не менять DefaultMode «потому что так удобнее»  
- [ ] Не добавлять Cloudflare `/12`  
- [ ] Не писать рабочий backend/API «заодно»  
- [ ] Не коммитить extract AAR / diag dumps  

---

## Чек-лист для Code Review

- [ ] Соответствие решениям PM + этому документу  
- [ ] FS §6 / ADR согласованы с `DefaultMode`  
- [ ] Нет user-facing full-tunnel без пометки diagnostic  
- [ ] Privacy: нет path/query в логах  
- [ ] FALLBACK имеет reason и не зацикливается  
- [ ] TunnelBridge совместим со старым/новым StartTunnel  
- [ ] Нет новых зависимостей  
- [ ] Тесты на defaults / forceMode / HostForIP / options parse  
- [ ] Diff не содержит secrets / binaries extract  
- [ ] Документация обновлена  

---

## Чек-лист для QA

### Smoke

- [ ] Cold start ≤2 с (ориентир)  
- [ ] Login → Connect → Connected  
- [ ] Disconnect без crash (`stop complete`, нет FATAL)  

### Матрица трафика (`scripts/VerifyAppSiteSwitch.ps1`)

- [ ] `site_yandex` / `2ip.ru` — DIRECT, RU IP  
- [ ] `site_youtube` / Instagram — RELAY (если PM не ослабил QUIC)  
- [ ] Госуслуги / банк — app bypass, нет «отключите VPN»  
- [ ] Chrome: RU сайт ок; YouTube в браузере ускоряется  

### Diagnostic (если разрешены)

- [ ] `direct_test`: LinkedIn/Telegram ок ⇒ проблема была в relay path  
- [ ] `full_relay`: всё через RELAY; регрессии логируются  
- [ ] `tcp_only`: UDP/443 drop виден в логе  
- [ ] После смены mode без reconnect — поведение старое (ожидаемо)  

### Регрессии

- [ ] Refresh token / auto reconnect  
- [ ] Region switch reconnect (BL-046)  
- [ ] Soft update dialog  
- [ ] Диагностика upload (TASK-01) всё ещё работает  

---

## Блокеры для Product Manager (обязательные ответы)

| ID | Вопрос | Варианты | Влияние |
|----|--------|----------|---------|
| **B1** | `DefaultMode` | **A)** оставить `DIRECT` (как FS/HEAD) · **B)** сменить на `RELAY` и обновить FS §6 + docs/33 | Критично для всего decision path |
| **B2** | Постоянный UDP/443 → DIRECT (`quic_direct_bypass`) | **A)** запретить в `split` · **B)** разрешить временно с expiry · **C)** сделать продуктовой политикой + ADR | Критично для Telegram/LinkedIn/HTTP3 |
| **B3** | RELAY dial fail → DIRECT | **A)** только через `ModeFallback` semantics · **B)** всегда silent fallback | Надёжность vs прозрачность |
| **B4** | Network Mode UI | **A)** только E09/debug · **B)** E05 с пометкой «тест» · **C)** не кораблить в release | Соответствие ТЗ «одна кнопка» |
| **B5** | Builtin CIDR defaults | **A)** domain-only + точечные CIDR · **B)** разрешить широкие диапазоны (нужен список) | Риск misroute / нагрузка |

Дополнительно: создать или официально отменить `docs/02.5_BusinessRules.md`; завести BL/TASK-02 в `docs/04_Backlog.md`.

---

## План тестирования (для разработчика после ✅)

### Unit

- `MergeWithDefaults` order + first-match  
- `SetForceMode` overrides DecideDetailed  
- `parseTunnelOptions` defaults / invalid → split  
- `HostForIP` after resolve; empty miss  
- Traffic matrix sites/apps expectations  
- FALLBACK reason на relay dial fail (mock)  

### Integration / device

- DNS-in-TUN: `[dns] query … via=yandex|doh`  
- Connect log: `vpn dns=10.10.0.1`, `networkMode=split`  
- `VerifyAppSiteSwitch.ps1 -WithUnit`  

### Smoke

- `flutter test`, `go test ./...` (go_core), install APK, connect/disconnect  

---

## Документация (обновить после реализации)

| Документ | Что |
|----------|-----|
| `docs/tasks/TASK-02_*.md` | Создать |
| `docs/04_Backlog.md` | BL/TASK status |
| `docs/11_Decisions.md` | ADR-015+ |
| `docs/07_Architecture.md` | DNS-in-TUN, options, FALLBACK |
| `docs/33_DirectVsVpnBypass.md` | DNS 10.10.0.1; режимы |
| `docs/02.2_FunctionalSpecification.md` | Только если PM меняет DefaultMode/UX |
| `docs/06_TestPlan.md` | Matrix + script |
| `docs/03_CurrentState.md` / `14_AIContext.md` / `17_CHANGELOG.md` | Факт после ship |
| `ai/CurrentTask.md` | Указать TASK-02 |

Обязательно обновить после REWORK: `docs/07.4_RoutingPolicy.md` (уже SoT), `docs/11_Decisions.md` ADR-015, `docs/33`, `docs/07_Architecture.md`.

---

## Вердикт

# ✅ Разрешить разработку — только REWORK по `docs/07.4_RoutingPolicy.md`

### Запрещено мержить WIP «как есть»

Причины: transport-level product decisions (`quic_direct_bypass`, silent fallback), `DefaultMode=RELAY`, Network Mode на E05, широкие CIDR.

### Разрешено

Переработка WIP строго по ADDENDUM и [`docs/07.4_RoutingPolicy.md`](../../docs/07.4_RoutingPolicy.md):

- product path: Domain → Rule Engine → Decision → Transport;  
- QUIC→DIRECT / silent fallback — только diagnostic, не default;  
- чинить RELAY data-plane; fallback — политика правила;  
- DefaultMode=DIRECT; DNS-in-TUN оставить.

### Code Review gate

Любой PR с `if udp/443 → DIRECT` или «всегда fallback DIRECT» в product path — **reject**, ссылка на §2 политики 07.4.

---

*Конец документа.*
