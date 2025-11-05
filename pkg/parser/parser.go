// Package parser parses .ffi schema files using Go's ast package.
package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"

	"github.com/shaban/ffire/pkg/schema"
)

// Parse parses a .ffi file and returns a Schema.
func Parse(filePath string) (*schema.Schema, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return ParseBytes(src)
}

// ParseBytes parses .ffi source code from bytes.
func ParseBytes(src []byte) (*schema.Schema, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse file: %w", err)
	}

	p := &schemaParser{
		fset:   fset,
		file:   file,
		types:  make(map[string]schema.Type),
		schema: &schema.Schema{},
	}

	return p.parse()
}

type schemaParser struct {
	fset   *token.FileSet
	file   *ast.File
	types  map[string]schema.Type
	schema *schema.Schema
}

func (p *schemaParser) parse() (*schema.Schema, error) {
	// Extract package name
	p.schema.Package = p.file.Name.Name

	// First pass: collect all type definitions
	for _, decl := range p.file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec := spec.(*ast.TypeSpec)
			if err := p.processTypeSpec(typeSpec); err != nil {
				return nil, err
			}
		}
	}

	// Second pass: resolve type references
	if err := p.resolveTypes(); err != nil {
		return nil, err
	}

	return p.schema, nil
}

func (p *schemaParser) processTypeSpec(spec *ast.TypeSpec) error {
	name := spec.Name.Name

	// Check if it's a type alias (message type)
	if spec.Assign != token.NoPos {
		// type Message = TargetType
		targetType, err := p.parseType(spec.Type)
		if err != nil {
			return fmt.Errorf("parse message type %s: %w", name, err)
		}
		p.schema.Messages = append(p.schema.Messages, schema.MessageType{
			Name:       name,
			TargetType: targetType,
		})
		return nil
	}

	// Regular type definition
	typ, err := p.parseType(spec.Type)
	if err != nil {
		return fmt.Errorf("parse type %s: %w", name, err)
	}

	// Store type
	p.types[name] = typ
	p.schema.Types = append(p.schema.Types, typ)

	return nil
}

func (p *schemaParser) parseType(expr ast.Expr) (schema.Type, error) {
	switch t := expr.(type) {
	case *ast.Ident:
		// Simple type name: int32, string, Device
		return &schema.PrimitiveType{Name: t.Name}, nil

	case *ast.StarExpr:
		// Optional type: *string, *int32
		innerType, err := p.parseType(t.X)
		if err != nil {
			return nil, err
		}
		// Mark as optional
		switch inner := innerType.(type) {
		case *schema.PrimitiveType:
			inner.Optional = true
			return inner, nil
		case *schema.StructType:
			inner.Optional = true
			return inner, nil
		case *schema.ArrayType:
			inner.Optional = true
			return inner, nil
		}
		return innerType, nil

	case *ast.ArrayType:
		// Array type: []int32, []Device
		if t.Len != nil {
			return nil, fmt.Errorf("fixed-size arrays not supported")
		}
		elemType, err := p.parseType(t.Elt)
		if err != nil {
			return nil, err
		}
		return &schema.ArrayType{ElementType: elemType}, nil

	case *ast.StructType:
		// Struct type definition
		return p.parseStruct(t)

	default:
		return nil, fmt.Errorf("unsupported type: %T", expr)
	}
}

func (p *schemaParser) parseStruct(structType *ast.StructType) (*schema.StructType, error) {
	var fields []schema.Field

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			return nil, fmt.Errorf("embedded fields not supported")
		}

		fieldType, err := p.parseType(field.Type)
		if err != nil {
			return nil, err
		}

		// Preserve full struct tag
		var fullTag string
		var jsonTag string
		if field.Tag != nil {
			fullTag = field.Tag.Value
			jsonTag = parseJSONTag(fullTag)
		}

		for _, name := range field.Names {
			f := schema.Field{
				Name: name.Name,
				Type: fieldType,
				Tag:  fullTag,
			}
			f.SetJSONTag(jsonTag)
			fields = append(fields, f)
		}
	}

	return &schema.StructType{Fields: fields}, nil
}

func (p *schemaParser) resolveTypes() error {
	// Resolve type references in all types
	for _, typ := range p.schema.Types {
		if err := p.resolveTypeReferences(typ); err != nil {
			return err
		}
	}

	// Resolve type references in message types
	for i, msg := range p.schema.Messages {
		resolved, err := p.resolveTypeReference(msg.TargetType)
		if err != nil {
			return fmt.Errorf("resolve message %s: %w", msg.Name, err)
		}
		p.schema.Messages[i].TargetType = resolved
	}

	return nil
}

func (p *schemaParser) resolveTypeReferences(typ schema.Type) error {
	switch t := typ.(type) {
	case *schema.StructType:
		// Update struct name if not set
		for name, storedType := range p.types {
			if storedType == t && t.Name == "" {
				t.Name = name
			}
		}
		// Resolve field types
		for i, field := range t.Fields {
			resolved, err := p.resolveTypeReference(field.Type)
			if err != nil {
				return err
			}
			t.Fields[i].Type = resolved
		}

	case *schema.ArrayType:
		resolved, err := p.resolveTypeReference(t.ElementType)
		if err != nil {
			return err
		}
		t.ElementType = resolved
	}

	return nil
}

func (p *schemaParser) resolveTypeReference(typ schema.Type) (schema.Type, error) {
	// Handle array types - need to resolve element type recursively
	if arrType, ok := typ.(*schema.ArrayType); ok {
		resolved, err := p.resolveTypeReference(arrType.ElementType)
		if err != nil {
			return nil, err
		}
		arrType.ElementType = resolved
		return arrType, nil
	}

	prim, ok := typ.(*schema.PrimitiveType)
	if !ok {
		return typ, nil
	}

	// If it's a known primitive, keep it
	if schema.IsPrimitive(prim.Name) {
		return typ, nil
	}

	// Look up in defined types
	resolved, exists := p.types[prim.Name]
	if !exists {
		return nil, fmt.Errorf("undefined type: %s", prim.Name)
	}

	// Preserve optional flag
	if prim.Optional {
		switch r := resolved.(type) {
		case *schema.StructType:
			copy := *r
			copy.Optional = true
			return &copy, nil
		case *schema.ArrayType:
			copy := *r
			copy.Optional = true
			return &copy, nil
		}
	}

	return resolved, nil
}

// parseJSONTag extracts the JSON field name from a struct tag.
// Example: `json:"name,omitempty"` returns "name"
func parseJSONTag(tagValue string) string {
	// Remove quotes
	if len(tagValue) >= 2 && tagValue[0] == '`' && tagValue[len(tagValue)-1] == '`' {
		tagValue = tagValue[1 : len(tagValue)-1]
	}

	// Parse struct tag
	tag := tagValue
	jsonPrefix := "json:\""
	start := 0
	for i := 0; i < len(tag); i++ {
		if i+len(jsonPrefix) <= len(tag) && tag[i:i+len(jsonPrefix)] == jsonPrefix {
			start = i + len(jsonPrefix)
			break
		}
	}

	if start == 0 {
		return ""
	}

	// Find the end quote
	end := start
	for end < len(tag) && tag[end] != '"' {
		end++
	}

	if end == start {
		return ""
	}

	jsonName := tag[start:end]

	// Strip options after comma (e.g., "name,omitempty" -> "name")
	for i, ch := range jsonName {
		if ch == ',' {
			return jsonName[:i]
		}
	}

	return jsonName
}

// Helper to extract type name for error messages
func typeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeName(t.X)
	case *ast.ArrayType:
		return "[]" + typeName(t.Elt)
	default:
		return fmt.Sprintf("%T", expr)
	}
}
