package benchmark

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/shaban/ffire/pkg/fixture"
	"github.com/shaban/ffire/pkg/generator"
	"github.com/shaban/ffire/pkg/schema"
)

// GenerateCSharp generates a native C# benchmark with embedded fixture
func GenerateCSharp(schema *schema.Schema, schemaName, messageName string, jsonData []byte, outputDir string, iterations int) error {
	// Create output directories
	csharpDir := filepath.Join(outputDir, "csharp")
	if err := os.MkdirAll(csharpDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Step 1: Generate C# code
	csCode, err := generator.GenerateCSharp(schema)
	if err != nil {
		return fmt.Errorf("failed to generate C# code: %w", err)
	}

	csPath := filepath.Join(csharpDir, "Generated.cs")
	if err := os.WriteFile(csPath, csCode, 0644); err != nil {
		return fmt.Errorf("failed to write C# code: %w", err)
	}

	// Step 2: Convert JSON to binary fixture
	binaryData, err := fixture.Convert(schema, messageName, jsonData)
	if err != nil {
		return fmt.Errorf("failed to convert JSON to binary: %w", err)
	}

	fixturePath := filepath.Join(csharpDir, "fixture.bin")
	if err := os.WriteFile(fixturePath, binaryData, 0644); err != nil {
		return fmt.Errorf("failed to write fixture: %w", err)
	}

	// Step 3: Generate benchmark harness
	namespace := schema.Package
	if namespace == "" {
		namespace = "FFire.Generated"
	} else {
		namespace = toPascalCase(namespace)
	}
	benchmarkCode := generateCSharpBenchmarkCode(namespace, messageName, iterations, schemaName)
	benchPath := filepath.Join(csharpDir, "Bench.cs")
	if err := os.WriteFile(benchPath, []byte(benchmarkCode), 0644); err != nil {
		return fmt.Errorf("failed to write benchmark: %w", err)
	}

	// Step 4: Generate project file
	projectCode := generateCSharpProjectFile()
	projectPath := filepath.Join(csharpDir, "Bench.csproj")
	if err := os.WriteFile(projectPath, []byte(projectCode), 0644); err != nil {
		return fmt.Errorf("failed to write project file: %w", err)
	}

	return nil
}

// toPascalCase converts a string to PascalCase (capitalize first letter)
func toPascalCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// generateCSharpBenchmarkCode generates the benchmark harness code
func generateCSharpBenchmarkCode(namespace, messageName string, iterations int, benchName string) string {
	// Append "Message" if not already present (C# generator adds it)
	csMessageName := messageName
	if !strings.HasSuffix(messageName, "Message") {
		csMessageName = messageName + "Message"
	}

	return fmt.Sprintf(`using System;
using System.Diagnostics;
using System.IO;
using System.Text.Json;

namespace FFire.Benchmark
{
    public class Bench
    {
        public static void Main(string[] args)
        {
            const int iterations = %d;
            byte[] fixture = File.ReadAllBytes("fixture.bin");

            // Warmup
            %s.%s warmupMsg = default;
            byte[] warmupEncoded = null;
            for (int i = 0; i < 100; i++)
            {
                warmupMsg = %s.%s.Decode(fixture);
                warmupEncoded = warmupMsg.Encode();
            }

            // Measure: decode then encode
            Stopwatch sw = Stopwatch.StartNew();
            long t0 = sw.ElapsedTicks;
            
            %s.%s msg = default;
            byte[] encoded = null;
            for (int i = 0; i < iterations; i++)
            {
                msg = %s.%s.Decode(fixture);
            }
            long t1 = sw.ElapsedTicks;
            
            for (int i = 0; i < iterations; i++)
            {
                encoded = msg.Encode();
            }
            long t2 = sw.ElapsedTicks;

            long decodeNs = (long)((double)(t1 - t0) / Stopwatch.Frequency * 1_000_000_000);
            long encodeNs = (long)((double)(t2 - t1) / Stopwatch.Frequency * 1_000_000_000);

            // Calculate per-operation metrics
            long encodeNsPerOp = encodeNs / iterations;
            long decodeNsPerOp = decodeNs / iterations;
            long totalNsPerOp = encodeNsPerOp + decodeNsPerOp;

            var result = new
            {
                language = "C#",
                format = "ffire",
                message = "%s",
                iterations = iterations,
                encode_ns = encodeNsPerOp,
                decode_ns = decodeNsPerOp,
                total_ns = totalNsPerOp,
                wire_size = encoded.Length,
                fixture_size = fixture.Length,
                timestamp = ""
            };

            Console.WriteLine(JsonSerializer.Serialize(result));
        }
    }
}
`, iterations, namespace, csMessageName, namespace, csMessageName, namespace, csMessageName, namespace, csMessageName, benchName)
}

// generateCSharpProjectFile generates the .csproj file
func generateCSharpProjectFile() string {
	return `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <OutputType>Exe</OutputType>
    <TargetFramework>net9.0</TargetFramework>
    <AllowUnsafeBlocks>true</AllowUnsafeBlocks>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
  </PropertyGroup>
</Project>
`
}
