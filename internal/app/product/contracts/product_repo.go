package contracts

import (
	"context"

	"cloud.google.com/go/spanner"
	"product-catalog-service/internal/app/product/domain"
)

// ProductRepo defines the write-side repository interface for Product aggregates.
// Implementations must return mutations instead of applying them.
type ProductRepo interface {
	// InsertMut returns a mutation to insert a new product.
	// Returns nil if product is nil.
	InsertMut(p *domain.Product) *spanner.Mutation

	// UpdateMut returns a mutation to update changed fields of a product.
	// Uses change tracker to build targeted updates.
	// Returns nil if no changes are dirty.
	UpdateMut(p *domain.Product) *spanner.Mutation

	// FindByID loads a product aggregate by ID.
	// Returns domain error if not found.
	FindByID(ctx context.Context, id string) (*domain.Product, error)
}

