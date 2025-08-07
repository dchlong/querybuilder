# GORM Repository Implementation

A comprehensive GORM-based repository implementation that seamlessly integrates with the existing query builder's filter and updater system.

## Overview

The `GormRepository` provides a production-ready, feature-rich repository implementation that leverages GORM's advanced ORM capabilities while maintaining compatibility with the generated filters and updaters from the querybuilder system.

## Features

### Core Repository Operations
- **Create**: Single and batch record creation with optimized batch sizing
- **FindOneByID**: Efficient single record lookup by primary key
- **FindOne**: Single record lookup with complex filtering
- **FindAll**: Multiple record retrieval with filtering, pagination, and sorting
- **Update**: Record updates using type-safe updaters

### Advanced GORM Features
- **Transactions**: Full transaction support with rollback capabilities
- **Batch Operations**: Optimized batch creation, updates, and deletions
- **Count & Exists**: Efficient existence and counting queries
- **Health Checks**: Database connection monitoring

### Performance & Monitoring
- **Connection Pooling**: Optimized database connection management
- **Health Checks**: Database connection monitoring

## Quick Start

### Basic Setup

```go
package main

import (
    "context"
    
    "github.com/dchlong/querybuilder/repository"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    // Setup database
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }
    
    // Create repository
    repo := repository.NewGormRepository[Product, *ProductFilters, *ProductUpdater](db)
    
    ctx := context.Background()
    
    // Use the repository
    products, err := repo.FindAll(ctx, NewProductFilters().IsActiveEq(true))
    if err != nil {
        log.Fatal(err)
    }
}
```


## Usage Examples

### Basic CRUD Operations

```go
// Create a single record
product := &Product{
    Name:     "Widget",
    Price:    19.99,
    IsActive: true,
}
err := repo.Create(ctx, product)

// Find by ID
product, found, err := repo.FindOneByID(ctx, 1)

// Find with filters
filter := NewProductFilters().
    IsActiveEq(true).
    PriceGte(10.0)
products, err := repo.FindAll(ctx, filter)

// Update record
updater := NewProductUpdater().
    SetPrice(24.99).
    SetStock(150)
err := repo.Update(ctx, product, updater)
```

### Advanced Operations

```go
// Batch creation
products := []*Product{ /* ... */ }
err := repo.CreateInBatches(ctx, 50, products...)

// Batch updates with filter
filter := NewProductFilters().CategoryIDEq(1)
updater := NewProductUpdater().SetIsActive(false)
rowsAffected, err := repo.UpdateWithFilter(ctx, filter, updater)

// Transaction
err := repo.WithTransaction(ctx, func(txRepo *repository.GormRepository[Product, *ProductFilters, *ProductUpdater]) error {
    // All operations within this function are in a transaction
    return txRepo.Create(ctx, product1, product2)
})
```

### Query Operations

```go
// Count records
count, err := repo.Count(ctx, NewProductFilters().IsActiveEq(true))

// Check existence
exists, err := repo.Exists(ctx, NewProductFilters().PriceGt(100))

// Pagination
products, err := repo.FindAll(ctx, filter,
    repository.WithLimit(20),
    repository.WithOffset(40),
    repository.WithSortField("created_at", "desc"),
)
```

## Integration with Generated Code

The repository seamlessly works with generated filters and updaters:

```go
// Generated filter usage
filter := NewProductFilters().
    NameLike("%widget%").
    PriceGte(10.0).
    PriceLte(100.0).
    IsActiveEq(true)

// Generated updater usage  
updater := NewProductUpdater().
    SetName("Updated Widget").
    SetPrice(29.99).
    SetStock(75)

// Generated options usage
options := NewProductOptions().
    OrderByPriceAsc().
    OrderByNameDesc()
```

## Error Handling

The repository provides comprehensive error handling:

```go
product, found, err := repo.FindOneByID(ctx, id)
if err != nil {
    // Database error occurred
    return fmt.Errorf("failed to find product: %w", err)
}
if !found {
    // Record not found (not an error)
    return ErrProductNotFound
}
```

## Performance Considerations

### Batch Operations
- Use `CreateInBatches` for large datasets
- Configure appropriate batch sizes based on your data
- Monitor memory usage with large batches

### Query Optimization
- Use `Count` instead of `FindAll` + `len()` for counting
- Use `Exists` for existence checks
- Apply appropriate database indexes

### Connection Management
- Configure connection pools appropriately
- Use health checks to monitor connection status
- Set reasonable query timeouts

## Testing

The repository includes comprehensive tests:

```bash
# Run repository tests
go test ./repository -v

# Run with coverage
go test ./repository -v -cover

# Run benchmarks
go test ./repository -v -bench=.
```


## Best Practices

### Repository Pattern
```go
// Define domain-specific repositories
type ProductRepository struct {
    *repository.GormRepository[Product, *ProductFilters, *ProductUpdater]
}

func (r *ProductRepository) GetActiveProductsByCategory(
    ctx context.Context, 
    categoryID int64,
) ([]*Product, error) {
    filter := NewProductFilters().
        CategoryIDEq(categoryID).
        IsActiveEq(true)
    
    return r.FindAll(ctx, filter)
}
```

### Transaction Management
```go
// Use transactions for multi-step operations
err := repo.WithTransaction(ctx, func(txRepo *GormRepository[...]) error {
    // Step 1: Create order
    if err := txRepo.Create(ctx, order); err != nil {
        return err
    }
    
    // Step 2: Update inventory
    return updateInventory(ctx, txRepo, orderItems)
})
```

### Error Handling
```go
// Distinguish between different error types
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return ErrProductNotFound
    }
    return fmt.Errorf("database error: %w", err)
}
```

## Migration from CommonStore

The `GormRepository` is fully compatible with the existing `CommonStore`:

```go
// Old way
store := repository.NewCommonStore[Product, *ProductFilters, *ProductUpdater](db)

// New way  
repo := repository.NewGormRepository[Product, *ProductFilters, *ProductUpdater](db)

// Same interface, more features
products, err := repo.FindAll(ctx, filter)
```

## Contributing

When contributing to the repository:

1. Add comprehensive tests for new features
2. Update documentation for any API changes
3. Follow existing code patterns and conventions
4. Ensure backward compatibility with existing interfaces

## License

This implementation follows the same license as the parent querybuilder project.