# CLI Reference

Command-line interface for ffire code generation.

## Installation

```bash
go install github.com/shaban/ffire/cmd/ffire@latest
```

## Commands

### `ffire gen`

Generate codec for a schema.

```bash
ffire gen --lang go --schema types.ffi --output ./generated
```

**Options:**
- `--lang` - Target language: `go`, `cpp`, `csharp`, `java`, `swift`, `dart`, `rust`, `zig`
- `--schema` - Input schema file (`.ffi`)
- `--output` - Output directory

### `ffire bench`

Generate benchmark harness.

```bash
ffire bench --lang go --schema array_int.ffi --json fixture.json --output ./bench
```

**Options:**
- `--lang` - Target language
- `--schema` - Schema file
- `--json` - Fixture data (JSON)
- `--output` - Output directory
- `--iterations` - Benchmark iterations (default: 10000)

## Examples

**Go package:**
```bash
ffire gen --lang go --schema api.ffi --output ./api
```

**C++ library:**
```bash
ffire gen --lang cpp --schema messages.ffi --output ./build
```

**Swift package:**
```bash
ffire gen --lang swift --schema types.ffi --output ./swift
cd swift && swift build
```

**Multi-language:**
```bash
for lang in go cpp csharp java swift dart rust zig; do
  ffire gen --lang $lang --schema schema.ffi --output ./gen/$lang
done
```
