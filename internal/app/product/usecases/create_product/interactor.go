package createproduct

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

// Request represents input for creating a product.
type Request struct {
	Name        string
	Description string
	Category    string
	// BasePriceNumerator and BasePriceDenominator represent the base price as a rational.
	// E.g., $19.99 = 1999/100.
	BasePriceNumerator   int64
	BasePriceDenominator int64
}

// Interactor implements the CreateProduct usecase following the Golden Mutation Pattern.
type Interactor struct {
	repo      contracts.ProductRepo
	outboxRepo contracts.OutboxRepo
	committer *committer.PlanCommitter
	clock     clock.Clock
}

// New creates a new CreateProduct interactor.
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

// Execute creates a new product and persists it atomically with events.
func (it *Interactor) Execute(ctx context.Context, req Request) (string, error) {
	// 1. Create aggregate
	basePrice, err := domain.NewMoneyFromFraction(
		req.BasePriceNumerator,
		req.BasePriceDenominator,
	)
	if err != nil {
		return "", fmt.Errorf("invalid base price: %w", err)
	}

	now := it.clock.Now()
	product := domain.NewProduct(
		generateID(), // TODO: use proper UUID generator
		req.Name,
		req.Description,
		req.Category,
		basePrice,
		now,
	)

	// 2. Domain validation (already done in constructor)

	// 3. Build commit plan
	plan := commitplan.NewPlan()

	// 4. Get mutations from repository
	if mut := it.repo.InsertMut(product); mut != nil {
		plan.Add(mut)
	}

	// 5. Add outbox events
	for _, event := range product.DomainEvents() {
		enriched := enrichEvent(product.ID(), event)
		if outboxMut := it.outboxRepo.InsertMut(enriched); outboxMut != nil {
			plan.Add(outboxMut)
		}
	}

	// 6. Apply plan (usecase applies, NOT handler!)
	if err := it.committer.Apply(ctx, plan); err != nil {
		return "", err
	}

	product.ClearDomainEvents()
	return product.ID(), nil
}

// enrichEvent converts a domain event to an enriched outbox event.
func enrichEvent(aggregateID string, event domain.DomainEvent) *contracts.EnrichedEvent {
	payload, _ := json.Marshal(event)
	return &contracts.EnrichedEvent{
		EventID:    generateID(),
		EventType:  eventType(event),
		AggregateID: aggregateID,
		Payload:    payload,
		Status:     "pending",
	}
}

// eventType returns a string identifier for the event type.
func eventType(event domain.DomainEvent) string {
	switch event.(type) {
	case domain.ProductCreatedEvent:
		return "product.created"
	case domain.ProductUpdatedEvent:
		return "product.updated"
	case domain.ProductActivatedEvent:
		return "product.activated"
	case domain.ProductDeactivatedEvent:
		return "product.deactivated"
	case domain.DiscountAppliedEvent:
		return "discount.applied"
	case domain.DiscountRemovedEvent:
		return "discount.removed"
	default:
		return "unknown"
	}
}

// generateID generates a simple ID. TODO: replace with proper UUID.
func generateID() string {
	return fmt.Sprintf("id-%d", time.Now().UnixNano())
}

