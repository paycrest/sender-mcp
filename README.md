# Paycrest MCP server

Go implementation of a [Model Context Protocol](https://modelcontextprotocol.io/) server that calls the **Paycrest aggregator** HTTP API (`/v2/...`). Use it from Cursor, Claude Desktop, or any MCP host that supports **stdio** transport.

## Install (senders) — fixed location (recommended)

Install the binary to a **stable path** so you do not rename Release assets or hand-edit long Download paths.

| OS | Install path |
|----|----------------|
| Windows | `%USERPROFILE%\.paycrest\paycrest-mcp.exe` |
| macOS / Linux | `~/.paycrest/paycrest-mcp` |

### One-line install

**Note:** If the GitHub repo is **private**, anonymous downloads fail. Friends need either:

- **GitHub CLI** logged in (`gh auth login`), then run the install from a clone / copied script, **or**
- A **public** Releases page, **or**
- Manual download of the `.exe` while logged into GitHub, then copy to the fixed path below.

**Windows (PowerShell — not Git Bash):**

```powershell
# From a clone (works with private repo if `gh auth login` succeeded):
cd path\to\sender-mcp
.\scripts\install.ps1
```

After scripts are on `main` **and** the repo is public (or you use a token):

```powershell
irm https://raw.githubusercontent.com/paycrest/sender-mcp/main/scripts/install.ps1 | iex
```

Optional: `.\scripts\install.ps1 -ApiKey "your-sender-api-key"`  
Or set user env `PAYCREST_API_KEY` first; the script will pick it up.

**macOS / Linux:**

```bash
# Preferred when repo is private (requires: gh auth login):
./scripts/install.sh
```

Public one-liner (after push to `main`):

```bash
curl -fsSL https://raw.githubusercontent.com/paycrest/sender-mcp/main/scripts/install.sh | bash
```

Optional: `PAYCREST_API_KEY=your-key bash scripts/install.sh`

The script:

1. Downloads the matching asset from [GitHub Releases](https://github.com/paycrest/sender-mcp/releases)
2. Saves it under `~/.paycrest/` with a **fixed name** (no rename by hand)
3. Merges a `paycrest` entry into `~/.cursor/mcp.json` (unless you pass `-SkipMcpJson` / `SKIP_MCP_JSON=1`)

Then **reload MCP** in Cursor (Settings → Tools & MCP) or restart Cursor.

### Cursor `mcp.json` (after install)

The install script writes something equivalent to (API key only — aggregator URL defaults to production inside the binary):

**Windows**

```json
{
  "mcpServers": {
    "paycrest": {
      "command": "C:/Users/YOU/.paycrest/paycrest-mcp.exe",
      "env": {
        "PAYCREST_API_KEY": "${env:PAYCREST_API_KEY}"
      }
    }
  }
}
```

**macOS / Linux**

```json
{
  "mcpServers": {
    "paycrest": {
      "command": "/Users/YOU/.paycrest/paycrest-mcp",
      "env": {
        "PAYCREST_API_KEY": "${env:PAYCREST_API_KEY}"
      }
    }
  }
}
```

Each sender uses **their own** dashboard API key (`PAYCREST_API_KEY`) — never share one key.

You can also use Cursor’s `${userHome}` if you prefer a portable config:

```json
"command": "${userHome}/.paycrest/paycrest-mcp.exe"
```

(omit `.exe` on macOS/Linux)

### Manual download (optional)

If you prefer not to run the script, download from **https://github.com/paycrest/sender-mcp/releases** and either:

- Save/rename into the fixed path above, **or**
- Point `command` at the downloaded file path as-is (no rename required)

| Your OS | Asset to download |
|---------|-------------------|
| Windows (x64) | `paycrest-mcp_vX.Y.Z_windows_amd64.exe` |
| macOS (Apple Silicon) | `paycrest-mcp_vX.Y.Z_darwin_arm64` |
| macOS (Intel) | `paycrest-mcp_vX.Y.Z_darwin_amd64` |
| Linux (x64) | `paycrest-mcp_vX.Y.Z_linux_amd64` |
| Linux (ARM64) | `paycrest-mcp_vX.Y.Z_linux_arm64` |

Ignore **Source code (zip/tar.gz)** on the Release page — those are not the MCP binary.

Optional: verify with `checksums.txt` on the same release.

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

Prefer the [fixed-location install](#install-senders--fixed-location-recommended) above. Manual example:

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
