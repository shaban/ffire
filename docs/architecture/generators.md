# Language Generators

How to implement a new language generator for ffire.

## Implementation Status

| Language | Package Generator | Benchmark Generator | Magefile Runner | Status |
|----------|------------------|---------------------|-----------------|--------|
| **Go** | ✅ Full | ✅ Yes | ✅ `mage runGo` | Production Ready |
| **C++** | ✅ Full | ✅ Yes | ✅ `mage runCpp` | Production Ready |
| **Python** | ✅ Full | ✅ Yes | ✅ `mage runPython` | Production Ready |
| **Swift** | ✅ Full | ✅ Yes | ✅ `mage runSwift` | Production Ready |
| **Dart** | ✅ Full | ✅ Yes | ✅ `mage runDart` | Production Ready |
| **JavaScript** | ✅ Full | ✅ Yes | ❌ Not integrated | Ready for integration |
| **Ruby** | ✅ Full | ✅ Yes | ❌ Not integrated | Ready for integration |
| **Java** | ✅ Full | ✅ Yes | ❌ Not integrated | Ready for integration |
| **C#** | ✅ Full | ✅ Yes | ❌ Not integrated | Ready for integration |
| **PHP** | ✅ Full | ✅ Yes | ❌ Not integrated | Ready for integration |

**Legend:**
- ✅ **Full**: Complete implementation with all features
- ✅ **Yes**: Benchmark harness generator exists
- ✅ **mage run***: Automated test runner integrated
- ❌ **Not integrated**: Exists but not in automated benchmark suite

**Tested Performance (10k iterations, struct benchmark):**
- Go: 178 ns/op (native, no FFI)
- C++: 255 ns/op (native, -O3 -march=native)
- Swift: 420 ns/op (FFI via C ABI)
- Dart: 2,370 ns/op (FFI via dart:ffi)
- Python: 1,619 ns/op (FFI via ctypes/pybind11)

See [Benchmarks](../development/benchmarks.md) for detailed performance analysis.

## Generator Interface

Each generator implements code generation for a target language.

Location: `pkg/generator/generator_LANG.go`

```go
func GenerateLANGPackage(config *PackageConfig) error {
    // 1. Parse schema
    // 2. Generate code files
    // 3. Optionally compile
    // 4. Create package structure
    return nil
}
```

## Implementation Steps

### 1. Create Generator File

`pkg/generator/generator_newlang.go`:

```go
package generator

func GenerateNewLangPackage(config *PackageConfig) error {
    outputDir := config.OutputDir
    schema := config.Schema
    
    // Generate source files
    for _, msg := range schema.Messages {
        code := generateNewLangType(msg)
        writeFile(outputDir, msg.Name+".ext", code)
    }
    
    return nil
}
```

### 2. Type Mapping

Map ffire types to target language:

```go
func newLangType(t schema.Type) string {
    switch t := t.(type) {
    case *schema.PrimitiveType:
        switch t.Name {
        case "int32": return "int"
        case "int64": return "long"
        case "float32": return "float"
        case "string": return "string"
        // ...
        }
    case *schema.ArrayType:
        return "List<" + newLangType(t.ElementType) + ">"
    case *schema.StructType:
        return t.Name
    }
}
```

### 3. Encode Generation

```go
func generateEncode(msg *schema.MessageType) string {
    buf := &bytes.Buffer{}
    
    fmt.Fprintf(buf, "function encode() {\n")
    fmt.Fprintf(buf, "  let buf = new Buffer();\n")
    
    // Generate encoding logic per field
    for _, field := range msg.Fields {
        generateEncodeField(buf, field)
    }
    
    fmt.Fprintf(buf, "  return buf.bytes();\n")
    fmt.Fprintf(buf, "}\n")
    
    return buf.String()
}
```

### 4. Decode Generation

```go
func generateDecode(msg *schema.MessageType) string {
    // Generate decoding logic
    // Read fields from buffer
    // Return decoded instance
}
```

### 5. Wire Format Implementation

Follow [Wire Format](../architecture/wire-format.md) specification:
- Varints for integers
- Length-prefixed strings
- Array length + elements
- Optional presence bits

### 6. Package Structure

Create idiomatic package for target language:

**Go**: Single package with encode/decode functions
**C++**: Header + implementation, compiled dylib
**Swift**: Package.swift + Sources/
**Python**: setup.py + module/
**JavaScript**: package.json + index.js

### 7. Reserved Keywords

Handle language keywords - see [Reserved Keywords](../development/keywords.md).

### 8. Add to CLI

In `cmd/ffire/gen.go`:

```go
case "newlang":
    if err := generator.GenerateNewLangPackage(config); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
```

### 9. Add Benchmark Support

In `pkg/benchmark/benchmark_newlang.go`:

```go
func GenerateNewLang(
    schema *schema.Schema,
    schemaName, messageName string,
    jsonData []byte,
    outputDir string,
    iterations int,
) error {
    // Generate benchmark harness
}
```

### 10. Test

```bash
# Generate test
ffire gen --lang newlang --schema testdata/schema/array_int.ffi --output /tmp/test

# Verify output
ls /tmp/test

# Run if executable
cd /tmp/test && ./run_test
```

## Examples

Reference implementations:
- **Go**: `generator_go.go` - Native, no FFI
- **C++**: `generator_cpp.go` - Header/impl, dylib compile
- **Swift**: `generator_swift.go` - Package + C ABI
- **Dart**: `generator_dart.go` - FFI wrapper generation
- **Python**: `generator_python.go` - ctypes bindings

## Common Patterns

### Native Implementation
Language has its own encoder (Go, C++, Rust).

### FFI Wrapper
Wraps C ABI dylib (Swift, Dart, Python, JavaScript).

### Hybrid
Some native, some FFI (Swift uses both).

## Performance Tips

- Pre-allocate buffers
- Bulk operations for arrays
- Zero-copy where possible
- Minimize allocations
- Profile before optimizing

See [Optimizations](../internals/optimizations.md) for techniques.

---

## Code Generation Best Practices

### Naming Convention

**Automatic `Message` Suffix for Generated Types**

Users write clean type names in schemas:
```go
type Config struct { ... }
type Device struct { ... }
type User struct { ... }
```

Generators automatically append `Message` suffix:
```go
// Generated Go
type ConfigMessage struct { ... }
type DeviceMessage struct { ... }
type UserMessage struct { ... }
```

**Implementation:**
```go
func messageTypeName(name string) string {
    return name + "Message"
}
```

**Rationale:**
- ✅ **User-friendly**: Clean schema names
- ✅ **Zero keyword collisions** across all 11 languages
- ✅ **Industry standard** (Protobuf, gRPC use "Message")
- ✅ **Clear boundary**: Generated types ≠ domain types
- ✅ **No maintenance** - no keyword lists needed

**Public API Function Names:** `Encode<TypeName>Message` / `Decode<TypeName>Message`

Examples:
```go
// Schema: type Config struct { ... }
// Generates: 
type ConfigMessage struct { ... }
func EncodeConfigMessage(msg *ConfigMessage) ([]byte, error)
func DecodeConfigMessage(data []byte) (*ConfigMessage, error)

// Schema: type Device struct { ... }
// Generates:
type DeviceMessage struct { ... }
func EncodeDeviceMessage(msg *DeviceMessage) ([]byte, error)
func DecodeDeviceMessage(data []byte) (*DeviceMessage, error)
```

**Rationale:**
- ✅ Better autocomplete (type prefix first)
- ✅ No collision with primitive helpers
- ✅ Clear distinction between internal helpers and public API

### Design Philosophy

**Generate correct code first, format later.**

Use language-specific formatters (go/format, clang-format, swift-format) to handle indentation and style. Focus generation logic on correctness, not aesthetics.

### Core Principles

#### 1. Programmatic Generation (Not Templates)

**Use bytes.Buffer + WriteString:**

```go
// ✅ Good: Type-safe, composable, debuggable
func generateDecodeFunction(buf *bytes.Buffer, typeName string) {
    buf.WriteString("func Decode")
    buf.WriteString(typeName)
    buf.WriteString("(data []byte) (")
    // ... more writes
}

// ❌ Avoid: Templates (hard to debug, escaping issues)
```

**Benefits:**
- Type-safe (compiler catches errors)
- Debuggable (step through generation)
- Refactorable (IDE support works)

#### 2. Upfront Variable Declarations

Use var() blocks, avoid := shadowing:

```go
// ✅ Generated code style
func DecodePerson(data []byte) (Person, error) {
    var (
        result Person
        offset int
        err    error
    )
    // All logic uses pre-declared vars
    return result, nil
}
```

#### 3. Let Formatters Handle Formatting

Generate minimal valid code, then format:

```go
func Generate(schema *schema.Schema) ([]byte, error) {
    buf := &bytes.Buffer{}
    // Write minimal valid code
    // ...
    // Let formatter make it pretty
    return formatGo(buf.Bytes())
}
```

### Language-Specific Formatting

**Go:** Use `go/format` (built-in)

**C++:** Use `clang-format` with LLVM style
```yaml
# .clang-format
BasedOnStyle: LLVM
IndentWidth: 4
ColumnLimit: 100
```

**Swift:** Use `swift-format`

### Multi-Pass Generation

Generate in logical sections, combine at end:

```go
type CodeGen struct {
    imports *bytes.Buffer
    types   *bytes.Buffer
    helpers *bytes.Buffer
    public  *bytes.Buffer
}

func (g *CodeGen) Generate() ([]byte, error) {
    g.generateImports()
    g.generateTypes()
    g.generateHelpers()
    g.generatePublicAPI()
    return g.combine(), nil
}
```

### Deterministic Output

**Always sort when iterating maps:**

```go
// ❌ Bad: Random order each run
for name, field := range schema.Fields {
    generateField(buf, name, field)
}

// ✅ Good: Deterministic order
names := make([]string, 0, len(schema.Fields))
for name := range schema.Fields {
    names = append(names, name)
}
sort.Strings(names)
for _, name := range names {
    generateField(buf, name, schema.Fields[name])
}
```

**Benefits:**
- Diff-able output
- Consistent git history
- Reproducible builds

### Error Context in Generated Code

Add helpful error messages:

```go
// Generated code includes context
if err != nil {
    return result, fmt.Errorf("decode field %s: %w", field.Name, err)
}
```

**Generated errors:**
```
decode field Name: unexpected EOF
decode field Parameters: invalid length
```

### Schema Analysis Pass

Before generating code, analyze the schema:

```go
type TypeInfo struct {
    IsFixedSize bool
    FixedSize   int
    HasStrings  bool
    HasArrays   bool
    NestDepth   int
}

func AnalyzeSchema(s *schema.Schema) map[string]TypeInfo
```

**Use analysis results to:**
- Simplify error handling for fixed-size types
- Pre-compute maximum message sizes
- Optimize generation strategy per type

### Implementation Order

1. **Analyze schema** - Compute type properties
2. **Generate minimal valid code** - Focus on correctness
3. **Add error handling** - Make it robust
4. **Add documentation** - Make it understandable
5. **Run through formatter** - Make it pretty
6. **Generate tests** - Make it verifiable

**Keep generation logic simple. Let formatters handle style.**
