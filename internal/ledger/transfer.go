package ledger

import "github.com/google/uuid"

// BuildEntries gera os dois lançamentos balanceados (débito na origem,
// crédito no destino) de uma transferência. Função pura: não toca no banco
// e não decide saldo suficiente (isso é responsabilidade do store, sob lock).
func BuildEntries(transferID uuid.UUID, from, to Account, amount int64) ([]Entry, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if from.ID == to.ID {
		return nil, ErrSameAccount
	}
	if from.Currency != to.Currency {
		return nil, ErrCurrencyMismatch
	}
	return []Entry{
		{TransferID: transferID, AccountID: from.ID, Direction: Debit, Amount: amount, Currency: from.Currency},
		{TransferID: transferID, AccountID: to.ID, Direction: Credit, Amount: amount, Currency: to.Currency},
	}, nil
}
