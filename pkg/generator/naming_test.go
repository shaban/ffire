package generator

import "testing"

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my_message", "MyMessage"},
		{"user-profile", "UserProfile"},
		{"simple", "Simple"},
		{"API_KEY", "APIKEY"},
		{"snake_case_name", "SnakeCaseName"},
		{"kebab-case-name", "KebabCaseName"},
		{"", ""},
	}

	for _, tt := range tests {
		result := ToPascalCase(tt.input)
		if result != tt.expected {
			t.Errorf("ToPascalCase(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my_message", "myMessage"},
		{"user-profile", "userProfile"},
		{"simple", "simple"},
		{"API_KEY", "aPIKEY"},
		{"", ""},
	}

	for _, tt := range tests {
		result := ToCamelCase(tt.input)
		if result != tt.expected {
			t.Errorf("ToCamelCase(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"MyMessage", "my_message"},
		{"userProfile", "user_profile"},
		{"simple", "simple"},
		{"HTTPResponse", "h_t_t_p_response"},
		{"", ""},
	}

	for _, tt := range tests {
		result := ToSnakeCase(tt.input)
		if result != tt.expected {
			t.Errorf("ToSnakeCase(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"MyMessage", "my-message"},
		{"userProfile", "user-profile"},
		{"simple", "simple"},
		{"", ""},
	}

	for _, tt := range tests {
		result := ToKebabCase(tt.input)
		if result != tt.expected {
			t.Errorf("ToKebabCase(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestToScreamingSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"MyMessage", "MY_MESSAGE"},
		{"userProfile", "USER_PROFILE"},
		{"simple", "SIMPLE"},
		{"", ""},
	}

	for _, tt := range tests {
		result := ToScreamingSnakeCase(tt.input)
		if result != tt.expected {
			t.Errorf("ToScreamingSnakeCase(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

// Test language-specific functions (they should all behave the same as ToPascalCase)
func TestLanguageSpecificClassNames(t *testing.T) {
	input := "my_message"
	expected := "MyMessage"

	funcs := map[string]func(string) string{
		"ToRubyClassName":      ToRubyClassName,
		"ToPythonClassName":    ToPythonClassName,
		"ToJavaScriptClassName": ToJavaScriptClassName,
		"ToSwiftClassName":     ToSwiftClassName,
		"ToGoTypeName":         ToGoTypeName,
		"ToRustTypeName":       ToRustTypeName,
		"ToCppClassName":       ToCppClassName,
	}

	for name, fn := range funcs {
		result := fn(input)
		if result != expected {
			t.Errorf("%s(%q) = %q; want %q", name, input, result, expected)
		}
	}
}
