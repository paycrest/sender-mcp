package mcpserver

import "testing"

func Test_isTerminalOrderStatus(t *testing.T) {
	for _, s := range []string{"settled", "Settled", "cancelled", "refunded", "expired"} {
		if !isTerminalOrderStatus(s) {
			t.Fatalf("expected terminal: %q", s)
		}
	}
	for _, s := range []string{"pending", "initiated", "fulfilling", "", "x"} {
		if isTerminalOrderStatus(s) {
			t.Fatalf("expected non-terminal: %q", s)
		}
	}
}

func Test_parseV2OrderGETSnapshot(t *testing.T) {
	body := []byte(`{"status":"success","data":{"status":"pending","txHash":"0xabc"}}`)
	snap, ok := parseV2OrderGETSnapshot(body)
	if !ok {
		t.Fatal("expected ok")
	}
	if snap.EnvelopeStatus != "success" || snap.OrderStatus != "pending" || snap.TxHash != "0xabc" {
		t.Fatalf("snap=%+v", snap)
	}
}
