package store_test

import (
	"context"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/stretchr/testify/require"
)

func TestCreditInput(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()
	u, _ := s.CreateUser(ctx, "a@b.com", "h")
	brl, _ := s.GetOrCreateWallet(ctx, u.ID, "BRL")

	p, _ := s.CreatePayment(ctx, store.PaymentParams{AccountID: brl, Kind: "deposit", Amount: 5000, Currency: "BRL", IdempotencyKey: "d1"})
	require.NoError(t, s.CreditDeposit(ctx, p.ID, brl, 5000))
	require.NoError(t, s.MarkPaymentStatus(ctx, p.ID, "completed"))

	it, _ := s.CreateIntent(ctx, store.IntentParams{IdempotencyKey: "i1", FromAccountID: brl, ToAccountID: ledger.FxBRL, Amount: 100, Currency: "BRL"})
	require.NoError(t, s.MarkUnderReview(ctx, it.ID, 60, "high", []byte("{}")))

	in, err := s.CreditInput(ctx, u.ID)
	require.NoError(t, err)
	require.Equal(t, 1, in.DepositCount)
	require.EqualValues(t, 5000, in.DepositTotal)
	require.EqualValues(t, 5000, in.Balance)
	require.Equal(t, 1, in.RiskFlags)
	require.GreaterOrEqual(t, in.AccountAgeDays, 0)
}
