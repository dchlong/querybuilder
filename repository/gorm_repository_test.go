package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestEntity represents a test entity for the repository tests
type TestEntity struct {
	ID        int64     `gorm:"primaryKey" db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Age       int       `db:"age"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// TestFilter implements EntityFilter for testing
type TestFilter struct {
	filters []*Filter
}

func (f *TestFilter) ListFilters() []*Filter {
	return f.filters
}

func (f *TestFilter) NameEq(name string) *TestFilter {
	f.filters = append(f.filters, &Filter{
		Field:    "name",
		Operator: OperatorEqual,
		Value:    name,
	})
	return f
}

func (f *TestFilter) EmailLike(pattern string) *TestFilter {
	f.filters = append(f.filters, &Filter{
		Field:    "email",
		Operator: OperatorLike,
		Value:    pattern,
	})
	return f
}

func (f *TestFilter) AgeGte(age int) *TestFilter {
	f.filters = append(f.filters, &Filter{
		Field:    "age",
		Operator: OperatorGreaterThanOrEqual,
		Value:    age,
	})
	return f
}

func (f *TestFilter) IsActiveEq(isActive bool) *TestFilter {
	f.filters = append(f.filters, &Filter{
		Field:    "is_active",
		Operator: OperatorEqual,
		Value:    isActive,
	})
	return f
}

// TestUpdater implements EntityUpdater for testing
type TestUpdater struct {
	fields map[string]interface{}
}

func (u *TestUpdater) GetChangeSet() map[string]interface{} {
	return u.fields
}

func (u *TestUpdater) SetName(name string) *TestUpdater {
	u.fields["name"] = name
	return u
}

func (u *TestUpdater) SetEmail(email string) *TestUpdater {
	u.fields["email"] = email
	return u
}

func (u *TestUpdater) SetAge(age int) *TestUpdater {
	u.fields["age"] = age
	return u
}

func (u *TestUpdater) SetIsActive(isActive bool) *TestUpdater {
	u.fields["is_active"] = isActive
	return u
}

// Helper functions for tests

func NewTestFilter() *TestFilter {
	return &TestFilter{
		filters: make([]*Filter, 0),
	}
}

func NewTestUpdater() *TestUpdater {
	return &TestUpdater{
		fields: make(map[string]interface{}),
	}
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto migrate the test entity
	err = db.AutoMigrate(&TestEntity{})
	require.NoError(t, err)

	return db
}

func setupTestRepository(t *testing.T) (*GormRepository[TestEntity, *TestFilter, *TestUpdater], *gorm.DB) {
	db := setupTestDB(t)
	repo := NewGormRepository[TestEntity, *TestFilter, *TestUpdater](db)
	return repo, db
}

func createTestEntities() []*TestEntity {
	return []*TestEntity{
		{
			Name:     "Alice",
			Email:    "alice@example.com",
			Age:      25,
			IsActive: true,
		},
		{
			Name:     "Bob",
			Email:    "bob@example.com",
			Age:      30,
			IsActive: true,
		},
		{
			Name:     "Charlie",
			Email:    "charlie@example.com",
			Age:      20,
			IsActive: false,
		},
		{
			Name:     "David",
			Email:    "david@example.com",
			Age:      35,
			IsActive: true,
		},
	}
}

// Test Cases

func TestGormRepository_Create(t *testing.T) {
	repo, _ := setupTestRepository(t)
	ctx := context.Background()

	t.Run("create single record", func(t *testing.T) {
		entity := &TestEntity{
			Name:     "Test Product",
			Email:    "test@example.com",
			Age:      25,
			IsActive: true,
		}

		err := repo.Create(ctx, entity)
		assert.NoError(t, err)
		assert.NotZero(t, entity.ID)
	})

	t.Run("create multiple records", func(t *testing.T) {
		entities := createTestEntities()

		err := repo.Create(ctx, entities...)
		assert.NoError(t, err)

		for _, entity := range entities {
			assert.NotZero(t, entity.ID)
		}
	})

	t.Run("create no records should return error", func(t *testing.T) {
		err := repo.Create(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no records provided")
	})
}

func TestGormRepository_FindOneByID(t *testing.T) {
	repo, _ := setupTestRepository(t)
	ctx := context.Background()

	// Create test data
	entity := &TestEntity{
		Name:     "Test Product",
		Email:    "test@example.com",
		Age:      25,
		IsActive: true,
	}
	err := repo.Create(ctx, entity)
	require.NoError(t, err)

	t.Run("find existing record", func(t *testing.T) {
		found, exists, err := repo.FindOneByID(ctx, entity.ID)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, entity.Name, found.Name)
		assert.Equal(t, entity.Email, found.Email)
	})

	t.Run("find non-existing record", func(t *testing.T) {
		found, exists, err := repo.FindOneByID(ctx, 99999)
		assert.NoError(t, err)
		assert.False(t, exists)
		assert.Nil(t, found)
	})
}

func TestGormRepository_FindOne(t *testing.T) {
	repo, _ := setupTestRepository(t)
	ctx := context.Background()

	// Create test data
	entities := createTestEntities()
	err := repo.Create(ctx, entities...)
	require.NoError(t, err)

	t.Run("find with single filter", func(t *testing.T) {
		filter := NewTestFilter().NameEq("Alice")
		found, exists, err := repo.FindOne(ctx, filter)

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, "Alice", found.Name)
	})

	t.Run("find with multiple filters", func(t *testing.T) {
		filter := NewTestFilter().IsActiveEq(true).AgeGte(30)
		found, exists, err := repo.FindOne(ctx, filter)

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.True(t, found.Age >= 30)
		assert.True(t, found.IsActive)
	})

	t.Run("find with no matches", func(t *testing.T) {
		filter := NewTestFilter().NameEq("NonExistent")
		found, exists, err := repo.FindOne(ctx, filter)

		assert.NoError(t, err)
		assert.False(t, exists)
		assert.Nil(t, found)
	})
}

func TestGormRepository_FindAll(t *testing.T) {
	repo, _ := setupTestRepository(t)
	ctx := context.Background()

	// Create test data
	entities := createTestEntities()
	err := repo.Create(ctx, entities...)
	require.NoError(t, err)

	t.Run("find all active users", func(t *testing.T) {
		filter := NewTestFilter().IsActiveEq(true)
		found, err := repo.FindAll(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, found, 3) // Alice, Bob, David

		for _, entity := range found {
			assert.True(t, entity.IsActive)
		}
	})

	t.Run("find with age filter", func(t *testing.T) {
		filter := NewTestFilter().AgeGte(25)
		found, err := repo.FindAll(ctx, filter)

		assert.NoError(t, err)
		assert.Len(t, found, 3) // Alice, Bob, David

		for _, entity := range found {
			assert.GreaterOrEqual(t, entity.Age, 25)
		}
	})

	t.Run("find with limit", func(t *testing.T) {
		filter := NewTestFilter().IsActiveEq(true)
		found, err := repo.FindAll(ctx, filter, WithLimit(2))

		assert.NoError(t, err)
		assert.Len(t, found, 2)
	})
}

func TestGormRepository_Update(t *testing.T) {
	repo, _ := setupTestRepository(t)
	ctx := context.Background()

	// Create test data
	entity := &TestEntity{
		Name:     "Test Product",
		Email:    "test@example.com",
		Age:      25,
		IsActive: true,
	}
	err := repo.Create(ctx, entity)
	require.NoError(t, err)

	t.Run("update single field", func(t *testing.T) {
		updater := NewTestUpdater().SetName("Updated Name")
		err := repo.Update(ctx, entity, updater)

		assert.NoError(t, err)

		// Verify update
		found, exists, err := repo.FindOneByID(ctx, entity.ID)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, "Updated Name", found.Name)
		assert.Equal(t, entity.Email, found.Email) // Should remain unchanged
	})

	t.Run("update multiple fields", func(t *testing.T) {
		updater := NewTestUpdater().SetName("Another Name").SetAge(30)
		err := repo.Update(ctx, entity, updater)

		assert.NoError(t, err)

		// Verify update
		found, exists, err := repo.FindOneByID(ctx, entity.ID)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, "Another Name", found.Name)
		assert.Equal(t, 30, found.Age)
	})

	t.Run("update with empty changeset should do nothing", func(t *testing.T) {
		updater := NewTestUpdater()
		err := repo.Update(ctx, entity, updater)

		assert.NoError(t, err)
	})
}

func TestGormRepository_Count(t *testing.T) {
	repo, _ := setupTestRepository(t)
	ctx := context.Background()

	// Create test data
	entities := createTestEntities()
	err := repo.Create(ctx, entities...)
	require.NoError(t, err)

	t.Run("count all records", func(t *testing.T) {
		filter := NewTestFilter()
		count, err := repo.Count(ctx, filter)

		assert.NoError(t, err)
		assert.Equal(t, int64(4), count)
	})

	t.Run("count with filter", func(t *testing.T) {
		filter := NewTestFilter().IsActiveEq(true)
		count, err := repo.Count(ctx, filter)

		assert.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})
}

func TestGormRepository_Exists(t *testing.T) {
	repo, _ := setupTestRepository(t)
	ctx := context.Background()

	// Create test data
	entities := createTestEntities()
	err := repo.Create(ctx, entities...)
	require.NoError(t, err)

	t.Run("exists with matching filter", func(t *testing.T) {
		filter := NewTestFilter().NameEq("Alice")
		exists, err := repo.Exists(ctx, filter)

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("exists with non-matching filter", func(t *testing.T) {
		filter := NewTestFilter().NameEq("NonExistent")
		exists, err := repo.Exists(ctx, filter)

		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestGormRepository_UpdateWithFilter(t *testing.T) {
	repo, _ := setupTestRepository(t)
	ctx := context.Background()

	// Create test data
	entities := createTestEntities()
	err := repo.Create(ctx, entities...)
	require.NoError(t, err)

	t.Run("update multiple records with filter", func(t *testing.T) {
		filter := NewTestFilter().IsActiveEq(true)
		updater := NewTestUpdater().SetEmail("updated@example.com")

		rowsAffected, err := repo.UpdateWithFilter(ctx, filter, updater)

		assert.NoError(t, err)
		assert.Equal(t, int64(3), rowsAffected)

		// Verify updates
		activeProducts, err := repo.FindAll(ctx, NewTestFilter().IsActiveEq(true))
		assert.NoError(t, err)

		for _, user := range activeProducts {
			assert.Equal(t, "updated@example.com", user.Email)
		}
	})
}

func TestGormRepository_DeleteWithFilter(t *testing.T) {
	repo, _ := setupTestRepository(t)
	ctx := context.Background()

	// Create test data
	entities := createTestEntities()
	err := repo.Create(ctx, entities...)
	require.NoError(t, err)

	t.Run("delete records with filter", func(t *testing.T) {
		filter := NewTestFilter().IsActiveEq(false)
		rowsAffected, err := repo.DeleteWithFilter(ctx, filter)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected) // Only Charlie is inactive

		// Verify deletion
		totalCount, err := repo.Count(ctx, NewTestFilter())
		assert.NoError(t, err)
		assert.Equal(t, int64(3), totalCount) // 3 remaining active users
	})
}

func TestGormRepository_CreateInBatches(t *testing.T) {
	repo, _ := setupTestRepository(t)
	ctx := context.Background()

	t.Run("create in batches", func(t *testing.T) {
		entities := make([]*TestEntity, 250) // More than default batch size
		for i := range entities {
			entities[i] = &TestEntity{
				Name:     fmt.Sprintf("Product %d", i),
				Email:    fmt.Sprintf("product%d@example.com", i),
				Age:      20 + (i % 50),
				IsActive: i%2 == 0,
			}
		}

		err := repo.CreateInBatches(ctx, 50, entities...)
		assert.NoError(t, err)

		// Verify all records were created
		count, err := repo.Count(ctx, NewTestFilter())
		assert.NoError(t, err)
		assert.Equal(t, int64(250), count)
	})
}

func TestGormRepository_WithTransaction(t *testing.T) {
	repo, _ := setupTestRepository(t)
	ctx := context.Background()

	t.Run("successful transaction", func(t *testing.T) {
		entity1 := &TestEntity{Name: "User1", Email: "user1@example.com", Age: 25, IsActive: true}
		entity2 := &TestEntity{Name: "User2", Email: "user2@example.com", Age: 30, IsActive: true}

		err := repo.WithTransaction(ctx, func(txRepo *GormRepository[TestEntity, *TestFilter, *TestUpdater]) error {
			if err := txRepo.Create(ctx, entity1); err != nil {
				return err
			}
			return txRepo.Create(ctx, entity2)
		})

		assert.NoError(t, err)

		// Verify both records were created
		count, err := repo.Count(ctx, NewTestFilter())
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("failed transaction should rollback", func(t *testing.T) {
		entity1 := &TestEntity{Name: "User3", Email: "user3@example.com", Age: 25, IsActive: true}

		err := repo.WithTransaction(ctx, func(txRepo *GormRepository[TestEntity, *TestFilter, *TestUpdater]) error {
			if err := txRepo.Create(ctx, entity1); err != nil {
				return err
			}
			// Simulate an error
			return errors.New("simulated error")
		})

		assert.Error(t, err)

		// Verify no new records were created (still 2 from previous test)
		count, err := repo.Count(ctx, NewTestFilter())
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})
}

func TestGormRepository_Health(t *testing.T) {
	repo, db := setupTestRepository(t)
	ctx := context.Background()

	t.Run("healthy connection", func(t *testing.T) {
		err := repo.Health(ctx)
		assert.NoError(t, err)
	})

	t.Run("unhealthy connection", func(t *testing.T) {
		// Close the database connection
		sqlDB, err := db.DB()
		require.NoError(t, err)
		_ = sqlDB.Close() // Ignore error in test cleanup

		err = repo.Health(ctx)
		assert.Error(t, err)
	})
}

// Benchmark tests

func BenchmarkGormRepository_Create(b *testing.B) {
	repo, _ := setupTestRepository(&testing.T{})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entity := &TestEntity{
			Name:     fmt.Sprintf("Product %d", i),
			Email:    fmt.Sprintf("product%d@example.com", i),
			Age:      20 + (i % 50),
			IsActive: true,
		}
		_ = repo.Create(ctx, entity)
	}
}

func BenchmarkGormRepository_FindAll(b *testing.B) {
	repo, _ := setupTestRepository(&testing.T{})
	ctx := context.Background()

	// Create test data
	entities := createTestEntities()
	_ = repo.Create(ctx, entities...)

	filter := NewTestFilter().IsActiveEq(true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindAll(ctx, filter)
	}
}
