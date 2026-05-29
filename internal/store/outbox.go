package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OutboxEvent struct {
	ID      uuid.UUID
	Topic   string
	Payload []byte
}

// writeEventTx grava um evento no outbox usando a transação fornecida
// (mesma tx da mudança de estado → atomicidade, sem dual-write).
func writeEventTx(ctx context.Context, tx pgx.Tx, topic string, payload []byte) (uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx,
		`INSERT INTO outbox (topic, payload) VALUES ($1, $2) RETURNING id`,
		topic, payload).Scan(&id)
	return id, err
}

// WriteEventDirect grava um evento fora de uma tx externa (uso em testes/admin).
func (s *Store) WriteEventDirect(ctx context.Context, topic string, payload []byte) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx,
		`INSERT INTO outbox (topic, payload) VALUES ($1, $2) RETURNING id`,
		topic, payload).Scan(&id)
	return id, err
}

func (s *Store) FetchUnpublished(ctx context.Context, limit int) ([]OutboxEvent, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, topic, payload FROM outbox WHERE published_at IS NULL ORDER BY created_at LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []OutboxEvent
	for rows.Next() {
		var e OutboxEvent
		if err := rows.Scan(&e.ID, &e.Topic, &e.Payload); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) MarkPublished(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `UPDATE outbox SET published_at = now() WHERE id = $1`, id)
	return err
}

func (s *Store) InsertEventLog(ctx context.Context, eventID uuid.UUID, topic string, payload []byte) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO event_log (event_id, topic, payload) VALUES ($1,$2,$3) ON CONFLICT (event_id) DO NOTHING`,
		eventID, topic, payload)
	return err
}

func (s *Store) CountEventLog(ctx context.Context) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT count(*) FROM event_log`).Scan(&n)
	return n, err
}
