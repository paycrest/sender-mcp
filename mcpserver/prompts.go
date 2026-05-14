package mcpserver

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var paycrestOrderIDArg = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// RegisterPrompts adds MCP prompts the host may show (e.g. Cursor MCP → Prompts).
// This does not render a button on tool output; it offers a repeatable “after payment” flow.
func RegisterPrompts(s *mcp.Server) {
	s.AddPrompt(&mcp.Prompt{
		Name:        "paycrest_after_payment_watch",
		Title:       "After payment — watch order",
		Description: "After you send funds to the provider account, run this with the order id from paycrest_create_order so the agent calls paycrest_watch_sender_order.",
		Arguments: []*mcp.PromptArgument{{
			Name:        "order_id",
			Title:       "Order id",
			Description: "The data.id UUID from the create_order response (also under Status polling).",
			Required:    true,
		}},
	}, afterPaymentWatchPrompt)
}

func afterPaymentWatchPrompt(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	if req.Params == nil || req.Params.Arguments == nil {
		return nil, fmt.Errorf("missing prompt arguments")
	}
	id := strings.TrimSpace(req.Params.Arguments["order_id"])
	if id == "" || !paycrestOrderIDArg.MatchString(id) {
		return nil, fmt.Errorf("order_id must be the UUID from create_order (data.id)")
	}
	text := fmt.Sprintf(
		"I have paid (funds sent). Call paycrest_watch_sender_order immediately with "+
			`{"id":%q,"poll_interval_sec":10,"max_wait_sec":3600} and poll until terminal status; then summarize final status and any txHash.`,
		id,
	)
	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{{
			Role:    "user",
			Content: &mcp.TextContent{Text: text},
		}},
	}, nil
}
