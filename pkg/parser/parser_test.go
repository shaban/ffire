package parser

import (
	"testing"

	"github.com/shaban/ffire/pkg/schema"
)

func TestParseSimpleSchema(t *testing.T) {
	src := `package test

type IntList []int32
`

	s, err := ParseBytes([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if s.Package != "test" {
		t.Errorf("Package = %q, want %q", s.Package, "test")
	}

	if len(s.Messages) != 1 {
		t.Fatalf("len(Messages) = %d, want 1", len(s.Messages))
	}

	msg := s.Messages[0]
	if msg.Name != "IntList" {
		t.Errorf("Message name = %q, want %q", msg.Name, "IntList")
	}

	arrayType, ok := msg.TargetType.(*schema.ArrayType)
	if !ok {
		t.Fatalf("Message type = %T, want *schema.ArrayType", msg.TargetType)
	}

	primType, ok := arrayType.ElementType.(*schema.PrimitiveType)
	if !ok {
		t.Fatalf("Element type = %T, want *schema.PrimitiveType", arrayType.ElementType)
	}

	if primType.Name != "int32" {
		t.Errorf("Element type name = %q, want %q", primType.Name, "int32")
	}
}

func TestParseStructSchema(t *testing.T) {
	src := `package test

type DeviceList []Device

type Device struct {
	Name     string
	Channels int32
}
`

	s, err := ParseBytes([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(s.Messages) != 1 {
		t.Fatalf("len(Messages) = %d, want 1", len(s.Messages))
	}

	// Now we have 2 types: DeviceList (ArrayType) and Device (StructType)
	if len(s.Types) != 2 {
		t.Fatalf("len(Types) = %d, want 2", len(s.Types))
	}

	// Find the Device struct type
	var structType *schema.StructType
	for _, typ := range s.Types {
		if st, ok := typ.(*schema.StructType); ok {
			structType = st
			break
		}
	}
	if structType == nil {
		t.Fatal("Device struct not found in Types")
	}

	if structType.Name != "Device" {
		t.Errorf("Struct name = %q, want %q", structType.Name, "Device")
	}

	if len(structType.Fields) != 2 {
		t.Fatalf("len(Fields) = %d, want 2", len(structType.Fields))
	}

	// Check fields
	if structType.Fields[0].Name != "Name" {
		t.Errorf("Field[0] name = %q, want %q", structType.Fields[0].Name, "Name")
	}

	if structType.Fields[1].Name != "Channels" {
		t.Errorf("Field[1] name = %q, want %q", structType.Fields[1].Name, "Channels")
	}
}

func TestParseOptionalFields(t *testing.T) {
	src := `package test

type Record struct {
	Required     string
	OptionalStr  *string
	OptionalInt  *int32
}
`

	s, err := ParseBytes([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Record is now a root type (unreferenced + exported)
	if len(s.Messages) != 1 {
		t.Fatalf("len(Messages) = %d, want 1", len(s.Messages))
	}

	structType, ok := s.Messages[0].TargetType.(*schema.StructType)
	if !ok {
		t.Fatalf("Message type = %T, want *schema.StructType", s.Messages[0].TargetType)
	}

	// Check required field
	if structType.Fields[0].Type.IsOptional() {
		t.Error("Required field should not be optional")
	}

	// Check optional fields
	if !structType.Fields[1].Type.IsOptional() {
		t.Error("OptionalStr should be optional")
	}

	if !structType.Fields[2].Type.IsOptional() {
		t.Error("OptionalInt should be optional")
	}
}

func TestParseNestedStructs(t *testing.T) {
	src := `package test

type Plugin struct {
	Name       string
	Parameters []Parameter
}

type Parameter struct {
	Label string
	Value float32
}
`

	s, err := ParseBytes([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Plugin is root (unreferenced), Parameter is helper (referenced)
	if len(s.Types) != 2 {
		t.Fatalf("len(Types) = %d, want 2", len(s.Types))
	}

	if len(s.Messages) != 1 {
		t.Fatalf("len(Messages) = %d, want 1 (Plugin should be root)", len(s.Messages))
	}

	// Find Plugin struct
	var pluginStruct *schema.StructType
	for _, typ := range s.Types {
		if typ.TypeName() == "Plugin" {
			pluginStruct = typ.(*schema.StructType)
			break
		}
	}

	if pluginStruct == nil {
		t.Fatal("Plugin struct not found")
	}

	// Check Parameters field is array of Parameter
	paramsField := pluginStruct.Fields[1]
	arrayType, ok := paramsField.Type.(*schema.ArrayType)
	if !ok {
		t.Fatalf("Parameters type = %T, want *schema.ArrayType", paramsField.Type)
	}

	paramStruct, ok := arrayType.ElementType.(*schema.StructType)
	if !ok {
		t.Fatalf("Parameters element type = %T, want *schema.StructType", arrayType.ElementType)
	}

	if paramStruct.Name != "Parameter" {
		t.Errorf("Parameters element name = %q, want %q", paramStruct.Name, "Parameter")
	}
}

func TestParseComplexSchema(t *testing.T) {
	src := `package audio

type DeviceList []Device

type PluginInfo struct {
	Name       string
	Vendor     string
	Parameters []Parameter
}

type Device struct {
	Name      string
	ID        string
	Channels  int32
	IsDefault bool
}

type Parameter struct {
	Label        string
	DefaultValue float32
	Unit         *string
}
`

	s, err := ParseBytes([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if s.Package != "audio" {
		t.Errorf("Package = %q, want %q", s.Package, "audio")
	}

	// DeviceList and PluginInfo are root types (unreferenced + exported)
	// Device and Parameter are helper types (referenced)
	if len(s.Messages) != 2 {
		t.Fatalf("len(Messages) = %d, want 2", len(s.Messages))
	}

	// We now have 4 types: DeviceList, PluginInfo, Device, Parameter
	if len(s.Types) != 4 {
		t.Fatalf("len(Types) = %d, want 4", len(s.Types))
	}
}

func TestParseFile(t *testing.T) {
	// Test parsing actual file from testdata
	s, err := Parse("../../testdata/schema/array_int.ffi")
	if err != nil {
		t.Fatalf("Parse file failed: %v", err)
	}

	if s.Package != "test" {
		t.Errorf("Package = %q, want %q", s.Package, "test")
	}

	if len(s.Messages) != 1 {
		t.Fatalf("len(Messages) = %d, want 1", len(s.Messages))
	}
}

func TestErrorUndefinedType(t *testing.T) {
	src := `package test

type Config struct {
	Value UndefinedType
}
`

	_, err := ParseBytes([]byte(src))
	if err == nil {
		t.Fatal("Expected error for undefined type, got nil")
	}
}

func TestParseStructTags(t *testing.T) {
	src := `package test

type User struct {
	ID    int64  ` + "`json:\"id\" db:\"user_id\" validate:\"required\"`" + `
	Name  string ` + "`json:\"name\" yaml:\"name\" xml:\"Name,attr\"`" + `
	Email string ` + "`json:\"email,omitempty\" validate:\"email\"`" + `
}
`

	s, err := ParseBytes([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// User is now a root type
	if len(s.Messages) != 1 {
		t.Fatalf("len(Messages) = %d, want 1", len(s.Messages))
	}

	structType, ok := s.Messages[0].TargetType.(*schema.StructType)
	if !ok {
		t.Fatalf("Message type = %T, want *schema.StructType", s.Messages[0].TargetType)
	}

	// Check first field - full tag preservation
	if structType.Fields[0].Tag != "`json:\"id\" db:\"user_id\" validate:\"required\"`" {
		t.Errorf("Field[0] Tag = %q, want full tag string", structType.Fields[0].Tag)
	}
	// Check JSON tag extraction
	if structType.Fields[0].JSONName() != "id" {
		t.Errorf("Field[0] JSONName = %q, want %q", structType.Fields[0].JSONName(), "id")
	}

	// Check second field - multiple tags
	if structType.Fields[1].Tag != "`json:\"name\" yaml:\"name\" xml:\"Name,attr\"`" {
		t.Errorf("Field[1] Tag = %q, want full tag string", structType.Fields[1].Tag)
	}
	if structType.Fields[1].JSONName() != "name" {
		t.Errorf("Field[1] JSONName = %q, want %q", structType.Fields[1].JSONName(), "name")
	}

	// Check third field - omitempty handling
	if structType.Fields[2].Tag != "`json:\"email,omitempty\" validate:\"email\"`" {
		t.Errorf("Field[2] Tag = %q, want full tag string", structType.Fields[2].Tag)
	}
	if structType.Fields[2].JSONName() != "email" {
		t.Errorf("Field[2] JSONName = %q, want %q (should strip ,omitempty)", structType.Fields[2].JSONName(), "email")
	}
}

func TestParseFieldWithoutTag(t *testing.T) {
	src := `package test

type User struct {
	Name string
}
`

	s, err := ParseBytes([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// User is now a root type
	if len(s.Messages) != 1 {
		t.Fatalf("len(Messages) = %d, want 1", len(s.Messages))
	}

	structType, ok := s.Messages[0].TargetType.(*schema.StructType)
	if !ok {
		t.Fatalf("Message type = %T, want *schema.StructType", s.Messages[0].TargetType)
	}

	// Field without tag should have empty Tag
	if structType.Fields[0].Tag != "" {
		t.Errorf("Field Tag = %q, want empty string", structType.Fields[0].Tag)
	}

	// JSONName should fall back to field name
	if structType.Fields[0].JSONName() != "Name" {
		t.Errorf("JSONName = %q, want %q", structType.Fields[0].JSONName(), "Name")
	}
}
