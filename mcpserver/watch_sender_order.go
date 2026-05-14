package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paycrest/paycrest/sender-mcp/paycrest"
	"github.com/paycrest/paycrest/sender-mcp/types"
)

// terminalOrderStatuses are payment-order states after which polling stops (aggregator paymentorder.Status).
var terminalOrderStatuses = map[string]struct{}{
	"settled":   {},
	"cancelled": {},
	"refunded":  {},
	"expired":   {},
}

func isTerminalOrderStatus(s string) bool {
	_, ok := terminalOrderStatuses[strings.ToLower(strings.TrimSpace(s))]
	return ok
}

func clampInt(v, min, max, def int) int {
	if v <= 0 {
		return def
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

type orderGETSnapshot struct {
	EnvelopeStatus string
	OrderStatus    string
	TxHash         string
}

func parseV2OrderGETSnapshot(body []byte) (snap orderGETSnapshot, ok bool) {
	var env struct {
		Status string          `json:"status"`
		Data   json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return orderGETSnapshot{}, false
	}
	snap.EnvelopeStatus = env.Status
	if len(env.Data) == 0 || string(env.Data) == "null" {
		return snap, true
	}
	var d struct {
		Status string `json:"status"`
		TxHash string `json:"txHash"`
	}
	_ = json.Unmarshal(env.Data, &d)
	snap.OrderStatus = d.Status
	snap.TxHash = d.TxHash
	return snap, true
}

func appendLastGETSnapshot(b *strings.Builder, lastBody []byte) {
	if len(lastBody) == 0 {
		return
	}
	b.WriteString("\n--- Last GET body ---\n")
	b.Write(lastBody)
	b.WriteByte('\n')
}

// watchSenderOrderTranscript polls GET /v2/sender/orders/{id} until terminal status, error, timeout, or ctx cancel.
// prog sends MCP progress notifications when the client supplied a progress token (optional); nil prog is a no-op.
func watchSenderOrderTranscript(ctx context.Context, c *paycrest.Client, in types.WatchSenderOrderIn, sectionTitle string, prog *watchProgressNotifier) (text string, isErr bool) {
	id := strings.TrimSpace(in.ID)
	if id == "" {
		return "watch error: id is required\n", true
	}
	intervalSec := clampInt(in.PollIntervalSec, 2, 30, 3)
	maxWaitSec := clampInt(in.MaxWaitSec, types.WatchWaitMinSec, types.WatchWaitMaxSec, types.WatchDefaultMaxWaitSec)
	interval := time.Duration(intervalSec) * time.Second
	deadline := time.Now().Add(time.Duration(maxWaitSec) * time.Second)
	path := "/v2/sender/orders/" + url.PathEscape(id)

	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", sectionTitle)
	fmt.Fprintf(&b, "order id: %s\n", id)
	fmt.Fprintf(&b, "poll every %ds, max wait %ds\n", intervalSec, maxWaitSec)
	fmt.Fprintf(&b, "stop when order status is one of: settled, cancelled, refunded, expired\n\n")

	prog.notify(ctx, 0, fmt.Sprintf("polling started (every %ds, max %ds)", intervalSec, maxWaitSec))

	poll := 0
	var lastBody []byte
	consecutive429 := 0
	for {
		if err := ctx.Err(); err != nil {
			b.WriteString(fmt.Sprintf("\nstopped: %v\n", err))
			appendLastGETSnapshot(&b, lastBody)
			return b.String(), true
		}
		if time.Now().After(deadline) {
			b.WriteString(fmt.Sprintf("\nmax_wait_sec (%d) exceeded — last snapshot below.\n", maxWaitSec))
			b.WriteString("Order still non-terminal when max_wait_sec elapsed. Run paycrest_watch_sender_order again on this id (optionally raise max_wait_sec up to 3600) — keep polling until Paycrest returns settled / cancelled / refunded / expired. Do not rely on a single paycrest_get_sender_order for that.\n")
			prog.notify(ctx, float64(poll), fmt.Sprintf("stopped: max_wait_sec (%d) exceeded", maxWaitSec))
			appendLastGETSnapshot(&b, lastBody)
			return b.String(), false
		}

		poll++
		httpSt, body, resHdr, err := c.GetForPoll(ctx, path, nil, true)
		lastBody = body
		ts := time.Now().UTC().Format(time.RFC3339Nano)

		if err != nil {
			fmt.Fprintf(&b, "[%s] poll #%d GET error: %v\n", ts, poll, err)
			appendLastGETSnapshot(&b, lastBody)
			return b.String(), true
		}

		if httpSt == http.StatusTooManyRequests {
			consecutive429++
			wait := interval
			if d, ok := paycrest.ParseRetryAfter(resHdr); ok && d > 0 {
				wait = d
			} else {
				shift := uint(min(consecutive429-1, 6))
				exp := 5 * time.Second << shift
				if exp > 2*time.Minute {
					exp = 2 * time.Minute
				}
				if exp < interval {
					exp = interval
				}
				wait = exp
			}
			if wait < interval {
				wait = interval
			}
			if wait > 2*time.Minute {
				wait = 2 * time.Minute
			}
			fmt.Fprintf(&b, "[%s] poll #%d HTTP 429 — backing off %v (consecutive=%d)\n",
				ts, poll, wait.Round(time.Millisecond), consecutive429)
			prog.notify(ctx, float64(poll), fmt.Sprintf("HTTP 429 — backing off %v", wait.Round(time.Millisecond)))
			select {
			case <-ctx.Done():
				b.WriteString(fmt.Sprintf("\nstopped: %v\n", ctx.Err()))
				appendLastGETSnapshot(&b, lastBody)
				return b.String(), true
			case <-time.After(wait):
			}
			continue
		}
		consecutive429 = 0

		if httpSt != http.StatusOK {
			fmt.Fprintf(&b, "[%s] poll #%d HTTP %d\n", ts, poll, httpSt)
			b.Write(body)
			b.WriteByte('\n')
			return b.String(), true
		}

		snap, parseOK := parseV2OrderGETSnapshot(body)
		if !parseOK {
			fmt.Fprintf(&b, "[%s] poll #%d could not parse JSON body\n", ts, poll)
			b.Write(body)
			b.WriteByte('\n')
			appendLastGETSnapshot(&b, lastBody)
			return b.String(), true
		}

		tx := snap.TxHash
		if len(tx) > 20 {
			tx = tx[:10] + "…" + tx[len(tx)-6:]
		}
		fmt.Fprintf(&b, "[%s] poll #%d HTTP %d envelope=%s order_status=%s txHash=%s\n",
			ts, poll, httpSt, strings.TrimSpace(snap.EnvelopeStatus), strings.TrimSpace(snap.OrderStatus), strings.TrimSpace(tx))

		prog.notify(ctx, float64(poll), fmt.Sprintf("order_status=%s", strings.TrimSpace(snap.OrderStatus)))

		envLower := strings.ToLower(strings.TrimSpace(snap.EnvelopeStatus))
		if envLower != "" && envLower != "success" {
			prog.notify(ctx, float64(poll), "envelope error: "+strings.TrimSpace(snap.EnvelopeStatus))
			b.WriteString("\n--- Last GET body ---\n")
			b.Write(body)
			b.WriteByte('\n')
			return b.String(), true
		}

		if isTerminalOrderStatus(snap.OrderStatus) {
			prog.notify(ctx, float64(poll), "terminal: "+strings.TrimSpace(snap.OrderStatus))
			b.WriteString("\n--- Terminal order status reached ---\n")
			b.Write(body)
			b.WriteByte('\n')
			return b.String(), false
		}

		select {
		case <-ctx.Done():
			b.WriteString(fmt.Sprintf("\nstopped: %v\n", ctx.Err()))
			appendLastGETSnapshot(&b, lastBody)
			return b.String(), true
		case <-time.After(interval):
		}
	}
}

func runWatchSenderOrder(ctx context.Context, c *paycrest.Client, req *mcp.CallToolRequest, in types.WatchSenderOrderIn) (*mcp.CallToolResult, any, error) {
	if strings.TrimSpace(in.ID) == "" {
		return toolResultErr(fmt.Errorf("id is required"))
	}
	prog := watchProgressFromCallToolRequest(req)
	text, isErr := watchSenderOrderTranscript(ctx, c, in, "--- paycrest_watch_sender_order ---", prog)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
		IsError: isErr,
	}, nil, nil
}
