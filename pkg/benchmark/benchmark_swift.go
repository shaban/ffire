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

// GenerateSwift generates a Swift benchmark with embedded fixture
func GenerateSwift(schema *schema.Schema, schemaName, messageName string, jsonData []byte, outputDir string, iterations int) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Step 1: Generate the Swift package
	config := &generator.PackageConfig{
		Schema:    schema,
		Language:  "swift",
		OutputDir: outputDir,
		Namespace: schemaName,
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

	// Step 4: Generate the benchmark harness
	benchmarkCode := generateSwiftBenchmarkCode(schemaName, messageName, iterations)
	benchPath := filepath.Join(swiftDir, "bench.swift")
	if err := os.WriteFile(benchPath, []byte(benchmarkCode), 0644); err != nil {
		return fmt.Errorf("failed to write benchmark: %w", err)
	}

	// Step 5: Generate a run script for convenience
	runScript := generateSwiftRunScript(schemaName)
	runPath := filepath.Join(swiftDir, "run.sh")
	if err := os.WriteFile(runPath, []byte(runScript), 0755); err != nil {
		return fmt.Errorf("failed to write run script: %w", err)
	}

	return nil
}

// generateSwiftBenchmarkCode generates the benchmark harness code
func generateSwiftBenchmarkCode(schemaName, messageName string, iterations int) string {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, `import Foundation

#if os(macOS)
typealias LibHandle = UnsafeMutableRawPointer
#else
typealias LibHandle = UnsafeMutableRawPointer?
#endif

// Load shared library
let libName: String
#if os(macOS)
libName = "lib%s.dylib"
#elseif os(Windows)
libName = "%s.dll"
#else
libName = "lib%s.so"
#endif

guard let lib = dlopen(libName, RTLD_LAZY) else {
    fatalError("Failed to load library: \(String(cString: dlerror()))")
}

// FFI function signatures
typealias EncodeFunc = @convention(c) (UnsafeRawPointer?, Int32, UnsafeMutablePointer<Int32>?) -> UnsafeMutableRawPointer?
typealias DecodeFunc = @convention(c) (UnsafeRawPointer?, Int32) -> UnsafeMutableRawPointer?
typealias FreeFunc = @convention(c) (UnsafeMutableRawPointer?) -> Void

guard let encodeSymbol = dlsym(lib, "ffire_encode_%s"),
      let decodeSymbol = dlsym(lib, "ffire_decode_%s"),
      let freeSymbol = dlsym(lib, "ffire_free_%s") else {
    fatalError("Failed to load symbols")
}

let encode = unsafeBitCast(encodeSymbol, to: EncodeFunc.self)
let decode = unsafeBitCast(decodeSymbol, to: DecodeFunc.self)
let freeMessage = unsafeBitCast(freeSymbol, to: FreeFunc.self)

// Load fixture
guard let fixtureData = try? Data(contentsOf: URL(fileURLWithPath: "fixture.bin")) else {
    fatalError("Failed to load fixture.bin")
}

let iterations = %d
let jsonOutput = ProcessInfo.processInfo.environment["BENCH_JSON"] == "1"

// Warmup
for _ in 0..<1000 {
    fixtureData.withUnsafeBytes { dataPtr in
        guard let msgPtr = decode(dataPtr.baseAddress, Int32(fixtureData.count)) else {
            fatalError("Decode failed during warmup")
        }
        var size: Int32 = 0
        _ = encode(msgPtr, 0, &size)
        freeMessage(msgPtr)
    }
}

// Benchmark decode
let decodeStart = DispatchTime.now()
for _ in 0..<iterations {
    fixtureData.withUnsafeBytes { dataPtr in
        guard let msgPtr = decode(dataPtr.baseAddress, Int32(fixtureData.count)) else {
            fatalError("Decode failed")
        }
        freeMessage(msgPtr)
    }
}
let decodeEnd = DispatchTime.now()
let decodeTimeNs = decodeEnd.uptimeNanoseconds - decodeStart.uptimeNanoseconds

// Benchmark encode (decode once, then encode many times)
var encodedData: Data?
var msgPtr: UnsafeMutableRawPointer?

fixtureData.withUnsafeBytes { dataPtr in
    msgPtr = decode(dataPtr.baseAddress, Int32(fixtureData.count))
}

guard let validMsgPtr = msgPtr else {
    fatalError("Decode failed for encode benchmark")
}

let encodeStart = DispatchTime.now()
for _ in 0..<iterations {
    var size: Int32 = 0
    if let resultPtr = encode(validMsgPtr, 0, &size) {
        encodedData = Data(bytes: resultPtr, count: Int(size))
    }
}
let encodeEnd = DispatchTime.now()
let encodeTimeNs = encodeEnd.uptimeNanoseconds - encodeStart.uptimeNanoseconds

freeMessage(validMsgPtr)

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

dlclose(lib)
`, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName, iterations, schemaName, schemaName)

	return buf.String()
}

// generateSwiftRunScript generates a convenience run script
func generateSwiftRunScript(schemaName string) string {
	return fmt.Sprintf(`#!/bin/bash
# Convenience script to run Swift benchmark

# Check if swift is available
if ! command -v swift &> /dev/null; then
    echo "Error: swift not found"
    exit 1
fi

# Compile and run
swift bench.swift -I. -L. -l%s "$@"
`, schemaName)
}
