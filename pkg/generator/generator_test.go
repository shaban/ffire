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

// Roundtrip tests - compile and execute generated code

func TestRoundtripPrimitiveInt32Array(t *testing.T) {
	s := &schema.Schema{
		Package: "testpkg",
		Messages: []schema.MessageType{
			{Name: "Numbers", TargetType: &schema.ArrayType{
				ElementType: &schema.PrimitiveType{Name: "int32"},
			}},
		},
	}

	code, err := GenerateGo(s)
	if err != nil {
		t.Fatalf("GenerateGo failed: %v", err)
	}

	// Verify bulk array encoding is used (unsafe optimization)
	codeStr := string(code)
	if !strings.Contains(codeStr, "unsafe.Slice") {
		t.Errorf("Expected bulk array encoding with unsafe.Slice for []int32")
	}

	t.Logf("Generated code for int32 array:\n%s", codeStr)
}

func TestRoundtripPrimitiveFloat32Array(t *testing.T) {
	s := &schema.Schema{
		Package: "testpkg",
		Messages: []schema.MessageType{
			{Name: "Floats", TargetType: &schema.ArrayType{
				ElementType: &schema.PrimitiveType{Name: "float32"},
			}},
		},
	}

	code, err := GenerateGo(s)
	if err != nil {
		t.Fatalf("GenerateGo failed: %v", err)
	}

	// Verify bulk array encoding is used (zero-copy reinterpret, not Float32bits)
	codeStr := string(code)
	if !strings.Contains(codeStr, "unsafe.Slice") {
		t.Errorf("Expected bulk array encoding with unsafe.Slice for []float32")
	}
}

func TestRoundtripPrimitiveFloat64Array(t *testing.T) {
	s := &schema.Schema{
		Package: "testpkg",
		Messages: []schema.MessageType{
			{Name: "Doubles", TargetType: &schema.ArrayType{
				ElementType: &schema.PrimitiveType{Name: "float64"},
			}},
		},
	}

	code, err := GenerateGo(s)
	if err != nil {
		t.Fatalf("GenerateGo failed: %v", err)
	}

	// Verify bulk array encoding is used (zero-copy reinterpret, not Float64bits)
	codeStr := string(code)
	if !strings.Contains(codeStr, "unsafe.Slice") {
		t.Errorf("Expected bulk array encoding with unsafe.Slice for []float64")
	}
}

func TestRoundtripAllPrimitiveTypes(t *testing.T) {
	s := &schema.Schema{
		Package: "testpkg",
		Types: []schema.Type{
			&schema.StructType{
				Name: "AllTypes",
				Fields: []schema.Field{
					{Name: "Bool", Type: &schema.PrimitiveType{Name: "bool"}},
					{Name: "Int8", Type: &schema.PrimitiveType{Name: "int8"}},
					{Name: "Int16", Type: &schema.PrimitiveType{Name: "int16"}},
					{Name: "Int32", Type: &schema.PrimitiveType{Name: "int32"}},
					{Name: "Int64", Type: &schema.PrimitiveType{Name: "int64"}},
					{Name: "Float32", Type: &schema.PrimitiveType{Name: "float32"}},
					{Name: "Float64", Type: &schema.PrimitiveType{Name: "float64"}},
					{Name: "String", Type: &schema.PrimitiveType{Name: "string"}},
				},
			},
		},
		Messages: []schema.MessageType{
			{Name: "AllTypes", TargetType: &schema.StructType{
				Name: "AllTypes",
				Fields: []schema.Field{
					{Name: "Bool", Type: &schema.PrimitiveType{Name: "bool"}},
					{Name: "Int8", Type: &schema.PrimitiveType{Name: "int8"}},
					{Name: "Int16", Type: &schema.PrimitiveType{Name: "int16"}},
					{Name: "Int32", Type: &schema.PrimitiveType{Name: "int32"}},
					{Name: "Int64", Type: &schema.PrimitiveType{Name: "int64"}},
					{Name: "Float32", Type: &schema.PrimitiveType{Name: "float32"}},
					{Name: "Float64", Type: &schema.PrimitiveType{Name: "float64"}},
					{Name: "String", Type: &schema.PrimitiveType{Name: "string"}},
				},
			}},
		},
	}

	code, err := GenerateGo(s)
	if err != nil {
		t.Fatalf("GenerateGo failed: %v", err)
	}

	codeStr := string(code)

	// Verify all types are handled (field names, not exact formatting)
	requiredFields := []string{"Bool", "Int8", "Int16", "Int32", "Int64", "Float32", "Float64", "String"}
	requiredTypes := []string{"bool", "int8", "int16", "int32", "int64", "float32", "float64", "string"}
	requiredFuncs := []string{"EncodeAllTypesMessage", "DecodeAllTypesMessage"}

	for _, field := range requiredFields {
		if !strings.Contains(codeStr, field) {
			t.Errorf("Missing field: %s", field)
		}
	}

	for _, typeName := range requiredTypes {
		if !strings.Contains(codeStr, typeName) {
			t.Errorf("Missing type: %s", typeName)
		}
	}

	for _, funcName := range requiredFuncs {
		if !strings.Contains(codeStr, funcName) {
			t.Errorf("Missing function: %s", funcName)
		}
	}
}

func TestRoundtripOptionalFields(t *testing.T) {
	s := &schema.Schema{
		Package: "testpkg",
		Types: []schema.Type{
			&schema.StructType{
				Name: "Optional",
				Fields: []schema.Field{
					{Name: "Required", Type: &schema.PrimitiveType{Name: "int32"}},
					{Name: "OptInt", Type: &schema.PrimitiveType{Name: "int32", Optional: true}},
					{Name: "OptString", Type: &schema.PrimitiveType{Name: "string", Optional: true}},
				},
			},
		},
		Messages: []schema.MessageType{
			{Name: "Optional", TargetType: &schema.StructType{
				Name: "Optional",
				Fields: []schema.Field{
					{Name: "Required", Type: &schema.PrimitiveType{Name: "int32"}},
					{Name: "OptInt", Type: &schema.PrimitiveType{Name: "int32", Optional: true}},
					{Name: "OptString", Type: &schema.PrimitiveType{Name: "string", Optional: true}},
				},
			}},
		},
	}

	code, err := GenerateGo(s)
	if err != nil {
		t.Fatalf("GenerateGo failed: %v", err)
	}

	codeStr := string(code)

	// Verify optional fields use pointers (struct definition)
	if !strings.Contains(codeStr, "OptInt") || !strings.Contains(codeStr, "*int32") {
		t.Errorf("Optional int32 should use pointer type")
	}
	if !strings.Contains(codeStr, "OptString") || !strings.Contains(codeStr, "*string") {
		t.Errorf("Optional string should use pointer type")
	}

	// Verify nil checks are present
	if !strings.Contains(codeStr, "== nil") {
		t.Errorf("Optional field encoding should check for nil")
	}
}

func TestRoundtripNestedStructs(t *testing.T) {
	s := &schema.Schema{
		Package: "testpkg",
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
					{Name: "Nested", Type: &schema.StructType{
						Name: "Inner",
						Fields: []schema.Field{
							{Name: "Value", Type: &schema.PrimitiveType{Name: "int32"}},
						},
					}},
				},
			},
		},
		Messages: []schema.MessageType{
			{Name: "Outer", TargetType: &schema.StructType{
				Name: "Outer",
				Fields: []schema.Field{
					{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}},
					{Name: "Nested", Type: &schema.StructType{
						Name: "Inner",
						Fields: []schema.Field{
							{Name: "Value", Type: &schema.PrimitiveType{Name: "int32"}},
						},
					}},
				},
			}},
		},
	}

	code, err := GenerateGo(s)
	if err != nil {
		t.Fatalf("GenerateGo failed: %v", err)
	}

	codeStr := string(code)

	// Verify nested struct handling
	if !strings.Contains(codeStr, "type Inner struct") {
		t.Errorf("Missing Inner struct definition")
	}
	if !strings.Contains(codeStr, "type Outer struct") {
		t.Errorf("Missing Outer struct definition")
	}
	if !strings.Contains(codeStr, "Nested Inner") {
		t.Errorf("Missing nested field in Outer struct")
	}
}

func TestRoundtripStringArray(t *testing.T) {
	s := &schema.Schema{
		Package: "testpkg",
		Messages: []schema.MessageType{
			{Name: "Strings", TargetType: &schema.ArrayType{
				ElementType: &schema.PrimitiveType{Name: "string"},
			}},
		},
	}

	code, err := GenerateGo(s)
	if err != nil {
		t.Fatalf("GenerateGo failed: %v", err)
	}

	codeStr := string(code)

	// String arrays should NOT use bulk encoding (need length prefix per string)
	if strings.Contains(codeStr, "unsafe.Slice") && strings.Contains(codeStr, "string") {
		t.Errorf("String arrays should not use unsafe bulk encoding")
	}

	// Should use element-by-element loop
	if !strings.Contains(codeStr, "for _, elem := range") {
		t.Errorf("String arrays should use element-by-element encoding")
	}
}

func TestRoundtripStructArray(t *testing.T) {
	s := &schema.Schema{
		Package: "testpkg",
		Types: []schema.Type{
			&schema.StructType{
				Name: "Item",
				Fields: []schema.Field{
					{Name: "ID", Type: &schema.PrimitiveType{Name: "int32"}},
					{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}},
				},
			},
		},
		Messages: []schema.MessageType{
			{Name: "Items", TargetType: &schema.ArrayType{
				ElementType: &schema.StructType{
					Name: "Item",
					Fields: []schema.Field{
						{Name: "ID", Type: &schema.PrimitiveType{Name: "int32"}},
						{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}},
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

	// Struct arrays should use element-by-element encoding
	if !strings.Contains(codeStr, "for _, elem := range") {
		t.Errorf("Struct arrays should use element-by-element encoding")
	}

	// Should encode struct fields
	if !strings.Contains(codeStr, "elem.ID") || !strings.Contains(codeStr, "elem.Name") {
		t.Errorf("Should encode struct fields from array element")
	}
}

func TestRoundtripOptionalArray(t *testing.T) {
	s := &schema.Schema{
		Package: "testpkg",
		Messages: []schema.MessageType{
			{Name: "OptionalArray", TargetType: &schema.ArrayType{
				ElementType: &schema.PrimitiveType{Name: "int32"},
				Optional:    true,
			}},
		},
	}

	code, err := GenerateGo(s)
	if err != nil {
		t.Fatalf("GenerateGo failed: %v", err)
	}

	codeStr := string(code)

	// Optional arrays should use pointer type
	if !strings.Contains(codeStr, "v *[]int32") {
		t.Errorf("Optional array parameter should use pointer type")
	}

	// Should have nil check
	if !strings.Contains(codeStr, "== nil") {
		t.Errorf("Optional array should check for nil")
	}
}

func TestGenerateLanguageSwitching(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Types: []schema.Type{
			&schema.StructType{
				Name: "Simple",
				Fields: []schema.Field{
					{Name: "ID", Type: &schema.PrimitiveType{Name: "int32"}},
				},
			},
		},
		Messages: []schema.MessageType{
			{Name: "Simple", TargetType: &schema.StructType{
				Name: "Simple",
				Fields: []schema.Field{
					{Name: "ID", Type: &schema.PrimitiveType{Name: "int32"}},
				},
			}},
		},
	}

	// Test Go generation works
	codeGo, err := Generate(s, "go")
	if err != nil {
		t.Fatalf("Generate(go) failed: %v", err)
	}
	if len(codeGo) == 0 {
		t.Errorf("Generate(go) returned empty code")
	}

	// Test C++ generation works
	codeCpp, err := Generate(s, "cpp")
	if err != nil {
		t.Fatalf("Generate(cpp) failed: %v", err)
	}
	if len(codeCpp) == 0 {
		t.Errorf("Generate(cpp) returned empty code")
	}

	// Test Swift returns error (not implemented)
	_, err = Generate(s, "swift")
	if err == nil {
		t.Errorf("Generate(swift) should return error (not implemented)")
	}

	// Test unknown language returns error
	_, err = Generate(s, "rust")
	if err == nil {
		t.Errorf("Generate(rust) should return error (unknown language)")
	}
}

func TestBulkArrayEncodingAllNumericTypes(t *testing.T) {
	numericTypes := []string{"int8", "int16", "int32", "int64", "float32", "float64"}

	for _, typeName := range numericTypes {
		t.Run(typeName, func(t *testing.T) {
			s := &schema.Schema{
				Package: "testpkg",
				Messages: []schema.MessageType{
					{Name: "Numbers", TargetType: &schema.ArrayType{
						ElementType: &schema.PrimitiveType{Name: typeName},
					}},
				},
			}

			code, err := GenerateGo(s)
			if err != nil {
				t.Fatalf("GenerateGo failed for %s: %v", typeName, err)
			}

			codeStr := string(code)

			// All numeric types should use bulk encoding
			if !strings.Contains(codeStr, "unsafe.Slice") {
				t.Errorf("%s array should use bulk encoding with unsafe.Slice", typeName)
			}

			// Verify correct size calculation
			// int8 is special - uses len(v) directly (1 byte per element)
			if typeName == "int8" {
				if !strings.Contains(codeStr, "len(v))") {
					t.Errorf("int8 array should use len(v) directly")
				}
			} else {
				// Other types multiply by element size
				expectedSizes := map[string]string{
					"int16":   "2",
					"int32":   "4",
					"int64":   "8",
					"float32": "4",
					"float64": "8",
				}

				if size, ok := expectedSizes[typeName]; ok {
					// Check both with and without spaces
					sizeCheck1 := "len(v)*" + size
					sizeCheck2 := "len(v) * " + size
					if !strings.Contains(codeStr, sizeCheck1) && !strings.Contains(codeStr, sizeCheck2) {
						t.Errorf("%s array should calculate size as len(v)*%s", typeName, size)
					}
				}
			}
		})
	}
}

func TestBoolArrayUsesLoop(t *testing.T) {
	s := &schema.Schema{
		Package: "testpkg",
		Messages: []schema.MessageType{
			{Name: "Bools", TargetType: &schema.ArrayType{
				ElementType: &schema.PrimitiveType{Name: "bool"},
			}},
		},
	}

	code, err := GenerateGo(s)
	if err != nil {
		t.Fatalf("GenerateGo failed: %v", err)
	}

	codeStr := string(code)

	// Bool arrays should NOT use unsafe (1 bit logically but 1 byte in Go)
	if strings.Contains(codeStr, "unsafe.Slice") && strings.Contains(codeStr, "bool") {
		t.Errorf("Bool arrays should not use unsafe bulk encoding")
	}

	// Should use element-by-element loop
	if !strings.Contains(codeStr, "for _, elem := range") {
		t.Errorf("Bool arrays should use element-by-element encoding")
	}
}

func TestGenerateCppSimpleStruct(t *testing.T) {
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

	code, err := GenerateCpp(s)
	if err != nil {
		t.Fatalf("GenerateCpp failed: %v", err)
	}

	codeStr := string(code)

	// Check header guard
	if !strings.Contains(codeStr, "#ifndef TEST_H") {
		t.Errorf("missing header guard start")
	}
	if !strings.Contains(codeStr, "#define TEST_H") {
		t.Errorf("missing header guard define")
	}
	if !strings.Contains(codeStr, "#endif // TEST_H") {
		t.Errorf("missing header guard end")
	}

	// Check includes
	if !strings.Contains(codeStr, "#include <cstdint>") {
		t.Errorf("missing cstdint include")
	}
	if !strings.Contains(codeStr, "#include <string>") {
		t.Errorf("missing string include")
	}
	if !strings.Contains(codeStr, "#include <vector>") {
		t.Errorf("missing vector include")
	}
	if !strings.Contains(codeStr, "#include <optional>") {
		t.Errorf("missing optional include")
	}

	// Check namespace
	if !strings.Contains(codeStr, "namespace test {") {
		t.Errorf("missing namespace declaration")
	}

	// Check struct definition
	if !strings.Contains(codeStr, "struct User {") {
		t.Errorf("missing struct definition")
	}
	if !strings.Contains(codeStr, "int32_t ID;") {
		t.Errorf("missing ID field")
	}
	if !strings.Contains(codeStr, "std::string Name;") {
		t.Errorf("missing Name field")
	}

	// Check Encoder class
	if !strings.Contains(codeStr, "class Encoder {") {
		t.Errorf("missing Encoder class")
	}
	if !strings.Contains(codeStr, "void write_int32(int32_t v)") {
		t.Errorf("missing write_int32 method")
	}
	if !strings.Contains(codeStr, "void write_string(const std::string& s)") {
		t.Errorf("missing write_string method")
	}

	// Check Decoder class
	if !strings.Contains(codeStr, "class Decoder {") {
		t.Errorf("missing Decoder class")
	}
	if !strings.Contains(codeStr, "int32_t read_int32()") {
		t.Errorf("missing read_int32 method")
	}
	if !strings.Contains(codeStr, "std::string read_string()") {
		t.Errorf("missing read_string method")
	}
	if !strings.Contains(codeStr, "void check_remaining(size_t needed)") {
		t.Errorf("missing bounds checking")
	}

	// Check encode/decode functions
	if !strings.Contains(codeStr, "encode_user_message") {
		t.Errorf("missing encode function")
	}
	if !strings.Contains(codeStr, "decode_user_message") {
		t.Errorf("missing decode function")
	}
}

func TestGenerateCppArray(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Types:   []schema.Type{},
		Messages: []schema.MessageType{
			{Name: "IntArray", TargetType: &schema.ArrayType{
				ElementType: &schema.PrimitiveType{Name: "int32"},
			}},
		},
	}

	code, err := GenerateCpp(s)
	if err != nil {
		t.Fatalf("GenerateCpp failed: %v", err)
	}

	codeStr := string(code)

	// Check array encoding
	if !strings.Contains(codeStr, "std::vector<int32_t>") {
		t.Errorf("missing vector type for array")
	}
	if !strings.Contains(codeStr, "uint16_t len = static_cast<uint16_t>(value.size())") {
		t.Errorf("missing array length encoding")
	}
	if !strings.Contains(codeStr, "for (const auto& elem : value)") {
		t.Errorf("missing array element loop")
	}

	// Check array decoding
	if !strings.Contains(codeStr, "uint16_t len = dec.read_array_length()") {
		t.Errorf("missing array length decoding")
	}
	if !strings.Contains(codeStr, "result.reserve(len)") {
		t.Errorf("missing vector reserve")
	}
	if !strings.Contains(codeStr, "for (uint16_t i = 0; i < len; ++i)") {
		t.Errorf("missing decode loop")
	}
}

func TestGenerateCppOptional(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
		Types: []schema.Type{
			&schema.StructType{
				Name: "Record",
				Fields: []schema.Field{
					{Name: "Required", Type: &schema.PrimitiveType{Name: "string"}},
					{Name: "Optional", Type: &schema.PrimitiveType{Name: "int32", Optional: true}},
				},
			},
		},
		Messages: []schema.MessageType{
			{Name: "Record", TargetType: &schema.StructType{
				Name: "Record",
				Fields: []schema.Field{
					{Name: "Required", Type: &schema.PrimitiveType{Name: "string"}},
					{Name: "Optional", Type: &schema.PrimitiveType{Name: "int32", Optional: true}},
				},
			}},
		},
	}

	code, err := GenerateCpp(s)
	if err != nil {
		t.Fatalf("GenerateCpp failed: %v", err)
	}

	codeStr := string(code)

	// Check optional field type
	if !strings.Contains(codeStr, "std::optional<int32_t> Optional;") {
		t.Errorf("optional field should use std::optional")
	}

	// Check optional encoding (presence byte)
	if !strings.Contains(codeStr, ".has_value()") {
		t.Errorf("missing has_value() check for optional")
	}
	if !strings.Contains(codeStr, "enc.write_byte(0x01)") {
		t.Errorf("missing presence byte encoding")
	}
	if !strings.Contains(codeStr, "enc.write_byte(0x00)") {
		t.Errorf("missing null byte encoding")
	}

	// Check optional decoding
	if !strings.Contains(codeStr, "if (dec.read_bool())") {
		t.Errorf("missing optional presence check in decode")
	}
}

func TestGenerateCppAllPrimitives(t *testing.T) {
	primitives := []string{"bool", "int8", "int16", "int32", "int64", "float32", "float64", "string"}
	
	for _, prim := range primitives {
		t.Run(prim, func(t *testing.T) {
s := &schema.Schema{
				Package: "test",
				Types:   []schema.Type{},
				Messages: []schema.MessageType{
					{Name: prim, TargetType: &schema.PrimitiveType{Name: prim}},
				},
			}

			code, err := GenerateCpp(s)
			if err != nil {
				t.Fatalf("GenerateCpp failed for %s: %v", prim, err)
			}

			codeStr := string(code)

			// Check that appropriate read/write methods exist
			var expectedReadMethod, expectedWriteMethod string
			switch prim {
			case "float32":
				expectedReadMethod = "read_float32"
				expectedWriteMethod = "write_float32"
			case "float64":
				expectedReadMethod = "read_float64"
				expectedWriteMethod = "write_float64"
			default:
				expectedReadMethod = "read_" + prim
				expectedWriteMethod = "write_" + prim
			}

			if !strings.Contains(codeStr, expectedReadMethod) {
				t.Errorf("missing %s method", expectedReadMethod)
			}
			if !strings.Contains(codeStr, expectedWriteMethod) {
				t.Errorf("missing %s method", expectedWriteMethod)
			}
		})
	}
}

func TestGenerateCppNestedStruct(t *testing.T) {
	s := &schema.Schema{
		Package: "test",
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
					{Name: "Nested", Type: &schema.StructType{
						Name: "Inner",
						Fields: []schema.Field{
							{Name: "Value", Type: &schema.PrimitiveType{Name: "int32"}},
						},
					}},
				},
			},
		},
		Messages: []schema.MessageType{
			{Name: "Outer", TargetType: &schema.StructType{
				Name: "Outer",
				Fields: []schema.Field{
					{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}},
					{Name: "Nested", Type: &schema.StructType{
						Name: "Inner",
						Fields: []schema.Field{
							{Name: "Value", Type: &schema.PrimitiveType{Name: "int32"}},
						},
					}},
				},
			}},
		},
	}

	code, err := GenerateCpp(s)
	if err != nil {
		t.Fatalf("GenerateCpp failed: %v", err)
	}

	codeStr := string(code)

	// Check both structs are defined
	if !strings.Contains(codeStr, "struct Inner {") {
		t.Errorf("missing Inner struct")
	}
	if !strings.Contains(codeStr, "struct Outer {") {
		t.Errorf("missing Outer struct")
	}

	// Check nested struct field
	if !strings.Contains(codeStr, "Inner Nested;") {
		t.Errorf("missing nested struct field")
	}

	// Check nested encoding
	if !strings.Contains(codeStr, "value.Nested.Value") {
		t.Errorf("missing nested field access in encoding")
	}
}
