package paycrest

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ParseRetryAfter interprets the Retry-After response header (seconds or HTTP-date per RFC 7231).
func ParseRetryAfter(h http.Header) (d time.Duration, ok bool) {
	if h == nil {
		return 0, false
	}
	ra := strings.TrimSpace(h.Get("Retry-After"))
	if ra == "" {
		return 0, false
	}
	if sec, err := strconv.Atoi(ra); err == nil && sec >= 0 {
		return time.Duration(sec) * time.Second, true
	}
	if t, err := http.ParseTime(ra); err == nil {
		dur := time.Until(t)
		if dur < 0 {
			dur = 0
		}
		return dur, true
	}
	return 0, false
}
