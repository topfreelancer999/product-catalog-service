package mproduct

import (
	"cloud.google.com/go/spanner"
	"time"
)

// Product represents a row in the products table.
// This is the database model, separate from the domain aggregate.
type Product struct {
	ProductID            string
	Name                 string
	Description          string
	Category             string
	BasePriceNumerator   int64
	BasePriceDenominator int64
	DiscountPercent      *spanner.NullNumeric
	DiscountStartDate    spanner.NullTime
	DiscountEndDate      spanner.NullTime
	Status               string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ArchivedAt           spanner.NullTime
}

// InsertMut returns a mutation to insert a new product.
func InsertMut(p *Product) *spanner.Mutation {
	if p == nil {
		return nil
	}
	return spanner.Insert(TableName, []string{
		ProductID,
		Name,
		Description,
		Category,
		BasePriceNumerator,
		BasePriceDenominator,
		DiscountPercent,
		DiscountStartDate,
		DiscountEndDate,
		Status,
		CreatedAt,
		UpdatedAt,
		ArchivedAt,
	}, []interface{}{
		p.ProductID,
		p.Name,
		p.Description,
		p.Category,
		p.BasePriceNumerator,
		p.BasePriceDenominator,
		p.DiscountPercent,
		p.DiscountStartDate,
		p.DiscountEndDate,
		p.Status,
		p.CreatedAt,
		p.UpdatedAt,
		p.ArchivedAt,
	})
}

// UpdateMut returns a mutation to update specific fields of a product.
func UpdateMut(productID string, updates map[string]interface{}) *spanner.Mutation {
	if len(updates) == 0 {
		return nil
	}
	return spanner.Update(TableName, updates)
}

