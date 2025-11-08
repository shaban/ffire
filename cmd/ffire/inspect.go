package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/shaban/ffire/pkg/inspector"
	"github.com/shaban/ffire/pkg/parser"
	"github.com/shaban/ffire/pkg/validator"
)

func runInspect(args []string) {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: ffire inspect [options]

Inspect and visualize binary wire format.

Options:
`)
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  ffire inspect --schema audio.ffi --binary output.bin
  ffire inspect --schema audio.ffi --binary output.bin --message AudioData
  ffire inspect --schema audio.ffi --binary output.bin --hex
`)
	}

	schemaFile := fs.String("schema", "", "Path to .ffi schema file (required)")
	binaryFile := fs.String("binary", "", "Path to binary wire format file (required)")
	messageName := fs.String("message", "Message", "Message type name")
	showHex := fs.Bool("hex", false, "Show hex dump")
	compact := fs.Bool("compact", false, "Compact output (no field annotations)")

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	// Validate required flags
	if *schemaFile == "" || *binaryFile == "" {
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

	// Read binary file
	data, err := os.ReadFile(*binaryFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading binary file: %v\n", err)
		os.Exit(1)
	}

	// Inspect binary data
	config := &inspector.Config{
		Schema:      schema,
		MessageName: *messageName,
		Data:        data,
		ShowHex:     *showHex,
		Compact:     *compact,
	}

	output, err := inspector.Inspect(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error inspecting binary: %s\n", formatError(err))
		os.Exit(1)
	}

	fmt.Print(output)
}
