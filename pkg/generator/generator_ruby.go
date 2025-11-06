package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// generateRubyWrapper generates the Ruby FFI bindings module
func generateRubyWrapper(config *PackageConfig, gemDir string) error {
	// Create lib/<gemname> directory
	libDir := filepath.Join(gemDir, "lib", config.Namespace)
	if err := os.MkdirAll(libDir, 0755); err != nil {
		return fmt.Errorf("failed to create lib directory: %w", err)
	}

	// Generate bindings.rb (FFI declarations)
	if err := generateRubyBindings(config, libDir); err != nil {
		return err
	}

	// Generate message wrapper classes
	if err := generateRubyMessageClasses(config, libDir); err != nil {
		return err
	}

	// Generate version.rb
	if err := generateRubyVersion(config, libDir); err != nil {
		return err
	}

	// Generate main module file lib/<gemname>.rb
	if err := generateRubyMainModule(config, gemDir); err != nil {
		return err
	}

	if config.Verbose {
		fmt.Println("✓ Generated Ruby FFI bindings")
	}

	return nil
}

// generateRubyBindings generates the FFI bindings (bindings.rb)
func generateRubyBindings(config *PackageConfig, libDir string) error {
	buf := &bytes.Buffer{}

	// Module header
	className := toRubyClassName(config.Namespace)
	fmt.Fprintf(buf, "# frozen_string_literal: true\n\n")
	fmt.Fprintf(buf, "require 'ffi'\n\n")
	fmt.Fprintf(buf, "module %s\n", className)
	fmt.Fprintf(buf, "  # FFI bindings to the native library\n")
	fmt.Fprintf(buf, "  module Bindings\n")
	fmt.Fprintf(buf, "    extend FFI::Library\n\n")

	// Platform detection and library loading
	buf.WriteString("    # Determine library name based on platform\n")
	buf.WriteString("    LIB_NAME = case RbConfig::CONFIG['host_os']\n")
	buf.WriteString("               when /darwin/i\n")
	buf.WriteString("                 'libffire.dylib'\n")
	buf.WriteString("               when /linux/i\n")
	buf.WriteString("                 'libffire.so'\n")
	buf.WriteString("               when /mswin|mingw|cygwin/i\n")
	buf.WriteString("                 'ffire.dll'\n")
	buf.WriteString("               else\n")
	buf.WriteString("                 'libffire.so'\n")
	buf.WriteString("               end\n\n")

	buf.WriteString("    # Load the native library\n")
	buf.WriteString("    lib_path = File.expand_path(File.join(File.dirname(__FILE__), '..', LIB_NAME))\n")
	buf.WriteString("    ffi_lib lib_path\n\n")

	// Generate FFI function declarations for each message type
	for _, msg := range config.Schema.Messages {
		msgName := strings.ToLower(msg.Name)

		buf.WriteString(fmt.Sprintf("    # FFI declarations for %s\n", msg.Name))
		buf.WriteString(fmt.Sprintf("    attach_function :%s_decode, [:pointer, :size_t, :pointer], :pointer\n", msgName))
		buf.WriteString(fmt.Sprintf("    attach_function :%s_encode, [:pointer, :pointer, :pointer], :size_t\n", msgName))
		buf.WriteString(fmt.Sprintf("    attach_function :%s_free, [:pointer], :void\n", msgName))
		buf.WriteString(fmt.Sprintf("    attach_function :%s_free_data, [:pointer], :void\n", msgName))
		buf.WriteString(fmt.Sprintf("    attach_function :%s_free_error, [:pointer], :void\n", msgName))
		buf.WriteString("\n")
	}

	buf.WriteString("  end\n")
	buf.WriteString("end\n")

	bindingsPath := filepath.Join(libDir, "bindings.rb")
	return os.WriteFile(bindingsPath, buf.Bytes(), 0644)
}

// generateRubyMessageClasses generates Ruby wrapper classes for messages
func generateRubyMessageClasses(config *PackageConfig, libDir string) error {
	className := toRubyClassName(config.Namespace)

	for _, msg := range config.Schema.Messages {
		buf := &bytes.Buffer{}

		msgClassName := toRubyClassName(msg.Name)
		msgName := strings.ToLower(msg.Name)

		// File header
		fmt.Fprintf(buf, "# frozen_string_literal: true\n\n")
		fmt.Fprintf(buf, "require_relative 'bindings'\n\n")
		fmt.Fprintf(buf, "module %s\n", className)
		fmt.Fprintf(buf, "  # %s message wrapper\n", msg.Name)
		fmt.Fprintf(buf, "  class %s\n", msgClassName)
		fmt.Fprintf(buf, "    # @return [FFI::Pointer] the native handle\n")
		fmt.Fprintf(buf, "    attr_reader :handle\n\n")

		// Constructor (private - use decode)
		buf.WriteString("    # @api private\n")
		buf.WriteString("    def initialize(handle)\n")
		buf.WriteString("      @handle = handle\n")
		buf.WriteString("      @freed = false\n")
		buf.WriteString("      # Register finalizer to free native resources\n")
		buf.WriteString("      ObjectSpace.define_finalizer(self, self.class.finalizer(handle))\n")
		buf.WriteString("    end\n\n")

		// Finalizer (class method)
		buf.WriteString("    # @api private\n")
		buf.WriteString("    def self.finalizer(handle)\n")
		buf.WriteString("      proc do\n")
		fmt.Fprintf(buf, "        Bindings.%s_free(handle) unless handle.null?\n", msgName)
		buf.WriteString("      end\n")
		buf.WriteString("    end\n\n")

		// Decode class method
		buf.WriteString("    # Decode a message from binary data\n")
		buf.WriteString("    #\n")
		buf.WriteString("    # @param data [String, Array<Integer>] the binary data to decode\n")
		buf.WriteString("    # @return [#{msgClassName}] the decoded message\n")
		buf.WriteString("    # @raise [RuntimeError] if decoding fails\n")
		buf.WriteString("    def self.decode(data)\n")
		buf.WriteString("      # Convert to byte array if needed\n")
		buf.WriteString("      data = data.bytes if data.is_a?(String)\n")
		buf.WriteString("      \n")
		buf.WriteString("      # Create FFI memory buffer\n")
		buf.WriteString("      buf = FFI::MemoryPointer.new(:uint8, data.size)\n")
		buf.WriteString("      buf.write_array_of_uint8(data)\n")
		buf.WriteString("      \n")
		buf.WriteString("      # Error pointer\n")
		buf.WriteString("      error_ptr = FFI::MemoryPointer.new(:pointer)\n")
		buf.WriteString("      \n")
		buf.WriteString("      # Call decode\n")
		fmt.Fprintf(buf, "      handle = Bindings.%s_decode(buf, data.size, error_ptr)\n", msgName)
		buf.WriteString("      \n")
		buf.WriteString("      # Check for errors\n")
		buf.WriteString("      if handle.null?\n")
		buf.WriteString("        error_msg_ptr = error_ptr.read_pointer\n")
		buf.WriteString("        if error_msg_ptr.null?\n")
		buf.WriteString("          raise 'Failed to decode: unknown error'\n")
		buf.WriteString("        else\n")
		buf.WriteString("          error_msg = error_msg_ptr.read_string\n")
		fmt.Fprintf(buf, "          Bindings.%s_free_error(error_msg_ptr)\n", msgName)
		buf.WriteString("          raise \"Failed to decode: #{error_msg}\"\n")
		buf.WriteString("        end\n")
		buf.WriteString("      end\n")
		buf.WriteString("      \n")
		buf.WriteString("      new(handle)\n")
		buf.WriteString("    end\n\n")

		// Encode instance method
		buf.WriteString("    # Encode this message to binary data\n")
		buf.WriteString("    #\n")
		buf.WriteString("    # @return [String] the encoded binary data\n")
		buf.WriteString("    # @raise [RuntimeError] if encoding fails or object is freed\n")
		buf.WriteString("    def encode\n")
		buf.WriteString("      raise 'Message already freed' if @freed\n")
		buf.WriteString("      \n")
		buf.WriteString("      # Output pointers\n")
		buf.WriteString("      data_ptr = FFI::MemoryPointer.new(:pointer)\n")
		buf.WriteString("      error_ptr = FFI::MemoryPointer.new(:pointer)\n")
		buf.WriteString("      \n")
		buf.WriteString("      # Call encode\n")
		fmt.Fprintf(buf, "      size = Bindings.%s_encode(@handle, data_ptr, error_ptr)\n", msgName)
		buf.WriteString("      \n")
		buf.WriteString("      # Check for errors\n")
		buf.WriteString("      if size.zero?\n")
		buf.WriteString("        error_msg_ptr = error_ptr.read_pointer\n")
		buf.WriteString("        if error_msg_ptr.null?\n")
		buf.WriteString("          raise 'Failed to encode: unknown error'\n")
		buf.WriteString("        else\n")
		buf.WriteString("          error_msg = error_msg_ptr.read_string\n")
		fmt.Fprintf(buf, "          Bindings.%s_free_error(error_msg_ptr)\n", msgName)
		buf.WriteString("          raise \"Failed to encode: #{error_msg}\"\n")
		buf.WriteString("        end\n")
		buf.WriteString("      end\n")
		buf.WriteString("      \n")
		buf.WriteString("      # Read the data\n")
		buf.WriteString("      data = data_ptr.read_pointer.read_bytes(size)\n")
		buf.WriteString("      \n")
		buf.WriteString("      # Free the native data\n")
		fmt.Fprintf(buf, "      Bindings.%s_free_data(data_ptr.read_pointer)\n", msgName)
		buf.WriteString("      \n")
		buf.WriteString("      data\n")
		buf.WriteString("    end\n\n")

		// Free instance method (explicit)
		buf.WriteString("    # Explicitly free native resources\n")
		buf.WriteString("    #\n")
		buf.WriteString("    # Normally not needed as the finalizer handles this, but can be called\n")
		buf.WriteString("    # manually for immediate cleanup.\n")
		buf.WriteString("    #\n")
		buf.WriteString("    # @return [void]\n")
		buf.WriteString("    def free\n")
		buf.WriteString("      unless @freed || @handle.null?\n")
		fmt.Fprintf(buf, "        Bindings.%s_free(@handle)\n", msgName)
		buf.WriteString("        @freed = true\n")
		buf.WriteString("      end\n")
		buf.WriteString("    end\n\n")

		// Freed? helper
		buf.WriteString("    # Check if the message has been freed\n")
		buf.WriteString("    #\n")
		buf.WriteString("    # @return [Boolean] true if freed\n")
		buf.WriteString("    def freed?\n")
		buf.WriteString("      @freed\n")
		buf.WriteString("    end\n")

		buf.WriteString("  end\n")
		buf.WriteString("end\n")

		// Write to file
		filename := fmt.Sprintf("%s.rb", strings.ToLower(msg.Name))
		msgPath := filepath.Join(libDir, filename)
		if err := os.WriteFile(msgPath, buf.Bytes(), 0644); err != nil {
			return err
		}
	}

	return nil
}

// generateRubyVersion generates version.rb
func generateRubyVersion(config *PackageConfig, libDir string) error {
	buf := &bytes.Buffer{}
	className := toRubyClassName(config.Namespace)

	fmt.Fprintf(buf, "# frozen_string_literal: true\n\n")
	fmt.Fprintf(buf, "module %s\n", className)
	fmt.Fprintf(buf, "  VERSION = '1.0.0'\n")
	fmt.Fprintf(buf, "end\n")

	versionPath := filepath.Join(libDir, "version.rb")
	return os.WriteFile(versionPath, buf.Bytes(), 0644)
}

// generateRubyMainModule generates the main lib/<gemname>.rb file
func generateRubyMainModule(config *PackageConfig, gemDir string) error {
	buf := &bytes.Buffer{}
	className := toRubyClassName(config.Namespace)

	fmt.Fprintf(buf, "# frozen_string_literal: true\n\n")
	fmt.Fprintf(buf, "require_relative '%s/version'\n", config.Namespace)
	fmt.Fprintf(buf, "require_relative '%s/bindings'\n", config.Namespace)

	// Require all message classes
	for _, msg := range config.Schema.Messages {
		filename := strings.ToLower(msg.Name)
		fmt.Fprintf(buf, "require_relative '%s/%s'\n", config.Namespace, filename)
	}

	buf.WriteString("\n")
	fmt.Fprintf(buf, "# FFire %s Ruby bindings\n", config.Schema.Package)
	fmt.Fprintf(buf, "module %s\n", className)
	fmt.Fprintf(buf, "  # Your code goes here...\n")
	fmt.Fprintf(buf, "end\n")

	mainPath := filepath.Join(gemDir, "lib", fmt.Sprintf("%s.rb", config.Namespace))
	return os.WriteFile(mainPath, buf.Bytes(), 0644)
}

// generateRubyGemspec generates the .gemspec file
func generateRubyGemspec(config *PackageConfig, gemDir string) error {
	buf := &bytes.Buffer{}
	className := toRubyClassName(config.Namespace)

	fmt.Fprintf(buf, "# frozen_string_literal: true\n\n")
	fmt.Fprintf(buf, "require_relative 'lib/%s/version'\n\n", config.Namespace)

	fmt.Fprintf(buf, "Gem::Specification.new do |spec|\n")
	fmt.Fprintf(buf, "  spec.name          = '%s'\n", config.Namespace)
	fmt.Fprintf(buf, "  spec.version       = %s::VERSION\n", className)
	fmt.Fprintf(buf, "  spec.authors       = ['Generated by FFire']\n")
	fmt.Fprintf(buf, "  spec.email         = []\n")
	fmt.Fprintf(buf, "  spec.summary       = 'FFire binary serialization library - %s schema'\n", config.Schema.Package)
	fmt.Fprintf(buf, "  spec.description   = 'Ruby bindings for the %s FFire schema using FFI'\n", config.Schema.Package)
	fmt.Fprintf(buf, "  spec.homepage      = 'https://github.com/shaban/ffire'\n")
	fmt.Fprintf(buf, "  spec.license       = 'SEE LICENSE IN LICENSE'\n")
	fmt.Fprintf(buf, "  spec.required_ruby_version = '>= 2.6.0'\n\n")

	buf.WriteString("  spec.metadata['homepage_uri'] = spec.homepage\n")
	buf.WriteString("  spec.metadata['source_code_uri'] = spec.homepage\n\n")

	buf.WriteString("  # Specify which files should be added to the gem when it is released.\n")
	buf.WriteString("  spec.files = Dir[\n")
	buf.WriteString("    'lib/**/*',\n")
	buf.WriteString("    'README.md',\n")
	buf.WriteString("    'LICENSE'\n")
	buf.WriteString("  ]\n\n")

	buf.WriteString("  spec.bindir        = 'exe'\n")
	buf.WriteString("  spec.executables   = spec.files.grep(%r{^exe/}) { |f| File.basename(f) }\n")
	buf.WriteString("  spec.require_paths = ['lib']\n\n")

	buf.WriteString("  # Runtime dependencies\n")
	buf.WriteString("  spec.add_runtime_dependency 'ffi', '~> 1.15'\n\n")

	buf.WriteString("  # Development dependencies\n")
	buf.WriteString("  spec.add_development_dependency 'bundler', '~> 2.0'\n")
	buf.WriteString("  spec.add_development_dependency 'rake', '~> 13.0'\n")
	buf.WriteString("end\n")

	gemspecPath := filepath.Join(gemDir, fmt.Sprintf("%s.gemspec", config.Namespace))
	return os.WriteFile(gemspecPath, buf.Bytes(), 0644)
}

// generateRubyGemfile generates the Gemfile
func generateRubyGemfile(config *PackageConfig, gemDir string) error {
	buf := &bytes.Buffer{}

	buf.WriteString("# frozen_string_literal: true\n\n")
	buf.WriteString("source 'https://rubygems.org'\n\n")
	fmt.Fprintf(buf, "gemspec\n")

	gemfilePath := filepath.Join(gemDir, "Gemfile")
	return os.WriteFile(gemfilePath, buf.Bytes(), 0644)
}

// generateRubyReadme generates the README.md
func generateRubyReadme(config *PackageConfig, gemDir string) error {
	buf := &bytes.Buffer{}
	className := toRubyClassName(config.Namespace)

	fmt.Fprintf(buf, "# %s - FFire Ruby Bindings\n\n", config.Namespace)
	fmt.Fprintf(buf, "Ruby bindings for the %s schema, generated by [FFire](https://github.com/shaban/ffire).\n\n", config.Schema.Package)

	buf.WriteString("## Installation\n\n")
	buf.WriteString("Add this line to your application's Gemfile:\n\n")
	buf.WriteString("```ruby\n")
	fmt.Fprintf(buf, "gem '%s'\n", config.Namespace)
	buf.WriteString("```\n\n")
	buf.WriteString("And then execute:\n\n")
	buf.WriteString("```bash\n")
	buf.WriteString("bundle install\n")
	buf.WriteString("```\n\n")
	buf.WriteString("Or install it yourself as:\n\n")
	buf.WriteString("```bash\n")
	fmt.Fprintf(buf, "gem install %s\n", config.Namespace)
	buf.WriteString("```\n\n")

	buf.WriteString("## Usage\n\n")
	buf.WriteString("```ruby\n")
	fmt.Fprintf(buf, "require '%s'\n\n", config.Namespace)

	// Generate example for first message type
	if len(config.Schema.Messages) > 0 {
		msg := config.Schema.Messages[0]
		msgClassName := toRubyClassName(msg.Name)

		buf.WriteString("# Decode from binary data\n")
		buf.WriteString("data = File.binread('data.bin')\n")
		fmt.Fprintf(buf, "msg = %s::%s.decode(data)\n\n", className, msgClassName)

		buf.WriteString("# Encode back to binary\n")
		buf.WriteString("encoded = msg.encode\n\n")

		buf.WriteString("# Free resources (optional - finalizer handles this)\n")
		buf.WriteString("msg.free\n")
	}

	buf.WriteString("```\n\n")

	buf.WriteString("## API\n\n")

	// Document each message type
	for _, msg := range config.Schema.Messages {
		msgClassName := toRubyClassName(msg.Name)

		fmt.Fprintf(buf, "### `%s::%s`\n\n", className, msgClassName)

		buf.WriteString("**`.decode(data)` → `#{msgClassName}`**\n\n")
		buf.WriteString("Decode a message from binary data.\n\n")
		buf.WriteString("- **Parameters:**\n")
		buf.WriteString("  - `data` - Binary data (String or Array of bytes)\n")
		fmt.Fprintf(buf, "- **Returns:** `%s` object\n", msgClassName)
		buf.WriteString("- **Raises:** `RuntimeError` if decoding fails\n\n")

		buf.WriteString("**`#encode` → `String`**\n\n")
		buf.WriteString("Encode this message to binary data.\n\n")
		buf.WriteString("- **Returns:** Binary data (String)\n")
		buf.WriteString("- **Raises:** `RuntimeError` if encoding fails\n\n")

		buf.WriteString("**`#free` → `void`**\n\n")
		buf.WriteString("Explicitly free native resources. Normally not needed as Ruby's garbage collector\n")
		buf.WriteString("will call the finalizer automatically.\n\n")

		buf.WriteString("**`#freed?` → `Boolean`**\n\n")
		buf.WriteString("Check if the message has been freed.\n\n")
	}

	buf.WriteString("## Memory Management\n\n")
	buf.WriteString("This gem uses Ruby finalizers to automatically free native resources when objects\n")
	buf.WriteString("are garbage collected. You don't need to call `free` explicitly unless you want\n")
	buf.WriteString("immediate cleanup.\n\n")

	buf.WriteString("## Platform Support\n\n")
	buf.WriteString("This gem includes pre-compiled libraries for:\n\n")
	buf.WriteString("- macOS (Darwin): `libffire.dylib`\n")
	buf.WriteString("- Linux: `libffire.so`\n")
	buf.WriteString("- Windows: `ffire.dll`\n\n")
	buf.WriteString("The correct library is automatically loaded based on your platform.\n\n")

	buf.WriteString("## Requirements\n\n")
	buf.WriteString("- Ruby 2.6.0 or higher\n")
	buf.WriteString("- The `ffi` gem (~> 1.15)\n\n")

	buf.WriteString("## Development\n\n")
	buf.WriteString("After checking out the repo, run `bundle install` to install dependencies.\n\n")

	buf.WriteString("## License\n\n")
	buf.WriteString("Generated by FFire. See your schema's license for terms.\n")

	readmePath := filepath.Join(gemDir, "README.md")
	return os.WriteFile(readmePath, buf.Bytes(), 0644)
}

// toRubyClassName converts a string to Ruby class name (CamelCase)
func toRubyClassName(s string) string {
	// Handle snake_case and kebab-case
	s = strings.ReplaceAll(s, "-", "_")
	parts := strings.Split(s, "_")

	for i, part := range parts {
		if part != "" {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}

	return strings.Join(parts, "")
}
