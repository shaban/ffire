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

// Note: Swift module keyword sanitization is handled by generator.sanitizeSwiftModuleName()
// See pkg/generator/generator_swift.go for the shared swiftModuleKeywords map.

// GenerateSwift generates a Swift benchmark with embedded fixture
func GenerateSwift(schema *schema.Schema, schemaName, messageName string, jsonData []byte, outputDir string, iterations int) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Step 1: Generate the Swift package
	// Use package name (not schema filename) as module name to avoid keyword collisions
	// e.g., struct.ffi with "package test" → module name "test"
	config := &generator.PackageConfig{
		Schema:    schema,
		Language:  "swift",
		OutputDir: outputDir,
		Namespace: schema.Package, // Use package name, not filename
		Optimize:  2,
		Platform:  "current",
		Arch:      "current",
		NoCompile: false,
		Verbose:   false,
	}

	if err := generator.GeneratePackage(config); err != nil {
		return fmt.Errorf("failed to generate Swift package: %w", err)
	}

	// Step 2: Convert JSON to binary fixture
	binaryData, err := fixture.Convert(schema, messageName, jsonData)
	if err != nil {
		return fmt.Errorf("failed to convert JSON to binary: %w", err)
	}

	// Step 3: Write the binary fixture
	swiftDir := filepath.Join(outputDir, "swift")
	fixturePath := filepath.Join(swiftDir, "fixture.bin")
	if err := os.WriteFile(fixturePath, binaryData, 0644); err != nil {
		return fmt.Errorf("failed to write fixture: %w", err)
	}

	// Step 4: Create bench directory for executable
	benchDir := filepath.Join(swiftDir, "Sources", "bench")
	if err := os.MkdirAll(benchDir, 0755); err != nil {
		return fmt.Errorf("failed to create bench directory: %w", err)
	}

	// Step 5: Generate the benchmark harness
	benchmarkCode := generateSwiftBenchmarkCode(schema, schemaName, messageName, iterations)
	benchPath := filepath.Join(benchDir, "main.swift")
	if err := os.WriteFile(benchPath, []byte(benchmarkCode), 0644); err != nil {
		return fmt.Errorf("failed to write benchmark: %w", err)
	}

	// Step 6: Update Package.swift to add bench executable
	if err := addBenchExecutableToPackage(swiftDir, schema, schemaName); err != nil {
		return fmt.Errorf("failed to update Package.swift: %w", err)
	}

	// Step 7: Generate a run script for convenience
	runScript := generateSwiftRunScript(schemaName)
	runPath := filepath.Join(swiftDir, "run.sh")
	if err := os.WriteFile(runPath, []byte(runScript), 0755); err != nil {
		return fmt.Errorf("failed to write run script: %w", err)
	}

	return nil
}

// generateSwiftBenchmarkCode generates the benchmark harness code
func generateSwiftBenchmarkCode(schema *schema.Schema, schemaName, messageName string, iterations int) string {
	buf := &bytes.Buffer{}

	// Swift generator appends "Message" suffix to avoid keyword collisions
	swiftTypeName := messageName + "Message"

	// Use package name (not schema filename) as module name to avoid keyword collisions
	// e.g., struct.ffi with "package test" → import test (not import struct)
	swiftModuleName := generator.SanitizeSwiftModuleName(schema.Package)

	// Use the generated Swift package bindings (like Dart/Python), not raw C FFI
	fmt.Fprintf(buf, `import Foundation
import %s

// Load fixture
guard let fixtureData = try? Data(contentsOf: URL(fileURLWithPath: "fixture.bin")) else {
    fatalError("Failed to load fixture.bin")
}

let iterations = %d
let jsonOutput = ProcessInfo.processInfo.environment["BENCH_JSON"] == "1"

// Warmup
for _ in 0..<1000 {
    do {
        let msg = try %s.decode(fixtureData)
        let _ = try msg.encode()
    } catch {
        fatalError("Warmup failed: \(error)")
    }
}

// Benchmark decode
let decodeStart = DispatchTime.now()
for _ in 0..<iterations {
    do {
        let _ = try %s.decode(fixtureData)
    } catch {
        fatalError("Decode failed: \(error)")
    }
}
let decodeEnd = DispatchTime.now()
let decodeTimeNs = decodeEnd.uptimeNanoseconds - decodeStart.uptimeNanoseconds

// Benchmark encode (decode once, then encode many times)
var encodedData: Data?
do {
    let msg = try %s.decode(fixtureData)
    let encodeStart = DispatchTime.now()
    for _ in 0..<iterations {
        encodedData = try msg.encode()
    }
    let encodeEnd = DispatchTime.now()
    let encodeTimeNs = encodeEnd.uptimeNanoseconds - encodeStart.uptimeNanoseconds
    
    // Calculate metrics
    let encodeNs = Int(encodeTimeNs) / iterations
    let decodeNs = Int(decodeTimeNs) / iterations
    let totalNs = encodeNs + decodeNs
    
    if jsonOutput {
        // Output JSON for automation
        let result: [String: Any] = [
            "language": "Swift",
            "format": "ffire",
            "message": "%s",
            "iterations": iterations,
            "encode_ns": encodeNs,
            "decode_ns": decodeNs,
            "total_ns": totalNs,
            "wire_size": encodedData?.count ?? 0,
            "fixture_size": fixtureData.count,
            "timestamp": ISO8601DateFormatter().string(from: Date())
        ]
        if let jsonData = try? JSONSerialization.data(withJSONObject: result),
           let jsonString = String(data: jsonData, encoding: .utf8) {
            print(jsonString)
        }
    } else {
        // Print human-readable results
        print("ffire benchmark: %s")
        print("Iterations:  \(iterations)")
        print("Encode:      \(encodeNs) ns/op")
        print("Decode:      \(decodeNs) ns/op")
        print("Total:       \(totalNs) ns/op")
        print("Wire size:   \(encodedData?.count ?? 0) bytes")
        print("Fixture:     \(fixtureData.count) bytes")
        let totalTimeS = Double(encodeTimeNs + decodeTimeNs) / 1_000_000_000.0
        print(String(format: "Total time:  %%.3fs", totalTimeS))
    }
} catch {
    fatalError("Benchmark failed: \(error)")
}
`,
		swiftModuleName, // import package (sanitized)
		iterations,
		swiftTypeName, // decode in warmup
		swiftTypeName, // decode in benchmark
		swiftTypeName, // decode for encode benchmark
		schemaName,    // message in JSON output
		schemaName)    // message in human output

	return buf.String()
}

// addBenchExecutableToPackage adds a bench executable target to Package.swift
func addBenchExecutableToPackage(swiftDir string, schema *schema.Schema, schemaName string) error {
	packagePath := filepath.Join(swiftDir, "Package.swift")
	content, err := os.ReadFile(packagePath)
	if err != nil {
		return err
	}

	// Check if bench executable already exists
	if bytes.Contains(content, []byte(`.executable(`)) {
		return nil // Already has executable
	}

	// Use schema's package name for the library and module name
	libName := schema.Package

	// Sanitize module name to match what's actually generated (package name, not schema filename)
	sanitizedName := generator.SanitizeSwiftModuleName(schema.Package)

	// Add bench executable to products and targets
	packageStr := string(content)

	// Add to products (after library product)
	packageBytes := bytes.Replace([]byte(packageStr),
		[]byte(`        .library(
            name: "`+sanitizedName+`",
            targets: ["`+sanitizedName+`"]
        ),`),
		[]byte(`        .library(
            name: "`+sanitizedName+`",
            targets: ["`+sanitizedName+`"]
        ),
        .executable(
            name: "bench",
            targets: ["bench"]
        ),`), 1)

	// Add to targets (after library target)
	targetInsert := fmt.Sprintf(`,
        .executableTarget(
            name: "bench",
            dependencies: ["%s"],
            path: "Sources/bench",
            linkerSettings: [
                .unsafeFlags(["-L", "lib"]),
                .linkedLibrary("%s")
            ]
        )`, sanitizedName, libName)

	packageBytes = bytes.ReplaceAll(packageBytes,
		[]byte(`        ),
    ]
)`),
		[]byte(`        )`+targetInsert+`
    ]
)`))

	return os.WriteFile(packagePath, packageBytes, 0644)
}

// generateSwiftRunScript generates a convenience run script
func generateSwiftRunScript(schemaName string) string {
	return `#!/bin/bash
# Convenience script to run Swift benchmark

# Check if swift is available
if ! command -v swift &> /dev/null; then
    echo "Error: swift not found"
    exit 1
fi

# Build and run the bench executable
export DYLD_LIBRARY_PATH=$(pwd)/lib:$DYLD_LIBRARY_PATH
swift run bench "$@"
`
}
