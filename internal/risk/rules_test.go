package risk

import "testing"

func TestRules(t *testing.T) {
	// drain: 95% do saldo dispara
	sig, ok := ruleDrain(Context{Amount: 9500, SenderBalance: 10000})
	if !ok || sig.Score == 0 {
		t.Fatal("drain deveria disparar")
	}
	if _, ok := ruleDrain(Context{Amount: 100, SenderBalance: 10000}); ok {
		t.Fatal("drain não deveria disparar com valor pequeno")
	}
	// new_recipient
	if _, ok := ruleNewRecipient(Context{RecipientIsNew: true}); !ok {
		t.Fatal("new_recipient deveria disparar")
	}
	if _, ok := ruleNewRecipient(Context{RecipientIsNew: false}); ok {
		t.Fatal("new_recipient não deveria disparar")
	}
	// velocity
	if _, ok := ruleVelocity(Context{RecentTransfers: 10}); !ok {
		t.Fatal("velocity deveria disparar")
	}
	if _, ok := ruleVelocity(Context{RecentTransfers: 1}); ok {
		t.Fatal("velocity não deveria disparar")
	}
	// large_amount
	if _, ok := ruleLargeAmount(Context{Amount: 2_000_000}); !ok {
		t.Fatal("large_amount deveria disparar")
	}
	if _, ok := ruleLargeAmount(Context{Amount: 5_000}); ok {
		t.Fatal("large_amount não deveria disparar")
	}
}
