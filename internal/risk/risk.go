// Package risk avalia o risco de uma transação (domínio puro, sem IO).
package risk

type Level string

const (
	Low  Level = "low"
	High Level = "high"
)

// Context reúne os dados necessários para pontuar uma transação.
// O chamador (worker) preenche a partir do store.
type Context struct {
	Amount            int64
	Currency          string
	SenderBalance     int64
	RecipientIsNew    bool
	RecentTransfers   int
	HistoricalAmounts []int64
}

// Signal é a contribuição de risco de um sinal individual (Score 0..100).
type Signal struct {
	Name   string
	Score  int
	Reason string
}

// Assessment é o resultado agregado da avaliação.
type Assessment struct {
	Score   int
	Level   Level
	Signals []Signal
	Advice  string
}
