# ffire Library Release Roadmap

## Current Status (Baseline)

### ✅ Working (10/10 benchmarks)
- Go (reference implementation)
- C++ (Tier A)
- Swift (Tier B, via C ABI)
- Dart (Tier B, via C ABI)
- **JavaScript (Tier B, via N-API)** ✅ FIXED
- **Python (Tier B, via pybind11)** ✅ FIXED
- **Protobuf (comparison baseline)** ✅ FIXED

### ❌ Not Working
- C#: Generator exists, no benchmark
- Java: Generator exists, no benchmark

### 🚫 Not Planned
- **PHP**: Generator exists but not pursuing
  - **Reasoning**: Web-centric ecosystem lacks high-performance IPC use cases
  - PHP deployments use HTTP/Redis for inter-service communication, not binary protocols
  - No persistent process model in typical PHP environments (Apache mod_php, PHP-FPM)
  - Market overlap with Python (scripting language with better IPC ecosystem)
  - Better to focus resources on languages with clear binary serialization demand
  
- **Ruby**: Generator exists but not pursuing
  - **Reasoning**: Same limitations as PHP, even less IPC culture
  - Ruby 3.x has YJIT (15-50% faster) but still web/scripting focused
  - Ecosystem dominated by Rails (web) and DevOps tools (Chef, Puppet)
  - Ruby users solve cross-process problems with HTTP/Redis/JSON, not binary IPC
  - No desktop or high-performance same-machine communication patterns
  - Python already covers "dynamic scripting language" use cases better

### 🚫 Missing
- Rust: No generator (needs native Tier A implementation)

---

## Release Requirements

1. **All benchmark languages fully working** (10/10 each)
2. **C#, Java working** (existing generators + benchmarks)
3. **Rust native implementation** (new Tier A generator)
4. **Protobuf baseline working** (for comparison)
5. **Performance optimization** for slow Tier B languages
6. **Complete documentation** (developer + user facing)

**Note**: PHP and Ruby are explicitly excluded from v1.0 due to lack of market fit for high-performance binary IPC in their ecosystems.

---

## Phase 1: Stabilization & Baseline (Priority: Critical)

**Goal**: Fix existing implementations to 10/10

### Milestone 1.1: Fix JavaScript (Target: 10/10) ✅ COMPLETE
**Status**: 6/10 → 10/10  
**Issue**: npm install failures - binding.cpp referenced wrong type names

- [x] Task 1.1.1: Investigate npm install errors for empty schema
  - Root cause: JavaScript binding used `msg.Name` instead of `msg.Name + "Message"`
  - Fix: Updated `generator_javascript.go` line 109 to use Message suffix
  - Commit: "fix: JavaScript binding uses Message suffix for type names"

- [x] Task 1.1.2: Verified all schemas working
  - All 10 schemas now build and run successfully
  - Test: `mage clean && mage genAll && mage runJavaScript` shows 10/10
  - No additional fixes needed - single root cause

**Verification**: ✅ `mage runJavaScript` shows 10/10 passing

---

### Milestone 1.2: Fix Python Array/Primitive Support (Target: 10/10) ✅ COMPLETE
**Status**: 4/10 → 10/10  
**Issue**: __init__.py exported classes that didn't exist for array/primitive types

- [x] Task 1.2.1: Remove class exports for non-struct message types
  - Solution: Python's dynamic typing handles raw vectors/primitives naturally
  - Fix: Only export classes for struct-based messages in __init__.py
  - Modified: `generator_python_pybind11.go` lines 282-350
  - Commit: "fix: Python __init__.py only exports struct message classes"

- [x] Task 1.2.2: Update benchmark to not expect Message classes
  - Issue: Benchmark tried to get {Name}Message class but only needs functions
  - Fix: Removed unused `MessageClass = getattr(pkg, ...)` line
  - Modified: `benchmark_python.go` removed className variable
  - Commit: "fix: Python benchmark uses only encode/decode functions"

- [x] Task 1.2.3: Verified all schemas working
  - All 10 schemas now build and run successfully
  - Test: `mage clean && mage genAll && mage runPython` shows 10/10
  - Arrays and primitives work via direct vector/value return

**Verification**: ✅ `mage runPython` shows 10/10 passing

---

### Milestone 1.3: Fix Protobuf Baseline (Target: 10/10) ✅ COMPLETE
**Status**: 0/10 → 10/10  
**Issue**: Benchmark generator incorrectly added Message suffix to proto type names

- [x] Task 1.3.1: Fix protobuf message name mapping
  - Root cause: `inferMessageType()` added Message suffix but proto doesn't use it
  - Protobuf uses exact message names from .proto files (EmptyTest not EmptyTestMessage)
  - Fix: Updated typeMap to use bare names matching .proto definitions
  - Modified: `benchmarks/magefile.go` lines 935-958
  - Commit: "fix: protobuf benchmark uses correct message names from .proto files"

- [x] Task 1.3.2: Verified fix is isolated to protobuf
  - Confirmed: `inferMessageType()` only called by `generateProtoBenchmark()`
  - No impact on other 6 languages (Go, C++, Python, Dart, Swift, JavaScript)
  - Test: `mage clean && mage genAll && mage runProto` shows 10/10

**Verification**: ✅ `mage runProto` shows 10/10 passing
  - Commit: "docs: protobuf benchmark regression analysis"

- [ ] Task 1.3.2: Fix protobuf regression
  - Fix: Root cause identified in 1.3.1
  - Likely: Schema package naming or generated code path issues
  - Test: Individual schema compiles and runs
  - Commit: "fix: protobuf benchmark regression"

- [ ] Task 1.3.3: Validate all protobuf benchmarks
  - Test: `mage runProto` shows 10/10
  - Compare: Performance against previous baseline (if available)
  - Commit: "test: validate protobuf baseline restoration"

**Verification**: `mage runProto` shows 10/10 passing

---

## Phase 2: Rust Native Implementation (Priority: High)

**Goal**: Add Rust as Tier A (native, no FFI) - strategic priority for performance-critical applications

### Milestone 2.1: Rust Generator Foundation
**Status**: Does not exist
**Rationale**: Rust + zero-copy = fastest possible binary serialization

- [ ] Task 2.1.1: Design Rust code generation architecture
  - Research: Rust serialization patterns (serde reference)
  - Design: Message suffix (ConfigMessage struct)
  - Design: Encode/decode API (methods vs functions)
  - Document: Architecture in YAML
  - Commit: "docs: Rust generator architecture design"

- [ ] Task 2.1.2: Create Rust generator skeleton
  - Create: `pkg/generator/generator_rust.go`
  - Implement: Basic struct generation (empty structs)
  - Pattern: Follow Go/C++ Tier A pattern (self-contained)
  - Commit: "feat: Rust generator skeleton"

- [ ] Task 2.1.3: Implement Rust type system
  - Implement: Primitive types (bool, i8, i16, i32, i64, f32, f64, String)
  - Implement: Optional types (Option<T>)
  - Implement: Array types (Vec<T>)
  - Implement: Nested structs
  - Test: Generate struct schema
  - Commit: "feat: Rust type system implementation"

- [ ] Task 2.1.4: Implement Rust encoder
  - Implement: Wire format encoding (match ffire spec)
  - Pattern: impl for each type
  - Test: Encode struct schema matches wire format
  - Commit: "feat: Rust encoder implementation"

- [ ] Task 2.1.5: Implement Rust decoder
  - Implement: Wire format decoding
  - Error handling: Result<T, Error> pattern
  - Test: Decode struct schema matches expected
  - Commit: "feat: Rust decoder implementation"

**Verification**: Rust generator can generate struct schema with working encode/decode

---

### Milestone 2.2: Rust Benchmark Integration
**Status**: Benchmark harness does not exist

- [ ] Task 2.2.1: Create Rust benchmark harness
  - Create: `pkg/benchmark/benchmark_rust.go`
  - Generate: Cargo.toml with dependencies
  - Generate: Rust benchmark driver (criterion or manual)
  - Commit: "feat: Rust benchmark harness"

- [ ] Task 2.2.2: Add Rust to magefile
  - Add: `RunRust()` function
  - Build: `cargo build --release` integration
  - Run: Benchmark execution
  - Commit: "feat: integrate Rust into benchmark suite"

- [ ] Task 2.2.3: Validate Rust across all schemas
  - Test: `mage runRust`
  - Fix: Any encoding/decoding bugs found
  - Iterate: Until 10/10 passing
  - Commit: "fix: Rust validation fixes"

**Verification**: `mage runRust` shows 10/10 passing

---

## Phase 3: C# & Java Native Implementations (Priority: Medium)

**Goal**: Replace C ABI bridges with pure native implementations for Tier A performance

**Rationale**: Modern JIT compilers (RyuJIT, HotSpot) achieve near-C++ speed. Native implementations provide:
- 🚀 10x faster than P/Invoke/JNI (no FFI overhead)
- 📦 Better distribution (single package, no native deps)
- 🔧 Easier debugging (pure managed code)
- ✨ Idiomatic APIs (native language features)

### Milestone 3.1: C# Native Generator (Target: 10/10)
**Status**: Replace existing P/Invoke implementation

- [ ] Task 3.1.1: Design C# native generator
  - Research: Span<byte> for zero-copy operations
  - Design: Pure C# encoder/decoder (no FFI)
  - Pattern: Copy from Go generator (similar syntax)
  - Document: Architecture design
  - Commit: "docs: C# native generator design"

- [ ] Task 3.1.2: Create C# native generator
  - Create: `pkg/generator/generator_csharp_native.go`
  - Implement: Struct generation with properties
  - Implement: Encode/decode using BinaryWriter/Span<byte>
  - Test: Generate struct schema
  - Commit: "feat: C# native generator"

- [ ] Task 3.1.3: Implement C# type system
  - Primitives: bool, sbyte, short, int, long, float, double, string
  - Optionals: Nullable<T> for value types, null for reference types
  - Arrays: List<T> or T[]
  - Nested: Child classes
  - Commit: "feat: C# native type system"

- [ ] Task 3.1.4: Validate and benchmark
  - Test: `mage runCSharp` shows 10/10
  - Compare: Performance vs old P/Invoke version
  - Expect: ~10x faster, competitive with Go/C++
  - Commit: "feat: C# native implementation complete"

**Verification**: C# native shows Tier A performance (~50-200ns per operation)

---

### Milestone 3.2: Java Native Generator (Target: 10/10)
**Status**: Create new native implementation

- [ ] Task 3.2.1: Design Java native generator
  - Research: ByteBuffer for efficient encoding
  - Design: Pure Java encoder/decoder (no JNI)
  - Pattern: Copy from C# native generator
  - Document: Architecture design
  - Commit: "docs: Java native generator design"

- [ ] Task 3.2.2: Create Java native generator
  - Create: `pkg/generator/generator_java.go`
  - Implement: Class generation with getters/setters
  - Implement: Encode/decode using ByteBuffer
  - Test: Generate struct schema
  - Commit: "feat: Java native generator"

- [ ] Task 3.2.3: Implement Java type system
  - Primitives: boolean, byte, short, int, long, float, double, String
  - Optionals: Optional<T> or null
  - Arrays: ArrayList<T> or native arrays
  - Nested: Inner/outer classes
  - Commit: "feat: Java native type system"

- [ ] Task 3.2.4: Create Java benchmark harness
  - Create: `pkg/benchmark/benchmark_java.go`
  - Generate: Maven/Gradle project structure
  - Generate: Benchmark driver (JMH or manual)
  - Commit: "feat: Java benchmark harness"

- [ ] Task 3.2.5: Validate and benchmark
  - Test: `mage runJava` shows 10/10
  - Compare: Performance expectations (Tier A)
  - Expect: Within 20% of Go/C++
  - Commit: "feat: Java native implementation complete"

**Verification**: Java native shows Tier A performance

---

## Phase 4: Performance Optimization & Analysis (Priority: Medium)

**Goal**: Optimize remaining Tier B languages and document tradeoffs

### Milestone 4.1: Performance Baseline & Analysis
**Status**: Need data

- [ ] Task 4.1.1: Collect comprehensive benchmark data
  - Run: `mage bench` and save results
  - Document: Performance comparison table
  - Identify: Slowest operations per language
  - Commit: "docs: baseline performance analysis"

- [ ] Task 4.1.2: Analyze Tier B FFI overhead
  - Profile: Swift, Dart, Python, JavaScript benchmarks
  - Measure: FFI call overhead vs native implementations
  - Identify: Major bottlenecks (memory allocation, data conversion, etc.)
  - Document: Findings with potential optimizations
  - Commit: "docs: Tier B FFI overhead analysis"

---

### Milestone 4.2: Optimization Implementation
**Status**: Depends on analysis

- [ ] Task 4.2.1: Optimize identified hotspots (per language)
  - Examples:
    - Reduce memory allocations in Swift
    - Batch FFI calls in Dart
    - Direct memory access in Python
    - Zero-copy where possible
  - Pattern: One commit per language optimized
  - Test: Performance improvement > 20%
  - Commit: "perf: {language} {specific optimization}"

- [ ] Task 4.2.2: Document optimization techniques
  - Update: Architecture YAML with patterns
  - Create: Performance tuning guide
  - Commit: "docs: Tier B performance optimization guide"

**Verification**: All Tier B languages within 5-10x of native (acceptable FFI overhead)

---

## Phase 5: Documentation & Examples (Priority: High)

**Goal**: Complete developer and user documentation

### Milestone 5.1: Developer Documentation
**Status**: Partial (architecture YAML exists)

- [ ] Task 5.1.1: Generator development guide
  - Create: `docs/GENERATOR_DEVELOPMENT.md`
  - Cover: Adding new language generators
  - Cover: Message suffix architecture
  - Cover: C ABI integration for Tier B
  - Cover: Testing and validation
  - Commit: "docs: generator development guide"

- [ ] Task 5.1.2: Architecture documentation
  - Create: `docs/ARCHITECTURE.md`
  - Cover: Wire format specification
  - Cover: Tier A vs Tier B design
  - Cover: Type system and encoding rules
  - Commit: "docs: architecture documentation"

- [ ] Task 5.1.3: Benchmark suite documentation
  - Create: `docs/BENCHMARKING.md`
  - Cover: How to run benchmarks
  - Cover: Adding new benchmark schemas
  - Cover: Interpreting results
  - Commit: "docs: benchmark suite documentation"

- [ ] Task 5.1.4: API reference for each language
  - Generate: Per-language API docs from code
  - Pattern: One doc per language
  - Location: `docs/api/{language}.md`
  - Commit: "docs: API reference for all languages"

**Verification**: Developers can add new generator by following docs

---

### Milestone 5.2: User Documentation
**Status**: Minimal

- [ ] Task 5.2.1: Getting started guide
  - Create: `docs/GETTING_STARTED.md`
  - Cover: Installation per language
  - Cover: Basic usage example (all languages)
  - Cover: Defining schemas (.ffi format)
  - Commit: "docs: getting started guide"

- [ ] Task 5.2.2: Language-specific guides
  - Create: `docs/languages/{lang}.md` for each language
  - Cover: Installation, usage patterns, best practices
  - Include: Complete working examples
  - Pattern: One commit per language guide
  - Commit: "docs: {language} user guide"

- [ ] Task 5.2.3: Schema definition guide
  - Create: `docs/SCHEMA.md`
  - Cover: Type system (primitives, structs, arrays, optionals)
  - Cover: Message definitions
  - Cover: Best practices
  - Commit: "docs: schema definition guide"

- [ ] Task 5.2.4: Performance guide
  - Create: `docs/PERFORMANCE.md`
  - Cover: Performance characteristics per language
  - Cover: When to use which language/tier
  - Cover: Optimization tips
  - Commit: "docs: performance guide"

**Verification**: Users can get started in < 5 minutes with any language

---

### Milestone 5.3: Examples & Tutorials
**Status**: None

- [ ] Task 5.3.1: Create example projects
  - Create: `examples/{language}/` for each language
  - Pattern: Simple client-server or data pipeline
  - Test: All examples run successfully
  - Commit: "examples: working examples for all languages"

- [ ] Task 5.3.2: Create tutorial series
  - Create: `docs/tutorials/` directory
  - Tutorial 1: Building a simple config system
  - Tutorial 2: Cross-language data exchange
  - Tutorial 3: Performance optimization
  - Commit: "docs: tutorial series"

**Verification**: Examples demonstrate real-world usage patterns

---

## Phase 6: Release Preparation (Priority: Critical)

**Goal**: Polish for public release

### Milestone 6.1: README & Project Polish

- [ ] Task 6.1.1: Write comprehensive README
  - Include: Feature highlights, quick start, language support matrix
  - Include: Performance comparison table
  - Include: Link to all documentation
  - Commit: "docs: comprehensive README"

- [ ] Task 6.1.2: Add CI/CD pipeline
  - Setup: GitHub Actions for all languages
  - Test: Run full benchmark suite on PR
  - Test: Verify 10/10 for all languages
  - Commit: "ci: GitHub Actions pipeline"

- [ ] Task 6.1.3: Prepare release artifacts
  - Package: Language-specific packages (npm, crates.io, etc.)
  - Versioning: Semantic versioning strategy
  - Changelog: Complete CHANGELOG.md
  - Commit: "chore: prepare v1.0.0 release artifacts"

**Verification**: Project ready for public announcement

---

## Summary & Prioritization

### Critical Path (Must complete for release):
1. **Phase 1**: Stabilization (JS, Python, Protobuf) ✅ COMPLETE
2. **Phase 2**: Rust implementation → 8 tasks
3. **Phase 3**: C# & Java native implementations → 10 tasks
4. **Phase 5**: Documentation → 12 tasks
5. **Phase 6**: Release prep → 3 tasks

**Total Critical Path**: ~33 tasks remaining

### Excluded from v1.0:
- **PHP**: Web-centric ecosystem lacks IPC use cases
- **Ruby**: Same as PHP, even less IPC culture

### Medium Priority (Can be post-1.0):
- **Phase 4**: Performance optimization → 4 tasks + language-specific

---

## Recommended Execution Order

1. ~~**Week 1-2**: Phase 1 (Stabilization)~~ ✅ COMPLETE
2. **Week 3-4**: Phase 2 (Rust) - Native Tier A implementation
3. **Week 5-6**: Phase 3 (C# & Java) - Native Tier A implementations
4. **Week 7-8**: Phase 5 (Documentation) - Complete docs & examples
5. **Week 9**: Phase 4 (Performance) - Optimize Tier B languages
6. **Week 10**: Phase 6 (Release) - Final polish & CI/CD

**Target Release**: ~8 weeks remaining

---

## Success Metrics

- [ ] All languages: 10/10 benchmark pass rate
- [ ] Total languages: 9 (Go, C++, Rust, Swift, Dart, Python, JS, C#, Java)
  - **Note**: PHP and Ruby excluded due to lack of market fit for binary IPC
- [ ] Documentation coverage: 100% (all sections complete)
- [ ] Examples: Working examples for each language
- [ ] CI/CD: Green builds on all PRs
- [ ] Performance: Tier A ~50-200ns, Tier B within 10x of Tier A
