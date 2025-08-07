# QueryBuilder CLI

A command-line tool for generating type-safe query builders for Go structs.

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/dchlong/querybuilder.git
cd querybuilder

# Install the CLI tool
make install

# Or build locally
make build
./bin/querybuilder --help
```

### Direct Installation

```bash
go install github.com/dchlong/querybuilder/cmd/querybuilder@latest
```

## Usage

### Basic Usage

```bash
# Generate query builder for a single file
querybuilder models.go

# Generate with custom output file
querybuilder -output models_qb.go models.go

# Generate with struct name suffix
querybuilder -suffix V1 models.go
```

### Directory Processing

```bash
# Process all Go files in a directory
querybuilder -dir ./models

# Process with verbose output
querybuilder -dir ./models -verbose
```

### Options

```bash
querybuilder [options] <input-file>

Options:
  -output, -o <file>    Output file path (default: <input>_querybuilder.go)
  -suffix, -s <suffix>  Suffix to append to struct names
  -dir, -d <directory>  Process all Go files in directory
  -types                Show supported field types
  -version, -v          Show version
  -help, -h             Show help
  -verbose              Verbose output
  -dry-run              Show what would be generated without writing files
```

## Examples

### 1. Basic Model Generation

**Input file: `product.go`**
```go
package models

import "time"

//gen:querybuilder
type Product struct {
    ID        int64     `db:"id"`
    Name      string    `db:"name"`
    Email     string    `db:"email"`
    CreatedAt time.Time `db:"created_at"`
    IsActive  bool      `db:"is_active"`
}
```

**Command:**
```bash
querybuilder product.go
```

**Generated: `product_querybuilder.go`**
- `ProductFilters` with methods like `NameEq()`, `EmailLike()`, `IsActiveEq()`
- `ProductUpdater` with methods like `SetName()`, `SetEmail()`
- `ProductOptions` with methods like `OrderByNameAsc()`, `OrderByCreatedAtDesc()`
- `ProductDBSchema` with field constants

### 2. Complex Model with JSON Fields

**Input file: `product.go`**
```go
package models

import (
    "time"
    "gorm.io/datatypes"
)

//gen:querybuilder
type Product struct {
    ID          int64                         `db:"id"`
    Name        string                        `db:"name"`
    Price       float64                       `db:"price"`
    Tags        []string                      `db:"tags"`
    Attributes  datatypes.JSONType[*Attrs]    `db:"attributes"`
    CreatedAt   time.Time                     `db:"created_at"`
    UpdatedAt   *time.Time                    `db:"updated_at"`
}

type Attrs struct {
    Color string `json:"color"`
    Size  string `json:"size"`
}
```

**Command:**
```bash
querybuilder -verbose product.go
```

### 3. Directory Processing

```bash
# Process all models in a directory
querybuilder -dir ./internal/models -verbose

# Dry run to see what would be generated
querybuilder -dir ./internal/models -dry-run
```

### 4. Custom Output and Suffixes

```bash
# Custom output file
querybuilder -output generated/user_queries.go user.go

# Add suffix to struct names (User becomes UserV1)
querybuilder -suffix V1 user.go
```

### 5. Integration with Build Process

**In Makefile:**
```makefile
generate:
	querybuilder -dir ./internal/models
	go fmt ./...

build: generate
	go build ./...
```

**In CI/CD:**
```yaml
- name: Generate Query Builders
  run: |
    go install github.com/dchlong/querybuilder/cmd/querybuilder@latest
    querybuilder -dir ./internal/models
    
- name: Verify Generated Code
  run: |
    git diff --exit-code || (echo "Generated code is out of date" && exit 1)
```

## Supported Field Types

### Fully Supported
- `string` - String operations (eq, ne, like, in, etc.)
- `int`, `int64` - Numeric operations (eq, ne, lt, gt, in, etc.)
- `float64` - Numeric operations with decimal support
- `bool` - Boolean operations (eq, ne)
- `time.Time` - Time operations (eq, ne, lt, gt, in, etc.)

### Pointer Types (Nullable)
- `*string` - Nullable string with IsNull/IsNotNull operations
- `*time.Time` - Nullable timestamp with IsNull/IsNotNull operations
- `*int64` - Nullable integer with IsNull/IsNotNull operations

### JSON Types
- `[]string` - JSON array stored as string
- `datatypes.JSONType[T]` - GORM JSON type with Go struct mapping

### Not Supported
- `map[string]interface{}` - Use `datatypes.JSONType[T]` instead
- `struct` (embedded) - Flatten or use JSON field
- Complex slices - Use JSON field for complex arrays

## Struct Annotation

Add the query builder annotation to structs you want to process:

```go
//gen:querybuilder
type MyStruct struct {
    // fields...
}
```

## Generated API

For each annotated struct, the tool generates:

### 1. Filters (`StructFilters`)
```go
filters := NewProductFilters().
    NameEq("John").
    AgeGt(18).
    IsActiveEq(true)

// Get repository filters
repoFilters := filters.ListFilters()
```

### 2. Updaters (`StructUpdater`)
```go
updater := NewProductUpdater().
    SetName("John Doe").
    SetAge(25)

// Get changeset
changes := updater.GetChangeSet()
```

### 3. Options (`StructOptions`)
```go
options := NewProductOptions().
    OrderByNameAsc().
    OrderByCreatedAtDesc()

// Apply to repository options
options.Apply(repoOptions)
```

### 4. Schema (`StructDBSchema`)
```go
// Type-safe field references
fieldName := ProductDBSchema.Name.String() // "name"
```

## Error Handling

The CLI provides clear error messages for common issues:

```bash
# File not found
querybuilder nonexistent.go
# Error: input file does not exist: nonexistent.go

# No annotated structs
querybuilder empty.go  
# Error: no structs with querybuilder annotations found in empty.go

# Invalid Go syntax
querybuilder invalid.go
# Error: failed to parse input file: invalid.go: expected 'package', found 'invalid'
```

## Debugging

Use verbose mode to see detailed information:

```bash
querybuilder -verbose -dry-run user.go
# Input file:  user.go
# Output file: user_querybuilder.go
# Would generate 15847 bytes of code for package 'models'
# Output would be written to: user_querybuilder.go
```

## Integration Examples

### With Repository Pattern

```go
type UserRepository struct {
    // your repository implementation
}

func (r *UserRepository) FindUsers(ctx context.Context) ([]*User, error) {
    filters := NewProductFilters().
        IsActiveEq(true).
        AgeGt(18)
    
    options := NewProductOptions().
        OrderByNameAsc()
    
    return r.Find(ctx, filters.ListFilters(), options)
}
```

### With GORM

```go
func FindActiveUsers(db *gorm.DB) ([]*User, error) {
    filters := NewProductFilters().
        IsActiveEq(true).
        CreatedAtGt(time.Now().AddDate(-1, 0, 0))
    
    query := db.Model(&User{})
    
    for _, filter := range filters.ListFilters() {
        query = query.Where(fmt.Sprintf("%s %s ?", filter.Field, filter.Operator), filter.Value)
    }
    
    var users []*User
    return users, query.Find(&users).Error
}
```

## Tips

1. **Use consistent struct tags**: The `db` tag defines the database field name
2. **Annotation placement**: Place `//gen:querybuilder` directly above the struct
3. **Naming conventions**: Follow Go naming conventions for generated code
4. **Version control**: Consider committing generated files or generating in CI
5. **Build integration**: Use `go generate` or Makefile targets for automation

## Troubleshooting

### Common Issues

1. **"No structs found"** - Ensure the `//gen:querybuilder` annotation is present
2. **Parse errors** - Check Go syntax in the input file
3. **Missing imports** - Generated files import required packages automatically
4. **Field type errors** - Use `querybuilder -types` to see supported types

### Getting Help

```bash
# Show all options
querybuilder -help

# Show supported field types
querybuilder -types

# Show version
querybuilder -version
```