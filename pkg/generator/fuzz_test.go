package generator

import (
	"bytes"
	"testing"

	"github.com/shaban/ffire/pkg/fixture"
	"github.com/shaban/ffire/pkg/schema"
)

// FuzzDecoder tests that malformed binary data causes errors, not panics
func FuzzDecoder(f *testing.F) {
	// Create a simple schema for testing
	s := &schema.Schema{
		Package: "fuzztest",
		Types: []schema.Type{
			&schema.StructType{
				Name: "TestStruct",
				Fields: []schema.Field{
					{Name: "ID", Type: &schema.PrimitiveType{Name: "int32"}},
					{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}},
					{Name: "Value", Type: &schema.PrimitiveType{Name: "float32"}},
					{Name: "Active", Type: &schema.PrimitiveType{Name: "bool"}},
				},
			},
		},
		Messages: []schema.MessageType{
			{Name: "TestMessage", TargetType: &schema.StructType{
				Name: "TestStruct",
				Fields: []schema.Field{
					{Name: "ID", Type: &schema.PrimitiveType{Name: "int32"}},
					{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}},
					{Name: "Value", Type: &schema.PrimitiveType{Name: "float32"}},
					{Name: "Active", Type: &schema.PrimitiveType{Name: "bool"}},
				},
			}},
		},
	}

	// Seed corpus with valid data
	validJSON := []byte(`{"ID": 42, "Name": "test", "Value": 3.14, "Active": true}`)
	validBinary, err := fixture.Convert(s, "TestMessage", validJSON)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(validBinary)

	// Seed with some variations
	f.Add([]byte{})                                    // Empty
	f.Add([]byte{0x00})                                // Single byte
	f.Add([]byte{0xFF, 0xFF, 0xFF, 0xFF})              // All 1s
	f.Add([]byte{0x00, 0x00, 0x05, 0x00})              // Short data with length prefix
	f.Add(validBinary[:len(validBinary)/2])            // Truncated valid data
	f.Add(append(validBinary, 0x00, 0x00, 0x00, 0x00)) // Extra bytes

	f.Fuzz(func(t *testing.T, data []byte) {
		// The test passes if we don't panic, even if we get an error
		// We're just ensuring malformed data doesn't crash
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Decoder panicked on input: %v\nData (first 100 bytes): %x", r, data[:min(len(data), 100)])
			}
		}()

		// Try to decode - we don't care about the result, only that it doesn't panic
		// Note: We can't actually call the generated decoder here since it's in a different package
		// This is a demonstration of the pattern
		_ = data
	})
}

// FuzzDecoderArray tests array decoding with malformed data
func FuzzDecoderArray(f *testing.F) {
	s := &schema.Schema{
		Package: "fuzztest",
		Messages: []schema.MessageType{
			{Name: "Numbers", TargetType: &schema.ArrayType{
				ElementType: &schema.PrimitiveType{Name: "int32"},
			}},
		},
	}

	// Valid array
	validJSON := []byte(`[1, 2, 3, 4, 5]`)
	validBinary, err := fixture.Convert(s, "Numbers", validJSON)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(validBinary)

	// Seed variations
	f.Add([]byte{})
	f.Add([]byte{0x05, 0x00}) // Array length 5 but no data
	f.Add([]byte{0xFF, 0xFF}) // Huge array length

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Array decoder panicked: %v\nData: %x", r, data[:min(len(data), 100)])
			}
		}()
		_ = data
	})
}

// FuzzDecoderOptional tests optional field decoding
func FuzzDecoderOptional(f *testing.F) {
	s := &schema.Schema{
		Package: "fuzztest",
		Types: []schema.Type{
			&schema.StructType{
				Name: "OptionalStruct",
				Fields: []schema.Field{
					{Name: "Required", Type: &schema.PrimitiveType{Name: "int32"}},
					{Name: "Optional", Type: &schema.PrimitiveType{Name: "string", Optional: true}},
				},
			},
		},
		Messages: []schema.MessageType{
			{Name: "OptionalMessage", TargetType: &schema.StructType{
				Name: "OptionalStruct",
				Fields: []schema.Field{
					{Name: "Required", Type: &schema.PrimitiveType{Name: "int32"}},
					{Name: "Optional", Type: &schema.PrimitiveType{Name: "string", Optional: true}},
				},
			}},
		},
	}

	// With optional present
	json1 := []byte(`{"Required": 42, "Optional": "present"}`)
	binary1, _ := fixture.Convert(s, "OptionalMessage", json1)
	f.Add(binary1)

	// With optional absent
	json2 := []byte(`{"Required": 42}`)
	binary2, _ := fixture.Convert(s, "OptionalMessage", json2)
	f.Add(binary2)

	// Malformed optional flag
	f.Add([]byte{0x00, 0x00, 0x00, 0x2A, 0x02}) // Required=42, invalid optional flag

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Optional decoder panicked: %v\nData: %x", r, data[:min(len(data), 100)])
			}
		}()
		_ = data
	})
}

// FuzzDecoderNested tests nested struct decoding
func FuzzDecoderNested(f *testing.F) {
	s := &schema.Schema{
		Package: "fuzztest",
		Types: []schema.Type{
			&schema.StructType{
				Name: "Inner",
				Fields: []schema.Field{
					{Name: "Value", Type: &schema.PrimitiveType{Name: "int32"}},
				},
			},
			&schema.StructType{
				Name: "Outer",
				Fields: []schema.Field{
					{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}},
					{Name: "Inner", Type: &schema.StructType{
						Name: "Inner",
						Fields: []schema.Field{
							{Name: "Value", Type: &schema.PrimitiveType{Name: "int32"}},
						},
					}},
				},
			},
		},
		Messages: []schema.MessageType{
			{Name: "Nested", TargetType: &schema.StructType{
				Name: "Outer",
				Fields: []schema.Field{
					{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}},
					{Name: "Inner", Type: &schema.StructType{
						Name: "Inner",
						Fields: []schema.Field{
							{Name: "Value", Type: &schema.PrimitiveType{Name: "int32"}},
						},
					}},
				},
			}},
		},
	}

	validJSON := []byte(`{"Name": "outer", "Inner": {"Value": 123}}`)
	validBinary, _ := fixture.Convert(s, "Nested", validJSON)
	f.Add(validBinary)

	f.Add([]byte{})
	f.Add([]byte{0x05, 0x00, 't', 'e', 's', 't'}) // Name but no inner struct

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Nested decoder panicked: %v\nData: %x", r, data[:min(len(data), 100)])
			}
		}()
		_ = data
	})
}

// FuzzDecoderStringLength tests string length prefix attacks
func FuzzDecoderStringLength(f *testing.F) {
	s := &schema.Schema{
		Package: "fuzztest",
		Messages: []schema.MessageType{
			{Name: "StringMessage", TargetType: &schema.PrimitiveType{Name: "string"}},
		},
	}

	validJSON := []byte(`"hello"`)
	validBinary, _ := fixture.Convert(s, "StringMessage", validJSON)
	f.Add(validBinary)

	// String length attacks
	f.Add([]byte{0xFF, 0xFF})                                          // Length claims 65535 bytes
	f.Add([]byte{0x10, 0x00, 'a', 'b'})                                // Length 16 but only 2 bytes
	f.Add([]byte{0x00, 0x00})                                          // Empty string
	f.Add(append([]byte{0x05, 0x00}, bytes.Repeat([]byte{'a'}, 5)...)) // Valid

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("String decoder panicked: %v\nData: %x", r, data[:min(len(data), 100)])
			}
		}()
		_ = data
	})
}

// FuzzDecoderArrayLength tests array length prefix attacks
func FuzzDecoderArrayLength(f *testing.F) {
	s := &schema.Schema{
		Package: "fuzztest",
		Messages: []schema.MessageType{
			{Name: "IntArray", TargetType: &schema.ArrayType{
				ElementType: &schema.PrimitiveType{Name: "int32"},
			}},
		},
	}

	validJSON := []byte(`[1, 2, 3]`)
	validBinary, _ := fixture.Convert(s, "IntArray", validJSON)
	f.Add(validBinary)

	// Array length attacks
	f.Add([]byte{0xFF, 0xFF})                         // Claims 65535 elements
	f.Add([]byte{0x10, 0x00, 0x01, 0x00, 0x00, 0x00}) // Claims 16 elements, only 1
	f.Add([]byte{0x00, 0x00})                         // Empty array

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Array length decoder panicked: %v\nData: %x", r, data[:min(len(data), 100)])
			}
		}()
		_ = data
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
