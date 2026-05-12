package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/paycrest/paycrest/sender-mcp/types"
)

// Load reads configuration from the environment.
func Load() types.Config {
	base := strings.TrimSpace(os.Getenv("PAYCREST_BASE_URL"))
	base = strings.TrimRight(base, "/")
	if base == "" {
		base = "http://127.0.0.1:8080"
	}

	timeout := 60 * time.Second
	if v := strings.TrimSpace(os.Getenv("PAYCREST_HTTP_TIMEOUT")); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			timeout = d
		}
	}

	maxBytes := int64(10 << 20) // 10 MiB
	if v := strings.TrimSpace(os.Getenv("PAYCREST_MAX_RESPONSE_BYTES")); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			maxBytes = n
		}
	}

	return types.Config{
		BaseURL:      base,
		APIKey:       strings.TrimSpace(os.Getenv("PAYCREST_API_KEY")),
		HTTPTimeout:  timeout,
		MaxRespBytes: maxBytes,
	}
}
