package errors

import "testing"

func TestErrorCodes(t *testing.T) {
	tests := []struct {
		name string
		code ErrorCode
		msg  string
		want string
	}{
		{
			name: "empty package",
			code: ErrEmptyPackage,
			msg:  "package name is required",
			want: "[E001] package name is required",
		},
		{
			name: "no messages",
			code: ErrNoMessages,
			msg:  "at least one message type is required",
			want: "[E002] at least one message type is required",
		},
		{
			name: "circular reference",
			code: ErrCircularReference,
			msg:  "circular reference detected: Node",
			want: "[E010] circular reference detected: Node",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.code, tt.msg)
			if got := err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsCode(t *testing.T) {
	err := New(ErrNoMessages, "test message")

	if !IsCode(err, ErrNoMessages) {
		t.Errorf("IsCode() should return true for matching code")
	}

	if IsCode(err, ErrEmptyPackage) {
		t.Errorf("IsCode() should return false for non-matching code")
	}
}

func TestGetCode(t *testing.T) {
	err := New(ErrCircularReference, "test")

	if got := GetCode(err); got != ErrCircularReference {
		t.Errorf("GetCode() = %v, want %v", got, ErrCircularReference)
	}
}

func TestErrorWithContext(t *testing.T) {
	err := New(ErrUndefinedType, "undefined type: Foo").
		WithContext("type", "Foo").
		WithContext("line", 42)

	errStr := err.Error()
	if errStr != "[E005] undefined type: Foo (type=Foo, line=42)" &&
		errStr != "[E005] undefined type: Foo (line=42, type=Foo)" {
		t.Errorf("Error() = %q, want context in output", errStr)
	}
}
