package igniffi

import (
	"fmt"
	"strings"

	"github.com/shaban/ffire/pkg/schema"
)

// GenerateStructs generates C struct definitions for all schema types
func GenerateStructs(s *schema.Schema) string {
	var b strings.Builder

	b.WriteString("// ============================================================================\n")
	b.WriteString("// Generated Struct Definitions\n")
	b.WriteString("// ============================================================================\n\n")

	// Forward declarations first
	b.WriteString("// Forward declarations\n")
	for _, typ := range s.Types {
		if structType, ok := typ.(*schema.StructType); ok {
			fmt.Fprintf(&b, "typedef struct igniffi_%s igniffi_%s;\n", toCIdentifier(structType.Name), toCIdentifier(structType.Name))
		}
	}
	// Root message types
	for _, msg := range s.Messages {
		if structType, ok := msg.TargetType.(*schema.StructType); ok {
			fmt.Fprintf(&b, "typedef struct igniffi_%s igniffi_%s;\n", toCIdentifier(structType.Name), toCIdentifier(structType.Name))
		} else if _, ok := msg.TargetType.(*schema.ArrayType); ok {
			// Array root type - create wrapper struct
			fmt.Fprintf(&b, "typedef struct igniffi_%s igniffi_%s;\n", toCIdentifier(msg.Name), toCIdentifier(msg.Name))
		}
	}
	b.WriteString("\n")

	// Generate struct definitions in order
	for _, typ := range s.Types {
		if structType, ok := typ.(*schema.StructType); ok {
			generateStructDef(&b, structType)
		}
	}

	// Generate root message types
	for _, msg := range s.Messages {
		if structType, ok := msg.TargetType.(*schema.StructType); ok {
			// Root struct type - just use the struct definition
			// (already generated above if it's in Types)
			isInTypes := false
			for _, typ := range s.Types {
				if st, ok := typ.(*schema.StructType); ok && st.Name == structType.Name {
					isInTypes = true
					break
				}
			}
			if !isInTypes {
				generateStructDef(&b, structType)
			}
		} else if arrayType, ok := msg.TargetType.(*schema.ArrayType); ok {
			// Array root type - create wrapper struct
			generateArrayRootStruct(&b, msg.Name, arrayType)
		}
	}

	return b.String()
}

func generateStructDef(b *strings.Builder, structType *schema.StructType) {
	structName := toCIdentifier(structType.Name)

	fmt.Fprintf(b, "struct igniffi_%s {\n", structName)

	for _, field := range structType.Fields {
		fieldName := toCIdentifier(field.Name)
		cType := getCType(field.Type)
		fmt.Fprintf(b, "    %s %s;\n", cType, fieldName)

		// If it's an array, add length field
		if _, isArray := field.Type.(*schema.ArrayType); isArray {
			fmt.Fprintf(b, "    uint16_t %s_len;\n", fieldName)
		}

		// If it's optional, add has_{field} boolean
		if field.Type.IsOptional() {
			fmt.Fprintf(b, "    bool has_%s;\n", fieldName)
		}
	}

	fmt.Fprintf(b, "};\n\n")
}

func generateArrayRootStruct(b *strings.Builder, msgName string, arrayType *schema.ArrayType) {
	structName := toCIdentifier(msgName)
	elementType := getCType(arrayType.ElementType)

	// Remove trailing * if present (for struct pointers)
	elementType = strings.TrimSuffix(elementType, "*")

	fmt.Fprintf(b, "struct igniffi_%s {\n", structName)
	fmt.Fprintf(b, "    %s* items;\n", elementType)
	fmt.Fprintf(b, "    uint16_t len;\n")
	fmt.Fprintf(b, "};\n\n")
}

// GenerateAccessors generates inline accessor functions for all structs
func GenerateAccessors(s *schema.Schema) string {
	var b strings.Builder

	b.WriteString("// ============================================================================\n")
	b.WriteString("// Inline Accessor Functions (zero overhead)\n")
	b.WriteString("// ============================================================================\n\n")

	// Generate accessors for embedded structs
	for _, typ := range s.Types {
		if structType, ok := typ.(*schema.StructType); ok {
			generateStructAccessors(&b, structType)
		}
	}

	// Generate accessors for root message types
	for _, msg := range s.Messages {
		if structType, ok := msg.TargetType.(*schema.StructType); ok {
			// Check if already generated (might be in Types too)
			isInTypes := false
			for _, typ := range s.Types {
				if st, ok := typ.(*schema.StructType); ok && st.Name == structType.Name {
					isInTypes = true
					break
				}
			}
			if !isInTypes {
				generateStructAccessors(&b, structType)
			}
		} else if arrayType, ok := msg.TargetType.(*schema.ArrayType); ok {
			// Array root type accessors
			generateArrayRootAccessors(&b, msg.Name, arrayType)
		}
	}

	return b.String()
}

func generateStructAccessors(b *strings.Builder, structType *schema.StructType) {
	structName := toCIdentifier(structType.Name)

	fmt.Fprintf(b, "// %s accessors\n", structType.Name)

	for _, field := range structType.Fields {
		fieldName := toCIdentifier(field.Name)

		// Generate getter
		if arrayType, ok := field.Type.(*schema.ArrayType); ok {
			// Array field - return pointer + length getter
			elementType := getCType(arrayType.ElementType)
			elementType = strings.TrimSuffix(elementType, "*")

			// Length getter
			fmt.Fprintf(b, "static inline uint16_t igniffi_%s_get_%s_len(const igniffi_%s* obj) {\n",
				structName, fieldName, structName)
			fmt.Fprintf(b, "    return obj->%s_len;\n", fieldName)
			fmt.Fprintf(b, "}\n\n")

			// Array element getter (const)
			fmt.Fprintf(b, "static inline const %s* igniffi_%s_get_%s(const igniffi_%s* obj, uint16_t index) {\n",
				elementType, structName, fieldName, structName)
			fmt.Fprintf(b, "    if (index >= obj->%s_len) return NULL;\n", fieldName)
			fmt.Fprintf(b, "    return &obj->%s[index];\n", fieldName)
			fmt.Fprintf(b, "}\n\n")

			// Array element getter (mutable)
			fmt.Fprintf(b, "static inline %s* igniffi_%s_get_mutable_%s(igniffi_%s* obj, uint16_t index) {\n",
				elementType, structName, fieldName, structName)
			fmt.Fprintf(b, "    if (index >= obj->%s_len) return NULL;\n", fieldName)
			fmt.Fprintf(b, "    return &obj->%s[index];\n", fieldName)
			fmt.Fprintf(b, "}\n\n")

		} else if primitiveType, ok := field.Type.(*schema.PrimitiveType); ok {
			cType := getCType(field.Type)

			// Getter
			fmt.Fprintf(b, "static inline %s igniffi_%s_get_%s(const igniffi_%s* obj) {\n",
				cType, structName, fieldName, structName)
			fmt.Fprintf(b, "    return obj->%s;\n", fieldName)
			fmt.Fprintf(b, "}\n\n")

			// Setter
			if primitiveType.Name == "string" {
				// String setter needs arena for copying
				fmt.Fprintf(b, "static inline void igniffi_%s_set_%s(igniffi_%s* obj, igniffi_StringView val, igniffi_Arena* arena) {\n",
					structName, fieldName, structName)
				fmt.Fprintf(b, "    obj->%s = igniffi_stringview_copy(val, arena);\n", fieldName)
				fmt.Fprintf(b, "}\n\n")
			} else {
				// Primitive setter
				fmt.Fprintf(b, "static inline void igniffi_%s_set_%s(igniffi_%s* obj, %s val) {\n",
					structName, fieldName, structName, cType)
				fmt.Fprintf(b, "    obj->%s = val;\n", fieldName)
				fmt.Fprintf(b, "}\n\n")
			}

		} else if _, ok := field.Type.(*schema.StructType); ok {
			// Nested struct - return pointer
			cType := getCType(field.Type)

			// Const getter
			fmt.Fprintf(b, "static inline const %s igniffi_%s_get_%s(const igniffi_%s* obj) {\n",
				cType, structName, fieldName, structName)
			fmt.Fprintf(b, "    return obj->%s;\n", fieldName)
			fmt.Fprintf(b, "}\n\n")

			// Mutable getter
			fmt.Fprintf(b, "static inline %s igniffi_%s_get_mutable_%s(igniffi_%s* obj) {\n",
				cType, structName, fieldName, structName)
			fmt.Fprintf(b, "    return obj->%s;\n", fieldName)
			fmt.Fprintf(b, "}\n\n")
		}
	}

	b.WriteString("\n")
}

func generateArrayRootAccessors(b *strings.Builder, msgName string, arrayType *schema.ArrayType) {
	structName := toCIdentifier(msgName)
	elementType := getCType(arrayType.ElementType)
	elementType = strings.TrimSuffix(elementType, "*")

	fmt.Fprintf(b, "// %s accessors (array root type)\n", msgName)

	// Length getter
	fmt.Fprintf(b, "static inline uint16_t igniffi_%s_len(const igniffi_%s* list) {\n",
		structName, structName)
	fmt.Fprintf(b, "    return list->len;\n")
	fmt.Fprintf(b, "}\n\n")

	// Element getter (const)
	fmt.Fprintf(b, "static inline const %s* igniffi_%s_get(const igniffi_%s* list, uint16_t index) {\n",
		elementType, structName, structName)
	fmt.Fprintf(b, "    if (index >= list->len) return NULL;\n")
	fmt.Fprintf(b, "    return &list->items[index];\n")
	fmt.Fprintf(b, "}\n\n")

	// Element getter (mutable)
	fmt.Fprintf(b, "static inline %s* igniffi_%s_get_mutable(igniffi_%s* list, uint16_t index) {\n",
		elementType, structName, structName)
	fmt.Fprintf(b, "    if (index >= list->len) return NULL;\n")
	fmt.Fprintf(b, "    return &list->items[index];\n")
	fmt.Fprintf(b, "}\n\n")
}

// getCType returns the C type for a schema type
func getCType(t schema.Type) string {
	switch typ := t.(type) {
	case *schema.PrimitiveType:
		switch typ.Name {
		case "bool":
			return "bool"
		case "int8":
			return "int8_t"
		case "int16":
			return "int16_t"
		case "int32":
			return "int32_t"
		case "int64":
			return "int64_t"
		case "float32":
			return "float"
		case "float64":
			return "double"
		case "string":
			return "igniffi_StringView"
		}
	case *schema.ArrayType:
		// Arrays are represented as pointer to element type
		elemType := getCType(typ.ElementType)
		elemType = strings.TrimSuffix(elemType, "*")
		return elemType + "*"
	case *schema.StructType:
		return "igniffi_" + toCIdentifier(typ.Name) + "*"
	}
	return "void*"
}

// toCIdentifier converts a name to a valid C identifier (lowercase with underscores)
func toCIdentifier(name string) string {
	var result strings.Builder
	for i, r := range name {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(r + 32) // Convert to lowercase
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
