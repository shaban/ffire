package schema

import "testing"

func TestPrimitiveTypes(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"bool", true},
		{"int8", true},
		{"int16", true},
		{"int32", true},
		{"int64", true},
		{"float32", true},
		{"float64", true},
		{"string", true},
		{"Device", false},
		{"[]int32", false},
	}

	for _, tt := range tests {
		got := IsPrimitive(tt.name)
		if got != tt.want {
			t.Errorf("IsPrimitive(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestSchemaConstruction(t *testing.T) {
	// Create a simple schema: type Message = []Device
	deviceStruct := &StructType{
		Name: "Device",
		Fields: []Field{
			{Name: "Name", Type: &PrimitiveType{Name: "string"}},
			{Name: "Channels", Type: &PrimitiveType{Name: "int32"}},
		},
	}

	arrayType := &ArrayType{
		ElementType: deviceStruct,
	}

	schema := &Schema{
		Package: "test",
		Messages: []MessageType{
			{Name: "Message", TargetType: arrayType},
		},
		Types: []Type{deviceStruct},
	}

	// Validate structure
	if schema.Package != "test" {
		t.Errorf("Package = %q, want %q", schema.Package, "test")
	}

	if len(schema.Messages) != 1 {
		t.Fatalf("len(Messages) = %d, want 1", len(schema.Messages))
	}

	msg := schema.Messages[0]
	if msg.Name != "Message" {
		t.Errorf("Message name = %q, want %q", msg.Name, "Message")
	}

	if msg.TargetType.TypeName() != "[]Device" {
		t.Errorf("Message type = %q, want %q", msg.TargetType.TypeName(), "[]Device")
	}
}

func TestOptionalTypes(t *testing.T) {
	optional := &PrimitiveType{Name: "string", Optional: true}
	required := &PrimitiveType{Name: "string", Optional: false}

	if !optional.IsOptional() {
		t.Error("Optional type should return IsOptional() = true")
	}

	if required.IsOptional() {
		t.Error("Required type should return IsOptional() = false")
	}
}

func TestFindType(t *testing.T) {
	deviceStruct := &StructType{
		Name: "Device",
		Fields: []Field{
			{Name: "Name", Type: &PrimitiveType{Name: "string"}},
		},
	}

	schema := &Schema{
		Package: "test",
		Types:   []Type{deviceStruct},
	}

	// Find existing type
	found := schema.FindType("Device")
	if found == nil {
		t.Fatal("FindType(Device) returned nil")
	}
	if found.TypeName() != "Device" {
		t.Errorf("Found type name = %q, want %q", found.TypeName(), "Device")
	}

	// Find primitive
	foundPrim := schema.FindType("int32")
	if foundPrim == nil {
		t.Fatal("FindType(int32) returned nil")
	}
	if foundPrim.TypeName() != "int32" {
		t.Errorf("Found primitive name = %q, want %q", foundPrim.TypeName(), "int32")
	}

	// Find non-existent type
	notFound := schema.FindType("NonExistent")
	if notFound != nil {
		t.Errorf("FindType(NonExistent) should return nil, got %v", notFound)
	}
}
