package gateway_test

import (
	"context"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/gateway"
	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestService_DepositConfirmCredits(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	svc := gateway.NewService(s, gateway.NewFakeProvider())
	ctx := context.Background()
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)

	p, err := svc.InitiateDeposit(ctx, a.ID, 5000, "dep-1")
	require.NoError(t, err)
	require.Equal(t, "pending", p.Status)
	bal, _ := s.Balance(ctx, a.ID)
	require.EqualValues(t, 0, bal)

	require.NoError(t, svc.HandleWebhook(ctx, gateway.WebhookEvent{
		EventID: uuid.New(), PaymentID: p.ID, Status: "succeeded",
	}))
	bal, _ = s.Balance(ctx, a.ID)
	require.EqualValues(t, 5000, bal)
}

func TestService_WithdrawHoldAndFailRefunds(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	svc := gateway.NewService(s, gateway.NewFakeProvider())
	ctx := context.Background()
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	dp, _ := svc.InitiateDeposit(ctx, a.ID, 5000, "dep-1")
	_ = svc.HandleWebhook(ctx, gateway.WebhookEvent{EventID: uuid.New(), PaymentID: dp.ID, Status: "succeeded"})

	wp, err := svc.InitiateWithdrawal(ctx, a.ID, 2000, "wd-1")
	require.NoError(t, err)
	bal, _ := s.Balance(ctx, a.ID)
	require.EqualValues(t, 3000, bal)

	require.NoError(t, svc.HandleWebhook(ctx, gateway.WebhookEvent{
		EventID: uuid.New(), PaymentID: wp.ID, Status: "failed",
	}))
	bal, _ = s.Balance(ctx, a.ID)
	require.EqualValues(t, 5000, bal)
}

func TestService_WebhookIdempotent(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	svc := gateway.NewService(s, gateway.NewFakeProvider())
	ctx := context.Background()
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	p, _ := svc.InitiateDeposit(ctx, a.ID, 5000, "dep-1")

	evt := gateway.WebhookEvent{EventID: uuid.New(), PaymentID: p.ID, Status: "succeeded"}
	require.NoError(t, svc.HandleWebhook(ctx, evt))
	require.NoError(t, svc.HandleWebhook(ctx, evt))
	bal, _ := s.Balance(ctx, a.ID)
	require.EqualValues(t, 5000, bal)
}

func TestService_InsufficientFundsNoHold(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	svc := gateway.NewService(s, gateway.NewFakeProvider())
	ctx := context.Background()
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	_, err := svc.InitiateWithdrawal(ctx, a.ID, 1000, "wd-1")
	require.ErrorIs(t, err, ledger.ErrInsufficientFunds)
}
