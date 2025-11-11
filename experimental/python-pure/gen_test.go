package main

import (
	"fmt"
	"os"

	"github.com/shaban/ffire/pkg/generator"
	"github.com/shaban/ffire/pkg/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <schema.ffi>\n", os.Args[0])
		os.Exit(1)
	}

	schema, err := parser.Parse(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing schema: %v\n", err)
		os.Exit(1)
	}

	code, err := generator.GeneratePythonPure(schema)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating Python: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(code))
}
