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
