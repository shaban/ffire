c abi improvements for better python js php integration
magefile root building
distribution
wire format change ordered wire alphabetically, static sizes first?


swift:
This is an excellent comprehensive analysis. Now let me provide you with the detailed optimization plan:

Plan: Swift Serialization Performance Optimization
Optimize the naive Swift ffire implementation by implementing bulk memory operations, eliminating unnecessary allocations, and leveraging Swift's low-level pointer APIs to achieve performance comparable to C++.

Key Findings
The current Swift implementation has several performance bottlenecks:

Element-by-element encoding - Each primitive in an array calls withUnsafeBytes separately
Intermediate allocations - Array(string.utf8) creates a temporary array for every string
Fixed buffer capacity - reserveCapacity(1024) regardless of actual size
No bulk memory operations - Missing memcpy-style bulk copies for primitive arrays
Individual position increments - Separate pos += N after every read
Steps
Implement bulk primitive array encoding in generator_swift.go - Add withUnsafeBytes { buffer.append(contentsOf: $0) } pattern for int8/16/32/64, float32/64 arrays instead of element-by-element encoding.

Implement bulk primitive array decoding in generator_swift.go - Use [Int32](unsafeUninitializedCapacity:initializingWith:) with memcpy for zero-copy array construction instead of .map { } loops.

Optimize string encoding in generateSwiftHelpers() - Replace Array(string.utf8) with string.withUTF8 { } to avoid intermediate array allocation.

Add dynamic buffer sizing - Calculate expected buffer size based on message contents instead of fixed 1024 bytes, or use heuristic message.count * averageElementSize.

Add helper functions for primitive reads - Create inline readInt32, readFloat32 etc. with defer { pos += N } to reduce code size and improve optimizer opportunities.

Further Considerations
ContiguousArray vs Array - Using ContiguousArray<T> guarantees contiguous storage but changes public API. Recommend for internal buffers only. Keep current Array for API compatibility.

Endianness handling - All Apple platforms are little-endian matching wire format. Add #if _endian(big) guards for theoretical portability, or document little-endian requirement.

Batch consecutive field encoding - For structs with multiple floats (e.g., DefaultValue, CurrentValue, MaxValue, MinValue), batch into single withUnsafeBytes call. Medium complexity, P2 priority.