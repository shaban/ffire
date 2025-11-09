# Fuzz Testing Guide

## Overview

Fuzz tests are included to ensure that ffire's decoders handle malformed binary data gracefully. The goal is to **return errors, not panic**, even when fed completely corrupted data.

## Why Fuzz Testing?

Decoders are security-critical components. Malformed input could come from:
- Network corruption
- Malicious actors
- Bugs in encoders
- Partial data reads

**Current issue**: Generated decoders use direct slice indexing without bounds checking, which can panic on malformed input.

## Running Fuzz Tests

### Quick Check (Pattern Validation)
```bash
go test -v ./pkg/generator -run Fuzz
```

### Full Fuzzing (Continuous)
```bash
# Fuzz for 30 seconds
go test -fuzz=FuzzDecoder -fuzztime=30s ./pkg/generator

# Fuzz until failure
go test -fuzz=FuzzDecoderArray ./pkg/generator

# Fuzz all tests
go test -fuzz=. ./pkg/generator
```

### Integration Fuzzing (With Generated Code)
```bash
go test -tags=fuzz -fuzz=FuzzGeneratedDecoder ./pkg/generator
```

## What Gets Tested

### 1. String Length Attacks (`FuzzDecoderStringLength`)
- Claims huge string length (65535 bytes) with minimal data
- Partial string data
- Zero-length strings
- **Expected**: Should detect insufficient data and error

### 2. Array Length Attacks (`FuzzDecoderArrayLength`)
- Claims huge array count but provides few elements
- Empty arrays
- Truncated array data
- **Expected**: Should validate array bounds

### 3. Optional Field Handling (`FuzzDecoderOptional`)
- Invalid optional presence flags (not 0x00 or 0x01)
- Missing data after "present" flag
- **Expected**: Should handle all flag values gracefully

### 4. Nested Structures (`FuzzDecoderNested`)
- Incomplete nested data
- Missing inner struct fields
- **Expected**: Should track nesting depth and validate each level

### 5. Truncated Data
- Valid data cut short at any point
- **Expected**: Should detect premature end-of-data

### 6. Garbage Data
- Random bytes
- All zeros
- All ones
- **Expected**: Should fail gracefully with descriptive errors

## Common Issues Found

### ❌ Index Out of Bounds
```go
// BAD: No bounds check
result.Name = string(data[pos : pos+int(length)])
```

**Fix**: Add bounds validation
```go
// GOOD: Bounds check
if pos+int(length) > len(data) {
    return result, fmt.Errorf("insufficient data for string: need %d bytes, have %d", length, len(data)-pos)
}
result.Name = string(data[pos : pos+int(length)])
```

### ❌ Integer Overflow
```go
// BAD: uint16 length could overflow when converted
length := uint16(data[pos]) | uint16(data[pos+1])<<8
// If length is 65535, pos+int(length) might overflow on 32-bit systems
```

**Fix**: Check for overflow
```go
length := uint16(data[pos]) | uint16(data[pos+1])<<8
if int(length) < 0 || pos+int(length) < pos {
    return result, fmt.Errorf("length overflow")
}
```

### ❌ Unchecked Array Allocation
```go
// BAD: Allocate before validating data exists
length := uint16(data[pos]) | uint16(data[pos+1])<<8
tmpSlice := make([]int32, length) // Could allocate 256KB+
// ... then try to fill from insufficient data
```

**Fix**: Validate first
```go
length := uint16(data[pos]) | uint16(data[pos+1])<<8
bytesNeeded := int(length) * 4
if pos+2+bytesNeeded > len(data) {
    return result, fmt.Errorf("insufficient data for array")
}
tmpSlice := make([]int32, length)
```

## Seed Corpus

The fuzz tests include seed corpus to guide fuzzing:
- Valid encoded data (to establish baseline)
- Empty data
- Truncated valid data
- Data with manipulated length prefixes
- Extra trailing bytes

## Integration Testing

For thorough testing, use the integration approach:

1. **Generate** decoder code from schema
2. **Compile** with fuzz harness
3. **Execute** with fuzzed input
4. **Verify** no panics occur

Example workflow:
```bash
# Create test schema
cat > test.ffi << EOF
type TestMessage {
    id: int32
    name: string
}

message TestMessage = TestMessage
EOF

# Generate decoder
ffire generate --schema test.ffi --lang go --output generated.go

# Create fuzz test
cat > fuzz_test.go << 'EOF'
package test

import "testing"

func FuzzDecode(f *testing.F) {
    f.Add([]byte{...}) // Valid data
    
    f.Fuzz(func(t *testing.T, data []byte) {
        defer func() {
            if r := recover(); r != nil {
                t.Fatalf("Decoder panicked: %v", r)
            }
        }()
        
        _, err := DecodeTestMessageMessage(data)
        _ = err // Errors are OK, panics are not
    })
}
EOF

# Run fuzzing
go test -fuzz=FuzzDecode
```

## Next Steps

To fix the issues found by fuzzing:

1. **Add bounds checking** to all array access in decoder generation
2. **Validate length prefixes** before allocation
3. **Check for integer overflow** in length calculations
4. **Add maximum limits** for strings/arrays (e.g., max 1MB per string)
5. **Track position** and ensure it doesn't exceed data length

## Metrics

Good fuzzing coverage should achieve:
- ✅ No panics on any input
- ✅ Descriptive error messages
- ✅ Maximum corpus of ~10,000 inputs tested per second
- ✅ All edge cases (empty, huge, truncated) handled

## Example Error Messages

Good error messages help debugging:

```go
// ❌ Bad: Generic panic
panic: runtime error: slice bounds out of range

// ✅ Good: Descriptive error
return nil, fmt.Errorf("decode string at pos %d: claimed length %d exceeds remaining data %d bytes", pos, length, len(data)-pos)
```

## CI Integration

Add to CI pipeline:
```yaml
- name: Fuzz Test
  run: go test -fuzz=. -fuzztime=60s ./pkg/generator
```

This catches regressions before they reach production.
