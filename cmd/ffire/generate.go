package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/shaban/ffire/pkg/generator"
	"github.com/shaban/ffire/pkg/parser"
	"github.com/shaban/ffire/pkg/schema"
	"github.com/shaban/ffire/pkg/validator"
)

func runGenerate(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	schemaFile := fs.String("schema", "", "Path to .ffi schema file (required)")
	lang := fs.String("lang", "", "Target language: go, cpp, python, swift, etc. (required)")
	output := fs.String("out", "./dist", "Output directory for generated package")
	optimize := fs.Int("O", 2, "Optimization level (0-3)")
	platform := fs.String("platform", "current", "Target platform: darwin, linux, windows, all")
	arch := fs.String("arch", "current", "Target architecture: arm64, x86_64, all")
	namespace := fs.String("ns", "", "Namespace/package name (defaults to schema name)")
	noCompile := fs.Bool("no-compile", false, "Skip dylib compilation (for testing)")
	verbose := fs.Bool("v", false, "Verbose output")

	// Legacy flags for backward compatibility
	legacyOutput := fs.String("output", "", "Legacy: Output file path (single file mode)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: ffire generate [options]

Generate production-ready packages for multiple languages.

Options:
`)
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  # Generate Python package
  ffire generate -lang python -schema audio.ffi -out ./dist
  
  # Generate C++ package with custom namespace
  ffire generate -lang cpp -schema audio.ffi -out ./dist -ns myaudio
  
  # Multi-platform build
  ffire generate -lang python -schema audio.ffi -platform all
  
  # Skip compilation (for template testing)
  ffire generate -lang ruby -schema audio.ffi --no-compile
  
  # Legacy single-file mode (backward compatible)
  ffire generate -schema schema.ffi -lang go -output generated.go
`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *schemaFile == "" || *lang == "" {
		fs.Usage()
		os.Exit(1)
	}

	// Parse schema
	schema, err := parser.Parse(*schemaFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing schema: %v\n", err)
		os.Exit(1)
	}

	// Validate schema
	if err := validator.ValidateSchema(schema); err != nil {
		fmt.Fprintf(os.Stderr, "Error validating schema: %v\n", err)
		os.Exit(1)
	}

	// Check for legacy single-file mode
	if *legacyOutput != "" {
		runLegacyGenerate(schema, *lang, *legacyOutput)
		return
	}

	// Package generation mode (new)
	config := &generator.PackageConfig{
		Schema:    schema,
		Language:  *lang,
		OutputDir: *output,
		Optimize:  *optimize,
		Platform:  *platform,
		Arch:      *arch,
		Namespace: *namespace,
		NoCompile: *noCompile,
		Verbose:   *verbose,
	}

	if err := generator.GeneratePackage(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating package: %v\n", err)
		os.Exit(1)
	}
}

// runLegacyGenerate handles single-file generation for backward compatibility
func runLegacyGenerate(schemaObj *schema.Schema, lang string, outputFile string) {
	var code []byte
	var err error

	switch lang {
	case "go":
		code, err = generator.GenerateGo(schemaObj)
	case "cpp":
		code, err = generator.GenerateCpp(schemaObj)
	case "swift":
		code, err = generator.GenerateSwift(schemaObj)
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported language: %s\n", lang)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		os.Exit(1)
	}

	// Write output
	if err := os.WriteFile(outputFile, code, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Generated %s code to %s\n", lang, outputFile)
}
