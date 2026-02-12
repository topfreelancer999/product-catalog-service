package domain

import (
	"fmt"
	"math/big"
)

// Money is a simple value object that wraps *big.Rat to represent
// monetary values with arbitrary precision.
//
// It is intentionally small and focused – all business rules live
// on the Product aggregate or domain services.
type Money struct {
	value *big.Rat
}

// NewMoneyFromFraction creates Money from integer numerator/denominator.
// Denominator must be > 0.
func NewMoneyFromFraction(numerator, denominator int64) (*Money, error) {
	if denominator <= 0 {
		return nil, fmt.Errorf("money denominator must be > 0")
	}
	r := big.NewRat(numerator, denominator)
	return &Money{value: r}, nil
}

// NewMoneyFromRat wraps a cloned *big.Rat as Money.
func NewMoneyFromRat(r *big.Rat) *Money {
	if r == nil {
		return nil
	}
	return &Money{value: new(big.Rat).Set(r)}
}

// Rat returns an immutable copy of the underlying value.
func (m *Money) Rat() *big.Rat {
	if m == nil || m.value == nil {
		return nil
	}
	return new(big.Rat).Set(m.value)
}

// MultiplyBy multiplies this Money by the given ratio and returns a new Money.
func (m *Money) MultiplyBy(ratio *big.Rat) *Money {
	if m == nil || m.value == nil || ratio == nil {
		return nil
	}
	out := new(big.Rat).Mul(m.value, ratio)
	return &Money{value: out}
}

// Subtract subtracts other from this Money and returns a new Money.
func (m *Money) Subtract(other *Money) *Money {
	if m == nil || m.value == nil || other == nil || other.value == nil {
		return nil
	}
	out := new(big.Rat).Sub(m.value, other.value)
	return &Money{value: out}
}

// Compare compares this Money with other.
// Returns -1 if m < other, 0 if equal, 1 if m > other.
func (m *Money) Compare(other *Money) int {
	if m == nil || m.value == nil {
		if other == nil || other.value == nil {
			return 0
		}
		return -1
	}
	if other == nil || other.value == nil {
		return 1
	}
	return m.value.Cmp(other.value)
}

// Fraction returns the internal numerator and denominator representation.
// This is primarily used by infrastructure mapping to the database schema.
func (m *Money) Fraction() (numerator, denominator int64) {
	if m == nil || m.value == nil {
		return 0, 1
	}
	return m.value.Num().Int64(), m.value.Denom().Int64()
}

