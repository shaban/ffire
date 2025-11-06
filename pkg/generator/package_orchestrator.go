package generator

import (
	"fmt"
	"os"
	"path/filepath"
)

// PackagePaths holds all the directory paths for a package
type PackagePaths struct {
	Root     string // Root package directory (e.g., dist/python/)
	Lib      string // Library directory (where dylib goes)
	Include  string // C++ headers directory
	Src      string // C++ source directory
	Package  string // Language package directory (for Python, Ruby modules)
	Examples string // Examples directory
}

// DirectoryLayout defines the directory structure for a language package
type DirectoryLayout struct {
	Name         string   // Language name (python, javascript, ruby, etc.)
	LibInPackage bool     // Whether dylib goes in package dir (Python) or separate lib/ (JS, Ruby)
	ExtraDirs    []string // Any additional directories needed
}

// Common directory layouts for different languages
var (
	PythonLayout = DirectoryLayout{
		Name:         "python",
		LibInPackage: true, // Python: package/libffire.so
	}

	JavaScriptLayout = DirectoryLayout{
		Name:         "javascript",
		LibInPackage: false, // JS: lib/libffire.dylib
	}

	RubyLayout = DirectoryLayout{
		Name:         "ruby",
		LibInPackage: false, // Ruby: lib/libffire.dylib
	}

	SwiftLayout = DirectoryLayout{
		Name:         "swift",
		LibInPackage: false, // Swift: lib/libffire.dylib
	}

	PHPLayout = DirectoryLayout{
		Name:         "php",
		LibInPackage: false, // PHP: lib/libffire.so
	}

	JavaLayout = DirectoryLayout{
		Name:         "java",
		LibInPackage: false, // Java: lib/libffire.so
	}

	CSharpLayout = DirectoryLayout{
		Name:         "csharp",
		LibInPackage: false, // C#: lib/libffire.dll
	}
)

// setupPackageDirectories creates the standard directory structure
func setupPackageDirectories(config *PackageConfig, layout DirectoryLayout) (*PackagePaths, error) {
	paths := &PackagePaths{
		Root:     filepath.Join(config.OutputDir, layout.Name),
		Include:  filepath.Join(config.OutputDir, layout.Name, "include"),
		Src:      filepath.Join(config.OutputDir, layout.Name, "src"),
		Examples: filepath.Join(config.OutputDir, layout.Name, "examples"),
	}

	// Determine lib and package paths based on layout
	if layout.LibInPackage {
		// Python-style: dylib goes in package directory
		paths.Package = filepath.Join(paths.Root, config.Namespace)
		paths.Lib = paths.Package
	} else {
		// JS/Ruby-style: dylib goes in lib/ directory
		paths.Lib = filepath.Join(paths.Root, "lib")
		paths.Package = filepath.Join(paths.Root, "lib", config.Namespace)
	}

	// Create standard directories
	dirsToCreate := []string{
		paths.Root,
		paths.Include,
		paths.Src,
		paths.Lib,
	}

	// Add package directory if different from lib
	if paths.Package != paths.Lib {
		dirsToCreate = append(dirsToCreate, paths.Package)
	}

	// Add any extra directories from layout
	for _, extraDir := range layout.ExtraDirs {
		dirsToCreate = append(dirsToCreate, filepath.Join(paths.Root, extraDir))
	}

	// Create all directories
	for _, dir := range dirsToCreate {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return paths, nil
}

// generateNativeComponents generates C++ header, C ABI, and compiles dylib
func generateNativeComponents(config *PackageConfig, paths *PackagePaths) error {
	// Generate C++ header
	cppCode, err := GenerateCpp(config.Schema)
	if err != nil {
		return fmt.Errorf("failed to generate C++ code: %w", err)
	}

	headerPath := filepath.Join(paths.Include, "generated.hpp")
	if err := os.WriteFile(headerPath, cppCode, 0644); err != nil {
		return fmt.Errorf("failed to write C++ header: %w", err)
	}

	// Generate C ABI wrapper
	if err := generateCABI(config, paths.Include, paths.Src); err != nil {
		return fmt.Errorf("failed to generate C ABI: %w", err)
	}

	// Compile dylib (unless --no-compile)
	if !config.NoCompile {
		if err := compileDylib(config, paths.Src, paths.Lib); err != nil {
			return fmt.Errorf("failed to compile dylib: %w", err)
		}
	}

	return nil
}

// orchestrateTierBPackage is the common orchestration for all Tier B languages
func orchestrateTierBPackage(
	config *PackageConfig,
	layout DirectoryLayout,
	generateWrapper func(*PackageConfig, *PackagePaths) error,
	generateMetadata func(*PackageConfig, *PackagePaths) error,
	printInstructions func(*PackageConfig, *PackagePaths),
) error {
	if config.Verbose {
		fmt.Printf("Generating %s package\n", layout.Name)
	}

	// Setup directories
	paths, err := setupPackageDirectories(config, layout)
	if err != nil {
		return err
	}

	// Generate native components (C++, C ABI, dylib)
	if err := generateNativeComponents(config, paths); err != nil {
		return err
	}

	// Generate language-specific wrapper
	if err := generateWrapper(config, paths); err != nil {
		return fmt.Errorf("failed to generate %s wrapper: %w", layout.Name, err)
	}

	// Generate package metadata (setup.py, package.json, gemspec, etc.)
	if err := generateMetadata(config, paths); err != nil {
		return fmt.Errorf("failed to generate %s metadata: %w", layout.Name, err)
	}

	// Print installation instructions
	if printInstructions != nil {
		printInstructions(config, paths)
	}

	return nil
}
