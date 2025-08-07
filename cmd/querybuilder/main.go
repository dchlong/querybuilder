package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dchlong/querybuilder"
	"github.com/dchlong/querybuilder/parser"
)

const (
	version = "v1.0.0"
	usage   = `querybuilder - Generate type-safe query builders for Go structs

USAGE:
    querybuilder [options] <input-file>

EXAMPLES:
    # Generate query builder for models.go
    querybuilder models.go

    # Generate with custom output file
    querybuilder -output models_querybuilder.go models.go

    # Generate with struct name suffix
    querybuilder -suffix V1 models.go

    # Generate for all Go files in directory
    querybuilder -dir ./models

    # Show supported field types
    querybuilder -types

OPTIONS:`
)

type config struct {
	inputFile   string
	outputFile  string
	suffix      string
	directory   string
	showTypes   bool
	showVersion bool
	showHelp    bool
	verbose     bool
	dryRun      bool
}

func main() {
	cfg := parseFlags()

	if cfg.showHelp {
		printUsage()
		os.Exit(0)
	}

	if cfg.showVersion {
		fmt.Printf("querybuilder %s\n", version)
		os.Exit(0)
	}

	if cfg.showTypes {
		printSupportedTypes()
		os.Exit(0)
	}

	ctx := context.Background()

	if cfg.directory != "" {
		if err := generateForDirectory(ctx, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if cfg.inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: input file is required\n\n")
		printUsage()
		os.Exit(1)
	}

	if err := generateForFile(ctx, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() *config {
	cfg := &config{}

	flag.StringVar(&cfg.outputFile, "output", "", "Output file path (default: <input>_querybuilder.go)")
	flag.StringVar(&cfg.outputFile, "o", "", "Output file path (short)")
	flag.StringVar(&cfg.suffix, "suffix", "", "Suffix to append to struct names")
	flag.StringVar(&cfg.suffix, "s", "", "Suffix to append to struct names (short)")
	flag.StringVar(&cfg.directory, "dir", "", "Process all Go files in directory")
	flag.StringVar(&cfg.directory, "d", "", "Process all Go files in directory (short)")
	flag.BoolVar(&cfg.showTypes, "types", false, "Show supported field types")
	flag.BoolVar(&cfg.showVersion, "version", false, "Show version")
	flag.BoolVar(&cfg.showVersion, "v", false, "Show version (short)")
	flag.BoolVar(&cfg.showHelp, "help", false, "Show help")
	flag.BoolVar(&cfg.showHelp, "h", false, "Show help (short)")
	flag.BoolVar(&cfg.verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&cfg.dryRun, "dry-run", false, "Show what would be generated without writing files")

	flag.Usage = printUsage
	flag.Parse()

	if len(flag.Args()) > 0 {
		cfg.inputFile = flag.Args()[0]
	}

	return cfg
}

func printUsage() {
	fmt.Println(usage)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("FIELD TYPES:")
	fmt.Println("  Supported: string, int, int64, float64, bool, time.Time, *time.Time")
	fmt.Println("  JSON:      []string, datatypes.JSONType[T]")
	fmt.Println("  Note:      Structs must have '//gen:querybuilder' annotation")
}

func printSupportedTypes() {
	structsParser := &parser.Structs{}
	generator := querybuilder.NewQueryBuilderGenerator(structsParser)

	fmt.Println("Supported Field Types:")
	for _, fieldType := range generator.GetSupportedFieldTypes() {
		fmt.Printf("  ✓ %s\n", fieldType)
	}

	fmt.Println("\nUnsupported Field Types:")
	for _, fieldType := range generator.GetUnsupportedFieldTypes() {
		fmt.Printf("  ✗ %s\n", fieldType)
	}

	fmt.Println("\nSpecial Types:")
	fmt.Println("  ✓ []string (JSON array)")
	fmt.Println("  ✓ datatypes.JSONType[T] (JSON object)")
	fmt.Println("  ✓ *time.Time (nullable timestamp)")
	fmt.Println("  ✓ *string (nullable string)")
}

func generateForFile(ctx context.Context, cfg *config) error {
	// Validate input file exists
	if _, err := os.Stat(cfg.inputFile); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", repository.ErrInputFileNotFound, cfg.inputFile)
	}

	// Determine output file
	outputFile := cfg.outputFile
	if outputFile == "" {
		outputFile = generateOutputFileName(cfg.inputFile)
	}

	if cfg.verbose {
		fmt.Printf("Input file:  %s\n", cfg.inputFile)
		fmt.Printf("Output file: %s\n", outputFile)
		if cfg.suffix != "" {
			fmt.Printf("Suffix:      %s\n", cfg.suffix)
		}
	}

	// Create generator
	structsParser := &parser.Structs{}
	generator := querybuilder.NewQueryBuilderGenerator(structsParser)

	if cfg.dryRun {
		// Generate in memory to check what would be generated
		code, packageName, err := generator.GenerateInMemory(ctx, cfg.inputFile, cfg.suffix)
		if err != nil {
			return fmt.Errorf("generation failed: %w", err)
		}

		fmt.Printf("Would generate %d bytes of code for package '%s'\n", len(code), packageName)
		fmt.Printf("Output would be written to: %s\n", outputFile)
		return nil
	}

	// Generate the query builder
	if err := generator.Generate(ctx, cfg.inputFile, outputFile, cfg.suffix); err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	fmt.Printf("Successfully generated query builder: %s\n", outputFile)
	return nil
}

func generateForDirectory(ctx context.Context, cfg *config) error {
	// Find all Go files in directory
	files, err := findGoFiles(cfg.directory)
	if err != nil {
		return fmt.Errorf("failed to find Go files in directory %s: %w", cfg.directory, err)
	}

	if len(files) == 0 {
		return fmt.Errorf("%w: %s", repository.ErrNoGoFiles, cfg.directory)
	}

	if cfg.verbose {
		fmt.Printf("Found %d Go files in %s\n", len(files), cfg.directory)
	}

	successCount := 0
	var errors []string

	// Process each file
	for _, file := range files {
		fileCfg := *cfg
		fileCfg.inputFile = file
		fileCfg.directory = "" // Clear directory to process single file

		if cfg.verbose {
			fmt.Printf("Processing: %s\n", file)
		}

		if err := generateForFile(ctx, &fileCfg); err != nil {
			if cfg.verbose {
				fmt.Printf("  Skipped: %v\n", err)
			}
			errors = append(errors, fmt.Sprintf("%s: %v", file, err))
		} else {
			successCount++
		}
	}

	fmt.Printf("Processed %d files successfully", successCount)
	if len(errors) > 0 {
		fmt.Printf(" (%d errors)", len(errors))
	}
	fmt.Println()

	if cfg.verbose && len(errors) > 0 {
		fmt.Println("\nErrors:")
		for _, err := range errors {
			fmt.Printf("  %s\n", err)
		}
	}

	return nil
}

func findGoFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files and generated files
		if strings.HasSuffix(path, "_test.go") ||
			strings.HasSuffix(path, "_querybuilder.go") ||
			strings.Contains(path, "generated") {
			return nil
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

func generateOutputFileName(inputFile string) string {
	ext := filepath.Ext(inputFile)
	base := strings.TrimSuffix(inputFile, ext)
	return base + "_querybuilder" + ext
}
