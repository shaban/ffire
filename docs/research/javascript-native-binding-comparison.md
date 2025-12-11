# JavaScript ↔ Native C/C++ Binding Approaches for High-Performance Binary Serialization

**Research Document for ffire Project**  
*Last Updated: January 2025*

---

## Executive Summary

This document compares the fastest approaches for interfacing JavaScript with native C/C++ code, specifically for high-performance binary serialization use cases like `ffire`. The analysis covers **N-API/Node-API**, **Koffi**, **WebAssembly (WASM/Emscripten)**, **Bun FFI**, and **Deno FFI**.

### Target API Signatures
```c
void* config_decode(uint8_t* data, size_t len, char** error);
int config_encode(void* handle, uint8_t** out, size_t* out_len, char** error);
```

### Quick Recommendation

| Use Case | Recommended Approach | Why |
|----------|---------------------|-----|
| **Production Node.js** | **N-API (node-addon-api)** | Lowest latency, ABI-stable, battle-tested |
| **Zero-Compile Distribution** | **Koffi** | No native compilation, ~50% overhead acceptable |
| **Browser + Node.js** | **WebAssembly** | Universal runtime, portable binaries |
| **Bun Projects** | **Bun FFI** or N-API | Bun's FFI is 2-6x faster, but experimental |
| **Deno Projects** | **Deno FFI** or N-API | Deno supports both well |

---

## 1. Comparison Table

| Feature | N-API (node-addon-api) | Koffi | WebAssembly (Emscripten) | Bun FFI | Deno FFI |
|---------|------------------------|-------|--------------------------|---------|----------|
| **Latency per Call** | ★★★★★ (~50-100ns) | ★★★★☆ (~80-150ns) | ★★★☆☆ (~200-500ns) | ★★★★★ (~30-80ns) | ★★★★☆ (~100-200ns) |
| **Memory Copy Overhead** | Zero-copy possible | Zero-copy possible | Copy required (WASM heap) | Zero-copy possible | Zero-copy possible |
| **Build Complexity** | Medium (node-gyp/cmake-js) | None (pure JS) | High (Emscripten toolchain) | None (pure JS) | None (pure JS) |
| **Cross-Platform** | ✅ All platforms | ✅ All platforms | ✅ All platforms + browsers | ⚠️ Bun platforms only | ⚠️ Deno platforms only |
| **Browser Support** | ❌ No | ❌ No | ✅ Yes | ❌ No | ❌ No |
| **Type Precision** | ★★★★★ Full | ★★★★★ Full | ★★★★☆ Good | ★★★★★ Full | ★★★★★ Full |
| **npm Installable (No Compile)** | ⚠️ With prebuild | ✅ Yes | ✅ Yes (.wasm file) | ✅ Yes | N/A |
| **Maturity** | ★★★★★ Production | ★★★★☆ Mature | ★★★★★ Production | ★★☆☆☆ Experimental | ★★★☆☆ Stable |

---

## 2. Detailed Analysis

### 2.1 N-API / Node-API (node-addon-api)

**Overview**: The official Node.js API for building native addons. Provides an ABI-stable interface across Node.js versions.

#### Performance Characteristics
- **Call Overhead**: ~50-100 nanoseconds per call (baseline reference)
- **Memory**: Supports zero-copy via `napi_create_external_arraybuffer` and `napi_create_external_buffer`
- **Threading**: Supports async workers via `napi_create_async_work`

#### Type Mapping

| C Type | N-API Type | JavaScript Type |
|--------|-----------|-----------------|
| `int8_t` | `napi_int8_array` | `Int8Array` |
| `int16_t` | `napi_int16_array` | `Int16Array` |
| `int32_t` | `napi_get_value_int32` | `number` |
| `int64_t` | `napi_get_value_int64` / `napi_get_value_bigint_int64` | `number` / `BigInt` |
| `float` | `napi_float32_array` | `Float32Array` |
| `double` | `napi_get_value_double` / `napi_float64_array` | `number` / `Float64Array` |
| `uint8_t*` | `napi_create_buffer` / `napi_create_external_buffer` | `Buffer` / `Uint8Array` |
| `void*` | `napi_create_external` | External handle |
| `char*` | `napi_create_string_utf8` | `string` |

#### Example Implementation

```cpp
// Native addon (C++)
#include <napi.h>

Napi::Value ConfigDecode(const Napi::CallbackInfo& info) {
    Napi::Env env = info.Env();
    
    // Get buffer (zero-copy access)
    Napi::Buffer<uint8_t> buffer = info[0].As<Napi::Buffer<uint8_t>>();
    uint8_t* data = buffer.Data();
    size_t len = buffer.Length();
    
    char* error = nullptr;
    void* handle = config_decode(data, len, &error);
    
    if (error) {
        Napi::Error::New(env, error).ThrowAsJavaScriptException();
        free(error);
        return env.Null();
    }
    
    // Return opaque handle as external
    return Napi::External<void>::New(env, handle, [](Napi::Env, void* h) {
        config_free(h); // Custom destructor
    });
}
```

```javascript
// JavaScript usage
const addon = require('./build/Release/ffire.node');
const buffer = fs.readFileSync('config.bin');
const handle = addon.configDecode(buffer);
```

#### Build & Distribution

```json
// package.json
{
  "scripts": {
    "install": "node-gyp rebuild",
    "prebuild": "prebuildify --napi"
  },
  "dependencies": {
    "node-addon-api": "^8.5.0"
  }
}
```

**Prebuilt Binaries**: Use `prebuildify` or `prebuild` to distribute precompiled binaries:
```bash
# Build for multiple platforms
prebuildify --napi --strip
# Creates prebuilds/linux-x64/node.napi.node, etc.
```

#### Pros & Cons

| Pros | Cons |
|------|------|
| Fastest possible performance | Requires native compilation |
| Zero-copy memory access | More complex build setup |
| ABI stable across Node versions | Platform-specific binaries |
| Battle-tested in production | Larger distribution size |
| Full type precision | |

---

### 2.2 Koffi

**Overview**: A pure JavaScript FFI library for Node.js that allows calling native functions without writing any C++ glue code.

#### Performance Characteristics
- **Call Overhead**: ~50-58% slower than N-API for simple calls (per Koffi benchmarks)
- **For Complex APIs**: Only ~23% overhead vs N-API (Raylib benchmark)
- **Memory**: Supports zero-copy via `koffi.view()` (external buffers)
- **vs node-ffi-napi**: ~10,000% faster (node-ffi-napi has massive overhead)

#### Type Mapping

| C Type | Koffi Type | Notes |
|--------|-----------|-------|
| `int8_t` | `'int8'` or `'int8_t'` | |
| `int16_t` | `'int16'` or `'int16_t'` | |
| `int32_t` | `'int32'` or `'int'` | |
| `int64_t` | `'int64'` or `'int64_t'` | Returns BigInt |
| `uint8_t` | `'uint8'` or `'uint8_t'` | |
| `uint64_t` | `'uint64'` | Returns BigInt |
| `float` | `'float32'` or `'float'` | |
| `double` | `'float64'` or `'double'` | |
| `void*` | `'void *'` | Opaque pointer |
| `char*` | `'str'` or `'const char *'` | Auto-converted to/from JS string |
| `uint8_t*` | `'uint8_t *'` | Use with TypedArrays |

#### Example Implementation

```javascript
const koffi = require('koffi');

// Load the library
const lib = koffi.load('./libffire.so'); // or .dll, .dylib

// Define pointer to error string (output parameter)
const ErrorPtr = koffi.pointer('char *');

// Define function signatures
const config_decode = lib.func(
    'void *config_decode(uint8_t *data, size_t len, _Out_ char **error)'
);

const config_encode = lib.func(
    'int config_encode(void *handle, _Out_ uint8_t **out, _Out_ size_t *out_len, _Out_ char **error)'
);

const config_free = lib.func('void config_free(void *handle)');
const buffer_free = lib.func('void buffer_free(uint8_t *buf)');

// Usage
function decode(buffer) {
    const error = [null]; // Single-element array for output
    
    // Pass TypedArray directly
    const handle = config_decode(
        new Uint8Array(buffer), 
        buffer.byteLength, 
        error
    );
    
    if (error[0]) {
        throw new Error(error[0]);
    }
    
    return handle;
}

function encode(handle) {
    const outPtr = [null];
    const outLen = [0];
    const error = [null];
    
    const result = config_encode(handle, outPtr, outLen, error);
    
    if (error[0] || result !== 0) {
        throw new Error(error[0] || 'Encode failed');
    }
    
    // Zero-copy view into native memory
    const view = koffi.view(outPtr[0], outLen[0]);
    const data = new Uint8Array(view).slice(); // Copy to JS-owned buffer
    
    buffer_free(outPtr[0]); // Free native buffer
    return data;
}
```

#### Advanced: Zero-Copy with External Buffers

```javascript
// Read directly from native memory without copying
const ptr = config_get_buffer(handle);
const len = config_get_length(handle);

// Create zero-copy view (lives as long as handle)
const view = koffi.view(ptr, len);
const array = new Uint8Array(view);

// Use array directly - no copy!
console.log(array[0], array[1]);
```

#### Struct Support

```javascript
// Define struct matching C layout
const Config = koffi.struct('Config', {
    version: 'int32',
    flags: 'uint64',
    name: 'char[64]',
    data_ptr: 'uint8_t *',
    data_len: 'size_t'
});

// Use struct in function calls
const get_config = lib.func('int get_config(_Out_ Config *config)');

let config = {};
get_config(config);
console.log(config.version, config.flags);
```

#### Pros & Cons

| Pros | Cons |
|------|------|
| No native compilation required | ~50% overhead vs N-API |
| npm install just works | Slightly larger memory footprint |
| Prebuilt binaries included | Less control over memory lifecycle |
| Excellent type mapping | |
| Cross-platform out of box | |
| Zero-copy possible via koffi.view() | |

---

### 2.3 WebAssembly (Emscripten)

**Overview**: Compile C/C++ to WebAssembly for universal execution in browsers and Node.js.

#### Performance Characteristics
- **Call Overhead**: ~200-500 nanoseconds per call (WASM boundary crossing)
- **Memory**: All data must be copied to/from WASM linear memory (HEAP8, etc.)
- **Computation**: Near-native speed within WASM module

#### Type Mapping (via ccall/cwrap)

| C Type | ccall/cwrap Type | Notes |
|--------|-----------------|-------|
| `int`, `int32_t` | `'number'` | |
| `float`, `double` | `'number'` | |
| `void*`, pointers | `'number'` | Pointer is just an offset |
| `char*` | `'string'` | Auto-converted |
| `uint8_t*` | `'array'` | Must copy data |

#### Example Implementation

```c
// ffire.c - Compile with Emscripten
#include <emscripten.h>
#include <stdlib.h>

EMSCRIPTEN_KEEPALIVE
void* config_decode(uint8_t* data, size_t len, char** error) {
    // Your decode logic
    return handle;
}

EMSCRIPTEN_KEEPALIVE
int config_encode(void* handle, uint8_t** out, size_t* out_len, char** error) {
    // Your encode logic
    return 0;
}

EMSCRIPTEN_KEEPALIVE
void config_free(void* handle) {
    free(handle);
}
```

```bash
# Compile
emcc ffire.c -o ffire.js \
    -sEXPORTED_FUNCTIONS='["_config_decode","_config_encode","_config_free","_malloc","_free"]' \
    -sEXPORTED_RUNTIME_METHODS='["ccall","cwrap","getValue","setValue"]' \
    -sMODULARIZE=1 \
    -sWASM=1 \
    -O3
```

```javascript
// JavaScript usage
const createModule = require('./ffire.js');

async function init() {
    const Module = await createModule();
    
    // Wrap functions
    const configDecode = Module.cwrap('config_decode', 'number', ['number', 'number', 'number']);
    const configEncode = Module.cwrap('config_encode', 'number', ['number', 'number', 'number', 'number']);
    
    function decode(buffer) {
        // Allocate memory in WASM heap
        const ptr = Module._malloc(buffer.length);
        Module.HEAPU8.set(buffer, ptr);
        
        // Allocate error pointer
        const errorPtr = Module._malloc(4); // pointer size
        Module.setValue(errorPtr, 0, 'i32');
        
        const handle = configDecode(ptr, buffer.length, errorPtr);
        
        // Check for error
        const errorStrPtr = Module.getValue(errorPtr, 'i32');
        if (errorStrPtr) {
            const error = Module.UTF8ToString(errorStrPtr);
            Module._free(errorPtr);
            Module._free(ptr);
            throw new Error(error);
        }
        
        Module._free(errorPtr);
        Module._free(ptr);
        return handle;
    }
    
    function encode(handle) {
        const outPtr = Module._malloc(4);  // uint8_t**
        const lenPtr = Module._malloc(4);  // size_t*
        const errorPtr = Module._malloc(4); // char**
        
        const result = configEncode(handle, outPtr, lenPtr, errorPtr);
        
        if (result !== 0) {
            const errorStrPtr = Module.getValue(errorPtr, 'i32');
            throw new Error(Module.UTF8ToString(errorStrPtr));
        }
        
        const dataPtr = Module.getValue(outPtr, 'i32');
        const dataLen = Module.getValue(lenPtr, 'i32');
        
        // Copy data from WASM heap
        const output = new Uint8Array(Module.HEAPU8.buffer, dataPtr, dataLen).slice();
        
        Module._free(dataPtr);
        Module._free(outPtr);
        Module._free(lenPtr);
        Module._free(errorPtr);
        
        return output;
    }
    
    return { decode, encode };
}
```

#### Using EM_JS for Direct JavaScript Interop

```c
// Inline JS in C code (lower overhead)
EM_JS(void, log_data, (const uint8_t* data, int len), {
    const view = HEAPU8.subarray(data, data + len);
    console.log('Data:', Array.from(view));
});
```

#### Pros & Cons

| Pros | Cons |
|------|------|
| Works in browsers | Memory must be copied (no true zero-copy) |
| Portable .wasm binary | Higher call overhead |
| No native dependencies | Complex build toolchain |
| Sandboxed execution | Larger bundle size |
| Same code for web & Node | Limited to WASM numeric types |

---

### 2.4 Bun FFI

**Overview**: Bun's native FFI module using TinyCC for JIT C compilation, claiming 2-6x faster performance than Node.js N-API.

#### Performance Characteristics
- **Call Overhead**: ~30-80 nanoseconds (2-6x faster than N-API per Bun claims)
- **Memory**: Zero-copy via `toArrayBuffer()` and `ptr()`
- **Status**: Experimental, known bugs, Bun recommends N-API for production

#### Type Mapping

| C Type | Bun FFI Type | Notes |
|--------|-------------|-------|
| `int8_t` | `FFIType.i8` | |
| `int16_t` | `FFIType.i16` | |
| `int32_t` | `FFIType.i32` | |
| `int64_t` | `FFIType.i64` | BigInt in JS |
| `uint8_t` | `FFIType.u8` | |
| `uint16_t` | `FFIType.u16` | |
| `uint32_t` | `FFIType.u32` | |
| `uint64_t` | `FFIType.u64` | BigInt in JS |
| `float` | `FFIType.f32` | |
| `double` | `FFIType.f64` | |
| `void*` | `FFIType.ptr` | |
| `char*` | `FFIType.cstring` | Read-only, use ptr for output |

#### Example Implementation

```javascript
import { dlopen, FFIType, ptr, toArrayBuffer, CString } from "bun:ffi";

const lib = dlopen("./libffire.so", {
    config_decode: {
        args: [FFIType.ptr, FFIType.u64, FFIType.ptr],
        returns: FFIType.ptr,
    },
    config_encode: {
        args: [FFIType.ptr, FFIType.ptr, FFIType.ptr, FFIType.ptr],
        returns: FFIType.i32,
    },
    config_free: {
        args: [FFIType.ptr],
        returns: FFIType.void,
    },
    buffer_free: {
        args: [FFIType.ptr],
        returns: FFIType.void,
    },
});

function decode(buffer: Uint8Array): number {
    // Get pointer to buffer data
    const dataPtr = ptr(buffer);
    
    // Allocate error pointer (as TypedArray for output)
    const errorPtr = new BigUint64Array(1);
    
    const handle = lib.symbols.config_decode(dataPtr, buffer.length, ptr(errorPtr));
    
    if (errorPtr[0] !== 0n) {
        const errorStr = new CString(Number(errorPtr[0]));
        throw new Error(errorStr.toString());
    }
    
    return handle;
}

function encode(handle: number): Uint8Array {
    const outPtr = new BigUint64Array(1);
    const outLen = new BigUint64Array(1);
    const errorPtr = new BigUint64Array(1);
    
    const result = lib.symbols.config_encode(
        handle,
        ptr(outPtr),
        ptr(outLen),
        ptr(errorPtr)
    );
    
    if (result !== 0) {
        throw new Error(new CString(Number(errorPtr[0])).toString());
    }
    
    // Zero-copy access via toArrayBuffer
    const buffer = toArrayBuffer(Number(outPtr[0]), 0, Number(outLen[0]));
    const copy = new Uint8Array(buffer).slice(); // Copy before freeing
    
    lib.symbols.buffer_free(Number(outPtr[0]));
    return copy;
}
```

#### Callbacks from C

```javascript
import { JSCallback } from "bun:ffi";

const callback = new JSCallback(
    (dataPtr, len) => {
        const buffer = toArrayBuffer(dataPtr, 0, len);
        console.log("Received:", new Uint8Array(buffer));
        return 0;
    },
    {
        args: [FFIType.ptr, FFIType.u64],
        returns: FFIType.i32,
    }
);

// Pass callback.ptr to C function
```

#### Pros & Cons

| Pros | Cons |
|------|------|
| Fastest FFI (2-6x vs N-API) | Bun-only runtime |
| Zero-copy memory access | Experimental, known bugs |
| No compilation needed | Bun recommends N-API for production |
| JIT compilation via TinyCC | Smaller ecosystem |

---

### 2.5 Deno FFI

**Overview**: Deno's foreign function interface via `Deno.dlopen`, similar to Koffi but built into Deno.

#### Performance Characteristics
- **Call Overhead**: ~100-200 nanoseconds (similar to Koffi, uses libffi internally)
- **Memory**: Zero-copy via `Deno.UnsafePointerView`
- **Security**: Requires `--allow-ffi` flag (sandboxed by default)

#### Type Mapping

| C Type | Deno FFI Type | Notes |
|--------|--------------|-------|
| `int8_t` | `"i8"` | |
| `int16_t` | `"i16"` | |
| `int32_t` | `"i32"` | |
| `int64_t` | `"i64"` | BigInt in JS |
| `uint8_t` | `"u8"` | |
| `uint16_t` | `"u16"` | |
| `uint32_t` | `"u32"` | |
| `uint64_t` | `"u64"` | BigInt in JS |
| `float` | `"f32"` | |
| `double` | `"f64"` | |
| `void*` | `"pointer"` | |
| `const char*` | `"buffer"` | For string input |
| Struct | `{ struct: [...] }` | By value |

#### Example Implementation

```typescript
// Run with: deno run --allow-ffi --unstable main.ts

const lib = Deno.dlopen("./libffire.so", {
    config_decode: {
        parameters: ["buffer", "usize", "pointer"],
        result: "pointer",
    },
    config_encode: {
        parameters: ["pointer", "pointer", "pointer", "pointer"],
        result: "i32",
    },
    config_free: {
        parameters: ["pointer"],
        result: "void",
    },
    buffer_free: {
        parameters: ["pointer"],
        result: "void",
    },
});

function decode(buffer: Uint8Array): Deno.PointerValue {
    // Allocate error pointer
    const errorBuffer = new BigUint64Array(1);
    
    const handle = lib.symbols.config_decode(
        buffer,
        buffer.length,
        Deno.UnsafePointer.of(errorBuffer)
    );
    
    if (errorBuffer[0] !== 0n) {
        const errorPtr = Deno.UnsafePointer.create(errorBuffer[0]);
        const errorView = new Deno.UnsafePointerView(errorPtr!);
        throw new Error(errorView.getCString());
    }
    
    return handle;
}

function encode(handle: Deno.PointerValue): Uint8Array {
    const outPtr = new BigUint64Array(1);
    const outLen = new BigUint64Array(1);
    const errorPtr = new BigUint64Array(1);
    
    const result = lib.symbols.config_encode(
        handle,
        Deno.UnsafePointer.of(outPtr),
        Deno.UnsafePointer.of(outLen),
        Deno.UnsafePointer.of(errorPtr)
    );
    
    if (result !== 0) {
        const errView = new Deno.UnsafePointerView(
            Deno.UnsafePointer.create(errorPtr[0])!
        );
        throw new Error(errView.getCString());
    }
    
    // Zero-copy view
    const dataPointer = Deno.UnsafePointer.create(outPtr[0])!;
    const view = new Deno.UnsafePointerView(dataPointer);
    const arrayBuffer = view.getArrayBuffer(Number(outLen[0]));
    
    // Copy to owned buffer before freeing
    const copy = new Uint8Array(arrayBuffer).slice();
    
    lib.symbols.buffer_free(dataPointer);
    return copy;
}

// Usage
const data = new Uint8Array([1, 2, 3, 4]);
const handle = decode(data);
const encoded = encode(handle);
lib.symbols.config_free(handle);
```

#### Callbacks (UnsafeCallback)

```typescript
const callback = new Deno.UnsafeCallback(
    {
        parameters: ["pointer", "usize"],
        result: "i32",
    },
    (dataPtr, len) => {
        const view = new Deno.UnsafePointerView(dataPtr!);
        const buffer = view.getArrayBuffer(Number(len));
        console.log("Received:", new Uint8Array(buffer));
        return 0;
    }
);

// Pass callback.pointer to C function
```

#### Pros & Cons

| Pros | Cons |
|------|------|
| No compilation needed | Deno-only runtime |
| Zero-copy memory access | Requires --allow-ffi flag |
| TypeScript-first | Slightly higher overhead than N-API |
| Built into Deno | Smaller ecosystem |
| Supports structs by value | |

---

## 3. Performance Benchmarks Summary

Based on available benchmarks (Koffi, Bun documentation):

### Call Overhead Comparison

| Approach | Relative Speed (higher is better) | Absolute Overhead |
|----------|----------------------------------|-------------------|
| **Bun FFI** | 200-600% of N-API | ~30-80ns |
| **N-API** | 100% (baseline) | ~50-100ns |
| **Koffi** | 63-77% of N-API | ~80-150ns |
| **Deno FFI** | ~60-80% of N-API | ~100-200ns |
| **WebAssembly** | 20-50% of N-API | ~200-500ns |
| **node-ffi-napi** | <1% of N-API | ~10,000ns+ |

### Memory Copy Overhead

| Approach | Zero-Copy Possible | Method |
|----------|-------------------|--------|
| **N-API** | ✅ Yes | `napi_create_external_buffer` |
| **Koffi** | ✅ Yes | `koffi.view()` |
| **Bun FFI** | ✅ Yes | `toArrayBuffer()` |
| **Deno FFI** | ✅ Yes | `UnsafePointerView.getArrayBuffer()` |
| **WebAssembly** | ❌ No | Must copy to WASM heap |

---

## 4. Recommendations for ffire

### Primary Recommendation: **N-API with Prebuild**

For a production binary serialization library, **N-API** provides:
- Lowest latency for repeated encode/decode calls
- Zero-copy buffer handling (critical for large messages)
- Stable ABI across Node.js versions
- Proven at scale (used by better-sqlite3, sharp, etc.)

**Distribution Strategy**:
```bash
# Use prebuildify for prebuilt binaries
npm install prebuildify
npx prebuildify --napi --strip

# Or use node-pre-gyp for more control
npm install @mapbox/node-pre-gyp
```

### Alternative: **Koffi for Zero-Compile Distribution**

If eliminating native compilation is a priority:
- ~50% overhead is acceptable for many use cases
- `npm install` just works everywhere
- Still supports zero-copy via `koffi.view()`

### For Multi-Runtime Support

If targeting multiple JavaScript runtimes:

1. **WebAssembly** for browser + Node.js + Deno + Bun (universal)
2. **N-API** for Node.js + Deno + Bun (all support Node addons)
3. **Runtime-specific FFI** if only targeting one runtime

### Recommended Architecture

```
ffire/
├── src/
│   └── lib.c                 # Core C implementation
├── bindings/
│   ├── node/                 # N-API binding
│   │   ├── binding.gyp
│   │   └── ffire.cc
│   ├── wasm/                 # Emscripten build
│   │   └── Makefile
│   └── js/                   # Koffi-based (fallback)
│       └── index.js
└── npm/
    ├── ffire/                # Main package (tries N-API, falls back to Koffi)
    ├── ffire-native/         # N-API only
    └── ffire-wasm/           # WASM only
```

---

## 5. Implementation Checklist

### For N-API Implementation

- [ ] Set up node-gyp or cmake-js build
- [ ] Use `node-addon-api` C++ wrapper for cleaner code
- [ ] Implement zero-copy buffer handling
- [ ] Add async workers for large operations
- [ ] Set up prebuildify for prebuilt binaries
- [ ] Test across Node.js 16, 18, 20, 22
- [ ] Add TypeScript declarations

### For Koffi Fallback

- [ ] Create type definitions matching C API
- [ ] Handle output parameters correctly
- [ ] Use `koffi.view()` for zero-copy where possible
- [ ] Test on Windows, Linux, macOS

### For WebAssembly Build

- [ ] Set up Emscripten toolchain
- [ ] Configure EXPORTED_FUNCTIONS and RUNTIME_METHODS
- [ ] Implement memory management wrappers
- [ ] Test in browser and Node.js
- [ ] Optimize with -O3 and -flto

---

## 6. References

- [Node-API Documentation](https://nodejs.org/api/n-api.html)
- [node-addon-api](https://github.com/nodejs/node-addon-api)
- [Koffi Documentation](https://koffi.dev/)
- [Emscripten Documentation](https://emscripten.org/)
- [Bun FFI Documentation](https://bun.sh/docs/api/ffi)
- [Deno FFI Documentation](https://deno.land/manual/runtime/ffi_api)
- [prebuildify](https://github.com/prebuild/prebuildify)
