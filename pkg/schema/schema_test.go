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

func TestOptionalStruct(t *testing.T) {
	required := &StructType{
		Name:     "Device",
		Fields:   []Field{{Name: "Name", Type: &PrimitiveType{Name: "string"}}},
		Optional: false,
	}

	optional := &StructType{
		Name:     "Device",
		Fields:   []Field{{Name: "Name", Type: &PrimitiveType{Name: "string"}}},
		Optional: true,
	}

	if required.IsOptional() {
		t.Error("Required struct should return IsOptional() = false")
	}

	if !optional.IsOptional() {
		t.Error("Optional struct should return IsOptional() = true")
	}
}

func TestOptionalArray(t *testing.T) {
	required := &ArrayType{
		ElementType: &PrimitiveType{Name: "int32"},
		Optional:    false,
	}

	optional := &ArrayType{
		ElementType: &PrimitiveType{Name: "int32"},
		Optional:    true,
	}

	if required.IsOptional() {
		t.Error("Required array should return IsOptional() = false")
	}

	if !optional.IsOptional() {
		t.Error("Optional array should return IsOptional() = true")
	}
}

func TestArrayTypeName(t *testing.T) {
	tests := []struct {
		name     string
		elemType Type
		want     string
	}{
		{
			name:     "primitive_array",
			elemType: &PrimitiveType{Name: "int32"},
			want:     "[]int32",
		},
		{
			name:     "struct_array",
			elemType: &StructType{Name: "Device"},
			want:     "[]Device",
		},
		{
			name: "nested_array",
			elemType: &ArrayType{
				ElementType: &PrimitiveType{Name: "string"},
			},
			want: "[][]string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arr := &ArrayType{ElementType: tt.elemType}
			got := arr.TypeName()
			if got != tt.want {
				t.Errorf("TypeName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFieldJSONName(t *testing.T) {
	tests := []struct {
		name    string
		field   Field
		want    string
		setTag  string
		wantTag string
	}{
		{
			name:  "no_json_tag",
			field: Field{Name: "DeviceID"},
			want:  "DeviceID",
		},
		{
			name:    "with_json_tag",
			field:   Field{Name: "DeviceID"},
			setTag:  "device_id",
			wantTag: "device_id",
		},
		{
			name:    "cached_json_tag",
			field:   Field{Name: "UserName"},
			setTag:  "user_name",
			wantTag: "user_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := tt.field

			if tt.setTag != "" {
				field.SetJSONTag(tt.setTag)
				got := field.JSONName()
				if got != tt.wantTag {
					t.Errorf("After SetJSONTag(%q), JSONName() = %q, want %q", tt.setTag, got, tt.wantTag)
				}
			} else {
				got := field.JSONName()
				if got != tt.want {
					t.Errorf("JSONName() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestFieldWithStructTag(t *testing.T) {
	field := Field{
		Name: "DeviceID",
		Type: &PrimitiveType{Name: "int32"},
		Tag:  "`json:\"device_id\" yaml:\"device_id\" db:\"device_id\"`",
	}

	// Tag should be preserved
	if field.Tag != "`json:\"device_id\" yaml:\"device_id\" db:\"device_id\"`" {
		t.Errorf("Tag not preserved correctly")
	}

	// JSONName should use field name until SetJSONTag is called
	if field.JSONName() != "DeviceID" {
		t.Errorf("JSONName() before SetJSONTag = %q, want %q", field.JSONName(), "DeviceID")
	}

	// After setting JSON tag
	field.SetJSONTag("device_id")
	if field.JSONName() != "device_id" {
		t.Errorf("JSONName() after SetJSONTag = %q, want %q", field.JSONName(), "device_id")
	}
}

func TestSchemaValidate(t *testing.T) {
	schema := &Schema{
		Package: "test",
		Messages: []MessageType{
			{Name: "Message", TargetType: &PrimitiveType{Name: "int32"}},
		},
	}

	// Currently returns nil (not implemented)
	err := schema.Validate()
	if err != nil {
		t.Errorf("Validate() returned error: %v", err)
	}
}

func TestPrimitiveTypeName(t *testing.T) {
	tests := []struct {
		name     string
		primType *PrimitiveType
		want     string
	}{
		{"bool", &PrimitiveType{Name: "bool"}, "bool"},
		{"int8", &PrimitiveType{Name: "int8"}, "int8"},
		{"int16", &PrimitiveType{Name: "int16"}, "int16"},
		{"int32", &PrimitiveType{Name: "int32"}, "int32"},
		{"int64", &PrimitiveType{Name: "int64"}, "int64"},
		{"float32", &PrimitiveType{Name: "float32"}, "float32"},
		{"float64", &PrimitiveType{Name: "float64"}, "float64"},
		{"string", &PrimitiveType{Name: "string"}, "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.primType.TypeName()
			if got != tt.want {
				t.Errorf("TypeName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStructTypeName(t *testing.T) {
	tests := []struct {
		name       string
		structType *StructType
		want       string
	}{
		{
			name:       "simple_struct",
			structType: &StructType{Name: "Device"},
			want:       "Device",
		},
		{
			name:       "nested_struct",
			structType: &StructType{Name: "AudioPlugin"},
			want:       "AudioPlugin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.structType.TypeName()
			if got != tt.want {
				t.Errorf("TypeName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComplexNestedSchema(t *testing.T) {
	// Test complex nested structure
	paramStruct := &StructType{
		Name: "Parameter",
		Fields: []Field{
			{Name: "Name", Type: &PrimitiveType{Name: "string"}},
			{Name: "Value", Type: &PrimitiveType{Name: "float64"}},
		},
	}

	pluginStruct := &StructType{
		Name: "Plugin",
		Fields: []Field{
			{Name: "Name", Type: &PrimitiveType{Name: "string"}},
			{Name: "Parameters", Type: &ArrayType{ElementType: paramStruct}},
		},
	}

	schema := &Schema{
		Package: "audio",
		Types:   []Type{paramStruct, pluginStruct},
		Messages: []MessageType{
			{Name: "PluginList", TargetType: &ArrayType{ElementType: pluginStruct}},
		},
	}

	// Verify we can find nested types
	foundParam := schema.FindType("Parameter")
	if foundParam == nil || foundParam.TypeName() != "Parameter" {
		t.Error("Failed to find Parameter type")
	}

	foundPlugin := schema.FindType("Plugin")
	if foundPlugin == nil || foundPlugin.TypeName() != "Plugin" {
		t.Error("Failed to find Plugin type")
	}

	// Verify message structure
	if len(schema.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(schema.Messages))
	}

	msg := schema.Messages[0]
	if msg.Name != "PluginList" {
		t.Errorf("Message name = %q, want %q", msg.Name, "PluginList")
	}

	// Check message type is array of Plugin
	arrayType, ok := msg.TargetType.(*ArrayType)
	if !ok {
		t.Fatal("Message target should be ArrayType")
	}

	if arrayType.ElementType.TypeName() != "Plugin" {
		t.Errorf("Array element type = %q, want %q", arrayType.ElementType.TypeName(), "Plugin")
	}
}

func TestFindAllPrimitiveTypes(t *testing.T) {
	schema := &Schema{Package: "test"}

	primitives := []string{"bool", "int8", "int16", "int32", "int64", "float32", "float64", "string"}
	for _, primName := range primitives {
		found := schema.FindType(primName)
		if found == nil {
			t.Errorf("FindType(%q) returned nil", primName)
			continue
		}
		if found.TypeName() != primName {
			t.Errorf("FindType(%q).TypeName() = %q, want %q", primName, found.TypeName(), primName)
		}
	}
}

func TestOptionalFieldInStruct(t *testing.T) {
	s := &StructType{
		Name: "Record",
		Fields: []Field{
			{Name: "ID", Type: &PrimitiveType{Name: "int32", Optional: false}},
			{Name: "Name", Type: &PrimitiveType{Name: "string", Optional: true}},
			{Name: "Tags", Type: &ArrayType{ElementType: &PrimitiveType{Name: "string"}, Optional: true}},
		},
	}

	// Check required field
	if s.Fields[0].Type.IsOptional() {
		t.Error("ID field should not be optional")
	}

	// Check optional primitive
	if !s.Fields[1].Type.IsOptional() {
		t.Error("Name field should be optional")
	}

	// Check optional array
	if !s.Fields[2].Type.IsOptional() {
		t.Error("Tags field should be optional")
	}
}

func TestCanonicalFieldOrdering(t *testing.T) {
	// Create a struct with mixed field types in non-canonical order
	s := &StructType{
		Name: "Mixed",
		Fields: []Field{
			{Name: "OptionalName", Type: &PrimitiveType{Name: "string", Optional: true}}, // optional
			{Name: "Name", Type: &PrimitiveType{Name: "string"}},                         // variable
			{Name: "Count", Type: &PrimitiveType{Name: "int32"}},                         // fixed4
			{Name: "Timestamp", Type: &PrimitiveType{Name: "int64"}},                     // fixed8
			{Name: "OptionalAge", Type: &PrimitiveType{Name: "int32", Optional: true}},   // optional
			{Name: "Active", Type: &PrimitiveType{Name: "bool"}},                         // fixed1
			{Name: "Tags", Type: &ArrayType{ElementType: &PrimitiveType{Name: "string"}}}, // variable
			{Name: "Score", Type: &PrimitiveType{Name: "float64"}},                       // fixed8
			{Name: "Status", Type: &PrimitiveType{Name: "int16"}},                        // fixed2
		},
	}

	sorted := SortFieldsCanonical(s.Fields)

	// Expected order:
	// 1. Fixed8 (alphabetical): Score, Timestamp
	// 2. Fixed4 (alphabetical): Count
	// 3. Fixed2 (alphabetical): Status
	// 4. Fixed1 (alphabetical): Active
	// 5. Variable (alphabetical): Name, Tags
	// 6. Optional (alphabetical): OptionalAge, OptionalName
	expected := []string{
		"Score",        // fixed8
		"Timestamp",    // fixed8
		"Count",        // fixed4
		"Status",       // fixed2
		"Active",       // fixed1
		"Name",         // variable
		"Tags",         // variable
		"OptionalAge",  // optional
		"OptionalName", // optional
	}

	if len(sorted) != len(expected) {
		t.Fatalf("Expected %d fields, got %d", len(expected), len(sorted))
	}

	for i, fieldName := range expected {
		if sorted[i].Name != fieldName {
			t.Errorf("Position %d: expected %s, got %s", i, fieldName, sorted[i].Name)
		}
	}
}

func TestCanonicalizeSchema(t *testing.T) {
	// Create a schema with unordered fields
	s := &Schema{
		Package: "test",
		Types: []Type{
			&StructType{
				Name: "Address",
				Fields: []Field{
					{Name: "City", Type: &PrimitiveType{Name: "string"}},
					{Name: "ZipCode", Type: &PrimitiveType{Name: "int32"}},
					{Name: "Street", Type: &PrimitiveType{Name: "string"}},
				},
			},
		},
		Messages: []MessageType{
			{
				Name: "Person",
				TargetType: &StructType{
					Name: "Person",
					Fields: []Field{
						{Name: "Name", Type: &PrimitiveType{Name: "string"}},
						{Name: "Age", Type: &PrimitiveType{Name: "int32"}},
						{Name: "Id", Type: &PrimitiveType{Name: "int64"}},
					},
				},
			},
		},
	}

	// Canonicalize
	s.Canonicalize()

	// Check that Address fields are reordered: ZipCode (int32), City, Street (strings alphabetical)
	addrType := s.Types[0].(*StructType)
	if addrType.Fields[0].Name != "ZipCode" {
		t.Errorf("Address field 0: expected ZipCode, got %s", addrType.Fields[0].Name)
	}
	if addrType.Fields[1].Name != "City" {
		t.Errorf("Address field 1: expected City, got %s", addrType.Fields[1].Name)
	}
	if addrType.Fields[2].Name != "Street" {
		t.Errorf("Address field 2: expected Street, got %s", addrType.Fields[2].Name)
	}

	// Check that Person fields are reordered: Id (int64), Age (int32), Name (string)
	personType := s.Messages[0].TargetType.(*StructType)
	if personType.Fields[0].Name != "Id" {
		t.Errorf("Person field 0: expected Id, got %s", personType.Fields[0].Name)
	}
	if personType.Fields[1].Name != "Age" {
		t.Errorf("Person field 1: expected Age, got %s", personType.Fields[1].Name)
	}
	if personType.Fields[2].Name != "Name" {
		t.Errorf("Person field 2: expected Name, got %s", personType.Fields[2].Name)
	}
}
