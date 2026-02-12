package unit

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/app/product/domain/services"
)

func TestMoneyCalculations(t *testing.T) {
	t.Run("Create money from fraction", func(t *testing.T) {
		money, err := domain.NewMoneyFromFraction(1999, 100)
		require.NoError(t, err)
		assert.NotNil(t, money)

		num, den := money.Fraction()
		assert.Equal(t, int64(1999), num)
		assert.Equal(t, int64(100), den)
	})

	t.Run("Invalid denominator", func(t *testing.T) {
		_, err := domain.NewMoneyFromFraction(100, 0)
		require.Error(t, err)
	})

	t.Run("Multiply money by ratio", func(t *testing.T) {
		money, _ := domain.NewMoneyFromFraction(10000, 100) // $100.00
		discount := big.NewRat(20, 100)                      // 20%

		result := money.MultiplyBy(discount)
		assert.NotNil(t, result)

		num, den := result.Fraction()
		assert.Equal(t, int64(2000), num)
		assert.Equal(t, int64(100), den) // $20.00
	})

	t.Run("Subtract money", func(t *testing.T) {
		money1, _ := domain.NewMoneyFromFraction(10000, 100) // $100.00
		money2, _ := domain.NewMoneyFromFraction(2000, 100)  // $20.00

		result := money1.Subtract(money2)
		assert.NotNil(t, result)

		num, den := result.Fraction()
		assert.Equal(t, int64(8000), num)
		assert.Equal(t, int64(100), den) // $80.00
	})

	t.Run("Compare money", func(t *testing.T) {
		money1, _ := domain.NewMoneyFromFraction(10000, 100)
		money2, _ := domain.NewMoneyFromFraction(8000, 100)
		money3, _ := domain.NewMoneyFromFraction(10000, 100)

		assert.Equal(t, 1, money1.Compare(money2))  // money1 > money2
		assert.Equal(t, -1, money2.Compare(money1)) // money2 < money1
		assert.Equal(t, 0, money1.Compare(money3))   // money1 == money3
	})
}

func TestDiscountValidation(t *testing.T) {
	t.Run("Valid discount", func(t *testing.T) {
		percentage := big.NewRat(20, 100) // 20%
		start := time.Now()
		end := start.Add(24 * time.Hour)

		discount, err := domain.NewDiscount(percentage, start, end)
		require.NoError(t, err)
		assert.NotNil(t, discount)
		assert.Equal(t, percentage, discount.Percentage())
		assert.Equal(t, start, discount.StartAt())
		assert.Equal(t, end, discount.EndAt())
	})

	t.Run("Invalid percentage - negative", func(t *testing.T) {
		percentage := big.NewRat(-10, 100)
		start := time.Now()
		end := start.Add(24 * time.Hour)

		_, err := domain.NewDiscount(percentage, start, end)
		require.Error(t, err)
	})

	t.Run("Invalid percentage - over 100%", func(t *testing.T) {
		percentage := big.NewRat(150, 100) // 150%
		start := time.Now()
		end := start.Add(24 * time.Hour)

		_, err := domain.NewDiscount(percentage, start, end)
		require.Error(t, err)
	})

	t.Run("Invalid period - end before start", func(t *testing.T) {
		percentage := big.NewRat(20, 100)
		start := time.Now()
		end := start.Add(-24 * time.Hour) // end before start

		_, err := domain.NewDiscount(percentage, start, end)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDiscountPeriod)
	})

	t.Run("IsValidAt - valid time", func(t *testing.T) {
		percentage := big.NewRat(20, 100)
		start := time.Now().Add(-1 * time.Hour)
		end := time.Now().Add(24 * time.Hour)

		discount, _ := domain.NewDiscount(percentage, start, end)
		assert.True(t, discount.IsValidAt(time.Now()))
	})

	t.Run("IsValidAt - before start", func(t *testing.T) {
		percentage := big.NewRat(20, 100)
		start := time.Now().Add(1 * time.Hour)
		end := time.Now().Add(24 * time.Hour)

		discount, _ := domain.NewDiscount(percentage, start, end)
		assert.False(t, discount.IsValidAt(time.Now()))
	})

	t.Run("IsValidAt - after end", func(t *testing.T) {
		percentage := big.NewRat(20, 100)
		start := time.Now().Add(-24 * time.Hour)
		end := time.Now().Add(-1 * time.Hour)

		discount, _ := domain.NewDiscount(percentage, start, end)
		assert.False(t, discount.IsValidAt(time.Now()))
	})
}

func TestPricingCalculator(t *testing.T) {
	calculator := services.PricingCalculator{}

	t.Run("Effective price without discount", func(t *testing.T) {
		basePrice, _ := domain.NewMoneyFromFraction(10000, 100) // $100.00
		product := domain.RehydrateProduct(
			"test-id",
			"Test",
			"Test",
			"test",
			basePrice,
			nil, // no discount
			domain.ProductStatusActive,
			nil,
			time.Now(),
			time.Now(),
		)

		effective := calculator.EffectivePrice(product, time.Now())
		assert.NotNil(t, effective)

		num, den := effective.Fraction()
		assert.Equal(t, int64(10000), num)
		assert.Equal(t, int64(100), den) // Same as base price
	})

	t.Run("Effective price with valid discount", func(t *testing.T) {
		basePrice, _ := domain.NewMoneyFromFraction(10000, 100) // $100.00
		discount, _ := domain.NewDiscount(
			big.NewRat(20, 100), // 20%
			time.Now().Add(-1*time.Hour),
			time.Now().Add(24*time.Hour),
		)

		product := domain.RehydrateProduct(
			"test-id",
			"Test",
			"Test",
			"test",
			basePrice,
			discount,
			domain.ProductStatusActive,
			nil,
			time.Now(),
			time.Now(),
		)

		effective := calculator.EffectivePrice(product, time.Now())
		assert.NotNil(t, effective)

		num, den := effective.Fraction()
		assert.Equal(t, int64(8000), num)
		assert.Equal(t, int64(100), den) // $80.00 (20% off)
	})

	t.Run("Effective price with expired discount", func(t *testing.T) {
		basePrice, _ := domain.NewMoneyFromFraction(10000, 100) // $100.00
		discount, _ := domain.NewDiscount(
			big.NewRat(20, 100), // 20%
			time.Now().Add(-24*time.Hour),
			time.Now().Add(-1*time.Hour), // expired
		)

		product := domain.RehydrateProduct(
			"test-id",
			"Test",
			"Test",
			"test",
			basePrice,
			discount,
			domain.ProductStatusActive,
			nil,
			time.Now(),
			time.Now(),
		)

		effective := calculator.EffectivePrice(product, time.Now())
		assert.NotNil(t, effective)

		// Should return base price since discount is expired
		num, den := effective.Fraction()
		assert.Equal(t, int64(10000), num)
		assert.Equal(t, int64(100), den)
	})

	t.Run("Precise decimal calculation", func(t *testing.T) {
		// Test with non-round numbers
		basePrice, _ := domain.NewMoneyFromFraction(9999, 100) // $99.99
		discount, _ := domain.NewDiscount(
			big.NewRat(15, 100), // 15%
			time.Now().Add(-1*time.Hour),
			time.Now().Add(24*time.Hour),
		)

		product := domain.RehydrateProduct(
			"test-id",
			"Test",
			"Test",
			"test",
			basePrice,
			discount,
			domain.ProductStatusActive,
			nil,
			time.Now(),
			time.Now(),
		)

		effective := calculator.EffectivePrice(product, time.Now())
		assert.NotNil(t, effective)

		// $99.99 * 0.85 = $84.9915
		// Using big.Rat preserves precision
		num, den := effective.Fraction()
		expected := big.NewRat(9999, 100)
		expected.Mul(expected, big.NewRat(85, 100))
		expectedNum, expectedDen := expected.Num().Int64(), expected.Denom().Int64()

		assert.Equal(t, expectedNum, num)
		assert.Equal(t, expectedDen, den)
	})
}

func TestStateMachineTransitions(t *testing.T) {
	t.Run("Product starts as inactive", func(t *testing.T) {
		basePrice, _ := domain.NewMoneyFromFraction(1000, 100)
		product := domain.NewProduct(
			"test-id",
			"Test",
			"Test",
			"test",
			basePrice,
			time.Now(),
		)

		assert.Equal(t, domain.ProductStatusInactive, product.Status())
	})

	t.Run("Activate inactive product", func(t *testing.T) {
		basePrice, _ := domain.NewMoneyFromFraction(1000, 100)
		product := domain.NewProduct(
			"test-id",
			"Test",
			"Test",
			"test",
			basePrice,
			time.Now(),
		)

		product.Activate(time.Now())
		assert.Equal(t, domain.ProductStatusActive, product.Status())
		assert.True(t, product.Changes().Dirty(domain.FieldStatus))
	})

	t.Run("Deactivate active product", func(t *testing.T) {
		basePrice, _ := domain.NewMoneyFromFraction(1000, 100)
		product := domain.NewProduct(
			"test-id",
			"Test",
			"Test",
			"test",
			basePrice,
			time.Now(),
		)

		product.Activate(time.Now())
		product.Deactivate(time.Now())
		assert.Equal(t, domain.ProductStatusInactive, product.Status())
	})

	t.Run("Archive product", func(t *testing.T) {
		basePrice, _ := domain.NewMoneyFromFraction(1000, 100)
		product := domain.NewProduct(
			"test-id",
			"Test",
			"Test",
			"test",
			basePrice,
			time.Now(),
		)

		now := time.Now()
		product.Archive(now)
		assert.Equal(t, domain.ProductStatusArchived, product.Status())
		assert.NotNil(t, product.ArchivedAt())
		assert.Equal(t, now, *product.ArchivedAt())
	})

	t.Run("Cannot activate archived product", func(t *testing.T) {
		basePrice, _ := domain.NewMoneyFromFraction(1000, 100)
		product := domain.NewProduct(
			"test-id",
			"Test",
			"Test",
			"test",
			basePrice,
			time.Now(),
		)

		product.Archive(time.Now())
		originalStatus := product.Status()
		product.Activate(time.Now())
		// Status should not change
		assert.Equal(t, originalStatus, product.Status())
	})

	t.Run("Cannot apply discount to inactive product", func(t *testing.T) {
		basePrice, _ := domain.NewMoneyFromFraction(1000, 100)
		product := domain.NewProduct(
			"test-id",
			"Test",
			"Test",
			"test",
			basePrice,
			time.Now(),
		)

		discount, _ := domain.NewDiscount(
			big.NewRat(20, 100),
			time.Now(),
			time.Now().Add(24*time.Hour),
		)

		err := product.ApplyDiscount(discount, time.Now())
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrProductNotActive)
	})

	t.Run("Can apply discount to active product", func(t *testing.T) {
		basePrice, _ := domain.NewMoneyFromFraction(1000, 100)
		product := domain.NewProduct(
			"test-id",
			"Test",
			"Test",
			"test",
			basePrice,
			time.Now(),
		)

		product.Activate(time.Now())
		discount, _ := domain.NewDiscount(
			big.NewRat(20, 100),
			time.Now(),
			time.Now().Add(24*time.Hour),
		)

		err := product.ApplyDiscount(discount, time.Now())
		assert.NoError(t, err)
		assert.NotNil(t, product.Discount())
		assert.True(t, product.Changes().Dirty(domain.FieldDiscount))
	})
}

func TestChangeTracking(t *testing.T) {
	t.Run("Track field changes", func(t *testing.T) {
		basePrice, _ := domain.NewMoneyFromFraction(1000, 100)
		product := domain.NewProduct(
			"test-id",
			"Original",
			"Original",
			"original",
			basePrice,
			time.Now(),
		)

		// Clear initial dirty flags
		product.Changes().Clear()

		product.UpdateDetails("Updated", "Updated", "updated", time.Now())
		assert.True(t, product.Changes().Dirty(domain.FieldName))
		assert.True(t, product.Changes().Dirty(domain.FieldDescription))
		assert.True(t, product.Changes().Dirty(domain.FieldCategory))
	})

	t.Run("No changes if values unchanged", func(t *testing.T) {
		basePrice, _ := domain.NewMoneyFromFraction(1000, 100)
		product := domain.NewProduct(
			"test-id",
			"Test",
			"Test",
			"test",
			basePrice,
			time.Now(),
		)

		product.Changes().Clear()

		// Update with same values
		product.UpdateDetails("Test", "Test", "test", time.Now())
		// Should still mark as dirty if called, but in practice
		// the implementation checks for changes
	})
}

func TestDomainEvents(t *testing.T) {
	t.Run("Product creation emits event", func(t *testing.T) {
		basePrice, _ := domain.NewMoneyFromFraction(1000, 100)
		product := domain.NewProduct(
			"test-id",
			"Test",
			"Test",
			"test",
			basePrice,
			time.Now(),
		)

		events := product.DomainEvents()
		require.Len(t, events, 1)
		assert.IsType(t, domain.ProductCreatedEvent{}, events[0])
	})

	t.Run("Update emits event", func(t *testing.T) {
		basePrice, _ := domain.NewMoneyFromFraction(1000, 100)
		product := domain.NewProduct(
			"test-id",
			"Test",
			"Test",
			"test",
			basePrice,
			time.Now(),
		)

		product.ClearDomainEvents()
		product.UpdateDetails("Updated", "", "", time.Now())

		events := product.DomainEvents()
		require.Len(t, events, 1)
		assert.IsType(t, domain.ProductUpdatedEvent{}, events[0])
	})

	t.Run("Clear domain events", func(t *testing.T) {
		basePrice, _ := domain.NewMoneyFromFraction(1000, 100)
		product := domain.NewProduct(
			"test-id",
			"Test",
			"Test",
			"test",
			basePrice,
			time.Now(),
		)

		require.Len(t, product.DomainEvents(), 1)
		product.ClearDomainEvents()
		assert.Len(t, product.DomainEvents(), 0)
	})
}
