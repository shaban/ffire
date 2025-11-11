package main

import (
	"protocompare/pb"

	"fmt"
	"os"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func main() {
	// Load JSON fixture
	jsonData, err := os.ReadFile("complex_wrapped.json")
	if err != nil {
		panic(err)
	}

	// Parse JSON to protobuf
	var list pb.PluginList
	if err := protojson.Unmarshal(jsonData, &list); err != nil {
		panic(err)
	}

	// Warmup
	for i := 0; i < 1000; i++ {
		data, _ := proto.Marshal(&list)
		var decoded pb.PluginList
		proto.Unmarshal(data, &decoded)
	}

	// Benchmark encode
	iterations := 100000
	start := time.Now()
	var encoded []byte
	for i := 0; i < iterations; i++ {
		encoded, _ = proto.Marshal(&list)
	}
	encodeNs := time.Since(start).Nanoseconds() / int64(iterations)

	// Benchmark decode
	start = time.Now()
	for i := 0; i < iterations; i++ {
		var decoded pb.PluginList
		proto.Unmarshal(encoded, &decoded)
	}
	decodeNs := time.Since(start).Nanoseconds() / int64(iterations)

	fmt.Printf("Go protobuf results:\n")
	fmt.Printf("encode_ns: %d\n", encodeNs)
	fmt.Printf("decode_ns: %d\n", decodeNs)
	fmt.Printf("total_ns: %d\n", encodeNs+decodeNs)
	fmt.Printf("wire_size: %d\n", len(encoded))
}
