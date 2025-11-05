package main

// Package main generates test schemas and JSON data for ffire benchmarking.
// Run once: go run generator.go

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	schemaDir = "../../testdata/schema"
	jsonDir   = "../../testdata/json"
)

func main() {
	// Create output directories
	os.MkdirAll(schemaDir, 0755)
	os.MkdirAll(jsonDir, 0755)

	generators := []struct {
		name   string
		schema string
		data   interface{}
	}{
		{"array_int", schemaArrayInt, generateArrayInt()},
		{"array_float", schemaArrayFloat, generateArrayFloat()},
		{"array_string", schemaArrayString, generateArrayString()},
		{"array_struct", schemaArrayStruct, generateArrayStruct()},
		{"struct", schemaStruct, generateStruct()},
		{"nested", schemaNested, generateNested()},
		{"complex", schemaComplex, generateComplex()},
		{"optional", schemaOptional, generateOptional()},
		{"empty", schemaEmpty, generateEmpty()},
	}

	for _, g := range generators {
		fmt.Printf("Generating %s...\n", g.name)

		// Write schema
		schemaPath := filepath.Join(schemaDir, g.name+".ffi")
		if err := os.WriteFile(schemaPath, []byte(g.schema), 0644); err != nil {
			panic(err)
		}

		// Write JSON
		jsonData, _ := json.MarshalIndent(g.data, "", "  ")
		jsonPath := filepath.Join(jsonDir, g.name+".json")
		if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
			panic(err)
		}

		fmt.Printf("  Schema: %s\n", schemaPath)
		fmt.Printf("  JSON: %s (%d bytes)\n\n", jsonPath, len(jsonData))
	}

	fmt.Println("✅ All test data generated successfully!")
}

// ============================================================================
// Schemas
// ============================================================================

const schemaArrayInt = `package test

// Message represents an array of integers
type Message = []int32
`

const schemaArrayFloat = `package test

// Message represents an array of floats
type Message = []float32
`

const schemaArrayString = `package test

// Message represents an array of strings
type Message = []string
`

const schemaArrayStruct = `package test

// Message represents an array of device structs
type Message = []Device

type Device struct {
	Name     string
	Channels int32
	Rate     int32
}
`

const schemaStruct = `package test

// Message represents a single configuration struct
type Message = Config

type Config struct {
	Host       string
	Port       int32
	EnableSSL  bool
	Timeout    float32
	MaxRetries int32
}
`

const schemaNested = `package test

// Message represents deeply nested structure (10 levels)
type Message = Level1

type Level1 struct {
	Level2 Level2
}

type Level2 struct {
	Level3 Level3
}

type Level3 struct {
	Level4 Level4
}

type Level4 struct {
	Level5 Level5
}

type Level5 struct {
	Level6 Level6
}

type Level6 struct {
	Level7 Level7
}

type Level7 struct {
	Level8 Level8
}

type Level8 struct {
	Level9 Level9
}

type Level9 struct {
	Level10 Level10
}

type Level10 struct {
	Data []int32
}
`

const schemaComplex = `package test

// Message represents realistic plugin enumeration data
type Message = []Plugin

type Plugin struct {
	Name           string
	ManufacturerID string
	Type           string
	Subtype        string
	Parameters     []Parameter
}

type Parameter struct {
	DisplayName         string
	DefaultValue        float32
	CurrentValue        float32
	Address             int32
	MaxValue            float32
	MinValue            float32
	Unit                string
	Identifier          string
	CanRamp             bool
	IsWritable          bool
	RawFlags            int64
	IndexedValues       *[]string
	IndexedValuesSource *string
}
`

const schemaOptional = `package test

// Message represents data with optional fields
type Message = []Record

type Record struct {
	Required     string
	OptionalStr  *string
	OptionalInt  *int32
	OptionalBool *bool
}
`

const schemaEmpty = `package test

// Message represents edge case with empty collections
type Message = EmptyTest

type EmptyTest struct {
	EmptyString string
	EmptyArray  []int32
	ZeroValue   int32
}
`

// ============================================================================
// Data Generators
// ============================================================================

func generateArrayInt() []int32 {
	// Target: ~5000 elements for similar runtime
	// int32 = 4 bytes wire + 4 byte array header
	count := 5000
	result := make([]int32, count)
	for i := range result {
		result[i] = int32(i + 1)
	}
	return result
}

func generateArrayFloat() []float32 {
	// Target: ~5000 elements for similar runtime
	count := 5000
	result := make([]float32, count)
	for i := range result {
		result[i] = float32(i) + 0.5
	}
	return result
}

func generateArrayString() []string {
	// Target: ~500 strings (avg 40 chars = ~44 bytes wire each)
	// 500 * 44 = 22KB wire ≈ similar decode time to 5000 primitives
	count := 500
	result := make([]string, count)
	for i := range result {
		result[i] = fmt.Sprintf("device_%04d_with_some_extra_text", i+1)
	}
	return result
}

type Device struct {
	Name     string `json:"name"`
	Channels int32  `json:"channels"`
	Rate     int32  `json:"rate"`
}

func generateArrayStruct() []Device {
	// Target: ~200 devices
	// Each device: ~50 bytes wire (string + 2 int32s)
	// 200 * 50 = 10KB ≈ similar runtime
	count := 200
	result := make([]Device, count)
	for i := range result {
		result[i] = Device{
			Name:     fmt.Sprintf("Audio Device %03d", i+1),
			Channels: 2,
			Rate:     48000,
		}
	}
	return result
}

type Config struct {
	Host       string  `json:"host"`
	Port       int32   `json:"port"`
	EnableSSL  bool    `json:"enableSSL"`
	Timeout    float32 `json:"timeout"`
	MaxRetries int32   `json:"maxRetries"`
}

func generateStruct() Config {
	return Config{
		Host:       "localhost",
		Port:       8080,
		EnableSSL:  true,
		Timeout:    30.0,
		MaxRetries: 3,
	}
}

type Level1 struct {
	Level2 Level2 `json:"level2"`
}
type Level2 struct {
	Level3 Level3 `json:"level3"`
}
type Level3 struct {
	Level4 Level4 `json:"level4"`
}
type Level4 struct {
	Level5 Level5 `json:"level5"`
}
type Level5 struct {
	Level6 Level6 `json:"level6"`
}
type Level6 struct {
	Level7 Level7 `json:"level7"`
}
type Level7 struct {
	Level8 Level8 `json:"level8"`
}
type Level8 struct {
	Level9 Level9 `json:"level9"`
}
type Level9 struct {
	Level10 Level10 `json:"level10"`
}
type Level10 struct {
	Data []int32 `json:"data"`
}

func generateNested() Level1 {
	// Target: Similar runtime - bulk data at deepest level
	// ~5000 int32s nested deep
	data := make([]int32, 5000)
	for i := range data {
		data[i] = int32(i)
	}

	return Level1{
		Level2: Level2{
			Level3: Level3{
				Level4: Level4{
					Level5: Level5{
						Level6: Level6{
							Level7: Level7{
								Level8: Level8{
									Level9: Level9{
										Level10: Level10{
											Data: data,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

type Plugin struct {
	Name           string      `json:"name"`
	ManufacturerID string      `json:"manufacturerID"`
	Type           string      `json:"type"`
	Subtype        string      `json:"subtype"`
	Parameters     []Parameter `json:"parameters"`
}

type Parameter struct {
	DisplayName         string    `json:"displayName"`
	DefaultValue        float32   `json:"defaultValue"`
	CurrentValue        float32   `json:"currentValue"`
	Address             int32     `json:"address"`
	MaxValue            float32   `json:"maxValue"`
	MinValue            float32   `json:"minValue"`
	Unit                string    `json:"unit"`
	Identifier          string    `json:"identifier"`
	CanRamp             bool      `json:"canRamp"`
	IsWritable          bool      `json:"isWritable"`
	RawFlags            int64     `json:"rawFlags"`
	IndexedValues       *[]string `json:"indexedValues,omitempty"`
	IndexedValuesSource *string   `json:"indexedValuesSource,omitempty"`
}

func generateComplex() []Plugin {
	// Target: Realistic plugin data, ~20 plugins with varying parameter counts
	// Total size should give similar runtime to other tests
	plugins := []Plugin{
		{
			Name:           "Apple: AUVarispeed",
			ManufacturerID: "appl",
			Type:           "aufc",
			Subtype:        "vari",
			Parameters: []Parameter{
				{
					DisplayName:  "Playback Rate",
					DefaultValue: 1.0,
					CurrentValue: 1.0,
					Address:      0,
					MaxValue:     4.0,
					MinValue:     0.25,
					Unit:         "Rate",
					Identifier:   "0",
					CanRamp:      false,
					IsWritable:   true,
					RawFlags:     3502243840,
				},
				{
					DisplayName:  "Playback Pitch",
					DefaultValue: 0.0,
					CurrentValue: 0.0,
					Address:      1,
					MaxValue:     2400.0,
					MinValue:     -2400.0,
					Unit:         "Cents",
					Identifier:   "1",
					CanRamp:      false,
					IsWritable:   true,
					RawFlags:     3498049536,
				},
			},
		},
		{
			Name:           "Apple: AUFilter",
			ManufacturerID: "appl",
			Type:           "aufx",
			Subtype:        "filt",
			Parameters: []Parameter{
				{
					DisplayName:         "Low Filter Type",
					DefaultValue:        0,
					CurrentValue:        0,
					Address:             0,
					MaxValue:            1,
					MinValue:            0,
					Unit:                "Indexed",
					Identifier:          "0",
					CanRamp:             false,
					IsWritable:          true,
					RawFlags:            3222274048,
					IndexedValues:       &[]string{"Low Shelf", "High Pass"},
					IndexedValuesSource: strPtr("valueStrings"),
				},
				{
					DisplayName:  "Low Frequency",
					DefaultValue: 100.0,
					CurrentValue: 100.0,
					Address:      1,
					MaxValue:     21829.5,
					MinValue:     10.0,
					Unit:         "Hertz",
					Identifier:   "1",
					CanRamp:      false,
					IsWritable:   true,
					RawFlags:     3226468352,
				},
				{
					DisplayName:  "Low Gain",
					DefaultValue: 0.0,
					CurrentValue: 0.0,
					Address:      2,
					MaxValue:     18.0,
					MinValue:     -18.0,
					Unit:         "Decibels",
					Identifier:   "2",
					CanRamp:      false,
					IsWritable:   true,
					RawFlags:     3222274048,
				},
			},
		},
	}

	// Duplicate and vary to reach target size
	result := make([]Plugin, 0, 20)
	for i := 0; i < 20; i++ {
		p := plugins[i%len(plugins)]
		p.Name = fmt.Sprintf("%s [Instance %d]", p.Name, i+1)
		result = append(result, p)
	}

	return result
}

type Record struct {
	Required     string  `json:"required"`
	OptionalStr  *string `json:"optionalStr,omitempty"`
	OptionalInt  *int32  `json:"optionalInt,omitempty"`
	OptionalBool *bool   `json:"optionalBool,omitempty"`
}

func generateOptional() []Record {
	// Mix of present and absent optional fields
	result := make([]Record, 1000)
	for i := range result {
		rec := Record{
			Required: fmt.Sprintf("record_%04d", i+1),
		}

		// Vary optional field presence
		if i%3 == 0 {
			rec.OptionalStr = strPtr(fmt.Sprintf("optional_%d", i))
		}
		if i%4 == 0 {
			rec.OptionalInt = int32Ptr(int32(i * 100))
		}
		if i%5 == 0 {
			rec.OptionalBool = boolPtr(true)
		}

		result[i] = rec
	}
	return result
}

type EmptyTest struct {
	EmptyString string  `json:"emptyString"`
	EmptyArray  []int32 `json:"emptyArray"`
	ZeroValue   int32   `json:"zeroValue"`
}

func generateEmpty() EmptyTest {
	return EmptyTest{
		EmptyString: "",
		EmptyArray:  []int32{},
		ZeroValue:   0,
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
