package risk

import (
	"context"
	"testing"
)

func TestEngineAssess_Low(t *testing.T) {
	e := NewEngine(RuleReasoningAdvisor{}, 50)
	a, err := e.Assess(context.Background(), Context{Amount: 100, SenderBalance: 100000, RecipientIsNew: false})
	if err != nil {
		t.Fatal(err)
	}
	if a.Level != Low {
		t.Fatalf("esperava low, veio %s (score %d)", a.Level, a.Score)
	}
}

func TestEngineAssess_High(t *testing.T) {
	e := NewEngine(RuleReasoningAdvisor{}, 50)
	// drain (35) + new_recipient (20) = 55 >= 50 -> high
	a, err := e.Assess(context.Background(), Context{Amount: 9500, SenderBalance: 10000, RecipientIsNew: true})
	if err != nil {
		t.Fatal(err)
	}
	if a.Level != High {
		t.Fatalf("esperava high, veio %s (score %d)", a.Level, a.Score)
	}
	if a.Advice == "" {
		t.Fatal("advice não deveria ser vazio")
	}
	if len(a.Signals) < 2 {
		t.Fatalf("esperava >=2 sinais, veio %d", len(a.Signals))
	}
}
