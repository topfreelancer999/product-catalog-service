package applydiscount

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/Vektor-AI/commitplan"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/pkg/clock"
	"product-catalog-service/internal/pkg/committer"
)

// Request represents input for applying a discount to a product.
type Request struct {
	ProductID string
	// PercentageNumerator and PercentageDenominator represent the discount percentage as a rational.
	// E.g., 20% = 20/100, 15.5% = 155/1000.
	PercentageNumerator   int64
	PercentageDenominator int64
	StartDate             time.Time
	EndDate               time.Time
}

// Interactor implements the ApplyDiscount usecase following the Golden Mutation Pattern.
// Enforces: only one active discount per product at a time (replaces existing).
type Interactor struct {
	repo      contracts.ProductRepo
	outboxRepo contracts.OutboxRepo
	committer *committer.PlanCommitter
	clock     clock.Clock
}

// New creates a new ApplyDiscount interactor.
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

// Execute applies a percentage-based discount to a product.
// The discount must have valid start/end dates, and the product must be active.
// If a discount already exists, it is replaced (only one active discount per product).
func (it *Interactor) Execute(ctx context.Context, req Request) error {
	// 1. Load aggregate
	product, err := it.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	// 2. Create discount value object (validates percentage and dates)
	percentage := big.NewRat(req.PercentageNumerator, req.PercentageDenominator)
	discount, err := domain.NewDiscount(percentage, req.StartDate, req.EndDate)
	if err != nil {
		return fmt.Errorf("invalid discount: %w", err)
	}

	// 3. Call domain method (validates product is active and discount is valid at current time)
	now := it.clock.Now()
	if err := product.ApplyDiscount(discount, now); err != nil {
		return err
	}

	// 4. Build commit plan
	plan := commitplan.NewPlan()

	// 5. Get mutations from repository
	if mut := it.repo.UpdateMut(product); mut != nil {
		plan.Add(mut)
	}

	// 6. Add outbox events
	for _, event := range product.DomainEvents() {
		enriched := enrichEvent(product.ID(), event)
		if outboxMut := it.outboxRepo.InsertMut(enriched); outboxMut != nil {
			plan.Add(outboxMut)
		}
	}

	// 7. Apply plan atomically
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
	case domain.DiscountAppliedEvent:
		return "discount.applied"
	default:
		return "unknown"
	}
}

func generateID() string {
	return fmt.Sprintf("id-%d", time.Now().UnixNano())
}
