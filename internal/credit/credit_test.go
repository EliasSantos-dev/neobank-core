package credit

import "testing"

func TestScore_HighProfile(t *testing.T) {
	r := Score(CreditInput{
		AccountAgeDays: 365, DepositCount: 10, Balance: 200_000,
		DepositTotal: 200_000, WithdrawalTotal: 10_000, RiskFlags: 0,
	})
	if r.Score < 80 {
		t.Fatalf("esperava score alto (>=80), veio %d", r.Score)
	}
	if r.Band != Excellent {
		t.Fatalf("esperava excellent, veio %s", r.Band)
	}
	sum := 50
	for _, f := range r.Factors {
		sum += f.Points
	}
	if clamp(sum) != r.Score {
		t.Fatalf("breakdown (%d) != score (%d)", clamp(sum), r.Score)
	}
}

func TestScore_LowProfile(t *testing.T) {
	r := Score(CreditInput{
		AccountAgeDays: 2, DepositCount: 1, Balance: 0,
		DepositTotal: 1_000, WithdrawalTotal: 5_000, RiskFlags: 5,
	})
	if r.Score >= 40 {
		t.Fatalf("esperava score baixo (<40), veio %d", r.Score)
	}
	if r.Band != Poor {
		t.Fatalf("esperava poor, veio %s", r.Band)
	}
}

func TestScore_Clamp(t *testing.T) {
	r := Score(CreditInput{RiskFlags: 100})
	if r.Score < 0 || r.Score > 100 {
		t.Fatalf("score fora de 0..100: %d", r.Score)
	}
}
