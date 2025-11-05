# ffire CLI Specification
## ffire - FFI Encoding

## Overview

Single unified CLI with subcommands following modern conventions (git, docker, cargo).

```bash
ffire <subcommand> [flags]
```

## Commands

### `ffire generate`
Generate encoder/decoder code from schema.

**Usage:**
```bash
ffire generate -schema <file> [-lang <langs>] [-output <dir>]
```

**Flags:**
- `-schema string` - Schema file path (required)
- `-output string` - Output directory (default: ".")
- `-lang string` - Languages to generate, comma-separated (default: "go")
  - Options: `go`, `cpp`, `swift`, `all`
  - Swift is a wrapper around C++ with Package.swift and module definition

**Examples:**
```bash
# Generate Go code in current directory
ffire generate -schema audio.ffi

# Generate Go and C++ code
ffire generate -schema audio.ffi -lang go,cpp

# Generate all languages in specific directory
ffire generate -schema audio.ffi -lang all -output ./generated
```

---

### `ffire validate`
Validate schema syntax and optionally validate JSON against schema.

**Usage:**
```bash
ffire validate -schema <file> [-json <file>]
```

**Flags:**
- `-schema string` - Schema file path (required)
- `-json string` - JSON data to validate against schema (optional)

**Examples:**
```bash
# Validate schema syntax only
ffire validate -schema audio.ffi

# Validate JSON matches schema
ffire validate -schema audio.ffi -json test_data.json
```

---

### `ffire fixture`
Generate binary test fixtures from JSON.

**Usage:**
```bash
ffire fixture -schema <file> -json <file> [-output <file>]
```

**Flags:**
- `-schema string` - Schema file path (required)
- `-json string` - JSON source data (required)
- `-output string` - Binary output path (default: `<schema_name>.bin`)

**Examples:**
```bash
# Generate fixture with default name (audio.bin)
ffire fixture -schema audio.ffi -json test_data.json

# Generate fixture with custom name
ffire fixture -schema audio.ffi -json test_data.json -output custom.bin
```

---

### `ffire bench`
Generate standalone benchmark executable for performance evaluation.

**Usage:**
```bash
ffire bench -schema <file> -json <file> [-lang <lang>] [-output <dir>] [-iterations <N>]
```

**Flags:**
- `-schema string` - Schema file path (required)
- `-json string` - JSON test data (required, used to generate fixture)
- `-lang string` - Language (default: "go", options: go, cpp, swift)
- `-output string` - Output directory (default: `/tmp/ffire-bench-<hash>`)
- `-iterations int` - Benchmark iterations (default: 1000000)

**Generated Output:**
```
<output>/
  <schema>_ffire.{go,cpp,h}  # Generated encoder/decoder
  bench_main.{go,cpp}         # Benchmark executable
  test_data.bin               # Binary fixture
  {go.mod,CMakeLists.txt}     # Build configuration
```

**Examples:**
```bash
# Generate Go benchmark in temp directory
ffire bench -schema audio.ffi -json test_data.json

# Generate C++ benchmark in specific directory
ffire bench -schema audio.ffi -json test_data.json -lang cpp -output ./bench_cpp

# Custom iteration count
ffire bench -schema audio.ffi -json test_data.json -iterations 10000000
```

---

## Global Flags

- `-h, --help` - Show help for command
- `-v, --version` - Show ffire version

**Examples:**
```bash
ffire --help
ffire generate --help
ffire --version
```

---

## Implementation Architecture

### CLI Structure
```
cmd/ffire/
├── main.go          # CLI entry point, subcommand routing
├── generate.go      # generate subcommand implementation
├── validate.go      # validate subcommand implementation
├── fixture.go       # fixture subcommand implementation
└── bench.go         # bench subcommand implementation
```

### Package Structure (DRY Principle)
```
pkg/
├── schema/          # Core type system and AST
├── parser/          # Parse .ffi files into schema
├── wire/            # Runtime wire format encode/decode
├── generator/       # Code generation orchestration
│   ├── go.go       # Go code generator
│   ├── cpp.go      # C++ code generator
│   └── swift.go    # Swift package generator (wraps C++)
├── validator/       # Schema and JSON validation
├── fixture/         # Binary fixture generation
└── benchmark/       # Benchmark code generation
```

### Command Composition (DRY)

Each CLI command is thin - parses flags and orchestrates pkg functions:

**Example: `ffire generate`**
```go
func runGenerate(schemaPath, output string, langs []string) error {
    // 1. Parse schema
    schema, err := parser.Parse(schemaPath)
    
    // 2. Validate schema
    if err := validator.ValidateSchema(schema); err != nil {
        return err
    }
    
    // 3. Generate code for each language
    return generator.GenerateAll(schema, langs, output)
}
```

**Example: `ffire bench`**
```go
func runBench(schemaPath, jsonPath, lang, output string, iterations int) error {
    // 1. Parse and validate schema
    schema, err := parser.Parse(schemaPath)
    if err := validator.ValidateSchema(schema); err != nil {
        return err
    }
    
    // 2. Generate fixture
    jsonData, _ := os.ReadFile(jsonPath)
    fixtureData, err := fixture.Generate(schema, jsonData)
    
    // 3. Generate benchmark code
    gen := benchmark.NewGenerator(lang)
    return gen.Generate(schema, fixtureData, output, iterations)
}
```

All functionality is in packages, CLI is just composition layer.

---

## Workflow Examples

### Complete Development Workflow
```bash
# 1. Create schema
cat > audio.ffi << EOF
package audio
type DeviceList = []Device
type Device struct {
    Name string
    Channels int32
}
EOF

# 2. Validate schema
ffire validate -schema audio.ffi

# 3. Generate code for all languages
ffire generate -schema audio.ffi -lang all -output ./gen

# 4. Create test data
cat > test_data.json << EOF
[{"Name": "Speaker", "Channels": 2}]
EOF

# 5. Validate test data
ffire validate -schema audio.ffi -json test_data.json

# 6. Generate binary fixture for tests
ffire fixture -schema audio.ffi -json test_data.json

# 7. Benchmark performance
ffire bench -schema audio.ffi -json test_data.json -lang go
ffire bench -schema audio.ffi -json test_data.json -lang cpp
```

### Testing Workflow
```bash
# Generate test fixtures for all test schemas
for schema in testdata/schema/*.ffi; do
    name=$(basename $schema .ffi)
    ffire fixture -schema $schema \
                  -json testdata/json/$name.json \
                  -output testdata/bin/$name.bin
done

# Run benchmarks for all test cases
for schema in testdata/schema/*.ffi; do
    name=$(basename $schema .ffi)
    ffire bench -schema $schema \
                -json testdata/json/$name.json \
                -output /tmp/bench_$name
done
```

---

## Design Benefits

✅ **Modern convention** - Follows git/docker/cargo patterns  
✅ **Single binary** - Install once, all functionality available  
✅ **Consistent interface** - Same flags across commands  
✅ **Easy discovery** - `ffire help` shows all commands  
✅ **DRY implementation** - CLI is thin layer over reusable packages  
✅ **Future-proof** - Easy to add new subcommands  
✅ **Good help text** - Command-specific help with `--help`