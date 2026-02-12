package getproduct

// ProductDTO is the response model for the GetProduct query.
// Prices are exposed as rational numerator/denominator pair to
// preserve full precision for callers.
type ProductDTO struct {
	ID          string
	Name        string
	Description string
	Category    string
	Status      string

	EffectivePriceNumerator   int64
	EffectivePriceDenominator int64
}

