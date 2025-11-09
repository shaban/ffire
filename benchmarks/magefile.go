//go:build mage || tools
// +build mage tools

// Magefile for cross-language benchmark comparison
//
// Usage:
//
//	mage genAll      - Generate all benchmarks
//	mage runGo       - Run Go benchmarks
//	mage compare     - Show comparison table
//	mage bench       - Full workflow (generate + run + compare)
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

	"github.com/magefile/mage/mg"
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

// Clean removes all generated benchmark code (but preserves results/)
func Clean() error {
	fmt.Println("🧹 Cleaning generated files...")
	os.RemoveAll(genDir)
	// Don't delete resultsDir - preserve historical benchmark results
	return nil
}

// CleanAll removes all generated files AND results
func CleanAll() error {
	fmt.Println("🧹 Cleaning all generated files and results...")
	os.RemoveAll(genDir)
	os.RemoveAll(resultsDir)
	return nil
}

// GenAll generates all benchmark variants
func GenAll() error {
	mg.Deps(Clean)

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
	if err := sh.RunV("sh", "-c", "cd .. && go install ./cmd/ffire"); err != nil {
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
			"--iterations", "10000",
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
			"--iterations", "10000",
		); err != nil {
			fmt.Printf("  ⚠️  Skipping %s: %v\n", suite.Name, err)
			continue
		}
	}

	// Generate ffire Python benchmarks
	for _, suite := range suites {
		fmt.Printf("🐍 Generating ffire Python benchmark: %s\n", suite.Name)
		if err := sh.Run("ffire", "bench",
			"--lang", "python",
			"--schema", suite.SchemaFile,
			"--json", suite.JSONFile,
			"--output", filepath.Join(genDir, "ffire_python_"+suite.Name),
			"--iterations", "10000",
		); err != nil {
			fmt.Printf("  ⚠️  Skipping %s: %v\n", suite.Name, err)
			continue
		}
	}

	// Generate ffire Dart benchmarks
	for _, suite := range suites {
		fmt.Printf("🎯 Generating ffire Dart benchmark: %s\n", suite.Name)
		if err := sh.Run("ffire", "bench",
			"--lang", "dart",
			"--schema", suite.SchemaFile,
			"--json", suite.JSONFile,
			"--output", filepath.Join(genDir, "ffire_dart_"+suite.Name),
			"--iterations", "10000",
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
			"--iterations", "10000",
		); err != nil {
			fmt.Printf("  ⚠️  Skipping %s: %v\n", suite.Name, err)
			continue
		}
	}

	// Generate ffire JavaScript benchmarks
	for _, suite := range suites {
		fmt.Printf("🟨 Generating ffire JavaScript benchmark: %s\n", suite.Name)
		if err := sh.Run("ffire", "bench",
			"--lang", "javascript",
			"--schema", suite.SchemaFile,
			"--json", suite.JSONFile,
			"--output", filepath.Join(genDir, "ffire_javascript_"+suite.Name),
			"--iterations", "10000",
		); err != nil {
			fmt.Printf("  ⚠️  Skipping %s: %v\n", suite.Name, err)
			continue
		}
	}

	// Generate protobuf benchmarks (only for those with .proto files)
	for _, suite := range suites {
		if _, err := os.Stat(suite.ProtoFile); err == nil {
			fmt.Printf("📦 Generating protobuf benchmark: %s\n", suite.Name)
			if err := genProto(suite.Name, suite.ProtoFile, suite.JSONFile); err != nil {
				return fmt.Errorf("failed to generate proto benchmark for %s: %w", suite.Name, err)
			}
		}
	}

	fmt.Println("\n✅ All benchmarks generated")
	return nil
}

// genProto generates protobuf benchmark
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

// RunGo runs the Go benchmarks
func RunGo() error {
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
			!strings.HasPrefix(base, "ffire_javascript_") {
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

// RunProto runs the protobuf benchmarks
func RunProto() error {
	fmt.Println("\n🏃 Running protobuf Go benchmarks...")

	// Find all protobuf benchmark directories
	pattern := filepath.Join(genDir, "proto_*")
	dirs, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(dirs) == 0 {
		fmt.Println("  ⚠️  No protobuf benchmarks found (skipping)")
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

// RunCpp runs the C++ benchmarks
func RunCpp() error {
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

// RunPython runs the Python benchmarks
func RunPython() error {
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

		result, err := runPythonBench(dir)
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

// RunDart runs the Dart benchmarks
func RunDart() error {
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

// RunSwift runs the Swift benchmarks
func RunSwift() error {
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

// RunJavaScript runs the JavaScript (Node.js) benchmarks
func RunJavaScript() error {
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

// Compare generates comparison table from all results
func Compare() error {
	fmt.Println("\n📊 Generating comparison table...")

	// Load all result files
	files, err := filepath.Glob(filepath.Join(resultsDir, "*.json"))
	if err != nil {
		return err
	}

	var allResults []BenchResult
	for _, file := range files {
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
		return fmt.Errorf("no results found - run 'mage runGo' first")
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

// Bench is the full workflow: generate, run, compare
func Bench() error {
	fmt.Println("🚀 Running full benchmark workflow...")
	mg.Deps(GenAll)

	if err := RunGo(); err != nil {
		return err
	}

	if err := RunCpp(); err != nil {
		return err
	}

	if err := RunPython(); err != nil {
		return err
	}

	if err := RunDart(); err != nil {
		return err
	}

	if err := RunSwift(); err != nil {
		return err
	}

	if err := RunProto(); err != nil {
		return err
	}

	return Compare()
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
	// Map schema names to generated message type names
	// Note: ffire generators automatically append "Message" suffix to avoid keyword collisions
	typeMap := map[string]string{
		"complex":      "PluginListMessage",
		"array_float":  "FloatListMessage",
		"array_int":    "IntListMessage",
		"array_string": "StringListMessage",
		"array_struct": "DeviceListMessage",
		"empty":        "EmptyTestMessage",
		"nested":       "Level1Message",
		"optional":     "RecordListMessage",
		"struct":       "ConfigMessage",
		"tags":         "UserMessage",
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
			Format:      "protobuf",
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
		fmt.Printf("protobuf benchmark: ` + name + `\n")
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
