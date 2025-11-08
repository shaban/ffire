package benchmark

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shaban/ffire/pkg/fixture"
	"github.com/shaban/ffire/pkg/generator"
	"github.com/shaban/ffire/pkg/schema"
)

// GenerateJavaScript generates a JavaScript/Node.js benchmark with embedded fixture
func GenerateJavaScript(schema *schema.Schema, schemaName, messageName string, jsonData []byte, outputDir string, iterations int) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Step 1: Generate the JavaScript package
	// Use package name (not schema filename) as module name
	config := &generator.PackageConfig{
		Schema:    schema,
		Language:  "javascript",
		OutputDir: outputDir,
		Namespace: schema.Package,
		Optimize:  2,
		Platform:  "current",
		Arch:      "current",
		NoCompile: false,
		Verbose:   false,
	}

	if err := generator.GeneratePackage(config); err != nil {
		return fmt.Errorf("failed to generate JavaScript package: %w", err)
	}

	// Step 2: Convert JSON to binary fixture
	binaryData, err := fixture.Convert(schema, messageName, jsonData)
	if err != nil {
		return fmt.Errorf("failed to convert JSON to binary: %w", err)
	}

	// Step 3: Write the binary fixture
	// JavaScript package is generated in outputDir/javascript subdirectory
	jsDir := filepath.Join(outputDir, "javascript")
	fixturePath := filepath.Join(jsDir, "fixture.bin")
	if err := os.WriteFile(fixturePath, binaryData, 0644); err != nil {
		return fmt.Errorf("failed to write fixture: %w", err)
	}

	// Step 4: Generate the benchmark harness
	benchmarkCode := generateJavaScriptBenchmarkCode(schemaName, messageName, iterations)
	benchPath := filepath.Join(jsDir, "bench.js")
	if err := os.WriteFile(benchPath, []byte(benchmarkCode), 0644); err != nil {
		return fmt.Errorf("failed to write benchmark: %w", err)
	}

	// Step 5: Generate a run script for convenience
	runScript := generateJavaScriptRunScript()
	runPath := filepath.Join(jsDir, "run.sh")
	if err := os.WriteFile(runPath, []byte(runScript), 0755); err != nil {
		return fmt.Errorf("failed to write run script: %w", err)
	}

	return nil
}

// generateJavaScriptBenchmarkCode generates the benchmark harness code
func generateJavaScriptBenchmarkCode(schemaName, messageName string, iterations int) string {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, `const fs = require('fs');
const { %sMessage } = require('./index');

function main() {
  // Load fixture
  const fixtureData = fs.readFileSync('fixture.bin');
  
  const iterations = %d;
  const jsonOutput = process.env.BENCH_JSON === '1';
  
  // Warmup
  for (let i = 0; i < 1000; i++) {
    const msg = %sMessage.decode(fixtureData);
    const encoded = msg.encode();
  }
  
  // Benchmark decode
  const decodeStart = process.hrtime.bigint();
  for (let i = 0; i < iterations; i++) {
    const msg = %sMessage.decode(fixtureData);
  }
  const decodeEnd = process.hrtime.bigint();
  const decodeTimeNs = decodeEnd - decodeStart;
  
  // Benchmark encode (decode once, then encode many times)
  const msg = %sMessage.decode(fixtureData);
  const encodeStart = process.hrtime.bigint();
  let encoded;
  for (let i = 0; i < iterations; i++) {
    encoded = msg.encode();
  }
  const encodeEnd = process.hrtime.bigint();
  const encodeTimeNs = encodeEnd - encodeStart;
  
  // Calculate metrics
  const encodeNs = Math.round(Number(encodeTimeNs) / iterations);
  const decodeNs = Math.round(Number(decodeTimeNs) / iterations);
  const totalNs = encodeNs + decodeNs;
  
  if (jsonOutput) {
    // Output JSON for automation
    const result = {
      language: 'JavaScript',
      format: 'ffire',
      message: '%s',
      iterations: iterations,
      encode_ns: encodeNs,
      decode_ns: decodeNs,
      total_ns: totalNs,
      wire_size: encoded.length,
      fixture_size: fixtureData.length,
      timestamp: new Date().toISOString(),
    };
    console.log(JSON.stringify(result));
  } else {
    // Print human-readable results
    console.log('ffire benchmark: %s');
    console.log('Iterations:  ' + iterations);
    console.log('Encode:      ' + encodeNs + ' ns/op');
    console.log('Decode:      ' + decodeNs + ' ns/op');
    console.log('Total:       ' + totalNs + ' ns/op');
    console.log('Wire size:   ' + encoded.length + ' bytes');
    console.log('Fixture:     ' + fixtureData.length + ' bytes');
    console.log('Total time:  ' + ((Number(encodeTimeNs) + Number(decodeTimeNs)) / 1e9).toFixed(3) + 's');
  }
}

main();
`, messageName, iterations, messageName, messageName, messageName, messageName, messageName)

	return buf.String()
}

// generateJavaScriptRunScript generates a convenience run script
func generateJavaScriptRunScript() string {
	return `#!/bin/bash
# Convenience script to run JavaScript benchmark

# Check if node is available
if ! command -v node &> /dev/null; then
    echo "Error: node not found"
    exit 1
fi

# Install dependencies if needed (npm install builds the N-API addon)
if [ ! -d "node_modules" ]; then
    npm install
fi

# Run benchmark
node bench.js "$@"
`
}
