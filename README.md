# ProbeShield

Minimal split frontend/backend auth stub.

## Behavior

- Serves only the authentication frontend and a fake post-login failure page.
- No PostgreSQL, Redis, Prisma, queues, nodes, metrics, OAuth, Telegram, passkey, or registration flow.
- If `auth.login` or `auth.password` is empty, successful authentication is impossible.
- Failed authentication is delayed by a hardcoded `5s` before returning `403`.
- Session TTL is hardcoded to `1h`.
- After successful authentication the frontend calls `GET /api/dashboard`, which returns real `500`.
- `GET /dashboard` with a valid cookie also returns real `500` and renders the original-style 500 page.

## Configuration

The config file is fixed as:

```text
configs/config.json
```

In Docker it is copied to:

```text
/app/configs/config.json
```

There is no `PROBE_SHIELD_CONFIG_FILE` and no `PROBE_SHIELD_STATIC_DIR`. Static files are resolved internally from the fixed project/container layout.

Example:

```json
{
  "server": {
    "address": "0.0.0.0",
    "port": 4000
  },
  "auth": {
    "login": "admin",
    "password": "admin"
  },
  "rateLimit": {
    "window": "5m",
    "maxFailedAttempts": 10
  },
  "branding": {
    "title": "{8195a3}Probe{eceddb}Shield",
    "description": "Authentication",
    "logoUrl": ""
  },
  "responseHeaders": {
    "enabled": true,
    "headers": {}
  }
}
```

`branding.title` is the single source for both the visible auth-page name and browser `<title>`. Color tags like `{8195a3}` are stripped only for the browser title.

`branding.description` is used as `<meta name="description">`.

`branding.logoUrl` may be an external URL or a root-relative static URL such as `/logo.svg`.

## Environment overrides

Only these variables are supported:

```env
PROBE_SHIELD_HTTP_ADDRESS=0.0.0.0
PROBE_SHIELD_HTTP_PORT=4000

PROBE_SHIELD_LOGIN=
PROBE_SHIELD_PASSWORD=

PROBE_SHIELD_RL_WINDOW=5m
PROBE_SHIELD_RL_MAX_FAILURES=10

PROBE_SHIELD_BRAND_TITLE={8195a3}Probe{eceddb}Shield
PROBE_SHIELD_BRAND_DESCRIPTION=Authentication
PROBE_SHIELD_BRAND_LOGO_URL=

PROBE_SHIELD_RESPONSE_HEADERS_ENABLED=true
PROBE_SHIELD_RESPONSE_HEADERS_JSON=

IS_TELEGRAM_NOTIFICATIONS_ENABLED=true
TELEGRAM_BOT_TOKEN=
TELEGRAM_BOT_PROXY=
TELEGRAM_NOTIFY_SERVICE=
```

Hardcoded and intentionally not configurable:

```text
session TTL: 1h
auth check delay: 5s
static dir: /app/frontend/dist in Docker, ../frontend/dist for local backend run
config path: configs/config.json or /app/configs/config.json
```

## Response headers

If `responseHeaders.enabled=false`, the backend adds no configurable security headers.

If `responseHeaders.enabled=true` and `responseHeaders.headers` is empty, the backend adds its built-in baseline headers.

If `responseHeaders.enabled=true` and `responseHeaders.headers` contains at least one header, only the configured headers are added; built-in headers are not merged.

## Docker

```bash
docker compose up -d --build
```

## Local run

Build frontend first:

```bash
cd frontend
npm ci --no-audit --no-fund
npm run start:build
```

Run backend:

```bash
cd ../backend
go run .
```

## API checks

```bash
curl -i http://127.0.0.1:4000/api/auth/status
curl -i http://127.0.0.1:4000/api/auth/me
```

Wrong password returns `403` after about 5 seconds.

Successful login returns `200` and sets `probe_shield_session` cookie. Then:

```bash
curl -i -b cookies.txt http://127.0.0.1:4000/api/dashboard
curl -i -b cookies.txt http://127.0.0.1:4000/dashboard
```

Both return `500` when authenticated.
