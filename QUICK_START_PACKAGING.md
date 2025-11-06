# FFire Multi-Language Packaging Quick Start

## What's New

FFire can now generate production-ready packages for multiple programming languages! Each package includes:

- **Pre-compiled library**: Fast C++ implementation compiled to dylib/so/dll
- **Language bindings**: Native wrappers using FFI
- **Package metadata**: setup.py, package.json, etc.
- **Documentation**: Comprehensive README with examples
- **Type definitions**: Where applicable (TypeScript .d.ts)

## Supported Languages

### Currently Available

1. **C++** (Tier A) - Direct dylib linkage
2. **Python** (Tier B) - ctypes wrapper + pip package
3. **JavaScript/Node.js** (Tier B) - ffi-napi wrapper + npm package
4. **TypeScript** (Tier B) - Same as JS + .d.ts definitions

### Coming Soon

- Swift (iOS/macOS)
- Ruby
- Rust
- Zig
- And more!

## Quick Examples

### Generate Python Package

```bash
# Generate
./ffire generate -lang python -schema myschema.ffi -out ./dist

# Install
cd dist/python
python3 -m venv venv
source venv/bin/activate
pip install .

# Use
python3
>>> from myschema import Message
>>> msg = Message.decode(data)
>>> encoded = msg.encode()
>>> msg.free()
```

### Generate JavaScript Package

```bash
# Generate
./ffire generate -lang javascript -schema myschema.ffi -out ./dist

# Install
cd dist/javascript
npm install

# Use (JavaScript)
const { Message } = require('myschema');
const msg = Message.decode(buffer);
const encoded = msg.encode();
msg.free();

# Use (TypeScript)
import { Message } from 'myschema';
const msg: Message = Message.decode(buffer);
const encoded: Buffer = msg.encode();
msg.free();
```

### Generate C++ Package

```bash
# Generate
./ffire generate -lang cpp -schema myschema.ffi -out ./dist

# The dylib is ready at dist/cpp/lib/libffire.dylib
# Link against it in your C++ project
```

## CLI Options

```bash
ffire generate [options]

Required:
  -lang string      Target language (python, javascript, cpp, etc.)
  -schema string    Path to .ffi schema file

Optional:
  -out string       Output directory (default "./dist")
  -ns string        Custom namespace/package name
  -O int            Optimization level 0-3 (default 2)
  -platform string  Target platform (darwin/linux/windows/all)
  -arch string      Target architecture (arm64/x86_64/all)
  -no-compile       Skip dylib compilation (testing only)
  -v                Verbose output
```

## Features

### Python Package
✅ ctypes FFI bindings  
✅ setuptools integration  
✅ Cross-platform support  
✅ PEP 668 compliant  
✅ Comprehensive README  

### JavaScript/Node.js Package
✅ ffi-napi bindings (modern fork)  
✅ JSDoc comments for autocomplete  
✅ TypeScript .d.ts definitions  
✅ npm package.json  
✅ Works with TypeScript, CoffeeScript, etc.  
✅ Node.js 14+ support  

### C++ Package
✅ Modern C++17 code  
✅ Cross-platform compilation  
✅ Optimization levels  
✅ Clean API  

## Architecture

All packages use the same **C ABI layer** for universal compatibility:

```
┌─────────────────┐
│  Your Language  │  (Python, JS, Swift, Ruby, etc.)
└────────┬────────┘
         │ FFI
┌────────▼────────┐
│    C ABI        │  (Opaque handles, extern "C")
└────────┬────────┘
         │
┌────────▼────────┐
│  C++ Codegen    │  (Fast serialization/deserialization)
└─────────────────┘
```

This means:
- **Consistent performance** across all languages
- **Single source of truth** (the C++ implementation)
- **Easy to add new languages** (just wrap the C ABI)

## Performance

Since all languages use the same compiled C++ dylib:

- **Serialization**: Native C++ speed (~1-2 GB/s)
- **FFI overhead**: Minimal (single function call)
- **Memory**: Efficient native allocations

## Testing

Each generated package can be tested:

### Python
```bash
cd dist/python
python3 -m venv venv
source venv/bin/activate
pip install .
python3 -c "from test import Message; print('✅ Success')"
```

### JavaScript
```bash
cd dist/javascript
npm install
node test.js  # Example test script included
```

## Generated File Structure

### Python
```
python/
├── myschema/
│   ├── __init__.py      # Package exports
│   └── bindings.py      # ctypes wrapper
├── lib/
│   └── libffire.dylib   # Compiled library
├── setup.py             # pip configuration
└── README.md            # Documentation
```

### JavaScript
```
javascript/
├── lib/
│   └── libffire.dylib   # Compiled library
├── index.js             # ffi-napi wrapper (JSDoc)
├── index.d.ts           # TypeScript definitions
├── package.json         # npm configuration
├── test.js              # Example test
└── README.md            # Documentation
```

## Next Steps

1. **Try it out**: Generate packages for your schemas
2. **Feedback**: Let us know what works and what doesn't
3. **Contribute**: Add support for your favorite language
4. **Star**: If you find this useful, star the repo!

## Documentation

- [PACKAGING_STATUS.md](PACKAGING_STATUS.md) - Detailed status and roadmap
- [docs/packaging-spec.md](docs/packaging-spec.md) - Complete specification
- Package READMEs - Generated in each package directory

## Support

Need help? Check the generated README.md in your package directory for:
- Installation instructions
- API documentation
- Usage examples
- Troubleshooting tips

## License

See LICENSE file for details.

---

**Happy coding with FFire!** 🔥
