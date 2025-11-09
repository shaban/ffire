package generator

// cAndGoKeywords lists reserved keywords shared between C and Go languages.
//
// Purpose:
//   - Detect potential C ABI function name collisions
//   - C ABI generates functions like: {lowercase_type}_decode, {lowercase_type}_encode
//   - If type name is a C/Go keyword, collision occurs (e.g., "type Register" → register_decode)
//
// Scope:
//   - Used ONLY for C ABI function prefix detection
//   - NOT used for type names (Message suffix handles type collisions universally)
//   - NOT used for field names or other identifiers
//
// Why C+Go merged:
//  1. Most keywords overlap between C and Go (break, case, const, continue, etc.)
//  2. Go generator also benefits from collision detection
//  3. Single source of truth reduces maintenance
//  4. Both languages would need similar handling anyway
//
// Current status:
//   - NOT YET IMPLEMENTED in generators (C compiler will error if collision occurs)
//   - Recommended fix: Detect keyword, append '_msg' suffix (register_msg_decode)
//   - See .copilot-instructions.md for implementation strategy
var cAndGoKeywords = map[string]bool{
	// C89/C99/C11 keywords
	"auto":     true,
	"break":    true,
	"case":     true,
	"char":     true,
	"const":    true,
	"continue": true,
	"default":  true,
	"do":       true,
	"double":   true,
	"else":     true,
	"enum":     true,
	"extern":   true,
	"float":    true,
	"for":      true,
	"goto":     true,
	"if":       true,
	"inline":   true, // C99+
	"int":      true,
	"long":     true,
	"register": true,
	"restrict": true, // C99+
	"return":   true,
	"short":    true,
	"signed":   true,
	"sizeof":   true,
	"static":   true,
	"struct":   true,
	"switch":   true,
	"typedef":  true,
	"union":    true,
	"unsigned": true,
	"void":     true,
	"volatile": true,
	"while":    true,

	// Go-specific keywords (not in C)
	"chan":        true,
	"defer":       true,
	"fallthrough": true,
	"func":        true,
	"go":          true,
	"import":      true,
	"interface":   true,
	"map":         true,
	"package":     true,
	"range":       true,
	"select":      true,
	"type":        true,
	"var":         true,
}

// IsCOrGoKeyword checks if a lowercase identifier is a C or Go keyword.
//
// Used to detect potential C ABI function name collisions:
//   - type Register → register_decode (collision!)
//   - type Config → config_decode (safe)
//
// Note: Type names themselves don't collide because they get Message suffix.
// Only C ABI function prefixes need checking.
func IsCOrGoKeyword(name string) bool {
	return cAndGoKeywords[name]
}
