# Generator Patterns

Common patterns across language generators.

## Generation Modes

ffire generators fall into three categories:

### 1. Native Implementation
**Languages:** Go, C++, Rust (future)

Generates complete encoder/decoder in target language. No FFI, no external dependencies.

**Example (Go):**
```
schema.ffi → pkg/generator/generator_go.go
           → generated.go (encode/decode functions)
```

**Advantages:**
- Fastest (no FFI overhead)
- Type-safe
- Debuggable
- Single-language stack

**When to use:** Language has good buffer manipulation, no GC issues with wire format.

### 2. C ABI Wrapper  
**Languages:** Swift, Dart, Python, JavaScript, Ruby

Generates language bindings that wrap a C ABI dylib.

**Example (Swift):**
```
schema.ffi → pkg/generator/generator_c_abi.go (C++ dylib)
           → pkg/generator/generator_swift.go (Swift bindings)
           
Result:
  lib/libtest.dylib        # C ABI implementation
  Sources/array_int.swift   # Swift wrapper
```

**Flow:**
1. Generate C++ encoder/decoder
2. Generate C ABI functions (`intlist_encode`, `intlist_decode`)
3. Compile to dylib
4. Generate language bindings that call dylib

**Advantages:**
- Write encoder once (C++), use everywhere
- Consistent wire format across languages
- Fast (compiled C++)

**When to use:** Language has good FFI, small stdlib for binary manipulation.

### 3. Hybrid
**Languages:** Swift (has both patterns)

Both native Swift implementation AND C ABI wrapper available.

**Use case:** Native for iOS/watchOS (embedded), C ABI for macOS/server.

## C ABI Pattern

### Generator Flow

```
schema.ffi
    ↓
generator_c_abi.go
    ↓
generated_c.h          // Header with function declarations
generated_c.cpp        // C++ implementation
    ↓
compile (clang++)
    ↓
libTEST.dylib         // Shared library
    ↓
generator_swift.go    // Language bindings
    ↓
swift_bindings.swift  // Wrapper types
```

### C ABI Functions

For each message type `IntList`:

**Decode:**
```c
IntListHandle intlist_decode(
    const uint8_t* data, 
    int32_t size,
    char** error_msg
);
```

**Encode:**
```c
size_t intlist_encode(
    IntListHandle handle,
    uint8_t** out_data,
    char** error_msg
);
```

**Free:**
```c
void intlist_free(IntListHandle handle);
void intlist_free_data(uint8_t* data);
void intlist_free_error(char* error);
```

### Naming Convention

C ABI uses lowercase message name + underscore + function:
- `Config` → `config_encode`, `config_decode`, `config_free`
- `IntList` → `intlist_encode`, `intlist_decode`, `intlist_free`

**Why lowercase?** C convention, avoids case-sensitivity issues across platforms.

### Handle Pattern

Messages are opaque handles (pointers):
```c
typedef void* IntListHandle;
```

Language bindings wrap handles:
```swift
public class IntList {
    private var handle: OpaquePointer
    
    public init(handle: OpaquePointer) {
        self.handle = handle
    }
    
    deinit {
        intlist_free(handle)
    }
}
```

### Memory Management

**Ownership:**
- Decode: C++ allocates handle, caller frees
- Encode: C++ allocates buffer, caller frees
- Errors: C++ allocates string, caller frees

**Pattern:**
```swift
// Decode
let handle = decode(data)   // C++ owns memory
defer { free(handle) }      // Swift ensures cleanup

// Encode
var outPtr: UnsafeMutablePointer<UInt8>?
let size = encode(handle, &outPtr, nil)
defer { free_data(outPtr) }  // Free C++ buffer
let data = Data(bytes: outPtr, count: size)
```

### Deployment Model

**Canonical Approach: Bundled dylib per package**

Each language package bundles its own copy of the dylib:
```
python_package/
  mypackage/
    lib/libmyschema.dylib    # Bundled, not system-wide
    __init__.py

dart_package/
  lib/libmyschema.dylib      # Separate copy
  myschema.dart
```

**Why bundled?**
- ✅ No version conflicts between applications
- ✅ No system-wide installation required
- ✅ Each app controls its own ffire version
- ✅ No symbol versioning needed (no shared library conflicts)

**Not supported: System-wide installation**
- ❌ `/usr/local/lib/libffire.dylib` - would require symbol versioning
- ❌ Plugin systems with multiple ffire versions - unsupported use case

**Symbol versioning:** Not implemented because bundled deployment avoids conflicts entirely.

## Library Naming

Libraries are named after the **package**, not the schema file:

```go
// schema.ffi
package test

// Generates:
lib/libtest.dylib     // NOT libarray_int.dylib
```

**Why?** Multiple schemas can share one package:
```
package myapp

type User struct { ... }
type Post struct { ... }

// Both compile into:
libmyapp.dylib  // Contains user_encode, post_encode, etc.
```

### Bundled Library Location

All generators place the dylib inside the package directory:

**Python (ctypes):**
```python
_lib_path = os.path.join(os.path.dirname(__file__), _lib_name)
_lib = ctypes.CDLL(_lib_path)
```

**Python (pybind11):**
```python
# Compiled as Python extension, automatically loaded
```

**Dart:**
```dart
DynamicLibrary.open('lib/libtest.dylib')  // Relative to package
```

**Swift:**
```swift
// Linked via Package.swift linkerSettings
.target(
    name: "test",
    linkerSettings: [
        .unsafeFlags(["-Llib", "-ltest"])
    ]
)
```

This ensures each package is self-contained with no system dependencies.

## Code Organization

### Go Generator
```
pkg/generator/generator_go.go
  → Single file per message
  → Pure Go, no FFI
```

### C++ Generator  
```
pkg/generator/generator_cpp.go
  → generated.h (types, declarations)
  → generated.cpp (encode/decode implementation)
  → Compiles to binary (bench) or library (for FFI)
```

### Swift Generator
```
pkg/generator/generator_swift.go
  → Package.swift (SPM manifest)
  → Sources/{package}/{package}.swift
  → Depends on C ABI dylib via linkerSettings
```

### Multi-file Generators
```
pkg/generator/
  generator_LANG.go      # Main generator
  generator_c_abi.go     # Shared C ABI generation
  package.go             # PackageConfig struct
```

## Type Mapping Reference

| ffire Type | Go | C++ | Swift | Python |
|------------|----|----|-------|--------|
| int32 | int32 | int32_t | Int32 | int |
| int64 | int64 | int64_t | Int64 | int |
| float32 | float32 | float | Float | float |
| float64 | float64 | double | Double | float |
| string | string | std::string | String | str |
| bool | bool | bool | Bool | bool |
| []T | []T | std::vector\<T\> | [T] | list[T] |
| *T (optional) | *T | std::optional\<T\> | T? | Optional[T] |

## Error Handling Patterns

**Go:** Return `(T, error)`
```go
func Decode(data []byte) (*Message, error)
```

**C++:** Throw exceptions
```cpp
Message decode(const std::vector<uint8_t>& data)
```

**Swift:** Throw errors
```swift
func decode(_ data: Data) throws -> Message
```

**C ABI:** Error out-parameter
```c
Handle decode(const uint8_t* data, int32_t size, char** error)
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
