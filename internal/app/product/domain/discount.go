package domain

import (
	"fmt"
	"math/big"
	"time"
)

// Discount is a value object that represents a percentage-based
// discount with a validity period.
//
// Percentage is represented as a rational number, e.g. 20% == 20/100.
type Discount struct {
	percentage *big.Rat
	startAt    time.Time
	endAt      time.Time
}

// NewDiscount creates a discount with the given percentage and time window.
// Percentage must be between 0 and 1 inclusive.
// start must be before end.
func NewDiscount(percentage *big.Rat, start, end time.Time) (*Discount, error) {
	if percentage == nil {
		return nil, fmt.Errorf("discount percentage is required")
	}

	// bounds: 0 <= percentage <= 1
	if percentage.Cmp(new(big.Rat).SetInt64(0)) < 0 ||
		percentage.Cmp(new(big.Rat).SetInt64(1)) > 0 {
		return nil, fmt.Errorf("discount percentage must be between 0 and 1")
	}

	if end.Before(start) {
		return nil, ErrInvalidDiscountPeriod
	}

	return &Discount{
		percentage: new(big.Rat).Set(percentage),
		startAt:    start,
		endAt:      end,
	}, nil
}

// Percentage returns an immutable copy of the percentage.
func (d *Discount) Percentage() *big.Rat {
	if d == nil || d.percentage == nil {
		return nil
	}
	return new(big.Rat).Set(d.percentage)
}

// StartAt returns the discount start time.
func (d *Discount) StartAt() time.Time {
	if d == nil {
		return time.Time{}
	}
	return d.startAt
}

// EndAt returns the discount end time.
func (d *Discount) EndAt() time.Time {
	if d == nil {
		return time.Time{}
	}
	return d.endAt
}

// IsValidAt returns true if the discount is valid at the given time.
func (d *Discount) IsValidAt(t time.Time) bool {
	if d == nil {
		return false
	}
	if t.Before(d.startAt) {
		return false
	}
	if t.After(d.endAt) {
		return false
	}
	return true
}

