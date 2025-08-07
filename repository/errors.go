package repository

import "errors"

// Common errors used throughout the querybuilder package.
// These static errors replace dynamic error creation for better performance,
// consistency, and easier error handling/testing.

// Input validation errors
var (
	// ErrEmptyInputFile indicates that an input file path was not provided
	ErrEmptyInputFile = errors.New("input file path cannot be empty")

	// ErrEmptyOutputFile indicates that an output file path was not provided
	ErrEmptyOutputFile = errors.New("output file path cannot be empty")

	// ErrNilParser indicates that a nil parser was provided to the generator
	ErrNilParser = errors.New("structs parser cannot be nil")

	// ErrInputFileNotFound indicates that the specified input file does not exist
	ErrInputFileNotFound = errors.New("input file does not exist")

	// ErrNoGoFiles indicates that no Go files were found in the specified directory
	ErrNoGoFiles = errors.New("no Go files found in directory")

	// ErrUnknownOperator indicates that an unknown operator was used in a filter
	ErrUnknownOperator = errors.New("unknown operator in filter")
)

// Generation errors
var (
	// ErrNoStructsProvided indicates that no structs were provided for generation
	ErrNoStructsProvided = errors.New("no structs provided for generation")

	// ErrNoAnnotatedStructs indicates that no structs with querybuilder annotations were found
	ErrNoAnnotatedStructs = errors.New("no structs with querybuilder annotations found")
)

// Repository operation errors
var (
	// ErrNoRecordsProvided indicates that no records were provided for a batch operation
	ErrNoRecordsProvided = errors.New("no records provided for creation")

	// ErrEmptyFieldName indicates that a filter has an empty field name
	ErrEmptyFieldName = errors.New("empty field name in filter")
)

// Template and formatting errors
var (
	// ErrTemplateExecution indicates that template execution failed
	ErrTemplateExecution = errors.New("failed to execute template")

	// ErrCodeFormatting indicates that code formatting failed
	ErrCodeFormatting = errors.New("failed to format generated code")
)

// File operation errors
var (
	// ErrCreateOutputDir indicates that the output directory could not be created
	ErrCreateOutputDir = errors.New("failed to create output directory")

	// ErrWriteGeneratedCode indicates that generated code could not be written to file
	ErrWriteGeneratedCode = errors.New("failed to write generated code")
)

// Parser errors
var (
	// ErrParseFile indicates that a file could not be parsed
	ErrParseFile = errors.New("failed to parse file")

	// ErrLoadPackage indicates that a package could not be loaded
	ErrLoadPackage = errors.New("failed to load package")

	// ErrTooManyPackages indicates that more packages were found than expected
	ErrTooManyPackages = errors.New("found more packages than expected")

	// ErrGetAbsPath indicates that an absolute path could not be determined
	ErrGetAbsPath = errors.New("failed to get absolute path")
)
