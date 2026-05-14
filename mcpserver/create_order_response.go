package mcpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/paycrest/paycrest/sender-mcp/types"
)

// afterPaySentinel appears inside the user-facing ACTION block; used to avoid duplicating that block.
const afterPaySentinel = "\nACTION\n"

func writeAfterPayTypePaidReminder(b *strings.Builder, instructionLine string) {
	line := strings.TrimSpace(instructionLine)
	if line == "" {
		line = `After you send funds, reply "paid" or "I have paid"`
	}
	b.WriteString("\n\nACTION\n\n")
	b.WriteString(line)
	b.WriteString("\n")
}

// writeOnrampFiatRefundSection appends a human-readable refund block from data.source (V2 onramp).
// The Paycrest API echoes normalized accountName; show it so the assistant can place the PAID block below this section.
func writeOnrampFiatRefundSection(b *strings.Builder, dataJSON []byte) {
	var wrap struct {
		Source json.RawMessage `json:"source"`
	}
	if err := json.Unmarshal(dataJSON, &wrap); err != nil || len(wrap.Source) == 0 || string(wrap.Source) == "null" {
		return
	}
	var src struct {
		Type          string `json:"type"`
		RefundAccount struct {
			Institution       string `json:"institution"`
			InstitutionName   string `json:"institutionName"`
			AccountIdentifier string `json:"accountIdentifier"`
			AccountName       string `json:"accountName"`
		} `json:"refundAccount"`
	}
	if err := json.Unmarshal(wrap.Source, &src); err != nil {
		return
	}
	if !strings.EqualFold(strings.TrimSpace(src.Type), "fiat") {
		return
	}
	ra := src.RefundAccount
	id := strings.TrimSpace(ra.AccountIdentifier)
	if id == "" {
		return
	}
	label := strings.TrimSpace(ra.InstitutionName)
	if label == "" {
		label = strings.TrimSpace(ra.Institution)
	}
	if label == "" {
		label = "refund"
	}
	b.WriteString("\n\n--- Refund (")
	b.WriteString(label)
	b.WriteString(") ---\n")
	b.WriteString("accountIdentifier: " + id + "\n")
	if nm := strings.TrimSpace(ra.AccountName); nm != "" {
		b.WriteString("accountName: " + nm + "\n")
	}
	if inst := strings.TrimSpace(ra.Institution); inst != "" {
		b.WriteString("institution: " + inst + "\n")
	}
}

// appendCreateOrderProviderSummary appends a human-readable provider section after the raw API JSON
// when POST /v2/sender/orders returns success with a providerAccount (onramp or offramp).
func appendCreateOrderProviderSummary(apiBody []byte) []byte {
	var env struct {
		Status string          `json:"status"`
		Data   json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(apiBody, &env); err != nil {
		return apiBody
	}
	if !strings.EqualFold(env.Status, "success") || len(env.Data) == 0 || string(env.Data) == "null" {
		return apiBody
	}

	var dataObj struct {
		ProviderAccount json.RawMessage `json:"providerAccount"`
	}
	if err := json.Unmarshal(env.Data, &dataObj); err != nil || len(dataObj.ProviderAccount) == 0 || string(dataObj.ProviderAccount) == "null" {
		return apiBody
	}

	var b strings.Builder
	b.Write(apiBody)
	b.WriteString("\n\n--- Paycrest providerAccount (where to send funds) ---\n")

	var fiat types.FiatProviderAccount
	if err := json.Unmarshal(dataObj.ProviderAccount, &fiat); err == nil && fiat.Institution != "" {
		b.WriteString("institution: " + fiat.Institution + "\n")
		b.WriteString("accountIdentifier: " + fiat.AccountIdentifier + "\n")
		b.WriteString("accountName: " + fiat.AccountName + "\n")
		b.WriteString("validUntil: " + fiat.ValidUntil + "\n")
		if fiat.AmountToTransfer != "" {
			b.WriteString("amountToTransfer: " + fiat.AmountToTransfer + "\n")
		}
		if fiat.Currency != "" {
			b.WriteString("currency: " + fiat.Currency + "\n")
		}
		paidLine := `After you send funds, reply "paid" or "I have paid"`
		if cur := strings.TrimSpace(fiat.Currency); cur != "" {
			paidLine = fmt.Sprintf(`After you send the %s, reply "paid" or "I have paid"`, cur)
		}
		writeOnrampFiatRefundSection(&b, env.Data)
		writeAfterPayTypePaidReminder(&b, paidLine)
		return []byte(b.String())
	}

	var crypto types.CryptoProviderAccount
	if err := json.Unmarshal(dataObj.ProviderAccount, &crypto); err == nil && crypto.ReceiveAddress != "" {
		b.WriteString("network: " + crypto.Network + "\n")
		b.WriteString("receiveAddress: " + crypto.ReceiveAddress + "\n")
		b.WriteString("validUntil: " + crypto.ValidUntil + "\n")
		writeAfterPayTypePaidReminder(&b, `After you send funds, reply "paid" or "I have paid"`)
		return []byte(b.String())
	}

	b.WriteString("(unparsed providerAccount; see JSON above)\n")
	writeAfterPayTypePaidReminder(&b, `After you send funds, reply "paid" or "I have paid"`)
	return []byte(b.String())
}

// appendCreateOrderWatchHint appends manual polling instructions when embedded auto-watch is off or order id could not be read.
func appendCreateOrderWatchHint(apiBody []byte) []byte {
	dec := json.NewDecoder(bytes.NewReader(apiBody))
	var env struct {
		Status string          `json:"status"`
		Data   json.RawMessage `json:"data"`
	}
	if err := dec.Decode(&env); err != nil {
		return apiBody
	}
	if !strings.EqualFold(env.Status, "success") || len(env.Data) == 0 || string(env.Data) == "null" {
		return apiBody
	}
	var d struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(env.Data, &d); err != nil || strings.TrimSpace(d.ID) == "" {
		return apiBody
	}
	var b strings.Builder
	b.Write(apiBody)
	if !strings.Contains(string(apiBody), afterPaySentinel) {
		writeAfterPayTypePaidReminder(&b, `After you send funds, reply "paid" or "I have paid"`)
	}
	b.WriteString("\n\n--- Status polling ---\n")
	fmt.Fprintf(&b, "Order id: %s\n", d.ID)
	fmt.Fprintf(&b, "Call paycrest_watch_sender_order with {\"id\":\"%s\"} (optional: \"poll_interval_sec\":10, \"max_wait_sec\":3600; max %d) to poll GET /v2/sender/orders until settled, cancelled, refunded, or expired.\n", d.ID, types.WatchWaitMaxSec)
	return []byte(b.String())
}

// ExtractV2CreateOrderID returns data.id from the first JSON object in body (e.g. create-order response, optionally with trailing non-JSON text).
func ExtractV2CreateOrderID(body []byte) (id string, ok bool) {
	dec := json.NewDecoder(bytes.NewReader(body))
	var env struct {
		Status string          `json:"status"`
		Data   json.RawMessage `json:"data"`
	}
	if err := dec.Decode(&env); err != nil {
		return "", false
	}
	if !strings.EqualFold(env.Status, "success") || len(env.Data) == 0 || string(env.Data) == "null" {
		return "", false
	}
	var d struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(env.Data, &d); err != nil || strings.TrimSpace(d.ID) == "" {
		return "", false
	}
	return strings.TrimSpace(d.ID), true
}
