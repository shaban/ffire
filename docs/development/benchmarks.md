# Benchmarks

Cross-language benchmark system for measuring encoding/decoding performance.

## Overview

```
testdata/                    benchmarks/
├── schema/array_int.ffi    ├── magefile.go (orchestration)
├── json/array_int.json  ───→ mage gen all ───→ generated/
└── proto/array_int.proto   │                   ├── ffire_go_array_int/
                            │                   ├── ffire_cpp_array_int/
                            │                   ├── ffire_rust_array_int/
                            │                   └── ...
                            │
                            └── mage run all ───→ results/
                                mage compare      ├── ffire_go.json
                                                 ├── ffire_cpp.json
                                                 └── ...
```

## Mage Commands

```bash
cd benchmarks

# Full workflow
mage bench              # Generate + run all + compare

# Generation
mage gen all            # Generate benchmarks for all languages
mage gen rust           # Generate benchmarks for one language

# Execution
mage run all            # Run all language benchmarks
mage run rust           # Run benchmarks for one language

# Analysis
mage compare            # Show comparison table from results/

# Cleanup
mage clean all          # Remove all generated benchmarks
mage clean rust         # Remove generated benchmarks for one language
```

## Supported Languages

All 8 languages have native implementations (no FFI):

| Language | Command | Status |
|----------|---------|--------|
| Go | `mage run go` | ✅ Native |
| C++ | `mage run cpp` | ✅ Native |
| C# | `mage run csharp` | ✅ Native |
| Java | `mage run java` | ✅ Native |
| Swift | `mage run swift` | ✅ Native |
| Dart | `mage run dart` | ✅ Native |
| Rust | `mage run rust` | ✅ Native |
| Zig | `mage run zig` | ✅ Native |

## Benchmark Suites

Current test cases in `testdata/schema/`:

- `array_int` - 5000 element int32 array
- `array_float` - 5000 element float32 array  
- `array_string` - 1000 strings
- `array_struct` - 1000 nested structs
- `struct` - Simple struct with primitives
- `optional` - Optional/nullable fields
- `nested` - Deeply nested types
- `complex` - Mixed types with arrays/structs
- `empty` - Empty message
- `tags` - Struct with Go tags

## Results Format

Each benchmark reports:
- **Encode**: Time per encode operation (ns/op)
- **Decode**: Time per decode operation (ns/op)  
- **Total**: Encode + Decode
- **Wire Size**: Serialized byte count

## Performance Comparison

**Array of 5000 float32 values (encode + decode):**

| Language | Encode | Decode | Total |
|----------|--------|--------|-------|
| **Rust** | 446 ns | 521 ns | 967 ns |
| **C++** | 784 ns | 537 ns | 1,321 ns |
| **Swift** | 769 ns | 570 ns | 1,339 ns |
| **Zig** | 1,076 ns | 556 ns | 1,632 ns |
| **C#** | 832 ns | 869 ns | 1,701 ns |
| **Java** | 1,017 ns | 1,326 ns | 2,343 ns |
| **Go** | 1,854 ns | 1,530 ns | 3,384 ns |
| **Dart** | 2,504 ns | 7,684 ns | 10,188 ns |

**Key Insights:**
- **Rust** is fastest due to unsafe bulk memcpy optimization
- **C++, Swift, Zig** are within 2x of Rust (excellent)
- **C#, Java** use managed memory but still very fast
- **Go** uses safe bounds checking, still sub-4μs
- **All languages** are production-ready with sub-20μs for 20KB payloads

## Generated Structure

After `mage gen all`, benchmarks land in `generated/`:

```
generated/
├── ffire_go_array_int/       # Go native
│   ├── bench.go
│   ├── generated.go
│   ├── fixture.bin
│   └── go.mod
├── ffire_cpp_array_int/      # C++ native
│   └── cpp/
│       ├── bench.cpp
│       ├── generated.h
│       └── fixture.bin
├── ffire_rust_array_int/     # Rust native
│   └── rust/
│       ├── Cargo.toml
│       ├── src/lib.rs
│       └── fixture.bin
└── proto_array_int/          # Protobuf reference
    ├── bench.go
    ├── array_int.pb.go
    └── fixture.bin
```

## Adding Benchmarks

**New test case:**
1. Create: `testdata/schema/my_test.ffi`
2. Create: `testdata/json/my_test.json`
3. (Optional) Create: `testdata/proto/my_test.proto`
4. Run: `mage gen all`
5. Auto-discovered by filename matching

**New language:**
1. Implement: `pkg/benchmark/benchmark_LANG.go`
2. Add to: `cmd/ffire/bench.go` (switch case)
3. Add runner in: `benchmarks/magefile.go`
4. Run: `mage gen LANG && mage run LANG`
