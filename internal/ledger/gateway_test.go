package ledger

import "testing"

func TestGatewayBRL(t *testing.T) {
	if GatewayBRL.String() != "00000000-0000-0000-0000-000000000002" {
		t.Fatalf("UUID inesperado: %s", GatewayBRL)
	}
}
