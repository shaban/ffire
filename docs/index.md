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
- **Multi-language**: Go, C++, Swift, Dart, Python, JavaScript, PHP, Ruby
- **Type-safe**: Strongly typed schemas in Go syntax
- **Portable**: Cross-platform, cross-language serialization

## Performance

Array of 5000 int32 values:

| Language | Encode | Decode | Wire Size |
|----------|--------|--------|-----------|
| Go       | 1.6 µs | 8.6 µs | 20 KB     |
| C++      | 1.4 µs | 2.7 µs | 20 KB     |
| Swift    | 1.5 µs | 10.4 µs| 20 KB     |

See [Benchmarks](development/benchmarks.md) for full results.

## Documentation

- [Architecture Overview](architecture/overview.md) - How ffire works
- [Schema Format](architecture/schema-format.md) - Define your types
- [CLI Reference](api/cli.md) - Command-line usage
- [Development Guide](development/testing.md) - Contributing

## License

MIT
