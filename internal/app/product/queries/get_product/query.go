package getproduct

import (
	"context"
	"fmt"
	"time"

	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/app/product/domain/services"
)

// Request represents input parameters for GetProduct query.
type Request struct {
	ProductID string
	// As-of time for price calculation; if zero, current time is used.
	Now time.Time
}

// Query implements "Get product by ID with current effective price".
type Query struct {
	readModel contracts.ReadModel
	pricing   services.PricingCalculator
}

func New(readModel contracts.ReadModel, pricing services.PricingCalculator) *Query {
	return &Query{
		readModel: readModel,
		pricing:   pricing,
	}
}

// Execute runs the query and returns a DTO with current effective price.
func (q *Query) Execute(ctx context.Context, req Request) (*ProductDTO, error) {
	record, err := q.readModel.GetProductByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	now := req.Now
	if now.IsZero() {
		now = time.Now()
	}

	basePrice, err := domain.NewMoneyFromFraction(
		record.BasePriceNumerator,
		record.BasePriceDenominator,
	)
	if err != nil {
		return nil, err
	}

	var discount *domain.Discount
	if record.DiscountPercent != nil && record.DiscountStart != nil && record.DiscountEnd != nil {
		discount, err = domain.NewDiscount(
			record.DiscountPercent,
			*record.DiscountStart,
			*record.DiscountEnd,
		)
		if err != nil {
			// if stored discount is invalid, treat as no discount
			discount = nil
		}
	}

	product := domain.RehydrateProduct(
		record.ProductID,
		record.Name,
		record.Description,
		record.Category,
		basePrice,
		discount,
		domain.ProductStatus(record.Status),
		nil, // archivedAt not required for this query
		time.Time{}, // createdAt not required
		time.Time{}, // updatedAt not required
	)

	// Calculate effective price at current time (only applies valid discounts)
	effective := q.pricing.EffectivePrice(product, now)
	if effective == nil {
		return nil, fmt.Errorf("failed to calculate effective price")
	}
	num, den := effective.Fraction()

	return &ProductDTO{
		ID:                       record.ProductID,
		Name:                     record.Name,
		Description:              record.Description,
		Category:                 record.Category,
		Status:                   record.Status,
		EffectivePriceNumerator:   num,
		EffectivePriceDenominator: den,
	}, nil
}

