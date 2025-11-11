# Pure Python Generator Strategy: NumPy vs array.array

## Context
- User installing FFire CLI + learning schema language
- `pip install numpy` is not a burden for this audience
- NumPy is **the standard** for Python numerical computing

## Comparison

### array.array (current)
```python
# Pros
✓ Built-in (no dependencies)
✓ Simple, lightweight

# Cons
✗ Not industry standard
✗ Limited ecosystem integration
✗ No multi-dimensional support
✗ Doesn't work with pandas/scipy/sklearn/pytorch/tensorflow
```

### numpy.ndarray (proposed)
```python
# Pros
✓ Industry standard (de facto for numerical Python)
✓ Already installed for 99% of target users
✓ Integrates with entire data science ecosystem
✓ Zero-copy views and slicing
✓ Direct memory access via buffer protocol
✓ Better performance for large arrays
✓ Multi-dimensional support

# Cons
✗ External dependency (~50MB)
  → But target users already have it!
```

## Decision: Dual Support

Generate code that:
1. **Prefers NumPy** if available (check `import numpy`)
2. **Falls back to array.array** if NumPy not installed
3. **Accepts both types** on encode (duck typing)

## Implementation

```python
# Generated imports
import struct
import array as builtin_array
from typing import List, Optional

try:
    import numpy as np
    HAS_NUMPY = True
except ImportError:
    HAS_NUMPY = False

class Encoder:
    def write_float32_array(self, arr):
        """Write float32 array - accepts numpy.ndarray or array.array"""
        # Accept numpy arrays
        if HAS_NUMPY and isinstance(arr, np.ndarray):
            if arr.dtype != np.float32:
                arr = arr.astype(np.float32)
            byte_data = arr.tobytes()
        # Accept array.array
        elif isinstance(arr, builtin_array.array):
            if arr.typecode != 'f':
                arr = builtin_array.array('f', arr)
            byte_data = arr.tobytes()
        # Accept lists
        else:
            if HAS_NUMPY:
                arr = np.array(arr, dtype=np.float32)
                byte_data = arr.tobytes()
            else:
                arr = builtin_array.array('f', arr)
                byte_data = arr.tobytes()
        
        self.ensure_capacity(len(byte_data))
        self.buffer[self.pos:self.pos+len(byte_data)] = byte_data
        self.pos += len(byte_data)

class Decoder:
    def read_float32_array(self, length):
        """Read float32 array - returns numpy.ndarray if available, else array.array"""
        byte_len = length * 4
        
        if HAS_NUMPY:
            # Zero-copy numpy view (fastest)
            arr = np.frombuffer(self.data, dtype=np.float32, count=length, offset=self.pos)
            self.pos += byte_len
            return arr.copy()  # Return copy for safety
        else:
            # Fallback to array.array
            arr = builtin_array.array('f')
            arr.frombytes(self.data[self.pos:self.pos+byte_len])
            self.pos += byte_len
            return arr
```

## Benefits

1. **Best experience for 99% of users** (who have NumPy)
2. **Still works without NumPy** (graceful degradation)
3. **Integrates with data science stack**
4. **Type flexibility** - accepts lists, array.array, OR numpy arrays

## Recommendation

✅ **Implement dual NumPy/array.array support**
✅ **Document NumPy as recommended** (not required)
✅ **Make it seamless** - users don't think about it
