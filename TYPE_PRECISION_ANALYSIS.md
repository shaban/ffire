# ffire Type Precision Analysis for Dynamic Language Bindings

## Current ffire Type System

**Primitives supported:**
- `bool` (1 byte)
- `int8` (1 byte, signed)
- `int16` (2 bytes, signed)
- `int32` (4 bytes, signed)
- `int64` (8 bytes, signed)
- `float32` (4 bytes, IEEE 754)
- `float64` (8 bytes, IEEE 754)
- `string` (2-byte length prefix + UTF-8)

**Missing unsigned types:**
- ❌ `uint8`
- ❌ `uint16`
- ❌ `uint32`
- ❌ `uint64`

## Type Mapping Challenges by Language

### Python

**Native types:**
- `int` - arbitrary precision (can represent any integer)
- `float` - always 64-bit (float64)
- `bool` - separate type
- `str` - Unicode strings
- `bytes` - byte arrays

**Mapping:**
| ffire Type | Python Type | Bidirectional? | Notes |
|------------|-------------|----------------|-------|
| `bool` | `bool` | ✅ Perfect | Direct mapping |
| `int8` | `int` | ✅ Perfect | Python int can hold -128 to 127 |
| `int16` | `int` | ✅ Perfect | Python int can hold -32768 to 32767 |
| `int32` | `int` | ✅ Perfect | Python int can hold ±2^31 |
| `int64` | `int` | ✅ Perfect | Python int can hold ±2^63 |
| `float32` | `float` | ⚠️ Precision loss | Python float is 64-bit, will widen |
| `float64` | `float` | ✅ Perfect | Direct mapping |
| `string` | `str` | ✅ Perfect | UTF-8 codec, max 65535 bytes |
| `[]int32` | `list[int]` | ✅ Perfect | Array length limited to 65535 |

**Issues:**
1. **float32 → float** - Decode widens to 64-bit (acceptable)
2. **float → float32** - Encode narrows (potential precision loss if value doesn't fit)
3. **Array length** - Limited to 65535 elements (uint16 length prefix)

### PHP

**Native types:**
- `int` - platform-dependent (32-bit or 64-bit)
- `float` - always 64-bit (float64)
- `bool` - separate type
- `string` - byte strings (not necessarily UTF-8)
- `array` - heterogeneous arrays

**Mapping:**
| ffire Type | PHP Type | Bidirectional? | Notes |
|------------|----------|----------------|-------|
| `bool` | `bool` | ✅ Perfect | Direct mapping |
| `int8` | `int` | ✅ Perfect | PHP int can hold -128 to 127 |
| `int16` | `int` | ✅ Perfect | PHP int can hold -32768 to 32767 |
| `int32` | `int` | ✅ Perfect | PHP int can hold ±2^31 |
| `int64` | `int` (64-bit) | ✅ Perfect | On 64-bit PHP |
| `int64` | `string` (32-bit) | ⚠️ String repr | On 32-bit PHP, use string |
| `float32` | `float` | ⚠️ Precision loss | PHP float is 64-bit, will widen |
| `float64` | `float` | ✅ Perfect | Direct mapping |
| `string` | `string` | ⚠️ Encoding | PHP strings are byte arrays, UTF-8 not guaranteed |
| `[]int32` | `array` | ✅ Perfect | Indexed array |

**Issues:**
1. **32-bit PHP**: int64 must be represented as string (like upb does)
2. **String encoding**: PHP strings are byte arrays, need UTF-8 validation
3. **float32**: Same precision loss as Python

### JavaScript/Node.js

**Native types:**
- `number` - always 64-bit float (!)
- `bigint` - arbitrary precision integers
- `boolean` - separate type
- `string` - UTF-16 strings (!)
- `Buffer` / `Uint8Array` - byte arrays

**Mapping:**
| ffire Type | JS Type | Bidirectional? | Notes |
|------------|---------|----------------|-------|
| `bool` | `boolean` | ✅ Perfect | Direct mapping |
| `int8` | `number` | ⚠️ Precision OK | Number can represent exactly |
| `int16` | `number` | ⚠️ Precision OK | Number can represent exactly |
| `int32` | `number` | ⚠️ Precision OK | Number can represent exactly |
| `int64` | `bigint` | ✅ Perfect | BigInt for full precision |
| `int64` | `number` | ❌ Precision loss | If using number, loses precision beyond ±2^53 |
| `float32` | `number` | ⚠️ Precision loss | Number is 64-bit, will widen |
| `float64` | `number` | ✅ Perfect | Direct mapping |
| `string` | `string` | ⚠️ UTF-16 → UTF-8 | Conversion required, max 65535 UTF-8 bytes |
| `[]int32` | `Array<number>` | ✅ Perfect | Typed arrays available |
| `[]int32` | `Int32Array` | ✅ Perfect | Zero-copy possible |

**Critical Issues:**
1. **int64 precision**: JavaScript `number` can only safely represent integers up to ±2^53
   - **Solution**: Use `bigint` for int64 (requires Node.js 10.4+)
   - **Alternative**: String representation (like protobuf.js does)
2. **String encoding**: UTF-16 ↔ UTF-8 conversion required
3. **Typed arrays**: Could use `Int32Array`, `Float32Array` for zero-copy

### Ruby

**Native types:**
- `Integer` - arbitrary precision (unified Fixnum/Bignum)
- `Float` - always 64-bit (float64)
- `TrueClass`/`FalseClass` - boolean
- `String` - byte strings with encoding
- `Array` - heterogeneous arrays

**Mapping:**
| ffire Type | Ruby Type | Bidirectional? | Notes |
|------------|-----------|----------------|-------|
| `bool` | `true`/`false` | ✅ Perfect | Direct mapping |
| `int8` | `Integer` | ✅ Perfect | Ruby Integer can hold any value |
| `int16` | `Integer` | ✅ Perfect | Ruby Integer can hold any value |
| `int32` | `Integer` | ✅ Perfect | Ruby Integer can hold any value |
| `int64` | `Integer` | ✅ Perfect | Ruby Integer can hold any value |
| `float32` | `Float` | ⚠️ Precision loss | Ruby Float is 64-bit, will widen |
| `float64` | `Float` | ✅ Perfect | Direct mapping |
| `string` | `String` (UTF-8) | ✅ Perfect | Force encoding to UTF-8 |
| `[]int32` | `Array` | ✅ Perfect | Array length limited to 65535 |

**Issues:**
1. **float32**: Same widening behavior as Python
2. **String encoding**: Must force UTF-8 encoding

---

## Recommendations for Type Precision

### 1. Should we add unsigned types?

**Pros:**
- More precise mapping for languages with unsigned types (C, C++, Rust, C#, Java)
- Can represent 0 to 2^32 instead of ±2^31 for uint32
- Matches common use cases (array indices, counts, IDs)

**Cons:**
- Python, Ruby, JavaScript don't have native unsigned types
- Increases language generator complexity
- Schema language needs syntax for unsigned types

**My Recommendation:** **NO**, don't add unsigned types yet. Here's why:

1. **Dynamic languages don't benefit**: Python/Ruby/JS all use arbitrary-precision or 64-bit ints
2. **Signed is sufficient**: -2^31 to 2^31 is enough for most use cases
3. **Complexity**: Adds ~50% more type mapping code for marginal benefit
4. **Wire format change**: Would need to update wire format spec

**If needed later**: Add as `uint8`, `uint16`, `uint32`, `uint64` with same wire format (just interpret differently in type system).

### 2. How to handle type precision issues?

**float32 precision loss (encode side):**
```python
# Python: float (64-bit) → float32 (32-bit)
# Issue: Value might not fit in float32 range or lose precision

# Solution 1: Clamp and warn
def encode_float32(value: float) -> bytes:
    if value > 3.4e38:
        warnings.warn(f"Value {value} exceeds float32 range, clamping")
        value = 3.4e38
    return struct.pack('<f', value)

# Solution 2: Raise error (strict mode)
def encode_float32_strict(value: float) -> bytes:
    f32_value = struct.pack('<f', value)
    roundtrip = struct.unpack('<f', f32_value)[0]
    if abs(value - roundtrip) > value * 1e-6:  # Relative error check
        raise ValueError(f"Value {value} cannot be represented precisely as float32")
    return f32_value
```

**int64 in JavaScript:**
```javascript
// JavaScript: Use bigint for int64
class Message {
  encodeInt64(value) {
    if (typeof value === 'bigint') {
      // Perfect: bigint can represent full range
      const buffer = Buffer.allocUnsafe(8);
      buffer.writeBigInt64LE(value);
      return buffer;
    } else if (typeof value === 'number') {
      // Warning: might lose precision
      if (!Number.isSafeInteger(value)) {
        console.warn(`Value ${value} is not a safe integer, may lose precision`);
      }
      return this.encodeInt64(BigInt(value));
    }
  }
  
  decodeInt64(buffer, offset) {
    // Return bigint by default
    return buffer.readBigInt64LE(offset);
  }
}
```

**String length limits:**
```python
# Wire format: uint16 length prefix = max 65535 bytes
def encode_string(value: str) -> bytes:
    utf8_bytes = value.encode('utf-8')
    if len(utf8_bytes) > 65535:
        raise ValueError(f"String too long: {len(utf8_bytes)} bytes (max 65535)")
    length = struct.pack('<H', len(utf8_bytes))
    return length + utf8_bytes
```

**Array length limits:**
```python
# Wire format: uint16 length prefix = max 65535 elements
def encode_array(values: list) -> bytes:
    if len(values) > 65535:
        raise ValueError(f"Array too long: {len(values)} elements (max 65535)")
    length = struct.pack('<H', len(values))
    # ... encode elements
```

---

## Summary: Type Precision Matrix

| Concern | Python | PHP | JavaScript | Ruby | Recommendation |
|---------|--------|-----|------------|------|----------------|
| int8-64 | ✅ Perfect | ✅ Perfect | ⚠️ Use bigint for int64 | ✅ Perfect | Document int64 → bigint for JS |
| float32 | ⚠️ Widens to 64 | ⚠️ Widens to 64 | ⚠️ Widens to 64 | ⚠️ Widens to 64 | Acceptable, document behavior |
| float64 | ✅ Perfect | ✅ Perfect | ✅ Perfect | ✅ Perfect | No issues |
| strings | ✅ Perfect | ⚠️ UTF-8 validation | ⚠️ UTF-16 ↔ UTF-8 | ✅ Perfect | Validate UTF-8, document limits |
| arrays | ✅ Max 65535 | ✅ Max 65535 | ✅ Max 65535 | ✅ Max 65535 | Document limit, throw on overflow |
| unsigned | ❌ N/A | ❌ N/A | ❌ N/A | ❌ N/A | **Don't add** |

**Bidirectional compatibility: ✅ ACHIEVABLE**

With proper handling of:
1. int64 → bigint in JavaScript
2. float32 precision loss (document as acceptable widening on decode)
3. String/array length limits (throw error on encode if exceeded)
4. UTF-8 validation for PHP strings

**Edge cases to handle:**
- Array > 65535 elements → raise error on encode
- String > 65535 UTF-8 bytes → raise error on encode
- int64 in JS → use bigint (or document precision loss with number)
- float32 precision → document widening behavior, optional strict mode

**No unsigned types needed** - signed types provide sufficient range for dynamic languages.
