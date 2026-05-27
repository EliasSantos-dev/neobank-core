package ledger

import "testing"

func TestAccountType_AllowsNegative(t *testing.T) {
	if Wallet.AllowsNegative() {
		t.Fatal("Wallet não pode permitir saldo negativo")
	}
	if !External.AllowsNegative() {
		t.Fatal("External deve permitir saldo negativo (fonte de funding)")
	}
}

func TestDirection_Valid(t *testing.T) {
	if !Debit.Valid() || !Credit.Valid() {
		t.Fatal("Debit/Credit devem ser válidos")
	}
	if Direction("foo").Valid() {
		t.Fatal("direção inválida não pode ser válida")
	}
}
