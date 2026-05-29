package gateway

import "github.com/google/uuid"

type WebhookEvent struct {
	EventID   uuid.UUID
	PaymentID uuid.UUID
	Status    string // "succeeded" | "failed"
}
