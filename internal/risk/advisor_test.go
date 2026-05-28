package risk

import (
	"context"
	"strings"
	"testing"
)

func TestRuleReasoningAdvisor(t *testing.T) {
	a := RuleReasoningAdvisor{}
	signals := []Signal{
		{Name: "drain", Score: 35, Reason: "transferência drena fração alta do saldo"},
		{Name: "new_recipient", Score: 20, Reason: "primeira transferência para este destinatário"},
	}
	advice, err := a.Advise(context.Background(), Context{}, signals, 55)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(advice, "55") {
		t.Fatalf("advice deveria citar o score: %q", advice)
	}
	if !strings.Contains(advice, "drena") {
		t.Fatalf("advice deveria citar o motivo: %q", advice)
	}
}

func TestRuleReasoningAdvisor_NoSignals(t *testing.T) {
	advice, err := RuleReasoningAdvisor{}.Advise(context.Background(), Context{}, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if advice == "" {
		t.Fatal("advice não deveria ser vazio")
	}
}
