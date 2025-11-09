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

// GenerateCSharp generates a C# benchmark with embedded fixture
func GenerateCSharp(schema *schema.Schema, schemaName, messageName string, jsonData []byte, outputDir string, iterations int) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Step 1: Generate the C# package
	// Use package name (not schema filename) as module name
	config := &generator.PackageConfig{
		Schema:    schema,
		Language:  "csharp",
		OutputDir: outputDir,
		Namespace: schema.Package,
		Optimize:  2,
		Platform:  "current",
		Arch:      "current",
		NoCompile: false,
		Verbose:   false,
	}

	if err := generator.GeneratePackage(config); err != nil {
		return fmt.Errorf("failed to generate C# package: %w", err)
	}

	// Step 2: Convert JSON to binary fixture
	binaryData, err := fixture.Convert(schema, messageName, jsonData)
	if err != nil {
		return fmt.Errorf("failed to convert JSON to binary: %w", err)
	}

	// Step 3: Write the binary fixture
	csharpDir := filepath.Join(outputDir, "csharp")
	fixturePath := filepath.Join(csharpDir, "fixture.bin")
	if err := os.WriteFile(fixturePath, binaryData, 0644); err != nil {
		return fmt.Errorf("failed to write fixture: %w", err)
	}

	// Step 4: Generate the benchmark harness
	benchmarkCode := generateCSharpBenchmarkCode(schemaName, messageName, iterations)
	benchPath := filepath.Join(csharpDir, "Benchmark.cs")
	if err := os.WriteFile(benchPath, []byte(benchmarkCode), 0644); err != nil {
		return fmt.Errorf("failed to write benchmark: %w", err)
	}

	// Step 5: Generate project file
	projectCode := generateCSharpProjectFile(schemaName)
	projectPath := filepath.Join(csharpDir, "Benchmark.csproj")
	if err := os.WriteFile(projectPath, []byte(projectCode), 0644); err != nil {
		return fmt.Errorf("failed to write project file: %w", err)
	}

	// Step 6: Generate a run script for convenience
	runScript := generateCSharpRunScript()
	runPath := filepath.Join(csharpDir, "run.sh")
	if err := os.WriteFile(runPath, []byte(runScript), 0755); err != nil {
		return fmt.Errorf("failed to write run script: %w", err)
	}

	return nil
}

// generateCSharpBenchmarkCode generates the benchmark harness code
func generateCSharpBenchmarkCode(schemaName, messageName string, iterations int) string {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, `using System;
using System.IO;
using System.Runtime.InteropServices;
using System.Diagnostics;
using System.Text.Json;
using System.Collections.Generic;

class Benchmark
{
    // P/Invoke declarations
    private const string LibName = "lib%s";
    
    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    private static extern IntPtr ffire_decode_%s(IntPtr data, int size);
    
    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    private static extern IntPtr ffire_encode_%s(IntPtr msg, int flags, out int size);
    
    [DllImport(LibName, CallingConvention = CallingConvention.Cdecl)]
    private static extern void ffire_free_%s(IntPtr msg);
    
    private static IntPtr Decode(byte[] data)
    {
        IntPtr dataPtr = Marshal.AllocHGlobal(data.Length);
        Marshal.Copy(data, 0, dataPtr, data.Length);
        IntPtr result = ffire_decode_%s(dataPtr, data.Length);
        Marshal.FreeHGlobal(dataPtr);
        
        if (result == IntPtr.Zero)
        {
            throw new Exception("Decode failed");
        }
        return result;
    }
    
    private static byte[] Encode(IntPtr msgPtr)
    {
        int size;
        IntPtr result = ffire_encode_%s(msgPtr, 0, out size);
        
        if (result == IntPtr.Zero)
        {
            throw new Exception("Encode failed");
        }
        
        byte[] encoded = new byte[size];
        Marshal.Copy(result, encoded, 0, size);
        return encoded;
    }
    
    private static void FreeMessage(IntPtr msgPtr)
    {
        ffire_free_%s(msgPtr);
    }
    
    static void Main(string[] args)
    {
        // Load fixture
        byte[] fixtureData = File.ReadAllBytes("fixture.bin");
        
        int iterations = %d;
        bool jsonOutput = Environment.GetEnvironmentVariable("BENCH_JSON") == "1";
        
        // Warmup
        for (int i = 0; i < 1000; i++)
        {
            IntPtr msgPtr = Decode(fixtureData);
            byte[] encoded = Encode(msgPtr);
            FreeMessage(msgPtr);
        }
        
        // Benchmark decode
        Stopwatch decodeWatch = Stopwatch.StartNew();
        for (int i = 0; i < iterations; i++)
        {
            IntPtr msgPtr = Decode(fixtureData);
            FreeMessage(msgPtr);
        }
        decodeWatch.Stop();
        long decodeTimeNs = decodeWatch.ElapsedTicks * 1000000000 / Stopwatch.Frequency;
        
        // Benchmark encode (decode once, then encode many times)
        IntPtr msg = Decode(fixtureData);
        Stopwatch encodeWatch = Stopwatch.StartNew();
        byte[] encoded = null;
        for (int i = 0; i < iterations; i++)
        {
            encoded = Encode(msg);
        }
        encodeWatch.Stop();
        long encodeTimeNs = encodeWatch.ElapsedTicks * 1000000000 / Stopwatch.Frequency;
        FreeMessage(msg);
        
        // Calculate metrics
        long encodeNs = encodeTimeNs / iterations;
        long decodeNs = decodeTimeNs / iterations;
        long totalNs = encodeNs + decodeNs;
        
        if (jsonOutput)
        {
            // Output JSON for automation
            var result = new Dictionary<string, object>
            {
                ["language"] = "C#",
                ["format"] = "ffire",
                ["message"] = "%s",
                ["iterations"] = iterations,
                ["encode_ns"] = encodeNs,
                ["decode_ns"] = decodeNs,
                ["total_ns"] = totalNs,
                ["wire_size"] = encoded.Length,
                ["fixture_size"] = fixtureData.Length,
                ["timestamp"] = DateTime.UtcNow.ToString("o")
            };
            Console.WriteLine(JsonSerializer.Serialize(result));
        }
        else
        {
            // Print human-readable results
            Console.WriteLine("ffire benchmark: %s");
            Console.WriteLine($"Iterations:  {iterations}");
            Console.WriteLine($"Encode:      {encodeNs} ns/op");
            Console.WriteLine($"Decode:      {decodeNs} ns/op");
            Console.WriteLine($"Total:       {totalNs} ns/op");
            Console.WriteLine($"Wire size:   {encoded.Length} bytes");
            Console.WriteLine($"Fixture:     {fixtureData.Length} bytes");
            double totalTimeS = (encodeTimeNs + decodeTimeNs) / 1e9;
            Console.WriteLine($"Total time:  {totalTimeS:F3}s");
        }
    }
}
`, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName, iterations, schemaName, schemaName)

	return buf.String()
}

// generateCSharpProjectFile generates the .csproj file
func generateCSharpProjectFile(schemaName string) string {
	_ = schemaName // Reserved for future use
	return `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <OutputType>Exe</OutputType>
    <TargetFramework>net6.0</TargetFramework>
    <AllowUnsafeBlocks>true</AllowUnsafeBlocks>
  </PropertyGroup>
</Project>
`
}

// generateCSharpRunScript generates a convenience run script
func generateCSharpRunScript() string {
	return `#!/bin/bash
# Convenience script to run C# benchmark

# Check if dotnet is available
if ! command -v dotnet &> /dev/null; then
    echo "Error: dotnet not found"
    exit 1
fi

# Check .NET version
DOTNET_VERSION=$(dotnet --version | cut -d. -f1)
if [ "$DOTNET_VERSION" -lt 6 ]; then
    echo "Error: .NET 6.0+ required"
    exit 1
fi

# Build if needed
if [ ! -d "bin" ] || [ "Benchmark.cs" -nt "bin" ]; then
    echo "Building..."
    dotnet build -c Release
fi

# Set library path
export LD_LIBRARY_PATH=.:$LD_LIBRARY_PATH
export DYLD_LIBRARY_PATH=.:$DYLD_LIBRARY_PATH

# Run benchmark
dotnet run -c Release --no-build "$@"
`
}
