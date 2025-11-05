package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/shaban/ffire/pkg/generator"
	"github.com/shaban/ffire/pkg/parser"
)

func runGenerate(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	schemaFile := fs.String("schema", "", "Path to .ffi schema file (required)")
	lang := fs.String("lang", "", "Target language: go, cpp, swift (required)")
	outputFile := fs.String("output", "", "Output file path (required)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: ffire generate [options]

Generate encoder/decoder code for target language.

Options:
`)
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  ffire generate --schema schema.ffi --lang go --output generated.go
  ffire generate --schema schema.ffi --lang cpp --output generated.hpp
  ffire generate --schema schema.ffi --lang swift --output generated.swift
`)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *schemaFile == "" || *lang == "" || *outputFile == "" {
		fs.Usage()
		os.Exit(1)
	}

	// Parse schema
	schema, err := parser.Parse(*schemaFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing schema: %v\n", err)
		os.Exit(1)
	}

	// Generate code
	var code []byte
	switch *lang {
	case "go":
		code, err = generator.GenerateGo(schema)
	case "cpp":
		code, err = generator.GenerateCpp(schema)
	case "swift":
		code, err = generator.GenerateSwift(schema)
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported language: %s\n", *lang)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		os.Exit(1)
	}

	// Write output
	if err := os.WriteFile(*outputFile, code, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Generated %s code from %s to %s\n", *lang, *schemaFile, *outputFile)
}
