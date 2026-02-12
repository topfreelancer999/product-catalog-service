package removediscount

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

// Request represents input for removing a discount from a product.
type Request struct {
	ProductID string
}

// Interactor implements the RemoveDiscount usecase following the Golden Mutation Pattern.
type Interactor struct {
	repo      contracts.ProductRepo
	outboxRepo contracts.OutboxRepo
	committer *committer.PlanCommitter
	clock     clock.Clock
}

// New creates a new RemoveDiscount interactor.
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

// Execute removes the current discount from a product (if any).
// Uses precise decimal arithmetic for pricing calculations via domain service.
func (it *Interactor) Execute(ctx context.Context, req Request) error {
	// 1. Load aggregate
	product, err := it.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	// 2. Call domain method (removes discount if present)
	now := it.clock.Now()
	product.RemoveDiscount(now)

	// 3. Build commit plan
	plan := commitplan.NewPlan()

	// 4. Get mutations from repository (only if discount was actually removed)
	if mut := it.repo.UpdateMut(product); mut != nil {
		plan.Add(mut)
	}

	// 5. Add outbox events (only if discount was removed)
	for _, event := range product.DomainEvents() {
		enriched := enrichEvent(product.ID(), event)
		if outboxMut := it.outboxRepo.InsertMut(enriched); outboxMut != nil {
			plan.Add(outboxMut)
		}
	}

	// 6. Apply plan atomically
	if err := it.committer.Apply(ctx, plan); err != nil {
		return err
	}

	product.ClearDomainEvents()
	return nil
}

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
	case domain.DiscountRemovedEvent:
		return "discount.removed"
	default:
		return "unknown"
	}
}

func generateID() string {
	return fmt.Sprintf("id-%d", time.Now().UnixNano())
}
