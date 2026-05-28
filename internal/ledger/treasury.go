package ledger

import "github.com/google/uuid"

// TreasuryBRL é a conta external (funding) semeada na migration 0002.
// Depósitos creditam a carteira a partir dela; saques a creditam de volta.
var TreasuryBRL = uuid.MustParse("00000000-0000-0000-0000-000000000001")
