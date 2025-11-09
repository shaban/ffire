# ffire Native Implementation Guide

**Target Languages**: Rust, C#, Java  
**Status**: Design document for Phase 2-3 implementation  
**Last Updated**: November 9, 2025

---

## Table of Contents

1. [Overview](#overview)
2. [Wire Format Specification](#wire-format-specification)
3. [Type System Mapping](#type-system-mapping)
4. [Optimization Patterns by Use Case](#optimization-patterns-by-use-case)
5. [Language-Specific Implementation Guidance](#language-specific-implementation-guidance)
6. [Code Generation Templates](#code-generation-templates)
7. [Performance Targets](#performance-targets)
8. [Testing & Validation](#testing--validation)

---

## Overview

### Purpose
This document captures the **proven optimization patterns** from Go and C++ implementations and provides **concrete guidance** for implementing Rust, C#, and Java native (Tier A) generators.

### Architecture Principles

1. **Message Suffix Convention**: All root message types use `{Name}Message` suffix to avoid keyword collisions
2. **Zero-Copy where possible**: Primitive arrays use memory reinterpretation for bulk encoding/decoding
3. **Little-endian wire format**: All multi-byte integers use little-endian byte order
4. **Length-prefixed strings**: UTF-8 strings with 2-byte (uint16) length prefix
5. **Direct slice indexing**: Decoding uses direct array indexing, not Reader abstractions

### Performance Philosophy

- **Tier A targets**: 50-200ns per operation (competitive with C++)
- **Key optimization**: Bulk operations for primitive arrays (10x faster than element-by-element)
- **Memory efficiency**: Pre-allocate buffers when sizes known, zero-copy where safe

---

## Wire Format Specification

### Primitive Types

All integers use **little-endian** byte order.

```
bool:    1 byte  (0x00 = false, 0x01 = true)
int8:    1 byte  (signed, two's complement)
int16:   2 bytes (little-endian)
int32:   4 bytes (little-endian)
int64:   8 bytes (little-endian)
float32: 4 bytes (IEEE 754, reinterpreted as uint32, little-endian)
float64: 8 bytes (IEEE 754, reinterpreted as uint64, little-endian)
string:  2-byte length + UTF-8 bytes (length is uint16, little-endian)
```

### Example: Encoding int32 value 0x12345678

```
Wire bytes: 78 56 34 12  (little-endian)
            ^  ^  ^  ^
            |  |  |  +-- Most significant byte
            |  |  +----- 
            |  +-------- 
            +----------- Least significant byte
```

### Optional Types

Optional fields use a **presence flag** (1 byte) followed by value if present:

```
Optional<T>:
  - 0x00 (absent) → no additional bytes
  - 0x01 (present) → followed by encoded T
```

### Array Types

Arrays use **2-byte length prefix** (uint16) followed by elements:

```
[]T:
  - 2 bytes: length (uint16, little-endian)
  - N * sizeof(T) bytes: elements
```

### Struct Types

Structs encode fields **in declaration order** with no padding or alignment:

```
struct Config {
    name: string,
    value: int32,
}

Wire format:
  [string length (2 bytes)][string data][int32 (4 bytes)]
```

### Nested Structs

Nested structs are encoded inline:

```
struct Inner {
    x: int16,
}

struct Outer {
    inner: Inner,
    y: int32,
}

Wire format:
  [x (2 bytes)][y (4 bytes)]
```

---

## Type System Mapping

### Primitive Type Mapping

| ffire Type | Go       | C++       | Rust      | C#        | Java      | Wire Size |
|------------|----------|-----------|-----------|-----------|-----------|-----------|
| `bool`     | `bool`   | `bool`    | `bool`    | `bool`    | `boolean` | 1 byte    |
| `int8`     | `int8`   | `int8_t`  | `i8`      | `sbyte`   | `byte`    | 1 byte    |
| `int16`    | `int16`  | `int16_t` | `i16`     | `short`   | `short`   | 2 bytes   |
| `int32`    | `int32`  | `int32_t` | `i32`     | `int`     | `int`     | 4 bytes   |
| `int64`    | `int64`  | `int64_t` | `i64`     | `long`    | `long`    | 8 bytes   |
| `float32`  | `float32`| `float`   | `f32`     | `float`   | `float`   | 4 bytes   |
| `float64`  | `float64`| `double`  | `f64`     | `double`  | `double`  | 8 bytes   |
| `string`   | `string` | `std::string` | `String` | `string` | `String`  | 2 + len   |

### Optional Type Mapping

| ffire Type | Go           | C++                      | Rust          | C#                    | Java          |
|------------|--------------|--------------------------|---------------|-----------------------|---------------|
| `?bool`    | `*bool`      | `std::optional<bool>`    | `Option<bool>`| `bool?`               | `Boolean`     |
| `?int32`   | `*int32`     | `std::optional<int32_t>` | `Option<i32>` | `int?`                | `Integer`     |
| `?string`  | `*string`    | `std::optional<std::string>` | `Option<String>` | `string` (null ok) | `String`      |
| `?Config`  | `*Config`    | `std::optional<Config>`  | `Option<Config>` | `Config` (null ok) | `Config`      |

### Array Type Mapping

| ffire Type | Go           | C++                   | Rust           | C#              | Java              |
|------------|--------------|-----------------------|----------------|-----------------|-------------------|
| `[]int32`  | `[]int32`    | `std::vector<int32_t>`| `Vec<i32>`     | `List<int>`     | `List<Integer>`   |
| `[]string` | `[]string`   | `std::vector<std::string>` | `Vec<String>` | `List<string>` | `List<String>`   |
| `[]Config` | `[]Config`   | `std::vector<Config>` | `Vec<Config>`  | `List<Config>`  | `List<Config>`    |

---

## Optimization Patterns by Use Case

### Use Case 1: Primitive Arrays (Zero-Copy)

**The Opportunity**: Primitive arrays (int8, int16, int32, int64, float32, float64) can be bulk-copied as raw bytes.

#### Go Implementation (Reference)

```go
// Encoding []int32
func encodeInt32Array(buf *bytes.Buffer, arr []int32) {
    // Write length
    length := uint16(len(arr))
    buf.WriteByte(byte(length))
    buf.WriteByte(byte(length >> 8))
    
    // Bulk write: reinterpret []int32 as []byte (zero-copy)
    if len(arr) > 0 {
        buf.Write(unsafe.Slice((*byte)(unsafe.Pointer(&arr[0])), len(arr)*4))
    }
}

// Decoding []int32
func decodeInt32Array(data []byte, pos *int) []int32 {
    // Read length
    length := uint16(data[*pos]) | uint16(data[*pos+1])<<8
    *pos += 2
    
    // Bulk read: reinterpret []byte as []int32 (zero-copy)
    result := make([]int32, length)
    if length > 0 {
        copy(result, unsafe.Slice((*int32)(unsafe.Pointer(&data[*pos])), int(length)))
        *pos += int(length) * 4
    }
    return result
}
```

**Key Insight**: This is 10-20x faster than element-by-element encoding because:
- Single memory operation instead of N WriteByte calls
- CPU can use SIMD/burst transfers
- Better cache locality

#### C++ Implementation (Reference)

```cpp
// Encoding std::vector<int32_t>
void encode_int32_array(Encoder& enc, const std::vector<int32_t>& arr) {
    // Write length
    uint16_t len = static_cast<uint16_t>(arr.size());
    enc.buffer.push_back(static_cast<uint8_t>(len));
    enc.buffer.push_back(static_cast<uint8_t>(len >> 8));
    
    // Bulk write: reinterpret vector data as bytes
    if (!arr.empty()) {
        const uint8_t* ptr = reinterpret_cast<const uint8_t*>(arr.data());
        enc.buffer.insert(enc.buffer.end(), ptr, ptr + arr.size() * 4);
    }
}

// Decoding std::vector<int32_t>
std::vector<int32_t> decode_int32_array(Decoder& dec) {
    // Read length
    uint16_t len = dec.read_array_length();
    
    // Bulk read: copy bytes and reinterpret
    std::vector<int32_t> result(len);
    if (len > 0) {
        dec.check_remaining(len * 4);
        std::memcpy(result.data(), dec.data + dec.pos, len * 4);
        dec.pos += len * 4;
    }
    return result;
}
```

#### Rust Implementation (Target)

```rust
// Encoding Vec<i32>
fn encode_i32_array(buf: &mut Vec<u8>, arr: &[i32]) {
    // Write length
    let len = arr.len() as u16;
    buf.push(len as u8);
    buf.push((len >> 8) as u8);
    
    // Bulk write: reinterpret &[i32] as &[u8]
    if !arr.is_empty() {
        let bytes = unsafe {
            std::slice::from_raw_parts(
                arr.as_ptr() as *const u8,
                arr.len() * 4
            )
        };
        buf.extend_from_slice(bytes);
    }
}

// Decoding Vec<i32>
fn decode_i32_array(data: &[u8], pos: &mut usize) -> Vec<i32> {
    // Read length
    let len = (data[*pos] as u16) | ((data[*pos + 1] as u16) << 8);
    *pos += 2;
    
    // Bulk read: copy bytes and reinterpret
    let mut result = vec![0i32; len as usize];
    if len > 0 {
        let bytes = unsafe {
            std::slice::from_raw_parts_mut(
                result.as_mut_ptr() as *mut u8,
                len as usize * 4
            )
        };
        bytes.copy_from_slice(&data[*pos..*pos + bytes.len()]);
        *pos += bytes.len();
    }
    result
}
```

#### C# Implementation (Target)

```csharp
// Encoding List<int>
void EncodeInt32Array(List<byte> buf, List<int> arr) {
    // Write length
    ushort len = (ushort)arr.Count;
    buf.Add((byte)len);
    buf.Add((byte)(len >> 8));
    
    // Bulk write using Span<T> (zero-copy)
    if (arr.Count > 0) {
        Span<int> span = CollectionsMarshal.AsSpan(arr);
        Span<byte> bytes = MemoryMarshal.AsBytes(span);
        buf.AddRange(bytes.ToArray()); // Or use ArrayPool for efficiency
    }
}

// Decoding List<int>
List<int> DecodeInt32Array(ReadOnlySpan<byte> data, ref int pos) {
    // Read length
    ushort len = (ushort)(data[pos] | (data[pos + 1] << 8));
    pos += 2;
    
    // Bulk read using Span<T>
    List<int> result = new List<int>(len);
    if (len > 0) {
        ReadOnlySpan<byte> bytes = data.Slice(pos, len * 4);
        Span<int> ints = MemoryMarshal.Cast<byte, int>(bytes);
        result.AddRange(ints.ToArray());
        pos += len * 4;
    }
    return result;
}
```

#### Java Implementation (Target)

```java
// Encoding List<Integer>
void encodeInt32Array(ByteBuffer buf, List<Integer> arr) {
    // Write length
    short len = (short) arr.size();
    buf.put((byte) len);
    buf.put((byte) (len >>> 8));
    
    // Bulk write using IntBuffer
    if (arr.size() > 0) {
        int[] array = arr.stream().mapToInt(Integer::intValue).toArray();
        buf.asIntBuffer().put(array);
        buf.position(buf.position() + array.length * 4);
    }
}

// Decoding List<Integer>
List<Integer> decodeInt32Array(ByteBuffer buf) {
    // Read length
    short len = (short) ((buf.get() & 0xFF) | ((buf.get() & 0xFF) << 8));
    
    // Bulk read using IntBuffer
    List<Integer> result = new ArrayList<>(len);
    if (len > 0) {
        int[] array = new int[len];
        buf.asIntBuffer().get(array);
        buf.position(buf.position() + len * 4);
        for (int v : array) result.add(v);
    }
    return result;
}
```

**Performance Impact**: 10-20x faster than element-by-element for large arrays (n > 100).

---

### Use Case 2: String Encoding

**The Pattern**: Strings are length-prefixed UTF-8 with no null terminator.

#### Go Implementation (Reference)

```go
func encodeString(buf *bytes.Buffer, s string) {
    length := uint16(len(s))
    buf.WriteByte(byte(length))
    buf.WriteByte(byte(length >> 8))
    buf.WriteString(s) // Already UTF-8 in Go
}

func decodeString(data []byte, pos *int) string {
    length := uint16(data[*pos]) | uint16(data[*pos+1])<<8
    *pos += 2
    s := string(data[*pos : *pos+int(length)]) // Safe copy
    *pos += int(length)
    return s
}
```

**Critical**: Use `string(data[start:end])` to create **independent copy**, avoiding lifetime issues.

#### Rust Implementation (Target)

```rust
fn encode_string(buf: &mut Vec<u8>, s: &str) {
    let len = s.len() as u16;
    buf.push(len as u8);
    buf.push((len >> 8) as u8);
    buf.extend_from_slice(s.as_bytes());
}

fn decode_string(data: &[u8], pos: &mut usize) -> String {
    let len = (data[*pos] as u16) | ((data[*pos + 1] as u16) << 8);
    *pos += 2;
    let s = String::from_utf8_lossy(&data[*pos..*pos + len as usize]).to_string();
    *pos += len as usize;
    s
}
```

#### C# Implementation (Target)

```csharp
void EncodeString(List<byte> buf, string s) {
    byte[] bytes = Encoding.UTF8.GetBytes(s);
    ushort len = (ushort)bytes.Length;
    buf.Add((byte)len);
    buf.Add((byte)(len >> 8));
    buf.AddRange(bytes);
}

string DecodeString(ReadOnlySpan<byte> data, ref int pos) {
    ushort len = (ushort)(data[pos] | (data[pos + 1] << 8));
    pos += 2;
    string s = Encoding.UTF8.GetString(data.Slice(pos, len));
    pos += len;
    return s;
}
```

#### Java Implementation (Target)

```java
void encodeString(ByteBuffer buf, String s) {
    byte[] bytes = s.getBytes(StandardCharsets.UTF_8);
    short len = (short) bytes.length;
    buf.put((byte) len);
    buf.put((byte) (len >>> 8));
    buf.put(bytes);
}

String decodeString(ByteBuffer buf) {
    short len = (short) ((buf.get() & 0xFF) | ((buf.get() & 0xFF) << 8));
    byte[] bytes = new byte[len];
    buf.get(bytes);
    return new String(bytes, StandardCharsets.UTF_8);
}
```

---

### Use Case 3: String Arrays

**Optimization**: Pre-calculate total size and pre-allocate buffer.

#### Go Implementation (Reference)

```go
func encodeStringArray(buf *bytes.Buffer, arr []string) {
    length := uint16(len(arr))
    buf.WriteByte(byte(length))
    buf.WriteByte(byte(length >> 8))
    
    // Optimization: calculate total size and Grow() once
    totalSize := 2 * len(arr) // length prefixes
    for _, s := range arr {
        totalSize += len(s)
    }
    buf.Grow(totalSize)
    
    // Now encode each string
    for _, s := range arr {
        l := uint16(len(s))
        buf.WriteByte(byte(l))
        buf.WriteByte(byte(l >> 8))
        buf.WriteString(s)
    }
}
```

**Key**: Single `Grow()` call prevents multiple buffer reallocations (5-10x faster for large arrays).

---

### Use Case 4: Struct Encoding

**Pattern**: Encode fields in declaration order, inline nested structs.

#### Go Implementation (Reference)

```go
type ConfigMessage struct {
    Name  string
    Value int32
}

func EncodeConfigMessage(msg ConfigMessage) []byte {
    buf := &bytes.Buffer{}
    
    // Encode Name (string)
    length := uint16(len(msg.Name))
    buf.WriteByte(byte(length))
    buf.WriteByte(byte(length >> 8))
    buf.WriteString(msg.Name)
    
    // Encode Value (int32)
    v := uint32(msg.Value)
    buf.WriteByte(byte(v))
    buf.WriteByte(byte(v >> 8))
    buf.WriteByte(byte(v >> 16))
    buf.WriteByte(byte(v >> 24))
    
    return buf.Bytes()
}
```

#### Rust Implementation (Target)

```rust
struct ConfigMessage {
    name: String,
    value: i32,
}

fn encode_config_message(msg: &ConfigMessage) -> Vec<u8> {
    let mut buf = Vec::new();
    
    // Encode name
    encode_string(&mut buf, &msg.name);
    
    // Encode value
    let v = msg.value as u32;
    buf.push(v as u8);
    buf.push((v >> 8) as u8);
    buf.push((v >> 16) as u8);
    buf.push((v >> 24) as u8);
    
    buf
}
```

---

### Use Case 5: Optional Fields

**Pattern**: 1-byte presence flag (0x00 = absent, 0x01 = present) followed by value.

#### Go Implementation (Reference)

```go
func encodeOptionalInt32(buf *bytes.Buffer, value *int32) {
    if value == nil {
        buf.WriteByte(0x00)
    } else {
        buf.WriteByte(0x01)
        v := uint32(*value)
        buf.WriteByte(byte(v))
        buf.WriteByte(byte(v >> 8))
        buf.WriteByte(byte(v >> 16))
        buf.WriteByte(byte(v >> 24))
    }
}
```

#### Rust Implementation (Target)

```rust
fn encode_optional_i32(buf: &mut Vec<u8>, value: &Option<i32>) {
    match value {
        None => buf.push(0x00),
        Some(v) => {
            buf.push(0x01);
            let u = *v as u32;
            buf.push(u as u8);
            buf.push((u >> 8) as u8);
            buf.push((u >> 16) as u8);
            buf.push((u >> 24) as u8);
        }
    }
}
```

---

### Use Case 6: Nested Structs

**Pattern**: Encode inline, recursively.

```go
type Inner struct {
    X int16
}

type OuterMessage struct {
    Inner Inner
    Y     int32
}

func EncodeOuterMessage(msg OuterMessage) []byte {
    buf := &bytes.Buffer{}
    
    // Encode Inner.X (int16)
    v := uint16(msg.Inner.X)
    buf.WriteByte(byte(v))
    buf.WriteByte(byte(v >> 8))
    
    // Encode Y (int32)
    v32 := uint32(msg.Y)
    buf.WriteByte(byte(v32))
    buf.WriteByte(byte(v32 >> 8))
    buf.WriteByte(byte(v32 >> 16))
    buf.WriteByte(byte(v32 >> 24))
    
    return buf.Bytes()
}
```

---

## Language-Specific Implementation Guidance

### Rust Guidelines

1. **Error Handling**: Use `Result<T, Error>` for all decode functions
2. **Zero-Copy**: Leverage `slice::from_raw_parts` for primitive arrays
3. **String Safety**: Use `String::from_utf8_lossy` to handle invalid UTF-8
4. **Ownership**: Encode functions take `&T`, decode functions return owned `T`
5. **Traits**: Consider implementing `Encode` and `Decode` traits for user types

**Example Trait**:
```rust
trait FfireEncode {
    fn encode(&self, buf: &mut Vec<u8>);
}

trait FfireDecode: Sized {
    fn decode(data: &[u8], pos: &mut usize) -> Result<Self, DecodeError>;
}
```

### C# Guidelines

1. **Use Span<T>**: `ReadOnlySpan<byte>` for decode, `Span<T>` for zero-copy
2. **MemoryMarshal**: Use `MemoryMarshal.Cast<byte, T>()` for bulk operations
3. **Nullable Reference Types**: Enable for proper null safety
4. **ArrayPool**: Use `ArrayPool<byte>.Shared` for temporary buffers
5. **Unsafe Code**: Allow in .csproj for zero-copy optimizations

**Example .csproj setting**:
```xml
<PropertyGroup>
    <AllowUnsafeBlocks>true</AllowUnsafeBlocks>
</PropertyGroup>
```

### Java Guidelines

1. **ByteBuffer**: Use `ByteBuffer` for all encoding/decoding
2. **Direct Buffers**: `ByteBuffer.allocateDirect()` for off-heap allocation
3. **Primitive Streams**: Use `IntBuffer`, `LongBuffer` for bulk operations
4. **Generics**: Avoid boxing by providing primitive-specific methods
5. **UTF-8**: Always use `StandardCharsets.UTF_8`

**Performance Tip**: Provide both `encode(Message)` and `encodeToBuffer(Message, ByteBuffer)` variants.

---

## Code Generation Templates

### Message Suffix Architecture

**All root message types** use `{Name}Message` suffix:

```
Schema Definition:      Generated Types:
-------------------     ----------------
message Config { ... }  → struct ConfigMessage { ... }
message Stats { ... }   → struct StatsMessage { ... }
message Empty { }       → struct EmptyMessage { }
```

**Helper/embedded types** do NOT use Message suffix:

```
type Inner { ... }      → struct Inner { ... }
```

### Generator Output Structure

Each generator should produce:

1. **Type Definitions**: Structs with Message suffix for root types
2. **Public Encode Functions**: `encode_{name}_message(value) -> bytes`
3. **Public Decode Functions**: `decode_{name}_message(bytes) -> value`
4. **Private Helpers**: Encoder/Decoder classes or functions

### Function Naming Convention

| Language | Encode Function         | Decode Function         |
|----------|-------------------------|-------------------------|
| Go       | `EncodeConfigMessage`   | `DecodeConfigMessage`   |
| C++      | `encode_config_message` | `decode_config_message` |
| Rust     | `encode_config_message` | `decode_config_message` |
| C#       | `EncodeConfigMessage`   | `DecodeConfigMessage`   |
| Java     | `encodeConfigMessage`   | `decodeConfigMessage`   |

---

## Performance Targets

### Tier A Performance Goals

| Operation                  | Target Latency | Notes                          |
|----------------------------|----------------|--------------------------------|
| Encode primitive (int32)   | < 5ns          | Inline, register operations    |
| Decode primitive (int32)   | < 10ns         | Array indexing + shifts        |
| Encode string (10 chars)   | < 50ns         | Includes length prefix         |
| Decode string (10 chars)   | < 80ns         | Includes UTF-8 validation      |
| Encode int32 array (100)   | < 100ns        | Bulk copy optimization         |
| Decode int32 array (100)   | < 200ns        | Bulk copy + allocation         |
| Encode struct (3 fields)   | < 50ns         | Sum of field costs             |
| Decode struct (3 fields)   | < 100ns        | Sum of field costs             |

### Comparison to Tier B (FFI)

Tier B languages (Python, JavaScript, Swift, Dart) incur ~100-500ns FFI overhead per call. Native implementations eliminate this entirely.

**Expected Speedup**: 5-10x faster than Tier B for typical workloads.

---

## Testing & Validation

### Cross-Language Test Strategy

**Goal**: Ensure all languages produce identical wire format.

#### Test Vector Generation

1. Generate test data in Go (reference implementation)
2. Encode to binary, save as `.bin` file
3. All other languages must:
   - Decode the `.bin` file successfully
   - Produce identical bytes when encoding same data

#### Test Schemas

Use all 10 benchmark schemas:
1. Empty (empty struct)
2. Primitive (single int32)
3. Struct (multiple fields)
4. Optional (optional fields)
5. Array (int32 array)
6. StringArray (string array)
7. Nested (nested structs)
8. PrimitiveArray (primitive array)
9. BoolArray (bool array)
10. ComplexTest (combination)

#### Validation Checklist

For each new language implementation:

- [ ] All 10 schemas generate without errors
- [ ] Generated code compiles without warnings
- [ ] Encodes test data matching Go reference bytes
- [ ] Decodes Go-encoded data correctly
- [ ] Round-trip (encode → decode) preserves data
- [ ] Performance meets Tier A targets
- [ ] Handles edge cases (empty arrays, null strings, max length)

### Binary Comparison Tool

```bash
# Generate reference binary from Go
go run main.go encode test.json > reference.bin

# Validate Rust implementation
rust run --release encode test.json > rust.bin
diff reference.bin rust.bin  # Should be identical

# Validate C# implementation
dotnet run encode test.json > csharp.bin
diff reference.bin csharp.bin  # Should be identical
```

---

## Implementation Checklist

### Phase 2: Rust Native Implementation

- [ ] Create `pkg/generator/generator_rust.go`
- [ ] Implement type mapping (primitives, Option<T>, Vec<T>)
- [ ] Implement struct generation with Message suffix
- [ ] Implement encode functions (inline primitives, bulk arrays)
- [ ] Implement decode functions (direct slice indexing)
- [ ] Generate tests for all 10 schemas
- [ ] Validate cross-language compatibility
- [ ] Benchmark performance (target: Tier A)

### Phase 3: C# Native Implementation

- [ ] Create `pkg/generator/generator_csharp_native.go`
- [ ] Implement type mapping (primitives, Nullable<T>, List<T>)
- [ ] Implement struct generation with Message suffix
- [ ] Implement encode using Span<byte> for zero-copy
- [ ] Implement decode using ReadOnlySpan<byte>
- [ ] Generate .csproj with unsafe blocks enabled
- [ ] Validate cross-language compatibility
- [ ] Benchmark performance (target: Tier A)

### Phase 3: Java Native Implementation

- [ ] Create `pkg/generator/generator_java.go`
- [ ] Implement type mapping (primitives, boxed types, List<T>)
- [ ] Implement struct generation with Message suffix
- [ ] Implement encode using ByteBuffer
- [ ] Implement decode using ByteBuffer with bulk operations
- [ ] Generate Maven/Gradle build files
- [ ] Validate cross-language compatibility
- [ ] Benchmark performance (target: Tier A)

---

## Appendix: Wire Format Examples

### Example 1: Simple Struct

```
struct ConfigMessage {
    name: string,
    value: int32,
}

Data: { name: "test", value: 42 }

Wire bytes (hex):
04 00          // string length = 4 (little-endian uint16)
74 65 73 74    // "test" UTF-8 bytes
2A 00 00 00    // value = 42 (little-endian int32)
```

### Example 2: Int32 Array

```
Data: [1, 2, 3]

Wire bytes (hex):
03 00          // array length = 3 (little-endian uint16)
01 00 00 00    // element 0 = 1
02 00 00 00    // element 1 = 2
03 00 00 00    // element 2 = 3
```

### Example 3: Optional Int32

```
Data: Some(42)

Wire bytes (hex):
01             // present flag = 0x01
2A 00 00 00    // value = 42

Data: None

Wire bytes (hex):
00             // absent flag = 0x00
```

---

## Conclusion

This guide provides everything needed to implement Rust, C#, and Java generators that:

1. ✅ Produce identical wire format to Go/C++
2. ✅ Achieve Tier A performance (50-200ns operations)
3. ✅ Use zero-copy optimizations where safe
4. ✅ Follow language-specific best practices
5. ✅ Pass cross-language compatibility tests

**Next Steps**:
1. Implement Rust generator (Phase 2)
2. Validate with all 10 benchmark schemas
3. Repeat for C# and Java (Phase 3)
4. Update roadmap with completion status

---

**Document Version**: 1.0  
**Authors**: Based on Go and C++ reference implementations  
**Target Completion**: Phase 2-3 of release roadmap
