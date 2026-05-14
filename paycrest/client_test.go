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

func TestParseRetryAfter(t *testing.T) {
	t.Parallel()
	h := http.Header{"Retry-After": {"42"}}
	d, ok := ParseRetryAfter(h)
	if !ok || d != 42*time.Second {
		t.Fatalf("seconds: got %v ok=%v", d, ok)
	}
	if _, ok := ParseRetryAfter(nil); ok {
		t.Fatal("nil header")
	}
	if _, ok := ParseRetryAfter(http.Header{}); ok {
		t.Fatal("empty")
	}
	future := time.Now().UTC().Add(90 * time.Second).Format(http.TimeFormat)
	h2 := http.Header{"Retry-After": {future}}
	d2, ok2 := ParseRetryAfter(h2)
	if !ok2 || d2 < 85*time.Second || d2 > 95*time.Second {
		t.Fatalf("http-date: got %v ok=%v", d2, ok2)
	}
}
