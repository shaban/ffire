package validator

import (
	"fmt"
	"testing"

	"github.com/shaban/ffire/pkg/schema"
)

func TestValidateSchemaBasic(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name:       "Message",
				TargetType: &schema.ArrayType{ElementType: &schema.PrimitiveType{Name: "int32"}},
			},
		},
	}

	if err := ValidateSchema(s); err != nil {
		t.Errorf("ValidateSchema failed: %v", err)
	}
}

func TestValidateSchemaNoPackage(t *testing.T) {
	s := &schema.Schema{
		Messages: []schema.MessageType{
			{
				Name:       "Message",
				TargetType: &schema.PrimitiveType{Name: "int32"},
			},
		},
	}

	if err := ValidateSchema(s); err == nil {
		t.Error("Expected error for missing package name")
	}
}

func TestValidateSchemaNoMessages(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
	}

	if err := ValidateSchema(s); err == nil {
		t.Error("Expected error for no message types")
	}
}

func TestValidateSchemaRootTypeNotExported(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name:       "message", // lowercase - not exported
				TargetType: &schema.PrimitiveType{Name: "int32"},
			},
		},
	}

	if err := ValidateSchema(s); err == nil {
		t.Error("Expected error for non-exported root type")
	}
}

func TestValidateSchemaEmptyStruct(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name:       "Message",
				TargetType: &schema.StructType{Name: "Empty", Fields: []schema.Field{}},
			},
		},
	}

	if err := ValidateSchema(s); err == nil {
		t.Error("Expected error for empty struct")
	}
}

func TestValidateSchemaMaxNesting(t *testing.T) {
	// Create a deeply nested type
	var typ schema.Type = &schema.PrimitiveType{Name: "int32"}
	for i := 0; i < 35; i++ {
		typ = &schema.ArrayType{ElementType: typ}
	}

	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name:       "Message",
				TargetType: typ,
			},
		},
	}

	if err := ValidateSchema(s); err == nil {
		t.Error("Expected error for exceeding max nesting depth")
	}
}

func TestValidateSchemaCircularReference(t *testing.T) {
	// Create a self-referencing struct (which should be caught)
	s := &schema.Schema{
		Package: "test",
		Types: []schema.Type{
			&schema.StructType{
				Name: "Node",
				Fields: []schema.Field{
					{Name: "Value", Type: &schema.PrimitiveType{Name: "int32"}},
					{Name: "Next", Type: &schema.PrimitiveType{Name: "Node"}},
				},
			},
		},
		Messages: []schema.MessageType{
			{
				Name:       "Message",
				TargetType: &schema.PrimitiveType{Name: "Node"},
			},
		},
	}

	// Note: This should detect the circular reference
	if err := ValidateSchema(s); err == nil {
		t.Error("Expected error for circular reference")
	}
}

func TestValidateJSONPrimitiveArray(t *testing.T) {
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

	if err := ValidateJSON(s, "Message", jsonData); err != nil {
		t.Errorf("ValidateJSON failed: %v", err)
	}
}

func TestValidateJSONStruct(t *testing.T) {
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

	if err := ValidateJSON(s, "Message", jsonData); err != nil {
		t.Errorf("ValidateJSON failed: %v", err)
	}
}

func TestValidateJSONMissingRequiredField(t *testing.T) {
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

	jsonData := []byte(`{"Name": "Speaker"}`)

	if err := ValidateJSON(s, "Message", jsonData); err == nil {
		t.Error("Expected error for missing required field")
	}
}

func TestValidateJSONOptionalField(t *testing.T) {
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

	// Without optional field
	jsonData1 := []byte(`{"Required": "value"}`)
	if err := ValidateJSON(s, "Message", jsonData1); err != nil {
		t.Errorf("ValidateJSON failed for missing optional field: %v", err)
	}

	// With optional field
	jsonData2 := []byte(`{"Required": "value", "Optional": "optional"}`)
	if err := ValidateJSON(s, "Message", jsonData2); err != nil {
		t.Errorf("ValidateJSON failed with optional field: %v", err)
	}

	// With null optional field
	jsonData3 := []byte(`{"Required": "value", "Optional": null}`)
	if err := ValidateJSON(s, "Message", jsonData3); err != nil {
		t.Errorf("ValidateJSON failed with null optional field: %v", err)
	}
}

func TestValidateJSONTypeMismatch(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name:       "Message",
				TargetType: &schema.ArrayType{ElementType: &schema.PrimitiveType{Name: "int32"}},
			},
		},
	}

	// String instead of int
	jsonData := []byte(`["not", "numbers"]`)

	if err := ValidateJSON(s, "Message", jsonData); err == nil {
		t.Error("Expected error for type mismatch")
	}
}

func TestValidateJSONIntegerRanges(t *testing.T) {
	tests := []struct {
		typeName string
		value    float64
		valid    bool
	}{
		{"int8", 127, true},
		{"int8", 128, false},
		{"int8", -128, true},
		{"int8", -129, false},
		{"int16", 32767, true},
		{"int16", 32768, false},
		{"int32", 2147483647, true},
		{"int32", 2147483648, false},
	}

	for _, tt := range tests {
		s := &schema.Schema{
			Package: "test",
			Messages: []schema.MessageType{
				{
					Name:       "Message",
					TargetType: &schema.PrimitiveType{Name: tt.typeName},
				},
			},
		}

		jsonData := []byte(fmt.Sprintf(`%v`, tt.value))
		err := ValidateJSON(s, "Message", jsonData)

		if tt.valid && err != nil {
			t.Errorf("%s with value %v should be valid, got error: %v", tt.typeName, tt.value, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("%s with value %v should be invalid", tt.typeName, tt.value)
		}
	}
}

func TestValidateJSONNestedStructs(t *testing.T) {
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
			&schema.StructType{
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

	if err := ValidateJSON(s, "Message", jsonData); err != nil {
		t.Errorf("ValidateJSON failed: %v", err)
	}
}

func TestValidateJSONWithMultipleTags(t *testing.T) {
	// Create a field with multiple tags, but only JSON tag should be used for validation
	userIDField := schema.Field{
		Name: "UserID",
		Type: &schema.PrimitiveType{Name: "int64"},
		Tag:  "`json:\"user_id\" db:\"user_id\" validate:\"required\"`",
	}
	userIDField.SetJSONTag("user_id")

	nameField := schema.Field{
		Name: "Name",
		Type: &schema.PrimitiveType{Name: "string"},
		Tag:  "`json:\"name\" yaml:\"name\" xml:\"Name,attr\"`",
	}
	nameField.SetJSONTag("name")

	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name: "Message",
				TargetType: &schema.StructType{
					Name: "User",
					Fields: []schema.Field{
						userIDField,
						nameField,
					},
				},
			},
		},
	}

	// JSON should use json tag names, not struct field names
	jsonData := []byte(`{"user_id": 123, "name": "Alice"}`)

	if err := ValidateJSON(s, "Message", jsonData); err != nil {
		t.Errorf("ValidateJSON failed: %v", err)
	}

	// Using struct field names should fail
	jsonDataBad := []byte(`{"UserID": 123, "Name": "Alice"}`)
	if err := ValidateJSON(s, "Message", jsonDataBad); err == nil {
		t.Error("Expected error when using struct field names instead of JSON tag names")
	}
}

func TestValidateUint16Bounds(t *testing.T) {
	tests := []struct {
		name      string
		schema    *schema.Schema
		jsonData  string
		shouldErr bool
		errMsg    string
	}{
		{
			name: "string within bounds (65,535 bytes)",
			schema: &schema.Schema{
				Package: "test",
				Messages: []schema.MessageType{
					{
						Name:       "Message",
						TargetType: &schema.PrimitiveType{Name: "string"},
					},
				},
			},
			jsonData: func() string {
				// Create a valid JSON string with 65,535 'a' characters
				b := make([]byte, 65535)
				for i := range b {
					b[i] = 'a'
				}
				return fmt.Sprintf(`"%s"`, string(b))
			}(),
			shouldErr: false,
		},
		{
			name: "string exceeds bounds (65,536 bytes)",
			schema: &schema.Schema{
				Package: "test",
				Messages: []schema.MessageType{
					{
						Name:       "Message",
						TargetType: &schema.PrimitiveType{Name: "string"},
					},
				},
			},
			jsonData: func() string {
				// Create a valid JSON string with 65,536 'a' characters
				b := make([]byte, 65536)
				for i := range b {
					b[i] = 'a'
				}
				return fmt.Sprintf(`"%s"`, string(b))
			}(),
			shouldErr: true,
			errMsg:    "string length 65536 exceeds maximum of 65,535 bytes",
		},
		{
			name: "array within bounds (65,535 elements)",
			schema: &schema.Schema{
				Package: "test",
				Messages: []schema.MessageType{
					{
						Name:       "Message",
						TargetType: &schema.ArrayType{ElementType: &schema.PrimitiveType{Name: "int32"}},
					},
				},
			},
			jsonData: func() string {
				// Create array with 65,535 zeros
				arr := "["
				for i := 0; i < 65535; i++ {
					if i > 0 {
						arr += ","
					}
					arr += "0"
				}
				arr += "]"
				return arr
			}(),
			shouldErr: false,
		},
		{
			name: "array exceeds bounds (65,536 elements)",
			schema: &schema.Schema{
				Package: "test",
				Messages: []schema.MessageType{
					{
						Name:       "Message",
						TargetType: &schema.ArrayType{ElementType: &schema.PrimitiveType{Name: "int32"}},
					},
				},
			},
			jsonData: func() string {
				// Create array with 65,536 zeros
				arr := "["
				for i := 0; i < 65536; i++ {
					if i > 0 {
						arr += ","
					}
					arr += "0"
				}
				arr += "]"
				return arr
			}(),
			shouldErr: true,
			errMsg:    "array length 65536 exceeds maximum of 65,535 elements",
		},
		{
			name: "nested string in struct exceeds bounds",
			schema: &schema.Schema{
				Package: "test",
				Messages: []schema.MessageType{
					{
						Name: "Message",
						TargetType: &schema.StructType{
							Name: "Config",
							Fields: []schema.Field{
								{
									Name: "Name",
									Type: &schema.PrimitiveType{Name: "string"},
								},
							},
						},
					},
				},
			},
			jsonData: func() string {
				// Create a valid JSON string with 65,536 'a' characters
				b := make([]byte, 65536)
				for i := range b {
					b[i] = 'a'
				}
				return fmt.Sprintf(`{"Name":"%s"}`, string(b))
			}(),
			shouldErr: true,
			errMsg:    "string length 65536 exceeds maximum of 65,535 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJSON(tt.schema, "Message", []byte(tt.jsonData))
			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error containing '%s', got no error", tt.errMsg)
				} else if tt.errMsg != "" {
					// Check if error message contains expected text
					if !contains(err.Error(), tt.errMsg) {
						t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
