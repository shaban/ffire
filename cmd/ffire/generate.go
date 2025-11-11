package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/shaban/ffire/pkg/generator"
	"github.com/shaban/ffire/pkg/parser"
	"github.com/shaban/ffire/pkg/validator"
)

func runGenerate(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	schemaFile := fs.String("schema", "", "Path to .ffi schema file (required)")
	lang := fs.String("lang", "", "Target language: go, cpp, js, python, swift, dart, java, csharp (required)")
	output := fs.String("out", "./dist", "Output directory for generated package")
	optimize := fs.Int("O", 2, "Optimization level (0-3)")
	platform := fs.String("platform", "current", "Target platform: darwin, linux, windows, all")
	arch := fs.String("arch", "current", "Target architecture: arm64, x86_64, all")
	namespace := fs.String("ns", "", "Namespace/package name (defaults to schema name)")
	noCompile := fs.Bool("no-compile", false, "Skip dylib compilation (for testing)")
	verbose := fs.Bool("v", false, "Verbose output")

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
  
  # Generate pure JavaScript package (no native bindings)
  ffire generate -lang js -schema audio.ffi -out ./dist
  
  # Generate C++ package with custom namespace
  ffire generate -lang cpp -schema audio.ffi -out ./dist -ns myaudio
  
  # Multi-platform build
  ffire generate -lang python -schema audio.ffi -platform all
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
		fmt.Fprintf(os.Stderr, "Error parsing schema: %v\n", formatError(err))
		os.Exit(1)
	}

	// Validate schema
	if err := validator.ValidateSchema(schema); err != nil {
		fmt.Fprintf(os.Stderr, "Error validating schema: %s\n", formatError(err))
		os.Exit(1)
	}

	// Generate package
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
		fmt.Fprintf(os.Stderr, "Error generating package: %s\n", formatError(err))
		os.Exit(1)
	}
}
