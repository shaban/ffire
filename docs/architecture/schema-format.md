# ffire Schema Format Specification
## ffire - FFI Encoding

## File Format
- **Extension**: `.ffi`
- **Syntax**: Go struct syntax
- **Required**: Package declaration
- **REQUIRED CONVENTION**: All message types MUST use `Message` suffix

## Naming Convention

**Generators automatically append `Message` suffix to all root types.**

Users write natural, clean type names:
```go
// ✅ User writes clean names
type Config struct { ... }
type Device struct { ... }
type LoginRequest struct { ... }
```

Generators produce types with `Message` suffix:
```swift
// Generated Swift
public class ConfigMessage { ... }
public class DeviceMessage { ... }
public class LoginRequestMessage { ... }
```

**Rationale:**
- **User-friendly**: Clean schema names without awkward suffixes
- **Zero keyword collisions**: Generated types safe across all 11 languages
- **Industry standard**: Protobuf, gRPC use "Message" terminology
- **Clear boundary**: Generated `ConfigMessage` ≠ user's domain `Config`
- **No maintenance**: No keyword lists needed
- **Future-proof**: Works for any language we add

## Type Definitions

### Root Types (Message-Capable Types)

Root types are automatically inferred based on two conditions:
1. **Not referenced** by any other type (as field, embedded type, or array element)
2. **Exported** (starts with uppercase letter)

These root types generate public encode/decode functions in the C ABI.

**Example:**
```go
package myapp

// Helper types (referenced by other types)
type Option struct {
    Name  string
    Value int32
}

type Metadata struct {
    Version string
    Author  string
}

// Root message types (unreferenced + exported)
type Config struct {                // ✓ Root: unreferenced + uppercase
    Host     string
    Port     int32
    Options  []Option                // Option is consumed here
    Metadata Metadata                // Metadata is consumed here
}

type DeviceList []Device            // ✓ Root: unreferenced + uppercase

type Device struct {                // Not root: referenced by DeviceList
    Name string
    ID   int32
}

type internal struct {              // Not root: lowercase (not exported)
    value int32
}
```

**Generated types (automatic Message suffix):**
```swift
// Swift
public class ConfigMessage { ... }       // From Config
public class DeviceListMessage { ... }   // From DeviceList
```

**Generated C ABI functions:**
```c
// C ABI uses lowercase of original name (no Message in function names)
ConfigMessageHandle config_decode(const uint8_t* data, size_t len, char** error);
size_t config_encode(ConfigMessageHandle handle, uint8_t** out_data, char** error);
void config_free(ConfigMessageHandle handle);

DeviceListMessageHandle devicelist_decode(const uint8_t* data, size_t len, char** error);
size_t devicelist_encode(DeviceListMessageHandle handle, uint8_t** out_data, char** error);
void devicelist_free(DeviceListMessageHandle handle);
```

**Rules:**
- At least one root type required (parser error otherwise)
- Root types must be exported (start with uppercase)
- Helper types can be public or private (user's choice)
- Function names derived from type name: `{type}_decode`, `{type}_encode`, `{type}_free`

### Struct Types
```go
type Person struct {
    Name    string
    Age     int32
    Email   *string  // Optional field
    Tags    []string
}
```
- Fields in declaration order
- Supported field types: primitives, strings, arrays, nested structs, optionals
- **All fields must be named** - embedded/anonymous structs are not supported

**Named vs Embedded Structs:**
```go
// ✓ Supported: Named nested struct
type Config struct {
    Port     int32
    Metadata Metadata  // Named field
}
// Access: config.Metadata.Version

// ✗ Not supported: Embedded struct
type Config struct {
    Port     int32
    Metadata          // Anonymous field - NOT SUPPORTED
}
```

**Rationale:** 
- Wire format encodes all structs identically (no distinction between named/embedded)
- Embedded structs are a Go-specific language feature for field promotion
- Not applicable to cross-language serialization (C++, Swift don't have this concept)
- Can be added in future if needed without breaking wire format compatibility

### Struct Tags

**Full Tag Preservation:**
```go
type User struct {
    ID    int64  `json:"id" db:"user_id" validate:"required"`
    Name  string `json:"name" yaml:"name" xml:"Name,attr"`
    Email string `json:"email,omitempty" validate:"email"`
}
```

**Behavior:**
- **All struct tags are preserved** verbatim in generated Go code
- Full tag string stored: `` `json:"id" db:"user_id" validate:"required"` ``
- JSON tags specifically parsed for validation and fixture conversion
- Other tags (yaml, xml, db, validate, etc.) passed through unchanged

**JSON Tag Usage:**
```go
type Plugin struct {
    Name           string `json:"name"`                    // Maps to "name" in JSON
    ManufacturerID string `json:"manufacturerID"`          // Maps to "manufacturerID" in JSON
    Type           string `json:"type,omitempty"`          // Optional in JSON output
    Parameters     []Parameter `json:"parameters"`
}
```
- Used by `ffire validate --json` to match JSON field names
- Used by `ffire fixture` to convert JSON to binary
- Supports `omitempty` and other standard JSON options
- If no JSON tag present, uses struct field name

**Multiple Tags Example:**
```go
type Config struct {
    Host     string `json:"host" yaml:"host" env:"APP_HOST"`
    Port     int32  `json:"port" yaml:"port" env:"APP_PORT" validate:"min=1,max=65535"`
    Database string `json:"db" yaml:"database" env:"DATABASE_URL" validate:"required"`
}
```
- All tags preserved for generated Go code
- JSON tag used for JSON operations only
- Generated C++/Swift code ignores all tags (not applicable)

**Tag Format:**
- Standard Go struct tag syntax: `` `key:"value" key2:"value2"` ``
- JSON tag format: `json:"fieldName"` or `json:"fieldName,omitempty"`
- Commas in tag values (like `omitempty`) are preserved
- Multiple spaces between tags are preserved

### Primitive Types
- `bool`, `int8`, `int16`, `int32`, `int64`
- `float32`, `float64`
- `string`

### Optional Fields
```go
type Config struct {
    Host string
    Port *int32  // Optional via pointer
}
```
- Use `*Type` for optional fields
- Encoded as: `[bool: present][value if present]`

### Arrays
```go
type Matrix = [][]float32  // 2D array
type Names = []string      // Array of strings
type Devices = []Device    // Array of structs
```
- Any type can be in array
- Max nesting: 32 levels
- Max array length: 65,535 elements (wire format limit)
- Max string length: 65,535 bytes (wire format limit)

## Wire Format Limits

All schemas must respect wire format constraints:
- **Strings**: Maximum 65,535 bytes (uint16) per string
- **Arrays**: Maximum 65,535 elements (uint16) per array
- **Nesting**: Maximum 32 levels deep
- **Messages**: Maximum 2GB total size

**Rationale**: These limits are enforced at the wire format level for safety by design. The uint16 length fields physically prevent buffer overflow and memory exhaustion attacks without requiring runtime bounds checking.

**Best practices**:
- Plugin names, config values: Well under 64KB ✓
- Device lists, parameter arrays: Usually < 1000 elements ✓
- Audio/graphics data: Keep on native side, pass handles only ✓

## Code Generation

### Go Output
```go
// For root type: DeviceList
func EncodeDeviceList(v []Device) []byte
func DecodeDeviceList(data []byte) ([]Device, error)

// Private helpers for nested types
func encodeDevice(buf *bytes.Buffer, v Device)
func decodeDevice(r *bytes.Reader) (Device, error)
```

### C++ Output
```cpp
// For root type: DeviceList
namespace package_name {
    std::vector<uint8_t> encode_devicelist_message(const std::vector<Device>& v);
    std::vector<Device> decode_devicelist_message(const std::vector<uint8_t>& data);
    
    // Private helpers
    void encode_device(std::vector<uint8_t>& buf, const Device& v);
    Device decode_device(const uint8_t*& ptr, const uint8_t* end);
}
```

### C ABI Output
```c
// For root type: DeviceList
DeviceListHandle deviceList_decode(const uint8_t* data, size_t len, char** error);
size_t deviceList_encode(DeviceListHandle handle, uint8_t** out_data, char** error);
void deviceList_free(DeviceListHandle handle);
void deviceList_free_data(uint8_t* data);
void deviceList_free_error(char* error);

// Helper types only get typedefs
typedef void* DeviceHandle;
```

### Naming Conventions
- **Go**: `PascalCase` for public, `camelCase` for private
- **C++**: `snake_case` for all functions/types
- **Transformation**: Via [strcase](https://github.com/iancoleman/strcase) package
- **Package → Namespace**: Direct mapping

## Example Schema

**audio.ffi:**
```go
package audio

// Root types (unreferenced + exported = auto-inferred)
type DeviceList []Device

type PluginInfo struct {
    Name   string
    Vendor string
    Params []Parameter
}

// Helper types (referenced by root types)
type Device struct {
    Name      string
    ID        string
    Channels  int32
    IsDefault bool
}

type Parameter struct {
    Label        string  `json:"label" yaml:"label"`
    DefaultValue float32 `json:"defaultValue" yaml:"default_value"`
    Unit         *string `json:"unit,omitempty" yaml:"unit,omitempty"`  // Optional
}
```

**Tag Behavior by Language:**

| Language | Tag Handling |
|----------|--------------|
| **Go** | All tags preserved exactly as written |
| **C++** | Tags ignored (not applicable) |
| **Swift** | Tags ignored (not applicable) |

**Internal Use:**
- **Schema Parser**: Extracts and stores full tag string
- **JSON Validator**: Parses `json:"fieldName"` for field name mapping
- **Fixture Converter**: Uses JSON tag to map JSON → binary
- **Code Generator**: Outputs full tag string in Go structs

**Example - JSON Tag Mapping:**
```go
// tags.ffi
package example

type User struct {  // Root type (unreferenced + exported)
    ID   int64  `json:"user_id"`
    Name string `json:"name"`
}

// tags.json - uses JSON tag names
{
  "user_id": 123,
  "name": "Alice"
}

// ✓ Valid - matches json tags
// ✗ Invalid - using struct field names
{
  "ID": 123,    // Wrong! Should be "user_id"
  "Name": "Alice"  // Wrong! Should be "name"
}
```

## Generated Usage

**Go:**
```go
import "yourproject/audio"

devices := []audio.Device{{Name: "Speaker", Channels: 2}}
bytes := audio.EncodeDeviceList(devices)

decoded, err := audio.DecodeDeviceList(bytes)
```

**C++:**
```cpp
#include "audio_ffire.h"

std::vector<audio::Device> devices = {{"Speaker", "spk1", 2, true}};
auto bytes = audio::encode_device_list(devices);

auto decoded = audio::decode_device_list(bytes);
```
