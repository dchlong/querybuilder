package domain

import "github.com/dchlong/querybuilder/repository"

// FieldType represents the type classification of a struct field
type FieldType int

const (
	FieldTypeUnknown FieldType = iota
	FieldTypeString
	FieldTypeNumeric
	FieldTypeTime
	FieldTypeBool
	FieldTypePointer
	FieldTypeSlice
	FieldTypeStruct
	FieldTypeMap
)

// String returns the string representation of FieldType
func (ft FieldType) String() string {
	switch ft {
	case FieldTypeString:
		return "string"
	case FieldTypeNumeric:
		return "numeric"
	case FieldTypeTime:
		return "time"
	case FieldTypeBool:
		return "bool"
	case FieldTypePointer:
		return "pointer"
	case FieldTypeSlice:
		return "slice"
	case FieldTypeStruct:
		return "struct"
	case FieldTypeMap:
		return "map"
	default:
		return "unknown"
	}
}

// Field represents a struct field with its metadata
type Field struct {
	Name     string    // Go field name
	DBName   string    // Database column name
	Type     FieldType // Field type classification
	TypeName string    // Go type name
	GoType   string    // Full Go type (e.g., "*time.Time")
}

// IsFilterable returns true if the field can be used in filters
func (f Field) IsFilterable() bool {
	return f.Type != FieldTypeSlice && f.Type != FieldTypeStruct && f.Type != FieldTypeMap
}

// SupportedOperators returns the operators supported by this field type
func (f Field) SupportedOperators() []repository.Operator {
	base := []repository.Operator{
		repository.OperatorEqual,
		repository.OperatorNotEqual,
	}

	switch f.Type {
	case FieldTypeString:
		return append(base,
			repository.OperatorLike,
			repository.OperatorNotLike,
			repository.OperatorIn,
			repository.OperatorNotIn,
			repository.OperatorLessThan,
			repository.OperatorGreaterThan,
			repository.OperatorLessThanOrEqual,
			repository.OperatorGreaterThanOrEqual,
		)
	case FieldTypeNumeric, FieldTypeTime:
		return append(base,
			repository.OperatorLessThan,
			repository.OperatorGreaterThan,
			repository.OperatorLessThanOrEqual,
			repository.OperatorGreaterThanOrEqual,
			repository.OperatorIn,
			repository.OperatorNotIn,
		)
	case FieldTypePointer:
		return append(base,
			repository.OperatorIsNull,
			repository.OperatorIsNotNull,
		)
	default:
		return base
	}
}

// Struct represents a Go struct with querybuilder generation metadata
type Struct struct {
	Name        string  // Go struct name
	PackageName string  // Package name
	Fields      []Field // Struct fields
}

// FilterableFields returns only the fields that can be used in filters
func (s Struct) FilterableFields() []Field {
	var filterable []Field
	for _, field := range s.Fields {
		if field.IsFilterable() {
			filterable = append(filterable, field)
		}
	}
	return filterable
}

// Method represents a generated method
type Method struct {
	Name          string // Method name
	Receiver      string // Receiver type and name
	Parameters    string // Parameter list
	ReturnType    string // Return type
	Body          string // Method body
	Documentation string // Method documentation
}
