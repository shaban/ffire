//go:build mage || tools
// +build mage tools

// Magefile for cross-language benchmark comparison
//
// Usage:
//
//	mage gen {target}    - Generate benchmarks for specific language or 'all'
//	                       Targets: all, go, cpp, java, python, dart, swift, js, proto
//	                       Example: mage gen java
//
//	mage run {target}    - Run benchmarks for specific language or 'all'
//	                       Targets: all, go, cpp, java, python, dart, swift, js, proto
//	                       Example: mage run go
//
//	mage clean {target}  - Clean generated files for specific language or 'all'
//	                       Targets: all, go, cpp, java, python, dart, swift, js, proto
//	                       Example: mage clean cpp
//
//	mage compare         - Generate comparison table from all benchmark results
//	                       Shows performance metrics across all languages and formats
//
//	mage graph {target}  - Display visual decode time comparison
//	                       Targets: all (average), or specific schema name (struct, array_int, etc.)
//	                       Example: mage graph struct
//	                       Shows all language implementations with performance bars
//
//	mage bench           - Full workflow: generate all → run all → compare
//	                       Comprehensive benchmark suite across all languages
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/magefile/mage/sh"
)

const (
	schemaDir  = "../testdata/schema"
	jsonDir    = "../testdata/json"
	protoDir   = "../testdata/proto"
	genDir     = "generated"
	resultsDir = "results"
)

type BenchmarkSuite struct {
	Name       string
	SchemaFile string
	JSONFile   string
	ProtoFile  string
}

func discoverBenchmarks() ([]BenchmarkSuite, error) {
	// Find all .ffi files
	schemaFiles, err := filepath.Glob(filepath.Join(schemaDir, "*.ffi"))
	if err != nil {
		return nil, err
	}

	var suites []BenchmarkSuite
	for _, schemaFile := range schemaFiles {
		base := filepath.Base(schemaFile)
		name := strings.TrimSuffix(base, ".ffi")

		jsonFile := filepath.Join(jsonDir, name+".json")
		protoFile := filepath.Join(protoDir, name+".proto")

		// Check if JSON file exists (required)
		if _, err := os.Stat(jsonFile); err != nil {
			fmt.Printf("  ⚠️  Skipping %s: no JSON fixture\n", name)
			continue
		}

		suites = append(suites, BenchmarkSuite{
			Name:       name,
			SchemaFile: schemaFile,
			JSONFile:   jsonFile,
			ProtoFile:  protoFile,
		})
	}

	return suites, nil
}

// BenchResult holds benchmark results in standardized format
type BenchResult struct {
	Language    string `json:"language"`
	Format      string `json:"format"`
	Message     string `json:"message"`
	Iterations  int    `json:"iterations"`
	EncodeNs    int64  `json:"encode_ns"`
	DecodeNs    int64  `json:"decode_ns"`
	TotalNs     int64  `json:"total_ns"`
	WireSize    int    `json:"wire_size"`
	FixtureSize int    `json:"fixture_size"`
	Timestamp   string `json:"timestamp"`
}

// cleanAll removes all generated files AND results (private helper)
func cleanAll() error {
	fmt.Println("🧹 Cleaning all generated files and results...")
	os.RemoveAll(genDir)
	os.RemoveAll(resultsDir)
	return nil
}

// genAll generates all benchmark variants (private helper)
func genAll() error {
	// Clean generated directory first
	fmt.Println("🧹 Cleaning generated files...")
	os.RemoveAll(genDir)

	if err := os.MkdirAll(genDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return err
	}

	// Discover all benchmarks
	fmt.Println("� Discovering benchmark suites...")
	suites, err := discoverBenchmarks()
	if err != nil {
		return err
	}

	fmt.Printf("Found %d benchmark suites\n\n", len(suites))

	// Ensure ffire is built and installed
	fmt.Println("🔨 Building ffire...")
	if err := sh.RunV("sh", "-c", "cd .. && go install -buildvcs=false ./cmd/ffire"); err != nil {
		return fmt.Errorf("failed to build ffire: %w", err)
	}

	// Generate ffire Go benchmarks
	for _, suite := range suites {
		fmt.Printf("🔧 Generating ffire Go benchmark: %s\n", suite.Name)
		if err := sh.Run("ffire", "bench",
			"--lang", "go",
			"--schema", suite.SchemaFile,
			"--json", suite.JSONFile,
			"--output", filepath.Join(genDir, "ffire_"+suite.Name),
			"--iterations", "100000",
		); err != nil {
			fmt.Printf("  ⚠️  Skipping %s: %v\n", suite.Name, err)
			continue
		}
	}

	// Generate ffire C++ benchmarks
	for _, suite := range suites {
		fmt.Printf("🔨 Generating ffire C++ benchmark: %s\n", suite.Name)
		if err := sh.Run("ffire", "bench",
			"--lang", "cpp",
			"--schema", suite.SchemaFile,
			"--json", suite.JSONFile,
			"--output", filepath.Join(genDir, "ffire_cpp_"+suite.Name),
			"--iterations", "100000",
		); err != nil {
			fmt.Printf("  ⚠️  Skipping %s: %v\n", suite.Name, err)
			continue
		}
	}

	// NOTE: Python and JavaScript excluded from 'all' - use explicit targets
	// Generate them with: mage gen python, mage gen javascript

	// Generate ffire Dart benchmarks
	for _, suite := range suites {
		fmt.Printf("🎯 Generating ffire Dart benchmark: %s\n", suite.Name)
		if err := sh.Run("ffire", "bench",
			"--lang", "dart",
			"--schema", suite.SchemaFile,
			"--json", suite.JSONFile,
			"--output", filepath.Join(genDir, "ffire_dart_"+suite.Name),
			"--iterations", "100000",
		); err != nil {
			fmt.Printf("  ⚠️  Skipping %s: %v\n", suite.Name, err)
			continue
		}
	}

	// Generate ffire Swift benchmarks
	for _, suite := range suites {
		fmt.Printf("🍎 Generating ffire Swift benchmark: %s\n", suite.Name)
		if err := sh.Run("ffire", "bench",
			"--lang", "swift",
			"--schema", suite.SchemaFile,
			"--json", suite.JSONFile,
			"--output", filepath.Join(genDir, "ffire_swift_"+suite.Name),
			"--iterations", "100000",
		); err != nil {
			fmt.Printf("  ⚠️  Skipping %s: %v\n", suite.Name, err)
			continue
		}
	}

	// Generate ffire Java benchmarks
	for _, suite := range suites {
		fmt.Printf("☕ Generating ffire Java benchmark: %s\n", suite.Name)
		if err := sh.Run("ffire", "bench",
			"--lang", "java",
			"--schema", suite.SchemaFile,
			"--json", suite.JSONFile,
			"--output", filepath.Join(genDir, "ffire_java_"+suite.Name),
			"--iterations", "100000",
		); err != nil {
			fmt.Printf("  ⚠️  Skipping %s: %v\n", suite.Name, err)
			continue
		}
	}

	// Generate ffire C# benchmarks
	for _, suite := range suites {
		fmt.Printf("💜 Generating ffire C# benchmark: %s\n", suite.Name)
		if err := sh.Run("ffire", "bench",
			"--lang", "csharp",
			"--schema", suite.SchemaFile,
			"--json", suite.JSONFile,
			"--output", filepath.Join(genDir, "ffire_csharp_"+suite.Name),
			"--iterations", "100000",
		); err != nil {
			fmt.Printf("  ⚠️  Skipping %s: %v\n", suite.Name, err)
			continue
		}
	}

	// Generate proto benchmarks (only for those with .proto files)
	for _, suite := range suites {
		if _, err := os.Stat(suite.ProtoFile); err == nil {
			fmt.Printf("📦 Generating proto benchmark: %s\n", suite.Name)
			if err := genProto(suite.Name, suite.ProtoFile, suite.JSONFile); err != nil {
				return fmt.Errorf("failed to generate proto benchmark for %s: %w", suite.Name, err)
			}
		}
	}

	fmt.Println("\n✅ All benchmarks generated")
	return nil
}

// genProto generates proto benchmark
func genProto(name, protoFile, jsonFile string) error {
	outDir := filepath.Join(genDir, "proto_"+name)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	// Generate Go code from proto
	fmt.Println("  Running protoc...")
	if err := sh.Run("protoc",
		"--go_out="+outDir,
		"--proto_path="+protoDir,
		protoFile,
	); err != nil {
		return fmt.Errorf("protoc failed: %w", err)
	}

	// Generate benchmark driver
	fmt.Println("  Generating benchmark driver...")
	return generateProtoBenchmark(name, outDir, jsonFile)
}

// Gen generates benchmarks for the specified target
// Usage:
//
//	mage gen all      - Generate all languages
//	mage gen go       - Generate only Go benchmarks
//	mage gen java     - Generate only Java benchmarks
//	mage gen cpp      - Generate only C++ benchmarks
//	(supports: go, cpp, java, python, dart, swift, javascript/js, proto)
func Gen(target string) error {
	target = strings.ToLower(target)

	// Handle 'all' target
	if target == "all" {
		return genAll()
	}

	// Handle 'js' alias
	if target == "js" {
		target = "javascript"
	}

	// Validate target
	validTargets := map[string]bool{
		"go": true, "cpp": true, "java": true, "csharp": true, "python": true, "python-pybind11": true,
		"dart": true, "swift": true, "javascript": true, "proto": true,
	}

	if !validTargets[target] {
		return fmt.Errorf("unknown target: %s\nValid targets: all, go, cpp, java, csharp, python, python-pybind11, dart, swift, javascript, proto", target)
	}

	// Create output directories
	if err := os.MkdirAll(genDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		return err
	}

	// Build ffire
	fmt.Println("🔨 Building ffire...")
	if err := sh.RunV("sh", "-c", "cd .. && go install -buildvcs=false ./cmd/ffire"); err != nil {
		return fmt.Errorf("failed to build ffire: %w", err)
	}

	// Discover benchmarks
	suites, err := discoverBenchmarks()
	if err != nil {
		return err
	}

	// Generate for the target language
	return genLanguage(target, suites)
}

// Run runs benchmarks for the specified target
// Usage:
//
//	mage run all      - Run all language benchmarks
//	mage run go       - Run only Go benchmarks
//	mage run java     - Run only Java benchmarks
//	(supports: go, cpp, java, python, dart, swift, javascript/js, proto)
func Run(target string) error {
	target = strings.ToLower(target)

	// Handle 'all' target - run stable languages (exclude js, python)
	if target == "all" {
		fmt.Println("🏃 Running all benchmarks (stable languages)...")
		fmt.Println("    Note: Python and JavaScript excluded - run with explicit targets")

		if err := runGo(); err != nil {
			fmt.Printf("⚠️  Go benchmarks failed: %v\n", err)
		}
		if err := runCpp(); err != nil {
			fmt.Printf("⚠️  C++ benchmarks failed: %v\n", err)
		}
		if err := runJava(); err != nil {
			fmt.Printf("⚠️  Java benchmarks failed: %v\n", err)
		}
		if err := runCSharp(); err != nil {
			fmt.Printf("⚠️  C# benchmarks failed: %v\n", err)
		}
		if err := runDart(); err != nil {
			fmt.Printf("⚠️  Dart benchmarks failed: %v\n", err)
		}
		if err := runSwift(); err != nil {
			fmt.Printf("⚠️  Swift benchmarks failed: %v\n", err)
		}
		if err := runProto(); err != nil {
			fmt.Printf("⚠️  Proto benchmarks failed: %v\n", err)
		}

		return nil
	}

	// Handle 'js' alias
	if target == "js" {
		target = "javascript"
	}

	// Route to specific runner
	switch target {
	case "go":
		return runGo()
	case "cpp":
		return runCpp()
	case "java":
		return runJava()
	case "csharp":
		return runCSharp()
	case "python":
		return runPython()
	case "dart":
		return runDart()
	case "swift":
		return runSwift()
	case "javascript":
		return runJavaScript()
	case "proto":
		return runProto()
	default:
		return fmt.Errorf("unknown target: %s\nValid targets: all, go, cpp, java, csharp, python, dart, swift, javascript, proto", target)
	}
}

// Clean removes generated files for the specified target
// Usage:
//
//	mage clean all    - Remove all generated files (keeps results/)
//	mage clean go     - Remove only Go generated files
//	mage clean java   - Remove only Java generated files
func Clean(target string) error {
	target = strings.ToLower(target)

	// Handle 'all' target
	if target == "all" {
		fmt.Println("🧹 Cleaning all generated files...")
		return os.RemoveAll(genDir)
	}

	// Handle 'js' alias
	if target == "js" {
		target = "javascript"
	}

	// Validate target
	validTargets := map[string]bool{
		"go": true, "cpp": true, "java": true, "csharp": true, "python": true,
		"dart": true, "swift": true, "javascript": true, "proto": true,
	}

	if !validTargets[target] {
		return fmt.Errorf("unknown target: %s\nValid targets: all, go, cpp, java, python, dart, swift, javascript, proto", target)
	}

	// Remove language-specific generated files
	fmt.Printf("🧹 Cleaning %s generated files...\n", target)

	var patterns []string
	if target == "go" {
		// Go benchmarks don't have language prefix
		patterns = []string{filepath.Join(genDir, "ffire_*")}
	} else if target == "proto" {
		patterns = []string{filepath.Join(genDir, "proto_*")}
	} else {
		patterns = []string{filepath.Join(genDir, fmt.Sprintf("ffire_%s_*", target))}
	}

	for _, pattern := range patterns {
		dirs, err := filepath.Glob(pattern)
		if err != nil {
			return err
		}

		for _, dir := range dirs {
			// For Go, skip other language variants
			if target == "go" {
				base := filepath.Base(dir)
				if strings.HasPrefix(base, "ffire_cpp_") ||
					strings.HasPrefix(base, "ffire_python_") ||
					strings.HasPrefix(base, "ffire_dart_") ||
					strings.HasPrefix(base, "ffire_swift_") ||
					strings.HasPrefix(base, "ffire_javascript_") ||
					strings.HasPrefix(base, "ffire_java_") {
					continue
				}
			}

			fmt.Printf("  Removing %s\n", dir)
			if err := os.RemoveAll(dir); err != nil {
				return err
			}
		}
	}

	return nil
}

// runGo runs the Go benchmarks
func runGo() error {
	fmt.Println("\n🏃 Running ffire Go benchmarks...")

	// Find all Go ffire benchmark directories (exclude cpp and python variants)
	pattern := filepath.Join(genDir, "ffire_*")
	allDirs, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	// Filter to only pure Go benchmarks (exclude other language variants)
	var dirs []string
	for _, dir := range allDirs {
		base := filepath.Base(dir)
		if !strings.HasPrefix(base, "ffire_cpp_") &&
			!strings.HasPrefix(base, "ffire_python_") &&
			!strings.HasPrefix(base, "ffire_dart_") &&
			!strings.HasPrefix(base, "ffire_swift_") &&
			!strings.HasPrefix(base, "ffire_javascript_") &&
			!strings.HasPrefix(base, "ffire_java_") &&
			!strings.HasPrefix(base, "ffire_csharp_") {
			dirs = append(dirs, dir)
		}
	}

	if len(dirs) == 0 {
		fmt.Println("  ⚠️  No Go benchmarks found (skipping)")
		return nil
	}

	var allResults []BenchResult
	for _, dir := range dirs {
		name := strings.TrimPrefix(filepath.Base(dir), "ffire_")
		fmt.Printf("\n  Testing: %s\n", name)

		result, err := runGoBench(dir)
		if err != nil {
			fmt.Printf("  ❌ Failed: %v\n", err)
			continue
		}

		// Print result
		fmt.Printf("  ✓ Encode: %d ns/op\n", result.EncodeNs)
		fmt.Printf("  ✓ Decode: %d ns/op\n", result.DecodeNs)
		fmt.Printf("  ✓ Total:  %d ns/op\n", result.TotalNs)
		fmt.Printf("  ✓ Size:   %d bytes\n", result.WireSize)

		allResults = append(allResults, result)
	}

	// Save all results
	return saveResults(allResults, "ffire_go")
}

// runProto runs the proto benchmarks
func runProto() error {
	fmt.Println("\n🏃 Running proto benchmarks...")

	// Find all proto benchmark directories
	pattern := filepath.Join(genDir, "proto_*")
	dirs, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(dirs) == 0 {
		fmt.Println("  ⚠️  No proto benchmarks found (skipping)")
		return nil
	}

	var allResults []BenchResult
	for _, dir := range dirs {
		name := strings.TrimPrefix(filepath.Base(dir), "proto_")
		fmt.Printf("\n  Testing: %s\n", name)

		result, err := runGoBench(dir)
		if err != nil {
			fmt.Printf("  ❌ Failed: %v\n", err)
			continue
		}

		// Print result
		fmt.Printf("  ✓ Encode: %d ns/op\n", result.EncodeNs)
		fmt.Printf("  ✓ Decode: %d ns/op\n", result.DecodeNs)
		fmt.Printf("  ✓ Total:  %d ns/op\n", result.TotalNs)
		fmt.Printf("  ✓ Size:   %d bytes\n", result.WireSize)

		allResults = append(allResults, result)
	}

	// Save all results
	return saveResults(allResults, "proto_go")
}

// runCpp runs the C++ benchmarks
func runCpp() error {
	fmt.Println("\n🏃 Running ffire C++ benchmarks...")

	// Find all C++ benchmark directories
	pattern := filepath.Join(genDir, "ffire_cpp_*")
	dirs, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(dirs) == 0 {
		fmt.Println("  ⚠️  No C++ benchmarks found (skipping)")
		return nil
	}

	var allResults []BenchResult
	for _, dir := range dirs {
		name := strings.TrimPrefix(filepath.Base(dir), "ffire_cpp_")
		fmt.Printf("\n  Testing: %s\n", name)

		result, err := runCppBench(dir)
		if err != nil {
			fmt.Printf("  ❌ Failed: %v\n", err)
			continue
		}

		// Print result
		fmt.Printf("  ✓ Encode: %d ns/op\n", result.EncodeNs)
		fmt.Printf("  ✓ Decode: %d ns/op\n", result.DecodeNs)
		fmt.Printf("  ✓ Total:  %d ns/op\n", result.TotalNs)
		fmt.Printf("  ✓ Size:   %d bytes\n", result.WireSize)

		allResults = append(allResults, result)
	}

	// Save all results
	return saveResults(allResults, "ffire_cpp")
}

// runPython runs the Pure Python benchmarks (now the default)
func runPython() error {
	fmt.Println("\n🏃 Running ffire Python benchmarks...")

	// Check if python3 is available
	if _, err := exec.LookPath("python3"); err != nil {
		fmt.Println("  ⚠️  python3 not found (skipping)")
		return nil
	}

	// Find all Python benchmark directories
	pattern := filepath.Join(genDir, "ffire_python_*")
	dirs, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(dirs) == 0 {
		fmt.Println("  ⚠️  No Python benchmarks found (skipping)")
		return nil
	}

	var allResults []BenchResult
	for _, dir := range dirs {
		name := strings.TrimPrefix(filepath.Base(dir), "ffire_python_")
		fmt.Printf("\n  Testing: %s\n", name)

		result, err := runPythonPureBench(dir)
		if err != nil {
			fmt.Printf("  ❌ Failed: %v\n", err)
			continue
		}

		// Override message name with schema name for consistent grouping
		result.Message = name

		// Print result
		fmt.Printf("  ✓ Encode: %d ns/op\n", result.EncodeNs)
		fmt.Printf("  ✓ Decode: %d ns/op\n", result.DecodeNs)
		fmt.Printf("  ✓ Total:  %d ns/op\n", result.TotalNs)
		fmt.Printf("  ✓ Size:   %d bytes\n", result.WireSize)

		allResults = append(allResults, result)
	}

	// Save all results
	return saveResults(allResults, "ffire_python")
}

// runPythonPyBind11 runs the PyBind11 benchmarks (legacy)
// runDart runs the Dart benchmarks
func runDart() error {
	fmt.Println("\n🏃 Running ffire Dart benchmarks...")

	// Check if dart is available
	if _, err := exec.LookPath("dart"); err != nil {
		fmt.Println("  ⚠️  dart not found (skipping)")
		return nil
	}

	// Find all Dart benchmark directories
	pattern := filepath.Join(genDir, "ffire_dart_*")
	dirs, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(dirs) == 0 {
		fmt.Println("  ⚠️  No Dart benchmarks found (skipping)")
		return nil
	}

	var allResults []BenchResult
	for _, dir := range dirs {
		name := strings.TrimPrefix(filepath.Base(dir), "ffire_dart_")
		fmt.Printf("\n  Testing: %s\n", name)

		result, err := runDartBench(dir)
		if err != nil {
			fmt.Printf("  ❌ Failed: %v\n", err)
			continue
		}

		// Override message name with schema name for consistent grouping
		result.Message = name

		// Print result
		fmt.Printf("  ✓ Encode: %d ns/op\n", result.EncodeNs)
		fmt.Printf("  ✓ Decode: %d ns/op\n", result.DecodeNs)
		fmt.Printf("  ✓ Total:  %d ns/op\n", result.TotalNs)
		fmt.Printf("  ✓ Size:   %d bytes\n", result.WireSize)

		allResults = append(allResults, result)
	}

	// Save all results
	return saveResults(allResults, "ffire_dart")
}

// runSwift runs the Swift benchmarks
func runSwift() error {
	fmt.Println("\n🏃 Running ffire Swift benchmarks...")

	// Check if swift is available
	if _, err := exec.LookPath("swift"); err != nil {
		fmt.Println("  ⚠️  swift not found (skipping)")
		return nil
	}

	// Find all Swift benchmark directories
	pattern := filepath.Join(genDir, "ffire_swift_*")
	dirs, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(dirs) == 0 {
		fmt.Println("  ⚠️  No Swift benchmarks found (skipping)")
		return nil
	}

	var allResults []BenchResult
	for _, dir := range dirs {
		name := strings.TrimPrefix(filepath.Base(dir), "ffire_swift_")
		fmt.Printf("\n  Testing: %s\n", name)

		result, err := runSwiftBench(dir)
		if err != nil {
			fmt.Printf("  ❌ Failed: %v\n", err)
			continue
		}

		// Override message name with schema name for consistent grouping
		result.Message = name

		// Print result
		fmt.Printf("  ✓ Encode: %d ns/op\n", result.EncodeNs)
		fmt.Printf("  ✓ Decode: %d ns/op\n", result.DecodeNs)
		fmt.Printf("  ✓ Total:  %d ns/op\n", result.TotalNs)
		fmt.Printf("  ✓ Size:   %d bytes\n", result.WireSize)

		allResults = append(allResults, result)
	}

	// Save all results
	return saveResults(allResults, "ffire_swift")
}

// runJavaScript runs the JavaScript (Node.js) benchmarks
func runJavaScript() error {
	fmt.Println("\n🏃 Running ffire JavaScript benchmarks...")

	// Check if node is available
	if _, err := exec.LookPath("node"); err != nil {
		fmt.Println("  ⚠️  node not found (skipping)")
		return nil
	}

	// Find all JavaScript benchmark directories
	pattern := filepath.Join(genDir, "ffire_javascript_*")
	dirs, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(dirs) == 0 {
		fmt.Println("  ⚠️  No JavaScript benchmarks found (skipping)")
		return nil
	}

	var allResults []BenchResult
	for _, dir := range dirs {
		name := strings.TrimPrefix(filepath.Base(dir), "ffire_javascript_")
		fmt.Printf("\n  Testing: %s\n", name)

		result, err := runJavaScriptBench(dir)
		if err != nil {
			fmt.Printf("  ❌ Failed: %v\n", err)
			continue
		}

		// Override message name with schema name for consistent grouping
		result.Message = name

		// Print result
		fmt.Printf("  ✓ Encode: %d ns/op\n", result.EncodeNs)
		fmt.Printf("  ✓ Decode: %d ns/op\n", result.DecodeNs)
		fmt.Printf("  ✓ Total:  %d ns/op\n", result.TotalNs)
		fmt.Printf("  ✓ Size:   %d bytes\n", result.WireSize)

		allResults = append(allResults, result)
	}

	// Save all results
	return saveResults(allResults, "ffire_javascript")
}

// runJava runs the Java benchmarks
func runJava() error {
	fmt.Println("\n🏃 Running ffire Java benchmarks...")

	// Check if java and javac are available
	if _, err := exec.LookPath("java"); err != nil {
		fmt.Println("  ⚠️  java not found (skipping)")
		return nil
	}
	if _, err := exec.LookPath("javac"); err != nil {
		fmt.Println("  ⚠️  javac not found (skipping)")
		return nil
	}

	// Find all Java benchmark directories
	pattern := filepath.Join(genDir, "ffire_java_*")
	dirs, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(dirs) == 0 {
		fmt.Println("  ⚠️  No Java benchmarks found (skipping)")
		return nil
	}

	var allResults []BenchResult
	for _, dir := range dirs {
		name := strings.TrimPrefix(filepath.Base(dir), "ffire_java_")
		fmt.Printf("\n  Testing: %s\n", name)

		result, err := runJavaBench(dir)
		if err != nil {
			fmt.Printf("  ❌ Failed: %v\n", err)
			continue
		}

		// Override message name with schema name for consistent grouping
		result.Message = name

		// Print result
		fmt.Printf("  ✓ Encode: %d ns/op\n", result.EncodeNs)
		fmt.Printf("  ✓ Decode: %d ns/op\n", result.DecodeNs)
		fmt.Printf("  ✓ Total:  %d ns/op\n", result.TotalNs)
		fmt.Printf("  ✓ Size:   %d bytes\n", result.WireSize)

		allResults = append(allResults, result)
	}

	// Save all results
	return saveResults(allResults, "ffire_java")
}

func runCSharp() error {
	fmt.Println("\n🏃 Running ffire C# benchmarks...")

	// Check if dotnet is available
	if _, err := exec.LookPath("dotnet"); err != nil {
		fmt.Println("  ⚠️  dotnet not found (skipping)")
		return nil
	}

	// Find all C# benchmark directories
	pattern := filepath.Join(genDir, "ffire_csharp_*")
	dirs, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(dirs) == 0 {
		fmt.Println("  ⚠️  No C# benchmarks found (skipping)")
		return nil
	}

	var allResults []BenchResult
	for _, dir := range dirs {
		name := strings.TrimPrefix(filepath.Base(dir), "ffire_csharp_")
		fmt.Printf("\n  Testing: %s\n", name)

		result, err := runCSharpBench(dir)
		if err != nil {
			fmt.Printf("  ❌ Failed: %v\n", err)
			continue
		}

		// Override message name with schema name for consistent grouping
		result.Message = name

		// Print result
		fmt.Printf("  ✓ Encode: %d ns/op\n", result.EncodeNs)
		fmt.Printf("  ✓ Decode: %d ns/op\n", result.DecodeNs)
		fmt.Printf("  ✓ Total:  %d ns/op\n", result.TotalNs)
		fmt.Printf("  ✓ Size:   %d bytes\n", result.WireSize)

		allResults = append(allResults, result)
	}

	// Save all results
	return saveResults(allResults, "ffire_csharp")
}

// Compare generates comparison table from all results
// Compare shows performance comparison across languages
// Usage:
//
//	mage compare           - Compare all languages
//	mage compare go java   - Compare specific languages
func Compare() error {
	fmt.Println("\n📊 Generating comparison table (stable languages)...")

	// Load all result files
	files, err := filepath.Glob(filepath.Join(resultsDir, "*.json"))
	if err != nil {
		return err
	}

	var allResults []BenchResult
	for _, file := range files {
		// Skip python and javascript results
		base := filepath.Base(file)
		if strings.Contains(base, "python") || strings.Contains(base, "javascript") {
			continue
		}

		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var results []BenchResult
		if err := json.Unmarshal(data, &results); err != nil {
			continue
		}
		allResults = append(allResults, results...)
	}

	if len(allResults) == 0 {
		return fmt.Errorf("no results found - run 'mage run all' first")
	}

	// Sort alphabetically by message name, then by format
	sort.Slice(allResults, func(i, j int) bool {
		if allResults[i].Message != allResults[j].Message {
			return allResults[i].Message < allResults[j].Message
		}
		return allResults[i].Format < allResults[j].Format
	})

	// Print table
	printComparisonTable(allResults)

	// Save markdown
	if err := saveMarkdownTable(allResults); err != nil {
		return err
	}

	fmt.Printf("\n📝 Results saved to %s/comparison.md\n", resultsDir)
	return nil
}

// Bench is the full workflow: generate, run, compare (stable languages only)
// Excludes Python and JavaScript - use explicit mage run python/javascript
func Bench() error {
	fmt.Println("🚀 Running full benchmark workflow (stable languages)...")
	fmt.Println("    Note: Python and JavaScript excluded from 'all'")

	if err := genAll(); err != nil {
		return err
	}

	if err := runGo(); err != nil {
		return err
	}

	if err := runCpp(); err != nil {
		return err
	}

	if err := runDart(); err != nil {
		return err
	}

	if err := runSwift(); err != nil {
		return err
	}

	if err := runJava(); err != nil {
		return err
	}

	if err := runCSharp(); err != nil {
		return err
	}

	if err := runProto(); err != nil {
		return err
	}

	return Compare()
}

// Graph displays terminal graphs for encode/decode/total times from benchmark results
// Usage:
//
//	mage graph all         - Show average across all schemas (encode, decode, total)
//	mage graph struct      - Show comparison for 'struct' schema only
//	mage graph array_int   - Show comparison for 'array_int' schema only
func Graph(target string) error {
	target = strings.ToLower(target)

	var targetSchema string
	if target != "all" {
		targetSchema = target
	}

	if targetSchema != "" {
		fmt.Printf("\n📈 Generating performance graphs for '%s' (stable languages)...\n", targetSchema)
	} else {
		fmt.Println("\n📈 Generating performance graphs (average across all schemas, stable languages)...")
	}

	// Load all result files
	files, err := filepath.Glob(filepath.Join(resultsDir, "*.json"))
	if err != nil {
		return err
	}

	var allResults []BenchResult
	for _, file := range files {
		// Skip python and javascript results
		base := filepath.Base(file)
		if strings.Contains(base, "python") || strings.Contains(base, "javascript") {
			continue
		}

		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var results []BenchResult
		if err := json.Unmarshal(data, &results); err != nil {
			continue
		}
		allResults = append(allResults, results...)
	}

	if len(allResults) == 0 {
		return fmt.Errorf("no results found - run 'mage bench' or 'mage run all' first")
	}

	// Filter by schema if specified
	if targetSchema != "" {
		var filtered []BenchResult
		for _, r := range allResults {
			if r.Message == targetSchema {
				filtered = append(filtered, r)
			}
		}
		if len(filtered) == 0 {
			return fmt.Errorf("no results found for schema '%s'\nAvailable schemas: run 'mage compare' to see all", targetSchema)
		}
		allResults = filtered
	}

	// Group results by language+format combination
	type LangFormat struct {
		language string
		format   string
	}

	type LangFormatData struct {
		encodeNs []float64
		decodeNs []float64
		totalNs  []float64
	}

	langFormatData := make(map[LangFormat]*LangFormatData)
	messageNames := make(map[string]bool)

	for _, r := range allResults {
		messageNames[r.Message] = true
		key := LangFormat{language: r.Language, format: r.Format}
		if langFormatData[key] == nil {
			langFormatData[key] = &LangFormatData{}
		}
		langFormatData[key].encodeNs = append(langFormatData[key].encodeNs, float64(r.EncodeNs))
		langFormatData[key].decodeNs = append(langFormatData[key].decodeNs, float64(r.DecodeNs))
		langFormatData[key].totalNs = append(langFormatData[key].totalNs, float64(r.TotalNs))
	}

	// Calculate averages for each language+format
	type LangFormatAvg struct {
		language string
		format   string
		label    string
		encodeNs float64
		decodeNs float64
		totalNs  float64
	}

	var langFormatAvgs []LangFormatAvg
	for key, data := range langFormatData {
		encodeSum := 0.0
		for _, v := range data.encodeNs {
			encodeSum += v
		}
		decodeSum := 0.0
		for _, v := range data.decodeNs {
			decodeSum += v
		}
		totalSum := 0.0
		for _, v := range data.totalNs {
			totalSum += v
		}

		count := float64(len(data.encodeNs))
		label := fmt.Sprintf("%s %s", key.format, key.language)
		langFormatAvgs = append(langFormatAvgs, LangFormatAvg{
			language: key.language,
			format:   key.format,
			label:    label,
			encodeNs: encodeSum / count,
			decodeNs: decodeSum / count,
			totalNs:  totalSum / count,
		})
	}

	// Helper function to display a chart
	displayChart := func(title string, getValue func(LangFormatAvg) float64) {
		// Sort by the metric
		sorted := make([]LangFormatAvg, len(langFormatAvgs))
		copy(sorted, langFormatAvgs)
		sort.Slice(sorted, func(i, j int) bool {
			return getValue(sorted[i]) < getValue(sorted[j])
		})

		// Print header
		fmt.Println("\n╔══════════════════════════════════════════════════════════════════╗")
		centerText := func(text string, width int) string {
			padding := width - len(text)
			if padding <= 0 {
				return text
			}
			leftPad := padding / 2
			rightPad := padding - leftPad
			return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
		}
		fmt.Printf("║%s║\n", centerText(title, 66))
		fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
		fmt.Println()

		// Find max for scaling
		maxVal := getValue(sorted[len(sorted)-1])
		barWidth := 50

		// Draw bars
		for i, lf := range sorted {
			val := getValue(lf)
			usVal := val / 1000.0 // Convert to microseconds
			barLen := int((val / maxVal) * float64(barWidth))
			if barLen < 1 {
				barLen = 1
			}

			bar := strings.Repeat("█", barLen)

			// Color coding
			var color string
			if i == 0 {
				color = "\033[32m" // Green
			} else if float64(i) < float64(len(sorted))/2 {
				color = "\033[33m" // Yellow
			} else {
				color = "\033[31m" // Red
			}
			resetColor := "\033[0m"

			fmt.Printf("  %-15s %s%s%s %.2f μs", lf.label, color, bar, resetColor, usVal)
			if i > 0 {
				ratio := val / getValue(sorted[0])
				fmt.Printf("  (%.2fx)", ratio)
			}
			fmt.Println()
		}

		fmt.Println()

		// Show comparison only for same language (e.g., ffire Go vs proto Go)
		// Find ffire and proto implementations of the same language
		for _, ffireLf := range sorted {
			if ffireLf.format == "ffire" {
				// Look for proto version of the same language
				for _, protoLf := range sorted {
					if protoLf.format == "proto" && protoLf.language == ffireLf.language {
						ratio := getValue(protoLf) / getValue(ffireLf)
						ffireVal := getValue(ffireLf) / 1000.0
						protoVal := getValue(protoLf) / 1000.0
						fmt.Printf("  → %s (%.2f μs) is %.2fx faster than %s (%.2f μs)\n",
							ffireLf.label, ffireVal, ratio, protoLf.label, protoVal)
					}
				}
			}
		}
	}

	// Display schema context
	if targetSchema != "" {
		fmt.Printf("\nPerformance Metrics for '%s':\n", targetSchema)
	} else {
		fmt.Printf("\nPerformance Metrics - Average across %d schemas:\n", len(messageNames))
	}

	// Display three charts
	displayChart("Encode Time (μs) - Lower is Better", func(lf LangFormatAvg) float64 { return lf.encodeNs })
	displayChart("Decode Time (μs) - Lower is Better", func(lf LangFormatAvg) float64 { return lf.decodeNs })
	displayChart("Total Time (μs) - Lower is Better", func(lf LangFormatAvg) float64 { return lf.totalNs })

	return nil
}

// Helper functions

func runGoBench(dir string) (BenchResult, error) {
	// Run benchmark with JSON output
	cmd := exec.Command("go", "run", ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "BENCH_JSON=1")

	output, err := cmd.Output()
	if err != nil {
		return BenchResult{}, fmt.Errorf("benchmark failed: %w", err)
	}

	var result BenchResult
	if err := json.Unmarshal(output, &result); err != nil {
		return BenchResult{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return result, nil
}

func runCppBench(dir string) (BenchResult, error) {
	// Build using make (simpler, works everywhere)
	fmt.Printf("    Building C++ benchmark...\n")
	if err := sh.RunV("make", "-C", dir); err != nil {
		return BenchResult{}, fmt.Errorf("build failed: %w", err)
	}

	// Run benchmark with JSON output (use absolute path)
	benchPath := filepath.Join(dir, "bench")
	cmd := exec.Command(benchPath)
	cmd.Env = append(os.Environ(), "BENCH_JSON=1")

	output, err := cmd.Output()
	if err != nil {
		return BenchResult{}, fmt.Errorf("benchmark failed: %w", err)
	}

	var result BenchResult
	if err := json.Unmarshal(output, &result); err != nil {
		return BenchResult{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return result, nil
}

func runPythonBench(dir string) (BenchResult, error) {
	// Python benchmarks are in the python/ subdirectory
	pythonDir := filepath.Join(dir, "python")

	// Install the pybind11 package (editable mode for fast iteration)
	// Use --break-system-packages for Homebrew Python or --user as fallback
	installCmd := exec.Command("pip3", "install", "-e", ".", "--quiet", "--break-system-packages")
	installCmd.Dir = pythonDir
	if err := installCmd.Run(); err != nil {
		// Try again with --user if --break-system-packages failed
		installCmd = exec.Command("pip3", "install", "-e", ".", "--quiet", "--user")
		installCmd.Dir = pythonDir
		if err := installCmd.Run(); err != nil {
			return BenchResult{}, fmt.Errorf("pip install failed: %w", err)
		}
	}

	// Run benchmark with JSON output
	cmd := exec.Command("python3", "bench.py")
	cmd.Dir = pythonDir
	cmd.Env = append(os.Environ(), "BENCH_JSON=1")

	output, err := cmd.Output()
	if err != nil {
		return BenchResult{}, fmt.Errorf("benchmark failed: %w", err)
	}

	var result BenchResult
	if err := json.Unmarshal(output, &result); err != nil {
		return BenchResult{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return result, nil
}

func runPythonPureBench(dir string) (BenchResult, error) {
	// Pure Python benchmarks are directly in the directory (no python/ subdirectory)

	// Run benchmark directly with JSON output
	cmd := exec.Command("python3", "bench.py")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "BENCH_JSON=1")

	output, err := cmd.Output()
	if err != nil {
		return BenchResult{}, fmt.Errorf("benchmark failed: %w", err)
	}

	var result BenchResult
	if err := json.Unmarshal(output, &result); err != nil {
		return BenchResult{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return result, nil
}

func runDartBench(dir string) (BenchResult, error) {
	// Dart benchmarks are in the dart/ subdirectory
	dartDir := filepath.Join(dir, "dart")

	// Install dependencies if not already done
	pubspecLock := filepath.Join(dartDir, ".dart_tool")
	if _, err := os.Stat(pubspecLock); os.IsNotExist(err) {
		fmt.Printf("    Installing Dart dependencies...\n")
		cmd := exec.Command("dart", "pub", "get")
		cmd.Dir = dartDir
		if err := cmd.Run(); err != nil {
			return BenchResult{}, fmt.Errorf("dart pub get failed: %w", err)
		}
	}

	// Run benchmark with JSON output
	cmd := exec.Command("dart", "run", "bench.dart")
	cmd.Dir = dartDir
	cmd.Env = append(os.Environ(), "BENCH_JSON=1")

	output, err := cmd.Output()
	if err != nil {
		return BenchResult{}, fmt.Errorf("benchmark failed: %w", err)
	}

	var result BenchResult
	if err := json.Unmarshal(output, &result); err != nil {
		return BenchResult{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return result, nil
}

func runSwiftBench(dir string) (BenchResult, error) {
	// Swift benchmarks are in the swift/ subdirectory
	swiftDir := filepath.Join(dir, "swift")

	// Workaround: ffire bench generates libtest.dylib but bench.swift expects lib{schema}.dylib
	// Create symlink if needed
	libDir := filepath.Join(swiftDir, "lib")
	testLib := filepath.Join(libDir, "libtest.dylib")
	if _, err := os.Stat(testLib); err == nil {
		// Extract schema name from directory
		schemaName := strings.TrimPrefix(filepath.Base(dir), "ffire_swift_")
		expectedLib := filepath.Join(libDir, fmt.Sprintf("lib%s.dylib", schemaName))
		if _, err := os.Stat(expectedLib); os.IsNotExist(err) {
			// Create symlink
			os.Symlink("libtest.dylib", expectedLib)
		}
	}

	// Build and run benchmark using Swift Package Manager
	// The benchmark is an executable target that imports the generated package
	absLibDir, err := filepath.Abs(filepath.Join(swiftDir, "lib"))
	if err != nil {
		return BenchResult{}, fmt.Errorf("failed to get absolute lib path: %w", err)
	}

	fmt.Printf("    Building and running Swift benchmark...\n")
	cmd := exec.Command("swift", "run", "-c", "release", "bench")
	cmd.Dir = swiftDir
	cmd.Env = append(os.Environ(),
		"BENCH_JSON=1",
		"DYLD_LIBRARY_PATH="+absLibDir,
	)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return BenchResult{}, fmt.Errorf("benchmark failed: %w\nStderr: %s", err, exitErr.Stderr)
		}
		return BenchResult{}, fmt.Errorf("benchmark failed: %w", err)
	}

	var result BenchResult
	if err := json.Unmarshal(output, &result); err != nil {
		return BenchResult{}, fmt.Errorf("failed to parse JSON: %w\nOutput: %s", err, output)
	}

	return result, nil
}

func runJavaScriptBench(dir string) (BenchResult, error) {
	// JavaScript benchmarks are in the javascript/ subdirectory
	jsDir := filepath.Join(dir, "javascript")

	// Check if node_modules exists, if not run npm install
	nodeModules := filepath.Join(jsDir, "node_modules")
	if _, err := os.Stat(nodeModules); os.IsNotExist(err) {
		fmt.Printf("    Installing dependencies...\n")
		cmd := exec.Command("npm", "install")
		cmd.Dir = jsDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return BenchResult{}, fmt.Errorf("npm install failed: %w\nOutput: %s", err, output)
		}
	}

	// Run benchmark
	fmt.Printf("    Running JavaScript benchmark...\n")
	cmd := exec.Command("node", "bench.js")
	cmd.Dir = jsDir
	cmd.Env = append(os.Environ(), "BENCH_JSON=1")

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return BenchResult{}, fmt.Errorf("benchmark failed: %w\nStderr: %s", err, exitErr.Stderr)
		}
		return BenchResult{}, fmt.Errorf("benchmark failed: %w", err)
	}

	var result BenchResult
	if err := json.Unmarshal(output, &result); err != nil {
		return BenchResult{}, fmt.Errorf("failed to parse JSON: %w\nOutput: %s", err, output)
	}

	return result, nil
}

func runJavaBench(dir string) (BenchResult, error) {
	// Java benchmarks are in the java/ subdirectory
	javaDir := filepath.Join(dir, "java")

	// Find all Java files
	javaFiles, err := filepath.Glob(filepath.Join(javaDir, "*.java"))
	if err != nil {
		return BenchResult{}, fmt.Errorf("failed to find Java files: %w", err)
	}
	if len(javaFiles) == 0 {
		return BenchResult{}, fmt.Errorf("no Java files found in %s", javaDir)
	}

	// Compile all Java files
	fmt.Printf("    Compiling Java benchmark...\n")
	args := []string{"-d", "."}
	for _, f := range javaFiles {
		args = append(args, filepath.Base(f))
	}
	cmd := exec.Command("javac", args...)
	cmd.Dir = javaDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return BenchResult{}, fmt.Errorf("javac failed: %w\nOutput: %s", err, output)
	}

	// Run benchmark with JSON output
	fmt.Printf("    Running Java benchmark...\n")
	cmd = exec.Command("java", "Bench")
	cmd.Dir = javaDir
	cmd.Env = append(os.Environ(), "BENCH_JSON=1")

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return BenchResult{}, fmt.Errorf("benchmark failed: %w\nStderr: %s", err, exitErr.Stderr)
		}
		return BenchResult{}, fmt.Errorf("benchmark failed: %w", err)
	}

	var result BenchResult
	if err := json.Unmarshal(output, &result); err != nil {
		return BenchResult{}, fmt.Errorf("failed to parse JSON: %w\nOutput: %s", err, output)
	}

	return result, nil
}

func runCSharpBench(dir string) (BenchResult, error) {
	// C# benchmarks are in the csharp/ subdirectory
	csharpDir := filepath.Join(dir, "csharp")

	// Find the .csproj file
	csprojFiles, err := filepath.Glob(filepath.Join(csharpDir, "*.csproj"))
	if err != nil {
		return BenchResult{}, fmt.Errorf("failed to find .csproj files: %w", err)
	}
	if len(csprojFiles) == 0 {
		return BenchResult{}, fmt.Errorf("no .csproj file found in %s", csharpDir)
	}

	// Compile C# benchmark
	fmt.Printf("    Compiling C# benchmark...\n")
	cmd := exec.Command("dotnet", "build", "-c", "Release", "--nologo", "-v", "q")
	cmd.Dir = csharpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return BenchResult{}, fmt.Errorf("dotnet build failed: %w\nOutput: %s", err, output)
	}

	// Run benchmark with JSON output
	fmt.Printf("    Running C# benchmark...\n")
	cmd = exec.Command("dotnet", "run", "-c", "Release", "--no-build", "--nologo")
	cmd.Dir = csharpDir
	cmd.Env = append(os.Environ(), "BENCH_JSON=1")

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return BenchResult{}, fmt.Errorf("benchmark failed: %w\nStderr: %s", err, exitErr.Stderr)
		}
		return BenchResult{}, fmt.Errorf("benchmark failed: %w", err)
	}

	var result BenchResult
	if err := json.Unmarshal(output, &result); err != nil {
		return BenchResult{}, fmt.Errorf("failed to parse JSON: %w\nOutput: %s", err, output)
	}

	return result, nil
}

func saveResults(results []BenchResult, name string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(resultsDir, fmt.Sprintf("%s.json", name))
	return os.WriteFile(filename, data, 0644)
}

func printComparisonTable(results []BenchResult) {
	fmt.Println("\n" + strings.Repeat("=", 95))
	fmt.Println("BENCHMARK COMPARISON")
	fmt.Println(strings.Repeat("=", 95))
	fmt.Printf("%-12s %-10s %-15s %12s %12s %12s %10s\n",
		"Language", "Format", "Message", "Encode", "Decode", "Total", "Size")
	fmt.Println(strings.Repeat("-", 95))

	lastMessage := ""
	for _, r := range results {
		// Add divider between different schemas
		if lastMessage != "" && r.Message != lastMessage {
			fmt.Println(strings.Repeat("-", 95))
		}
		lastMessage = r.Message

		fmt.Printf("%-12s %-10s %-15s %10d ns %10d ns %10d ns %8d B\n",
			r.Language, r.Format, r.Message,
			r.EncodeNs, r.DecodeNs, r.TotalNs,
			r.WireSize)
	}
	fmt.Println(strings.Repeat("=", 95))
}

func saveMarkdownTable(results []BenchResult) error {
	var buf strings.Builder

	buf.WriteString("# ffire Benchmark Comparison\n\n")
	buf.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))

	buf.WriteString("| Language | Format | Message | Encode (ns) | Decode (ns) | Total (ns) | Wire Size |\n")
	buf.WriteString("|----------|--------|---------|-------------|-------------|------------|----------|\n")

	lastMessage := ""
	for _, r := range results {
		// Add divider between different schemas
		if lastMessage != "" && r.Message != lastMessage {
			buf.WriteString("|----------|--------|---------|-------------|-------------|------------|----------|\n")
		}
		lastMessage = r.Message

		buf.WriteString(fmt.Sprintf("| %s | %s | %s | %d | %d | %d | %d |\n",
			r.Language, r.Format, r.Message,
			r.EncodeNs, r.DecodeNs, r.TotalNs,
			r.WireSize))
	}

	buf.WriteString("\n## Notes\n\n")
	buf.WriteString("- All benchmarks use the same test fixture\n")
	buf.WriteString("- Measurements exclude warmup and fixture loading\n")
	buf.WriteString("- Results are averaged over multiple iterations\n")

	return os.WriteFile(
		filepath.Join(resultsDir, "comparison.md"),
		[]byte(buf.String()),
		0644,
	)
}

func inferMessageType(name string, jsonData []byte) string {
	// Map schema names to proto message type names (as defined in .proto files)
	// Note: proto uses the exact message names from .proto files, no suffix added
	typeMap := map[string]string{
		"complex":      "PluginList",
		"array_float":  "FloatList",
		"array_int":    "IntList",
		"array_string": "StringList",
		"array_struct": "DeviceList",
		"empty":        "EmptyTest",
		"nested":       "Level1",
		"optional":     "RecordList",
		"struct":       "Config",
		"tags":         "User",
	}

	if typeName, ok := typeMap[name]; ok {
		return typeName
	}

	// Fallback: capitalize first letter
	if len(name) > 0 {
		return strings.ToUpper(name[:1]) + name[1:]
	}
	return "Message"
}

func getFieldName(name string) string {
	// Map schema names to their proto field names for wrapping raw arrays
	// Empty string means the JSON is already an object and doesn't need wrapping
	fieldMap := map[string]string{
		"complex":      "plugins",
		"array_float":  "values",
		"array_int":    "values",
		"array_string": "values",
		"array_struct": "devices",
		"empty":        "",
		"nested":       "", // already an object
		"optional":     "records",
		"struct":       "",
		"tags":         "",
	}

	if fieldName, ok := fieldMap[name]; ok {
		return fieldName
	}
	return "items"
}

func generateProtoBenchmark(name, outDir, jsonFile string) error {
	// Read JSON fixture to determine message type
	jsonData, err := os.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("failed to read JSON: %w", err)
	}

	// Infer message type from JSON structure
	msgTypeName := inferMessageType(name, jsonData)
	fieldName := getFieldName(name)

	// Generate JSON wrapping code if needed (for raw arrays)
	wrapCode := ""
	if fieldName != "" {
		wrapCode = `
	// Wrap raw array JSON in object with field name for protojson
	wrappedJSON := []byte(fmt.Sprintf("{%q: %s}", "` + fieldName + `", string(fixtureJSON)))
	jsonToUnmarshal := wrappedJSON`
	} else {
		wrapCode = `
	jsonToUnmarshal := fixtureJSON`
	}

	// Create benchmark driver using protojson for generic unmarshalling
	benchCode := `package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"time"

	testproto "protobench/github.com/shaban/ffire/testdata/proto"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

//go:embed fixture.json
var fixtureJSON []byte

type BenchResult struct {
	Language    string ` + "`json:\"language\"`" + `
	Format      string ` + "`json:\"format\"`" + `
	Message     string ` + "`json:\"message\"`" + `
	Iterations  int    ` + "`json:\"iterations\"`" + `
	EncodeNs    int64  ` + "`json:\"encode_ns\"`" + `
	DecodeNs    int64  ` + "`json:\"decode_ns\"`" + `
	TotalNs     int64  ` + "`json:\"total_ns\"`" + `
	WireSize    int    ` + "`json:\"wire_size\"`" + `
	FixtureSize int    ` + "`json:\"fixture_size\"`" + `
	Timestamp   string ` + "`json:\"timestamp\"`" + `
}

func main() {
	iterations := 10000
	jsonOutput := os.Getenv("BENCH_JSON") == "1"
` + wrapCode + `

	// Parse JSON into protobuf message using protojson
	msg := &testproto.` + msgTypeName + `{}
	if err := protojson.Unmarshal(jsonToUnmarshal, msg); err != nil {
		panic(fmt.Sprintf("failed to parse fixture: %v", err))
	}

	// Warmup
	for i := 0; i < 1000; i++ {
		encoded, _ := proto.Marshal(msg)
		decoded := &testproto.` + msgTypeName + `{}
		proto.Unmarshal(encoded, decoded)
	}

	// Benchmark encode
	start := time.Now()
	var encoded []byte
	for i := 0; i < iterations; i++ {
		encoded, _ = proto.Marshal(msg)
	}
	encodeTime := time.Since(start)

	// Benchmark decode
	start = time.Now()
	for i := 0; i < iterations; i++ {
		decoded := &testproto.` + msgTypeName + `{}
		proto.Unmarshal(encoded, decoded)
	}
	decodeTime := time.Since(start)

	// Calculate metrics
	encodeNs := encodeTime.Nanoseconds() / int64(iterations)
	decodeNs := decodeTime.Nanoseconds() / int64(iterations)
	totalNs := encodeNs + decodeNs

	if jsonOutput {
		result := BenchResult{
			Language:    "Go",
			Format:      "proto",
			Message:     "` + name + `",
			Iterations:  iterations,
			EncodeNs:    encodeNs,
			DecodeNs:    decodeNs,
			TotalNs:     totalNs,
			WireSize:    len(encoded),
			FixtureSize: len(fixtureJSON),
			Timestamp:   time.Now().Format(time.RFC3339),
		}
		json.NewEncoder(os.Stdout).Encode(result)
	} else {
		fmt.Printf("proto benchmark: ` + name + `\n")
		fmt.Printf("Iterations:  %d\n", iterations)
		fmt.Printf("Encode:      %d ns/op\n", encodeNs)
		fmt.Printf("Decode:      %d ns/op\n", decodeNs)
		fmt.Printf("Total:       %d ns/op\n", totalNs)
		fmt.Printf("Wire size:   %d bytes\n", len(encoded))
		fmt.Printf("Fixture:     %d bytes\n", len(fixtureJSON))
		fmt.Printf("Total time:  %.2fs\n", (encodeTime + decodeTime).Seconds())
	}
}
`

	benchFile := filepath.Join(outDir, "bench.go")
	if err := os.WriteFile(benchFile, []byte(benchCode), 0644); err != nil {
		return fmt.Errorf("failed to write bench.go: %w", err)
	}

	// Copy JSON fixture
	fixtureFile := filepath.Join(outDir, "fixture.json")
	if err := os.WriteFile(fixtureFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write fixture: %w", err)
	}

	// Create go.mod
	goMod := `module protobench

go 1.21

require google.golang.org/protobuf v1.31.0
`
	modFile := filepath.Join(outDir, "go.mod")
	if err := os.WriteFile(modFile, []byte(goMod), 0644); err != nil {
		return fmt.Errorf("failed to write go.mod: %w", err)
	}

	// Run go mod tidy to generate go.sum
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = outDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run go mod tidy: %w", err)
	}

	return nil
}

// Helper function for language generation
func genLanguage(lang string, suites []BenchmarkSuite) error {
	switch lang {
	case "go":
		for _, suite := range suites {
			fmt.Printf("🔧 Generating Go benchmark: %s\n", suite.Name)
			if err := sh.Run("ffire", "bench",
				"--lang", "go",
				"--schema", suite.SchemaFile,
				"--json", suite.JSONFile,
				"--output", filepath.Join(genDir, "ffire_"+suite.Name),
				"--iterations", "100000",
			); err != nil {
				return err
			}
		}
	case "cpp":
		for _, suite := range suites {
			fmt.Printf("🔨 Generating C++ benchmark: %s\n", suite.Name)
			if err := sh.Run("ffire", "bench",
				"--lang", "cpp",
				"--schema", suite.SchemaFile,
				"--json", suite.JSONFile,
				"--output", filepath.Join(genDir, "ffire_cpp_"+suite.Name),
				"--iterations", "100000",
			); err != nil {
				return err
			}
		}
	case "java":
		for _, suite := range suites {
			fmt.Printf("☕ Generating Java benchmark: %s\n", suite.Name)
			if err := sh.Run("ffire", "bench",
				"--lang", "java",
				"--schema", suite.SchemaFile,
				"--json", suite.JSONFile,
				"--output", filepath.Join(genDir, "ffire_java_"+suite.Name),
				"--iterations", "100000",
			); err != nil {
				return err
			}
		}
	case "csharp":
		for _, suite := range suites {
			fmt.Printf("💜 Generating C# benchmark: %s\n", suite.Name)
			if err := sh.Run("ffire", "bench",
				"--lang", "csharp",
				"--schema", suite.SchemaFile,
				"--json", suite.JSONFile,
				"--output", filepath.Join(genDir, "ffire_csharp_"+suite.Name),
				"--iterations", "100000",
			); err != nil {
				return err
			}
		}
	case "swift":
		for _, suite := range suites {
			fmt.Printf("🍎 Generating Swift benchmark: %s\n", suite.Name)
			if err := sh.Run("ffire", "bench",
				"--lang", "swift",
				"--schema", suite.SchemaFile,
				"--json", suite.JSONFile,
				"--output", filepath.Join(genDir, "ffire_swift_"+suite.Name),
				"--iterations", "100000",
			); err != nil {
				return err
			}
		}
	case "dart":
		for _, suite := range suites {
			fmt.Printf("🎯 Generating Dart benchmark: %s\n", suite.Name)
			if err := sh.Run("ffire", "bench",
				"--lang", "dart",
				"--schema", suite.SchemaFile,
				"--json", suite.JSONFile,
				"--output", filepath.Join(genDir, "ffire_dart_"+suite.Name),
				"--iterations", "100000",
			); err != nil {
				return err
			}
		}
	case "python":
		for _, suite := range suites {
			fmt.Printf("🐍 Generating Python benchmark: %s\n", suite.Name)
			if err := sh.Run("ffire", "bench",
				"--lang", "python",
				"--schema", suite.SchemaFile,
				"--json", suite.JSONFile,
				"--output", filepath.Join(genDir, "ffire_python_"+suite.Name),
				"--iterations", "100000",
			); err != nil {
				return err
			}
		}
	case "python-pybind11":
		for _, suite := range suites {
			fmt.Printf("🐍 Generating Python-PyBind11 benchmark: %s\n", suite.Name)
			if err := sh.Run("ffire", "bench",
				"--lang", "python-pybind11",
				"--schema", suite.SchemaFile,
				"--json", suite.JSONFile,
				"--output", filepath.Join(genDir, "ffire_python_pybind11_"+suite.Name),
				"--iterations", "100000",
			); err != nil {
				return err
			}
		}
	case "javascript":
		for _, suite := range suites {
			fmt.Printf("🟨 Generating JavaScript benchmark: %s\n", suite.Name)
			if err := sh.Run("ffire", "bench",
				"--lang", "javascript",
				"--schema", suite.SchemaFile,
				"--json", suite.JSONFile,
				"--output", filepath.Join(genDir, "ffire_js_"+suite.Name),
				"--iterations", "100000",
			); err != nil {
				return err
			}
		}
	case "proto":
		for _, suite := range suites {
			fmt.Printf("📦 Generating proto benchmark: %s\n", suite.Name)
			if _, err := os.Stat(suite.ProtoFile); os.IsNotExist(err) {
				fmt.Printf("  ⚠️  Skipping %s: no proto file\n", suite.Name)
				continue
			}
			outDir := filepath.Join(genDir, "proto_"+suite.Name)
			if err := genProto(suite.Name, suite.ProtoFile, suite.JSONFile); err != nil {
				return err
			}
			fmt.Printf("  Generating benchmark driver...\n")
			if err := generateProtoBenchmark(suite.Name, outDir, suite.JSONFile); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unknown language: %s", lang)
	}
	return nil
}

func runLanguage(lang string) error {
	switch lang {
	case "go":
		return runGo()
	case "cpp":
		return runCpp()
	case "java":
		return runJava()
	case "csharp":
		return runCSharp()
	case "swift":
		return runSwift()
	case "dart":
		return runDart()
	case "python":
		return runPython()
	case "javascript":
		return runJavaScript()
	case "proto":
		return runProto()
	default:
		return fmt.Errorf("unknown language: %s", lang)
	}
}
