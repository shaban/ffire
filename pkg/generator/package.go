package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shaban/ffire/pkg/schema"
)

// PackageConfig holds configuration for package generation
type PackageConfig struct {
	Schema    *schema.Schema
	Language  string
	OutputDir string
	Optimize  int
	Platform  string // "darwin", "linux", "windows", "current", "all"
	Arch      string // "arm64", "x86_64", "current", "all"
	Namespace string // Optional namespace/package name override
	NoCompile bool   // Skip dylib compilation
	Verbose   bool   // Verbose output
}

// GeneratePackage generates a complete production-ready package
func GeneratePackage(config *PackageConfig) error {
	if config.Verbose {
		fmt.Printf("Generating %s package for schema: %s\n", config.Language, config.Schema.Package)
	}

	// Set default namespace if not provided
	if config.Namespace == "" {
		config.Namespace = config.Schema.Package
	}

	// Resolve platform/arch if set to "current"
	if config.Platform == "current" {
		config.Platform = runtime.GOOS
	}
	if config.Arch == "current" {
		config.Arch = runtime.GOARCH
	}

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Determine package type (Tier A vs Tier B)
	switch config.Language {
	case "c", "cpp", "c++", "rust", "zig", "d", "nim", "crystal", "odin", "v":
		return generateTierAPackage(config)
	case "python", "javascript", "js", "node", "swift", "ruby", "java", "csharp", "cs", "c#", "go", "dart", "php", "perl", "lua", "r":
		return generateTierBPackage(config)
	default:
		return fmt.Errorf("unsupported language: %s", config.Language)
	}
}

// generateTierAPackage generates native code + C ABI (no wrapper layer)
func generateTierAPackage(config *PackageConfig) error {
	if config.Verbose {
		fmt.Println("Generating Tier A package (native code + C ABI)")
	}

	// Create directory structure
	langDir := filepath.Join(config.OutputDir, config.Language)
	includeDir := filepath.Join(langDir, "include")
	libDir := filepath.Join(langDir, "lib")
	srcDir := filepath.Join(langDir, "src")
	examplesDir := filepath.Join(langDir, "examples")

	for _, dir := range []string{includeDir, libDir, srcDir, examplesDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate C++ header
	cppCode, err := GenerateCpp(config.Schema)
	if err != nil {
		return fmt.Errorf("failed to generate C++ code: %w", err)
	}

	headerPath := filepath.Join(includeDir, "generated.hpp")
	if err := os.WriteFile(headerPath, cppCode, 0644); err != nil {
		return fmt.Errorf("failed to write C++ header: %w", err)
	}
	fmt.Printf("✓ Generated C++ code: %s\n", headerPath)

	// Generate C ABI wrapper
	if err := generateCABI(config, includeDir, srcDir); err != nil {
		return fmt.Errorf("failed to generate C ABI: %w", err)
	}

	// Compile dylib (unless --no-compile)
	if !config.NoCompile {
		if err := compileDylib(config, srcDir, libDir); err != nil {
			return fmt.Errorf("failed to compile dylib: %w", err)
		}
	}

	// Generate examples
	if err := generateExamples(config, examplesDir); err != nil {
		return fmt.Errorf("failed to generate examples: %w", err)
	}

	// Generate README
	if err := generateREADME(config, langDir); err != nil {
		return fmt.Errorf("failed to generate README: %w", err)
	}

	fmt.Printf("\n✅ Package ready at: %s\n", langDir)
	return nil
}

// generateTierBPackage generates complete package with language-specific wrapper
func generateTierBPackage(config *PackageConfig) error {
	if config.Verbose {
		fmt.Println("Generating Tier B package (with language wrapper)")
	}

	// Route to language-specific generator
	switch config.Language {
	case "python", "py":
		return generatePythonPackage(config)
	case "javascript", "js", "node", "nodejs":
		return generateJavaScriptPackage(config)
	case "swift":
		return generateSwiftPackage(config)
	case "ruby", "rb":
		return generateRubyPackage(config)
	default:
		return fmt.Errorf("Tier B package generation not yet implemented for %s", config.Language)
	}
}

// generateCABI generates C ABI wrapper files
func generateCABI(config *PackageConfig, includeDir, srcDir string) error {
	headerPath := filepath.Join(includeDir, "generated_c.h")
	implPath := filepath.Join(srcDir, "generated_c.cpp")

	// Generate C ABI header
	headerCode, err := GenerateCABIHeader(config.Schema)
	if err != nil {
		return fmt.Errorf("failed to generate C ABI header: %w", err)
	}

	if err := os.WriteFile(headerPath, headerCode, 0644); err != nil {
		return fmt.Errorf("failed to write C ABI header: %w", err)
	}
	fmt.Printf("✓ Generated C ABI header: %s\n", headerPath)

	// Generate C ABI implementation
	implCode, err := GenerateCABIImpl(config.Schema)
	if err != nil {
		return fmt.Errorf("failed to generate C ABI implementation: %w", err)
	}

	if err := os.WriteFile(implPath, implCode, 0644); err != nil {
		return fmt.Errorf("failed to write C ABI implementation: %w", err)
	}
	fmt.Printf("✓ Generated C ABI implementation: %s\n", implPath)

	return nil
}

// compileDylib compiles the C++ code into a dynamic library
func compileDylib(config *PackageConfig, srcDir, libDir string) error {
	if config.Verbose {
		fmt.Printf("Compiling dylib for platform=%s arch=%s optimize=%d\n",
			config.Platform, config.Arch, config.Optimize)
	}

	// Determine compiler and flags based on platform
	var compiler string
	var outputFile string
	var compileFlags []string

	switch config.Platform {
	case "darwin":
		compiler = "clang++"
		outputFile = filepath.Join(libDir, "libffire.dylib")
		compileFlags = []string{
			"-std=c++17",
			"-dynamiclib",
			"-fPIC",
			fmt.Sprintf("-O%d", config.Optimize),
			"-Wall",
			"-Wextra",
		}

		// Add architecture flag for macOS
		if config.Arch == "arm64" {
			compileFlags = append(compileFlags, "-arch", "arm64")
		} else if config.Arch == "x86_64" {
			compileFlags = append(compileFlags, "-arch", "x86_64")
		}

	case "linux":
		compiler = "g++"
		outputFile = filepath.Join(libDir, "libffire.so")
		compileFlags = []string{
			"-std=c++17",
			"-shared",
			"-fPIC",
			fmt.Sprintf("-O%d", config.Optimize),
			"-Wall",
			"-Wextra",
		}

	case "windows":
		compiler = "x86_64-w64-mingw32-g++"
		outputFile = filepath.Join(libDir, "ffire.dll")
		compileFlags = []string{
			"-std=c++17",
			"-shared",
			fmt.Sprintf("-O%d", config.Optimize),
			"-Wall",
			"-Wextra",
		}

	default:
		return fmt.Errorf("unsupported platform: %s", config.Platform)
	}

	// Build the command
	includeDir := filepath.Join(filepath.Dir(srcDir), "include")
	srcFile := filepath.Join(srcDir, "generated_c.cpp")

	// Convert to absolute paths
	absIncludeDir, err := filepath.Abs(includeDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for include dir: %w", err)
	}
	absSrcFile, err := filepath.Abs(srcFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for source file: %w", err)
	}
	absOutputFile, err := filepath.Abs(outputFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for output file: %w", err)
	}

	args := compileFlags
	args = append(args, "-I"+absIncludeDir)
	args = append(args, "-o", absOutputFile)
	args = append(args, absSrcFile)

	if config.Verbose {
		fmt.Printf("Running: %s %s\n", compiler, strings.Join(args, " "))
	}

	// Execute compilation
	cmd := exec.Command(compiler, args...)
	// Don't set cmd.Dir - we're using absolute paths

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compilation failed: %w\nOutput: %s", err, string(output))
	}

	if len(output) > 0 && config.Verbose {
		fmt.Printf("Compiler output:\n%s\n", string(output))
	}

	fmt.Printf("✓ Compiled dylib: %s\n", outputFile)
	return nil
}

// generateExamples generates example code
func generateExamples(config *PackageConfig, examplesDir string) error {
	// TODO: Generate language-specific examples
	fmt.Printf("TODO: Generate examples in %s\n", examplesDir)

	return nil
}

// generateREADME generates package README
func generateREADME(config *PackageConfig, langDir string) error {
	// TODO: Generate comprehensive README
	readmePath := filepath.Join(langDir, "README.md")
	fmt.Printf("TODO: Generate README at %s\n", readmePath)

	return nil
}

// generatePythonPackage generates a complete Python package with ctypes wrapper
func generatePythonPackage(config *PackageConfig) error {
	if config.Verbose {
		fmt.Println("Generating Python package")
	}

	// Create directory structure
	langDir := filepath.Join(config.OutputDir, "python")
	packageDir := filepath.Join(langDir, config.Namespace)

	for _, dir := range []string{langDir, packageDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate C++ code and C ABI (we need the dylib)
	includeDir := filepath.Join(langDir, "include")
	srcDir := filepath.Join(langDir, "src")
	libDir := filepath.Join(packageDir) // Put dylib in package dir

	for _, dir := range []string{includeDir, srcDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate C++ header
	cppCode, err := GenerateCpp(config.Schema)
	if err != nil {
		return fmt.Errorf("failed to generate C++ code: %w", err)
	}

	headerPath := filepath.Join(includeDir, "generated.hpp")
	if err := os.WriteFile(headerPath, cppCode, 0644); err != nil {
		return fmt.Errorf("failed to write C++ header: %w", err)
	}

	// Generate C ABI wrapper
	if err := generateCABI(config, includeDir, srcDir); err != nil {
		return fmt.Errorf("failed to generate C ABI: %w", err)
	}

	// Compile dylib
	if !config.NoCompile {
		if err := compileDylib(config, srcDir, libDir); err != nil {
			return fmt.Errorf("failed to compile dylib: %w", err)
		}
	}

	// Generate Python wrapper
	if err := generatePythonWrapper(config, packageDir); err != nil {
		return fmt.Errorf("failed to generate Python wrapper: %w", err)
	}

	// Generate setup.py
	if err := generatePythonSetup(config, langDir); err != nil {
		return fmt.Errorf("failed to generate setup.py: %w", err)
	}

	// Generate __init__.py
	if err := generatePythonInit(config, packageDir); err != nil {
		return fmt.Errorf("failed to generate __init__.py: %w", err)
	}

	// Generate README.md
	if err := generatePythonReadme(config, langDir); err != nil {
		return fmt.Errorf("failed to generate README.md: %w", err)
	}

	// Print installation instructions
	fmt.Printf("\n✅ Python package ready at: %s\n\n", langDir)
	fmt.Println("Installation:")
	fmt.Printf("  cd %s\n\n", langDir)
	fmt.Println("  # Recommended: Use a virtual environment")
	fmt.Println("  python3 -m venv venv")
	fmt.Println("  source venv/bin/activate  # On Windows: venv\\Scripts\\activate")
	fmt.Println("  pip install .")
	fmt.Println()
	fmt.Println("  # Alternative: Install for current user only")
	fmt.Println("  pip install --user .")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  from %s import Message\n", config.Namespace)
	fmt.Println("  msg = Message.decode(data)")
	fmt.Println("  encoded = msg.encode()")
	fmt.Println()

	return nil
}

func generateJavaScriptPackage(config *PackageConfig) error {
	if config.Verbose {
		fmt.Println("Generating JavaScript/Node.js package")
	}

	// Create directory structure
	langDir := filepath.Join(config.OutputDir, "javascript")
	libDir := filepath.Join(langDir, "lib")

	for _, dir := range []string{langDir, libDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate C++ code and C ABI (we need the dylib)
	includeDir := filepath.Join(langDir, "include")
	srcDir := filepath.Join(langDir, "src")

	for _, dir := range []string{includeDir, srcDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate C++ header
	cppCode, err := GenerateCpp(config.Schema)
	if err != nil {
		return fmt.Errorf("failed to generate C++ code: %w", err)
	}

	headerPath := filepath.Join(includeDir, "generated.hpp")
	if err := os.WriteFile(headerPath, cppCode, 0644); err != nil {
		return fmt.Errorf("failed to write C++ header: %w", err)
	}

	// Generate C ABI wrapper
	if err := generateCABI(config, includeDir, srcDir); err != nil {
		return fmt.Errorf("failed to generate C ABI: %w", err)
	}

	// Compile dylib
	if !config.NoCompile {
		if err := compileDylib(config, srcDir, libDir); err != nil {
			return fmt.Errorf("failed to compile dylib: %w", err)
		}
	}

	// Generate JavaScript wrapper
	if err := generateJavaScriptWrapper(config, langDir); err != nil {
		return fmt.Errorf("failed to generate JavaScript wrapper: %w", err)
	}

	// Generate TypeScript definitions
	if err := generateTypeScriptDefinitions(config, langDir); err != nil {
		return fmt.Errorf("failed to generate TypeScript definitions: %w", err)
	}

	// Generate package.json
	if err := generateJavaScriptPackageJson(config, langDir); err != nil {
		return fmt.Errorf("failed to generate package.json: %w", err)
	}

	// Generate README.md
	if err := generateJavaScriptReadme(config, langDir); err != nil {
		return fmt.Errorf("failed to generate README.md: %w", err)
	}

	// Print installation instructions
	fmt.Printf("\n✅ JavaScript/Node.js package ready at: %s\n\n", langDir)
	fmt.Println("Installation:")
	fmt.Printf("  cd %s\n", langDir)
	fmt.Println("  npm install")
	fmt.Println()
	fmt.Println("Usage (JavaScript):")
	fmt.Printf("  const { Message } = require('%s');\n", config.Namespace)
	fmt.Println("  const msg = Message.decode(data);")
	fmt.Println("  const encoded = msg.encode();")
	fmt.Println("  msg.free();")
	fmt.Println()
	fmt.Println("Usage (TypeScript):")
	fmt.Printf("  import { Message } from '%s';\n", config.Namespace)
	fmt.Println("  const msg: Message = Message.decode(data);")
	fmt.Println("  const encoded: Buffer = msg.encode();")
	fmt.Println()

	return nil
}

func generateSwiftPackage(config *PackageConfig) error {
	return fmt.Errorf("Swift package generation not yet implemented")
}

func generateRubyPackage(config *PackageConfig) error {
	return GenerateRubyPackage(config)
}
