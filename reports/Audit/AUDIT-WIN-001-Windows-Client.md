# AUDIT-WIN-001 — Аудит Windows Client

**Дата:** 2026-08-15  
**Репозиторий:** `Zilola5404/StreamPass`  
**Ветка:** `main`  
**Тип:** независимый архитектурный и readiness-аудит

---

# 1. Итоговый вердикт

## ❌ Windows Client пока нельзя считать готовым к QA или production.

В репозитории присутствует реальная реализация Windows-sidecar, Flutter Windows UI, IPC, Wintun/TUN integration, Decision Engine/DNS integration и исправления, направленные на устранение routing loop.

Однако ключевой критерий продукта — реальная передача пользовательского трафика Windows-приложений через DIRECT/RELAY — в репозитории не подтверждён полноценным E2E-тестом.

Текущий статус: **Engineering / Device Validation**, а не Production Ready.

---

# 2. Что реально реализовано

Проверено наличие Windows-specific реализации, включая:

- Windows Go sidecar;
- локальный IPC;
- start/stop/update rules/ping;
- Hysteria client integration;
- Decision Engine;
- DNS cache / HostForIP integration;
- TUN bridge;
- Flutter Windows client integration;
- `traffic_ready` state;
- попытку исключения tunnel interfaces при выборе underlay;
- cleanup default route;
- verification script для Windows.

**Вердикт:** реализация существует и не является только UI/mock.

---

# 3. Что сделано хорошо

## 3.1. Разделение Hysteria handshake и traffic readiness

Использование состояния `traffic_ready` — правильное решение. Сам факт Hysteria Connected не должен считаться доказательством работающего пользовательского трафика.

## 3.2. Защита от routing loop

Последние изменения пытаются исключить Wintun/StreamPass и другие виртуальные интерфейсы при выборе физического underlay для Hysteria. Направление корректное.

## 3.3. Recovery от stale default route

Есть логика очистки default route после аварийного завершения/старта. Это важная защита от полного blackhole Windows traffic.

## 3.4. Общий Go Core

Decision/DNS/transport logic вынесены в общий Go слой, что соответствует цели минимизировать платформенную дубликацию.

---

# 4. Главный P0 — не доказан реальный Windows data plane

Verification script проверяет build/unit/integration/IPC и наличие Wintun logic. Live Wintun test требует elevated Administrator execution.

Даже успешное создание TUN не доказывает:

```text
Windows application
  -> TUN
  -> Decision
  -> DIRECT / RELAY
  -> Internet
  -> response
```

Необходим отдельный физический E2E на Windows.

Обязательные доказательства:

- DIRECT browser traffic;
- RELAY browser traffic;
- DNS resolution;
- HostForIP;
- Decision result;
- first byte;
- bytes transferred;
- disconnect/reconnect;
- Relay unavailable behavior.

**Статус: BLOCKER.**

---

# 5. P0 — архитектурное расхождение WFP vs Wintun

Проектная документация определяет Windows Platform Adapter через Windows Filtering Platform (WFP), тогда как фактическая текущая реализация Windows построена вокруг Wintun/TUN.

Это не доказательство технической непригодности Wintun, но это прямое расхождение реализации и утверждённой архитектурной документации.

Необходимо принять одно официальное решение:

### Вариант A

Утвердить Wintun для Windows MVP и оформить ADR/изменение архитектуры.

### Вариант B

Перевести Windows interception на WFP согласно текущему ТЗ.

До принятия решения нельзя считать архитектуру согласованной.

**Статус: BLOCKER.**

---

# 6. P0 — конфликт DefaultMode в документации

В разных документах проекта зафиксированы противоречащие значения DefaultMode: встречается `DIRECT`, а в Routing Policy — `RELAY`.

Это опасно, поскольку реализация Windows может выбрать неверную политику для unknown traffic.

Необходимо установить Single Source of Truth и синхронизировать:

- TZ;
- Functional Specification;
- CurrentState;
- Architecture;
- Routing Policy;
- Windows task;
- Go implementation.

**Статус: BLOCKER.**

---

# 7. P0 — `insecure=1` нельзя оставлять в production

В текущем направлении реализации присутствует предупреждение о `insecure=1` и необходимости TLS certificate validation/pinning перед production.

Для production relay transport должен использовать валидную TLS/security policy без отключения проверки сертификата.

**Статус: BLOCKER для production.**

---

# 8. P1 — слишком жёсткая очистка route

Cleanup содержит привязку к конкретному TUN gateway (`10.10.0.2`). Это хрупко.

Правильнее хранить routes, созданные конкретной connection session, и удалять именно их по interface/index/ownership.

Нельзя полагаться только на фиксированный gateway.

---

# 9. P1 — выбор physical underlay основан на эвристике

Список исключений вроде Wintun/WireGuard/TAP/WSL/Hyper-V полезен как защита, но имя интерфейса не является надёжным источником истины.

Рекомендуется выбирать underlay через фактический Windows default route/interface index и затем проверять, что выбранный interface не принадлежит StreamPass.

---

# 10. Риск ложного PASS

Automated verification может показывать PASS для unit/contract/build/IPC проверок, хотя реальный browser traffic ещё не проверен.

В QA отчётах необходимо разделять:

- Unit PASS;
- Integration PASS;
- Build PASS;
- Live Device NOT RUN/PASS;
- Browser E2E NOT RUN/PASS;
- Relay E2E NOT RUN/PASS.

Единый PASS без указания уровня тестирования недопустим.

---

# 11. Что пока НЕ доказано

| Проверка | Статус |
|---|---|
| Windows build | Реализовано/проверяется |
| IPC | Реализовано |
| Wintun code | Реализовано |
| Live Wintun | Не доказано в репозитории |
| DIRECT real traffic | Не доказано |
| RELAY real traffic | Не доказано |
| Split routing | Не доказано E2E |
| DNS E2E | Не доказано на физическом Windows |
| HostForIP E2E | Не доказано |
| Browser E2E | Не доказано |
| IPv6 E2E | Не доказано |
| MTU E2E | Не доказано |
| Relay-down behavior | Не доказано |
| Production TLS | Не готово |

---

# 12. Обязательный E2E сценарий перед QA

На физическом Windows 10/11 устройстве:

## DIRECT

Проверить:

- `ya.ru`;
- `2ip.ru`;
- `gosuslugi.ru`;
- `s7.ru`.

Ожидание: `Decision=DIRECT`, реальный response и bytes > 0.

## RELAY

Проверить утверждённые relay destinations, например:

- Google;
- YouTube;
- GitHub;
- Instagram.

Ожидание:

```text
Decision=RELAY
Relay dial=success
first_byte > 0
bytes_transferred > 0
```

## DNS

Проверить наличие:

```text
DNS_QUERY
DNS_RESPONSE
HOST_INDEXED
HOST_LOOKUP
DECISION
```

`host=` не должен систематически оставаться пустым.

## Relay Down

Отключить Relay и убедиться, что DIRECT ресурсы продолжают работать.

---

# 13. Что запрещено делать до E2E

Не добавлять новые списки доменов/CIDR как workaround.

Не расширять Cloudflare/CDN CIDR без архитектурного решения.

Не менять fallback только для получения визуального Connected.

Не считать Hysteria handshake доказательством работающего трафика.

Не скрывать отсутствие live E2E за Unit PASS.

---

# 14. Оценка

### Архитектура: **72/100**

Сильные стороны:

- общий Go Core;
- отдельные Decision/Transport/DNS уровни;
- попытка защитить underlay от routing loop;
- traffic readiness state.

Слабые стороны:

- WFP/Wintun conflict;
- DefaultMode documentation conflict;
- production TLS blocker;
- отсутствие подтверждённого Windows E2E;
- хрупкий route cleanup;
- эвристический выбор physical interface.

### Готовность Windows Client: **45/100**

Основная причина низкой оценки — отсутствие доказательства главного пользовательского сценария: реальный Windows traffic через DIRECT/RELAY.

---

# 15. Definition of Done для закрытия аудита

- [ ] Принято архитектурное решение WFP или Wintun.
- [ ] Устранён конфликт DefaultMode.
- [ ] Убран `insecure=1` из production path.
- [ ] Live Wintun test пройден на физическом Windows.
- [ ] DIRECT E2E пройден.
- [ ] RELAY E2E пройден.
- [ ] DNS E2E пройден.
- [ ] HostForIP E2E пройден.
- [ ] Split routing E2E пройден.
- [ ] Relay-down scenario пройден.
- [ ] Reconnect/Disconnect пройден.
- [ ] IPv4/IPv6 policy подтверждена тестами.
- [ ] MTU проверен.
- [ ] Логи показывают фактический Decision и first byte.
- [ ] QA отчёт разделяет automated и real-device results.

---

# 16. Финальный вердикт

## ❌ Вернуть на Engineering Validation

Windows Client не следует передавать в полноценный QA и тем более выпускать в production до закрытия P0.

Текущая реализация является рабочей основой, но её ключевой data-plane результат пока не доказан.

Главный следующий шаг — не добавление новых routing rules, а **физический Windows E2E, который докажет DIRECT → Internet и RELAY → Internet на реальном устройстве**.
