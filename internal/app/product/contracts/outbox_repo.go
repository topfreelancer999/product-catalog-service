package contracts

import (
	"context"

	"cloud.google.com/go/spanner"
)

// EnrichedEvent represents a domain event enriched with metadata for outbox storage.
type EnrichedEvent struct {
	EventID   string
	EventType string
	AggregateID string
	Payload []byte
	// Status is typically "pending" for new events.
	Status string
}

// OutboxRepo defines the interface for storing events in the transactional outbox.
type OutboxRepo interface {
	// InsertMut returns a mutation to insert an enriched event.
	// Returns nil if event is nil.
	InsertMut(event *EnrichedEvent) *spanner.Mutation
}
