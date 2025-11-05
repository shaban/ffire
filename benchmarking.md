# ffire Benchmarking
## ffire - FFI Encoding

## Problem Statement

Benchmarking cross-language encode/decode performance is challenging:
- CGO can't be used in Go `*_test.go` files easily
- Need to compare C++/Swift implementations against Go
- Want simple, portable benchmark executables
- Need consistent test data across languages

## Solution: Standalone Benchmark Executable

Generate a complete, self-contained benchmark program for each language that:
1. Includes generated encoder/decoder code
2. Embeds binary fixture data (from JSON)
3. Runs N iterations and measures timing
4. Prints standardized results
5. No dependency on testing frameworks

## CLI Usage

```bash
ffire bench [options]
  --schema string      Schema file (required)
  --json string        JSON test data (required)
  --output string      Output directory (required)
  --message string     Message type name (default: "Message")
  --iterations int     Benchmark iterations (default: 100000)
```

Note: C++ and other language support is planned but not yet implemented.

## Example Workflow

```bash
# Generate benchmark for Go
ffire bench --schema schema.ffi --json test_data.json --output ./bench_go

# Run benchmark (human-readable output)
cd bench_go && go run .

# Run benchmark with JSON output (for automation)
cd bench_go && BENCH_JSON=1 go run .

# Use Magefile for cross-language comparison
cd examples/benchmark-comparison
mage bench
```

## Generated Structure

### Go Output (Current)
```
bench_go/
  bench.go         # Benchmark main with //go:embed
  generated.go     # Generated encoder/decoder (package main)
  fixture.bin      # Binary fixture
  go.mod           # Minimal module definition
```

Simply run: `cd bench_go && go run .`

### C++ Output (Planned)
```
bench_cpp/
  bench.cpp        # Benchmark main
  generated.h      # Generated header
  generated.cpp    # Generated implementation
  fixture.bin      # Binary fixture
```

## Benchmark Output Format

### Human-Readable (default)
```
ffire benchmark: Config
Iterations:  100000
Encode:      203 ns/op
Decode:      210 ns/op
Total:       413 ns/op
Wire size:   24 bytes
Fixture:     24 bytes
Total time:  0.04s
```

### JSON (with BENCH_JSON=1)
```json
{
  "language": "Go",
  "format": "ffire",
  "message": "Config",
  "iterations": 100000,
  "encode_ns": 203,
  "decode_ns": 210,
  "total_ns": 413,
  "wire_size": 24,
  "fixture_size": 24,
  "timestamp": "2025-11-05T15:00:00Z"
}
```

## Go Benchmark Template

```go
package main

import (
    _ "embed"
    "fmt"
    "time"
)

//go:embed test_data.bin
var fixtureData []byte

func main() {
    iterations := 1_000_000
    
    // Decode fixture once
    original, err := DecodeTypeName(fixtureData)
    if err != nil {
        panic(err)
    }
    
    // Benchmark encode
    start := time.Now()
    var encoded []byte
    for i := 0; i < iterations; i++ {
        encoded = EncodeTypeName(original)
    }
    encodeTime := time.Since(start)
    
    // Benchmark decode
    start = time.Now()
    for i := 0; i < iterations; i++ {
        _, _ = DecodeTypeName(encoded)
    }
    decodeTime := time.Since(start)
    
    // Print results
    fmt.Printf("ffire benchmark: TypeName\n")
    fmt.Printf("Iterations:  %d\n", iterations)
    fmt.Printf("Encode:      %d ns/op\n", encodeTime.Nanoseconds()/int64(iterations))
    fmt.Printf("Decode:      %d ns/op\n", decodeTime.Nanoseconds()/int64(iterations))
    fmt.Printf("Total:       %d ns/op\n", (encodeTime+decodeTime).Nanoseconds()/int64(iterations))
    fmt.Printf("Wire size:   %d bytes\n", len(encoded))
    fmt.Printf("Total time:  %.2fs\n", (encodeTime+decodeTime).Seconds())
}
```

## C++ Benchmark Template

```cpp
#include "package_ffire.h"
#include <chrono>
#include <fstream>
#include <iostream>
#include <vector>

std::vector<uint8_t> load_fixture(const char* path) {
    std::ifstream file(path, std::ios::binary);
    return std::vector<uint8_t>((std::istreambuf_iterator<char>(file)),
                                 std::istreambuf_iterator<char>());
}

int main() {
    const int iterations = 1'000'000;
    
    // Load and decode fixture
    auto fixture_data = load_fixture("test_data.bin");
    auto original = namespace::decode_type_name(fixture_data);
    
    // Benchmark encode
    std::vector<uint8_t> encoded;
    auto start = std::chrono::high_resolution_clock::now();
    for (int i = 0; i < iterations; ++i) {
        encoded = namespace::encode_type_name(original);
    }
    auto encode_time = std::chrono::high_resolution_clock::now() - start;
    
    // Benchmark decode
    start = std::chrono::high_resolution_clock::now();
    for (int i = 0; i < iterations; ++i) {
        auto decoded = namespace::decode_type_name(encoded);
    }
    auto decode_time = std::chrono::high_resolution_clock::now() - start;
    
    // Calculate metrics
    auto encode_ns = std::chrono::duration_cast<std::chrono::nanoseconds>(encode_time).count();
    auto decode_ns = std::chrono::duration_cast<std::chrono::nanoseconds>(decode_time).count();
    auto total_ns = encode_ns + decode_ns;
    
    // Print results
    std::cout << "ffire benchmark: namespace::TypeName\n";
    std::cout << "Iterations:  " << iterations << "\n";
    std::cout << "Encode:      " << encode_ns / iterations << " ns/op\n";
    std::cout << "Decode:      " << decode_ns / iterations << " ns/op\n";
    std::cout << "Total:       " << total_ns / iterations << " ns/op\n";
    std::cout << "Wire size:   " << encoded.size() << " bytes\n";
    std::cout << "Total time:  " << total_ns / 1e9 << "s\n";
    
    return 0;
}
```

## Benefits

✅ **Minimalistic** - Just 3-4 files per benchmark  
✅ **Self-contained** - No external dependencies  
✅ **Portable** - Run with `go run .`  
✅ **Debuggable** - Standard Go code, use any debugger  
✅ **Comparable** - Same fixture across implementations  
✅ **Automation-friendly** - JSON output for tooling  
✅ **Fast iteration** - No test framework overhead  
✅ **Real-world** - Measures actual encode/decode performance

## Implementation Notes

### Fixture Generation
- JSON data converted to binary via `pkg/fixture`
- Binary fixture embedded with `//go:embed`
- Validation happens before encoding

### Iteration Count
- Default 100K iterations (configurable via `--iterations`)
- 1000 warmup iterations (not measured)
- Suitable for microsecond-level operations

### JSON Output
- Set `BENCH_JSON=1` environment variable
- Used by Magefile orchestration for comparison
- Standard format across all implementations

### Architecture
- Generator creates `package main` code
- No subdirectories or complex imports
- Single `go run .` command to execute

## Cross-Language Comparison

See `examples/benchmark-comparison/` for Magefile-based orchestration:

```bash
cd examples/benchmark-comparison
mage bench
```

This generates benchmarks, runs them, and produces comparison tables automatically.

## Future Enhancements

- **C++ generation**: Complete C++ benchmark generator
- **Protobuf comparison**: Add `ffire to-proto` for protobuf baseline
- **Profile mode**: Generate pprof/perf-compatible output
- **Memory tracking**: Add detailed allocation metrics
- **Rust support**: Generate Rust benchmarks
