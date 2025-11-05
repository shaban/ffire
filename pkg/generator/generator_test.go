package generator

import (
	"strings"
	"testing"

	"github.com/shaban/ffire/pkg/schema"
)

func TestGenerateGoSimpleStruct(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Types: []schema.Type{
			&schema.StructType{
				Name: "User",
				Fields: []schema.Field{
					{Name: "ID", Type: &schema.PrimitiveType{Name: "int32"}},
					{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}},
				},
			},
		},
		Messages: []schema.MessageType{
			{Name: "User", TargetType: &schema.StructType{
				Name: "User",
				Fields: []schema.Field{
					{Name: "ID", Type: &schema.PrimitiveType{Name: "int32"}},
					{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}},
				},
			}},
		},
	}

	code, err := GenerateGo(s)
	if err != nil {
		t.Fatalf("GenerateGo failed: %v", err)
	}

	codeStr := string(code)

	// Check package declaration
	if !strings.Contains(codeStr, "package test") {
		t.Errorf("missing package declaration")
	}

	// Check imports
	if !strings.Contains(codeStr, "import (") {
		t.Errorf("missing imports")
	}

	// Check struct definition
	if !strings.Contains(codeStr, "type User struct") {
		t.Errorf("missing User struct")
	}

	// Check encode function
	if !strings.Contains(codeStr, "func EncodeUserMessage") {
		t.Errorf("missing EncodeUserMessage function")
	}

	// Check decode function
	if !strings.Contains(codeStr, "func DecodeUserMessage") {
		t.Errorf("missing DecodeUserMessage function")
	}

	t.Logf("Generated code:\n%s", codeStr)
}

func TestGenerateGoArray(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Types: []schema.Type{
			&schema.StructType{
				Name: "Item",
				Fields: []schema.Field{
					{Name: "Value", Type: &schema.PrimitiveType{Name: "int32"}},
				},
			},
		},
		Messages: []schema.MessageType{
			{Name: "Items", TargetType: &schema.ArrayType{
				ElementType: &schema.StructType{
					Name: "Item",
					Fields: []schema.Field{
						{Name: "Value", Type: &schema.PrimitiveType{Name: "int32"}},
					},
				},
			}},
		},
	}

	code, err := GenerateGo(s)
	if err != nil {
		t.Fatalf("GenerateGo failed: %v", err)
	}

	codeStr := string(code)

	// Check encode function for array
	if !strings.Contains(codeStr, "func EncodeItemMessage") {
		t.Errorf("missing EncodeItemMessage function")
	}

	// Check array encoding
	if !strings.Contains(codeStr, "for _, elem := range") {
		t.Errorf("missing array loop")
	}

	t.Logf("Generated code:\n%s", codeStr)
}

func TestGenerateGoOptional(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Types: []schema.Type{
			&schema.StructType{
				Name: "Record",
				Fields: []schema.Field{
					{Name: "ID", Type: &schema.PrimitiveType{Name: "int32"}},
					{Name: "Label", Type: &schema.PrimitiveType{Name: "string", Optional: true}},
				},
			},
		},
		Messages: []schema.MessageType{
			{Name: "Record", TargetType: &schema.StructType{
				Name: "Record",
				Fields: []schema.Field{
					{Name: "ID", Type: &schema.PrimitiveType{Name: "int32"}},
					{Name: "Label", Type: &schema.PrimitiveType{Name: "string", Optional: true}},
				},
			}},
		},
	}

	code, err := GenerateGo(s)
	if err != nil {
		t.Fatalf("GenerateGo failed: %v", err)
	}

	codeStr := string(code)

	// Check optional field type
	if !strings.Contains(codeStr, "Label *string") {
		t.Errorf("missing optional Label field")
	}

	// Check nil check in encoding
	if !strings.Contains(codeStr, "if") && !strings.Contains(codeStr, "== nil") {
		t.Errorf("missing nil check for optional field")
	}

	t.Logf("Generated code:\n%s", codeStr)
}

func TestGenerateGoPrimitiveMessage(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{Name: "Count", TargetType: &schema.PrimitiveType{Name: "int32"}},
		},
	}

	code, err := GenerateGo(s)
	if err != nil {
		t.Fatalf("GenerateGo failed: %v", err)
	}

	codeStr := string(code)

	// Check function name uses capitalized primitive
	if !strings.Contains(codeStr, "func EncodeInt32Message") {
		t.Errorf("missing EncodeInt32Message function")
	}

	if !strings.Contains(codeStr, "func DecodeInt32Message") {
		t.Errorf("missing DecodeInt32Message function")
	}

	t.Logf("Generated code:\n%s", codeStr)
}

func TestGenerateGoStructWithTags(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Types: []schema.Type{
			&schema.StructType{
				Name: "Person",
				Fields: []schema.Field{
					{Name: "ID", Type: &schema.PrimitiveType{Name: "int32"}, Tag: "`json:\"id\"`"},
					{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}, Tag: "`json:\"name\" db:\"full_name\"`"},
				},
			},
		},
		Messages: []schema.MessageType{
			{Name: "Person", TargetType: &schema.StructType{
				Name: "Person",
				Fields: []schema.Field{
					{Name: "ID", Type: &schema.PrimitiveType{Name: "int32"}, Tag: "`json:\"id\"`"},
					{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}, Tag: "`json:\"name\" db:\"full_name\"`"},
				},
			}},
		},
	}

	code, err := GenerateGo(s)
	if err != nil {
		t.Fatalf("GenerateGo failed: %v", err)
	}

	codeStr := string(code)

	// Check tags are preserved
	if !strings.Contains(codeStr, "`json:\"id\"`") {
		t.Errorf("missing json tag on ID field")
	}

	if !strings.Contains(codeStr, "`json:\"name\" db:\"full_name\"`") {
		t.Errorf("missing tags on Name field")
	}

	t.Logf("Generated code:\n%s", codeStr)
}
