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

// GeneratePHP generates a PHP benchmark with embedded fixture
func GeneratePHP(schema *schema.Schema, schemaName, messageName string, jsonData []byte, outputDir string, iterations int) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Step 1: Generate the PHP package
	config := &generator.PackageConfig{
		Schema:    schema,
		Language:  "php",
		OutputDir: outputDir,
		Namespace: schemaName,
		Optimize:  2,
		Platform:  "current",
		Arch:      "current",
		NoCompile: false,
		Verbose:   false,
	}

	if err := generator.GeneratePackage(config); err != nil {
		return fmt.Errorf("failed to generate PHP package: %w", err)
	}

	// Step 2: Convert JSON to binary fixture
	binaryData, err := fixture.Convert(schema, messageName, jsonData)
	if err != nil {
		return fmt.Errorf("failed to convert JSON to binary: %w", err)
	}

	// Step 3: Write the binary fixture
	phpDir := filepath.Join(outputDir, "php")
	fixturePath := filepath.Join(phpDir, "fixture.bin")
	if err := os.WriteFile(fixturePath, binaryData, 0644); err != nil {
		return fmt.Errorf("failed to write fixture: %w", err)
	}

	// Step 4: Generate the benchmark harness
	benchmarkCode := generatePHPBenchmarkCode(schemaName, messageName, iterations)
	benchPath := filepath.Join(phpDir, "bench.php")
	if err := os.WriteFile(benchPath, []byte(benchmarkCode), 0644); err != nil {
		return fmt.Errorf("failed to write benchmark: %w", err)
	}

	// Step 5: Generate a run script for convenience
	runScript := generatePHPRunScript()
	runPath := filepath.Join(phpDir, "run.sh")
	if err := os.WriteFile(runPath, []byte(runScript), 0755); err != nil {
		return fmt.Errorf("failed to write run script: %w", err)
	}

	return nil
}

// generatePHPBenchmarkCode generates the benchmark harness code
func generatePHPBenchmarkCode(schemaName, messageName string, iterations int) string {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, `<?php

// Load shared library using FFI
$libName = match (PHP_OS_FAMILY) {
    'Darwin' => 'lib%s.dylib',
    'Windows' => '%s.dll',
    default => 'lib%s.so',
};

$ffi = FFI::cdef(
    "void* ffire_decode_%s(const void* data, int size);
     void* ffire_encode_%s(const void* msg, int flags, int* size);
     void ffire_free_%s(void* msg);",
    __DIR__ . '/' . $libName
);

function decode($ffi, $data) {
    $ptr = $ffi->ffire_decode_%s($data, strlen($data));
    if ($ptr === null) {
        throw new Exception('Decode failed');
    }
    return $ptr;
}

function encode($ffi, $msgPtr) {
    $size = FFI::new('int');
    $ptr = $ffi->ffire_encode_%s($msgPtr, 0, FFI::addr($size));
    if ($ptr === null) {
        throw new Exception('Encode failed');
    }
    return FFI::string($ptr, $size->cdata);
}

function freeMessage($ffi, $msgPtr) {
    $ffi->ffire_free_%s($msgPtr);
}

// Load fixture
$fixtureData = file_get_contents('fixture.bin');
if ($fixtureData === false) {
    die("Failed to load fixture.bin\n");
}

$iterations = %d;
$jsonOutput = getenv('BENCH_JSON') === '1';

// Warmup
for ($i = 0; $i < 1000; $i++) {
    $msgPtr = decode($ffi, $fixtureData);
    $encoded = encode($ffi, $msgPtr);
    freeMessage($ffi, $msgPtr);
}

// Benchmark decode
$decodeStart = hrtime(true);
for ($i = 0; $i < $iterations; $i++) {
    $msgPtr = decode($ffi, $fixtureData);
    freeMessage($ffi, $msgPtr);
}
$decodeEnd = hrtime(true);
$decodeTimeNs = $decodeEnd - $decodeStart;

// Benchmark encode (decode once, then encode many times)
$msgPtr = decode($ffi, $fixtureData);
$encodeStart = hrtime(true);
$encoded = null;
for ($i = 0; $i < $iterations; $i++) {
    $encoded = encode($ffi, $msgPtr);
}
$encodeEnd = hrtime(true);
$encodeTimeNs = $encodeEnd - $encodeStart;
freeMessage($ffi, $msgPtr);

// Calculate metrics
$encodeNs = (int)($encodeTimeNs / $iterations);
$decodeNs = (int)($decodeTimeNs / $iterations);
$totalNs = $encodeNs + $decodeNs;

if ($jsonOutput) {
    // Output JSON for automation
    $result = [
        'language' => 'PHP',
        'format' => 'ffire',
        'message' => '%s',
        'iterations' => $iterations,
        'encode_ns' => $encodeNs,
        'decode_ns' => $decodeNs,
        'total_ns' => $totalNs,
        'wire_size' => strlen($encoded),
        'fixture_size' => strlen($fixtureData),
        'timestamp' => date('c'),
    ];
    echo json_encode($result) . "\n";
} else {
    // Print human-readable results
    echo "ffire benchmark: %s\n";
    echo "Iterations:  {$iterations}\n";
    echo "Encode:      {$encodeNs} ns/op\n";
    echo "Decode:      {$decodeNs} ns/op\n";
    echo "Total:       {$totalNs} ns/op\n";
    echo "Wire size:   " . strlen($encoded) . " bytes\n";
    echo "Fixture:     " . strlen($fixtureData) . " bytes\n";
    $totalTimeS = ($encodeTimeNs + $decodeTimeNs) / 1e9;
    echo sprintf("Total time:  %%.3fs\n", $totalTimeS);
}
`, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName,
		iterations, schemaName, schemaName)

	return buf.String()
}

// generatePHPRunScript generates a convenience run script
func generatePHPRunScript() string {
	return `#!/bin/bash
# Convenience script to run PHP benchmark

# Check if php is available
if ! command -v php &> /dev/null; then
    echo "Error: php not found"
    exit 1
fi

# Check PHP version (need 7.4+ for FFI)
PHP_VERSION=$(php -r 'echo PHP_VERSION;')
PHP_MAJOR=$(echo $PHP_VERSION | cut -d. -f1)
PHP_MINOR=$(echo $PHP_VERSION | cut -d. -f2)

if [ "$PHP_MAJOR" -lt 7 ] || ([ "$PHP_MAJOR" -eq 7 ] && [ "$PHP_MINOR" -lt 4 ]); then
    echo "Error: PHP 7.4+ required for FFI support (found $PHP_VERSION)"
    exit 1
fi

# Run benchmark
php bench.php "$@"
`
}
