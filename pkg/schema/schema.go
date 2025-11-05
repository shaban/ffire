// Package schema provides the AST representation for ffire schemas.
package schema

// Schema represents a complete .ffi schema file.
type Schema struct {
	Package  string        // Package name
	Messages []MessageType // Message types (public encode/decode)
	Types    []Type        // All type definitions
}

// MessageType represents a type alias that generates public encode/decode.
// Example: type DeviceList = []Device
type MessageType struct {
	Name       string // Alias name (e.g., "DeviceList")
	TargetType Type   // What it aliases (e.g., ArrayType of Device)
}

// Type represents any type in the schema.
type Type interface {
	TypeName() string
	IsOptional() bool
}

// PrimitiveType represents built-in types.
type PrimitiveType struct {
	Name     string // "bool", "int8", "int16", "int32", "int64", "float32", "float64", "string"
	Optional bool
}

func (p *PrimitiveType) TypeName() string { return p.Name }
func (p *PrimitiveType) IsOptional() bool { return p.Optional }

// StructType represents a struct definition.
type StructType struct {
	Name     string
	Fields   []Field
	Optional bool
}

func (s *StructType) TypeName() string { return s.Name }
func (s *StructType) IsOptional() bool { return s.Optional }

// Field represents a struct field.
type Field struct {
	Name    string
	Type    Type
	Tag     string // Full struct tag (e.g., `json:"name" yaml:"name" db:"name"`)
	jsonTag string // Cached JSON tag name for internal use
}

// JSONName returns the JSON field name (from json tag if present, otherwise field name).
func (f *Field) JSONName() string {
	if f.jsonTag != "" {
		return f.jsonTag
	}
	return f.Name
}

// SetJSONTag sets the cached JSON tag for internal use.
func (f *Field) SetJSONTag(tag string) {
	f.jsonTag = tag
}

// ArrayType represents an array/slice type.
type ArrayType struct {
	ElementType Type
	Optional    bool
}

func (a *ArrayType) TypeName() string {
	return "[]" + a.ElementType.TypeName()
}
func (a *ArrayType) IsOptional() bool { return a.Optional }

// Validate checks if the schema is well-formed.
func (s *Schema) Validate() error {
	// TODO: Implement validation rules:
	// - No circular type references
	// - All referenced types exist
	// - Max nesting depth (32)
	// - Valid primitive type names
	return nil
}

// FindType looks up a type by name in the schema.
func (s *Schema) FindType(name string) Type {
	// Check primitives
	if IsPrimitive(name) {
		return &PrimitiveType{Name: name}
	}

	// Check defined types
	for _, t := range s.Types {
		if t.TypeName() == name {
			return t
		}
	}

	return nil
}

// IsPrimitive checks if a type name is a built-in primitive.
func IsPrimitive(name string) bool {
	primitives := map[string]bool{
		"bool":    true,
		"int8":    true,
		"int16":   true,
		"int32":   true,
		"int64":   true,
		"float32": true,
		"float64": true,
		"string":  true,
	}
	return primitives[name]
}
