package store

import (
	"context"
	"errors"

	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/EliasSantos-dev/neobank-core/internal/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// CreateUser cria a carteira (accounts) e o usuário na mesma transação.
func (s *Store) CreateUser(ctx context.Context, email, passwordHash string) (user.User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return user.User{}, err
	}
	defer tx.Rollback(ctx)

	var walletID uuid.UUID
	if err := tx.QueryRow(ctx,
		`INSERT INTO accounts (currency, type) VALUES ('BRL', 'wallet') RETURNING id`).Scan(&walletID); err != nil {
		return user.User{}, err
	}

	var u user.User
	err = tx.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, wallet_account_id) VALUES ($1, $2, $3)
		 RETURNING id, email, wallet_account_id, created_at`,
		email, passwordHash, walletID).Scan(&u.ID, &u.Email, &u.WalletAccountID, &u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return user.User{}, user.ErrEmailTaken
		}
		return user.User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return user.User{}, err
	}
	return u, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (user.User, string, error) {
	var u user.User
	var hash string
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, wallet_account_id, created_at, password_hash FROM users WHERE email = $1`,
		email).Scan(&u.ID, &u.Email, &u.WalletAccountID, &u.CreatedAt, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return user.User{}, "", user.ErrUserNotFound
	}
	return u, hash, err
}

func (s *Store) GetUserByID(ctx context.Context, id uuid.UUID) (user.User, error) {
	var u user.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, wallet_account_id, created_at FROM users WHERE id = $1`,
		id).Scan(&u.ID, &u.Email, &u.WalletAccountID, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return user.User{}, user.ErrUserNotFound
	}
	return u, err
}

// ListEntries devolve o extrato de uma conta, mais recentes primeiro.
func (s *Store) ListEntries(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]ledger.Entry, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, seq, transfer_id, account_id, direction, amount, currency, created_at
		 FROM entries WHERE account_id = $1 ORDER BY seq DESC LIMIT $2 OFFSET $3`,
		accountID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ledger.Entry
	for rows.Next() {
		var e ledger.Entry
		var dir string
		if err := rows.Scan(&e.ID, &e.Seq, &e.TransferID, &e.AccountID, &dir, &e.Amount, &e.Currency, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Direction = ledger.Direction(dir)
		out = append(out, e)
	}
	return out, rows.Err()
}
