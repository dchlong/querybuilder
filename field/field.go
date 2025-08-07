package field

import (
	"fmt"
	"go/types"
	"reflect"
	"strings"

	"gorm.io/gorm/schema"
)

// TimeTypePattern represents a pattern for detecting time-related types.
type TimeTypePattern struct {
	Pattern   string // Type name pattern (exact match)
	IsNumeric bool   // Whether this time type behaves like numeric for filtering
}

// DefaultTimeTypes contains the built-in time type patterns.
var DefaultTimeTypes = []TimeTypePattern{
	{Pattern: "time.Time", IsNumeric: true},
	{Pattern: "datatypes.Date", IsNumeric: true},
	{Pattern: "datatypes.Time", IsNumeric: true},
	{Pattern: "datatypes.DateTime", IsNumeric: true},
	{Pattern: "sql.NullTime", IsNumeric: true},
	{Pattern: "pq.NullTime", IsNumeric: true},
}

// BaseInfo contains basic information about a struct field.
type BaseInfo struct {
	Name     string // Go field name
	DBName   string // Database column name
	TypeName string // Go type name

	// Type classification flags
	IsStruct  bool // Is a struct type
	IsNumeric bool // Is a numeric type (int, float, etc.)
	IsTime    bool // Is a time-related type
	IsString  bool // Is a string type
	IsSlice   bool // Is a slice type
	IsMap     bool // Is a map type
}

// Info contains comprehensive field information including type metadata.
type Info struct {
	BaseInfo           // Embedded base information
	pointed  *BaseInfo // Information about pointed-to type (for pointers)
	typeArgs []*Info   // Type arguments (for generics)

	// Enhanced type flags
	IsPointer bool // Is a pointer type
	IsGeneric bool // Has generic type parameters
}

func (fi Info) TypeArgs() []*Info {
	return fi.typeArgs
}

// GetTypeName returns the full type name including generics and pointer information.
func (fi Info) GetTypeName() string {
	name := fi.TypeName

	if fi.IsPointer {
		name = "*" + fi.pointed.TypeName
	}

	if fi.IsGeneric && len(fi.typeArgs) > 0 {
		var args []string
		for _, arg := range fi.typeArgs {
			args = append(args, arg.GetTypeName())
		}
		name = fmt.Sprintf("%s<%s>", name, strings.Join(args, ", "))
	}

	return name
}

// GetPointed returns the pointed-to type for pointer fields.
// Returns zero value if field is not a pointer.
func (fi Info) GetPointed() Info {
	if !fi.IsPointer || fi.pointed == nil {
		return Info{}
	}
	return Info{
		BaseInfo: *fi.pointed,
	}
}

// InfoGenerator generates field information from Go types.
// Contains package context for proper type name resolution and configurable time type detection.
type InfoGenerator struct {
	pkg       *types.Package    // Package context for type resolution
	timeTypes []TimeTypePattern // Configurable time type patterns
}

// Field interface defines the contract for struct field information.
// Provides access to field metadata needed for code generation.
type Field interface {
	Name() string           // Go field name
	Type() types.Type       // Go type information
	Tag() reflect.StructTag // Struct tags (db:"", json:"", etc.)
}

type field struct {
	name string
	typ  types.Type
	tag  reflect.StructTag
}

func (f field) Name() string {
	return f.name
}

func (f field) Type() types.Type {
	return f.typ
}

func (f field) Tag() reflect.StructTag {
	return f.tag
}

// NewInfoGenerator creates a new InfoGenerator with default time type patterns.
func NewInfoGenerator(pkg *types.Package) *InfoGenerator {
	return &InfoGenerator{
		pkg:       pkg,
		timeTypes: DefaultTimeTypes,
	}
}

// NewInfoGeneratorWithTimeTypes creates a new InfoGenerator with custom time type patterns.
func NewInfoGeneratorWithTimeTypes(pkg *types.Package, timeTypes []TimeTypePattern) *InfoGenerator {
	return &InfoGenerator{
		pkg:       pkg,
		timeTypes: timeTypes,
	}
}

// AddTimeType adds a custom time type pattern to the generator.
func (g *InfoGenerator) AddTimeType(pattern string, isNumeric bool) {
	g.timeTypes = append(g.timeTypes, TimeTypePattern{
		Pattern:   pattern,
		IsNumeric: isNumeric,
	})
}

// matchTimeType checks if a type name matches any configured time type patterns.
// Returns the matching pattern or nil if no match is found.
func (g *InfoGenerator) matchTimeType(typeName string) *TimeTypePattern {
	for i := range g.timeTypes {
		if g.timeTypes[i].Pattern == typeName {
			return &g.timeTypes[i]
		}
	}
	return nil
}

// getOriginalTypeName returns the properly qualified type name.
// Returns unqualified name for types in the same package, qualified name for imports.
func (g InfoGenerator) getOriginalTypeName(t *types.Named) string {
	obj := t.Obj()
	if obj.Pkg() == g.pkg {
		// Same package - use unqualified name
		return obj.Name()
	}

	// Different package - use qualified name
	return fmt.Sprintf("%s.%s", obj.Pkg().Name(), obj.Name())
}

// parseTagSetting parses struct tag settings for database field configuration.
// Based on GORM's tag parsing implementation.
// Supports both 'sql' and 'gorm' tag formats.
func parseTagSetting(tags reflect.StructTag) map[string]string {
	setting := make(map[string]string)

	// Process both sql and gorm tags
	tagSources := []string{tags.Get("sql"), tags.Get("gorm")}
	for _, tagSource := range tagSources {
		if tagSource == "" {
			continue
		}

		// Parse individual tag components
		tagParts := strings.Split(tagSource, ";")
		for _, tagPart := range tagParts {
			if tagPart == "" {
				continue
			}

			// Split key:value pairs
			keyValue := strings.Split(tagPart, ":")
			key := strings.TrimSpace(strings.ToUpper(keyValue[0]))

			if len(keyValue) >= 2 {
				setting[key] = strings.Join(keyValue[1:], ":")
			} else {
				setting[key] = key
			}
		}
	}

	return setting
}

// GenFieldInfo generates field information for code generation.
// Returns nil if the field should be skipped (e.g., tagged with "-").
func (g InfoGenerator) GenFieldInfo(f Field) *Info {
	// Check if field should be skipped
	if g.shouldSkipField(f) {
		return nil
	}

	// Create base field information
	baseInfo := g.createBaseInfo(f)

	// Handle time types using configurable patterns
	if timePattern := g.matchTimeType(baseInfo.TypeName); timePattern != nil {
		return g.createTimeFieldInfo(baseInfo, *timePattern)
	}

	// Process field based on its type
	return g.processFieldType(f, baseInfo)
}

// shouldSkipField checks if a field should be skipped based on its tags.
func (g InfoGenerator) shouldSkipField(f Field) bool {
	tagSetting := parseTagSetting(f.Tag())
	return tagSetting["-"] != ""
}

// createBaseInfo creates the base field information structure.
func (g InfoGenerator) createBaseInfo(f Field) BaseInfo {
	tagSetting := parseTagSetting(f.Tag())

	dbName := schema.NamingStrategy{}.ColumnName("", f.Name())
	if dbColName := tagSetting["COLUMN"]; dbColName != "" {
		dbName = dbColName
	}

	return BaseInfo{
		Name:     f.Name(),
		TypeName: f.Type().String(),
		DBName:   dbName,
	}
}

// createTimeFieldInfo creates field info for time-related fields using the matched pattern.
func (g InfoGenerator) createTimeFieldInfo(baseInfo BaseInfo, pattern TimeTypePattern) *Info {
	baseInfo.IsTime = true
	baseInfo.IsNumeric = pattern.IsNumeric
	return &Info{BaseInfo: baseInfo}
}

// processFieldType processes a field based on its Go type.
func (g InfoGenerator) processFieldType(f Field, baseInfo BaseInfo) *Info {
	switch t := f.Type().(type) {
	case *types.Basic:
		return g.processBasicType(t, baseInfo)
	case *types.Slice:
		return g.processSliceType(baseInfo)
	case *types.Named:
		return g.processNamedType(f, t)
	case *types.Struct:
		return g.processStructType(baseInfo)
	case *types.Pointer:
		return g.processPointerType(f, t, baseInfo)
	case *types.Map:
		return g.processMapType(baseInfo)
	default:
		// Unknown type - no filtering needed
		return nil
	}
}

// processBasicType handles basic Go types (string, int, bool, etc.).
func (g InfoGenerator) processBasicType(t *types.Basic, baseInfo BaseInfo) *Info {
	baseInfo.IsString = t.Info()&types.IsString != 0
	baseInfo.IsNumeric = t.Info()&types.IsNumeric != 0
	return &Info{BaseInfo: baseInfo}
}

// processSliceType handles slice types.
func (g InfoGenerator) processSliceType(baseInfo BaseInfo) *Info {
	baseInfo.IsSlice = true
	return &Info{BaseInfo: baseInfo}
}

// processStructType handles struct types.
func (g InfoGenerator) processStructType(baseInfo BaseInfo) *Info {
	baseInfo.IsStruct = true
	return &Info{BaseInfo: baseInfo}
}

// processMapType handles map types.
func (g InfoGenerator) processMapType(baseInfo BaseInfo) *Info {
	baseInfo.IsMap = true
	return &Info{BaseInfo: baseInfo}
}

// processNamedType handles named types (custom types, generics).
func (g InfoGenerator) processNamedType(f Field, t *types.Named) *Info {
	// Recursively process the underlying type
	r := g.GenFieldInfo(field{
		name: f.Name(),
		typ:  t.Underlying(),
		tag:  f.Tag(),
	})

	if r == nil {
		return nil
	}

	// Set the original type name
	r.TypeName = g.getOriginalTypeName(t)

	// Handle time types using configurable patterns
	if timePattern := g.matchTimeType(r.TypeName); timePattern != nil {
		r.IsTime = true
		r.IsStruct = false
		r.IsNumeric = timePattern.IsNumeric
	}

	// Handle generic types
	if t.TypeArgs().Len() > 0 {
		r.TypeName = g.processGenericType(f, t, r.TypeName)
		r.IsGeneric = true
	}

	return r
}

// processGenericType handles generic type arguments.
func (g InfoGenerator) processGenericType(f Field, t *types.Named, baseName string) string {
	var typeArgs []string
	for i := 0; i < t.TypeArgs().Len(); i++ {
		typeArg := t.TypeArgs().At(i)
		argInfo := g.GenFieldInfo(field{
			name: f.Name(),
			typ:  typeArg,
			tag:  f.Tag(),
		})

		if argInfo != nil {
			typeArgs = append(typeArgs, argInfo.TypeName)
		}
	}

	return fmt.Sprintf("%s[%s]", baseName, strings.Join(typeArgs, ", "))
}

// processPointerType handles pointer types.
func (g InfoGenerator) processPointerType(f Field, t *types.Pointer, baseInfo BaseInfo) *Info {
	pointedField := g.GenFieldInfo(field{
		name: f.Name(),
		typ:  t.Elem(),
		tag:  f.Tag(),
	})

	if pointedField == nil {
		return nil
	}

	return &Info{
		BaseInfo: BaseInfo{
			Name:     baseInfo.Name,
			TypeName: fmt.Sprintf("*%s", pointedField.TypeName),
			DBName:   baseInfo.DBName,
		},
		IsPointer: true,
		pointed:   &pointedField.BaseInfo,
	}
}
