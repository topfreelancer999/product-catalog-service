package listproducts

// ProductListItemDTO represents a single item in the products list.
type ProductListItemDTO struct {
	ID       string
	Name     string
	Category string
	Status   string

	EffectivePriceNumerator   int64
	EffectivePriceDenominator int64
}

// ListResultDTO is the result of the ListProducts query.
type ListResultDTO struct {
	Items         []ProductListItemDTO
	NextPageToken string
}

