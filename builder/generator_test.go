package builder

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dchlong/querybuilder/domain"
)

func TestNewGenerator(t *testing.T) {
	generator := NewGenerator()

	if generator == nil {
		t.Fatal("NewGenerator() returned nil")
	}

	if generator.methodFactory == nil {
		t.Error("methodFactory not initialized")
	}

	if generator.templates == nil {
		t.Error("templates not initialized")
	}
}

func TestGenerator_GenerateCode_EmptyStructs(t *testing.T) {
	generator := NewGenerator()
	ctx := context.Background()

	// Test with empty structs slice
	_, err := generator.GenerateCode(ctx, []domain.Struct{}, "test")
	if err == nil {
		t.Error("GenerateCode should return error for empty structs slice")
	}

	if !strings.Contains(err.Error(), "no structs provided") {
		t.Errorf("Error message should mention no structs provided, got: %v", err)
	}
}

func TestGenerator_GenerateCode_SingleStruct(t *testing.T) {
	generator := NewGenerator()
	ctx := context.Background()

	// Create test struct
	testStruct := domain.Struct{
		Name:        "Product",
		PackageName: "models",
		Fields: []domain.Field{
			{
				Name:     "ID",
				DBName:   "id",
				TypeName: "int64",
				Type:     domain.FieldTypeNumeric,
			},
			{
				Name:     "Name",
				DBName:   "name",
				TypeName: "string",
				Type:     domain.FieldTypeString,
			},
			{
				Name:     "Email",
				DBName:   "email",
				TypeName: "string",
				Type:     domain.FieldTypeString,
			},
		},
	}

	code, err := generator.GenerateCode(ctx, []domain.Struct{testStruct}, "models")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}

	if len(code) == 0 {
		t.Error("Generated code is empty")
	}

	codeStr := string(code)

	// Test package declaration
	if !strings.Contains(codeStr, "package models") {
		t.Error("Generated code missing package declaration")
	}

	// Test import statement
	if !strings.Contains(codeStr, `"github.com/dchlong/querybuilder/repository"`) {
		t.Error("Generated code missing repository import")
	}

	// Test filter struct generation
	if !strings.Contains(codeStr, "type ProductFilters struct") {
		t.Error("Generated code missing ProductFilters struct")
	}

	// Test constructor generation
	if !strings.Contains(codeStr, "func NewProductFilters() *ProductFilters") {
		t.Error("Generated code missing ProductFilters constructor")
	}

	// Test method generation (at least some methods should be present)
	if !strings.Contains(codeStr, "func (p *ProductFilters)") {
		t.Error("Generated code missing filter methods")
	}

	// Test updater generation
	if !strings.Contains(codeStr, "type ProductUpdater struct") {
		t.Error("Generated code missing ProductUpdater struct")
	}

	// Test options generation
	if !strings.Contains(codeStr, "type ProductOptions struct") {
		t.Error("Generated code missing ProductOptions struct")
	}

	// Test schema generation
	if !strings.Contains(codeStr, "type ProductDBSchemaField string") {
		t.Error("Generated code missing schema field type")
	}

	if !strings.Contains(codeStr, "var ProductDBSchema = struct") {
		t.Error("Generated code missing schema variable")
	}
}

func TestGenerator_GenerateCode_MultipleStructs(t *testing.T) {
	generator := NewGenerator()
	ctx := context.Background()

	structs := []domain.Struct{
		{
			Name:        "Product",
			PackageName: "models",
			Fields: []domain.Field{
				{Name: "ID", DBName: "id", TypeName: "int64", Type: domain.FieldTypeNumeric},
			},
		},
		{
			Name:        "Post",
			PackageName: "models",
			Fields: []domain.Field{
				{Name: "Title", DBName: "title", TypeName: "string", Type: domain.FieldTypeString},
			},
		},
	}

	code, err := generator.GenerateCode(ctx, structs, "models")
	if err != nil {
		t.Fatalf("GenerateCode with multiple structs failed: %v", err)
	}

	codeStr := string(code)

	// Test both structs are generated
	expectedTypes := []string{
		"type ProductFilters struct",
		"type ProductUpdater struct",
		"type ProductOptions struct",
		"type PostFilters struct",
		"type PostUpdater struct",
		"type PostOptions struct",
	}

	for _, expectedType := range expectedTypes {
		if !strings.Contains(codeStr, expectedType) {
			t.Errorf("Generated code missing type: %s", expectedType)
		}
	}
}

func TestGenerator_GenerateFile(t *testing.T) {
	generator := NewGenerator()
	ctx := context.Background()

	// Create temporary directory
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "generated.go")

	testStruct := domain.Struct{
		Name:        "Product",
		PackageName: "store",
		Fields: []domain.Field{
			{
				Name:     "Price",
				DBName:   "price",
				TypeName: "decimal.Decimal",
				Type:     domain.FieldTypeNumeric,
			},
		},
	}

	err := generator.GenerateFile(ctx, []domain.Struct{testStruct}, "store", outputPath)
	if err != nil {
		t.Fatalf("GenerateFile failed: %v", err)
	}

	// Test file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Generated file does not exist")
	}

	// Test file contents
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Test basic structure
	if !strings.Contains(contentStr, "package store") {
		t.Error("Generated file missing package declaration")
	}

	if !strings.Contains(contentStr, "type ProductFilters struct") {
		t.Error("Generated file missing struct definition")
	}
}

func TestGenerator_GenerateFile_DirectoryCreation(t *testing.T) {
	generator := NewGenerator()
	ctx := context.Background()

	// Create path with non-existent directory
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "subdir", "generated.go")

	testStruct := domain.Struct{
		Name:        "Test",
		PackageName: "test",
		Fields: []domain.Field{
			{Name: "ID", DBName: "id", TypeName: "int", Type: domain.FieldTypeNumeric},
		},
	}

	err := generator.GenerateFile(ctx, []domain.Struct{testStruct}, "test", outputPath)
	if err != nil {
		t.Fatalf("GenerateFile with directory creation failed: %v", err)
	}

	// Test that directory was created
	if _, err := os.Stat(filepath.Dir(outputPath)); os.IsNotExist(err) {
		t.Error("Output directory was not created")
	}

	// Test that file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Generated file does not exist")
	}
}

func TestGenerator_buildTemplateData(t *testing.T) {
	generator := NewGenerator()

	testStructs := []domain.Struct{
		{
			Name: "Product",
			Fields: []domain.Field{
				{
					Name:     "Name",
					DBName:   "name",
					TypeName: "string",
					Type:     domain.FieldTypeString,
				},
				{
					Name:     "Age",
					DBName:   "age",
					TypeName: "int",
					Type:     domain.FieldTypeNumeric,
				},
				{
					Name:     "Tags",
					DBName:   "tags",
					TypeName: "[]string",
					Type:     domain.FieldTypeSlice, // Not filterable
				},
			},
		},
	}

	templateData := generator.buildTemplateData(testStructs)

	// Test structure of template data
	structs, ok := templateData["Structs"].([]map[string]interface{})
	if !ok {
		t.Fatal("Template data Structs is not the expected type")
	}

	if len(structs) != 1 {
		t.Errorf("Expected 1 struct in template data, got %d", len(structs))
	}

	productStruct := structs[0]

	// Test struct name
	if productStruct["Name"] != "Product" {
		t.Errorf("Expected struct name 'Product', got %v", productStruct["Name"])
	}

	// Test fields (should include all fields)
	fields, ok := productStruct["Fields"].([]domain.Field)
	if !ok {
		t.Fatal("Fields is not the expected type")
	}

	// Should have 3 total fields (Name, Age, Tags)
	if len(fields) != 3 {
		t.Errorf("Expected 3 total fields, got %d", len(fields))
	}

	// Test that methods are generated
	filterMethods, ok := productStruct["FilterMethods"].([]domain.Method)
	if !ok {
		t.Fatal("FilterMethods is not the expected type")
	}

	if len(filterMethods) == 0 {
		t.Error("No filter methods generated")
	}

	updaterMethods, ok := productStruct["UpdaterMethods"].([]domain.Method)
	if !ok {
		t.Fatal("UpdaterMethods is not the expected type")
	}

	// Should have updater methods for all fields (including non-filterable)
	if len(updaterMethods) != 3 {
		t.Errorf("Expected 3 updater methods, got %d", len(updaterMethods))
	}

	orderMethods, ok := productStruct["OrderMethods"].([]domain.Method)
	if !ok {
		t.Fatal("OrderMethods is not the expected type")
	}

	// Should have 4 order methods (2 fields * 2 directions) for filterable fields only
	if len(orderMethods) != 4 {
		t.Errorf("Expected 4 order methods, got %d", len(orderMethods))
	}
}

func TestGenerator_buildPackageHeader(t *testing.T) {
	generator := NewGenerator()

	header := generator.buildPackageHeader("testpkg")

	expectedElements := []string{
		"// Code generated by querybuilder. DO NOT EDIT.",
		"package testpkg",
		`"github.com/dchlong/querybuilder/repository"`,
	}

	for _, element := range expectedElements {
		if !strings.Contains(header, element) {
			t.Errorf("Package header missing element: %s", element)
		}
	}
}

// Generic type tests for builder
func TestGenerator_GenericTypeHandling(t *testing.T) {
	generator := NewGenerator()
	ctx := context.Background()

	// Create struct with generic types
	genericStruct := domain.Struct{
		Name:        "Container",
		PackageName: "generic",
		Fields: []domain.Field{
			{
				Name:     "Value",
				DBName:   "value",
				TypeName: "T",
				Type:     domain.FieldTypeString, // Assume T behaves like string
			},
			{
				Name:     "Pointer",
				DBName:   "pointer",
				TypeName: "*T",
				Type:     domain.FieldTypePointer,
			},
			{
				Name:     "Slice",
				DBName:   "slice",
				TypeName: "[]T",
				Type:     domain.FieldTypeSlice, // Not filterable
			},
		},
	}

	code, err := generator.GenerateCode(ctx, []domain.Struct{genericStruct}, "generic")
	if err != nil {
		t.Fatalf("GenerateCode with generics failed: %v", err)
	}

	codeStr := string(code)

	// Test that generic types are preserved
	expectedGenericElements := []string{
		"type ContainerFilters struct",
		"func NewContainerFilters()",
		"ContainerDBSchemaField(\"value\")",
		"ContainerDBSchemaField(\"pointer\")",
		"ContainerDBSchemaField(\"slice\")",
	}

	for _, element := range expectedGenericElements {
		if !strings.Contains(codeStr, element) {
			t.Errorf("Generic code missing element: %s", element)
		}
	}

	// Test that generic parameters are used in methods
	// Should have methods for Value and Pointer (both filterable), but not Slice
	if !strings.Contains(codeStr, "func (c *ContainerFilters)") {
		t.Error("Generic struct should have filter methods")
	}
}

func TestGenerator_ErrorHandling(t *testing.T) {
	generator := NewGenerator()
	ctx := context.Background()

	// Test with invalid output path (root directory which should fail on most systems)
	invalidPath := "/root/invalid/path/test.go"

	testStruct := domain.Struct{
		Name:        "Test",
		PackageName: "test",
		Fields: []domain.Field{
			{Name: "ID", DBName: "id", TypeName: "int", Type: domain.FieldTypeNumeric},
		},
	}

	err := generator.GenerateFile(ctx, []domain.Struct{testStruct}, "test", invalidPath)
	if err == nil {
		t.Error("GenerateFile should fail with invalid path")
	}

	// Error should mention path creation or writing
	if !strings.Contains(err.Error(), "failed to") {
		t.Errorf("Error should be descriptive, got: %v", err)
	}
}

func TestGenerator_PerformanceWithLargeStruct(t *testing.T) {
	generator := NewGenerator()
	ctx := context.Background()

	// Create struct with many fields to test performance
	fields := make([]domain.Field, 50)
	for i := 0; i < 50; i++ {
		fields[i] = domain.Field{
			Name:     "Field" + string(rune('A'+i%26)) + string(rune('0'+i/26)),
			DBName:   "field_" + string(rune('a'+i%26)) + string(rune('0'+i/26)),
			TypeName: "string",
			Type:     domain.FieldTypeString,
		}
	}

	largeStruct := domain.Struct{
		Name:        "Large",
		PackageName: "test",
		Fields:      fields,
	}

	start := time.Now()
	code, err := generator.GenerateCode(ctx, []domain.Struct{largeStruct}, "test")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("GenerateCode with large struct failed: %v", err)
	}

	if len(code) == 0 {
		t.Error("Generated code is empty")
	}

	// Test that generation completes in reasonable time (less than 1 second)
	if duration > time.Second {
		t.Errorf("Code generation took too long: %v", duration)
	}

	t.Logf("Generated code for 50-field struct in %v", duration)
}
