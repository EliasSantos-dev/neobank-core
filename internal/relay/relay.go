// Package relay publica eventos do outbox no barramento (at-least-once).
package relay

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/EliasSantos-dev/neobank-core/internal/events"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
)

type Relay struct {
	s   *store.Store
	bus *events.Bus
}

func New(s *store.Store, bus *events.Bus) *Relay {
	return &Relay{s: s, bus: bus}
}

// RunOnce publica todos os eventos não-publicados do outbox e os marca.
// Retorna quantos publicou. At-least-once: se cair entre publicar e marcar,
// o evento é reenviado na próxima rodada (consumidores idempotentes).
func (r *Relay) RunOnce(ctx context.Context) (int, error) {
	pending, err := r.s.FetchUnpublished(ctx, 100)
	if err != nil {
		return 0, err
	}
	published := 0
	for _, e := range pending {
		var payload map[string]any
		_ = json.Unmarshal(e.Payload, &payload)
		env := events.Envelope{EventID: e.ID.String(), Topic: e.Topic, Payload: payload}
		data, err := json.Marshal(env)
		if err != nil {
			slog.Error("relay: marshal", "err", err, "id", e.ID)
			continue
		}
		if err := r.bus.Publish(e.Topic, data); err != nil {
			slog.Error("relay: publish", "err", err, "id", e.ID)
			continue
		}
		if err := r.s.MarkPublished(ctx, e.ID); err != nil {
			slog.Error("relay: mark", "err", err, "id", e.ID)
			continue
		}
		published++
	}
	return published, nil
}

func (r *Relay) Run(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			_, _ = r.RunOnce(ctx)
		}
	}
}
