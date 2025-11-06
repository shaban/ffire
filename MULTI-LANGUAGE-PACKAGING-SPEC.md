# FFire Multi-Language Packaging Specification

**Version:** 1.0-draft  
**Date:** November 6, 2025  
**Status:** Design & Review Phase

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [CLI Design](#cli-design)
4. [Output Structure](#output-structure)
5. [Language-Specific Specs](#language-specific-specs)
6. [Build & Distribution](#build--distribution)
7. [Potential Issues & Mitigations](#potential-issues--mitigations)
8. [Implementation Phases](#implementation-phases)
9. [Testing Strategy](#testing-strategy)
10. [Documentation Requirements](#documentation-requirements)

---

## 1. Overview

### Goal
Enable FFire to generate production-ready packages for 24+ programming languages from a single `.ffi` schema, using a unified C ABI dynamic library approach.

### Key Principles
- **Single Source of Truth**: One `.ffi` schema
- **Single Binary**: One `libffire.{dylib,so,dll}` for all languages
- **Zero Manual Work**: Generated packages are ready to publish
- **Language Idiomatic**: Each package follows ecosystem conventions
- **Cross-Platform**: macOS, Linux, Windows support

### User Experience
```bash
# Generate native code (Tier A)
ffire generate -lang cpp -schema audio.ffi -out ./dist

# Generate complete package (Tier B)
ffire generate -lang python -schema audio.ffi -out ./dist
cd dist/python && pip install .  # Works immediately!
```

---

## 2. Architecture

### System Components

```
┌─────────────────────────────────────────────────────────┐
│                    FFire CLI                            │
│  (ffire generate -lang <lang> -schema <file>)          │
└────────────────┬────────────────────────────────────────┘
                 │
                 ├──> Schema Parser (.ffi → AST)
                 │
                 ├──> C++ Code Generator (generated.hpp)
                 │
                 ├──> C ABI Generator (generated_c.cpp/h)
                 │
                 ├──> Dylib Compiler (libffire.dylib/so/dll)
                 │
                 └──> Language Package Generator
                      │
                      ├──> Tier A: Code Only (C, C++, Rust, etc.)
                      └──> Tier B: Full Package (Python, Swift, etc.)

                 Output: Ready-to-use package
```

### Data Flow

```
.ffi schema
    ↓
AST (internal representation)
    ↓
├─> C++ code (generated.hpp)
│   ↓
│   C ABI wrapper (generated_c.cpp/h)
│   ↓
│   Compiled dylib (libffire.dylib/so/dll)
│
└─> Language-specific wrapper
    ↓
    Complete package (with metadata, docs, examples)
```

---

## 3. CLI Design

### Command Structure

```bash
ffire generate [OPTIONS] -lang <language> -schema <schema.ffi>
```

### Options

| Option | Short | Type | Default | Description |
|--------|-------|------|---------|-------------|
| `--language` | `-lang` | string | required | Target language (cpp, python, swift, etc.) |
| `--schema` | `-schema` | path | required | Input .ffi schema file |
| `--output` | `-out` | path | `./dist` | Output directory |
| `--optimize` | `-O` | int | `2` | Optimization level (0-3) |
| `--platform` | `-p` | string | current | Target platform (darwin, linux, windows, all) |
| `--arch` | `-a` | string | current | Target architecture (arm64, x86_64, all) |
| `--namespace` | `-ns` | string | schema name | C++ namespace / package name |
| `--no-compile` | | flag | false | Skip dylib compilation (for testing) |
| `--verbose` | `-v` | flag | false | Verbose output |

### Examples

```bash
# Basic usage
ffire generate -lang python -schema audio.ffi

# Multi-platform
ffire generate -lang python -schema audio.ffi -platform all

# Custom output
ffire generate -lang swift -schema audio.ffi -out ./packages/swift

# Skip compilation (template testing)
ffire generate -lang ruby -schema test.ffi --no-compile

# Verbose mode
ffire generate -lang javascript -schema audio.ffi -v
```

### Output Messages

**Tier A (Native) Output:**
```
✓ Generated C++ code: dist/cpp/generated.hpp
✓ Generated C ABI:    dist/cpp/generated_c.cpp
✓ Compiled library:   dist/cpp/libffire.dylib
✓ Created examples:   dist/cpp/examples/

Usage:
  #include "generated.hpp"
  Link against: libffire.dylib

See dist/cpp/README.md for details.
```

**Tier B (Wrapper) Output:**
```
✓ Generated C++ code:   dist/python/src/generated.hpp
✓ Generated C ABI:      dist/python/src/generated_c.cpp
✓ Compiled library:     dist/python/ffire/libffire.dylib
✓ Generated wrapper:    dist/python/ffire/__init__.py
✓ Created package:      dist/python/setup.py
✓ Created examples:     dist/python/examples/example.py

Package ready! Install with:
  cd dist/python && pip install .

Or publish to PyPI:
  cd dist/python && python -m build && twine upload dist/*
```

---

## 4. Output Structure

### Universal Distribution (Multi-Platform)

When generating with `-platform all`:

```
dist/
├── ffire-<schema>-<version>/
│   ├── README.md                    # Universal guide
│   ├── LICENSE                      # MIT or chosen license
│   ├── lib/                         # Compiled binaries
│   │   ├── darwin-arm64/
│   │   │   └── libffire.dylib
│   │   ├── darwin-x86_64/
│   │   │   └── libffire.dylib
│   │   ├── linux-x86_64/
│   │   │   └── libffire.so
│   │   ├── linux-arm64/
│   │   │   └── libffire.so
│   │   └── windows-x64/
│   │       └── ffire.dll
│   ├── include/
│   │   ├── generated.hpp            # C++ header
│   │   ├── generated_c.h            # C ABI header
│   │   └── ffire_common.h           # Common types/macros
│   ├── src/
│   │   ├── generated.hpp            # C++ implementation
│   │   └── generated_c.cpp          # C ABI implementation
│   └── packages/
│       ├── python/
│       ├── swift/
│       ├── javascript/
│       └── ... (all Tier B languages)
```

### Single-Language Package (Tier A: C++)

```
dist/cpp/
├── README.md
├── include/
│   ├── generated.hpp
│   └── generated_c.h
├── lib/
│   └── libffire.dylib
├── examples/
│   ├── basic_usage.cpp
│   └── Makefile
└── tests/
    └── test_generated.cpp
```

### Single-Language Package (Tier B: Python)

```
dist/python/
├── README.md
├── setup.py                         # or pyproject.toml
├── MANIFEST.in
├── ffire/
│   ├── __init__.py                  # Main wrapper
│   ├── _native.py                   # ctypes bindings
│   ├── types.py                     # Python type hints
│   ├── lib/
│   │   ├── darwin-arm64/
│   │   │   └── libffire.dylib
│   │   ├── linux-x86_64/
│   │   │   └── libffire.so
│   │   └── windows-x64/
│   │       └── ffire.dll
│   └── py.typed                     # PEP 561 marker
├── examples/
│   ├── basic_usage.py
│   └── advanced_example.py
├── tests/
│   └── test_ffire.py
└── src/                             # C++ source (for reference)
    ├── generated.hpp
    └── generated_c.cpp
```

---

## 5. Language-Specific Specs

### 5.1 Python Package Spec

**Package Name:** `ffire-<schema>`  
**Distribution:** PyPI wheel + source  
**Files Required:**

```python
# setup.py
from setuptools import setup, find_packages

setup(
    name="ffire-{schema}",
    version="{version}",
    packages=find_packages(),
    package_data={
        "ffire": [
            "lib/darwin-arm64/*.dylib",
            "lib/darwin-x86_64/*.dylib",
            "lib/linux-x86_64/*.so",
            "lib/linux-arm64/*.so",
            "lib/windows-x64/*.dll",
        ]
    },
    install_requires=[],
    python_requires=">=3.7",
)
```

**API Design:**
```python
# ffire/__init__.py
from typing import List, Optional

class {MessageType}:
    def __init__(self, data: bytes):
        """Decode from binary data"""
        
    @property
    def field_name(self) -> str:
        """Access field"""
    
    def encode(self) -> bytes:
        """Encode to binary"""

def decode(data: bytes) -> {MessageType}:
    """Convenience function"""

def encode(obj: {MessageType}) -> bytes:
    """Convenience function"""
```

**Platform Detection:**
```python
# ffire/_native.py
import sys
import platform
from pathlib import Path

def _get_lib_path():
    system = platform.system().lower()
    machine = platform.machine().lower()
    
    if system == "darwin":
        arch = "arm64" if machine == "arm64" else "x86_64"
        lib_name = "libffire.dylib"
    elif system == "linux":
        arch = "x86_64" if machine in ["x86_64", "amd64"] else "arm64"
        lib_name = "libffire.so"
    elif system == "windows":
        arch = "x64"
        lib_name = "ffire.dll"
    else:
        raise RuntimeError(f"Unsupported platform: {system}")
    
    lib_dir = Path(__file__).parent / "lib" / f"{system}-{arch}"
    return lib_dir / lib_name
```

---

### 5.2 Swift Package Spec

**Package Name:** `ffire-{schema}`  
**Distribution:** Swift Package Manager  
**Files Required:**

```swift
// Package.swift
// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "FFire{Schema}",
    platforms: [.macOS(.v13), .iOS(.v16)],
    products: [
        .library(name: "FFire{Schema}", targets: ["FFire{Schema}"])
    ],
    targets: [
        .binaryTarget(
            name: "libffire",
            path: "lib/libffire.xcframework"  // For multi-platform
        ),
        .target(
            name: "FFire{Schema}",
            dependencies: ["libffire"],
            path: "Sources",
            swiftSettings: [.interoperabilityMode(.Cxx)]
        )
    ]
)
```

**XCFramework Creation:**
```bash
# Create XCFramework for iOS + macOS + Simulator
xcodebuild -create-xcframework \
  -library lib/darwin-arm64/libffire.dylib \
  -library lib/darwin-x86_64/libffire.dylib \
  -library lib/ios-arm64/libffire.dylib \
  -library lib/ios-x86_64-simulator/libffire.dylib \
  -output lib/libffire.xcframework
```

**API Design:**
```swift
// Sources/FFire{Schema}/FFire.swift
import Foundation

public class {MessageType} {
    private let handle: OpaquePointer
    
    public init(data: Data) throws {
        // Decode using C ABI
    }
    
    public var fieldName: String {
        // Getter via C ABI
    }
    
    public func encode() -> Data {
        // Encode via C ABI
    }
    
    deinit {
        // Free handle
    }
}

public func decode(_ data: Data) throws -> {MessageType} {
    try {MessageType}(data: data)
}
```

---

### 5.3 JavaScript/Node.js Package Spec

**Package Name:** `@ffire/{schema}`  
**Distribution:** npm  
**Files Required:**

```json
// package.json
{
  "name": "@ffire/{schema}",
  "version": "{version}",
  "main": "index.js",
  "types": "index.d.ts",
  "engines": {
    "node": ">=14.0.0"
  },
  "dependencies": {
    "ffi-napi": "^4.0.3",
    "ref-napi": "^3.0.3"
  },
  "files": [
    "index.js",
    "index.d.ts",
    "lib/**/*"
  ]
}
```

**API Design:**
```javascript
// index.js
const ffi = require('ffi-napi');
const ref = require('ref-napi');
const path = require('path');

// Platform detection
function getLibPath() {
  const platform = process.platform;
  const arch = process.arch;
  
  let libName;
  if (platform === 'darwin') libName = 'libffire.dylib';
  else if (platform === 'linux') libName = 'libffire.so';
  else if (platform === 'win32') libName = 'ffire.dll';
  else throw new Error(`Unsupported platform: ${platform}`);
  
  return path.join(__dirname, 'lib', `${platform}-${arch}`, libName);
}

// FFI declarations
const lib = ffi.Library(getLibPath(), {
  'plugin_decode': ['pointer', ['pointer', 'size_t', 'pointer']],
  'plugin_encode': ['size_t', ['pointer', 'pointer', 'pointer']],
  'plugin_free': ['void', ['pointer']],
  'plugin_free_data': ['void', ['pointer']]
});

class {MessageType} {
  constructor(buffer) {
    // Decode
  }
  
  get fieldName() {
    // Getter
  }
  
  encode() {
    // Encode
  }
}

module.exports = { {MessageType}, decode, encode };
```

**TypeScript Definitions:**
```typescript
// index.d.ts
export class {MessageType} {
  constructor(buffer: Buffer);
  readonly fieldName: string;
  encode(): Buffer;
}

export function decode(buffer: Buffer): {MessageType};
export function encode(obj: {MessageType}): Buffer;
```

---

### 5.4 Ruby Gem Spec

**Package Name:** `ffire-{schema}`  
**Distribution:** RubyGems  
**Files Required:**

```ruby
# ffire.gemspec
Gem::Specification.new do |spec|
  spec.name          = "ffire-{schema}"
  spec.version       = "{version}"
  spec.authors       = ["FFire"]
  spec.summary       = "FFire bindings for {schema}"
  spec.files         = Dir["lib/**/*", "README.md"]
  spec.require_paths = ["lib"]
  spec.add_dependency "ffi", "~> 1.15"
  spec.required_ruby_version = ">= 2.7"
end
```

**API Design:**
```ruby
# lib/ffire.rb
require 'ffi'

module FFire
  extend FFI::Library
  
  # Platform detection
  def self.lib_path
    platform = case RbConfig::CONFIG['host_os']
    when /darwin/ then 'darwin'
    when /linux/ then 'linux'
    when /mswin|mingw/ then 'windows'
    else raise "Unsupported platform"
    end
    
    arch = RbConfig::CONFIG['host_cpu']
    arch = 'arm64' if arch =~ /arm|aarch64/
    arch = 'x86_64' if arch =~ /x86_64|amd64/
    
    ext = platform == 'darwin' ? 'dylib' : (platform == 'windows' ? 'dll' : 'so')
    File.join(__dir__, 'ffire', 'lib', "#{platform}-#{arch}", "libffire.#{ext}")
  end
  
  ffi_lib lib_path
  
  # FFI declarations
  attach_function :plugin_decode, [:pointer, :size_t, :pointer], :pointer
  attach_function :plugin_encode, [:pointer, :pointer, :pointer], :size_t
  attach_function :plugin_free, [:pointer], :void
  
  class {MessageType}
    def initialize(data)
      # Decode
    end
    
    def field_name
      # Getter
    end
    
    def encode
      # Encode
    end
  end
end
```

---

### 5.5 Java/Maven Package Spec

**Package Name:** `com.ffire:{schema}`  
**Distribution:** Maven Central  
**Files Required:**

```xml
<!-- pom.xml -->
<project>
  <groupId>com.ffire</groupId>
  <artifactId>{schema}</artifactId>
  <version>{version}</version>
  
  <dependencies>
    <dependency>
      <groupId>net.java.dev.jna</groupId>
      <artifactId>jna</artifactId>
      <version>5.13.0</version>
    </dependency>
  </dependencies>
  
  <build>
    <resources>
      <resource>
        <directory>src/main/resources</directory>
        <includes>
          <include>**/*.dylib</include>
          <include>**/*.so</include>
          <include>**/*.dll</include>
        </includes>
      </resource>
    </resources>
  </build>
</project>
```

**API Design:**
```java
// src/main/java/com/ffire/{schema}/{MessageType}.java
package com.ffire.{schema};

import com.sun.jna.*;
import java.nio.file.*;

public class {MessageType} implements AutoCloseable {
    
    private static final FFireLibrary LIB = loadLibrary();
    
    private static FFireLibrary loadLibrary() {
        String os = System.getProperty("os.name").toLowerCase();
        String arch = System.getProperty("os.arch");
        
        String platform = os.contains("mac") ? "darwin" :
                          os.contains("linux") ? "linux" : "windows";
        String archStr = arch.contains("aarch64") || arch.contains("arm") ? "arm64" : "x86_64";
        
        String libName = platform.equals("windows") ? "ffire.dll" : "libffire." + 
                         (platform.equals("darwin") ? "dylib" : "so");
        
        String libPath = String.format("/lib/%s-%s/%s", platform, archStr, libName);
        
        // Extract from JAR to temp location
        // ...
        
        return Native.load(extractedLibPath, FFireLibrary.class);
    }
    
    private Pointer handle;
    
    public {MessageType}(byte[] data) {
        // Decode using JNA
    }
    
    public String getFieldName() {
        // Getter via JNA
    }
    
    public byte[] encode() {
        // Encode via JNA
    }
    
    @Override
    public void close() {
        if (handle != null) {
            LIB.plugin_free(handle);
            handle = null;
        }
    }
    
    interface FFireLibrary extends Library {
        Pointer plugin_decode(byte[] data, long size, PointerByReference error);
        long plugin_encode(Pointer handle, PointerByReference outData, PointerByReference error);
        void plugin_free(Pointer handle);
        void plugin_free_data(Pointer data);
    }
}
```

---

## 6. Build & Distribution

### 6.1 Cross-Platform Compilation

**Makefile for Multi-Platform:**
```makefile
CXX = clang++
CXXFLAGS = -std=c++17 -O2 -Wall -Wextra

# Detect platform
UNAME := $(shell uname -s)
ARCH := $(shell uname -m)

ifeq ($(UNAME),Darwin)
    PLATFORM = darwin
    LIB_EXT = dylib
    SHARED_FLAGS = -dynamiclib
else ifeq ($(UNAME),Linux)
    PLATFORM = linux
    LIB_EXT = so
    SHARED_FLAGS = -shared -fPIC
else
    PLATFORM = windows
    LIB_EXT = dll
    SHARED_FLAGS = -shared
endif

# Output directory
OUT_DIR = lib/$(PLATFORM)-$(ARCH)
LIB = $(OUT_DIR)/libffire.$(LIB_EXT)

all: $(LIB)

$(OUT_DIR):
	mkdir -p $(OUT_DIR)

$(LIB): generated_c.cpp generated.hpp | $(OUT_DIR)
	$(CXX) $(CXXFLAGS) $(SHARED_FLAGS) -o $(LIB) generated_c.cpp

clean:
	rm -rf lib/
```

**Cross-Compilation (Linux → Windows):**
```bash
# Install mingw-w64
brew install mingw-w64

# Compile for Windows
x86_64-w64-mingw32-g++ -std=c++17 -O2 -shared \
  -o lib/windows-x64/ffire.dll \
  generated_c.cpp
```

**Docker for Linux Builds:**
```dockerfile
# Dockerfile
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    build-essential \
    clang \
    cmake

WORKDIR /build
COPY . .

RUN make PLATFORM=linux ARCH=x86_64
RUN make PLATFORM=linux ARCH=arm64
```

---

### 6.2 Platform-Specific Considerations

**macOS:**
- Universal binary (arm64 + x86_64): `lipo -create -output libffire.dylib libffire_arm64.dylib libffire_x86_64.dylib`
- Code signing may be required for distribution
- XCFramework for Swift/iOS

**Linux:**
- Build for x86_64 and arm64
- Use oldest supported glibc (e.g., Ubuntu 20.04 LTS)
- Ship with RPATH for relative library loading

**Windows:**
- MSVC or MinGW-w64
- Include vcruntime140.dll or statically link CRT
- Handle DLL search paths

---

### 6.3 Version Management

**Semantic Versioning:**
```
{major}.{minor}.{patch}

Example: ffire-audio-1.2.3
```

**Version in Code:**
```cpp
// generated.hpp
#define FFIRE_VERSION_MAJOR 1
#define FFIRE_VERSION_MINOR 2
#define FFIRE_VERSION_PATCH 3
#define FFIRE_VERSION "1.2.3"
```

**ABI Compatibility:**
- Major version bump: Breaking changes
- Minor version bump: Backward-compatible additions
- Patch version bump: Bug fixes only

---

## 7. Potential Issues & Mitigations

### 7.1 Platform Detection Issues

**Problem:** Runtime platform detection fails  
**Mitigation:**
- Fallback to multiple detection methods
- Allow manual library path override via environment variable
- Ship platform-specific installers if needed

```python
# Python example
import os

lib_path = os.environ.get('FFIRE_LIB_PATH') or _detect_platform()
```

---

### 7.2 Symbol Conflicts

**Problem:** Multiple packages using different ffire versions  
**Mitigation:**
- Version namespace in C ABI: `ffire_v1_plugin_decode`
- Hide symbols with visibility attributes
- Document that only one version can be loaded per process

```cpp
// generated_c.h
#define FFIRE_API __attribute__((visibility("default")))

extern "C" {
    FFIRE_API PluginHandle ffire_v1_plugin_decode(...);
}
```

---

### 7.3 Memory Management

**Problem:** Memory leaks across language boundary  
**Mitigation:**
- Clear ownership rules in docs
- RAII wrappers for Tier A languages
- Finalizers/destructors for Tier B languages
- Comprehensive memory tests

```python
# Python with context manager
with ffire.decode(data) as plugin:
    print(plugin.name)
# Automatically freed
```

---

### 7.4 ABI Stability

**Problem:** C++ ABI changes break compatibility  
**Mitigation:**
- Only expose C ABI (extern "C")
- Opaque handles only
- Never expose C++ types in C API
- Version check at runtime

```cpp
extern "C" {
    // Good: Opaque handle
    PluginHandle plugin_decode(...);
    
    // Bad: Exposed C++ type
    std::vector<Plugin> plugin_decode_bad(...);  // DON'T DO THIS
}
```

---

### 7.5 Dependency Hell

**Problem:** Language package managers conflict with system libs  
**Mitigation:**
- Bundle libffire in each package (isolated)
- Use relative RPATH / @rpath
- No external C++ dependencies (self-contained)

```bash
# Set RPATH on macOS
install_name_tool -add_rpath @loader_path/lib libffire.dylib

# Set RPATH on Linux
patchelf --set-rpath '$ORIGIN/lib' libffire.so
```

---

### 7.6 Large Binary Size

**Problem:** Bundling dylib in every package increases size  
**Mitigation:**
- Strip debug symbols: `strip -x libffire.dylib`
- Use link-time optimization (LTO): `-flto`
- Compress in package (e.g., wheel compression)
- Consider system-wide install option for advanced users

```bash
# Before stripping
ls -lh libffire.dylib
# 2.3M

# After stripping
strip -x libffire.dylib
ls -lh libffire.dylib
# 187K (87% reduction!)
```

---

### 7.7 Thread Safety

**Problem:** Concurrent access to C ABI from multiple threads  
**Mitigation:**
- Document thread-safety guarantees
- Use thread-local storage for error messages
- Atomic operations for reference counting
- Comprehensive threading tests

```cpp
// generated_c.cpp
thread_local char* last_error = nullptr;

extern "C" const char* ffire_get_last_error() {
    return last_error;
}
```

---

### 7.8 Error Handling Across Languages

**Problem:** C++ exceptions don't cross language boundaries  
**Mitigation:**
- Catch all exceptions at C ABI boundary
- Return error codes + error messages
- Language wrappers throw native exceptions

```cpp
extern "C" PluginHandle plugin_decode(const uint8_t* data, size_t size, char** error_msg) {
    try {
        auto plugins = test::decode_plugin_message(data, size);
        // ...
        return handle;
    } catch (const std::exception& e) {
        *error_msg = strdup(e.what());
        return nullptr;
    } catch (...) {
        *error_msg = strdup("Unknown error");
        return nullptr;
    }
}
```

---

## 8. Implementation Phases

### Phase 1: Foundation (Weeks 1-2)
- ✅ **Already Done:** C++ code generation
- ✅ **Already Done:** C ABI wrapper working
- ✅ **Already Done:** Python & Swift tested
- ⬜ Finalize C ABI design
- ⬜ Multi-platform build system
- ⬜ Symbol versioning strategy

**Deliverable:** Stable C ABI + cross-platform dylib

---

### Phase 2: CLI Framework (Weeks 3-4)
- ⬜ Implement `ffire generate` command
- ⬜ Template system for language packages
- ⬜ Platform detection logic
- ⬜ Output directory structure
- ⬜ Verbose logging

**Deliverable:** CLI that generates C++, C, Rust packages

---

### Phase 3: Tier B Languages - Batch 1 (Weeks 5-6)
- ⬜ Python package generator (highest priority)
- ⬜ JavaScript/Node.js package generator
- ⬜ Swift package generator
- ⬜ Ruby package generator

**Deliverable:** 4 language packages working end-to-end

---

### Phase 4: Tier B Languages - Batch 2 (Weeks 7-8)
- ⬜ Java package generator
- ⬜ C# package generator
- ⬜ Go package generator (via cgo)
- ⬜ Dart package generator

**Deliverable:** 8 languages total

---

### Phase 5: Testing & Polish (Weeks 9-10)
- ⬜ Comprehensive test suite per language
- ⬜ Memory leak tests (valgrind, sanitizers)
- ⬜ Thread safety tests
- ⬜ Cross-platform CI/CD (GitHub Actions)
- ⬜ Benchmark suite

**Deliverable:** Production-ready quality

---

### Phase 6: Documentation (Weeks 11-12)
- ⬜ API documentation per language
- ⬜ Usage guides
- ⬜ Migration guides
- ⬜ Performance tuning docs
- ⬜ Troubleshooting guide

**Deliverable:** Complete documentation

---

### Phase 7: Release & Distribution (Week 13)
- ⬜ Package registry accounts (PyPI, npm, SPM, etc.)
- ⬜ Automated publishing pipeline
- ⬜ GitHub releases
- ⬜ Website/docs hosting

**Deliverable:** Public release

---

## 9. Testing Strategy

### 9.1 Unit Tests

**Per Language Package:**
```python
# tests/test_python.py
def test_decode():
    data = load_fixture("test.bin")
    obj = ffire.decode(data)
    assert obj.field == expected_value

def test_encode():
    obj = create_test_object()
    data = ffire.encode(obj)
    assert len(data) == expected_size

def test_round_trip():
    original = load_fixture("test.bin")
    obj = ffire.decode(original)
    encoded = ffire.encode(obj)
    assert encoded == original

def test_memory_cleanup():
    # Ensure no leaks
    for _ in range(1000):
        obj = ffire.decode(data)
        _ = ffire.encode(obj)
```

---

### 9.2 Integration Tests

**Cross-Language Interop:**
```bash
# Encode in Python, decode in Swift
python tests/encode.py > test.bin
swift run decode test.bin

# Encode in Swift, decode in JavaScript
swift run encode > test.bin
node tests/decode.js test.bin
```

---

### 9.3 Performance Tests

**Benchmarks:**
```python
# tests/benchmark.py
import time

data = load_large_fixture()
iterations = 10000

start = time.perf_counter()
for _ in range(iterations):
    obj = ffire.decode(data)
end = time.perf_counter()

print(f"Decode: {(end - start) / iterations * 1e6:.2f} µs/op")
```

---

### 9.4 Memory Tests

**Valgrind (Linux):**
```bash
valgrind --leak-check=full --show-leak-kinds=all \
  python tests/test_memory.py
```

**AddressSanitizer:**
```bash
clang++ -fsanitize=address -g -O1 generated_c.cpp -o test_asan
./test_asan
```

---

### 9.5 CI/CD Pipeline

**GitHub Actions:**
```yaml
# .github/workflows/test.yml
name: Test All Languages

on: [push, pull_request]

jobs:
  test-python:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        python-version: ['3.8', '3.9', '3.10', '3.11', '3.12']
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-python@v4
        with:
          python-version: ${{ matrix.python-version }}
      - run: cd dist/python && pip install -e .
      - run: pytest tests/

  test-swift:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3
      - run: cd dist/swift && swift test

  test-javascript:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - run: cd dist/javascript && npm install && npm test
```

---

## 10. Documentation Requirements

### 10.1 User Documentation

**README per Package:**
```markdown
# FFire {Schema} - {Language}

## Installation

[Language-specific installation instructions]

## Quick Start

```{language}
[Basic usage example]
```

## API Reference

[Generated API docs]

## Examples

- [Basic usage](examples/basic.{ext})
- [Advanced usage](examples/advanced.{ext})

## Performance

- Decode: ~X µs/op
- Encode: ~X µs/op

## Troubleshooting

[Common issues and solutions]
```

---

### 10.2 Developer Documentation

**Contributing Guide:**
- How to add a new language
- Template system explanation
- Testing requirements
- Code style guidelines

**Architecture Documentation:**
- System design
- C ABI specification
- Memory management rules
- Thread safety guarantees

---

### 10.3 Website & GitHub Pages

**Structure:**
```
docs/
├── index.md                    # Homepage
├── getting-started.md          # Quick start guide
├── languages/
│   ├── python.md
│   ├── swift.md
│   ├── javascript.md
│   └── ... (per language)
├── api/
│   └── c-abi.md               # C API reference
├── guides/
│   ├── performance.md
│   ├── troubleshooting.md
│   └── migration.md
└── blog/
    └── announcing-ffire.md    # Release post
```

**Material for MkDocs:**
```yaml
# mkdocs.yml
site_name: FFire
site_url: https://ffire.dev
theme:
  name: material
  palette:
    primary: deep orange  # Fire theme
    accent: amber
  logo: assets/logo.svg   # Flame logo
  favicon: assets/favicon.ico

nav:
  - Home: index.md
  - Getting Started: getting-started.md
  - Languages:
    - Python: languages/python.md
    - Swift: languages/swift.md
    - JavaScript: languages/javascript.md
  - API Reference: api/c-abi.md
  - Guides:
    - Performance: guides/performance.md
    - Troubleshooting: guides/troubleshooting.md
```

---

## 11. Long-Term Roadmap

### Milestone 1: Foundation ✅
- [x] Schema design (.ffi format)
- [x] C++ code generation
- [x] C ABI wrapper
- [x] Initial testing (Python, Swift)

### Milestone 2: Multi-Language Support (Current Phase)
- [ ] Finalize C++ bridging
- [ ] Complete CLI packaging tool
- [ ] Generate 8-10 language bindings
- [ ] Comprehensive testing

### Milestone 3: Optimization
- [ ] C++ performance brainstorming
  - SIMD optimizations
  - Zero-copy decoding
  - Memory pool allocator
  - Compile-time schema validation
- [ ] Benchmark against protobuf, flatbuffers, capnproto
- [ ] Profile-guided optimization

### Milestone 4: Project Polish
- [ ] Clean up obsolete exploratory code
- [ ] Remove deprecated docs
- [ ] Consistent code style across all generators
- [ ] Final API review

### Milestone 5: Documentation & Branding
- [ ] Complete user documentation
- [ ] GitHub Pages with Material theme
- [ ] Brand design (logo: "F" with flame element)
- [ ] Color scheme: Orange/red fire colors
- [ ] Example projects gallery

### Milestone 6: Public Release
- [ ] Version 1.0.0
- [ ] Publish to package registries
- [ ] Press release / blog post
- [ ] Social media announcements

### Milestone 7: Community Building
- [ ] Notify relevant communities:
  - r/programming
  - Hacker News
  - Product Hunt
  - Language-specific forums (r/python, r/rust, etc.)
  - Audio plugin developer forums
  - Game development communities
- [ ] Create Discord/Slack community
- [ ] Set up discussion forums (GitHub Discussions)
- [ ] Encourage community contributions

---

## 12. Success Metrics

### Technical Metrics
- **Performance:** <10µs encode/decode for common schemas
- **Size:** <200KB dylib (stripped)
- **Coverage:** 10+ languages with complete packages
- **Compatibility:** macOS, Linux, Windows (x86_64 + arm64)
- **Quality:** >90% test coverage, zero memory leaks

### Adoption Metrics
- **Downloads:** 1K+ in first month
- **GitHub Stars:** 100+ in first 3 months
- **Contributors:** 5+ external contributors
- **Issues:** <10 open bugs at any time
- **Documentation:** 95%+ positive feedback

---

## 13. Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| C++ ABI instability | Medium | High | Use extern "C" only |
| Platform-specific bugs | High | Medium | Comprehensive CI/CD testing |
| Memory leaks | Medium | High | Extensive memory testing |
| Poor performance | Low | High | Benchmark early and often |
| Adoption failure | Medium | High | Strong documentation, examples |
| Maintenance burden | Medium | Medium | Automate everything possible |
| Competition | Low | Medium | Focus on ease of use |

---

## 14. Decision Log

### Decision 1: C ABI vs. C++ Direct Export
**Chosen:** C ABI  
**Rationale:** Maximum language compatibility, ABI stability  
**Trade-offs:** Extra layer, slightly verbose API  

### Decision 2: Bundle dylib vs. System Install
**Chosen:** Bundle in each package  
**Rationale:** No dependency hell, works out of box  
**Trade-offs:** Larger package size, some duplication  

### Decision 3: Single Generator vs. Per-Language Tools
**Chosen:** Single generator with templates  
**Rationale:** Consistency, easier maintenance  
**Trade-offs:** More complex generator logic  

### Decision 4: Manual vs. Auto-Generated Wrappers
**Chosen:** Auto-generated from templates  
**Rationale:** Scalability, consistency  
**Trade-offs:** Less flexibility for language-specific optimizations  

---

## 15. Open Questions

1. **Schema Evolution:** How to handle breaking changes in .ffi schemas?
   - Proposal: Semantic versioning in schema + migration tools

2. **Large Messages:** Should we support streaming/chunked encoding?
   - Proposal: Phase 2 feature after initial release

3. **Code Size:** How to minimize dylib size for embedded systems?
   - Proposal: Compile-time feature flags to exclude unused code

4. **WebAssembly:** Should we support WASM as a target?
   - Proposal: Experimental support via Emscripten

5. **Plugin System:** Allow custom language generators?
   - Proposal: Plugin API in v2.0

---

## 16. Next Steps

### Immediate (This Week)
1. Review this spec with stakeholders
2. Identify any missing requirements
3. Create detailed task breakdown for Phase 2
4. Set up project board

### Short-Term (Next 2 Weeks)
1. Finalize C ABI design
2. Implement CLI framework
3. Create template system
4. Generate Python package end-to-end

### Medium-Term (Next Month)
1. Complete 4-5 language generators
2. Set up CI/CD pipeline
3. Write initial documentation
4. Start performance optimization

---

## Appendix A: Example Session

```bash
# User creates a schema
cat > audio.ffi << EOF
struct AudioPlugin {
    name: string
    version: string
    parameters: [AudioParameter]
}

struct AudioParameter {
    id: string
    value: float
    min: float
    max: float
}
EOF

# Generate Python package
ffire generate -lang python -schema audio.ffi -out ./dist

# Output
✓ Generated C++ code:   dist/python/src/generated.hpp
✓ Generated C ABI:      dist/python/src/generated_c.cpp
✓ Compiled library:     dist/python/ffire/libffire.dylib
✓ Generated wrapper:    dist/python/ffire/__init__.py
✓ Created package:      dist/python/setup.py

Package ready! Install with:
  cd dist/python && pip install .

# User installs and uses
cd dist/python
pip install .

python
>>> import ffire
>>> data = open('plugin.bin', 'rb').read()
>>> plugin = ffire.AudioPlugin(data)
>>> print(plugin.name)
'My Plugin'
>>> print(plugin.version)
'1.0.0'
>>> encoded = plugin.encode()
>>> len(encoded)
4293

# Success! 🎉
```

---

**End of Specification**

---

**Review Checklist:**
- [ ] All 14 Tier B languages specified
- [ ] Cross-platform build process defined
- [ ] Potential issues identified and mitigated
- [ ] Testing strategy comprehensive
- [ ] Documentation requirements clear
- [ ] Implementation phases realistic
- [ ] Long-term roadmap aligned with goals
- [ ] Success metrics measurable
- [ ] Risks assessed and mitigated
- [ ] Open questions documented

**Status:** Ready for review and feedback
