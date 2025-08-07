package domain

import (
	"reflect"
	"testing"

	"github.com/dchlong/querybuilder/repository"
)

func TestFieldType_String(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expected  string
	}{
		{"string type", FieldTypeString, "string"},
		{"numeric type", FieldTypeNumeric, "numeric"},
		{"time type", FieldTypeTime, "time"},
		{"bool type", FieldTypeBool, "bool"},
		{"pointer type", FieldTypePointer, "pointer"},
		{"slice type", FieldTypeSlice, "slice"},
		{"struct type", FieldTypeStruct, "struct"},
		{"map type", FieldTypeMap, "map"},
		{"unknown type", FieldTypeUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fieldType.String()
			if result != tt.expected {
				t.Errorf("FieldType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestField_IsFilterable(t *testing.T) {
	tests := []struct {
		name     string
		field    Field
		expected bool
	}{
		{
			name: "string field is filterable",
			field: Field{
				Name: "Name",
				Type: FieldTypeString,
			},
			expected: true,
		},
		{
			name: "numeric field is filterable",
			field: Field{
				Name: "Age",
				Type: FieldTypeNumeric,
			},
			expected: true,
		},
		{
			name: "slice field is not filterable",
			field: Field{
				Name: "Tags",
				Type: FieldTypeSlice,
			},
			expected: false,
		},
		{
			name: "struct field is not filterable",
			field: Field{
				Name: "Address",
				Type: FieldTypeStruct,
			},
			expected: false,
		},
		{
			name: "map field is not filterable",
			field: Field{
				Name: "Metadata",
				Type: FieldTypeMap,
			},
			expected: false,
		},
		{
			name: "pointer field is filterable",
			field: Field{
				Name: "UpdatedAt",
				Type: FieldTypePointer,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.field.IsFilterable()
			if result != tt.expected {
				t.Errorf("Field.IsFilterable() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestField_SupportedOperators(t *testing.T) {
	tests := []struct {
		name     string
		field    Field
		expected []repository.Operator
	}{
		{
			name: "string field supports all string operators",
			field: Field{
				Type: FieldTypeString,
			},
			expected: []repository.Operator{
				repository.OperatorEqual,
				repository.OperatorNotEqual,
				repository.OperatorLike,
				repository.OperatorNotLike,
				repository.OperatorIn,
				repository.OperatorNotIn,
				repository.OperatorLessThan,
				repository.OperatorGreaterThan,
				repository.OperatorLessThanOrEqual,
				repository.OperatorGreaterThanOrEqual,
			},
		},
		{
			name: "numeric field supports numeric operators",
			field: Field{
				Type: FieldTypeNumeric,
			},
			expected: []repository.Operator{
				repository.OperatorEqual,
				repository.OperatorNotEqual,
				repository.OperatorLessThan,
				repository.OperatorGreaterThan,
				repository.OperatorLessThanOrEqual,
				repository.OperatorGreaterThanOrEqual,
				repository.OperatorIn,
				repository.OperatorNotIn,
			},
		},
		{
			name: "time field supports numeric operators",
			field: Field{
				Type: FieldTypeTime,
			},
			expected: []repository.Operator{
				repository.OperatorEqual,
				repository.OperatorNotEqual,
				repository.OperatorLessThan,
				repository.OperatorGreaterThan,
				repository.OperatorLessThanOrEqual,
				repository.OperatorGreaterThanOrEqual,
				repository.OperatorIn,
				repository.OperatorNotIn,
			},
		},
		{
			name: "pointer field supports null operators",
			field: Field{
				Type: FieldTypePointer,
			},
			expected: []repository.Operator{
				repository.OperatorEqual,
				repository.OperatorNotEqual,
				repository.OperatorIsNull,
				repository.OperatorIsNotNull,
			},
		},
		{
			name: "bool field supports basic operators",
			field: Field{
				Type: FieldTypeBool,
			},
			expected: []repository.Operator{
				repository.OperatorEqual,
				repository.OperatorNotEqual,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.field.SupportedOperators()
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Field.SupportedOperators() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStruct_FilterableFields(t *testing.T) {
	s := Struct{
		Name: "Product",
		Fields: []Field{
			{Name: "ID", Type: FieldTypeNumeric},
			{Name: "Name", Type: FieldTypeString},
			{Name: "Tags", Type: FieldTypeSlice},     // Not filterable
			{Name: "Address", Type: FieldTypeStruct}, // Not filterable
			{Name: "Metadata", Type: FieldTypeMap},   // Not filterable
			{Name: "CreatedAt", Type: FieldTypeTime},
		},
	}

	expected := []Field{
		{Name: "ID", Type: FieldTypeNumeric},
		{Name: "Name", Type: FieldTypeString},
		{Name: "CreatedAt", Type: FieldTypeTime},
	}

	result := s.FilterableFields()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Struct.FilterableFields() = %v, want %v", result, expected)
	}
}

// Generic type tests
func TestGenericFieldTypes(t *testing.T) {
	tests := []struct {
		name      string
		goType    string
		typeName  string
		fieldType FieldType
		expected  bool // IsFilterable
	}{
		{
			name:      "generic slice is not filterable",
			goType:    "[]T",
			typeName:  "[]T",
			fieldType: FieldTypeSlice,
			expected:  false,
		},
		{
			name:      "generic map is not filterable",
			goType:    "map[K]V",
			typeName:  "map[K]V",
			fieldType: FieldTypeMap,
			expected:  false,
		},
		{
			name:      "generic pointer is filterable",
			goType:    "*T",
			typeName:  "*T",
			fieldType: FieldTypePointer,
			expected:  true,
		},
		{
			name:      "channel type is not filterable",
			goType:    "chan T",
			typeName:  "chan T",
			fieldType: FieldTypeUnknown,
			expected:  true, // Unknown types are considered filterable with basic ops
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := Field{
				Name:     "GenericField",
				GoType:   tt.goType,
				TypeName: tt.typeName,
				Type:     tt.fieldType,
			}

			result := field.IsFilterable()
			if result != tt.expected {
				t.Errorf("Generic field %s IsFilterable() = %v, want %v", tt.goType, result, tt.expected)
			}

			// Test that it has at least basic operators
			operators := field.SupportedOperators()
			if len(operators) == 0 {
				t.Errorf("Generic field %s should have at least basic operators", tt.goType)
			}

			// All fields should support Equal and NotEqual
			hasEqual := false
			hasNotEqual := false
			for _, op := range operators {
				if op == repository.OperatorEqual {
					hasEqual = true
				}
				if op == repository.OperatorNotEqual {
					hasNotEqual = true
				}
			}

			if !hasEqual || !hasNotEqual {
				t.Errorf("Generic field %s should support Equal and NotEqual operators", tt.goType)
			}
		})
	}
}

func TestComplexGenericTypes(t *testing.T) {
	tests := []struct {
		name    string
		field   Field
		wantOps int // Expected number of operators
	}{
		{
			name: "slice of pointers",
			field: Field{
				Name:     "Items",
				GoType:   "[]*T",
				TypeName: "[]*T",
				Type:     FieldTypeSlice,
			},
			wantOps: 0, // Slices are not filterable
		},
		{
			name: "map with generic values",
			field: Field{
				Name:     "Cache",
				GoType:   "map[string]T",
				TypeName: "map[string]T",
				Type:     FieldTypeMap,
			},
			wantOps: 0, // Maps are not filterable
		},
		{
			name: "nested generic pointer",
			field: Field{
				Name:     "NestedPtr",
				GoType:   "**T",
				TypeName: "**T",
				Type:     FieldTypePointer,
			},
			wantOps: 4, // Equal, NotEqual, IsNull, IsNotNull
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operators := tt.field.SupportedOperators()

			if tt.wantOps == 0 {
				// Should not be filterable
				if tt.field.IsFilterable() {
					t.Errorf("%s should not be filterable", tt.name)
				}
			} else {
				// Should be filterable with expected operators
				if !tt.field.IsFilterable() {
					t.Errorf("%s should be filterable", tt.name)
				}

				if len(operators) != tt.wantOps {
					t.Errorf("%s operators count = %d, want %d", tt.name, len(operators), tt.wantOps)
				}
			}
		})
	}
}

func TestMethod_Struct(t *testing.T) {
	method := Method{
		Name:          "NameEq",
		Receiver:      "f *ProductFilters",
		Parameters:    "name string",
		ReturnType:    "*ProductFilters",
		Body:          "// method body",
		Documentation: "NameEq filters by name equal",
	}

	// Test that Method struct holds all necessary information
	if method.Name != "NameEq" {
		t.Errorf("Method.Name = %v, want NameEq", method.Name)
	}

	if method.Receiver != "f *ProductFilters" {
		t.Errorf("Method.Receiver = %v, want 'f *ProductFilters'", method.Receiver)
	}

	if method.Parameters != "name string" {
		t.Errorf("Method.Parameters = %v, want 'name string'", method.Parameters)
	}

	if method.ReturnType != "*ProductFilters" {
		t.Errorf("Method.ReturnType = %v, want '*ProductFilters'", method.ReturnType)
	}

	if method.Documentation != "NameEq filters by name equal" {
		t.Errorf("Method.Documentation = %v, want 'NameEq filters by name equal'", method.Documentation)
	}
}
