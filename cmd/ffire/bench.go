package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shaban/ffire/pkg/benchmark"
	"github.com/shaban/ffire/pkg/parser"
	"github.com/shaban/ffire/pkg/validator"
)

func runBench(args []string) {
	fs := flag.NewFlagSet("bench", flag.ExitOnError)
	schemaFile := fs.String("schema", "", "Path to .ffi schema file (required)")
	jsonFile := fs.String("json", "", "Path to JSON fixture file (required)")
	outputDir := fs.String("output", "", "Output directory (required)")
	lang := fs.String("lang", "go", "Target language: go, cpp, python (default: go)")
	messageName := fs.String("message", "Message", "Message type name to encode (default: Message)")
	iterations := fs.Int("iterations", 100000, "Number of benchmark iterations (default: 100000)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: ffire bench [options]

Generate benchmark executables with embedded fixtures.

Options:
`)
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  ffire bench --schema schema.ffi --json data.json --output bench/
  ffire bench --lang cpp --schema schema.ffi --json data.json --output bench_cpp/
  ffire bench --schema schema.ffi --json data.json --output bench/ --iterations 10000000
`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *schemaFile == "" || *jsonFile == "" || *outputDir == "" {
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

	// Read JSON file
	jsonData, err := os.ReadFile(*jsonFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading JSON file: %v\n", err)
		os.Exit(1)
	}

	// Validate JSON against schema
	if err := validator.ValidateJSON(schema, *messageName, jsonData); err != nil {
		fmt.Fprintf(os.Stderr, "Error validating JSON: %v\n", err)
		os.Exit(1)
	}

	// Extract schema name from file path
	schemaName := filepath.Base(*schemaFile)
	schemaName = strings.TrimSuffix(schemaName, filepath.Ext(schemaName))

	// Generate benchmark based on language
	switch *lang {
	case "go":
		if err := benchmark.GenerateGo(schema, schemaName, *messageName, jsonData, *outputDir, *iterations); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating benchmark: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Generated Go benchmark in %s\n", *outputDir)
		fmt.Printf("  Run with: cd %s && go run .\n", *outputDir)

	case "cpp":
		if err := benchmark.GenerateCpp(schema, schemaName, *messageName, jsonData, *outputDir, *iterations); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating benchmark: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Generated C++ benchmark in %s\n", *outputDir)
		fmt.Printf("\n  Build with CMake:\n")
		fmt.Printf("    cd %s && cmake -B build && cmake --build build && ./build/bench\n", *outputDir)
		fmt.Printf("\n  Or build with Make (fallback):\n")
		fmt.Printf("    cd %s && make && ./bench\n", *outputDir)

	case "python":
		if err := benchmark.GeneratePython(schema, schemaName, *messageName, jsonData, *outputDir, *iterations); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating benchmark: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Generated Python benchmark in %s\n", *outputDir)
		fmt.Printf("  Run with: cd %s/python && python3 bench.py\n", *outputDir)

	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported language '%s' (supported: go, cpp, python)\n", *lang)
		os.Exit(1)
	}
}
