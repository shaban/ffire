# C# Native Implementation Research

**Goal**: Design high-performance native C# implementation matching Tier A performance (within 2x of Go/C++)

## Current Status
- C#: Has P/Invoke-based benchmark (via C ABI)
- Java: ✅ Native implementation complete, 10/10 passing (1.47x C++ baseline)
- Target: Replace P/Invoke with pure C# for better performance and distribution

---

## 1. Type Safety

### Modern C# Type System (C# 9.0+)
```csharp
// Nullable reference types (enabled by default in .NET 6+)
#nullable enable

public class ConfigMessage
{
    // Non-null by default
    public string Name { get; set; } = "";
    
    // Optional fields use ? suffix
    public string? Description { get; set; }
    public int? MaxRetries { get; set; }
    
    // Arrays never null, but can be empty
    public List<Plugin> Plugins { get; set; } = new();
}
```

### Record Types for Immutability (C# 9+)
```csharp
// Value-based equality, immutable by default
public record PluginMessage(
    string Name,
    string Version,
    bool Enabled
);

// With mutation via 'with' expression
var updated = original with { Enabled = false };
```

### Init-only Properties (C# 9+)
```csharp
public class Message
{
    // Can only be set during object initialization
    public string Name { get; init; } = "";
    public int Version { get; init; }
}

var msg = new Message { Name = "test", Version = 1 };
// msg.Name = "x";  // Compile error!
```

---

## 2. Bulk Operations & Zero-Copy

### Span<T> and Memory<T> (.NET Core 2.1+)
Zero-copy slicing and bulk operations:

```csharp
public static void EncodePrimitiveArray(Span<byte> buffer, int[] values)
{
    // Zero-copy reinterpretation: int[] → Span<byte>
    Span<byte> valueBytes = MemoryMarshal.AsBytes(values.AsSpan());
    
    // Bulk copy to output buffer
    valueBytes.CopyTo(buffer);
}

public static int[] DecodePrimitiveArray(ReadOnlySpan<byte> buffer, int count)
{
    // Zero-copy reinterpretation: byte[] → int[]
    return MemoryMarshal.Cast<byte, int>(buffer.Slice(0, count * 4)).ToArray();
}
```

### ArrayPool<T> for Reduced Allocations
```csharp
using System.Buffers;

public byte[] Encode(ConfigMessage msg)
{
    // Rent buffer from pool instead of allocating
    byte[] buffer = ArrayPool<byte>.Shared.Rent(4096);
    try
    {
        int written = EncodeToBuffer(buffer, msg);
        return buffer[..written];  // Return only used portion
    }
    finally
    {
        ArrayPool<byte>.Shared.Return(buffer);
    }
}
```

### BinaryPrimitives for Endianness
```csharp
using System.Buffers.Binary;

public static void WriteInt32(Span<byte> buffer, int value)
{
    // Little-endian write, no allocations
    BinaryPrimitives.WriteInt32LittleEndian(buffer, value);
}

public static int ReadInt32(ReadOnlySpan<byte> buffer)
{
    return BinaryPrimitives.ReadInt32LittleEndian(buffer);
}
```

---

## 3. Performance Optimizations

### Aggressive Inlining
```csharp
using System.Runtime.CompilerServices;

[MethodImpl(MethodImplOptions.AggressiveInlining)]
private static void WriteVarint32(Span<byte> buffer, ref int offset, int value)
{
    // Inline small hot-path methods
    buffer[offset++] = (byte)value;
    buffer[offset++] = (byte)(value >> 8);
    buffer[offset++] = (byte)(value >> 16);
    buffer[offset++] = (byte)(value >> 24);
}
```

### Unsafe Code for Maximum Performance
```csharp
public unsafe byte[] EncodeFast(int[] values)
{
    byte[] result = new byte[values.Length * 4];
    
    fixed (int* src = values)
    fixed (byte* dst = result)
    {
        // Direct memory copy, fastest possible
        Buffer.MemoryCopy(src, dst, result.Length, result.Length);
    }
    
    return result;
}
```

### stackalloc for Small Buffers
```csharp
public void EncodeSmallMessage()
{
    // Stack allocation, no GC pressure
    Span<byte> buffer = stackalloc byte[256];
    
    int written = EncodeToSpan(buffer);
    // Use buffer...
}
```

### CollectionsMarshal for List<T> Access (.NET 5+)
```csharp
using System.Runtime.InteropServices;

public static void ProcessList(List<int> values)
{
    // Zero-copy access to List<T> internal array
    Span<int> span = CollectionsMarshal.AsSpan(values);
    
    // Direct manipulation without bounds checks
    for (int i = 0; i < span.Length; i++)
    {
        span[i] *= 2;
    }
}
```

---

## 4. Implementation Strategy

### Tier A: Pure C# Native Implementation

**Pattern**: Follow Java generator architecture

```csharp
// Generated message class
public class ConfigMessage
{
    public string Name { get; set; } = "";
    public int Version { get; set; }
    public List<PluginMessage> Plugins { get; set; } = new();
    
    // Encode to Span<byte> for zero-copy
    public int Encode(Span<byte> buffer)
    {
        int offset = 0;
        
        // Encode string length + data
        BinaryPrimitives.WriteInt32LittleEndian(buffer[offset..], Name.Length);
        offset += 4;
        
        int bytesWritten = Encoding.UTF8.GetBytes(Name, buffer[offset..]);
        offset += bytesWritten;
        
        // Encode version
        BinaryPrimitives.WriteInt32LittleEndian(buffer[offset..], Version);
        offset += 4;
        
        // Encode array length
        BinaryPrimitives.WriteInt32LittleEndian(buffer[offset..], Plugins.Count);
        offset += 4;
        
        // Encode each plugin
        foreach (var plugin in Plugins)
        {
            offset += plugin.Encode(buffer[offset..]);
        }
        
        return offset;
    }
    
    // Decode from ReadOnlySpan<byte> for zero-copy
    public static ConfigMessage Decode(ReadOnlySpan<byte> data)
    {
        var msg = new ConfigMessage();
        int offset = 0;
        
        // Decode string
        int nameLen = BinaryPrimitives.ReadInt32LittleEndian(data[offset..]);
        offset += 4;
        msg.Name = Encoding.UTF8.GetString(data.Slice(offset, nameLen));
        offset += nameLen;
        
        // Decode version
        msg.Version = BinaryPrimitives.ReadInt32LittleEndian(data[offset..]);
        offset += 4;
        
        // Decode array
        int pluginCount = BinaryPrimitives.ReadInt32LittleEndian(data[offset..]);
        offset += 4;
        
        for (int i = 0; i < pluginCount; i++)
        {
            var plugin = PluginMessage.Decode(data[offset..]);
            msg.Plugins.Add(plugin);
            offset += plugin.EncodedSize;
        }
        
        return msg;
    }
}
```

---

## 5. Primitive Array Optimizations

### Using MemoryMarshal for Bulk Copy
```csharp
public static void EncodeIntArray(Span<byte> buffer, int[] values)
{
    // Zero-copy conversion: int[] → byte[]
    ReadOnlySpan<int> valueSpan = values.AsSpan();
    ReadOnlySpan<byte> byteSpan = MemoryMarshal.AsBytes(valueSpan);
    
    // Write length
    BinaryPrimitives.WriteInt32LittleEndian(buffer, values.Length);
    
    // Bulk copy data (single memory operation)
    byteSpan.CopyTo(buffer[4..]);
}

public static int[] DecodeIntArray(ReadOnlySpan<byte> buffer)
{
    // Read length
    int count = BinaryPrimitives.ReadInt32LittleEndian(buffer);
    
    // Zero-copy conversion: byte[] → int[]
    return MemoryMarshal.Cast<byte, int>(buffer[4..(4 + count * 4)]).ToArray();
}
```

### Specialized Slice Classes (Like Java)
```csharp
// Primitive array wrapper with Go-like API
public class IntSlice
{
    private int[] data;
    
    public IntSlice(int capacity)
    {
        data = new int[capacity];
    }
    
    public IntSlice(int[] array)
    {
        data = array;
    }
    
    public int Length => data.Length;
    
    public int this[int index]
    {
        get => data[index];
        set => data[index] = value;
    }
    
    public Span<int> AsSpan() => data.AsSpan();
    
    public IntSlice Append(params int[] values)
    {
        int[] newData = new int[data.Length + values.Length];
        data.CopyTo(newData, 0);
        values.CopyTo(newData, data.Length);
        return new IntSlice(newData);
    }
    
    public IntSlice Slice(int start, int end)
    {
        return new IntSlice(data[start..end]);
    }
}
```

---

## 6. Memory Management

### IDisposable Pattern for Large Messages
```csharp
public class LargeMessage : IDisposable
{
    private byte[] buffer;
    
    public LargeMessage()
    {
        buffer = ArrayPool<byte>.Shared.Rent(65536);
    }
    
    public void Dispose()
    {
        ArrayPool<byte>.Shared.Return(buffer);
    }
}

// Usage with 'using' statement
using (var msg = new LargeMessage())
{
    // Use message...
}  // Automatically returned to pool
```

### Struct vs Class Trade-offs
```csharp
// Small immutable messages: use struct (stack allocated, no GC)
public readonly struct Point
{
    public int X { get; init; }
    public int Y { get; init; }
}

// Large or mutable messages: use class (heap allocated, GC managed)
public class ComplexMessage
{
    public string Name { get; set; }
    public List<Plugin> Plugins { get; set; }
}
```

---

## 7. .NET Runtime Optimizations

### Target Framework: .NET 6.0+
- **Span<T>**: Zero-copy slicing and bulk operations
- **MemoryMarshal**: Unsafe but fast type reinterpretation
- **BinaryPrimitives**: Endianness-aware reads/writes
- **ArrayPool<T>**: Reduce GC pressure
- **Aggressive inlining**: JIT optimizations
- **CollectionsMarshal**: Direct List<T> access

### Performance Characteristics
- **RyuJIT**: Modern JIT compiler, near-C++ performance
- **Tiered Compilation**: Quick startup, optimized steady-state
- **GC**: Gen0 collections ~1ms, minimize allocations
- **SIMD**: Auto-vectorization for array operations

---

## 8. Comparison: C# vs Java

| Feature | C# | Java |
|---------|----|----|
| Zero-copy slicing | `Span<T>`, `Memory<T>` | `ByteBuffer` (heap alloc) |
| Primitive arrays | `int[]` with `MemoryMarshal` | Custom slice classes |
| Unsafe code | `unsafe` block, pointers | Not available (without JNI) |
| Stack allocation | `stackalloc` | Not available |
| Nullable types | Built-in (`string?`) | Requires `Optional<T>` |
| Immutability | `record`, `init` | `final` fields |
| Memory pooling | `ArrayPool<T>` | Manual or Netty |

**Verdict**: C# has **more powerful** zero-copy primitives than Java

---

## 9. Expected Performance

### Java Baseline (Validated)
- **Decode**: 6,125 ns/op (1.47x C++ baseline)
- **Architecture**: Native ByteBuffer, no JNI

### C# Target (Estimated)
- **Decode**: 4,000-6,000 ns/op (1.0-1.5x C++ baseline)
- **Rationale**:
  - `Span<T>` + `MemoryMarshal` = zero allocations
  - Aggressive inlining + RyuJIT = near-C++ codegen
  - No FFI overhead (vs current P/Invoke implementation)
  - Stack allocation for small buffers
- **Expected**: **Faster than Java** due to better zero-copy primitives

---

## 10. Implementation Checklist

### Phase 1: Generator Foundation
- [ ] Create `generator_csharp_native.go`
- [ ] Generate message classes with properties
- [ ] Implement `Encode(Span<byte>)` methods
- [ ] Implement `Decode(ReadOnlySpan<byte>)` static methods

### Phase 2: Type System
- [ ] Primitives: bool, sbyte, short, int, long, float, double, string
- [ ] Optionals: Nullable<T> for value types, string? for reference types
- [ ] Arrays: List<T> with CollectionsMarshal optimization
- [ ] Nested: Recursive message encoding/decoding

### Phase 3: Optimizations
- [ ] Use `BinaryPrimitives` for endianness
- [ ] Use `MemoryMarshal` for primitive array bulk copy
- [ ] Add `[MethodImpl(AggressiveInlining)]` to hot paths
- [ ] Use `stackalloc` for small temporary buffers
- [ ] Optional: `ArrayPool<T>` for large messages

### Phase 4: Benchmark & Validation
- [ ] Create benchmark harness (native C#, no P/Invoke)
- [ ] Integrate with mage build system
- [ ] Test all 10 schemas
- [ ] Validate performance: target 4,000-6,000 ns/op (1.0-1.5x C++)

---

## 11. Key Insights from Java Implementation

From `generator_java.go` line 79:
```
// Slice classes provide 11x faster encoding vs ArrayList<Integer>
// and 4.25x better memory efficiency
```

**Lesson for C#**:
1. **Primitive arrays are critical** - avoid boxing (List<int> not List<Int32>)
2. **Direct memory access** - use `Span<T>` and `MemoryMarshal`
3. **Bulk operations** - single copy for entire array, not element-by-element
4. **Stack allocation** - use `stackalloc` for temporary buffers

---

## 12. References

### Official Documentation
- [Span<T> Documentation](https://learn.microsoft.com/en-us/dotnet/api/system.span-1)
- [Memory<T> and Span<T> usage guidelines](https://learn.microsoft.com/en-us/dotnet/standard/memory-and-spans/)
- [BinaryPrimitives](https://learn.microsoft.com/en-us/dotnet/api/system.buffers.binary.binaryprimitives)
- [MemoryMarshal](https://learn.microsoft.com/en-us/dotnet/api/system.runtime.interopservices.memorymarshal)
- [ArrayPool<T>](https://learn.microsoft.com/en-us/dotnet/api/system.buffers.arraypool-1)

### Performance Resources
- [Writing High-Performance .NET Code](https://github.com/adamsitnik/awesome-dot-net-performance)
- [.NET Performance Blog](https://devblogs.microsoft.com/dotnet/category/performance/)
- [BenchmarkDotNet](https://benchmarkdotnet.org/) - for micro-benchmarking

---

## Summary

**C# native implementation should be FASTER than Java** due to:
1. ✅ **Span<T>** - true zero-copy slicing (vs Java's ByteBuffer heap allocation)
2. ✅ **MemoryMarshal** - unsafe type reinterpretation without JNI
3. ✅ **stackalloc** - stack allocation for small buffers
4. ✅ **BinaryPrimitives** - optimized endianness handling
5. ✅ **AggressiveInlining** - JIT hint for hot paths

**Target Performance**: 4,000-6,000 ns/op (1.0-1.5x C++ baseline)
**Better than**: Java 6,125 ns/op (1.47x)
**Comparison**: Pure C# faster than Java, comparable to Go (5,424 ns, 1.30x)
