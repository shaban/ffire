# Test Data Generator

Generates test schemas (`.ffi`) and JSON data for ffire benchmarking.

## Usage

```bash
cd experimental/testdata
go run generator.go
```

This will generate:
- `../../testdata/schema/*.ffi` - Schema definitions
- `../../testdata/json/*.json` - Test data

## Generated Test Cases

1. **array_int** - Array of 5000 int32s
2. **array_float** - Array of 5000 float32s
3. **array_string** - Array of 500 strings (~40 chars each)
4. **array_struct** - Array of 200 device structs
5. **struct** - Single config struct
6. **nested** - 10-level nested structure with array at bottom
7. **complex** - 20 plugins with parameters (realistic use case)
8. **optional** - 1000 records with optional fields
9. **empty** - Edge case with empty string/array

All tests are sized to produce similar benchmark runtimes (±40% deviation).

## Regenerating

If you need to change test data:
1. Edit `generator.go`
2. Run `go run generator.go`
3. Commit the new generated files

Generated data is deterministic and checked into git.
