### Jira Issue

Jira Issue: <!-- e.g. https://paycrest-io.atlassian.net/browse/KAN-123 -->

### Description

> Describe the purpose of this PR (MCP tools, prompts, Paycrest client, install scripts, etc.).
>
> Spec and acceptance criteria live on the linked Jira ticket — not in this template.

### Self-review

- [ ] Reviewed diff against Jira acceptance criteria (including failure cases)
- [ ] CodeRabbit / CI green

### References

> - GitHub Issue/PR number addressed or fixed e.g closes #407 (legacy GitHub issues only)
> - Related aggregator API changes (optional)

### Testing

> Describe `go test ./...`, manual MCP smoke tests in Cursor, and release workflow checks if applicable.

### Staging

- [ ] MCP smoke test against staging Paycrest API

### Checklist

- [ ] Tests added or updated where applicable
- [ ] [docs/USER_GUIDE.md](docs/USER_GUIDE.md) updated if sender-facing behavior changed
- [ ] Order create/watch flow still matches `.cursor/rules/paycrest-mcp-order-flow.mdc`
