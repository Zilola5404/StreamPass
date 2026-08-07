# Current Task

> Updated: 2026-08-08

## Active task

**BL-001: Реализация и техническая приёмка Hysteria2 VPN Tunnel**

### Статус

**In Progress — re-validation / hardening.**

Независимый Team Lead аудит показал, что базовая реализация BL-001 уже существует в `client/go_core` и ранее была отмечена как Done, однако текущая документация содержит противоречие: BL-001 закрыт в backlog, а доказательство полного физического Android E2E находится вне подтверждённых артефактов текущего состояния. Поэтому задача не считается окончательно принятой до прохождения критериев ниже.

### Источники требований

- `docs/02_TZ.md` — Hysteria2 как готовый transport, без разработки собственного протокола; fallback UDP 443 → 8443 → 24443 → TCP 443 → TCP 8443.
- `docs/02.2_FunctionalSpecification.md` — Connect одним действием, состояния Connecting/Connected/Error, техническая телеметрия без URL/контента.
- `docs/07_Architecture.md` — Flutter → Android VPNService → Go Core Hysteria2; `mobile/tunnel.go` — ключевой файл Go Core.
- `docs/07.4_RoutingPolicy.md` — единственный источник политики DIRECT/RELAY/FALLBACK; Transport не принимает продуктовые решения.
- `docs/08_API.md` — `GET /api/v1/servers` предоставляет клиенту relay host/port/connection_config.

### Что уже существует

- `client/go_core/mobile/tunnel.go` — lifecycle `PrepareRelay`, `StartTunnel`, `StopTunnel`, callbacks, socket protection, подключение через `hyconfig.ConnectWithFallback`.
- `client/go_core/internal/hyconfig/parse.go` — разбор `hysteria2://` / `hy2://`, auth, SNI, Salamander obfs, TLS/pin handling, protected UDP socket.
- `client/go_core/internal/hyconfig/fallback.go` — UDP 443/8443/24443 + TCP underlay 8443/24443.
- `client/go_core/internal/tunbridge/bridge.go` — TUN system stack, Hysteria TCP/UDP forwarding, Decision Engine integration, DNS path, diagnostics.
- `client/go_core/go.mod` — `github.com/apernet/hysteria/core/v2 v2.6.1`, `extras/v2 v2.6.1`, `sing`, `sing-tun`; gomobile через локальный `vendor-src/mobile`.
- Unit tests для URI parsing и fallback candidates существуют.
- Ранее подтверждались `go test`, `go build` и Android Gradle build; physical-device E2E должен быть подтверждён отдельно.

### Архитектурное решение

Новое архитектурное решение **не требуется**. Реализация должна строго оставаться в утверждённой архитектуре `docs/07_Architecture.md` и routing policy `docs/07.4_RoutingPolicy.md`.

Не разрешается передавать разработчику задачу на изменение продуктовой маршрутизации, создание собственного Hysteria transport или замену утверждённого Hysteria2 core.

### Developer hand-off

Разработчик должен:

1. Провести gap-analysis текущей реализации BL-001 против ТЗ/FS/Architecture/Routing Policy.
2. Не переписывать рабочий transport без подтверждённого дефекта.
3. Исправить только подтверждённые несоответствия и ошибки.
4. Добавить/усилить unit/integration tests для Hysteria2 config, fallback, lifecycle, protected socket и relay data path.
5. Проверить сборку Go Core и AAR.
6. Провести реальный Android E2E с live relay: Connect → VPNService/TUN → Hysteria2 → relay → data transfer.
7. Подтвердить корректный Disconnect/StopTunnel и отсутствие зависших сессий.
8. Не логировать password, `connection_config`, URL или содержимое пользовательского трафика.

### Acceptance gate

BL-001 не переводить в Done, пока не подтверждены:

- Go Core tests PASS;
- Go Core build PASS;
- AAR build PASS;
- Android build PASS;
- live relay handshake PASS;
- реальный TCP и UDP data transfer через Hysteria2 PASS;
- socket `protect()` подтверждён для relay underlay;
- fallback candidates работают в утверждённом порядке;
- Disconnect корректно закрывает TUN/Hysteria/session;
- связанные routing/Decision Engine функции не сломаны;
- QA получил задачу и подтвердил результат.

### Security gate

Пароли Hysteria и полные `connection_config` не должны попадать в Git, APK assets, публичный `/downloads/` или пользовательские логи. Тестовые секреты — только через environment/secret store. Для production relay предпочтительно использовать проверяемый TLS certificate/pin; `insecure` допускается только там, где это явно подтверждено текущей relay-конфигурацией и зафиксировано в диагностике/документации.

## Previous context

- APK: `StreamPass-v0.1.1+34-signed-arm64.apk`.
- Routing policy: `routing-policy-v1`.
- Последние изменения включали BL-017 TCP underlay fallback и последующий routing-policy rework.
- BL-003 physical-device E2E ранее оставался точкой ручной проверки; это является обязательным evidence для окончательной приёмки transport.
