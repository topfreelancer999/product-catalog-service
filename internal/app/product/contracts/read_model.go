package contracts

import (
	"context"
	"math/big"
	"time"
)

// ProductRecord is a read-model representation of a product row.
// It is intentionally close to the storage model but independent from it.
type ProductRecord struct {
	ProductID string
	Name      string
	Description string
	Category  string

	BasePriceNumerator   int64
	BasePriceDenominator int64

	// DiscountPercent is expressed as a rational number (e.g. 20% == 20/100).
	DiscountPercent *big.Rat
	DiscountStart   *time.Time
	DiscountEnd     *time.Time

	Status string
}

// ReadModel defines interfaces for query-side data access.
type ReadModel interface {
	// GetProductByID returns a single product by ID or an error
	// if it does not exist or the read fails.
	GetProductByID(ctx context.Context, id string) (*ProductRecord, error)

	// ListActiveProducts returns active products, optionally filtered by category,
	// using simple cursor-based pagination.
	ListActiveProducts(
		ctx context.Context,
		category *string,
		pageSize int,
		pageToken string,
	) (records []*ProductRecord, nextPageToken string, err error)
}

