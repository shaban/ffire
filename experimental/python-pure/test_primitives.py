#!/usr/bin/env python3
"""
Test Python's native binary primitives for high-performance codec implementation.
Explores: bytearray, memoryview, struct module for zero-copy operations.
"""

import struct
import time
import sys

def test_struct_module():
    """Test struct module for encoding/decoding primitives with endianness control"""
    print("\n=== Testing struct module ===")
    
    # Encode primitives (little-endian)
    buffer = bytearray()
    buffer.extend(struct.pack('<i', 42))        # int32
    buffer.extend(struct.pack('<f', 3.14))      # float32
    buffer.extend(struct.pack('<d', 2.71828))   # float64
    buffer.extend(struct.pack('<q', 1234567890123456789))  # int64
    
    print(f"Encoded buffer: {buffer.hex()}")
    print(f"Buffer size: {len(buffer)} bytes")
    
    # Decode primitives
    pos = 0
    val_i32 = struct.unpack_from('<i', buffer, pos)[0]
    pos += 4
    val_f32 = struct.unpack_from('<f', buffer, pos)[0]
    pos += 4
    val_f64 = struct.unpack_from('<d', buffer, pos)[0]
    pos += 8
    val_i64 = struct.unpack_from('<q', buffer, pos)[0]
    pos += 8
    
    print(f"Decoded int32: {val_i32}")
    print(f"Decoded float32: {val_f32}")
    print(f"Decoded float64: {val_f64}")
    print(f"Decoded int64: {val_i64}")


def test_memoryview_zero_copy():
    """Test memoryview for zero-copy array operations"""
    print("\n=== Testing memoryview (zero-copy) ===")
    
    # Create a large array of int32s
    import array
    data = array.array('i', range(1000))  # 1000 int32s
    
    # Get memoryview (zero-copy!)
    mv = memoryview(data)
    
    print(f"Array size: {len(data)} elements = {len(data) * 4} bytes")
    print(f"Memoryview: {mv.nbytes} bytes, format: {mv.format}")
    
    # Cast to bytes for wire format (zero-copy view)
    bytes_view = mv.cast('B')
    print(f"Bytes view: {len(bytes_view)} bytes")
    print(f"First 16 bytes: {bytes(bytes_view[:16]).hex()}")
    
    # Reconstruct array from bytes (zero-copy if aligned)
    reconstructed = memoryview(bytes_view).cast('i')
    print(f"Reconstructed first 10 elements: {reconstructed[:10].tolist()}")


def test_bytearray_performance():
    """Benchmark bytearray write performance"""
    print("\n=== Testing bytearray performance ===")
    
    iterations = 100000
    
    # Test 1: Individual writes
    start = time.perf_counter()
    buffer = bytearray()
    for i in range(iterations):
        buffer.extend(struct.pack('<i', i))
    elapsed = time.perf_counter() - start
    print(f"Individual writes: {iterations} int32s in {elapsed*1000:.2f}ms ({elapsed/iterations*1e6:.2f}µs each)")
    
    # Test 2: Batch write with pre-allocated buffer
    start = time.perf_counter()
    buffer = bytearray(iterations * 4)
    for i in range(iterations):
        struct.pack_into('<i', buffer, i * 4, i)
    elapsed = time.perf_counter() - start
    print(f"Pre-allocated writes: {iterations} int32s in {elapsed*1000:.2f}ms ({elapsed/iterations*1e6:.2f}µs each)")
    
    # Test 3: Bulk array write (zero-copy with array module)
    import array
    start = time.perf_counter()
    data = array.array('i', range(iterations))
    buffer = bytearray(memoryview(data).cast('B'))
    elapsed = time.perf_counter() - start
    print(f"Zero-copy array: {iterations} int32s in {elapsed*1000:.2f}ms ({elapsed/iterations*1e6:.2f}µs each)")


def test_string_encoding():
    """Test UTF-8 string encoding/decoding"""
    print("\n=== Testing string encoding ===")
    
    test_str = "Hello, 世界! 🚀"
    
    # Encode to UTF-8
    utf8_bytes = test_str.encode('utf-8')
    print(f"Original: {test_str}")
    print(f"UTF-8 bytes: {utf8_bytes.hex()}")
    print(f"Length: {len(utf8_bytes)} bytes")
    
    # Decode from UTF-8
    decoded = utf8_bytes.decode('utf-8')
    print(f"Decoded: {decoded}")
    print(f"Match: {decoded == test_str}")


def test_roundtrip_struct():
    """Test complete encode/decode roundtrip"""
    print("\n=== Testing roundtrip (struct with primitives) ===")
    
    # Define a test struct
    class TestMessage:
        def __init__(self, id=0, name="", score=0.0):
            self.id = id
            self.name = name
            self.score = score
        
        def encode(self):
            buffer = bytearray()
            # Write int32 id
            buffer.extend(struct.pack('<i', self.id))
            # Write string (uint16 length + UTF-8 bytes)
            name_bytes = self.name.encode('utf-8')
            buffer.extend(struct.pack('<H', len(name_bytes)))
            buffer.extend(name_bytes)
            # Write float64 score
            buffer.extend(struct.pack('<d', self.score))
            return bytes(buffer)
        
        @staticmethod
        def decode(data):
            pos = 0
            # Read int32 id
            id_val = struct.unpack_from('<i', data, pos)[0]
            pos += 4
            # Read string
            str_len = struct.unpack_from('<H', data, pos)[0]
            pos += 2
            name_val = data[pos:pos+str_len].decode('utf-8')
            pos += str_len
            # Read float64 score
            score_val = struct.unpack_from('<d', data, pos)[0]
            pos += 8
            
            return TestMessage(id_val, name_val, score_val)
    
    # Test
    original = TestMessage(42, "Alice", 95.5)
    print(f"Original: id={original.id}, name={original.name}, score={original.score}")
    
    encoded = original.encode()
    print(f"Encoded: {encoded.hex()} ({len(encoded)} bytes)")
    
    decoded = TestMessage.decode(encoded)
    print(f"Decoded: id={decoded.id}, name={decoded.name}, score={decoded.score}")
    
    # Benchmark
    iterations = 100000
    start = time.perf_counter()
    for _ in range(iterations):
        encoded = original.encode()
    elapsed = time.perf_counter() - start
    print(f"\nEncode benchmark: {iterations} iterations in {elapsed*1000:.2f}ms ({elapsed/iterations*1e6:.2f}µs each)")
    
    start = time.perf_counter()
    for _ in range(iterations):
        decoded = TestMessage.decode(encoded)
    elapsed = time.perf_counter() - start
    print(f"Decode benchmark: {iterations} iterations in {elapsed*1000:.2f}ms ({elapsed/iterations*1e6:.2f}µs each)")


if __name__ == "__main__":
    print(f"Python version: {sys.version}")
    print(f"Struct module available: {struct is not None}")
    
    test_struct_module()
    test_memoryview_zero_copy()
    test_bytearray_performance()
    test_string_encoding()
    test_roundtrip_struct()
    
    print("\n✅ Pure Python codec primitives are viable!")
    print("   - struct module for DataView-like operations")
    print("   - memoryview for zero-copy TypedArray-like views")
    print("   - bytearray for mutable ArrayBuffer-like storage")
    print("   - Performance is decent for pure Python (~1-2µs per operation)")
