# Go API

Programmatic usage of ffire from Go.

## Generator API

```go
import "github.com/shaban/ffire/pkg/generator"

// Parse schema
schema, err := parser.Parse("schema.ffi")

// Generate package
config := &generator.PackageConfig{
    Schema:    schema,
    Language:  "go",
    OutputDir: "./output",
    Optimize:  2,
}
err = generator.GeneratePackage(config)
```

## Using Generated Code

```go
import "your-module/generated"

// Encode
msg := &Message{Value: 42}
data, err := msg.Encode()

// Decode  
msg, err := DecodeMessage(data)
```

## Schema Parsing

```go
import "github.com/shaban/ffire/pkg/parser"

schema, err := parser.Parse("types.ffi")
if err != nil {
    log.Fatal(err)
}

// Inspect
for _, msg := range schema.Messages {
    fmt.Printf("Message: %s\n", msg.Name)
}
```

## Benchmark Utilities

```go
import "github.com/shaban/ffire/pkg/benchmark"

// Generate benchmark
err := benchmark.GenerateGo(
    schema,
    "array_int",
    "IntList", 
    jsonData,
    "./output",
    10000,
)
```

## Fixture Conversion

```go
import "github.com/shaban/ffire/pkg/fixture"

// JSON to binary
binary, err := fixture.Convert(schema, "Message", jsonData)

// Binary to JSON
json, err := fixture.ToJSON(schema, "Message", binary)
```
