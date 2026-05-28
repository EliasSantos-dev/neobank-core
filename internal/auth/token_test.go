package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIssueAndParse(t *testing.T) {
	secret := []byte("test-secret")
	id := uuid.New()
	tok, err := Issue(secret, id, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Parse(secret, tok)
	if err != nil {
		t.Fatal(err)
	}
	if got != id {
		t.Fatalf("sub %s != %s", got, id)
	}
}

func TestParseRejectsExpired(t *testing.T) {
	secret := []byte("s")
	tok, _ := Issue(secret, uuid.New(), -time.Minute)
	if _, err := Parse(secret, tok); err == nil {
		t.Fatal("expirado deveria falhar")
	}
}

func TestParseRejectsTampered(t *testing.T) {
	tok, _ := Issue([]byte("s"), uuid.New(), time.Hour)
	if _, err := Parse([]byte("outro"), tok); err == nil {
		t.Fatal("assinatura inválida deveria falhar")
	}
}
