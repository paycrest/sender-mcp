package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// watchProgressNotifier sends MCP progress notifications to the host (e.g. Cursor)
// when the client included a progress token on the tools/call request.
type watchProgressNotifier struct {
	session *mcp.ServerSession
	token   any
}

func watchProgressFromCallToolRequest(req *mcp.CallToolRequest) *watchProgressNotifier {
	if req == nil || req.Params == nil {
		return nil
	}
	ss, ok := req.GetSession().(*mcp.ServerSession)
	if !ok {
		return nil
	}
	tok := req.Params.GetProgressToken()
	if tok == nil {
		return nil
	}
	return &watchProgressNotifier{session: ss, token: tok}
}

// notify reports monotonic progress (poll index) and a short status message. Errors are ignored
// so polling is never aborted because the client ignored progress. A nil receiver is a no-op.
func (p *watchProgressNotifier) notify(ctx context.Context, progress float64, message string) {
	if p == nil || p.session == nil || p.token == nil {
		return
	}
	_ = p.session.NotifyProgress(ctx, &mcp.ProgressNotificationParams{
		ProgressToken: p.token,
		Progress:      progress,
		Message:       message,
	})
}
