#!/bin/bash
# Run all cpp-bindings tests and collect results

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR/.."

echo "========================================"
echo "C++ Bindings Test Matrix"
echo "========================================"
echo ""

# Array of test directories
TESTS=(
    "01-cpp-direct"
    "02-cpp-static"
    "03-cpp-dynamic"
    "06-objcpp-direct"
    "07-objcpp-static"
    "08-objcpp-dynamic"
    "09-python-dynamic"
)

# Run each test
for test in "${TESTS[@]}"; do
    echo ""
    echo "========================================" 
    echo "Running: $test"
    echo "========================================" 
    cd "$test"
    ./run.sh
    cd ..
    echo ""
done

echo ""
echo "========================================"
echo "All tests complete!"
echo "========================================"
echo ""
echo "To generate results summary:"
echo "  cd scripts && ./generate_report.sh"
