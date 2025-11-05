# ffire Optimization Plan

## Current Performance Baseline (Data-Driven)

**Before optimizations - All test schemas:**

| Schema | ffire Total | ffire Encode | ffire Decode | Wire Size | Status |
|--------|-------------|--------------|--------------|-----------|--------|
| **struct** | 393 ns | 111 ns | 282 ns | 24 B | ✅ |
| **empty** | 444 ns | 236 ns | 208 ns | 8 B | ✅ |
| **tags** | 619 ns | 281 ns | 338 ns | 34 B | ✅ |
| **array_struct** | 34,034 ns | 15,892 ns | 18,142 ns | 5,202 B | ✅ |
| **complex** | 37,941 ns | 13,745 ns | 24,196 ns | 4,293 B | ✅ |
| **array_string** | 41,306 ns | 15,390 ns | 25,916 ns | 17,002 B | ✅ |
| **optional** | 184,288 ns | 45,782 ns | 138,506 ns | 21,840 B | ⚠️ SLOW |
| array_float | - | - | - | - | ❌ Failed |
| array_int | - | - | - | - | ❌ Failed |
| nested | - | - | - | - | ❌ Failed |

**Protobuf comparison (complex schema only):**
- Protobuf: 24,147 ns total (9,273 encode + 14,874 decode) - 3,921 B
- ffire: 37,941 ns total (13,745 encode + 24,196 decode) - 4,293 B
- **Gap: 1.57x slower than protobuf**

**Target: 2.5x speedup** → ~15,000 ns total (beat protobuf by 38%)

---

## Critical Performance Issues Found

### 1. ❗ `binary.Write()` for every uint16/primitive (BIGGEST ISSUE)

**Current code:**
```go
binary.Write(buf, binary.LittleEndian, uint16(len(elem.Name)))  // SLOW!
```

**Problem:** `binary.Write()` uses reflection and allocates! ~10-20x slower than manual encoding.

**Fix:**
```go
// Instead of binary.Write(buf, binary.LittleEndian, uint16(len))
buf.WriteByte(byte(len))       // low byte
buf.WriteByte(byte(len >> 8))  // high byte
```

**Expected Impact:** 40-50% speedup on both encode/decode

---

### 2. ❗ `buf.WriteString()` does string→[]byte copy

**Current code:**
```go
buf.WriteString(elem.Name)  // Copies string to []byte internally
```

**Problem:** `WriteString()` converts string to `[]byte`, allocating and copying.

**Fix:**
```go
buf.Write([]byte(elem.Name))  // Or use unsafe.StringData() + unsafe.Slice()
```

**Expected Impact:** 5-10% speedup

---

### 3. ❗ `binary.Read()` for every primitive in decoder

**Current code:**
```go
binary.Read(r, binary.LittleEndian, &length1)  // SLOW - uses reflection!
```

**Fix:**
```go
var b [2]byte
r.Read(b[:])
length1 := uint16(b[0]) | uint16(b[1])<<8
```

**Expected Impact:** 40-50% speedup on decode

---

### 4. `bytes.NewReader()` allocates

**Current code:**
```go
r := bytes.NewReader(data)  // Allocates Reader struct
```

**Fix:** Manual index tracking (no allocation):
```go
var pos int
// Then use data[pos:pos+n] slicing
```

**Expected Impact:** 10-15% speedup on decode

---

### 5. Growing buffer allocations

**Current code:**
```go
buf := &bytes.Buffer{}  // Starts small, grows multiple times
```

**Fix:** Pre-allocate with estimated size:
```go
buf := bytes.NewBuffer(make([]byte, 0, 8192))  // Pre-allocate 8KB
```

**Expected Impact:** 10-15% speedup on encode

---

## Summary Table

| Optimization | Encode Speedup | Decode Speedup | Difficulty |
|--------------|----------------|----------------|------------|
| Remove `binary.Write/Read` | **40-50%** | **40-50%** | Easy |
| Pre-allocate buffer | **10-15%** | N/A | Trivial |
| String copy optimization | **5-10%** | **5-10%** | Medium |
| Remove `bytes.Reader` | N/A | **10-15%** | Easy |
| **TOTAL ESTIMATED** | **55-75%** | **55-75%** | - |

---

## Implementation Plan

### Phase 1: Quick Wins (Target: 55-75% speedup)
1. ✅ Replace `binary.Write()` with manual byte writes
2. ✅ Replace `binary.Read()` with manual byte reads  
3. ✅ Pre-allocate encode buffer
4. ✅ Remove `bytes.Reader` allocation in decoder

### Phase 2: Medium Wins (Target: additional 10-20%)
5. String encoding optimization (eliminate copies)
6. Inline small functions
7. Better error handling (remove checks from hot paths)

### Phase 3: Advanced (Target: additional 10-20%)
8. Zero-copy decoding (use views into buffer)
9. Buffer pooling (reuse buffers)
10. SIMD for bulk operations

---

## Notes

- **BLOCKER:** Need data-driven benchmarking first!
- Current benchmark only tests `complex.ffi` with hardcoded paths
- Must traverse all `testdata/schema/*.ffi` and corresponding `.json` files
- Need both ffire and protobuf benchmarks for each schema
- Get comprehensive baseline before starting optimizations

---

## Next Steps

1. 🔴 **Fix benchmarking to be data-driven** (traverse testdata/)
2. Run full baseline benchmarks on all schemas
3. Implement Phase 1 optimizations
4. Re-run benchmarks to measure actual gains
5. Iterate based on results
