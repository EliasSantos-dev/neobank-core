package store_test

import (
	"context"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/stretchr/testify/require"
)

func TestOutboxFetchMark(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()

	id, err := s.WriteEventDirect(ctx, "transfer.completed", []byte(`{"a":1}`))
	require.NoError(t, err)

	pend, err := s.FetchUnpublished(ctx, 10)
	require.NoError(t, err)
	require.Len(t, pend, 1)
	require.Equal(t, id, pend[0].ID)
	require.Equal(t, "transfer.completed", pend[0].Topic)

	require.NoError(t, s.MarkPublished(ctx, id))
	pend, _ = s.FetchUnpublished(ctx, 10)
	require.Len(t, pend, 0)
}

func TestInsertEventLogIdempotent(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()
	id, _ := s.WriteEventDirect(ctx, "transfer.completed", []byte(`{"a":1}`))

	require.NoError(t, s.InsertEventLog(ctx, id, "transfer.completed", []byte(`{"a":1}`)))
	require.NoError(t, s.InsertEventLog(ctx, id, "transfer.completed", []byte(`{"a":1}`)))
	n, err := s.CountEventLog(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, n)
}

func TestTransferEmitsOutboxEvent(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()
	ext, _ := s.CreateAccount(ctx, "BRL", ledger.External)
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	_, err := s.Transfer(ctx, store.TransferParams{IdempotencyKey: "t1", FromAccountID: ext.ID, ToAccountID: a.ID, Amount: 1000, Currency: "BRL"})
	require.NoError(t, err)

	pend, err := s.FetchUnpublished(ctx, 10)
	require.NoError(t, err)
	require.Len(t, pend, 1)
	require.Equal(t, "transfer.completed", pend[0].Topic)
}

func TestIntentTransitionsEmitEvents(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	b, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)

	it1, _ := s.CreateIntent(ctx, store.IntentParams{IdempotencyKey: "i1", FromAccountID: a.ID, ToAccountID: b.ID, Amount: 100, Currency: "BRL"})
	require.NoError(t, s.MarkUnderReview(ctx, it1.ID, 60, "high", []byte(`{}`)))

	it2, _ := s.CreateIntent(ctx, store.IntentParams{IdempotencyKey: "i2", FromAccountID: a.ID, ToAccountID: b.ID, Amount: 100, Currency: "BRL"})
	require.NoError(t, s.RejectIntent(ctx, it2.ID))

	pend, _ := s.FetchUnpublished(ctx, 10)
	topics := map[string]bool{}
	for _, e := range pend {
		topics[e.Topic] = true
	}
	require.True(t, topics["intent.under_review"])
	require.True(t, topics["intent.rejected"])
}
