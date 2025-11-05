# ffire Code Generation Guide
## ffire - FFI Encoding

## Naming Convention

**Public API Function Names:** `Encode<TypeName>Message` / `Decode<TypeName>Message`

Where `<TypeName>` is the root type name with proper capitalization:
- Primitives: Capitalize first letter (`int32` → `Int32`, `string` → `String`, `bool` → `Bool`)
- Structs: Use struct name as-is (`Device` → `Device`, `Plugin` → `Plugin`)
- Arrays: Use element type name (`[]int32` → `Int32`, `[]Device` → `Device`)

**Examples:**
```
type Message = []int32       → EncodeInt32Message / DecodeInt32Message
type Message = []string      → EncodeStringMessage / DecodeStringMessage
type Message = []Device      → EncodeDeviceMessage / DecodeDeviceMessage
type Message = Device        → EncodeDeviceMessage / DecodeDeviceMessage
type PluginList = []Plugin   → EncodePluginMessage / DecodePluginMessage
```

**Rationale:**
- ✅ Better autocomplete (type prefix first)
- ✅ No collision with primitive helpers (`EncodeInt32` vs `EncodeInt32Message`)
- ✅ Clear distinction between internal helpers and public API
- ✅ Works consistently across all type combinations

## Design Philosophy

**Generate correct code first, format later.**

Use language-specific formatters to handle indentation, spacing, and style. Focus code generation logic on correctness and completeness, not aesthetics.

## Core Principles

### 1. Programmatic Generation (Not Templates)

**Use bytes.Buffer + WriteString operations:**

```go
// ✅ Good: Type-safe, composable, debuggable
func generateDecodeFunction(buf *bytes.Buffer, typeName string) {
    buf.WriteString("func Decode")
    buf.WriteString(typeName)
    buf.WriteString("(data []byte) (")
    buf.WriteString(typeName)
    buf.WriteString(", error) {\n")
    // ... more writes
    buf.WriteString("}\n")
}

// ❌ Avoid: Templates (AI struggles with escaping, hard to debug)
const tmpl = `func Decode{{.TypeName}}(data []byte) ({{.TypeName}}, error) {
    {{range .Fields}}...{{end}}
}`
```

**Benefits:**
- Type-safe (compiler catches errors)
- Composable (easy to build complex logic)
- Debuggable (step through generation)
- Refactorable (IDE support works)
- No escaping hell

### 2. Upfront Variable Declarations

**Use var() blocks, avoid := shadowing:**

```go
// ✅ Generated code style
func DecodePerson(data []byte) (Person, error) {
    var (
        result Person
        offset int
        length uint32
        err    error
    )
    
    // All logic uses pre-declared vars
    length, err = decodeUint32(data[offset:])
    if err != nil {
        return result, err
    }
    offset += 4
    
    result.Name, err = decodeString(data[offset:])
    if err != nil {
        return result, err
    }
    
    return result, nil
}
```

**Benefits:**
- No shadowing bugs
- Clear scope visibility
- Easier to generate (collect vars first, then logic)
- Go best practice (common in stdlib)

### 3. Let Formatters Handle Formatting

**Don't track indentation - generate ugly, format at end:**

```go
// Generate without worrying about spacing/indentation
func Generate(schema *schema.Schema) ([]byte, error) {
    buf := &bytes.Buffer{}
    
    // Write minimal valid code
    buf.WriteString("package ")
    buf.WriteString(schema.Package)
    buf.WriteString("\n")
    
    buf.WriteString("func Decode")
    buf.WriteString(typeName)
    buf.WriteString("(data []byte)(")
    buf.WriteString(typeName)
    buf.WriteString(",error){")
    buf.WriteString("var result ")
    buf.WriteString(typeName)
    buf.WriteString("\nreturn result,nil}")
    
    // Let formatter make it pretty
    return formatGo(buf.Bytes())
}
```

## Language-Specific Formatting

### Go - Use go/format

```go
import "go/format"

func formatGo(code []byte) ([]byte, error) {
    formatted, err := format.Source(code)
    if err != nil {
        // Return unformatted code + error for debugging
        return code, fmt.Errorf("format go code: %w", err)
    }
    return formatted, nil
}
```

**Built-in, always available, perfect for Go code.**

### C++ - Use clang-format

```go
import "os/exec"

func formatCpp(code []byte) ([]byte, error) {
    // Check availability
    if _, err := exec.LookPath("clang-format"); err != nil {
        log.Println("Warning: clang-format not found, C++ output not formatted")
        return code, nil  // Graceful degradation
    }
    
    cmd := exec.Command("clang-format", "--style=LLVM")
    cmd.Stdin = bytes.NewReader(code)
    
    formatted, err := cmd.Output()
    if err != nil {
        log.Printf("Warning: clang-format failed: %v", err)
        return code, nil  // Return unformatted rather than fail
    }
    
    return formatted, nil
}
```

**Configuration (.clang-format):**
```yaml
BasedOnStyle: LLVM
IndentWidth: 4
ColumnLimit: 100
AllowShortFunctionsOnASingleLine: Empty
PointerAlignment: Left
```

### Swift - Use swift-format

```go
func formatSwift(code []byte) ([]byte, error) {
    if _, err := exec.LookPath("swift-format"); err != nil {
        log.Println("Warning: swift-format not found, Swift output not formatted")
        return code, nil
    }
    
    cmd := exec.Command("swift-format")
    cmd.Stdin = bytes.NewReader(code)
    
    formatted, err := cmd.Output()
    if err != nil {
        return code, nil
    }
    
    return formatted, nil
}
```

## Multi-Pass Generation

Generate in logical sections, combine at end:

```go
type CodeGen struct {
    schema  *schema.Schema
    imports *bytes.Buffer
    types   *bytes.Buffer
    helpers *bytes.Buffer
    public  *bytes.Buffer
}

func (g *CodeGen) Generate() ([]byte, error) {
    // Pass 1: Imports
    g.generateImports()
    
    // Pass 2: Type definitions (if needed)
    g.generateTypes()
    
    // Pass 3: Private helpers
    for _, typ := range g.schema.Types {
        g.generateEncodeHelper(typ)
        g.generateDecodeHelper(typ)
    }
    
    // Pass 4: Public API
    for _, msg := range g.schema.Messages {
        g.generateEncodePublic(msg)
        g.generateDecodePublic(msg)
    }
    
    // Combine
    final := g.combine()
    
    // Format
    return formatGo(final)
}

func (g *CodeGen) combine() []byte {
    buf := &bytes.Buffer{}
    buf.Write(g.imports.Bytes())
    buf.WriteString("\n")
    buf.Write(g.types.Bytes())
    buf.WriteString("\n")
    buf.Write(g.helpers.Bytes())
    buf.WriteString("\n")
    buf.Write(g.public.Bytes())
    return buf.Bytes()
}
```

## Helper Functions for Common Patterns

```go
// Write function signature
func writeFunc(buf *bytes.Buffer, name, params, returns string) {
    fmt.Fprintf(buf, "func %s(%s) %s {\n", name, params, returns)
}

// Write var block
func writeVarBlock(buf *bytes.Buffer, vars map[string]string) {
    buf.WriteString("var (\n")
    for name, typ := range vars {
        fmt.Fprintf(buf, "%s %s\n", name, typ)
    }
    buf.WriteString(")\n")
}

// Write error check
func writeErrorCheck(buf *bytes.Buffer, returnZero string) {
    fmt.Fprintf(buf, "if err != nil {\nreturn %s, err\n}\n", returnZero)
}

// Write error check with context
func writeErrorCheckWithContext(buf *bytes.Buffer, context, returnZero string) {
    fmt.Fprintf(buf, "if err != nil {\nreturn %s, fmt.Errorf(\"%s: %%w\", err)\n}\n",
        returnZero, context)
}
```

## Generate Documentation

```go
func generateDecodeFunc(buf *bytes.Buffer, msg schema.MessageType) {
    // Doc comment
    fmt.Fprintf(buf, "// Decode%s decodes a %s from wire format.\n", 
        msg.Name, msg.Name)
    fmt.Fprintf(buf, "// Returns error if data is invalid or incomplete.\n")
    
    // Function
    fmt.Fprintf(buf, "func Decode%s(data []byte) (%s, error) {\n",
        msg.Name, msg.TargetType.Name())
    // ...
}
```

## Deterministic Output

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

## Error Context in Generated Code

Add helpful error messages:

```go
func generateFieldDecode(buf *bytes.Buffer, field schema.Field) {
    fmt.Fprintf(buf, "result.%s, err = decode%s(r)\n", 
        field.Name, field.Type.Name())
    
    // Add context to error
    fmt.Fprintf(buf, "if err != nil {\n")
    fmt.Fprintf(buf, "return result, fmt.Errorf(\"decode field %s: %%w\", err)\n",
        field.Name)
    fmt.Fprintf(buf, "}\n")
}
```

**Generated code produces errors like:**
```
decode field Name: unexpected EOF
decode field Channels: invalid length
```

## Complete Generation Example

```go
type GoGenerator struct {
    schema *schema.Schema
}

func (g *GoGenerator) Generate() ([]byte, error) {
    buf := &bytes.Buffer{}
    
    // Package
    fmt.Fprintf(buf, "package %s\n", g.schema.Package)
    
    // Imports
    buf.WriteString("import(\n")
    buf.WriteString("\"bytes\"\n")
    buf.WriteString("\"encoding/binary\"\n")
    buf.WriteString("\"fmt\"\n")
    buf.WriteString(")\n")
    
    // For each message type, generate encode/decode
    for _, msg := range g.schema.Messages {
        g.generateEncode(buf, msg)
        g.generateDecode(buf, msg)
    }
    
    // Format and return
    return format.Source(buf.Bytes())
}

func (g *GoGenerator) generateDecode(buf *bytes.Buffer, msg schema.MessageType) {
    typeName := msg.Name
    
    // Doc comment
    fmt.Fprintf(buf, "// Decode%s decodes %s from wire format.\n", 
        typeName, typeName)
    
    // Signature
    fmt.Fprintf(buf, "func Decode%s(data []byte)(%s,error){\n",
        typeName, msg.TargetType.Name())
    
    // Var block
    buf.WriteString("var(\n")
    fmt.Fprintf(buf, "result %s\n", msg.TargetType.Name())
    buf.WriteString("r=bytes.NewReader(data)\n")
    buf.WriteString("err error\n")
    buf.WriteString(")\n")
    
    // Decode logic based on type
    g.generateDecodeLogic(buf, msg.TargetType)
    
    // Return
    buf.WriteString("return result,nil\n")
    buf.WriteString("}\n")
}
```

## Testing Generated Code

Generate tests alongside code:

```go
func GenerateWithTests(schema *schema.Schema) (code, tests []byte) {
    code = generateCode(schema)
    tests = generateTests(schema)
    return
}

func generateTests(schema *schema.Schema) []byte {
    buf := &bytes.Buffer{}
    
    fmt.Fprintf(buf, "package %s\n", schema.Package)
    buf.WriteString("import \"testing\"\n")
    
    // Generate round-trip test for each message type
    for _, msg := range schema.Messages {
        fmt.Fprintf(buf, "func TestRoundTrip%s(t *testing.T){\n", msg.Name)
        // ... test logic
        buf.WriteString("}\n")
    }
    
    return buf.Bytes()
}
```

## Best Practices Summary

1. ✅ **bytes.Buffer + WriteString** - No templates
2. ✅ **Upfront var() blocks** - No := shadowing
3. ✅ **Format at end** - go/format, clang-format, swift-format
4. ✅ **Multi-pass generation** - Imports, types, helpers, API
5. ✅ **Helper functions** - DRY common patterns
6. ✅ **Generate comments** - Document generated code
7. ✅ **Sort deterministically** - Consistent output
8. ✅ **Error context** - Helpful error messages
9. ✅ **Generate tests** - Validate correctness
10. ✅ **Graceful degradation** - Work without formatters

## Schema Analysis Pass

Before generating code, analyze the schema to enable optimizations:

```go
type TypeInfo struct {
    IsFixedSize bool  // All fields are non-optional primitives?
    FixedSize   int   // If IsFixedSize=true, exact byte size
    MaxSize     int   // Maximum possible size (with all optionals present)
    HasStrings  bool  // Contains any string fields?
    HasArrays   bool  // Contains any array fields?
    NestDepth   int   // Maximum nesting depth
}

func AnalyzeSchema(s *schema.Schema) map[string]TypeInfo
```

**Use analysis results to:**
- Simplify error handling for fixed-size types
- Pre-compute maximum message sizes
- Detect types that need no bounds checking (thanks to uint16 limits)
- Optimize generation strategy per type

## Implementation Order

1. **Analyze schema** - Compute type properties and constraints
2. **Generate minimal valid code** - Focus on correctness
3. **Add var blocks and error handling** - Make it robust
4. **Add documentation** - Make it understandable
5. **Run through formatter** - Make it pretty
6. **Generate tests** - Make it verifiable

**Keep generation logic simple. Let formatters handle style.**
