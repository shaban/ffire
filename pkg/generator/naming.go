package generator

import (
	"strings"
	"unicode"
)

// String transformation utilities for cross-language code generation
// These functions help convert schema names to language-specific naming conventions

// ToPascalCase converts a string to PascalCase (UpperCamelCase)
// Examples: "my_message" -> "MyMessage", "user-profile" -> "UserProfile"
func ToPascalCase(s string) string {
	// Handle snake_case and kebab-case
	s = strings.ReplaceAll(s, "-", "_")
	parts := strings.Split(s, "_")

	for i, part := range parts {
		if part != "" {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}

	return strings.Join(parts, "")
}

// ToCamelCase converts a string to camelCase (lowerCamelCase)
// Examples: "my_message" -> "myMessage", "user-profile" -> "userProfile"
func ToCamelCase(s string) string {
	pascal := ToPascalCase(s)
	if pascal == "" {
		return ""
	}
	return strings.ToLower(pascal[:1]) + pascal[1:]
}

// ToSnakeCase converts a string to snake_case
// Examples: "MyMessage" -> "my_message", "userProfile" -> "user_profile"
func ToSnakeCase(s string) string {
	var result strings.Builder

	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// ToKebabCase converts a string to kebab-case
// Examples: "MyMessage" -> "my-message", "user_profile" -> "user-profile"
func ToKebabCase(s string) string {
	return strings.ReplaceAll(ToSnakeCase(s), "_", "-")
}

// ToScreamingSnakeCase converts a string to SCREAMING_SNAKE_CASE
// Examples: "myMessage" -> "MY_MESSAGE", "user-profile" -> "USER_PROFILE"
func ToScreamingSnakeCase(s string) string {
	return strings.ToUpper(ToSnakeCase(s))
}

// Language-specific convenience functions

// ToRubyClassName converts to Ruby class name (PascalCase)
// Examples: "my_message" -> "MyMessage"
func ToRubyClassName(s string) string {
	return ToPascalCase(s)
}

// ToPythonClassName converts to Python class name (PascalCase)
// Examples: "my_message" -> "MyMessage"
func ToPythonClassName(s string) string {
	return ToPascalCase(s)
}

// ToJavaScriptClassName converts to JavaScript class name (PascalCase)
// Examples: "my_message" -> "MyMessage"
func ToJavaScriptClassName(s string) string {
	return ToPascalCase(s)
}

// ToSwiftClassName converts to Swift class/struct name (PascalCase)
// Examples: "my_message" -> "MyMessage"
func ToSwiftClassName(s string) string {
	return ToPascalCase(s)
}

// ToGoTypeName converts to Go type name (PascalCase, exported)
// Examples: "my_message" -> "MyMessage"
func ToGoTypeName(s string) string {
	return ToPascalCase(s)
}

// ToRustTypeName converts to Rust type name (PascalCase)
// Examples: "my_message" -> "MyMessage"
func ToRustTypeName(s string) string {
	return ToPascalCase(s)
}

// ToCppClassName converts to C++ class name (PascalCase)
// Examples: "my_message" -> "MyMessage"
func ToCppClassName(s string) string {
	return ToPascalCase(s)
}
