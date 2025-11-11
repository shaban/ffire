# Protocol Buffers vs ffire: Performance Comparison

Comparing protobuf (Go and C#) against ffire on the `complex` benchmark.

## Results

### Protocol Buffers (Go)
```
encode_ns: 8,582
decode_ns: 14,515
total_ns:  23,097
wire_size: 3,921 bytes
```

### Protocol Buffers (C#)
```
encode_ns: 9,857
decode_ns: 10,116
total_ns:  19,973
wire_size: 3,921 bytes
```

### ffire (C#)
```
encode_ns: 22,913
decode_ns: 14,966
total_ns:  37,879
wire_size: 4,293 bytes
```

### ffire (C++)
```
encode_ns: 4,986
decode_ns: 3,365
total_ns:  8,351
wire_size: 4,293 bytes
```

## Analysis

**C# protobuf vs ffire C#:**
- protobuf encode: 9,857 ns (0.43x - **57% faster** than ffire!)
- protobuf decode: 10,116 ns (0.68x - **32% faster** than ffire!)
- protobuf total:  19,973 ns (0.53x - **47% faster** overall)

**Go protobuf vs C# protobuf:**
- Go encode: 8,582 ns (0.87x - 13% faster)
- Go decode: 14,515 ns (1.43x slower)
- Go total:  23,097 ns (1.16x slower)

**C# protobuf vs C++ ffire:**
- protobuf encode: 9,857 ns (1.98x slower than C++)
- protobuf decode: 10,116 ns (3.01x slower than C++)
- protobuf total:  19,973 ns (2.39x slower than C++)

**ffire C# vs C++ ffire:**
- ffire C# encode: 22,913 ns (4.60x slower than C++)
- ffire C# decode: 14,966 ns (4.45x slower than C++)
- ffire C# total:  37,879 ns (4.54x slower than C++)

## Key Findings

1. **Protobuf C# is significantly faster than ffire C#**: 47% faster overall (19,973 vs 37,879 ns)
   - 57% faster encode (9,857 vs 22,913 ns)
   - 32% faster decode (10,116 vs 14,966 ns)

2. **Both ffire C# and protobuf C# show similar ratios vs C++**:
   - Protobuf C#: 2.39x slower than ffire C++
   - ffire C#: 4.54x slower than ffire C++
   - The ~2x gap is the difference in implementation maturity

3. **Wire size**: Almost identical (ffire: 4,293 bytes, protobuf: 3,921 bytes = 9% difference)

4. **Go protobuf is competitive with C# protobuf**: Only 16% slower overall, faster encode

## Conclusions

The protobuf C# comparison reveals that **ffire's C# implementation has significant room for optimization**:

- **4.54x vs C++** is the current ffire C# ratio on complex
- **2.39x vs C++** is what mature protobuf C# achieves
- This suggests ffire C# could potentially reach **~2.5x vs C++** with similar optimizations

The gap is likely due to:
- Protobuf's highly optimized code generation patterns
- More aggressive use of unsafe code and direct memory manipulation
- Years of production optimization in the protobuf C# generator

**Current status**: ffire C# is functional and ~2x slower than the industry-standard protobuf implementation. This is acceptable for an initial release, but there's clear potential for 2x improvement by studying protobuf's techniques.
