package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paycrest/paycrest/sender-mcp/paycrest"
	"github.com/paycrest/paycrest/sender-mcp/types"
)

// RegisterTools attaches Paycrest API tools and prompts to the MCP server.
func RegisterTools(s *mcp.Server, c *paycrest.Client, cfg types.Config) {
	RegisterPrompts(s)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "paycrest_get_currencies",
		Description: "GET /v2/currencies — list fiat currencies supported by Paycrest.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ types.Empty) (*mcp.CallToolResult, any, error) {
		status, data, err := c.Get(ctx, "/v2/currencies", nil, false)
		if err != nil {
			return nil, nil, err
		}
		return httpToolResult(status, data)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "paycrest_get_institutions",
		Description: "GET /v2/institutions/{currency_code} — institutions for a fiat currency code (e.g. NGN).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in types.InstitutionsIn) (*mcp.CallToolResult, any, error) {
		code := strings.TrimSpace(in.CurrencyCode)
		if code == "" {
			return toolResultErr(fmt.Errorf("currency_code is required"))
		}
		path := "/v2/institutions/" + url.PathEscape(code)
		status, data, err := c.Get(ctx, path, nil, false)
		if err != nil {
			return nil, nil, err
		}
		return httpToolResult(status, data)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "paycrest_get_tokens",
		Description: "GET /v2/tokens — supported tokens and networks.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ types.Empty) (*mcp.CallToolResult, any, error) {
		status, data, err := c.Get(ctx, "/v2/tokens", nil, false)
		if err != nil {
			return nil, nil, err
		}
		return httpToolResult(status, data)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "paycrest_get_pubkey",
		Description: "GET /v2/pubkey — aggregator public key.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ types.Empty) (*mcp.CallToolResult, any, error) {
		status, data, err := c.Get(ctx, "/v2/pubkey", nil, false)
		if err != nil {
			return nil, nil, err
		}
		return httpToolResult(status, data)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "paycrest_get_rates",
		Description: "GET /v2/rates/{network}/{token}/{amount}/{fiat} — V2 buy/sell quote. Optional query: provider_id (8 letters), side (buy|sell).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in types.RatesIn) (*mcp.CallToolResult, any, error) {
		if strings.TrimSpace(in.Network) == "" || strings.TrimSpace(in.Token) == "" || strings.TrimSpace(in.Amount) == "" || strings.TrimSpace(in.Fiat) == "" {
			return toolResultErr(fmt.Errorf("network, token, amount, and fiat are required"))
		}
		path := fmt.Sprintf("/v2/rates/%s/%s/%s/%s",
			url.PathEscape(strings.TrimSpace(in.Network)),
			url.PathEscape(strings.TrimSpace(in.Token)),
			url.PathEscape(strings.TrimSpace(in.Amount)),
			url.PathEscape(strings.TrimSpace(in.Fiat)),
		)
		q := url.Values{}
		if strings.TrimSpace(in.ProviderID) != "" {
			q.Set("provider_id", strings.TrimSpace(in.ProviderID))
		}
		if strings.TrimSpace(in.Side) != "" {
			q.Set("side", strings.TrimSpace(in.Side))
		}
		status, data, err := c.Get(ctx, path, q, false)
		if err != nil {
			return nil, nil, err
		}
		return httpToolResult(status, data)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "paycrest_list_sender_orders",
		Description: "GET /v2/sender/orders — list payment orders for the authenticated sender. Requires PAYCREST_API_KEY. Optional: page, pageSize, ordering (asc|desc), search, from, to, export.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in types.ListOrdersIn) (*mcp.CallToolResult, any, error) {
		q := url.Values{}
		if in.Page > 0 {
			q.Set("page", fmt.Sprintf("%d", in.Page))
		}
		if in.PageSize > 0 {
			q.Set("pageSize", fmt.Sprintf("%d", in.PageSize))
		}
		if strings.TrimSpace(in.Ordering) != "" {
			q.Set("ordering", strings.TrimSpace(in.Ordering))
		}
		if strings.TrimSpace(in.Search) != "" {
			q.Set("search", strings.TrimSpace(in.Search))
		}
		if strings.TrimSpace(in.From) != "" {
			q.Set("from", strings.TrimSpace(in.From))
		}
		if strings.TrimSpace(in.To) != "" {
			q.Set("to", strings.TrimSpace(in.To))
		}
		if strings.TrimSpace(in.Export) != "" {
			q.Set("export", strings.TrimSpace(in.Export))
		}
		status, data, err := c.Get(ctx, "/v2/sender/orders", q, true)
		if err != nil {
			return nil, nil, err
		}
		return httpToolResult(status, data)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "paycrest_watch_sender_order",
		Description: "Use this after the user pays or asks for live status until the order finishes. Polls GET /v2/sender/orders/{id} until terminal status (settled, cancelled, refunded, expired) or max_wait_sec elapses. Requires PAYCREST_API_KEY. Optional: poll_interval_sec (2–30, default 3), max_wait_sec (10–3600, default 3600 when omitted — up to 1 hour of polling). HTTP 429 waits (Retry-After or exponential backoff) and polling continues. When the MCP client sends a progress token with the call, the server emits notifications/progress with order_status updates for the host UI. Returns a timestamped transcript plus the last successful GET JSON when stopped.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in types.WatchSenderOrderIn) (*mcp.CallToolResult, any, error) {
		return runWatchSenderOrder(ctx, c, req, in)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "paycrest_get_sender_order",
		Description: "GET /v2/sender/orders/{id} — one snapshot (single GET) of a payment order by id (V2 schema). Requires PAYCREST_API_KEY. Does not poll. After the user says paid / funds sent / transfer confirmed on-chain, do not call this tool alone — call paycrest_watch_sender_order on that id so status updates repeat until terminal (MCP progress when the host sends a progress token). Use get only for a deliberate one-off peek or to resolve an id before immediately calling watch.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in types.GetOrderIn) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(in.ID)
		if id == "" {
			return toolResultErr(fmt.Errorf("id is required"))
		}
		path := "/v2/sender/orders/" + url.PathEscape(id)
		status, data, err := c.Get(ctx, path, nil, true)
		if err != nil {
			return nil, nil, err
		}
		return httpToolResult(status, data)
	})

	s.AddTool(&mcp.Tool{
		Name:        "paycrest_create_order",
		Description: "POST /v2/sender/orders — create a payment order (V2 payload). Requires PAYCREST_API_KEY. On HTTP 201: raw JSON + providerAccount pay-in summary + mandatory user ACTION block (lines \"ACTION\" then \"After you send the <fiat>…\" / generic funds line). When summarizing, copy that ACTION block verbatim after refund (if any) — do not replace with tool names or order UUIDs in the user-facing sentence. By default (PAYCREST_AUTO_WATCH_AFTER_CREATE off) returns quickly; you then call paycrest_watch_sender_order with data.id (see sender-mcp/.cursor/rules/paycrest-mcp-order-flow.mdc). Set PAYCREST_AUTO_WATCH_AFTER_CREATE=true to embed the same poll as paycrest_watch_sender_order (15s interval, 429 backoff, MCP progress) until terminal or PAYCREST_CREATE_ORDER_MAX_WAIT_SEC (default 900, max 3600).",
		InputSchema: json.RawMessage(`{"type":"object","additionalProperties":true,"description":"JSON body for POST /v2/sender/orders (V2PaymentOrderPayload)."}`),
	}, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		raw := req.Params.Arguments
		if len(strings.TrimSpace(string(raw))) == 0 {
			return rawToolErr(fmt.Errorf("arguments must be a JSON object matching V2PaymentOrderPayload"))
		}
		if !json.Valid(raw) {
			return rawToolErr(fmt.Errorf("arguments must be valid JSON"))
		}
		status, data, err := c.PostJSON(ctx, "/v2/sender/orders", raw, true)
		if err != nil {
			return rawToolErr(err)
		}
		if status == http.StatusCreated {
			data = appendCreateOrderProviderSummary(data)
			if cfg.AutoWatchAfterCreate {
				if id, ok := ExtractV2CreateOrderID(data); ok {
					prog := watchProgressFromCallToolRequest(req)
					watchText, _ := watchSenderOrderTranscript(ctx, c, types.WatchSenderOrderIn{
						ID:              id,
						PollIntervalSec: 15,
						MaxWaitSec:      cfg.CreateOrderMaxWaitSec,
					}, "--- Automated status watch (after create_order) ---", prog)
					data = append(data, "\n\n"...)
					data = append(data, watchText...)
				} else {
					data = appendCreateOrderWatchHint(data)
				}
			} else {
				data = appendCreateOrderWatchHint(data)
			}
		}
		return httpRawToolResult(status, data)
	})
}

func httpToolResult(status int, data []byte) (*mcp.CallToolResult, any, error) {
	text := string(data)
	if status >= http.StatusBadRequest {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
			IsError: true,
		}, nil, nil
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, nil, nil
}

func httpRawToolResult(status int, data []byte) (*mcp.CallToolResult, error) {
	text := string(data)
	if status >= http.StatusBadRequest {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
			IsError: true,
		}, nil
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, nil
}

func rawToolErr(err error) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
		IsError: true,
	}, nil
}

func toolResultErr(err error) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
		IsError: true,
	}, nil, nil
}
