// Package generator generates encoder/decoder code for various languages.
package generator

import (
"fmt"

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

// GenerateSwift generates Swift encoder/decoder code.
func GenerateSwift(s *schema.Schema) ([]byte, error) {
	return nil, fmt.Errorf("Swift generation not yet implemented")
}
