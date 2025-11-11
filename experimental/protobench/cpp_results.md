# Protobuf Benchmark Results - FileDescriptorProto Parsing

## C++ Implementation Times (from bazel run //benchmarks:benchmark)

Extracted from benchmark output - Parse FileDesc mean times:

### Upb Implementation (Optimized C)
- `BM_Parse_Upb_FileDesc<CompiledIn, UseArena, Copy>_mean`: **89,937 ns**
- `BM_Parse_Upb_FileDesc<CompiledIn, InitBlock, Copy>_mean`: **91,747 ns**
- `BM_Parse_Upb_FileDesc<Parsed, UseArena, Alias>_mean`: **90,529 ns**
- `BM_Parse_Upb_FileDesc<Parsed, InitBlock, Alias>_mean`: **89,830 ns** ⭐ (fastest)

**Best Upb time: ~90,000 ns (90 µs)**

### Proto2 Implementation (Standard C++)
- `BM_Parse_Proto2<FileDesc, NoArena, Copy>_mean`: **215,114 ns**
- `BM_Parse_Proto2<FileDesc, UseArena, Copy>_mean`: **178,364 ns**
- `BM_Parse_Proto2<FileDesc, InitBlock, Copy>_mean`: **174,689 ns** ⭐ (fastest)
- `BM_Parse_Proto2<FileDescSV, InitBlock, Alias>_mean`: **174,727 ns**

**Best Proto2 time: ~175,000 ns (175 µs)**

## Performance Ratio

**Proto2 / Upb ratio: 175,000 / 90,000 = 1.94x**

Standard C++ implementation is ~2x slower than optimized upb implementation.

## Next Steps

1. Run Python protobuf benchmarks to get Python decode times
2. Calculate Python/C++ ratio in protobuf
3. Compare with our ffire Python/C++ ratio from existing benchmarks
4. Validate if the ratios are similar (indicating similar maturity)

## Benchmark Schema

The benchmark uses `descriptor.proto` (FileDescriptorProto) which is:
- ~7KB serialized size
- Complex nested structure with multiple message types
- Representative of real-world protocol buffer usage
- Used to describe .proto schema definitions
