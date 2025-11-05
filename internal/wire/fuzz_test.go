package wire

import (
	"bytes"
	"testing"
)

// FuzzDecodeString tests string decoding with malformed length prefixes
func FuzzDecodeString(f *testing.F) {
	// Seed with valid strings
	f.Add([]byte{0x05, 0x00, 'h', 'e', 'l', 'l', 'o'}) // "hello"
	f.Add([]byte{0x00, 0x00})                          // ""
	f.Add([]byte{0x01, 0x00, 'a'})                     // "a"

	// Malformed
	f.Add([]byte{0xFF, 0xFF})           // Huge length claim, no data
	f.Add([]byte{0x10, 0x00, 'a', 'b'}) // Length 16, only 2 bytes
	f.Add([]byte{0x05, 0x00})           // Length 5, no data

	f.Fuzz(func(t *testing.T, data []byte) {
		r := bytes.NewReader(data)
		_, err := DecodeString(r)
		// We expect errors on malformed input, but no panics
		_ = err
	})
}

// FuzzDecodeInt32 tests int32 decoding with insufficient data
func FuzzDecodeInt32(f *testing.F) {
	f.Add([]byte{0x01, 0x02, 0x03, 0x04}) // Valid int32
	f.Add([]byte{0xFF, 0xFF, 0xFF, 0xFF}) // -1
	f.Add([]byte{0x00, 0x00, 0x00, 0x00}) // 0

	// Malformed
	f.Add([]byte{})                 // Empty
	f.Add([]byte{0x01})             // Only 1 byte
	f.Add([]byte{0x01, 0x02})       // Only 2 bytes
	f.Add([]byte{0x01, 0x02, 0x03}) // Only 3 bytes

	f.Fuzz(func(t *testing.T, data []byte) {
		r := bytes.NewReader(data)
		_, err := DecodeInt32(r)
		_ = err
	})
}

// FuzzDecodeInt64 tests int64 decoding
func FuzzDecodeInt64(f *testing.F) {
	f.Add([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}) // Valid
	f.Add([]byte{})                                               // Empty
	f.Add([]byte{0x01, 0x02, 0x03})                               // Too short

	f.Fuzz(func(t *testing.T, data []byte) {
		r := bytes.NewReader(data)
		_, err := DecodeInt64(r)
		_ = err
	})
}

// FuzzDecodeFloat32 tests float32 decoding
func FuzzDecodeFloat32(f *testing.F) {
	f.Add([]byte{0x00, 0x00, 0x00, 0x00}) // 0.0
	f.Add([]byte{0xC3, 0xF5, 0x48, 0x40}) // 3.14
	f.Add([]byte{})                       // Empty
	f.Add([]byte{0x01, 0x02})             // Too short

	f.Fuzz(func(t *testing.T, data []byte) {
		r := bytes.NewReader(data)
		_, err := DecodeFloat32(r)
		_ = err
	})
}

// FuzzDecodeFloat64 tests float64 decoding
func FuzzDecodeFloat64(f *testing.F) {
	f.Add([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) // 0.0
	f.Add([]byte{})                                               // Empty
	f.Add([]byte{0x01, 0x02, 0x03, 0x04})                         // Too short

	f.Fuzz(func(t *testing.T, data []byte) {
		r := bytes.NewReader(data)
		_, err := DecodeFloat64(r)
		_ = err
	})
}

// FuzzDecodeBool tests bool decoding
func FuzzDecodeBool(f *testing.F) {
	f.Add([]byte{0x00}) // false
	f.Add([]byte{0x01}) // true
	f.Add([]byte{0x02}) // Invalid but should not panic
	f.Add([]byte{0xFF}) // Invalid
	f.Add([]byte{})     // Empty

	f.Fuzz(func(t *testing.T, data []byte) {
		r := bytes.NewReader(data)
		_, err := DecodeBool(r)
		_ = err
	})
}

// FuzzDecodeArrayHeader tests array length decoding
func FuzzDecodeArrayHeader(f *testing.F) {
	f.Add([]byte{0x00, 0x00}) // length 0
	f.Add([]byte{0x05, 0x00}) // length 5
	f.Add([]byte{0xFF, 0xFF}) // length 65535
	f.Add([]byte{})           // Empty
	f.Add([]byte{0x01})       // Only 1 byte

	f.Fuzz(func(t *testing.T, data []byte) {
		r := bytes.NewReader(data)
		_, err := DecodeArrayHeader(r)
		_ = err
	})
}

// FuzzDecodeInt8 tests int8 decoding
func FuzzDecodeInt8(f *testing.F) {
	f.Add([]byte{0x00})
	f.Add([]byte{0x7F})
	f.Add([]byte{0xFF})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		r := bytes.NewReader(data)
		_, err := DecodeInt8(r)
		_ = err
	})
}

// FuzzDecodeInt16 tests int16 decoding
func FuzzDecodeInt16(f *testing.F) {
	f.Add([]byte{0x00, 0x00})
	f.Add([]byte{0xFF, 0xFF})
	f.Add([]byte{})
	f.Add([]byte{0x01})

	f.Fuzz(func(t *testing.T, data []byte) {
		r := bytes.NewReader(data)
		_, err := DecodeInt16(r)
		_ = err
	})
}

// FuzzDecodeMultipleValues tests decoding multiple values in sequence
func FuzzDecodeMultipleValues(f *testing.F) {
	// Seed with valid sequence: bool, int32, string
	validSeq := []byte{
		0x01,                   // bool: true
		0x2A, 0x00, 0x00, 0x00, // int32: 42
		0x04, 0x00, 't', 'e', 's', 't', // string: "test"
	}
	f.Add(validSeq)

	// Partial sequences
	f.Add([]byte{0x01})                                     // Just bool
	f.Add([]byte{0x01, 0x2A, 0x00})                         // Bool + partial int32
	f.Add([]byte{0x01, 0x2A, 0x00, 0x00, 0x00})             // Bool + int32, no string
	f.Add([]byte{0x01, 0x2A, 0x00, 0x00, 0x00, 0x10, 0x00}) // Bool + int32 + string length claim

	f.Fuzz(func(t *testing.T, data []byte) {
		r := bytes.NewReader(data)

		// Try to decode sequence - should handle EOF gracefully
		_, err1 := DecodeBool(r)
		if err1 != nil {
			return // Expected on short data
		}

		_, err2 := DecodeInt32(r)
		if err2 != nil {
			return // Expected on short data
		}

		_, err3 := DecodeString(r)
		_ = err3 // Expected on short data
	})
}
