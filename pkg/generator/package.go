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

	// Handle Go as Tier 0 (native reference implementation)
	if config.Language == "go" {
		return generateGoPackage(config)
	}

	// Determine package type (Tier A vs Tier B)
	switch config.Language {
	case "c", "cpp", "c++", "rust", "zig", "d", "nim", "crystal", "odin", "v":
		return generateTierAPackage(config)
	case "python", "javascript", "js", "node", "swift", "ruby", "java", "csharp", "cs", "c#", "dart", "php", "perl", "lua", "r":
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
	case "php":
		return generatePHPPackage(config)
	case "java":
		return generateJavaPackage(config)
	case "csharp":
		return generateCSharpPackage(config)
	case "dart":
		return generateDartPackage(config)
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
		outputFile = filepath.Join(libDir, fmt.Sprintf("lib%s.dylib", config.Schema.Package))
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
		outputFile = filepath.Join(libDir, fmt.Sprintf("lib%s.so", config.Schema.Package))
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
		outputFile = filepath.Join(libDir, fmt.Sprintf("%s.dll", config.Schema.Package))
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

// generateGoPackage generates a native Go package (Tier 0 reference implementation)
func generateGoPackage(config *PackageConfig) error {
	if config.Verbose {
		fmt.Println("Generating Go package (native implementation)")
	}

	// Generate Go code for all message types
	code, err := GenerateGo(config.Schema)
	if err != nil {
		return fmt.Errorf("failed to generate Go code: %w", err)
	}

	// Write to output file
	outputPath := filepath.Join(config.OutputDir, config.Namespace+".go")
	if err := os.WriteFile(outputPath, code, 0644); err != nil {
		return fmt.Errorf("failed to write Go code: %w", err)
	}

	fmt.Printf("✓ Generated Go package: %s\n", outputPath)
	return nil
}

// generatePythonPackage generates a complete Python package with pybind11 wrapper
func generatePythonPackage(config *PackageConfig) error {
	return GeneratePythonPybind11Package(config)
}

func generateJavaScriptPackage(config *PackageConfig) error {
	return GenerateJavaScriptPackage(config)
}

func generateSwiftPackage(config *PackageConfig) error {
	return GenerateSwiftPackage(config)
}

func generateRubyPackage(config *PackageConfig) error {
	return GenerateRubyPackage(config)
}

func generatePHPPackage(config *PackageConfig) error {
	return GeneratePHPPackage(config)
}

func generateJavaPackage(config *PackageConfig) error {
	return GenerateJavaPackage(config)
}

func generateCSharpPackage(config *PackageConfig) error {
	return GenerateCSharpPackage(config)
}

func generateDartPackage(config *PackageConfig) error {
	return GenerateDartPackage(config)
}
