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

	// Step 1: Generate the Java code for all message types
	javaDir := filepath.Join(outputDir, "java")
	if err := os.MkdirAll(javaDir, 0755); err != nil {
		return fmt.Errorf("failed to create java directory: %w", err)
	}

	// Generate Java code for the schema
	javaCode, err := generator.GenerateJava(schema)
	if err != nil {
		return fmt.Errorf("failed to generate Java code: %w", err)
	}

	// Write the generated message class
	messagePath := filepath.Join(javaDir, messageName+"Message.java")
	if err := os.WriteFile(messagePath, []byte(javaCode), 0644); err != nil {
		return fmt.Errorf("failed to write message class: %w", err)
	}

	// Step 2: Convert JSON to binary fixture
	binaryData, err := fixture.Convert(schema, messageName, jsonData)
	if err != nil {
		return fmt.Errorf("failed to convert JSON to binary: %w", err)
	}

	// Step 3: Write the binary fixture
	fixturePath := filepath.Join(javaDir, "fixture.bin")
	if err := os.WriteFile(fixturePath, binaryData, 0644); err != nil {
		return fmt.Errorf("failed to write fixture: %w", err)
	}

	// Step 4: Generate the benchmark harness
	benchmarkCode := generateJavaBenchmarkCode(schema.Package, schemaName, messageName, iterations)
	benchPath := filepath.Join(javaDir, "Bench.java")
	if err := os.WriteFile(benchPath, []byte(benchmarkCode), 0644); err != nil {
		return fmt.Errorf("failed to write benchmark: %w", err)
	}

	return nil
}

// generateJavaBenchmarkCode generates the benchmark harness code for native Java
func generateJavaBenchmarkCode(packageName, schemaName, messageName string, iterations int) string {
	buf := &bytes.Buffer{}

	// Write imports
	fmt.Fprintf(buf, "import %s.%sMessage;\n", packageName, messageName)
	buf.WriteString(`import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.time.Instant;

public class Bench {
    public static void main(String[] args) throws IOException {
        // Load fixture
        byte[] fixtureData = Files.readAllBytes(Paths.get("fixture.bin"));
        
`)
	fmt.Fprintf(buf, "        int iterations = %d;\n", iterations)
	buf.WriteString(`        boolean jsonOutput = "1".equals(System.getenv("BENCH_JSON"));
        
        // Warmup
        for (int i = 0; i < 1000; i++) {
`)
	fmt.Fprintf(buf, "            %sMessage msg = %sMessage.decode(fixtureData);\n", messageName, messageName)
	buf.WriteString(`            byte[] encoded = msg.encode();
        }
        
        // Benchmark decode
        long decodeStart = System.nanoTime();
        for (int i = 0; i < iterations; i++) {
`)
	fmt.Fprintf(buf, "            %sMessage msg = %sMessage.decode(fixtureData);\n", messageName, messageName)
	buf.WriteString(`        }
        long decodeEnd = System.nanoTime();
        long decodeTimeNs = decodeEnd - decodeStart;
        
        // Benchmark encode (decode once, then encode many times)
`)
	fmt.Fprintf(buf, "        %sMessage msg = %sMessage.decode(fixtureData);\n", messageName, messageName)
	buf.WriteString(`        long encodeStart = System.nanoTime();
        byte[] encoded = null;
        for (int i = 0; i < iterations; i++) {
            encoded = msg.encode();
        }
        long encodeEnd = System.nanoTime();
        long encodeTimeNs = encodeEnd - encodeStart;
        
        // Calculate metrics
        long encodeNs = encodeTimeNs / iterations;
        long decodeNs = decodeTimeNs / iterations;
        long totalNs = encodeNs + decodeNs;
        
        if (jsonOutput) {
            // Output JSON for automation
            System.out.println("{");
            System.out.println("  \"language\": \"Java\",");
            System.out.println("  \"format\": \"ffire\",");
`)
	fmt.Fprintf(buf, "            System.out.println(\"  \\\"message\\\": \\\"%s\\\",\");\n", schemaName)
	buf.WriteString(`            System.out.println("  \"iterations\": " + iterations + ",");
            System.out.println("  \"encode_ns\": " + encodeNs + ",");
            System.out.println("  \"decode_ns\": " + decodeNs + ",");
            System.out.println("  \"total_ns\": " + totalNs + ",");
            System.out.println("  \"wire_size\": " + encoded.length + ",");
            System.out.println("  \"fixture_size\": " + fixtureData.length + ",");
            System.out.println("  \"timestamp\": \"" + Instant.now().toString() + "\"");
            System.out.println("}");
        } else {
            // Print human-readable results
`)
	fmt.Fprintf(buf, "            System.out.println(\"ffire benchmark: %s\");\n", schemaName)
	buf.WriteString(`            System.out.println("Iterations:  " + iterations);
            System.out.println("Encode:      " + encodeNs + " ns/op");
            System.out.println("Decode:      " + decodeNs + " ns/op");
            System.out.println("Total:       " + totalNs + " ns/op");
            System.out.println("Wire size:   " + encoded.length + " bytes");
            System.out.println("Fixture:     " + fixtureData.length + " bytes");
            double totalTimeS = (encodeTimeNs + decodeTimeNs) / 1e9;
            System.out.printf("Total time:  %.3fs%n", totalTimeS);
        }
    }
}
`)

	return buf.String()
}
