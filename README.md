# Paycrest MCP server

Go implementation of a [Model Context Protocol](https://modelcontextprotocol.io/) server that calls the **Paycrest aggregator** HTTP API (`/v2/...`). Use it from Cursor, Claude Desktop, or any MCP host that supports **stdio** transport.

**Users / senders:** see [docs/USER_GUIDE.md](docs/USER_GUIDE.md) for install and how to use in Cursor.

## Install (senders)

End users: follow the full walkthrough in **[docs/USER_GUIDE.md](docs/USER_GUIDE.md)** (PowerShell / curl installers, `mcp.json`, smoke tests, order flow, troubleshooting).

Quick reference — fixed install paths:

| OS | Install path |
|----|----------------|
| Windows | `%USERPROFILE%\.paycrest\paycrest-mcp.exe` |
| macOS / Linux | `~/.paycrest/paycrest-mcp` |

**Windows (PowerShell — not Git Bash):**

```powershell
irm https://raw.githubusercontent.com/paycrest/sender-mcp/main/scripts/install.ps1 | iex
```

**macOS / Linux:**

```bash
curl -fsSL https://raw.githubusercontent.com/paycrest/sender-mcp/main/scripts/install.sh | bash
```

Set `PAYCREST_API_KEY` in `~/.cursor/mcp.json`, then reload MCP. Production API URL is the built-in default — no base URL needed for senders.

Releases: https://github.com/paycrest/sender-mcp/releases (ignore **Source code** zip/tar.gz; use the platform binary assets).

### Maintainers — publish a release

From this repo (after changes are on `main`):

```bash
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions (`.github/workflows/release.yml`) runs tests, builds all platform binaries, and attaches them to the release for that tag.

## Requirements

- **Senders using Releases:** no Go install required
- **Building from source:** Go **1.25+** (required by [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk))

## Layout

- `main.go` — stdio MCP entrypoint
- `config/` — env loading (`Load` → `types.Config`)
- `paycrest/` — HTTP client for the aggregator
- `mcpserver/` — MCP tools + prompts, create-order helpers, order watch + MCP progress
- `.cursor/rules/` — Cursor rule (canonical under **`sender-mcp`**; duplicate at repo **`.cursor/rules/`** when workspace is `paycrest` root so rules load)
- `scripts/install.ps1` / `scripts/install.sh` — install latest Release binary to `~/.paycrest/` and update Cursor `mcp.json`
- `docs/USER_GUIDE.md` — end-user install and Cursor usage guide
- `types/` — shared structs (config, tool inputs, provider account DTOs)

## Environment

| Variable | Where to set | Required | Description |
|----------|----------------|----------|-------------|
| `PAYCREST_BASE_URL` | Optional `.env` (developers only) | No | Aggregator base URL. **Default: `https://api.paycrest.io`**. Senders should **not** set this in Cursor. Use e.g. `http://127.0.0.1:8080` only for a local aggregator. |
| `PAYCREST_API_KEY` | **MCP host `env` only** (per user) | For sender tools | Sender API key, sent as `API-Key`. Each integrator uses their own key — **do not** commit it in `.env` in shared repos. |
| `PAYCREST_HTTP_TIMEOUT` | `.env` or host `env` | No | Client timeout (Go duration, e.g. `30s`). Default: `60s` |
| `PAYCREST_MAX_RESPONSE_BYTES` | `.env` or host `env` | No | Max response body read (default `10485760`) |
| `PAYCREST_AUTO_WATCH_AFTER_CREATE` | `.env` or host `env` | No | **`true`/`1`/`yes`/`on`** enables embedded status poll after `paycrest_create_order` **201** (15s interval, 429 backoff, MCP progress when supported). **Default off** — create returns in ~seconds with a manual `paycrest_watch_sender_order` hint. |
| `PAYCREST_CREATE_ORDER_MAX_WAIT_SEC` | `.env` or host `env` | No | Max seconds for that embedded poll (clamped **10–3600**). Default **900**. |
| `PAYCREST_WATCH_RETURN_ON_STATUS_CHANGE` | `.env` or host `env` | No | Default **off**. Set **`true`** only if you want the watch tool to return on each non-terminal status change (agent must re-call — can force multiple Run clicks). Leave unset for **one** poll after “paid” until terminal. |

### `.env` file (local, optional)

Use `.env` for **non-secret** defaults such as `PAYCREST_BASE_URL` when running from the `sender-mcp/` directory.

1. Copy: `cp .env.example .env` (or copy `sender-mcp/.env.example` to `sender-mcp/.env` on Windows).
2. Set `PAYCREST_BASE_URL` if you want a local default.
3. **`PAYCREST_API_KEY` is intentionally omitted** from `.env.example` — agents and CI should not rely on a shared key file. Each user configures their key in **Cursor / Claude / etc.** under the server’s `env` block.

The server calls `godotenv.Load()` at startup; missing `.env` is ignored. Host-provided environment variables override `.env`.

**Git:** only `.env.example` is tracked; `.env` is gitignored. Binaries belong on [Releases](https://github.com/paycrest/sender-mcp/releases), not in the repo tree.

## Build (from source)

```bash
cd sender-mcp
go build -o paycrest-mcp .
```

## Cursor (example)

Prefer the [Install (senders)](#install-senders) section or [docs/USER_GUIDE.md](docs/USER_GUIDE.md). Manual example:

Each user adds their own `PAYCREST_API_KEY` in **their** MCP config (e.g. user `~/.cursor/mcp.json`). Do not commit real keys in project repos.

```json
{
  "mcpServers": {
    "paycrest": {
      "command": "${userHome}/.paycrest/paycrest-mcp.exe",
      "env": {
        "PAYCREST_API_KEY": "${env:PAYCREST_API_KEY}"
      }
    }
  }
}
```

On macOS/Linux use `"${userHome}/.paycrest/paycrest-mcp"` (no `.exe`).

Omit `PAYCREST_API_KEY` if you only use public tools (`paycrest_get_currencies`, `paycrest_get_rates`, etc.).

Optional keys in the same `env` object: `PAYCREST_AUTO_WATCH_AFTER_CREATE` (`true` to poll inside create; default fast/off), `PAYCREST_CREATE_ORDER_MAX_WAIT_SEC`, etc.

## Tools (MVP)

| Tool | HTTP |
|------|------|
| `paycrest_get_currencies` | `GET /v2/currencies` |
| `paycrest_get_institutions` | `GET /v2/institutions/{currency_code}` |
| `paycrest_get_tokens` | `GET /v2/tokens` |
| `paycrest_get_pubkey` | `GET /v2/pubkey` |
| `paycrest_get_rates` | `GET /v2/rates/{network}/{token}/{amount}/{fiat}` |
| `paycrest_create_order` | `POST /v2/sender/orders` (body = V2 payment order JSON) |
| `paycrest_list_sender_orders` | `GET /v2/sender/orders` |
| `paycrest_watch_sender_order` | Polls `GET /v2/sender/orders/{id}` until terminal status or `max_wait_sec` (default **3600** s if omitted; MCP progress when supported) |
| `paycrest_get_sender_order` | `GET /v2/sender/orders/{id}` (single snapshot; prefer watch after payment) |

Responses are returned as **JSON text** (the raw HTTP body from the API). HTTP status `>= 400` is surfaced as an MCP tool error (`isError`).

## MCP prompts

Hosts (e.g. Cursor) may list these under **MCP → Prompts**. They do **not** add a Run button under a tool result card; they are a separate, repeatable way to pre-fill a user message after payment.

| Prompt name | Purpose |
|-------------|---------|
| `paycrest_after_payment_watch` (**After payment — watch order**) | Argument `order_id` = `data.id` from create_order. Inserts text so the agent calls `paycrest_watch_sender_order` with that id. |

## Notes

- **`paycrest_create_order` (HTTP 201):** By default (**`PAYCREST_AUTO_WATCH_AFTER_CREATE`** off) the tool **returns immediately** with pay-in details and a hint. For embedded polling, set **`PAYCREST_AUTO_WATCH_AFTER_CREATE=true`**. Cursor rules: **`sender-mcp/.cursor/rules/paycrest-mcp-order-flow.mdc`** (canonical) and **`.cursor/rules/paycrest-mcp-order-flow.mdc`** at repo root (duplicate so rules apply when the workspace is **`paycrest`**). After create, wait ~**10s** then **`paycrest_watch_sender_order`**; when the user says they **sent funds**, **call `paycrest_watch_sender_order` immediately** with the order id — do not only suggest manual checks.
- This binary does **not** add routes to the aggregator; it is an HTTP client only.
- Sender tools require the same credentials as other non-web API clients (`API-Key` header).
- For HMAC-based sender auth without an API key, extend this server or use the documented HMAC flow separately (not implemented here).
