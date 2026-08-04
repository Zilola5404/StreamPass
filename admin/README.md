# StreamPass Admin Panel (BL-020)

Статический UI для операторов. Авторизация — заголовок `X-Admin-Key`.

## URL

```text
https://<host>/admin/
```

## Возможности

| Вкладка | API |
|---------|-----|
| Health | `GET /health`, `GET /servers/all` |
| Users | `GET /users` |
| Relays | `GET /servers/all`, `POST /servers`, `DELETE /servers/{id}` |
| Rules | `GET /rules`, `POST /rules` |
| Config | `GET /config`, `POST /config` |

Ключ только в `sessionStorage` вкладки.

## Deploy

`docker-compose` монтирует `./admin` в Caddy (`/srv/admin`). После pull на сервере:

```bash
docker compose up -d caddy
```
