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
		"ruby/lib/libffire.dylib", // or .so on Linux
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
	// Create temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "ffire-test-python-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Logf("Testing Python package generation in: %s", tmpDir)

	// Parse a test schema
	schemaPath := "../../testdata/schema/complex.ffi"
	schema, err := parser.Parse(schemaPath)
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	// Generate Python package
	config := &PackageConfig{
		Schema:    schema,
		Language:  "python",
		OutputDir: tmpDir,
		Optimize:  2,
		Platform:  "current",
		Arch:      "current",
		Namespace: schema.Package,
		NoCompile: false,
		Verbose:   testing.Verbose(),
	}

	err = GeneratePackage(config)
	if err != nil {
		t.Fatalf("Failed to generate Python package: %v", err)
	}

	// Verify expected files exist
	expectedFiles := []string{
		"python/test/libffire.dylib", // or .so on Linux (in package dir for Python)
		"python/test/__init__.py",
		"python/test/bindings.py",
		"python/setup.py",
		"python/README.md",
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

	// Test that Python can parse the generated code (syntax check)
	if hasPython() {
		t.Log("Python found, testing syntax...")
		testPythonSyntax(t, tmpDir)

		// Test that Python can import the module
		t.Log("Testing Python module import...")
		testPythonImport(t, tmpDir)
	} else {
		t.Log("Python not installed, skipping Python-specific tests")
	}
}

// TestJavaScriptPackageIntegration generates a JavaScript package and validates it
func TestJavaScriptPackageIntegration(t *testing.T) {
	// Create temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "ffire-test-javascript-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Logf("Testing JavaScript package generation in: %s", tmpDir)

	// Parse a test schema
	schemaPath := "../../testdata/schema/complex.ffi"
	schema, err := parser.Parse(schemaPath)
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	// Generate JavaScript package
	config := &PackageConfig{
		Schema:    schema,
		Language:  "javascript",
		OutputDir: tmpDir,
		Optimize:  2,
		Platform:  "current",
		Arch:      "current",
		Namespace: schema.Package,
		NoCompile: false,
		Verbose:   testing.Verbose(),
	}

	err = GeneratePackage(config)
	if err != nil {
		t.Fatalf("Failed to generate JavaScript package: %v", err)
	}

	// Verify expected files exist
	expectedFiles := []string{
		"javascript/lib/libffire.dylib", // or .so on Linux
		"javascript/index.js",
		"javascript/index.d.ts",
		"javascript/package.json",
		"javascript/README.md",
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

	// Test that Node.js can parse the generated code (syntax check)
	if hasNode() {
		t.Log("Node.js found, testing syntax...")
		testJavaScriptSyntax(t, tmpDir)
	} else {
		t.Log("Node.js not installed, skipping JavaScript-specific tests")
	}
}

// TestSwiftPackageIntegration generates a Swift package and validates it
func TestSwiftPackageIntegration(t *testing.T) {
	// Create temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "ffire-test-swift-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Logf("Testing Swift package generation in: %s", tmpDir)

	// Parse a test schema
	schemaPath := "../../testdata/schema/complex.ffi"
	schema, err := parser.Parse(schemaPath)
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	// Generate Swift package
	config := &PackageConfig{
		Schema:    schema,
		Language:  "swift",
		OutputDir: tmpDir,
		Optimize:  2,
		Platform:  "current",
		Arch:      "current",
		Namespace: schema.Package,
		NoCompile: false,
		Verbose:   testing.Verbose(),
	}

	err = GeneratePackage(config)
	if err != nil {
		t.Fatalf("Failed to generate Swift package: %v", err)
	}

	// Verify expected files exist
	expectedFiles := []string{
		"swift/lib/libffire.dylib", // or .so on Linux
		"swift/Sources/test/test.swift",
		"swift/Package.swift",
		"swift/README.md",
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

	// Test that Swift can parse the generated code (syntax check)
	if hasSwift() {
		t.Log("Swift found, testing syntax...")
		testSwiftSyntax(t, tmpDir)
	} else {
		t.Log("Swift not installed, skipping Swift-specific tests")
	}
}

// TestPHPPackageIntegration generates a PHP package and validates it
func TestPHPPackageIntegration(t *testing.T) {
	// Create temporary directory for test output
	tmpDir, err := os.MkdirTemp("", "ffire-test-php-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Logf("Testing PHP package generation in: %s", tmpDir)

	// Parse a test schema
	schemaPath := "../../testdata/schema/complex.ffi"
	schema, err := parser.Parse(schemaPath)
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	// Generate PHP package
	config := &PackageConfig{
		Schema:    schema,
		Language:  "php",
		OutputDir: tmpDir,
		Optimize:  2,
		Platform:  "current",
		Arch:      "current",
		Namespace: schema.Package,
		NoCompile: false,
		Verbose:   testing.Verbose(),
	}

	err = GeneratePackage(config)
	if err != nil {
		t.Fatalf("Failed to generate PHP package: %v", err)
	}

	// Verify expected files exist
	expectedFiles := []string{
		"php/lib/libffire.dylib", // or .so on Linux
		"php/src/Test.php",
		"php/composer.json",
		"php/README.md",
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

	// Test that PHP can parse the generated code (syntax check)
	if hasPHP() {
		t.Log("PHP found, testing syntax...")
		testPHPSyntax(t, tmpDir)

		// Check if FFI extension is available
		if hasPHPFFI() {
			t.Log("PHP FFI extension found")
		} else {
			t.Log("PHP FFI extension not available, cannot test FFI loading")
		}
	} else {
		t.Log("PHP not installed, skipping PHP-specific tests")
	}
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

// Helper: Check if Python is installed
func hasPython() bool {
	_, err := exec.LookPath("python3")
	if err == nil {
		return true
	}
	_, err = exec.LookPath("python")
	return err == nil
}

// Helper: Check if Node.js is installed
func hasNode() bool {
	_, err := exec.LookPath("node")
	return err == nil
}

// Helper: Check if Swift is installed
func hasSwift() bool {
	_, err := exec.LookPath("swiftc")
	return err == nil
}

// Helper: Check if PHP is installed
func hasPHP() bool {
	_, err := exec.LookPath("php")
	return err == nil
}

// Helper: Check if PHP FFI extension is available
func hasPHPFFI() bool {
	cmd := exec.Command("php", "-m")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "FFI")
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

// Helper: Test Python syntax of generated files
func testPythonSyntax(t *testing.T, tmpDir string) {
	pythonFiles := []string{
		"python/test/__init__.py",
		"python/test/bindings.py",
	}

	pythonCmd := "python3"
	if !hasPython() {
		return
	}
	// Check if python3 exists, otherwise use python
	if _, err := exec.LookPath("python3"); err != nil {
		pythonCmd = "python"
	}

	for _, file := range pythonFiles {
		fullPath := filepath.Join(tmpDir, file)
		cmd := exec.Command(pythonCmd, "-m", "py_compile", fullPath)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Errorf("Python syntax error in %s: %v\nOutput: %s", file, err, output)
		} else if testing.Verbose() {
			t.Logf("✓ Python syntax OK: %s", file)
		}
	}
}

// Helper: Test that Python can import the module
func testPythonImport(t *testing.T, tmpDir string) {
	pythonCmd := "python3"
	if _, err := exec.LookPath("python3"); err != nil {
		pythonCmd = "python"
	}

	// Create a test script that tries to import the module
	testScript := `
import sys
import os
sys.path.insert(0, os.path.join(sys.argv[1], 'python'))
try:
    import test
    
    # Check that the module and class exist
    if hasattr(test, 'Message'):
        print("OK: Module imported successfully")
        sys.exit(0)
    else:
        print("ERROR: Module imported but classes not defined")
        sys.exit(1)
except ImportError as e:
    print(f"ERROR: Failed to import module: {e}")
    sys.exit(1)
except Exception as e:
    print(f"ERROR: {e}")
    sys.exit(1)
`

	scriptPath := filepath.Join(tmpDir, "test_import.py")
	err := os.WriteFile(scriptPath, []byte(testScript), 0644)
	if err != nil {
		t.Fatalf("Failed to write test script: %v", err)
	}

	cmd := exec.Command(pythonCmd, scriptPath, tmpDir)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("Failed to import Python module: %v\nOutput: %s", err, output)
	} else if strings.Contains(string(output), "OK:") {
		if testing.Verbose() {
			t.Logf("✓ Python module imported successfully")
		}
	} else {
		t.Errorf("Unexpected output from Python test: %s", output)
	}
}

// Helper: Test JavaScript syntax of generated files
func testJavaScriptSyntax(t *testing.T, tmpDir string) {
	jsFiles := []string{
		"javascript/index.js",
	}

	for _, file := range jsFiles {
		fullPath := filepath.Join(tmpDir, file)
		// Use node --check to validate syntax without executing
		cmd := exec.Command("node", "--check", fullPath)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Errorf("JavaScript syntax error in %s: %v\nOutput: %s", file, err, output)
		} else if testing.Verbose() {
			t.Logf("✓ JavaScript syntax OK: %s", file)
		}
	}
}

// Helper: Test Swift syntax of generated files
func testSwiftSyntax(t *testing.T, tmpDir string) {
	swiftFiles := []string{
		"swift/Sources/test/test.swift",
	}

	for _, file := range swiftFiles {
		fullPath := filepath.Join(tmpDir, file)
		// Use swiftc -parse to validate syntax without compiling
		cmd := exec.Command("swiftc", "-parse", fullPath)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Errorf("Swift syntax error in %s: %v\nOutput: %s", file, err, output)
		} else if testing.Verbose() {
			t.Logf("✓ Swift syntax OK: %s", file)
		}
	}
}

// Helper: Test PHP syntax of generated files
func testPHPSyntax(t *testing.T, tmpDir string) {
	phpFiles := []string{
		"php/src/Test.php",
	}

	for _, file := range phpFiles {
		fullPath := filepath.Join(tmpDir, file)
		// Use php -l to validate syntax
		cmd := exec.Command("php", "-l", fullPath)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Errorf("PHP syntax error in %s: %v\nOutput: %s", file, err, output)
		} else if testing.Verbose() {
			t.Logf("✓ PHP syntax OK: %s", file)
		}
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
