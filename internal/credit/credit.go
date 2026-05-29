// Package credit calcula um score de crédito explicável a partir do histórico
// transacional (domínio puro, sem IO).
package credit

type Band string

const (
	Poor      Band = "poor"
	Fair      Band = "fair"
	Good      Band = "good"
	Excellent Band = "excellent"
)

type CreditInput struct {
	AccountAgeDays  int
	DepositCount    int
	Balance         int64
	DepositTotal    int64
	WithdrawalTotal int64
	RiskFlags       int
}

type Factor struct {
	Name   string
	Points int
}

type Result struct {
	Score   int
	Band    Band
	Factors []Factor
}

const base = 50

func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Score aplica os fatores sobre a base e classifica em faixas.
func Score(in CreditInput) Result {
	factors := []Factor{
		{"account_age", min(15, in.AccountAgeDays/30)},
		{"deposit_activity", min(15, in.DepositCount*3)},
		{"balance", min(10, int(in.Balance/10_000))},
		{"withdrawal_ratio", -withdrawalPenalty(in)},
		{"risk_flags", -min(25, in.RiskFlags*10)},
	}
	total := base
	for _, f := range factors {
		total += f.Points
	}
	score := clamp(total)
	return Result{Score: score, Band: band(score), Factors: factors}
}

func withdrawalPenalty(in CreditInput) int {
	if in.DepositTotal <= 0 {
		return 0
	}
	ratioX15 := int(in.WithdrawalTotal * 15 / in.DepositTotal)
	if ratioX15 <= 7 { // ~0.5*15
		return 0
	}
	return min(15, ratioX15)
}

func band(score int) Band {
	switch {
	case score >= 80:
		return Excellent
	case score >= 60:
		return Good
	case score >= 40:
		return Fair
	default:
		return Poor
	}
}
