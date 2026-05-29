package ledger

import "github.com/google/uuid"

// Contas de câmbio (external), uma por moeda — semeadas na migration 0006.
var (
	FxBRL = uuid.MustParse("00000000-0000-0000-0000-000000000010")
	FxUSD = uuid.MustParse("00000000-0000-0000-0000-000000000011")
	FxEUR = uuid.MustParse("00000000-0000-0000-0000-000000000012")
)

// FxAccount devolve a conta de câmbio da moeda (e se a moeda é suportada).
func FxAccount(currency string) (uuid.UUID, bool) {
	switch currency {
	case "BRL":
		return FxBRL, true
	case "USD":
		return FxUSD, true
	case "EUR":
		return FxEUR, true
	default:
		return uuid.Nil, false
	}
}
