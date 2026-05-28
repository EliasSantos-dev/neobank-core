package store_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/stretchr/testify/require"
)

func TestIntentLifecycle(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()
	ext, _ := s.CreateAccount(ctx, "BRL", ledger.External)
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	b, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	_, err := s.Transfer(ctx, store.TransferParams{IdempotencyKey: "f", FromAccountID: ext.ID, ToAccountID: a.ID, Amount: 10000, Currency: "BRL"})
	require.NoError(t, err)

	intent, err := s.CreateIntent(ctx, store.IntentParams{
		IdempotencyKey: "i1", FromAccountID: a.ID, ToAccountID: b.ID, Amount: 2000, Currency: "BRL",
	})
	require.NoError(t, err)
	require.Equal(t, "pending", intent.Status)

	pend, err := s.ListPending(ctx, 10)
	require.NoError(t, err)
	require.Len(t, pend, 1)

	reasons, _ := json.Marshal([]string{"drain"})
	require.NoError(t, s.MarkUnderReview(ctx, intent.ID, 60, "high", reasons))
	got, err := s.GetIntent(ctx, intent.ID)
	require.NoError(t, err)
	require.Equal(t, "under_review", got.Status)
	require.EqualValues(t, 60, got.RiskScore)

	pend, _ = s.ListPending(ctx, 10)
	require.Len(t, pend, 0)
}

func TestCompleteIntent(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()
	ext, _ := s.CreateAccount(ctx, "BRL", ledger.External)
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	b, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	_, _ = s.Transfer(ctx, store.TransferParams{IdempotencyKey: "f", FromAccountID: ext.ID, ToAccountID: a.ID, Amount: 10000, Currency: "BRL"})
	it, _ := s.CreateIntent(ctx, store.IntentParams{IdempotencyKey: "i1", FromAccountID: a.ID, ToAccountID: b.ID, Amount: 2000, Currency: "BRL"})

	require.NoError(t, s.CompleteIntent(ctx, it.ID))
	got, _ := s.GetIntent(ctx, it.ID)
	require.Equal(t, "completed", got.Status)
	require.NotNil(t, got.TransferID)

	balB, _ := s.Balance(ctx, b.ID)
	require.EqualValues(t, 2000, balB)

	require.NoError(t, s.CompleteIntent(ctx, it.ID)) // idempotente
	balB2, _ := s.Balance(ctx, b.ID)
	require.EqualValues(t, 2000, balB2)
}

func TestRejectIntent(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	b, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	it, _ := s.CreateIntent(ctx, store.IntentParams{IdempotencyKey: "i1", FromAccountID: a.ID, ToAccountID: b.ID, Amount: 2000, Currency: "BRL"})
	require.NoError(t, s.RejectIntent(ctx, it.ID))
	got, _ := s.GetIntent(ctx, it.ID)
	require.Equal(t, "rejected", got.Status)
	balB, _ := s.Balance(ctx, b.ID)
	require.EqualValues(t, 0, balB)
}

func TestRiskInputs(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()
	ext, _ := s.CreateAccount(ctx, "BRL", ledger.External)
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	b, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	_, _ = s.Transfer(ctx, store.TransferParams{IdempotencyKey: "f", FromAccountID: ext.ID, ToAccountID: a.ID, Amount: 10000, Currency: "BRL"})
	_, _ = s.Transfer(ctx, store.TransferParams{IdempotencyKey: "h", FromAccountID: a.ID, ToAccountID: b.ID, Amount: 500, Currency: "BRL"})

	isNew, err := s.RecipientIsNew(ctx, a.ID, b.ID)
	require.NoError(t, err)
	require.False(t, isNew)

	c, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	isNew, _ = s.RecipientIsNew(ctx, a.ID, c.ID)
	require.True(t, isNew)

	hist, err := s.DebitHistory(ctx, a.ID, 100)
	require.NoError(t, err)
	require.Contains(t, hist, int64(500))
}
