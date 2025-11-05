package fixture

import (
	"bytes"
	"testing"

	"github.com/shaban/ffire/internal/wire"
	"github.com/shaban/ffire/pkg/schema"
)

func TestConvertPrimitiveArray(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name:       "Message",
				TargetType: &schema.ArrayType{ElementType: &schema.PrimitiveType{Name: "int32"}},
			},
		},
	}

	jsonData := []byte(`[1, 2, 3, 4, 5]`)

	binary, err := Convert(s, "Message", jsonData)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Decode and verify
	r := bytes.NewReader(binary)

	// Read array length
	length, err := wire.DecodeArrayHeader(r)
	if err != nil {
		t.Fatalf("DecodeArrayHeader failed: %v", err)
	}

	if length != 5 {
		t.Errorf("Array length = %d, want 5", length)
	}

	// Read elements
	for i := 0; i < 5; i++ {
		val, err := wire.DecodeInt32(r)
		if err != nil {
			t.Fatalf("DecodeInt32 failed at index %d: %v", i, err)
		}
		if val != int32(i+1) {
			t.Errorf("Element %d = %d, want %d", i, val, i+1)
		}
	}
}

func TestConvertStruct(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name: "Message",
				TargetType: &schema.StructType{
					Name: "Device",
					Fields: []schema.Field{
						{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}},
						{Name: "Channels", Type: &schema.PrimitiveType{Name: "int32"}},
					},
				},
			},
		},
	}

	jsonData := []byte(`{"Name": "Speaker", "Channels": 2}`)

	binary, err := Convert(s, "Message", jsonData)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Decode and verify
	r := bytes.NewReader(binary)

	// Read Name field
	name, err := wire.DecodeString(r)
	if err != nil {
		t.Fatalf("DecodeString failed: %v", err)
	}
	if name != "Speaker" {
		t.Errorf("Name = %q, want %q", name, "Speaker")
	}

	// Read Channels field
	channels, err := wire.DecodeInt32(r)
	if err != nil {
		t.Fatalf("DecodeInt32 failed: %v", err)
	}
	if channels != 2 {
		t.Errorf("Channels = %d, want 2", channels)
	}
}

func TestConvertStructArray(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name: "Message",
				TargetType: &schema.ArrayType{
					ElementType: &schema.StructType{
						Name: "Device",
						Fields: []schema.Field{
							{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}},
							{Name: "Channels", Type: &schema.PrimitiveType{Name: "int32"}},
						},
					},
				},
			},
		},
	}

	jsonData := []byte(`[
		{"Name": "Speaker", "Channels": 2},
		{"Name": "Microphone", "Channels": 1}
	]`)

	binary, err := Convert(s, "Message", jsonData)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Decode and verify
	r := bytes.NewReader(binary)

	// Read array length
	length, err := wire.DecodeArrayHeader(r)
	if err != nil {
		t.Fatalf("DecodeArrayHeader failed: %v", err)
	}
	if length != 2 {
		t.Errorf("Array length = %d, want 2", length)
	}

	// First device
	name1, _ := wire.DecodeString(r)
	channels1, _ := wire.DecodeInt32(r)
	if name1 != "Speaker" || channels1 != 2 {
		t.Errorf("Device 0: got %q, %d; want %q, %d", name1, channels1, "Speaker", 2)
	}

	// Second device
	name2, _ := wire.DecodeString(r)
	channels2, _ := wire.DecodeInt32(r)
	if name2 != "Microphone" || channels2 != 1 {
		t.Errorf("Device 1: got %q, %d; want %q, %d", name2, channels2, "Microphone", 1)
	}
}

func TestConvertOptionalFields(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name: "Message",
				TargetType: &schema.StructType{
					Name: "Record",
					Fields: []schema.Field{
						{Name: "Required", Type: &schema.PrimitiveType{Name: "string"}},
						{Name: "Optional", Type: &schema.PrimitiveType{Name: "string", Optional: true}},
					},
				},
			},
		},
	}

	// Test with optional field present
	jsonData1 := []byte(`{"Required": "value", "Optional": "present"}`)
	binary1, err := Convert(s, "Message", jsonData1)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	r1 := bytes.NewReader(binary1)
	required1, _ := wire.DecodeString(r1)
	present1, _ := wire.DecodeBool(r1)
	optional1, _ := wire.DecodeString(r1)

	if required1 != "value" {
		t.Errorf("Required = %q, want %q", required1, "value")
	}
	if !present1 {
		t.Error("Optional field should be present")
	}
	if optional1 != "present" {
		t.Errorf("Optional = %q, want %q", optional1, "present")
	}

	// Test with optional field absent
	jsonData2 := []byte(`{"Required": "value"}`)
	binary2, err := Convert(s, "Message", jsonData2)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	r2 := bytes.NewReader(binary2)
	required2, _ := wire.DecodeString(r2)
	present2, _ := wire.DecodeBool(r2)

	if required2 != "value" {
		t.Errorf("Required = %q, want %q", required2, "value")
	}
	if present2 {
		t.Error("Optional field should not be present")
	}
}

func TestConvertNestedStructs(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Types: []schema.Type{
			&schema.StructType{
				Name: "Parameter",
				Fields: []schema.Field{
					{Name: "Label", Type: &schema.PrimitiveType{Name: "string"}},
					{Name: "Value", Type: &schema.PrimitiveType{Name: "float32"}},
				},
			},
		},
		Messages: []schema.MessageType{
			{
				Name: "Message",
				TargetType: &schema.StructType{
					Name: "Plugin",
					Fields: []schema.Field{
						{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}},
						{Name: "Parameters", Type: &schema.ArrayType{
							ElementType: &schema.StructType{
								Name: "Parameter",
								Fields: []schema.Field{
									{Name: "Label", Type: &schema.PrimitiveType{Name: "string"}},
									{Name: "Value", Type: &schema.PrimitiveType{Name: "float32"}},
								},
							},
						}},
					},
				},
			},
		},
	}

	jsonData := []byte(`{
		"Name": "Reverb",
		"Parameters": [
			{"Label": "Size", "Value": 0.5},
			{"Label": "Damping", "Value": 0.3}
		]
	}`)

	binary, err := Convert(s, "Message", jsonData)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Decode and verify
	r := bytes.NewReader(binary)

	// Read plugin name
	pluginName, _ := wire.DecodeString(r)
	if pluginName != "Reverb" {
		t.Errorf("Plugin name = %q, want %q", pluginName, "Reverb")
	}

	// Read parameters array length
	paramLength, _ := wire.DecodeArrayHeader(r)
	if paramLength != 2 {
		t.Errorf("Parameters length = %d, want 2", paramLength)
	}

	// First parameter
	label1, _ := wire.DecodeString(r)
	value1, _ := wire.DecodeFloat32(r)
	if label1 != "Size" || value1 != 0.5 {
		t.Errorf("Parameter 0: got %q, %f; want %q, %f", label1, value1, "Size", 0.5)
	}

	// Second parameter
	label2, _ := wire.DecodeString(r)
	value2, _ := wire.DecodeFloat32(r)
	if label2 != "Damping" || value2 != 0.3 {
		t.Errorf("Parameter 1: got %q, %f; want %q, %f", label2, value2, "Damping", 0.3)
	}
}

func TestConvertErrorInvalidJSON(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name:       "Message",
				TargetType: &schema.PrimitiveType{Name: "int32"},
			},
		},
	}

	jsonData := []byte(`{invalid json}`)

	_, err := Convert(s, "Message", jsonData)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestConvertErrorMissingMessageType(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name:       "Message",
				TargetType: &schema.PrimitiveType{Name: "int32"},
			},
		},
	}

	jsonData := []byte(`42`)

	_, err := Convert(s, "NonExistent", jsonData)
	if err == nil {
		t.Error("Expected error for missing message type")
	}
}

// Test all primitive types
func TestConvertAllPrimitiveTypes(t *testing.T) {
	tests := []struct {
		name      string
		typeName  string
		jsonValue string
		wantBytes int
	}{
		{"bool_true", "bool", "true", 1},
		{"bool_false", "bool", "false", 1},
		{"int8", "int8", "42", 1},
		{"int8_negative", "int8", "-42", 1},
		{"int16", "int16", "1234", 2},
		{"int16_negative", "int16", "-1234", 2},
		{"int32", "int32", "123456", 4},
		{"int32_negative", "int32", "-123456", 4},
		{"int64", "int64", "9876543210", 8},
		{"int64_negative", "int64", "-9876543210", 8},
		{"float32", "float32", "3.14", 4},
		{"float32_negative", "float32", "-3.14", 4},
		{"float64", "float64", "3.141592653589793", 8},
		{"float64_negative", "float64", "-3.141592653589793", 8},
		{"string", "string", `"hello"`, 2 + 5}, // uint16 length + content
		{"string_empty", "string", `""`, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &schema.Schema{
				Package: "test",
				Messages: []schema.MessageType{
					{
						Name:       "Message",
						TargetType: &schema.PrimitiveType{Name: tt.typeName},
					},
				},
			}

			binary, err := Convert(s, "Message", []byte(tt.jsonValue))
			if err != nil {
				t.Fatalf("Convert failed: %v", err)
			}

			if len(binary) != tt.wantBytes {
				t.Errorf("Binary length = %d, want %d", len(binary), tt.wantBytes)
			}
		})
	}
}

func TestConvertPrimitiveArrayAllTypes(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		json     string
		count    int
	}{
		{"bool_array", "bool", `[true, false, true]`, 3},
		{"int8_array", "int8", `[1, 2, 3, 4, 5]`, 5},
		{"int16_array", "int16", `[100, 200, 300]`, 3},
		{"int32_array", "int32", `[1000, 2000, 3000]`, 3},
		{"int64_array", "int64", `[10000, 20000, 30000]`, 3},
		{"float32_array", "float32", `[1.1, 2.2, 3.3]`, 3},
		{"float64_array", "float64", `[1.1, 2.2, 3.3]`, 3},
		{"string_array", "string", `["a", "b", "c"]`, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &schema.Schema{
				Package: "test",
				Messages: []schema.MessageType{
					{
						Name: "Message",
						TargetType: &schema.ArrayType{
							ElementType: &schema.PrimitiveType{Name: tt.typeName},
						},
					},
				},
			}

			binary, err := Convert(s, "Message", []byte(tt.json))
			if err != nil {
				t.Fatalf("Convert failed: %v", err)
			}

			// Verify array header
			r := bytes.NewReader(binary)
			length, err := wire.DecodeArrayHeader(r)
			if err != nil {
				t.Fatalf("DecodeArrayHeader failed: %v", err)
			}

			if int(length) != tt.count {
				t.Errorf("Array length = %d, want %d", length, tt.count)
			}
		})
	}
}

func TestConvertErrorTypeCheck(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		json     string
		wantErr  string
	}{
		{"bool_wrong_type", "bool", `"not a bool"`, "expected bool"},
		{"int8_wrong_type", "int8", `"not a number"`, "expected number"},
		{"int16_wrong_type", "int16", `"not a number"`, "expected number"},
		{"int32_wrong_type", "int32", `"not a number"`, "expected number"},
		{"int64_wrong_type", "int64", `"not a number"`, "expected number"},
		{"float32_wrong_type", "float32", `"not a number"`, "expected number"},
		{"float64_wrong_type", "float64", `"not a number"`, "expected number"},
		{"string_wrong_type", "string", `123`, "expected string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &schema.Schema{
				Package: "test",
				Messages: []schema.MessageType{
					{
						Name:       "Message",
						TargetType: &schema.PrimitiveType{Name: tt.typeName},
					},
				},
			}

			_, err := Convert(s, "Message", []byte(tt.json))
			if err == nil {
				t.Fatalf("Expected error containing %q", tt.wantErr)
			}
		})
	}
}

func TestConvertStructWithMissingRequiredField(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name: "Message",
				TargetType: &schema.StructType{
					Name: "Device",
					Fields: []schema.Field{
						{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}},
						{Name: "Required", Type: &schema.PrimitiveType{Name: "int32"}},
					},
				},
			},
		},
	}

	// Missing required field "Required"
	jsonData := []byte(`{"Name": "Test"}`)

	_, err := Convert(s, "Message", jsonData)
	if err == nil {
		t.Error("Expected error for missing required field")
	}
}

func TestConvertStructWrongType(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name: "Message",
				TargetType: &schema.StructType{
					Name: "Device",
					Fields: []schema.Field{
						{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}},
					},
				},
			},
		},
	}

	// Not a struct
	jsonData := []byte(`"not an object"`)

	_, err := Convert(s, "Message", jsonData)
	if err == nil {
		t.Error("Expected error for wrong type")
	}
}

func TestConvertArrayWrongType(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name: "Message",
				TargetType: &schema.ArrayType{
					ElementType: &schema.PrimitiveType{Name: "int32"},
				},
			},
		},
	}

	// Not an array
	jsonData := []byte(`"not an array"`)

	_, err := Convert(s, "Message", jsonData)
	if err == nil {
		t.Error("Expected error for wrong type")
	}
}

func TestConvertOptionalArrayPresent(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name: "Message",
				TargetType: &schema.ArrayType{
					ElementType: &schema.PrimitiveType{Name: "int32"},
					Optional:    true,
				},
			},
		},
	}

	jsonData := []byte(`[1, 2, 3]`)

	binary, err := Convert(s, "Message", jsonData)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	r := bytes.NewReader(binary)
	
	// Read present flag
	present, _ := wire.DecodeBool(r)
	if !present {
		t.Error("Optional array should be present")
	}

	// Read array length
	length, _ := wire.DecodeArrayHeader(r)
	if length != 3 {
		t.Errorf("Array length = %d, want 3", length)
	}
}

func TestConvertOptionalStructPresent(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name: "Message",
				TargetType: &schema.StructType{
					Name: "Device",
					Fields: []schema.Field{
						{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}},
					},
					Optional: true,
				},
			},
		},
	}

	jsonData := []byte(`{"Name": "Test"}`)

	binary, err := Convert(s, "Message", jsonData)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	r := bytes.NewReader(binary)
	
	// Read present flag
	present, _ := wire.DecodeBool(r)
	if !present {
		t.Error("Optional struct should be present")
	}

	// Read struct field
	name, _ := wire.DecodeString(r)
	if name != "Test" {
		t.Errorf("Name = %q, want %q", name, "Test")
	}
}

func TestConvertOptionalPrimitivePresent(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name:       "Message",
				TargetType: &schema.PrimitiveType{Name: "int32", Optional: true},
			},
		},
	}

	jsonData := []byte(`42`)

	binary, err := Convert(s, "Message", jsonData)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	r := bytes.NewReader(binary)
	
	// Read present flag
	present, _ := wire.DecodeBool(r)
	if !present {
		t.Error("Optional primitive should be present")
	}

	// Read value
	value, _ := wire.DecodeInt32(r)
	if value != 42 {
		t.Errorf("Value = %d, want 42", value)
	}
}

func TestConvertEmptyArray(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name: "Message",
				TargetType: &schema.ArrayType{
					ElementType: &schema.PrimitiveType{Name: "int32"},
				},
			},
		},
	}

	jsonData := []byte(`[]`)

	binary, err := Convert(s, "Message", jsonData)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	r := bytes.NewReader(binary)
	length, _ := wire.DecodeArrayHeader(r)
	if length != 0 {
		t.Errorf("Empty array length = %d, want 0", length)
	}
}

func TestConvertEmptyString(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name:       "Message",
				TargetType: &schema.PrimitiveType{Name: "string"},
			},
		},
	}

	jsonData := []byte(`""`)

	binary, err := Convert(s, "Message", jsonData)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	r := bytes.NewReader(binary)
	str, _ := wire.DecodeString(r)
	if str != "" {
		t.Errorf("String = %q, want empty string", str)
	}
}

func TestConvertArrayElementError(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name: "Message",
				TargetType: &schema.ArrayType{
					ElementType: &schema.PrimitiveType{Name: "int32"},
				},
			},
		},
	}

	// Array with wrong type element
	jsonData := []byte(`[1, 2, "not a number", 4]`)

	_, err := Convert(s, "Message", jsonData)
	if err == nil {
		t.Error("Expected error for wrong element type in array")
	}
}
