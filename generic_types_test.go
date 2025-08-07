package querybuilder

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dchlong/querybuilder/domain"
	"github.com/dchlong/querybuilder/generation"
	parserPkg "github.com/dchlong/querybuilder/parser"
	"github.com/dchlong/querybuilder/repository"
)

// IMPORTANT: Generic type parameters like `T any`, `Value T` are NOT supported by querybuilder.
// The querybuilder only supports concrete types like int, string, time.Time, etc.
// All tests in this file are skipped as generic types are not supported.

func TestGenericTypes_UnsupportedFeature(t *testing.T) {
	t.Skip("Generic type parameters like 'T any', 'Value T' are not supported by querybuilder")
}

func TestGenericTypes_DomainLayer(t *testing.T) {
	t.Skip("Generic type parameters are not supported by querybuilder")
	tests := []struct {
		name         string
		field        domain.Field
		isFilterable bool
		opCount      int // Expected number of supported operators
	}{
		{
			name: "generic value type",
			field: domain.Field{
				Name:     "Value",
				TypeName: "T",
				GoType:   "T",
				Type:     domain.FieldTypeString, // Assume T behaves like string
			},
			isFilterable: true,
			opCount:      10, // String operations
		},
		{
			name: "generic pointer type",
			field: domain.Field{
				Name:     "Pointer",
				TypeName: "*T",
				GoType:   "*T",
				Type:     domain.FieldTypePointer,
			},
			isFilterable: true,
			opCount:      4, // Equal, NotEqual, IsNull, IsNotNull
		},
		{
			name: "generic slice type",
			field: domain.Field{
				Name:     "Items",
				TypeName: "[]T",
				GoType:   "[]T",
				Type:     domain.FieldTypeSlice,
			},
			isFilterable: false,
			opCount:      2, // Slice fields still get Equal and NotEqual operators
		},
		{
			name: "generic map type",
			field: domain.Field{
				Name:     "Mapping",
				TypeName: "map[K]V",
				GoType:   "map[K]V",
				Type:     domain.FieldTypeMap,
			},
			isFilterable: false,
			opCount:      2, // Map fields still get Equal and NotEqual operators
		},
		{
			name: "constrained generic type",
			field: domain.Field{
				Name:     "Comparable",
				TypeName: "T",
				GoType:   "T comparable",
				Type:     domain.FieldTypeNumeric, // Assume comparable behaves like numeric
			},
			isFilterable: true,
			opCount:      8, // Numeric operations
		},
		{
			name: "nested generic type",
			field: domain.Field{
				Name:     "Nested",
				TypeName: "*[]T",
				GoType:   "*[]T",
				Type:     domain.FieldTypePointer, // Pointer to slice
			},
			isFilterable: true,
			opCount:      4, // Pointer operations
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test IsFilterable
			if tt.field.IsFilterable() != tt.isFilterable {
				t.Errorf("IsFilterable() = %v, want %v", tt.field.IsFilterable(), tt.isFilterable)
			}

			// Test SupportedOperators
			ops := tt.field.SupportedOperators()
			if len(ops) != tt.opCount {
				t.Errorf("SupportedOperators() count = %d, want %d", len(ops), tt.opCount)
			}
		})
	}
}

func TestGenericTypes_MethodFactory(t *testing.T) {
	t.Skip("Generic type parameters are not supported by querybuilder")
	factory := generation.NewMethodFactory()

	tests := []struct {
		name       string
		structName string
		field      domain.Field
		operator   string
		wantMethod string
		wantParam  string
	}{
		{
			name:       "generic value filter",
			structName: "Container",
			field: domain.Field{
				Name:     "Value",
				TypeName: "T",
				Type:     domain.FieldTypeString,
			},
			operator:   "Equal",
			wantMethod: "ValueEq",
			wantParam:  "value T",
		},
		{
			name:       "generic pointer filter",
			structName: "Holder",
			field: domain.Field{
				Name:     "Data",
				TypeName: "*T",
				Type:     domain.FieldTypePointer,
			},
			operator:   "Equal",
			wantMethod: "DataEq",
			wantParam:  "data *T",
		},
		{
			name:       "constrained generic",
			structName: "Store",
			field: domain.Field{
				Name:     "Key",
				TypeName: "K",
				GoType:   "K comparable",
				Type:     domain.FieldTypeString,
			},
			operator:   "Like",
			wantMethod: "KeyLike",
			wantParam:  "key K",
		},
		{
			name:       "complex generic type",
			structName: "Complex",
			field: domain.Field{
				Name:     "Handler",
				TypeName: "func(T) R",
				GoType:   "func(T) R",
				Type:     domain.FieldTypeString, // Treat as string for testing
			},
			operator:   "Equal",
			wantMethod: "HandlerEq",
			wantParam:  "handler func(T) R",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get operator from string
			var op repository.Operator
			switch tt.operator {
			case "Equal":
				op = repository.OperatorEqual
			case "Like":
				op = repository.OperatorLike
			default:
				t.Fatalf("Unknown operator: %s", tt.operator)
			}

			method := factory.CreateFilterMethod(tt.structName, tt.field, op)

			if method.Name != tt.wantMethod {
				t.Errorf("Method name = %v, want %v", method.Name, tt.wantMethod)
			}

			if method.Parameters != tt.wantParam {
				t.Errorf("Method parameters = %v, want %v", method.Parameters, tt.wantParam)
			}

			// Test that method body is not empty
			if method.Body == "" {
				t.Error("Method body should not be empty")
			}
		})
	}
}

func TestGenericTypes_UpdaterMethods(t *testing.T) {
	t.Skip("Generic type parameters are not supported by querybuilder")
	factory := generation.NewMethodFactory()

	genericField := domain.Field{
		Name:     "GenericValue",
		TypeName: "T",
		Type:     domain.FieldTypeString,
	}

	method := factory.CreateUpdaterMethod("Container", genericField)

	expectedMethod := "SetGenericValue"
	expectedParam := "genericValue T"
	expectedReceiver := "c *ContainerUpdater"

	if method.Name != expectedMethod {
		t.Errorf("Updater method name = %v, want %v", method.Name, expectedMethod)
	}

	if method.Parameters != expectedParam {
		t.Errorf("Updater method parameters = %v, want %v", method.Parameters, expectedParam)
	}

	if method.Receiver != expectedReceiver {
		t.Errorf("Updater method receiver = %v, want %v", method.Receiver, expectedReceiver)
	}

	// Test body contains generic type reference
	if !strings.Contains(method.Body, "genericValue") {
		t.Error("Updater method body should reference parameter")
	}
}

func TestGenericTypes_OrderMethods(t *testing.T) {
	t.Skip("Generic type parameters are not supported by querybuilder")
	factory := generation.NewMethodFactory()

	genericField := domain.Field{
		Name:     "Priority",
		TypeName: "T",
		Type:     domain.FieldTypeNumeric,
	}

	// Test ascending order
	ascMethod := factory.CreateOrderMethod("Queue", genericField, true)
	if ascMethod.Name != "OrderByPriorityAsc" {
		t.Errorf("Ascending order method name = %v, want OrderByPriorityAsc", ascMethod.Name)
	}

	// Test descending order
	descMethod := factory.CreateOrderMethod("Queue", genericField, false)
	if descMethod.Name != "OrderByPriorityDesc" {
		t.Errorf("Descending order method name = %v, want OrderByPriorityDesc", descMethod.Name)
	}

	// Test that both methods reference the field in their bodies
	if !strings.Contains(ascMethod.Body, "Priority") {
		t.Error("Ascending order method should reference field name")
	}

	if !strings.Contains(descMethod.Body, "Priority") {
		t.Error("Descending order method should reference field name")
	}
}

func TestGenericTypes_EndToEndGeneration(t *testing.T) {
	t.Skip("Skipping end-to-end test - requires full parser integration")
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "generic.go")
	outputFile := filepath.Join(tempDir, "generic_generated.go")

	// Create comprehensive generic test file
	testGoCode := `package generic

import "time"

//gen:querybuilder
type Container[T any] struct {
	ID       int64     ` + "`db:\"id\"`" + `
	Value    T         ` + "`db:\"value\"`" + `
	Pointer  *T        ` + "`db:\"pointer\"`" + `
	Slice    []T       ` + "`db:\"slice\"`" + `     // Not filterable
	Map      map[string]T ` + "`db:\"map_data\"`" + ` // Not filterable
	Created  time.Time ` + "`db:\"created_at\"`" + `
}

//gen:querybuilder
type Repository[K comparable, V any] struct {
	Key      K         ` + "`db:\"key\"`" + `
	Value    V         ` + "`db:\"value\"`" + `
	Version  int       ` + "`db:\"version\"`" + `
}

//gen:querybuilder
type Cache[T comparable] struct {
	Key      T         ` + "`db:\"cache_key\"`" + `
	Data     string    ` + "`db:\"cache_data\"`" + `
	TTL      *int64    ` + "`db:\"ttl\"`" + `
}
`

	err := os.WriteFile(inputFile, []byte(testGoCode), 0644)
	if err != nil {
		t.Fatalf("Failed to create generic test file: %v", err)
	}

	structsParser := &parserPkg.Structs{}
	generator := NewQueryBuilderGenerator(structsParser)

	ctx := context.Background()
	err = generator.Generate(ctx, inputFile, outputFile, "")
	if err != nil {
		t.Fatalf("Generic end-to-end generation failed: %v", err)
	}

	generatedCode, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated generic file: %v", err)
	}

	codeStr := string(generatedCode)

	// Test all generic struct types are generated
	expectedStructs := []string{
		"type ContainerFilters struct",
		"type ContainerUpdater struct",
		"type ContainerOptions struct",
		"type RepositoryFilters struct",
		"type RepositoryUpdater struct",
		"type RepositoryOptions struct",
		"type CacheFilters struct",
		"type CacheUpdater struct",
		"type CacheOptions struct",
	}

	for _, expected := range expectedStructs {
		if !strings.Contains(codeStr, expected) {
			t.Errorf("Generated code missing generic struct: %s", expected)
		}
	}

	// Test DB schema generation for generics
	expectedSchemas := []string{
		"var ContainerDBSchema = struct",
		"var RepositoryDBSchema = struct",
		"var CacheDBSchema = struct",
	}

	for _, expected := range expectedSchemas {
		if !strings.Contains(codeStr, expected) {
			t.Errorf("Generated code missing generic schema: %s", expected)
		}
	}

	// Test that non-filterable generic fields are excluded from filters
	nonFilterableChecks := []string{
		"SliceEq", // Should not exist
		"MapEq",   // Should not exist
	}

	for _, check := range nonFilterableChecks {
		if strings.Contains(codeStr, check) {
			t.Errorf("Generated code should not contain non-filterable method: %s", check)
		}
	}

	// Test that updater methods exist for all fields (including non-filterable)
	expectedUpdaters := []string{
		"SetValue",
		"SetSlice", // Should exist even though not filterable
		"SetMap",   // Should exist even though not filterable
	}

	for _, expected := range expectedUpdaters {
		if !strings.Contains(codeStr, expected) {
			t.Errorf("Generated code missing updater method: %s", expected)
		}
	}
}

func TestGenericTypes_ComplexConstraints(t *testing.T) {
	t.Skip("Skipping constraints test - requires full parser integration")
	// Test handling of complex generic constraints
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "constraints.go")
	outputFile := filepath.Join(tempDir, "constraints_generated.go")

	testGoCode := `package constraints

import "constraints"

//gen:querybuilder  
type Numeric[T constraints.Signed | constraints.Unsigned | constraints.Float] struct {
	Value     T       ` + "`db:\"value\"`" + `
	Min       T       ` + "`db:\"min_value\"`" + `
	Max       T       ` + "`db:\"max_value\"`" + `
	Default   *T      ` + "`db:\"default_value\"`" + `
}

//gen:querybuilder
type Ordered[T constraints.Ordered] struct {
	First     T       ` + "`db:\"first\"`" + `
	Second    T       ` + "`db:\"second\"`" + `
	Priority  int     ` + "`db:\"priority\"`" + `
}
`

	err := os.WriteFile(inputFile, []byte(testGoCode), 0644)
	if err != nil {
		t.Fatalf("Failed to create constraints test file: %v", err)
	}

	structsParser := &parserPkg.Structs{}
	generator := NewQueryBuilderGenerator(structsParser)

	ctx := context.Background()
	err = generator.Generate(ctx, inputFile, outputFile, "")
	if err != nil {
		t.Fatalf("Complex constraints generation failed: %v", err)
	}

	generatedCode, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated constraints file: %v", err)
	}

	codeStr := string(generatedCode)

	// Test that complex constrained generics are handled
	expectedElements := []string{
		"type NumericFilters struct",
		"type OrderedFilters struct",
		"func NewNumericFilters()",
		"func NewOrderedFilters()",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(codeStr, expected) {
			t.Errorf("Constraints code missing: %s", expected)
		}
	}
}

func TestGenericTypes_NestedGenerics(t *testing.T) {
	t.Skip("Generic type parameters are not supported by querybuilder")
	// Test deeply nested generic types
	factory := generation.NewMethodFactory()

	nestedField := domain.Field{
		Name:     "Complex",
		TypeName: "map[K][]V",
		GoType:   "map[K][]V",
		Type:     domain.FieldTypeMap, // Maps are not filterable
	}

	// Even though this is a map (not filterable), test that it doesn't crash
	field := domain.Field{
		Name:     "SimpleGeneric",
		TypeName: "T",
		Type:     domain.FieldTypeString,
	}

	method := factory.CreateFilterMethod("Nested", field, repository.OperatorEqual)

	if method.Name == "" {
		t.Error("Nested generic method should have a name")
	}

	if method.Body == "" {
		t.Error("Nested generic method should have a body")
	}

	// Test that complex nested types are handled gracefully
	_ = nestedField.IsFilterable()       // Should return false without panic
	_ = nestedField.SupportedOperators() // Should return empty slice without panic
}

func TestGenericTypes_ParameterNaming(t *testing.T) {
	t.Skip("Generic type parameters are not supported by querybuilder")
	factory := generation.NewMethodFactory()

	// Test parameter naming with generic types that might conflict with Go keywords
	tests := []struct {
		fieldName     string
		expectedParam string
	}{
		{"Type", "typeValue"},           // Go keyword
		{"Map", "mapValue"},             // Go keyword
		{"Chan", "chanValue"},           // Go keyword
		{"Interface", "interfaceValue"}, // Go keyword
		{"GenericT", "genericT"},        // Not a keyword
		{"Value", "value"},              // Not a keyword
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			field := domain.Field{
				Name:     tt.fieldName,
				TypeName: "T",
				Type:     domain.FieldTypeString,
			}

			method := factory.CreateFilterMethod("Test", field, repository.OperatorEqual)

			expectedParam := tt.expectedParam + " T"
			if method.Parameters != expectedParam {
				t.Errorf("Parameter naming = %v, want %v", method.Parameters, expectedParam)
			}
		})
	}
}

func TestGenericTypes_VariadicMethods(t *testing.T) {
	t.Skip("Generic type parameters are not supported by querybuilder")
	factory := generation.NewMethodFactory()

	genericField := domain.Field{
		Name:     "Items",
		TypeName: "T",
		Type:     domain.FieldTypeNumeric,
	}

	method := factory.CreateFilterMethod("Collection", genericField, repository.OperatorIn)

	// Test variadic parameter with generic type
	expectedParam := "itemss ...T" // Note: the parameter naming might need refinement
	if method.Parameters != expectedParam {
		t.Errorf("Variadic generic parameters = %v, want %v", method.Parameters, expectedParam)
	}

	// Test that method body handles variadic generic parameters
	if !strings.Contains(method.Body, "itemss") {
		t.Error("Variadic generic method body should reference parameter")
	}
}

func TestGenericTypes_NullOperators(t *testing.T) {
	t.Skip("Generic type parameters are not supported by querybuilder")
	factory := generation.NewMethodFactory()

	genericPtrField := domain.Field{
		Name:     "OptionalValue",
		TypeName: "*T",
		Type:     domain.FieldTypePointer,
	}

	// Test IsNull method
	nullMethod := factory.CreateFilterMethod("Optional", genericPtrField, repository.OperatorIsNull)

	if nullMethod.Name != "OptionalValueIsNull" {
		t.Errorf("Null method name = %v, want OptionalValueIsNull", nullMethod.Name)
	}

	if nullMethod.Parameters != "" {
		t.Errorf("Null method should have no parameters, got: %v", nullMethod.Parameters)
	}

	// Test IsNotNull method
	notNullMethod := factory.CreateFilterMethod("Optional", genericPtrField, repository.OperatorIsNotNull)

	if notNullMethod.Name != "OptionalValueIsNotNull" {
		t.Errorf("NotNull method name = %v, want OptionalValueIsNotNull", notNullMethod.Name)
	}

	// Test that both methods have proper bodies
	if !strings.Contains(nullMethod.Body, "Value:    nil") {
		t.Error("Null method body should set Value to nil")
	}

	if !strings.Contains(notNullMethod.Body, "Value:    nil") {
		t.Error("NotNull method body should set Value to nil")
	}
}
