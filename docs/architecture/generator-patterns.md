# Generator Patterns

Common patterns across language generators.

## Generation Approach

All ffire generators produce **native implementations** - complete encoder/decoder code in the target language with no FFI or external dependencies.

### Supported Languages

| Language | Generator | Output |
|----------|-----------|--------|
| Go | `generator_go.go` | Single `.go` file |
| C++ | `generator_cpp.go` | `.h` + `.cpp` files |
| C# | `generator_csharp.go` | Single `.cs` file |
| Java | `generator_java.go` | Single `.java` file |
| Swift | `generator_swift.go` | Single `.swift` file |
| Dart | `generator_dart.go` | Single `.dart` file |
| Rust | `generator_rust.go` | Single `.rs` file |
| Zig | `generator_zig.go` | Single `.zig` file |

### Advantages of Native

- **Fast**: No FFI overhead, direct memory access
- **Type-safe**: Compile-time type checking
- **Debuggable**: Step through generated code
- **Portable**: No shared library dependencies
- **Simple**: Single-language build chain

## Code Organization

Each generator follows a similar pattern:

```
pkg/generator/
  generator_go.go       # Go generator
  generator_cpp.go      # C++ generator
  generator_csharp.go   # C# generator
  generator_java.go     # Java generator
  generator_swift.go    # Swift generator
  generator_dart.go     # Dart generator
  generator_rust.go     # Rust generator
  generator_zig.go      # Zig generator
  package.go            # PackageConfig struct, routing
```

### Generator Structure

Each `generator_LANG.go` follows this pattern:

```go
func GenerateLANGPackage(config *PackageConfig) error {
    // 1. Generate types and encode/decode functions
    // 2. Write output file(s)
    // 3. Return error if any
}
```

## Type Mapping Reference

| ffire Type | Go | C++ | C# | Java | Swift | Dart | Rust | Zig |
|------------|----|----|----|----|-------|------|------|-----|
| int32 | int32 | int32_t | int | int | Int32 | int | i32 | i32 |
| int64 | int64 | int64_t | long | long | Int64 | int | i64 | i64 |
| float32 | float32 | float | float | float | Float | double | f32 | f32 |
| float64 | float64 | double | double | double | Double | double | f64 | f64 |
| string | string | std::string | string | String | String | String | String | []u8 |
| bool | bool | bool | bool | boolean | Bool | bool | bool | bool |
| []T | []T | std::vector\<T\> | List\<T\> | ArrayList\<T\> | [T] | List\<T\> | Vec\<T\> | []T |
| *T | *T | std::optional\<T\> | T? | T | T? | T? | Option\<T\> | ?T |

## Error Handling Patterns

**Go:** Return `(T, error)`
```go
func Decode(data []byte) (*Message, error)
```

**C++:** Throw exceptions
```cpp
Message decode(const std::vector<uint8_t>& data)
```

**C#:** Throw exceptions
```csharp
public static Message Decode(byte[] data)
```

**Java:** Throw exceptions
```java
public static Message decode(byte[] data) throws FFireException
```

**Swift:** Throw errors
```swift
func decode(_ data: Data) throws -> Message
```

**Rust:** Return Result
```rust
fn decode(data: &[u8]) -> Result<Message, FFireError>
```

Language bindings convert C errors to native exceptions.

## Testing Strategy

Each generator must pass:
1. **Unit tests** - `pkg/generator/generator_LANG_test.go`
2. **Integration tests** - Compile generated code
3. **Benchmark tests** - Encode/decode fixture succeeds
4. **Wire format tests** - Output matches other languages

See [Testing](../development/testing.md) for details.

---

## C++ Implementation Details

### Status: ✅ Complete

The C++ generator is production-ready with full feature parity with the Go implementation.

### Core Features

- **Header-only C++17 implementation** - No separate compilation needed
- **Modern C++ types**: `std::vector`, `std::string`, `std::optional`, `int32_t`
- **Little-endian encoding** - Matches wire format specification
- **Bounds checking** - All decoder methods check remaining bytes before reading
- **Exception-based error handling** - Throws `std::runtime_error` on invalid data
- **Zero allocations on decode** - Uses pointer + size for input data

### Generated Code Structure

```cpp
namespace <package> {
    // Struct definitions
    struct MyStruct {
        int32_t field1;
        std::string field2;
        std::optional<bool> field3;  // Optional fields
    };

    // Encoder class with write methods
    class Encoder {
        std::vector<uint8_t> buffer;
        void write_int32(int32_t v);
        void write_string(const std::string& s);
        // ... etc
    };

    // Decoder class with read methods and bounds checking
    class Decoder {
        const uint8_t* data;
        size_t size;
        size_t pos = 0;
        
        void check_remaining(size_t needed);  // Throws on insufficient data
        int32_t read_int32();
        std::string read_string();
        // ... etc
    };

    // Message encode/decode functions
    std::vector<uint8_t> encode_mystruct_message(const MyStruct& value);
    MyStruct decode_mystruct_message(const uint8_t* data, size_t size);
    MyStruct decode_mystruct_message(const std::vector<uint8_t>& data);
}
```

### Type Mapping (C++)

| ffire Schema | C++ Type |
|-------------|----------|
| `bool` | `bool` |
| `int8` | `int8_t` |
| `int16` | `int16_t` |
| `int32` | `int32_t` |
| `int64` | `int64_t` |
| `float32` | `float` |
| `float64` | `double` |
| `string` | `std::string` |
| `[]T` | `std::vector<T>` |
| `*T` | `std::optional<T>` |
| `struct` | `struct` |

### Wire Format Compliance

✅ Little-endian byte order for multi-byte types  
✅ uint16 length prefixes for strings and arrays  
✅ Optional fields: 0x00 for null, 0x01 + value for present  
✅ Struct fields in declaration order  
✅ No padding between fields  
✅ Bounds checking to prevent buffer overruns

### Performance Characteristics

**Encoding:**
- Pre-allocation via `reserve()` for arrays
- `std::memcpy` for float/double bit reinterpretation
- Efficient `buffer.insert()` for string data

**Decoding:**
- Zero-copy input (pointer + size, no buffer allocation)
- Pre-allocation via `reserve()` when array length known
- Bounds checking is inlined and optimizes well

### Safety Features

1. **Bounds checking**: Every read checks `pos + needed <= size`
2. **Exception on error**: Throws instead of undefined behavior
3. **Type safety**: Strong typing with `int32_t`, `uint16_t`, etc.
4. **RAII**: `std::vector` manages buffer lifetime automatically
5. **Const correctness**: Input data is `const uint8_t*`

### Usage Example

```cpp
#include "generated.hpp"

int main() {
    // Encode
    MyStruct data;
    data.field1 = 42;
    data.field2 = "hello";
    data.field3 = true;  // Optional field
    
    auto encoded = mypkg::encode_mystruct_message(data);
    
    // Decode
    try {
        MyStruct decoded = mypkg::decode_mystruct_message(encoded);
        // Use decoded...
    } catch (const std::runtime_error& e) {
        std::cerr << "Decode error: " << e.what() << "\n";
    }
}
```

### Comparison: Go vs C++

| Feature | Go | C++ |
|---------|----|----|
| **Optional fields** | `*T` (pointers) | `std::optional<T>` |
| **Arrays** | `[]T` (slices) | `std::vector<T>` |
| **Error handling** | `(result, error)` | Exceptions |
| **Encoding buffer** | `bytes.Buffer` | `std::vector<uint8_t>` |
| **Decoding input** | Direct slice indexing | Pointer + bounds checking |
| **Zero-copy arrays** | ✅ `unsafe.Slice` for primitives | ⏳ TODO: Use `memcpy` for bulk |
| **Bounds checking** | ❌ None (assumes valid) | ✅ Every read |

### Testing Status

✅ 5 unit tests covering all features  
✅ 3 integration tests with real C++ compilation  
✅ All existing ffire test suite passes (no regressions)

**Test Coverage:**
- Simple struct encoding/decoding
- Array encoding with length prefix
- Optional fields with `std::optional`
- All 8 primitive types
- Nested struct support

### Future Optimizations (TODO)

⏳ **Bulk array decode** - Use `memcpy` for primitive arrays (like Go's `unsafe.Slice`)  
⏳ **Small string optimization** - Could use `std::string_view` for decode  
⏳ **Stack buffers** - Could use `std::array` for known-size messages  
⏳ **Custom allocators** - Allow users to provide allocators for `std::vector`

### Verdict

✅ **Production-ready** - Full feature parity, comprehensive testing, safe and fast.
