// ffire - FFI Encoding code generator and tooling
package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/shaban/ffire/pkg/errors"
)

func main() {
	// Panic recovery to provide clean error messages
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Unexpected error occurred:\n")
			fmt.Fprintf(os.Stderr, "%v\n\n", r)
			
			// Print stack trace in verbose mode or if FFIRE_DEBUG is set
			if os.Getenv("FFIRE_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "Stack trace:\n%s\n", debug.Stack())
			} else {
				fmt.Fprintf(os.Stderr, "Run with FFIRE_DEBUG=1 for stack trace\n")
			}
			
			os.Exit(2)
		}
	}()

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

// formatError formats an error with helpful hints if available
func formatError(err error) string {
	if ffErr, ok := err.(*errors.Error); ok {
		return ffErr.ErrorWithHint()
	}
	// Try to unwrap and check again
	if ffErr := errors.Unwrap(err); ffErr != nil {
		if e, ok := ffErr.(*errors.Error); ok {
			return e.ErrorWithHint()
		}
	}
	return err.Error()
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
