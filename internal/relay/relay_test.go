package relay_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/EliasSantos-dev/neobank-core/internal/events"
	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/EliasSantos-dev/neobank-core/internal/relay"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/stretchr/testify/require"
)

func TestRelayPublishesAndMarks(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	bus, err := events.NewBus()
	require.NoError(t, err)
	defer bus.Close()
	ctx := context.Background()

	var mu sync.Mutex
	received := []events.Envelope{}
	var wg sync.WaitGroup
	wg.Add(1)
	require.NoError(t, bus.Subscribe("transfer.>", func(topic string, payload []byte) {
		var env events.Envelope
		_ = json.Unmarshal(payload, &env)
		mu.Lock()
		received = append(received, env)
		mu.Unlock()
		wg.Done()
	}))

	ext, _ := s.CreateAccount(ctx, "BRL", ledger.External)
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	_, _ = s.Transfer(ctx, store.TransferParams{IdempotencyKey: "t1", FromAccountID: ext.ID, ToAccountID: a.ID, Amount: 1000, Currency: "BRL"})

	r := relay.New(s, bus)
	n, err := r.RunOnce(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, n)

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
	mu.Lock()
	require.Len(t, received, 1)
	require.Equal(t, "transfer.completed", received[0].Topic)
	require.NotEmpty(t, received[0].EventID)
	mu.Unlock()

	// segunda chamada não republica
	n, err = r.RunOnce(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, n)
}
