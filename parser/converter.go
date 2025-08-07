package parser

import (
	"go/ast"
	"strings"

	"github.com/dchlong/querybuilder/domain"
	"github.com/dchlong/querybuilder/field"
)

// Converter converts from existing parser types to clean domain types
type Converter struct {
	fieldInfoGenerator *field.InfoGenerator
}

// NewConverter creates a new converter
func NewConverter(fieldInfoGenerator *field.InfoGenerator) *Converter {
	return &Converter{
		fieldInfoGenerator: fieldInfoGenerator,
	}
}

// ConvertStruct converts a ParsedStruct to domain.Struct.
// Only includes fields that can be processed by the field info generator.
func (c *Converter) ConvertStruct(s ParsedStruct) domain.Struct {
	domainStruct := domain.Struct{
		Name:        s.TypeName,
		PackageName: "", // Will be set by caller
		Fields:      make([]domain.Field, 0, len(s.Fields)),
	}

	for _, f := range s.Fields {
		fieldInfo := c.fieldInfoGenerator.GenFieldInfo(f)
		if fieldInfo != nil {
			domainField := c.convertField(*fieldInfo)
			domainStruct.Fields = append(domainStruct.Fields, domainField)
		}
	}

	return domainStruct
}

// convertField converts field.Info to domain.Field.
// Maps all relevant field metadata from the parsed field info.
func (c *Converter) convertField(fi field.Info) domain.Field {
	return domain.Field{
		Name:     fi.Name,
		DBName:   fi.DBName,
		Type:     c.convertFieldType(fi),
		TypeName: fi.TypeName,
		GoType:   fi.GetTypeName(), // Use full type name including generics
	}
}

// convertFieldType converts field.Info to domain.FieldType.
// Uses a priority-based approach where more specific types take precedence.
func (c *Converter) convertFieldType(fi field.Info) domain.FieldType {
	// Handle special types first (most specific)
	if fi.IsTime {
		return domain.FieldTypeTime
	}

	// Handle container types
	if fi.IsSlice {
		return domain.FieldTypeSlice
	}
	if fi.IsMap {
		return domain.FieldTypeMap
	}
	if fi.IsStruct {
		return domain.FieldTypeStruct
	}

	// Handle pointer types
	if fi.IsPointer {
		return domain.FieldTypePointer
	}

	// Handle basic types
	if fi.IsString {
		return domain.FieldTypeString
	}
	if fi.IsNumeric {
		return domain.FieldTypeNumeric
	}

	// Check for boolean type (fallback to string matching)
	if c.isBooleanType(fi.TypeName) {
		return domain.FieldTypeBool
	}

	return domain.FieldTypeUnknown
}

// isBooleanType checks if a type name represents a boolean type.
func (c *Converter) isBooleanType(typeName string) bool {
	lowerTypeName := strings.ToLower(typeName)
	return strings.Contains(lowerTypeName, "bool")
}

// ShouldGenerateQueryBuilder checks if struct should have querybuilder generated
func (c *Converter) ShouldGenerateQueryBuilder(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}

	for _, comment := range doc.List {
		text := strings.TrimSpace(comment.Text)

		// Check for both old and new format annotations
		if c.hasQueryBuilderAnnotation(text) {
			return true
		}
	}

	return false
}

// hasQueryBuilderAnnotation checks if a comment contains querybuilder annotation.
// Supports multiple annotation formats for flexibility.
func (c *Converter) hasQueryBuilderAnnotation(comment string) bool {
	// Clean up comment text
	text := c.cleanCommentText(comment)
	if text == "" {
		return false
	}

	// Check against supported annotation formats
	annotations := []string{
		"gen:querybuilder",
		"@querybuilder",
		"+querybuilder",
		"//go:generate querybuilder",
	}

	lowerText := strings.ToLower(text)
	for _, annotation := range annotations {
		if strings.Contains(lowerText, strings.ToLower(annotation)) {
			return true
		}
	}

	return false
}

// cleanCommentText removes comment prefixes/suffixes and normalizes whitespace.
func (c *Converter) cleanCommentText(comment string) string {
	text := strings.TrimSpace(comment)
	text = strings.TrimPrefix(text, "//")
	text = strings.TrimPrefix(text, "/*")
	text = strings.TrimSuffix(text, "*/")
	return strings.TrimSpace(text)
}
