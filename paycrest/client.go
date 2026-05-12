package paycrest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/paycrest/paycrest/sender-mcp/types"
)

// Client calls the Paycrest aggregator HTTP API.
type Client struct {
	http    *http.Client
	baseURL string
	apiKey  string
	maxBody int64
}

// NewClient builds an HTTP client for the given config.
func NewClient(cfg types.Config) *Client {
	return &Client{
		http: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		maxBody: cfg.MaxRespBytes,
	}
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body []byte, senderAuth bool) (status int, respBody []byte, err error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return 0, nil, fmt.Errorf("parse url: %w", err)
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}

	var rdr io.Reader
	if len(body) > 0 {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), rdr)
	if err != nil {
		return 0, nil, err
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	if senderAuth {
		if c.apiKey == "" {
			return 0, nil, fmt.Errorf("PAYCREST_API_KEY is required for sender routes")
		}
		req.Header.Set("API-Key", c.apiKey)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()

	limited := io.LimitReader(res.Body, c.maxBody+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return res.StatusCode, nil, err
	}
	if int64(len(data)) > c.maxBody {
		return res.StatusCode, nil, fmt.Errorf("response exceeds PAYCREST_MAX_RESPONSE_BYTES (%d)", c.maxBody)
	}
	return res.StatusCode, data, nil
}

// Get performs a GET request. senderAuth adds API-Key for /v2/sender/* routes.
func (c *Client) Get(ctx context.Context, path string, query url.Values, senderAuth bool) (status int, body []byte, err error) {
	return c.do(ctx, http.MethodGet, path, query, nil, senderAuth)
}

// PostJSON sends a JSON body. senderAuth adds API-Key.
func (c *Client) PostJSON(ctx context.Context, path string, payload []byte, senderAuth bool) (status int, body []byte, err error) {
	return c.do(ctx, http.MethodPost, path, nil, payload, senderAuth)
}
