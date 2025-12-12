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
	lang := fs.String("lang", "go", "Target language: go, cpp, swift, dart, java, csharp, rust, zig (default: go)")
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
		fmt.Fprintf(os.Stderr, "Error parsing schema: %s\n", formatError(err))
		os.Exit(1)
	}

	// Validate schema
	if err := validator.ValidateSchema(schema); err != nil {
		fmt.Fprintf(os.Stderr, "Error validating schema: %s\n", formatError(err))
		os.Exit(1)
	}

	// Read JSON file
	jsonData, err := os.ReadFile(*jsonFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading JSON file: %v\n", err)
		os.Exit(1)
	}

	// Auto-detect message name if not specified or if default "Message" doesn't exist
	actualMessageName := *messageName
	if len(schema.Messages) == 0 {
		fmt.Fprintf(os.Stderr, "Error: schema has no root types\n")
		os.Exit(1)
	}

	// If using default "Message" but it doesn't exist, use first root type
	if actualMessageName == "Message" {
		found := false
		for _, msg := range schema.Messages {
			if msg.Name == "Message" {
				found = true
				break
			}
		}
		if !found {
			actualMessageName = schema.Messages[0].Name
			fmt.Printf("Note: Using root type '%s' (no 'Message' type found)\n", actualMessageName)
		}
	}

	// Validate JSON against schema
	if err := validator.ValidateJSON(schema, actualMessageName, jsonData); err != nil {
		fmt.Fprintf(os.Stderr, "Error validating JSON: %s\n", formatError(err))
		os.Exit(1)
	}

	// Extract schema name from file path
	schemaName := filepath.Base(*schemaFile)
	schemaName = strings.TrimSuffix(schemaName, filepath.Ext(schemaName))

	// Generate benchmark based on language
	switch *lang {
	case "go":
		if err := benchmark.GenerateGo(schema, schemaName, actualMessageName, jsonData, *outputDir, *iterations); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating benchmark: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Generated Go benchmark in %s\n", *outputDir)
		fmt.Printf("  Run with: cd %s && go run .\n", *outputDir)

	case "cpp":
		if err := benchmark.GenerateCpp(schema, schemaName, actualMessageName, jsonData, *outputDir, *iterations); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating benchmark: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Generated C++ benchmark in %s\n", *outputDir)
		fmt.Printf("\n  Build with CMake:\n")
		fmt.Printf("    cd %s && cmake -B build && cmake --build build && ./build/bench\n", *outputDir)
		fmt.Printf("\n  Or build with Make (fallback):\n")
		fmt.Printf("    cd %s && make && ./bench\n", *outputDir)

	case "dart":
		if err := benchmark.GenerateDart(schema, schemaName, actualMessageName, jsonData, *outputDir, *iterations); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating benchmark: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Generated Dart benchmark in %s\n", *outputDir)
		fmt.Printf("  Run with: cd %s/dart && dart run bench.dart\n", *outputDir)

	case "swift":
		if err := benchmark.GenerateSwift(schema, schemaName, actualMessageName, jsonData, *outputDir, *iterations); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating benchmark: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Generated Swift benchmark in %s\n", *outputDir)
		fmt.Printf("  Run with: cd %s/swift && swift bench.swift\n", *outputDir)

	case "java":
		if err := benchmark.GenerateJava(schema, schemaName, actualMessageName, jsonData, *outputDir, *iterations); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating benchmark: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Generated Java benchmark in %s\n", *outputDir)
		fmt.Printf("  Run with: cd %s/java && javac *.java && java Bench\n", *outputDir)

	case "csharp":
		if err := benchmark.GenerateCSharp(schema, schemaName, actualMessageName, jsonData, *outputDir, *iterations); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating benchmark: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Generated C# benchmark in %s\n", *outputDir)
		fmt.Printf("  Run with: cd %s/csharp && dotnet run -c Release\n", *outputDir)

	case "zig":
		if err := benchmark.GenerateZig(schema, schemaName, actualMessageName, jsonData, *outputDir, *iterations); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating benchmark: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Generated Zig benchmark in %s\n", *outputDir)
		fmt.Printf("  Run with: cd %s/zig && zig build -Doptimize=ReleaseFast && ./zig-out/bin/bench\n", *outputDir)

	case "rust":
		if err := benchmark.GenerateRust(schema, schemaName, actualMessageName, jsonData, *outputDir, *iterations); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating benchmark: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Generated Rust benchmark in %s\n", *outputDir)
		fmt.Printf("  Run with: cd %s/rust && cargo build --release --bin bench && ./target/release/bench\n", *outputDir)

	case "js", "javascript", "igniffi-js":
		if err := benchmark.GenerateIgniffiJS(schema, schemaName, actualMessageName, jsonData, *outputDir, *iterations); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating benchmark: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Generated JavaScript benchmark in %s\n", *outputDir)
		fmt.Printf("  Run with: cd %s/javascript && npm install && node bench.js\n", *outputDir)

	case "python", "py", "igniffi-python":
		if err := benchmark.GenerateIgniffiPython(schema, schemaName, actualMessageName, jsonData, *outputDir, *iterations); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating benchmark: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Generated Python benchmark in %s\n", *outputDir)
		fmt.Printf("  Run with: cd %s/python && pip install . && python bench.py\n", *outputDir)

	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported language '%s' (supported: go, cpp, js, python, swift, dart, java, csharp, zig, rust)\n", *lang)
		os.Exit(1)
	}
}
