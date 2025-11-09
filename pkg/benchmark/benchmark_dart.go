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

// GenerateDart generates a Dart benchmark with embedded fixture
func GenerateDart(schema *schema.Schema, schemaName, messageName string, jsonData []byte, outputDir string, iterations int) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Step 1: Generate the Dart package
	// Use package name (not schema filename) as module name
	config := &generator.PackageConfig{
		Schema:    schema,
		Language:  "dart",
		OutputDir: outputDir,
		Namespace: schema.Package,
		Optimize:  2,
		Platform:  "current",
		Arch:      "current",
		NoCompile: false,
		Verbose:   false,
	}

	if err := generator.GeneratePackage(config); err != nil {
		return fmt.Errorf("failed to generate Dart package: %w", err)
	}

	// Step 2: Convert JSON to binary fixture
	binaryData, err := fixture.Convert(schema, messageName, jsonData)
	if err != nil {
		return fmt.Errorf("failed to convert JSON to binary: %w", err)
	}

	// Step 3: Write the binary fixture
	dartDir := filepath.Join(outputDir, "dart")
	fixturePath := filepath.Join(dartDir, "fixture.bin")
	if err := os.WriteFile(fixturePath, binaryData, 0644); err != nil {
		return fmt.Errorf("failed to write fixture: %w", err)
	}

	// Step 4: Generate the benchmark harness
	benchmarkCode := generateDartBenchmarkCode(schema.Package, messageName, iterations)
	benchPath := filepath.Join(dartDir, "bench.dart")
	if err := os.WriteFile(benchPath, []byte(benchmarkCode), 0644); err != nil {
		return fmt.Errorf("failed to write benchmark: %w", err)
	}

	// Step 5: Generate a run script for convenience
	runScript := generateDartRunScript()
	runPath := filepath.Join(dartDir, "run.sh")
	if err := os.WriteFile(runPath, []byte(runScript), 0755); err != nil {
		return fmt.Errorf("failed to write run script: %w", err)
	}

	return nil
}

// generateDartBenchmarkCode generates the benchmark harness code
func generateDartBenchmarkCode(schemaName, messageName string, iterations int) string {
	buf := &bytes.Buffer{}

	// Add Message suffix to class name (generator adds this suffix)
	className := messageName + "Message"

	fmt.Fprintf(buf, `import 'dart:io';
import 'dart:convert';
import 'package:%s/%s.dart';

void main() async {
  // Load fixture
  final fixtureData = await File('fixture.bin').readAsBytes();
  
  final iterations = %d;
  final jsonOutput = Platform.environment['BENCH_JSON'] == '1';
  
  // Warmup
  for (var i = 0; i < 1000; i++) {
    final msg = %s.decode(fixtureData);
    final encoded = msg.encode();
  }
  
  // Benchmark decode
  final decodeStart = DateTime.now();
  for (var i = 0; i < iterations; i++) {
    final msg = %s.decode(fixtureData);
  }
  final decodeEnd = DateTime.now();
  final decodeTime = decodeEnd.difference(decodeStart);
  
  // Benchmark encode (decode once, then encode many times)
  final msg = %s.decode(fixtureData);
  final encodeStart = DateTime.now();
  var encoded;
  for (var i = 0; i < iterations; i++) {
    encoded = msg.encode();
  }
  final encodeEnd = DateTime.now();
  final encodeTime = encodeEnd.difference(encodeStart);
  
  // Calculate metrics
  final encodeNs = (encodeTime.inMicroseconds * 1000 / iterations).round();
  final decodeNs = (decodeTime.inMicroseconds * 1000 / iterations).round();
  final totalNs = encodeNs + decodeNs;
  
  if (jsonOutput) {
    // Output JSON for automation
    final result = {
      'language': 'Dart',
      'format': 'ffire',
      'message': '%s',
      'iterations': iterations,
      'encode_ns': encodeNs,
      'decode_ns': decodeNs,
      'total_ns': totalNs,
      'wire_size': encoded.length,
      'fixture_size': fixtureData.length,
      'timestamp': DateTime.now().toIso8601String(),
    };
    print(jsonEncode(result));
  } else {
    // Print human-readable results
    print('ffire benchmark: %s');
    print('Iterations:  $iterations');
    print('Encode:      $encodeNs ns/op');
    print('Decode:      $decodeNs ns/op');
    print('Total:       $totalNs ns/op');
    print('Wire size:   ${encoded.length} bytes');
    print('Fixture:     ${fixtureData.length} bytes');
    print('Total time:  ${(encodeTime.inMilliseconds + decodeTime.inMilliseconds) / 1000}s');
  }
}
`, schemaName, schemaName, iterations, className, className, className, messageName, messageName)

	return buf.String()
}

// generateDartRunScript generates a convenience run script
func generateDartRunScript() string {
	return `#!/bin/bash
# Convenience script to run Dart benchmark

# Check if dart is available
if ! command -v dart &> /dev/null; then
    echo "Error: dart not found"
    exit 1
fi

# Install dependencies if needed
if [ ! -d ".dart_tool" ]; then
    dart pub get
fi

# Run benchmark
dart run bench.dart "$@"
`
}
