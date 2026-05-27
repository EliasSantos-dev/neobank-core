package store

import (
	"context"

	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/google/uuid"
)

func (s *Store) CreateAccount(ctx context.Context, currency string, t ledger.AccountType) (ledger.Account, error) {
	var a ledger.Account
	var typ string
	err := s.pool.QueryRow(ctx,
		`INSERT INTO accounts (currency, type) VALUES ($1, $2)
		 RETURNING id, currency, type, created_at`,
		currency, string(t)).Scan(&a.ID, &a.Currency, &typ, &a.CreatedAt)
	a.Type = ledger.AccountType(typ)
	return a, err
}

// Balance deriva o saldo via SUM dos lançamentos (history-as-truth).
func (s *Store) Balance(ctx context.Context, accountID uuid.UUID) (int64, error) {
	return balance(ctx, s.pool, accountID)
}

func balance(ctx context.Context, q querier, id uuid.UUID) (int64, error) {
	var b int64
	err := q.QueryRow(ctx,
		`SELECT COALESCE(SUM(CASE WHEN direction = 'credit' THEN amount ELSE -amount END), 0)::bigint
		 FROM entries WHERE account_id = $1`, id).Scan(&b)
	return b, err
}

func getTransferByKey(ctx context.Context, q querier, key string) (ledger.Transfer, error) {
	var t ledger.Transfer
	err := q.QueryRow(ctx,
		`SELECT id, idempotency_key, status, created_at FROM transfers WHERE idempotency_key = $1`,
		key).Scan(&t.ID, &t.IdempotencyKey, &t.Status, &t.CreatedAt)
	return t, err
}
