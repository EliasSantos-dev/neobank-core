package ledger

import "github.com/google/uuid"

// GatewayBRL é a conta external que representa o gateway de pagamento
// (contraparte de depósitos e saques), semeada na migration 0005.
var GatewayBRL = uuid.MustParse("00000000-0000-0000-0000-000000000002")
