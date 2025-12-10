c abi improvements for better python js php integration
magefile root building
distribution
wire format change ordered wire alphabetically, static sizes first?

Based on my deep understanding of the wire format and optimization techniques from Swift, here are the biggest opportunities for improvement:

High-Impact Optimization Opportunities
1. C++ Decode (array_float, array_int, nested) - HUGE potential
Current: 2,389-2,404 ns decode for primitive arrays
Swift: 570-596 ns decode (4x faster!)
Issue: C++ is likely doing element-by-element decoding instead of bulk memcpy
Fix: Same technique as Swift - single memcpy for entire primitive array
Expected gain: 4x faster decode → totals drop from ~3,100 ns to ~1,500 ns
2. Dart (all benchmarks) - Pervasive issues
array_float decode: 13,629 ns (vs Swift 570 ns = 24x slower)
array_int decode: 12,807 ns (vs Swift 596 ns = 21x slower)
nested decode: 12,581 ns (vs Swift 466 ns = 27x slower)
Issue: Almost certainly per-element decoding + typed array creation overhead
Fix: Use ByteData views and bulk typed list creation
Expected gain: 10-20x on primitive array operations
3. Go (array_int, array_float, nested) - Moderate potential
array_int decode: 2,472 ns (vs Swift 596 ns = 4x slower)
Issue: Likely using binary.Read per element instead of unsafe cast
Fix: Use unsafe.Slice to reinterpret byte slice as []int32/[]float32
Expected gain: 2-3x on primitive array decode
4. Java array_string - Encode bottleneck
Current: 10,052 ns encode (vs Go 5,496 ns)
Issue: String handling, possibly charset encoding overhead
Fix: Pre-size buffer, use String.getBytes() with direct buffer writes
Expected gain: ~40% improvement
5. Swift array_struct decode - Still has room
Current: 8,915 ns (vs C++ 1,536 ns = 6x slower)
Issue: We optimized primitive arrays inside nested, but array_struct has strings
Fix: Pre-allocate array, reduce String construction overhead
Expected gain: 2-3x possible
Priority Ranking
Priority	Language	Benchmark	Current	Target	Effort
|🥇 1	C++	array_float/int decode	2,400 ns	~600 ns	Low
|🥇 1	Dart	array_float/int decode	13,000 ns	~2,000 ns	Medium
|🥈 2	C++	nested decode	2,393 ns	~500 ns	Low
|🥈 2	Dart	nested decode	12,581 ns	~2,000 ns	Medium
|🥉 3	Go	array_int decode	2,472 ns	~800 ns	Medium
Recommendation
Start with C++ array decode - it's the lowest-hanging fruit:

C++ should be the fastest language, but Swift is beating it on decode by 4x
The fix is straightforward: replace element-by-element decoding with memcpy
High confidence in 4x improvement
Would you like me to investigate and optimize C++ first?