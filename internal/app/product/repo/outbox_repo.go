package repo

import (
	"time"

	"cloud.google.com/go/spanner"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/models/moutbox"
)

// OutboxRepo implements the transactional outbox pattern for event storage.
type OutboxRepo struct{}

// NewOutboxRepo creates a new OutboxRepo instance.
func NewOutboxRepo() *OutboxRepo {
	return &OutboxRepo{}
}

// InsertMut converts an EnrichedEvent to an OutboxEvent and returns a Spanner mutation.
// This implements the contracts.OutboxRepo interface.
// Returns nil if event is nil.
func (r *OutboxRepo) InsertMut(event *contracts.EnrichedEvent) *spanner.Mutation {
	if event == nil {
		return nil
	}

	// Map EnrichedEvent to OutboxEvent model
	outboxEvent := &moutbox.OutboxEvent{
		EventID:     event.EventID,
		EventType:   event.EventType,
		AggregateID: event.AggregateID,
		Payload:     event.Payload,
		Status:      event.Status,
		CreatedAt:   time.Now(), // Use current time for created_at
	}

	// Use the model's InsertMut helper to create the Spanner mutation
	return moutbox.InsertMut(outboxEvent)
}
