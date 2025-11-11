# Java Slice Types for ffire

This directory contains validated, production-ready slice implementations that can be included verbatim in generated Java code.

## Slice Types

| Slice Type | Java Primitive | Schema Types | Element Size |
|------------|----------------|--------------|--------------|
| ByteSlice | byte[] | int8, uint8, byte | 1 byte |
| ShortSlice | short[] | int16, uint16 | 2 bytes |
| IntSlice | int[] | int32, uint32 | 4 bytes |
| LongSlice | long[] | int64, uint64 | 8 bytes |
| FloatSlice | float[] | float32 | 4 bytes |
| DoubleSlice | double[] | float64 | 8 bytes |

## Design Goals

✅ **Zero boxing in codec paths** - encode/decode use primitive arrays directly
✅ **Iterable for ergonomics** - `for (int v : slice)` works (accepts boxing here)
✅ **Go-like API** - `len()`, `get()`, `set()`, `append()`, `slice()`
✅ **Memory safe** - no leaks, proper GC
✅ **Type safe** - bounds checking on all operations

## Performance

Compared to `ArrayList<Integer>`:
- **11x faster encode** (357 ns vs 3,988 ns for 5000 elements)
- **4.25x more memory efficient** (20MB vs 85MB for 1000 slices)
- Zero boxing overhead in codec operations

## API

```java
// Creation
IntSlice slice = new IntSlice(capacity);
IntSlice slice = new IntSlice(int[] array);

// Access
int len = slice.len();
int value = slice.get(index);
slice.set(index, value);

// Slicing (Go-like)
IntSlice sub = slice.slice(start, end);  // [start:end]
IntSlice sub = slice.sliceFrom(start);   // [start:]
IntSlice sub = slice.sliceTo(end);       // [:end]

// Append (immutable)
IntSlice newSlice = slice.append(value);
IntSlice newSlice = slice.appendAll(1, 2, 3);

// Iteration (boxes values)
for (int value : slice) {
    // use value
}

// Direct array access (performance)
int[] arr = slice.array();

// Conversion to List (boxes all)
List<Integer> list = slice.asList();

// Codec (zero boxing)
void encodeTo(ByteBuffer buf);
static IntSlice decodeFrom(ByteBuffer buf, int length);
```

## Tests

- `AllSlicesTest.java` - Basic functionality for all types
- `APITest.java` - API ergonomics demonstration
- `PerformanceTest.java` - vs ArrayList<Integer> benchmark
- `MemoryLeakTest.java` - Memory safety validation
- `ErgonomicsTest.java` - Real-world usage patterns

All tests pass ✓

## Integration

These files can be included verbatim in generated Java code. The generator should:

1. Include the appropriate Slice class(es) based on schema types used
2. Use SliceType instead of `List<Boxed>` for primitive array fields
3. Generate encode/decode using slice.encodeTo()/decodeFrom()
4. Use slice.len() instead of list.size()

## Wire Format Compatibility

Slices use the same wire format as other languages:
- Little-endian byte order
- 2-byte length prefix for arrays
- Direct binary encoding of elements

Compatible with Go, C++, Swift, Python, etc.
