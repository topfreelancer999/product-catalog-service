package domain

import "errors"

// Domain error placeholders.

var (
	ErrProductNotActive      = errors.New("product not active")
	ErrInvalidDiscountPeriod = errors.New("invalid discount period")
)

