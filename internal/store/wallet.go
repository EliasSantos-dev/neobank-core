package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type WalletBalance struct {
	Currency string
	Balance  int64
}

// GetOrCreateWallet devolve a conta-wallet do usuário para a moeda, criando se preciso.
// Para BRL, reconcilia com users.wallet_account_id (criado no M2).
func (s *Store) GetOrCreateWallet(ctx context.Context, userID uuid.UUID, currency string) (uuid.UUID, error) {
	var accountID uuid.UUID
	err := s.pool.QueryRow(ctx,
		`SELECT account_id FROM user_wallets WHERE user_id=$1 AND currency=$2`, userID, currency).Scan(&accountID)
	if err == nil {
		return accountID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	if currency == "BRL" {
		if err := tx.QueryRow(ctx, `SELECT wallet_account_id FROM users WHERE id=$1`, userID).Scan(&accountID); err != nil {
			return uuid.Nil, err
		}
	} else {
		if err := tx.QueryRow(ctx,
			`INSERT INTO accounts (currency, type) VALUES ($1,'wallet') RETURNING id`, currency).Scan(&accountID); err != nil {
			return uuid.Nil, err
		}
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO user_wallets (user_id, currency, account_id) VALUES ($1,$2,$3)
		 ON CONFLICT (user_id, currency) DO NOTHING`,
		userID, currency, accountID); err != nil {
		return uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	// relê (cobre corrida onde outro inseriu antes)
	if err := s.pool.QueryRow(ctx,
		`SELECT account_id FROM user_wallets WHERE user_id=$1 AND currency=$2`, userID, currency).Scan(&accountID); err != nil {
		return uuid.Nil, err
	}
	return accountID, nil
}

// ListWallets devolve as wallets do usuário com saldo. Sempre inclui BRL.
func (s *Store) ListWallets(ctx context.Context, userID uuid.UUID) ([]WalletBalance, error) {
	if _, err := s.GetOrCreateWallet(ctx, userID, "BRL"); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx,
		`SELECT uw.currency,
		        COALESCE(SUM(CASE WHEN e.direction='credit' THEN e.amount ELSE -e.amount END),0)::bigint
		 FROM user_wallets uw
		 LEFT JOIN entries e ON e.account_id = uw.account_id
		 WHERE uw.user_id=$1
		 GROUP BY uw.currency
		 ORDER BY uw.currency`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []WalletBalance
	for rows.Next() {
		var wb WalletBalance
		if err := rows.Scan(&wb.Currency, &wb.Balance); err != nil {
			return nil, err
		}
		out = append(out, wb)
	}
	return out, rows.Err()
}
