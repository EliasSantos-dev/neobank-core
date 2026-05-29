package store_test

import (
	"context"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func mustUUID(s string) uuid.UUID { return uuid.MustParse(s) }

func TestPaymentCRUD(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)

	p, err := s.CreatePayment(ctx, store.PaymentParams{
		AccountID: a.ID, Kind: "deposit", Amount: 5000, Currency: "BRL", IdempotencyKey: "p1",
	})
	require.NoError(t, err)
	require.Equal(t, "pending", p.Status)

	require.NoError(t, s.SetPaymentProviderRef(ctx, p.ID, "ref-123"))
	require.NoError(t, s.MarkPaymentStatus(ctx, p.ID, "completed"))

	got, err := s.GetPayment(ctx, p.ID)
	require.NoError(t, err)
	require.Equal(t, "completed", got.Status)
	require.Equal(t, "ref-123", got.ProviderRef)

	first, err := s.RecordWebhookEvent(ctx, mustUUID("11111111-1111-1111-1111-111111111111"), p.ID)
	require.NoError(t, err)
	require.True(t, first)
	again, err := s.RecordWebhookEvent(ctx, mustUUID("11111111-1111-1111-1111-111111111111"), p.ID)
	require.NoError(t, err)
	require.False(t, again)
}

func TestPaymentLedgerMoves(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)

	require.NoError(t, s.CreditDeposit(ctx, mustUUID("aaaaaaaa-0000-0000-0000-000000000001"), a.ID, 5000))
	bal, _ := s.Balance(ctx, a.ID)
	require.EqualValues(t, 5000, bal)

	require.NoError(t, s.HoldWithdrawal(ctx, mustUUID("aaaaaaaa-0000-0000-0000-000000000002"), a.ID, 2000))
	bal, _ = s.Balance(ctx, a.ID)
	require.EqualValues(t, 3000, bal)

	require.NoError(t, s.RefundWithdrawal(ctx, mustUUID("aaaaaaaa-0000-0000-0000-000000000003"), a.ID, 2000))
	bal, _ = s.Balance(ctx, a.ID)
	require.EqualValues(t, 5000, bal)
}
