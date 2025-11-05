package main

import (
	"encoding/binary"
	"fmt"
	"time"
	"unsafe"
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

// Simulate array decoding with varint length
func decodeArrayVarint(data []byte) []float32 {
	length, offset := decodeVarint(data)

	// Allocate destination
	result := make([]float32, length)

	// Bulk copy the array data
	dataBytes := data[offset : offset+int(length)*4]
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&result[0])), len(result)*4), dataBytes)

	return result
}

// Simulate array decoding with uint32 length
func decodeArrayUint32(data []byte) []float32 {
	length := binary.LittleEndian.Uint32(data[0:4])

	// Allocate destination
	result := make([]float32, length)

	// Bulk copy the array data
	dataBytes := data[4 : 4+int(length)*4]
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&result[0])), len(result)*4), dataBytes)

	return result
}

// Encode array with varint
func encodeArrayVarint(arr []float32) []byte {
	lenBytes := encodeVarint(uint32(len(arr)))
	result := make([]byte, len(lenBytes)+len(arr)*4)
	copy(result, lenBytes)
	copy(result[len(lenBytes):], unsafe.Slice((*byte)(unsafe.Pointer(&arr[0])), len(arr)*4))
	return result
}

// Encode array with uint32
func encodeArrayUint32(arr []float32) []byte {
	result := make([]byte, 4+len(arr)*4)
	binary.LittleEndian.PutUint32(result[0:4], uint32(len(arr)))
	copy(result[4:], unsafe.Slice((*byte)(unsafe.Pointer(&arr[0])), len(arr)*4))
	return result
}

func benchmarkArrays(arraySize int, numArrays int, iterations int) {
	fmt.Printf("=== Array size: %d floats (%d bytes data) ===\n", arraySize, arraySize*4)

	// Create test arrays
	arrays := make([][]float32, numArrays)
	for i := range arrays {
		arrays[i] = make([]float32, arraySize)
		for j := range arrays[i] {
			arrays[i][j] = float32(i*1000 + j)
		}
	}

	// Pre-encode with varint
	encodedVarint := make([][]byte, numArrays)
	for i, arr := range arrays {
		encodedVarint[i] = encodeArrayVarint(arr)
	}

	// Pre-encode with uint32
	encodedUint32 := make([][]byte, numArrays)
	for i, arr := range arrays {
		encodedUint32[i] = encodeArrayUint32(arr)
	}

	// Benchmark varint
	start := time.Now()
	for iter := 0; iter < iterations; iter++ {
		for _, data := range encodedVarint {
			result := decodeArrayVarint(data)
			_ = result[0] // Prevent optimization
		}
	}
	varTime := time.Since(start)

	// Benchmark uint32
	start = time.Now()
	for iter := 0; iter < iterations; iter++ {
		for _, data := range encodedUint32 {
			result := decodeArrayUint32(data)
			_ = result[0] // Prevent optimization
		}
	}
	u32Time := time.Since(start)

	totalOps := numArrays * iterations
	varPerOp := float64(varTime.Nanoseconds()) / float64(totalOps)
	u32PerOp := float64(u32Time.Nanoseconds()) / float64(totalOps)

	fmt.Printf("Varint:  %v for %d decodes (%.2f ns/array)\n", varTime, totalOps, varPerOp)
	fmt.Printf("Uint32:  %v for %d decodes (%.2f ns/array)\n", u32Time, totalOps, u32PerOp)
	fmt.Printf("Varint is %.2fx slower (%.2f ns overhead)\n", varPerOp/u32PerOp, varPerOp-u32PerOp)

	// Size comparison
	varSize := len(encodedVarint[0])
	u32Size := len(encodedUint32[0])
	fmt.Printf("Size: varint=%d bytes, uint32=%d bytes (%d bytes saved)\n", varSize, u32Size, u32Size-varSize)
	fmt.Printf("Overhead: varint=%.2f%%, uint32=%.2f%% of data\n\n",
		100.0*float64(varSize-arraySize*4)/float64(arraySize*4),
		100.0*float64(u32Size-arraySize*4)/float64(arraySize*4))
}

func main() {
	iterations := 100_000
	numArrays := 10

	// Small arrays (typical for your use case - parameter lists, device lists)
	benchmarkArrays(10, numArrays, iterations)  // 40 bytes
	benchmarkArrays(50, numArrays, iterations)  // 200 bytes
	benchmarkArrays(100, numArrays, iterations) // 400 bytes

	// Medium arrays (plugin parameter tree)
	benchmarkArrays(500, numArrays, iterations/10)  // 2KB
	benchmarkArrays(1000, numArrays, iterations/10) // 4KB

	// Large arrays (if you ever pass bulk data)
	benchmarkArrays(10000, numArrays, iterations/100) // 40KB
}
