package services

import (
	"math/big"
	"time"

	"product-catalog-service/internal/app/product/domain"
)

// PricingCalculator encapsulates rules for computing effective price.
type PricingCalculator struct{}

// EffectivePrice returns the effective price for a product at the given time,
// taking into account its discount (if valid at that time).
//
// If no valid discount exists at the given time, base price is returned.
// Uses precise decimal arithmetic via big.Rat.
func (PricingCalculator) EffectivePrice(p *domain.Product, at time.Time) *domain.Money {
	if p == nil || p.BasePrice() == nil {
		return nil
	}

	base := p.BasePrice()
	d := p.Discount()
	
	// Only apply discount if it exists and is valid at the given time
	if d == nil || !d.IsValidAt(at) {
		return base
	}

	// finalPrice = base * (1 - percentage)
	// Uses big.Rat for precise decimal arithmetic
	one := big.NewRat(1, 1)
	discountPart := new(big.Rat).Sub(one, d.Percentage())
	return base.MultiplyBy(discountPart)
}

