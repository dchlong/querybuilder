package examples

import (
	"context"
	"testing"

	"github.com/dchlong/querybuilder"
	"github.com/dchlong/querybuilder/parser"
)

// ExampleUsage demonstrates how to use the clean querybuilder
func TestExampleProduct(t *testing.T) {
	ctx := context.Background()

	// Initialize the generator with a cleaner API
	structsParser := &parser.Structs{} // Initialize with your parser
	generator := querybuilder.NewQueryBuilderGenerator(structsParser)

	// Generate code with simple, clear method call
	err := generator.Generate(ctx, "./product.go", "product_querybuilder.go", "")
	if err != nil {
		panic(err)
	}
}
