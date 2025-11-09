# Testing

ffire uses multiple test strategies to ensure correctness.

## Unit Tests

```bash
# All packages
go test ./...

# Specific package
go test ./pkg/generator
go test ./pkg/parser

# With coverage
go test -cover ./...
```

## Generator Tests

Each generator has tests in `pkg/generator/generator_*_test.go`:

```bash
go test ./pkg/generator -run TestGenerateGo
go test ./pkg/generator -run TestGenerateCpp
go test ./pkg/generator -run TestGenerateSwift
```

## Integration Tests

Located in `testdata/`:
- Schemas: `testdata/schema/*.ffi`
- Fixtures: `testdata/json/*.json`
- Proto reference: `testdata/proto/*.proto`

```bash
# Generate all test outputs
cd benchmarks
mage genAll

# Run cross-language tests
mage test
```

## Benchmark Tests

Benchmarks serve as integration tests - if encoding/decoding succeeds, codecs are working.

```bash
cd benchmarks
mage runGo      # Test Go codec
mage runCpp     # Test C++ codec  
mage runSwift   # Test Swift codec
```

## Adding Tests

### New Schema Test

1. Create: `testdata/schema/new.ffi`
2. Create: `testdata/json/new.json`
3. Run: `mage genAll` 
4. Verify: Generated code compiles

### New Generator Test

```go
func TestGenerateNewLang(t *testing.T) {
    schema := parseTestSchema(t)
    
    config := &PackageConfig{
        Schema:   schema,
        Language: "newlang",
        OutputDir: t.TempDir(),
    }
    
    err := GeneratePackage(config)
    if err != nil {
        t.Fatal(err)
    }
    
    // Verify output
}
```

## CI/CD

Tests run on GitHub Actions:
- Go unit tests
- Generator tests  
- Cross-compilation checks
- Benchmark smoke tests (verify they run, not performance)
