# StreamPass — API Reference

> Base URL: `https://212-43-156-33.nip.io/api/v1` (production) | `http://localhost:8080/api/v1` (local)  
> Дата: 2026-08-03 | Версия: v1

---

## Общие сведения

- Все business endpoints: prefix `/api/v1`
- Content-Type: `application/json`
- Timestamps: RFC 3339 (`2006-01-02T15:04:05Z07:00`)
- Rate limiting: 20 req/min на public endpoints (register, login, webhook)

### Формат ошибок

```json
{
  "error": {
    "code": "invalid_credentials",
    "message": "invalid email or password",
    "details": {}
  }
}
```

| HTTP Status | Code examples |
|-------------|---------------|
| 400 | `invalid_input`, `rule_set_invalid` |
| 401 | `unauthorized`, `invalid_credentials`, `token_expired`, `token_invalid` |
| 402 | `subscription_expired`, `payment_failed` |
| 403 | `forbidden` |
| 404 | `not_found` |
| 409 | `already_exists`, `conflict` |
| 429 | `rate_limited` |
| 500 | `internal` |

### Authentication

| Тип | Header | Endpoints |
|-----|--------|-----------|
| Public | — | /health, /rules, /config, /register, /login, /logout, /payments/webhook |
| Bearer JWT | `Authorization: Bearer <access_token>` | /servers, /telemetry, /payments, /subscription/* |
| Admin Key | `X-Admin-Key: <ADMIN_API_KEY>` | /rules POST, /config POST, /servers/*, /users |

---

## Endpoints

### GET /health

| | |
|---|---|
| **URL** | `/health` или `/api/v1/health` |
| **Auth** | Public |
| **Response 200** | `{"status":"ok"}` |

---

### POST /api/v1/register

| | |
|---|---|
| **Назначение** | Регистрация + автоматический login |
| **Auth** | Public (rate limited) |
| **Request** | `{"email":"user@example.com","password":"secret123"}` |
| **Response 201** | Token pair (см. login) |
| **Errors** | 400 invalid_input, 409 already_exists, 429 rate_limited |

---

### POST /api/v1/login

| | |
|---|---|
| **Назначение** | Авторизация |
| **Auth** | Public (rate limited) |
| **Request** | `{"email":"user@example.com","password":"secret123"}` |
| **Response 200** | |
```json
{
  "access_token": "eyJ...",
  "access_expires_at": "2026-08-03T16:30:00Z",
  "refresh_token": "abc...",
  "refresh_expires_at": "2026-09-02T16:15:00Z"
}
```
| **Errors** | 401 invalid_credentials, 429 rate_limited |

---

### POST /api/v1/logout

| | |
|---|---|
| **Назначение** | Revoke refresh token |
| **Auth** | Public |
| **Request** | `{"refresh_token":"abc..."}` |
| **Response 204** | No content |
| **Errors** | 400 invalid_input |

---

### GET /api/v1/rules

| | |
|---|---|
| **Назначение** | Последний набор правил маршрутизации |
| **Auth** | Public |
| **Response 200** | |
```json
{
  "version": 1,
  "rules": [
    {"kind": "domain", "pattern": "*.ru", "mode": "direct"},
    {"kind": "domain", "pattern": "youtube.com", "mode": "relay"}
  ],
  "created_at": "2026-08-01T10:00:00Z"
}
```

---

### POST /api/v1/rules

| | |
|---|---|
| **Назначение** | Publish новый набор правил (version++) |
| **Auth** | X-Admin-Key |
| **Request** | `{"rules":[{"kind":"domain","pattern":"*.ru","mode":"direct"}]}` |
| **Response 201** | Rule set (как GET) |
| **Errors** | 401 unauthorized, 400 rule_set_invalid |

---

### GET /api/v1/config

| | |
|---|---|
| **Назначение** | Динамическая конфигурация клиента |
| **Auth** | Public |
| **Response 200** | |
```json
{
  "version": 1,
  "min_supported_client_version": "0.1.0",
  "telemetry_enabled": true,
  "rule_poll_interval_sec": 86400,
  "relay_poll_interval_sec": 300,
  "updated_at": "2026-08-01T10:00:00Z"
}
```

---

### POST /api/v1/config

| | |
|---|---|
| **Назначение** | Publish новую конфигурацию |
| **Auth** | X-Admin-Key |
| **Request** | Fields from config response |
| **Response 201** | Config object |

---

### GET /api/v1/servers

| | |
|---|---|
| **Назначение** | Доступные relay-серверы для клиента (healthy only) |
| **Auth** | Bearer JWT |
| **Response 200** | Array of servers |
```json
[
  {
    "id": "de-frankfurt-1",
    "region": "de",
    "host": "212.43.159.198",
    "port": 443,
    "healthy": true,
    "load_ratio": 0.3,
    "rtt_ms": 68,
    "connection_config": "hysteria2://..."
  }
]
```
| **Errors** | 401 unauthorized |

---

### GET /api/v1/servers/all

| | |
|---|---|
| **Назначение** | Все relay (включая unhealthy) — для healthmonitor |
| **Auth** | X-Admin-Key |
| **Response 200** | Array (как GET /servers) |

---

### POST /api/v1/servers

| | |
|---|---|
| **Назначение** | Register relay server |
| **Auth** | X-Admin-Key |
| **Request** | `{"id":"...","region":"de","host":"...","port":443,"connection_config":"hysteria2://..."}` |
| **Response 201** | Server object |

---

### DELETE /api/v1/servers/{id}

| | |
|---|---|
| **Назначение** | Удалить relay |
| **Auth** | X-Admin-Key |
| **Response 204** | No content |
| **Errors** | 404 not_found |

---

### POST /api/v1/servers/health

| | |
|---|---|
| **Назначение** | Record health check result |
| **Auth** | X-Admin-Key |
| **Request** | `{"id":"...","healthy":true,"load_ratio":0.3,"rtt_ms":68}` |
| **Response 204** | No content |

---

### POST /api/v1/telemetry

| | |
|---|---|
| **Назначение** | Отправка технических метрик (без PII) |
| **Auth** | Bearer JWT |
| **Request** | |
```json
{
  "rtt_ms": 68,
  "packet_loss_pct": 0.1,
  "relay_id": "de-frankfurt-1",
  "client_version": "0.1.0",
  "os": "android",
  "connect_ms": 3200,
  "error_code": ""
}
```
| **Response 204** | No content |

---

### POST /api/v1/payments

| | |
|---|---|
| **Назначение** | Создать платёж ЮKassa |
| **Auth** | Bearer JWT |
| **Response 201** | `{"confirmation_url":"https://yoomoney.ru/..."}` |
| **Errors** | 402 payment_failed, 401 unauthorized |

---

### POST /api/v1/payments/webhook

| | |
|---|---|
| **Назначение** | Webhook от платёжного провайдера |
| **Auth** | Public (rate limited) |
| **Request** | `{"provider_payment_id":"..."}` |
| **Response 204** | No content |
| **Note** | Backend re-fetches status from provider, не доверяет body |

---

### GET /api/v1/subscription

| | |
|---|---|
| **Назначение** | Статус подписки пользователя |
| **Auth** | Bearer JWT |
| **Response 200** | `{"status":"active","active_until":"2026-09-03T00:00:00Z"}` |
| **Status values** | `active`, `expired`, `none` |

---

### POST /api/v1/subscription/cancel

| | |
|---|---|
| **Назначение** | Отмена подписки (немедленное прекращение) |
| **Auth** | Bearer JWT |
| **Response 204** | No content |

---

### GET /api/v1/users

| | |
|---|---|
| **Назначение** | Список пользователей (admin) |
| **Auth** | X-Admin-Key |
| **Response 200** | Array of user summaries |
```json
[
  {
    "id": "usr_...",
    "email": "user@example.com",
    "created_at": "2026-08-01T10:00:00Z",
    "subscription_active_until": "2026-09-03T00:00:00Z",
    "subscription_active": true
  }
]
```

---

## Source

Маршруты: `backend/internal/infrastructure/http/router/router.go`  
Handlers: `backend/internal/infrastructure/http/handler/`
