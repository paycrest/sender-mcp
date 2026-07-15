#!/usr/bin/env bash
# Install paycrest-mcp into a fixed location and (optionally) wire Cursor mcp.json.
#
#   curl -fsSL https://raw.githubusercontent.com/paycrest/sender-mcp/main/scripts/install.sh | bash
#
# Env:
#   PAYCREST_API_KEY  — if set, written into mcp.json; else uses ${env:PAYCREST_API_KEY}
#   TAG               — release tag (default: latest), e.g. v0.1.0
#   SKIP_MCP_JSON=1   — only download binary
#   REPO              — default paycrest/sender-mcp
set -euo pipefail

REPO="${REPO:-paycrest/sender-mcp}"
TAG="${TAG:-latest}"
INSTALL_DIR="${HOME}/.paycrest"
DEST="${INSTALL_DIR}/paycrest-mcp"
MCP_PATH="${HOME}/.cursor/mcp.json"
API_KEY="${PAYCREST_API_KEY:-}"

uname_s="$(uname -s)"
uname_m="$(uname -m)"
case "${uname_s}" in
  Darwin) goos=darwin ;;
  Linux)  goos=linux ;;
  *) echo "Unsupported OS: ${uname_s}" >&2; exit 1 ;;
esac
case "${uname_m}" in
  x86_64|amd64) goarch=amd64 ;;
  arm64|aarch64) goarch=arm64 ;;
  *) echo "Unsupported arch: ${uname_m}" >&2; exit 1 ;;
esac

ASSET_SUFFIX="${goos}_${goarch}"

if [[ "${TAG}" == "latest" ]]; then
  API_URL="https://api.github.com/repos/${REPO}/releases/latest"
else
  API_URL="https://api.github.com/repos/${REPO}/releases/tags/${TAG}"
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required for install.sh (JSON parsing)." >&2
  exit 1
fi

echo "Fetching release (${TAG}) from github.com/${REPO} ..."
RELEASE_JSON="$(curl -fsSL -H "Accept: application/vnd.github+json" -H "User-Agent: paycrest-mcp-install" "${API_URL}")"
eval "$(printf '%s' "${RELEASE_JSON}" | ASSET_SUFFIX="${ASSET_SUFFIX}" python3 -c '
import json, os, sys
rel = json.load(sys.stdin)
suffix = os.environ["ASSET_SUFFIX"]
tag = rel.get("tag_name") or ""
url = ""
name = ""
for a in rel.get("assets") or []:
    n = a.get("name") or ""
    if suffix in n and "checksums" not in n:
        url = a.get("browser_download_url") or ""
        name = n
        break
if not url:
    sys.stderr.write(f"No asset matching *{suffix}* on release {tag}\n")
    sys.exit(1)
# shell-safe export
def sh(s: str) -> str:
    return "'" + s.replace("'", "'"'"'") + "'"
print(f"TAG_NAME={sh(tag)}")
print(f"DOWNLOAD_URL={sh(url)}")
print(f"ASSET_NAME={sh(name)}")
')"

mkdir -p "${INSTALL_DIR}"
echo "Downloading ${ASSET_NAME} → ${DEST}"
tmp="$(mktemp)"
curl -fsSL -o "${tmp}" "${DOWNLOAD_URL}"
chmod +x "${tmp}"
mv "${tmp}" "${DEST}"

echo "Installed: ${DEST} (${TAG_NAME})"

if [[ "${SKIP_MCP_JSON:-0}" != "1" ]]; then
  mkdir -p "$(dirname "${MCP_PATH}")"
  PAYCREST_COMMAND="${DEST}" PAYCREST_API_KEY_VALUE="${API_KEY}" MCP_PATH="${MCP_PATH}" python3 <<'PY'
import json, os
path = os.environ["MCP_PATH"]
command = os.environ["PAYCREST_COMMAND"]
api_key = os.environ.get("PAYCREST_API_KEY_VALUE") or ""
key = api_key if api_key.strip() else "${env:PAYCREST_API_KEY}"
server = {
    "command": command,
    "env": {
        "PAYCREST_API_KEY": key,
    },
}
if os.path.isfile(path):
    with open(path, encoding="utf-8") as f:
        root = json.load(f)
else:
    root = {}
servers = root.setdefault("mcpServers", {})
servers["paycrest"] = server
with open(path, "w", encoding="utf-8") as f:
    json.dump(root, f, indent=2)
    f.write("\n")
print(f"Updated Cursor MCP config: {path}")
PY
fi

echo ""
echo "Next:"
if [[ -z "${API_KEY}" ]]; then
  echo "  1. export PAYCREST_API_KEY=<your-dashboard-key>"
  echo "     (or set it in ${MCP_PATH})"
else
  echo "  1. API key written into mcp.json (do not commit that file)."
fi
echo "  2. Reload MCP in Cursor, or restart Cursor."
echo "  3. Test: ask the agent to list Paycrest currencies."
echo ""
echo "Fixed binary path: ${DEST}"
