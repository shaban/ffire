# Reserved Keywords

Language generators must handle identifiers that conflict with target language keywords.

## Strategy

**Two-Level Keyword Collision Avoidance:**

1. **Type Names**: Automatic `Message` suffix for all root types
2. **Module Names**: Use package name (from schema), optionally sanitize per-language

### Type Level

Users write clean, natural type names in schemas:
```go
type Config struct { ... }
type Device struct { ... }
type User struct { ... }
```

Generators automatically append `Message` suffix to generated types:
```swift
public class ConfigMessage { ... }    // Generated from Config
public class DeviceMessage { ... }    // Generated from Device
public class UserMessage { ... }      // Generated from User
```

### Module/Package Level

**Module names use the package declaration**, not the schema filename:

```go
// struct.ffi - filename doesn't matter for naming
package test

type Config struct { ... }
```

Generates:
- **Swift**: `import test` (not `import struct`)
- **Python**: `import test` (not `import struct`)
- **Library**: `libtest.dylib` (not `libstruct.dylib`)
- **C functions**: `config_decode()` (uses type name, not module)

**Why package name?**
- ✅ Semantic: matches schema's declared package
- ✅ Consistent: library, module, namespace all use same name
- ✅ Multi-file: multiple schemas can share one package
- ✅ No collisions: package names rarely conflict with keywords

**Per-language sanitization**: If package name is a keyword in target language, generators can append safe suffix (e.g., Swift: `struct` → `structModule`).

## Rationale

### Why automatic `Message` suffix?

1. **User-Friendly Schemas**: Users write natural names (`Config`, not `ConfigMessage`)
2. **Zero Keyword Collisions**: Generated types cannot conflict with keywords
3. **Industry Standard**: Protobuf, gRPC use "Message" terminology
4. **Clear Boundary**: Generated `ConfigMessage` ≠ user's domain `Config` type
5. **Zero Maintenance**: No keyword lists to maintain across 11 languages
6. **Future-Proof**: Works for any language we add

### Example: Clear Separation

Schema (user writes):
```go
// schema.ffi
package myapp

type Config struct {
    timeout int32
    host    string
}
```

Generated Swift:
```swift
// ffire generates
public class ConfigMessage {
    var timeout: Int32
    var host: String
}

// User's domain model (separate)
struct Config {
    let timeout: Int
    let host: String
}

// Clear separation between generated and domain types
let msg = ConfigMessage.decode(data)
let config = Config(from: msg)  // Your mapping
```

## Implementation

All generators apply the transformation:
```go
func messageTypeName(name string) string {
    return name + "Message"
}

// Config → ConfigMessage
// Device → DeviceMessage
```

## Examples

### Schema Definition (User Writes)
```go
package test

type Config struct {
    timeout int32
    host    string
}

type Device struct {
    name   string
    active bool
}
```

### Generated Code (Automatic Message Suffix)

**Go:**
```go
type ConfigMessage struct {
    Timeout int32
    Host    string
}
```

**Swift:**
```swift
public class ConfigMessage {
    public var timeout: Int32
    public var host: String
}
```

**Python:**
```python
class ConfigMessage:
    def __init__(self, timeout: int, host: str):
        self.timeout = timeout
        self.host = host
```

**C++:**
```cpp
struct ConfigMessage {
    int32_t timeout;
    std::string host;
};
```

## Common Patterns

### Request/Response (User Writes)
```go
type LoginRequest struct {
    username string
    password string
}

type LoginResponse struct {
    token   string
    expires int64
}
```

Generates: `LoginRequestMessage`, `LoginResponseMessage`

### Events (User Writes)
```go
type DeviceConnected struct {
    deviceID string
    timestamp int64
}
```

Generates: `DeviceConnectedMessage`

## Status

✅ **IMPLEMENTED** - Generators automatically append `Message` suffix to all root types.
