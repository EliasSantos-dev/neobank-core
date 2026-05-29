package store

import (
	"context"

	"github.com/EliasSantos-dev/neobank-core/internal/credit"
	"github.com/google/uuid"
)

// CreditInput agrega os sinais de crédito do usuário a partir do histórico.
func (s *Store) CreditInput(ctx context.Context, userID uuid.UUID) (credit.CreditInput, error) {
	var in credit.CreditInput

	u, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return in, err
	}

	if err := s.pool.QueryRow(ctx,
		`SELECT GREATEST(0, EXTRACT(DAY FROM (now() - created_at)))::int FROM users WHERE id=$1`,
		userID).Scan(&in.AccountAgeDays); err != nil {
		return in, err
	}

	if err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*), COALESCE(SUM(amount),0)::bigint
		 FROM payments WHERE account_id=$1 AND kind='deposit' AND status='completed'`,
		u.WalletAccountID).Scan(&in.DepositCount, &in.DepositTotal); err != nil {
		return in, err
	}

	if err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount),0)::bigint
		 FROM payments WHERE account_id=$1 AND kind='withdrawal' AND status='completed'`,
		u.WalletAccountID).Scan(&in.WithdrawalTotal); err != nil {
		return in, err
	}

	bal, err := s.Balance(ctx, u.WalletAccountID)
	if err != nil {
		return in, err
	}
	in.Balance = bal

	if err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM transfer_intents
		 WHERE from_account_id=$1 AND status IN ('under_review','rejected')`,
		u.WalletAccountID).Scan(&in.RiskFlags); err != nil {
		return in, err
	}

	return in, nil
}
