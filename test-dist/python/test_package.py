#!/usr/bin/env python3
"""Test the generated Python package."""

import sys
sys.path.insert(0, '/Users/shaban/Code/ffire/test-dist/python')

from test import Message

# Load test data
with open('/Users/shaban/Code/ffire/experimental/cpp-bindings/common/complex.bin', 'rb') as f:
    data = f.read()

print(f"Loaded {len(data)} bytes")

# Test decode
try:
    msg = Message.decode(data)
    print("✓ Decode successful")
except Exception as e:
    print(f"✗ Decode failed: {e}")
    sys.exit(1)

# Test encode
try:
    encoded = msg.encode()
    print(f"✓ Encode successful: {len(encoded)} bytes")
except Exception as e:
    print(f"✗ Encode failed: {e}")
    sys.exit(1)

# Verify round-trip
if len(encoded) == len(data):
    print("✓ Round-trip size matches!")
else:
    print(f"✗ Size mismatch: {len(encoded)} vs {len(data)}")
    sys.exit(1)

print("\n✅ All tests passed!")
