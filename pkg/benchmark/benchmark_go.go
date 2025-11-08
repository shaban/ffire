// Package benchmark generates self-contained benchmark executables for ffire schemas.
package benchmark

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/shaban/ffire/pkg/fixture"
	"github.com/shaban/ffire/pkg/generator"
	"github.com/shaban/ffire/pkg/schema"
)

// GenerateGo creates a complete Go benchmark executable in the output directory.
func GenerateGo(s *schema.Schema, schemaName string, messageName string, jsonData []byte, outputDir string, iterations int) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the encoder/decoder code as package main
	generatedCode, err := generator.Generate(s, generator.LanguageGo)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Rewrite package declaration to "main" for single-directory benchmark
	generatedCodeStr := string(generatedCode)
	// Find the package line (after the generated comment)
	packageLine := "package " + s.Package + "\n"
	packageIdx := strings.Index(generatedCodeStr, packageLine)
	if packageIdx == -1 {
		return fmt.Errorf("could not find package declaration in generated code")
	}
	// Keep everything before the package line, replace it with "package main", then add the rest
	generatedCodeStr = generatedCodeStr[:packageIdx] + "package main\n" + generatedCodeStr[packageIdx+len(packageLine):]

	// Write generated code
	generatedFile := filepath.Join(outputDir, "generated.go")
	if err := os.WriteFile(generatedFile, []byte(generatedCodeStr), 0644); err != nil {
		return fmt.Errorf("failed to write generated code: %w", err)
	}

	// Convert JSON to binary fixture
	binaryData, err := fixture.Convert(s, messageName, jsonData)
	if err != nil {
		return fmt.Errorf("failed to convert fixture: %w", err)
	}

	// Write binary fixture
	fixtureFile := filepath.Join(outputDir, "fixture.bin")
	if err := os.WriteFile(fixtureFile, binaryData, 0644); err != nil {
		return fmt.Errorf("failed to write fixture: %w", err)
	}

	// Find the message type
	var messageType *schema.MessageType
	for i := range s.Messages {
		if s.Messages[i].Name == messageName {
			messageType = &s.Messages[i]
			break
		}
	}
	if messageType == nil {
		return fmt.Errorf("message type %s not found", messageName)
	}

	// Determine root type name for function naming
	rootTypeName := getRootTypeName(messageType.TargetType)

	// Generate benchmark main
	benchData := BenchmarkData{
		Package:      s.Package,
		SchemaName:   schemaName,
		MessageName:  messageName,
		TypeName:     rootTypeName,
		Iterations:   iterations,
		FixtureBytes: len(binaryData),
	}

	var buf bytes.Buffer
	if err := goBenchTemplate.Execute(&buf, benchData); err != nil {
		return fmt.Errorf("failed to generate benchmark: %w", err)
	}

	// Write benchmark main
	benchFile := filepath.Join(outputDir, "bench.go")
	if err := os.WriteFile(benchFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write benchmark main: %w", err)
	}

	// Write minimal go.mod (needed for go:embed)
	goMod := "module bench\n\ngo 1.21\n"
	modFile := filepath.Join(outputDir, "go.mod")
	if err := os.WriteFile(modFile, []byte(goMod), 0644); err != nil {
		return fmt.Errorf("failed to write go.mod: %w", err)
	}

	return nil
}

// BenchmarkData holds template data for benchmark generation.
type BenchmarkData struct {
	Package      string
	SchemaName   string
	MessageName  string
	TypeName     string
	Iterations   int
	FixtureBytes int
}

// getRootTypeName extracts the type name for function naming.
// Matches generator.rootTypeName - for arrays, returns the element type name (not "Array" suffix).
func getRootTypeName(typ schema.Type) string {
	switch t := typ.(type) {
	case *schema.PrimitiveType:
		// Capitalize first letter to match generator
		if len(t.Name) > 0 {
			return string(t.Name[0]-32) + t.Name[1:] // Simple uppercase first char
		}
		return t.Name
	case *schema.StructType:
		return t.Name
	case *schema.ArrayType:
		// For arrays, return the element type name (generator doesn't add "Array" suffix)
		return getRootTypeName(t.ElementType)
	default:
		return "Message"
	}
}

var goBenchTemplate = template.Must(template.New("bench").Parse(`package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

//go:embed fixture.bin
var fixtureData []byte

type BenchResult struct {
	Language   string ` + "`json:\"language\"`" + `
	Format     string ` + "`json:\"format\"`" + `
	Message    string ` + "`json:\"message\"`" + `
	Iterations int    ` + "`json:\"iterations\"`" + `
	EncodeNs   int64  ` + "`json:\"encode_ns\"`" + `
	DecodeNs   int64  ` + "`json:\"decode_ns\"`" + `
	TotalNs    int64  ` + "`json:\"total_ns\"`" + `
	WireSize   int    ` + "`json:\"wire_size\"`" + `
	FixtureSize int   ` + "`json:\"fixture_size\"`" + `
	Timestamp  string ` + "`json:\"timestamp\"`" + `
}

func main() {
	iterations := {{.Iterations}}
	jsonOutput := os.Getenv("BENCH_JSON") == "1"
	
	// Decode fixture to get original data
	original, err := Decode{{.TypeName}}Message(fixtureData)
	if err != nil {
		panic(fmt.Sprintf("failed to decode fixture: %v", err))
	}
	
	// Warmup
	for i := 0; i < 1000; i++ {
		encoded := Encode{{.TypeName}}Message(original)
		_, _ = Decode{{.TypeName}}Message(encoded)
	}
	
	// Benchmark encode
	start := time.Now()
	var encoded []byte
	for i := 0; i < iterations; i++ {
		encoded = Encode{{.TypeName}}Message(original)
	}
	encodeTime := time.Since(start)
	
	// Benchmark decode
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_, _ = Decode{{.TypeName}}Message(encoded)
	}
	decodeTime := time.Since(start)
	
	// Calculate metrics
	encodeNs := encodeTime.Nanoseconds() / int64(iterations)
	decodeNs := decodeTime.Nanoseconds() / int64(iterations)
	totalNs := encodeNs + decodeNs
	
	if jsonOutput {
		// Output JSON for automation
		result := BenchResult{
			Language:    "Go",
			Format:      "ffire",
			Message:     "{{.SchemaName}}",
			Iterations:  iterations,
			EncodeNs:    encodeNs,
			DecodeNs:    decodeNs,
			TotalNs:     totalNs,
			WireSize:    len(encoded),
			FixtureSize: len(fixtureData),
			Timestamp:   time.Now().Format(time.RFC3339),
		}
		json.NewEncoder(os.Stdout).Encode(result)
	} else {
		// Print human-readable results
		fmt.Printf("ffire benchmark: {{.SchemaName}}\n")
		fmt.Printf("Iterations:  %d\n", iterations)
		fmt.Printf("Encode:      %d ns/op\n", encodeNs)
		fmt.Printf("Decode:      %d ns/op\n", decodeNs)
		fmt.Printf("Total:       %d ns/op\n", totalNs)
		fmt.Printf("Wire size:   %d bytes\n", len(encoded))
		fmt.Printf("Fixture:     %d bytes\n", len(fixtureData))
		fmt.Printf("Total time:  %.2fs\n", (encodeTime + decodeTime).Seconds())
	}
}
`))
