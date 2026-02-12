package repo

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"cloud.google.com/go/spanner"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/models/mproduct"
)

// ProductRepo implements contracts.ProductRepo using Spanner.
type ProductRepo struct {
	client *spanner.Client
}

// NewProductRepo creates a new ProductRepo with the given Spanner client.
func NewProductRepo(client *spanner.Client) *ProductRepo {
	return &ProductRepo{client: client}
}

// InsertMut returns a mutation to insert a new product.
// Returns nil if product is nil.
func (r *ProductRepo) InsertMut(p *domain.Product) *spanner.Mutation {
	if p == nil {
		return nil
	}

	baseNum, baseDen := p.BasePrice().Fraction()

	model := &mproduct.Product{
		ProductID:            p.ID(),
		Name:                 p.Name(),
		Description:          p.Description(),
		Category:             p.Category(),
		BasePriceNumerator:   baseNum,
		BasePriceDenominator: baseDen,
		Status:               string(p.Status()),
		CreatedAt:            p.CreatedAt(),
		UpdatedAt:            p.UpdatedAt(),
	}

	if discount := p.Discount(); discount != nil {
		// Convert discount percentage to NUMERIC
		percent := discount.Percentage()
		// NUMERIC in Spanner is stored as string representation
		model.DiscountPercent = &spanner.NullNumeric{
			Numeric: spanner.Numeric(percent.String()),
			Valid:   true,
		}
		model.DiscountStartDate = spanner.NullTime{
			Time:  discount.StartAt(),
			Valid: true,
		}
		model.DiscountEndDate = spanner.NullTime{
			Time:  discount.EndAt(),
			Valid: true,
		}
	}

	if archivedAt := p.ArchivedAt(); archivedAt != nil {
		model.ArchivedAt = spanner.NullTime{
			Time:  *archivedAt,
			Valid: true,
		}
	}

	return mproduct.InsertMut(model)
}

// UpdateMut returns a mutation to update changed fields of a product.
// Uses change tracker to build targeted updates.
// Returns nil if no changes are dirty.
func (r *ProductRepo) UpdateMut(p *domain.Product) *spanner.Mutation {
	if p == nil {
		return nil
	}

	updates := make(map[string]interface{})

	if p.Changes().Dirty(domain.FieldName) {
		updates[mproduct.Name] = p.Name()
	}

	if p.Changes().Dirty(domain.FieldDescription) {
		updates[mproduct.Description] = p.Description()
	}

	if p.Changes().Dirty(domain.FieldCategory) {
		updates[mproduct.Category] = p.Category()
	}

	if p.Changes().Dirty(domain.FieldStatus) {
		updates[mproduct.Status] = string(p.Status())
	}

	if p.Changes().Dirty(domain.FieldDiscount) {
		if discount := p.Discount(); discount != nil {
			percent := discount.Percentage()
			updates[mproduct.DiscountPercent] = spanner.NullNumeric{
				Numeric: spanner.Numeric(percent.String()),
				Valid:   true,
			}
			updates[mproduct.DiscountStartDate] = spanner.NullTime{
				Time:  discount.StartAt(),
				Valid: true,
			}
			updates[mproduct.DiscountEndDate] = spanner.NullTime{
				Time:  discount.EndAt(),
				Valid: true,
			}
		} else {
			// Clear discount
			updates[mproduct.DiscountPercent] = spanner.NullNumeric{Valid: false}
			updates[mproduct.DiscountStartDate] = spanner.NullTime{Valid: false}
			updates[mproduct.DiscountEndDate] = spanner.NullTime{Valid: false}
		}
	}

	if p.Changes().Dirty(domain.FieldArchivedAt) {
		if archivedAt := p.ArchivedAt(); archivedAt != nil {
			updates[mproduct.ArchivedAt] = spanner.NullTime{
				Time:  *archivedAt,
				Valid: true,
			}
		} else {
			updates[mproduct.ArchivedAt] = spanner.NullTime{Valid: false}
		}
	}

	// Always update updated_at if there are any changes
	if len(updates) > 0 {
		updates[mproduct.UpdatedAt] = p.UpdatedAt()
		return mproduct.UpdateMut(p.ID(), updates)
	}

	return nil // No changes
}

// FindByID loads a product aggregate by ID.
// Returns domain error if not found.
func (r *ProductRepo) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	row, err := r.client.Single().ReadRow(ctx, mproduct.TableName, spanner.Key{id}, []string{
		mproduct.ProductID,
		mproduct.Name,
		mproduct.Description,
		mproduct.Category,
		mproduct.BasePriceNumerator,
		mproduct.BasePriceDenominator,
		mproduct.DiscountPercent,
		mproduct.DiscountStartDate,
		mproduct.DiscountEndDate,
		mproduct.Status,
		mproduct.CreatedAt,
		mproduct.UpdatedAt,
		mproduct.ArchivedAt,
	})
	if err != nil {
		if spanner.ErrCode(err) == spanner.ErrCode(spanner.ErrNotFound) {
			return nil, fmt.Errorf("product not found")
		}
		return nil, err
	}

	var model mproduct.Product
	if err := row.ToStruct(&model); err != nil {
		return nil, fmt.Errorf("failed to parse product row: %w", err)
	}

	return r.toDomain(&model)
}

// toDomain converts a database model to a domain aggregate.
func (r *ProductRepo) toDomain(model *mproduct.Product) (*domain.Product, error) {
	basePrice, err := domain.NewMoneyFromFraction(
		model.BasePriceNumerator,
		model.BasePriceDenominator,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid base price: %w", err)
	}

	var discount *domain.Discount
	if model.DiscountPercent.Valid && model.DiscountStartDate.Valid && model.DiscountEndDate.Valid {
		// Parse NUMERIC string to big.Rat
		percentStr := string(model.DiscountPercent.Numeric)
		percent := new(big.Rat)
		if _, ok := percent.SetString(percentStr); !ok {
			return nil, fmt.Errorf("invalid discount percentage: %s", percentStr)
		}

		discount, err = domain.NewDiscount(
			percent,
			model.DiscountStartDate.Time,
			model.DiscountEndDate.Time,
		)
		if err != nil {
			return nil, fmt.Errorf("invalid discount: %w", err)
		}
	}

	var archivedAt *time.Time
	if model.ArchivedAt.Valid {
		archivedAt = &model.ArchivedAt.Time
	}

	return domain.RehydrateProduct(
		model.ProductID,
		model.Name,
		model.Description,
		model.Category,
		basePrice,
		discount,
		domain.ProductStatus(model.Status),
		archivedAt,
		model.CreatedAt,
		model.UpdatedAt,
	), nil
}
