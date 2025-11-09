# Architecture Overview

ffire is a binary serialization system with three main components:

```
┌──────────┐     ┌───────────┐     ┌─────────┐
│  Schema  │ --> │ Generator │ --> │  Codec  │
│ (.ffi)   │     │           │     │ (Go/C++/│
└──────────┘     └───────────┘     │ Swift)  │
                                    └─────────┘
```

## Pipeline

1. **Parse**: Schema file → AST
2. **Generate**: AST → Language-specific code
3. **Compile**: Code → Binary codec (Go/C++/etc)
4. **Use**: Codec → Encode/Decode messages

## Components

### Parser (`pkg/parser`)
Reads `.ffi` files (Go syntax) and builds schema AST.

### Schema (`pkg/schema`)
In-memory representation of types, messages, and structure.

### Generators (`pkg/generator`)
Per-language code generation:
- `generator_go.go` - Native Go codec
- `generator_cpp.go` - C++ with header/implementation
- `generator_swift.go` - Swift package + C ABI
- `generator_dart.go` - Dart package + FFI
- etc.

### Encoder/Decoder
Runtime libraries for each language implementing the wire format.

## Data Flow

```
User Schema (.ffi)
    ↓
Parser (Go AST)
    ↓
Schema (typed AST)
    ↓
Generator (per language)
    ↓
Generated Code
    ↓
Compiled Library
    ↓
User Application
```

## Wire Format

See [Wire Format](wire-format.md) for encoding specification.

## Adding Languages

See [Generators](generators.md) for implementing new language support.
