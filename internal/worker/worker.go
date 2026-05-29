// Package worker processa intenções de transferência pendentes em background.
package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/EliasSantos-dev/neobank-core/internal/risk"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
)

type Worker struct {
	s      *store.Store
	engine *risk.Engine
}

func New(s *store.Store, engine *risk.Engine) *Worker {
	return &Worker{s: s, engine: engine}
}

// ProcessOnce pontua todas as intenções pendentes e as transiciona.
// Retorna quantas processou.
func (w *Worker) ProcessOnce(ctx context.Context) (int, error) {
	pending, err := w.s.ListPending(ctx, 50)
	if err != nil {
		return 0, err
	}
	for _, it := range pending {
		if err := w.processIntent(ctx, it); err != nil {
			return 0, err
		}
	}
	return len(pending), nil
}

func (w *Worker) processIntent(ctx context.Context, it store.Intent) error {
	balance, err := w.s.Balance(ctx, it.FromAccountID)
	if err != nil {
		return err
	}
	isNew, err := w.s.RecipientIsNew(ctx, it.FromAccountID, it.ToAccountID)
	if err != nil {
		return err
	}
	hist, err := w.s.DebitHistory(ctx, it.FromAccountID, 100)
	if err != nil {
		return err
	}
	recent, err := w.s.RecentDebitCount(ctx, it.FromAccountID, 20)
	if err != nil {
		return err
	}

	a, err := w.engine.Assess(ctx, risk.Context{
		Amount:            it.Amount,
		Currency:          it.Currency,
		SenderBalance:     balance,
		RecipientIsNew:    isNew,
		RecentTransfers:   recent,
		HistoricalAmounts: hist,
	})
	if err != nil {
		return err
	}

	reasons, _ := json.Marshal(a)
	if a.Level == risk.Low {
		// grava o score (observabilidade) e efetiva
		if err := w.s.SetRisk(ctx, it.ID, a.Score, string(a.Level), reasons); err != nil {
			return err
		}
		return w.s.CompleteIntent(ctx, it.ID)
	}
	return w.s.MarkUnderReview(ctx, it.ID, a.Score, string(a.Level), reasons)
}

// Run executa ProcessOnce num ticker até o contexto ser cancelado.
func (w *Worker) Run(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			_, _ = w.ProcessOnce(ctx)
		}
	}
}
