# Probe Shield

Minimal split backend/frontend authentication stub. Runtime/project names are isolated as `probe-shield`, while the auth and 500 screens intentionally keep the original Exodus-style visual shell.

The project intentionally contains no PostgreSQL, Redis, queue, node, metrics, Telegram, subscription, OAuth, passkey, dashboard, or external panel integrations.

## Behavior

- The backend serves the frontend from `frontend/dist` at the domain root only. There is no base-path/runtime-path override.
- Auth-page branding is read from config/env and returned through the original `/api/auth/status` `branding.title` / `branding.logoUrl` contract.
- Login is checked only against plaintext values from config/env.
- Failed login attempts are delayed by `PROBE_SHIELD_AUTH_CHECK_DELAY` / `auth.authCheckDelay` and then rate-limited in memory.
- If login or password is empty, successful authentication is impossible.
- After successful authentication, `/dashboard` returns a real HTTP `500 Internal Server Error` and renders the original-style 500 page (`Something bad just happened...`) with a return-to-authentication button.
- Without a valid session cookie, `/dashboard` redirects to `/`.

## Docker Compose

All runtime settings are isolated under the `PROBE_SHIELD_*` namespace and can live in one `.env` file.

```bash
cp .env.example .env
nano .env
docker compose up -d --build
```

Default `.env`:

```dotenv
PROBE_SHIELD_CONFIG_FILE=/app/configs/probe-shield.json
PROBE_SHIELD_HTTP_ADDRESS=0.0.0.0
PROBE_SHIELD_HTTP_PORT=3000
PROBE_SHIELD_STATIC_DIR=/app/frontend/dist
PROBE_SHIELD_LOGIN=
PROBE_SHIELD_PASSWORD=
PROBE_SHIELD_SESSION_TTL=12h
PROBE_SHIELD_AUTH_CHECK_DELAY=5s
PROBE_SHIELD_RL_WINDOW=5m
PROBE_SHIELD_RL_MAX_FAILURES=10
PROBE_SHIELD_BRAND_TITLE=
PROBE_SHIELD_BRAND_LOGO_URL=
PROBE_SHIELD_RESPONSE_HEADERS_ENABLED=true
PROBE_SHIELD_RESPONSE_HEADERS_JSON=
```

Set credentials like this:

```dotenv
PROBE_SHIELD_LOGIN=admin
PROBE_SHIELD_PASSWORD=change-me
```

Leave either value empty to make successful authentication impossible.

## Branding

Branding is intentionally implemented through the same frontend contract as the original auth page, but values come from config/env instead of the database.

In `configs/probe-shield.json`:

```json
"branding": {
  "title": "",
  "logoUrl": ""
}
```

Empty values keep the original Exodus logo and title. To replace the visible name:

```json
"branding": {
  "title": "Custom Panel",
  "logoUrl": ""
}
```

The title supports the original colored text syntax:

```json
"branding": {
  "title": "{8195a3}Custom{eceddb}Panel",
  "logoUrl": ""
}
```

To replace the logo, set `logoUrl` to an absolute URL or a root-relative static file URL:

```json
"branding": {
  "title": "{8195a3}Custom{eceddb}Panel",
  "logoUrl": "https://example.com/logo.svg"
}
```

The same can be overridden from the shared `.env` file:

```dotenv
PROBE_SHIELD_BRAND_TITLE={8195a3}Custom{eceddb}Panel
PROBE_SHIELD_BRAND_LOGO_URL=https://example.com/logo.svg
```

For a local static logo inside the container, place the file into `frontend/dist` before building the Docker image, for example `frontend/dist/logo.svg`, and set:

```dotenv
PROBE_SHIELD_BRAND_LOGO_URL=/logo.svg
```


## Response headers

`configs/probe-shield.json` contains a `responseHeaders` block. When `enabled` is `true` and `headers` contains at least one item, the backend emits exactly those configured headers and does not append built-in defaults. When `enabled` is `true` and `headers` is empty or omitted, the backend emits the built-in Exodus-like baseline headers.

```json
"responseHeaders": {
  "enabled": true,
  "headers": {
    "X-Content-Type-Options": "nosniff",
    "X-XSS-Protection": "0",
    "X-Frame-Options": "SAMEORIGIN",
    "Cross-Origin-Opener-Policy": "same-origin-allow-popups",
    "Cross-Origin-Resource-Policy": "same-site",
    "Referrer-Policy": "strict-origin-when-cross-origin",
    "Strict-Transport-Security": "max-age=31536000; includeSubDomains",
    "X-Robots-Tag": "noindex, nofollow, noarchive, nosnippet, noimageindex",
    "Content-Security-Policy": "default-src 'self' *;script-src 'self' 'unsafe-inline' 'unsafe-eval' 'wasm-unsafe-eval' *;img-src 'self' data: *;connect-src 'self' *;worker-src 'self' blob: *;frame-src 'self' oauth.telegram.org *;frame-ancestors 'self' *;base-uri 'self';font-src 'self' https: data:;form-action 'self';object-src 'none';script-src-attr 'none';style-src 'self' https: 'unsafe-inline';upgrade-insecure-requests"
  }
}
```

From `.env`, full override is available through JSON:

```dotenv
PROBE_SHIELD_RESPONSE_HEADERS_ENABLED=true
PROBE_SHIELD_RESPONSE_HEADERS_JSON={"X-Content-Type-Options":"nosniff"}
```

With that example, only `X-Content-Type-Options` is emitted from the configurable response-header layer.

## Local build

```bash
cd frontend
npm ci --no-audit --no-fund
npm run start:build

cd ../backend
PROBE_SHIELD_CONFIG_FILE=../configs/probe-shield.json \
PROBE_SHIELD_STATIC_DIR=../frontend/dist \
go run .
```

## Docker build note

The container build is intentionally Go-only. It expects `frontend/dist` to already exist. The archive already contains this directory; after editing frontend sources manually, rebuild it first:

```bash
cd frontend
npm ci --no-audit --no-fund
npm run start:build
cd ..
docker compose up -d --build
```

## API compatibility checks

Auth status is wrapped like the original frontend contract expects:

```bash
curl http://127.0.0.1:3000/api/auth/status
```

A failed password attempt waits for `PROBE_SHIELD_AUTH_CHECK_DELAY` and returns `403` with the original-style error message for frontend notifications:

```bash
curl -i -X POST http://127.0.0.1:3000/api/auth/login \
  -H 'Content-Type: application/json' \
  --data '{"username":"admin","password":"wrong"}'
```

Successful login returns:

```json
{"response":{"accessToken":"..."}}
```

Health check:

```bash
curl http://127.0.0.1:3000/api/health
```

## GitHub Actions

The workflow at `.github/workflows/build.yml` runs frontend dependency install, typecheck/build, backend test/build, and container build. The Dockerfile intentionally does not run `npm install`; it copies `frontend/dist` into the final image. It does not inject application versions, create releases, or push images.
