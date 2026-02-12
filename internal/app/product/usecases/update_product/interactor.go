package updateproduct

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Vektor-AI/commitplan"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/pkg/clock"
	"product-catalog-service/internal/pkg/committer"
)

// Request represents input for updating product details.
type Request struct {
	ProductID   string
	Name        *string // nil means no change
	Description *string
	Category    *string
}

// Interactor implements the UpdateProduct usecase following the Golden Mutation Pattern.
type Interactor struct {
	repo      contracts.ProductRepo
	outboxRepo contracts.OutboxRepo
	committer *committer.PlanCommitter
	clock     clock.Clock
}

// New creates a new UpdateProduct interactor.
func New(
	repo contracts.ProductRepo,
	outboxRepo contracts.OutboxRepo,
	committer *committer.PlanCommitter,
	clock clock.Clock,
) *Interactor {
	return &Interactor{
		repo:      repo,
		outboxRepo: outboxRepo,
		committer: committer,
		clock:     clock,
	}
}

// Execute updates product details atomically with events.
func (it *Interactor) Execute(ctx context.Context, req Request) error {
	// 1. Load aggregate
	product, err := it.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	// 2. Call domain method
	now := it.clock.Now()
	name := ""
	desc := ""
	cat := ""

	if req.Name != nil {
		name = *req.Name
	}
	if req.Description != nil {
		desc = *req.Description
	}
	if req.Category != nil {
		cat = *req.Category
	}

	product.UpdateDetails(name, desc, cat, now)

	// 3. Build commit plan
	plan := commitplan.NewPlan()

	// 4. Get mutations from repository
	if mut := it.repo.UpdateMut(product); mut != nil {
		plan.Add(mut)
	}

	// 5. Add outbox events
	for _, event := range product.DomainEvents() {
		enriched := enrichEvent(product.ID(), event)
		if outboxMut := it.outboxRepo.InsertMut(enriched); outboxMut != nil {
			plan.Add(outboxMut)
		}
	}

	// 6. Apply plan
	if err := it.committer.Apply(ctx, plan); err != nil {
		return err
	}

	product.ClearDomainEvents()
	return nil
}

// enrichEvent converts a domain event to an enriched outbox event.
func enrichEvent(aggregateID string, event domain.DomainEvent) *contracts.EnrichedEvent {
	payload, _ := json.Marshal(event)
	return &contracts.EnrichedEvent{
		EventID:     generateID(),
		EventType:   eventType(event),
		AggregateID: aggregateID,
		Payload:     payload,
		Status:      "pending",
	}
}

func eventType(event domain.DomainEvent) string {
	switch event.(type) {
	case domain.ProductUpdatedEvent:
		return "product.updated"
	default:
		return "unknown"
	}
}

func generateID() string {
	return fmt.Sprintf("id-%d", time.Now().UnixNano())
}
