package risk

import "context"

type Engine struct {
	advisor   Advisor
	threshold int
}

func NewEngine(advisor Advisor, threshold int) *Engine {
	return &Engine{advisor: advisor, threshold: threshold}
}

// Assess roda regras + estatística, agrega o score (clamp 0..100),
// classifica o nível e pede a explicação ao advisor.
func (e *Engine) Assess(ctx context.Context, c Context) (Assessment, error) {
	var signals []Signal
	for _, r := range allRules {
		if sig, ok := r(c); ok {
			signals = append(signals, sig)
		}
	}
	if sig, ok := statisticalSignal(c); ok {
		signals = append(signals, sig)
	}

	score := 0
	for _, s := range signals {
		score += s.Score
	}
	if score > 100 {
		score = 100
	}

	level := Low
	if score >= e.threshold {
		level = High
	}

	advice, err := e.advisor.Advise(ctx, c, signals, score)
	if err != nil {
		return Assessment{}, err
	}
	return Assessment{Score: score, Level: level, Signals: signals, Advice: advice}, nil
}
