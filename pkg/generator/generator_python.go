package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shaban/ffire/pkg/schema"
)

// generatePythonWrapper generates the ctypes wrapper module
func generatePythonWrapper(config *PackageConfig, packageDir string) error {
	buf := &bytes.Buffer{}

	// Module docstring
	fmt.Fprintf(buf, `"""
FFire %s bindings using ctypes.

This module provides Python bindings to the FFire binary serialization library
via a C ABI dynamic library.
"""

import ctypes
import os
import platform
from typing import List, Optional

`, config.Schema.Package)

	// Determine library name based on platform
	buf.WriteString("# Determine library name based on platform\n")
	buf.WriteString("_lib_name = {\n")
	buf.WriteString("    'Darwin': 'libffire.dylib',\n")
	buf.WriteString("    'Linux': 'libffire.so',\n")
	buf.WriteString("    'Windows': 'ffire.dll'\n")
	buf.WriteString("}.get(platform.system(), 'libffire.so')\n\n")

	// Load library
	buf.WriteString("# Load the C library\n")
	buf.WriteString("_lib_path = os.path.join(os.path.dirname(__file__), _lib_name)\n")
	buf.WriteString("_lib = ctypes.CDLL(_lib_path)\n\n")

	// Generate bindings for each message type
	for _, msg := range config.Schema.Messages {
		if err := generatePythonMessageBindings(buf, config.Schema, &msg); err != nil {
			return err
		}
	}

	// Write to file
	wrapperPath := filepath.Join(packageDir, "bindings.py")
	if err := os.WriteFile(wrapperPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write Python wrapper: %w", err)
	}

	fmt.Printf("✓ Generated Python bindings: %s\n", wrapperPath)
	return nil
}

func generatePythonMessageBindings(buf *bytes.Buffer, s *schema.Schema, msg *schema.MessageType) error {
	handleName := msg.Name + "Handle"
	baseName := strings.ToLower(msg.Name[:1]) + msg.Name[1:]

	// Define C types
	buf.WriteString("# Handle type\n")
	fmt.Fprintf(buf, "%s = ctypes.c_void_p\n\n", handleName)

	// Define function signatures
	buf.WriteString("# Decode function\n")
	fmt.Fprintf(buf, "_lib.%s_decode.argtypes = [ctypes.POINTER(ctypes.c_uint8), ctypes.c_size_t, ctypes.POINTER(ctypes.c_char_p)]\n", baseName)
	fmt.Fprintf(buf, "_lib.%s_decode.restype = %s\n\n", baseName, handleName)

	buf.WriteString("# Encode function\n")
	fmt.Fprintf(buf, "_lib.%s_encode.argtypes = [%s, ctypes.POINTER(ctypes.POINTER(ctypes.c_uint8)), ctypes.POINTER(ctypes.c_char_p)]\n", baseName, handleName)
	fmt.Fprintf(buf, "_lib.%s_encode.restype = ctypes.c_size_t\n\n", baseName)

	buf.WriteString("# Memory management functions\n")
	fmt.Fprintf(buf, "_lib.%s_free.argtypes = [%s]\n", baseName, handleName)
	fmt.Fprintf(buf, "_lib.%s_free.restype = None\n", baseName)
	fmt.Fprintf(buf, "_lib.%s_free_data.argtypes = [ctypes.POINTER(ctypes.c_uint8)]\n", baseName)
	fmt.Fprintf(buf, "_lib.%s_free_data.restype = None\n", baseName)
	fmt.Fprintf(buf, "_lib.%s_free_error.argtypes = [ctypes.c_char_p]\n", baseName)
	fmt.Fprintf(buf, "_lib.%s_free_error.restype = None\n\n", baseName)

	// Generate Python class wrapper
	className := msg.Name
	fmt.Fprintf(buf, "class %s:\n", className)
	fmt.Fprintf(buf, `    """Wrapper for %s message type."""
    
    def __init__(self, handle: %s):
        self._handle = handle
    
    def __del__(self):
        if hasattr(self, '_handle') and self._handle:
            _lib.%s_free(self._handle)
    
    @staticmethod
    def decode(data: bytes) -> '%s':
        """Decode a %s from binary data."""
        data_array = (ctypes.c_uint8 * len(data)).from_buffer_copy(data)
        error = ctypes.c_char_p()
        
        handle = _lib.%s_decode(data_array, len(data), ctypes.byref(error))
        
        if not handle:
            error_msg = error.value.decode('utf-8') if error.value else 'Unknown error'
            if error:
                _lib.%s_free_error(error)
            raise RuntimeError(f"Failed to decode %s: {error_msg}")
        
        return %s(handle)
    
    def encode(self) -> bytes:
        """Encode this %s to binary data."""
        encoded_data = ctypes.POINTER(ctypes.c_uint8)()
        error = ctypes.c_char_p()
        
        size = _lib.%s_encode(self._handle, ctypes.byref(encoded_data), ctypes.byref(error))
        
        if size == 0:
            error_msg = error.value.decode('utf-8') if error.value else 'Unknown error'
            if error:
                _lib.%s_free_error(error)
            raise RuntimeError(f"Failed to encode %s: {error_msg}")
        
        # Copy data to Python bytes
        result = bytes(encoded_data[:size])
        _lib.%s_free_data(encoded_data)
        
        return result

`, className, handleName, baseName, className, className, baseName, baseName, className, className, className, baseName, baseName, className, baseName)

	return nil
}

// generatePythonSetup generates setup.py for the package
func generatePythonSetup(config *PackageConfig, langDir string) error {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, `from setuptools import setup, find_packages
import platform

# Determine which library to include based on platform
data_files = []
system = platform.system()
if system == 'Darwin':
    data_files = ['%s/libffire.dylib']
elif system == 'Linux':
    data_files = ['%s/libffire.so']
elif system == 'Windows':
    data_files = ['%s/ffire.dll']

setup(
    name='%s',
    version='1.0.0',
    description='FFire binary serialization library - %s schema',
    author='Generated by FFire',
    packages=find_packages(),
    package_data={
        '%s': data_files,
    },
    python_requires='>=3.7',
    classifiers=[
        'Development Status :: 4 - Beta',
        'Intended Audience :: Developers',
        'Programming Language :: Python :: 3',
        'Programming Language :: Python :: 3.7',
        'Programming Language :: Python :: 3.8',
        'Programming Language :: Python :: 3.9',
        'Programming Language :: Python :: 3.10',
        'Programming Language :: Python :: 3.11',
    ],
)
`, config.Namespace, config.Namespace, config.Namespace, config.Namespace, config.Schema.Package, config.Namespace)

	setupPath := filepath.Join(langDir, "setup.py")
	if err := os.WriteFile(setupPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write setup.py: %w", err)
	}

	fmt.Printf("✓ Generated setup.py: %s\n", setupPath)
	return nil
}

// generatePythonInit generates __init__.py for the package
func generatePythonInit(config *PackageConfig, packageDir string) error {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, `"""
%s - FFire bindings for %s schema.
"""

from .bindings import (
`, config.Namespace, config.Schema.Package)

	// Import all message types
	for i, msg := range config.Schema.Messages {
		if i > 0 {
			buf.WriteString(",\n")
		}
		fmt.Fprintf(buf, "    %s", msg.Name)
	}

	buf.WriteString(",\n)\n\n__all__ = [\n")
	
	for i, msg := range config.Schema.Messages {
		if i > 0 {
			buf.WriteString(",\n")
		}
		fmt.Fprintf(buf, "    '%s'", msg.Name)
	}
	
	buf.WriteString(",\n]\n")

	initPath := filepath.Join(packageDir, "__init__.py")
	if err := os.WriteFile(initPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write __init__.py: %w", err)
	}

	fmt.Printf("✓ Generated __init__.py: %s\n", initPath)
	return nil
}
