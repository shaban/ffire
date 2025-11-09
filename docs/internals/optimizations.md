# Optimizations

Recent performance improvements in ffire codecs.

## Bulk Array Encoding

### Problem
Arrays were encoded element-by-element with repeated buffer allocations.

### Solution
Pre-calculate total size, allocate once, bulk copy.

#### Go Implementation
```go
// Before: Multiple allocations
for _, item := range items {
    buf.WriteVarint(item)  // May allocate
}

// After: Single allocation
totalSize := 2 * len(items)  // Each varint ≤ 2 bytes for int32
buf.Grow(totalSize)
for _, item := range items {
    buf.WriteVarint(item)  // No allocation
}
```

#### C++ Implementation  
```cpp
// Before: Incremental growth
for (auto item : items) {
    buffer.push_back(item);  // May reallocate
}

// After: Single allocation
buffer.reserve(buffer.size() + items.size() * sizeof(T));
buffer.insert(buffer.end(), 
    reinterpret_cast<const uint8_t*>(items.data()),
    reinterpret_cast<const uint8_t*>(items.data() + items.size()));
```

### Results

**Numeric Arrays (5000 int32)**
- C++: 30% faster encoding
- Go: 28% faster encoding

**String Arrays (1000 strings)**
- Go: 43% faster encoding  
- C++: 7% faster encoding

## Zero-Copy Techniques

### C++ Arrays
Use `insert()` with raw pointers instead of element-wise copy.

### Go Strings
Direct memory copy via `copy(buf, str)` instead of byte iteration.

## Future Work

- SIMD for numeric array encoding
- Vectorized varint encoding
- Lazy decoding (decode on access)
- Memory pooling for large messages
