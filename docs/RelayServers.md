# StreamPass Relay Servers

> Обновлено: 2026-08-03

## Роли серверов

| IP | Роль | Сервисы |
|----|------|---------|
| `212.43.156.33` | **Backend / API** | Docker: postgres, redis, backend, caddy, healthmonitor |
| `212.43.157.167` | **VPN Relay** | Hiddify Manager + Hysteria2 (sing-box), HAProxy :443 |

> `212.43.159.198` из старого ТЗ — **не используется**.

## Hysteria2 на relay (212.43.157.167)

Hiddify поднимает Hysteria2 на UDP-портах **32527** / **32528** (не 443).

Параметры берутся из Hiddify Admin / `singbox/configs/05_inbounds_4100_hysteria.json` на сервере.

Формат `connection_config` для go_core:

```
hysteria2://<user-uuid>@212.43.157.167.sslip.io:32528/?obfs=salamander&obfs-password=<obfs-secret>&sni=212.43.157.167.sslip.io&insecure=1
```

**Важно:** в PostgreSQL должна храниться ссылка `hysteria2://…`, а не URL подписки Hiddify (`https://…/uuid/`).

## Регистрация / обновление relay в БД

На backend-сервере (`212.43.156.33`):

```bash
docker exec -it streampass-postgres-1 psql -U streampass -d streampass
```

```sql
UPDATE relay_servers
SET host = '212.43.157.167',
    port = 32528,
    connection_config = 'hysteria2://...',  -- см. Hiddify
    healthy = true,
    updated_at = NOW()
WHERE id = 'nl-amsterdam-1';
```

Или через Admin API (`POST /api/v1/admin/relays`).

## Локальная проверка клиента

```powershell
$env:STREAMPASS_RELAY_URI = "hysteria2://..."   # из Hiddify / БД
cd client/go_core
go test -v -timeout 1m -run TestIntegrationHysteria ./internal/hyconfig/
```

Ожидание: `hysteria handshake OK`, foreign IP ≈ IP relay.

## SSH

- Backend: `ssh root@212.43.156.33`
- Relay: `ssh root@212.43.157.167`

**Пароли и секреты Hiddify не хранить в Git.** Ротация паролей root после передачи в чат.

## Установка native Hysteria (альтернатива Hiddify)

Скрипт: `scripts/setup-relay-hysteria.sh` — для чистого StreamPass relay без Hiddify.
