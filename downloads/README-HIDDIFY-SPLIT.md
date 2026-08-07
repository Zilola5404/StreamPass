# Hiddify / sing-box temporary split profiles

## Security

- **Do not** put relay passwords in `/downloads/` on the public VPS.
- Generate a local profile from the template (or ask ops for a one-off private file).
- `file_server browse` is disabled on `/downloads/`.

## Why plain `hysteria2://` fails TZ behavior

| Link | Symptom | Cause |
|------|---------|--------|
| Hiddify `157.167:32528` full tunnel | Foreign OK, `ya.ru` no | Server rejects `geoip-ru`/`geosite-ru`; RU sites also often block VPS IPs. Need **client** DIRECT for RU. |
| Native `156.33:443` full tunnel | Often broken foreign | VPS has **no IPv6**; clients sending AAAA get `no IPv4 address available`. Need **IPv4-only DNS** + split. |

Go backend (`backend/`) does **not** route packets — only Hysteria/Hiddify on the VPS does.

## Working approach (TZ)

RU → `direct`, foreign → `proxy`, DNS `ipv4_only`, rule-sets that **exist** (HTTP 200):

- geoip: `https://raw.githubusercontent.com/SagerNet/sing-geoip/rule-set/geoip-ru.srs`
- geosite: `https://raw.githubusercontent.com/SagerNet/sing-geosite/rule-set/geosite-category-ru.srs`

(Local mirrors after deploy: `https://212-43-156-33.nip.io/downloads/rules/…`)

Do **not** use `Chocolate4U/.../geosite-ru.srs` — **404**.

## Template

See `hiddify-split.template.json` — placeholders `__HY2_SERVER__`, `__HY2_PORT__`, `__HY2_PASSWORD__`, `__HY2_OBFS__`, `__HY2_SNI__`.
