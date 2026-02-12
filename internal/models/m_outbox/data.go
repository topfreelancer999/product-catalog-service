package moutbox

import (
	"cloud.google.com/go/spanner"
	"time"
)

// OutboxEvent represents a row in the outbox_events table.
type OutboxEvent struct {
	EventID    string
	EventType  string
	AggregateID string
	Payload   []byte
	Status    string
	CreatedAt time.Time
	ProcessedAt *time.Time
}

// InsertMut returns a mutation to insert a new outbox event.
func InsertMut(e *OutboxEvent) *spanner.Mutation {
	if e == nil {
		return nil
	}
	return spanner.Insert(TableName, []string{
		"event_id",
		"event_type",
		"aggregate_id",
		"payload",
		"status",
		"created_at",
	}, []interface{}{
		e.EventID,
		e.EventType,
		e.AggregateID,
		e.Payload,
		e.Status,
		e.CreatedAt,
	})
}

