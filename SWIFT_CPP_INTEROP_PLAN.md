# Swift C++ Interop Implementation Plan

## Current State
- Swift uses C ABI wrapper with OpaquePointer
- Data copying: Swift Data → C pointer → OpaquePointer (decode)
- Data copying: C function → pointer → Swift Data → free (encode)
- Performance: 11.15 μs (1.41x slower than C++ 7.89 μs)

## Goal
**Use C++ types directly with ZERO translation/copying**

## Requirements from Swift Documentation

### 1. Package.swift Changes
```swift
// swift-tools-version:5.9  // MUST be 5.9+
import PackageDescription

let package = Package(
    name: "ffire_swift_complex",
    platforms: [
        .macOS(.v13),  // MUST be macOS 13+ (was 10.15)
        .iOS(.v16),    // Updated from iOS 13
        .tvOS(.v16),   // Updated from tvOS 13
        .watchOS(.v9)  // Updated from watchOS 6
    ],
    products: [
        .library(
            name: "ffire_swift_complex",
            targets: ["ffire_swift_complex"]
        ),
    ],
    targets: [
        .target(
            name: "ffire_swift_complex",
            dependencies: [],
            path: "Sources/ffire_swift_complex",
            cxxSettings: [
                .headerSearchPath("../../generated/ffire_cpp_complex"),
            ],
            swiftSettings: [
                .interoperabilityMode(.Cxx)  // CRITICAL: Enable C++ interop
            ]
        ),
    ],
    cxxLanguageStandard: .cxx17  // Match C++ generator standard
)
```

### 2. Module Map (module.modulemap)
**Create** `Sources/ffire_swift_complex/include/module.modulemap`:
```
module ffire_cpp_complex {
    header "generated.hpp"
    export *
}
```

### 3. Import in Swift Code
```swift
import ffire_cpp_complex  // Import C++ module directly
```

### 4. Use C++ Types Directly

#### Current (with copying):
```swift
public func decode(from data: Data) throws -> Plugin {
    return data.withUnsafeBytes { (ptr: UnsafeRawBufferPointer) -> Plugin in
        let cPlugin = decode_plugin_message(ptr.baseAddress, data.count)
        // ... copy fields from OpaquePointer to Swift struct
        c_plugin_free(cPlugin)
    }
}
```

#### With C++ Interop (ZERO copying):
```swift
import ffire_cpp_complex

public struct Plugin {
    private var cppPlugin: test.Plugin  // Direct C++ type!
    
    public init() {
        self.cppPlugin = test.Plugin()
    }
    
    public func encode() -> Data {
        var encoder = test.Encoder()
        test.encode_plugin_message(cppPlugin, &encoder)
        // std::vector<uint8_t> becomes Swift collection automatically
        return Data(encoder.buffer)
    }
    
    public static func decode(from data: Data) throws -> Plugin {
        var plugin = Plugin()
        let bytes = [UInt8](data)
        bytes.withUnsafeBufferPointer { ptr in
            plugin.cppPlugin = test.decode_plugin_message(ptr.baseAddress!, bytes.count)
        }
        return plugin
    }
    
    // Properties access C++ directly
    public var name: String {
        get { String(cppPlugin.Name) }  // std::string → Swift String
        set { cppPlugin.Name = std.string(newValue) }
    }
}
```

### 5. C++ API Structure (from generated.hpp)
```cpp
namespace test {
    struct Plugin { /* fields */ };
    struct Parameter { /* fields */ };
    
    // FREE FUNCTIONS (not methods!)
    std::vector<uint8_t> encode_plugin_message(const Plugin& msg);
    Plugin decode_plugin_message(const uint8_t* data, size_t len);
    
    // OR with encoder pattern:
    class Encoder {
        std::vector<uint8_t> buffer;
        // methods...
    };
    void encode_plugin_message(const Plugin& msg, Encoder& enc);
}
```

## Implementation Approach

### Option A: Wrap C++ structs directly
```swift
// Expose C++ struct fields as computed properties
extension test.Plugin {
    public var swiftName: String {
        get { String(Name) }
        set { Name = std.string(newValue) }
    }
}
```

### Option B: Swift wrapper with C++ storage
```swift
public struct Plugin {
    internal var cpp: test.Plugin  // Store C++ struct directly
    
    public init() {
        self.cpp = test.Plugin()
    }
    
    public var name: String {
        get { String(cpp.Name) }
        set { cpp.Name = std.string(newValue) }
    }
}
```

## What Gets Generated

### Instead of generating:
1. C ABI functions with @_silgen_name
2. OpaquePointer wrappers
3. Data copying code
4. C pointer management

### Generate:
1. `import ffire_cpp_complex` statement
2. Swift structs that store `test.Plugin` directly
3. Computed properties that access C++ fields
4. encode/decode that call C++ free functions directly
5. module.modulemap to expose C++ headers
6. Package.swift with .interoperabilityMode(.Cxx)

## Performance Impact
- **Current**: Swift Data → copy to C buffer → decode → copy to OpaquePointer → copy to Swift struct
- **With interop**: Swift [UInt8] → direct pointer to C++ decode → store C++ struct directly
- **Expected improvement**: 11.15 μs → ~8 μs (match C++ performance)

## Compatibility
- Swift 5.9+ only (released Sept 2023)
- macOS 13+ / iOS 16+ / tvOS 16+ / watchOS 9+
- All Apple platforms support this
- CI/CD may need toolchain updates

## Implementation Steps

1. **Update Package.swift template** in generator_swift.go
   - swift-tools-version:5.9
   - Updated platform requirements
   - Add swiftSettings: [.interoperabilityMode(.Cxx)]
   - Add cxxSettings with headerSearchPath

2. **Generate module.modulemap**
   - New function to create modulemap
   - Reference C++ generated.hpp

3. **Rewrite Swift wrapper generation**
   - Remove all @_silgen_name declarations
   - Remove OpaquePointer usage
   - Add `import ffire_cpp_complex` (use schema package name)
   - Generate structs that wrap `test.Plugin` (use schema namespace)
   - Generate computed properties accessing C++ fields
   - Generate encode/decode calling C++ free functions directly

4. **Type mappings**:
   - `std::string` ↔ `String` (automatic)
   - `std::vector<T>` ↔ Array-like access (automatic)
   - `std::optional<T>` ↔ `Optional<T>` (automatic)
   - C++ structs stored directly in Swift structs

5. **No C library needed**
   - Remove C library build from mage
   - Remove .linkedLibrary() from Package.swift
   - Swift directly uses C++ generated code

## Key Swift C++ Interop Features We Use

1. **Direct import**: `import ffire_cpp_complex` brings C++ namespace
2. **Value types**: C++ structs become Swift value types automatically
3. **std::vector**: Conforms to RandomAccessCollection in Swift
4. **std::string**: Convertible to/from Swift String
5. **Free functions**: Callable as `test.encode_plugin_message(...)`
6. **Nested types**: `test.Plugin`, `test.Parameter` accessible directly

## What We DON'T Need
- ❌ C ABI layer (no more dlopen/dlsym)
- ❌ OpaquePointer wrappers
- ❌ Data copying between languages
- ❌ Separate C library compilation
- ❌ Manual memory management (retain/release)
- ❌ Translation functions between C and Swift types

## What We GET
- ✅ Direct C++ struct storage in Swift
- ✅ Zero-copy access to C++ data
- ✅ Automatic std::vector/string conversions
- ✅ Type-safe Swift API
- ✅ Performance matching C++
- ✅ Simpler generated code
