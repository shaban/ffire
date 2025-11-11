# Protocol Buffers Benchmark Comparison

This directory runs protobuf's official benchmarks to extract performance ratios for comparison with ffire.

## Goal

Extract language performance ratios from protobuf to validate ffire implementation maturity.

## Setup

1. Clone protobuf repository
2. Build and run benchmarks
3. Extract performance ratios (Go/C++, Python/C++, etc.)
4. Compare with ffire ratios

## Expected Output

Performance ratios like:
- Go decode: 1.5x slower than C++
- Python decode: 8x slower than C++
- Java decode: 2x slower than C++

These ratios serve as reference to validate ffire's language implementations.
