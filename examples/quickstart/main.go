// Quickstart example demonstrating ffire encode/decode roundtrip.
package main

import (
	"fmt"

	person "github.com/shaban/ffire/examples/quickstart/generated"
)

func main() {
	// Create a message
	msg := person.PersonMessage{
		Name: "Alice",
		Age:  30,
	}
	fmt.Printf("Original: %+v\n", msg)

	// Encode to binary (method style)
	data := msg.Encode()
	fmt.Printf("Encoded:  %d bytes: %v\n", len(data), data)

	// Decode back (method style)
	var decoded person.PersonMessage
	if err := decoded.Decode(data); err != nil {
		panic(err)
	}
	fmt.Printf("Decoded:  %+v\n", decoded)

	// Verify roundtrip
	if msg.Name == decoded.Name && msg.Age == decoded.Age {
		fmt.Println("✓ Roundtrip successful!")
	}
}
