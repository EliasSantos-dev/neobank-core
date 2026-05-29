package store

import (
	"context"

	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/EliasSantos-dev/neobank-core/internal/recon"
)

// ReconInput coleta os dados agregados para a reconciliação dos invariantes.
func (s *Store) ReconInput(ctx context.Context) (recon.Input, error) {
	var in recon.Input

	if err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(CASE WHEN direction='credit' THEN amount ELSE -amount END),0)::bigint FROM entries`).
		Scan(&in.GlobalSum); err != nil {
		return in, err
	}

	rows, err := s.pool.Query(ctx,
		`SELECT transfer_id,
		        COALESCE(SUM(CASE WHEN direction='debit'  THEN amount ELSE 0 END),0)::bigint,
		        COALESCE(SUM(CASE WHEN direction='credit' THEN amount ELSE 0 END),0)::bigint
		 FROM entries GROUP BY transfer_id`)
	if err != nil {
		return in, err
	}
	for rows.Next() {
		var t recon.TransferSums
		if err := rows.Scan(&t.TransferID, &t.Debits, &t.Credits); err != nil {
			rows.Close()
			return in, err
		}
		in.Transfers = append(in.Transfers, t)
	}
	rows.Close()

	wrows, err := s.pool.Query(ctx,
		`SELECT a.id,
		        COALESCE(SUM(CASE WHEN e.direction='credit' THEN e.amount ELSE -e.amount END),0)::bigint
		 FROM accounts a
		 LEFT JOIN entries e ON e.account_id = a.id
		 WHERE a.type='wallet'
		 GROUP BY a.id`)
	if err != nil {
		return in, err
	}
	for wrows.Next() {
		var wb recon.WalletBalance
		if err := wrows.Scan(&wb.AccountID, &wb.Balance); err != nil {
			wrows.Close()
			return in, err
		}
		in.Wallets = append(in.Wallets, wb)
	}
	wrows.Close()
	if err := wrows.Err(); err != nil {
		return in, err
	}

	// saldo da conta gateway no ledger
	if err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(CASE WHEN direction='credit' THEN amount ELSE -amount END),0)::bigint
		 FROM entries WHERE account_id = $1`, ledger.GatewayBRL).Scan(&in.GatewayLedger); err != nil {
		return in, err
	}
	// esperado: Σ(holds de saques ativos: pending/completed) − Σ(depósitos completed)
	if err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(CASE
		           WHEN kind='withdrawal' AND status IN ('pending','completed') THEN amount
		           WHEN kind='deposit'    AND status='completed'                THEN -amount
		           ELSE 0 END),0)::bigint
		 FROM payments`).Scan(&in.GatewayExpected); err != nil {
		return in, err
	}
	return in, nil
}
