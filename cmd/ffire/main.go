// ffire - FFI Encoding code generator and tooling
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "fixture":
		runFixture(os.Args[2:])
	case "validate":
		runValidate(os.Args[2:])
	case "generate":
		runGenerate(os.Args[2:])
	case "bench":
		runBench(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`ffire - FFI Encoding code generator and tooling

Usage:
  ffire <command> [options]

Commands:
  fixture     Convert JSON fixture to binary wire format
  validate    Validate schema and fixture files
  generate    Generate encoder/decoder code (Go, C++, Swift)
  bench       Generate benchmark executables

Examples:
  ffire fixture --schema testdata/schema/complex.ffi --json testdata/json/complex.json --output out.bin
  ffire validate --schema testdata/schema/complex.ffi --json testdata/json/complex.json
  ffire generate --schema testdata/schema/complex.ffi --lang go --output generated/
  ffire bench --schema testdata/schema/complex.ffi --output bench/

Use "ffire <command> --help" for more information about a command.`)
}
