// Package fixture converts JSON test data to binary wire format.
package fixture

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/shaban/ffire/internal/wire"
	"github.com/shaban/ffire/pkg/schema"
)

// Convert converts JSON data to binary wire format according to schema.
func Convert(s *schema.Schema, messageName string, jsonData []byte) ([]byte, error) {
	// Find the message type
	var messageType *schema.MessageType
	for i := range s.Messages {
		if s.Messages[i].Name == messageName {
			messageType = &s.Messages[i]
			break
		}
	}

	if messageType == nil {
		return nil, fmt.Errorf("message type %s not found in schema", messageName)
	}

	// Parse JSON
	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Encode to binary
	buf := &bytes.Buffer{}
	if err := encodeValue(buf, s, messageType.TargetType, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// encodeValue encodes a JSON value to binary format.
func encodeValue(buf *bytes.Buffer, s *schema.Schema, typ schema.Type, value interface{}) error {
	// Handle optional types
	if typ.IsOptional() {
		if value == nil {
			// Write present flag = false
			wire.EncodeBool(buf, false)
			return nil
		}
		// Write present flag = true
		wire.EncodeBool(buf, true)
	}

	switch t := typ.(type) {
	case *schema.PrimitiveType:
		return encodePrimitive(buf, t, value)

	case *schema.StructType:
		return encodeStruct(buf, s, t, value)

	case *schema.ArrayType:
		return encodeArray(buf, s, t, value)

	default:
		return fmt.Errorf("unknown type: %T", typ)
	}
}

// encodePrimitive encodes a primitive value.
func encodePrimitive(buf *bytes.Buffer, typ *schema.PrimitiveType, value interface{}) error {
	if value == nil && typ.Optional {
		return nil // Already handled by encodeValue
	}

	switch typ.Name {
	case "bool":
		v, ok := value.(bool)
		if !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
		wire.EncodeBool(buf, v)
		return nil

	case "int8":
		num, ok := value.(float64)
		if !ok {
			return fmt.Errorf("expected number, got %T", value)
		}
		wire.EncodeInt8(buf, int8(num))
		return nil

	case "int16":
		num, ok := value.(float64)
		if !ok {
			return fmt.Errorf("expected number, got %T", value)
		}
		wire.EncodeInt16(buf, int16(num))
		return nil

	case "int32":
		num, ok := value.(float64)
		if !ok {
			return fmt.Errorf("expected number, got %T", value)
		}
		wire.EncodeInt32(buf, int32(num))
		return nil

	case "int64":
		num, ok := value.(float64)
		if !ok {
			return fmt.Errorf("expected number, got %T", value)
		}
		wire.EncodeInt64(buf, int64(num))
		return nil

	case "float32":
		num, ok := value.(float64)
		if !ok {
			return fmt.Errorf("expected number, got %T", value)
		}
		wire.EncodeFloat32(buf, float32(num))
		return nil

	case "float64":
		num, ok := value.(float64)
		if !ok {
			return fmt.Errorf("expected number, got %T", value)
		}
		wire.EncodeFloat64(buf, num)
		return nil

	case "string":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
		wire.EncodeString(buf, str)
		return nil

	default:
		return fmt.Errorf("unknown primitive type: %s", typ.Name)
	}
}

// encodeStruct encodes a struct value.
func encodeStruct(buf *bytes.Buffer, s *schema.Schema, typ *schema.StructType, value interface{}) error {
	if value == nil && typ.Optional {
		return nil // Already handled by encodeValue
	}

	obj, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected object, got %T", value)
	}

	// Encode each field in order
	for _, field := range typ.Fields {
		jsonName := field.JSONName()
		fieldValue, exists := obj[jsonName]
		if !exists {
			if !field.Type.IsOptional() {
				return fmt.Errorf("required field %s missing", field.Name)
			}
			// For optional fields, encode as not present
			wire.EncodeBool(buf, false)
			continue
		}

		if err := encodeValue(buf, s, field.Type, fieldValue); err != nil {
			return fmt.Errorf("encode field %s: %w", field.Name, err)
		}
	}

	return nil
}

// encodeArray encodes an array value.
func encodeArray(buf *bytes.Buffer, s *schema.Schema, typ *schema.ArrayType, value interface{}) error {
	if value == nil && typ.Optional {
		return nil // Already handled by encodeValue
	}

	arr, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("expected array, got %T", value)
	}

	// Write array length (uint16 - validator ensures < 65536)
	wire.EncodeArrayHeader(buf, uint16(len(arr)))

	// Write each element
	for i, elem := range arr {
		if err := encodeValue(buf, s, typ.ElementType, elem); err != nil {
			return fmt.Errorf("encode element %d: %w", i, err)
		}
	}

	return nil
}
