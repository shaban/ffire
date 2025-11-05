# ffire Benchmarks

Cross-language benchmark comparison using Magefile for orchestration.

## Setup

1. **Install mage** (one-time setup):
```bash
go install github.com/magefile/mage@latest
```

This adds `mage` to your `$GOPATH/bin` (typically `~/go/bin`). Make sure this is in your `$PATH`.

2. **Build ffire CLI** (from repo root):
```bash
cd ..
go build -o ffire ./cmd/ffire
```

## Usage

### Full Workflow
```bash
mage bench
```

This will:
1. Generate ffire Go benchmark
2. Run the benchmark
3. Display comparison table
4. Save results to `results/comparison.md`

### Step by Step

```bash
# Generate all benchmarks
mage genAll

# Run Go benchmarks
mage runGo

# Show comparison table
mage compare

# Clean generated files
mage clean
```

## Output Structure

```
benchmarks/
├── magefile.go              # Orchestration
├── generated/               # Generated benchmarks (gitignored)
│   └── ffire_go/
│       ├── bench.go
│       ├── generated.go
│       ├── fixture.bin
│       └── go.mod
└── results/                 # Benchmark results (gitignored)
    ├── ffire_go.json
    └── comparison.md
```

## Example Output

### Terminal
```
=========================================================================================
BENCHMARK COMPARISON
=========================================================================================
Language     Format     Message         Encode       Decode        Total       Size
-----------------------------------------------------------------------------------------
Go           ffire      complex       13291 ns     23799 ns     37090 ns     4293 B
=========================================================================================
```

### JSON Output (for automation)
```json
{
  "language": "Go",
  "format": "ffire",
  "message": "complex",
  "iterations": 10000,
  "encode_ns": 13291,
  "decode_ns": 23799,
  "total_ns": 37090,
  "wire_size": 4293,
  "fixture_size": 4293,
  "timestamp": "2025-11-05T17:46:23+01:00"
}
```

## Adding More Languages

To add C++ or protobuf comparisons, extend the `magefile.go`:

```go
// Add to GenAll()
fmt.Println("📦 Generating ffire C++ benchmark...")
if err := sh.Run("../ffire", "bench",
    "--schema", schemaFile,
    "--json", dataFile,
    "--output", filepath.Join(genDir, "ffire_cpp"),
    "--lang", "cpp",
); err != nil {
    return err
}

// Add RunCpp() function
func RunCpp() error {
    // Build and run C++ benchmark
}
```

## Future Enhancements

- [ ] Add C++ benchmark support
- [ ] Add protobuf comparison (requires `to-proto` command)
- [ ] Add Rust benchmark support
- [ ] Add memory profiling
- [ ] Add flamegraph generation
- [ ] Support custom fixture files
