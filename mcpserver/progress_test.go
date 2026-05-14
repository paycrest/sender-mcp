package mcpserver

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func Test_watchProgressFromCallToolRequest_nil(t *testing.T) {
	if p := watchProgressFromCallToolRequest(nil); p != nil {
		t.Fatalf("expected nil, got %+v", p)
	}
}

func Test_watchProgressNotifier_notify_nilReceiver(t *testing.T) {
	var p *watchProgressNotifier
	p.notify(t.Context(), 1, "x") // must not panic
}

func Test_watchProgressFromCallToolRequest_withoutServerSession(t *testing.T) {
	req := &mcp.CallToolRequest{}
	if p := watchProgressFromCallToolRequest(req); p != nil {
		t.Fatalf("expected nil without session, got %+v", p)
	}
}
