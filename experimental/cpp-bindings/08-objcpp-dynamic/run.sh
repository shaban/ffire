#!/bin/bash
set -e

echo "=== 08-objcpp-dynamic: Objective-C++ with dynamic library (C ABI) ==="
echo ""

echo "=== Building (cold) ==="
make clean > /dev/null 2>&1
TIME=$(/usr/bin/time -p make 2>&1 | grep real | awk '{print $2}')
echo "Build time: ${TIME}s"

echo ""
echo "=== Binary Size ==="
ls -lh test libffire.dylib | awk '{print $5 " " $9}'

echo ""
echo "=== Performance (hyperfine) ==="
hyperfine --warmup 5 --runs 50 --export-json perf.json './test' 2>&1 | grep -E "(Time|Mean)"

echo ""
echo "=== Memory Usage ==="
/usr/bin/time -l ./test 2>&1 | grep "maximum resident set size" | awk '{print "Peak RSS: " $1/1024 " KB"}'

echo ""
echo "=== Test Output ==="
./test

echo ""
echo "✓ Test complete"
