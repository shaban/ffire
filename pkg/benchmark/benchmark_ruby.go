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

// GenerateRuby generates a Ruby benchmark with embedded fixture
func GenerateRuby(schema *schema.Schema, schemaName, messageName string, jsonData []byte, outputDir string, iterations int) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Step 1: Generate the Ruby package
	config := &generator.PackageConfig{
		Schema:    schema,
		Language:  "ruby",
		OutputDir: outputDir,
		Namespace: schemaName,
		Optimize:  2,
		Platform:  "current",
		Arch:      "current",
		NoCompile: false,
		Verbose:   false,
	}

	if err := generator.GeneratePackage(config); err != nil {
		return fmt.Errorf("failed to generate Ruby package: %w", err)
	}

	// Step 2: Convert JSON to binary fixture
	binaryData, err := fixture.Convert(schema, messageName, jsonData)
	if err != nil {
		return fmt.Errorf("failed to convert JSON to binary: %w", err)
	}

	// Step 3: Write the binary fixture
	rubyDir := filepath.Join(outputDir, "ruby")
	fixturePath := filepath.Join(rubyDir, "fixture.bin")
	if err := os.WriteFile(fixturePath, binaryData, 0644); err != nil {
		return fmt.Errorf("failed to write fixture: %w", err)
	}

	// Step 4: Generate the benchmark harness
	benchmarkCode := generateRubyBenchmarkCode(schemaName, messageName, iterations)
	benchPath := filepath.Join(rubyDir, "bench.rb")
	if err := os.WriteFile(benchPath, []byte(benchmarkCode), 0644); err != nil {
		return fmt.Errorf("failed to write benchmark: %w", err)
	}

	// Step 5: Generate a run script for convenience
	runScript := generateRubyRunScript()
	runPath := filepath.Join(rubyDir, "run.sh")
	if err := os.WriteFile(runPath, []byte(runScript), 0755); err != nil {
		return fmt.Errorf("failed to write run script: %w", err)
	}

	return nil
}

// generateRubyBenchmarkCode generates the benchmark harness code
func generateRubyBenchmarkCode(schemaName, messageName string, iterations int) string {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, `require 'ffi'
require 'json'

module FFire%s
  extend FFI::Library
  
  # Load shared library
  lib_name = case RbConfig::CONFIG['host_os']
             when /darwin/
               'lib%s.dylib'
             when /mswin|mingw|cygwin/
               '%s.dll'
             else
               'lib%s.so'
             end
  
  ffi_lib File.join(__dir__, lib_name)
  
  # FFI function declarations
  attach_function :ffire_encode_%s, [:pointer, :int32, :pointer], :pointer
  attach_function :ffire_decode_%s, [:pointer, :int32], :pointer
  attach_function :ffire_free_%s, [:pointer], :void
  
  def self.decode(data)
    ptr = FFI::MemoryPointer.new(:char, data.bytesize)
    ptr.write_bytes(data)
    result = ffire_decode_%s(ptr, data.bytesize)
    raise 'Decode failed' if result.null?
    result
  end
  
  def self.encode(msg_ptr)
    size_ptr = FFI::MemoryPointer.new(:int32)
    result = ffire_encode_%s(msg_ptr, 0, size_ptr)
    raise 'Encode failed' if result.null?
    size = size_ptr.read_int32
    result.read_bytes(size)
  end
  
  def self.free_message(msg_ptr)
    ffire_free_%s(msg_ptr)
  end
end

# Load fixture
fixture_data = File.binread('fixture.bin')

iterations = %d
json_output = ENV['BENCH_JSON'] == '1'

# Warmup
1000.times do
  msg_ptr = FFire%s.decode(fixture_data)
  encoded = FFire%s.encode(msg_ptr)
  FFire%s.free_message(msg_ptr)
end

# Benchmark decode
decode_start = Process.clock_gettime(Process::CLOCK_MONOTONIC, :nanosecond)
iterations.times do
  msg_ptr = FFire%s.decode(fixture_data)
  FFire%s.free_message(msg_ptr)
end
decode_end = Process.clock_gettime(Process::CLOCK_MONOTONIC, :nanosecond)
decode_time_ns = decode_end - decode_start

# Benchmark encode (decode once, then encode many times)
msg_ptr = FFire%s.decode(fixture_data)
encode_start = Process.clock_gettime(Process::CLOCK_MONOTONIC, :nanosecond)
encoded = nil
iterations.times do
  encoded = FFire%s.encode(msg_ptr)
end
encode_end = Process.clock_gettime(Process::CLOCK_MONOTONIC, :nanosecond)
encode_time_ns = encode_end - encode_start
FFire%s.free_message(msg_ptr)

# Calculate metrics
encode_ns = (encode_time_ns / iterations).round
decode_ns = (decode_time_ns / iterations).round
total_ns = encode_ns + decode_ns

if json_output
  # Output JSON for automation
  result = {
    language: 'Ruby',
    format: 'ffire',
    message: '%s',
    iterations: iterations,
    encode_ns: encode_ns,
    decode_ns: decode_ns,
    total_ns: total_ns,
    wire_size: encoded.bytesize,
    fixture_size: fixture_data.bytesize,
    timestamp: Time.now.iso8601
  }
  puts JSON.generate(result)
else
  # Print human-readable results
  puts "ffire benchmark: %s"
  puts "Iterations:  #{iterations}"
  puts "Encode:      #{encode_ns} ns/op"
  puts "Decode:      #{decode_ns} ns/op"
  puts "Total:       #{total_ns} ns/op"
  puts "Wire size:   #{encoded.bytesize} bytes"
  puts "Fixture:     #{fixture_data.bytesize} bytes"
  total_time_s = (encode_time_ns + decode_time_ns) / 1_000_000_000.0
  puts "Total time:  #{'%%.3f' %% total_time_s}s"
end
`, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName,
		iterations, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName, schemaName,
		schemaName, schemaName)

	return buf.String()
}

// generateRubyRunScript generates a convenience run script
func generateRubyRunScript() string {
	return `#!/bin/bash
# Convenience script to run Ruby benchmark

# Check if ruby is available
if ! command -v ruby &> /dev/null; then
    echo "Error: ruby not found"
    exit 1
fi

# Install FFI gem if needed
if ! ruby -e "require 'ffi'" 2>/dev/null; then
    echo "Installing ffi gem..."
    gem install ffi
fi

# Run benchmark
ruby bench.rb "$@"
`
}
