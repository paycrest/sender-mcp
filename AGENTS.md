# Paycrest Sender MCP — Agent instructions

Go MCP server exposing Paycrest aggregator `/v2` APIs to Cursor and other MCP hosts. See [README.md](README.md) and [docs/USER_GUIDE.md](docs/USER_GUIDE.md).

## Issue tracker (Jira)

File new work in Jira project **KAN** ([paycrest-io.atlassian.net](https://paycrest-io.atlassian.net)), label **`sender-mcp`**. See [docs/agents/issue-tracker.md](docs/agents/issue-tracker.md) for issue types, Atlassian MCP usage, and PR linking. **Ticket spec:** [docs/agents/ticket-spec-template.md](docs/agents/ticket-spec-template.md) (Engineering Working Agreement). **PR template:** [.github/pull_request_template.md](.github/pull_request_template.md) (separate from ticket spec).

Do **not** use `gh issue create` or GitHub Issues for new tickets.

## GitHub PRs

Use [.github/pull_request_template.md](.github/pull_request_template.md). Prefix **branch** with the Jira key (e.g. `KAN-123-short-description`) and **PR title** with `KAN-123: Short description`. Include `Jira Issue: https://paycrest-io.atlassian.net/browse/KAN-123` in the description. Do not add AI attribution to PR titles or descriptions.

## Conventions

- Go stdio MCP server; tools in `mcpserver/`, HTTP client in `paycrest/`.
- Order create/watch flow rules: [.cursor/rules/paycrest-mcp-order-flow.mdc](.cursor/rules/paycrest-mcp-order-flow.mdc).
- Tag releases (`v*`) to publish binaries via GitHub Actions.
