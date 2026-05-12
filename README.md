# Paycrest MCP server



Go implementation of a [Model Context Protocol](https://modelcontextprotocol.io/) server that calls the **Paycrest aggregator** HTTP API (`/v2/...`). Use it from Cursor, Claude Desktop, or any MCP host that supports **stdio** transport.



## Requirements



- Go **1.25+** (required by [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk))



## Layout



- `main.go` — stdio MCP entrypoint  
- `config/` — env loading (`Load` → `types.Config`)  
- `paycrest/` — HTTP client for the aggregator  
- `mcpserver/` — MCP tool registration and create-order response helpers  
- `types/` — shared structs (config, tool inputs, provider account DTOs)



## Environment



| Variable | Where to set | Required | Description |

|----------|----------------|----------|-------------|

| `PAYCREST_BASE_URL` | `.env` and/or MCP host `env` | No | Aggregator base URL (no trailing path). Default: `http://127.0.0.1:8080` |

| `PAYCREST_API_KEY` | **MCP host `env` only** (per user) | For sender tools | Sender API key, sent as `API-Key`. Each integrator uses their own key — **do not** commit it in `.env` in shared repos. |

| `PAYCREST_HTTP_TIMEOUT` | `.env` or host `env` | No | Client timeout (Go duration, e.g. `30s`). Default: `60s` |

| `PAYCREST_MAX_RESPONSE_BYTES` | `.env` or host `env` | No | Max response body read (default `10485760`) |



### `.env` file (local, optional)



Use `.env` for **non-secret** defaults such as `PAYCREST_BASE_URL` when running from the `sender-mcp/` directory.



1. Copy: `cp .env.example .env` (or copy `sender-mcp/.env.example` to `sender-mcp/.env` on Windows).

2. Set `PAYCREST_BASE_URL` if you want a local default.

3. **`PAYCREST_API_KEY` is intentionally omitted** from `.env.example` — agents and CI should not rely on a shared key file. Each user configures their key in **Cursor / Claude / etc.** under the server’s `env` block.



The server calls `godotenv.Load()` at startup; missing `.env` is ignored. Host-provided environment variables override `.env`.



**Git:** only `.env.example` is tracked; `.env` is gitignored.



## Build



```bash

cd sender-mcp

go build -o paycrest-mcp .

```



## Cursor (example)



Each user adds their own `PAYCREST_API_KEY` in **their** MCP config (e.g. user `~/.cursor/mcp.json`). Do not commit real keys in project repos.



```json

{

  "mcpServers": {

    "paycrest": {

      "command": "C:/full/path/to/paycrest-mcp.exe",

      "env": {

        "PAYCREST_BASE_URL": "https://api.paycrest.io",

        "PAYCREST_API_KEY": "your-sender-api-key"

      }

    }

  }

}

```



Omit `PAYCREST_API_KEY` if you only use public tools (`paycrest_get_currencies`, `paycrest_get_rates`, etc.).



Use the full path to `paycrest-mcp` if it is not on `PATH`.



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

| `paycrest_get_sender_order` | `GET /v2/sender/orders/{id}` |



Responses are returned as **JSON text** (the raw HTTP body from the API). HTTP status `>= 400` is surfaced as an MCP tool error (`isError`).



## Notes



- This binary does **not** add routes to the aggregator; it is an HTTP client only.

- Sender tools require the same credentials as other non-web API clients (`API-Key` header).

- For HMAC-based sender auth without an API key, extend this server or use the documented HMAC flow separately (not implemented here).

