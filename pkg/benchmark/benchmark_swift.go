package benchmark

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shaban/ffire/pkg/fixture"
	"github.com/shaban/ffire/pkg/generator"
	"github.com/shaban/ffire/pkg/schema"
)

// Note: Swift module keyword sanitization is handled by generator.sanitizeSwiftModuleName()
// See pkg/generator/generator_swift.go for the shared swiftModuleKeywords map.

// rootTypeName returns the root type name for C++ function naming
// This matches the logic in pkg/generator/generator_cpp.go
func rootTypeName(typ schema.Type) string {
	switch t := typ.(type) {
	case *schema.PrimitiveType:
		return strings.Title(t.Name)
	case *schema.StructType:
		return t.Name
	case *schema.ArrayType:
		return rootTypeName(t.ElementType)
	default:
		return "Unknown"
	}
}

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

// generateSwiftBenchmarkCode generates the benchmark harness code using C++ interop
func generateSwiftBenchmarkCode(s *schema.Schema, schemaName, messageName string, iterations int) string {
	// C++ module name: {package}_cpp_{package}
	cppModuleName := fmt.Sprintf("%s_cpp_%s", s.Package, s.Package)

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
		funcSuffix := strings.ToLower(messageName)
		return generateSwiftBench(s, cppModuleName, funcSuffix, schemaName, iterations)
	}

	// Function name suffix: computed from root type (matches C++ generator logic)
	funcSuffix := strings.ToLower(rootTypeName(msg.TargetType))
	return generateSwiftBench(s, cppModuleName, funcSuffix, schemaName, iterations)
}

func generateSwiftBench(s *schema.Schema, cppModuleName, funcSuffix, schemaName string, iterations int) string {
	buf := &bytes.Buffer{}

	// Import C++ module directly - Swift uses C++ types with zero translation
	fmt.Fprintf(buf, `import Foundation
import %s

// Load fixture
guard let fixtureData = try? Data(contentsOf: URL(fileURLWithPath: "fixture.bin")) else {
    fatalError("Failed to load fixture.bin")
}

let iterations = %d
let jsonOutput = ProcessInfo.processInfo.environment["BENCH_JSON"] == "1"

// Use C++ function signature verbatim: decode_plugin_message(const uint8_t* data, size_t size)
let fixtureBytes = [UInt8](fixtureData)

// Warmup - call C++ functions with exact signature from header
for _ in 0..<1000 {
    let decoded = fixtureBytes.withUnsafeBufferPointer { ptr in
        %s.decode_%s_message(ptr.baseAddress!, fixtureBytes.count)
    }
    let _ = %s.encode_%s_message(decoded)
}

// Benchmark decode - C++ signature: std::vector<T> decode(const uint8_t*, size_t)
// Use batched autoreleasepool to reduce ARC overhead (4%% faster than no pool)
let decodeStart = DispatchTime.now()
fixtureBytes.withUnsafeBufferPointer { ptr in
    let batchSize = 1000
    for _ in 0..<(iterations / batchSize) {
        autoreleasepool {
            for _ in 0..<batchSize {
                let _ = %s.decode_%s_message(ptr.baseAddress!, fixtureBytes.count)
            }
        }
    }
}
let decodeEnd = DispatchTime.now()
let decodeTimeNs = decodeEnd.uptimeNanoseconds - decodeStart.uptimeNanoseconds

// Benchmark encode - decode once, then encode many times
let decoded = fixtureBytes.withUnsafeBufferPointer { ptr in
    %s.decode_%s_message(ptr.baseAddress!, fixtureBytes.count)
}

// Use autoreleasepool per iteration for encode (3%% faster)
let encodeStart = DispatchTime.now()
var wireSize = 0
for _ in 0..<iterations {
    autoreleasepool {
        let encoded = %s.encode_%s_message(decoded)
        wireSize = Int(encoded.size())  // Get size from last iteration
    }
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
		cppModuleName, // import C++ module
		iterations,
		s.Package, funcSuffix, // warmup decode
		s.Package, funcSuffix, // warmup encode
		s.Package, funcSuffix, // benchmark decode
		s.Package, funcSuffix, // decode for encode benchmark
		s.Package, funcSuffix, // benchmark encode
		schemaName, // message in JSON output
		schemaName) // message in human output

	return buf.String()
}

// addTestExecutableToPackage adds a test executable target to Package.swift with C++ interop
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

	// Add to targets with C++ interop settings
	// test_bench target needs C++ interop to import the C++ module
	targetInsert := fmt.Sprintf(`,
        .executableTarget(
            name: "test_bench",
            dependencies: ["%s"],
            path: "Sources/test_bench",
            swiftSettings: [
                .interoperabilityMode(.Cxx)
            ]
        )`, sanitizedName)

	// Add test_bench target after the library target (before closing of targets array)
	// Match the closing of the last target in the targets array
	packageBytes = bytes.ReplaceAll(packageBytes,
		[]byte(`        ),
    ],
    cxxLanguageStandard: .cxx17`),
		[]byte(`        )`+targetInsert+`,
    ],
    cxxLanguageStandard: .cxx17`))

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
