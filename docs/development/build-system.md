# Build System

ffire uses [Mage](https://magefile.org/) for benchmark orchestration.

## Why Mage?

Cross-platform task runner in Go. Alternative to Makefiles with:
- Cross-language compilation (Go, C++, Swift, etc.)
- Parallel execution
- Dependency management between tasks
- Type-safe task definitions

## Location

`benchmarks/magefile.go` - 1000+ lines orchestrating the full benchmark workflow.

## Key Functions

### Discovery
```go
discoverBenchmarks() → finds testdata/schema/*.ffi files
```
Auto-discovers benchmark suites by matching `.ffi` + `.json` files.

### Generation Tasks

**`GenAll()`** - Generate all language benchmarks
- Calls `ffire bench` for each (suite × language) combination
- Creates `generated/ffire_LANG_SUITE/` directories
- Converts JSON fixtures to binary
- Go, C++, C#, Java, Swift, Dart, Rust, Zig

**Language-specific generators:**
- `genGoSuite()` → `ffire bench --lang go`
- `genCppSuite()` → `ffire bench --lang cpp`
- `genSwiftSuite()` → `ffire bench --lang swift`
- etc.

### Execution Tasks

**`RunGo()`** - Run Go benchmarks
1. Find all `generated/ffire_*/bench.go`
2. For each: `go run bench.go`
3. Parse JSON output
4. Save to `results/ffire_go.json`

**`RunCpp()`** - Run C++ benchmarks
1. Find all `generated/ffire_cpp_*/`
2. Build with CMake if needed
3. Run compiled binary
4. Collect results

**`RunSwift()`** - Run Swift benchmarks  
1. Find all `generated/ffire_swift_*/`
2. Build: `swift run -c release bench`
3. Set `DYLD_LIBRARY_PATH` for dylib
4. Collect results

**Pattern:** Each `Run*()` follows:
```go
func RunLANG() error {
    dirs := findBenchmarkDirs(pattern)
    results := []BenchResult{}
    
    for dir := range dirs {
        result := runLANGBench(dir)
        results = append(results, result)
    }
    
    saveResults(results, "ffire_LANG")
}
```

### Analysis Tasks

**`Compare()`** - Generate comparison table
- Loads all `results/*.json` files
- Groups by message type
- Prints markdown table:
  ```
  | Message | Lang | Encode | Decode | Size |
  |---------|------|--------|--------|------|
  ```

**`Bench()`** - Full workflow
```go
func Bench() error {
    mg.Deps(Clean)
    mg.Deps(GenAll)
    mg.SerialDeps(RunGo, RunCpp, RunSwift)
    return Compare()
}
```

## Result Format

`results/ffire_LANG.json`:
```json
[
  {
    "language": "Go",
    "format": "ffire",
    "message": "array_int",
    "iterations": 10000,
    "encode_ns": 1574,
    "decode_ns": 10379,
    "total_ns": 11953,
    "wire_size": 20002,
    "fixture_size": 20002,
    "timestamp": "2025-11-08T13:45:00Z"
  }
]
```

## Common Workflows

**Initial setup:**
```bash
cd benchmarks
mage genAll        # Generate everything once
```

**Test single language:**
```bash
mage clean         # Remove old generated code
mage genAll        # Regenerate with latest ffire
mage runGo         # Test Go only
```

**Full comparison:**
```bash
mage bench         # Clean + generate + run all + compare
```

**Iterate on generator:**
```bash
# Fix generator code
go install ./cmd/ffire

# Regenerate single language
rm -rf generated/ffire_swift_*
cd benchmarks && mage genAll
mage runSwift
```

## Parallelization

Mage can run tasks in parallel:
```bash
mage -j 4 runGo runCpp runSwift runDart  # Run 4 languages concurrently
```

## Environment Variables

**`BENCH_JSON=1`** - Output JSON for parsing (vs human-readable)
- Set by mage runners automatically
- Benchmark harnesses check this variable

**`DYLD_LIBRARY_PATH`** (macOS) - Find shared libraries
- Swift/Dart benchmarks need C ABI dylib
- Set to `generated/*/lib` directory

## Dependencies

Required for full benchmark suite:
- `go` - Go benchmarks
- `g++` or `clang++` - C++ benchmarks
- `dotnet` - C# benchmarks
- `javac` - Java benchmarks
- `swift` - Swift benchmarks (macOS)
- `dart` - Dart benchmarks
- `rustc`/`cargo` - Rust benchmarks
- `zig` - Zig benchmarks
- `protoc` - Protobuf comparison

Missing dependencies are skipped with warnings.
