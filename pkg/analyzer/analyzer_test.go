package analyzer

import (
	"testing"

	"github.com/shaban/ffire/pkg/schema"
)

func TestAnalyzePrimitives(t *testing.T) {
	tests := []struct {
		name        string
		typ         *schema.PrimitiveType
		wantFixed   bool
		wantSize    int
		wantMaxSize int
	}{
		{
			name:        "int32",
			typ:         &schema.PrimitiveType{Name: "int32"},
			wantFixed:   true,
			wantSize:    4,
			wantMaxSize: 4,
		},
		{
			name:        "bool",
			typ:         &schema.PrimitiveType{Name: "bool"},
			wantFixed:   true,
			wantSize:    1,
			wantMaxSize: 1,
		},
		{
			name:        "int64",
			typ:         &schema.PrimitiveType{Name: "int64"},
			wantFixed:   true,
			wantSize:    8,
			wantMaxSize: 8,
		},
		{
			name:        "string",
			typ:         &schema.PrimitiveType{Name: "string"},
			wantFixed:   false,
			wantSize:    0,
			wantMaxSize: 65537, // 2 + 65535
		},
		{
			name:        "optional int32",
			typ:         &schema.PrimitiveType{Name: "int32", Optional: true},
			wantFixed:   false,
			wantSize:    0,
			wantMaxSize: 5, // 1 flag + 4 bytes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &schema.Schema{
				Package: "test",
				Types:   []schema.Type{tt.typ},
			}

			result := Analyze(s)
			if len(result) > 0 {
				t.Errorf("primitives shouldn't appear in result map")
			}

			// Test via direct computation
			a := &analyzer{
				schema:   s,
				typeInfo: make(map[string]*TypeInfo),
				visiting: make(map[string]bool),
			}
			info := a.computeTypeInfo(tt.typ)

			if info.IsFixedSize != tt.wantFixed {
				t.Errorf("IsFixedSize = %v, want %v", info.IsFixedSize, tt.wantFixed)
			}
			if info.FixedSize != tt.wantSize {
				t.Errorf("FixedSize = %d, want %d", info.FixedSize, tt.wantSize)
			}
			if info.MaxSize != tt.wantMaxSize {
				t.Errorf("MaxSize = %d, want %d", info.MaxSize, tt.wantMaxSize)
			}
		})
	}
}

func TestAnalyzeFixedStruct(t *testing.T) {
	// Struct with only fixed-size primitives
	s := &schema.Schema{
		Package: "test",
		Types: []schema.Type{
			&schema.StructType{
				Name: "Point",
				Fields: []schema.Field{
					{Name: "X", Type: &schema.PrimitiveType{Name: "float32"}},
					{Name: "Y", Type: &schema.PrimitiveType{Name: "float32"}},
				},
			},
		},
	}

	result := Analyze(s)
	info := result["Point"]

	if info == nil {
		t.Fatal("Point not in result")
	}

	if !info.IsFixedSize {
		t.Errorf("IsFixedSize = false, want true")
	}
	if info.FixedSize != 8 {
		t.Errorf("FixedSize = %d, want 8", info.FixedSize)
	}
	if info.MaxSize != 8 {
		t.Errorf("MaxSize = %d, want 8", info.MaxSize)
	}
	if info.HasStrings {
		t.Errorf("HasStrings = true, want false")
	}
	if info.HasArrays {
		t.Errorf("HasArrays = true, want false")
	}
	if info.NestDepth != 0 {
		t.Errorf("NestDepth = %d, want 0", info.NestDepth)
	}
}

func TestAnalyzeStructWithString(t *testing.T) {
	// Struct with string field
	s := &schema.Schema{
		Package: "test",
		Types: []schema.Type{
			&schema.StructType{
				Name: "Config",
				Fields: []schema.Field{
					{Name: "Host", Type: &schema.PrimitiveType{Name: "string"}},
					{Name: "Port", Type: &schema.PrimitiveType{Name: "int32"}},
				},
			},
		},
	}

	result := Analyze(s)
	info := result["Config"]

	if info == nil {
		t.Fatal("Config not in result")
	}

	if info.IsFixedSize {
		t.Errorf("IsFixedSize = true, want false")
	}
	if !info.HasStrings {
		t.Errorf("HasStrings = false, want true")
	}
	if info.MaxSize != 65541 { // 2+65535 (string) + 4 (int32)
		t.Errorf("MaxSize = %d, want 65541", info.MaxSize)
	}
}

func TestAnalyzeArray(t *testing.T) {
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

	result := Analyze(s)
	info := result["Item"]

	if info == nil {
		t.Fatal("Item not in result")
	}

	if !info.IsFixedSize {
		t.Errorf("Item.IsFixedSize = false, want true")
	}
	if info.FixedSize != 4 {
		t.Errorf("Item.FixedSize = %d, want 4", info.FixedSize)
	}

	// Test array analysis directly
	a := &analyzer{
		schema:   s,
		typeInfo: result,
		visiting: make(map[string]bool),
	}

	arrayInfo := a.computeTypeInfo(s.Messages[0].TargetType)

	if arrayInfo.IsFixedSize {
		t.Errorf("Array.IsFixedSize = true, want false")
	}
	if !arrayInfo.HasArrays {
		t.Errorf("Array.HasArrays = false, want true")
	}
	// 2 (length) + 65535 * 4 (max elements * element size)
	expectedMax := 2 + (65535 * 4)
	if arrayInfo.MaxSize != expectedMax {
		t.Errorf("Array.MaxSize = %d, want %d", arrayInfo.MaxSize, expectedMax)
	}
}

func TestAnalyzeNesting(t *testing.T) {
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
					{Name: "Data", Type: &schema.StructType{
						Name: "Inner",
						Fields: []schema.Field{
							{Name: "Value", Type: &schema.PrimitiveType{Name: "int32"}},
						},
					}},
				},
			},
		},
	}

	result := Analyze(s)

	innerInfo := result["Inner"]
	if innerInfo.NestDepth != 0 {
		t.Errorf("Inner.NestDepth = %d, want 0", innerInfo.NestDepth)
	}

	outerInfo := result["Outer"]
	if outerInfo.NestDepth != 1 {
		t.Errorf("Outer.NestDepth = %d, want 1", outerInfo.NestDepth)
	}
}

func TestAnalyzeOptionalStruct(t *testing.T) {
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
	}

	result := Analyze(s)
	info := result["Record"]

	if info == nil {
		t.Fatal("Record not in result")
	}

	if info.IsFixedSize {
		t.Errorf("IsFixedSize = true, want false (has optional field)")
	}
	if !info.HasStrings {
		t.Errorf("HasStrings = false, want true")
	}
	// 4 (int32) + 1 (optional flag) + 2 + 65535 (max string)
	expectedMax := 4 + 1 + 2 + 65535
	if info.MaxSize != expectedMax {
		t.Errorf("MaxSize = %d, want %d", info.MaxSize, expectedMax)
	}
}
