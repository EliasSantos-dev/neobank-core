// Package events é o barramento de eventos (NATS embutido in-process).
package events

const (
	TopicTransferCompleted = "transfer.completed"
	TopicIntentUnderReview = "intent.under_review"
	TopicIntentRejected    = "intent.rejected"
)

// Envelope é o que trafega no barramento: id para idempotência + payload.
type Envelope struct {
	EventID string         `json:"event_id"`
	Topic   string         `json:"topic"`
	Payload map[string]any `json:"payload"`
}
