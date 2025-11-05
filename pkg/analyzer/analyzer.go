// Package analyzer analyzes schemas to enable code generation optimizations.
package analyzer

import (
	"github.com/shaban/ffire/pkg/schema"
)

// TypeInfo contains analysis results for a type.
type TypeInfo struct {
	IsFixedSize bool // All fields are non-optional primitives (no strings/arrays)?
	FixedSize   int  // Exact byte size if IsFixedSize=true
	MaxSize     int  // Maximum possible size with all optionals present
	HasStrings  bool // Contains any string fields?
	HasArrays   bool // Contains any array fields?
	NestDepth   int  // Maximum nesting depth
}

// Analyze analyzes all types in a schema and returns type information map.
// The map key is the type name.
func Analyze(s *schema.Schema) map[string]*TypeInfo {
	a := &analyzer{
		schema:   s,
		typeInfo: make(map[string]*TypeInfo),
		visiting: make(map[string]bool),
	}

	// Analyze all struct types
	for _, typ := range s.Types {
		if structType, ok := typ.(*schema.StructType); ok {
			a.analyzeType(structType.Name, structType)
		}
	}

	return a.typeInfo
}

type analyzer struct {
	schema   *schema.Schema
	typeInfo map[string]*TypeInfo
	visiting map[string]bool // Detect circular references
}

func (a *analyzer) analyzeType(name string, typ schema.Type) *TypeInfo {
	// Check cache
	if info, ok := a.typeInfo[name]; ok {
		return info
	}

	// Detect circular references
	if a.visiting[name] {
		// Circular reference - not fixed size, infinite max size
		return &TypeInfo{
			IsFixedSize: false,
			MaxSize:     -1, // Infinite
			NestDepth:   999,
		}
	}

	a.visiting[name] = true
	defer delete(a.visiting, name)

	info := a.computeTypeInfo(typ)
	a.typeInfo[name] = info
	return info
}

func (a *analyzer) computeTypeInfo(typ schema.Type) *TypeInfo {
	switch t := typ.(type) {
	case *schema.PrimitiveType:
		return a.analyzePrimitive(t)
	case *schema.StructType:
		return a.analyzeStruct(t)
	case *schema.ArrayType:
		return a.analyzeArray(t)
	default:
		return &TypeInfo{}
	}
}

func (a *analyzer) analyzePrimitive(typ *schema.PrimitiveType) *TypeInfo {
	info := &TypeInfo{
		NestDepth: 0,
	}

	// Non-optional primitive
	size := a.primitiveSize(typ.Name)

	if typ.Name == "string" {
		// Strings are not fixed size
		info.IsFixedSize = false
		info.HasStrings = true
		info.MaxSize = 2 + 65535 // uint16 length + max string
	} else {
		// Fixed-size primitive
		info.IsFixedSize = true
		info.FixedSize = size
		info.MaxSize = size
	}

	if typ.Optional {
		info.IsFixedSize = false
		info.FixedSize = 0
		info.MaxSize += 1 // Add optional flag to max size
	}

	return info
}

func (a *analyzer) analyzeStruct(typ *schema.StructType) *TypeInfo {
	info := &TypeInfo{
		IsFixedSize: true, // Assume fixed until proven otherwise
		NestDepth:   0,
	}

	if typ.Optional {
		info.IsFixedSize = false
		info.FixedSize = 0
		info.MaxSize = 1 // Start with optional flag
	}

	// Analyze each field
	maxFieldDepth := 0
	for _, field := range typ.Fields {
		fieldInfo := a.computeTypeInfo(field.Type)

		// Update flags
		if !fieldInfo.IsFixedSize {
			info.IsFixedSize = false
		}
		if fieldInfo.HasStrings {
			info.HasStrings = true
		}
		if fieldInfo.HasArrays {
			info.HasArrays = true
		}

		// Update sizes
		if info.IsFixedSize {
			info.FixedSize += fieldInfo.FixedSize
		}
		info.MaxSize += fieldInfo.MaxSize

		// Track depth - increment for nested structs/arrays
		fieldDepth := fieldInfo.NestDepth
		if _, isStruct := field.Type.(*schema.StructType); isStruct {
			fieldDepth++ // Nested struct adds a level
		} else if _, isArray := field.Type.(*schema.ArrayType); isArray {
			fieldDepth++ // Nested array adds a level
		}

		if fieldDepth > maxFieldDepth {
			maxFieldDepth = fieldDepth
		}
	}

	info.NestDepth = maxFieldDepth

	return info
}

func (a *analyzer) analyzeArray(typ *schema.ArrayType) *TypeInfo {
	elemInfo := a.computeTypeInfo(typ.ElementType)

	info := &TypeInfo{
		IsFixedSize: false, // Arrays are never fixed size (length varies)
		HasArrays:   true,
		MaxSize:     2 + (65535 * elemInfo.MaxSize), // uint16 length + max elements
		NestDepth:   elemInfo.NestDepth + 1,
	}

	if elemInfo.HasStrings {
		info.HasStrings = true
	}

	if typ.Optional {
		info.MaxSize += 1 // Optional flag
	}

	return info
}

func (a *analyzer) primitiveSize(name string) int {
	switch name {
	case "bool", "int8":
		return 1
	case "int16":
		return 2
	case "int32", "float32":
		return 4
	case "int64", "float64":
		return 8
	case "string":
		return 2 + 65535 // uint16 length + max data
	default:
		return 0
	}
}
