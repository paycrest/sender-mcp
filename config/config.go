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
		// Production aggregator — senders need not set PAYCREST_BASE_URL in Cursor mcp.json.
		// Override via env/.env for local aggregator (e.g. http://127.0.0.1:8080).
		base = "https://api.paycrest.io"
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

	autoWatch := false
	switch strings.ToLower(strings.TrimSpace(os.Getenv("PAYCREST_AUTO_WATCH_AFTER_CREATE"))) {
	case "1", "true", "yes", "on":
		autoWatch = true
	}

	createMaxWait := 900
	if v := strings.TrimSpace(os.Getenv("PAYCREST_CREATE_ORDER_MAX_WAIT_SEC")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			createMaxWait = n
			if createMaxWait < types.WatchWaitMinSec {
				createMaxWait = types.WatchWaitMinSec
			}
			if createMaxWait > types.WatchWaitMaxSec {
				createMaxWait = types.WatchWaitMaxSec
			}
		}
	}

	return types.Config{
		BaseURL:               base,
		APIKey:                strings.TrimSpace(os.Getenv("PAYCREST_API_KEY")),
		HTTPTimeout:           timeout,
		MaxRespBytes:          maxBytes,
		AutoWatchAfterCreate:  autoWatch,
		CreateOrderMaxWaitSec: createMaxWait,
	}
}
