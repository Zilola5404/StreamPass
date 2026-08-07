# QA Report — Product / Traffic Acceleration (независимая проверка)

> **Task ID:** PRODUCT-QA (BL-017, BL-035, TASK-02 routing, traffic MVP)  
> **Дата:** 2026-08-07  
> **QA Lead:** независимая проверка (без опоры на «готово» разработчика)  
> **Код на main:** `5d82fa6` (TASK-02 REWORK, `connectFlow=routing-policy-v1`, build **+33**)  
> **Prod OTA / docs:** заявлен **+25** (`tcp-underlay-v1`)

---

## 1. Общая оценка

Продукт **не готов** к передаче на следующий этап как стабильный consumer-ready ускоритель зарубежного трафика.

**Backend и CI-уровень** — в рабочем состоянии (SmokeTest, unit/integration tests green).

**Ключевая пользовательская ценность** (YouTube, Instagram, Gemini, зарубежные сайты через relay; Госуслуги без блокировки VPN) — **не подтверждена** на устройстве и **противоречит** заявлениям пользователя и частичным прогонам QA.

**Итоговый вердикт: FAIL** (см. §18).

---

## 2. Что проверено

### Фактически проверено (runtime)

| # | Проверка | Метод | Результат |
|---|----------|-------|-----------|
| R1 | Prod API health + public endpoints | `scripts/SmokeTest.ps1` | **PASS** 8/8 |
| R2 | Backend tests | `go test ./...` (backend) | **PASS** |
| R3 | go_core tests (decision, tunbridge, hyconfig, mobile) | `go test ./...` (client/go_core) | **PASS** |
| R4 | Flutter traffic/lifecycle contract | `flutter test` traffic_behavior, traffic_switch_scenarios, vpn_lifecycle | **PASS** 21 test |
| R5 | Relay TCP underlay ports | TCP connect `212.43.156.33:443/8443/24443` | **PASS** |
| R6 | Traffic path unit diagnosis | `go test -run TrafficPathDiagnosis` | **PASS** (документирует риск IP-only→DIRECT) |
| R7 | Device adb (кратковременно) | `adb devices` | **PASS** (RFGYB48SWAF), затем отключён |
| R8 | DiagnoseTrafficBlock (без VPN logs) | `scripts/DiagnoseTrafficBlock.ps1 -WithUnit` | **PASS** infra, **WARN** нет connect logs без VPN |
| R9 | Пользовательский прогон (из сессии) | `VerifyAppSiteSwitch.ps1 -AutoLaunch` | **PARTIAL** — Gemini logs OK; Instagram/Gosuslugi WARN; пользователь сообщает реальные сайты **не работают** |

### Проверено анализом кода / документов

| # | Область | Источник |
|---|---------|----------|
| A1 | Routing policy v1 (`DefaultMode=DIRECT`, HostForIP, DNS `10.10.0.1`) | `docs/07.4_RoutingPolicy.md`, `StreamPassVpnService.kt`, `bridge.go` |
| A2 | Риск IP-only без hostname → DIRECT (geo-block) | `traffic_path_test.go`, `07.4` §5.2 |
| A3 | Version drift +25 vs +33 | `build_info.dart`, `03_CurrentState.md`, `/api/v1/config` |
| A4 | BL-035 off-site backup | `27_BackupRecovery.md`, скрипты — без VPS runtime |
| A5 | BL-017 TCP underlay | hyconfig tests + открытые порты |

### Не проверено (BLOCKED / вне сессии)

| # | Что | Причина |
|---|-----|---------|
| N1 | Полный device E2E с **+33 APK** routing-policy-v1 | OTA/prod APK = старая сборка; новый AAR не задеплоен на `/downloads/` |
| N2 | LiveProbe с VPN connected | Телефон отключился от adb во время QA-сессии |
| N3 | Off-site backup cron на VPS | Нет SSH-доступа QA к prod |
| N4 | ЮKassa live billing | NOT_REQUIRED (Skipped по backlog) |
| N5 | iOS / Windows clients | NOT_REQUIRED |

---

## 3. Что работает

| Область | Статус | Доказательство |
|---------|--------|----------------|
| Auth/API/Smoke | **PASS** | SmokeTest 8/8 |
| Rules/Config/Regions public API | **PASS** | SmokeTest |
| Backend unit + integration | **PASS** | go test |
| go_core decision/DNS/tunbridge unit | **PASS** | go test |
| Relay infrastructure TCP | **PASS** | ports 443, 8443, 24443 open |
| VPN connect lifecycle (mock) | **PASS** | vpn_lifecycle_test.dart |
| TASK-02 diagnostics logging | **PASS** (код) | `[decision]`, `[diag]`, `[tun]` в bridge.go |
| BL-017 fallback (unit) | **PASS** | hyconfig tests |

---

## 4. Что не работает

| Требование (ТЗ §6 / FS §6) | Ожидание | Факт | Статус |
|----------------------------|----------|------|--------|
| YouTube / зарубежные через RELAY | Стабильный доступ через relay egress | Пользователь: **не работает**; ранее «не доступно в стране» (Gemini) | **FAIL** |
| Instagram | RELAY + CDN | Пользователь: **не работает**; QA logs WARN (DNS/RELAY не подтверждены) | **FAIL** |
| Gemini / Google AI | RELAY | Пользователь: **«недоступно в вашей стране»** (RU geo) | **FAIL** |
| Госуслуги (E08 app bypass) | Без «отключите VPN» | Пользователь: **не открывается из-за VPN** | **FAIL** |
| Prod OTA = актуальный клиент | Последняя routing-policy-v1 | OTA/config без build +33; docs +25 | **FAIL** |

---

## 5. Что реализовано частично

| Область | Статус | Комментарий |
|---------|--------|-------------|
| Foreign acceleration | **PARTIAL** | Engine/rules/DNS architecture обновлены (TASK-02), но end-user результат не подтверждён |
| App bypass (Госуслуги) | **PARTIAL** | Пакеты в `VpnBypassApps.kt`; bypass log не всегда в step logs; пользователь FAIL |
| Split DNS + HostForIP | **PARTIAL** | Требует `vpn dns=10.10.0.1` (+33); старый APK с Yandex OS DNS ломает цепочку |
| BL-035 off-site backup | **PARTIAL** | Скрипты есть; prod cron/.enc не верифицирован |
| BL-017 TCP underlay | **PARTIAL** | Порты открыты, unit OK; полный connect на device не завершён QA |
| QA tooling | **PARTIAL** | `DiagnoseTrafficBlock.ps1`, `VerifyAppSiteSwitch.ps1` добавлены, не в CI |

---

## 6. Что отсутствует

| Элемент | Статус |
|---------|--------|
| Автоматический device E2E в CI | **MISSING** |
| Published APK +33 на OTA | **MISSING** (vs codebase) |
| Code Review / Audit для TASK-02 (`5d82fa6`) | **MISSING** |
| Подтверждение off-site backup на secondary VPS | **MISSING** |
| `docs/18_KnownIssues.md` | **MISSING** (файл не найден) |

---

## 7. Найденные баги

### BUG-001

**Приоритет:** P0  
**Название:** Зарубежные сервисы (YouTube, Instagram, Gemini) недоступны / показывают geo-block при подключённом VPN.

**Требование:** ТЗ §6 (RELAY для YouTube, Google, foreign CDN); FS §6 п.4–5.

**Предусловия:** VPN Connected, подписка ACTIVE, Chrome.

**Шаги:**
1. Подключить StreamPass VPN.
2. Открыть youtube.com, instagram.com, gemini.google.com в Chrome.

**Ожидаемый результат:** Сайты загружаются через relay (foreign egress IP); ускорение/доступ работает.

**Фактический результат:** Пользователь: сайты **не работают**; Gemini — «недоступно в вашей стране» (признак RU egress / не-relay path).

**Причина (анализ + unit):** IP-only TUN flows без `HostForIP` → `DefaultMode=DIRECT` → protected direct с RU IP (`traffic_path_test.go`); либо установлен **старый APK** (`vpn dns=77.88.8.8` вместо `10.10.0.1`).

**Влияние:** Основной сценарий продукта не выполняется.

**Где обнаружено:** User report; `DiagnoseTrafficBlock.ps1` cheat sheet; go unit tests.

**Рекомендация:** Задеплоить +33 APK; проверить connect log `vpn dns=10.10.0.1`; `[decision] action=RELAY` для foreign hosts; `-LiveProbe`.

**Статус:** OPEN

---

### BUG-002

**Приоритет:** P1  
**Название:** Приложение Госуслуги блокируется VPN (E08 не работает для пользователя).

**Требование:** FS E08; ТЗ §6 DIRECT + app bypass для гос/банков.

**Шаги:** VPN ON → открыть Госуслуги (ru.rostel).

**Ожидаемый результат:** Приложение открывается без требования отключить VPN.

**Фактический результат:** Пользователь: **не открывается из-за VPN**.

**Причина:** Возможны: старый APK без расширенного bypass; `VPN app-bypass applied=0`; неверный package id на устройстве.

**Рекомендация:** `adb shell pm list packages | findstr gosuslugi`; connect log `VPN app-bypass: ru.rostel`; переподключить VPN после нового APK.

**Статус:** OPEN

---

### BUG-003

**Приоритет:** P1  
**Название:** Расхождение версии клиента: codebase +33 routing-policy-v1 vs prod/docs +25.

**Требование:** DoD §3, FS soft-update; `03_CurrentState.md`.

**Факт:** `BuildInfo.buildNumber=33`, `connectFlow=routing-policy-v1`; prod config `latest_client_version=0.1.1` без build; OTA/docs — +25.

**Влияние:** Пользователи и QA тестируют **не тот** бинарник с TASK-02 routing fixes.

**Рекомендация:** Собрать signed APK +33, обновить `/downloads/StreamPass.apk`, `03_CurrentState.md`, config API.

**Статус:** OPEN

---

### BUG-004

**Приоритет:** P2  
**Название:** UI показывает «подписка не активна» при недоступности backend (ложный negative).

**Требование:** FS error states; корректный Offline/Error.

**Факт (анализ кода, prior session):** `home_screen.dart` catch → inactive subscription on network error.

**Статус:** OPEN (частично исправлено в working tree — **не верифицировано QA**)

---

### BUG-005

**Приоритет:** P2  
**Название:** BL-035 off-site backup не подтверждён на production.

**Требование:** BL-035, `27_BackupRecovery.md`.

**Факт:** Только code review скриптов; нет `.enc` на secondary / cron proof.

**Статус:** OPEN

---

## 8. Критические блокеры (P0/P1)

| ID | Блокер | Блокирует этап |
|----|--------|----------------|
| BUG-001 | Foreign traffic / geo-block | MVP / Release |
| BUG-002 | Gosuslugi VPN block | FS E08 |
| BUG-003 | OTA version drift +25 vs +33 | Любое device QA |

---

## 9. Несоответствия ТЗ

| ТЗ / FS | Должно быть | Есть | Статус |
|---------|-------------|------|--------|
| §6 RELAY YouTube/Google/foreign | Рабочий relay egress | User FAIL | **FAIL** |
| §10 Fallback TCP 443 | В цепочке fallback | Пропущен (Caddy) — задокументировано в BL-017 QA | **PARTIAL** |
| FS §6 DefaultMode DIRECT | Да | Да — но требует HostForIP/DNS chain | **PARTIAL** |
| FS E08 app bypass | Госуслуги без VPN block | User FAIL | **FAIL** |
| DoD §2 all tests green | green | green (2026-08-07 run) | **PASS** |

---

## 10. Регрессионные проблемы

| Область | Наблюдение |
|---------|------------|
| BL-017 prior QA | go_core mobile test **FAIL** в отчёте 2026-08-06 — **исправлено** (green 2026-08-07) |
| Backend availability | Был downtime в сессии 2026-08-06 — **восстановлен** (SmokeTest PASS 2026-08-07) |
| Traffic после TASK-02 | Новая архитектура — **regression foreign access** по user report |

---

## 11. Проблемы Backend

| Проверка | Статус |
|----------|--------|
| SmokeTest | **PASS** |
| Integration tests | **PASS** |
| Rules v7 (54 rules) | **PASS** (runtime fetch) |
| Downtime resilience | Был инцидент; сейчас OK |

---

## 12. Проблемы Frontend / Mobile

| Проблема | Приоритет |
|----------|-----------|
| Foreign sites geo-block / no traffic | P0 |
| Gosuslugi VPN block | P1 |
| OTA stale APK | P1 |
| Misleading subscription UI on network error | P2 |
| flutter analyze 65 infos/warnings (no errors) | P4 |
| App task-switch stability | Не перепроверено в этой сессии (ранее fixes +25) |

---

## 13. Проблемы API

| Проверка | Статус |
|----------|--------|
| Client ↔ API contract (mock tests) | **PASS** |
| `/api/v1/config` missing build number for OTA | **PARTIAL** — только semver `0.1.1` |
| Prod недоступность → client UX | **FAIL** (BUG-004) |

---

## 14. Проблемы безопасности

| Проверка | Статус |
|----------|--------|
| Metrics blocked publicly | **PASS** (404) |
| Servers without auth → 401 | **PASS** |
| Secrets in repo | Не аудировалось в этой сессии |
| PII in connect logs | **WARN** — hostname-only policy заявлена; перепроверить на device |

---

## 15. Проблемы производительности

| Наблюдение | Статус |
|------------|--------|
| Backend loadtest | **PASS** (go test loadtest) |
| Relay blackhole detection | Код есть (`relay_blackhole`); device не проверен |
| Log spam `[tun] tcp mode` | **WARN** — возможен шум; throttle в diag |

---

## 16. Что необходимо исправить (Developer)

1. **P0:** Восстановить foreign RELAY path на device — подтвердить `vpn dns=10.10.0.1`, `[decision] action=RELAY`, отсутствие `host=` empty + DIRECT для 142.250/157.240.
2. **P1:** Gosuslugi bypass — package id на устройстве + connect log `VPN app-bypass: ru.rostel`.
3. **P1:** Опубликовать APK **+33** (`routing-policy-v1`) на OTA; обновить docs/config.
4. **P2:** BL-035 — доказать off-site cron + `.enc` на `212.43.157.167`.
5. **P2:** Network error vs inactive subscription UX — QA retest.

---

## 17. Что необходимо повторно протестировать (QA)

После исправлений:

```powershell
# 1. Deploy +33 APK
adb install -r client\build\app\outputs\flutter-apk\app-release.apk

# 2. Connect VPN, then:
.\scripts\DiagnoseTrafficBlock.ps1 -LiveProbe -ReportPath reports\QA\traffic-block-retest.md
.\scripts\VerifyAppSiteSwitch.ps1 -AutoLaunch -SkipManual -ReportPath reports\QA\traffic-switch-retest.md

# 3. Regression
.\scripts\SmokeTest.ps1
cd client\go_core && go test ./...
cd client && flutter test
```

**Pass criteria device:**
- `[decision] host=youtube.com action=RELAY` (не пустой host для foreign)
- YouTube / Instagram / Gemini открываются
- Госуслуги без VPN block
- `vpn dns=10.10.0.1 (Go dnscache)`

---

## 18. Итоговый вердикт

# FAIL

### Ответы на 4 вопроса QA Lead

| # | Вопрос | Ответ |
|---|--------|-------|
| 1 | Работает ли функция по ТЗ? | **Нет** для core value (foreign acceleration, E08) — **FAIL** |
| 2 | Не сломались ли существующие функции? | Backend/API/tests — **OK**; user-facing traffic — **регрессия/не закрыто** |
| 3 | Что не реализовано / неправильно? | End-to-end relay egress на device; Gosuslugi bypass; OTA +33; off-site proof |
| 4 | Можно ли передавать дальше? | **Нет** — есть P0/P1 (BUG-001…003) |

### Условие PASS

- BUG-001…003 закрыты с device proof (+33 APK на OTA)
- `DiagnoseTrafficBlock -LiveProbe` — no blockers
- `VerifyAppSiteSwitch` — 0 FAIL on instagram/gemini/gosuslugi/youtube
- BL-035 verified или formal risk acceptance
- Docs synced to +33

---

*Отчёт QA Lead 2026-08-07. Runtime: SmokeTest, go test, flutter test, relay TCP, частичный adb. Device full E2E BLOCKED частично (adb disconnect); user reports включены как фактическое поведение.*
