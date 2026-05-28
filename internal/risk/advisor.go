package risk

import (
	"context"
	"fmt"
	"strings"
)

// Advisor produz uma explicação em linguagem natural sobre a avaliação.
// A implementação default é determinística; um LLMAdvisor pode plugar aqui.
type Advisor interface {
	Advise(ctx context.Context, c Context, signals []Signal, score int) (string, error)
}

// RuleReasoningAdvisor sintetiza o racional a partir dos sinais disparados.
// Determinístico, sem rede — é a camada de "agente" entregue no M3.
type RuleReasoningAdvisor struct{}

func (RuleReasoningAdvisor) Advise(_ context.Context, _ Context, signals []Signal, score int) (string, error) {
	if len(signals) == 0 {
		return fmt.Sprintf("Risco %d/100: nenhum sinal relevante; transação dentro do padrão.", score), nil
	}
	reasons := make([]string, 0, len(signals))
	for _, s := range signals {
		reasons = append(reasons, s.Reason)
	}
	return fmt.Sprintf("Risco %d/100. Sinais: %s.", score, strings.Join(reasons, "; ")), nil
}
