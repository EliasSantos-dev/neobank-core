package recon_test

import (
	"context"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/EliasSantos-dev/neobank-core/internal/recon"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/stretchr/testify/require"
)

func TestRunner_HealthyOnRealData(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()
	ext, _ := s.CreateAccount(ctx, "BRL", ledger.External)
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	b, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	_, _ = s.Transfer(ctx, store.TransferParams{IdempotencyKey: "f", FromAccountID: ext.ID, ToAccountID: a.ID, Amount: 5000, Currency: "BRL"})
	_, _ = s.Transfer(ctx, store.TransferParams{IdempotencyKey: "t", FromAccountID: a.ID, ToAccountID: b.ID, Amount: 2000, Currency: "BRL"})

	rep, err := recon.Run(ctx, s)
	require.NoError(t, err)
	require.True(t, rep.Healthy, "divergências: %+v", rep.Discrepancies)
}
