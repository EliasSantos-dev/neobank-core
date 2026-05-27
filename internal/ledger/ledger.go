// Package ledger contém o domínio puro do ledger (sem IO).
package ledger

import (
	"time"

	"github.com/google/uuid"
)

type Direction string

const (
	Debit  Direction = "debit"
	Credit Direction = "credit"
)

func (d Direction) Valid() bool { return d == Debit || d == Credit }

type AccountType string

const (
	Wallet   AccountType = "wallet"   // carteira de usuário: não pode ficar negativa
	External AccountType = "external" // fonte/sumidouro de funding: pode ficar negativa
)

// AllowsNegative indica se o tipo de conta pode ter saldo < 0.
func (t AccountType) AllowsNegative() bool { return t == External }

type Account struct {
	ID        uuid.UUID
	Currency  string
	Type      AccountType
	CreatedAt time.Time
}

type Entry struct {
	ID         uuid.UUID
	Seq        int64
	TransferID uuid.UUID
	AccountID  uuid.UUID
	Direction  Direction
	Amount     int64
	Currency   string
	CreatedAt  time.Time
}

type Transfer struct {
	ID             uuid.UUID
	IdempotencyKey string
	Status         string
	CreatedAt      time.Time
}
