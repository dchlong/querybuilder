package querybuilder

import (
	"context"
	"fmt"

	"github.com/dchlong/querybuilder/builder"
	"github.com/dchlong/querybuilder/domain"
	"github.com/dchlong/querybuilder/field"
	"github.com/dchlong/querybuilder/parser"
	"github.com/dchlong/querybuilder/repository"
)

// Generator provides a clean, readable API for querybuilder generation
type Generator struct {
	structsParser *parser.Structs
	converter     *parser.Converter
	generator     *builder.Generator
}

// NewQueryBuilderGenerator creates a new querybuilder generator
func NewQueryBuilderGenerator(structsParser *parser.Structs) *Generator {
	fieldInfoGen := field.NewInfoGenerator(nil) // Will be set when parsing

	return &Generator{
		structsParser: structsParser,
		converter:     parser.NewConverter(fieldInfoGen),
		generator:     builder.NewGenerator(),
	}
}

// Generate generates querybuilder code for a Go source file
func (g *Generator) Generate(ctx context.Context, inputFile, outputFile, suffix string) error {
	// Validate inputs
	if err := g.validateInputs(inputFile, outputFile); err != nil {
		return fmt.Errorf("invalid inputs: %w", err)
	}

	// Parse the input file
	parsedFile, err := g.structsParser.ParseFile(ctx, inputFile)
	if err != nil {
		return fmt.Errorf("%w %s: %w", repository.ErrParseFile, inputFile, err)
	}

	// Update field info generator with parsed types
	fieldInfoGen := field.NewInfoGenerator(parsedFile.Types)
	g.converter = parser.NewConverter(fieldInfoGen)

	// Convert to domain structs
	var domainStructs []domain.Struct
	for _, parsedStruct := range parsedFile.Structs {
		if !g.converter.ShouldGenerateQueryBuilder(parsedStruct.Doc) {
			continue
		}

		// Apply suffix if provided
		structWithSuffix := parsedStruct
		if suffix != "" {
			structWithSuffix.TypeName = parsedStruct.TypeName + suffix
		}

		domainStruct := g.converter.ConvertStruct(structWithSuffix)
		domainStruct.PackageName = parsedFile.PackageName
		domainStructs = append(domainStructs, domainStruct)
	}

	// Check if we have any structs to generate
	if len(domainStructs) == 0 {
		return fmt.Errorf("%w in %s", repository.ErrNoAnnotatedStructs, inputFile)
	}

	// Generate the code
	if err := g.generator.GenerateFile(ctx, domainStructs, parsedFile.PackageName, outputFile); err != nil {
		return fmt.Errorf("failed to generate querybuilder code: %w", err)
	}

	return nil
}

// GenerateInMemory generates querybuilder code and returns it as bytes
func (g *Generator) GenerateInMemory(ctx context.Context, inputFile, suffix string) ([]byte, string, error) {
	// Parse the input file
	parsedFile, err := g.structsParser.ParseFile(ctx, inputFile)
	if err != nil {
		return nil, "", fmt.Errorf("%w %s: %w", repository.ErrParseFile, inputFile, err)
	}

	// Update field info generator with parsed types
	fieldInfoGen := field.NewInfoGenerator(parsedFile.Types)
	g.converter = parser.NewConverter(fieldInfoGen)

	// Convert to domain structs
	var domainStructs []domain.Struct
	for _, parsedStruct := range parsedFile.Structs {
		if !g.converter.ShouldGenerateQueryBuilder(parsedStruct.Doc) {
			continue
		}

		// Apply suffix if provided
		structWithSuffix := parsedStruct
		if suffix != "" {
			structWithSuffix.TypeName = parsedStruct.TypeName + suffix
		}

		domainStruct := g.converter.ConvertStruct(structWithSuffix)
		domainStruct.PackageName = parsedFile.PackageName
		domainStructs = append(domainStructs, domainStruct)
	}

	// Check if we have any structs to generate
	if len(domainStructs) == 0 {
		return nil, "", fmt.Errorf("%w in %s", repository.ErrNoAnnotatedStructs, inputFile)
	}

	// Generate the code
	code, err := g.generator.GenerateCode(ctx, domainStructs, parsedFile.PackageName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate querybuilder code: %w", err)
	}

	return code, parsedFile.PackageName, nil
}

// validateInputs validates the input parameters
func (g *Generator) validateInputs(inputFile, outputFile string) error {
	if inputFile == "" {
		return repository.ErrEmptyInputFile
	}
	if outputFile == "" {
		return repository.ErrEmptyOutputFile
	}
	if g.structsParser == nil {
		return repository.ErrNilParser
	}
	return nil
}

// GetSupportedFieldTypes returns the field types supported by the querybuilder
func (g *Generator) GetSupportedFieldTypes() []string {
	return []string{
		domain.FieldTypeString.String(),
		domain.FieldTypeNumeric.String(),
		domain.FieldTypeTime.String(),
		domain.FieldTypeBool.String(),
		domain.FieldTypePointer.String(),
	}
}

// GetUnsupportedFieldTypes returns the field types not supported by the querybuilder
func (g *Generator) GetUnsupportedFieldTypes() []string {
	return []string{
		domain.FieldTypeSlice.String(),
		domain.FieldTypeStruct.String(),
		domain.FieldTypeMap.String(),
	}
}
