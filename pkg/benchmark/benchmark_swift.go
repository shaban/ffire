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

// Note: Swift module keyword sanitization is handled by generator.SanitizeSwiftModuleName()
// See pkg/generator/generator_swift.go for the shared swiftModuleKeywords map.
// This benchmark package generates native Swift benchmarks that use the unsafe pointer-based
// implementation with zero-copy operations and @inlinable functions.

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

	// Step 4: Create test directory for executable
	testDir := filepath.Join(swiftDir, "Sources", "test_bench")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return fmt.Errorf("failed to create test_bench directory: %w", err)
	}

	// Step 5: Generate the benchmark harness
	benchmarkCode := generateSwiftBenchmarkCode(schema, schemaName, messageName, iterations)
	testPath := filepath.Join(testDir, "main.swift")
	if err := os.WriteFile(testPath, []byte(benchmarkCode), 0644); err != nil {
		return fmt.Errorf("failed to write benchmark: %w", err)
	}

	// Step 6: Update Package.swift to add test executable
	if err := addTestExecutableToPackage(swiftDir, schema, schemaName); err != nil {
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

// generateSwiftBenchmarkCode generates the benchmark harness code using native Swift
func generateSwiftBenchmarkCode(s *schema.Schema, schemaName, messageName string, iterations int) string {
	// Find the message to get its target type
	var msg *schema.MessageType
	for i := range s.Messages {
		if s.Messages[i].Name == messageName {
			msg = &s.Messages[i]
			break
		}
	}
	if msg == nil {
		// Fallback: use message name
		return generateSwiftBenchNative(s, messageName, schemaName, iterations)
	}

	// Use message name for function naming (matches native Swift generator)
	return generateSwiftBenchNative(s, messageName, schemaName, iterations)
}

// generateSwiftBenchNative generates native Swift benchmark code (no C++ interop)
func generateSwiftBenchNative(s *schema.Schema, messageName, schemaName string, iterations int) string {
	// Sanitize module name to match generated package
	moduleName := generator.SanitizeSwiftModuleName(s.Package)

	buf := &bytes.Buffer{}

	// Import native Swift module directly
	fmt.Fprintf(buf, `import Foundation
import %s

// Load fixture
guard let fixtureData = try? Data(contentsOf: URL(fileURLWithPath: "fixture.bin")) else {
    fatalError("Failed to load fixture.bin")
}

let iterations = %d
let jsonOutput = ProcessInfo.processInfo.environment["BENCH_JSON"] == "1"

// Warmup - call native Swift functions
do {
    for _ in 0..<1000 {
        let decoded = try decode%sMessage(fixtureData)
        let _ = encode%sMessage(decoded)
    }
} catch {
    fatalError("Warmup failed: \(error)")
}

// Benchmark decode - native Swift with unsafe pointers
let decodeStart = DispatchTime.now()
do {
    for _ in 0..<iterations {
        let _ = try decode%sMessage(fixtureData)
    }
} catch {
    fatalError("Decode benchmark failed: \(error)")
}
let decodeEnd = DispatchTime.now()
let decodeTimeNs = decodeEnd.uptimeNanoseconds - decodeStart.uptimeNanoseconds

// Benchmark encode - decode once, then encode many times
let decoded: %sMessage
do {
    decoded = try decode%sMessage(fixtureData)
} catch {
    fatalError("Failed to decode for encode benchmark: \(error)")
}

let encodeStart = DispatchTime.now()
var wireSize = 0
for _ in 0..<iterations {
    let encoded = encode%sMessage(decoded)
    wireSize = encoded.count  // Get size from last iteration
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
        "wire_size": wireSize,
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
    print("Wire size:   \(wireSize) bytes")
    print("Fixture:     \(fixtureData.count) bytes")
    let totalTimeS = Double(encodeTimeNs + decodeTimeNs) / 1_000_000_000.0
    print(String(format: "Total time:  %%.3fs", totalTimeS))
}
`,
		moduleName,  // import native Swift module
		iterations,
		messageName, messageName, // warmup decode/encode
		messageName, // benchmark decode
		messageName, messageName, // decode for encode (with type annotation)
		messageName, // benchmark encode
		schemaName,  // message in JSON output
		schemaName)  // message in human output

	return buf.String()
}

// addTestExecutableToPackage adds a test executable target to Package.swift for native Swift benchmarks
func addTestExecutableToPackage(swiftDir string, schema *schema.Schema, schemaName string) error {
	packagePath := filepath.Join(swiftDir, "Package.swift")
	content, err := os.ReadFile(packagePath)
	if err != nil {
		return err
	}

	// Check if test_bench target already exists
	if bytes.Contains(content, []byte(`name: "test"`)) && bytes.Contains(content, []byte(`targets: ["test_bench"]`)) {
		return nil // Already has test executable target
	}

	// Sanitize module name to match what's actually generated (package name, not schema filename)
	sanitizedName := generator.SanitizeSwiftModuleName(schema.Package)

	// Add test executable to products and targets
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
            targets: ["test_bench"]
        ),`), 1)

	// Add to targets - native Swift only, no C++ interop needed
	targetInsert := fmt.Sprintf(`,
        .executableTarget(
            name: "test_bench",
            dependencies: ["%s"],
            path: "Sources/test_bench"
        )`, sanitizedName)

	// Add test_bench target after the library target
	// Match the pattern with library target closing
	packageBytes = bytes.ReplaceAll(packageBytes,
		[]byte(`        .target(
            name: "`+sanitizedName+`",
            dependencies: [],
            path: "Sources/`+sanitizedName+`"
        ),`),
		[]byte(`        .target(
            name: "`+sanitizedName+`",
            dependencies: [],
            path: "Sources/`+sanitizedName+`"
        )`+targetInsert+`,`))

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
