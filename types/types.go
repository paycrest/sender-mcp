package types

import "time"

// Bounds for paycrest_watch_sender_order max_wait_sec and embedded create-order poll (seconds).
const (
	WatchWaitMinSec = 10
	WatchWaitMaxSec = 3600
	// WatchDefaultMaxWaitSec is the default max_wait_sec when the client omits it (poll until terminal or this cap).
	WatchDefaultMaxWaitSec = 3600
)

// Config holds runtime settings for the MCP server and Paycrest HTTP client.
// PAYCREST_API_KEY is expected from the MCP host env (per user), not from a shared .env in repos.
type Config struct {
	BaseURL      string
	APIKey       string // sent as API-Key for /v2/sender/* (see aggregator DynamicAuthMiddleware)
	HTTPTimeout  time.Duration
	MaxRespBytes int64
	// AutoWatchAfterCreate runs order status polling after HTTP 201 from paycrest_create_order (default false for fast response). Set PAYCREST_AUTO_WATCH_AFTER_CREATE=true|1|yes|on to enable embedded poll + MCP progress.
	AutoWatchAfterCreate bool
	// CreateOrderMaxWaitSec is max seconds for that embedded poll (clamped WatchWaitMinSec–WatchWaitMaxSec). From PAYCREST_CREATE_ORDER_MAX_WAIT_SEC; default 900.
	CreateOrderMaxWaitSec int
}

// Empty is a zero-size tool input for tools that take no arguments.
type Empty struct{}

// InstitutionsIn is the input for paycrest_get_institutions.
type InstitutionsIn struct {
	CurrencyCode string `json:"currency_code" jsonschema:"required fiat currency code, e.g. NGN"`
}

// RatesIn is the input for paycrest_get_rates.
type RatesIn struct {
	Network    string `json:"network" jsonschema:"required network identifier, e.g. base"`
	Token      string `json:"token" jsonschema:"required token symbol"`
	Amount     string `json:"amount" jsonschema:"required amount as decimal string"`
	Fiat       string `json:"fiat" jsonschema:"required fiat currency code"`
	ProviderID string `json:"provider_id,omitempty" jsonschema:"optional 8-char provider id"`
	Side       string `json:"side,omitempty" jsonschema:"optional buy or sell"`
}

// ListOrdersIn is the input for paycrest_list_sender_orders.
type ListOrdersIn struct {
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
	Ordering string `json:"ordering,omitempty"`
	Search   string `json:"search,omitempty"`
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	Export   string `json:"export,omitempty"`
}

// GetOrderIn is the input for paycrest_get_sender_order.
type GetOrderIn struct {
	ID string `json:"id" jsonschema:"required payment order id"`
}

// WatchSenderOrderIn is the input for paycrest_watch_sender_order.
type WatchSenderOrderIn struct {
	ID string `json:"id" jsonschema:"required payment order id (UUID from create or get order)"`
	// PollIntervalSec is seconds between GET polls (clamped 2–30). Default 3.
	PollIntervalSec int `json:"poll_interval_sec,omitempty"`
	// MaxWaitSec is maximum seconds to keep polling (clamped WatchWaitMinSec–WatchWaitMaxSec). Default WatchDefaultMaxWaitSec (1 hour).
	MaxWaitSec int `json:"max_wait_sec,omitempty"`
}

// FiatProviderAccount matches Paycrest onramp pay-in / noblocks providerAccount JSON fields.
type FiatProviderAccount struct {
	Institution       string `json:"institution"`
	AccountIdentifier string `json:"accountIdentifier"`
	AccountName       string `json:"accountName"`
	ValidUntil        string `json:"validUntil"`
	AmountToTransfer  string `json:"amountToTransfer,omitempty"`
	Currency          string `json:"currency,omitempty"`
}

// CryptoProviderAccount matches offramp providerAccount crypto receive JSON.
type CryptoProviderAccount struct {
	Network        string `json:"network"`
	ReceiveAddress string `json:"receiveAddress"`
	ValidUntil     string `json:"validUntil"`
}
