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

// RegisterTools attaches Paycrest API tools to the MCP server.
func RegisterTools(s *mcp.Server, c *paycrest.Client) {
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
		Name:        "paycrest_get_sender_order",
		Description: "GET /v2/sender/orders/{id} — get one payment order by id (V2 schema). Requires PAYCREST_API_KEY.",
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
		Description: "POST /v2/sender/orders — create a payment order (V2 payload: source/destination JSON objects, amount, etc.). Requires PAYCREST_API_KEY. On success (201), the tool output includes the raw JSON plus a short \"providerAccount\" section: onramp lists institution, accountIdentifier, accountName, validUntil, amountToTransfer, currency (fiat pay-in); offramp lists network, receiveAddress, validUntil (crypto receive).",
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
