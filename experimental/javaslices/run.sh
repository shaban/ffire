#!/bin/bash

echo "Compiling IntSlice classes..."
javac IntSlice.java APITest.java PerformanceTest.java

if [ $? -eq 0 ]; then
    echo ""
    echo "=== Running API Test ==="
    java APITest
    
    echo ""
    echo ""
    echo "=== Running Performance Test ==="
    java PerformanceTest
else
    echo "Compilation failed!"
    exit 1
fi
