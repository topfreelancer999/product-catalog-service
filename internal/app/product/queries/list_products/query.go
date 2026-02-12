package listproducts

import (
	"context"
	"time"

	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/app/product/domain/services"
)

// Request represents input parameters for the ListProducts query.
type Request struct {
	Category   *string
	PageSize   int
	PageToken  string
	// As-of time for price calculation; if zero, current time is used.
	Now time.Time
}

// Query implements "List active products with pagination" and
// optional filtering by category.
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

// Execute runs the list query.
func (q *Query) Execute(ctx context.Context, req Request) (*ListResultDTO, error) {
	now := req.Now
	if now.IsZero() {
		now = time.Now()
	}

	records, nextToken, err := q.readModel.ListActiveProducts(
		ctx,
		req.Category,
		req.PageSize,
		req.PageToken,
	)
	if err != nil {
		return nil, err
	}

	items := make([]ProductListItemDTO, 0, len(records))

	for _, r := range records {
		basePrice, err := domain.NewMoneyFromFraction(
			r.BasePriceNumerator,
			r.BasePriceDenominator,
		)
		if err != nil {
			return nil, err
		}

		var discount *domain.Discount
		if r.DiscountPercent != nil && r.DiscountStart != nil && r.DiscountEnd != nil {
			discount, err = domain.NewDiscount(
				r.DiscountPercent,
				*r.DiscountStart,
				*r.DiscountEnd,
			)
			if err != nil {
				discount = nil
			}
		}

		product := domain.RehydrateProduct(
			r.ProductID,
			r.Name,
			r.Description,
			r.Category,
			basePrice,
			discount,
			domain.ProductStatus(r.Status),
			nil,
			time.Time{},
			time.Time{},
		)

		// Calculate effective price at current time (only applies valid discounts)
		effective := q.pricing.EffectivePrice(product, now)
		if effective == nil {
			// Skip this item if price calculation fails
			continue
		}
		num, den := effective.Fraction()

		items = append(items, ProductListItemDTO{
			ID:                        r.ProductID,
			Name:                      r.Name,
			Category:                  r.Category,
			Status:                    r.Status,
			EffectivePriceNumerator:   num,
			EffectivePriceDenominator: den,
		})
	}

	return &ListResultDTO{
		Items:         items,
		NextPageToken: nextToken,
	}, nil
}

