# ffire

Fast binary serialization for complete data transfer and FFI.

## Why ffire?

Existing serialization formats are optimized for specific use cases:

- **Protocol Buffers** → Cross-architecture transport (HTTP/RPC), schema evolution, backward compatibility
- **FlatBuffers** → Zero-copy access, time-to-first-byte, random field access without full parsing

**ffire is different.** It's optimized for:

1. **Fast complete transfer** — When you need all the data, as fast as possible
2. **FFI-friendly native types** — Generated code that works seamlessly across language boundaries

No other format optimizes for generating idiomatic, native types that can be easily imported as a package/module to interface between low-level and high-level languages.

## Performance vs Protocol Buffers

Real benchmark data — array of 5000 integers (encode + decode):

| Format | Language | Total Time | vs ffire |
|--------|----------|------------|----------|
| ffire  | Rust     | 1,184 ns   | —        |
| ffire  | C++      | 1,405 ns   | —        |
| ffire  | Swift    | 1,347 ns   | —        |
| ffire  | C#       | 1,666 ns   | —        |
| ffire  | Zig      | 1,721 ns   | —        |
| ffire  | Java     | 2,226 ns   | —        |
| ffire  | Go       | 3,278 ns   | —        |
| proto  | Go       | 31,252 ns  | **9.5x slower** |

Nested message with 5000 integers:

| Format | Language | Total Time | vs ffire |
|--------|----------|------------|----------|
| ffire  | Rust     | 1,259 ns   | —        |
| ffire  | C++      | 1,327 ns   | —        |
| ffire  | Go       | 3,229 ns   | —        |
| proto  | Go       | 32,422 ns  | **10x slower** |

Optional fields with mixed types:

| Format | Language | Total Time | vs ffire |
|--------|----------|------------|----------|
| ffire  | C++      | 33,507 ns  | —        |
| ffire  | Zig      | 34,558 ns  | —        |
| ffire  | Go       | 58,542 ns  | —        |
| proto  | Go       | 140,870 ns | **2.4x slower** |

!!! info "Benchmark Methodology"
    All benchmarks run on the same machine, same data, measuring encode + decode roundtrip.
    See [full benchmark results](development/benchmarks.md) for all message types and languages.

## Quick Start

```bash
# Install
go install github.com/shaban/ffire/cmd/ffire@latest

# Generate code for your language
ffire gen --lang go --schema schema.ffire --output ./generated
```

=== "Go"
    ```go
    import "your-module/generated"

    // Encode
    msg := &Person{Name: "Alice", Age: 30}
    data := msg.Encode()

    // Decode
    decoded := &Person{}
    decoded.Decode(data)
    ```

=== "Rust"
    ```rust
    use generated::Person;

    // Encode
    let msg = Person { name: "Alice".into(), age: 30 };
    let data = msg.encode();

    // Decode
    let decoded = Person::decode(&data)?;
    ```

=== "C++"
    ```cpp
    #include "generated/person.hpp"

    // Encode
    Person msg{"Alice", 30};
    auto data = msg.encode();

    // Decode
    Person decoded;
    decoded.decode(data);
    ```

=== "Swift"
    ```swift
    import Generated

    // Encode
    let msg = Person(name: "Alice", age: 30)
    let data = msg.encode()

    // Decode
    let decoded = try Person.decode(data)
    ```

## Supported Languages

| Language | Status | Notes |
|----------|--------|-------|
| Go       | ✅ Stable | Reference implementation |
| Rust     | ✅ Stable | Fastest encode/decode |
| C++      | ✅ Stable | C++17, header-only |
| Swift    | ✅ Stable | Swift 5.9+ |
| C#       | ✅ Stable | .NET 6+ |
| Java     | ✅ Stable | Java 11+ |
| Dart     | ✅ Stable | Dart 3.0+ |
| Zig      | ✅ Stable | Zig 0.11+ |

## Documentation

<div class="grid cards" markdown>

-   :material-cube-outline:{ .lg .middle } __Architecture__

    ---

    How ffire works internally

    [:octicons-arrow-right-24: Learn more](architecture/index.md)

-   :material-console:{ .lg .middle } __CLI Reference__

    ---

    Command-line usage and options

    [:octicons-arrow-right-24: CLI docs](api/cli.md)

-   :material-speedometer:{ .lg .middle } __Benchmarks__

    ---

    Full performance comparison

    [:octicons-arrow-right-24: See results](development/benchmarks.md)

-   :material-file-document:{ .lg .middle } __Schema Format__

    ---

    Define your message types

    [:octicons-arrow-right-24: Schema docs](architecture/schema-format.md)

</div>

## License

MIT
