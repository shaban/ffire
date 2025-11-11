package benchmark

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/shaban/ffire/pkg/fixture"
	"github.com/shaban/ffire/pkg/generator"
	"github.com/shaban/ffire/pkg/schema"
)

// GenerateCpp creates a complete C++ benchmark executable in the output directory.
func GenerateCpp(s *schema.Schema, schemaName string, messageName string, jsonData []byte, outputDir string, iterations int) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the encoder/decoder code
	generatedCode, err := generator.GenerateCpp(s)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Write generated code
	generatedFile := filepath.Join(outputDir, "generated.hpp")
	if err := os.WriteFile(generatedFile, generatedCode, 0644); err != nil {
		return fmt.Errorf("failed to write generated code: %w", err)
	}

	// Convert JSON to binary fixture
	binaryData, err := fixture.Convert(s, messageName, jsonData)
	if err != nil {
		return fmt.Errorf("failed to convert fixture: %w", err)
	}

	// Write binary fixture as C++ byte array
	fixtureCode := generateCppFixture(binaryData)
	fixtureFile := filepath.Join(outputDir, "fixture.hpp")
	if err := os.WriteFile(fixtureFile, []byte(fixtureCode), 0644); err != nil {
		return fmt.Errorf("failed to write fixture: %w", err)
	}

	// Find the message type
	var messageType *schema.MessageType
	for i := range s.Messages {
		if s.Messages[i].Name == messageName {
			messageType = &s.Messages[i]
			break
		}
	}
	if messageType == nil {
		return fmt.Errorf("message type %s not found", messageName)
	}

	// Determine root type name for function naming
	rootTypeName := getRootTypeName(messageType.TargetType)

	// Generate benchmark main
	benchData := CppBenchmarkData{
		Namespace:    s.Package,
		SchemaName:   schemaName,
		MessageName:  messageName,
		TypeName:     rootTypeName,
		Iterations:   iterations,
		FixtureBytes: len(binaryData),
	}

	var buf bytes.Buffer
	if err := cppBenchTemplate.Execute(&buf, benchData); err != nil {
		return fmt.Errorf("failed to generate benchmark: %w", err)
	}

	// Write benchmark main
	benchFile := filepath.Join(outputDir, "bench.cpp")
	if err := os.WriteFile(benchFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write benchmark main: %w", err)
	}

	// Generate CMakeLists.txt
	var cmakeBuf bytes.Buffer
	if err := cmakeTemplate.Execute(&cmakeBuf, benchData); err != nil {
		return fmt.Errorf("failed to generate CMakeLists.txt: %w", err)
	}

	cmakeFile := filepath.Join(outputDir, "CMakeLists.txt")
	if err := os.WriteFile(cmakeFile, cmakeBuf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write CMakeLists.txt: %w", err)
	}

	// Generate Makefile (fallback for systems without cmake)
	var makeBuf bytes.Buffer
	if err := makefileTemplate.Execute(&makeBuf, benchData); err != nil {
		return fmt.Errorf("failed to generate Makefile: %w", err)
	}

	makeFile := filepath.Join(outputDir, "Makefile")
	if err := os.WriteFile(makeFile, makeBuf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write Makefile: %w", err)
	}

	return nil
}

// CppBenchmarkData holds template data for C++ benchmark generation.
type CppBenchmarkData struct {
	Namespace    string
	SchemaName   string
	MessageName  string
	TypeName     string
	Iterations   int
	FixtureBytes int
}

// generateCppFixture converts binary data to a C++ byte array
func generateCppFixture(data []byte) string {
	var buf bytes.Buffer
	buf.WriteString("#ifndef FIXTURE_HPP\n")
	buf.WriteString("#define FIXTURE_HPP\n\n")
	buf.WriteString("#include <cstdint>\n")
	buf.WriteString("#include <cstddef>\n\n")
	buf.WriteString("const uint8_t FIXTURE_DATA[] = {\n")

	for i, b := range data {
		if i > 0 {
			buf.WriteString(", ")
		}
		if i%16 == 0 {
			buf.WriteString("\n    ")
		}
		fmt.Fprintf(&buf, "0x%02x", b)
	}

	buf.WriteString("\n};\n\n")
	fmt.Fprintf(&buf, "const size_t FIXTURE_SIZE = %d;\n\n", len(data))
	buf.WriteString("#endif // FIXTURE_HPP\n")
	return buf.String()
}

var cppBenchTemplate = template.Must(template.New("cppbench").Funcs(template.FuncMap{
	"ToLower": strings.ToLower,
}).Parse(`#include <iostream>
#include <iomanip>
#include <chrono>
#include <vector>
#include <cstring>
#include "generated.hpp"
#include "fixture.hpp"

using namespace std::chrono;

struct BenchResult {
    const char* language;
    const char* format;
    const char* message;
    int iterations;
    int64_t encode_ns;
    int64_t decode_ns;
    int64_t total_ns;
    size_t wire_size;
    size_t fixture_size;
};

int main(int argc, char** argv) {
    const int iterations = {{.Iterations}};
    const bool json_output = std::getenv("BENCH_JSON") != nullptr;
    
    try {
        // Decode fixture to get original data
        auto original = {{.Namespace}}::decode_{{.TypeName | ToLower}}_message(FIXTURE_DATA, FIXTURE_SIZE);
        
        // Warmup
        for (int i = 0; i < 1000; ++i) {
            auto encoded = {{.Namespace}}::encode_{{.TypeName | ToLower}}_message(original);
            auto decoded = {{.Namespace}}::decode_{{.TypeName | ToLower}}_message(encoded);
        }
        
        // Benchmark encode
        auto encode_start = high_resolution_clock::now();
        std::vector<uint8_t> encoded;
        for (int i = 0; i < iterations; ++i) {
            encoded = {{.Namespace}}::encode_{{.TypeName | ToLower}}_message(original);
        }
        auto encode_end = high_resolution_clock::now();
        auto encode_time = duration_cast<nanoseconds>(encode_end - encode_start).count();
        
        // Benchmark decode
        auto decode_start = high_resolution_clock::now();
        for (int i = 0; i < iterations; ++i) {
            auto decoded = {{.Namespace}}::decode_{{.TypeName | ToLower}}_message(encoded);
        }
        auto decode_end = high_resolution_clock::now();
        auto decode_time = duration_cast<nanoseconds>(decode_end - decode_start).count();
        
        // Calculate metrics
        int64_t encode_ns = encode_time / iterations;
        int64_t decode_ns = decode_time / iterations;
        int64_t total_ns = encode_ns + decode_ns;
        
        if (json_output) {
            // Output JSON for automation
            std::cout << "{"
                      << "\"language\":\"C++\","
                      << "\"format\":\"ffire\","
                      << "\"message\":\"{{.SchemaName}}\","
                      << "\"iterations\":" << iterations << ","
                      << "\"encode_ns\":" << encode_ns << ","
                      << "\"decode_ns\":" << decode_ns << ","
                      << "\"total_ns\":" << total_ns << ","
                      << "\"wire_size\":" << encoded.size() << ","
                      << "\"fixture_size\":" << FIXTURE_SIZE
                      << "}\n";
        } else {
            // Print human-readable results
            std::cout << "ffire benchmark: {{.SchemaName}}\n";
            std::cout << "Iterations:  " << iterations << "\n";
            std::cout << "Encode:      " << encode_ns << " ns/op\n";
            std::cout << "Decode:      " << decode_ns << " ns/op\n";
            std::cout << "Total:       " << total_ns << " ns/op\n";
            std::cout << "Wire size:   " << encoded.size() << " bytes\n";
            std::cout << "Fixture:     " << FIXTURE_SIZE << " bytes\n";
            std::cout << "Total time:  " << std::fixed << std::setprecision(2) 
                      << (encode_time + decode_time) / 1e9 << "s\n";
        }
        
        return 0;
    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << "\n";
        return 1;
    }
}
`))

var cmakeTemplate = template.Must(template.New("cmake").Parse(`cmake_minimum_required(VERSION 3.10)
project(ffire_bench)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Enable optimizations
set(CMAKE_CXX_FLAGS_RELEASE "-O3 -march=native")
set(CMAKE_BUILD_TYPE Release)

add_executable(bench bench.cpp)
`))

var makefileTemplate = template.Must(template.New("makefile").Parse(`# Makefile for ffire benchmark
# Works with clang++ (macOS Xcode tools) or g++ (Linux)

CXX := $(shell command -v clang++ 2>/dev/null || echo g++)
CXXFLAGS := -std=c++17 -O3 -march=native -Wall
TARGET := bench
SOURCES := bench.cpp
HEADERS := generated.hpp fixture.hpp

.PHONY: all clean

all: $(TARGET)

$(TARGET): $(SOURCES) $(HEADERS)
	$(CXX) $(CXXFLAGS) -o $(TARGET) $(SOURCES)
	@echo "Built $(TARGET) successfully"

clean:
	rm -f $(TARGET)

run: $(TARGET)
	./$(TARGET)
`))
