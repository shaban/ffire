// Package errors provides structured error codes for ffire.
package errors

import "fmt"

// ErrorCode represents a unique error identifier.
type ErrorCode string

const (
	// Schema validation errors (E001-E010)
	ErrEmptyPackage      ErrorCode = "E001" // Package name is required
	ErrNoMessages        ErrorCode = "E002" // At least one message type is required
	ErrEmptyMessageName  ErrorCode = "E003" // Message type name cannot be empty
	ErrNilTargetType     ErrorCode = "E004" // Message target type cannot be nil
	ErrUndefinedType     ErrorCode = "E005" // Reference to undefined type
	ErrEmptyStruct       ErrorCode = "E006" // Struct has no fields
	ErrEmptyFieldName    ErrorCode = "E007" // Field name cannot be empty
	ErrNilFieldType      ErrorCode = "E008" // Field type cannot be nil
	ErrNilArrayElement   ErrorCode = "E009" // Array element type cannot be nil
	ErrCircularReference ErrorCode = "E010" // Circular type reference detected

	// Nesting and complexity errors (E011-E015)
	ErrMaxNestingDepth ErrorCode = "E011" // Nesting depth exceeds maximum
	ErrUnknownType     ErrorCode = "E012" // Unknown type encountered

	// JSON validation errors (E013-E020)
	ErrMessageNotFound ErrorCode = "E013" // Message type not found in schema
	ErrInvalidJSON     ErrorCode = "E014" // Invalid JSON format
	ErrRequiredField   ErrorCode = "E015" // Required field is missing or null
	ErrTypeMismatch    ErrorCode = "E016" // Value type doesn't match schema
	ErrIntegerExpected ErrorCode = "E017" // Integer value expected
	ErrNumberExpected  ErrorCode = "E018" // Number value expected
	ErrStringExpected  ErrorCode = "E019" // String value expected
	ErrObjectExpected  ErrorCode = "E020" // Object value expected
	ErrArrayExpected   ErrorCode = "E021" // Array value expected
	ErrBoolExpected    ErrorCode = "E022" // Boolean value expected

	// Range validation errors (E023-E028)
	ErrInt8OutOfRange   ErrorCode = "E023" // Value out of range for int8
	ErrInt16OutOfRange  ErrorCode = "E024" // Value out of range for int16
	ErrInt32OutOfRange  ErrorCode = "E025" // Value out of range for int32
	ErrStringTooLong    ErrorCode = "E026" // String exceeds maximum length (65535 bytes)
	ErrArrayTooLong     ErrorCode = "E027" // Array exceeds maximum length (65535 elements)
	ErrUnknownPrimitive ErrorCode = "E028" // Unknown primitive type

	// File I/O errors (E029-E032)
	ErrFileRead   ErrorCode = "E029" // Failed to read file
	ErrFileWrite  ErrorCode = "E030" // Failed to write file
	ErrFileParse  ErrorCode = "E031" // Failed to parse schema file
	ErrFileCreate ErrorCode = "E032" // Failed to create file or directory
)

// errorHints provides helpful hints for each error code
var errorHints = map[ErrorCode]string{
	ErrEmptyPackage:      "Add a package declaration at the top of your schema file, e.g., 'package myapp'",
	ErrNoMessages:        "Define at least one message type, e.g., 'type Message = YourType'",
	ErrEmptyMessageName:  "Message type must have a name, e.g., 'type Message = ...'",
	ErrUndefinedType:     "Make sure the type is defined before using it, or use a built-in type (string, int32, float32, etc.)",
	ErrEmptyStruct:       "Structs must have at least one field",
	ErrCircularReference: "Types cannot reference themselves directly or indirectly",
	ErrMaxNestingDepth:   "Reduce nesting depth by flattening your data structure or using separate types",
	ErrMessageNotFound:   "Check that the message name matches one defined in your schema",
	ErrInvalidJSON:       "Ensure your JSON is well-formed (use a JSON validator)",
	ErrInt8OutOfRange:    "int8 values must be between -128 and 127",
	ErrInt16OutOfRange:   "int16 values must be between -32768 and 32767",
	ErrInt32OutOfRange:   "int32 values must be between -2147483648 and 2147483647",
	ErrStringTooLong:     "Strings are limited to 65,535 bytes in the wire format",
	ErrArrayTooLong:      "Arrays are limited to 65,535 elements in the wire format",
}

// Error represents a structured error with code and context.
type Error struct {
	Code    ErrorCode
	Message string
	Context map[string]interface{}
}

// Error implements the error interface.
func (e *Error) Error() string {
	if len(e.Context) == 0 {
		return fmt.Sprintf("[%s] %s", e.Code, e.Message)
	}

	// Build context string
	contextStr := ""
	for k, v := range e.Context {
		if contextStr != "" {
			contextStr += ", "
		}
		contextStr += fmt.Sprintf("%s=%v", k, v)
	}
	return fmt.Sprintf("[%s] %s (%s)", e.Code, e.Message, contextStr)
}

// New creates a new Error with the given code and message.
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// Newf creates a new Error with formatted message.
func Newf(code ErrorCode, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context to an error.
func (e *Error) WithContext(key string, value interface{}) *Error {
	e.Context[key] = value
	return e
}

// Hint returns a helpful hint for the error, if available.
func (e *Error) Hint() string {
	if hint, ok := errorHints[e.Code]; ok {
		return hint
	}
	return ""
}

// ErrorWithHint returns the error message with an optional hint.
func (e *Error) ErrorWithHint() string {
	msg := e.Error()
	if hint := e.Hint(); hint != "" {
		return fmt.Sprintf("%s\n💡 Hint: %s", msg, hint)
	}
	return msg
}

// IsCode checks if an error has a specific error code.
func IsCode(err error, code ErrorCode) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == code
	}
	// Try unwrapping if it's a wrapped error
	if unwrapped := Unwrap(err); unwrapped != nil {
		return IsCode(unwrapped, code)
	}
	return false
}

// GetCode extracts the error code from an error, or returns empty string.
func GetCode(err error) ErrorCode {
	if e, ok := err.(*Error); ok {
		return e.Code
	}
	// Try unwrapping if it's a wrapped error
	if unwrapped := Unwrap(err); unwrapped != nil {
		return GetCode(unwrapped)
	}
	return ""
}

// Unwrap returns the wrapped error if it exists.
func Unwrap(err error) error {
	type unwrapper interface {
		Unwrap() error
	}
	if u, ok := err.(unwrapper); ok {
		return u.Unwrap()
	}
	return nil
}
