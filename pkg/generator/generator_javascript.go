package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shaban/ffire/pkg/schema"
)

// GenerateJavaScriptPackage generates a complete JavaScript/Node.js package using the orchestrator
func GenerateJavaScriptPackage(config *PackageConfig) error {
	return orchestrateTierBPackage(
		config,
		JavaScriptLayout,
		generateJavaScriptWrapperOrchestrated,
		generateJavaScriptMetadataOrchestrated,
		printJavaScriptInstructions,
	)
}

func generateJavaScriptWrapperOrchestrated(config *PackageConfig, paths *PackagePaths) error {
	// Generate JavaScript wrapper
	if err := generateJavaScriptWrapper(config, paths.Root); err != nil {
		return err
	}

	// Generate TypeScript definitions
	if err := generateTypeScriptDefinitions(config, paths.Root); err != nil {
		return err
	}

	return nil
}

func generateJavaScriptMetadataOrchestrated(config *PackageConfig, paths *PackagePaths) error {
	// Generate package.json
	if err := generateJavaScriptPackageJson(config, paths.Root); err != nil {
		return err
	}

	// Generate README.md
	if err := generateJavaScriptReadme(config, paths.Root); err != nil {
		return err
	}

	return nil
}

func printJavaScriptInstructions(config *PackageConfig, paths *PackagePaths) {
	fmt.Printf("\n✅ JavaScript/Node.js package ready at: %s\n\n", paths.Root)
	fmt.Println("Installation:")
	fmt.Printf("  cd %s\n", paths.Root)
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
}

// generateJavaScriptWrapper generates the FFI wrapper with JSDoc
func generateJavaScriptWrapper(config *PackageConfig, packageDir string) error {
	buf := &bytes.Buffer{}

	// Header comment
	fmt.Fprintf(buf, `/**
 * FFire %s bindings using ffi-napi
 * 
 * This module provides Node.js bindings to the FFire binary serialization library
 * via a C ABI dynamic library.
 * 
 * @module %s
 */

const ffi = require('ffi-napi');
const ref = require('ref-napi');
const path = require('path');
const os = require('os');

`, config.Schema.Package, config.Namespace)

	// Platform detection
	buf.WriteString("// Determine library name based on platform\n")
	buf.WriteString("const libName = (() => {\n")
	buf.WriteString("  const platform = os.platform();\n")
	buf.WriteString("  switch (platform) {\n")
	buf.WriteString("    case 'darwin': return 'libffire.dylib';\n")
	buf.WriteString("    case 'linux': return 'libffire.so';\n")
	buf.WriteString("    case 'win32': return 'ffire.dll';\n")
	buf.WriteString("    default: return 'libffire.so';\n")
	buf.WriteString("  }\n")
	buf.WriteString("})();\n\n")

	// Load library
	buf.WriteString("// Load the C library\n")
	buf.WriteString("const libPath = path.join(__dirname, 'lib', libName);\n\n")

	// Generate FFI bindings for each message type
	for _, msg := range config.Schema.Messages {
		if err := generateJavaScriptMessageBindings(buf, config.Schema, &msg); err != nil {
			return err
		}
	}

	// Write to file
	wrapperPath := filepath.Join(packageDir, "index.js")
	if err := os.WriteFile(wrapperPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write JavaScript wrapper: %w", err)
	}

	fmt.Printf("✓ Generated JavaScript bindings: %s\n", wrapperPath)
	return nil
}

func generateJavaScriptMessageBindings(buf *bytes.Buffer, s *schema.Schema, msg *schema.MessageType) error {
	baseName := strings.ToLower(msg.Name[:1]) + msg.Name[1:]
	className := msg.Name

	// Define FFI library interface
	buf.WriteString("// FFI library interface\n")
	fmt.Fprintf(buf, "const lib_%s = ffi.Library(libPath, {\n", baseName)
	fmt.Fprintf(buf, "  '%s_decode': ['pointer', ['pointer', 'size_t', 'pointer']],\n", baseName)
	fmt.Fprintf(buf, "  '%s_encode': ['size_t', ['pointer', 'pointer', 'pointer']],\n", baseName)
	fmt.Fprintf(buf, "  '%s_free': ['void', ['pointer']],\n", baseName)
	fmt.Fprintf(buf, "  '%s_free_data': ['void', ['pointer']],\n", baseName)
	fmt.Fprintf(buf, "  '%s_free_error': ['void', ['pointer']]\n", baseName)
	buf.WriteString("});\n\n")

	// Generate class with JSDoc
	fmt.Fprintf(buf, `/**
 * %s wrapper class
 */
class %s {
  /**
   * @private
   * @param {Buffer} handle - Native handle pointer
   */
  constructor(handle) {
    this._handle = handle;
    this._freed = false;
  }

  /**
   * Decode a %s from binary data
   * @param {Buffer} data - Binary data to decode
   * @returns {%s} Decoded message object
   * @throws {Error} If decoding fails
   */
  static decode(data) {
    if (!Buffer.isBuffer(data)) {
      throw new TypeError('data must be a Buffer');
    }

    const errorPtr = ref.alloc(ref.refType(ref.types.CString));
    const handle = lib_%s.%s_decode(
      data,
      data.length,
      errorPtr
    );

    if (handle.isNull()) {
      const errorMsg = errorPtr.deref();
      const error = errorMsg.isNull() ? 'Unknown error' : errorMsg.readCString();
      if (!errorMsg.isNull()) {
        lib_%s.%s_free_error(errorMsg);
      }
      throw new Error('Failed to decode %s: ' + error);
    }

    return new %s(handle);
  }

  /**
   * Encode this %s to binary data
   * @returns {Buffer} Encoded binary data
   * @throws {Error} If encoding fails
   */
  encode() {
    if (this._freed) {
      throw new Error('%s already freed');
    }

    const dataPtr = ref.alloc(ref.refType(ref.types.uint8));
    const errorPtr = ref.alloc(ref.refType(ref.types.CString));

    const size = lib_%s.%s_encode(
      this._handle,
      dataPtr,
      errorPtr
    );

    if (size === 0) {
      const errorMsg = errorPtr.deref();
      const error = errorMsg.isNull() ? 'Unknown error' : errorMsg.readCString();
      if (!errorMsg.isNull()) {
        lib_%s.%s_free_error(errorMsg);
      }
      throw new Error('Failed to encode %s: ' + error);
    }

    const data = dataPtr.deref();
    const result = Buffer.from(ref.reinterpret(data, size, 0));
    lib_%s.%s_free_data(data);

    return result;
  }

  /**
   * Free the native resources
   * @returns {void}
   */
  free() {
    if (!this._freed && !this._handle.isNull()) {
      lib_%s.%s_free(this._handle);
      this._freed = true;
    }
  }
}

`, className, className, className, className, baseName, baseName, baseName, baseName, className, className, className, className, baseName, baseName, baseName, baseName, className, baseName, baseName, baseName, baseName)

	// Export
	fmt.Fprintf(buf, "module.exports = { %s };\n\n", className)

	return nil
}

// generateTypeScriptDefinitions generates .d.ts file
func generateTypeScriptDefinitions(config *PackageConfig, packageDir string) error {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, `/**
 * FFire %s TypeScript definitions
 */

`, config.Schema.Package)

	// Generate class definitions
	for _, msg := range config.Schema.Messages {
		className := msg.Name

		fmt.Fprintf(buf, `/**
 * %s message type
 */
export class %s {
  /**
   * Decode a %s from binary data
   * @param data Binary data to decode
   * @returns Decoded message object
   * @throws Error if decoding fails
   */
  static decode(data: Buffer): %s;

  /**
   * Encode this %s to binary data
   * @returns Encoded binary data
   * @throws Error if encoding fails
   */
  encode(): Buffer;

  /**
   * Free the native resources
   */
  free(): void;
}

`, className, className, className, className, className)
	}

	// Write to file
	defsPath := filepath.Join(packageDir, "index.d.ts")
	if err := os.WriteFile(defsPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write TypeScript definitions: %w", err)
	}

	fmt.Printf("✓ Generated TypeScript definitions: %s\n", defsPath)
	return nil
}

// generateJavaScriptPackageJson generates package.json
func generateJavaScriptPackageJson(config *PackageConfig, packageDir string) error {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, `{
  "name": "%s",
  "version": "1.0.0",
  "description": "FFire binary serialization library - %s schema",
  "main": "index.js",
  "types": "index.d.ts",
  "scripts": {
    "test": "node test.js"
  },
  "keywords": [
    "ffire",
    "serialization",
    "binary",
    "codec"
  ],
  "author": "Generated by FFire",
  "license": "SEE LICENSE IN LICENSE",
  "dependencies": {
    "ffi-napi": "^4.0.3",
    "ref-napi": "^3.0.3"
  },
  "engines": {
    "node": ">=14.0.0"
  },
  "os": [
    "darwin",
    "linux",
    "win32"
  ]
}
`, config.Namespace, config.Schema.Package)

	packageJsonPath := filepath.Join(packageDir, "package.json")
	if err := os.WriteFile(packageJsonPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write package.json: %w", err)
	}

	fmt.Printf("✓ Generated package.json: %s\n", packageJsonPath)
	return nil
}

// generateJavaScriptReadme generates README.md
func generateJavaScriptReadme(config *PackageConfig, packageDir string) error {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, `# %s - FFire Node.js Bindings

Node.js bindings for the %s schema, generated by [FFire](https://github.com/shaban/ffire).

## Installation

`, config.Namespace, config.Schema.Package)

	buf.WriteString("```bash\n")
	fmt.Fprintf(buf, "cd %s\n", filepath.Base(packageDir))
	buf.WriteString("npm install\n")
	buf.WriteString("```\n\n")

	buf.WriteString("## Usage\n\n")
	buf.WriteString("### JavaScript\n\n")
	buf.WriteString("```javascript\n")
	fmt.Fprintf(buf, "const { Message } = require('%s');\n\n", config.Namespace)
	buf.WriteString("// Decode from binary\n")
	buf.WriteString("const fs = require('fs');\n")
	buf.WriteString("const data = fs.readFileSync('data.bin');\n\n")
	buf.WriteString("const msg = Message.decode(data);\n\n")
	buf.WriteString("// Encode back to binary\n")
	buf.WriteString("const encoded = msg.encode();\n\n")
	buf.WriteString("// Don't forget to free resources\n")
	buf.WriteString("msg.free();\n")
	buf.WriteString("```\n\n")

	buf.WriteString("### TypeScript\n\n")
	buf.WriteString("```typescript\n")
	fmt.Fprintf(buf, "import { Message } from '%s';\n\n", config.Namespace)
	buf.WriteString("// Full type safety!\n")
	buf.WriteString("const data: Buffer = fs.readFileSync('data.bin');\n")
	buf.WriteString("const msg: Message = Message.decode(data);\n")
	buf.WriteString("const encoded: Buffer = msg.encode();\n")
	buf.WriteString("msg.free();\n")
	buf.WriteString("```\n\n")

	buf.WriteString("## API\n\n")

	for _, msg := range config.Schema.Messages {
		fmt.Fprintf(buf, "### `%s`\n\n", msg.Name)
		fmt.Fprintf(buf, "**`%s.decode(data: Buffer): %s`**\n\n", msg.Name, msg.Name)
		fmt.Fprintf(buf, "Decode a `%s` from binary data.\n\n", msg.Name)
		buf.WriteString("- **Parameters:** `data` - Binary data (Buffer)\n")
		fmt.Fprintf(buf, "- **Returns:** `%s` object\n", msg.Name)
		buf.WriteString("- **Throws:** `Error` if decoding fails\n\n")

		fmt.Fprintf(buf, "**`%s.encode(): Buffer`**\n\n", msg.Name)
		fmt.Fprintf(buf, "Encode this `%s` to binary data.\n\n", msg.Name)
		buf.WriteString("- **Returns:** Binary data (Buffer)\n")
		buf.WriteString("- **Throws:** `Error` if encoding fails\n\n")

		fmt.Fprintf(buf, "**`%s.free(): void`**\n\n", msg.Name)
		buf.WriteString("Free the native resources. Should be called when done with the object.\n\n")
	}

	buf.WriteString("## Platform Support\n\n")
	buf.WriteString("This package includes pre-compiled libraries for:\n\n")
	buf.WriteString("- macOS (Darwin): `libffire.dylib`\n")
	buf.WriteString("- Linux: `libffire.so`\n")
	buf.WriteString("- Windows: `ffire.dll`\n\n")
	buf.WriteString("The correct library is automatically loaded based on your platform.\n\n")

	buf.WriteString("## Requirements\n\n")
	buf.WriteString("- Node.js 14.0.0 or higher\n")
	buf.WriteString("- npm (comes with Node.js)\n\n")

	buf.WriteString("## Dependencies\n\n")
	buf.WriteString("- `ffi-napi`: Native FFI bindings for Node.js\n")
	buf.WriteString("- `ref-napi`: Turn Buffer instances into pointers\n\n")

	buf.WriteString("These are automatically installed when you run `npm install`.\n\n")

	buf.WriteString("## License\n\n")
	buf.WriteString("Generated by FFire. See your schema's license for terms.\n")

	readmePath := filepath.Join(packageDir, "README.md")
	if err := os.WriteFile(readmePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	fmt.Printf("✓ Generated README.md: %s\n", readmePath)
	return nil
}
