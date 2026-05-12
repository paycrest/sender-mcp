package mcpserver

import (
	"encoding/json"
	"strings"

	"github.com/paycrest/paycrest/sender-mcp/types"
)

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
		return []byte(b.String())
	}

	var crypto types.CryptoProviderAccount
	if err := json.Unmarshal(dataObj.ProviderAccount, &crypto); err == nil && crypto.ReceiveAddress != "" {
		b.WriteString("network: " + crypto.Network + "\n")
		b.WriteString("receiveAddress: " + crypto.ReceiveAddress + "\n")
		b.WriteString("validUntil: " + crypto.ValidUntil + "\n")
		return []byte(b.String())
	}

	b.WriteString("(unparsed providerAccount; see JSON above)\n")
	return []byte(b.String())
}
