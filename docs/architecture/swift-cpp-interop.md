# Swift + C++ Interoperability Complete Guide

## Overview

Swift has native C++ interoperability starting from Swift 5.9, allowing bidirectional use of C++ APIs from Swift and Swift APIs from C++. The integration is automatic with no bridging layer needed, and there's no overhead when calling C++ APIs in Swift or vice versa.

## Performance Characteristics

- **Zero overhead**: Direct and native API calls with no performance penalty
- **Automatic memory management**: C++ destructors are invoked automatically when values are no longer used in Swift, and Swift invokes the underlying copy constructor when copying C++ values
- **Value semantics**: C++ structs/classes are imported as value types (except types without copy constructors)

## Ways to Include C++ Code in Swift

### 1. **Direct Header Import with Module Maps** (Recommended)

**File Structure:**
```
MyProject/
├── Sources/
│   ├── CxxModule/
│   │   ├── include/
│   │   │   ├── module.modulemap
│   │   │   └── MyLib.hpp
│   │   └── MyLib.cpp
│   └── MySwiftApp/
│       └── main.swift
└── Package.swift
```

**module.modulemap:**
```c
module MyCxxModule {
    header "MyLib.hpp"
    requires cplusplus
    export *
}
```

**MyLib.hpp:**
```cpp
#pragma once
#include <vector>
#include <string>

namespace MyLib {
    class Calculator {
    public:
        int add(int a, int b);
        std::vector<int> getNumbers();
    };
}
```

**Swift Usage:**
```swift
import MyCxxModule

let calc = MyLib.Calculator()
let result = calc.add(5, 3)
```

### 2. **Swift Package Manager Integration**

**Package.swift:**
```swift
import PackageDescription

let package = Package(
    name: "CxxInterop",
    platforms: [.macOS(.v12), .iOS(.v15)],
    products: [
        .library(name: "CxxLib", targets: ["CxxLib"]),
        .executable(name: "MyApp", targets: ["MyApp"])
    ],
    targets: [
        // C++ target
        .target(
            name: "CxxLib",
            dependencies: [],
            path: "Sources/CxxLib",
            sources: ["MyLib.cpp"],
            publicHeadersPath: "include"
        ),
        // Swift target with C++ interop
        .executableTarget(
            name: "MyApp",
            dependencies: ["CxxLib"],
            swiftSettings: [
                .unsafeFlags([
                    "-I", "Sources/CxxLib/include",
                    "-cxx-interoperability-mode=default"
                ])
            ]
        )
    ]
)
```

### 3. **Xcode Project Setup**

**Steps:**

1. **Enable C++ Interoperability:**
   - In Xcode 15+, change "C++ and Objective-C Interoperability" setting from "C and Objective-C" to "C++ and Objective-C++"
   - Or manually add to Build Settings → Other Swift Flags: `-cxx-interoperability-mode=default`

2. **Add C++ Files:**
   - Add `.cpp` and `.hpp` files to your project
   - Xcode will ask to create a bridging header - agree to that

3. **Create Module Map (for frameworks):**
   ```
   MyFramework/
   ├── Headers/
   │   ├── module.modulemap
   │   └── MyFramework.hpp
   └── Sources/
       └── MyFramework.cpp
   ```

### 4. **CMake Integration**

**CMakeLists.txt:**
```cmake
cmake_minimum_required(VERSION 3.26)
project(MyCxxSwiftApp)

# C++ library
add_library(MyCxxLib STATIC
    src/MyLib.cpp
)
target_include_directories(MyCxxLib PUBLIC include)

# Swift executable
add_executable(MySwiftApp
    swift/main.swift
)

# Link C++ library to Swift
target_link_libraries(MySwiftApp PRIVATE MyCxxLib)

# Enable C++ interop
target_compile_options(MySwiftApp PRIVATE
    -cxx-interoperability-mode=default
)
```

## Supported Features

### C++ → Swift

Swift can import and use the following C++ features:

**Fully Supported:**
- Functions and methods
- Enums (including `enum class`)
- Structs and classes (as value types)
- Templates (class template specializations)
- Constructors (including implicit)
- Destructors (automatic invocation)
- Copy constructors
- Operator overloading (some limitations)
- Namespaces
- Standard library types (`std::vector`, `std::string`, etc.)

**Limited/Not Supported:**
- Virtual (dynamic) methods do not integrate with Swift yet
- C++ exceptions cannot be caught in Swift - uncaught exceptions cause program termination
- Inline namespaces have issues

### Swift → C++

Swift can expose the following to C++ via generated headers:

**Supported:**
- Functions
- Structs and enums
- Instance methods on structs/enums
- Class methods (partial support)
- Properties (via getters/setters)
- Swift collections (`Array`, `Dictionary`, etc.)

**Not Yet Supported:**
- Full class virtual dispatch
- Swift protocol conformances
- Generic Swift types

## Advanced Features

### 1. **Swift Annotations for Better C++ APIs**

**In C++ headers, use Swift bridging attributes:**

```cpp
#include <swift/bridging>

class ImageProcessor {
    std::vector<Image> _images;
    
public:
    // Map getter/setter to Swift computed property
    SWIFT_COMPUTED_PROPERTY
    const std::vector<Image>& getImages() const { return _images; }
    
    SWIFT_COMPUTED_PROPERTY
    void setImages(const std::vector<Image>& imgs) { _images = imgs; }
};
```

**Swift sees this as:**
```swift
let processor = ImageProcessor()
processor.images = [img1, img2]  // Natural Swift property syntax
```

### 2. **Extending C++ Types in Swift**

C++ types can be extended in Swift to conform to Swift protocols:

```swift
import CxxStdlib

extension std.vector: RandomAccessCollection {
    public var startIndex: Int { 0 }
    public var endIndex: Int { size() }
}

// Now you can use Swift collection methods
let vec = std.vector<Int32>()
vec.push_back(1)
vec.push_back(2)
vec.forEach { print($0) }
```

### 3. **Memory Management**

**C++ object lifetimes are managed automatically:**

```swift
func useTemporary() {
    let obj = CppClass()  // Constructor called
    obj.doWork()
    // Destructor automatically called when obj goes out of scope
}
```

**For shared ownership, use `std::shared_ptr`:**

```cpp
std::shared_ptr<Resource> createResource();
```

```swift
let resource = createResource()  // Reference counted
// Automatically deallocated when last reference drops
```

## Limitations & Constraints

### Current Limitations

1. C++ interoperability requires all dependencies in Swift Package Manager to also enable C++ interoperability
2. For-in loops over C++ containers may make deep copies - no performance guarantees yet
3. C++ code must be built with compatible compilers: Clang on Apple platforms/Linux, MSVC on Windows
4. C++ standard library types are not automatically bridged to Swift native types (std::string ≠ String)

### Platform Support

C++ interoperability is supported on all Swift platforms: macOS, iOS, tvOS, watchOS, Linux, and Windows

**C++ Standard Libraries by Platform:**
- **Apple platforms**: libc++
- **Linux**: libstdc++ (default) or libc++ (with `-Xcc -stdlib=libc++`)
- **Windows**: MSVC STL

## Complete Example: Cross-Platform Math Library

**Directory Structure:**
```
MathLib/
├── include/
│   ├── module.modulemap
│   └── Math.hpp
├── src/
│   └── Math.cpp
├── swift/
│   └── main.swift
└── Package.swift
```

**include/module.modulemap:**
```c
module MathLib {
    header "Math.hpp"
    requires cplusplus
    export *
}
```

**include/Math.hpp:**
```cpp
#pragma once
#include <vector>
#include <cmath>

namespace Math {
    class Statistics {
    public:
        static double mean(const std::vector<double>& values);
        static double stddev(const std::vector<double>& values);
    };
    
    double factorial(int n);
}
```

**src/Math.cpp:**
```cpp
#include "Math.hpp"
#include <numeric>

double Math::Statistics::mean(const std::vector<double>& values) {
    return std::accumulate(values.begin(), values.end(), 0.0) / values.size();
}

double Math::Statistics::stddev(const std::vector<double>& values) {
    double m = mean(values);
    double variance = 0.0;
    for (double v : values) {
        variance += (v - m) * (v - m);
    }
    return std::sqrt(variance / values.size());
}

double Math::factorial(int n) {
    return n <= 1 ? 1 : n * factorial(n - 1);
}
```

**swift/main.swift:**
```swift
import MathLib
import CxxStdlib

// Create C++ vector
var data = std.vector<Double>()
data.push_back(1.0)
data.push_back(2.0)
data.push_back(3.0)
data.push_back(4.0)
data.push_back(5.0)

// Call C++ static methods
let average = Math.Statistics.mean(data)
let stdDev = Math.Statistics.stddev(data)

print("Mean: \(average)")
print("StdDev: \(stdDev)")

// Call C++ function
let fact = Math.factorial(5)
print("5! = \(fact)")
```

**Package.swift:**
```swift
// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "MathLib",
    platforms: [.macOS(.v13)],
    products: [
        .executable(name: "MathDemo", targets: ["MathDemo"])
    ],
    targets: [
        .target(
            name: "MathLibCxx",
            path: ".",
            sources: ["src/Math.cpp"],
            publicHeadersPath: "include"
        ),
        .executableTarget(
            name: "MathDemo",
            dependencies: ["MathLibCxx"],
            path: "swift",
            swiftSettings: [
                .interoperabilityMode(.Cxx)
            ]
        )
    ],
    cxxLanguageStandard: .cxx17
)
```

**Build and Run:**
```bash
swift build
swift run MathDemo
```

## Best Practices

1. **Use module maps** for clean separation and reusability
2. **Prefer value semantics** - let Swift manage C++ object lifetimes
3. **Handle exceptions in C++** before they reach Swift
4. **Use Swift bridging annotations** to make C++ APIs feel native
5. **Keep C++ headers minimal** - only expose what Swift needs
6. **Use `std::shared_ptr`** for shared ownership across languages
7. **Test incrementally** - start small and add complexity gradually

## Resources

- Official Swift C++ Interop Documentation: https://www.swift.org/documentation/cxx-interop/
- Status and Supported Features: https://www.swift.org/documentation/cxx-interop/status/
- WWDC23 Session "Mix Swift and C++": https://developer.apple.com/videos/play/wwdc2023/10172/
