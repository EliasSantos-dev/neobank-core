// Package store é o adapter PostgreSQL do ledger (pgx escrito à mão).
package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxuuid "github.com/vgarvardt/pgx-google-uuid/v5"
)

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// NewPool cria um pool de conexões com o codec de UUID (google/uuid) registrado,
// para que uuid.UUID seja encodado/decodado de forma transparente.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxuuid.Register(conn.TypeMap())
		return nil
	}
	return pgxpool.NewWithConfig(ctx, cfg)
}

// querier abstrai *pgxpool.Pool e pgx.Tx (ambos implementam estes métodos),
// permitindo reusar queries dentro ou fora de uma transação.
type querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}
