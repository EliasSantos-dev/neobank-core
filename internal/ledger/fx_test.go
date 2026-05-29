package ledger

import "testing"

func TestFxAccounts(t *testing.T) {
	if FxBRL.String() != "00000000-0000-0000-0000-000000000010" {
		t.Fatalf("FxBRL: %s", FxBRL)
	}
	if FxUSD.String() != "00000000-0000-0000-0000-000000000011" {
		t.Fatalf("FxUSD: %s", FxUSD)
	}
	if FxEUR.String() != "00000000-0000-0000-0000-000000000012" {
		t.Fatalf("FxEUR: %s", FxEUR)
	}
	if _, ok := FxAccount("JPY"); ok {
		t.Fatal("JPY não deveria ser suportada")
	}
}
