package archiveproduct

import (
	"context"
	"fmt"
	"time"

	"github.com/Vektor-AI/commitplan"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/pkg/clock"
	"product-catalog-service/internal/pkg/committer"
)

// Request represents input for archiving a product (soft delete).
type Request struct {
	ProductID string
}

// Interactor implements the ArchiveProduct usecase following the Golden Mutation Pattern.
type Interactor struct {
	repo      contracts.ProductRepo
	committer *committer.PlanCommitter
	clock     clock.Clock
}

// New creates a new ArchiveProduct interactor.
func New(
	repo contracts.ProductRepo,
	committer *committer.PlanCommitter,
	clock clock.Clock,
) *Interactor {
	return &Interactor{
		repo:      repo,
		committer: committer,
		clock:     clock,
	}
}

// Execute archives a product (soft delete).
// Note: Archive does not emit domain events per the task spec.
func (it *Interactor) Execute(ctx context.Context, req Request) error {
	// 1. Load aggregate
	product, err := it.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	// 2. Call domain method
	now := it.clock.Now()
	product.Archive(now)

	// 3. Build commit plan
	plan := commitplan.NewPlan()

	// 4. Get mutations from repository
	if mut := it.repo.UpdateMut(product); mut != nil {
		plan.Add(mut)
	}

	// 5. No outbox events for archive (per task spec)

	// 6. Apply plan
	if err := it.committer.Apply(ctx, plan); err != nil {
		return err
	}

	return nil
}
