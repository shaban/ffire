package generator

import (
"bytes"
"fmt"
"os"
"path/filepath"
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
	buf.WriteString("import 'dart:typed_data';\n\n")
	
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
	buf.WriteString("      return DynamicLibrary.open('libffire.dylib');\n")
	buf.WriteString("    } else if (Platform.isLinux) {\n")
	buf.WriteString("      return DynamicLibrary.open('libffire.so');\n")
	buf.WriteString("    } else if (Platform.isWindows) {\n")
	buf.WriteString("      return DynamicLibrary.open('ffire.dll');\n")
	buf.WriteString("    }\n")
	buf.WriteString("    throw UnsupportedError('Platform not supported');\n")
	buf.WriteString("  }\n\n")
	
	// Native function declarations
	buf.WriteString("  static final ffire_decode = _lib\n")
	buf.WriteString("      .lookup<NativeFunction<Pointer<Void> Function(Pointer<Uint8>, Int32)>>('ffire_decode')\n")
	buf.WriteString("      .asFunction<Pointer<Void> Function(Pointer<Uint8>, int)>();\n\n")
	
	buf.WriteString("  static final ffire_encode = _lib\n")
	buf.WriteString("      .lookup<NativeFunction<Pointer<Void> Function(Pointer<Void>)>>('ffire_encode')\n")
	buf.WriteString("      .asFunction<Pointer<Void> Function(Pointer<Void>)>();\n\n")
	
	buf.WriteString("  static final ffire_free = _lib\n")
	buf.WriteString("      .lookup<NativeFunction<Void Function(Pointer<Void>)>>('ffire_free')\n")
	buf.WriteString("      .asFunction<void Function(Pointer<Void>)>();\n\n")
	
	buf.WriteString("  static final ffire_get_error = _lib\n")
	buf.WriteString("      .lookup<NativeFunction<Pointer<Utf8> Function()>>('ffire_get_error')\n")
	buf.WriteString("      .asFunction<Pointer<Utf8> Function()>();\n\n")
	
	buf.WriteString("  static String getLastError() {\n")
	buf.WriteString("    final errPtr = ffire_get_error();\n")
	buf.WriteString("    if (errPtr.address == 0) return 'Unknown error';\n")
	buf.WriteString("    return errPtr.toDartString();\n")
	buf.WriteString("  }\n")
	buf.WriteString("}\n\n")
	
	// Message class
	buf.WriteString("class Message {\n")
	buf.WriteString("  Pointer<Void>? _handle;\n")
	buf.WriteString("  bool _disposed = false;\n\n")
	
	buf.WriteString("  Message._(this._handle);\n\n")
	
	buf.WriteString("  static Message decode(Uint8List data) {\n")
	buf.WriteString("    final dataPtr = malloc.allocate<Uint8>(data.length);\n")
	buf.WriteString("    final dataList = dataPtr.asTypedList(data.length);\n")
	buf.WriteString("    dataList.setAll(0, data);\n\n")
	
	buf.WriteString("    final handle = _NativeLibrary.ffire_decode(dataPtr, data.length);\n")
	buf.WriteString("    malloc.free(dataPtr);\n\n")
	
	buf.WriteString("    if (handle.address == 0) {\n")
	buf.WriteString("      final error = _NativeLibrary.getLastError();\n")
	fmt.Fprintf(buf, "      throw %sException('Decode failed: $error');\n", ToPascalCase(packageName))
	buf.WriteString("    }\n\n")
	
	buf.WriteString("    return Message._(handle);\n")
	buf.WriteString("  }\n\n")
	
	buf.WriteString("  Uint8List encode() {\n")
	buf.WriteString("    if (_disposed) {\n")
	buf.WriteString("      throw StateError('Message has been disposed');\n")
	buf.WriteString("    }\n\n")
	
	buf.WriteString("    final resultPtr = _NativeLibrary.ffire_encode(_handle!);\n")
	buf.WriteString("    if (resultPtr.address == 0) {\n")
	buf.WriteString("      final error = _NativeLibrary.getLastError();\n")
	fmt.Fprintf(buf, "      throw %sException('Encode failed: $error');\n", ToPascalCase(packageName))
	buf.WriteString("    }\n\n")
	
	buf.WriteString("    // Read length (first 4 bytes) and data\n")
	buf.WriteString("    final lengthPtr = resultPtr.cast<Uint32>();\n")
	buf.WriteString("    final length = lengthPtr.value;\n")
	buf.WriteString("    final dataPtr = Pointer<Uint8>.fromAddress(resultPtr.address + 4);\n")
	buf.WriteString("    final data = Uint8List.fromList(dataPtr.asTypedList(length));\n\n")
	
	buf.WriteString("    return data;\n")
	buf.WriteString("  }\n\n")
	
	buf.WriteString("  void dispose() {\n")
	buf.WriteString("    if (!_disposed && _handle != null) {\n")
	buf.WriteString("      _NativeLibrary.ffire_free(_handle!);\n")
	buf.WriteString("      _handle = null;\n")
	buf.WriteString("      _disposed = true;\n")
	buf.WriteString("    }\n")
	buf.WriteString("  }\n")
	buf.WriteString("}\n")

	filePath := filepath.Join(libDir, packageName+".dart")
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write Dart library: %w", err)
	}

	fmt.Printf("✓ Generated %s.dart\n", packageName)
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
	buf.WriteString("- Native library (libffire.dylib/.so/.dll) must be in library search path\n\n")
	
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
