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

// Canonicalize sorts all struct fields in canonical wire format order.
// This should be called once before code generation.
// The canonical order is:
// 1. Fixed-size fields (8-byte, 4-byte, 2-byte, 1-byte), alphabetically within size
// 2. Variable-size fields (strings, arrays), alphabetically
// 3. Optional fields, alphabetically
func (s *Schema) Canonicalize() {
	// Canonicalize all struct types
	for _, t := range s.Types {
		if st, ok := t.(*StructType); ok {
			st.Fields = SortFieldsCanonical(st.Fields)
		}
	}
	// Canonicalize root message types that are structs
	for _, msg := range s.Messages {
		if st, ok := msg.TargetType.(*StructType); ok {
			st.Fields = SortFieldsCanonical(st.Fields)
		}
	}
}

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

// PrimitiveSize returns the byte size of a primitive type.
// Returns 0 for variable-size types like string.
func PrimitiveSize(name string) int {
	sizes := map[string]int{
		"bool":    1,
		"int8":    1,
		"int16":   2,
		"int32":   4,
		"int64":   8,
		"float32": 4,
		"float64": 8,
		"string":  0, // variable size
	}
	return sizes[name]
}

// FieldCategory represents the ordering category for canonical field ordering.
type FieldCategory int

const (
	CategoryFixed8    FieldCategory = iota // 8-byte fixed (int64, float64)
	CategoryFixed4                         // 4-byte fixed (int32, float32)
	CategoryFixed2                         // 2-byte fixed (int16)
	CategoryFixed1                         // 1-byte fixed (bool, int8)
	CategoryVariable                       // variable size (string, arrays)
	CategoryOptional                       // optional fields (any type)
)

// GetFieldCategory returns the canonical ordering category for a field.
func GetFieldCategory(f Field) FieldCategory {
	return getTypeCategory(f.Type)
}

func getTypeCategory(t Type) FieldCategory {
	switch typ := t.(type) {
	case *PrimitiveType:
		if typ.Optional {
			return CategoryOptional
		}
		switch typ.Name {
		case "int64", "float64":
			return CategoryFixed8
		case "int32", "float32":
			return CategoryFixed4
		case "int16":
			return CategoryFixed2
		case "bool", "int8":
			return CategoryFixed1
		case "string":
			return CategoryVariable
		}
	case *ArrayType:
		if typ.Optional {
			return CategoryOptional
		}
		return CategoryVariable
	case *StructType:
		if typ.Optional {
			return CategoryOptional
		}
		// Check if struct is fixed-size (all fields are non-optional fixed primitives)
		if IsFixedSizeStruct(typ) {
			return CategoryFixed8 // Treat as large fixed for ordering purposes
		}
		return CategoryVariable
	}
	return CategoryVariable
}

// IsFixedSizeStruct returns true if the struct has only fixed-size, non-optional fields.
func IsFixedSizeStruct(s *StructType) bool {
	for _, f := range s.Fields {
		if !IsFixedSizeType(f.Type) {
			return false
		}
	}
	return true
}

// IsFixedSizeType returns true if the type has a fixed byte size.
func IsFixedSizeType(t Type) bool {
	switch typ := t.(type) {
	case *PrimitiveType:
		if typ.Optional {
			return false
		}
		return typ.Name != "string"
	case *StructType:
		if typ.Optional {
			return false
		}
		return IsFixedSizeStruct(typ)
	case *ArrayType:
		return false // Arrays are always variable size
	}
	return false
}

// GetPrimitiveSize returns the byte size of a primitive type, or 0 if not fixed-size.
func GetPrimitiveSize(t Type) int {
	prim, ok := t.(*PrimitiveType)
	if !ok || prim.Optional {
		return 0
	}
	switch prim.Name {
	case "int64", "float64":
		return 8
	case "int32", "float32":
		return 4
	case "int16":
		return 2
	case "bool", "int8":
		return 1
	default:
		return 0 // string is variable
	}
}

// FixedFieldRun represents a contiguous run of fixed-size primitive fields.
type FixedFieldRun struct {
	StartIndex int
	EndIndex   int // exclusive
	TotalBytes int
}

// GetFixedFieldRuns returns runs of contiguous fixed-size primitive fields.
// After canonical ordering, all fixed-size fields are at the front.
func GetFixedFieldRuns(fields []Field) []FixedFieldRun {
	if len(fields) == 0 {
		return nil
	}

	var runs []FixedFieldRun
	runStart := -1
	runBytes := 0

	for i, field := range fields {
		size := GetPrimitiveSize(field.Type)
		if size > 0 {
			if runStart == -1 {
				runStart = i
				runBytes = size
			} else {
				runBytes += size
			}
		} else {
			// End of run
			if runStart != -1 && runBytes > 0 {
				runs = append(runs, FixedFieldRun{
					StartIndex: runStart,
					EndIndex:   i,
					TotalBytes: runBytes,
				})
			}
			runStart = -1
			runBytes = 0
		}
	}

	// Handle final run
	if runStart != -1 && runBytes > 0 {
		runs = append(runs, FixedFieldRun{
			StartIndex: runStart,
			EndIndex:   len(fields),
			TotalBytes: runBytes,
		})
	}

	return runs
}

// SortFieldsCanonical returns a copy of fields sorted in canonical wire format order:
// 1. Fixed-size fields (8-byte, then 4-byte, then 2-byte, then 1-byte), alphabetically within size
// 2. Variable-size fields (strings, arrays), alphabetically
// 3. Optional fields, alphabetically
func SortFieldsCanonical(fields []Field) []Field {
	// Make a copy to avoid modifying original
	sorted := make([]Field, len(fields))
	copy(sorted, fields)
	
	// Sort by category first, then alphabetically by name within category
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			catI := GetFieldCategory(sorted[i])
			catJ := GetFieldCategory(sorted[j])
			
			// Compare by category first
			if catI > catJ {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			} else if catI == catJ && sorted[i].Name > sorted[j].Name {
				// Same category: sort alphabetically
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	
	return sorted
}

