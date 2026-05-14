package mcpserver

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func Test_afterPaymentWatchPrompt_valid(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"order_id": "e3447fe7-731a-424c-9e07-ae6d6675dedc",
			},
		},
	}
	res, err := afterPaymentWatchPrompt(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Messages) != 1 {
		t.Fatalf("messages=%d", len(res.Messages))
	}
	tc, ok := res.Messages[0].Content.(*mcp.TextContent)
	if !ok || !strings.Contains(tc.Text, "paycrest_watch_sender_order") || !strings.Contains(tc.Text, "e3447fe7-731a-424c-9e07-ae6d6675dedc") {
		t.Fatalf("unexpected content: %#v", res.Messages[0].Content)
	}
}

func Test_afterPaymentWatchPrompt_invalid(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{"order_id": "not-a-uuid"},
		},
	}
	_, err := afterPaymentWatchPrompt(context.Background(), req)
	if err == nil {
		t.Fatal("expected error")
	}
}
