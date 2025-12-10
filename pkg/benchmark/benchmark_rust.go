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

// GenerateRust generates a Rust benchmark with embedded fixture
// This uses native Rust (no FFI) for maximum safety and idiomatic code
func GenerateRust(schema *schema.Schema, schemaName, messageName string, jsonData []byte, outputDir string, iterations int) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Step 1: Generate the Rust package
	config := &generator.PackageConfig{
		Schema:    schema,
		Language:  "rust",
		OutputDir: outputDir,
		Namespace: schema.Package,
		Optimize:  2,
		Platform:  "current",
		Arch:      "current",
		NoCompile: false,
		Verbose:   false,
	}

	if err := generator.GeneratePackage(config); err != nil {
		return fmt.Errorf("failed to generate Rust package: %w", err)
	}

	// Step 2: Convert JSON to binary fixture
	binaryData, err := fixture.Convert(schema, messageName, jsonData)
	if err != nil {
		return fmt.Errorf("failed to convert JSON to binary: %w", err)
	}

	// Step 3: Write the binary fixture
	rustDir := filepath.Join(outputDir, "rust")
	fixturePath := filepath.Join(rustDir, "fixture.bin")
	if err := os.WriteFile(fixturePath, binaryData, 0644); err != nil {
		return fmt.Errorf("failed to write fixture: %w", err)
	}

	// Step 4: Create bench directory
	benchDir := filepath.Join(rustDir, "src", "bin")
	if err := os.MkdirAll(benchDir, 0755); err != nil {
		return fmt.Errorf("failed to create bench directory: %w", err)
	}

	// Step 5: Generate the benchmark main.rs
	benchmarkCode := generateRustBenchmarkCode(schema.Package, messageName, iterations)
	mainPath := filepath.Join(benchDir, "bench.rs")
	if err := os.WriteFile(mainPath, []byte(benchmarkCode), 0644); err != nil {
		return fmt.Errorf("failed to write benchmark: %w", err)
	}

	// Step 6: Update Cargo.toml to add [[bin]] section
	if err := addBenchBinaryToCargo(rustDir); err != nil {
		return fmt.Errorf("failed to update Cargo.toml: %w", err)
	}

	// Step 7: Generate a run script for convenience
	runScript := generateRustRunScript()
	runPath := filepath.Join(rustDir, "run.sh")
	if err := os.WriteFile(runPath, []byte(runScript), 0755); err != nil {
		return fmt.Errorf("failed to write run script: %w", err)
	}

	return nil
}

// generateRustBenchmarkCode generates the benchmark harness code for native Rust
func generateRustBenchmarkCode(schemaName, messageName string, iterations int) string {
	buf := &bytes.Buffer{}

	// Convert message name to snake_case to match Rust function names
	snakeName := generator.ToSnakeCase(messageName)

	fmt.Fprintf(buf, `use std::fs;
use std::time::Instant;
use std::env;

use %s::{encode_%s_message, decode_%s_message};

fn main() {
    // Read fixture
    let fixture_data = fs::read("fixture.bin")
        .expect("Failed to read fixture.bin");

    let iterations: usize = %d;
    let json_output = env::var("BENCH_JSON").map(|v| v == "1").unwrap_or(false);

    // Warmup
    for _ in 0..1000 {
        let msg = decode_%s_message(&fixture_data).expect("Warmup decode failed");
        let _ = encode_%s_message(&msg);
    }

    // Benchmark decode
    let decode_start = Instant::now();
    for _ in 0..iterations {
        let _ = decode_%s_message(&fixture_data).expect("Decode failed");
    }
    let decode_duration = decode_start.elapsed();
    let decode_time_ns = decode_duration.as_nanos() as u64;

    // Benchmark encode (decode once, then encode many times)
    let msg = decode_%s_message(&fixture_data).expect("Decode for encode benchmark failed");

    let encode_start = Instant::now();
    let mut encoded: Vec<u8> = Vec::new();
    for _ in 0..iterations {
        encoded = encode_%s_message(&msg);
    }
    let encode_duration = encode_start.elapsed();
    let encode_time_ns = encode_duration.as_nanos() as u64;

    // Calculate metrics
    let encode_ns = encode_time_ns / iterations as u64;
    let decode_ns = decode_time_ns / iterations as u64;
    let total_ns = encode_ns + decode_ns;
    let wire_size = encoded.len();
    let fixture_size = fixture_data.len();

    if json_output {
        // Output JSON for automation
        println!(
            r#"{{"language": "Rust", "format": "ffire", "message": "%s", "iterations": {}, "encode_ns": {}, "decode_ns": {}, "total_ns": {}, "wire_size": {}, "fixture_size": {}}}"#,
            iterations, encode_ns, decode_ns, total_ns, wire_size, fixture_size
        );
    } else {
        // Print human-readable results
        println!("ffire benchmark: %s");
        println!("Iterations:  {}", iterations);
        println!("Encode:      {} ns/op", encode_ns);
        println!("Decode:      {} ns/op", decode_ns);
        println!("Total:       {} ns/op", total_ns);
        println!("Wire size:   {} bytes", wire_size);
        println!("Fixture:     {} bytes", fixture_size);
        let total_time_s = (encode_time_ns + decode_time_ns) as f64 / 1_000_000_000.0;
        println!("Total time:  {:.3}s", total_time_s);
    }
}
`, schemaName, snakeName, snakeName, // use statement
		iterations,
		snakeName, snakeName, // warmup
		snakeName,            // benchmark decode
		snakeName, snakeName, // decode for encode, encode
		messageName,          // JSON message name (original case for display)
		messageName)          // human-readable message name

	return buf.String()
}

// addBenchBinaryToCargo adds a [[bin]] section to Cargo.toml for the benchmark
func addBenchBinaryToCargo(rustDir string) error {
	cargoPath := filepath.Join(rustDir, "Cargo.toml")
	content, err := os.ReadFile(cargoPath)
	if err != nil {
		return err
	}

	// Check if bench binary already exists
	if bytes.Contains(content, []byte(`name = "bench"`)) {
		return nil // Already has bench binary
	}

	// Add [[bin]] section
	binSection := `

[[bin]]
name = "bench"
path = "src/bin/bench.rs"
`

	content = append(content, []byte(binSection)...)
	return os.WriteFile(cargoPath, content, 0644)
}

// generateRustRunScript generates a convenience run script
func generateRustRunScript() string {
	return `#!/bin/bash
# Convenience script to run Rust benchmark
set -e

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Build in release mode
cargo build --release --bin bench

# Run the benchmark
./target/release/bench "$@"
`
}
