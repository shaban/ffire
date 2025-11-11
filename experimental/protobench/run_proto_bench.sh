#!/bin/bash
# Quick script to run protobuf benchmarks and extract key decode times

echo "=========================================="
echo "Protobuf C++ Benchmark Results"
echo "=========================================="

cd protobuf

echo ""
echo "Running Parse benchmarks (this may take a minute)..."
bazel run //benchmarks:benchmark -- \
  --benchmark_filter="BM_Parse.*FileDesc.*_mean" \
  --benchmark_repetitions=5 \
  2>&1 | grep "_mean" | grep -v "stddev\|cv" | awk '{print $1, $2}'

echo ""
echo "=========================================="
echo "Key takeaways:"
echo "  - BM_Parse_Proto2: Standard C++ implementation"
echo "  - BM_Parse_Upb: Optimized upb implementation"
echo "=========================================="
