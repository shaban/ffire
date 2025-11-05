package main

/*
int noop() {
    return 42;
}

int add(int a, int b) {
    return a + b;
}
*/
import "C"
import (
	"fmt"
	"time"
)

func main() {
	// Warmup
	for i := 0; i < 1000; i++ {
		C.noop()
	}

	// Benchmark noop call
	iterations := 10_000_000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		C.noop()
	}
	elapsed := time.Since(start)
	perCall := elapsed.Nanoseconds() / int64(iterations)

	fmt.Printf("CGO noop call overhead:\n")
	fmt.Printf("  Total: %v for %d calls\n", elapsed, iterations)
	fmt.Printf("  Per call: %d ns\n", perCall)
	fmt.Printf("  Throughput: %.2f million calls/sec\n\n", float64(iterations)/elapsed.Seconds()/1_000_000)

	// Benchmark with minimal work
	start = time.Now()
	sum := 0
	for i := 0; i < iterations; i++ {
		sum += int(C.add(C.int(i), C.int(1)))
	}
	elapsed = time.Since(start)
	perCall = elapsed.Nanoseconds() / int64(iterations)

	fmt.Printf("CGO add(i, 1) call overhead:\n")
	fmt.Printf("  Total: %v for %d calls\n", elapsed, iterations)
	fmt.Printf("  Per call: %d ns\n", perCall)
	fmt.Printf("  Throughput: %.2f million calls/sec\n", float64(iterations)/elapsed.Seconds()/1_000_000)
	fmt.Printf("  (sum=%d to prevent optimization)\n", sum)
}
