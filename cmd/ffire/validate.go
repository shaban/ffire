package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/shaban/ffire/pkg/parser"
	"github.com/shaban/ffire/pkg/validator"
)

func runValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	schemaFile := fs.String("schema", "", "Path to .ffi schema file (required)")
	jsonFile := fs.String("json", "", "Path to JSON fixture file (optional)")
	messageName := fs.String("message", "Message", "Message type name (default: Message)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: ffire validate [options]

Validate schema and optionally validate JSON fixture against schema.

Options:
`)
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  ffire validate --schema schema.ffi
  ffire validate --schema schema.ffi --json data.json
  ffire validate --schema schema.ffi --json data.json --message DeviceList
`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	// Validate required flags
	if *schemaFile == "" {
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

	fmt.Printf("✓ Schema %s is valid\n", *schemaFile)

	// If JSON file is provided, validate it too
	if *jsonFile != "" {
		jsonData, err := os.ReadFile(*jsonFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading JSON file: %v\n", err)
			os.Exit(1)
		}

		if err := validator.ValidateJSON(schema, *messageName, jsonData); err != nil {
			fmt.Fprintf(os.Stderr, "Error validating JSON: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ JSON %s is valid\n", *jsonFile)
	}
}
