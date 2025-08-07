# QueryBuilder 🚀

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![ORM Agnostic](https://img.shields.io/badge/ORM-Agnostic-purple.svg)](#orm-agnostic-design)
[![Architecture](https://img.shields.io/badge/Architecture-Clean-brightgreen.svg)](#architecture)
[![Generic Types](https://img.shields.io/badge/Generic%20Types-Not%20Supported-red.svg)](#generic-type-limitations)

**QueryBuilder** is a powerful, **ORM-agnostic** Go code generator that **decouples filtering and updating logic from ORM implementations**. Built with **Clean Architecture** principles, it generates type-safe query builders that work with any database layer - GORM, SQLx, or your custom ORM.

## ✨ Features

- 🔌 **ORM-Agnostic Design** - Works with GORM, SQLx, database/sql, or any custom ORM
- 🏗️ **Clean Architecture** - Layered design with clear separation of concerns
- ⚠️ **Generic Type Limitations** - QueryBuilder currently does not support generic type parameters like `T any`
- 🔍 **Advanced Filtering** - Type-safe filters with operators (Equal, Like, GreaterThan, In, etc.)
- 📝 **Smart Updates** - Fluent updater API for all field types
- 📊 **Flexible Ordering** - Ascending/descending sorting with multiple fields
- 🛡️ **Type Safety** - Compile-time validation of queries and updates
- 🚀 **High Performance** - Optimized code generation with minimal runtime overhead
- 📚 **Self-Documenting** - Generated code includes comprehensive documentation

## 🏗️ Architecture

QueryBuilder follows Clean Architecture principles with four distinct layers:

```
┌─────────────────┐
│   Templates     │  ← Code generation templates
├─────────────────┤
│   Generation    │  ← Method factories and builders  
├─────────────────┤
│    Builder      │  ← Orchestration and coordination
├─────────────────┤
│    Domain       │  ← Core types and business logic
└─────────────────┘
```

### Layer Responsibilities

- **Domain Layer**: Core types, field classification, and business rules
- **Generation Layer**: Method factories for filters, updaters, and ordering
- **Builder Layer**: Code generation orchestration and template coordination
- **Templates Layer**: Go template system for clean code output

## 🚀 Quick Start

### CLI Installation

```bash
# Install the CLI tool
curl -fsSL https://raw.githubusercontent.com/dchlong/querybuilder/main/install.sh | bash

# Or install directly with Go
go install github.com/dchlong/querybuilder/cmd/querybuilder@latest

# Verify installation
querybuilder --version
```

### Library Installation

```bash
go install github.com/dchlong/querybuilder/cmd/querybuilder@latest
```

### Basic Usage

1. **Add annotations to your structs**:

```go
package models

import (
    "time"
    "gorm.io/datatypes"
)

//gen:querybuilder
type Product struct {
    ID          int64                           `gorm:"column:id"`
    Name        string                          `gorm:"column:name"`
    SKU         string                          `gorm:"column:sku"`
    Price       float64                         `gorm:"column:price"`
    Stock       int                             `gorm:"column:stock"`
    CategoryID  int64                           `gorm:"column:category_id"`
    IsActive    bool                            `gorm:"column:is_active"`
    Tags        []string                        `gorm:"column:tags"` // JSON array
    Attributes  datatypes.JSONType[*Attributes] `gorm:"column:attributes"` // Generic type!
    CreatedAt   time.Time                       `gorm:"column:created_at"`
    UpdatedAt   *time.Time                      `gorm:"column:updated_at"`
}

type Attributes struct {
    Color      string  `json:"color"`
    Size       string  `json:"size"`
    Weight     float64 `json:"weight"`
    Dimensions string  `json:"dimensions"`
}
```

2. **Generate the querybuilder code**:

```bash
# Generate query builder for models.go
querybuilder models.go

# Or with custom output file
querybuilder -output models_queries.go models.go

# Process entire directory
querybuilder -dir ./internal/models
```

3. **Use the generated fluent API**:

```go
// Create complex filters
filters := NewProductFilters().
    NameLike("%widget%").
    PriceGt(10.0).
    SKULike("%PRD-%").
    IsActiveEq(true).
    CreatedAtGte(time.Now().AddDate(-1, 0, 0))

// Create updates
updater := NewProductUpdater().
    SetName("Premium Widget").
    SetPrice(29.99).
    SetStock(100).
    SetAttributes(datatypes.JSONType[*Attributes]{}) // Note: Concrete generic types like JSONType work

// Create ordering
options := NewProductOptions().
    OrderByCreatedAtDesc().
    OrderByNameAsc()

// Use with your repository
products, err := productRepo.FindAll(ctx, filters.ListFilters(), options)
rowsAffected, err := productRepo.UpdateWithFilter(ctx, filters.ListFilters(), updater.GetChangeSet())
```

## ⚠️ Generic Type Limitations

QueryBuilder currently **does not support generic type parameters** like `T any` or constrained generics like `T comparable`.

### Unsupported Generic Patterns

```go
//gen:querybuilder
type Container[T any] struct {
    Value    T                    `gorm:"column:value"`     // ❌ Generic type parameter not supported
    Pointer  *T                   `gorm:"column:pointer"`   // ❌ Generic pointer not supported  
    Slice    []T                  `gorm:"column:slice"`     // ❌ Generic slice not supported
    Map      map[string]T         `gorm:"column:mapping"`   // ❌ Generic map not supported
}

//gen:querybuilder  
type Repository[K comparable, V any] struct {
    Key      K                    `gorm:"column:key"`       // ❌ Constrained generic not supported
    Value    V                    `gorm:"column:value"`     // ❌ Any type generic not supported
    Version  int                  `gorm:"column:version"`   // ✅ Regular types are supported
}
```

### Recommended Alternatives

Use concrete types instead of generic type parameters:

```go
//gen:querybuilder
type ProductContainer struct {
    Value    Product              `gorm:"column:value"`     // ✅ Concrete type supported
    Pointer  *Product             `gorm:"column:pointer"`   // ✅ Concrete pointer supported  
    Slice    []Product            `gorm:"column:slice"`     // ✅ Concrete slice supported (updatable only)
    Map      map[string]Product   `gorm:"column:mapping"`   // ✅ Concrete map supported (updatable only)
}

//gen:querybuilder
type StringRepository struct {
    Key      string               `gorm:"column:key"`       // ✅ Concrete string supported
    Value    string               `gorm:"column:value"`     // ✅ Concrete string supported
    Version  int                  `gorm:"column:version"`   // ✅ Regular types are supported
}
```

## 📚 Comprehensive Examples

### Advanced Filtering

```go
// Numeric operations
filters := NewProductFilters().
    PriceGt(10.0).                // Greater than
    PriceLte(100.0).              // Less than or equal
    StockIn(25, 50, 100, 200)     // In list

// String operations  
filters = NewProductFilters().
    NameLike("%widget%").          // Pattern matching
    NameNotLike("%discontinued%"). // Negative pattern
    SKUIn("PRD-001", "PRD-002")

// Time operations
filters = NewProductFilters().
    CreatedAtGte(startDate).      // Greater than or equal
    CreatedAtLt(endDate).         // Less than
    UpdatedAtIsNull().            // Null checks
    UpdatedAtIsNotNull()          // Not null checks

// Boolean operations
filters = NewProductFilters().
    IsActiveEq(true).             // Boolean equality
    IsActiveNe(false)             // Boolean inequality

// Combine multiple conditions
complexFilters := NewProductFilters().
    NameLike("%premium%").
    PriceGt(50.0).
    IsActiveEq(true).
    CreatedAtGte(time.Now().AddDate(-2, 0, 0)).
    SKUNotLike("%temp%")
```

### Flexible Updates

```go
// Update individual fields
updater := NewProductUpdater().
    SetName("Updated Product").
    SetPrice(49.99).
    SetStock(150)

// Update with nil values for pointers
updater = NewProductUpdater().
    SetUpdatedAt(nil)             // Set to NULL
    
// Update with current timestamp
now := time.Now()
updater = NewProductUpdater().
    SetUpdatedAt(&now)            // Set to current time

// Update concrete generic types (specific instantiations work)
attributes := datatypes.JSONType[*Attributes]{
    Data: &Attributes{
        Color:      "blue",
        Size:       "large",
        Weight:     2.5,
        Dimensions: "10x5x2",
    },
}
updater = NewProductUpdater().
    SetAttributes(attributes)     // Concrete generic type update

// Chain multiple updates
updater = NewProductUpdater().
    SetName("Premium Widget").
    SetPrice(29.99).
    SetIsActive(true).
    SetUpdatedAt(&now).
    SetAttributes(attributes)
```

### Multi-Field Ordering

```go
// Single field ordering
options := NewProductOptions().
    OrderByNameAsc()              // A-Z sorting

options = NewProductOptions().
    OrderByCreatedAtDesc()        // Newest first

// Multi-field ordering
options = NewProductOptions().
    OrderByIsActiveDesc().        // Active products first
    OrderByCreatedAtDesc().       // Then by newest
    OrderByNameAsc()              // Then by name A-Z

// Complex sorting scenarios
options = NewProductOptions().
    OrderByPriceDesc().           // Most expensive first
    OrderByCreatedAtAsc().        // Then by creation date  
    OrderBySKUAsc()               // Then alphabetically by SKU
```

## 🔌 ORM-Agnostic Design

QueryBuilder **decouples filtering and updating logic from ORM implementations**, providing a clean separation between business logic and data access. The generated code produces standard Go types that work with any database layer.

### Key Benefits

- 🚫 **No ORM Lock-in** - Switch between GORM, SQLx, database/sql without changing business logic
- 🧩 **Clean Separation** - Business rules separated from database implementation details  
- 🔄 **Easy Migration** - Migrate between ORMs without rewriting query logic
- 🛡️ **Type Safety** - Compile-time validation regardless of ORM choice
- 🧪 **Testable** - Mock repositories easily without ORM dependencies

### Generated Output Structure

```go
// Generated types are ORM-agnostic
type ProductFilters struct {
    filters map[ProductDBSchemaField][]*repository.Filter
}

// Standard Go types for updates
func (u *ProductUpdater) GetChangeSet() map[string]interface{} {
    return u.fields // Plain map[string]interface{}
}

// Standard repository.Filter structure
type Filter struct {
    Field    string      // Database field name
    Operator string      // SQL operator (=, LIKE, >, etc.)
    Value    interface{} // Field value
}
```

### GORM Integration Example

```go
type ProductRepository struct {
    db *gorm.DB
}

func (r *ProductRepository) FindAll(ctx context.Context, filters []*repository.Filter, options *ProductOptions) ([]*Product, error) {
    query := r.db.WithContext(ctx)
    
    // Apply filters - ORM-specific implementation
    for _, filter := range filters {
        query = query.Where(fmt.Sprintf("%s %s ?", filter.Field, filter.Operator), filter.Value)
    }
    
    // Apply ordering - ORM-specific implementation
    if options != nil {
        var repoOptions repository.Options
        options.Apply(&repoOptions)
        
        for _, sortField := range repoOptions.SortFields {
            query = query.Order(fmt.Sprintf("%s %s", sortField.Field, sortField.Direction))
        }
    }
    
    var products []*Product
    err := query.Find(&products).Error
    return products, err
}

func (r *ProductRepository) UpdateWithFilter(ctx context.Context, filters []*repository.Filter, changeSet map[string]interface{}) (int64, error) {
    query := r.db.WithContext(ctx).Model(&Product{})
    
    // Apply filters
    for _, filter := range filters {
        query = query.Where(fmt.Sprintf("%s %s ?", filter.Field, filter.Operator), filter.Value)
    }
    
    result := query.Updates(changeSet)
    return result.RowsAffected, result.Error
}
```

### SQLx Integration Example

```go
import (
    "github.com/jmoiron/sqlx"
    "github.com/Masterminds/squirrel"
)

type ProductRepository struct {
    db *sqlx.DB
}

func (r *ProductRepository) FindAll(ctx context.Context, filters []*repository.Filter, options *ProductOptions) ([]*Product, error) {
    // Build query with Squirrel
    query := squirrel.Select("*").From("products")
    
    // Apply filters - different ORM, same input
    for _, filter := range filters {
        query = query.Where(squirrel.Expr(fmt.Sprintf("%s %s ?", filter.Field, filter.Operator), filter.Value))
    }
    
    // Apply ordering
    if options != nil {
        var repoOptions repository.Options
        options.Apply(&repoOptions)
        
        for _, sortField := range repoOptions.SortFields {
            query = query.OrderBy(fmt.Sprintf("%s %s", sortField.Field, sortField.Direction))
        }
    }
    
    sql, args, err := query.ToSql()
    if err != nil {
        return nil, err
    }
    
    var products []*Product
    err = r.db.SelectContext(ctx, &products, sql, args...)
    return products, err
}

func (r *ProductRepository) UpdateWithFilter(ctx context.Context, filters []*repository.Filter, changeSet map[string]interface{}) (int64, error) {
    query := squirrel.Update("products")
    
    // Apply updates - same changeSet format
    for field, value := range changeSet {
        query = query.Set(field, value)
    }
    
    // Apply filters
    for _, filter := range filters {
        query = query.Where(squirrel.Expr(fmt.Sprintf("%s %s ?", filter.Field, filter.Operator), filter.Value))
    }
    
    sql, args, err := query.ToSql()
    if err != nil {
        return 0, err
    }
    
    result, err := r.db.ExecContext(ctx, sql, args...)
    if err != nil {
        return 0, err
    }
    
    return result.RowsAffected()
}
```

### Raw database/sql Integration Example

```go
import (
    "database/sql"
    "strings"
)

type ProductRepository struct {
    db *sql.DB
}

func (r *ProductRepository) FindAll(ctx context.Context, filters []*repository.Filter, options *ProductOptions) ([]*Product, error) {
    query := "SELECT id, name, sku, price, stock, category_id, is_active, created_at, updated_at FROM products"
    var args []interface{}
    
    // Apply filters - pure SQL
    if len(filters) > 0 {
        var conditions []string
        for _, filter := range filters {
            conditions = append(conditions, fmt.Sprintf("%s %s ?", filter.Field, filter.Operator))
            args = append(args, filter.Value)
        }
        query += " WHERE " + strings.Join(conditions, " AND ")
    }
    
    // Apply ordering
    if options != nil {
        var repoOptions repository.Options
        options.Apply(&repoOptions)
        
        if len(repoOptions.SortFields) > 0 {
            var orderClauses []string
            for _, sortField := range repoOptions.SortFields {
                orderClauses = append(orderClauses, fmt.Sprintf("%s %s", sortField.Field, sortField.Direction))
            }
            query += " ORDER BY " + strings.Join(orderClauses, ", ")
        }
    }
    
    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var products []*Product
    for rows.Next() {
        product := &Product{}
        err := rows.Scan(&product.ID, &product.Name, &product.SKU, &product.Price, &product.Stock, &product.CategoryID, &product.IsActive, &product.CreatedAt, &product.UpdatedAt)
        if err != nil {
            return nil, err
        }
        products = append(products, product)
    }
    
    return products, rows.Err()
}

func (r *ProductRepository) UpdateWithFilter(ctx context.Context, filters []*repository.Filter, changeSet map[string]interface{}) (int64, error) {
    if len(changeSet) == 0 {
        return 0, nil
    }
    
    query := "UPDATE products SET "
    var setParts []string
    var args []interface{}
    
    // Apply updates
    for field, value := range changeSet {
        setParts = append(setParts, fmt.Sprintf("%s = ?", field))
        args = append(args, value)
    }
    query += strings.Join(setParts, ", ")
    
    // Apply filters
    if len(filters) > 0 {
        var conditions []string
        for _, filter := range filters {
            conditions = append(conditions, fmt.Sprintf("%s %s ?", filter.Field, filter.Operator))
            args = append(args, filter.Value)
        }
        query += " WHERE " + strings.Join(conditions, " AND ")
    }
    
    result, err := r.db.ExecContext(ctx, query, args...)
    if err != nil {
        return 0, err
    }
    
    return result.RowsAffected()
}
```

### Business Logic Layer (ORM-Independent)

```go
// Service layer is completely ORM-agnostic
type ProductService struct {
    repo ProductRepositoryInterface
}

type ProductRepositoryInterface interface {
    FindAll(ctx context.Context, filters []*repository.Filter, options *ProductOptions) ([]*Product, error)
    UpdateWithFilter(ctx context.Context, filters []*repository.Filter, changeSet map[string]interface{}) (int64, error)
}

func (s *ProductService) GetActivePremiumProducts(ctx context.Context) ([]*Product, error) {
    // Business logic using generated querybuilder - ORM independent!
    filters := NewProductFilters().
        IsActiveEq(true).
        NameLike("%premium%").
        PriceGt(50.0)
    
    options := NewProductOptions().
        OrderByCreatedAtDesc().
        OrderByNameAsc()
    
    return s.repo.FindAll(ctx, filters.ListFilters(), options)
}

func (s *ProductService) UpdateCategoryPricing(ctx context.Context, categoryID int64, priceAdjustment float64) (int64, error) {
    // Update logic is also ORM independent
    filters := NewProductFilters().
        IsActiveEq(true).
        CategoryIDEq(categoryID)
    
    now := time.Now()
    updater := NewProductUpdater().
        SetPrice(priceAdjustment).
        SetUpdatedAt(&now)
    
    return s.repo.UpdateWithFilter(ctx, filters.ListFilters(), updater.GetChangeSet())
}
```

## 🎯 Field Type Support

### Filterable Types (Support all operators)

| Type | Operators | Example |
|------|-----------|---------|
| `string` | Eq, Ne, Like, NotLike, In, NotIn, Lt, Gt, Lte, Gte | `NameLike("%widget%")` |
| `int`, `int64`, `float64` | Eq, Ne, Lt, Gt, Lte, Gte, In, NotIn | `PriceGt(10.0)` |
| `time.Time` | Eq, Ne, Lt, Gt, Lte, Gte, In, NotIn | `CreatedAtGte(startDate)` |
| `bool` | Eq, Ne | `IsActiveEq(true)` |
| `*T` (pointers) | Eq, Ne, IsNull, IsNotNull | `UpdatedAtIsNull()` |

### Updatable-Only Types (Can be set but not filtered)

| Type | Capability | Example |
|------|------------|---------|
| `[]T` (slices) | Update only | `SetTags([]string{"electronics", "gadgets"})` |
| `map[K]V` (maps) | Update only | `SetAttributes(map[string]string{})` |
| `struct` | Update only | `SetConfig(ConfigStruct{})` |
| `datatypes.JSONType[T]` | Update only | `SetAttributes(attributesData)` |

### Note on Concrete Generic Types

While generic type parameters like `T any` are not supported, concrete instantiations of generic types (like `datatypes.JSONType[*Attributes]`) work normally and follow standard type behavior rules.

## 🛠️ Command Line Interface

The `querybuilder` command provides a rich CLI experience:

### Basic Commands

```bash
# Generate with defaults
querybuilder -in models.go

# Custom output file
querybuilder -in models.go -out generated_queries.go

# Add suffix to generated types
querybuilder -in models.go -suffix V1

# Verbose output
querybuilder -in models.go -v
```

### Information Commands

```bash
# Show version and features
querybuilder -version

# List supported field types
querybuilder -supported

# Show help and examples
querybuilder -help
```

### Advanced Options

```bash
# Custom timeout for large files
querybuilder -in large_models.go -timeout 5m

# Generate with suffix and verbose output
querybuilder -in models.go -out custom_name.go -suffix V1 -v
```

## 🏗️ Programmatic Usage

Use QueryBuilder programmatically in your applications:

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/dchlong/querybuilder/parser"
    "github.com/dchlong/querybuilder"
)

func main() {
    ctx := context.Background()
    
    // Create generator
    structsParser := &parser.Structs{}
    generator := querybuilder.NewQueryBuilderGenerator(structsParser)
    
    // Generate to file
    err := generator.Generate(ctx, "models.go", "models_querybuilder.go", "")
    if err != nil {
        panic(err)
    }
    
    // Generate in memory
    code, packageName, err := generator.GenerateInMemory(ctx, "models.go", "")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Generated %d bytes for package %s\n", len(code), packageName)
    
    // Check supported types
    supported := generator.GetSupportedFieldTypes()
    unsupported := generator.GetUnsupportedFieldTypes()
    
    fmt.Printf("Supported: %v\n", supported)
    fmt.Printf("Unsupported: %v\n", unsupported)
}
```

## 🔧 Configuration

### Annotation Formats

QueryBuilder supports multiple annotation formats:

```go
//gen:querybuilder          // Preferred format
type Product struct { ... }

//@querybuilder             // Alternative format  
type Product struct { ... }

//+querybuilder             // Another alternative
type Order struct { ... }
```

### DB Field Mapping

Use struct tags to map Go fields to database columns:

```go
// Using GORM tags (recommended)
type Product struct {
    ID          int64     `gorm:"column:id"`           // Maps to "id" column
    Name        string    `gorm:"column:product_name"` // Maps to "product_name" column  
    SKU         string    `gorm:"column:sku_code"`     // Maps to "sku_code" column
    Price       float64   `gorm:"column:price"`        // Maps to "price" column
    CreatedAt   time.Time `gorm:"column:created_at"`   // Maps to "created_at" column
}

// Alternative using SQL tags
type Product struct {
    ID          int64     `sql:"column:id"`            // Maps to "id" column
    Name        string    `sql:"column:product_name"`  // Maps to "product_name" column  
    SKU         string    `sql:"column:sku_code"`      // Maps to "sku_code" column
    Price       float64   `sql:"column:price"`         // Maps to "price" column
    CreatedAt   time.Time `sql:"column:created_at"`    // Maps to "created_at" column
}

// Without explicit tags, uses GORM naming strategy (snake_case conversion)
type Product struct {
    ID         int64     // Maps to "id" column
    Name       string    // Maps to "name" column
    CategoryID int64     // Maps to "category_id" column (snake_case)
    CreatedAt  time.Time // Maps to "created_at" column (snake_case)
}
```

## 🧪 Testing

QueryBuilder includes comprehensive tests:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test suites
go test ./domain -v          # Domain layer tests
go test ./generation -v      # Generation layer tests  
go test ./templates -v       # Template tests
go test ./builder -v         # Builder layer tests

# Run integration tests
go test . -v                 # Root package integration tests
```

### Test Coverage

- **Domain Layer**: Field type classification, operator support, concrete type handling
- **Generation Layer**: Method factory, parameter naming, body generation
- **Templates Layer**: Template rendering, output formatting
- **Builder Layer**: Code generation orchestration, file operations
- **Integration Tests**: End-to-end generation, real-world scenarios
- **Type Tests**: Field type classification and concrete type validation

## 🚀 Performance

QueryBuilder is optimized for performance:

- **Fast Generation**: Efficient template rendering and code generation
- **Minimal Runtime Overhead**: Generated code has minimal performance impact
- **Memory Efficient**: Smart memory usage during generation
- **Concurrent Safe**: Thread-safe generation for parallel processing

### Benchmarks

```
BenchmarkGeneration-8        1000    1.2ms/op    245KB/op
BenchmarkTemplateRender-8    5000    0.3ms/op     87KB/op
BenchmarkMethodFactory-8    10000    0.1ms/op     23KB/op
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/dchlong/querybuilder.git
cd querybuilder

# Install dependencies
go mod download

# Run tests
go test ./...

# Build command
go build -o querybuilder ./cmd/querybuilder
```

### Code Style

- Follow standard Go conventions
- Add comprehensive tests for new features
- Update documentation for API changes
- Ensure all tests pass before submitting PR

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Clean Architecture principles by Robert C. Martin
- Go community for excellent tooling and libraries
- GORM team for database integration patterns
- All contributors who helped shape this project

## 📧 Support

- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/dchlong/querybuilder/issues)
- 💡 **Feature Requests**: [GitHub Discussions](https://github.com/dchlong/querybuilder/discussions)
- 📚 **Documentation**: [Wiki Pages](https://github.com/dchlong/querybuilder/wiki)
- 💬 **Community**: [Discord Server](https://discord.gg/querybuilder)

---

**Built with ❤️ for the Go community**

*QueryBuilder - Type-safe, ORM-Agnostic, Clean Architecture*