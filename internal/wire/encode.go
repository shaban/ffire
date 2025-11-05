// Package wire implements the ffire wire format encoding and decoding.
// This is internal tooling used by fixture generation and benchmarking.
// Generated code does NOT import this - it's self-contained.
package wire

import (
	"bytes"
	"encoding/binary"
)

// Wire format version: Uses uint16 for string/array lengths (max 65,535)
// This provides safety by design - physically impossible to overflow

// EncodeBool encodes a boolean value.
func EncodeBool(buf *bytes.Buffer, v bool) {
	if v {
		buf.WriteByte(0x01)
	} else {
		buf.WriteByte(0x00)
	}
}

// EncodeInt8 encodes an int8 as 1 byte.
func EncodeInt8(buf *bytes.Buffer, v int8) {
	buf.WriteByte(byte(v))
}

// EncodeInt16 encodes an int16 as 2 bytes (little-endian).
func EncodeInt16(buf *bytes.Buffer, v int16) {
	binary.Write(buf, binary.LittleEndian, v)
}

// EncodeInt32 encodes an int32 as 4 bytes (little-endian).
func EncodeInt32(buf *bytes.Buffer, v int32) {
	binary.Write(buf, binary.LittleEndian, v)
}

// EncodeInt64 encodes an int64 as 8 bytes (little-endian).
func EncodeInt64(buf *bytes.Buffer, v int64) {
	binary.Write(buf, binary.LittleEndian, v)
}

// EncodeFloat32 encodes a float32 as 4 bytes (IEEE 754, little-endian).
func EncodeFloat32(buf *bytes.Buffer, v float32) {
	binary.Write(buf, binary.LittleEndian, v)
}

// EncodeFloat64 encodes a float64 as 8 bytes (IEEE 754, little-endian).
func EncodeFloat64(buf *bytes.Buffer, v float64) {
	binary.Write(buf, binary.LittleEndian, v)
}

// EncodeString encodes a string as [uint16_le: byte_length][utf8_bytes...].
// No null terminator. Empty string is encoded as 0x00 0x00.
// Max length: 65,535 bytes (enforced by validator).
func EncodeString(buf *bytes.Buffer, s string) {
	binary.Write(buf, binary.LittleEndian, uint16(len(s)))
	buf.WriteString(s)
}

// EncodeArrayHeader encodes an array length as uint16_le.
// This should be called before encoding array elements.
// Max count: 65,535 elements (enforced by validator).
func EncodeArrayHeader(buf *bytes.Buffer, count uint16) {
	binary.Write(buf, binary.LittleEndian, count)
}
