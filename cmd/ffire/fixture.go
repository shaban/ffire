package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/shaban/ffire/pkg/fixture"
	"github.com/shaban/ffire/pkg/parser"
	"github.com/shaban/ffire/pkg/validator"
)

func runFixture(args []string) {
	fs := flag.NewFlagSet("fixture", flag.ExitOnError)
	schemaFile := fs.String("schema", "", "Path to .ffi schema file (required)")
	jsonFile := fs.String("json", "", "Path to JSON fixture file (required)")
	outputFile := fs.String("output", "", "Path to output binary file (required)")
	messageName := fs.String("message", "", "Message type name to encode (auto-detected if only one root type)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: ffire fixture [options]

Convert JSON fixture to binary wire format.

Options:
`)
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  ffire fixture --schema schema.ffi --json data.json --output data.bin
  ffire fixture --schema schema.ffi --json data.json --output data.bin --message DeviceList
`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	// Validate required flags
	if *schemaFile == "" || *jsonFile == "" || *outputFile == "" {
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

	// Auto-detect message name if not specified
	if *messageName == "" {
		if len(schema.Messages) == 0 {
			fmt.Fprintf(os.Stderr, "Error: No root types found in schema\n")
			os.Exit(1)
		}
		if len(schema.Messages) == 1 {
			*messageName = schema.Messages[0].Name
			fmt.Printf("Auto-detected root type: %s\n", *messageName)
		} else {
			fmt.Fprintf(os.Stderr, "Error: Multiple root types found, please specify --message:\n")
			for _, msg := range schema.Messages {
				fmt.Fprintf(os.Stderr, "  - %s\n", msg.Name)
			}
			os.Exit(1)
		}
	}

	// Read JSON file
	jsonData, err := os.ReadFile(*jsonFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading JSON file: %v\n", err)
		os.Exit(1)
	}

	// Validate JSON against schema
	if err := validator.ValidateJSON(schema, *messageName, jsonData); err != nil {
		fmt.Fprintf(os.Stderr, "Error validating JSON: %s\n", formatError(err))
		os.Exit(1)
	}

	// Convert to binary
	binary, err := fixture.Convert(schema, *messageName, jsonData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting to binary: %v\n", err)
		os.Exit(1)
	}

	// Write output file
	if err := os.WriteFile(*outputFile, binary, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Converted %s to %s (%d bytes)\n", *jsonFile, *outputFile, len(binary))
}
