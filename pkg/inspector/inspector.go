// Package inspector provides tools to visualize binary wire format.
package inspector

import (
	"bytes"
	"fmt"
	"math"
	"strings"

	"github.com/shaban/ffire/pkg/errors"
	"github.com/shaban/ffire/pkg/schema"
)

// Config holds configuration for inspection.
type Config struct {
	Schema      *schema.Schema
	MessageName string
	Data        []byte
	ShowHex     bool
	Compact     bool
}

// Inspect analyzes binary wire format and returns a human-readable visualization.
func Inspect(cfg *Config) (string, error) {
	// Find message type
	var messageType *schema.MessageType
	for i := range cfg.Schema.Messages {
		if cfg.Schema.Messages[i].Name == cfg.MessageName {
			messageType = &cfg.Schema.Messages[i]
			break
		}
	}

	if messageType == nil {
		return "", errors.Newf(errors.ErrMessageNotFound, "message type %s not found in schema", cfg.MessageName)
	}

	var buf bytes.Buffer

	// Header
	buf.WriteString(fmt.Sprintln("📦 Wire Format Inspector"))
	buf.WriteString(fmt.Sprintln("═══════════════════════════════════════"))
	buf.WriteString(fmt.Sprintf("Schema:  %s\n", cfg.Schema.Package))
	buf.WriteString(fmt.Sprintf("Message: %s\n", cfg.MessageName))
	buf.WriteString(fmt.Sprintf("Size:    %d bytes\n", len(cfg.Data)))
	buf.WriteString(fmt.Sprintln("═══════════════════════════════════════"))

	// Hex dump
	if cfg.ShowHex {
		buf.WriteString("Hex Dump:\n")
		buf.WriteString(hexDump(cfg.Data))
		buf.WriteString("\n")
	}

	// Parse and annotate
	buf.WriteString("Wire Format Breakdown:\n")
	buf.WriteString("───────────────────────────────────────\n")

	pos := 0
	annotations := &bytes.Buffer{}
	err := inspectValue(cfg.Data, &pos, messageType.TargetType, "", annotations, cfg.Compact, 0)
	if err != nil {
		return "", err
	}

	buf.WriteString(annotations.String())

	// Footer
	if pos != len(cfg.Data) {
		buf.WriteString(fmt.Sprintf("\n⚠️  Warning: %d bytes remaining (expected 0)\n", len(cfg.Data)-pos))
	} else {
		buf.WriteString("\n✅ Successfully parsed all bytes\n")
	}

	return buf.String(), nil
}

func hexDump(data []byte) string {
	var buf bytes.Buffer
	lines := (len(data) + 15) / 16

	for i := 0; i < lines; i++ {
		offset := i * 16
		end := offset + 16
		if end > len(data) {
			end = len(data)
		}

		// Offset
		buf.WriteString(fmt.Sprintf("%08x  ", offset))

		// Hex bytes
		for j := 0; j < 16; j++ {
			if offset+j < end {
				buf.WriteString(fmt.Sprintf("%02x ", data[offset+j]))
			} else {
				buf.WriteString("   ")
			}
			if j == 7 {
				buf.WriteString(" ")
			}
		}

		// ASCII
		buf.WriteString(" |")
		for j := 0; j < 16 && offset+j < end; j++ {
			b := data[offset+j]
			if b >= 32 && b <= 126 {
				buf.WriteByte(b)
			} else {
				buf.WriteByte('.')
			}
		}
		buf.WriteString("|\n")
	}

	return buf.String()
}

func inspectValue(data []byte, pos *int, typ schema.Type, path string, buf *bytes.Buffer, compact bool, indent int) error {
	startPos := *pos

	switch t := typ.(type) {
	case *schema.PrimitiveType:
		return inspectPrimitive(data, pos, t, path, buf, compact, indent, startPos)
	case *schema.ArrayType:
		return inspectArray(data, pos, t, path, buf, compact, indent, startPos)
	case *schema.StructType:
		return inspectStruct(data, pos, t, path, buf, compact, indent, startPos)
	default:
		return errors.Newf(errors.ErrUnknownType, "unknown type: %T", typ)
	}
}

func inspectPrimitive(data []byte, pos *int, typ *schema.PrimitiveType, path string, buf *bytes.Buffer, compact bool, indent int, startPos int) error {
	indentStr := strings.Repeat("  ", indent)

	// Optional flag
	if typ.Optional {
		if *pos >= len(data) {
			return fmt.Errorf("unexpected end of data at offset %d", *pos)
		}
		present := data[*pos]
		*pos++

		if present == 0x00 {
			if !compact {
				buf.WriteString(fmt.Sprintf("%s[%04x] %s: null (optional)\n", indentStr, startPos, path))
			}
			return nil
		}
	}

	switch typ.Name {
	case "bool":
		if *pos >= len(data) {
			return fmt.Errorf("unexpected end of data at offset %d", *pos)
		}
		value := data[*pos] == 0x01
		buf.WriteString(fmt.Sprintf("%s[%04x] %s: %v (bool, 1 byte)\n", indentStr, startPos, path, value))
		*pos++

	case "int8":
		if *pos >= len(data) {
			return fmt.Errorf("unexpected end of data at offset %d", *pos)
		}
		value := int8(data[*pos])
		buf.WriteString(fmt.Sprintf("%s[%04x] %s: %d (int8, 1 byte)\n", indentStr, startPos, path, value))
		*pos++

	case "int16":
		if *pos+1 >= len(data) {
			return fmt.Errorf("unexpected end of data at offset %d", *pos)
		}
		value := int16(uint16(data[*pos]) | uint16(data[*pos+1])<<8)
		buf.WriteString(fmt.Sprintf("%s[%04x] %s: %d (int16, 2 bytes)\n", indentStr, startPos, path, value))
		*pos += 2

	case "int32":
		if *pos+3 >= len(data) {
			return fmt.Errorf("unexpected end of data at offset %d", *pos)
		}
		value := int32(uint32(data[*pos]) | uint32(data[*pos+1])<<8 | uint32(data[*pos+2])<<16 | uint32(data[*pos+3])<<24)
		buf.WriteString(fmt.Sprintf("%s[%04x] %s: %d (int32, 4 bytes)\n", indentStr, startPos, path, value))
		*pos += 4

	case "int64":
		if *pos+7 >= len(data) {
			return fmt.Errorf("unexpected end of data at offset %d", *pos)
		}
		value := int64(uint64(data[*pos]) | uint64(data[*pos+1])<<8 | uint64(data[*pos+2])<<16 | uint64(data[*pos+3])<<24 |
			uint64(data[*pos+4])<<32 | uint64(data[*pos+5])<<40 | uint64(data[*pos+6])<<48 | uint64(data[*pos+7])<<56)
		buf.WriteString(fmt.Sprintf("%s[%04x] %s: %d (int64, 8 bytes)\n", indentStr, startPos, path, value))
		*pos += 8

	case "float32":
		if *pos+3 >= len(data) {
			return fmt.Errorf("unexpected end of data at offset %d", *pos)
		}
		bits := uint32(data[*pos]) | uint32(data[*pos+1])<<8 | uint32(data[*pos+2])<<16 | uint32(data[*pos+3])<<24
		value := math.Float32frombits(bits)
		buf.WriteString(fmt.Sprintf("%s[%04x] %s: %f (float32, 4 bytes)\n", indentStr, startPos, path, value))
		*pos += 4

	case "float64":
		if *pos+7 >= len(data) {
			return fmt.Errorf("unexpected end of data at offset %d", *pos)
		}
		bits := uint64(data[*pos]) | uint64(data[*pos+1])<<8 | uint64(data[*pos+2])<<16 | uint64(data[*pos+3])<<24 |
			uint64(data[*pos+4])<<32 | uint64(data[*pos+5])<<40 | uint64(data[*pos+6])<<48 | uint64(data[*pos+7])<<56
		value := math.Float64frombits(bits)
		buf.WriteString(fmt.Sprintf("%s[%04x] %s: %f (float64, 8 bytes)\n", indentStr, startPos, path, value))
		*pos += 8

	case "string":
		if *pos+1 >= len(data) {
			return fmt.Errorf("unexpected end of data at offset %d", *pos)
		}
		length := uint16(data[*pos]) | uint16(data[*pos+1])<<8
		*pos += 2

		if *pos+int(length) > len(data) {
			return fmt.Errorf("string length %d exceeds remaining data at offset %d", length, *pos)
		}
		value := string(data[*pos : *pos+int(length)])
		buf.WriteString(fmt.Sprintf("%s[%04x] %s: \"%s\" (string, %d bytes + 2 byte length)\n", indentStr, startPos, path, value, length))
		*pos += int(length)

	default:
		return errors.Newf(errors.ErrUnknownPrimitive, "unknown primitive type: %s", typ.Name)
	}

	return nil
}

func inspectArray(data []byte, pos *int, typ *schema.ArrayType, path string, buf *bytes.Buffer, compact bool, indent int, startPos int) error {
	indentStr := strings.Repeat("  ", indent)

	// Optional flag
	if typ.Optional {
		if *pos >= len(data) {
			return fmt.Errorf("unexpected end of data at offset %d", *pos)
		}
		present := data[*pos]
		*pos++

		if present == 0x00 {
			if !compact {
				buf.WriteString(fmt.Sprintf("%s[%04x] %s: null (optional array)\n", indentStr, startPos, path))
			}
			return nil
		}
	}

	// Array length
	if *pos+1 >= len(data) {
		return fmt.Errorf("unexpected end of data at offset %d", *pos)
	}
	length := uint16(data[*pos]) | uint16(data[*pos+1])<<8
	*pos += 2

	buf.WriteString(fmt.Sprintf("%s[%04x] %s: array[%d]\n", indentStr, startPos, path, length))

	// Array elements
	for i := 0; i < int(length); i++ {
		elemPath := fmt.Sprintf("%s[%d]", path, i)
		if err := inspectValue(data, pos, typ.ElementType, elemPath, buf, compact, indent+1); err != nil {
			return err
		}
	}

	return nil
}

func inspectStruct(data []byte, pos *int, typ *schema.StructType, path string, buf *bytes.Buffer, compact bool, indent int, startPos int) error {
	indentStr := strings.Repeat("  ", indent)

	// Optional flag
	if typ.Optional {
		if *pos >= len(data) {
			return fmt.Errorf("unexpected end of data at offset %d", *pos)
		}
		present := data[*pos]
		*pos++

		if present == 0x00 {
			if !compact {
				buf.WriteString(fmt.Sprintf("%s[%04x] %s: null (optional struct)\n", indentStr, startPos, path))
			}
			return nil
		}
	}

	if path == "" {
		buf.WriteString(fmt.Sprintf("%s[%04x] %s {\n", indentStr, startPos, typ.Name))
	} else {
		buf.WriteString(fmt.Sprintf("%s[%04x] %s: %s {\n", indentStr, startPos, path, typ.Name))
	}

	// Struct fields
	for _, field := range typ.Fields {
		fieldPath := field.Name
		if path != "" {
			fieldPath = path + "." + field.Name
		}
		if err := inspectValue(data, pos, field.Type, fieldPath, buf, compact, indent+1); err != nil {
			return err
		}
	}

	buf.WriteString(fmt.Sprintf("%s}\n", indentStr))

	return nil
}
