package main

import (
	"encoding/binary"
	"fmt"
	"time"
)

// Decode varint (up to 32-bit)
func decodeVarint(data []byte) (uint32, int) {
	var result uint32
	var shift uint
	for i := 0; i < len(data) && i < 5; i++ {
		b := data[i]
		result |= uint32(b&0x7F) << shift
		if b&0x80 == 0 {
			return result, i + 1
		}
		shift += 7
	}
	return result, 5
}

// Encode varint
func encodeVarint(val uint32) []byte {
	buf := make([]byte, 0, 5)
	for val >= 0x80 {
		buf = append(buf, byte(val)|0x80)
		val >>= 7
	}
	buf = append(buf, byte(val))
	return buf
}

func benchmarkVarint(values []uint32, iterations int) time.Duration {
	// Pre-encode all values
	encoded := make([][]byte, len(values))
	for i, v := range values {
		encoded[i] = encodeVarint(v)
	}

	start := time.Now()
	sum := uint64(0)
	for iter := 0; iter < iterations; iter++ {
		for _, data := range encoded {
			val, _ := decodeVarint(data)
			sum += uint64(val)
		}
	}
	elapsed := time.Since(start)

	// Prevent optimization
	if sum == 0 {
		fmt.Println("impossible")
	}

	return elapsed
}

func benchmarkUint32(values []uint32, iterations int) time.Duration {
	// Pre-encode all values
	encoded := make([]byte, len(values)*4)
	for i, v := range values {
		binary.LittleEndian.PutUint32(encoded[i*4:], v)
	}

	start := time.Now()
	sum := uint64(0)
	for iter := 0; iter < iterations; iter++ {
		for i := 0; i < len(values); i++ {
			val := binary.LittleEndian.Uint32(encoded[i*4:])
			sum += uint64(val)
		}
	}
	elapsed := time.Since(start)

	// Prevent optimization
	if sum == 0 {
		fmt.Println("impossible")
	}

	return elapsed
}

func main() {
	// Test case 1: Small values (< 128) - typical for your use case
	fmt.Println("=== Test 1: Small values (< 128) - 95% of your strings ===")
	smallValues := make([]uint32, 100)
	for i := range smallValues {
		smallValues[i] = uint32(10 + i%50) // 10-59 bytes
	}

	iterations := 100_000

	varTime := benchmarkVarint(smallValues, iterations)
	u32Time := benchmarkUint32(smallValues, iterations)

	totalOps := len(smallValues) * iterations
	fmt.Printf("Varint:  %v for %d decodes (%.2f ns/decode)\n",
		varTime, totalOps, float64(varTime.Nanoseconds())/float64(totalOps))
	fmt.Printf("Uint32:  %v for %d decodes (%.2f ns/decode)\n",
		u32Time, totalOps, float64(u32Time.Nanoseconds())/float64(totalOps))
	fmt.Printf("Varint is %.2fx slower\n", float64(varTime)/float64(u32Time))

	// Calculate size difference
	varSize := 0
	for _, v := range smallValues {
		varSize += len(encodeVarint(v))
	}
	u32Size := len(smallValues) * 4
	fmt.Printf("Size: varint=%d bytes, uint32=%d bytes (%.1f%% savings)\n\n",
		varSize, u32Size, 100.0*float64(u32Size-varSize)/float64(u32Size))

	// Test case 2: Mixed values (realistic distribution)
	fmt.Println("=== Test 2: Mixed values (small + some large) ===")
	mixedValues := make([]uint32, 100)
	for i := range mixedValues {
		if i < 90 {
			mixedValues[i] = uint32(10 + i%50) // Small
		} else {
			mixedValues[i] = uint32(1000 + i*100) // Large (2-byte varint)
		}
	}

	varTime = benchmarkVarint(mixedValues, iterations)
	u32Time = benchmarkUint32(mixedValues, iterations)

	fmt.Printf("Varint:  %v for %d decodes (%.2f ns/decode)\n",
		varTime, totalOps, float64(varTime.Nanoseconds())/float64(totalOps))
	fmt.Printf("Uint32:  %v for %d decodes (%.2f ns/decode)\n",
		u32Time, totalOps, float64(u32Time.Nanoseconds())/float64(totalOps))
	fmt.Printf("Varint is %.2fx slower\n", float64(varTime)/float64(u32Time))

	varSize = 0
	for _, v := range mixedValues {
		varSize += len(encodeVarint(v))
	}
	u32Size = len(mixedValues) * 4
	fmt.Printf("Size: varint=%d bytes, uint32=%d bytes (%.1f%% savings)\n\n",
		varSize, u32Size, 100.0*float64(u32Size-varSize)/float64(u32Size))

	// Test case 3: Worst case - large values
	fmt.Println("=== Test 3: Large values (worst case for varint) ===")
	largeValues := make([]uint32, 100)
	for i := range largeValues {
		largeValues[i] = uint32(1_000_000 + i*1000) // 3-4 byte varints
	}

	varTime = benchmarkVarint(largeValues, iterations)
	u32Time = benchmarkUint32(largeValues, iterations)

	fmt.Printf("Varint:  %v for %d decodes (%.2f ns/decode)\n",
		varTime, totalOps, float64(varTime.Nanoseconds())/float64(totalOps))
	fmt.Printf("Uint32:  %v for %d decodes (%.2f ns/decode)\n",
		u32Time, totalOps, float64(u32Time.Nanoseconds())/float64(totalOps))
	fmt.Printf("Varint is %.2fx slower\n", float64(varTime)/float64(u32Time))

	varSize = 0
	for _, v := range largeValues {
		varSize += len(encodeVarint(v))
	}
	u32Size = len(largeValues) * 4
	fmt.Printf("Size: varint=%d bytes, uint32=%d bytes (%.1f%% savings)\n",
		varSize, u32Size, 100.0*float64(u32Size-varSize)/float64(u32Size))
}
