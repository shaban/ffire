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

// GenerateZig generates a Zig benchmark with embedded fixture
func GenerateZig(schema *schema.Schema, schemaName, messageName string, jsonData []byte, outputDir string, iterations int) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Step 1: Generate the Zig package
	config := &generator.PackageConfig{
		Schema:    schema,
		Language:  "zig",
		OutputDir: outputDir,
		Namespace: schema.Package,
		Optimize:  2,
		Platform:  "current",
		Arch:      "current",
		NoCompile: false,
		Verbose:   false,
	}

	if err := generator.GeneratePackage(config); err != nil {
		return fmt.Errorf("failed to generate Zig package: %w", err)
	}

	// Step 2: Convert JSON to binary fixture
	binaryData, err := fixture.Convert(schema, messageName, jsonData)
	if err != nil {
		return fmt.Errorf("failed to convert JSON to binary: %w", err)
	}

	// Step 3: Write the binary fixture
	zigDir := filepath.Join(outputDir, "zig")
	fixturePath := filepath.Join(zigDir, "fixture.bin")
	if err := os.WriteFile(fixturePath, binaryData, 0644); err != nil {
		return fmt.Errorf("failed to write fixture: %w", err)
	}

	// Step 4: Generate the benchmark main.zig
	benchmarkCode := generateZigBenchmarkCode(schema.Package, messageName, iterations)
	mainPath := filepath.Join(zigDir, "src", "main.zig")
	if err := os.WriteFile(mainPath, []byte(benchmarkCode), 0644); err != nil {
		return fmt.Errorf("failed to write benchmark: %w", err)
	}

	// Step 5: Generate a run script for convenience
	runScript := generateZigRunScript()
	runPath := filepath.Join(zigDir, "run.sh")
	if err := os.WriteFile(runPath, []byte(runScript), 0755); err != nil {
		return fmt.Errorf("failed to write run script: %w", err)
	}

	return nil
}

// generateZigBenchmarkCode generates the benchmark harness code
func generateZigBenchmarkCode(schemaName, messageName string, iterations int) string {
	buf := &bytes.Buffer{}

	// Add Message suffix to type name (generator adds this suffix)
	typeName := messageName + "Message"

	// Use alias if schema name conflicts with Zig keywords
	importAlias := schemaName
	if isZigKeyword(schemaName) {
		importAlias = "ffire_" + schemaName
	}

	fmt.Fprintf(buf, `const std = @import("std");
const %s = @import("%s");

pub fn main() !void {
    // Read fixture
    const cwd = std.fs.cwd();
    const file = cwd.openFile("fixture.bin", .{}) catch |err| {
        std.debug.print("Failed to open fixture.bin: {}\n", .{err});
        return err;
    };
    defer file.close();

    const stat = try file.stat();
    const allocator = std.heap.page_allocator;
    const fixtureData = try allocator.alloc(u8, stat.size);
    defer allocator.free(fixtureData);
    _ = try file.readAll(fixtureData);

    const iterations: usize = %d;
    const jsonOutput = if (std.posix.getenv("BENCH_JSON")) |_| true else false;

    // Warmup
    var i: usize = 0;
    while (i < 1000) : (i += 1) {
        const msg = try %s.%s.decode(fixtureData);
        const encoded = try msg.encode();
        %s.%s.freeEncodedData(encoded);
        msg.deinit();
    }

    // Benchmark decode
    const decodeStart = std.time.nanoTimestamp();
    i = 0;
    while (i < iterations) : (i += 1) {
        const msg = try %s.%s.decode(fixtureData);
        msg.deinit();
    }
    const decodeEnd = std.time.nanoTimestamp();
    const decodeTime: u64 = @intCast(decodeEnd - decodeStart);

    // Benchmark encode (decode once, then encode many times)
    const msg = try %s.%s.decode(fixtureData);
    defer msg.deinit();

    const encodeStart = std.time.nanoTimestamp();
    var encoded: []u8 = undefined;
    i = 0;
    while (i < iterations) : (i += 1) {
        encoded = try msg.encode();
        if (i < iterations - 1) {
            %s.%s.freeEncodedData(encoded);
        }
    }
    const encodeEnd = std.time.nanoTimestamp();
    const encodeTime: u64 = @intCast(encodeEnd - encodeStart);

    // Calculate metrics
    const encodeNs = @divFloor(encodeTime, iterations);
    const decodeNs = @divFloor(decodeTime, iterations);
    const totalNs = encodeNs + decodeNs;

    const stdout = std.fs.File.stdout().deprecatedWriter();
    if (jsonOutput) {
        // Output JSON for automation
        try stdout.print(
            \\{{"language": "Zig", "format": "ffire", "message": "%s", "iterations": {d}, "encode_ns": {d}, "decode_ns": {d}, "total_ns": {d}, "wire_size": {d}, "fixture_size": {d}}}
            \\
        , .{ iterations, encodeNs, decodeNs, totalNs, encoded.len, fixtureData.len });
    } else {
        // Print human-readable results
        try stdout.print("ffire benchmark: %s\n", .{});
        try stdout.print("Iterations:  {d}\n", .{iterations});
        try stdout.print("Encode:      {d} ns/op\n", .{encodeNs});
        try stdout.print("Decode:      {d} ns/op\n", .{decodeNs});
        try stdout.print("Total:       {d} ns/op\n", .{totalNs});
        try stdout.print("Wire size:   {d} bytes\n", .{encoded.len});
        try stdout.print("Fixture:     {d} bytes\n", .{fixtureData.len});
    }

    %s.%s.freeEncodedData(encoded);
}
`, importAlias, schemaName,
		iterations,
		importAlias, typeName,
		importAlias, typeName,
		importAlias, typeName,
		importAlias, typeName,
		importAlias, typeName,
		messageName,
		messageName,
		importAlias, typeName)

	return buf.String()
}

// isZigKeyword checks if a name is a Zig reserved keyword
func isZigKeyword(name string) bool {
	keywords := map[string]bool{
		"test": true, "error": true, "type": true, "align": true,
		"and": true, "or": true, "break": true, "return": true,
		"continue": true, "defer": true, "else": true, "enum": true,
		"for": true, "if": true, "import": true, "pub": true,
		"struct": true, "switch": true, "union": true, "unreachable": true,
		"var": true, "while": true, "fn": true, "const": true,
	}
	return keywords[name]
}

// generateZigRunScript generates a convenience run script
func generateZigRunScript() string {
	return `#!/bin/bash
# Convenience script to run Zig benchmark
set -e

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Set library path
export DYLD_LIBRARY_PATH="$SCRIPT_DIR/lib:$DYLD_LIBRARY_PATH"
export LD_LIBRARY_PATH="$SCRIPT_DIR/lib:$LD_LIBRARY_PATH"

# Build in release mode
zig build -Doptimize=ReleaseFast

# Run the benchmark
./zig-out/bin/bench
`
}
