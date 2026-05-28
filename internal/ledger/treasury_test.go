package ledger

import "testing"

func TestTreasuryBRL(t *testing.T) {
	if TreasuryBRL.String() != "00000000-0000-0000-0000-000000000001" {
		t.Fatalf("UUID inesperado: %s", TreasuryBRL)
	}
}
