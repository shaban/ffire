#!/usr/bin/env python3
"""
Python test using ctypes to call C ABI wrapper
"""
import ctypes
import time
import json

# Build the library first if it doesn't exist
import subprocess
import os

if not os.path.exists('./libffire.dylib'):
    subprocess.run(['make'], check=True)

# Load the shared library
lib = ctypes.CDLL('./libffire.dylib')

# Define C types
PluginHandle = ctypes.c_void_p

# Define function signatures
lib.plugin_decode.argtypes = [ctypes.POINTER(ctypes.c_uint8), ctypes.c_size_t, ctypes.POINTER(ctypes.c_char_p)]
lib.plugin_decode.restype = PluginHandle

lib.plugin_encode.argtypes = [PluginHandle, ctypes.POINTER(ctypes.POINTER(ctypes.c_uint8)), ctypes.POINTER(ctypes.c_char_p)]
lib.plugin_encode.restype = ctypes.c_size_t

lib.plugin_free.argtypes = [PluginHandle]
lib.plugin_free.restype = None

lib.plugin_free_data.argtypes = [ctypes.POINTER(ctypes.c_uint8)]
lib.plugin_free_data.restype = None

# Read binary file
with open('../common/complex.bin', 'rb') as f:
    data = f.read()

# Convert to C array
data_array = (ctypes.c_uint8 * len(data)).from_buffer_copy(data)

ITERATIONS = 100

# Warmup
for _ in range(10):
    error = ctypes.c_char_p()
    plugin = lib.plugin_decode(data_array, len(data), ctypes.byref(error))
    if plugin:
        lib.plugin_free(plugin)

# Decode benchmark
decode_start = time.perf_counter()
for _ in range(ITERATIONS):
    error = ctypes.c_char_p()
    plugin = lib.plugin_decode(data_array, len(data), ctypes.byref(error))
    lib.plugin_free(plugin)
decode_end = time.perf_counter()
decode_us = int((decode_end - decode_start) * 1_000_000 / ITERATIONS)

# Get plugin for encoding
error = ctypes.c_char_p()
plugin = lib.plugin_decode(data_array, len(data), ctypes.byref(error))

# Encode benchmark
encode_start = time.perf_counter()
encoded_data = ctypes.POINTER(ctypes.c_uint8)()
for _ in range(ITERATIONS):
    if encoded_data:
        lib.plugin_free_data(encoded_data)
    encoded_size = lib.plugin_encode(plugin, ctypes.byref(encoded_data), ctypes.byref(error))
encode_end = time.perf_counter()
encode_us = int((encode_end - encode_start) * 1_000_000 / ITERATIONS)

# Output JSON
result = {
    "decode_us": decode_us,
    "encode_us": encode_us,
    "size_bytes": encoded_size,
    "iterations": ITERATIONS
}
print(json.dumps(result))

# Cleanup
lib.plugin_free_data(encoded_data)
lib.plugin_free(plugin)
