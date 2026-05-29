package recon

import "testing"

func TestCheck_Healthy(t *testing.T) {
	in := Input{
		GlobalSum: 0,
		Transfers: []TransferSums{{TransferID: "t1", Debits: 100, Credits: 100}},
		Wallets:   []WalletBalance{{AccountID: "w1", Balance: 50}},
	}
	r := Check(in)
	if !r.Healthy {
		t.Fatalf("esperava saudável, veio %+v", r.Discrepancies)
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
		GlobalSum: 5,
		Transfers: []TransferSums{{TransferID: "t1", Debits: 100, Credits: 90}},
		Wallets:   []WalletBalance{{AccountID: "w1", Balance: -10}},
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
