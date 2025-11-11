#!/usr/bin/env python3
"""
Simple benchmark to measure Python protobuf parsing performance.
This will help us compare with our ffire Python implementation.
"""

import sys
import time
import os

try:
    from google.protobuf import descriptor_pb2
    print("✅ Successfully imported protobuf")
except ImportError as e:
    print(f"❌ Failed to import protobuf: {e}")
    print("\nPlease install protobuf:")
    print("  pip3 install protobuf --break-system-packages")
    sys.exit(1)

def benchmark_parse(data, message_class, iterations=1000):
    """Benchmark parsing performance."""
    times = []
    
    for i in range(iterations):
        msg = message_class()
        start = time.perf_counter_ns()
        msg.ParseFromString(data)
        end = time.perf_counter_ns()
        times.append(end - start)
    
    # Remove outliers (first few runs may be slower due to warmup)
    times = sorted(times)[10:-10]
    
    avg_time = sum(times) / len(times)
    min_time = min(times)
    max_time = max(times)
    
    return {
        'avg_ns': avg_time,
        'min_ns': min_time,
        'max_ns': max_time,
        'iterations': len(times)
    }

def create_sample_file_descriptor():
    """Create a sample FileDescriptorProto for benchmarking (~7.5KB like C++ benchmark)."""
    file_desc = descriptor_pb2.FileDescriptorProto()
    file_desc.name = "benchmark_test.proto"
    file_desc.package = "benchmark.test"
    file_desc.syntax = "proto2"
    
    # Adjust to hit ~7.5KB serialized size (similar to descriptor.proto in C++ benchmark)
    # Through trial: 10 messages x 12 fields + 5 enums gets us close
    for i in range(10):
        msg = file_desc.message_type.add()
        msg.name = f"TestMessage{i}"
        
        # Add fields
        for j in range(12):
            field = msg.field.add()
            field.name = f"field_{j}"
            field.number = j + 1
            field.label = descriptor_pb2.FieldDescriptorProto.LABEL_OPTIONAL
            
            # Mix field types
            if j % 4 == 0:
                field.type = descriptor_pb2.FieldDescriptorProto.TYPE_STRING
                field.default_value = f"default_{j}"
            elif j % 4 == 1:
                field.type = descriptor_pb2.FieldDescriptorProto.TYPE_INT32
                field.default_value = str(j * 10)
            elif j % 4 == 2:
                field.type = descriptor_pb2.FieldDescriptorProto.TYPE_BOOL
                field.default_value = "false"
            else:
                field.type = descriptor_pb2.FieldDescriptorProto.TYPE_MESSAGE
                field.type_name = f".benchmark.test.TestMessage{(i+1)%10}"
        
        # Add some nested types
        if i % 3 == 0:
            nested = msg.nested_type.add()
            nested.name = "NestedType"
            for k in range(4):
                nested_field = nested.field.add()
                nested_field.name = f"nested_{k}"
                nested_field.number = k + 1
                nested_field.type = descriptor_pb2.FieldDescriptorProto.TYPE_INT32
                nested_field.label = descriptor_pb2.FieldDescriptorProto.LABEL_OPTIONAL
    
    # Add enums
    for i in range(5):
        enum = file_desc.enum_type.add()
        enum.name = f"TestEnum{i}"
        for j in range(8):
            value = enum.value.add()
            value.name = f"ENUM_{i}_VALUE_{j}"
            value.number = j
    
    # Add some options
    file_desc.options.java_package = "com.benchmark.test"
    file_desc.options.optimize_for = descriptor_pb2.FileOptions.SPEED
    
    return file_desc

def main():
    print("=" * 60)
    print("Python Protobuf Parsing Benchmark")
    print("=" * 60)
    
    # Check which implementation is being used
    from google.protobuf.internal import api_implementation
    
    # Try to force pure Python for comparison
    try:
        os.environ['PROTOCOL_BUFFERS_PYTHON_IMPLEMENTATION'] = 'python'
        # Reimport to pick up the environment variable
        import importlib
        importlib.reload(api_implementation)
    except:
        pass
    
    impl = api_implementation.Type()
    print(f"\nProtobuf implementation: {impl}")
    print(f"  - 'cpp': C++ extension (fast)")
    print(f"  - 'python': Pure Python (slow)")
    print(f"  - 'upb': upb binding (fast)")
    
    # Create test data
    print("\nCreating test FileDescriptorProto...")
    file_desc = create_sample_file_descriptor()
    serialized = file_desc.SerializeToString()
    print(f"Serialized size: {len(serialized)} bytes")
    
    target_size = 7500  # Match C++ benchmark size
    if len(serialized) < target_size * 0.8:
        print(f"⚠️  Warning: Message is smaller than C++ benchmark ({target_size} bytes)")
        print(f"   Results may not be directly comparable")
    elif len(serialized) > target_size * 1.5:
        print(f"⚠️  Warning: Message is larger than C++ benchmark ({target_size} bytes)")
        print(f"   Results may not be directly comparable")
    else:
        print(f"✅ Message size is comparable to C++ benchmark (~{target_size} bytes)")
    
    # Benchmark parsing
    print(f"\nRunning benchmark with 1000 iterations...")
    results = benchmark_parse(serialized, descriptor_pb2.FileDescriptorProto, iterations=1000)
    
    print("\n" + "=" * 60)
    print("RESULTS")
    print("=" * 60)
    print(f"Average: {results['avg_ns']:.0f} ns ({results['avg_ns']/1000:.2f} µs)")
    print(f"Min:     {results['min_ns']:.0f} ns ({results['min_ns']/1000:.2f} µs)")
    print(f"Max:     {results['max_ns']:.0f} ns ({results['max_ns']/1000:.2f} µs)")
    print(f"Samples: {results['iterations']}")
    
    # Compare with C++ baseline from our earlier measurement
    cpp_time_ns = 90000  # From protobuf C++ benchmark
    ratio = results['avg_ns'] / cpp_time_ns
    
    print("\n" + "=" * 60)
    print("COMPARISON")
    print("=" * 60)
    print(f"C++ (upb):      {cpp_time_ns:,.0f} ns")
    print(f"Python ({impl}): {results['avg_ns']:,.0f} ns")
    print(f"Ratio:          {ratio:.1f}x slower")
    
    # Compare with ffire
    ffire_python_ns = 148979  # From our benchmarks
    ffire_cpp_ns = 4170      # From our benchmarks
    ffire_ratio = ffire_python_ns / ffire_cpp_ns
    
    print("\n" + "=" * 60)
    print("FFIRE COMPARISON")
    print("=" * 60)
    print(f"FFire C++:      {ffire_cpp_ns:,.0f} ns")
    print(f"FFire Python:   {ffire_python_ns:,.0f} ns")
    print(f"FFire Ratio:    {ffire_ratio:.1f}x slower")
    print(f"\nProtobuf ratio: {ratio:.1f}x")
    print(f"FFire ratio:    {ffire_ratio:.1f}x")
    
    if abs(ratio - ffire_ratio) / ratio < 0.3:
        print("\n✅ Similar ratios - FFire Python implementation maturity is comparable to protobuf!")
    elif ffire_ratio > ratio * 2:
        print("\n⚠️  FFire Python is significantly slower - optimization opportunity")
    else:
        print("\n✅ FFire Python is competitive")

if __name__ == "__main__":
    main()
