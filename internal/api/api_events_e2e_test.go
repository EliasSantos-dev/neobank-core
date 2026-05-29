package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/EliasSantos-dev/neobank-core/internal/audit"
	"github.com/EliasSantos-dev/neobank-core/internal/events"
	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/EliasSantos-dev/neobank-core/internal/relay"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/stretchr/testify/require"
)

func TestE2E_OutboxRelayAudit(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	bus, err := events.NewBus()
	require.NoError(t, err)
	defer bus.Close()
	ctx := context.Background()
	require.NoError(t, audit.Start(ctx, bus, s))

	ext, _ := s.CreateAccount(ctx, "BRL", ledger.External)
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	_, _ = s.Transfer(ctx, store.TransferParams{IdempotencyKey: "f", FromAccountID: ext.ID, ToAccountID: a.ID, Amount: 5000, Currency: "BRL"})

	r := relay.New(s, bus)
	_, err = r.RunOnce(ctx)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		n, _ := s.CountEventLog(ctx)
		return n == 1
	}, 2*time.Second, 20*time.Millisecond)
}
