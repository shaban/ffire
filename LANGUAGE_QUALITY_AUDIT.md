# Language Generator Quality Audit

**Criteria for v1.0 Inclusion:**
- ✅ Sane, usable data types
- ✅ Non-atrocious performance
- ✅ Proper memory management
- ✅ Complete feature coverage (primitives, structs, arrays, optionals)

**Action:** Generators not meeting criteria should be dropped or marked experimental until proper implementation is possible.

---

## Tier 0: Native Implementation

### Go ✅
- **Status:** Reference implementation
- **Data Types:** Native Go types (`int32`, `string`, `[]T`, `*T`)
- **Performance:** Excellent (native)
- **Memory:** Automatic GC
- **Verdict:** ✅ **INCLUDE in v1.0**

---

## Tier A: Native Code + C ABI

### C++ ✅
- **Status:** Primary implementation, basis for C ABI
- **Data Types:** Modern C++17 (`std::vector`, `std::string`, `std::optional`, `int32_t`)
- **Performance:** Excellent (native, zero-copy)
- **Memory:** RAII, automatic
- **Feature Coverage:** Complete
- **Verdict:** ✅ **INCLUDE in v1.0**

### C ✅
- **Status:** C ABI header generated
- **Data Types:** Opaque handles, primitive C types
- **Performance:** Good (direct C ABI)
- **Memory:** Manual (`free` functions provided)
- **Deployment:** Bundled dylib per package (canonical)
- **Symbol Versioning:** Not needed (bundled deployment avoids conflicts)
- **Verdict:** ✅ **INCLUDE in v1.0** - bundled deployment is canonical approach

---

## Tier B: FFI Wrapper Languages

### Python (pybind11) ✅
- **Status:** Only Python implementation (ctypes removed)
- **Data Types:** Native Python types (int, str, list, Optional)
- **Performance:** Excellent (C++ extension, minimal overhead)
- **Memory:** Automatic GC
- **Feature Coverage:** Complete via pybind11 bindings
- **Verdict:** ✅ **INCLUDE in v1.0**

### JavaScript/Node.js 🔬
- **Status:** ffi-napi wrapper generated (not yet benchmarked)
- **Data Types:** JS types (Number, String, Array, null/undefined)
- **Performance:** 🔬 Needs benchmarking - FFI overhead unknown
- **Memory:** GC, but FFI overhead unknown
- **Dependencies:** `ffi-napi`, `ref-napi`, `ref-struct-di`
- **Current State:** Generator exists, needs evaluation
- **Action Required:** Implement benchmarks, measure performance
- **Verdict:** 🔬 **NEEDS EVALUATION** - generator ready, benchmarking required

### Ruby 🔬
- **Status:** ruby-ffi wrapper generated (not yet benchmarked)
- **Data Types:** Ruby types (Integer, String, Array, nil)
- **Performance:** 🔬 Needs benchmarking - Ruby FFI overhead unknown
- **Memory:** GC with finalizers for cleanup
- **Current State:** Generator exists, needs evaluation
- **Note:** Ruby FFI is known to be slow, but Ruby itself is slow
- **Action Required:** Implement benchmarks, measure performance
- **Verdict:** 🔬 **NEEDS EVALUATION** - generator ready, benchmarking required

### Swift ✅
- **Status:** Swift package with C++ interop, benchmarked
- **Data Types:** Swift types (Int32, String, Array, Optional)
- **Performance:** ✅ **Good** - performs well in benchmarks
  - Array of ints: 1574 ns encode, 11525 ns decode
  - Working benchmarks in `benchmarks/generated/ffire_swift_*`
- **Memory:** ARC (Automatic Reference Counting)
- **Requirements:** Swift 5.9+ with C++ interop mode
- **Note:** iOS support (XCFramework) not yet implemented
- **Verdict:** ✅ **STRONG CANDIDATE** - performs well, benchmarks working

### PHP 🔬
- **Status:** FFI wrapper generated (not yet tested)
- **Data Types:** PHP types (int, string, array)
- **Performance:** 🔬 Needs benchmarking - completely unknown
- **Memory:** GC, PHP FFI untested in this context
- **Requirements:** PHP 7.4+ (FFI extension)
- **Current State:** Generator exists, needs evaluation
- **Note:** PHP FFI is relatively new feature
- **Action Required:** Test if it works, implement benchmarks
- **Verdict:** � **NEEDS EVALUATION** - generator ready, testing required

### Java 🔬
- **Status:** JNA wrapper generated (not yet tested)
- **Data Types:** Java types (int, String, ArrayList, null)
- **Performance:** 🔬 Needs benchmarking - JNA has overhead
- **Memory:** GC, native memory cleanup needs verification
- **Current State:** Generator exists, needs evaluation
- **Note:** JNA is easier but slower than JNI
- **Action Required:** Test if it works, implement benchmarks
- **Alternative:** Consider JNI instead if JNA performance is poor
- **Verdict:** � **NEEDS EVALUATION** - generator ready, testing required

### C# 🔬
- **Status:** P/Invoke wrapper generated (not yet tested)
- **Data Types:** C# types (int, string, List, null)
- **Performance:** 🔬 Needs benchmarking - P/Invoke overhead unknown
- **Memory:** GC, native memory cleanup needs verification
- **Platforms:** Windows, .NET Core, Mono
- **Current State:** Generator exists, needs evaluation
- **Action Required:** Test on target platforms, implement benchmarks
- **Verdict:** � **NEEDS EVALUATION** - generator ready, testing required

### Dart ✅
- **Status:** dart:ffi wrapper generated and tested
- **Data Types:** Dart types (int, String, List, null)
- **Performance:** ✅ **Admirably** - performs well in benchmarks
  - Empty: 941 ns encode, 1557 ns decode
  - Struct: 961 ns encode, 1452 ns decode
  - Complex: 4757 ns encode, 5346 ns decode
  - Array of floats: 2485 ns encode, 13508 ns decode
- **Memory:** GC with finalizers for native memory
- **Benchmark Results:** Comparable to C++ and Go, much faster than Python
- **Verdict:** ✅ **STRONG CANDIDATE** - performs admirably, working in benchmarks

---

## Summary

### ✅ Confirmed Working (Benchmarked)
1. **Go** - Reference implementation, excellent performance
2. **C++** - Primary implementation, excellent performance
3. **Python (pybind11)** - Far superior to ctypes, benchmarked
4. **Dart** - ✅ **Performs admirably** - verified in benchmarks
5. **Swift** - ✅ **Good performance** - benchmarks working

### ✅ Confirmed Working (C ABI Foundation)
6. **C** - C ABI working, bundled deployment is canonical (no symbol versioning needed)

### 🔬 Needs Evaluation (Generator Ready)
7. **JavaScript/Node.js** - Generator exists, needs benchmarks
8. **Ruby** - Generator exists, needs benchmarks
9. **PHP** - Generator exists, needs testing and benchmarks
10. **Java** - Generator exists, needs testing and benchmarks
11. **C#** - Generator exists, needs testing and benchmarks

**Status:** All generators kept for evaluation. Some proven, some need experimentation to determine v1.0 inclusion.

---

## Action Items

### Immediate (This Week)
1. ✅ Mark Python as pybind11-only in all docs
2. ✅ Document bundled dylib as canonical deployment (no symbol versioning needed)
3. 🔬 Benchmark JavaScript/Node.js generator
4. 🔬 Benchmark Ruby generator
5. 🔬 Benchmark Swift generator

### Short-Term (Next 2 Weeks)
6. 🗑️ Drop or mark experimental: PHP, Java, C#, Dart (unless proven)
7. 📊 Performance comparison table for all generators
8. 📝 Document which generators ship in v1.0

### Evaluation Process
Each language needs:
1. **Functionality Test** - Does it compile and work?
2. **Performance Benchmark** - Is it non-atrocious?
3. **Data Type Audit** - Are types sane and usable?
4. **Memory Test** - No leaks, proper cleanup?

Once evaluated, language will be marked:
- ✅ **Include in v1.0** - meets quality bar
- 🔬 **Experimental** - works but needs improvement
- ❌ **Defer** - doesn't meet bar, revisit later

---

## Deployment Model: Bundled Dylibs (Canonical)

**Decision:** Each package bundles its own dylib. This is the canonical approach.

**Architecture:**
```
Python package:
  myschema/
    lib/libmyschema.dylib    # Bundled with package
    __init__.py

Dart package:
  lib/libmyschema.dylib      # Separate bundled copy
  myschema.dart
```

**Why bundled?**
- ✅ No version conflicts between applications
- ✅ No system-wide installation required
- ✅ Each app controls its own ffire version
- ✅ Simple deployment (pip install, dart pub get, etc.)
- ✅ No symbol versioning needed

**What we DON'T support:**
- ❌ System-wide installation (`/usr/local/lib/libffire.dylib`)
- ❌ Plugin systems with multiple ffire versions in same process
- ❌ Shared library approach across multiple apps

**Symbol Versioning:** Not implemented and not needed because bundled deployment eliminates conflicts entirely. Each package has its own isolated dylib.

---

## Current Language Status for v1.0

**Proven (Benchmarked):**
- ✅ Go
- ✅ C++
- ✅ Python (pybind11)
- ✅ Dart (performs admirably!)
- ✅ Swift (good performance)

**Needs Benchmarking:**
- 🔬 JavaScript/Node.js
- 🔬 Ruby
- 🔬 PHP
- 🔬 Java
- 🔬 C#

**C ABI:** Working, bundled deployment is canonical (symbol versioning not needed)

**v1.0 Strategy:** Include languages that pass evaluation. Undecided means experimentation still needed.
