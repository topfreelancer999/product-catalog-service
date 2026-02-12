package domain

import "time"

// DomainEvent is a marker interface for all product domain events.
// Events are simple intent-carrying structs without behavior.
type DomainEvent interface {
	OccurredAt() time.Time
}

// baseEvent provides common fields for all events.
type baseEvent struct {
	occurredAt time.Time
}

func (e baseEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// ProductCreatedEvent is raised when a product is created.
type ProductCreatedEvent struct {
	baseEvent
	ProductID string
}

// ProductUpdatedEvent is raised when mutable product details change.
type ProductUpdatedEvent struct {
	baseEvent
	ProductID string
}

// ProductActivatedEvent is raised when a product becomes active.
type ProductActivatedEvent struct {
	baseEvent
	ProductID string
}

// ProductDeactivatedEvent is raised when a product becomes inactive.
type ProductDeactivatedEvent struct {
	baseEvent
	ProductID string
}

// DiscountAppliedEvent is raised when a discount is added or changed.
type DiscountAppliedEvent struct {
	baseEvent
	ProductID string
}

// DiscountRemovedEvent is raised when the product discount is removed.
type DiscountRemovedEvent struct {
	baseEvent
	ProductID string
}

