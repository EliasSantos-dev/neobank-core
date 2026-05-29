// Package audit consome eventos e materializa um read-model (event_log).
package audit

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/EliasSantos-dev/neobank-core/internal/events"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	"github.com/google/uuid"
)

// Start assina os eventos de domínio e persiste cada um no event_log.
// Idempotente: event_id é PK com ON CONFLICT DO NOTHING.
func Start(ctx context.Context, bus *events.Bus, s *store.Store) error {
	handler := func(topic string, payload []byte) {
		var env events.Envelope
		if err := json.Unmarshal(payload, &env); err != nil {
			slog.Error("audit: unmarshal", "err", err)
			return
		}
		id, err := uuid.Parse(env.EventID)
		if err != nil {
			slog.Error("audit: event_id inválido", "id", env.EventID)
			return
		}
		body, _ := json.Marshal(env.Payload)
		if err := s.InsertEventLog(ctx, id, env.Topic, body); err != nil {
			slog.Error("audit: insert", "err", err)
		}
	}
	if err := bus.Subscribe("transfer.>", handler); err != nil {
		return err
	}
	return bus.Subscribe("intent.>", handler)
}
