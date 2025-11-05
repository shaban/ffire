package wire

import (
	"encoding/binary"
	"fmt"
	"io"
)

// DecodeBool decodes a boolean from 1 byte (0x00 = false, 0x01 = true).
func DecodeBool(r io.Reader) (bool, error) {
	var (
		b   [1]byte
		err error
	)
	_, err = io.ReadFull(r, b[:])
	if err != nil {
		return false, fmt.Errorf("decode bool: %w", err)
	}
	return b[0] != 0x00, nil
}

// DecodeInt8 decodes an int8 from 1 byte.
func DecodeInt8(r io.Reader) (int8, error) {
	var (
		b   [1]byte
		err error
	)
	_, err = io.ReadFull(r, b[:])
	if err != nil {
		return 0, fmt.Errorf("decode int8: %w", err)
	}
	return int8(b[0]), nil
}

// DecodeInt16 decodes an int16 from 2 bytes (little-endian).
func DecodeInt16(r io.Reader) (int16, error) {
	var (
		v   int16
		err error
	)
	err = binary.Read(r, binary.LittleEndian, &v)
	if err != nil {
		return 0, fmt.Errorf("decode int16: %w", err)
	}
	return v, nil
}

// DecodeInt32 decodes an int32 from 4 bytes (little-endian).
func DecodeInt32(r io.Reader) (int32, error) {
	var (
		v   int32
		err error
	)
	err = binary.Read(r, binary.LittleEndian, &v)
	if err != nil {
		return 0, fmt.Errorf("decode int32: %w", err)
	}
	return v, nil
}

// DecodeInt64 decodes an int64 from 8 bytes (little-endian).
func DecodeInt64(r io.Reader) (int64, error) {
	var (
		v   int64
		err error
	)
	err = binary.Read(r, binary.LittleEndian, &v)
	if err != nil {
		return 0, fmt.Errorf("decode int64: %w", err)
	}
	return v, nil
}

// DecodeFloat32 decodes a float32 from 4 bytes (IEEE 754, little-endian).
func DecodeFloat32(r io.Reader) (float32, error) {
	var (
		v   float32
		err error
	)
	err = binary.Read(r, binary.LittleEndian, &v)
	if err != nil {
		return 0, fmt.Errorf("decode float32: %w", err)
	}
	return v, nil
}

// DecodeFloat64 decodes a float64 from 8 bytes (IEEE 754, little-endian).
func DecodeFloat64(r io.Reader) (float64, error) {
	var (
		v   float64
		err error
	)
	err = binary.Read(r, binary.LittleEndian, &v)
	if err != nil {
		return 0, fmt.Errorf("decode float64: %w", err)
	}
	return v, nil
}

// DecodeString decodes a string from [uint16_le: byte_length][utf8_bytes...].
// No bounds checking needed - uint16 physically limits to 65,535 bytes.
func DecodeString(r io.Reader) (string, error) {
	var (
		length uint16
		err    error
	)
	err = binary.Read(r, binary.LittleEndian, &length)
	if err != nil {
		return "", fmt.Errorf("decode string length: %w", err)
	}

	if length == 0 {
		return "", nil
	}

	buf := make([]byte, length)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return "", fmt.Errorf("decode string data: %w", err)
	}

	return string(buf), nil
}

// DecodeArrayHeader decodes an array length from uint16_le.
// No bounds checking needed - uint16 physically limits to 65,535 elements.
func DecodeArrayHeader(r io.Reader) (uint16, error) {
	var (
		count uint16
		err   error
	)
	err = binary.Read(r, binary.LittleEndian, &count)
	if err != nil {
		return 0, fmt.Errorf("decode array header: %w", err)
	}
	return count, nil
}
