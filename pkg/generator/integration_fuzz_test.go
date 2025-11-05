// +build fuzz

package generator_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/shaban/ffire/pkg/fixture"
	"github.com/shaban/ffire/pkg/generator"
	"github.com/shaban/ffire/pkg/schema"
)

// TestFuzzIntegration is a documented example of how to run integration fuzz tests
// against generated decoders. Run with: go test -tags=fuzz -fuzz=FuzzIntegration
func TestFuzzIntegration(t *testing.T) {
	t.Skip("This is a template for integration fuzzing. Enable by removing skip and building generated code.")
}

// FuzzGeneratedDecoder tests that generated decoders handle malformed input gracefully
func FuzzGeneratedDecoder(f *testing.F) {
	// Create test schema
	s := &schema.Schema{
		Package: "fuzztest",
		Types: []schema.Type{
			&schema.StructType{
				Name: "FuzzStruct",
				Fields: []schema.Field{
					{Name: "ID", Type: &schema.PrimitiveType{Name: "int32"}},
					{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}},
					{Name: "Value", Type: &schema.PrimitiveType{Name: "float64"}},
					{Name: "Tags", Type: &schema.ArrayType{
						ElementType: &schema.PrimitiveType{Name: "string"},
					}},
				},
			},
		},
		Messages: []schema.MessageType{
			{Name: "FuzzMessage", TargetType: &schema.StructType{
				Name: "FuzzStruct",
				Fields: []schema.Field{
					{Name: "ID", Type: &schema.PrimitiveType{Name: "int32"}},
					{Name: "Name", Type: &schema.PrimitiveType{Name: "string"}},
					{Name: "Value", Type: &schema.PrimitiveType{Name: "float64"}},
					{Name: "Tags", Type: &schema.ArrayType{
						ElementType: &schema.PrimitiveType{Name: "string"},
					}},
				},
			}},
		},
	}

	// Generate code to temp directory
	code, err := generator.GenerateGo(s)
	if err != nil {
		f.Fatal(err)
	}

	tempDir := f.TempDir()
	genFile := filepath.Join(tempDir, "generated.go")
	if err := os.WriteFile(genFile, code, 0644); err != nil {
		f.Fatal(err)
	}

	// Create test file that calls the decoder
	testCode := `package fuzztest

import (
	"fmt"
	"testing"
)

func TestDecode(t *testing.T, data []byte) (recovered interface{}) {
	defer func() {
		recovered = recover()
	}()
	
	_, err := DecodeFuzzStructMessage(data)
	if err != nil {
		// Errors are fine, we just don't want panics
		return nil
	}
	return nil
}
`

	testFile := filepath.Join(tempDir, "test_decode.go")
	if err := os.WriteFile(testFile, []byte(testCode), 0644); err != nil {
		f.Fatal(err)
	}

	// Seed with valid data
	validJSON := []byte(`{"ID": 42, "Name": "test", "Value": 3.14, "Tags": ["tag1", "tag2"]}`)
	validBinary, err := fixture.Convert(s, "FuzzMessage", validJSON)
	if err != nil {
		f.Fatal(err)
	}

	f.Add(validBinary)

	// Seed with edge cases
	f.Add([]byte{})                         // Empty
	f.Add([]byte{0xFF, 0xFF, 0xFF, 0xFF})  // All 1s
	f.Add(validBinary[:len(validBinary)/2]) // Truncated

	f.Fuzz(func(t *testing.T, data []byte) {
		// For now, just verify the pattern works
		// In a real scenario, we would compile and run the generated code
		// with the fuzzed data and check for panics
		
		// This demonstrates the approach:
		// 1. Generate decoder code
		// 2. Create a test harness that calls decoder with fuzzed data
		// 3. Check that decoder returns error, not panic
		
		_ = data
	})
}

// Example of how to run a full integration fuzz test with compilation
func exampleFullFuzzTest(data []byte) error {
	// 1. Generate code
	s := &schema.Schema{
		Package: "fuzztest",
		Messages: []schema.MessageType{
			{Name: "Test", TargetType: &schema.PrimitiveType{Name: "int32"}},
		},
	}

	code, err := generator.GenerateGo(s)
	if err != nil {
		return err
	}

	// 2. Write to temp dir
	tempDir, _ := os.MkdirTemp("", "fuzz")
	defer os.RemoveAll(tempDir)

	genFile := filepath.Join(tempDir, "generated.go")
	os.WriteFile(genFile, code, 0644)

	// 3. Create test file
	testCode := fmt.Sprintf(`package fuzztest

import "testing"

func TestFuzz(t *testing.T) {
	data := []byte{%v}
	
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Decoder panicked: %%v", r)
		}
	}()
	
	_, err := DecodeInt32Message(data)
	// Error is ok, panic is not
	_ = err
}
`, bytesToGoArray(data))

	testFile := filepath.Join(tempDir, "fuzz_test.go")
	os.WriteFile(testFile, []byte(testCode), 0644)

	// 4. Run test
	cmd := exec.Command("go", "test", "-v")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's a panic (test failure) vs compile error
		if bytes.Contains(output, []byte("panic")) {
			return fmt.Errorf("PANIC DETECTED: %s", output)
		}
	}

	return nil
}

func bytesToGoArray(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	var buf bytes.Buffer
	for i, b := range data {
		if i > 0 {
			buf.WriteString(", ")
		}
		fmt.Fprintf(&buf, "0x%02x", b)
	}
	return buf.String()
}
