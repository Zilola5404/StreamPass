# StreamPass Relay Servers

> Обновлено: 2026-08-04

## Роли серверов

| IP | Роль | Сервисы |
|----|------|---------|
| `212.43.156.33` | **Backend / API + Native Relay** | Docker: postgres, redis, backend, caddy, healthmonitor; **Hysteria2 native** UDP :443 |
| `212.43.157.167` | **VPN Relay (Hiddify)** | Hiddify Manager + Hysteria2 (sing-box), HAProxy :443 |

> `212.43.159.198` из старого ТЗ — **не используется**.

## Relay в PostgreSQL (production)

| id | region | host | port | тип |
|----|--------|------|------|-----|
| `nl-native-1` | `nl` | `212.43.156.33` | 443 | Native Hysteria (StreamPass) |
| `nl-amsterdam-1` | `nl` | `212.43.157.167` | 32528 | Hiddify / sing-box |

Канонические коды регионов (BL-026): `de` Frankfurt, `nl` Amsterdam, `pl` Warsaw, `fi` Helsinki.
Каталог: `GET /api/v1/regions`. Новые VPS регистрируются через Admin → Relays (`POST /api/v1/servers`) с `region` из каталога.

### nl-native-1 — native Hysteria на backend-сервере

Установлен скриптом `scripts/setup-relay-hysteria.sh` (UDP 443, TCP 443 занят Caddy).

Формат `connection_config` (пароли **не хранить в Git** — подставлять из env/секрет-хранилища):

```
hysteria2://<AUTH_PASSWORD>@212.43.156.33:443/?obfs=salamander&obfs-password=<OBFS_PASSWORD>
```

Для test-relay с self-signed TLS добавьте `&insecure=1` **только** в dev/staging.
В production используйте валидный сертификат или `pinSHA256=...` — без `insecure=1`.

### nl-amsterdam-1 — Hiddify relay

Hiddify поднимает Hysteria2 на UDP-портах **32527** / **32528** (не 443).

Параметры берутся из Hiddify Admin / `singbox/configs/05_inbounds_4100_hysteria.json` на сервере.

```
hysteria2://<user-uuid>@212.43.157.167.sslip.io:32528/?obfs=salamander&obfs-password=<obfs-secret>&sni=212.43.157.167.sslip.io
```

**Важно:** в PostgreSQL должна храниться ссылка `hysteria2://…`, а не URL подписки Hiddify (`https://…/uuid/`).

## Регистрация / обновление relay в БД

На backend-сервере (`212.43.156.33`):

```bash
docker exec -it streampass-postgres-1 psql -U streampass -d streampass
```

```sql
-- Native relay (156.33) — подставьте реальные пароли из env, не из Git
INSERT INTO relay_servers (id, region, host, port, healthy, connection_config, updated_at)
VALUES (
  'nl-native-1', 'nl', '212.43.156.33', 443, true,
  'hysteria2://<AUTH_PASSWORD>@212.43.156.33:443/?obfs=salamander&obfs-password=<OBFS_PASSWORD>',
  NOW()
)
ON CONFLICT (id) DO UPDATE SET connection_config = EXCLUDED.connection_config, region = EXCLUDED.region, updated_at = NOW();

-- Hiddify relay (157.167)
UPDATE relay_servers
SET host = '212.43.157.167',
    port = 32528,
    region = 'nl',
    connection_config = 'hysteria2://...',
    healthy = true,
    updated_at = NOW()
WHERE id = 'nl-amsterdam-1';
```

Или через Admin UI `/admin/` → Relays / Admin API `POST /api/v1/servers`.

## Установка native Hysteria

Скрипт: `scripts/setup-relay-hysteria.sh` — чистый StreamPass relay без Hiddify.

```bash
# На Ubuntu/Debian (root). UDP 443 — Caddy держит только TCP 443.
AUTH_PASSWORD='<secret>' OBFS_PASSWORD='<secret>' LISTEN_PORT=443 bash setup-relay-hysteria.sh
```

Переменные окружения: `AUTH_PASSWORD` (обязательно), `OBFS_PASSWORD` (обязательно), `LISTEN_PORT`, `HYSTERIA_VERSION`.

При загрузке с Windows используйте LF (не CRLF):

```powershell
$script = (Get-Content scripts/setup-relay-hysteria.sh -Raw) -replace "`r`n","`n"
$script | ssh root@212.43.156.33 "cat > /root/setup-relay-hysteria.sh && chmod +x /root/setup-relay-hysteria.sh"
```

## Локальная проверка клиента

```powershell
$env:STREAMPASS_RELAY_URI = "hysteria2://<AUTH_PASSWORD>@212.43.156.33:443/?obfs=salamander&obfs-password=<OBFS_PASSWORD>"
cd client/go_core
go test -v -timeout 1m -run TestIntegrationHysteria ./internal/hyconfig/
```

Ожидание: `hysteria handshake OK`, foreign IP ≈ IP relay.

## SSH

- Backend + native relay: `ssh root@212.43.156.33`
- Hiddify relay: `ssh root@212.43.157.167`

**Пароли и секреты не хранить в Git.** После утечки в истории git — ротировать пароли relay на сервере.
