package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shaban/ffire/pkg/schema"
)

// GeneratePHPPackage generates a complete PHP Composer package using the orchestrator
func GeneratePHPPackage(config *PackageConfig) error {
	return orchestrateTierBPackage(
		config,
		PHPLayout,
		generatePHPWrapperOrchestrated,
		generatePHPMetadataOrchestrated,
		printPHPInstructions,
	)
}

func generatePHPWrapperOrchestrated(config *PackageConfig, paths *PackagePaths) error {
	// Create src directory structure
	srcDir := filepath.Join(paths.Root, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return fmt.Errorf("failed to create src directory: %w", err)
	}

	// Generate PHP wrapper
	if err := generatePHPWrapper(config, srcDir, paths.Lib); err != nil {
		return err
	}

	return nil
}

func generatePHPMetadataOrchestrated(config *PackageConfig, paths *PackagePaths) error {
	// Generate composer.json
	if err := generatePHPComposerJson(config, paths.Root); err != nil {
		return err
	}

	// Generate README.md
	if err := generatePHPReadme(config, paths.Root); err != nil {
		return err
	}

	return nil
}

func printPHPInstructions(config *PackageConfig, paths *PackagePaths) {
	fmt.Printf("\n✅ PHP Composer package ready at: %s\n\n", paths.Root)
	fmt.Println("Installation:")
	fmt.Printf("  cd %s\n", paths.Root)
	fmt.Println("  composer install")
	fmt.Println()
	fmt.Println("Requirements:")
	fmt.Println("  - PHP 7.4 or higher (FFI extension required)")
	fmt.Println("  - Enable FFI in php.ini: ffi.enable=true")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  require_once 'vendor/autoload.php';\n")
	fmt.Printf("  use %s\\Message;\n", ToPascalCase(config.Namespace))
	fmt.Println()
	fmt.Println("  $data = file_get_contents('data.bin');")
	fmt.Println("  try {")
	fmt.Println("      $msg = Message::decode($data);")
	fmt.Println("      $encoded = $msg->encode();")
	fmt.Println("  } catch (Exception $e) {")
	fmt.Println("      echo \"Error: \" . $e->getMessage();")
	fmt.Println("  }")
	fmt.Println()
}

// generatePHPWrapper generates the PHP wrapper using FFI
func generatePHPWrapper(config *PackageConfig, srcDir string, libDir string) error {
	buf := &bytes.Buffer{}

	namespace := ToPascalCase(config.Namespace)

	// PHP header
	buf.WriteString("<?php\n\n")
	fmt.Fprintf(buf, "namespace %s;\n\n", namespace)

	// Use statements
	buf.WriteString("use FFI;\n")
	buf.WriteString("use FFI\\CData;\n")
	buf.WriteString("use Exception;\n\n")

	// Exception class
	buf.WriteString("/**\n")
	buf.WriteString(" * FFire exception\n")
	buf.WriteString(" */\n")
	buf.WriteString("class FFireException extends Exception {}\n\n")

	// FFI Loader class
	buf.WriteString("/**\n")
	buf.WriteString(" * FFI Library loader\n")
	buf.WriteString(" */\n")
	buf.WriteString("class FFILibrary {\n")
	buf.WriteString("    private static ?FFI $ffi = null;\n\n")
	buf.WriteString("    public static function load(): FFI {\n")
	buf.WriteString("        if (self::$ffi !== null) {\n")
	buf.WriteString("            return self::$ffi;\n")
	buf.WriteString("        }\n\n")
	buf.WriteString("        // Determine library path based on OS\n")
	buf.WriteString("        $libName = match(PHP_OS_FAMILY) {\n")
	fmt.Fprintf(buf, "            'Darwin' => 'lib%s.dylib',\n", config.Schema.Package)
	fmt.Fprintf(buf, "            'Linux' => 'lib%s.so',\n", config.Schema.Package)
	fmt.Fprintf(buf, "            'Windows' => '%s.dll',\n", config.Schema.Package)
	fmt.Fprintf(buf, "            default => 'lib%s.so'\n", config.Schema.Package)
	buf.WriteString("        };\n\n")
	buf.WriteString("        $libPath = __DIR__ . '/../lib/' . $libName;\n\n")
	buf.WriteString("        if (!file_exists($libPath)) {\n")
	buf.WriteString("            throw new FFireException(\"Library not found: {$libPath}\");\n")
	buf.WriteString("        }\n\n")

	// Build FFI declarations
	buf.WriteString("        $ffiCode = \"\n")

	for _, msg := range config.Schema.Messages {
		baseName := strings.ToLower(msg.Name) // All lowercase to match C ABI
		fmt.Fprintf(buf, "            void* %s_decode(const uint8_t* data, size_t size, char** error);\n", baseName)
		fmt.Fprintf(buf, "            size_t %s_encode(void* handle, uint8_t** data, char** error);\n", baseName)
		fmt.Fprintf(buf, "            void %s_free(void* handle);\n", baseName)
		fmt.Fprintf(buf, "            void %s_free_data(uint8_t* data);\n", baseName)
		fmt.Fprintf(buf, "            void %s_free_error(char* error);\n", baseName)
	}

	buf.WriteString("        \";\n\n")
	buf.WriteString("        self::$ffi = FFI::cdef($ffiCode, $libPath);\n")
	buf.WriteString("        return self::$ffi;\n")
	buf.WriteString("    }\n")
	buf.WriteString("}\n\n")

	// Generate class for each message type
	for _, msg := range config.Schema.Messages {
		if err := generatePHPMessageClass(buf, namespace, &msg); err != nil {
			return err
		}
	}

	// Write to file
	wrapperPath := filepath.Join(srcDir, namespace+".php")
	if err := os.WriteFile(wrapperPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write PHP wrapper: %w", err)
	}

	fmt.Printf("✓ Generated PHP bindings: %s\n", wrapperPath)
	return nil
}

func generatePHPMessageClass(buf *bytes.Buffer, namespace string, msg *schema.MessageType) error {
	className := msg.Name
	baseName := strings.ToLower(msg.Name) // All lowercase to match C ABI

	buf.WriteString("/**\n")
	fmt.Fprintf(buf, " * %s message type\n", className)
	buf.WriteString(" */\n")
	fmt.Fprintf(buf, "class %s {\n", className)
	buf.WriteString("    private CData $handle;\n")
	buf.WriteString("    private FFI $ffi;\n")
	buf.WriteString("    private bool $freed = false;\n\n")

	// Private constructor
	buf.WriteString("    private function __construct(CData $handle, FFI $ffi) {\n")
	buf.WriteString("        $this->handle = $handle;\n")
	buf.WriteString("        $this->ffi = $ffi;\n")
	buf.WriteString("    }\n\n")

	// Destructor
	buf.WriteString("    public function __destruct() {\n")
	buf.WriteString("        $this->free();\n")
	buf.WriteString("    }\n\n")

	// Decode method
	buf.WriteString("    /**\n")
	fmt.Fprintf(buf, "     * Decode a %s from binary data\n", className)
	buf.WriteString("     * @param string $data Binary data to decode\n")
	fmt.Fprintf(buf, "     * @return %s Decoded message object\n", className)
	buf.WriteString("     * @throws FFireException if decoding fails\n")
	buf.WriteString("     */\n")
	buf.WriteString("    public static function decode(string $data): self {\n")
	buf.WriteString("        $ffi = FFILibrary::load();\n")
	buf.WriteString("        $size = strlen($data);\n\n")
	buf.WriteString("        // Allocate memory for data\n")
	buf.WriteString("        $dataPtr = $ffi->new('uint8_t[' . $size . ']', false);\n")
	buf.WriteString("        FFI::memcpy($dataPtr, $data, $size);\n\n")
	buf.WriteString("        // Allocate error pointer\n")
	buf.WriteString("        $errorPtr = $ffi->new('char*');\n")
	buf.WriteString("        $errorPtr->cdata = null;\n\n")
	fmt.Fprintf(buf, "        $handle = $ffi->%s_decode($dataPtr, $size, FFI::addr($errorPtr));\n\n", baseName)
	buf.WriteString("        if (FFI::isNull($handle)) {\n")
	buf.WriteString("            $error = 'Unknown error';\n")
	buf.WriteString("            if (!FFI::isNull($errorPtr->cdata)) {\n")
	buf.WriteString("                $error = FFI::string($errorPtr->cdata);\n")
	fmt.Fprintf(buf, "                $ffi->%s_free_error($errorPtr->cdata);\n", baseName)
	buf.WriteString("            }\n")
	fmt.Fprintf(buf, "            throw new FFireException(\"Failed to decode %s: {$error}\");\n", className)
	buf.WriteString("        }\n\n")
	buf.WriteString("        return new self($handle, $ffi);\n")
	buf.WriteString("    }\n\n")

	// Encode method
	buf.WriteString("    /**\n")
	fmt.Fprintf(buf, "     * Encode this %s to binary data\n", className)
	buf.WriteString("     * @return string Encoded binary data\n")
	buf.WriteString("     * @throws FFireException if encoding fails\n")
	buf.WriteString("     */\n")
	buf.WriteString("    public function encode(): string {\n")
	buf.WriteString("        if ($this->freed) {\n")
	fmt.Fprintf(buf, "            throw new FFireException('%s already freed');\n", className)
	buf.WriteString("        }\n\n")
	buf.WriteString("        $dataPtr = $this->ffi->new('uint8_t*');\n")
	buf.WriteString("        $dataPtr->cdata = null;\n")
	buf.WriteString("        $errorPtr = $this->ffi->new('char*');\n")
	buf.WriteString("        $errorPtr->cdata = null;\n\n")
	fmt.Fprintf(buf, "        $size = $this->ffi->%s_encode($this->handle, FFI::addr($dataPtr), FFI::addr($errorPtr));\n\n", baseName)
	buf.WriteString("        if ($size === 0) {\n")
	buf.WriteString("            $error = 'Unknown error';\n")
	buf.WriteString("            if (!FFI::isNull($errorPtr->cdata)) {\n")
	buf.WriteString("                $error = FFI::string($errorPtr->cdata);\n")
	fmt.Fprintf(buf, "                $this->ffi->%s_free_error($errorPtr->cdata);\n", baseName)
	buf.WriteString("            }\n")
	fmt.Fprintf(buf, "            throw new FFireException(\"Failed to encode %s: {$error}\");\n", className)
	buf.WriteString("        }\n\n")
	buf.WriteString("        $result = FFI::string($dataPtr->cdata, $size);\n")
	fmt.Fprintf(buf, "        $this->ffi->%s_free_data($dataPtr->cdata);\n\n", baseName)
	buf.WriteString("        return $result;\n")
	buf.WriteString("    }\n\n")

	// Free method
	buf.WriteString("    /**\n")
	buf.WriteString("     * Free the native resources\n")
	buf.WriteString("     */\n")
	buf.WriteString("    public function free(): void {\n")
	buf.WriteString("        if (!$this->freed) {\n")
	fmt.Fprintf(buf, "            $this->ffi->%s_free($this->handle);\n", baseName)
	buf.WriteString("            $this->freed = true;\n")
	buf.WriteString("        }\n")
	buf.WriteString("    }\n")
	buf.WriteString("}\n\n")

	return nil
}

// generatePHPComposerJson generates composer.json
func generatePHPComposerJson(config *PackageConfig, packageDir string) error {
	buf := &bytes.Buffer{}

	namespace := ToPascalCase(config.Namespace)

	fmt.Fprintf(buf, `{
    "name": "%s/%s",
    "description": "FFire binary serialization library - %s schema",
    "type": "library",
    "require": {
        "php": ">=7.4",
        "ext-ffi": "*"
    },
    "autoload": {
        "psr-4": {
            "%s\\": "src/"
        }
    },
    "keywords": [
        "ffire",
        "serialization",
        "binary",
        "codec",
        "ffi"
    ],
    "license": "SEE LICENSE FILE",
    "authors": [
        {
            "name": "Generated by FFire"
        }
    ]
}
`, strings.ToLower(config.Namespace), strings.ToLower(config.Namespace),
		config.Schema.Package, namespace)

	composerPath := filepath.Join(packageDir, "composer.json")
	if err := os.WriteFile(composerPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write composer.json: %w", err)
	}

	fmt.Printf("✓ Generated composer.json: %s\n", composerPath)
	return nil
}

// generatePHPReadme generates README.md
func generatePHPReadme(config *PackageConfig, packageDir string) error {
	buf := &bytes.Buffer{}

	namespace := ToPascalCase(config.Namespace)

	fmt.Fprintf(buf, `# %s - FFire PHP Bindings

PHP bindings for the %s schema, generated by [FFire](https://github.com/shaban/ffire).

## Requirements

- PHP 7.4 or higher
- FFI extension enabled
- The FFI extension must be enabled in php.ini:

`, config.Namespace, config.Schema.Package)

	buf.WriteString("```ini\n")
	buf.WriteString("ffi.enable=true\n")
	buf.WriteString("```\n\n")

	buf.WriteString("## Installation\n\n")
	buf.WriteString("### Using Composer\n\n")
	buf.WriteString("```bash\n")
	fmt.Fprintf(buf, "cd %s\n", filepath.Base(packageDir))
	buf.WriteString("composer install\n")
	buf.WriteString("```\n\n")

	buf.WriteString("## Usage\n\n")
	buf.WriteString("```php\n")
	buf.WriteString("<?php\n\n")
	buf.WriteString("require_once 'vendor/autoload.php';\n\n")
	fmt.Fprintf(buf, "use %s\\Message;\n", namespace)
	fmt.Fprintf(buf, "use %s\\FFireException;\n\n", namespace)
	buf.WriteString("// Decode from binary\n")
	buf.WriteString("$data = file_get_contents('data.bin');\n\n")
	buf.WriteString("try {\n")
	buf.WriteString("    $msg = Message::decode($data);\n")
	buf.WriteString("    \n")
	buf.WriteString("    // Encode back to binary\n")
	buf.WriteString("    $encoded = $msg->encode();\n")
	buf.WriteString("    \n")
	buf.WriteString("    // Resources are automatically freed when $msg goes out of scope\n")
	buf.WriteString("    // Or manually call: $msg->free();\n")
	buf.WriteString("} catch (FFireException $e) {\n")
	buf.WriteString("    echo \"Error: \" . $e->getMessage() . \"\\n\";\n")
	buf.WriteString("}\n")
	buf.WriteString("```\n\n")

	buf.WriteString("## API\n\n")

	for _, msg := range config.Schema.Messages {
		fmt.Fprintf(buf, "### `%s`\n\n", msg.Name)
		fmt.Fprintf(buf, "**`static function decode(string $data): %s`**\n\n", msg.Name)
		fmt.Fprintf(buf, "Decode a `%s` from binary data.\n\n", msg.Name)
		buf.WriteString("- **Parameter:** `$data` - Binary data (string)\n")
		fmt.Fprintf(buf, "- **Returns:** `%s` object\n", msg.Name)
		buf.WriteString("- **Throws:** `FFireException` if decoding fails\n\n")

		fmt.Fprintf(buf, "**`function encode(): string`**\n\n")
		fmt.Fprintf(buf, "Encode this `%s` to binary data.\n\n", msg.Name)
		buf.WriteString("- **Returns:** Binary data (string)\n")
		buf.WriteString("- **Throws:** `FFireException` if encoding fails\n\n")

		buf.WriteString("**`function free(): void`**\n\n")
		buf.WriteString("Manually free the native resources. Called automatically by destructor.\n\n")
	}

	buf.WriteString("## Error Handling\n\n")
	buf.WriteString("The library throws `FFireException` for all errors:\n\n")
	buf.WriteString("```php\n")
	buf.WriteString("try {\n")
	buf.WriteString("    $msg = Message::decode($invalidData);\n")
	buf.WriteString("} catch (FFireException $e) {\n")
	buf.WriteString("    echo \"Decoding failed: \" . $e->getMessage();\n")
	buf.WriteString("}\n")
	buf.WriteString("```\n\n")

	buf.WriteString("## Platform Support\n\n")
	buf.WriteString("This package includes pre-compiled libraries for:\n\n")
	fmt.Fprintf(buf, "- macOS (Darwin): `lib%s.dylib`\n", config.Schema.Package)
	fmt.Fprintf(buf, "- Linux: `lib%s.so`\n", config.Schema.Package)
	fmt.Fprintf(buf, "- Windows: `%s.dll`\n\n", config.Schema.Package)
	buf.WriteString("The correct library is automatically loaded based on your platform.\n\n")

	buf.WriteString("## Memory Management\n\n")
	buf.WriteString("Resources are automatically freed when objects are garbage collected. You can also manually call `free()` if needed for immediate cleanup.\n\n")

	buf.WriteString("## FFI Extension\n\n")
	buf.WriteString("This library requires PHP's FFI extension, which provides a way to call C functions from PHP. Make sure it's enabled in your php.ini:\n\n")
	buf.WriteString("```bash\n")
	buf.WriteString("php -m | grep FFI\n")
	buf.WriteString("```\n\n")
	buf.WriteString("If FFI is not listed, enable it in php.ini and restart PHP.\n\n")

	buf.WriteString("## License\n\n")
	buf.WriteString("Generated by FFire. See your schema's license for terms.\n")

	readmePath := filepath.Join(packageDir, "README.md")
	if err := os.WriteFile(readmePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	fmt.Printf("✓ Generated README.md: %s\n", readmePath)
	return nil
}
