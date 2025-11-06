# FFire Multi-Language Packaging Status

## Overview

FFire now supports generating production-ready packages for multiple languages using a C ABI dylib approach. Each package includes:

- Pre-compiled dynamic library (dylib/so/dll)
- Language-specific wrapper/bindings
- Package metadata (setup.py, package.json, etc.)
- Comprehensive README with usage examples
- Type definitions where applicable

## Tier System

### Tier A: Native Languages
Languages that can use C ABI directly or prefer native code.
- **Output:** Dylib only (users link against it)
- **Examples:** C, C++, Rust, Zig

### Tier B: FFI Wrapper Languages
Languages that use FFI to call into native libraries.
- **Output:** Dylib + wrapper + package metadata
- **Examples:** Python, JavaScript, Swift, Ruby

## Implementation Status

### ✅ Completed

#### C ABI Layer
- [x] C header generation (opaque handles)
- [x] C++ implementation (HandleImpl pattern)
- [x] Cross-platform compilation (darwin/linux/windows)
- [x] Symbol export verification
- [x] Multiple message type support

#### CLI Infrastructure
- [x] Flag parsing (`-lang`, `-out`, `-O`, `-platform`, `-arch`)
- [x] Tier routing (Tier A vs Tier B)
- [x] Verbose output mode
- [x] Legacy single-file mode compatibility

#### C++ Package (Tier A)
- [x] C++ header generation
- [x] Dylib compilation
- [x] Optimization levels (0-3)
- [x] Platform detection
- [x] Architecture detection

#### Python Package (Tier B)
- [x] ctypes wrapper generation
- [x] setup.py generation
- [x] \_\_init\_\_.py generation
- [x] README.md with examples
- [x] Platform-specific package_data
- [x] PEP 668 documentation
- [x] Tested and verified (4293 bytes round-trip)

#### JavaScript/Node.js Package (Tier B)
- [x] ffi-napi wrapper generation
- [x] package.json generation
- [x] JSDoc comments for autocomplete
- [x] TypeScript .d.ts definitions
- [x] README.md with JS and TS examples
- [x] Platform detection (darwin/linux/win32)
- [x] Cross-compile-to-JS support (TypeScript, CoffeeScript, etc.)

### 🚧 In Progress

None currently.

### 📋 Planned

#### Swift Package (Tier B)
- [ ] Swift wrapper generation
- [ ] Package.swift generation
- [ ] README.md with examples
- [ ] iOS/macOS support

#### Ruby Package (Tier B)
- [ ] FFI gem wrapper
- [ ] Gemspec generation
- [ ] README.md with examples

#### Template System
- [ ] Generic template engine
- [ ] Easy addition of new languages
- [ ] Community contribution support

#### Multi-Platform Builds
- [ ] Cross-compilation support
- [ ] `-platform all` flag
- [ ] `-arch all` flag
- [ ] Fat binaries for macOS

#### Example Generation
- [ ] Example code for each package
- [ ] Integration tests
- [ ] Benchmark examples

## Generated Package Structure

### Python Package
```
python/
├── lib/
│   └── libffire.dylib          # Compiled dylib
├── test/
│   ├── __init__.py             # Package exports
│   └── bindings.py             # ctypes wrapper
├── setup.py                    # setuptools configuration
└── README.md                   # Documentation
```

### JavaScript Package
```
javascript/
├── lib/
│   └── libffire.dylib          # Compiled dylib
├── include/                    # C++ headers (for reference)
├── src/                        # C++ source (for reference)
├── index.js                    # ffi-napi wrapper with JSDoc
├── index.d.ts                  # TypeScript definitions
├── package.json                # npm package configuration
├── test.js                     # Example test script
└── README.md                   # Documentation
```

## CLI Usage

### Generate Python Package
```bash
ffire generate -lang python -schema audio.ffi -out ./dist
cd dist/python
python3 -m venv venv
source venv/bin/activate
pip install .
```

### Generate JavaScript Package
```bash
ffire generate -lang javascript -schema audio.ffi -out ./dist
cd dist/javascript
npm install
node test.js
```

### Generate C++ Package
```bash
ffire generate -lang cpp -schema audio.ffi -out ./dist
# Dylib ready at dist/cpp/lib/libffire.dylib
```

### Advanced Options
```bash
# Custom namespace
ffire generate -lang python -schema audio.ffi -ns myaudio

# Skip compilation (template testing)
ffire generate -lang ruby -schema audio.ffi --no-compile

# Optimization level
ffire generate -lang cpp -schema audio.ffi -O 3

# Verbose output
ffire generate -lang python -schema audio.ffi -v
```

## Testing

### Python
```bash
cd test-dist/python
python3 -m venv venv
source venv/bin/activate
pip install .
python -c "from test import Message; print('✅ Import successful')"
```

### JavaScript
```bash
cd test-dist/javascript
npm install
node test.js
```

## Technical Details

### C ABI Design

**Opaque Handles:**
```c
typedef struct message_handle_impl* message_handle;
```

**Decode Function:**
```c
message_handle message_decode(const uint8_t* data, size_t len, char** error);
```

**Encode Function:**
```c
size_t message_encode(message_handle handle, uint8_t** out_data, char** error);
```

**Free Functions:**
```c
void message_free(message_handle handle);
void message_free_data(uint8_t* data);
void message_free_error(char* error);
```

### Compilation

**macOS (clang++):**
```bash
clang++ -std=c++17 -dynamiclib -fPIC -O2 -arch arm64 \
  -I include -o lib/libffire.dylib src/generated_c.cpp
```

**Linux (g++):**
```bash
g++ -std=c++17 -shared -fPIC -O2 \
  -I include -o lib/libffire.so src/generated_c.cpp
```

**Windows (mingw):**
```bash
x86_64-w64-mingw32-g++ -std=c++17 -shared -O2 \
  -I include -o lib/ffire.dll src/generated_c.cpp
```

### Symbol Export Verification

All packages include properly exported C ABI symbols:

```bash
$ nm -gU lib/libffire.dylib | grep message
00000000000004c8 T _message_decode
00000000000013dc T _message_encode
0000000000001dc4 T _message_free
0000000000001e50 T _message_free_data
0000000000001e5c T _message_free_error
```

## Language Support Matrix

| Language   | Tier | Status      | FFI Library    | Package Manager |
|------------|------|-------------|----------------|-----------------|
| C++        | A    | ✅ Complete | Native         | Manual          |
| Python     | B    | ✅ Complete | ctypes         | pip/setuptools  |
| JavaScript | B    | ✅ Complete | ffi-napi       | npm             |
| TypeScript | B    | ✅ Complete | ffi-napi + .d.ts | npm         |
| Swift      | B    | 📋 Planned  | Swift FFI      | SPM             |
| Ruby       | B    | 📋 Planned  | FFI gem        | gem             |
| Rust       | A    | 📋 Planned  | bindgen        | cargo           |
| Zig        | A    | 📋 Planned  | @cImport       | zig build       |

## Developer Experience

### Python
- **Installation:** Standard pip/setuptools
- **Import:** `from test import Message`
- **Type hints:** Not included (can be added)
- **Documentation:** Docstrings in wrapper

### JavaScript
- **Installation:** Standard npm
- **Import:** `const { Message } = require('test')`
- **Type safety:** JSDoc comments for autocomplete
- **Documentation:** Inline JSDoc

### TypeScript
- **Installation:** Same as JavaScript (npm)
- **Import:** `import { Message } from 'test'`
- **Type safety:** Full .d.ts definitions
- **Documentation:** Inline JSDoc + types

## Performance

All packages use the same compiled C++ dylib, so performance is identical:

- **Serialization:** Native C++ speed
- **FFI overhead:** Minimal (single function call)
- **Memory:** Efficient (native allocations)

## Compatibility

### Python
- Requires: Python 3.6+
- Platforms: macOS, Linux, Windows
- Dependencies: ctypes (stdlib only)

### JavaScript/Node.js
- Requires: Node.js 14.0.0+
- Platforms: macOS, Linux, Windows
- Dependencies: ffi-napi, ref-napi

### TypeScript
- Requires: Node.js 14.0.0+ + TypeScript
- Platforms: macOS, Linux, Windows
- Dependencies: Same as JavaScript

## Next Steps

1. **Swift Package** - Implement Swift wrapper for iOS/macOS development
2. **Ruby Package** - Implement FFI gem for Ruby ecosystem
3. **Multi-platform Builds** - Support cross-compilation and fat binaries
4. **Template System** - Make it easy to add new language generators
5. **Example Gallery** - Create comprehensive examples for each language
6. **CI/CD Integration** - Automated multi-platform builds

## Contributing

To add support for a new language:

1. Create `pkg/generator/generator_<lang>.go`
2. Implement wrapper generation using C ABI
3. Add package metadata generation
4. Add language to Tier routing in `package.go`
5. Test with `--no-compile` flag first
6. Write README with examples

## License

See LICENSE file for details.
