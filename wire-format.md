# ffire Wire Format Specification
## ffire - FFI Encoding

## Design Principles
- **Natural sizes, no padding** - Optimized for metadata/orchestration (strings, small structs)
- **Little-endian** - Native for x64/ARM64
- **Fixed-size lengths** - uint16 for safety by design (no overflow checks needed)
- **No versioning** - Encoder/decoder compiled together for same-machine IPC
- **Target languages**: C, C++, Go, Swift, Objective-C++
- **Safety by design** - Physical limits prevent runtime security checks

## Primitive Types
- `bool`: 1 byte (0x00 = false, 0x01 = true)
- `int8`: 1 byte
- `int16`: 2 bytes
- `int32`: 4 bytes
- `int64`: 8 bytes
- `float32`: 4 bytes (IEEE 754)
- `float64`: 8 bytes (IEEE 754)

## Composite Types

### String
```
[uint16_le: byte_length][utf8_bytes...]
```
- Length is number of UTF-8 bytes (not characters)
- No null terminator
- Empty string: `00 00`
- Max length: 65,535 bytes (64KB - 1)
- **Safety**: uint16 physically prevents overflow attacks

### Array
```
[uint16_le: element_count][element_0][element_1]...[element_n]
```
- Count is number of elements
- Elements encoded sequentially (primitives or structs)
- Empty array: `00 00`
- Max count: 65,535 elements
- **Safety**: uint16 physically prevents memory exhaustion attacks

### Struct
```
[field_0][field_1]...[field_n]
```
- Fields in declaration order, no padding between fields

## Message Structure
```
[root_value]
```
- No message-level size prefix (buffer length is known from IPC mechanism)
- `root_value`: One of: primitive, string, array, struct

## Constraints
- **Max nesting depth**: 32 levels (prevents stack overflow)
- **Max message size**: 2^31 bytes (2GB - allows safe int casting)
- **Max string length**: 65,535 bytes (uint16 - physically impossible to overflow)
- **Max array length**: 65,535 elements (uint16 - prevents memory exhaustion)

**Rationale**: These limits are enforced at the wire format level, eliminating the need for runtime bounds checking. A malicious or corrupt message cannot cause buffer overflows or memory exhaustion because the type system prevents it.

## Example
```go
type Device struct {
    Name string    // "Speaker"
    Channels int32 // 2
}
devices := []Device{{Name: "Speaker", Channels: 2}}
```

**Wire bytes (hex)**:
```
01 00                    # array length = 1 (uint16 LE)
07 00                    # string length = 7 (uint16 LE)
53 70 65 61 6B 65 72     # "Speaker" (UTF-8)
02 00 00 00              # channels = 2 (int32 LE)
```
Total: 14 bytes (4 bytes smaller than uint32 version)