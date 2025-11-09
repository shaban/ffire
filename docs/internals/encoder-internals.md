# ffire Encoder Internals

## Overview

The ffire encoder generates optimized Go code that serializes data structures to a deterministic binary wire format. This document explains how the encoding process works, from schema analysis to code generation.

---

## Wire Format Design

### Basic Principles

- **Little-endian byte order** - All multi-byte integers use little-endian
- **Length-prefixed** - Strings and arrays have uint16 length prefix (max 65,535 elements)
- **Optional fields** - Marked with presence byte (0x00 = absent, 0x01 = present)
- **No padding** - Fields are packed tightly with no alignment padding
- **Deterministic** - Same input always produces identical output

### Primitive Type Encoding

| Type | Size | Encoding |
|------|------|----------|
| `bool` | 1 byte | `0x00` = false, `0x01` = true |
| `int8` | 1 byte | Direct byte value |
| `int16` | 2 bytes | Little-endian: `[byte0, byte1]` |
| `int32` | 4 bytes | Little-endian: `[byte0, byte1, byte2, byte3]` |
| `int64` | 8 bytes | Little-endian: 8 bytes |
| `float32` | 4 bytes | IEEE 754 bits as little-endian uint32 |
| `float64` | 8 bytes | IEEE 754 bits as little-endian uint64 |
| `string` | 2 + N bytes | `[length_uint16][utf8_bytes]` |

### Complex Type Encoding

**Optional Fields:**
```
[1 byte presence][value if present]
```
- Presence byte: `0x00` = null/nil, `0x01` = value follows
- If present, encode the value normally

**Arrays:**
```
[2 byte length][element0][element1]...[elementN]
```
- Length as uint16 (max 65,535 elements)
- Each element encoded according to its type

**Structs:**
```
[field0][field1]...[fieldN]
```
- No struct boundary markers (fields encoded sequentially)
- Non-optional structs have NO presence byte
- Optional structs: `[1 byte presence][fields if present]`

---

## Code Generation Process

### Phase 1: Schema Analysis

The generator analyzes the schema to determine:

1. **Import requirements:**
   - `bytes` - Always imported (for `bytes.Buffer`)
   - `math` - Only if schema contains float32/float64
   - `unsafe` - Only if schema has primitive arrays (for zero-copy)

2. **Type characteristics:**
   - Which types are optional
   - Which arrays can use bulk encoding
   - Nesting depth and structure

### Phase 2: Type Definitions

Generated Go structs preserve the schema structure:

```go
// From schema:
type Config struct {
    Host       string  `json:"host"`
    Port       int32   `json:"port"`
    EnableSSL  bool    `json:"enableSSL"`
}

// Generated (struct tags preserved):
type Config struct {
    Host       string  `json:"host"`
    Port       int32   `json:"port"`
    EnableSSL  bool    `json:"enableSSL"`
}
```

### Phase 3: Encode Function Generation

For each message type, generate a public encode function:

```go
func EncodeConfigMessage(v Config) []byte {
    buf := &bytes.Buffer{}
    // Encoding logic here
    return buf.Bytes()
}
```

---

## Encoding Strategies

### Strategy 1: Primitive Fields (Direct Byte Writes)

**int32 example:**
```go
// Generated code:
{ 
    v := uint32(v.Port)
    buf.WriteByte(byte(v))
    buf.WriteByte(byte(v>>8))
    buf.WriteByte(byte(v>>16))
    buf.WriteByte(byte(v>>24))
}
```

**Why this approach:**
- ✅ No `binary.Write()` overhead (no reflection)
- ✅ Explicit byte ordering (no endianness checks)
- ✅ Compiler can inline these operations
- ✅ Each `WriteByte()` is ~2-3 instructions

**Alternative considered but rejected:**
```go
// Slower: Create temp array + Write()
var b [4]byte
b[0] = byte(v)
b[1] = byte(v>>8)
b[2] = byte(v>>16)
b[3] = byte(v>>24)
buf.Write(b[:])  // Extra function call overhead
```

### Strategy 2: String Encoding

**Generated code:**
```go
// Write length prefix (uint16)
{ 
    l := uint16(len(v.Host))
    buf.WriteByte(byte(l))
    buf.WriteByte(byte(l>>8))
}
// Write string data
buf.WriteString(v.Host)  // Efficient: no allocation
```

**Why length-prefixed:**
- ✅ Know string size before reading (for decoding)
- ✅ No escaping needed (unlike null-terminated)
- ✅ Max 65,535 bytes (reasonable for most use cases)

### Strategy 3: Array Encoding (Zero-Copy for Primitives)

**For primitive arrays (int16, int32, int64, float32, float64):**

```go
// Write array length
{ 
    l := uint16(len(array))
    buf.WriteByte(byte(l))
    buf.WriteByte(byte(l>>8))
}

// Zero-copy bulk write using unsafe
if len(array) > 0 {
    buf.Write(unsafe.Slice((*byte)(unsafe.Pointer(&array[0])), len(array)*4))
}
```

**How it works:**
1. `&array[0]` - Get pointer to first element
2. `unsafe.Pointer()` - Convert to generic pointer
3. `(*byte)` - Reinterpret as byte pointer
4. `unsafe.Slice()` - Create `[]byte` view of the memory
5. `buf.Write()` - Copy bytes directly (memcpy)

**Performance:**
- ✅ **25x faster** than element-by-element encoding
- ✅ No loops, no type conversions
- ✅ Single memcpy operation
- ✅ Works because wire format is little-endian (same as x86/ARM64)

**Safety considerations:**
- ✅ Read-only operation (no dangling pointers)
- ✅ Source data remains valid during copy
- ✅ Only used for non-optional primitive arrays
- ⚠️ Requires little-endian architecture (documented limitation)

**For non-primitive arrays (strings, structs, optional elements):**

```go
// Fallback to element-by-element encoding
for _, elem := range array {
    // Encode each element normally
}
```

### Strategy 4: Struct Encoding (Field Flattening)

**Non-optional structs are transparent:**

```go
type Metadata struct {
    Version int32
    Author  string
}

type Config struct {
    Port     int32
    Metadata Metadata  // Non-optional nested struct
}

// Generated encoding (no struct boundary):
func EncodeConfigMessage(v Config) []byte {
    buf := &bytes.Buffer{}
    
    // Port
    { v := uint32(v.Port); buf.WriteByte(...) }
    
    // Metadata.Version (no boundary marker!)
    { v := uint32(v.Metadata.Version); buf.WriteByte(...) }
    
    // Metadata.Author
    { l := uint16(len(v.Metadata.Author)); buf.WriteByte(...) }
    buf.WriteString(v.Metadata.Author)
    
    return buf.Bytes()
}
```

**Wire format:**
```
[4 bytes Port][4 bytes Version][2 bytes len][N bytes Author]
```

**Why no struct markers:**
- ✅ Non-optional structs are always present (no ambiguity)
- ✅ Reduces wire size (no extra bytes)
- ✅ Struct nesting is compile-time information
- ✅ Decoder knows structure from schema

**Optional structs DO have markers:**
```go
// Optional nested struct
type Config struct {
    Metadata *Metadata  // Pointer = optional
}

// Generated code:
if v.Metadata == nil {
    buf.WriteByte(0x00)  // Absent marker
} else {
    buf.WriteByte(0x01)  // Present marker
    // Encode fields...
}
```

### Strategy 5: Optional Field Encoding

**Pattern for all optional types:**
```go
if v.OptionalField == nil {
    buf.WriteByte(0x00)  // Not present
} else {
    buf.WriteByte(0x01)  // Present
    // Encode dereferenced value: *v.OptionalField
}
```

**Example with optional int32:**
```go
OptionalPort *int32

// Generated:
if v.OptionalPort == nil {
    buf.WriteByte(0x00)
} else {
    buf.WriteByte(0x01)
    v := uint32(*v.OptionalPort)  // Dereference pointer
    buf.WriteByte(byte(v))
    buf.WriteByte(byte(v>>8))
    buf.WriteByte(byte(v>>16))
    buf.WriteByte(byte(v>>24))
}
```

---

## Buffer Management

### Current Approach: `bytes.Buffer`

```go
func EncodeMessage(v Type) []byte {
    buf := &bytes.Buffer{}  // Allocate new buffer
    // ... encode operations ...
    return buf.Bytes()       // Return slice to internal buffer
}
```

**How `bytes.Buffer` works internally:**
1. Starts with 64-byte bootstrap array (on stack or first alloc)
2. Grows exponentially when capacity exceeded: 64 → 128 → 256 → 512...
3. Growth factor: 2x (similar to Go slices)
4. `Bytes()` returns slice to internal buffer (no copy if size fits)

**Performance characteristics:**
- Small messages (<64 bytes): 1 allocation (~90 bytes)
- Medium messages (64-1024 bytes): 2-4 allocations (grows 2-3 times)
- Large messages (>1KB): 4-6 allocations
- Allocation cost: ~50-100ns per allocation on modern CPUs

**Why we use it:**
- ✅ Simple API (`WriteByte`, `Write`, `WriteString`)
- ✅ Handles growth automatically
- ✅ Efficient for variable-size data
- ✅ No need to calculate exact size upfront

**Alternatives considered:**

1. **Pre-allocated `[]byte` with append:**
   ```go
   data := make([]byte, 0, estimatedSize)
   data = append(data, byte1)
   data = append(data, byte2)
   ```
   - ❌ Doesn't save allocations (append grows similarly)
   - ❌ More complex code generation
   - ✅ Slightly less overhead (no `bytes.Buffer` struct)

2. **Exact size calculation + single allocation:**
   ```go
   size := calculateExactSize(v)  // Traverse structure once
   data := make([]byte, size)
   pos := 0
   // Write directly to data[pos]
   ```
   - ✅ Single allocation (no growth)
   - ❌ Double traversal (calculate, then encode)
   - ❌ Only works for deterministic sizes (no arrays/strings)
   - ⚠️ Likely slower overall for variable-size data

3. **Buffer pooling:**
   ```go
   var pool = sync.Pool{New: func() { return &bytes.Buffer{} }}
   buf := pool.Get().(*bytes.Buffer)
   defer pool.Put(buf)
   ```
   - ✅ Reduces GC pressure for high-throughput batch encoding
   - ❌ Adds API complexity
   - ❌ Minimal benefit for typical request/response pattern
   - ⚠️ Users can add pooling at call site if needed

---

## Optimization History

### Baseline (Initial Implementation)
- Used `binary.Write()` for multi-byte values
- Used `io.Reader` for decoding
- **Performance:** ~37µs for complex benchmark

### Step 1: Remove binary.Write/Read
- Manual byte encoding with bit shifts
- Direct `WriteByte()` calls
- **Speedup:** 2x (37µs → 19µs)
- **Why:** Eliminated reflection and interface overhead

### Step 2: Direct Slice Indexing (Decode)
- Removed `io.Reader` allocation
- Direct indexing: `data[pos]`, `data[pos+1]`, etc.
- Position tracking with `pos` variable
- **Speedup:** 3-11x on decode path
- **Why:** Eliminated Reader interface overhead, better inlining

### Step 3: Bulk Array Writes
- Attempted temp buffer approach for multi-byte values
- **Result:** Regression (slower)
- **Reverted:** Individual `WriteByte()` is faster for small values

### Step 4: Zero-Copy Array Encoding
- Use `unsafe.Slice()` to reinterpret arrays as `[]byte`
- Single `Write()` call for entire array
- **Speedup:** 25x on array encoding (50µs → 2µs)
- **Why:** Eliminated loop overhead, became a single memcpy

### Final Results
- **2-5x faster than protobuf** on most benchmarks
- **array_int:** 4.94x faster than protobuf
- **nested:** 4.70x faster than protobuf
- **struct:** 3.31x faster than protobuf

---

## Current Limitations & Design Decisions

### 1. Array/String Size Limit: 65,535 elements
- **Reason:** Length stored as uint16 (2 bytes)
- **Rationale:** Prevents buffer overflow attacks, reasonable for most data
- **Workaround:** Split large arrays into chunks or use streaming

### 2. Little-Endian Only
- **Reason:** Zero-copy array encoding assumes little-endian
- **Platforms:** x86, x86-64, ARM, ARM64 (covers 99%+ of systems)
- **Workaround:** Could add byte swapping for big-endian (with performance cost)

### 3. No Schema Versioning
- **Reason:** Focused on deterministic encoding, not evolution
- **Use case:** FFI boundaries where both sides compiled from same schema
- **Workaround:** Use different message types for different versions

### 4. No Compression
- **Reason:** Keep encoding simple and fast
- **Use case:** Small messages where compression overhead > size savings
- **Workaround:** Apply compression at transport layer (gzip, zstd)

### 5. Embedded Structs Not Supported
- **Reason:** Go-specific language feature, not cross-language
- **Wire format:** Wouldn't change (same encoding)
- **Status:** Could be added for convenience without breaking changes

---

## Generated Code Example

**Input Schema (struct.ffi):**
```go
package test

type Message = Config

type Config struct {
    Host       string  `json:"host"`
    Port       int32   `json:"port"`
    EnableSSL  bool    `json:"enableSSL"`
    Timeout    float32 `json:"timeout"`
    MaxRetries int32   `json:"maxRetries"`
}
```

**Generated Encoder:**
```go
package test

import (
    "bytes"
    "math"
)

type Config struct {
    Host       string  `json:"host"`
    Port       int32   `json:"port"`
    EnableSSL  bool    `json:"enableSSL"`
    Timeout    float32 `json:"timeout"`
    MaxRetries int32   `json:"maxRetries"`
}

// EncodeConfigMessage encodes Message to binary wire format.
func EncodeConfigMessage(v Config) []byte {
    buf := &bytes.Buffer{}
    
    // Host (string)
    { 
        l := uint16(len(v.Host))
        buf.WriteByte(byte(l))
        buf.WriteByte(byte(l>>8))
    }
    buf.WriteString(v.Host)
    
    // Port (int32)
    { 
        v := uint32(v.Port)
        buf.WriteByte(byte(v))
        buf.WriteByte(byte(v>>8))
        buf.WriteByte(byte(v>>16))
        buf.WriteByte(byte(v>>24))
    }
    
    // EnableSSL (bool)
    if v.EnableSSL {
        buf.WriteByte(0x01)
    } else {
        buf.WriteByte(0x00)
    }
    
    // Timeout (float32)
    { 
        v := math.Float32bits(v.Timeout)
        buf.WriteByte(byte(v))
        buf.WriteByte(byte(v>>8))
        buf.WriteByte(byte(v>>16))
        buf.WriteByte(byte(v>>24))
    }
    
    // MaxRetries (int32)
    { 
        v := uint32(v.MaxRetries)
        buf.WriteByte(byte(v))
        buf.WriteByte(byte(v>>8))
        buf.WriteByte(byte(v>>16))
        buf.WriteByte(byte(v>>24))
    }
    
    return buf.Bytes()
}
```

**Wire Format Layout:**
```
[2 bytes Host length]
[N bytes Host UTF-8 data]
[4 bytes Port]
[1 byte EnableSSL]
[4 bytes Timeout as IEEE 754]
[4 bytes MaxRetries]

Total: 15 + len(Host) bytes
```

---

## Performance Characteristics

### Encoding Performance by Type

| Type | Operation | Cost | Notes |
|------|-----------|------|-------|
| `int8`, `bool` | 1 `WriteByte()` | ~2ns | Single byte, no conversion |
| `int16` | 2 `WriteByte()` | ~4ns | Two bytes with shift |
| `int32`, `float32` | 4 `WriteByte()` | ~8ns | Four bytes with shifts |
| `int64`, `float64` | 8 `WriteByte()` | ~16ns | Eight bytes with shifts |
| `string` | Length + `WriteString()` | ~5ns + len | Length overhead minimal |
| `[]int32` (5000 elem) | Unsafe bulk | ~1.5µs | Zero-copy, single memcpy |
| Struct (5 fields) | Field sum | ~51ns | Sum of individual fields |

### Memory Allocation

| Message Size | Allocations | Total Memory | Growth Pattern |
|--------------|-------------|--------------|----------------|
| <64 bytes | 1 | ~90 bytes | Bootstrap array |
| 64-256 bytes | 2 | ~320 bytes | Grow to 256 |
| 256-1024 bytes | 3 | ~1200 bytes | Grow to 1024 |
| 1KB-4KB | 4 | ~5KB | Grow to 4096 |
| >4KB | 5+ | 2x final | Exponential |

### Comparison with Protobuf

| Benchmark | ffire | Protobuf | Speedup | Why? |
|-----------|-------|----------|---------|------|
| struct | 51ns | 195ns | **3.8x** | No reflection, direct writes |
| array_int | 6.6µs | 32.7µs | **4.9x** | Zero-copy vs element loop |
| complex | 12.1µs | 22.4µs | **1.9x** | Simpler wire format |
| nested | 7.0µs | 32.9µs | **4.7x** | No varint overhead |

---

## Future Optimization Opportunities (Not Implemented)

### 1. SIMD for Bulk Operations
- **Potential:** 10-15% speedup on array encoding
- **Complexity:** High (assembly for x86/ARM64, testing)
- **Decision:** Not worth complexity for marginal gain

### 2. Buffer Pre-allocation
- **Potential:** 10-20% speedup for fixed-size messages
- **Complexity:** Medium (calculate size, handle variable data)
- **Decision:** `bytes.Buffer` already efficient enough

### 3. Bounds Check Elimination
- **Potential:** 5-10% speedup on decode
- **Complexity:** Low (add bounds check hints)
- **Decision:** Modern Go compiler already optimizes this

### 4. Direct `[]byte` for Fixed-Size Structs
- **Potential:** 2-3x speedup on small structs
- **Complexity:** High (dual code paths, size calculation)
- **Decision:** Only benefits 2-3 benchmarks, not worth it

---

## Summary

The ffire encoder achieves high performance through:

1. **Zero-copy operations** - `unsafe.Slice()` for primitive arrays
2. **Direct byte manipulation** - No reflection or interfaces
3. **Simple wire format** - No varints, no zigzag encoding
4. **Deterministic layout** - Enables decoder optimizations
5. **Smart code generation** - Conditional imports, type-specific paths

The result is **2-5x faster than protobuf** while maintaining:
- Simple, readable generated code
- Cross-language compatibility (Go, C++, Swift)
- Deterministic output (same input = same bytes)
- Memory safety (no dangling pointers or buffer overflows)
