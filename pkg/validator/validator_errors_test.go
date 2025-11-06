package validator

import (
	"testing"

	"github.com/shaban/ffire/pkg/errors"
	"github.com/shaban/ffire/pkg/schema"
)

func TestValidateSchema_ErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		schema   *schema.Schema
		wantCode errors.ErrorCode
	}{
		{
			name: "empty package",
			schema: &schema.Schema{
				Package: "",
				Messages: []schema.MessageType{
					{Name: "Test", TargetType: &schema.PrimitiveType{Name: "string"}},
				},
			},
			wantCode: errors.ErrEmptyPackage,
		},
		{
			name: "no messages",
			schema: &schema.Schema{
				Package:  "test",
				Messages: []schema.MessageType{},
			},
			wantCode: errors.ErrNoMessages,
		},
		{
			name: "empty message name",
			schema: &schema.Schema{
				Package: "test",
				Messages: []schema.MessageType{
					{Name: "", TargetType: &schema.PrimitiveType{Name: "string"}},
				},
			},
			wantCode: errors.ErrEmptyMessageName,
		},
		{
			name: "nil target type",
			schema: &schema.Schema{
				Package: "test",
				Messages: []schema.MessageType{
					{Name: "Test", TargetType: nil},
				},
			},
			wantCode: errors.ErrNilTargetType,
		},
		{
			name: "undefined type",
			schema: &schema.Schema{
				Package: "test",
				Messages: []schema.MessageType{
					{Name: "Test", TargetType: &schema.PrimitiveType{Name: "UndefinedType"}},
				},
			},
			wantCode: errors.ErrUndefinedType,
		},
		{
			name: "empty struct",
			schema: &schema.Schema{
				Package: "test",
				Messages: []schema.MessageType{
					{Name: "Test", TargetType: &schema.StructType{Name: "Empty", Fields: []schema.Field{}}},
				},
			},
			wantCode: errors.ErrEmptyStruct,
		},
		{
			name: "empty field name",
			schema: &schema.Schema{
				Package: "test",
				Messages: []schema.MessageType{
					{
						Name: "Test",
						TargetType: &schema.StructType{
							Name: "Person",
							Fields: []schema.Field{
								{Name: "", Type: &schema.PrimitiveType{Name: "string"}},
							},
						},
					},
				},
			},
			wantCode: errors.ErrEmptyFieldName,
		},
		{
			name: "nil field type",
			schema: &schema.Schema{
				Package: "test",
				Messages: []schema.MessageType{
					{
						Name: "Test",
						TargetType: &schema.StructType{
							Name: "Person",
							Fields: []schema.Field{
								{Name: "name", Type: nil},
							},
						},
					},
				},
			},
			wantCode: errors.ErrNilFieldType,
		},
		{
			name: "nil array element",
			schema: &schema.Schema{
				Package: "test",
				Messages: []schema.MessageType{
					{
						Name:       "Test",
						TargetType: &schema.ArrayType{ElementType: nil},
					},
				},
			},
			wantCode: errors.ErrNilArrayElement,
		},
		{
			name: "circular reference",
			schema: &schema.Schema{
				Package: "test",
				Types: []schema.Type{
					&schema.StructType{
						Name: "Node",
						Fields: []schema.Field{
							{Name: "next", Type: &schema.PrimitiveType{Name: "Node"}},
						},
					},
				},
				Messages: []schema.MessageType{
					{Name: "Test", TargetType: &schema.PrimitiveType{Name: "Node"}},
				},
			},
			wantCode: errors.ErrCircularReference,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSchema(tt.schema)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !errors.IsCode(err, tt.wantCode) {
				t.Errorf("expected error code %s, got %s (error: %v)", 
					tt.wantCode, errors.GetCode(err), err)
			}
		})
	}
}

func TestValidateJSON_ErrorCodes(t *testing.T) {
	schema := &schema.Schema{
		Package: "test",
		Messages: []schema.MessageType{
			{
				Name: "Person",
				TargetType: &schema.StructType{
					Name: "Person",
					Fields: []schema.Field{
						{Name: "name", Type: &schema.PrimitiveType{Name: "string"}},
						{Name: "age", Type: &schema.PrimitiveType{Name: "int32"}},
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		json     string
		wantCode errors.ErrorCode
	}{
		{
			name:     "message not found",
			json:     `{}`,
			wantCode: errors.ErrMessageNotFound,
		},
		{
			name:     "invalid JSON",
			json:     `{invalid}`,
			wantCode: errors.ErrInvalidJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgName := "Person"
			if tt.name == "message not found" {
				msgName = "Unknown"
			}

			err := ValidateJSON(schema, msgName, []byte(tt.json))
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !errors.IsCode(err, tt.wantCode) {
				t.Errorf("expected error code %s, got %s (error: %v)", 
					tt.wantCode, errors.GetCode(err), err)
			}
		})
	}
}
