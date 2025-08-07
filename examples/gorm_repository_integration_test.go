package examples

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dchlong/querybuilder/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestGormRepositoryIntegration demonstrates full integration with generated Product querybuilder
func TestGormRepositoryIntegration(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)

	// Create GORM repository using generated types
	repo := repository.NewGormRepository[Product, *ProductFilters, *ProductUpdater](db)

	ctx := context.Background()

	t.Run("create products with complex data", func(t *testing.T) {
		products := createTestProducts()

		err := repo.Create(ctx, products...)
		assert.NoError(t, err)

		// Verify all products have IDs
		for _, product := range products {
			assert.NotZero(t, product.ID)
		}
	})

	t.Run("find products using generated filters", func(t *testing.T) {
		// Test basic equality filter
		filter := NewProductFilters().IsActiveEq(true)
		products, err := repo.FindAll(ctx, filter)

		assert.NoError(t, err)
		assert.Greater(t, len(products), 0)

		for _, product := range products {
			assert.True(t, product.IsActive)
		}
	})

	t.Run("complex filtering with multiple conditions", func(t *testing.T) {
		// Test chained filters
		filter := NewProductFilters().
			IsActiveEq(true).
			PriceGte(10.0).
			PriceLte(100.0).
			StockGt(0)

		products, err := repo.FindAll(ctx, filter)
		assert.NoError(t, err)

		for _, product := range products {
			assert.True(t, product.IsActive)
			assert.GreaterOrEqual(t, product.Price, 10.0)
			assert.LessOrEqual(t, product.Price, 100.0)
			assert.Greater(t, product.Stock, 0)
		}
	})

	t.Run("find products with IN filter", func(t *testing.T) {
		// Test IN filter for categories
		filter := NewProductFilters().CategoryIDIn(1, 2)
		products, err := repo.FindAll(ctx, filter)

		assert.NoError(t, err)
		for _, product := range products {
			assert.Contains(t, []int64{1, 2}, product.CategoryID)
		}
	})

	t.Run("find products with LIKE filter", func(t *testing.T) {
		// Test LIKE filter for names
		filter := NewProductFilters().NameLike("%Widget%")
		products, err := repo.FindAll(ctx, filter)

		assert.NoError(t, err)
		for _, product := range products {
			assert.Contains(t, product.Name, "Widget")
		}
	})

	t.Run("update products using generated updater", func(t *testing.T) {
		// Find a product to update
		filter := NewProductFilters().NameEq("Awesome Widget")
		product, found, err := repo.FindOne(ctx, filter)
		require.NoError(t, err)
		require.True(t, found)

		// Update using generated updater
		updater := NewProductUpdater().
			SetPrice(29.99).
			SetStock(75).
			SetIsActive(true)

		err = repo.Update(ctx, product, updater)
		assert.NoError(t, err)

		// Verify update
		updated, found, err := repo.FindOneByID(ctx, product.ID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, 29.99, updated.Price)
		assert.Equal(t, 75, updated.Stock)
		assert.True(t, updated.IsActive)
	})

	t.Run("batch update with filter", func(t *testing.T) {
		// Update all products in category 1
		filter := NewProductFilters().CategoryIDEq(1)
		updater := NewProductUpdater().SetIsActive(false)

		rowsAffected, err := repo.UpdateWithFilter(ctx, filter, updater)
		assert.NoError(t, err)
		assert.Greater(t, rowsAffected, int64(0))

		// Verify updates
		products, err := repo.FindAll(ctx, NewProductFilters().CategoryIDEq(1))
		assert.NoError(t, err)

		for _, product := range products {
			assert.False(t, product.IsActive)
		}
	})

	t.Run("count and exists operations", func(t *testing.T) {
		// Count total products
		totalCount, err := repo.Count(ctx, NewProductFilters())
		assert.NoError(t, err)
		assert.Greater(t, totalCount, int64(0))

		// Count active products
		activeCount, err := repo.Count(ctx, NewProductFilters().IsActiveEq(true))
		assert.NoError(t, err)
		assert.LessOrEqual(t, activeCount, totalCount)

		// Check if expensive products exist
		hasExpensive, err := repo.Exists(ctx, NewProductFilters().PriceGt(100.0))
		assert.NoError(t, err)
		assert.False(t, hasExpensive) // Based on our test data

		// Check if products exist in category
		hasInCategory, err := repo.Exists(ctx, NewProductFilters().CategoryIDEq(2))
		assert.NoError(t, err)
		assert.True(t, hasInCategory)
	})

	t.Run("pagination with generated options", func(t *testing.T) {
		// Test pagination
		filter := NewProductFilters().IsActiveEq(true)

		// Get first page
		page1, err := repo.FindAll(ctx, filter,
			repository.WithLimit(2),
			repository.WithOffset(0),
		)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(page1), 2)

		// Get second page
		page2, err := repo.FindAll(ctx, filter,
			repository.WithLimit(2),
			repository.WithOffset(2),
		)
		assert.NoError(t, err)

		// Verify pages are different (if we have enough data)
		if len(page1) > 0 && len(page2) > 0 {
			assert.NotEqual(t, page1[0].ID, page2[0].ID)
		}
	})

	t.Run("transaction operations", func(t *testing.T) {
		originalCount, err := repo.Count(ctx, NewProductFilters())
		require.NoError(t, err)

		// Test successful transaction
		newProduct := &Product{
			Name:       "Transaction Test Product",
			SKU:        "TTP-001",
			Price:      50.0,
			Stock:      10,
			CategoryID: 1,
			IsActive:   true,
			CreatedAt:  time.Now(),
		}

		err = repo.WithTransaction(ctx, func(txRepo *repository.GormRepository[Product, *ProductFilters, *ProductUpdater]) error {
			if err := txRepo.Create(ctx, newProduct); err != nil {
				return err
			}

			// Update the product within transaction
			updater := NewProductUpdater().SetStock(20)
			return txRepo.Update(ctx, newProduct, updater)
		})

		assert.NoError(t, err)

		// Verify transaction committed
		newCount, err := repo.Count(ctx, NewProductFilters())
		assert.NoError(t, err)
		assert.Equal(t, originalCount+1, newCount)

		// Verify the product was updated
		found, exists, err := repo.FindOne(ctx, NewProductFilters().NameEq("Transaction Test Product"))
		require.NoError(t, err)
		require.True(t, exists)
		assert.Equal(t, 20, found.Stock)
	})

	t.Run("batch creation", func(t *testing.T) {
		// Create products in batches
		batchProducts := make([]*Product, 25)
		for i := range batchProducts {
			batchProducts[i] = &Product{
				Name:       fmt.Sprintf("Batch Product %d", i),
				SKU:        fmt.Sprintf("BP-%03d", i),
				Price:      float64(10 + i),
				Stock:      i + 1,
				CategoryID: int64((i % 3) + 1),
				IsActive:   i%2 == 0,
				CreatedAt:  time.Now(),
			}
		}

		err := repo.CreateInBatches(ctx, 10, batchProducts...)
		assert.NoError(t, err)

		// Verify all were created
		count, err := repo.Count(ctx, NewProductFilters().NameLike("Batch Product%"))
		assert.NoError(t, err)
		assert.Equal(t, int64(25), count)
	})

	t.Run("delete with filter", func(t *testing.T) {
		// Delete all batch products
		rowsAffected, err := repo.DeleteWithFilter(ctx, NewProductFilters().NameLike("Batch Product%"))
		assert.NoError(t, err)
		assert.Equal(t, int64(25), rowsAffected)

		// Verify deletion
		count, err := repo.Count(ctx, NewProductFilters().NameLike("Batch Product%"))
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("JSON attributes operations", func(t *testing.T) {
		// Create a product with complex attributes
		attrs := &Attributes{
			Color:      "red",
			Size:       "large",
			Weight:     2.5,
			Dimensions: "15x10x5",
		}

		product := &Product{
			Name:       "JSON Test Product",
			SKU:        "JTP-001",
			Price:      75.00,
			Stock:      30,
			CategoryID: 1,
			IsActive:   true,
			Attributes: datatypes.NewJSONType(attrs),
			CreatedAt:  time.Now(),
		}

		err := repo.Create(ctx, product)
		assert.NoError(t, err)
		assert.NotZero(t, product.ID)

		// Retrieve and verify the JSON attributes
		found, exists, err := repo.FindOneByID(ctx, product.ID)
		require.NoError(t, err)
		require.True(t, exists)

		// Verify the attributes were stored and retrieved correctly
		retrievedAttrs := found.Attributes.Data()
		require.NotNil(t, retrievedAttrs)
		assert.Equal(t, "red", retrievedAttrs.Color)
		assert.Equal(t, "large", retrievedAttrs.Size)
		assert.Equal(t, 2.5, retrievedAttrs.Weight)
		assert.Equal(t, "15x10x5", retrievedAttrs.Dimensions)

		// Update attributes using the generated updater
		newAttrs := &Attributes{
			Color:      "blue",
			Size:       "medium",
			Weight:     1.8,
			Dimensions: "12x8x4",
		}

		updater := NewProductUpdater().
			SetAttributes(datatypes.NewJSONType(newAttrs)).
			SetPrice(65.00)

		err = repo.Update(ctx, found, updater)
		assert.NoError(t, err)

		// Verify the update
		updated, exists, err := repo.FindOneByID(ctx, product.ID)
		require.NoError(t, err)
		require.True(t, exists)

		updatedAttrs := updated.Attributes.Data()
		require.NotNil(t, updatedAttrs)
		assert.Equal(t, "blue", updatedAttrs.Color)
		assert.Equal(t, "medium", updatedAttrs.Size)
		assert.Equal(t, 1.8, updatedAttrs.Weight)
		assert.Equal(t, "12x8x4", updatedAttrs.Dimensions)
		assert.Equal(t, 65.00, updated.Price)
	})

	t.Run("health check", func(t *testing.T) {
		err := repo.Health(ctx)
		assert.NoError(t, err)
	})
}

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto migrate the Product schema
	err = db.AutoMigrate(&Product{})
	require.NoError(t, err)

	return db
}

// createTestProducts creates sample products for testing
func createTestProducts() []*Product {
	now := time.Now()
	desc1 := "An amazing widget for all your needs"
	desc2 := "Professional gadget for business use"

	return []*Product{
		{
			ID:          0,
			Name:        "Awesome Widget",
			SKU:         "AWG-001",
			Description: &desc1,
			Price:       19.99,
			Stock:       100,
			CategoryID:  1,
			IsActive:    true,
			Tags:        []string{"widget", "awesome", "useful"},
			Attributes: datatypes.NewJSONType(&Attributes{
				Color:      "blue",
				Size:       "medium",
				Weight:     1.5,
				Dimensions: "10x5x2",
			}),
			CreatedAt: now,
			UpdatedAt: nil,
		},
		{
			Name:        "Super Gadget",
			SKU:         "SGD-002",
			Description: &desc2,
			Price:       49.99,
			Stock:       50,
			CategoryID:  2,
			IsActive:    true,
			Tags:        []string{"gadget", "professional", "business"},
			CreatedAt:   now,
		},
		{
			Name:       "Basic Tool",
			SKU:        "BTL-003",
			Price:      9.99,
			Stock:      200,
			CategoryID: 1,
			IsActive:   false,
			Tags:       []string{"tool", "basic"},
			CreatedAt:  now,
		},
		{
			Name:       "Premium Device",
			SKU:        "PRM-004",
			Price:      99.99,
			Stock:      25,
			CategoryID: 3,
			IsActive:   true,
			Tags:       []string{"premium", "device", "luxury"},
			CreatedAt:  now,
		},
	}
}
