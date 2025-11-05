// Package validator validates schemas and JSON fixtures.
package validator

import (
	"encoding/json"
	"fmt"

	"github.com/shaban/ffire/pkg/schema"
)

const maxNestingDepth = 32

// ValidateSchema checks if a schema is well-formed.
func ValidateSchema(s *schema.Schema) error {
	if s.Package == "" {
		return fmt.Errorf("package name is required")
	}

	if len(s.Messages) == 0 {
		return fmt.Errorf("at least one message type is required")
	}

	// Check all message types reference valid types
	for _, msg := range s.Messages {
		if msg.Name == "" {
			return fmt.Errorf("message type name cannot be empty")
		}
		if msg.TargetType == nil {
			return fmt.Errorf("message %s: target type cannot be nil", msg.Name)
		}
		if err := validateType(s, msg.TargetType, 0); err != nil {
			return fmt.Errorf("message %s: %w", msg.Name, err)
		}
	}

	// Check all defined types are valid
	for _, typ := range s.Types {
		if err := validateType(s, typ, 0); err != nil {
			return fmt.Errorf("type %s: %w", typ.TypeName(), err)
		}
	}

	// Check for circular references
	if err := checkCircularReferences(s); err != nil {
		return err
	}

	return nil
}

// validateType recursively validates a type and its nesting depth.
func validateType(s *schema.Schema, typ schema.Type, depth int) error {
	if depth > maxNestingDepth {
		return fmt.Errorf("nesting depth exceeds maximum of %d", maxNestingDepth)
	}

	switch t := typ.(type) {
	case *schema.PrimitiveType:
		if !schema.IsPrimitive(t.Name) {
			// Check if it's a defined type
			if s.FindType(t.Name) == nil {
				return fmt.Errorf("undefined type: %s", t.Name)
			}
		}

	case *schema.StructType:
		if len(t.Fields) == 0 {
			return fmt.Errorf("struct %s has no fields", t.Name)
		}
		for _, field := range t.Fields {
			if field.Name == "" {
				return fmt.Errorf("struct %s: field name cannot be empty", t.Name)
			}
			if field.Type == nil {
				return fmt.Errorf("struct %s: field %s has nil type", t.Name, field.Name)
			}
			if err := validateType(s, field.Type, depth+1); err != nil {
				return fmt.Errorf("struct %s: field %s: %w", t.Name, field.Name, err)
			}
		}

	case *schema.ArrayType:
		if t.ElementType == nil {
			return fmt.Errorf("array element type cannot be nil")
		}
		if err := validateType(s, t.ElementType, depth+1); err != nil {
			return fmt.Errorf("array element: %w", err)
		}

	default:
		return fmt.Errorf("unknown type: %T", typ)
	}

	return nil
}

// checkCircularReferences detects circular type references.
func checkCircularReferences(s *schema.Schema) error {
	for _, typ := range s.Types {
		visited := make(map[string]bool)
		if err := detectCycle(s, typ, visited); err != nil {
			return err
		}
	}
	return nil
}

// detectCycle performs DFS to detect cycles.
func detectCycle(s *schema.Schema, typ schema.Type, visited map[string]bool) error {
	name := typ.TypeName()

	// Skip primitives
	if schema.IsPrimitive(name) {
		return nil
	}

	// Check for cycle
	if visited[name] {
		return fmt.Errorf("circular reference detected: %s", name)
	}

	visited[name] = true
	defer delete(visited, name)

	switch t := typ.(type) {
	case *schema.StructType:
		for _, field := range t.Fields {
			if err := detectCycle(s, field.Type, visited); err != nil {
				return err
			}
		}

	case *schema.ArrayType:
		if err := detectCycle(s, t.ElementType, visited); err != nil {
			return err
		}
	}

	return nil
}

// ValidateJSON validates that JSON data matches the schema.
func ValidateJSON(s *schema.Schema, messageName string, jsonData []byte) error {
	// Find the message type
	var messageType *schema.MessageType
	for i := range s.Messages {
		if s.Messages[i].Name == messageName {
			messageType = &s.Messages[i]
			break
		}
	}

	if messageType == nil {
		return fmt.Errorf("message type %s not found in schema", messageName)
	}

	// Parse JSON
	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Validate against type
	return validateJSONValue(s, messageType.TargetType, data, "")
}

// validateJSONValue recursively validates a JSON value against a type.
func validateJSONValue(s *schema.Schema, typ schema.Type, value interface{}, path string) error {
	// Handle optional types
	if typ.IsOptional() {
		if value == nil {
			return nil
		}
	} else {
		if value == nil {
			return fmt.Errorf("%s: required field is null", path)
		}
	}

	switch t := typ.(type) {
	case *schema.PrimitiveType:
		return validatePrimitive(t, value, path)

	case *schema.StructType:
		return validateStruct(s, t, value, path)

	case *schema.ArrayType:
		return validateArray(s, t, value, path)

	default:
		return fmt.Errorf("%s: unknown type %T", path, typ)
	}
}

// validatePrimitive validates a primitive value.
func validatePrimitive(typ *schema.PrimitiveType, value interface{}, path string) error {
	if value == nil && typ.Optional {
		return nil
	}

	switch typ.Name {
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("%s: expected bool, got %T", path, value)
		}

	case "int8", "int16", "int32", "int64":
		// JSON unmarshals numbers as float64
		if num, ok := value.(float64); ok {
			// Check if it's an integer
			if num != float64(int64(num)) {
				return fmt.Errorf("%s: expected integer, got %v", path, num)
			}
			// Check range for specific types
			switch typ.Name {
			case "int8":
				if num < -128 || num > 127 {
					return fmt.Errorf("%s: value %v out of range for int8", path, num)
				}
			case "int16":
				if num < -32768 || num > 32767 {
					return fmt.Errorf("%s: value %v out of range for int16", path, num)
				}
			case "int32":
				if num < -2147483648 || num > 2147483647 {
					return fmt.Errorf("%s: value %v out of range for int32", path, num)
				}
			}
		} else {
			return fmt.Errorf("%s: expected number, got %T", path, value)
		}

	case "float32", "float64":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("%s: expected number, got %T", path, value)
		}

	case "string":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("%s: expected string, got %T", path, value)
		}
		// Validate string length (uint16 wire format limit)
		if len(str) > 65535 {
			return fmt.Errorf("%s: string length %d exceeds maximum of 65,535 bytes", path, len(str))
		}

	default:
		return fmt.Errorf("%s: unknown primitive type: %s", path, typ.Name)
	}

	return nil
}

// validateStruct validates a struct value.
func validateStruct(s *schema.Schema, typ *schema.StructType, value interface{}, path string) error {
	if value == nil && typ.Optional {
		return nil
	}

	obj, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("%s: expected object, got %T", path, value)
	}

	// Check all required fields are present
	for _, field := range typ.Fields {
		fieldPath := path + "." + field.Name
		if path == "" {
			fieldPath = field.Name
		}

		jsonName := field.JSONName()
		fieldValue, exists := obj[jsonName]
		if !exists {
			if !field.Type.IsOptional() {
				return fmt.Errorf("%s: required field missing", fieldPath)
			}
			continue
		}

		if err := validateJSONValue(s, field.Type, fieldValue, fieldPath); err != nil {
			return err
		}
	}

	return nil
}

// validateArray validates an array value.
func validateArray(s *schema.Schema, typ *schema.ArrayType, value interface{}, path string) error {
	if value == nil && typ.Optional {
		return nil
	}

	arr, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("%s: expected array, got %T", path, value)
	}

	// Validate array length (uint16 wire format limit)
	if len(arr) > 65535 {
		return fmt.Errorf("%s: array length %d exceeds maximum of 65,535 elements", path, len(arr))
	}

	// Validate each element
	for i, elem := range arr {
		elemPath := fmt.Sprintf("%s[%d]", path, i)
		if err := validateJSONValue(s, typ.ElementType, elem, elemPath); err != nil {
			return err
		}
	}

	return nil
}
