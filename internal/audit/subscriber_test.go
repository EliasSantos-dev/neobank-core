package audit_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/EliasSantos-dev/neobank-core/internal/audit"
	"github.com/EliasSantos-dev/neobank-core/internal/events"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/stretchr/testify/require"
)

func TestAuditPersistsIdempotent(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	bus, err := events.NewBus()
	require.NoError(t, err)
	defer bus.Close()
	ctx := context.Background()

	require.NoError(t, audit.Start(ctx, bus, s))

	env := events.Envelope{EventID: "11111111-1111-1111-1111-111111111111", Topic: "transfer.completed", Payload: map[string]any{"a": 1.0}}
	data, _ := json.Marshal(env)
	require.NoError(t, bus.Publish("transfer.completed", data))
	require.NoError(t, bus.Publish("transfer.completed", data)) // 2x

	require.Eventually(t, func() bool {
		n, _ := s.CountEventLog(ctx)
		return n == 1
	}, 2*time.Second, 20*time.Millisecond)
}
