package events_test

import (
	"sync"
	"testing"
	"time"

	"github.com/EliasSantos-dev/neobank-core/internal/events"
	"github.com/stretchr/testify/require"
)

func TestBusPublishSubscribe(t *testing.T) {
	bus, err := events.NewBus()
	require.NoError(t, err)
	defer bus.Close()

	var mu sync.Mutex
	got := map[string][]byte{}
	var wg sync.WaitGroup
	wg.Add(1)
	require.NoError(t, bus.Subscribe("transfer.>", func(topic string, payload []byte) {
		mu.Lock()
		got[topic] = payload
		mu.Unlock()
		wg.Done()
	}))

	require.NoError(t, bus.Publish("transfer.completed", []byte(`{"x":1}`)))

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout esperando o evento")
	}
	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, []byte(`{"x":1}`), got["transfer.completed"])
}
