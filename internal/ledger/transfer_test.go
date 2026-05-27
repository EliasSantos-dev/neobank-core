package ledger

import (
	"testing"

	"github.com/google/uuid"
)

func TestBuildEntries_Balanced(t *testing.T) {
	from := Account{ID: uuid.New(), Currency: "BRL", Type: Wallet}
	to := Account{ID: uuid.New(), Currency: "BRL", Type: Wallet}
	tid := uuid.New()

	entries, err := BuildEntries(tid, from, to, 5000)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("esperava 2 lançamentos, veio %d", len(entries))
	}
	var deb, cred int64
	for _, e := range entries {
		if e.Direction == Debit {
			deb += e.Amount
		} else {
			cred += e.Amount
		}
	}
	if deb != cred {
		t.Fatalf("não balanceou: deb=%d cred=%d", deb, cred)
	}
}

func TestBuildEntries_Errors(t *testing.T) {
	a := Account{ID: uuid.New(), Currency: "BRL", Type: Wallet}
	usd := Account{ID: uuid.New(), Currency: "USD", Type: Wallet}

	if _, err := BuildEntries(uuid.New(), a, usd, 10); err != ErrCurrencyMismatch {
		t.Fatalf("esperava ErrCurrencyMismatch, veio %v", err)
	}
	if _, err := BuildEntries(uuid.New(), a, a, 10); err != ErrSameAccount {
		t.Fatalf("esperava ErrSameAccount, veio %v", err)
	}
	b := Account{ID: uuid.New(), Currency: "BRL", Type: Wallet}
	if _, err := BuildEntries(uuid.New(), a, b, 0); err != ErrInvalidAmount {
		t.Fatalf("esperava ErrInvalidAmount, veio %v", err)
	}
}
