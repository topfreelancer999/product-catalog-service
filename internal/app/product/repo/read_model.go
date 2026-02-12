package repo

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/models/mproduct"
)

// ReadModel implements contracts.ReadModel using Spanner for query-side reads.
type ReadModel struct {
	client *spanner.Client
}

// NewReadModel creates a new ReadModel with the given Spanner client.
func NewReadModel(client *spanner.Client) *ReadModel {
	return &ReadModel{client: client}
}

// GetProductByID returns a single product by ID or an error if it does not exist.
func (r *ReadModel) GetProductByID(ctx context.Context, id string) (*contracts.ProductRecord, error) {
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

	return r.toRecord(&model), nil
}

// ListActiveProducts returns active products, optionally filtered by category,
// using simple cursor-based pagination.
func (r *ReadModel) ListActiveProducts(
	ctx context.Context,
	category *string,
	pageSize int,
	pageToken string,
) ([]*contracts.ProductRecord, string, error) {
	// Handle pagination defaults
	if pageSize <= 0 {
		pageSize = 50 // default
	}
	if pageSize > 1000 {
		pageSize = 1000 // max
	}
	limit := pageSize + 1 // fetch one extra to check for next page

	// Build query with proper WHERE clause
	sql := `SELECT product_id, name, description, category, 
	           base_price_numerator, base_price_denominator,
	           discount_percent, discount_start_date, discount_end_date,
	           status
	      FROM products
	      WHERE status = @status`
	
	params := map[string]interface{}{
		"status": "active",
	}

	// Add category filter if provided
	if category != nil && *category != "" {
		sql += " AND category = @category"
		params["category"] = *category
	}

	// Handle cursor-based pagination
	if pageToken != "" {
		decoded, err := base64.StdEncoding.DecodeString(pageToken)
		if err == nil {
			sql += " AND product_id > @cursor"
			params["cursor"] = string(decoded)
		}
	}

	sql += " ORDER BY product_id LIMIT @limit"
	params["limit"] = limit

	stmt := spanner.Statement{
		SQL:    sql,
		Params: params,
	}

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var records []*contracts.ProductRecord
	var lastID string

	for {
		row, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, "", err
		}

		var model mproduct.Product
		if err := row.ToStruct(&model); err != nil {
			return nil, "", fmt.Errorf("failed to parse product row: %w", err)
		}

		// Check if we've exceeded page size
		if len(records) >= pageSize {
			lastID = model.ProductID
			break
		}

		records = append(records, r.toRecord(&model))
	}

	// Generate next page token if there are more results
	nextToken := ""
	if lastID != "" {
		nextToken = base64.StdEncoding.EncodeToString([]byte(lastID))
	}

	return records, nextToken, nil
}

// toRecord converts a database model to a ProductRecord.
func (r *ReadModel) toRecord(model *mproduct.Product) *contracts.ProductRecord {
	record := &contracts.ProductRecord{
		ProductID:            model.ProductID,
		Name:                 model.Name,
		Description:          model.Description,
		Category:             model.Category,
		BasePriceNumerator:   model.BasePriceNumerator,
		BasePriceDenominator: model.BasePriceDenominator,
		Status:               model.Status,
	}

	if model.DiscountPercent.Valid && model.DiscountStartDate.Valid && model.DiscountEndDate.Valid {
		// Parse NUMERIC string to big.Rat
		percentStr := string(model.DiscountPercent.Numeric)
		percent := new(big.Rat)
		if _, ok := percent.SetString(percentStr); ok {
			record.DiscountPercent = percent
			record.DiscountStart = &model.DiscountStartDate.Time
			record.DiscountEnd = &model.DiscountEndDate.Time
		}
	}

	return record
}
