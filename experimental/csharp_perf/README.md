# C# Performance Experiments

Testing different C# serialization approaches to inform native generator design.

## Approaches to Test

1. **BinaryWriter/BinaryReader** (traditional)
   - Easy to use, but creates intermediate allocations
   - BufferedStream can help

2. **Span<byte> + MemoryMarshal** (modern, zero-copy)
   - .NET Core 2.1+ feature
   - Direct memory access, no allocations
   - Uses `BitConverter` or `BinaryPrimitives`

3. **Unsafe code with pointers** (fastest)
   - Direct memory manipulation
   - Requires `unsafe` context
   - Best performance but less portable

4. **ArrayPool<byte>** (for arrays)
   - Reuse buffers, reduce GC pressure
   - Important for encoding arrays

## Performance Targets

Based on Java (1.47x C++ baseline):
- Target: < 2.0x C++ baseline
- Stretch: Match Java at ~1.5x

## Key Optimizations

1. **Type Safety**: Use generics and spans
2. **Bulk Operations**: `Buffer.BlockCopy`, `Span.CopyTo`
3. **Zero Allocations**: Stackalloc for small buffers, ArrayPool for large
4. **JIT-friendly**: Inline methods, avoid virtual calls
