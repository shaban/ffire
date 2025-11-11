#!/usr/bin/env python3
"""
Calculate the message size from C++ benchmark output.
From the benchmark we saw: bytes_per_second=33.2786Mi/s at 215,114 ns
"""

# From C++ benchmark output:
# BM_Parse_Proto2<FileDesc, NoArena, Copy>_mean: 215,114 ns
# bytes_per_second=33.2786Mi/s

time_ns = 215114
bytes_per_sec = 33.2786 * 1024 * 1024  # Convert MiB/s to bytes/s

# Calculate message size: bytes = (bytes/sec) * (time in seconds)
time_sec = time_ns / 1_000_000_000
message_bytes = bytes_per_sec * time_sec

print(f"Calculation from C++ benchmark:")
print(f"  Time: {time_ns:,} ns ({time_sec*1000:.4f} ms)")
print(f"  Throughput: {bytes_per_sec:,.0f} bytes/sec ({bytes_per_sec/1024/1024:.2f} MiB/s)")
print(f"  Message size: {message_bytes:,.0f} bytes (~{message_bytes/1024:.1f} KB)")

print(f"\nThe C++ benchmarks are parsing a ~{int(message_bytes)} byte message")
