"""
FFire test bindings using ctypes.

This module provides Python bindings to the FFire binary serialization library
via a C ABI dynamic library.
"""

import ctypes
import os
import platform
from typing import List, Optional

# Determine library name based on platform
_lib_name = {
    'Darwin': 'libffire.dylib',
    'Linux': 'libffire.so',
    'Windows': 'ffire.dll'
}.get(platform.system(), 'libffire.so')

# Load the C library
_lib_path = os.path.join(os.path.dirname(__file__), _lib_name)
_lib = ctypes.CDLL(_lib_path)

# Handle type
MessageHandle = ctypes.c_void_p

# Decode function
_lib.message_decode.argtypes = [ctypes.POINTER(ctypes.c_uint8), ctypes.c_size_t, ctypes.POINTER(ctypes.c_char_p)]
_lib.message_decode.restype = MessageHandle

# Encode function
_lib.message_encode.argtypes = [MessageHandle, ctypes.POINTER(ctypes.POINTER(ctypes.c_uint8)), ctypes.POINTER(ctypes.c_char_p)]
_lib.message_encode.restype = ctypes.c_size_t

# Memory management functions
_lib.message_free.argtypes = [MessageHandle]
_lib.message_free.restype = None
_lib.message_free_data.argtypes = [ctypes.POINTER(ctypes.c_uint8)]
_lib.message_free_data.restype = None
_lib.message_free_error.argtypes = [ctypes.c_char_p]
_lib.message_free_error.restype = None

class Message:
    """Wrapper for Message message type."""
    
    def __init__(self, handle: MessageHandle):
        self._handle = handle
    
    def __del__(self):
        if hasattr(self, '_handle') and self._handle:
            _lib.message_free(self._handle)
    
    @staticmethod
    def decode(data: bytes) -> 'Message':
        """Decode a Message from binary data."""
        data_array = (ctypes.c_uint8 * len(data)).from_buffer_copy(data)
        error = ctypes.c_char_p()
        
        handle = _lib.message_decode(data_array, len(data), ctypes.byref(error))
        
        if not handle:
            error_msg = error.value.decode('utf-8') if error.value else 'Unknown error'
            if error:
                _lib.message_free_error(error)
            raise RuntimeError(f"Failed to decode Message: {error_msg}")
        
        return Message(handle)
    
    def encode(self) -> bytes:
        """Encode this Message to binary data."""
        encoded_data = ctypes.POINTER(ctypes.c_uint8)()
        error = ctypes.c_char_p()
        
        size = _lib.message_encode(self._handle, ctypes.byref(encoded_data), ctypes.byref(error))
        
        if size == 0:
            error_msg = error.value.decode('utf-8') if error.value else 'Unknown error'
            if error:
                _lib.message_free_error(error)
            raise RuntimeError(f"Failed to encode Message: {error_msg}")
        
        # Copy data to Python bytes
        result = bytes(encoded_data[:size])
        _lib.message_free_data(encoded_data)
        
        return result

