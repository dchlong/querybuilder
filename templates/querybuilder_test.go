package templates

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dchlong/querybuilder/domain"
)

func TestNewQueryBuilderTemplates(t *testing.T) {
	templates := NewQueryBuilderTemplates()

	if templates == nil {
		t.Fatal("NewQueryBuilderTemplates() returned nil")
	}

	if templates.Main == nil {
		t.Fatal("Main template not initialized")
	}

	// Test that template can be parsed without errors
	if templates.Main.Tree == nil {
		t.Error("Template not properly parsed")
	}
}

func TestTemplate_BasicExecution(t *testing.T) {
	templates := NewQueryBuilderTemplates()

	// Create test data
	testData := map[string]interface{}{
		"Structs": []map[string]interface{}{
			{
				"Name": "Product",
				"Fields": []domain.Field{
					{Name: "ID", DBName: "id", TypeName: "int64"},
					{Name: "Name", DBName: "name", TypeName: "string"},
				},
				"FilterMethods": []domain.Method{
					{
						Name:          "IDEq",
						Receiver:      "p *ProductFilters",
						Parameters:    "id int64",
						ReturnType:    "*ProductFilters",
						Body:          "// filter body",
						Documentation: "IDEq filters by ID equal",
					},
				},
				"UpdaterMethods": []domain.Method{
					{
						Name:          "SetName",
						Receiver:      "p *ProductUpdater",
						Parameters:    "name string",
						ReturnType:    "*ProductUpdater",
						Body:          "// updater body",
						Documentation: "SetName sets the name field",
					},
				},
				"OrderMethods": []domain.Method{
					{
						Name:          "OrderByNameAsc",
						Receiver:      "p *ProductOptions",
						Parameters:    "",
						ReturnType:    "*ProductOptions",
						Body:          "// order body",
						Documentation: "OrderByNameAsc orders by name ascending",
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := templates.Main.Execute(&buf, testData)
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	result := buf.String()

	// Test that output contains expected elements
	expectedElements := []string{
		"type ProductFilters struct",
		"func NewProductFilters()",
		"func (p *ProductFilters) IDEq(id int64) *ProductFilters",
		"type ProductUpdater struct",
		"func NewProductUpdater()",
		"func (p *ProductUpdater) SetName(name string) *ProductUpdater",
		"type ProductOptions struct",
		"func NewProductOptions()",
		"func (p *ProductOptions) OrderByNameAsc() *ProductOptions",
		"type ProductDBSchemaField string",
		"var ProductDBSchema = struct",
		"ID ProductDBSchemaField",
		"Name ProductDBSchemaField",
		"ID: ProductDBSchemaField(\"id\")",
		"Name: ProductDBSchemaField(\"name\")",
	}

	for _, element := range expectedElements {
		if !strings.Contains(result, element) {
			t.Errorf("Template output missing expected element: %s", element)
		}
	}
}

func TestTemplate_MultipleStructs(t *testing.T) {
	templates := NewQueryBuilderTemplates()

	// Test with multiple structs
	testData := map[string]interface{}{
		"Structs": []map[string]interface{}{
			{
				"Name": "Product",
				"Fields": []domain.Field{
					{Name: "ID", DBName: "id", TypeName: "int64"},
				},
				"FilterMethods":  []domain.Method{},
				"UpdaterMethods": []domain.Method{},
				"OrderMethods":   []domain.Method{},
			},
			{
				"Name": "Post",
				"Fields": []domain.Field{
					{Name: "Title", DBName: "title", TypeName: "string"},
				},
				"FilterMethods":  []domain.Method{},
				"UpdaterMethods": []domain.Method{},
				"OrderMethods":   []domain.Method{},
			},
		},
	}

	var buf bytes.Buffer
	err := templates.Main.Execute(&buf, testData)
	if err != nil {
		t.Fatalf("Template execution with multiple structs failed: %v", err)
	}

	result := buf.String()

	// Test both structs are generated
	expectedStructs := []string{
		"type ProductFilters struct",
		"func NewProductFilters()",
		"var ProductDBSchema",
		"type PostFilters struct",
		"func NewPostFilters()",
		"var PostDBSchema",
	}

	for _, element := range expectedStructs {
		if !strings.Contains(result, element) {
			t.Errorf("Multi-struct template output missing: %s", element)
		}
	}
}

func TestTemplate_EmptyData(t *testing.T) {
	templates := NewQueryBuilderTemplates()

	// Test with empty structs slice
	testData := map[string]interface{}{
		"Structs": []map[string]interface{}{},
	}

	var buf bytes.Buffer
	err := templates.Main.Execute(&buf, testData)
	if err != nil {
		t.Fatalf("Template execution with empty data failed: %v", err)
	}

	result := buf.String()

	// Should produce minimal output (just whitespace/newlines)
	trimmed := strings.TrimSpace(result)
	if trimmed != "" {
		t.Errorf("Empty data should produce empty output, got: %s", trimmed)
	}
}

func TestTemplate_MethodGeneration(t *testing.T) {
	templates := NewQueryBuilderTemplates()

	// Test detailed method generation
	testData := map[string]interface{}{
		"Structs": []map[string]interface{}{
			{
				"Name": "Product",
				"Fields": []domain.Field{
					{Name: "Price", DBName: "price", TypeName: "decimal.Decimal"},
				},
				"FilterMethods": []domain.Method{
					{
						Name:       "PriceGt",
						Receiver:   "p *ProductFilters",
						Parameters: "price decimal.Decimal",
						ReturnType: "*ProductFilters",
						Body: `p.filters[ProductDBSchema.Price] = append(p.filters[ProductDBSchema.Price], 
	&repository.Filter{
		Field:    string(ProductDBSchema.Price),
		Operator: repository.OperatorGreaterThan,
		Value:    price,
	})
return p`,
						Documentation: "PriceGt filters by price greater than",
					},
				},
				"UpdaterMethods": []domain.Method{
					{
						Name:       "SetPrice",
						Receiver:   "p *ProductUpdater",
						Parameters: "price decimal.Decimal",
						ReturnType: "*ProductUpdater",
						Body: `p.fields[string(ProductDBSchema.Price)] = price
return p`,
						Documentation: "SetPrice sets the price field for update",
					},
				},
				"OrderMethods": []domain.Method{
					{
						Name:       "OrderByPriceDesc",
						Receiver:   "p *ProductOptions",
						Parameters: "",
						ReturnType: "*ProductOptions",
						Body: `p.options = append(p.options, func(options *repository.Options) {
	options.SortFields = append(options.SortFields, &repository.SortField{
		Field:     string(ProductDBSchema.Price),
		Direction: "desc",
	})
})
return p`,
						Documentation: "OrderByPriceDesc orders results by price descending",
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := templates.Main.Execute(&buf, testData)
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	result := buf.String()

	// Test method documentation is included
	if !strings.Contains(result, "// PriceGt filters by price greater than") {
		t.Error("Method documentation not included in output")
	}

	// Test method signature is correct
	if !strings.Contains(result, "func (p *ProductFilters) PriceGt(price decimal.Decimal) *ProductFilters {") {
		t.Error("Filter method signature not correctly generated")
	}

	// Test method body is included
	if !strings.Contains(result, "repository.OperatorGreaterThan") {
		t.Error("Method body not included in output")
	}

	// Test updater method
	if !strings.Contains(result, "func (p *ProductUpdater) SetPrice(price decimal.Decimal) *ProductUpdater {") {
		t.Error("Updater method signature not correctly generated")
	}

	// Test order method
	if !strings.Contains(result, "func (p *ProductOptions) OrderByPriceDesc() *ProductOptions {") {
		t.Error("Order method signature not correctly generated")
	}
}

func TestTemplate_SpecialCharacters(t *testing.T) {
	templates := NewQueryBuilderTemplates()

	// Test with field names that might cause template issues
	testData := map[string]interface{}{
		"Structs": []map[string]interface{}{
			{
				"Name": "Special",
				"Fields": []domain.Field{
					{Name: "JSON", DBName: "json_data", TypeName: "string"},
					{Name: "XMLData", DBName: "xml_data", TypeName: "string"},
					{Name: "HTMLContent", DBName: "html_content", TypeName: "string"},
				},
				"FilterMethods":  []domain.Method{},
				"UpdaterMethods": []domain.Method{},
				"OrderMethods":   []domain.Method{},
			},
		},
	}

	var buf bytes.Buffer
	err := templates.Main.Execute(&buf, testData)
	if err != nil {
		t.Fatalf("Template execution with special characters failed: %v", err)
	}

	result := buf.String()

	// Test that special field names are handled correctly
	expectedMappings := []string{
		"JSON: SpecialDBSchemaField(\"json_data\")",
		"XMLData: SpecialDBSchemaField(\"xml_data\")",
		"HTMLContent: SpecialDBSchemaField(\"html_content\")",
	}

	for _, mapping := range expectedMappings {
		if !strings.Contains(result, mapping) {
			t.Errorf("Special character field mapping missing: %s", mapping)
		}
	}
}

// Generic type template tests
func TestTemplate_GenericTypes(t *testing.T) {
	templates := NewQueryBuilderTemplates()

	// Test template with generic type names
	testData := map[string]interface{}{
		"Structs": []map[string]interface{}{
			{
				"Name": "Container",
				"Fields": []domain.Field{
					{Name: "Value", DBName: "value", TypeName: "T"},
					{Name: "Pointer", DBName: "pointer", TypeName: "*T"},
					{Name: "Slice", DBName: "slice", TypeName: "[]T"},
				},
				"FilterMethods": []domain.Method{
					{
						Name:          "ValueEq",
						Receiver:      "c *ContainerFilters",
						Parameters:    "value T", // Generic parameter
						ReturnType:    "*ContainerFilters",
						Body:          "// generic filter body",
						Documentation: "ValueEq filters by value equal",
					},
				},
				"UpdaterMethods": []domain.Method{
					{
						Name:          "SetValue",
						Receiver:      "c *ContainerUpdater",
						Parameters:    "value T", // Generic parameter
						ReturnType:    "*ContainerUpdater",
						Body:          "// generic updater body",
						Documentation: "SetValue sets the generic value",
					},
				},
				"OrderMethods": []domain.Method{},
			},
		},
	}

	var buf bytes.Buffer
	err := templates.Main.Execute(&buf, testData)
	if err != nil {
		t.Fatalf("Template execution with generic types failed: %v", err)
	}

	result := buf.String()

	// Test that generic types are preserved in output
	if !strings.Contains(result, "func (c *ContainerFilters) ValueEq(value T) *ContainerFilters") {
		t.Error("Generic parameter type not preserved in filter method")
	}

	if !strings.Contains(result, "func (c *ContainerUpdater) SetValue(value T) *ContainerUpdater") {
		t.Error("Generic parameter type not preserved in updater method")
	}

	// Test that generic field mappings work
	if !strings.Contains(result, "Value: ContainerDBSchemaField(\"value\")") {
		t.Error("Generic field mapping not generated correctly")
	}
}

func TestTemplate_ComplexGenericScenarios(t *testing.T) {
	templates := NewQueryBuilderTemplates()

	// Test with complex generic constraints and nested types
	testData := map[string]interface{}{
		"Structs": []map[string]interface{}{
			{
				"Name": "Generic",
				"Fields": []domain.Field{
					{Name: "Constraint", DBName: "constraint", TypeName: "T comparable"},
					{Name: "MapField", DBName: "map_field", TypeName: "map[K]V"},
					{Name: "ChanField", DBName: "chan_field", TypeName: "chan T"},
				},
				"FilterMethods": []domain.Method{
					{
						Name:          "ConstraintEq",
						Receiver:      "g *GenericFilters",
						Parameters:    "constraint T", // This might be "T comparable" in real scenarios
						ReturnType:    "*GenericFilters",
						Body:          "// constraint body",
						Documentation: "ConstraintEq filters by constraint equal",
					},
				},
				"UpdaterMethods": []domain.Method{},
				"OrderMethods":   []domain.Method{},
			},
		},
	}

	var buf bytes.Buffer
	err := templates.Main.Execute(&buf, testData)
	if err != nil {
		t.Fatalf("Template execution with complex generics failed: %v", err)
	}

	result := buf.String()

	// Test that complex types don't break template
	if !strings.Contains(result, "type GenericFilters struct") {
		t.Error("Complex generic struct not generated")
	}

	// Test field mappings for complex types
	expectedMappings := []string{
		"Constraint: GenericDBSchemaField(\"constraint\")",
		"MapField: GenericDBSchemaField(\"map_field\")",
		"ChanField: GenericDBSchemaField(\"chan_field\")",
	}

	for _, mapping := range expectedMappings {
		if !strings.Contains(result, mapping) {
			t.Errorf("Complex generic field mapping missing: %s", mapping)
		}
	}
}
