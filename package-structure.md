# ffire Package Structure
## ffire - FFI Encoding

## Design Principle: DRY (Don't Repeat Yourself)

All CLI tools compose functionality from reusable packages. No logic duplication between commands.

## Package Organization

```
ffire/
├── cmd/
│   └── ffire/
│       ├── main.go              # CLI entry point
│       ├── generate.go          # generate subcommand
│       ├── validate.go          # validate subcommand
│       ├── fixture.go           # fixture subcommand
│       └── bench.go             # bench subcommand
│
├── pkg/
│   ├── schema/                  # Schema representation and AST
│   │   ├── ast.go              # Schema AST types
│   │   ├── types.go            # Type system (primitives, composites)
│   │   └── validate.go         # Schema validation rules
│   │
│   ├── parser/                  # Parse .ffir files
│   │   ├── parser.go           # Go syntax parser (uses go/parser)
│   │   └── loader.go           # Load and resolve type references
│   │
│   ├── analyzer/                # Schema analysis for optimization
│   │   ├── analyzer.go         # Analyze type properties
│   │   └── typeinfo.go         # TypeInfo struct and utilities
│   │
│   ├── wire/                    # Wire format encode/decode (runtime)
│   │   ├── encode.go           # Generic encoding logic
│   │   ├── decode.go           # Generic decoding logic
│   │   └── primitives.go       # Primitive type handling
│   │
│   ├── generator/               # Code generation orchestration
│   │   ├── generator.go        # Main generation interface
│   │   ├── go.go               # Go code generator
│   │   ├── cpp.go              # C++ code generator
│   │   └── swift.go            # Swift package generator (wraps C++)
│   │
│   ├── validator/               # Schema and data validation
│   │   ├── schema.go           # Validate schema structure
│   │   ├── json.go             # Validate JSON against schema
│   │   └── errors.go           # Validation error types
│   │
│   ├── fixture/                 # Binary test fixture generation
│   │   ├── fixture.go          # Generate .bin from JSON
│   │   └── json.go             # JSON parsing and conversion
│   │
│   └── benchmark/               # Benchmark code generation
│       ├── benchmark.go        # Benchmark generation interface
│       ├── go.go               # Go benchmark template
│       └── cpp.go              # C++ benchmark template
│
└── internal/
    └── testdata/                # Test schemas and fixtures
```

## Package Responsibilities

### `schema` - Schema Representation
**Purpose**: Core type system and AST representation

```go
type Schema struct {
    Package    string
    Messages   []MessageType    // Types with public encode/decode
    Types      []Type           // All type definitions
}

type MessageType struct {
    Name       string
    TargetType Type             // What it aliases
}

type Type interface {
    Name() string
    Validate() error
}

type StructType struct {
    Name   string
    Fields []Field
}

type Field struct {
    Name     string
    Type     Type
    Optional bool
}

// Validate ensures schema is well-formed
func (s *Schema) Validate() error
```

**Used by**: parser, generator, validator, fixture, benchmark

### `parser` - Parse .ffi Files
**Purpose**: Parse Go-syntax schema files into AST

```go
// Parse a .ffi file into Schema
func Parse(filepath string) (*schema.Schema, error)

// ParseBytes parses from in-memory source
func ParseBytes(source []byte) (*schema.Schema, error)
```

**Dependencies**: `go/parser`, `go/ast`, `schema`  
**Used by**: All CLI commands

### `analyzer` - Schema Analysis
**Purpose**: Analyze schemas to enable code generation optimizations

```go
type TypeInfo struct {
    IsFixedSize bool  // All non-optional primitives?
    FixedSize   int   // Exact size if fixed
    MaxSize     int   // Max size with optionals present
    HasStrings  bool
    HasArrays   bool
    NestDepth   int
}

// Analyze all types in schema
func Analyze(schema *schema.Schema) map[string]TypeInfo
```

**Dependencies**: `schema`  
**Used by**: `generator`

### `wire` - Runtime Wire Format
**Purpose**: Core encoding/decoding logic (used by generated code)

```go
// Low-level primitives
func EncodeUint32(buf *bytes.Buffer, v uint32)
func DecodeUint32(r io.Reader) (uint32, error)

func EncodeString(buf *bytes.Buffer, s string)
func DecodeString(r io.Reader) (string, error)

// Array helpers
func EncodeArrayHeader(buf *bytes.Buffer, count uint32)
func DecodeArrayHeader(r io.Reader) (uint32, error)
```

**Dependencies**: None (pure encode/decode)  
**Used by**: Generated code (imported at runtime)

### `generator` - Code Generation
**Purpose**: Generate encoder/decoder code for target languages

```go
type Generator interface {
    Generate(schema *schema.Schema, output string) error
}

// Language-specific generators
func NewGoGenerator() Generator
func NewCppGenerator() Generator
func NewSwiftGenerator() Generator

// Generate all requested languages
func GenerateAll(schema *schema.Schema, langs []string, output string) error
```

**Dependencies**: `schema`, `text/template`  
**Used by**: `generate` CLI command

### `validator` - Validation
**Purpose**: Validate schemas and JSON data

```go
// Validate schema syntax and semantics
func ValidateSchema(schema *schema.Schema) error

// Validate JSON matches schema
func ValidateJSON(schema *schema.Schema, jsonData []byte) error

// Validation errors with context
type ValidationError struct {
    Field   string
    Message string
}
```

**Dependencies**: `schema`, `encoding/json`  
**Used by**: `validate` CLI command, `fixture` package

### `fixture` - Binary Fixture Generation
**Purpose**: Convert JSON test data to binary format

```go
// Generate binary fixture from JSON
func Generate(schema *schema.Schema, jsonData []byte) ([]byte, error)

// Generate and write to file
func GenerateFile(schema *schema.Schema, jsonPath, outputPath string) error
```

**Dependencies**: `schema`, `validator`, `wire`, `encoding/json`  
**Used by**: `fixture` CLI command, `benchmark` package

### `benchmark` - Benchmark Generation
**Purpose**: Generate standalone benchmark executables

```go
type BenchmarkGenerator interface {
    Generate(schema *schema.Schema, fixtureData []byte, output string, iterations int) error
}

func NewGoBenchmarkGenerator() BenchmarkGenerator
func NewCppBenchmarkGenerator() BenchmarkGenerator
```

**Dependencies**: `schema`, `generator`, `fixture`, `text/template`  
**Used by**: `bench` CLI command

## CLI Command Composition

### `ffire generate`
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

### `ffire validate`
```go
func runValidate(schemaPath, jsonPath string) error {
    // 1. Parse schema
    schema, err := parser.Parse(schemaPath)
    
    // 2. Validate schema
    if err := validator.ValidateSchema(schema); err != nil {
        return err
    }
    
    // 3. Validate JSON if provided
    if jsonPath != "" {
        jsonData, _ := os.ReadFile(jsonPath)
        return validator.ValidateJSON(schema, jsonData)
    }
    
    return nil
}
```

### `ffire fixture`
```go
func runFixture(schemaPath, jsonPath, output string) error {
    // 1. Parse schema
    schema, err := parser.Parse(schemaPath)
    
    // 2. Validate schema
    if err := validator.ValidateSchema(schema); err != nil {
        return err
    }
    
    // 3. Generate fixture (validates JSON internally)
    return fixture.GenerateFile(schema, jsonPath, output)
}
```

### `ffire bench`
```go
func runBench(schemaPath, jsonPath, lang, output string, iterations int) error {
    // 1. Parse schema
    schema, err := parser.Parse(schemaPath)
    
    // 2. Validate schema
    if err := validator.ValidateSchema(schema); err != nil {
        return err
    }
    
    // 3. Generate fixture
    fixtureData, err := fixture.Generate(schema, jsonData)
    
    // 4. Generate benchmark code
    var gen benchmark.BenchmarkGenerator
    switch lang {
    case "go":
        gen = benchmark.NewGoBenchmarkGenerator()
    case "cpp":
        gen = benchmark.NewCppBenchmarkGenerator()
    }
    
    return gen.Generate(schema, fixtureData, output, iterations)
}
```

## Dependency Graph

```
         ┌──────────┐
         │  schema  │  (Core type system)
         └────┬─────┘
              │
    ┌─────────┼──────────┬─────────┐
    │         │          │         │
┌───▼───┐ ┌──▼──────┐ ┌─▼──────┐ ┌▼────────┐
│parser │ │validator│ │analyzer│ │  wire   │
└───┬───┘ └──┬──────┘ └─┬──────┘ └─────────┘
    │        │           │             │
    │   ┌────▼─────┐◄────┘             │
    └───►  fixture │◄──────────────────┘
        └────┬─────┘
             │
    ┌────────┼────────┐
    │        │        │
┌───▼─────┐  │  ┌─────▼──────┐
│generator│◄─┘  │ benchmark  │
│(uses    │     │            │
│analyzer)│     └────────────┘
└─────────┘
```

**Key principle**: Lower packages have no dependencies on higher packages. Pure bottom-up composition.

## Implementation Order

1. **schema** - Define AST and type system ✅
2. **parser** - Parse .ffi files into schema ✅
3. **validator** - Validate schemas ✅
4. **wire** - Core encode/decode primitives ✅
5. **analyzer** - Analyze schemas for optimization 🔄
6. **generator** (Go) - Generate Go code 🔄
7. **fixture** - JSON to binary conversion ✅
8. **benchmark** (Go) - Generate Go benchmarks
9. **generator** (C++) - Generate C++ code
10. **benchmark** (C++) - Generate C++ benchmarks
11. **generator** (Swift) - Swift package wrapper

## Testing Strategy

Each package has unit tests with testdata:
```
pkg/parser/parser_test.go
pkg/parser/testdata/
  valid_schema.ffi
  invalid_schema.ffi

pkg/generator/go_test.go
pkg/generator/testdata/
  simple.ffi
  nested.ffi
  expected_output.go
```

Integration tests in `cmd/ffire`:
```
cmd/ffire/integration_test.go
cmd/ffire/testdata/
  audio.ffi
  audio_test.json
```

## Benefits of This Structure

✅ **DRY** - Zero duplication between CLI commands  
✅ **Testable** - Each package tested independently  
✅ **Composable** - Easy to add new commands/languages  
✅ **Clear dependencies** - Bottom-up, no cycles  
✅ **Maintainable** - Changes isolated to single package  
✅ **Reusable** - Packages usable as library (not just CLI)
