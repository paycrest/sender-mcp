package mcpserver

import (
	"strings"
	"testing"
)

func Test_appendCreateOrderProviderSummary_onrampFiat(t *testing.T) {
	raw := `{"status":"success","message":"ok","data":{"id":"00000000-0000-0000-0000-000000000001","providerAccount":{"institution":"Test Bank","accountIdentifier":"1234567890","accountName":"Jane Doe","validUntil":"2026-01-15T12:00:00Z","amountToTransfer":"50000","currency":"NGN"}}}`
	out := string(appendCreateOrderProviderSummary([]byte(raw)))
	if !strings.Contains(out, "--- Paycrest providerAccount") {
		t.Fatal("expected provider header")
	}
	if !strings.Contains(out, "institution: Test Bank") || !strings.Contains(out, "amountToTransfer: 50000") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func Test_appendCreateOrderProviderSummary_offrampCrypto(t *testing.T) {
	raw := `{"status":"success","message":"ok","data":{"providerAccount":{"network":"base","receiveAddress":"0xabc","validUntil":"2026-01-15T12:00:00Z"}}}`
	out := string(appendCreateOrderProviderSummary([]byte(raw)))
	if !strings.Contains(out, "receiveAddress: 0xabc") || !strings.Contains(out, "network: base") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func Test_appendCreateOrderProviderSummary_nonSuccess(t *testing.T) {
	raw := `{"status":"error","message":"bad","data":null}`
	out := appendCreateOrderProviderSummary([]byte(raw))
	if string(out) != raw {
		t.Fatal("expected unchanged body on error")
	}
}
