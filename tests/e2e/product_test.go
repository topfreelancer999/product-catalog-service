package e2e

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"

	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/app/product/domain/services"
	"product-catalog-service/internal/app/product/queries/getproduct"
	"product-catalog-service/internal/app/product/queries/listproducts"
	"product-catalog-service/internal/app/product/repo"
	createproduct "product-catalog-service/internal/app/product/usecases/create_product"
	updateproduct "product-catalog-service/internal/app/product/usecases/update_product"
	activateproduct "product-catalog-service/internal/app/product/usecases/activate_product"
	deactivateproduct "product-catalog-service/internal/app/product/usecases/deactivate_product"
	applydiscount "product-catalog-service/internal/app/product/usecases/apply_discount"
	removediscount "product-catalog-service/internal/app/product/usecases/remove_discount"
	archiveproduct "product-catalog-service/internal/app/product/usecases/archive_product"
	"product-catalog-service/internal/pkg/clock"
	"product-catalog-service/internal/pkg/committer"
)

var (
	testDB     *spanner.Client
	testCtx    context.Context
	testClock  clock.Clock
	committer_ *committer.PlanCommitter
)

func setupTestDB(t *testing.T) {
	// TODO: Initialize Spanner emulator connection
	// For now, tests will need Spanner emulator running via docker-compose
	databaseName := "projects/test-project/instances/test-instance/databases/test-db"
	
	client, err := spanner.NewClient(context.Background(), databaseName)
	if err != nil {
		t.Skipf("Skipping test: Spanner emulator not available: %v", err)
		return
	}
	
	testDB = client
	testCtx = context.Background()
	testClock = clock.SystemClock{}
	committer_ = committer.New(client)
}

func teardownTestDB(t *testing.T) {
	if testDB != nil {
		testDB.Close()
	}
}

func getOutboxEvents(t *testing.T, aggregateID string) []OutboxEvent {
	stmt := spanner.Statement{
		SQL: `SELECT event_id, event_type, aggregate_id, payload, status, created_at
		      FROM outbox_events
		      WHERE aggregate_id = @aggregate_id
		      ORDER BY created_at`,
		Params: map[string]interface{}{
			"aggregate_id": aggregateID,
		},
	}

	iter := testDB.Single().Query(testCtx, stmt)
	defer iter.Stop()

	var events []OutboxEvent
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		require.NoError(t, err)

		var event OutboxEvent
		err = row.ToStruct(&event)
		require.NoError(t, err)
		events = append(events, event)
	}

	return events
}

type OutboxEvent struct {
	EventID     string
	EventType   string
	AggregateID string
	Payload     []byte
	Status      string
	CreatedAt   time.Time
}

func TestProductCreationFlow(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	productRepo := repo.NewProductRepo(testDB)
	outboxRepo := repo.NewOutboxRepo()
	readModel := repo.NewReadModel(testDB)
	pricing := services.PricingCalculator{}

	createUsecase := createproduct.New(productRepo, outboxRepo, committer_, testClock)
	getQuery := getproduct.New(readModel, pricing)

	// Test: Create product
	productID, err := createUsecase.Execute(testCtx, createproduct.Request{
		Name:                 "Test Product",
		Description:          "A test product",
		Category:             "electronics",
		BasePriceNumerator:   1999,
		BasePriceDenominator: 100, // $19.99
	})
	require.NoError(t, err)
	require.NotEmpty(t, productID)

	// Verify: Query returns correct data
	product, err := getQuery.Execute(testCtx, getproduct.Request{
		ProductID: productID,
	})
	require.NoError(t, err)
	assert.Equal(t, "Test Product", product.Name)
	assert.Equal(t, "A test product", product.Description)
	assert.Equal(t, "electronics", product.Category)
	assert.Equal(t, "inactive", product.Status)
	assert.Equal(t, int64(1999), product.EffectivePriceNumerator)
	assert.Equal(t, int64(100), product.EffectivePriceDenominator)

	// Verify: Outbox event was created
	events := getOutboxEvents(t, productID)
	require.Len(t, events, 1)
	assert.Equal(t, "product.created", events[0].EventType)
	assert.Equal(t, productID, events[0].AggregateID)
	assert.Equal(t, "pending", events[0].Status)
}

func TestProductUpdateFlow(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	productRepo := repo.NewProductRepo(testDB)
	outboxRepo := repo.NewOutboxRepo()
	readModel := repo.NewReadModel(testDB)
	pricing := services.PricingCalculator{}

	createUsecase := createproduct.New(productRepo, outboxRepo, committer_, testClock)
	updateUsecase := updateproduct.New(productRepo, outboxRepo, committer_, testClock)
	getQuery := getproduct.New(readModel, pricing)

	// Setup: Create product
	productID, err := createUsecase.Execute(testCtx, createproduct.Request{
		Name:                 "Original Name",
		Description:          "Original Description",
		Category:             "original",
		BasePriceNumerator:   1000,
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)

	// Test: Update product
	newName := "Updated Name"
	newDesc := "Updated Description"
	newCat := "updated"
	err = updateUsecase.Execute(testCtx, updateproduct.Request{
		ProductID:   productID,
		Name:        &newName,
		Description: &newDesc,
		Category:    &newCat,
	})
	require.NoError(t, err)

	// Verify: Query returns updated data
	product, err := getQuery.Execute(testCtx, getproduct.Request{
		ProductID: productID,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", product.Name)
	assert.Equal(t, "Updated Description", product.Description)
	assert.Equal(t, "updated", product.Category)

	// Verify: Outbox event was created
	events := getOutboxEvents(t, productID)
	require.GreaterOrEqual(t, len(events), 2) // created + updated
	assert.Equal(t, "product.updated", events[len(events)-1].EventType)
}

func TestDiscountApplicationFlow(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	productRepo := repo.NewProductRepo(testDB)
	outboxRepo := repo.NewOutboxRepo()
	readModel := repo.NewReadModel(testDB)
	pricing := services.PricingCalculator{}

	createUsecase := createproduct.New(productRepo, outboxRepo, committer_, testClock)
	activateUsecase := activateproduct.New(productRepo, outboxRepo, committer_, testClock)
	applyDiscountUsecase := applydiscount.New(productRepo, outboxRepo, committer_, testClock)
	getQuery := getproduct.New(readModel, pricing)

	// Setup: Create and activate product
	productID, err := createUsecase.Execute(testCtx, createproduct.Request{
		Name:                 "Discounted Product",
		Description:          "A product with discount",
		Category:             "electronics",
		BasePriceNumerator:   10000, // $100.00
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)

	err = activateUsecase.Execute(testCtx, activateproduct.Request{
		ProductID: productID,
	})
	require.NoError(t, err)

	// Test: Apply 20% discount
	now := time.Now()
	err = applyDiscountUsecase.Execute(testCtx, applydiscount.Request{
		ProductID:            productID,
		PercentageNumerator:   20,
		PercentageDenominator: 100, // 20%
		StartDate:             now.Add(-1 * time.Hour),
		EndDate:               now.Add(24 * time.Hour),
	})
	require.NoError(t, err)

	// Verify: Effective price is calculated correctly (80% of base price)
	product, err := getQuery.Execute(testCtx, getproduct.Request{
		ProductID: productID,
		Now:        now,
	})
	require.NoError(t, err)
	
	// Expected: $100.00 * 0.8 = $80.00 = 8000/100
	assert.Equal(t, int64(8000), product.EffectivePriceNumerator)
	assert.Equal(t, int64(100), product.EffectivePriceDenominator)

	// Verify: Outbox event was created
	events := getOutboxEvents(t, productID)
	require.GreaterOrEqual(t, len(events), 3) // created + activated + discount applied
	discountApplied := false
	for _, e := range events {
		if e.EventType == "discount.applied" {
			discountApplied = true
			break
		}
	}
	assert.True(t, discountApplied, "discount.applied event should exist")
}

func TestProductActivationDeactivation(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	productRepo := repo.NewProductRepo(testDB)
	outboxRepo := repo.NewOutboxRepo()
	readModel := repo.NewReadModel(testDB)
	pricing := services.PricingCalculator{}

	createUsecase := createproduct.New(productRepo, outboxRepo, committer_, testClock)
	activateUsecase := activateproduct.New(productRepo, outboxRepo, committer_, testClock)
	deactivateUsecase := deactivateproduct.New(productRepo, outboxRepo, committer_, testClock)
	getQuery := getproduct.New(readModel, pricing)

	// Setup: Create product (starts as inactive)
	productID, err := createUsecase.Execute(testCtx, createproduct.Request{
		Name:                 "Test Product",
		Description:          "Test",
		Category:             "test",
		BasePriceNumerator:   1000,
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)

	// Verify: Product starts as inactive
	product, err := getQuery.Execute(testCtx, getproduct.Request{
		ProductID: productID,
	})
	require.NoError(t, err)
	assert.Equal(t, "inactive", product.Status)

	// Test: Activate product
	err = activateUsecase.Execute(testCtx, activateproduct.Request{
		ProductID: productID,
	})
	require.NoError(t, err)

	// Verify: Product is now active
	product, err = getQuery.Execute(testCtx, getproduct.Request{
		ProductID: productID,
	})
	require.NoError(t, err)
	assert.Equal(t, "active", product.Status)

	// Test: Deactivate product
	err = deactivateUsecase.Execute(testCtx, deactivateproduct.Request{
		ProductID: productID,
	})
	require.NoError(t, err)

	// Verify: Product is inactive again
	product, err = getQuery.Execute(testCtx, getproduct.Request{
		ProductID: productID,
	})
	require.NoError(t, err)
	assert.Equal(t, "inactive", product.Status)

	// Verify: Events were created
	events := getOutboxEvents(t, productID)
	require.GreaterOrEqual(t, len(events), 3) // created + activated + deactivated
}

func TestBusinessRuleValidation(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	productRepo := repo.NewProductRepo(testDB)
	outboxRepo := repo.NewOutboxRepo()

	createUsecase := createproduct.New(productRepo, outboxRepo, committer_, testClock)
	applyDiscountUsecase := applydiscount.New(productRepo, outboxRepo, committer_, testClock)

	// Setup: Create inactive product
	productID, err := createUsecase.Execute(testCtx, createproduct.Request{
		Name:                 "Inactive Product",
		Description:          "Test",
		Category:             "test",
		BasePriceNumerator:   1000,
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)

	// Test: Cannot apply discount to inactive product
	now := time.Now()
	err = applyDiscountUsecase.Execute(testCtx, applydiscount.Request{
		ProductID:            productID,
		PercentageNumerator:   10,
		PercentageDenominator: 100,
		StartDate:             now,
		EndDate:               now.Add(24 * time.Hour),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrProductNotActive)

	// Test: Invalid discount period (end before start)
	err = applyDiscountUsecase.Execute(testCtx, applydiscount.Request{
		ProductID:            productID,
		PercentageNumerator:   10,
		PercentageDenominator: 100,
		StartDate:             now.Add(24 * time.Hour),
		EndDate:               now, // end before start
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidDiscountPeriod)
}

func TestOutboxEventCreation(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	productRepo := repo.NewProductRepo(testDB)
	outboxRepo := repo.NewOutboxRepo()

	createUsecase := createproduct.New(productRepo, outboxRepo, committer_, testClock)
	updateUsecase := updateproduct.New(productRepo, outboxRepo, committer_, testClock)
	activateUsecase := activateproduct.New(productRepo, outboxRepo, committer_, testClock)

	// Test: Create product generates event
	productID, err := createUsecase.Execute(testCtx, createproduct.Request{
		Name:                 "Event Test Product",
		Description:          "Test",
		Category:             "test",
		BasePriceNumerator:   1000,
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)

	events := getOutboxEvents(t, productID)
	require.Len(t, events, 1)
	assert.Equal(t, "product.created", events[0].EventType)
	assert.Equal(t, "pending", events[0].Status)

	// Verify payload is valid JSON
	var payload map[string]interface{}
	err = json.Unmarshal(events[0].Payload, &payload)
	require.NoError(t, err)
	assert.Equal(t, productID, payload["ProductID"])

	// Test: Update generates event
	newName := "Updated"
	err = updateUsecase.Execute(testCtx, updateproduct.Request{
		ProductID: productID,
		Name:      &newName,
	})
	require.NoError(t, err)

	events = getOutboxEvents(t, productID)
	require.GreaterOrEqual(t, len(events), 2)
	assert.Equal(t, "product.updated", events[len(events)-1].EventType)

	// Test: Activate generates event
	err = activateUsecase.Execute(testCtx, activateproduct.Request{
		ProductID: productID,
	})
	require.NoError(t, err)

	events = getOutboxEvents(t, productID)
	require.GreaterOrEqual(t, len(events), 3)
	hasActivated := false
	for _, e := range events {
		if e.EventType == "product.activated" {
			hasActivated = true
			break
		}
	}
	assert.True(t, hasActivated, "product.activated event should exist")
}

func TestRemoveDiscount(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	productRepo := repo.NewProductRepo(testDB)
	outboxRepo := repo.NewOutboxRepo()
	readModel := repo.NewReadModel(testDB)
	pricing := services.PricingCalculator{}

	createUsecase := createproduct.New(productRepo, outboxRepo, committer_, testClock)
	activateUsecase := activateproduct.New(productRepo, outboxRepo, committer_, testClock)
	applyDiscountUsecase := applydiscount.New(productRepo, outboxRepo, committer_, testClock)
	removeDiscountUsecase := removediscount.New(productRepo, outboxRepo, committer_, testClock)
	getQuery := getproduct.New(readModel, pricing)

	// Setup: Create, activate, and apply discount
	productID, err := createUsecase.Execute(testCtx, createproduct.Request{
		Name:                 "Test Product",
		Description:          "Test",
		Category:             "test",
		BasePriceNumerator:   10000,
		BasePriceDenominator: 100,
	})
	require.NoError(t, err)

	err = activateUsecase.Execute(testCtx, activateproduct.Request{
		ProductID: productID,
	})
	require.NoError(t, err)

	now := time.Now()
	err = applyDiscountUsecase.Execute(testCtx, applydiscount.Request{
		ProductID:            productID,
		PercentageNumerator:   20,
		PercentageDenominator: 100,
		StartDate:             now.Add(-1 * time.Hour),
		EndDate:               now.Add(24 * time.Hour),
	})
	require.NoError(t, err)

	// Test: Remove discount
	err = removeDiscountUsecase.Execute(testCtx, removediscount.Request{
		ProductID: productID,
	})
	require.NoError(t, err)

	// Verify: Effective price is back to base price
	product, err := getQuery.Execute(testCtx, getproduct.Request{
		ProductID: productID,
		Now:        now,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(10000), product.EffectivePriceNumerator)
	assert.Equal(t, int64(100), product.EffectivePriceDenominator)

	// Verify: Discount removed event was created
	events := getOutboxEvents(t, productID)
	hasRemoved := false
	for _, e := range events {
		if e.EventType == "discount.removed" {
			hasRemoved = true
			break
		}
	}
	assert.True(t, hasRemoved, "discount.removed event should exist")
}
