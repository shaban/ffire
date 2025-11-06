# C ABI Dynamic Library - Language Support Matrix

## Languages That Work Out-of-the-Box with C ABI

Our C ABI dylib (`libffire.dylib` / `libffire.so`) can be called from virtually any language with Foreign Function Interface (FFI) support.

### ✅ **Tier 1: Proven Working**

| Language | Status | FFI Mechanism | Notes |
|----------|--------|---------------|-------|
| **Python** | ✅ Tested | `ctypes` | Works perfectly, 4µs per op |
| **Swift** | ✅ Tested | `@_silgen_name` | Works perfectly, 3-4µs per op, Swift 5.5+ |
| **C++** | ✅ Tested | Direct C linkage | Reference implementation |
| **Objective-C++** | ✅ Tested | Direct C linkage | Works with Foundation |

### ✅ **Tier 2: Should Work Immediately (Standard FFI)**

These languages have mature, stable FFI layers that work with C ABI:

| Language | FFI Mechanism | Maturity | Notes |
|----------|---------------|----------|-------|
| **Rust** | `extern "C"` + `unsafe` | Excellent | Zero-cost, can wrap in safe API |
| **Go** | `cgo` | Excellent | Built-in C interop |
| **JavaScript/Node.js** | `node-ffi-napi` or `ffi-rs` | Excellent | Popular npm packages |
| **Zig** | `@cImport()` | Excellent | First-class C interop |
| **C** | Direct | Perfect | Native target |
| **Objective-C** | Direct | Perfect | Same as C |
| **Java** | JNA / JNI | Excellent | Standard approach |
| **Kotlin/JVM** | JNA / JNI | Excellent | Via Java interop |
| **C#/.NET** | P/Invoke | Excellent | Built-in DllImport |
| **F#** | P/Invoke | Excellent | Via .NET |
| **Ruby** | `fiddle` or `ffi` gem | Good | Standard library + gem |
| **Lua** | `ffi` (LuaJIT) | Good | Fast FFI in LuaJIT |
| **PHP** | FFI extension (PHP 7.4+) | Good | Modern PHP |
| **Perl** | `FFI::Platypus` | Good | CPAN module |
| **Julia** | `ccall` | Excellent | Built-in, zero overhead |
| **Nim** | `{.importc.}` | Excellent | Native C interop |
| **D** | `extern(C)` | Excellent | First-class C compat |
| **Haskell** | FFI | Good | Via Foreign module |
| **OCaml** | Ctypes | Good | Well-established |
| **Common Lisp** | CFFI | Good | Standard approach |
| **Racket** | FFI | Good | Built-in foreign interface |
| **Elixir/Erlang** | NIFs or Ports | Good | Native Implemented Functions |
| **Dart** | `dart:ffi` | Good | Built-in (Dart 2.6+) |
| **R** | `.Call()` interface | Good | Standard for packages |
| **Crystal** | `lib` declaration | Excellent | Compile-time C bindings |

### 🔧 **Tier 3: Needs Wrapper/Bridge (But Possible)**

| Language | Approach | Difficulty | Notes |
|----------|----------|------------|-------|
| **WebAssembly** | Emscripten | Medium | Need to compile C wrapper to WASM |
| **TypeScript** | Via Node.js FFI | Easy | Same as JavaScript |
| **Kotlin/Native** | cinterop | Medium | Different from Kotlin/JVM |
| **Scala** | JNA/JNI | Easy | Via JVM |
| **Clojure** | JNA | Easy | Via JVM |

### ❌ **Languages That Don't Support C FFI**

| Language | Reason |
|----------|--------|
| **Pure Bash/Shell** | No FFI mechanism |
| **SQL** | Not a general-purpose language |
| **HTML/CSS** | Markup languages |

---

## Ergonomics: Working with C Structs & Arrays

Our C ABI exposes:
- Opaque handles (`PluginHandle`)
- Getter functions for strings, numbers, booleans, arrays
- Manual memory management (`plugin_free`, `plugin_free_data`)

### 🟢 **Tier A: Comfortable (Native C Interop)**

These languages handle C types naturally with minimal friction:

| Language | Why Comfortable | Example |
|----------|----------------|---------|
| **C** | Native types | Direct struct access |
| **C++** | Native types | Can wrap in RAII classes |
| **Objective-C** | Native types | Same as C |
| **Objective-C++** | Native types | Best of both worlds |
| **Rust** | Excellent FFI | Can create safe wrappers easily |
| **Zig** | C-compatible | Can use C types directly |
| **Nim** | Good C interop | Clean wrapper syntax |
| **D** | Native C compat | Can interface directly |
| **Go** | cgo support | Good tooling, some overhead |
| **Crystal** | C-like syntax | Compile-time bindings |

**User Experience:** Natural, feels like native code. Can create ergonomic wrappers.

### 🟡 **Tier B: Manageable (With Helper Library/Package)**

These languages need wrapper code but it's straightforward:

| Language | Approach | Pain Points |
|----------|----------|-------------|
| **Swift** | Define functions with `@_silgen_name` | Manual pointer management |
| **Python** | ctypes with wrapper classes | Boilerplate for each function |
| **JavaScript/Node.js** | FFI library + wrapper | Async complications, type safety |
| **Ruby** | FFI gem + wrapper class | Need to define signatures |
| **Java** | JNA + wrapper classes | Verbose type mappings |
| **Kotlin/JVM** | JNA + idiomatic API | Same as Java |
| **C#/.NET** | P/Invoke + wrapper | Marshal boilerplate |
| **F#** | P/Invoke + functional API | Same as C# but nicer |
| **Julia** | ccall with types | Manual struct definitions |
| **Lua** | LuaJIT FFI | Manual declarations |
| **PHP** | FFI extension + class | Verbose, need PHP 7.4+ |
| **Perl** | FFI::Platypus | Older ecosystem |
| **Dart** | dart:ffi + wrapper | Manual struct layouts |
| **R** | .Call interface | Academic focus, less tooling |

**User Experience:** Need to write/generate a wrapper layer, but it's a one-time cost. Package maintainers would handle this.

**Recommendation:** Provide code generators or pre-built wrappers for popular languages.

### 🔴 **Tier C: Painful (Manual or Difficult FFI)**

These languages have FFI but it's uncomfortable:

| Language | Why Painful | Issues |
|----------|-------------|--------|
| **Haskell** | Type system friction | Monads, IO boundaries, verbose FFI |
| **OCaml** | Manual binding | Ctypes is verbose, GC complications |
| **Common Lisp** | CFFI setup | Implementation-dependent, varies by runtime |
| **Racket** | FFI complexity | Less common, documentation sparse |
| **Elixir/Erlang** | NIFs are tricky | Need to write C wrapper, memory safety concerns |
| **Kotlin/Native** | Different from JVM | Limited ecosystem, less mature |
| **WebAssembly** | Indirect only | Need Emscripten compilation, no direct dylib |

**User Experience:** Significant effort required. Better to use a higher-level language wrapper.

**Recommendation:** Not recommended for direct use. Use Python/Node.js wrapper instead.

### ❌ **Tier D: Not Practical**

| Language | Why Not Practical |
|----------|------------------|
| **Pure Bash/Shell** | No FFI, would need external tool |
| **SQL** | Not applicable |
| **HTML/CSS** | Not programming languages |
| **TypeScript** | No direct FFI, must use Node.js |
| **Scala** | Better to use Kotlin/Java wrapper |
| **Clojure** | Better to use Java wrapper |

---

## Packaging Strategy by Tier

### **Tier A (Native):**
```
✅ Ship dylib + C header
✅ Users compile/link directly
✅ No wrapper needed (optional RAII wrapper)
```

**Example:** Rust crate with just FFI declarations + safe wrapper

### **Tier B (Wrapper Needed):**
```
✅ Ship dylib
✅ Ship/generate language-specific wrapper
✅ Publish to package registry (PyPI, npm, RubyGems, etc.)
```

**Example:** Python package with:
- `libffire.dylib` in wheel
- `ffire.py` wrapper with classes
- `pip install ffire`

### **Tier C (Not Recommended):**
```
⚠️ Document that they should use Tier B language
⚠️ Or provide pre-built wrappers if demanded
```

---

## Performance Expectations

Based on our testing, expected per-operation overhead:

| Overhead Level | Languages | Time per Op |
|----------------|-----------|-------------|
| **Zero overhead** | C, C++, Rust, Zig, Nim, D, Crystal | 3-4µs |
| **Minimal overhead** | Swift, Go, Objective-C, Julia | 3-5µs |
| **Small overhead** | Python, Ruby, Lua, Node.js | 4-10µs |
| **Moderate overhead** | Java/JVM, C#/.NET, Dart | 5-20µs |
| **Higher overhead** | PHP, Perl, interpreted langs | 10-50µs |

*Note: These are estimates based on FFI overhead characteristics. Actual performance depends on implementation quality and runtime.*

---

## Implementation Complexity

### **Trivial (< 30 lines):**
- C, Objective-C
- Rust (unsafe block)
- Zig
- Nim
- D
- Crystal

### **Easy (30-100 lines):**
- Python (ctypes)
- Swift (@_silgen_name)
- Ruby (fiddle/ffi)
- JavaScript/Node.js
- Go (cgo)
- Julia
- Lua

### **Moderate (100-200 lines):**
- Java/Kotlin (JNA)
- C#/.NET (P/Invoke)
- Haskell
- OCaml
- Common Lisp

---

## Key Takeaways

1. **Universal Compatibility**: C ABI is the lingua franca of programming languages
2. **~40 languages** can use our dylib without recompilation
3. **Performance**: Most languages achieve <10µs overhead per operation
4. **Python & Swift proven**: Our tests show 3-5µs per operation
5. **Binary distribution**: Ship one `.dylib`/`.so` for all these languages!

### **Practical Distribution Strategy:**

**Phase 1 - Core Native Support (Ship dylib only):**
- C, C++, Rust, Zig, Objective-C, Objective-C++
- Target: Systems programmers who are comfortable with FFI

**Phase 2 - Popular High-Level Languages (Ship dylib + wrapper):**
- Python (PyPI), Swift (SPM), Node.js (npm), Ruby (RubyGems), Go (module)
- Target: Mainstream application developers

**Phase 3 - Enterprise/Specialized (On demand):**
- Java (Maven), C# (NuGet), Dart (pub.dev)
- Target: Enterprise and mobile developers

**Not Recommended:**
- Use Python/Node.js/Ruby as bridge for Haskell, OCaml, Lisp, etc.
- WebAssembly: Compile C ABI wrapper to WASM separately

---

## Next Steps for Testing

**High Priority:**
- [ ] Rust (most requested systems language)
- [ ] Go (compare with native Go ffire implementation)
- [ ] JavaScript/Node.js (web/server use case)
- [ ] Java (enterprise use case)

**Medium Priority:**
- [ ] Ruby (Rails ecosystem)
- [ ] C# (Unity, .NET ecosystem)
- [ ] Lua (game scripting)

**Research Interest:**
- [ ] Zig (C replacement)
- [ ] Julia (scientific computing)
- [ ] WebAssembly (browser support)

---

## Addendum: Tier B Packaging Requirements

### Universal Distribution Strategy

**Goal:** Ship a single package structure that works for all languages without linker/symbol conflicts.

```
ffire-dist/
├── lib/
│   ├── libffire.dylib          # macOS
│   ├── libffire.so             # Linux
│   └── ffire.dll               # Windows
├── include/
│   └── ffire.h                 # C header with all declarations
├── wrappers/
│   ├── python/
│   ├── swift/
│   ├── javascript/
│   ├── ruby/
│   ├── java/
│   ├── csharp/
│   ├── dart/
│   └── ... (other Tier B languages)
└── examples/
    ├── python/
    ├── swift/
    └── ... (usage examples per language)
```

---

### Tier B Language Requirements

Each language needs minimal files to work with the C ABI dylib:

#### **1. Python** (PyPI: `ffire`)
**What's Needed:**
- `ffire.py` - ctypes wrapper (~150 lines)
  - Class `Plugin` with properties
  - Functions `decode()`, `encode()`
  - Memory management helpers
- `setup.py` or `pyproject.toml` - package metadata
- Bundle `libffire.dylib/so/dll` in wheel

**Usage:**
```python
import ffire
plugin = ffire.decode(data)
print(plugin.name)
encoded = ffire.encode(plugin)
```

**Generator Output:**
```
-lang python
  Creates: python/ffire.py, python/setup.py
  Includes: libffire.dylib bundled in wheel
  Install: pip install ./python
```

---

#### **2. Swift** (SPM: `ffire`)
**What's Needed:**
- `Package.swift` - SPM manifest
- `FFire.swift` - wrapper (~100 lines)
  - Class `Plugin` with computed properties
  - Functions `decode()`, `encode()`
  - Automatic memory management (deinit)
- Link against `libffire.dylib` at runtime

**Usage:**
```swift
import FFire
let plugin = try FFire.decode(data)
print(plugin.name)
let encoded = FFire.encode(plugin)
```

**Generator Output:**
```
-lang swift
  Creates: swift/Package.swift, swift/Sources/FFire/FFire.swift
  Includes: libffire.dylib linked via SPM
  Install: swift build
```

---

#### **3. JavaScript/Node.js** (npm: `ffire`)
**What's Needed:**
- `package.json` - npm metadata
- `index.js` - FFI wrapper using `ffi-napi` or `ffi-rs` (~200 lines)
  - Class `Plugin` with getters
  - Functions `decode()`, `encode()`
  - Promise-based API
- Bundle `libffire.node` (compiled addon) OR use FFI at runtime

**Usage:**
```javascript
const ffire = require('ffire');
const plugin = ffire.decode(buffer);
console.log(plugin.name);
const encoded = ffire.encode(plugin);
```

**Generator Output:**
```
-lang javascript (or node, nodejs)
  Creates: javascript/package.json, javascript/index.js
  Includes: libffire.dylib loaded via FFI
  Install: npm install ./javascript
```

---

#### **4. Ruby** (RubyGems: `ffire`)
**What's Needed:**
- `ffire.gemspec` - gem specification
- `lib/ffire.rb` - FFI wrapper using `fiddle` or `ffi` gem (~150 lines)
  - Class `Plugin` with attr_readers
  - Module methods `decode`, `encode`
  - Automatic GC cleanup
- Bundle `libffire.bundle` (dylib)

**Usage:**
```ruby
require 'ffire'
plugin = FFire.decode(data)
puts plugin.name
encoded = FFire.encode(plugin)
```

**Generator Output:**
```
-lang ruby
  Creates: ruby/ffire.gemspec, ruby/lib/ffire.rb
  Includes: libffire.dylib bundled in gem
  Install: gem install ./ruby/ffire-1.0.0.gem
```

---

#### **5. Java** (Maven: `com.ffire:ffire`)
**What's Needed:**
- `pom.xml` - Maven configuration
- `src/main/java/com/ffire/Plugin.java` - JNA wrapper (~300 lines)
  - Interface for JNA bindings
  - Class `Plugin` with getters
  - Static methods `decode()`, `encode()`
  - Resource management (try-with-resources)
- Bundle `libffire.dylib/so/dll` in resources

**Usage:**
```java
import com.ffire.Plugin;
Plugin plugin = Plugin.decode(data);
System.out.println(plugin.getName());
byte[] encoded = plugin.encode();
```

**Generator Output:**
```
-lang java
  Creates: java/pom.xml, java/src/main/java/com/ffire/*.java
  Includes: libffire.dylib in src/main/resources
  Install: mvn install
```

---

#### **6. Kotlin/JVM** (Maven/Gradle)
**What's Needed:**
- Same as Java + `build.gradle.kts`
- `src/main/kotlin/com/ffire/Plugin.kt` - Idiomatic Kotlin wrapper (~200 lines)
  - Data class for Plugin
  - Extension functions
  - Inline classes for type safety

**Usage:**
```kotlin
import com.ffire.decode
val plugin = decode(data)
println(plugin.name)
val encoded = plugin.encode()
```

**Generator Output:**
```
-lang kotlin
  Creates: kotlin/build.gradle.kts, kotlin/src/main/kotlin/com/ffire/*.kt
  Includes: Reuses Java JNA bindings + Kotlin wrapper
  Install: gradle build
```

---

#### **7. C#/.NET** (NuGet: `FFire`)
**What's Needed:**
- `FFire.csproj` - project file
- `Plugin.cs` - P/Invoke wrapper (~250 lines)
  - DllImport declarations
  - Class `Plugin` with properties
  - IDisposable pattern for memory management
- Bundle `libffire.dylib/so/dll` in runtimes folder

**Usage:**
```csharp
using FFire;
using var plugin = Plugin.Decode(data);
Console.WriteLine(plugin.Name);
var encoded = plugin.Encode();
```

**Generator Output:**
```
-lang csharp (or dotnet)
  Creates: csharp/FFire.csproj, csharp/Plugin.cs
  Includes: libffire.dll in runtimes/win-x64/native
  Install: dotnet pack
```

---

#### **8. F#** (.NET)
**What's Needed:**
- Same as C# + `FFire.fsproj`
- `Plugin.fs` - Functional wrapper (~200 lines)
  - Same P/Invoke as C#
  - Functional API with discriminated unions
  - Computation expressions for resource management

**Usage:**
```fsharp
open FFire
let plugin = decode data
printfn "%s" plugin.Name
let encoded = encode plugin
```

**Generator Output:**
```
-lang fsharp
  Creates: fsharp/FFire.fsproj, fsharp/Plugin.fs
  Includes: Reuses C# P/Invoke + F# functional wrapper
  Install: dotnet pack
```

---

#### **9. Julia** (Pkg)
**What's Needed:**
- `Project.toml` - package metadata
- `src/FFire.jl` - ccall wrapper (~100 lines)
  - Struct definitions
  - Functions using `ccall`
  - Finalizers for cleanup

**Usage:**
```julia
using FFire
plugin = decode(data)
println(plugin.name)
encoded = encode(plugin)
```

**Generator Output:**
```
-lang julia
  Creates: julia/Project.toml, julia/src/FFire.jl
  Includes: libffire.dylib loaded via ccall
  Install: Julia Pkg.add(path="./julia")
```

---

#### **10. Lua** (LuaRocks)
**What's Needed:**
- `ffire-1.0-1.rockspec` - LuaRocks spec
- `ffire.lua` - LuaJIT FFI wrapper (~100 lines)
  - ffi.cdef declarations
  - Module with decode/encode functions
  - Metatables for object methods

**Usage:**
```lua
local ffire = require("ffire")
local plugin = ffire.decode(data)
print(plugin.name)
local encoded = ffire.encode(plugin)
```

**Generator Output:**
```
-lang lua
  Creates: lua/ffire-1.0-1.rockspec, lua/ffire.lua
  Includes: libffire.so loaded via LuaJIT FFI
  Install: luarocks install ./lua/ffire-1.0-1.rockspec
```

---

#### **11. PHP** (Composer: `ffire/ffire`)
**What's Needed:**
- `composer.json` - Composer metadata
- `src/Plugin.php` - FFI wrapper using PHP 7.4+ FFI (~250 lines)
  - Class with FFI declarations
  - Static methods for decode/encode
  - Destructor for cleanup

**Usage:**
```php
use FFire\Plugin;
$plugin = Plugin::decode($data);
echo $plugin->getName();
$encoded = $plugin->encode();
```

**Generator Output:**
```
-lang php
  Creates: php/composer.json, php/src/Plugin.php
  Includes: libffire.so loaded via FFI::cdef
  Install: composer install
```

---

#### **12. Dart** (pub.dev: `ffire`)
**What's Needed:**
- `pubspec.yaml` - Dart package spec
- `lib/ffire.dart` - dart:ffi wrapper (~200 lines)
  - Class extending Struct
  - Functions using DynamicLibrary
  - Finalizers for native memory

**Usage:**
```dart
import 'package:ffire/ffire.dart';
final plugin = decode(data);
print(plugin.name);
final encoded = encode(plugin);
```

**Generator Output:**
```
-lang dart
  Creates: dart/pubspec.yaml, dart/lib/ffire.dart
  Includes: libffire.dylib loaded via DynamicLibrary
  Install: dart pub get
```

---

#### **13. Perl** (CPAN: `FFire`)
**What's Needed:**
- `Makefile.PL` or `dist.ini` - CPAN metadata
- `lib/FFire.pm` - FFI::Platypus wrapper (~150 lines)
  - Package with Platypus declarations
  - Subroutines for decode/encode
  - DESTROY for cleanup

**Usage:**
```perl
use FFire;
my $plugin = FFire::decode($data);
print $plugin->name;
my $encoded = FFire::encode($plugin);
```

**Generator Output:**
```
-lang perl
  Creates: perl/Makefile.PL, perl/lib/FFire.pm
  Includes: libffire.so loaded via FFI::Platypus
  Install: perl Makefile.PL && make install
```

---

#### **14. R** (CRAN: `ffire`)
**What's Needed:**
- `DESCRIPTION` - R package metadata
- `R/ffire.R` - .Call interface (~100 lines)
  - Functions using .C or .Call
  - S3 or R6 classes for Plugin
  - On-load hook for dylib

**Usage:**
```r
library(ffire)
plugin <- decode(data)
print(plugin$name)
encoded <- encode(plugin)
```

**Generator Output:**
```
-lang r
  Creates: r/DESCRIPTION, r/R/ffire.R, r/src/wrapper.c
  Includes: libffire.so as package shared library
  Install: R CMD INSTALL r/
```

---

### Summary: Generator Strategy

**For Tier A Languages (Native):**
```bash
ffire generate -lang cpp -schema my.ffi
ffire generate -lang rust -schema my.ffi
ffire generate -lang c -schema my.ffi
```

**Output:**
- `generated.hpp/rs/h` - Native language code
- `README.md` - Basic usage instructions
- Prints: "Link against libffire.dylib or include directly"

**For Tier B Languages (Wrapper):**
```bash
ffire generate -lang python -schema my.ffi
ffire generate -lang swift -schema my.ffi
ffire generate -lang javascript -schema my.ffi
```

**Output:**
- Language-specific wrapper package (complete with build config)
- `libffire.dylib` copied into package
- `README.md` - Installation and usage guide
- Prints: "Run 'pip install .' (or language-specific install command)"

---

### Unified Distribution Structure

```
ffire-release/
├── README.md                          # Universal installation guide
├── lib/
│   ├── darwin-arm64/libffire.dylib
│   ├── darwin-x86_64/libffire.dylib
│   ├── linux-x86_64/libffire.so
│   ├── linux-arm64/libffire.so
│   └── windows-x64/ffire.dll
├── include/
│   └── ffire.h                        # Universal C header
├── packages/
│   ├── python/
│   │   ├── ffire/
│   │   │   ├── __init__.py
│   │   │   └── libffire.dylib -> ../../lib/darwin-arm64/libffire.dylib
│   │   ├── setup.py
│   │   └── README.md
│   ├── swift/
│   │   ├── Package.swift
│   │   ├── Sources/FFire/FFire.swift
│   │   └── README.md
│   ├── javascript/
│   │   ├── package.json
│   │   ├── index.js
│   │   └── README.md
│   └── ... (other Tier B languages)
└── examples/
    ├── python/example.py
    ├── swift/example.swift
    └── ... (working examples per language)
```

**Key Design Principles:**
1. ✅ Single dylib - no duplicate symbols
2. ✅ Platform-specific folders - avoid linker conflicts
3. ✅ Symlinks to shared lib - DRY principle
4. ✅ Self-contained packages - each can be distributed independently
5. ✅ Language-idiomatic structure - familiar to users

---

### Generator Implementation

**Command:**
```bash
ffire generate -lang <language> -schema my.ffi -out ./dist
```

**Behavior:**

**Tier A (C, C++, Rust, Zig, etc.):**
- Generates native code only
- Prints basic linking instructions
- No wrapper needed

**Tier B (Python, Swift, Node.js, etc.):**
- Generates wrapper package with all files
- Copies appropriate `libffire.dylib/so/dll`
- Includes package metadata (setup.py, Package.swift, etc.)
- Prints: "Package ready at ./dist/<lang>. Run: <install-command>"

**Benefits:**
- ✅ One generator handles all languages
- ✅ Consistent output structure
- ✅ No manual packaging needed
- ✅ Ready to publish (PyPI, npm, SPM, etc.)
- ✅ Examples included for quick start

---

**Bottom Line:** The C ABI dylib approach makes ffire usable in virtually every mainstream programming language with minimal effort!
