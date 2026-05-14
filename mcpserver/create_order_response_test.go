package mcpserver

import (
	"strings"
	"testing"
)

func Test_appendCreateOrderProviderSummary_onrampFiat(t *testing.T) {
	raw := `{"status":"success","message":"ok","data":{"id":"00000000-0000-0000-0000-000000000001","providerAccount":{"institution":"Test Bank","accountIdentifier":"1234567890","accountName":"Jane Doe","validUntil":"2026-01-15T12:00:00Z","amountToTransfer":"50000","currency":"NGN"},"source":{"type":"fiat","currency":"NGN","refundAccount":{"institution":"OPay","institutionName":"OPay","accountIdentifier":"8166796176","accountName":"UKO SUNDAY ONAH"}}}}`
	out := string(appendCreateOrderProviderSummary([]byte(raw)))
	if !strings.Contains(out, "--- Paycrest providerAccount") {
		t.Fatal("expected provider header")
	}
	if !strings.Contains(out, "institution: Test Bank") || !strings.Contains(out, "amountToTransfer: 50000") {
		t.Fatalf("unexpected output:\n%s", out)
	}
	if !strings.Contains(out, "--- Refund (OPay) ---") || !strings.Contains(out, "accountIdentifier: 8166796176") {
		t.Fatalf("expected refund section:\n%s", out)
	}
	iRefund := strings.Index(out, "--- Refund (OPay) ---")
	iAction := strings.Index(out, "\nACTION\n")
	if iRefund < 0 || iAction < 0 || !(iRefund < iAction) {
		t.Fatalf("expected refund block before ACTION block:\n%s", out)
	}
	if !strings.Contains(out, `After you send the NGN, reply "paid" or "I have paid"`) {
		t.Fatalf("expected NGN paid line:\n%s", out)
	}
}

func Test_appendCreateOrderProviderSummary_onrampFiat_noRefundSource(t *testing.T) {
	raw := `{"status":"success","message":"ok","data":{"id":"00000000-0000-0000-0000-000000000001","providerAccount":{"institution":"Test Bank","accountIdentifier":"1234567890","accountName":"Jane Doe","validUntil":"2026-01-15T12:00:00Z","amountToTransfer":"50000","currency":"NGN"}}}`
	out := string(appendCreateOrderProviderSummary([]byte(raw)))
	if strings.Contains(out, "--- Refund (") {
		t.Fatal("did not expect refund section without source.refundAccount")
	}
	if !strings.Contains(out, "\nACTION\n") {
		t.Fatalf("expected ACTION block after pay-in:\n%s", out)
	}
}

func Test_appendCreateOrderProviderSummary_offrampCrypto(t *testing.T) {
	raw := `{"status":"success","message":"ok","data":{"providerAccount":{"network":"base","receiveAddress":"0xabc","validUntil":"2026-01-15T12:00:00Z"}}}`
	out := string(appendCreateOrderProviderSummary([]byte(raw)))
	if !strings.Contains(out, "receiveAddress: 0xabc") || !strings.Contains(out, "network: base") {
		t.Fatalf("unexpected output:\n%s", out)
	}
	if !strings.Contains(out, "\nACTION\n") {
		t.Fatalf("expected after-pay reminder:\n%s", out)
	}
}

func Test_appendCreateOrderProviderSummary_nonSuccess(t *testing.T) {
	raw := `{"status":"error","message":"bad","data":null}`
	out := appendCreateOrderProviderSummary([]byte(raw))
	if string(out) != raw {
		t.Fatal("expected unchanged body on error")
	}
}

func Test_appendCreateOrderWatchHint(t *testing.T) {
	raw := `{"status":"success","message":"ok","data":{"id":"11111111-1111-1111-1111-111111111111","status":"pending"}}`
	out := string(appendCreateOrderWatchHint([]byte(raw)))
	if !strings.Contains(out, "Status polling") || !strings.Contains(out, "11111111-1111-1111-1111-111111111111") {
		t.Fatalf("unexpected output:\n%s", out)
	}
	if !strings.Contains(out, "paycrest_watch_sender_order") {
		t.Fatal("expected watch tool name in hint")
	}
	if !strings.Contains(out, `After you send funds, reply "paid" or "I have paid"`) {
		t.Fatalf("expected ACTION paid line in hint:\n%s", out)
	}
}

func Test_appendCreateOrderWatchHint_singleActionBlockAfterProviderSummary(t *testing.T) {
	raw := []byte(`{"status":"success","message":"ok","data":{"id":"22222222-2222-2222-2222-222222222222","providerAccount":{"institution":"X","accountIdentifier":"1","accountName":"Y","validUntil":"2026-01-01T00:00:00Z","currency":"NGN"}}}`)
	withProvider := appendCreateOrderProviderSummary(raw)
	out := string(appendCreateOrderWatchHint(withProvider))
	if strings.Count(out, "\nACTION\n") != 1 {
		t.Fatalf("expected exactly one ACTION block, got:\n%s", out)
	}
	if !strings.Contains(out, "--- Status polling ---") {
		t.Fatal("expected status polling section")
	}
}

func TestExtractV2CreateOrderID(t *testing.T) {
	raw := []byte(`{"status":"success","data":{"id":"33333333-3333-3333-3333-333333333333"}}`)
	id, ok := ExtractV2CreateOrderID(raw)
	if !ok || id != "33333333-3333-3333-3333-333333333333" {
		t.Fatalf("id=%q ok=%v", id, ok)
	}
}
