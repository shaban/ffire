// Package generator generates encoder/decoder code for various languages.
package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"

	"github.com/shaban/ffire/pkg/schema"
)

// Language represents target language for code generation.
type Language string

const (
	LanguageGo    Language = "go"
	LanguageCpp   Language = "cpp"
	LanguageSwift Language = "swift"
)

// Generate generates encoder/decoder code for the specified language.
func Generate(s *schema.Schema, lang Language) ([]byte, error) {
	switch lang {
	case LanguageGo:
		return GenerateGo(s)
	case LanguageCpp:
		return GenerateCpp(s)
	case LanguageSwift:
		return GenerateSwift(s)
	default:
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}
}

// GenerateGo generates Go encoder/decoder code.
func GenerateGo(s *schema.Schema) ([]byte, error) {
	gen := &goGenerator{schema: s, buf: &bytes.Buffer{}}
	return gen.generate()
}

type goGenerator struct {
	schema     *schema.Schema
	buf        *bytes.Buffer
	varCounter int
}

func (g *goGenerator) uniqueVar(prefix string) string {
	g.varCounter++
	return fmt.Sprintf("%s%d", prefix, g.varCounter)
}

func (g *goGenerator) schemaHasStrings() bool {
	return g.typeContainsString(g.schema.Messages[0].TargetType)
}

func (g *goGenerator) typeContainsString(typ schema.Type) bool {
	switch t := typ.(type) {
	case *schema.PrimitiveType:
		return t.Name == "string"
	case *schema.ArrayType:
		return g.typeContainsString(t.ElementType)
	case *schema.StructType:
		for _, field := range t.Fields {
			if g.typeContainsString(field.Type) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (g *goGenerator) schemaHasFloats() bool {
	return g.typeContainsFloat(g.schema.Messages[0].TargetType)
}

func (g *goGenerator) typeContainsFloat(typ schema.Type) bool {
	switch t := typ.(type) {
	case *schema.PrimitiveType:
		return t.Name == "float32" || t.Name == "float64"
	case *schema.ArrayType:
		return g.typeContainsFloat(t.ElementType)
	case *schema.StructType:
		for _, field := range t.Fields {
			if g.typeContainsFloat(field.Type) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (g *goGenerator) schemaHasPrimitiveArrays() bool {
	return g.typeContainsPrimitiveArray(g.schema.Messages[0].TargetType)
}

func (g *goGenerator) typeContainsPrimitiveArray(typ schema.Type) bool {
	switch t := typ.(type) {
	case *schema.ArrayType:
		// Check if this is an array of non-optional, non-string primitives (uses unsafe)
		if primType, ok := t.ElementType.(*schema.PrimitiveType); ok && !primType.Optional && primType.Name != "string" && primType.Name != "bool" {
			return true
		}
		// Recursively check element type
		return g.typeContainsPrimitiveArray(t.ElementType)
	case *schema.StructType:
		for _, field := range t.Fields {
			if g.typeContainsPrimitiveArray(field.Type) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (g *goGenerator) generate() ([]byte, error) {
	// Package declaration
	fmt.Fprintf(g.buf, "package %s\n\n", g.schema.Package)

	// Imports
	g.buf.WriteString("import (\n")
	g.buf.WriteString("\"bytes\"\n")
	// Only import math if schema contains floats (needed for Float32bits/Float64bits)
	if g.schemaHasFloats() {
		g.buf.WriteString("\"math\"\n")
	}
	// Import unsafe for zero-copy array encoding (reinterpret []T as []byte)
	if g.schemaHasPrimitiveArrays() {
		g.buf.WriteString("\"unsafe\"\n")
	}
	g.buf.WriteString(")\n\n")

	// Generate type definitions (structs)
	for _, typ := range g.schema.Types {
		if structType, ok := typ.(*schema.StructType); ok {
			g.generateStruct(structType)
		}
	}

	// Generate public message encode/decode functions
	for _, msg := range g.schema.Messages {
		g.generateMessageEncode(msg)
		g.generateMessageDecode(msg)
	}

	// Generate private helper functions
	for _, typ := range g.schema.Types {
		if structType, ok := typ.(*schema.StructType); ok {
			g.generateStructHelpers(structType)
		}
	}

	// Format the code
	formatted, err := format.Source(g.buf.Bytes())
	if err != nil {
		// Return unformatted code with error for debugging
		return g.buf.Bytes(), fmt.Errorf("format go code: %w", err)
	}

	return formatted, nil
}

func (g *goGenerator) generateStruct(structType *schema.StructType) {
	fmt.Fprintf(g.buf, "type %s struct {\n", structType.Name)
	for _, field := range structType.Fields {
		typeStr := g.goTypeString(field.Type)
		if field.Tag != "" {
			fmt.Fprintf(g.buf, "%s %s %s\n", field.Name, typeStr, field.Tag)
		} else {
			fmt.Fprintf(g.buf, "%s %s\n", field.Name, typeStr)
		}
	}
	g.buf.WriteString("}\n\n")
}

func (g *goGenerator) goTypeString(typ schema.Type) string {
	switch t := typ.(type) {
	case *schema.PrimitiveType:
		prefix := ""
		if t.Optional {
			prefix = "*"
		}
		return prefix + t.Name

	case *schema.StructType:
		prefix := ""
		if t.Optional {
			prefix = "*"
		}
		return prefix + t.Name

	case *schema.ArrayType:
		prefix := ""
		if t.Optional {
			prefix = "*"
		}
		return prefix + "[]" + g.goTypeString(t.ElementType)

	default:
		return "interface{}"
	}
}

func (g *goGenerator) generateMessageEncode(msg schema.MessageType) {
	// Determine root type name for function naming
	rootTypeName := g.rootTypeName(msg.TargetType)
	funcName := fmt.Sprintf("Encode%sMessage", rootTypeName)

	// Function signature
	paramType := g.goTypeString(msg.TargetType)
	fmt.Fprintf(g.buf, "// %s encodes %s to binary wire format.\n", funcName, msg.Name)
	fmt.Fprintf(g.buf, "func %s(v %s) []byte {\n", funcName, paramType)

	// Use default buffer - bytes.Buffer automatically grows efficiently
	g.buf.WriteString("buf := &bytes.Buffer{}\n")
	g.generateEncodeValue("buf", "v", msg.TargetType)
	g.buf.WriteString("return buf.Bytes()\n")
	g.buf.WriteString("}\n\n")
}

func (g *goGenerator) generateMessageDecode(msg schema.MessageType) {
	// Determine root type name for function naming
	rootTypeName := g.rootTypeName(msg.TargetType)
	funcName := fmt.Sprintf("Decode%sMessage", rootTypeName)

	// Function signature
	returnType := g.goTypeString(msg.TargetType)
	fmt.Fprintf(g.buf, "// %s decodes %s from binary wire format.\n", funcName, msg.Name)
	fmt.Fprintf(g.buf, "func %s(data []byte) (%s, error) {\n", funcName, returnType)

	// Direct slice indexing - no Reader allocation
	g.buf.WriteString("var (\n")
	g.buf.WriteString("result " + returnType + "\n")
	g.buf.WriteString("pos int\n")
	g.buf.WriteString(")\n")

	g.generateDecodeValueDirect("data", "pos", "result", msg.TargetType, false)
	g.buf.WriteString("return result, nil\n")
	g.buf.WriteString("}\n\n")
}
func (g *goGenerator) rootTypeName(typ schema.Type) string {
	switch t := typ.(type) {
	case *schema.PrimitiveType:
		// Capitalize first letter: int32 -> Int32
		return strings.Title(t.Name)
	case *schema.StructType:
		return t.Name
	case *schema.ArrayType:
		return g.rootTypeName(t.ElementType)
	default:
		return "Unknown"
	}
}

func (g *goGenerator) generateEncodeValue(bufVar, valueVar string, typ schema.Type) {
	switch t := typ.(type) {
	case *schema.PrimitiveType:
		g.generateEncodePrimitive(bufVar, valueVar, t)
	case *schema.StructType:
		g.generateEncodeStruct(bufVar, valueVar, t)
	case *schema.ArrayType:
		g.generateEncodeArray(bufVar, valueVar, t)
	}
}

func (g *goGenerator) generateEncodePrimitive(bufVar, valueVar string, typ *schema.PrimitiveType) {
	if typ.Optional {
		fmt.Fprintf(g.buf, "if %s == nil {\n", valueVar)
		fmt.Fprintf(g.buf, "%s.WriteByte(0x00)\n", bufVar)
		g.buf.WriteString("} else {\n")
		fmt.Fprintf(g.buf, "%s.WriteByte(0x01)\n", bufVar)
		valueVar = "*" + valueVar
	}

	switch typ.Name {
	case "bool":
		fmt.Fprintf(g.buf, "if %s {\n", valueVar)
		fmt.Fprintf(g.buf, "%s.WriteByte(0x01)\n", bufVar)
		g.buf.WriteString("} else {\n")
		fmt.Fprintf(g.buf, "%s.WriteByte(0x00)\n", bufVar)
		g.buf.WriteString("}\n")
	case "int8":
		fmt.Fprintf(g.buf, "%s.WriteByte(byte(%s))\n", bufVar, valueVar)
	case "int16":
		fmt.Fprintf(g.buf, "{ v := uint16(%s); %s.WriteByte(byte(v)); %s.WriteByte(byte(v>>8)) }\n", valueVar, bufVar, bufVar)
	case "int32":
		fmt.Fprintf(g.buf, "{ v := uint32(%s); %s.WriteByte(byte(v)); %s.WriteByte(byte(v>>8)); %s.WriteByte(byte(v>>16)); %s.WriteByte(byte(v>>24)) }\n", valueVar, bufVar, bufVar, bufVar, bufVar)
	case "int64":
		fmt.Fprintf(g.buf, "{ v := uint64(%s); %s.WriteByte(byte(v)); %s.WriteByte(byte(v>>8)); %s.WriteByte(byte(v>>16)); %s.WriteByte(byte(v>>24)); %s.WriteByte(byte(v>>32)); %s.WriteByte(byte(v>>40)); %s.WriteByte(byte(v>>48)); %s.WriteByte(byte(v>>56)) }\n", valueVar, bufVar, bufVar, bufVar, bufVar, bufVar, bufVar, bufVar, bufVar)
	case "float32":
		fmt.Fprintf(g.buf, "{ v := math.Float32bits(%s); %s.WriteByte(byte(v)); %s.WriteByte(byte(v>>8)); %s.WriteByte(byte(v>>16)); %s.WriteByte(byte(v>>24)) }\n", valueVar, bufVar, bufVar, bufVar, bufVar)
	case "float64":
		fmt.Fprintf(g.buf, "{ v := math.Float64bits(%s); %s.WriteByte(byte(v)); %s.WriteByte(byte(v>>8)); %s.WriteByte(byte(v>>16)); %s.WriteByte(byte(v>>24)); %s.WriteByte(byte(v>>32)); %s.WriteByte(byte(v>>40)); %s.WriteByte(byte(v>>48)); %s.WriteByte(byte(v>>56)) }\n", valueVar, bufVar, bufVar, bufVar, bufVar, bufVar, bufVar, bufVar, bufVar)
	case "string":
		fmt.Fprintf(g.buf, "{ l := uint16(len(%s)); %s.WriteByte(byte(l)); %s.WriteByte(byte(l>>8)) }\n", valueVar, bufVar, bufVar)
		fmt.Fprintf(g.buf, "%s.WriteString(%s)\n", bufVar, valueVar)
	}

	if typ.Optional {
		g.buf.WriteString("}\n")
	}
}

func (g *goGenerator) generateEncodeStruct(bufVar, valueVar string, typ *schema.StructType) {
	if typ.Optional {
		fmt.Fprintf(g.buf, "if %s == nil {\n", valueVar)
		fmt.Fprintf(g.buf, "%s.WriteByte(0x00)\n", bufVar)
		g.buf.WriteString("} else {\n")
		fmt.Fprintf(g.buf, "%s.WriteByte(0x01)\n", bufVar)
		valueVar = "*" + valueVar
	}

	for _, field := range typ.Fields {
		fieldVar := valueVar + "." + field.Name
		g.generateEncodeValue(bufVar, fieldVar, field.Type)
	}

	if typ.Optional {
		g.buf.WriteString("}\n")
	}
}

func (g *goGenerator) generateEncodeArray(bufVar, valueVar string, typ *schema.ArrayType) {
	if typ.Optional {
		fmt.Fprintf(g.buf, "if %s == nil {\n", valueVar)
		fmt.Fprintf(g.buf, "%s.WriteByte(0x00)\n", bufVar)
		g.buf.WriteString("} else {\n")
		fmt.Fprintf(g.buf, "%s.WriteByte(0x01)\n", bufVar)
		valueVar = "*" + valueVar
	}

	// Write array length
	fmt.Fprintf(g.buf, "{ l := uint16(len(%s)); %s.WriteByte(byte(l)); %s.WriteByte(byte(l>>8)) }\n", valueVar, bufVar, bufVar)

	// Check if we can do bulk write for primitive arrays
	if primType, ok := typ.ElementType.(*schema.PrimitiveType); ok && !primType.Optional {
		g.generateBulkArrayEncode(bufVar, valueVar, primType)
	} else {
		// Fallback to element-by-element encoding
		fmt.Fprintf(g.buf, "for _, elem := range %s {\n", valueVar)
		g.generateEncodeValue(bufVar, "elem", typ.ElementType)
		g.buf.WriteString("}\n")
	}

	if typ.Optional {
		g.buf.WriteString("}\n")
	}
}

func (g *goGenerator) generateBulkArrayEncode(bufVar, valueVar string, primType *schema.PrimitiveType) {
	switch primType.Name {
	case "bool":
		// Bools need individual handling (can't bulk write due to 0x00/0x01 encoding)
		fmt.Fprintf(g.buf, "for _, elem := range %s {\n", valueVar)
		fmt.Fprintf(g.buf, "if elem { %s.WriteByte(0x01) } else { %s.WriteByte(0x00) }\n", bufVar, bufVar)
		g.buf.WriteString("}\n")
	case "int8":
		// int8/uint8 can be reinterpreted directly as []byte (no endianness issue)
		// No len check needed - unsafe.Slice handles empty slices correctly
		fmt.Fprintf(g.buf, "if len(%s) > 0 {\n", valueVar)
		fmt.Fprintf(g.buf, "%s.Write(unsafe.Slice((*byte)(unsafe.Pointer(&%s[0])), len(%s)))\n", bufVar, valueVar, valueVar)
		g.buf.WriteString("}\n")
	case "int16", "int32", "int64", "float32", "float64":
		// Zero-copy reinterpret for multi-byte types (little-endian wire format)
		typeSize := map[string]int{
			"int16":   2,
			"int32":   4,
			"int64":   8,
			"float32": 4,
			"float64": 8,
		}[primType.Name]

		// Reinterpret array as []byte using unsafe - zero-copy, no allocation
		// Keep len check for safety with unsafe pointer
		fmt.Fprintf(g.buf, "if len(%s) > 0 {\n", valueVar)
		fmt.Fprintf(g.buf, "%s.Write(unsafe.Slice((*byte)(unsafe.Pointer(&%s[0])), len(%s)*%d))\n",
			bufVar, valueVar, valueVar, typeSize)
		g.buf.WriteString("}\n")

	case "string":
		// Strings need individual length prefixes
		fmt.Fprintf(g.buf, "for _, elem := range %s {\n", valueVar)
		fmt.Fprintf(g.buf, "{ l := uint16(len(elem)); %s.WriteByte(byte(l)); %s.WriteByte(byte(l>>8)) }\n", bufVar, bufVar)
		fmt.Fprintf(g.buf, "%s.WriteString(elem)\n", bufVar)
		g.buf.WriteString("}\n")
	}
}

func (g *goGenerator) generateDecodeValue(readerVar, resultVar string, typ schema.Type, isPointer bool) {
	switch t := typ.(type) {
	case *schema.PrimitiveType:
		g.generateDecodePrimitive(readerVar, resultVar, t, isPointer)
	case *schema.StructType:
		g.generateDecodeStruct(readerVar, resultVar, t, isPointer)
	case *schema.ArrayType:
		g.generateDecodeArray(readerVar, resultVar, t, isPointer)
	}
}

func (g *goGenerator) generateDecodePrimitive(readerVar, resultVar string, typ *schema.PrimitiveType, isPointer bool) {
	if typ.Optional {
		presentVar := g.uniqueVar("present")
		fmt.Fprintf(g.buf, "%s, err := %s.ReadByte()\n", presentVar, readerVar)
		g.buf.WriteString("if err != nil {\n")
		g.buf.WriteString("return result, fmt.Errorf(\"read optional flag: %w\", err)\n")
		g.buf.WriteString("}\n")
		fmt.Fprintf(g.buf, "if %s == 0x01 {\n", presentVar)

		// Allocate pointer
		tmpVar := g.uniqueVar("tmp")
		fmt.Fprintf(g.buf, "var %s %s\n", tmpVar, typ.Name)
		g.decodeNonOptionalPrimitive(readerVar, tmpVar, typ)
		fmt.Fprintf(g.buf, "%s = &%s\n", resultVar, tmpVar)

		g.buf.WriteString("}\n")
		return
	}

	if isPointer {
		tmpVar := g.uniqueVar("tmp")
		fmt.Fprintf(g.buf, "var %s %s\n", tmpVar, typ.Name)
		g.decodeNonOptionalPrimitive(readerVar, tmpVar, typ)
		fmt.Fprintf(g.buf, "%s = &%s\n", resultVar, tmpVar)
	} else {
		g.decodeNonOptionalPrimitive(readerVar, resultVar, typ)
	}
}
func (g *goGenerator) decodeNonOptionalPrimitive(readerVar, resultVar string, typ *schema.PrimitiveType) {
	switch typ.Name {
	case "bool":
		bVar := g.uniqueVar("b")
		fmt.Fprintf(g.buf, "var %s byte\n", bVar)
		fmt.Fprintf(g.buf, "%s, err = %s.ReadByte()\n", bVar, readerVar)
		g.buf.WriteString("if err != nil {\n")
		g.buf.WriteString("return result, fmt.Errorf(\"read bool: %w\", err)\n")
		g.buf.WriteString("}\n")
		fmt.Fprintf(g.buf, "%s = %s == 0x01\n", resultVar, bVar)
	case "int8":
		bVar := g.uniqueVar("b")
		fmt.Fprintf(g.buf, "%s, err := %s.ReadByte()\n", bVar, readerVar)
		g.buf.WriteString("if err != nil {\n")
		g.buf.WriteString("return result, fmt.Errorf(\"read int8: %w\", err)\n")
		g.buf.WriteString("}\n")
		fmt.Fprintf(g.buf, "%s = int8(%s)\n", resultVar, bVar)
	case "int16":
		bVar := g.uniqueVar("b")
		fmt.Fprintf(g.buf, "{ var %s [2]byte; _, err = %s.Read(%s[:]); if err != nil { return result, fmt.Errorf(\"read int16: %%w\", err) }; %s = int16(uint16(%s[0]) | uint16(%s[1])<<8) }\n", bVar, readerVar, bVar, resultVar, bVar, bVar)
	case "int32":
		bVar := g.uniqueVar("b")
		fmt.Fprintf(g.buf, "{ var %s [4]byte; _, err = %s.Read(%s[:]); if err != nil { return result, fmt.Errorf(\"read int32: %%w\", err) }; %s = int32(uint32(%s[0]) | uint32(%s[1])<<8 | uint32(%s[2])<<16 | uint32(%s[3])<<24) }\n", bVar, readerVar, bVar, resultVar, bVar, bVar, bVar, bVar)
	case "int64":
		bVar := g.uniqueVar("b")
		fmt.Fprintf(g.buf, "{ var %s [8]byte; _, err = %s.Read(%s[:]); if err != nil { return result, fmt.Errorf(\"read int64: %%w\", err) }; %s = int64(uint64(%s[0]) | uint64(%s[1])<<8 | uint64(%s[2])<<16 | uint64(%s[3])<<24 | uint64(%s[4])<<32 | uint64(%s[5])<<40 | uint64(%s[6])<<48 | uint64(%s[7])<<56) }\n", bVar, readerVar, bVar, resultVar, bVar, bVar, bVar, bVar, bVar, bVar, bVar, bVar)
	case "float32":
		bVar := g.uniqueVar("b")
		fmt.Fprintf(g.buf, "{ var %s [4]byte; _, err = %s.Read(%s[:]); if err != nil { return result, fmt.Errorf(\"read float32: %%w\", err) }; %s = math.Float32frombits(uint32(%s[0]) | uint32(%s[1])<<8 | uint32(%s[2])<<16 | uint32(%s[3])<<24) }\n", bVar, readerVar, bVar, resultVar, bVar, bVar, bVar, bVar)
	case "float64":
		bVar := g.uniqueVar("b")
		fmt.Fprintf(g.buf, "{ var %s [8]byte; _, err = %s.Read(%s[:]); if err != nil { return result, fmt.Errorf(\"read float64: %%w\", err) }; %s = math.Float64frombits(uint64(%s[0]) | uint64(%s[1])<<8 | uint64(%s[2])<<16 | uint64(%s[3])<<24 | uint64(%s[4])<<32 | uint64(%s[5])<<40 | uint64(%s[6])<<48 | uint64(%s[7])<<56) }\n", bVar, readerVar, bVar, resultVar, bVar, bVar, bVar, bVar, bVar, bVar, bVar, bVar)
	case "string":
		lenVar := g.uniqueVar("length")
		bytesVar := g.uniqueVar("strBytes")
		bVar := g.uniqueVar("b")
		fmt.Fprintf(g.buf, "{ var %s [2]byte; _, err = %s.Read(%s[:]); if err != nil { return result, fmt.Errorf(\"read string length: %%w\", err) }; %s := uint16(%s[0]) | uint16(%s[1])<<8\n", bVar, readerVar, bVar, lenVar, bVar, bVar)
		fmt.Fprintf(g.buf, "%s := make([]byte, %s)\n", bytesVar, lenVar)
		fmt.Fprintf(g.buf, "_, err = io.ReadFull(%s, %s)\n", readerVar, bytesVar)
		g.buf.WriteString("if err != nil {\n")
		g.buf.WriteString("return result, fmt.Errorf(\"read string data: %w\", err)\n")
		g.buf.WriteString("}\n")
		fmt.Fprintf(g.buf, "%s = string(%s) }\n", resultVar, bytesVar)
	}
}

func (g *goGenerator) generateDecodeStruct(readerVar, resultVar string, typ *schema.StructType, isPointer bool) {
	if typ.Optional {
		presentVar := g.uniqueVar("present")
		fmt.Fprintf(g.buf, "%s, err := %s.ReadByte()\n", presentVar, readerVar)
		g.buf.WriteString("if err != nil {\n")
		g.buf.WriteString("return result, fmt.Errorf(\"read optional flag: %w\", err)\n")
		g.buf.WriteString("}\n")
		fmt.Fprintf(g.buf, "if %s == 0x01 {\n", presentVar)

		// Allocate pointer and decode into it
		tmpVar := g.uniqueVar("tmp")
		fmt.Fprintf(g.buf, "%s := &%s{}\n", tmpVar, typ.Name)
		for _, field := range typ.Fields {
			g.generateDecodeValue(readerVar, tmpVar+"."+field.Name, field.Type, false)
		}
		fmt.Fprintf(g.buf, "%s = %s\n", resultVar, tmpVar)

		g.buf.WriteString("}\n")
		return
	}

	// For non-optional structs
	if isPointer {
		tmpVar := g.uniqueVar("tmp")
		fmt.Fprintf(g.buf, "%s := &%s{}\n", tmpVar, typ.Name)
		for _, field := range typ.Fields {
			g.generateDecodeValue(readerVar, tmpVar+"."+field.Name, field.Type, false)
		}
		fmt.Fprintf(g.buf, "%s = %s\n", resultVar, tmpVar)
	} else {
		for _, field := range typ.Fields {
			g.generateDecodeValue(readerVar, resultVar+"."+field.Name, field.Type, false)
		}
	}
}
func (g *goGenerator) generateDecodeArray(readerVar, resultVar string, typ *schema.ArrayType, isPointer bool) {
	if typ.Optional {
		presentVar := g.uniqueVar("present")
		fmt.Fprintf(g.buf, "%s, err := %s.ReadByte()\n", presentVar, readerVar)
		g.buf.WriteString("if err != nil {\n")
		g.buf.WriteString("return result, fmt.Errorf(\"read optional flag: %w\", err)\n")
		g.buf.WriteString("}\n")
		fmt.Fprintf(g.buf, "if %s == 0x01 {\n", presentVar)
	}

	// Read array length
	lenVar := g.uniqueVar("length")
	bVar := g.uniqueVar("b")
	fmt.Fprintf(g.buf, "{ var %s [2]byte; _, err = %s.Read(%s[:]); if err != nil { return result, fmt.Errorf(\"read array length: %%w\", err) }; %s := uint16(%s[0]) | uint16(%s[1])<<8\n", bVar, readerVar, bVar, lenVar, bVar, bVar)

	// Determine element type string
	elemTypeStr := g.goTypeString(typ.ElementType)

	// Allocate slice
	sliceVar := g.uniqueVar("tmpSlice")
	fmt.Fprintf(g.buf, "%s := make([]%s, %s)\n", sliceVar, elemTypeStr, lenVar)
	fmt.Fprintf(g.buf, "for i := range %s {\n", sliceVar)
	g.generateDecodeValue(readerVar, sliceVar+"[i]", typ.ElementType, false)
	g.buf.WriteString("}\n")

	if typ.Optional {
		fmt.Fprintf(g.buf, "%s = &%s }\n", resultVar, sliceVar)
		g.buf.WriteString("}\n")
	} else if isPointer {
		fmt.Fprintf(g.buf, "%s = &%s }\n", resultVar, sliceVar)
	} else {
		fmt.Fprintf(g.buf, "%s = %s }\n", resultVar, sliceVar)
	}
}
func (g *goGenerator) generateStructHelpers(structType *schema.StructType) {
	// Private helper functions will be added here
	// For now, we just generate encode/decode for message types
}

// Direct slice indexing decode methods (zero-copy, no Reader allocation)
func (g *goGenerator) generateDecodeValueDirect(dataVar, posVar, resultVar string, typ schema.Type, isPointer bool) {
	switch t := typ.(type) {
	case *schema.PrimitiveType:
		g.generateDecodePrimitiveDirect(dataVar, posVar, resultVar, t, isPointer)
	case *schema.StructType:
		g.generateDecodeStructDirect(dataVar, posVar, resultVar, t, isPointer)
	case *schema.ArrayType:
		g.generateDecodeArrayDirect(dataVar, posVar, resultVar, t, isPointer)
	}
}

func (g *goGenerator) generateDecodePrimitiveDirect(dataVar, posVar, resultVar string, typ *schema.PrimitiveType, isPointer bool) {
	if typ.Optional {
		presentVar := g.uniqueVar("present")
		fmt.Fprintf(g.buf, "%s := %s[%s]; %s++\n", presentVar, dataVar, posVar, posVar)
		fmt.Fprintf(g.buf, "if %s == 0x01 {\n", presentVar)

		tmpVar := g.uniqueVar("tmp")
		fmt.Fprintf(g.buf, "var %s %s\n", tmpVar, typ.Name)
		g.decodeNonOptionalPrimitiveDirect(dataVar, posVar, tmpVar, typ)
		fmt.Fprintf(g.buf, "%s = &%s\n", resultVar, tmpVar)

		g.buf.WriteString("}\n")
		return
	}

	if isPointer {
		tmpVar := g.uniqueVar("tmp")
		fmt.Fprintf(g.buf, "var %s %s\n", tmpVar, typ.Name)
		g.decodeNonOptionalPrimitiveDirect(dataVar, posVar, tmpVar, typ)
		fmt.Fprintf(g.buf, "%s = &%s\n", resultVar, tmpVar)
	} else {
		g.decodeNonOptionalPrimitiveDirect(dataVar, posVar, resultVar, typ)
	}
}

func (g *goGenerator) decodeNonOptionalPrimitiveDirect(dataVar, posVar, resultVar string, typ *schema.PrimitiveType) {
	switch typ.Name {
	case "bool":
		fmt.Fprintf(g.buf, "%s = %s[%s] == 0x01; %s++\n", resultVar, dataVar, posVar, posVar)
	case "int8":
		fmt.Fprintf(g.buf, "%s = int8(%s[%s]); %s++\n", resultVar, dataVar, posVar, posVar)
	case "int16":
		fmt.Fprintf(g.buf, "%s = int16(uint16(%s[%s]) | uint16(%s[%s+1])<<8); %s += 2\n", resultVar, dataVar, posVar, dataVar, posVar, posVar)
	case "int32":
		fmt.Fprintf(g.buf, "%s = int32(uint32(%s[%s]) | uint32(%s[%s+1])<<8 | uint32(%s[%s+2])<<16 | uint32(%s[%s+3])<<24); %s += 4\n", resultVar, dataVar, posVar, dataVar, posVar, dataVar, posVar, dataVar, posVar, posVar)
	case "int64":
		fmt.Fprintf(g.buf, "%s = int64(uint64(%s[%s]) | uint64(%s[%s+1])<<8 | uint64(%s[%s+2])<<16 | uint64(%s[%s+3])<<24 | uint64(%s[%s+4])<<32 | uint64(%s[%s+5])<<40 | uint64(%s[%s+6])<<48 | uint64(%s[%s+7])<<56); %s += 8\n", resultVar, dataVar, posVar, dataVar, posVar, dataVar, posVar, dataVar, posVar, dataVar, posVar, dataVar, posVar, dataVar, posVar, dataVar, posVar, posVar)
	case "float32":
		fmt.Fprintf(g.buf, "%s = math.Float32frombits(uint32(%s[%s]) | uint32(%s[%s+1])<<8 | uint32(%s[%s+2])<<16 | uint32(%s[%s+3])<<24); %s += 4\n", resultVar, dataVar, posVar, dataVar, posVar, dataVar, posVar, dataVar, posVar, posVar)
	case "float64":
		fmt.Fprintf(g.buf, "%s = math.Float64frombits(uint64(%s[%s]) | uint64(%s[%s+1])<<8 | uint64(%s[%s+2])<<16 | uint64(%s[%s+3])<<24 | uint64(%s[%s+4])<<32 | uint64(%s[%s+5])<<40 | uint64(%s[%s+6])<<48 | uint64(%s[%s+7])<<56); %s += 8\n", resultVar, dataVar, posVar, dataVar, posVar, dataVar, posVar, dataVar, posVar, dataVar, posVar, dataVar, posVar, dataVar, posVar, dataVar, posVar, posVar)
	case "string":
		lenVar := g.uniqueVar("length")
		fmt.Fprintf(g.buf, "%s := uint16(%s[%s]) | uint16(%s[%s+1])<<8; %s += 2\n", lenVar, dataVar, posVar, dataVar, posVar, posVar)
		// Safe string copy - creates independent copy to avoid lifetime issues
		fmt.Fprintf(g.buf, "%s = string(%s[%s:%s+int(%s)]); %s += int(%s)\n", resultVar, dataVar, posVar, posVar, lenVar, posVar, lenVar)
	}
}

func (g *goGenerator) generateDecodeStructDirect(dataVar, posVar, resultVar string, typ *schema.StructType, isPointer bool) {
	if typ.Optional {
		presentVar := g.uniqueVar("present")
		fmt.Fprintf(g.buf, "%s := %s[%s]; %s++\n", presentVar, dataVar, posVar, posVar)
		fmt.Fprintf(g.buf, "if %s == 0x01 {\n", presentVar)

		tmpVar := g.uniqueVar("tmp")
		fmt.Fprintf(g.buf, "%s := &%s{}\n", tmpVar, typ.Name)
		for _, field := range typ.Fields {
			g.generateDecodeValueDirect(dataVar, posVar, tmpVar+"."+field.Name, field.Type, false)
		}
		fmt.Fprintf(g.buf, "%s = %s\n", resultVar, tmpVar)

		g.buf.WriteString("}\n")
		return
	}

	if isPointer {
		tmpVar := g.uniqueVar("tmp")
		fmt.Fprintf(g.buf, "%s := &%s{}\n", tmpVar, typ.Name)
		for _, field := range typ.Fields {
			g.generateDecodeValueDirect(dataVar, posVar, tmpVar+"."+field.Name, field.Type, false)
		}
		fmt.Fprintf(g.buf, "%s = %s\n", resultVar, tmpVar)
	} else {
		for _, field := range typ.Fields {
			g.generateDecodeValueDirect(dataVar, posVar, resultVar+"."+field.Name, field.Type, false)
		}
	}
}

func (g *goGenerator) generateDecodeArrayDirect(dataVar, posVar, resultVar string, typ *schema.ArrayType, isPointer bool) {
	if typ.Optional {
		presentVar := g.uniqueVar("present")
		fmt.Fprintf(g.buf, "%s := %s[%s]; %s++\n", presentVar, dataVar, posVar, posVar)
		fmt.Fprintf(g.buf, "if %s == 0x01 {\n", presentVar)
	}

	// Read array length
	lenVar := g.uniqueVar("length")
	fmt.Fprintf(g.buf, "%s := uint16(%s[%s]) | uint16(%s[%s+1])<<8; %s += 2\n", lenVar, dataVar, posVar, dataVar, posVar, posVar)

	// Determine element type string
	elemTypeStr := g.goTypeString(typ.ElementType)

	// Allocate slice
	sliceVar := g.uniqueVar("tmpSlice")
	fmt.Fprintf(g.buf, "%s := make([]%s, %s)\n", sliceVar, elemTypeStr, lenVar)
	fmt.Fprintf(g.buf, "for i := range %s {\n", sliceVar)
	g.generateDecodeValueDirect(dataVar, posVar, sliceVar+"[i]", typ.ElementType, false)
	g.buf.WriteString("}\n")

	if typ.Optional {
		fmt.Fprintf(g.buf, "%s = &%s\n", resultVar, sliceVar)
		g.buf.WriteString("}\n")
	} else if isPointer {
		fmt.Fprintf(g.buf, "%s = &%s\n", resultVar, sliceVar)
	} else {
		fmt.Fprintf(g.buf, "%s = %s\n", resultVar, sliceVar)
	}
}

// GenerateCpp generates C++ encoder/decoder code.
func GenerateCpp(s *schema.Schema) ([]byte, error) {
	return nil, fmt.Errorf("C++ generation not yet implemented")
}

// GenerateSwift generates Swift encoder/decoder code.
func GenerateSwift(s *schema.Schema) ([]byte, error) {
	return nil, fmt.Errorf("Swift generation not yet implemented")
}
