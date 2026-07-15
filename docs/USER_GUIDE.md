# Paycrest Sender MCP — User Guide

Use **Paycrest** from **Cursor** through the Model Context Protocol (MCP). This small app runs on your machine, talks to the Paycrest API, and lets the Cursor agent create and track payment orders in plain language.

You do **not** need Go or a local clone for normal use — only Cursor and a sender API key.

---

## Prerequisites

1. **[Cursor](https://cursor.com/)** installed
2. A **sender API key** from the [Paycrest dashboard](https://paycrest.io) (Settings → API keys, or equivalent)
3. Access to [GitHub Releases](https://github.com/paycrest/sender-mcp/releases) for this repo (public repo, or you’re logged in / using `gh`)

---

## Install

The installer downloads the latest release binary to a **fixed path** and can wire Cursor’s MCP config for you.

| OS | Binary path |
|----|-------------|
| Windows | `%USERPROFILE%\.paycrest\paycrest-mcp.exe` |
| macOS / Linux | `~/.paycrest/paycrest-mcp` |

### Windows (PowerShell — not Git Bash)

Open **Windows PowerShell** (or Terminal → PowerShell). Do **not** use Git Bash for this one-liner.

```powershell
irm https://raw.githubusercontent.com/paycrest/sender-mcp/main/scripts/install.ps1 | iex
```

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/paycrest/sender-mcp/main/scripts/install.sh | bash
```

### What the script does

1. Downloads the matching binary from [Releases](https://github.com/paycrest/sender-mcp/releases)
2. Saves it under `~/.paycrest/` with a fixed name (no manual rename)
3. Merges a `paycrest` entry into `~/.cursor/mcp.json` (unless you skip that step)

**Ignore** the **Source code (zip/tar.gz)** links on the Releases page — those are source archives, not the MCP binary.

---

## Configure Cursor (`mcp.json`)

You only need **`PAYCREST_API_KEY`**.

Typical config after install (path may match your username):

**Windows**

```json
{
  "mcpServers": {
    "paycrest": {
      "command": "C:/Users/YOU/.paycrest/paycrest-mcp.exe",
      "env": {
        "PAYCREST_API_KEY": "your-sender-api-key"
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
        "PAYCREST_API_KEY": "your-sender-api-key"
      }
    }
  }
}
```

Portable alternative using Cursor’s `${userHome}`:

```json
"command": "${userHome}/.paycrest/paycrest-mcp.exe"
```

(Omit `.exe` on macOS/Linux.)

Use **your own** dashboard key. Never share or commit it.

### Reload MCP

After install or editing `mcp.json`:

1. Open **Cursor Settings → Tools & MCP**, or
2. Restart Cursor

Confirm **paycrest** shows as connected / no error on the server.

---

## Smoke tests

In Cursor chat, try:

- “List Paycrest currencies”
- “What’s the rate for 100 USDC on base to NGN?”

If those succeed, the MCP server and API key are working.

---

## Order flow (create → pay → watch)

Typical flow the agent should follow:

1. **Create** an order (onramp or offramp) via natural language.
2. The agent shows **pay-in details** (bank / wallet / account info) and the order id.
3. You send the funds (fiat or crypto, as instructed).
4. Reply **`paid`** (or “I have paid” / “transfer confirmed”).
5. The agent runs **one continuous watch** until the order is **settled**, **cancelled**, **refunded**, or **expired** — you should not need to keep clicking Run for each status hop.

You can also use the MCP prompt **“After payment — watch order”** and paste the order id if the chat lost context.

### Example prompts

**Offramp** (crypto → fiat):

> offramp 0.5 USDC Base, refund 0x0c89b146E1dB0A1cf325e0bBA4FfFF24e3F1d0D4, NGN OPay 0987654321, John Doe.

**Onramp** (fiat → crypto):

> onramp 1000 NGN to USDC on Base at 0xCfb28A38b6083B2Ef0B4F99C7cC3aB76b71F0426, refund OPay 0987654321, name John Doe.

After you’ve paid:

> paid

---

## Upgrade

1. **Quit Cursor completely** first (the running MCP holds a lock on the `.exe` / binary).
2. Re-run the same install one-liner (or `.\scripts\install.ps1` / `./scripts/install.sh` from a clone).
3. Open Cursor and reload MCP if needed.

If you skip quitting Cursor, Windows often fails with “file in use” / access denied.

---

## Troubleshooting

| Problem | What to do |
|---------|------------|
| `irm` / curl returns **404** | Repo or raw script URL not reachable (private repo or wrong branch). Use a [Release](https://github.com/paycrest/sender-mcp/releases) download while logged into GitHub, or clone and run `.\scripts\install.ps1` / `./scripts/install.sh` with `gh auth login`. |
| **File in use** / cannot overwrite binary | Quit Cursor fully, then re-run install. |
| Install fails in **Git Bash** | Use **PowerShell** on Windows for `irm … \| iex`. Git Bash is not the supported path for that one-liner. |
| Auth / **401** / sender tools fail | Check `PAYCREST_API_KEY` in `~/.cursor/mcp.json` — wrong, expired, or missing key. Reload MCP after fixing. |
| Downloaded something that isn’t an app | Ignore **Source code (zip/tar.gz)** on Releases. Use the platform asset (e.g. `…_windows_amd64.exe`, `…_darwin_arm64`). |
| MCP server red / not listed | Confirm `command` path exists (`%USERPROFILE%\.paycrest\paycrest-mcp.exe` or `~/.paycrest/paycrest-mcp`), then reload MCP / restart Cursor. |

---

## Links

- **Releases (binaries):** https://github.com/paycrest/sender-mcp/releases
- **Maintainer / technical README:** [../README.md](../README.md)
