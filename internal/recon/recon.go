// Package recon verifica os invariantes contábeis do ledger.
package recon

import "fmt"

type TransferSums struct {
	TransferID string
	Debits     int64
	Credits    int64
}

type WalletBalance struct {
	AccountID string
	Balance   int64
}

type CurrencySum struct {
	Currency string
	Sum      int64
}

type Input struct {
	PerCurrency     []CurrencySum // Σ (créditos - débitos) por moeda — deve ser 0 em cada
	Transfers       []TransferSums
	Wallets         []WalletBalance
	GatewayLedger   int64 // saldo da conta gateway no ledger
	GatewayExpected int64 // esperado pela tabela payments
}

type Discrepancy struct {
	Kind   string
	Detail string
}

type Report struct {
	Healthy       bool
	Discrepancies []Discrepancy
}

// Check aplica os invariantes; função pura, testável isoladamente.
func Check(in Input) Report {
	var d []Discrepancy
	for _, c := range in.PerCurrency {
		if c.Sum != 0 {
			d = append(d, Discrepancy{Kind: "conservation", Detail: fmt.Sprintf("moeda %s: Σ saldos = %d (esperado 0)", c.Currency, c.Sum)})
		}
	}
	for _, t := range in.Transfers {
		if t.Debits != t.Credits {
			d = append(d, Discrepancy{Kind: "double_entry", Detail: fmt.Sprintf("transfer %s: débitos=%d créditos=%d", t.TransferID, t.Debits, t.Credits)})
		}
	}
	for _, w := range in.Wallets {
		if w.Balance < 0 {
			d = append(d, Discrepancy{Kind: "negative_wallet", Detail: fmt.Sprintf("wallet %s: saldo=%d", w.AccountID, w.Balance)})
		}
	}
	if in.GatewayLedger != in.GatewayExpected {
		d = append(d, Discrepancy{Kind: "gateway_mismatch", Detail: fmt.Sprintf("gateway ledger=%d esperado=%d", in.GatewayLedger, in.GatewayExpected)})
	}
	return Report{Healthy: len(d) == 0, Discrepancies: d}
}
