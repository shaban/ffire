package main

import (
	"fmt"
	"log"

	"github.com/shaban/ffire/pkg/analyzer"
	"github.com/shaban/ffire/pkg/parser"
)

func main() {
	// Parse complex.ffi schema
	schema, err := parser.Parse("testdata/schema/complex.ffi")
	if err != nil {
		log.Fatal(err)
	}

	// Analyze all types
	typeInfo := analyzer.Analyze(schema)

	fmt.Println("=== Schema Analysis ===")
	fmt.Println()
	fmt.Printf("Package: %s\n\n", schema.Package)

	// Print analysis for each type
	for name, info := range typeInfo {
		fmt.Printf("Type: %s\n", name)
		fmt.Printf("  IsFixedSize: %v\n", info.IsFixedSize)
		if info.IsFixedSize {
			fmt.Printf("  FixedSize:   %d bytes\n", info.FixedSize)
		}
		fmt.Printf("  MaxSize:     %d bytes (%.2f KB)\n", info.MaxSize, float64(info.MaxSize)/1024)
		fmt.Printf("  HasStrings:  %v\n", info.HasStrings)
		fmt.Printf("  HasArrays:   %v\n", info.HasArrays)
		fmt.Printf("  NestDepth:   %d\n", info.NestDepth)
		fmt.Println()
	}
}
