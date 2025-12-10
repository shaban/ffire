package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shaban/ffire/pkg/schema"
)

// swiftModuleKeywords lists Swift keywords that cannot be used as module/package names.
//
// Purpose:
//   - Sanitize schema.Package names for Swift module system
//   - Example: "package struct" → "import structModule" (struct is Swift keyword)
//
// Scope:
//   - Used ONLY for module/package name sanitization
//   - NOT used for type names (they get Message suffix: StructMessage)
//   - NOT used for field names or other identifiers
//
// Why type names don't need this:
//   - Type names automatically get "Message" suffix
//   - "type Struct" → "StructMessage" (no collision)
//   - "type Class" → "ClassMessage" (no collision)
//   - Universal collision avoidance across all 11 languages
//
// Swift's module system:
//   - Modules must not conflict with language keywords
//   - "import struct" is invalid (struct is keyword)
//   - "import structModule" is valid (appended suffix)
var swiftModuleKeywords = map[string]bool{
	"Any": true, "as": true, "associatedtype": true, "break": true, "case": true,
	"catch": true, "class": true, "continue": true, "default": true, "defer": true,
	"deinit": true, "do": true, "else": true, "enum": true, "extension": true,
	"fallthrough": true, "false": true, "fileprivate": true, "for": true, "func": true,
	"guard": true, "if": true, "import": true, "in": true, "init": true,
	"inout": true, "internal": true, "is": true, "let": true, "nil": true,
	"open": true, "operator": true, "private": true, "protocol": true, "public": true,
	"repeat": true, "rethrows": true, "return": true, "self": true, "Self": true,
	"static": true, "struct": true, "subscript": true, "super": true, "switch": true,
	"throw": true, "throws": true, "true": true, "try": true, "typealias": true,
	"var": true, "where": true, "while": true,
}

// SanitizeSwiftModuleName ensures the module name is not a Swift keyword.
// Appends "Module" suffix if the name conflicts with a Swift keyword.
//
// Used for: schema.Package (module names)
// Not used for: Type names (already get Message suffix)
//
// Exported for use by benchmark generation (pkg/benchmark/benchmark_swift.go).
func SanitizeSwiftModuleName(name string) string {
	if swiftModuleKeywords[name] {
		return name + "Module"
	}
	return name
}

// estimateStructSize calculates the minimum encoded size of a struct type.
// This is used for capacity pre-allocation to minimize buffer reallocations.
// Returns the sum of:
// - Fixed-size primitive fields
// - 2 bytes per string/array length prefix
// - 1 byte per optional presence flag
// - 16 bytes per string (heuristic for average string content)
// - Recursively calculates nested struct sizes
func estimateStructSize(structType *schema.StructType) int {
	size := 0
	for _, field := range structType.Fields {
		size += estimateFieldSize(field.Type)
	}
	return size
}

func estimateFieldSize(fieldType schema.Type) int {
	switch t := fieldType.(type) {
	case *schema.PrimitiveType:
		baseSize := 0
		switch t.Name {
		case "bool", "int8":
			baseSize = 1
		case "int16":
			baseSize = 2
		case "int32", "float32":
			baseSize = 4
		case "int64", "float64":
			baseSize = 8
		case "string":
			baseSize = 2 + 16 // length prefix + avg string content
		}
		if t.Optional {
			return 1 + baseSize // presence flag + value
		}
		return baseSize
	case *schema.ArrayType:
		elemSize := estimateFieldSize(t.ElementType)
		// Assume avg 5 elements + 2 bytes length prefix
		if t.Optional {
			return 1 + 2 + elemSize*5
		}
		return 2 + elemSize*5
	case *schema.StructType:
		return estimateStructSize(t)
	}
	return 32 // fallback
}

// GenerateSwiftPackage generates a complete Swift package using the orchestrator
func GenerateSwiftPackage(config *PackageConfig) error {
	// Sanitize the namespace to avoid Swift keywords
	config.Namespace = SanitizeSwiftModuleName(config.Namespace)

	return orchestrateTierBPackage(
		config,
		SwiftLayout,
		generateSwiftWrapperOrchestrated,
		generateSwiftMetadataOrchestrated,
		printSwiftInstructions,
	)
}

func generateSwiftWrapperOrchestrated(config *PackageConfig, paths *PackagePaths) error {
	// Generate native Swift code
	swiftCode, err := generateSwiftNative(config.Schema)
	if err != nil {
		return fmt.Errorf("failed to generate Swift code: %w", err)
	}

	// Create Sources directory structure
	sourcesDir := filepath.Join(paths.Root, "Sources", config.Namespace)
	if err := os.MkdirAll(sourcesDir, 0755); err != nil {
		return fmt.Errorf("failed to create Sources directory: %w", err)
	}

	// Write Swift source file
	swiftPath := filepath.Join(sourcesDir, "Generated.swift")
	if err := os.WriteFile(swiftPath, swiftCode, 0644); err != nil {
		return fmt.Errorf("failed to write Swift source: %w", err)
	}
	fmt.Printf("✓ Generated Swift source: %s\n", swiftPath)

	return nil
}

// generateSwiftNative generates pure Swift code optimized for maximum performance
func generateSwiftNative(s *schema.Schema) ([]byte, error) {
	var buf bytes.Buffer

	// File header
	buf.WriteString("// Generated by ffire - Native Swift implementation\n")
	buf.WriteString("// DO NOT EDIT - This file is auto-generated\n\n")
	buf.WriteString("import Foundation\n\n")

	// Generate message type definitions (root types with Message suffix)
	for _, msg := range s.Messages {
		if structType, ok := msg.TargetType.(*schema.StructType); ok {
			generateSwiftMessageStruct(&buf, msg.Name, structType)
		} else if arrayType, ok := msg.TargetType.(*schema.ArrayType); ok {
			// Array type alias
			elemTypeStr := getSwiftTypeString(arrayType.ElementType)
			buf.WriteString(fmt.Sprintf("public typealias %sMessage = [%s]\n\n", msg.Name, elemTypeStr))
		} else {
			// Primitive type alias
			typeStr := getSwiftTypeString(msg.TargetType)
			buf.WriteString(fmt.Sprintf("public typealias %sMessage = %s\n\n", msg.Name, typeStr))
		}
	}

	// Generate helper structs (embedded types, no Message suffix)
	for _, typ := range s.Types {
		if structType, ok := typ.(*schema.StructType); ok {
			// Skip if this is a root message type
			isRootType := false
			for _, msg := range s.Messages {
				if st, ok := msg.TargetType.(*schema.StructType); ok && st.Name == structType.Name {
					isRootType = true
					break
				}
			}
			if !isRootType {
				generateSwiftStruct(&buf, structType)
			}
		}
	}

	// Generate encode functions
	buf.WriteString("// MARK: - Encoding\n\n")
	for _, msg := range s.Messages {
		generateSwiftEncoderFunc(&buf, msg)
	}

	// Generate decode functions
	buf.WriteString("// MARK: - Decoding\n\n")
	for _, msg := range s.Messages {
		generateSwiftDecoderFunc(&buf, msg)
	}

	// Generate struct helper functions (only for referenced types, not root messages)
	buf.WriteString("// MARK: - Struct Helpers\n\n")
	// Build a set of root message type names
	rootMessageTypes := make(map[string]bool)
	for _, msg := range s.Messages {
		if st, ok := msg.TargetType.(*schema.StructType); ok {
			rootMessageTypes[st.Name] = true
		}
	}
	// Only generate helpers for non-root struct types
	for _, typ := range s.Types {
		if structType, ok := typ.(*schema.StructType); ok {
			// Skip if this is a root message type
			if !rootMessageTypes[structType.Name] {
				generateSwiftStructHelpers(&buf, structType)
			}
		}
	}

	// Generate helper functions
	generateSwiftHelpers(&buf)

	return buf.Bytes(), nil
}

func generateSwiftMessageStruct(buf *bytes.Buffer, messageName string, structType *schema.StructType) {
	structName := messageName + "Message"
	buf.WriteString(fmt.Sprintf("public struct %s {\n", structName))

	for _, field := range structType.Fields {
		swiftType := getSwiftTypeString(field.Type)
		fieldName := escapeSwiftFieldName(field.Name)
		buf.WriteString(fmt.Sprintf("    public var %s: %s\n", fieldName, swiftType))
	}

	// Generate memberwise initializer
	buf.WriteString("\n    public init(\n")
	for i, field := range structType.Fields {
		swiftType := getSwiftTypeString(field.Type)
		fieldName := escapeSwiftFieldName(field.Name)
		buf.WriteString(fmt.Sprintf("        %s: %s", fieldName, swiftType))
		if i < len(structType.Fields)-1 {
			buf.WriteString(",\n")
		} else {
			buf.WriteString("\n")
		}
	}
	buf.WriteString("    ) {\n")
	for _, field := range structType.Fields {
		fieldName := escapeSwiftFieldName(field.Name)
		buf.WriteString(fmt.Sprintf("        self.%s = %s\n", fieldName, fieldName))
	}
	buf.WriteString("    }\n")
	buf.WriteString("}\n\n")
}

func generateSwiftStruct(buf *bytes.Buffer, structType *schema.StructType) {
	buf.WriteString(fmt.Sprintf("public struct %s {\n", structType.Name))

	for _, field := range structType.Fields {
		swiftType := getSwiftTypeString(field.Type)
		fieldName := escapeSwiftFieldName(field.Name)
		buf.WriteString(fmt.Sprintf("    public var %s: %s\n", fieldName, swiftType))
	}

	// Generate memberwise initializer
	buf.WriteString("\n    public init(\n")
	for i, field := range structType.Fields {
		swiftType := getSwiftTypeString(field.Type)
		fieldName := escapeSwiftFieldName(field.Name)
		buf.WriteString(fmt.Sprintf("        %s: %s", fieldName, swiftType))
		if i < len(structType.Fields)-1 {
			buf.WriteString(",\n")
		} else {
			buf.WriteString("\n")
		}
	}
	buf.WriteString("    ) {\n")
	for _, field := range structType.Fields {
		fieldName := escapeSwiftFieldName(field.Name)
		buf.WriteString(fmt.Sprintf("        self.%s = %s\n", fieldName, fieldName))
	}
	buf.WriteString("    }\n")
	buf.WriteString("}\n\n")
}

func generateSwiftEncoderFunc(buf *bytes.Buffer, msg schema.MessageType) {
	structName := msg.Name + "Message"
	funcName := fmt.Sprintf("encode%sMessage", msg.Name)

	buf.WriteString("@inlinable\n")
	buf.WriteString(fmt.Sprintf("public func %s(_ message: %s) -> Data {\n", funcName, structName))
	buf.WriteString("    var buffer = [UInt8]()\n")
	// Dynamic capacity based on message type
	if arrayType, ok := msg.TargetType.(*schema.ArrayType); ok {
		if primType, ok := arrayType.ElementType.(*schema.PrimitiveType); ok {
			switch primType.Name {
			case "int8", "bool":
				buf.WriteString("    buffer.reserveCapacity(2 + message.count)\n")
			case "int16":
				buf.WriteString("    buffer.reserveCapacity(2 + message.count * 2)\n")
			case "int32", "float32":
				buf.WriteString("    buffer.reserveCapacity(2 + message.count * 4)\n")
			case "int64", "float64":
				buf.WriteString("    buffer.reserveCapacity(2 + message.count * 8)\n")
			default:
				// For strings, use heuristic based on average string length
				buf.WriteString("    buffer.reserveCapacity(max(1024, message.count * 32))\n")
			}
		} else if structType, ok := arrayType.ElementType.(*schema.StructType); ok {
			// For struct arrays, estimate based on struct field sizes
			estimatedSize := estimateStructSize(structType)
			buf.WriteString(fmt.Sprintf("    buffer.reserveCapacity(max(1024, message.count * %d))\n", estimatedSize))
		} else {
			// Fallback for other types
			buf.WriteString("    buffer.reserveCapacity(max(1024, message.count * 64))\n")
		}
	} else if structType, ok := msg.TargetType.(*schema.StructType); ok {
		// For struct messages, estimate based on struct field sizes
		estimatedSize := estimateStructSize(structType)
		buf.WriteString(fmt.Sprintf("    buffer.reserveCapacity(%d)\n", max(1024, estimatedSize)))
	} else {
		// Fallback
		buf.WriteString("    buffer.reserveCapacity(1024)\n")
	}

	switch t := msg.TargetType.(type) {
	case *schema.StructType:
		for _, field := range t.Fields {
			generateSwiftEncodeField(buf, field, "message."+field.Name)
		}
	case *schema.ArrayType:
		// For array types, encode as array
		buf.WriteString("    let len = UInt16(message.count)\n")
		if primType, ok := t.ElementType.(*schema.PrimitiveType); ok {
			switch primType.Name {
			case "bool":
				// Bool arrays need element-by-element conversion
				buf.WriteString("    withUnsafeBytes(of: len.littleEndian) { buffer.append(contentsOf: $0) }\n")
				buf.WriteString("    for item in message { buffer.append(item ? 1 : 0) }\n")
			case "int8":
				// Int8 arrays need bitPattern conversion
				buf.WriteString("    withUnsafeBytes(of: len.littleEndian) { buffer.append(contentsOf: $0) }\n")
				buf.WriteString("    for item in message { buffer.append(UInt8(bitPattern: item)) }\n")
			case "int16":
				// Optimized: pre-allocated Data with memcpy for Int16 arrays
				buf.WriteString("    let totalSize = 2 + message.count * 2\n")
				buf.WriteString("    var result = Data(count: totalSize)\n")
				buf.WriteString("    result.withUnsafeMutableBytes { ptr in\n")
				buf.WriteString("        let base = ptr.baseAddress!\n")
				buf.WriteString("        base.storeBytes(of: len.littleEndian, as: UInt16.self)\n")
				buf.WriteString("        message.withUnsafeBytes { src in\n")
				buf.WriteString("            _ = memcpy(base.advanced(by: 2), src.baseAddress!, message.count * 2)\n")
				buf.WriteString("        }\n")
				buf.WriteString("    }\n")
				buf.WriteString("    return result\n")
			case "int32":
				// Optimized: pre-allocated Data with memcpy for Int32 arrays
				buf.WriteString("    let totalSize = 2 + message.count * 4\n")
				buf.WriteString("    var result = Data(count: totalSize)\n")
				buf.WriteString("    result.withUnsafeMutableBytes { ptr in\n")
				buf.WriteString("        let base = ptr.baseAddress!\n")
				buf.WriteString("        base.storeBytes(of: len.littleEndian, as: UInt16.self)\n")
				buf.WriteString("        message.withUnsafeBytes { src in\n")
				buf.WriteString("            _ = memcpy(base.advanced(by: 2), src.baseAddress!, message.count * 4)\n")
				buf.WriteString("        }\n")
				buf.WriteString("    }\n")
				buf.WriteString("    return result\n")
			case "int64":
				// Optimized: pre-allocated Data with memcpy for Int64 arrays
				buf.WriteString("    let totalSize = 2 + message.count * 8\n")
				buf.WriteString("    var result = Data(count: totalSize)\n")
				buf.WriteString("    result.withUnsafeMutableBytes { ptr in\n")
				buf.WriteString("        let base = ptr.baseAddress!\n")
				buf.WriteString("        base.storeBytes(of: len.littleEndian, as: UInt16.self)\n")
				buf.WriteString("        message.withUnsafeBytes { src in\n")
				buf.WriteString("            _ = memcpy(base.advanced(by: 2), src.baseAddress!, message.count * 8)\n")
				buf.WriteString("        }\n")
				buf.WriteString("    }\n")
				buf.WriteString("    return result\n")
			case "float32":
				// Optimized: pre-allocated Data with memcpy for Float arrays
				buf.WriteString("    let totalSize = 2 + message.count * 4\n")
				buf.WriteString("    var result = Data(count: totalSize)\n")
				buf.WriteString("    result.withUnsafeMutableBytes { ptr in\n")
				buf.WriteString("        let base = ptr.baseAddress!\n")
				buf.WriteString("        base.storeBytes(of: len.littleEndian, as: UInt16.self)\n")
				buf.WriteString("        message.withUnsafeBytes { src in\n")
				buf.WriteString("            _ = memcpy(base.advanced(by: 2), src.baseAddress!, message.count * 4)\n")
				buf.WriteString("        }\n")
				buf.WriteString("    }\n")
				buf.WriteString("    return result\n")
			case "float64":
				// Optimized: pre-allocated Data with memcpy for Double arrays
				buf.WriteString("    let totalSize = 2 + message.count * 8\n")
				buf.WriteString("    var result = Data(count: totalSize)\n")
				buf.WriteString("    result.withUnsafeMutableBytes { ptr in\n")
				buf.WriteString("        let base = ptr.baseAddress!\n")
				buf.WriteString("        base.storeBytes(of: len.littleEndian, as: UInt16.self)\n")
				buf.WriteString("        message.withUnsafeBytes { src in\n")
				buf.WriteString("            _ = memcpy(base.advanced(by: 2), src.baseAddress!, message.count * 8)\n")
				buf.WriteString("        }\n")
				buf.WriteString("    }\n")
				buf.WriteString("    return result\n")
			case "string":
				// Two-pass encoding: calculate total size, then bulk write
				buf.WriteString("    // Fast path: pre-calculate total size and write in one pass\n")
				buf.WriteString("    var strings = message\n")
				buf.WriteString("    var totalSize = 2 // array length prefix\n")
				buf.WriteString("    for i in 0..<strings.count {\n")
				buf.WriteString("        strings[i].makeContiguousUTF8()\n")
				buf.WriteString("        totalSize += 2 + strings[i].utf8.count // length prefix + bytes\n")
				buf.WriteString("    }\n")
				buf.WriteString("    // Single allocation\n")
				buf.WriteString("    var result = Data(count: totalSize)\n")
				buf.WriteString("    result.withUnsafeMutableBytes { ptr in\n")
				buf.WriteString("        let base = ptr.baseAddress!\n")
				buf.WriteString("        var pos = 0\n")
				buf.WriteString("        // Write array length\n")
				buf.WriteString("        base.storeBytes(of: len.littleEndian, as: UInt16.self)\n")
				buf.WriteString("        pos = 2\n")
				buf.WriteString("        // Write each string using index to access mutated array\n")
				buf.WriteString("        for i in 0..<strings.count {\n")
				buf.WriteString("            strings[i].withUTF8 { utf8 in\n")
				buf.WriteString("                let strLen = UInt16(utf8.count)\n")
				buf.WriteString("                base.storeBytes(of: strLen.littleEndian, toByteOffset: pos, as: UInt16.self)\n")
				buf.WriteString("                pos += 2\n")
				buf.WriteString("                if let src = utf8.baseAddress {\n")
				buf.WriteString("                    memcpy(base.advanced(by: pos), src, utf8.count)\n")
				buf.WriteString("                }\n")
				buf.WriteString("                pos += utf8.count\n")
				buf.WriteString("            }\n")
				buf.WriteString("        }\n")
				buf.WriteString("    }\n")
				buf.WriteString("    return result\n")
			}
		} else if structType, ok := t.ElementType.(*schema.StructType); ok {
			// Check if struct has only primitive fields (no strings, arrays, or nested structs)
			hasOnlyPrimitives := true
			fixedSize := 0
			for _, field := range structType.Fields {
				if primType, ok := field.Type.(*schema.PrimitiveType); ok {
					switch primType.Name {
					case "bool", "int8":
						fixedSize += 1
					case "int16":
						fixedSize += 2
					case "int32", "float32":
						fixedSize += 4
					case "int64", "float64":
						fixedSize += 8
					case "string":
						hasOnlyPrimitives = false
					}
				} else {
					hasOnlyPrimitives = false
				}
			}

			if hasOnlyPrimitives && fixedSize > 0 {
				// For structs with only primitive fields, we can bulk allocate
				buf.WriteString(fmt.Sprintf("    // Optimized: struct has only primitives, fixed size = %d bytes\n", fixedSize))
				buf.WriteString(fmt.Sprintf("    let structSize = %d\n", fixedSize))
				buf.WriteString("    let totalSize = 2 + message.count * structSize\n")
				buf.WriteString("    var result = Data(count: totalSize)\n")
				buf.WriteString("    result.withUnsafeMutableBytes { ptr in\n")
				buf.WriteString("        let base = ptr.baseAddress!\n")
				buf.WriteString("        base.storeBytes(of: len.littleEndian, as: UInt16.self)\n")
				buf.WriteString("        var pos = 2\n")
				buf.WriteString("        for item in message {\n")
				// Generate inline field writes
				for _, field := range structType.Fields {
					if primType, ok := field.Type.(*schema.PrimitiveType); ok {
						accessor := fmt.Sprintf("item.%s", field.Name)
						switch primType.Name {
						case "bool":
							buf.WriteString(fmt.Sprintf("            base.storeBytes(of: %s ? UInt8(1) : UInt8(0), toByteOffset: pos, as: UInt8.self)\n", accessor))
							buf.WriteString("            pos += 1\n")
						case "int8":
							buf.WriteString(fmt.Sprintf("            base.storeBytes(of: UInt8(bitPattern: %s), toByteOffset: pos, as: UInt8.self)\n", accessor))
							buf.WriteString("            pos += 1\n")
						case "int16":
							buf.WriteString(fmt.Sprintf("            base.storeBytes(of: %s.littleEndian, toByteOffset: pos, as: Int16.self)\n", accessor))
							buf.WriteString("            pos += 2\n")
						case "int32":
							buf.WriteString(fmt.Sprintf("            base.storeBytes(of: %s.littleEndian, toByteOffset: pos, as: Int32.self)\n", accessor))
							buf.WriteString("            pos += 4\n")
						case "int64":
							buf.WriteString(fmt.Sprintf("            base.storeBytes(of: %s.littleEndian, toByteOffset: pos, as: Int64.self)\n", accessor))
							buf.WriteString("            pos += 8\n")
						case "float32":
							buf.WriteString(fmt.Sprintf("            base.storeBytes(of: %s.bitPattern.littleEndian, toByteOffset: pos, as: UInt32.self)\n", accessor))
							buf.WriteString("            pos += 4\n")
						case "float64":
							buf.WriteString(fmt.Sprintf("            base.storeBytes(of: %s.bitPattern.littleEndian, toByteOffset: pos, as: UInt64.self)\n", accessor))
							buf.WriteString("            pos += 8\n")
						}
					}
				}
				buf.WriteString("        }\n")
				buf.WriteString("    }\n")
				buf.WriteString("    return result\n")
			} else {
				// Check if struct has only non-optional primitives and strings (no arrays, nested structs, or optionals)
				hasOnlySimpleFields := true
				stringFields := []string{}
				fixedSize := 0
				for _, field := range structType.Fields {
					if primType, ok := field.Type.(*schema.PrimitiveType); ok {
						// Skip optional fields - they have variable encoding
						if primType.Optional {
							hasOnlySimpleFields = false
							continue
						}
						switch primType.Name {
						case "bool", "int8":
							fixedSize += 1
						case "int16":
							fixedSize += 2
						case "int32", "float32":
							fixedSize += 4
						case "int64", "float64":
							fixedSize += 8
						case "string":
							fixedSize += 2 // length prefix
							stringFields = append(stringFields, field.Name)
						}
					} else {
						hasOnlySimpleFields = false
					}
				}

				if hasOnlySimpleFields && len(stringFields) > 0 {
					// Two-pass optimization for structs with primitives and strings
					buf.WriteString("    // Two-pass: calculate total size, then bulk write\n")
					buf.WriteString("    var items = message\n")
					buf.WriteString(fmt.Sprintf("    var totalSize = 2 + message.count * %d // array len + fixed portion\n", fixedSize))
					buf.WriteString("    for i in 0..<items.count {\n")
					for _, sf := range stringFields {
						buf.WriteString(fmt.Sprintf("        items[i].%s.makeContiguousUTF8()\n", sf))
						buf.WriteString(fmt.Sprintf("        totalSize += items[i].%s.utf8.count\n", sf))
					}
					buf.WriteString("    }\n")
					buf.WriteString("    var result = Data(count: totalSize)\n")
					buf.WriteString("    result.withUnsafeMutableBytes { ptr in\n")
					buf.WriteString("        let base = ptr.baseAddress!\n")
					buf.WriteString("        base.storeBytes(of: len.littleEndian, as: UInt16.self)\n")
					buf.WriteString("        var pos = 2\n")
					buf.WriteString("        for i in 0..<items.count {\n")
					// Generate inline field writes
					for _, field := range structType.Fields {
						if primType, ok := field.Type.(*schema.PrimitiveType); ok {
							accessor := fmt.Sprintf("items[i].%s", field.Name)
							switch primType.Name {
							case "bool":
								buf.WriteString(fmt.Sprintf("            base.storeBytes(of: %s ? UInt8(1) : UInt8(0), toByteOffset: pos, as: UInt8.self)\n", accessor))
								buf.WriteString("            pos += 1\n")
							case "int8":
								buf.WriteString(fmt.Sprintf("            base.storeBytes(of: UInt8(bitPattern: %s), toByteOffset: pos, as: UInt8.self)\n", accessor))
								buf.WriteString("            pos += 1\n")
							case "int16":
								buf.WriteString(fmt.Sprintf("            base.storeBytes(of: %s.littleEndian, toByteOffset: pos, as: Int16.self)\n", accessor))
								buf.WriteString("            pos += 2\n")
							case "int32":
								buf.WriteString(fmt.Sprintf("            base.storeBytes(of: %s.littleEndian, toByteOffset: pos, as: Int32.self)\n", accessor))
								buf.WriteString("            pos += 4\n")
							case "int64":
								buf.WriteString(fmt.Sprintf("            base.storeBytes(of: %s.littleEndian, toByteOffset: pos, as: Int64.self)\n", accessor))
								buf.WriteString("            pos += 8\n")
							case "float32":
								buf.WriteString(fmt.Sprintf("            base.storeBytes(of: %s.bitPattern.littleEndian, toByteOffset: pos, as: UInt32.self)\n", accessor))
								buf.WriteString("            pos += 4\n")
							case "float64":
								buf.WriteString(fmt.Sprintf("            base.storeBytes(of: %s.bitPattern.littleEndian, toByteOffset: pos, as: UInt64.self)\n", accessor))
								buf.WriteString("            pos += 8\n")
							case "string":
								buf.WriteString(fmt.Sprintf("            %s.withUTF8 { utf8 in\n", accessor))
								buf.WriteString("                let strLen = UInt16(utf8.count)\n")
								buf.WriteString("                base.storeBytes(of: strLen.littleEndian, toByteOffset: pos, as: UInt16.self)\n")
								buf.WriteString("                pos += 2\n")
								buf.WriteString("                if let src = utf8.baseAddress {\n")
								buf.WriteString("                    memcpy(base.advanced(by: pos), src, utf8.count)\n")
								buf.WriteString("                }\n")
								buf.WriteString("                pos += utf8.count\n")
								buf.WriteString("            }\n")
							}
						}
					}
					buf.WriteString("        }\n")
					buf.WriteString("    }\n")
					buf.WriteString("    return result\n")
				} else {
					// Check if struct has only primitives/strings with some optionals (no arrays or nested structs)
					hasOnlyPrimitivesAndOptionals := true
					hasOptionals := false
					for _, field := range structType.Fields {
						if primType, ok := field.Type.(*schema.PrimitiveType); ok {
							if primType.Optional {
								hasOptionals = true
							}
						} else {
							hasOnlyPrimitivesAndOptionals = false
						}
					}

					if hasOnlyPrimitivesAndOptionals && hasOptionals {
						// Two-pass optimization for structs with optional fields
						buf.WriteString("    // Two-pass: calculate total size including optionals, then bulk write\n")
						buf.WriteString("    var items = message\n")
						buf.WriteString("    var totalSize = 2 // array length prefix\n")
						buf.WriteString("    for i in 0..<items.count {\n")
						// Calculate size for each field
						for _, field := range structType.Fields {
							if primType, ok := field.Type.(*schema.PrimitiveType); ok {
								accessor := fmt.Sprintf("items[i].%s", field.Name)
								if primType.Optional {
									// Optional field: 1 byte presence + value if present
									switch primType.Name {
									case "bool":
										buf.WriteString(fmt.Sprintf("        totalSize += 1 + (%s != nil ? 1 : 0)\n", accessor))
									case "int8":
										buf.WriteString(fmt.Sprintf("        totalSize += 1 + (%s != nil ? 1 : 0)\n", accessor))
									case "int16":
										buf.WriteString(fmt.Sprintf("        totalSize += 1 + (%s != nil ? 2 : 0)\n", accessor))
									case "int32", "float32":
										buf.WriteString(fmt.Sprintf("        totalSize += 1 + (%s != nil ? 4 : 0)\n", accessor))
									case "int64", "float64":
										buf.WriteString(fmt.Sprintf("        totalSize += 1 + (%s != nil ? 8 : 0)\n", accessor))
									case "string":
										buf.WriteString(fmt.Sprintf("        if var s = %s {\n", accessor))
										buf.WriteString("            s.makeContiguousUTF8()\n")
										buf.WriteString(fmt.Sprintf("            items[i].%s = s\n", field.Name))
										buf.WriteString("            totalSize += 1 + 2 + s.utf8.count\n")
										buf.WriteString("        } else {\n")
										buf.WriteString("            totalSize += 1\n")
										buf.WriteString("        }\n")
									}
								} else {
									// Non-optional field
									switch primType.Name {
									case "bool", "int8":
										buf.WriteString("        totalSize += 1\n")
									case "int16":
										buf.WriteString("        totalSize += 2\n")
									case "int32", "float32":
										buf.WriteString("        totalSize += 4\n")
									case "int64", "float64":
										buf.WriteString("        totalSize += 8\n")
									case "string":
										buf.WriteString(fmt.Sprintf("        items[i].%s.makeContiguousUTF8()\n", field.Name))
										buf.WriteString(fmt.Sprintf("        totalSize += 2 + items[i].%s.utf8.count\n", field.Name))
									}
								}
							}
						}
						buf.WriteString("    }\n")
						buf.WriteString("    var result = Data(count: totalSize)\n")
						buf.WriteString("    result.withUnsafeMutableBytes { ptr in\n")
						buf.WriteString("        let base = ptr.baseAddress!\n")
						buf.WriteString("        base.storeBytes(of: len.littleEndian, as: UInt16.self)\n")
						buf.WriteString("        var pos = 2\n")
						buf.WriteString("        for i in 0..<items.count {\n")
						// Write each field
						for _, field := range structType.Fields {
							if primType, ok := field.Type.(*schema.PrimitiveType); ok {
								accessor := fmt.Sprintf("items[i].%s", field.Name)
								if primType.Optional {
									switch primType.Name {
									case "bool":
										buf.WriteString(fmt.Sprintf("            if let v = %s {\n", accessor))
										buf.WriteString("                base.storeBytes(of: UInt8(1), toByteOffset: pos, as: UInt8.self)\n")
										buf.WriteString("                pos += 1\n")
										buf.WriteString("                base.storeBytes(of: v ? UInt8(1) : UInt8(0), toByteOffset: pos, as: UInt8.self)\n")
										buf.WriteString("                pos += 1\n")
										buf.WriteString("            } else {\n")
										buf.WriteString("                base.storeBytes(of: UInt8(0), toByteOffset: pos, as: UInt8.self)\n")
										buf.WriteString("                pos += 1\n")
										buf.WriteString("            }\n")
									case "int32":
										buf.WriteString(fmt.Sprintf("            if let v = %s {\n", accessor))
										buf.WriteString("                base.storeBytes(of: UInt8(1), toByteOffset: pos, as: UInt8.self)\n")
										buf.WriteString("                pos += 1\n")
										buf.WriteString("                base.storeBytes(of: v.littleEndian, toByteOffset: pos, as: Int32.self)\n")
										buf.WriteString("                pos += 4\n")
										buf.WriteString("            } else {\n")
										buf.WriteString("                base.storeBytes(of: UInt8(0), toByteOffset: pos, as: UInt8.self)\n")
										buf.WriteString("                pos += 1\n")
										buf.WriteString("            }\n")
									case "int64":
										buf.WriteString(fmt.Sprintf("            if let v = %s {\n", accessor))
										buf.WriteString("                base.storeBytes(of: UInt8(1), toByteOffset: pos, as: UInt8.self)\n")
										buf.WriteString("                pos += 1\n")
										buf.WriteString("                base.storeBytes(of: v.littleEndian, toByteOffset: pos, as: Int64.self)\n")
										buf.WriteString("                pos += 8\n")
										buf.WriteString("            } else {\n")
										buf.WriteString("                base.storeBytes(of: UInt8(0), toByteOffset: pos, as: UInt8.self)\n")
										buf.WriteString("                pos += 1\n")
										buf.WriteString("            }\n")
									case "float32":
										buf.WriteString(fmt.Sprintf("            if let v = %s {\n", accessor))
										buf.WriteString("                base.storeBytes(of: UInt8(1), toByteOffset: pos, as: UInt8.self)\n")
										buf.WriteString("                pos += 1\n")
										buf.WriteString("                base.storeBytes(of: v.bitPattern.littleEndian, toByteOffset: pos, as: UInt32.self)\n")
										buf.WriteString("                pos += 4\n")
										buf.WriteString("            } else {\n")
										buf.WriteString("                base.storeBytes(of: UInt8(0), toByteOffset: pos, as: UInt8.self)\n")
										buf.WriteString("                pos += 1\n")
										buf.WriteString("            }\n")
									case "float64":
										buf.WriteString(fmt.Sprintf("            if let v = %s {\n", accessor))
										buf.WriteString("                base.storeBytes(of: UInt8(1), toByteOffset: pos, as: UInt8.self)\n")
										buf.WriteString("                pos += 1\n")
										buf.WriteString("                base.storeBytes(of: v.bitPattern.littleEndian, toByteOffset: pos, as: UInt64.self)\n")
										buf.WriteString("                pos += 8\n")
										buf.WriteString("            } else {\n")
										buf.WriteString("                base.storeBytes(of: UInt8(0), toByteOffset: pos, as: UInt8.self)\n")
										buf.WriteString("                pos += 1\n")
										buf.WriteString("            }\n")
									case "string":
										buf.WriteString(fmt.Sprintf("            if var s = %s {\n", accessor))
										buf.WriteString("                base.storeBytes(of: UInt8(1), toByteOffset: pos, as: UInt8.self)\n")
										buf.WriteString("                pos += 1\n")
										buf.WriteString("                s.withUTF8 { utf8 in\n")
										buf.WriteString("                    base.storeBytes(of: UInt16(utf8.count).littleEndian, toByteOffset: pos, as: UInt16.self)\n")
										buf.WriteString("                    pos += 2\n")
										buf.WriteString("                    if let src = utf8.baseAddress {\n")
										buf.WriteString("                        memcpy(base.advanced(by: pos), src, utf8.count)\n")
										buf.WriteString("                    }\n")
										buf.WriteString("                    pos += utf8.count\n")
										buf.WriteString("                }\n")
										buf.WriteString("            } else {\n")
										buf.WriteString("                base.storeBytes(of: UInt8(0), toByteOffset: pos, as: UInt8.self)\n")
										buf.WriteString("                pos += 1\n")
										buf.WriteString("            }\n")
									}
								} else {
									// Non-optional field
									switch primType.Name {
									case "bool":
										buf.WriteString(fmt.Sprintf("            base.storeBytes(of: %s ? UInt8(1) : UInt8(0), toByteOffset: pos, as: UInt8.self)\n", accessor))
										buf.WriteString("            pos += 1\n")
									case "int8":
										buf.WriteString(fmt.Sprintf("            base.storeBytes(of: UInt8(bitPattern: %s), toByteOffset: pos, as: UInt8.self)\n", accessor))
										buf.WriteString("            pos += 1\n")
									case "int16":
										buf.WriteString(fmt.Sprintf("            base.storeBytes(of: %s.littleEndian, toByteOffset: pos, as: Int16.self)\n", accessor))
										buf.WriteString("            pos += 2\n")
									case "int32":
										buf.WriteString(fmt.Sprintf("            base.storeBytes(of: %s.littleEndian, toByteOffset: pos, as: Int32.self)\n", accessor))
										buf.WriteString("            pos += 4\n")
									case "int64":
										buf.WriteString(fmt.Sprintf("            base.storeBytes(of: %s.littleEndian, toByteOffset: pos, as: Int64.self)\n", accessor))
										buf.WriteString("            pos += 8\n")
									case "float32":
										buf.WriteString(fmt.Sprintf("            base.storeBytes(of: %s.bitPattern.littleEndian, toByteOffset: pos, as: UInt32.self)\n", accessor))
										buf.WriteString("            pos += 4\n")
									case "float64":
										buf.WriteString(fmt.Sprintf("            base.storeBytes(of: %s.bitPattern.littleEndian, toByteOffset: pos, as: UInt64.self)\n", accessor))
										buf.WriteString("            pos += 8\n")
									case "string":
										buf.WriteString(fmt.Sprintf("            %s.withUTF8 { utf8 in\n", accessor))
										buf.WriteString("                base.storeBytes(of: UInt16(utf8.count).littleEndian, toByteOffset: pos, as: UInt16.self)\n")
										buf.WriteString("                pos += 2\n")
										buf.WriteString("                if let src = utf8.baseAddress {\n")
										buf.WriteString("                    memcpy(base.advanced(by: pos), src, utf8.count)\n")
										buf.WriteString("                }\n")
										buf.WriteString("                pos += utf8.count\n")
										buf.WriteString("            }\n")
									}
								}
							}
						}
						buf.WriteString("        }\n")
						buf.WriteString("    }\n")
						buf.WriteString("    return result\n")
					} else {
						// Fallback: structs with arrays or nested structs
						buf.WriteString(fmt.Sprintf("    for item in message { encodeStruct_%s(&buffer, item) }\n", structType.Name))
					}
				}
			}
		}
	}

	buf.WriteString("    return Data(buffer)\n")
	buf.WriteString("}\n\n")
}

func generateSwiftEncodeField(buf *bytes.Buffer, field schema.Field, accessor string) {
	// Check if optional
	isOptional := field.Type.IsOptional()

	// For optional primitives and strings, use dedicated helper functions
	if isOptional {
		switch t := field.Type.(type) {
		case *schema.PrimitiveType:
			switch t.Name {
			case "bool":
				buf.WriteString(fmt.Sprintf("    writeOptionalBool(&buffer, %s)\n", accessor))
				return
			case "int32":
				buf.WriteString(fmt.Sprintf("    writeOptionalInt32(&buffer, %s)\n", accessor))
				return
			case "int64":
				buf.WriteString(fmt.Sprintf("    writeOptionalInt64(&buffer, %s)\n", accessor))
				return
			case "float32":
				buf.WriteString(fmt.Sprintf("    writeOptionalFloat(&buffer, %s)\n", accessor))
				return
			case "float64":
				buf.WriteString(fmt.Sprintf("    writeOptionalDouble(&buffer, %s)\n", accessor))
				return
			case "string":
				buf.WriteString(fmt.Sprintf("    writeOptionalString(&buffer, %s)\n", accessor))
				return
			}
		}
	}

	// Fallback for int8, int16 optionals and other types
	if isOptional {
		buf.WriteString(fmt.Sprintf("    if let unwrapped = %s {\n", accessor))
		buf.WriteString("        buffer.append(1) // present\n")
		accessor = "unwrapped"
	}

	switch t := field.Type.(type) {
	case *schema.PrimitiveType:
		generateSwiftEncodePrimitive(buf, t.Name, accessor)
	case *schema.ArrayType:
		generateSwiftEncodeArray(buf, t, accessor)
	case *schema.StructType:
		buf.WriteString(fmt.Sprintf("        encodeStruct_%s(&buffer, %s)\n", t.Name, accessor))
	}

	if isOptional {
		buf.WriteString("    } else {\n")
		buf.WriteString("        buffer.append(0) // absent\n")
		buf.WriteString("    }\n")
	}
}

func generateSwiftEncodePrimitive(buf *bytes.Buffer, typeName string, accessor string) {
	switch typeName {
	case "bool":
		buf.WriteString(fmt.Sprintf("    buffer.append(%s ? 1 : 0)\n", accessor))
	case "int8":
		buf.WriteString(fmt.Sprintf("    buffer.append(UInt8(bitPattern: %s))\n", accessor))
	case "int16":
		buf.WriteString(fmt.Sprintf("    withUnsafeBytes(of: %s.littleEndian) { buffer.append(contentsOf: $0) }\n", accessor))
	case "int32":
		buf.WriteString(fmt.Sprintf("    withUnsafeBytes(of: %s.littleEndian) { buffer.append(contentsOf: $0) }\n", accessor))
	case "int64":
		buf.WriteString(fmt.Sprintf("    withUnsafeBytes(of: %s.littleEndian) { buffer.append(contentsOf: $0) }\n", accessor))
	case "float32":
		buf.WriteString(fmt.Sprintf("    withUnsafeBytes(of: %s.bitPattern.littleEndian) { buffer.append(contentsOf: $0) }\n", accessor))
	case "float64":
		buf.WriteString(fmt.Sprintf("    withUnsafeBytes(of: %s.bitPattern.littleEndian) { buffer.append(contentsOf: $0) }\n", accessor))
	case "string":
		buf.WriteString(fmt.Sprintf("    encodeString(&buffer, %s)\n", accessor))
	}
}

func generateSwiftEncodeArray(buf *bytes.Buffer, arrayType *schema.ArrayType, accessor string) {
	buf.WriteString(fmt.Sprintf("    let len = UInt16(%s.count)\n", accessor))
	buf.WriteString("    withUnsafeBytes(of: len.littleEndian) { buffer.append(contentsOf: $0) }\n")

	if primType, ok := arrayType.ElementType.(*schema.PrimitiveType); ok {
		switch primType.Name {
		case "bool":
			// Bool arrays need element-by-element conversion
			buf.WriteString(fmt.Sprintf("    for item in %s { buffer.append(item ? 1 : 0) }\n", accessor))
		case "int8":
			// Int8 arrays need bitPattern conversion
			buf.WriteString(fmt.Sprintf("    for item in %s { buffer.append(UInt8(bitPattern: item)) }\n", accessor))
		case "int16":
			// Bulk copy for Int16 arrays (little-endian platforms)
			buf.WriteString(fmt.Sprintf("    %s.withUnsafeBytes { buffer.append(contentsOf: $0) }\n", accessor))
		case "int32":
			// Bulk copy for Int32 arrays (little-endian platforms)
			buf.WriteString(fmt.Sprintf("    %s.withUnsafeBytes { buffer.append(contentsOf: $0) }\n", accessor))
		case "int64":
			// Bulk copy for Int64 arrays (little-endian platforms)
			buf.WriteString(fmt.Sprintf("    %s.withUnsafeBytes { buffer.append(contentsOf: $0) }\n", accessor))
		case "float32":
			// Bulk copy for Float arrays (little-endian platforms, IEEE 754)
			buf.WriteString(fmt.Sprintf("    %s.withUnsafeBytes { buffer.append(contentsOf: $0) }\n", accessor))
		case "float64":
			// Bulk copy for Double arrays (little-endian platforms, IEEE 754)
			buf.WriteString(fmt.Sprintf("    %s.withUnsafeBytes { buffer.append(contentsOf: $0) }\n", accessor))
		case "string":
			buf.WriteString(fmt.Sprintf("    for item in %s { encodeString(&buffer, item) }\n", accessor))
		}
	} else if structType, ok := arrayType.ElementType.(*schema.StructType); ok {
		buf.WriteString(fmt.Sprintf("    for item in %s { encodeStruct_%s(&buffer, item) }\n", accessor, structType.Name))
	}
}

func generateSwiftDecoderFunc(buf *bytes.Buffer, msg schema.MessageType) {
	structName := msg.Name + "Message"
	funcName := fmt.Sprintf("decode%sMessage", msg.Name)

	buf.WriteString("@inlinable\n")
	buf.WriteString(fmt.Sprintf("public func %s(_ data: Data) throws -> %s {\n", funcName, structName))
	buf.WriteString("    return try data.withUnsafeBytes { (ptr: UnsafeRawBufferPointer) in\n")
	buf.WriteString("        guard let base = ptr.baseAddress else { throw FFireError.invalidData }\n")
	buf.WriteString("        var pos = 0\n")

	switch t := msg.TargetType.(type) {
	case *schema.StructType:
		for _, field := range t.Fields {
			generateSwiftDecodeField(buf, field)
		}

		buf.WriteString(fmt.Sprintf("        return %s(\n", structName))
		for i, field := range t.Fields {
			buf.WriteString(fmt.Sprintf("            %s: %s", field.Name, field.Name))
			if i < len(t.Fields)-1 {
				buf.WriteString(",\n")
			} else {
				buf.WriteString("\n")
			}
		}
		buf.WriteString("        )\n")
	case *schema.ArrayType:
		// Decode array
		buf.WriteString("        let len = Int(UInt16(littleEndian: base.load(fromByteOffset: pos, as: UInt16.self)))\n")
		buf.WriteString("        pos += 2\n")
		if primType, ok := t.ElementType.(*schema.PrimitiveType); ok {
			switch primType.Name {
			case "bool":
				// Bool arrays need element-by-element conversion
				buf.WriteString("        return (0..<len).map { _ in\n")
				buf.WriteString("            let v = base.load(fromByteOffset: pos, as: UInt8.self) != 0\n")
				buf.WriteString("            pos += 1\n")
				buf.WriteString("            return v\n")
				buf.WriteString("        }\n")
			case "int8":
				// Int8 arrays need element-by-element read
				buf.WriteString("        return (0..<len).map { _ in\n")
				buf.WriteString("            let v = base.load(fromByteOffset: pos, as: Int8.self)\n")
				buf.WriteString("            pos += 1\n")
				buf.WriteString("            return v\n")
				buf.WriteString("        }\n")
			case "int16":
				// Bulk copy for Int16 arrays (little-endian platforms)
				buf.WriteString("        let byteCount = len * MemoryLayout<Int16>.stride\n")
				buf.WriteString("        let result = [Int16](unsafeUninitializedCapacity: len) { buffer, initializedCount in\n")
				buf.WriteString("            let src = UnsafeRawPointer(base.advanced(by: pos))\n")
				buf.WriteString("            let dst = UnsafeMutableRawPointer(buffer.baseAddress!)\n")
				buf.WriteString("            dst.copyMemory(from: src, byteCount: byteCount)\n")
				buf.WriteString("            initializedCount = len\n")
				buf.WriteString("        }\n")
				buf.WriteString("        pos += byteCount\n")
				buf.WriteString("        return result\n")
			case "int32":
				// Bulk copy for Int32 arrays (little-endian platforms)
				buf.WriteString("        let byteCount = len * MemoryLayout<Int32>.stride\n")
				buf.WriteString("        let result = [Int32](unsafeUninitializedCapacity: len) { buffer, initializedCount in\n")
				buf.WriteString("            let src = UnsafeRawPointer(base.advanced(by: pos))\n")
				buf.WriteString("            let dst = UnsafeMutableRawPointer(buffer.baseAddress!)\n")
				buf.WriteString("            dst.copyMemory(from: src, byteCount: byteCount)\n")
				buf.WriteString("            initializedCount = len\n")
				buf.WriteString("        }\n")
				buf.WriteString("        pos += byteCount\n")
				buf.WriteString("        return result\n")
			case "int64":
				// Bulk copy for Int64 arrays (little-endian platforms)
				buf.WriteString("        let byteCount = len * MemoryLayout<Int64>.stride\n")
				buf.WriteString("        let result = [Int64](unsafeUninitializedCapacity: len) { buffer, initializedCount in\n")
				buf.WriteString("            let src = UnsafeRawPointer(base.advanced(by: pos))\n")
				buf.WriteString("            let dst = UnsafeMutableRawPointer(buffer.baseAddress!)\n")
				buf.WriteString("            dst.copyMemory(from: src, byteCount: byteCount)\n")
				buf.WriteString("            initializedCount = len\n")
				buf.WriteString("        }\n")
				buf.WriteString("        pos += byteCount\n")
				buf.WriteString("        return result\n")
			case "float32":
				// Bulk copy for Float arrays (little-endian platforms, IEEE 754)
				buf.WriteString("        let byteCount = len * MemoryLayout<Float>.stride\n")
				buf.WriteString("        let result = [Float](unsafeUninitializedCapacity: len) { buffer, initializedCount in\n")
				buf.WriteString("            let src = UnsafeRawPointer(base.advanced(by: pos))\n")
				buf.WriteString("            let dst = UnsafeMutableRawPointer(buffer.baseAddress!)\n")
				buf.WriteString("            dst.copyMemory(from: src, byteCount: byteCount)\n")
				buf.WriteString("            initializedCount = len\n")
				buf.WriteString("        }\n")
				buf.WriteString("        pos += byteCount\n")
				buf.WriteString("        return result\n")
			case "float64":
				// Bulk copy for Double arrays (little-endian platforms, IEEE 754)
				buf.WriteString("        let byteCount = len * MemoryLayout<Double>.stride\n")
				buf.WriteString("        let result = [Double](unsafeUninitializedCapacity: len) { buffer, initializedCount in\n")
				buf.WriteString("            let src = UnsafeRawPointer(base.advanced(by: pos))\n")
				buf.WriteString("            let dst = UnsafeMutableRawPointer(buffer.baseAddress!)\n")
				buf.WriteString("            dst.copyMemory(from: src, byteCount: byteCount)\n")
				buf.WriteString("            initializedCount = len\n")
				buf.WriteString("        }\n")
				buf.WriteString("        pos += byteCount\n")
				buf.WriteString("        return result\n")
			case "string":
				// Optimized string array decoding - inline string creation, pre-allocate array
				buf.WriteString("        var result = [String]()\n")
				buf.WriteString("        result.reserveCapacity(len)\n")
				buf.WriteString("        for _ in 0..<len {\n")
				buf.WriteString("            let strLen = Int(UInt16(littleEndian: base.load(fromByteOffset: pos, as: UInt16.self)))\n")
				buf.WriteString("            pos += 2\n")
				buf.WriteString("            let str = String(decoding: UnsafeBufferPointer(start: base.advanced(by: pos).assumingMemoryBound(to: UInt8.self), count: strLen), as: UTF8.self)\n")
				buf.WriteString("            result.append(str)\n")
				buf.WriteString("            pos += strLen\n")
				buf.WriteString("        }\n")
				buf.WriteString("        return result\n")
			}
		} else if structType, ok := t.ElementType.(*schema.StructType); ok {
			buf.WriteString(fmt.Sprintf("        return try (0..<len).map { _ in try decodeStruct_%s(base, &pos) }\n", structType.Name))
		}
	}

	buf.WriteString("    }\n")
	buf.WriteString("}\n\n")
}

func generateSwiftDecodeField(buf *bytes.Buffer, field schema.Field) {
	varName := field.Name
	isOptional := field.Type.IsOptional()

	// For optional primitives and strings, use dedicated helper functions
	if isOptional {
		switch t := field.Type.(type) {
		case *schema.PrimitiveType:
			switch t.Name {
			case "bool":
				buf.WriteString(fmt.Sprintf("        let %s = readOptionalBool(base, &pos)\n", varName))
			case "int32":
				buf.WriteString(fmt.Sprintf("        let %s = readOptionalInt32(base, &pos)\n", varName))
			case "int64":
				buf.WriteString(fmt.Sprintf("        let %s = readOptionalInt64(base, &pos)\n", varName))
			case "float32":
				buf.WriteString(fmt.Sprintf("        let %s = readOptionalFloat(base, &pos)\n", varName))
			case "float64":
				buf.WriteString(fmt.Sprintf("        let %s = readOptionalDouble(base, &pos)\n", varName))
			case "string":
				buf.WriteString(fmt.Sprintf("        let %s = readOptionalString(base, &pos)\n", varName))
			default:
				// Fallback for int8, int16 - use branching approach
				generateSwiftDecodeOptionalFallback(buf, field)
			}
			return
		}
	}

	// Non-optional primitives and strings, or types that need branching
	if isOptional {
		buf.WriteString(fmt.Sprintf("        let %sPresent = base.load(fromByteOffset: pos, as: UInt8.self) != 0\n", varName))
		buf.WriteString("        pos += 1\n")
		buf.WriteString(fmt.Sprintf("        let %s: %s\n", varName, getSwiftTypeString(field.Type)))
		buf.WriteString(fmt.Sprintf("        if %sPresent {\n", varName))
	}

	switch t := field.Type.(type) {
	case *schema.PrimitiveType:
		generateSwiftDecodePrimitive(buf, t.Name, varName)
	case *schema.ArrayType:
		if isOptional {
			generateSwiftDecodeArray(buf, t, varName+"Value")
			buf.WriteString(fmt.Sprintf("            %s = %sValue\n", varName, varName))
		} else {
			generateSwiftDecodeArray(buf, t, varName)
		}
	case *schema.StructType:
		if isOptional {
			buf.WriteString(fmt.Sprintf("            let %sValue = try decodeStruct_%s(base, &pos)\n", varName, t.Name))
			buf.WriteString(fmt.Sprintf("            %s = %sValue\n", varName, varName))
		} else {
			buf.WriteString(fmt.Sprintf("        let %s = try decodeStruct_%s(base, &pos)\n", varName, t.Name))
		}
	}

	if isOptional {
		buf.WriteString("        } else {\n")
		buf.WriteString(fmt.Sprintf("            %s = nil\n", varName))
		buf.WriteString("        }\n")
	}
}

func generateSwiftDecodeOptionalFallback(buf *bytes.Buffer, field schema.Field) {
	varName := field.Name
	buf.WriteString(fmt.Sprintf("        let %sPresent = base.load(fromByteOffset: pos, as: UInt8.self) != 0\n", varName))
	buf.WriteString("        pos += 1\n")
	buf.WriteString(fmt.Sprintf("        let %s: %s\n", varName, getSwiftTypeString(field.Type)))
	buf.WriteString(fmt.Sprintf("        if %sPresent {\n", varName))

	if t, ok := field.Type.(*schema.PrimitiveType); ok {
		generateSwiftDecodePrimitive(buf, t.Name, varName+"Value")
		buf.WriteString(fmt.Sprintf("            %s = %sValue\n", varName, varName))
	}

	buf.WriteString("        } else {\n")
	buf.WriteString(fmt.Sprintf("            %s = nil\n", varName))
	buf.WriteString("        }\n")
}

func generateSwiftDecodePrimitive(buf *bytes.Buffer, typeName string, varName string) {
	switch typeName {
	case "bool":
		buf.WriteString(fmt.Sprintf("        let %s = readBool(base, &pos)\n", varName))
	case "int8":
		buf.WriteString(fmt.Sprintf("        let %s = readInt8(base, &pos)\n", varName))
	case "int16":
		buf.WriteString(fmt.Sprintf("        let %s = readInt16(base, &pos)\n", varName))
	case "int32":
		buf.WriteString(fmt.Sprintf("        let %s = readInt32(base, &pos)\n", varName))
	case "int64":
		buf.WriteString(fmt.Sprintf("        let %s = readInt64(base, &pos)\n", varName))
	case "float32":
		buf.WriteString(fmt.Sprintf("        let %s = readFloat(base, &pos)\n", varName))
	case "float64":
		buf.WriteString(fmt.Sprintf("        let %s = readDouble(base, &pos)\n", varName))
	case "string":
		buf.WriteString(fmt.Sprintf("        let %s = try decodeString(base, &pos)\n", varName))
	}
}

func generateSwiftDecodeArray(buf *bytes.Buffer, arrayType *schema.ArrayType, varName string) {
	elemSwiftType := getSwiftTypeString(arrayType.ElementType)
	buf.WriteString(fmt.Sprintf("        let %sLen = Int(UInt16(littleEndian: base.load(fromByteOffset: pos, as: UInt16.self)))\n", varName))
	buf.WriteString("        pos += 2\n")

	if primType, ok := arrayType.ElementType.(*schema.PrimitiveType); ok {
		switch primType.Name {
		case "bool":
			// Bool arrays need element-by-element conversion (UInt8 to Bool)
			buf.WriteString(fmt.Sprintf("        let %s: [Bool] = (0..<%sLen).map { _ in\n", varName, varName))
			buf.WriteString("            let v = base.load(fromByteOffset: pos, as: UInt8.self) != 0\n")
			buf.WriteString("            pos += 1\n")
			buf.WriteString("            return v\n")
			buf.WriteString("        }\n")
		case "int8":
			// Int8 arrays need element-by-element read
			buf.WriteString(fmt.Sprintf("        let %s: [Int8] = (0..<%sLen).map { _ in\n", varName, varName))
			buf.WriteString("            let v = base.load(fromByteOffset: pos, as: Int8.self)\n")
			buf.WriteString("            pos += 1\n")
			buf.WriteString("            return v\n")
			buf.WriteString("        }\n")
		case "int16":
			// Bulk copy for Int16 arrays (little-endian platforms)
			buf.WriteString(fmt.Sprintf("        let %sByteCount = %sLen * MemoryLayout<Int16>.stride\n", varName, varName))
			buf.WriteString(fmt.Sprintf("        let %s = [Int16](unsafeUninitializedCapacity: %sLen) { buffer, initializedCount in\n", varName, varName))
			buf.WriteString("            let src = UnsafeRawPointer(base.advanced(by: pos))\n")
			buf.WriteString("            let dst = UnsafeMutableRawPointer(buffer.baseAddress!)\n")
			buf.WriteString(fmt.Sprintf("            dst.copyMemory(from: src, byteCount: %sByteCount)\n", varName))
			buf.WriteString(fmt.Sprintf("            initializedCount = %sLen\n", varName))
			buf.WriteString("        }\n")
			buf.WriteString(fmt.Sprintf("        pos += %sByteCount\n", varName))
		case "int32":
			// Bulk copy for Int32 arrays (little-endian platforms)
			buf.WriteString(fmt.Sprintf("        let %sByteCount = %sLen * MemoryLayout<Int32>.stride\n", varName, varName))
			buf.WriteString(fmt.Sprintf("        let %s = [Int32](unsafeUninitializedCapacity: %sLen) { buffer, initializedCount in\n", varName, varName))
			buf.WriteString("            let src = UnsafeRawPointer(base.advanced(by: pos))\n")
			buf.WriteString("            let dst = UnsafeMutableRawPointer(buffer.baseAddress!)\n")
			buf.WriteString(fmt.Sprintf("            dst.copyMemory(from: src, byteCount: %sByteCount)\n", varName))
			buf.WriteString(fmt.Sprintf("            initializedCount = %sLen\n", varName))
			buf.WriteString("        }\n")
			buf.WriteString(fmt.Sprintf("        pos += %sByteCount\n", varName))
		case "int64":
			// Bulk copy for Int64 arrays (little-endian platforms)
			buf.WriteString(fmt.Sprintf("        let %sByteCount = %sLen * MemoryLayout<Int64>.stride\n", varName, varName))
			buf.WriteString(fmt.Sprintf("        let %s = [Int64](unsafeUninitializedCapacity: %sLen) { buffer, initializedCount in\n", varName, varName))
			buf.WriteString("            let src = UnsafeRawPointer(base.advanced(by: pos))\n")
			buf.WriteString("            let dst = UnsafeMutableRawPointer(buffer.baseAddress!)\n")
			buf.WriteString(fmt.Sprintf("            dst.copyMemory(from: src, byteCount: %sByteCount)\n", varName))
			buf.WriteString(fmt.Sprintf("            initializedCount = %sLen\n", varName))
			buf.WriteString("        }\n")
			buf.WriteString(fmt.Sprintf("        pos += %sByteCount\n", varName))
		case "float32":
			// Bulk copy for Float arrays (little-endian platforms, IEEE 754)
			buf.WriteString(fmt.Sprintf("        let %sByteCount = %sLen * MemoryLayout<Float>.stride\n", varName, varName))
			buf.WriteString(fmt.Sprintf("        let %s = [Float](unsafeUninitializedCapacity: %sLen) { buffer, initializedCount in\n", varName, varName))
			buf.WriteString("            let src = UnsafeRawPointer(base.advanced(by: pos))\n")
			buf.WriteString("            let dst = UnsafeMutableRawPointer(buffer.baseAddress!)\n")
			buf.WriteString(fmt.Sprintf("            dst.copyMemory(from: src, byteCount: %sByteCount)\n", varName))
			buf.WriteString(fmt.Sprintf("            initializedCount = %sLen\n", varName))
			buf.WriteString("        }\n")
			buf.WriteString(fmt.Sprintf("        pos += %sByteCount\n", varName))
		case "float64":
			// Bulk copy for Double arrays (little-endian platforms, IEEE 754)
			buf.WriteString(fmt.Sprintf("        let %sByteCount = %sLen * MemoryLayout<Double>.stride\n", varName, varName))
			buf.WriteString(fmt.Sprintf("        let %s = [Double](unsafeUninitializedCapacity: %sLen) { buffer, initializedCount in\n", varName, varName))
			buf.WriteString("            let src = UnsafeRawPointer(base.advanced(by: pos))\n")
			buf.WriteString("            let dst = UnsafeMutableRawPointer(buffer.baseAddress!)\n")
			buf.WriteString(fmt.Sprintf("            dst.copyMemory(from: src, byteCount: %sByteCount)\n", varName))
			buf.WriteString(fmt.Sprintf("            initializedCount = %sLen\n", varName))
			buf.WriteString("        }\n")
			buf.WriteString(fmt.Sprintf("        pos += %sByteCount\n", varName))
		case "string":
			// Optimized string array decoding - inline string creation, pre-allocate array
			buf.WriteString(fmt.Sprintf("        var %s = [String]()\n", varName))
			buf.WriteString(fmt.Sprintf("        %s.reserveCapacity(%sLen)\n", varName, varName))
			buf.WriteString(fmt.Sprintf("        for _ in 0..<%sLen {\n", varName))
			buf.WriteString("            let strLen = Int(UInt16(littleEndian: base.load(fromByteOffset: pos, as: UInt16.self)))\n")
			buf.WriteString("            pos += 2\n")
			buf.WriteString("            let str = String(decoding: UnsafeBufferPointer(start: base.advanced(by: pos).assumingMemoryBound(to: UInt8.self), count: strLen), as: UTF8.self)\n")
			buf.WriteString(fmt.Sprintf("            %s.append(str)\n", varName))
			buf.WriteString("            pos += strLen\n")
			buf.WriteString("        }\n")
		}
	} else if structType, ok := arrayType.ElementType.(*schema.StructType); ok {
		buf.WriteString(fmt.Sprintf("        let %s: [%s] = try (0..<%sLen).map { _ in try decodeStruct_%s(base, &pos) }\n", 
			varName, elemSwiftType, varName, structType.Name))
	}
}

func generateSwiftStructHelpers(buf *bytes.Buffer, structType *schema.StructType) {
	// Encode helper
	buf.WriteString("@inlinable\n")
	buf.WriteString(fmt.Sprintf("func encodeStruct_%s(_ buffer: inout [UInt8], _ value: %s) {\n", structType.Name, structType.Name))
	for _, field := range structType.Fields {
		generateSwiftEncodeField(buf, field, "value."+field.Name)
	}
	buf.WriteString("}\n\n")

	// Decode helper
	buf.WriteString("@inlinable\n")
	buf.WriteString(fmt.Sprintf("func decodeStruct_%s(_ base: UnsafeRawPointer, _ pos: inout Int) throws -> %s {\n", structType.Name, structType.Name))
	for _, field := range structType.Fields {
		generateSwiftDecodeField(buf, field)
	}
	buf.WriteString(fmt.Sprintf("    return %s(\n", structType.Name))
	for i, field := range structType.Fields {
		buf.WriteString(fmt.Sprintf("        %s: %s", field.Name, field.Name))
		if i < len(structType.Fields)-1 {
			buf.WriteString(",\n")
		} else {
			buf.WriteString("\n")
		}
	}
	buf.WriteString("    )\n")
	buf.WriteString("}\n\n")
}

func generateSwiftHelpers(buf *bytes.Buffer) {
	buf.WriteString("// MARK: - Helper Functions\n\n")
	
	buf.WriteString("public enum FFireError: Error {\n")
	buf.WriteString("    case invalidData\n")
	buf.WriteString("    case invalidString\n")
	buf.WriteString("}\n\n")

	// Add inline helper functions for primitive reads
	buf.WriteString("@inlinable\n")
	buf.WriteString("func readInt16(_ base: UnsafeRawPointer, _ pos: inout Int) -> Int16 {\n")
	buf.WriteString("    defer { pos += 2 }\n")
	buf.WriteString("    return Int16(littleEndian: base.load(fromByteOffset: pos, as: Int16.self))\n")
	buf.WriteString("}\n\n")

	buf.WriteString("@inlinable\n")
	buf.WriteString("func readInt32(_ base: UnsafeRawPointer, _ pos: inout Int) -> Int32 {\n")
	buf.WriteString("    defer { pos += 4 }\n")
	buf.WriteString("    return Int32(littleEndian: base.load(fromByteOffset: pos, as: Int32.self))\n")
	buf.WriteString("}\n\n")

	buf.WriteString("@inlinable\n")
	buf.WriteString("func readInt64(_ base: UnsafeRawPointer, _ pos: inout Int) -> Int64 {\n")
	buf.WriteString("    defer { pos += 8 }\n")
	buf.WriteString("    return Int64(littleEndian: base.load(fromByteOffset: pos, as: Int64.self))\n")
	buf.WriteString("}\n\n")

	buf.WriteString("@inlinable\n")
	buf.WriteString("func readFloat(_ base: UnsafeRawPointer, _ pos: inout Int) -> Float {\n")
	buf.WriteString("    defer { pos += 4 }\n")
	buf.WriteString("    return Float(bitPattern: UInt32(littleEndian: base.load(fromByteOffset: pos, as: UInt32.self)))\n")
	buf.WriteString("}\n\n")

	buf.WriteString("@inlinable\n")
	buf.WriteString("func readDouble(_ base: UnsafeRawPointer, _ pos: inout Int) -> Double {\n")
	buf.WriteString("    defer { pos += 8 }\n")
	buf.WriteString("    return Double(bitPattern: UInt64(littleEndian: base.load(fromByteOffset: pos, as: UInt64.self)))\n")
	buf.WriteString("}\n\n")

	buf.WriteString("@inlinable\n")
	buf.WriteString("func readBool(_ base: UnsafeRawPointer, _ pos: inout Int) -> Bool {\n")
	buf.WriteString("    defer { pos += 1 }\n")
	buf.WriteString("    return base.load(fromByteOffset: pos, as: UInt8.self) != 0\n")
	buf.WriteString("}\n\n")

	buf.WriteString("@inlinable\n")
	buf.WriteString("func readInt8(_ base: UnsafeRawPointer, _ pos: inout Int) -> Int8 {\n")
	buf.WriteString("    defer { pos += 1 }\n")
	buf.WriteString("    return base.load(fromByteOffset: pos, as: Int8.self)\n")
	buf.WriteString("}\n\n")

	// Optional primitive readers - combine presence check + value read
	buf.WriteString("@inlinable\n")
	buf.WriteString("func readOptionalInt32(_ base: UnsafeRawPointer, _ pos: inout Int) -> Int32? {\n")
	buf.WriteString("    let present = base.load(fromByteOffset: pos, as: UInt8.self)\n")
	buf.WriteString("    pos += 1\n")
	buf.WriteString("    guard present != 0 else { return nil }\n")
	buf.WriteString("    defer { pos += 4 }\n")
	buf.WriteString("    return Int32(littleEndian: base.load(fromByteOffset: pos, as: Int32.self))\n")
	buf.WriteString("}\n\n")

	buf.WriteString("@inlinable\n")
	buf.WriteString("func readOptionalInt64(_ base: UnsafeRawPointer, _ pos: inout Int) -> Int64? {\n")
	buf.WriteString("    let present = base.load(fromByteOffset: pos, as: UInt8.self)\n")
	buf.WriteString("    pos += 1\n")
	buf.WriteString("    guard present != 0 else { return nil }\n")
	buf.WriteString("    defer { pos += 8 }\n")
	buf.WriteString("    return Int64(littleEndian: base.load(fromByteOffset: pos, as: Int64.self))\n")
	buf.WriteString("}\n\n")

	buf.WriteString("@inlinable\n")
	buf.WriteString("func readOptionalFloat(_ base: UnsafeRawPointer, _ pos: inout Int) -> Float? {\n")
	buf.WriteString("    let present = base.load(fromByteOffset: pos, as: UInt8.self)\n")
	buf.WriteString("    pos += 1\n")
	buf.WriteString("    guard present != 0 else { return nil }\n")
	buf.WriteString("    defer { pos += 4 }\n")
	buf.WriteString("    return Float(bitPattern: UInt32(littleEndian: base.load(fromByteOffset: pos, as: UInt32.self)))\n")
	buf.WriteString("}\n\n")

	buf.WriteString("@inlinable\n")
	buf.WriteString("func readOptionalDouble(_ base: UnsafeRawPointer, _ pos: inout Int) -> Double? {\n")
	buf.WriteString("    let present = base.load(fromByteOffset: pos, as: UInt8.self)\n")
	buf.WriteString("    pos += 1\n")
	buf.WriteString("    guard present != 0 else { return nil }\n")
	buf.WriteString("    defer { pos += 8 }\n")
	buf.WriteString("    return Double(bitPattern: UInt64(littleEndian: base.load(fromByteOffset: pos, as: UInt64.self)))\n")
	buf.WriteString("}\n\n")

	buf.WriteString("@inlinable\n")
	buf.WriteString("func readOptionalBool(_ base: UnsafeRawPointer, _ pos: inout Int) -> Bool? {\n")
	buf.WriteString("    let present = base.load(fromByteOffset: pos, as: UInt8.self)\n")
	buf.WriteString("    pos += 1\n")
	buf.WriteString("    guard present != 0 else { return nil }\n")
	buf.WriteString("    defer { pos += 1 }\n")
	buf.WriteString("    return base.load(fromByteOffset: pos, as: UInt8.self) != 0\n")
	buf.WriteString("}\n\n")

	buf.WriteString("@inlinable\n")
	buf.WriteString("func readOptionalString(_ base: UnsafeRawPointer, _ pos: inout Int) -> String? {\n")
	buf.WriteString("    let present = base.load(fromByteOffset: pos, as: UInt8.self)\n")
	buf.WriteString("    pos += 1\n")
	buf.WriteString("    guard present != 0 else { return nil }\n")
	buf.WriteString("    let len = Int(UInt16(littleEndian: base.load(fromByteOffset: pos, as: UInt16.self)))\n")
	buf.WriteString("    pos += 2\n")
	buf.WriteString("    let result = String(decoding: UnsafeBufferPointer(start: base.advanced(by: pos).assumingMemoryBound(to: UInt8.self), count: len), as: UTF8.self)\n")
	buf.WriteString("    pos += len\n")
	buf.WriteString("    return result\n")
	buf.WriteString("}\n\n")

	// Optional primitive writers - combine presence byte + value in single call
	buf.WriteString("@inlinable\n")
	buf.WriteString("func writeOptionalInt32(_ buffer: inout [UInt8], _ value: Int32?) {\n")
	buf.WriteString("    guard let v = value else { buffer.append(0); return }\n")
	buf.WriteString("    buffer.append(1)\n")
	buf.WriteString("    withUnsafeBytes(of: v.littleEndian) { buffer.append(contentsOf: $0) }\n")
	buf.WriteString("}\n\n")

	buf.WriteString("@inlinable\n")
	buf.WriteString("func writeOptionalInt64(_ buffer: inout [UInt8], _ value: Int64?) {\n")
	buf.WriteString("    guard let v = value else { buffer.append(0); return }\n")
	buf.WriteString("    buffer.append(1)\n")
	buf.WriteString("    withUnsafeBytes(of: v.littleEndian) { buffer.append(contentsOf: $0) }\n")
	buf.WriteString("}\n\n")

	buf.WriteString("@inlinable\n")
	buf.WriteString("func writeOptionalFloat(_ buffer: inout [UInt8], _ value: Float?) {\n")
	buf.WriteString("    guard let v = value else { buffer.append(0); return }\n")
	buf.WriteString("    buffer.append(1)\n")
	buf.WriteString("    withUnsafeBytes(of: v.bitPattern.littleEndian) { buffer.append(contentsOf: $0) }\n")
	buf.WriteString("}\n\n")

	buf.WriteString("@inlinable\n")
	buf.WriteString("func writeOptionalDouble(_ buffer: inout [UInt8], _ value: Double?) {\n")
	buf.WriteString("    guard let v = value else { buffer.append(0); return }\n")
	buf.WriteString("    buffer.append(1)\n")
	buf.WriteString("    withUnsafeBytes(of: v.bitPattern.littleEndian) { buffer.append(contentsOf: $0) }\n")
	buf.WriteString("}\n\n")

	buf.WriteString("@inlinable\n")
	buf.WriteString("func writeOptionalBool(_ buffer: inout [UInt8], _ value: Bool?) {\n")
	buf.WriteString("    guard let v = value else { buffer.append(0); return }\n")
	buf.WriteString("    buffer.append(1)\n")
	buf.WriteString("    buffer.append(v ? 1 : 0)\n")
	buf.WriteString("}\n\n")

	buf.WriteString("@inlinable\n")
	buf.WriteString("func writeOptionalString(_ buffer: inout [UInt8], _ value: String?) {\n")
	buf.WriteString("    guard let v = value else { buffer.append(0); return }\n")
	buf.WriteString("    buffer.append(1)\n")
	buf.WriteString("    // Reuse encodeString for consistency\n")
	buf.WriteString("    encodeString(&buffer, v)\n")
	buf.WriteString("}\n\n")

	// String encoding - use append for [UInt8] (very fast compared to Data.append)
	buf.WriteString("@inlinable\n")
	buf.WriteString("func encodeString(_ buffer: inout [UInt8], _ string: String) {\n")
	buf.WriteString("    var s = string\n")
	buf.WriteString("    s.withUTF8 { utf8 in\n")
	buf.WriteString("        let len = UInt16(utf8.count)\n")
	buf.WriteString("        buffer.append(UInt8(len & 0xFF))\n")
	buf.WriteString("        buffer.append(UInt8(len >> 8))\n")
	buf.WriteString("        if let base = utf8.baseAddress {\n")
	buf.WriteString("            buffer.append(contentsOf: UnsafeBufferPointer(start: base, count: utf8.count))\n")
	buf.WriteString("        }\n")
	buf.WriteString("    }\n")
	buf.WriteString("}\n\n")

	buf.WriteString("@inlinable\n")
	buf.WriteString("func decodeString(_ base: UnsafeRawPointer, _ pos: inout Int) throws -> String {\n")
	buf.WriteString("    let len = Int(UInt16(littleEndian: base.load(fromByteOffset: pos, as: UInt16.self)))\n")
	buf.WriteString("    pos += 2\n")
	buf.WriteString("    // Use unsafe decoding - assumes valid UTF-8 (ffire guarantees this)\n")
	buf.WriteString("    let result = String(decoding: UnsafeBufferPointer(start: base.advanced(by: pos).assumingMemoryBound(to: UInt8.self), count: len), as: UTF8.self)\n")
	buf.WriteString("    pos += len\n")
	buf.WriteString("    return result\n")
	buf.WriteString("}\n")
}

func getSwiftTypeString(typ schema.Type) string {
	switch t := typ.(type) {
	case *schema.PrimitiveType:
		baseType := getSwiftPrimitiveType(t.Name)
		if t.Optional {
			return baseType + "?"
		}
		return baseType
	case *schema.ArrayType:
		elemType := getSwiftTypeString(t.ElementType)
		arrayType := fmt.Sprintf("[%s]", elemType)
		if t.Optional {
			return arrayType + "?"
		}
		return arrayType
	case *schema.StructType:
		if t.Optional {
			return t.Name + "?"
		}
		return t.Name
	default:
		return "Any"
	}
}

func getSwiftPrimitiveType(name string) string {
	switch name {
	case "bool":
		return "Bool"
	case "int8":
		return "Int8"
	case "int16":
		return "Int16"
	case "int32":
		return "Int32"
	case "int64":
		return "Int64"
	case "float32":
		return "Float"
	case "float64":
		return "Double"
	case "string":
		return "String"
	default:
		return "Any"
	}
}

// Swift keywords that need escaping when used as field names
var swiftFieldKeywords = map[string]bool{
	"Type": true, "Self": true, "self": true, "Protocol": true,
	"class": true, "struct": true, "enum": true, "protocol": true,
	"extension": true, "func": true, "var": true, "let": true,
	"init": true, "deinit": true, "subscript": true, "typealias": true,
	"operator": true, "precedencegroup": true, "associatedtype": true,
	"import": true, "static": true, "public": true, "private": true,
	"fileprivate": true, "internal": true, "open": true,
}

func escapeSwiftFieldName(name string) string {
	if swiftFieldKeywords[name] {
		return "`" + name + "`"
	}
	return name
}

func generateSwiftMetadataOrchestrated(config *PackageConfig, paths *PackagePaths) error {
	// Generate Package.swift
	if err := generateSwiftPackageManifest(config, paths.Root); err != nil {
		return err
	}

	// Generate README.md
	if err := generateSwiftReadme(config, paths.Root); err != nil {
		return err
	}

	return nil
}

func printSwiftInstructions(config *PackageConfig, paths *PackagePaths) {
	fmt.Printf("\n✅ Native Swift package ready at: %s\n\n", paths.Root)
	fmt.Println("Build:")
	fmt.Printf("  cd %s\n", paths.Root)
	fmt.Println("  swift build")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  import %s\n", config.Namespace)
	fmt.Printf("  let msg = PluginMessage(name: \"test\", version: \"1.0\")\n")
	fmt.Printf("  let encoded = encodePluginMessage(msg)\n")
	fmt.Printf("  let decoded = try decodePluginMessage(encoded)\n")
	fmt.Println()
}

// generateSwiftPackageManifest generates Package.swift for native Swift
func generateSwiftPackageManifest(config *PackageConfig, packageDir string) error {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, `// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "%s",
    platforms: [
        .macOS(.v13),
        .iOS(.v16),
        .tvOS(.v16),
        .watchOS(.v9)
    ],
    products: [
        .library(
            name: "%s",
            targets: ["%s"]
        ),
    ],
    dependencies: [],
    targets: [
        .target(
            name: "%s",
            dependencies: [],
            path: "Sources/%s"
        ),
    ]
)
`, config.Namespace, config.Namespace, config.Namespace, config.Namespace, config.Namespace)

	manifestPath := filepath.Join(packageDir, "Package.swift")
	if err := os.WriteFile(manifestPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write Package.swift: %w", err)
	}

	fmt.Printf("✓ Generated Package.swift: %s\n", manifestPath)
	return nil
}

// generateSwiftReadme generates README.md
func generateSwiftReadme(config *PackageConfig, packageDir string) error {
	buf := &bytes.Buffer{}

	packageName := config.Schema.Package

	fmt.Fprintf(buf, "# %s - FFire Native Swift Package\n\n", config.Namespace)
	fmt.Fprintf(buf, "Native Swift implementation for the %s schema.\n", packageName)
	buf.WriteString("Generated by [FFire](https://github.com/shaban/ffire).\n\n")

	buf.WriteString("## Overview\n\n")
	buf.WriteString("This package provides a **native Swift implementation** optimized for maximum performance:\n\n")
	buf.WriteString("- Direct memory access using unsafe pointers\n")
	buf.WriteString("- Zero-copy operations for primitive arrays\n")
	buf.WriteString("- Inlinable functions for compiler optimization\n")
	buf.WriteString("- Value types (structs) for efficient memory usage\n")
	buf.WriteString("- ~10% faster than Go\n\n")

	buf.WriteString("## Requirements\n\n")
	buf.WriteString("- Swift 5.9+\n")
	buf.WriteString("- macOS 13+, iOS 16+, tvOS 16+, or watchOS 9+\n\n")

	buf.WriteString("## Installation\n\n")
	buf.WriteString("### Swift Package Manager\n\n")
	buf.WriteString("Add this package as a dependency in your Package.swift:\n\n")
	buf.WriteString("```swift\n")
	buf.WriteString("dependencies: [\n")
	fmt.Fprintf(buf, "    .package(path: \"%s\")\n", packageDir)
	buf.WriteString("]\n")
	buf.WriteString("```\n\n")

	buf.WriteString("## Usage\n\n")
	buf.WriteString("```swift\n")
	fmt.Fprintf(buf, "import %s\n", config.Namespace)
	buf.WriteString("import Foundation\n\n")

	// Example with first message
	if len(config.Schema.Messages) > 0 {
		msg := config.Schema.Messages[0]
		fmt.Fprintf(buf, "// Create message\n")
		fmt.Fprintf(buf, "let message = %sMessage(...)\n\n", msg.Name)
		
		fmt.Fprintf(buf, "// Encode\n")
		fmt.Fprintf(buf, "let encoded = encode%sMessage(message)\n\n", msg.Name)
		
		fmt.Fprintf(buf, "// Decode\n")
		fmt.Fprintf(buf, "let decoded = try decode%sMessage(encoded)\n", msg.Name)
	}
	
	buf.WriteString("```\n\n")

	buf.WriteString("## API\n\n")
	buf.WriteString("### Message Types\n\n")
	buf.WriteString("All message types are Swift structs:\n\n")

	for _, msg := range config.Schema.Messages {
		fmt.Fprintf(buf, "- `%sMessage`\n", msg.Name)
	}

	buf.WriteString("\n### Encoder Functions\n\n")
	for _, msg := range config.Schema.Messages {
		fmt.Fprintf(buf, "```swift\n")
		fmt.Fprintf(buf, "func encode%sMessage(_ message: %sMessage) -> Data\n", msg.Name, msg.Name)
		fmt.Fprintf(buf, "```\n\n")
	}

	buf.WriteString("### Decoder Functions\n\n")
	for _, msg := range config.Schema.Messages {
		fmt.Fprintf(buf, "```swift\n")
		fmt.Fprintf(buf, "func decode%sMessage(_ data: Data) throws -> %sMessage\n", msg.Name, msg.Name)
		fmt.Fprintf(buf, "```\n\n")
	}

	buf.WriteString("## Performance\n\n")
	buf.WriteString("Native Swift optimizations:\n")
	buf.WriteString("- **Unsafe pointers**: Direct memory access, no bounds checking\n")
	buf.WriteString("- **Inlinable functions**: Compiler can inline encode/decode operations\n")
	buf.WriteString("- **Value types**: Stack allocation, no heap overhead\n")
	buf.WriteString("- **Zero-copy**: Primitive arrays use bulk memory operations\n")
	buf.WriteString("- **~10% faster than Go**: Benchmarked on complex schemas\n\n")

	buf.WriteString("## Platform Support\n\n")
	buf.WriteString("- macOS 13+\n")
	buf.WriteString("- iOS 16+\n")
	buf.WriteString("- tvOS 16+\n")
	buf.WriteString("- watchOS 9+\n\n")

	buf.WriteString("## License\n\n")
	buf.WriteString("Generated by FFire. See your schema's license for terms.\n")

	readmePath := filepath.Join(packageDir, "README.md")
	if err := os.WriteFile(readmePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	fmt.Printf("✓ Generated README.md: %s\n", readmePath)
	return nil
}
