#!/usr/bin/env python3
"""
Test both upb (fast) and pure Python implementations separately.
"""
import os
import sys

# Force pure Python BEFORE importing protobuf
os.environ['PROTOCOL_BUFFERS_PYTHON_IMPLEMENTATION'] = 'python'

import time
from google.protobuf import descriptor_pb2
from google.protobuf.internal import api_implementation

print(f"Implementation: {api_implementation.Type()}")

# Create a ~7.5KB message
def create_message():
    file_desc = descriptor_pb2.FileDescriptorProto()
    file_desc.name = "test.proto"
    file_desc.package = "test"
    
    for i in range(10):
        msg = file_desc.message_type.add()
        msg.name = f"Msg{i}"
        for j in range(12):
            field = msg.field.add()
            field.name = f"f{j}"
            field.number = j + 1
            field.type = descriptor_pb2.FieldDescriptorProto.TYPE_INT32
            field.label = descriptor_pb2.FieldDescriptorProto.LABEL_OPTIONAL
    
    return file_desc.SerializeToString()

data = create_message()
print(f"Message size: {len(data)} bytes\n")

# Benchmark
times = []
for _ in range(1000):
    msg = descriptor_pb2.FileDescriptorProto()
    start = time.perf_counter_ns()
    msg.ParseFromString(data)
    end = time.perf_counter_ns()
    times.append(end - start)

times = sorted(times)[10:-10]
avg = sum(times) / len(times)

print(f"Average: {avg:.0f} ns ({avg/1000:.2f} µs)")
print(f"Min: {min(times):.0f} ns")
print(f"Max: {max(times):.0f} ns")
