package paycrest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paycrest/paycrest/sender-mcp/types"
)

func TestClient_Get_public(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/currencies" {
			t.Fatalf("path: %s", r.URL.Path)
		}
		if r.Header.Get("API-Key") != "" {
			t.Fatal("unexpected API-Key on public route")
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(types.Config{
		BaseURL:      srv.URL,
		HTTPTimeout:  5 * time.Second,
		MaxRespBytes: 1 << 20,
	})
	status, body, err := c.Get(context.Background(), "/v2/currencies", nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Fatalf("status=%d", status)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("body=%q", body)
	}
}

func TestClient_Get_sender_requires_key(t *testing.T) {
	c := NewClient(types.Config{
		BaseURL:      "http://example.com",
		HTTPTimeout:  5 * time.Second,
		MaxRespBytes: 1 << 20,
		APIKey:       "",
	})
	_, _, err := c.Get(context.Background(), "/v2/sender/orders", nil, true)
	if err == nil {
		t.Fatal("expected error without API key")
	}
}
