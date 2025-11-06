package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

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

	// TODO: Implement Tier B package generation
	// This will include language-specific wrappers, package metadata, etc.

	return fmt.Errorf("Tier B package generation not yet implemented for %s", config.Language)
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
	// TODO: Implement cross-platform compilation
	// For now, just return placeholder
	fmt.Printf("TODO: Compile dylib for platform=%s arch=%s\n", config.Platform, config.Arch)

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
