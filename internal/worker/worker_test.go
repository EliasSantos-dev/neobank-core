package worker_test

import (
	"context"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/EliasSantos-dev/neobank-core/internal/risk"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	"github.com/EliasSantos-dev/neobank-core/internal/worker"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (*store.Store, *worker.Worker, context.Context) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	w := worker.New(s, risk.NewEngine(risk.RuleReasoningAdvisor{}, 50))
	return s, w, context.Background()
}

func TestProcessOnce_LowRiskCompletes(t *testing.T) {
	s, w, ctx := setup(t)
	ext, _ := s.CreateAccount(ctx, "BRL", ledger.External)
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	b, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	_, _ = s.Transfer(ctx, store.TransferParams{IdempotencyKey: "f", FromAccountID: ext.ID, ToAccountID: a.ID, Amount: 100000, Currency: "BRL"})
	_, _ = s.Transfer(ctx, store.TransferParams{IdempotencyKey: "h1", FromAccountID: a.ID, ToAccountID: b.ID, Amount: 1000, Currency: "BRL"})
	_, _ = s.Transfer(ctx, store.TransferParams{IdempotencyKey: "h2", FromAccountID: a.ID, ToAccountID: b.ID, Amount: 1100, Currency: "BRL"})
	_, _ = s.Transfer(ctx, store.TransferParams{IdempotencyKey: "h3", FromAccountID: a.ID, ToAccountID: b.ID, Amount: 900, Currency: "BRL"})
	it, _ := s.CreateIntent(ctx, store.IntentParams{IdempotencyKey: "i1", FromAccountID: a.ID, ToAccountID: b.ID, Amount: 1000, Currency: "BRL"})

	n, err := w.ProcessOnce(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	got, _ := s.GetIntent(ctx, it.ID)
	require.Equal(t, "completed", got.Status)
}

func TestProcessOnce_HighRiskUnderReview(t *testing.T) {
	s, w, ctx := setup(t)
	ext, _ := s.CreateAccount(ctx, "BRL", ledger.External)
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	b, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	_, _ = s.Transfer(ctx, store.TransferParams{IdempotencyKey: "f", FromAccountID: ext.ID, ToAccountID: a.ID, Amount: 10000, Currency: "BRL"})
	it, _ := s.CreateIntent(ctx, store.IntentParams{IdempotencyKey: "i1", FromAccountID: a.ID, ToAccountID: b.ID, Amount: 9500, Currency: "BRL"})

	n, err := w.ProcessOnce(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	got, _ := s.GetIntent(ctx, it.ID)
	require.Equal(t, "under_review", got.Status)
	balB, _ := s.Balance(ctx, b.ID)
	require.EqualValues(t, 0, balB)
}
