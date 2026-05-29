package recon

import "testing"

func TestCheck_Healthy(t *testing.T) {
	in := Input{
		PerCurrency: []CurrencySum{{"BRL", 0}, {"USD", 0}},
		Transfers:   []TransferSums{{TransferID: "t1", Debits: 100, Credits: 100}},
		Wallets:     []WalletBalance{{AccountID: "w1", Balance: 50}},
	}
	r := Check(in)
	if !r.Healthy {
		t.Fatalf("esperava saudável, veio %+v", r.Discrepancies)
	}
}

func TestCheck_PerCurrencyConservation(t *testing.T) {
	healthy := Input{PerCurrency: []CurrencySum{{"BRL", 0}, {"USD", 0}}}
	if !Check(healthy).Healthy {
		t.Fatal("esperava saudável")
	}
	broken := Input{PerCurrency: []CurrencySum{{"BRL", 0}, {"USD", 5}}}
	r := Check(broken)
	if r.Healthy {
		t.Fatal("esperava não-saudável")
	}
	found := false
	for _, d := range r.Discrepancies {
		if d.Kind == "conservation" {
			found = true
		}
	}
	if !found {
		t.Fatalf("esperava conservation: %+v", r.Discrepancies)
	}
}

func TestCheck_GatewayMismatch(t *testing.T) {
	in := Input{GatewayLedger: 100, GatewayExpected: 100}
	if !Check(in).Healthy {
		t.Fatal("esperava saudável quando gateway bate")
	}
	in.GatewayExpected = 999
	r := Check(in)
	kinds := map[string]bool{}
	for _, d := range r.Discrepancies {
		kinds[d.Kind] = true
	}
	if !kinds["gateway_mismatch"] {
		t.Fatalf("esperava gateway_mismatch: %+v", r.Discrepancies)
	}
}

func TestCheck_Discrepancies(t *testing.T) {
	in := Input{
		PerCurrency: []CurrencySum{{"BRL", 5}},
		Transfers:   []TransferSums{{TransferID: "t1", Debits: 100, Credits: 90}},
		Wallets:     []WalletBalance{{AccountID: "w1", Balance: -10}},
	}
	r := Check(in)
	if r.Healthy {
		t.Fatal("esperava não-saudável")
	}
	kinds := map[string]bool{}
	for _, d := range r.Discrepancies {
		kinds[d.Kind] = true
	}
	if !kinds["conservation"] || !kinds["double_entry"] || !kinds["negative_wallet"] {
		t.Fatalf("faltou alguma divergência: %+v", r.Discrepancies)
	}
}
