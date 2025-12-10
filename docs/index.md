# ffire

Fast, compact binary serialization format with generated codecs.

## Quick Start

```bash
# Install
go install github.com/shaban/ffire/cmd/ffire@latest

# Generate
ffire gen --lang go --schema schema.ffi --output ./generated

# Use
import "your-module/generated"

data, _ := Message.Encode(msg)
msg, _ := Message.Decode(data)
```

## Features

- **Fast**: Optimized bulk array encoding, zero-copy where possible
- **Compact**: Efficient wire format, minimal overhead
- **Multi-language**: Go, C++, C#, Java, Swift, Dart, Rust, Zig
- **Type-safe**: Strongly typed schemas in Go syntax
- **Portable**: Cross-platform, cross-language serialization

## Performance

Array of 5000 float32 values (encode + decode):

| Language | Total   | 
|----------|---------|
| Rust     | 967 ns  |
| C++      | 1,321 ns|
| Swift    | 1,339 ns|
| Zig      | 1,632 ns|
| C#       | 1,701 ns|
| Java     | 2,343 ns|
| Go       | 3,384 ns|
| Dart     | 10,188 ns|

See [Benchmarks](development/benchmarks.md) for full results.

## Documentation

- [Architecture Overview](architecture/overview.md) - How ffire works
- [Schema Format](architecture/schema-format.md) - Define your types
- [CLI Reference](api/cli.md) - Command-line usage
- [Development Guide](development/testing.md) - Contributing

## License

MIT
