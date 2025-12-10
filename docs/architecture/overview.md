# Architecture Overview

ffire is a binary serialization system with three main components:

```
┌──────────┐     ┌───────────┐     ┌─────────┐
│  Schema  │ --> │ Generator │ --> │  Codec  │
│ (.ffi)   │     │           │     │ (Native)│
└──────────┘     └───────────┘     └─────────┘
```

## Pipeline

1. **Parse**: Schema file → AST
2. **Generate**: AST → Language-specific code
3. **Compile**: Code → Binary codec
4. **Use**: Codec → Encode/Decode messages

## Components

### Parser (`pkg/parser`)
Reads `.ffi` files (Go syntax) and builds schema AST.

### Schema (`pkg/schema`)
In-memory representation of types, messages, and structure.

### Generators (`pkg/generator`)
Native code generation for 8 languages:
- `generator_go.go` - Go
- `generator_cpp.go` - C++
- `generator_csharp.go` - C#
- `generator_java.go` - Java
- `generator_swift.go` - Swift
- `generator_dart.go` - Dart
- `generator_rust.go` - Rust
- `generator_zig.go` - Zig

All generators produce standalone native code with no FFI dependencies.

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
