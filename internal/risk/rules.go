package risk

const (
	velocityThreshold    = 5
	largeAmountThreshold = 1_000_000 // R$ 10.000,00 em centavos
	drainFractionNum     = 9         // 90% = 9/10
	drainFractionDen     = 10
)

type rule func(Context) (Signal, bool)

func ruleVelocity(c Context) (Signal, bool) {
	if c.RecentTransfers > velocityThreshold {
		return Signal{Name: "velocity", Score: 25, Reason: "muitas transferências em curto período"}, true
	}
	return Signal{}, false
}

func ruleLargeAmount(c Context) (Signal, bool) {
	if c.Amount > largeAmountThreshold {
		return Signal{Name: "large_amount", Score: 30, Reason: "valor acima do limiar de alto valor"}, true
	}
	return Signal{}, false
}

func ruleNewRecipient(c Context) (Signal, bool) {
	if c.RecipientIsNew {
		return Signal{Name: "new_recipient", Score: 20, Reason: "primeira transferência para este destinatário"}, true
	}
	return Signal{}, false
}

func ruleDrain(c Context) (Signal, bool) {
	if c.SenderBalance > 0 && c.Amount*drainFractionDen >= c.SenderBalance*drainFractionNum {
		return Signal{Name: "drain", Score: 35, Reason: "transferência drena fração alta do saldo"}, true
	}
	return Signal{}, false
}

var allRules = []rule{ruleVelocity, ruleLargeAmount, ruleNewRecipient, ruleDrain}
