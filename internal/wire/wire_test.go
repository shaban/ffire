package wire

import (
	"bytes"
	"testing"
)

func TestRoundTripBool(t *testing.T) {
	tests := []bool{true, false}
	for _, want := range tests {
		buf := &bytes.Buffer{}
		EncodeBool(buf, want)

		got, err := DecodeBool(bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Fatalf("DecodeBool failed: %v", err)
		}
		if got != want {
			t.Errorf("bool round-trip: got %v, want %v", got, want)
		}
	}
}

func TestRoundTripInt8(t *testing.T) {
	tests := []int8{-128, -1, 0, 1, 127}
	for _, want := range tests {
		buf := &bytes.Buffer{}
		EncodeInt8(buf, want)

		got, err := DecodeInt8(bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Fatalf("DecodeInt8 failed: %v", err)
		}
		if got != want {
			t.Errorf("int8 round-trip: got %v, want %v", got, want)
		}
	}
}

func TestRoundTripInt16(t *testing.T) {
	tests := []int16{-32768, -1, 0, 1, 32767}
	for _, want := range tests {
		buf := &bytes.Buffer{}
		EncodeInt16(buf, want)

		got, err := DecodeInt16(bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Fatalf("DecodeInt16 failed: %v", err)
		}
		if got != want {
			t.Errorf("int16 round-trip: got %v, want %v", got, want)
		}
	}
}

func TestRoundTripInt32(t *testing.T) {
	tests := []int32{-2147483648, -1, 0, 1, 2147483647}
	for _, want := range tests {
		buf := &bytes.Buffer{}
		EncodeInt32(buf, want)

		got, err := DecodeInt32(bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Fatalf("DecodeInt32 failed: %v", err)
		}
		if got != want {
			t.Errorf("int32 round-trip: got %v, want %v", got, want)
		}
	}
}

func TestRoundTripInt64(t *testing.T) {
	tests := []int64{-9223372036854775808, -1, 0, 1, 9223372036854775807}
	for _, want := range tests {
		buf := &bytes.Buffer{}
		EncodeInt64(buf, want)

		got, err := DecodeInt64(bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Fatalf("DecodeInt64 failed: %v", err)
		}
		if got != want {
			t.Errorf("int64 round-trip: got %v, want %v", got, want)
		}
	}
}

func TestRoundTripFloat32(t *testing.T) {
	tests := []float32{-1.5, 0.0, 1.5, 3.14159}
	for _, want := range tests {
		buf := &bytes.Buffer{}
		EncodeFloat32(buf, want)

		got, err := DecodeFloat32(bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Fatalf("DecodeFloat32 failed: %v", err)
		}
		if got != want {
			t.Errorf("float32 round-trip: got %v, want %v", got, want)
		}
	}
}

func TestRoundTripFloat64(t *testing.T) {
	tests := []float64{-1.5, 0.0, 1.5, 3.14159265359}
	for _, want := range tests {
		buf := &bytes.Buffer{}
		EncodeFloat64(buf, want)

		got, err := DecodeFloat64(bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Fatalf("DecodeFloat64 failed: %v", err)
		}
		if got != want {
			t.Errorf("float64 round-trip: got %v, want %v", got, want)
		}
	}
}

func TestRoundTripString(t *testing.T) {
	tests := []string{
		"",
		"hello",
		"Hello, 世界",
		"device_0001_with_some_extra_text",
	}
	for _, want := range tests {
		buf := &bytes.Buffer{}
		EncodeString(buf, want)

		got, err := DecodeString(bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Fatalf("DecodeString failed for %q: %v", want, err)
		}
		if got != want {
			t.Errorf("string round-trip: got %q, want %q", got, want)
		}
	}
}

func TestRoundTripArrayHeader(t *testing.T) {
	tests := []uint16{0, 1, 100, 5000, 65535} // Max changed from 4294967295 to 65535 (uint16)
	for _, want := range tests {
		buf := &bytes.Buffer{}
		EncodeArrayHeader(buf, want)

		got, err := DecodeArrayHeader(bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Fatalf("DecodeArrayHeader failed: %v", err)
		}
		if got != want {
			t.Errorf("array header round-trip: got %v, want %v", got, want)
		}
	}
}

// Test wire format according to spec
func TestWireFormatSpec(t *testing.T) {
	// Example from wire-format.md:
	// Device{Name: "Speaker", Channels: 2}
	// Expected wire bytes (uint16 for length):
	// 07 00                    # string length = 7 (uint16 LE)
	// 53 70 65 61 6B 65 72     # "Speaker" (UTF-8)
	// 02 00 00 00              # channels = 2 (int32 LE)

	buf := &bytes.Buffer{}
	EncodeString(buf, "Speaker")
	EncodeInt32(buf, 2)

	want := []byte{
		0x07, 0x00, // length = 7 (uint16 instead of uint32)
		0x53, 0x70, 0x65, 0x61, 0x6B, 0x65, 0x72, // "Speaker"
		0x02, 0x00, 0x00, 0x00, // channels = 2
	}

	got := buf.Bytes()
	if !bytes.Equal(got, want) {
		t.Errorf("wire format mismatch:\ngot:  %x\nwant: %x", got, want)
	}
}
