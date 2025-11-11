# C# UTF-8 Native Structs Plan

## THE REAL PROBLEM
C# strings are UTF-16 internally, wire format is UTF-8.
Current approach does UTF-8 conversions on EVERY encode/decode:
- Decode: UTF-8 bytes → UTF-16 string (Encoding.UTF8.GetString)
- Encode: UTF-16 string → UTF-8 bytes (Encoding.UTF8.GetBytes)

Result: 29,572 ns encode + 16,077 ns decode = 45,649 ns total

## THE SOLUTION: Store UTF-8 Bytes Internally (Like Go Does!)

Go strings are UTF-8 byte slices → ZERO encoding during encode/decode!

We can do the same in C# by storing byte[] internally with string property wrappers:

```csharp
public struct Parameter {
    // Internal UTF-8 storage (used for encoding/decoding)
    internal byte[] _displayNameUtf8;
    
    // User-facing string property (normal C# usage)
    public string DisplayName {
        get => _displayNameUtf8 == null ? "" : Encoding.UTF8.GetString(_displayNameUtf8);
        set => _displayNameUtf8 = Encoding.UTF8.GetBytes(value ?? "");
    }
}
```

## IMPLEMENTATION STEPS

1. **Change field generation: Use byte[] with string property wrapper**
   ```csharp
   internal byte[] _displayNameUtf8;
   public string DisplayName {
       get => _displayNameUtf8 == null ? "" : Encoding.UTF8.GetString(_displayNameUtf8);
       set => _displayNameUtf8 = Encoding.UTF8.GetBytes(value ?? "");
   }
   ```

2. **Encode: NO ComputeSize(), allocate max, write directly, trim**
   ```csharp
   byte[] Encode() {
       int maxSize = ComputeMaxSize();  // Fast: string.Length * 3
       byte[] buffer = new byte[maxSize];
       int offset = 0;
       EncodeTo(buffer, ref offset);
       if (offset < maxSize) Array.Resize(ref buffer, offset);
       return buffer;
   }
   
   void EncodeTo(Span<byte> buffer, ref int offset) {
       // Just copy UTF-8 bytes (NO encoding!)
       BinaryPrimitives.WriteUInt16LittleEndian(buffer.Slice(offset), (ushort)_displayNameUtf8.Length);
       _displayNameUtf8.CopyTo(buffer.Slice(offset + 2));
       offset += 2 + _displayNameUtf8.Length;
   }
   ```

3. **Decode: Copy bytes directly (NO decoding!)**
   ```csharp
   static Parameter DecodeFrom(ReadOnlySpan<byte> buffer, ref int offset) {
       var obj = new Parameter();
       int len = BinaryPrimitives.ReadUInt16LittleEndian(buffer.Slice(offset));
       offset += 2;
       obj._displayNameUtf8 = buffer.Slice(offset, len).ToArray();
       offset += len;
       return obj;
   }
   ```

4. **For arrays: Same approach - copy UTF-8 bytes**
   - Array string elements also just copy bytes
   - NO UTF-8 conversions anywhere!

## WHAT THIS ACHIEVES
- **ZERO UTF-8 encoding during Encode()** - just copy bytes (like Go!)
- **ZERO UTF-8 decoding during Decode()** - just copy bytes (like Go!)
- User-facing properties still work as normal C# strings
- UTF-8 conversion only happens if/when user accesses string properties
- For pure encode/decode benchmarks: ZERO conversions!

Expected results:
- Encode: ~5,000-8,000 ns (close to Go's 4,900 ns)
- Decode: ~4,000-6,000 ns (close to Go's 4,300 ns)
- Total: ~9,000-14,000 ns vs current 45,649 ns = **3-5x faster!**
