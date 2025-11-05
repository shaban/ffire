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
