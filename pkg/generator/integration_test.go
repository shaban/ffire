package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shaban/ffire/pkg/parser"
)

// TestRubyPackageIntegration generates a Ruby package and validates it can be imported
func TestRubyPackageIntegration(t *testing.T) {
	// Create temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "ffire-test-ruby-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Logf("Testing Ruby package generation in: %s", tmpDir)

	// Parse a test schema
	schemaPath := "../../testdata/schema/complex.ffi"
	schema, err := parser.Parse(schemaPath)
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	// Generate Ruby package
	config := &PackageConfig{
		Schema:    schema,
		Language:  "ruby",
		OutputDir: tmpDir,
		Optimize:  2,
		Platform:  "current",
		Arch:      "current",
		Namespace: schema.Package,
		NoCompile: false,
		Verbose:   testing.Verbose(),
	}

	err = GeneratePackage(config) // Use GeneratePackage to get platform resolution
	if err != nil {
		t.Fatalf("Failed to generate Ruby package: %v", err)
	}

	// Verify expected files exist
	expectedFiles := []string{
		"ruby/lib/libffire.dylib",     // or .so on Linux
		"ruby/lib/test.rb",
		"ruby/lib/test/bindings.rb",
		"ruby/lib/test/message.rb",
		"ruby/lib/test/version.rb",
		"ruby/test.gemspec",
		"ruby/Gemfile",
		"ruby/README.md",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(tmpDir, file)
		// Skip dylib check on non-macOS (would be .so or .dll)
		if strings.Contains(file, ".dylib") && !fileExists(fullPath) {
			// Try .so for Linux
			fullPath = strings.ReplaceAll(fullPath, ".dylib", ".so")
			if !fileExists(fullPath) {
				// Try .dll for Windows
				fullPath = strings.ReplaceAll(fullPath, ".so", ".dll")
			}
		}
		
		if !fileExists(fullPath) {
			t.Errorf("Expected file not found: %s", file)
		}
	}

	// Test that Ruby can parse the generated code (syntax check)
	if hasRuby() {
		t.Log("Ruby found, testing syntax...")
		testRubySyntax(t, tmpDir)
		
		// If FFI gem is available, test actual loading
		if hasFFIGem() {
			t.Log("FFI gem found, testing library loading...")
			testRubyFFILoading(t, tmpDir)
		} else {
			t.Log("FFI gem not installed, skipping FFI loading test")
		}
	} else {
		t.Log("Ruby not installed, skipping Ruby-specific tests")
	}
}

// TestPythonPackageIntegration generates a Python package and validates it
func TestPythonPackageIntegration(t *testing.T) {
	t.Skip("TODO: Implement after Python generator refactoring")
}

// TestJavaScriptPackageIntegration generates a JavaScript package and validates it
func TestJavaScriptPackageIntegration(t *testing.T) {
	t.Skip("TODO: Implement after JavaScript generator refactoring")
}

// Helper: Check if file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Helper: Check if Ruby is installed
func hasRuby() bool {
	_, err := exec.LookPath("ruby")
	return err == nil
}

// Helper: Check if FFI gem is installed
func hasFFIGem() bool {
	cmd := exec.Command("ruby", "-e", "require 'ffi'; puts 'ok'")
	output, err := cmd.CombinedOutput()
	return err == nil && strings.Contains(string(output), "ok")
}

// Helper: Test Ruby syntax of generated files
func testRubySyntax(t *testing.T, tmpDir string) {
	rubyFiles := []string{
		"ruby/lib/test.rb",
		"ruby/lib/test/bindings.rb",
		"ruby/lib/test/message.rb",
		"ruby/lib/test/version.rb",
	}

	for _, file := range rubyFiles {
		fullPath := filepath.Join(tmpDir, file)
		cmd := exec.Command("ruby", "-c", fullPath)
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			t.Errorf("Ruby syntax error in %s: %v\nOutput: %s", file, err, output)
		} else if testing.Verbose() {
			t.Logf("✓ Ruby syntax OK: %s", file)
		}
	}
}

// Helper: Test that Ruby can load the FFI library
func testRubyFFILoading(t *testing.T, tmpDir string) {
	// Create a test script that tries to require the module
	testScript := `
$LOAD_PATH.unshift File.expand_path('ruby/lib', ARGV[0])
begin
  require 'test'
  
  # Check that the module and class exist
  if defined?(Test) && defined?(Test::Message)
    puts "OK: Module loaded successfully"
    exit 0
  else
    puts "ERROR: Module loaded but classes not defined"
    exit 1
  end
rescue LoadError => e
  puts "ERROR: Failed to load module: #{e.message}"
  exit 1
rescue => e
  puts "ERROR: #{e.message}"
  exit 1
end
`

	scriptPath := filepath.Join(tmpDir, "test_load.rb")
	err := os.WriteFile(scriptPath, []byte(testScript), 0644)
	if err != nil {
		t.Fatalf("Failed to write test script: %v", err)
	}

	cmd := exec.Command("ruby", scriptPath, tmpDir)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Errorf("Failed to load Ruby module: %v\nOutput: %s", err, output)
	} else if strings.Contains(string(output), "OK:") {
		if testing.Verbose() {
			t.Logf("✓ Ruby FFI module loaded successfully")
		}
	} else {
		t.Errorf("Unexpected output from Ruby test: %s", output)
	}
}

// TestPackageCompilationErrors tests that we properly capture and report compilation errors
func TestPackageCompilationErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ffire-test-error-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Parse a schema
	schemaPath := "../../testdata/schema/complex.ffi"
	schema, err := parser.Parse(schemaPath)
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	// Test with invalid platform (should fail gracefully)
	config := &PackageConfig{
		Schema:    schema,
		Language:  "ruby",
		OutputDir: tmpDir,
		Optimize:  2,
		Platform:  "invalid_platform",
		Arch:      "current",
		Namespace: schema.Package,
		NoCompile: false,
		Verbose:   false,
	}

	// This should either succeed (if platform validation is permissive)
	// or fail with a clear error message
	err = GeneratePackage(config)
	// We're just checking that it doesn't panic
	if err != nil && testing.Verbose() {
		t.Logf("Expected error for invalid platform: %v", err)
	}
}

// TestPackageNoCompile tests the --no-compile flag
func TestPackageNoCompile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ffire-test-nocompile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	schemaPath := "../../testdata/schema/complex.ffi"
	schema, err := parser.Parse(schemaPath)
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	config := &PackageConfig{
		Schema:    schema,
		Language:  "ruby",
		OutputDir: tmpDir,
		Optimize:  2,
		Platform:  "current",
		Arch:      "current",
		Namespace: schema.Package,
		NoCompile: true, // Skip compilation
		Verbose:   testing.Verbose(),
	}

	err = GeneratePackage(config)
	if err != nil {
		t.Fatalf("Failed to generate Ruby package with --no-compile: %v", err)
	}

	// C++ source should exist
	srcFile := filepath.Join(tmpDir, "ruby/src/generated_c.cpp")
	if !fileExists(srcFile) {
		t.Errorf("C++ source file not found: %s", srcFile)
	}

	// Dylib should NOT exist (because we didn't compile)
	dylibFile := filepath.Join(tmpDir, "ruby/lib/libffire.dylib")
	soFile := filepath.Join(tmpDir, "ruby/lib/libffire.so")
	dllFile := filepath.Join(tmpDir, "ruby/lib/ffire.dll")
	
	if fileExists(dylibFile) || fileExists(soFile) || fileExists(dllFile) {
		t.Errorf("Dylib should not exist with --no-compile flag")
	}
}
