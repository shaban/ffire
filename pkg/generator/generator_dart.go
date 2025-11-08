package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shaban/ffire/pkg/schema"
)

// GenerateDartPackage generates a complete Dart package with FFI bindings
func GenerateDartPackage(config *PackageConfig) error {
	return orchestrateTierBPackage(
		config,
		DartLayout,
		generateDartWrapperOrchestrated,
		generateDartMetadataOrchestrated,
		printDartInstructions,
	)
}

func generateDartWrapperOrchestrated(config *PackageConfig, paths *PackagePaths) error {
	// Create Dart lib directory structure
	libDir := filepath.Join(paths.Root, "lib")
	if err := os.MkdirAll(libDir, 0755); err != nil {
		return fmt.Errorf("failed to create lib directory: %w", err)
	}

	// Generate Dart FFI wrapper
	return generateDartFiles(config, libDir, paths.Lib)
}

func generateDartMetadataOrchestrated(config *PackageConfig, paths *PackagePaths) error {
	// Generate pubspec.yaml
	if err := generateDartPubspec(config, paths.Root); err != nil {
		return err
	}

	// Generate README.md
	return generateDartReadme(config, paths.Root)
}

func printDartInstructions(config *PackageConfig, paths *PackagePaths) {
	fmt.Printf("\n✅ Dart package ready at: %s\n\n", paths.Root)
	fmt.Println("Build:")
	fmt.Printf("  cd %s\n", paths.Root)
	fmt.Println("  dart pub get")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  import 'package:%s/%s.dart';\n\n", config.Namespace, config.Namespace)
	fmt.Println("  final data = await File('data.bin').readAsBytes();")
	fmt.Println("  final msg = Message.decode(data);")
	fmt.Println("  final encoded = msg.encode();")
	fmt.Println()
}

func generateDartFiles(config *PackageConfig, libDir, nativeLibDir string) error {
	packageName := config.Namespace

	// Main library file
	buf := &bytes.Buffer{}

	buf.WriteString("import 'dart:ffi';\n")
	buf.WriteString("import 'dart:io';\n")
	buf.WriteString("import 'dart:typed_data';\n")
	buf.WriteString("import 'package:ffi/ffi.dart';\n\n")

	// Exception class
	fmt.Fprintf(buf, "class %sException implements Exception {\n", ToPascalCase(packageName))
	buf.WriteString("  final String message;\n")
	fmt.Fprintf(buf, "  %sException(this.message);\n", ToPascalCase(packageName))
	buf.WriteString("  @override\n")
	buf.WriteString("  String toString() => message;\n")
	buf.WriteString("}\n\n")

	// Native library loader
	buf.WriteString("class _NativeLibrary {\n")
	buf.WriteString("  static final DynamicLibrary _lib = _loadLibrary();\n\n")
	buf.WriteString("  static DynamicLibrary _loadLibrary() {\n")
	buf.WriteString("    if (Platform.isMacOS) {\n")
	fmt.Fprintf(buf, "      return DynamicLibrary.open('lib/lib%s.dylib');\n", config.Schema.Package)
	buf.WriteString("    } else if (Platform.isLinux) {\n")
	fmt.Fprintf(buf, "      return DynamicLibrary.open('lib/lib%s.so');\n", config.Schema.Package)
	buf.WriteString("    } else if (Platform.isWindows) {\n")
	fmt.Fprintf(buf, "      return DynamicLibrary.open('lib/%s.dll');\n", config.Schema.Package)
	buf.WriteString("    }\n")
	buf.WriteString("    throw UnsupportedError('Platform not supported');\n")
	buf.WriteString("  }\n")
	buf.WriteString("}\n\n")

	// Generate bindings for each message type
	for _, msg := range config.Schema.Messages {
		if err := generateDartMessageBindings(buf, config.Schema, &msg, packageName); err != nil {
			return err
		}
	}

	filePath := filepath.Join(libDir, packageName+".dart")
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write Dart library: %w", err)
	}

	fmt.Printf("✓ Generated %s.dart\n", packageName)
	return nil
}

func generateDartMessageBindings(buf *bytes.Buffer, s *schema.Schema, msg *schema.MessageType, packageName string) error {
	baseName := strings.ToLower(msg.Name) // All lowercase to match C ABI
	className := msg.Name

	// Native function references in _NativeLibrary class extension
	fmt.Fprintf(buf, "extension _%sFunctions on _NativeLibrary {\n", className)

	// Decode function: Pointer<Void> func(Pointer<Uint8> data, int32_t len, Pointer<Pointer<Utf8>> error)
	fmt.Fprintf(buf, "  static final %s_decode = _NativeLibrary._lib\n", baseName)
	fmt.Fprintf(buf, "      .lookup<NativeFunction<Pointer<Void> Function(Pointer<Uint8>, Int32, Pointer<Pointer<Utf8>>)>>('%s_decode')\n", baseName)
	fmt.Fprintf(buf, "      .asFunction<Pointer<Void> Function(Pointer<Uint8>, int, Pointer<Pointer<Utf8>>)>();\n\n")

	// Encode function: size_t func(Pointer<Void> handle, Pointer<Pointer<Uint8>> out_data, Pointer<Pointer<Utf8>> error)
	fmt.Fprintf(buf, "  static final %s_encode = _NativeLibrary._lib\n", baseName)
	fmt.Fprintf(buf, "      .lookup<NativeFunction<Size Function(Pointer<Void>, Pointer<Pointer<Uint8>>, Pointer<Pointer<Utf8>>)>>('%s_encode')\n", baseName)
	fmt.Fprintf(buf, "      .asFunction<int Function(Pointer<Void>, Pointer<Pointer<Uint8>>, Pointer<Pointer<Utf8>>)>();\n\n")

	// Free functions
	fmt.Fprintf(buf, "  static final %s_free = _NativeLibrary._lib\n", baseName)
	fmt.Fprintf(buf, "      .lookup<NativeFunction<Void Function(Pointer<Void>)>>('%s_free')\n", baseName)
	fmt.Fprintf(buf, "      .asFunction<void Function(Pointer<Void>)>();\n\n")

	fmt.Fprintf(buf, "  static final %s_free_data = _NativeLibrary._lib\n", baseName)
	fmt.Fprintf(buf, "      .lookup<NativeFunction<Void Function(Pointer<Uint8>)>>('%s_free_data')\n", baseName)
	fmt.Fprintf(buf, "      .asFunction<void Function(Pointer<Uint8>)>();\n\n")

	fmt.Fprintf(buf, "  static final %s_free_error = _NativeLibrary._lib\n", baseName)
	fmt.Fprintf(buf, "      .lookup<NativeFunction<Void Function(Pointer<Utf8>)>>('%s_free_error')\n", baseName)
	fmt.Fprintf(buf, "      .asFunction<void Function(Pointer<Utf8>)>();\n")

	buf.WriteString("}\n\n")

	// Message class
	fmt.Fprintf(buf, "class %s {\n", className)
	buf.WriteString("  Pointer<Void>? _handle;\n")
	buf.WriteString("  bool _disposed = false;\n\n")

	fmt.Fprintf(buf, "  %s._(this._handle);\n\n", className)

	// Decode method
	fmt.Fprintf(buf, "  static %s decode(Uint8List data) {\n", className)
	buf.WriteString("    final dataPtr = malloc.allocate<Uint8>(data.length);\n")
	buf.WriteString("    final dataList = dataPtr.asTypedList(data.length);\n")
	buf.WriteString("    dataList.setAll(0, data);\n\n")

	buf.WriteString("    final errorPtr = malloc.allocate<Pointer<Utf8>>(sizeOf<Pointer<Utf8>>());\n")
	buf.WriteString("    errorPtr.value = nullptr;\n\n")

	fmt.Fprintf(buf, "    final handle = _%sFunctions.%s_decode(dataPtr, data.length, errorPtr);\n", className, baseName)
	buf.WriteString("    malloc.free(dataPtr);\n\n")

	buf.WriteString("    if (handle.address == 0) {\n")
	buf.WriteString("      final errMsg = errorPtr.value.address != 0 \n")
	buf.WriteString("        ? errorPtr.value.toDartString() \n")
	buf.WriteString("        : 'Unknown error';\n")
	buf.WriteString("      if (errorPtr.value.address != 0) {\n")
	fmt.Fprintf(buf, "        _%sFunctions.%s_free_error(errorPtr.value);\n", className, baseName)
	buf.WriteString("      }\n")
	buf.WriteString("      malloc.free(errorPtr);\n")
	buf.WriteString("      throw ")
	fmt.Fprintf(buf, "%sException('Decode failed: $errMsg');\n", ToPascalCase(packageName))
	buf.WriteString("    }\n\n")

	buf.WriteString("    malloc.free(errorPtr);\n")
	fmt.Fprintf(buf, "    return %s._(handle);\n", className)
	buf.WriteString("  }\n\n")

	// Encode method
	buf.WriteString("  Uint8List encode() {\n")
	buf.WriteString("    if (_disposed) {\n")
	buf.WriteString("      throw StateError('Message has been disposed');\n")
	buf.WriteString("    }\n\n")

	buf.WriteString("    final outDataPtr = malloc.allocate<Pointer<Uint8>>(sizeOf<Pointer<Uint8>>());\n")
	buf.WriteString("    outDataPtr.value = nullptr;\n")
	buf.WriteString("    final errorPtr = malloc.allocate<Pointer<Utf8>>(sizeOf<Pointer<Utf8>>());\n")
	buf.WriteString("    errorPtr.value = nullptr;\n\n")

	fmt.Fprintf(buf, "    final size = _%sFunctions.%s_encode(_handle!, outDataPtr, errorPtr);\n", className, baseName)

	buf.WriteString("    if (size == 0) {\n")
	buf.WriteString("      final errMsg = errorPtr.value.address != 0 \n")
	buf.WriteString("        ? errorPtr.value.toDartString() \n")
	buf.WriteString("        : 'Unknown error';\n")
	buf.WriteString("      if (errorPtr.value.address != 0) {\n")
	fmt.Fprintf(buf, "        _%sFunctions.%s_free_error(errorPtr.value);\n", className, baseName)
	buf.WriteString("      }\n")
	buf.WriteString("      malloc.free(outDataPtr);\n")
	buf.WriteString("      malloc.free(errorPtr);\n")
	buf.WriteString("      throw ")
	fmt.Fprintf(buf, "%sException('Encode failed: $errMsg');\n", ToPascalCase(packageName))
	buf.WriteString("    }\n\n")

	buf.WriteString("    // Copy data to Dart\n")
	buf.WriteString("    final result = Uint8List.fromList(outDataPtr.value.asTypedList(size));\n")
	fmt.Fprintf(buf, "    _%sFunctions.%s_free_data(outDataPtr.value);\n", className, baseName)
	buf.WriteString("    malloc.free(outDataPtr);\n")
	buf.WriteString("    malloc.free(errorPtr);\n\n")

	buf.WriteString("    return result;\n")
	buf.WriteString("  }\n\n")

	// Dispose method
	buf.WriteString("  void dispose() {\n")
	buf.WriteString("    if (!_disposed && _handle != null) {\n")
	fmt.Fprintf(buf, "      _%sFunctions.%s_free(_handle!);\n", className, baseName)
	buf.WriteString("      _handle = null;\n")
	buf.WriteString("      _disposed = true;\n")
	buf.WriteString("    }\n")
	buf.WriteString("  }\n")
	buf.WriteString("}\n\n")

	return nil
}

func generateDartPubspec(config *PackageConfig, dartDir string) error {
	buf := &bytes.Buffer{}
	packageName := config.Namespace

	fmt.Fprintf(buf, "name: %s\n", packageName)
	buf.WriteString("description: Dart FFI bindings for ffire schema\n")
	buf.WriteString("version: 1.0.0\n\n")

	buf.WriteString("environment:\n")
	buf.WriteString("  sdk: '>=2.17.0 <4.0.0'\n\n")

	buf.WriteString("dependencies:\n")
	buf.WriteString("  ffi: ^2.0.0\n")

	filePath := filepath.Join(dartDir, "pubspec.yaml")
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write pubspec.yaml: %w", err)
	}

	fmt.Println("✓ Generated pubspec.yaml")
	return nil
}

func generateDartReadme(config *PackageConfig, dartDir string) error {
	buf := &bytes.Buffer{}
	packageName := config.Namespace

	fmt.Fprintf(buf, "# %s - Dart FFI Bindings\n\n", packageName)
	buf.WriteString("Dart package with FFI bindings for native library interop.\n\n")

	buf.WriteString("## Installation\n\n")
	buf.WriteString("Add to your `pubspec.yaml`:\n\n")
	buf.WriteString("```yaml\n")
	buf.WriteString("dependencies:\n")
	fmt.Fprintf(buf, "  %s:\n", packageName)
	buf.WriteString("    path: ../path/to/package\n")
	buf.WriteString("```\n\n")

	buf.WriteString("Then run:\n\n")
	buf.WriteString("```bash\n")
	buf.WriteString("dart pub get\n")
	buf.WriteString("```\n\n")

	buf.WriteString("## Requirements\n\n")
	buf.WriteString("- Dart SDK 2.17.0 or higher\n")
	fmt.Fprintf(buf, "- Native library (lib%s.dylib/.so or %s.dll) must be in library search path\n\n", packageName, packageName)

	buf.WriteString("## Usage\n\n")
	buf.WriteString("```dart\n")
	fmt.Fprintf(buf, "import 'package:%s/%s.dart';\n", packageName, packageName)
	buf.WriteString("import 'dart:io';\n\n")

	buf.WriteString("void main() async {\n")
	buf.WriteString("  // Read binary data\n")
	buf.WriteString("  final data = await File('data.bin').readAsBytes();\n\n")

	buf.WriteString("  try {\n")
	buf.WriteString("    // Decode message\n")
	buf.WriteString("    final msg = Message.decode(data);\n\n")

	buf.WriteString("    // Encode back\n")
	buf.WriteString("    final encoded = msg.encode();\n")
	buf.WriteString("    await File('output.bin').writeAsBytes(encoded);\n\n")

	buf.WriteString("    // Clean up\n")
	buf.WriteString("    msg.dispose();\n")
	buf.WriteString("  } catch (e) {\n")
	buf.WriteString("    print('Error: $e');\n")
	buf.WriteString("  }\n")
	buf.WriteString("}\n")
	buf.WriteString("```\n\n")

	buf.WriteString("## API\n\n")
	buf.WriteString("### Message Class\n\n")
	buf.WriteString("- **`static Message decode(Uint8List data)`**  \n")
	buf.WriteString("  Decode binary data into a Message.\n\n")

	buf.WriteString("- **`Uint8List encode()`**  \n")
	buf.WriteString("  Encode the message back to binary.\n\n")

	buf.WriteString("- **`void dispose()`**  \n")
	buf.WriteString("  Free native resources. Always call when done.\n\n")

	buf.WriteString("## License\n\n")
	buf.WriteString("Generated by FFireGenerator\n")

	filePath := filepath.Join(dartDir, "README.md")
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	fmt.Println("✓ Generated README.md")
	return nil
}
