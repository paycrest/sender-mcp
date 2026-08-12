# Ticket / Spec template (Jira KAN)

Copy this into the **Jira ticket description** when creating work for the sender-mcp repo. The ticket is the single source of truth — no spec, no build. Add a **flowchart in a Jira comment** when the change touches multi-step order create/watch flows.

**Label:** `sender-mcp` (required)

---

## User Story

Add the details of this issue from the user's POV.

---

## Acceptance Criteria

Include at least one **failure-case** scenario, not only the happy path.

1. **GIVEN** …
   **WHEN** …
   **THEN** …

2. **GIVEN** … (failure / edge case)
   **WHEN** …
   **THEN** …

---

## Tech Details

- `mcpserver/` tools, prompts, or handlers affected
- Paycrest API client changes
- [docs/USER_GUIDE.md](../USER_GUIDE.md) updates if sender-facing behavior changes
- Order create/watch flow vs `.cursor/rules/paycrest-mcp-order-flow.mdc`

---

## Money-safety

- Touches order creation, payment instructions, or watch/poll behavior on live orders? **Yes / No**
- If **Yes**: second human reviewer required before prod; call out failure cases and never auto-submit payments.
- Tooling, logging, or docs with no order side effects: usually **No**.

---

## Notes / Assumptions

- Assumptions that must stay true for this change to remain correct.

---

## Open Questions

- …

---

## Bug tickets (shorter variant)

For **Bug** issue type, use at minimum:

**Describe the bug** — what is wrong (tool error, wrong pay-in details, watch stops early).

**To reproduce** — MCP tool, args, environment.

**Expected** — correct tool output or flow.

**Environment** — staging vs production API.

**Acceptance criteria** — **GIVEN / WHEN / THEN** for the fix, including one regression check.
