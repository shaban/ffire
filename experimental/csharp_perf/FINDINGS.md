# C# Native Generator - Performance Research Results

## Executive Summary

Based on experimental benchmarks, the optimal C# approach for ffire is:
- **Span<byte> + BinaryPrimitives** for scalar encoding (28x faster than BinaryWriter)
- **MemoryMarshal.Cast** for array operations (45x faster than loop copy)
- **No unsafe code needed** - Span achieves 92% of unsafe performance while remaining safe

## Benchmark Results

### Scalar Serialization (struct with 5 fields)
```
BinaryWriter:     28.03x slower (traditional approach)
Span (modern):    1.00x (baseline) ✅ RECOMMENDED
Unsafe pointers:  0.92x (only 8% faster, not worth complexity)
```

**Decision**: Use **Span<byte> + BinaryPrimitives** for encoding primitives
- Modern, safe, JIT-friendly
- Nearly as fast as unsafe code
- Works on all platforms
- Type-safe with compile-time guarantees

### Array Serialization (1000 int array)
```
Loop copy:            44.72x slower
Buffer.BlockCopy:     1.17x
Span.CopyTo:          1.00x (baseline)
MemoryMarshal.Cast:   1.06x ✅ RECOMMENDED
```

**Decision**: Use **MemoryMarshal.AsBytes()** for bulk array operations
- Zero-copy view into memory
- Handles endianness automatically
- Type-safe reinterpretation
- Optimal for large arrays

## Recommended Architecture

### 1. Type Safety
```csharp
// Use generic encode/decode with constraints
public static void EncodeInt32(Span<byte> buffer, ref int offset, int value)
{
    BinaryPrimitives.WriteInt32LittleEndian(buffer.Slice(offset, 4), value);
    offset += 4;
}

// Array encoding with MemoryMarshal
public static void EncodeInt32Array(Span<byte> buffer, ref int offset, int[] values)
{
    Span<int> intSpan = values;
    Span<byte> byteSpan = MemoryMarshal.AsBytes(intSpan);
    byteSpan.CopyTo(buffer.Slice(offset));
    offset += byteSpan.Length;
}
```

### 2. Bulk Operations
- **Primitives**: Use `BinaryPrimitives.WriteXXX` methods
- **Arrays**: Use `MemoryMarshal.AsBytes()` + `CopyTo()`
- **Strings**: UTF8 encoding with `Encoding.UTF8.GetBytes()`

### 3. Memory Management
- **Small buffers**: `stackalloc byte[256]` for local buffers
- **Large buffers**: `ArrayPool<byte>.Shared.Rent()` to reduce GC
- **Return buffers**: Always return rented arrays with `ArrayPool.Return()`

### 4. JIT Optimizations
- Mark hot methods with `[MethodImpl(MethodImplOptions.AggressiveInlining)]`
- Use `readonly struct` for message types
- Avoid virtual calls in encode/decode paths
- Use `ref` parameters to avoid copying large structs

## Performance Targets

Based on Java achieving **1.47x C++ baseline**:

**Target**: < 2.0x C++ baseline
**Stretch**: Match Java at ~1.5x

**Rationale**:
- RyuJIT has comparable performance to HotSpot JVM
- Span operations are highly optimized in .NET Core 3.0+
- Modern .NET achieves near-native performance for these workloads

## Implementation Strategy

### Phase 1: Core Types (Week 1)
- [ ] Implement primitive encoders (bool, int8-64, float32/64, string)
- [ ] Implement primitive decoders
- [ ] Use Span<byte> + BinaryPrimitives throughout

### Phase 2: Collections (Week 2)  
- [ ] Implement array encoders using MemoryMarshal
- [ ] Implement optional type handling (nullable types)
- [ ] Use ArrayPool for large buffer allocations

### Phase 3: Generator (Week 3)
- [ ] Create generator_csharp.go
- [ ] Generate classes with readonly struct pattern
- [ ] Generate encode/decode methods
- [ ] Add aggressive inlining attributes

### Phase 4: Validation (Week 4)
- [ ] Run all 10 benchmark schemas
- [ ] Verify < 2.0x C++ baseline
- [ ] Profile and optimize hot paths

## Key Learnings

1. **Span is the sweet spot**: 92% of unsafe performance, 100% safety
2. **Avoid BinaryWriter**: 28x slower due to stream overhead
3. **Bulk operations matter**: 45x difference for arrays
4. **Modern .NET is fast**: Can match Java performance with proper patterns

## Next Steps

1. ✅ Research complete - ready to implement
2. Create `pkg/generator/generator_csharp.go`
3. Follow Span<byte> patterns throughout
4. Target .NET 9.0 for latest JIT optimizations
5. Benchmark against Java as reference
