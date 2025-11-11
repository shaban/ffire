# Protocol Buffers: Go vs C# Performance Comparison

Simple benchmark comparing protobuf performance in Go and C# on the `complex` schema.

## Setup

```bash
# Generate Go code
protoc --go_out=. complex.proto

# Generate C# code  
protoc --csharp_out=. complex.proto

# Install Go dependencies
go mod tidy
```

## Run Benchmarks

```bash
# Go benchmark
go run bench_go.go complex.pb.go

# C# benchmark
dotnet run -c Release
```

## Schema

The `complex.proto` schema matches the ffire `complex.ffi` benchmark - a nested structure with:
- PluginList → Plugin[] → Parameter[]
- 13 fields per Parameter including strings, floats, bools, optional nested message
- Real audio plugin data from complex.json fixture
