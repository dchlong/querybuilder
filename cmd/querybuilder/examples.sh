#!/bin/bash

# QueryBuilder CLI Examples and Demo Script
# This script demonstrates all features of the QueryBuilder CLI tool

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Logging functions
log() { echo -e "${BLUE}[DEMO]${NC} $1"; }
success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }
header() { echo -e "\n${YELLOW}=== $1 ===${NC}"; }

# Ensure we're in the right directory (project root)
cd "$(dirname "$0")/../.."

# Build the CLI tool first
header "Building QueryBuilder CLI"
log "Building binary using Makefile..."
make build
success "CLI built successfully at ./bin/querybuilder"

QUERYBUILDER="./bin/querybuilder"

# Verify CLI is working
log "Verifying CLI is working..."
$QUERYBUILDER -version
success "CLI verification completed"

# Example 1: Basic usage with actual files
header "Example 1: Basic Usage"
log "Generating query builder for examples/product.go"
$QUERYBUILDER -verbose examples/product.go
success "Generated examples/product_querybuilder.go"

# Example 2: Custom output file
header "Example 2: Custom Output File"
log "Generating with custom output filename"
$QUERYBUILDER -output examples/product_queries.go examples/product.go
success "Generated examples/product_queries.go"

# Example 3: Struct suffix
header "Example 3: Struct Name Suffix"
log "Generating with V1 suffix for product types"
$QUERYBUILDER -suffix V1 -output examples/product_v1.go examples/product.go
success "Generated with V1 suffix (ProductV1Filters, ProductV1Updater, etc.)"

# Example 4: Short flags demonstration
header "Example 4: Short Flags Usage"
log "Using short flags: -o (output), -s (suffix), -v (version)"
$QUERYBUILDER -o examples/product_short.go -s V2 examples/product.go
success "Generated using short flags"

# Example 5: Directory processing with short flag
header "Example 5: Directory Processing"
log "Cleaning any previously generated files first..."
rm -f examples/*_querybuilder.go
log "Processing all Go files in examples directory using short flag -d"
$QUERYBUILDER -d examples -verbose
success "Processed all files in examples directory"

# Example 6: Dry run
header "Example 6: Dry Run (Preview Mode)"
log "Showing what would be generated without creating files"
$QUERYBUILDER -dry-run examples/product.go
success "Dry run completed - no files were created"

# Example 7: Show supported types
header "Example 7: Supported Field Types"
log "Displaying supported field types"
$QUERYBUILDER -types

# Example 8: Show version
header "Example 8: Version Information"
$QUERYBUILDER -version

# Demonstrate generated code usage
header "Example 9: Generated Code Demo"
log "Creating demonstration of generated API usage"

cat > /tmp/demo_usage.go << 'EOF'
package main

import (
    "fmt"
)

// Simulated generated code structure
type ProductFilters struct {
    filters map[string]interface{}
}

func NewProductFilters() *ProductFilters {
    return &ProductFilters{filters: make(map[string]interface{})}
}

func (f *ProductFilters) NameEq(name string) *ProductFilters {
    f.filters["name_eq"] = name
    return f
}

func (f *ProductFilters) PriceGt(price float64) *ProductFilters {
    f.filters["price_gt"] = price
    return f
}

func (f *ProductFilters) IsActiveEq(active bool) *ProductFilters {
    f.filters["is_active"] = active
    return f
}

func (f *ProductFilters) ListFilters() map[string]interface{} {
    return f.filters
}

func main() {
    // Example of fluent API usage
    filters := NewProductFilters().
        NameEq("Gaming Laptop").
        PriceGt(999.99).
        IsActiveEq(true)
    
    fmt.Println("Generated filters:")
    for key, value := range filters.ListFilters() {
        fmt.Printf("  %s: %v\n", key, value)
    }
}
EOF

log "Running generated code demo"
go run /tmp/demo_usage.go
rm /tmp/demo_usage.go
success "Demo completed"

# Show generated file structure
header "Example 10: Generated File Structure"
log "Examining generated file structure"

if [ -f "examples/product_querybuilder.go" ]; then
    log "Generated file contains:"
    echo "  - $(grep -c "func.*Filters" examples/product_querybuilder.go || echo 0) filter methods"
    echo "  - $(grep -c "func.*Updater" examples/product_querybuilder.go || echo 0) updater methods"
    echo "  - $(grep -c "func.*Order" examples/product_querybuilder.go || echo 0) order methods"
    echo "  - $(wc -l < examples/product_querybuilder.go) total lines"
    success "File structure analyzed"
fi

# Performance test
header "Example 11: Performance Test"
log "Testing generation performance"

time_start=$(date +%s%N)
$QUERYBUILDER examples/product.go > /dev/null 2>&1
time_end=$(date +%s%N)
duration=$(( (time_end - time_start) / 1000000 ))

log "Generation completed in ${duration}ms"
success "Performance test completed"

# Integration example
header "Example 12: Build Integration"
log "Example Makefile integration:"

cat << 'EOF'
# Add to your Makefile:
generate:
    querybuilder -d ./internal/models
    go fmt ./...

build: generate
    go build ./...

clean-generated:
    find . -name "*_querybuilder.go" -delete
EOF

success "Integration examples shown"

# Cleanup demonstration
header "Cleanup"
log "Cleaning up demo files..."
rm -f examples/product_queries.go examples/product_v1.go examples/product_short.go
success "Demo completed successfully!"

echo
header "Summary"
log "✅ Basic file generation"
log "✅ Custom output files"
log "✅ Struct name suffixes"
log "✅ Directory processing"
log "✅ Complex field types"
log "✅ Dry run mode"
log "✅ Type information"
log "✅ Performance testing"
log "✅ Integration examples"

echo
log "🎉 All examples completed successfully!"
log "Try running: $QUERYBUILDER --help"
log "Or process your own files: $QUERYBUILDER your-models.go"