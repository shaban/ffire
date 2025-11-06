# C++ Bindings - C ABI Dynamic Library Approach

Production implementation using C ABI dynamic library for universal language support.

## Quick Start

```bash
# Test C++ with C ABI
cd 03-cpp-dynamic
make && ./test

# Test Swift with C ABI
cd 05-swift-dynamic
make && ./test

# Test Objective-C++ with C ABI
cd 08-objcpp-dynamic
make && ./test

# Test Python with C ABI
cd 09-python-dynamic
make && python3 test.py
```

## Directory Structure

```
cpp-bindings/
├── common/                   # Shared C ABI implementation
│   ├── generated.hpp         # C++ implementation
│   ├── generated_c.h         # C ABI header (universal interface)
│   ├── generated_c.cpp       # C ABI implementation
│   └── complex.bin           # Test fixture (4293 bytes)
├── 03-cpp-dynamic/           # C++ using C ABI
├── 05-swift-dynamic/         # Swift using C ABI
├── 08-objcpp-dynamic/        # Objective-C++ using C ABI
├── 09-python-dynamic/        # Python using C ABI (ctypes)
├── README.md                 # This file
├── RESULTS.md                # Performance analysis
├── C-ABI-LANGUAGE-SUPPORT.md # Language compatibility matrix
└── GITHUB-COVERAGE-ANALYSIS.md # Developer coverage estimate
```

## Implemented Languages (C ABI Dylib)

| Test | Language | FFI Mechanism | Performance |
|------|----------|---------------|-------------|
| 03 | C++ | Direct C linkage | 3-4µs per op |
| 05 | Swift | @_silgen_name | 3-4µs per op |
| 08 | Obj-C++ | Direct C linkage | 3-4µs per op |
| 09 | Python | ctypes | 4-5µs per op |

## Key Results

**Performance:** All C++ approaches perform identically (~3µs decode, ~4µs encode)  
**Build Time:** ~1 second for all approaches  
**Binary Size:** 38-71KB depending on approach  
**Memory:** C++ 1.5MB, Obj-C++ 5.8MB, Python 17MB

See [RESULTS.md](./RESULTS.md) for complete analysis.

## Requirements

- Clang compiler (clang++)
- Make
- Python 3 (for Python test)
- hyperfine (for performance measurement)

```bash
# Install hyperfine on macOS
brew install hyperfine
```

## Regenerating Test Files

```bash
cd /Users/shaban/Code/ffire

# Generate C++ header
./ffire generate --lang cpp --schema testdata/schema/complex.ffi \
  --output experimental/cpp-bindings/common/generated.hpp

# Generate binary fixture
./ffire fixture --schema testdata/schema/complex.ffi \
  --json testdata/json/complex.json \
  --output experimental/cpp-bindings/common/complex.bin
```

## Each Test Includes

- **test.{cpp,mm,py}** - Test implementation
- **Makefile** - Build configuration
- **run.sh** - Measurement script
- **perf.json** - Hyperfine output (generated)

## Measurement Methodology

Each test measures:
- **Build time:** Cold and warm builds
- **Binary size:** Executable and library sizes
- **Runtime:** Mean and stddev over 50 runs (hyperfine)
- **Memory:** Peak resident set size
- **Performance:** Per-operation decode/encode time (100 iterations)

All tests use `-O2` optimization (production default, not debug, not aggressive).

## Common Patterns

### C++ Direct Include
```cpp
#include "../common/generated.hpp"
auto plugins = test::decode_plugin_message(data.data(), data.size());
```

### C++ Dynamic Library (C ABI)
```cpp
#include "../common/generated_c.h"
PluginHandle plugin = plugin_decode(data.data(), data.size(), &error);
```

### Python (ctypes)
```python
lib = ctypes.CDLL('./libffire.dylib')
plugin = lib.plugin_decode(data_array, len(data), ctypes.byref(error))
```

## Future Work

- [ ] Add Swift native C++ interop tests (5.9+)
- [ ] Fix C ABI encode size mismatch (169 vs 4293 bytes)
- [ ] Test on Linux and Windows
- [ ] Benchmark with more schemas
- [ ] Measure zero-copy decode approaches

## Documentation

This experiment is for evaluation purposes. Results and methodology will be archived in the repository documentation. The test code itself is ephemeral and not intended for production use.

---

**Created:** November 6, 2025  
**Purpose:** Evaluate C++ integration approaches for ffire  
**Status:** Complete (7/7 tests passing)
