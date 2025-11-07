package benchmark

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shaban/ffire/pkg/fixture"
	"github.com/shaban/ffire/pkg/generator"
	"github.com/shaban/ffire/pkg/schema"
)

// GenerateJava generates a Java benchmark with embedded fixture
func GenerateJava(schema *schema.Schema, schemaName, messageName string, jsonData []byte, outputDir string, iterations int) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Step 1: Generate the Java package
	config := &generator.PackageConfig{
		Schema:    schema,
		Language:  "java",
		OutputDir: outputDir,
		Namespace: schemaName,
		Optimize:  2,
		Platform:  "current",
		Arch:      "current",
		NoCompile: false,
		Verbose:   false,
	}

	if err := generator.GeneratePackage(config); err != nil {
		return fmt.Errorf("failed to generate Java package: %w", err)
	}

	// Step 2: Convert JSON to binary fixture
	binaryData, err := fixture.Convert(schema, messageName, jsonData)
	if err != nil {
		return fmt.Errorf("failed to convert JSON to binary: %w", err)
	}

	// Step 3: Write the binary fixture
	javaDir := filepath.Join(outputDir, "java")
	fixturePath := filepath.Join(javaDir, "fixture.bin")
	if err := os.WriteFile(fixturePath, binaryData, 0644); err != nil {
		return fmt.Errorf("failed to write fixture: %w", err)
	}

	// Step 4: Generate the benchmark harness
	benchmarkCode := generateJavaBenchmarkCode(schemaName, messageName, iterations)
	benchPath := filepath.Join(javaDir, "Benchmark.java")
	if err := os.WriteFile(benchPath, []byte(benchmarkCode), 0644); err != nil {
		return fmt.Errorf("failed to write benchmark: %w", err)
	}

	// Step 5: Generate a run script for convenience
	runScript := generateJavaRunScript(schemaName)
	runPath := filepath.Join(javaDir, "run.sh")
	if err := os.WriteFile(runPath, []byte(runScript), 0755); err != nil {
		return fmt.Errorf("failed to write run script: %w", err)
	}

	return nil
}

// generateJavaBenchmarkCode generates the benchmark harness code
func generateJavaBenchmarkCode(schemaName, messageName string, iterations int) string {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, `import com.sun.jna.*;
import com.sun.jna.ptr.IntByReference;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.time.Instant;
import com.google.gson.Gson;
import java.util.HashMap;
import java.util.Map;

public class Benchmark {
    // JNA interface for native library
    public interface FFire%sLib extends Library {
        String libName = Platform.isWindows() ? "%s" : 
                        Platform.isMac() ? "lib%s" : "lib%s";
        
        FFire%sLib INSTANCE = Native.load(libName, FFire%sLib.class);
        
        Pointer ffire_decode_%s(Pointer data, int size);
        Pointer ffire_encode_%s(Pointer msg, int flags, IntByReference size);
        void ffire_free_%s(Pointer msg);
    }
    
    private static final FFire%sLib lib = FFire%sLib.INSTANCE;
    
    private static Pointer decode(byte[] data) {
        Memory mem = new Memory(data.length);
        mem.write(0, data, 0, data.length);
        Pointer result = lib.ffire_decode_%s(mem, data.length);
        if (result == null) {
            throw new RuntimeException("Decode failed");
        }
        return result;
    }
    
    private static byte[] encode(Pointer msgPtr) {
        IntByReference size = new IntByReference();
        Pointer result = lib.ffire_encode_%s(msgPtr, 0, size);
        if (result == null) {
            throw new RuntimeException("Encode failed");
        }
        int len = size.getValue();
        return result.getByteArray(0, len);
    }
    
    private static void freeMessage(Pointer msgPtr) {
        lib.ffire_free_%s(msgPtr);
    }
    
    public static void main(String[] args) throws IOException {
        // Load fixture
        byte[] fixtureData = Files.readAllBytes(Paths.get("fixture.bin"));
        
        int iterations = %d;
        boolean jsonOutput = "1".equals(System.getenv("BENCH_JSON"));
        
        // Warmup
        for (int i = 0; i < 1000; i++) {
            Pointer msgPtr = decode(fixtureData);
            byte[] encoded = encode(msgPtr);
            freeMessage(msgPtr);
        }
        
        // Benchmark decode
        long decodeStart = System.nanoTime();
        for (int i = 0; i < iterations; i++) {
            Pointer msgPtr = decode(fixtureData);
            freeMessage(msgPtr);
        }
        long decodeEnd = System.nanoTime();
        long decodeTimeNs = decodeEnd - decodeStart;
        
        // Benchmark encode (decode once, then encode many times)
        Pointer msgPtr = decode(fixtureData);
        long encodeStart = System.nanoTime();
        byte[] encoded = null;
        for (int i = 0; i < iterations; i++) {
            encoded = encode(msgPtr);
        }
        long encodeEnd = System.nanoTime();
        long encodeTimeNs = encodeEnd - encodeStart;
        freeMessage(msgPtr);
        
        // Calculate metrics
        long encodeNs = encodeTimeNs / iterations;
        long decodeNs = decodeTimeNs / iterations;
        long totalNs = encodeNs + decodeNs;
        
        if (jsonOutput) {
            // Output JSON for automation
            Map<String, Object> result = new HashMap<>();
            result.put("language", "Java");
            result.put("format", "ffire");
            result.put("message", "%s");
            result.put("iterations", iterations);
            result.put("encode_ns", encodeNs);
            result.put("decode_ns", decodeNs);
            result.put("total_ns", totalNs);
            result.put("wire_size", encoded.length);
            result.put("fixture_size", fixtureData.length);
            result.put("timestamp", Instant.now().toString());
            
            Gson gson = new Gson();
            System.out.println(gson.toJson(result));
        } else {
            // Print human-readable results
            System.out.println("ffire benchmark: %s");
            System.out.println("Iterations:  " + iterations);
            System.out.println("Encode:      " + encodeNs + " ns/op");
            System.out.println("Decode:      " + decodeNs + " ns/op");
            System.out.println("Total:       " + totalNs + " ns/op");
            System.out.println("Wire size:   " + encoded.length + " bytes");
            System.out.println("Fixture:     " + fixtureData.length + " bytes");
            double totalTimeS = (encodeTimeNs + decodeTimeNs) / 1e9;
            System.out.printf("Total time:  %%.3fs%%n", totalTimeS);
        }
    }
}
`, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName,
		schemaName, schemaName, schemaName, schemaName, schemaName, iterations, schemaName, schemaName)

	return buf.String()
}

// generateJavaRunScript generates a convenience run script
func generateJavaRunScript(schemaName string) string {
	_ = schemaName // Reserved for future use
	return `#!/bin/bash
# Convenience script to run Java benchmark

# Check if java is available
if ! command -v java &> /dev/null; then
    echo "Error: java not found"
    exit 1
fi

if ! command -v javac &> /dev/null; then
    echo "Error: javac not found"
    exit 1
fi

# Download dependencies if needed
if [ ! -f "jna-5.13.0.jar" ]; then
    echo "Downloading JNA..."
    curl -L -o jna-5.13.0.jar https://repo1.maven.org/maven2/net/java/dev/jna/jna/5.13.0/jna-5.13.0.jar
fi

if [ ! -f "gson-2.10.1.jar" ]; then
    echo "Downloading Gson..."
    curl -L -o gson-2.10.1.jar https://repo1.maven.org/maven2/com/google/code/gson/gson/2.10.1/gson-2.10.1.jar
fi

# Compile if needed
if [ ! -f "Benchmark.class" ] || [ "Benchmark.java" -nt "Benchmark.class" ]; then
    echo "Compiling..."
    javac -cp ".:jna-5.13.0.jar:gson-2.10.1.jar" Benchmark.java
fi

# Set library path
export LD_LIBRARY_PATH=.:$LD_LIBRARY_PATH
export DYLD_LIBRARY_PATH=.:$DYLD_LIBRARY_PATH

# Run benchmark
java -cp ".:jna-5.13.0.jar:gson-2.10.1.jar" Benchmark "$@"
`
}
