package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shaban/ffire/pkg/schema"
)

// GenerateJavaPackage generates a complete Java Maven package using the orchestrator
func GenerateJavaPackage(config *PackageConfig) error {
	return orchestrateTierBPackage(
		config,
		JavaLayout,
		generateJavaWrapperOrchestrated,
		generateJavaMetadataOrchestrated,
		printJavaInstructions,
	)
}

func generateJavaWrapperOrchestrated(config *PackageConfig, paths *PackagePaths) error {
	// Create Java source directory structure (Maven style: src/main/java/com/ffire/packagename)
	packagePath := strings.ReplaceAll("com.ffire."+strings.ToLower(config.Namespace), ".", "/")
	javaDir := filepath.Join(paths.Root, "src", "main", "java", packagePath)

	if err := os.MkdirAll(javaDir, 0755); err != nil {
		return fmt.Errorf("failed to create Java source directory: %w", err)
	}

	// Also create resources directory for the native library
	resourcesDir := filepath.Join(paths.Root, "src", "main", "resources", "native")
	if err := os.MkdirAll(resourcesDir, 0755); err != nil {
		return fmt.Errorf("failed to create resources directory: %w", err)
	}

	// Generate Java wrapper classes
	if err := generateJavaWrapper(config, javaDir); err != nil {
		return err
	}

	return nil
}

func generateJavaMetadataOrchestrated(config *PackageConfig, paths *PackagePaths) error {
	// Generate pom.xml (Maven)
	if err := generateJavaPom(config, paths.Root); err != nil {
		return err
	}

	// Generate README.md
	if err := generateJavaReadme(config, paths.Root); err != nil {
		return err
	}

	return nil
}

func printJavaInstructions(config *PackageConfig, paths *PackagePaths) {
	fmt.Printf("\n✅ Java Maven package ready at: %s\n\n", paths.Root)
	fmt.Println("Build:")
	fmt.Printf("  cd %s\n", paths.Root)
	fmt.Println("  mvn clean install")
	fmt.Println()
	fmt.Println("Requirements:")
	fmt.Println("  - Java 8 or higher")
	fmt.Println("  - Maven 3.x")
	fmt.Println("  - Native library must be in java.library.path or bundled in resources")
	fmt.Println()
	fmt.Println("Usage:")
	packageName := "com.ffire." + strings.ToLower(config.Namespace)
	fmt.Printf("  import %s.Message;\n", packageName)
	fmt.Println()
	fmt.Println("  byte[] data = Files.readAllBytes(Paths.get(\"data.bin\"));")
	fmt.Println("  try {")
	fmt.Println("      Message msg = Message.decode(data);")
	fmt.Println("      byte[] encoded = msg.encode();")
	fmt.Println("      msg.free();")
	fmt.Println("  } catch (Exception e) {")
	fmt.Println("      System.err.println(\"Error: \" + e.getMessage());")
	fmt.Println("  }")
	fmt.Println()
}

// generateJavaWrapper generates Java JNI wrapper classes
func generateJavaWrapper(config *PackageConfig, javaDir string) error {
	packageName := "com.ffire." + strings.ToLower(config.Namespace)

	// Generate exception class
	if err := generateJavaException(javaDir, packageName); err != nil {
		return err
	}

	// Generate NativeLibrary loader class
	if err := generateJavaNativeLoader(javaDir, packageName); err != nil {
		return err
	}

	// Generate class for each message type
	for _, msg := range config.Schema.Messages {
		if err := generateJavaMessageClass(javaDir, packageName, &msg); err != nil {
			return err
		}
	}

	return nil
}

func generateJavaException(javaDir string, packageName string) error {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, "package %s;\n\n", packageName)
	buf.WriteString("/**\n")
	buf.WriteString(" * FFire exception for encoding/decoding errors\n")
	buf.WriteString(" */\n")
	buf.WriteString("public class FFireException extends Exception {\n")
	buf.WriteString("    public FFireException(String message) {\n")
	buf.WriteString("        super(message);\n")
	buf.WriteString("    }\n")
	buf.WriteString("}\n")

	exceptionPath := filepath.Join(javaDir, "FFireException.java")
	if err := os.WriteFile(exceptionPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write FFireException.java: %w", err)
	}

	fmt.Printf("✓ Generated FFireException.java\n")
	return nil
}

func generateJavaNativeLoader(javaDir string, packageName string) error {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, "package %s;\n\n", packageName)
	buf.WriteString("/**\n")
	buf.WriteString(" * Native library loader\n")
	buf.WriteString(" */\n")
	buf.WriteString("class NativeLibrary {\n")
	buf.WriteString("    private static boolean loaded = false;\n\n")
	buf.WriteString("    static {\n")
	buf.WriteString("        try {\n")
	buf.WriteString("            // Try to load from java.library.path\n")
	buf.WriteString("            System.loadLibrary(\"ffire\");\n")
	buf.WriteString("            loaded = true;\n")
	buf.WriteString("        } catch (UnsatisfiedLinkError e) {\n")
	buf.WriteString("            // Library not in java.library.path\n")
	buf.WriteString("            loaded = false;\n")
	buf.WriteString("        }\n")
	buf.WriteString("    }\n\n")
	buf.WriteString("    static void ensureLoaded() {\n")
	buf.WriteString("        if (!loaded) {\n")
	buf.WriteString("            throw new UnsatisfiedLinkError(\n")
	buf.WriteString("                \"Native library 'ffire' not found. \" +\n")
	buf.WriteString("                \"Add library path with -Djava.library.path=<path>\"\n")
	buf.WriteString("            );\n")
	buf.WriteString("        }\n")
	buf.WriteString("    }\n")
	buf.WriteString("}\n")

	loaderPath := filepath.Join(javaDir, "NativeLibrary.java")
	if err := os.WriteFile(loaderPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write NativeLibrary.java: %w", err)
	}

	fmt.Printf("✓ Generated NativeLibrary.java\n")
	return nil
}

func generateJavaMessageClass(javaDir string, packageName string, msg *schema.MessageType) error {
	buf := &bytes.Buffer{}
	className := msg.Name
	baseName := strings.ToLower(msg.Name[:1]) + msg.Name[1:]

	fmt.Fprintf(buf, "package %s;\n\n", packageName)
	buf.WriteString("/**\n")
	fmt.Fprintf(buf, " * %s message type\n", className)
	buf.WriteString(" */\n")
	fmt.Fprintf(buf, "public class %s implements AutoCloseable {\n", className)
	buf.WriteString("    static {\n")
	buf.WriteString("        NativeLibrary.ensureLoaded();\n")
	buf.WriteString("    }\n\n")
	buf.WriteString("    private long handle;\n")
	buf.WriteString("    private boolean freed = false;\n\n")

	// Private constructor
	buf.WriteString("    private " + className + "(long handle) {\n")
	buf.WriteString("        this.handle = handle;\n")
	buf.WriteString("    }\n\n")

	// Native method declarations
	buf.WriteString("    // Native method declarations\n")
	fmt.Fprintf(buf, "    private static native long %sDecode(byte[] data, int size);\n", baseName)
	fmt.Fprintf(buf, "    private static native byte[] %sEncode(long handle);\n", baseName)
	fmt.Fprintf(buf, "    private static native void %sFree(long handle);\n", baseName)
	fmt.Fprintf(buf, "    private static native String %sGetError();\n\n", baseName)

	// Decode method
	buf.WriteString("    /**\n")
	fmt.Fprintf(buf, "     * Decode a %s from binary data\n", className)
	buf.WriteString("     * @param data Binary data to decode\n")
	fmt.Fprintf(buf, "     * @return Decoded %s object\n", className)
	buf.WriteString("     * @throws FFireException if decoding fails\n")
	buf.WriteString("     */\n")
	fmt.Fprintf(buf, "    public static %s decode(byte[] data) throws FFireException {\n", className)
	fmt.Fprintf(buf, "        long handle = %sDecode(data, data.length);\n", baseName)
	buf.WriteString("        if (handle == 0) {\n")
	fmt.Fprintf(buf, "            String error = %sGetError();\n", baseName)
	buf.WriteString("            if (error == null) error = \"Unknown error\";\n")
	fmt.Fprintf(buf, "            throw new FFireException(\"Failed to decode %s: \" + error);\n", className)
	buf.WriteString("        }\n")
	fmt.Fprintf(buf, "        return new %s(handle);\n", className)
	buf.WriteString("    }\n\n")

	// Encode method
	buf.WriteString("    /**\n")
	fmt.Fprintf(buf, "     * Encode this %s to binary data\n", className)
	buf.WriteString("     * @return Encoded binary data\n")
	buf.WriteString("     * @throws FFireException if encoding fails\n")
	buf.WriteString("     */\n")
	buf.WriteString("    public byte[] encode() throws FFireException {\n")
	buf.WriteString("        if (freed) {\n")
	fmt.Fprintf(buf, "            throw new FFireException(\"%s already freed\");\n", className)
	buf.WriteString("        }\n\n")
	fmt.Fprintf(buf, "        byte[] result = %sEncode(handle);\n", baseName)
	buf.WriteString("        if (result == null) {\n")
	fmt.Fprintf(buf, "            String error = %sGetError();\n", baseName)
	buf.WriteString("            if (error == null) error = \"Unknown error\";\n")
	fmt.Fprintf(buf, "            throw new FFireException(\"Failed to encode %s: \" + error);\n", className)
	buf.WriteString("        }\n")
	buf.WriteString("        return result;\n")
	buf.WriteString("    }\n\n")

	// Free method
	buf.WriteString("    /**\n")
	buf.WriteString("     * Free the native resources\n")
	buf.WriteString("     */\n")
	buf.WriteString("    public void free() {\n")
	buf.WriteString("        if (!freed && handle != 0) {\n")
	fmt.Fprintf(buf, "            %sFree(handle);\n", baseName)
	buf.WriteString("            freed = true;\n")
	buf.WriteString("            handle = 0;\n")
	buf.WriteString("        }\n")
	buf.WriteString("    }\n\n")

	// AutoCloseable interface
	buf.WriteString("    @Override\n")
	buf.WriteString("    public void close() {\n")
	buf.WriteString("        free();\n")
	buf.WriteString("    }\n\n")

	// Finalizer
	buf.WriteString("    @Override\n")
	buf.WriteString("    protected void finalize() throws Throwable {\n")
	buf.WriteString("        try {\n")
	buf.WriteString("            free();\n")
	buf.WriteString("        } finally {\n")
	buf.WriteString("            super.finalize();\n")
	buf.WriteString("        }\n")
	buf.WriteString("    }\n")
	buf.WriteString("}\n")

	classPath := filepath.Join(javaDir, className+".java")
	if err := os.WriteFile(classPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write %s.java: %w", className, err)
	}

	fmt.Printf("✓ Generated %s.java\n", className)
	return nil
}

// generateJavaPom generates pom.xml for Maven
func generateJavaPom(config *PackageConfig, packageDir string) error {
	buf := &bytes.Buffer{}

	groupId := "com.ffire"
	artifactId := strings.ToLower(config.Namespace)

	buf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	buf.WriteString("<project xmlns=\"http://maven.apache.org/POM/4.0.0\"\n")
	buf.WriteString("         xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"\n")
	buf.WriteString("         xsi:schemaLocation=\"http://maven.apache.org/POM/4.0.0\n")
	buf.WriteString("         http://maven.apache.org/xsd/maven-4.0.0.xsd\">\n")
	buf.WriteString("    <modelVersion>4.0.0</modelVersion>\n\n")

	fmt.Fprintf(buf, "    <groupId>%s</groupId>\n", groupId)
	fmt.Fprintf(buf, "    <artifactId>%s</artifactId>\n", artifactId)
	buf.WriteString("    <version>1.0.0</version>\n")
	buf.WriteString("    <packaging>jar</packaging>\n\n")

	fmt.Fprintf(buf, "    <name>%s</name>\n", config.Namespace)
	fmt.Fprintf(buf, "    <description>FFire binary serialization library - %s schema</description>\n\n", config.Schema.Package)

	buf.WriteString("    <properties>\n")
	buf.WriteString("        <maven.compiler.source>1.8</maven.compiler.source>\n")
	buf.WriteString("        <maven.compiler.target>1.8</maven.compiler.target>\n")
	buf.WriteString("        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>\n")
	buf.WriteString("    </properties>\n\n")

	buf.WriteString("    <build>\n")
	buf.WriteString("        <plugins>\n")
	buf.WriteString("            <plugin>\n")
	buf.WriteString("                <groupId>org.apache.maven.plugins</groupId>\n")
	buf.WriteString("                <artifactId>maven-compiler-plugin</artifactId>\n")
	buf.WriteString("                <version>3.8.1</version>\n")
	buf.WriteString("                <configuration>\n")
	buf.WriteString("                    <source>1.8</source>\n")
	buf.WriteString("                    <target>1.8</target>\n")
	buf.WriteString("                </configuration>\n")
	buf.WriteString("            </plugin>\n")
	buf.WriteString("        </plugins>\n")
	buf.WriteString("    </build>\n")
	buf.WriteString("</project>\n")

	pomPath := filepath.Join(packageDir, "pom.xml")
	if err := os.WriteFile(pomPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write pom.xml: %w", err)
	}

	fmt.Printf("✓ Generated pom.xml\n")
	return nil
}

// generateJavaReadme generates README.md
func generateJavaReadme(config *PackageConfig, packageDir string) error {
	buf := &bytes.Buffer{}

	packageName := "com.ffire." + strings.ToLower(config.Namespace)

	fmt.Fprintf(buf, `# %s - FFire Java Bindings

Java bindings for the %s schema, generated by [FFire](https://github.com/shaban/ffire).

## Requirements

- Java 8 or higher
- Maven 3.x
- Native library (libffire.so/dylib/dll) must be accessible

## Installation

### Build with Maven

`, config.Namespace, config.Schema.Package)

	buf.WriteString("```bash\n")
	fmt.Fprintf(buf, "cd %s\n", filepath.Base(packageDir))
	buf.WriteString("mvn clean install\n")
	buf.WriteString("```\n\n")

	buf.WriteString("### Running with Native Library\n\n")
	buf.WriteString("The native library must be in your `java.library.path`:\n\n")
	buf.WriteString("```bash\n")
	buf.WriteString("java -Djava.library.path=/path/to/lib -jar your-app.jar\n")
	buf.WriteString("```\n\n")

	buf.WriteString("Or copy the library to a system library path:\n")
	buf.WriteString("- **macOS**: `/usr/local/lib/`\n")
	buf.WriteString("- **Linux**: `/usr/lib/` or `/usr/local/lib/`\n")
	buf.WriteString("- **Windows**: System PATH\n\n")

	buf.WriteString("## Usage\n\n")
	buf.WriteString("```java\n")
	fmt.Fprintf(buf, "import %s.Message;\n", packageName)
	fmt.Fprintf(buf, "import %s.FFireException;\n", packageName)
	buf.WriteString("import java.nio.file.Files;\n")
	buf.WriteString("import java.nio.file.Paths;\n\n")
	buf.WriteString("public class Example {\n")
	buf.WriteString("    public static void main(String[] args) {\n")
	buf.WriteString("        try {\n")
	buf.WriteString("            // Decode from binary\n")
	buf.WriteString("            byte[] data = Files.readAllBytes(Paths.get(\"data.bin\"));\n")
	buf.WriteString("            Message msg = Message.decode(data);\n\n")
	buf.WriteString("            // Encode back to binary\n")
	buf.WriteString("            byte[] encoded = msg.encode();\n\n")
	buf.WriteString("            // Free resources\n")
	buf.WriteString("            msg.free();\n\n")
	buf.WriteString("            // Or use try-with-resources (implements AutoCloseable)\n")
	buf.WriteString("            try (Message msg2 = Message.decode(data)) {\n")
	buf.WriteString("                byte[] encoded2 = msg2.encode();\n")
	buf.WriteString("            }\n")
	buf.WriteString("        } catch (FFireException e) {\n")
	buf.WriteString("            System.err.println(\"FFire error: \" + e.getMessage());\n")
	buf.WriteString("        } catch (Exception e) {\n")
	buf.WriteString("            e.printStackTrace();\n")
	buf.WriteString("        }\n")
	buf.WriteString("    }\n")
	buf.WriteString("}\n")
	buf.WriteString("```\n\n")

	buf.WriteString("## API\n\n")

	for _, msg := range config.Schema.Messages {
		fmt.Fprintf(buf, "### `%s`\n\n", msg.Name)
		fmt.Fprintf(buf, "**`static %s decode(byte[] data) throws FFireException`**\n\n", msg.Name)
		fmt.Fprintf(buf, "Decode a `%s` from binary data.\n\n", msg.Name)
		buf.WriteString("- **Parameter:** `data` - Binary data (byte[])\n")
		fmt.Fprintf(buf, "- **Returns:** `%s` object\n", msg.Name)
		buf.WriteString("- **Throws:** `FFireException` if decoding fails\n\n")

		buf.WriteString("**`byte[] encode() throws FFireException`**\n\n")
		fmt.Fprintf(buf, "Encode this `%s` to binary data.\n\n", msg.Name)
		buf.WriteString("- **Returns:** Binary data (byte[])\n")
		buf.WriteString("- **Throws:** `FFireException` if encoding fails\n\n")

		buf.WriteString("**`void free()`**\n\n")
		buf.WriteString("Manually free the native resources. Called automatically by finalizer or when used with try-with-resources.\n\n")
	}

	buf.WriteString("## Memory Management\n\n")
	buf.WriteString("This library implements `AutoCloseable`, so you can use try-with-resources:\n\n")
	buf.WriteString("```java\n")
	buf.WriteString("try (Message msg = Message.decode(data)) {\n")
	buf.WriteString("    // Use msg\n")
	buf.WriteString("} // Automatically freed\n")
	buf.WriteString("```\n\n")
	buf.WriteString("Alternatively, call `free()` manually or let the garbage collector handle it via the finalizer.\n\n")

	buf.WriteString("## Platform Support\n\n")
	buf.WriteString("Requires the appropriate native library for your platform:\n\n")
	buf.WriteString("- macOS: `libffire.dylib`\n")
	buf.WriteString("- Linux: `libffire.so`\n")
	buf.WriteString("- Windows: `ffire.dll`\n\n")

	buf.WriteString("## License\n\n")
	buf.WriteString("Generated by FFire. See your schema's license for terms.\n")

	readmePath := filepath.Join(packageDir, "README.md")
	if err := os.WriteFile(readmePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	fmt.Printf("✓ Generated README.md\n")
	return nil
}
