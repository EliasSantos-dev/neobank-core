package store_test

import (
	"context"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/store"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/stretchr/testify/require"
)

func TestUserWallets(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()
	u, err := s.CreateUser(ctx, "a@b.com", "h")
	require.NoError(t, err)

	brl, err := s.GetOrCreateWallet(ctx, u.ID, "BRL")
	require.NoError(t, err)
	require.Equal(t, u.WalletAccountID, brl) // reaproveita a wallet BRL do usuário

	usd1, err := s.GetOrCreateWallet(ctx, u.ID, "USD")
	require.NoError(t, err)
	usd2, err := s.GetOrCreateWallet(ctx, u.ID, "USD")
	require.NoError(t, err)
	require.Equal(t, usd1, usd2) // idempotente

	ws, err := s.ListWallets(ctx, u.ID)
	require.NoError(t, err)
	require.Len(t, ws, 2) // BRL + USD
}
