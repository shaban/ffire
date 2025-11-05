package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println(`
⚠️  This directory uses Mage for task automation.

To install mage:
  go install github.com/magefile/mage@latest

Make sure ~/go/bin is in your PATH, then run:
  mage -l                  # List available targets
  mage bench               # Run full benchmark workflow
  mage clean               # Clean generated files
  mage genAll              # Generate all benchmarks
  mage runGo               # Run Go benchmarks
  mage compare             # Show comparison table

See README.md for more information.`)
	os.Exit(1)
}
