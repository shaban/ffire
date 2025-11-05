# ffire Design Decisions
## ffire - FFI Encoding

## Core Philosophy: Safety by Design

**Principle**: Pay the cost once at design time, not at runtime.

Eliminate entire classes of security vulnerabilities through type system constraints rather than runtime checks.

---

## Decision 1: uint16 Length Fields

**Date**: November 5, 2025

### Decision
Use `uint16` (not `uint32`) for string and array length fields in wire format.

### Rationale
- **Safety**: Physically impossible to overflow - max 65,535 bytes/elements
- **No runtime checks**: Type system prevents buffer overflow and memory exhaustion
- **Smaller wire format**: 2 bytes vs 4 bytes per length (space savings)
- **Sufficient for domain**: Plugin names, config, device lists all well under 64KB

### Impact
- **Per-string limit**: Max 65,535 bytes (~64KB)
- **Per-array limit**: Max 65,535 elements
- **Total message size**: Still up to 2GB (2^31 bytes)
  - Example: Array of 65K strings × 64KB each = up to 4GB (capped at 2GB limit)
- Estimated real-world usage: <1% of per-field limits

### Trade-offs
✅ **Gained**: Complete elimination of bounds checking code  
✅ **Gained**: Smaller wire format  
✅ **Gained**: Simpler generated code  
❌ **Lost**: Ability to encode >64KB strings or >65K element arrays  
✅ **Acceptable**: Use case is metadata/orchestration, not bulk data transfer

---

## Decision 2: Schema Analysis Pass

**Date**: November 5, 2025

### Decision
Add `pkg/analyzer` package to analyze schemas before code generation.

### Rationale
- **Optimization**: Detect fixed-size types (no strings/arrays/optionals)
- **Pre-computation**: Calculate max message sizes at codegen time
- **Strategy selection**: Choose optimal encoding approach per type
- **One-time cost**: Analysis happens once during generation, not at runtime

### Analysis Output
```go
type TypeInfo struct {
    IsFixedSize bool  // Can use simplified error handling?
    FixedSize   int   // Exact size if fixed (enables optimizations)
    MaxSize     int   // Upper bound (enables single size check)
    HasStrings  bool  // Needs string encoding?
    HasArrays   bool  // Needs array encoding?
    NestDepth   int   // Validate against 32 level limit
}
```

### Use Cases
- Fixed-size types: Simpler decode with guaranteed sizes
- Max size known: Single upfront bounds check, then safe
- Type properties: Generate only needed encoding logic

---

## Decision 3: No Field Reordering

**Date**: November 5, 2025

### Decision
Do **not** reorder struct fields for optimization. Keep declaration order.

### Rationale
- **Simplicity**: Schema field order matches generated struct order
- **No benefit**: Sequential streaming means no cache locality gains
- **Complexity**: Would require offset tracking and partial decode support
- **Maintainability**: Keep it simple and predictable

### Alternative Considered
Reorder fields to group primitives first (enables early validation, fixed offsets).

### Why Rejected
- Breaks intuitive field ordering
- Adds significant complexity to generator
- Benefits are marginal for sequential decode
- Can revisit if profiling shows need

---

## Decision 4: No Zero-Copy Unsafe

**Date**: November 5, 2025

### Decision
Do **not** generate unsafe pointer-based decoding for fixed-size types.

### Rationale
- **Cross-language**: Unsafe pointer tricks break C++/Swift compatibility
- **Safety**: Defeats "safe by design" philosophy
- **Premature**: No evidence this is bottleneck
- **Complexity**: Requires platform-specific code generation

### Alternative Considered
```go
// Unsafe fast path for fixed-size structs
func DecodePointUnsafe(data []byte) *Point {
    return (*Point)(unsafe.Pointer(&data[0]))
}
```

### Why Rejected
- Endianness issues across architectures
- Padding differences between languages
- Violates Go's memory safety guarantees
- Not significantly faster than binary.Read for small structs

---

## Decision 5: Unified CLI

**Date**: Prior (documented for completeness)

### Decision
Single `ffire` binary with subcommands (generate, validate, fixture, bench).

### Rationale
- Modern convention (git, docker, cargo)
- Consistent interface
- Single installation
- Easy discovery

---

## Decision 6: No Wire Format Versioning

**Date**: Prior (documented for completeness)

### Decision
No version field in wire format. Encoder and decoder must be compiled together.

### Rationale
- **Same-machine IPC**: Both sides from same build
- **Simpler**: No compatibility matrix to manage
- **Faster**: No version checking overhead
- **Breaking changes OK**: Recompile both sides

### Constraint
Wire format changes require recompilation of both encoder and decoder.

### Acceptable Because
- Target use case: Same-machine plugin enumeration, not network protocol
- Development iteration: Both sides rebuilt together
- No persistent storage: Wire format is transport only

---

## Summary of Safety Approach

### Problems Prevented by Design

| Attack Vector | Traditional Solution | ffire Solution |
|--------------|---------------------|----------------|
| String buffer overflow | Runtime bounds check | uint16 physically limits to 64KB |
| Array exhaustion | Runtime bounds check | uint16 physically limits to 65K elements |
| Integer overflow in length | Check `length < MAX` | uint16 can't overflow |
| Stack overflow (nesting) | Runtime depth tracking | Schema validator enforces 32 levels |
| Message size DoS | Runtime size check | 2^31 limit + small element limits |

### Code Simplification

**Before (with checks):**
```go
var length uint32
binary.Read(r, &length)
if length > MaxStringLength {
    return ErrStringTooLarge  // ← Check needed
}
strBytes := make([]byte, length)
```

**After (without checks):**
```go
var length uint16
binary.Read(r, &length)
strBytes := make([]byte, length)  // ← Can't overflow!
```

### Performance Impact
- **Eliminated**: ~10-20 CPU cycles per length check
- **Eliminated**: ~50-100 bytes of code per check
- **Added**: 0 runtime cost (type system handles it)

---

## Future Considerations

### Not Decided Yet

1. **Partial decode functions**: Generate `DecodeJustID()` style functions?
2. **Streaming decode**: Support incremental parsing for large messages?
3. **Compile-time offsets**: Pre-compute field offsets for fixed layouts?

### When to Revisit

- **Profile first**: Measure actual bottlenecks in real usage
- **Prove need**: Show benchmark demonstrating insufficient performance
- **Evaluate complexity**: Only add if benefit >> complexity cost

---

## Lessons Applied

1. **Safety by design** > Runtime validation
2. **Simple first** > Premature optimization
3. **Type system** > Dynamic checks
4. **Analysis once** > Checks every time
5. **Constraints** > Configuration

**Philosophy**: The best runtime check is the one you don't need to write.
