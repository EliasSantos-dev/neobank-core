// Package user é o domínio de identidade (puro).
package user

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

type User struct {
	ID              uuid.UUID
	Email           string
	WalletAccountID uuid.UUID
	CreatedAt       time.Time
}

// Validate checa formato de email e força mínima de senha.
func Validate(email, password string) error {
	if !emailRe.MatchString(email) {
		return ErrInvalidEmail
	}
	if len(password) < 8 {
		return ErrWeakPassword
	}
	return nil
}
