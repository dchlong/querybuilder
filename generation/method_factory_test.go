package generation

import (
	"strings"
	"testing"

	"github.com/dchlong/querybuilder/domain"
	"github.com/dchlong/querybuilder/repository"
)

func TestNewMethodFactory(t *testing.T) {
	factory := NewMethodFactory()

	if factory == nil {
		t.Fatal("NewMethodFactory() returned nil")
	}

	if factory.operatorNames == nil {
		t.Error("operatorNames map not initialized")
	}

	if factory.methodSuffixes == nil {
		t.Error("methodSuffixes map not initialized")
	}

	// Test that all operators are mapped
	expectedOperators := []repository.Operator{
		repository.OperatorEqual,
		repository.OperatorNotEqual,
		repository.OperatorLessThan,
		repository.OperatorLessThanOrEqual,
		repository.OperatorGreaterThan,
		repository.OperatorGreaterThanOrEqual,
		repository.OperatorLike,
		repository.OperatorNotLike,
		repository.OperatorIsNull,
		repository.OperatorIsNotNull,
		repository.OperatorIn,
		repository.OperatorNotIn,
	}

	for _, op := range expectedOperators {
		if _, exists := factory.operatorNames[op]; !exists {
			t.Errorf("operatorNames missing mapping for %v", op)
		}
		if _, exists := factory.methodSuffixes[op]; !exists {
			t.Errorf("methodSuffixes missing mapping for %v", op)
		}
	}
}

func TestMethodFactory_CreateFilterMethod_Binary(t *testing.T) {
	factory := NewMethodFactory()

	field := domain.Field{
		Name:     "Name",
		TypeName: "string",
		Type:     domain.FieldTypeString,
	}

	method := factory.CreateFilterMethod("Product", field, repository.OperatorEqual)

	// Test method properties
	if method.Name != "NameEq" {
		t.Errorf("Method name = %v, want NameEq", method.Name)
	}

	if method.Receiver != "p *ProductFilters" {
		t.Errorf("Method receiver = %v, want 'p *ProductFilters'", method.Receiver)
	}

	if method.Parameters != "name string" {
		t.Errorf("Method parameters = %v, want 'name string'", method.Parameters)
	}

	if method.ReturnType != "*ProductFilters" {
		t.Errorf("Method return type = %v, want '*ProductFilters'", method.ReturnType)
	}

	// Test method body contains expected elements
	expectedBodyParts := []string{
		"p.filters[ProductDBSchema.Name]",
		"repository.OperatorEqual",
		"name",
		"return p",
	}

	for _, part := range expectedBodyParts {
		if !strings.Contains(method.Body, part) {
			t.Errorf("Method body missing expected part: %s\nBody: %s", part, method.Body)
		}
	}

	// Test documentation
	if !strings.Contains(method.Documentation, "NameEq") {
		t.Errorf("Method documentation should contain method name")
	}
}

func TestMethodFactory_CreateFilterMethod_Variadic(t *testing.T) {
	factory := NewMethodFactory()

	field := domain.Field{
		Name:     "ID",
		TypeName: "int64",
		Type:     domain.FieldTypeNumeric,
	}

	method := factory.CreateFilterMethod("Product", field, repository.OperatorIn)

	// Test variadic parameters
	if method.Parameters != "iDs ...int64" {
		t.Errorf("Variadic method parameters = %v, want 'iDs ...int64'", method.Parameters)
	}

	if method.Name != "IDIn" {
		t.Errorf("Method name = %v, want IDIn", method.Name)
	}

	// Test body contains variadic parameter
	if !strings.Contains(method.Body, "iDs") {
		t.Errorf("Variadic method body should contain parameter name 'iDs'")
	}
}

func TestMethodFactory_CreateFilterMethod_Unary(t *testing.T) {
	factory := NewMethodFactory()

	field := domain.Field{
		Name:     "UpdatedAt",
		TypeName: "*time.Time",
		Type:     domain.FieldTypePointer,
	}

	method := factory.CreateFilterMethod("Product", field, repository.OperatorIsNull)

	// Test unary method has no parameters
	if method.Parameters != "" {
		t.Errorf("Unary method parameters = %v, want empty string", method.Parameters)
	}

	if method.Name != "UpdatedAtIsNull" {
		t.Errorf("Method name = %v, want UpdatedAtIsNull", method.Name)
	}

	// Test body contains nil value
	if !strings.Contains(method.Body, "Value:    nil") {
		t.Errorf("Unary method body should contain 'Value:    nil'")
	}
}

func TestMethodFactory_CreateUpdaterMethod(t *testing.T) {
	factory := NewMethodFactory()

	field := domain.Field{
		Name:     "Email",
		TypeName: "string",
		Type:     domain.FieldTypeString,
	}

	method := factory.CreateUpdaterMethod("Product", field)

	// Test updater method properties
	if method.Name != "SetEmail" {
		t.Errorf("Updater method name = %v, want SetEmail", method.Name)
	}

	if method.Receiver != "p *ProductUpdater" {
		t.Errorf("Updater method receiver = %v, want 'p *ProductUpdater'", method.Receiver)
	}

	if method.Parameters != "email string" {
		t.Errorf("Updater method parameters = %v, want 'email string'", method.Parameters)
	}

	if method.ReturnType != "*ProductUpdater" {
		t.Errorf("Updater method return type = %v, want '*ProductUpdater'", method.ReturnType)
	}

	// Test method body
	expectedBodyParts := []string{
		"p.fields[string(ProductDBSchema.Email)]",
		"email",
		"return p",
	}

	for _, part := range expectedBodyParts {
		if !strings.Contains(method.Body, part) {
			t.Errorf("Updater method body missing expected part: %s", part)
		}
	}
}

func TestMethodFactory_CreateOrderMethod(t *testing.T) {
	factory := NewMethodFactory()

	field := domain.Field{
		Name:     "CreatedAt",
		TypeName: "time.Time",
		Type:     domain.FieldTypeTime,
	}

	// Test ascending order method
	ascMethod := factory.CreateOrderMethod("Product", field, true)

	if ascMethod.Name != "OrderByCreatedAtAsc" {
		t.Errorf("Ascending order method name = %v, want OrderByCreatedAtAsc", ascMethod.Name)
	}

	if !strings.Contains(ascMethod.Body, `Direction: "asc"`) {
		t.Errorf("Ascending order method body should contain 'Direction: \"asc\"'")
	}

	// Test descending order method
	descMethod := factory.CreateOrderMethod("Product", field, false)

	if descMethod.Name != "OrderByCreatedAtDesc" {
		t.Errorf("Descending order method name = %v, want OrderByCreatedAtDesc", descMethod.Name)
	}

	if !strings.Contains(descMethod.Body, `Direction: "desc"`) {
		t.Errorf("Descending order method body should contain 'Direction: \"desc\"'")
	}

	// Test receiver type
	if ascMethod.Receiver != "p *ProductOptions" {
		t.Errorf("Order method receiver = %v, want 'p *ProductOptions'", ascMethod.Receiver)
	}

	// Test return type
	if ascMethod.ReturnType != "*ProductOptions" {
		t.Errorf("Order method return type = %v, want '*ProductOptions'", ascMethod.ReturnType)
	}
}

func TestMethodFactory_fieldNameToParamName(t *testing.T) {
	factory := NewMethodFactory()

	tests := []struct {
		name      string
		fieldName string
		expected  string
	}{
		{"normal field", "Name", "name"},
		{"empty field", "", "value"},
		{"single char", "A", "a"},
		{"keyword field", "Type", "typeValue"},
		{"func keyword", "Func", "funcValue"},
		{"import keyword", "Import", "importValue"},
		{"non-keyword", "Email", "email"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := factory.fieldNameToParamName(tt.fieldName)
			if result != tt.expected {
				t.Errorf("fieldNameToParamName(%v) = %v, want %v", tt.fieldName, result, tt.expected)
			}
		})
	}
}

func TestMethodFactory_OperatorHelpers(t *testing.T) {
	factory := NewMethodFactory()

	// Test unary operators
	unaryOps := []repository.Operator{
		repository.OperatorIsNull,
		repository.OperatorIsNotNull,
	}

	for _, op := range unaryOps {
		if !factory.isUnaryOperator(op) {
			t.Errorf("isUnaryOperator(%v) should return true", op)
		}
	}

	// Test variadic operators
	variadicOps := []repository.Operator{
		repository.OperatorIn,
		repository.OperatorNotIn,
	}

	for _, op := range variadicOps {
		if !factory.isVariadicOperator(op) {
			t.Errorf("isVariadicOperator(%v) should return true", op)
		}
	}

	// Test that other operators return false
	binaryOps := []repository.Operator{
		repository.OperatorEqual,
		repository.OperatorNotEqual,
		repository.OperatorLessThan,
		repository.OperatorGreaterThan,
	}

	for _, op := range binaryOps {
		if factory.isUnaryOperator(op) {
			t.Errorf("isUnaryOperator(%v) should return false", op)
		}
		if factory.isVariadicOperator(op) {
			t.Errorf("isVariadicOperator(%v) should return false", op)
		}
	}
}

// Generic type specific tests
func TestMethodFactory_GenericTypeHandling(t *testing.T) {
	factory := NewMethodFactory()

	tests := []struct {
		name      string
		field     domain.Field
		operator  repository.Operator
		wantName  string
		wantParam string
	}{
		{
			name: "generic pointer field",
			field: domain.Field{
				Name:     "GenericPtr",
				TypeName: "*T",
				Type:     domain.FieldTypePointer,
			},
			operator:  repository.OperatorEqual,
			wantName:  "GenericPtrEq",
			wantParam: "genericPtr *T",
		},
		{
			name: "generic constraint field",
			field: domain.Field{
				Name:     "Constraint",
				TypeName: "T",
				Type:     domain.FieldTypeString, // Assuming T is string-like
			},
			operator:  repository.OperatorLike,
			wantName:  "ConstraintLike",
			wantParam: "constraint T",
		},
		{
			name: "generic variadic field",
			field: domain.Field{
				Name:     "IDs",
				TypeName: "T",
				Type:     domain.FieldTypeNumeric,
			},
			operator:  repository.OperatorIn,
			wantName:  "IDsIn",
			wantParam: "iDss ...T", // Note: this might need refinement
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := factory.CreateFilterMethod("Generic", tt.field, tt.operator)

			if method.Name != tt.wantName {
				t.Errorf("Generic method name = %v, want %v", method.Name, tt.wantName)
			}

			if method.Parameters != tt.wantParam {
				t.Errorf("Generic method parameters = %v, want %v", method.Parameters, tt.wantParam)
			}

			// Test that generic types are properly handled in body
			// Note: Method body may not directly reference the type name, but should contain parameter reference
			if method.Body == "" {
				t.Error("Method body should not be empty")
			}
		})
	}
}

func TestMethodFactory_ComplexGenericScenarios(t *testing.T) {
	factory := NewMethodFactory()

	// Test with complex generic types
	complexField := domain.Field{
		Name:     "ComplexGeneric",
		TypeName: "map[K][]V",            // This shouldn't be filterable in practice
		Type:     domain.FieldTypeString, // But if it were...
		GoType:   "map[K][]V",
	}

	method := factory.CreateFilterMethod("Container", complexField, repository.OperatorEqual)

	// Test that complex generics are handled without crashing
	if method.Name == "" {
		t.Error("Complex generic method should have a name")
	}

	if method.Body == "" {
		t.Error("Complex generic method should have a body")
	}

	// Test method contains expected structure
	expectedParts := []string{
		"c.filters[ContainerDBSchema.ComplexGeneric]",
		"repository.OperatorEqual",
		"return c",
	}

	for _, part := range expectedParts {
		if !strings.Contains(method.Body, part) {
			t.Errorf("Complex generic method body missing: %s", part)
		}
	}
}
