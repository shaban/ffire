# ffire Library Release Roadmap

## Current Status (Baseline)

### ✅ Working (10/10 benchmarks)
- Go (reference implementation)
- C++ (Tier A)
- Swift (Tier B, via C ABI)
- Dart (Tier B, via C ABI)

### ⚠️ Partial (need fixes)
- JavaScript: 6/10 (npm install issues in 4 benchmarks)
- Python: 4/10 (pybind11 only supports struct types, not arrays/primitives)

### ❌ Not Working
- Protobuf (comparison baseline): 0/10 - **Regression** (was working, now broken)
- C#: Generator exists, no benchmark
- Java: Generator exists, no benchmark
- PHP: Generator exists, no benchmark
- Ruby: Generator exists, no benchmark

### 🚫 Missing
- Rust: No generator (needs native Tier A implementation)

---

## Release Requirements

1. **All benchmark languages fully working** (10/10 each)
2. **PHP, C#, Java working** (existing generators + benchmarks)
3. **Rust native implementation** (new Tier A generator)
4. **Protobuf baseline working** (for comparison)
5. **Performance optimization** for slow Tier B languages
6. **Complete documentation** (developer + user facing)

---

## Phase 1: Stabilization & Baseline (Priority: Critical)

**Goal**: Fix existing implementations to 10/10

### Milestone 1.1: Fix JavaScript (Target: 10/10)
**Status**: 6/10 → 10/10  
**Issue**: npm install failures in 4 schemas (empty, nested, struct, tags)

- [ ] Task 1.1.1: Investigate npm install errors for empty schema
  - Run: `cd benchmarks/generated/ffire_javascript_empty/javascript && npm install`
  - Debug: Check package.json, native module build
  - Test: `mage runJavaScript` shows 7/10
  - Commit: "fix: JavaScript empty schema npm install"

- [ ] Task 1.1.2: Fix remaining npm install errors (nested, struct, tags)
  - Debug: Likely similar root cause
  - Test: `mage runJavaScript` shows 10/10
  - Commit: "fix: JavaScript npm install for all schemas"

**Verification**: `mage runJavaScript` shows 10/10 passing

---

### Milestone 1.2: Fix Python Array/Primitive Support (Target: 10/10)
**Status**: 4/10 → 10/10  
**Issue**: pybind11 only binds struct types, not array/primitive root messages

- [ ] Task 1.2.1: Design Python bindings for array root types
  - Research: How to bind `std::vector<T>` as root message in pybind11
  - Design: Wrapper class pattern or direct binding?
  - Document: Update architecture YAML with approach
  - Commit: "docs: design Python array/primitive message bindings"

- [ ] Task 1.2.2: Implement array type bindings in pybind11 generator
  - Modify: `pkg/generator/generator_python_pybind11.go`
  - Add: Bindings for array root messages (std::vector wrapper)
  - Test: Generate array_float, array_int, array_string schemas
  - Commit: "feat: Python pybind11 array root message support"

- [ ] Task 1.2.3: Implement primitive type bindings (if needed)
  - Handle: Primitive root messages (e.g., single int32)
  - Test: Generate any primitive-based schemas
  - Commit: "feat: Python pybind11 primitive root message support"

- [ ] Task 1.2.4: Fix complex schema (nested struct + optional handling)
  - Debug: Why complex schema fails (likely optional fields)
  - Fix: Optional field handling in pybind11 bindings
  - Test: `mage runPython` shows 10/10
  - Commit: "fix: Python complex schema with optional fields"

**Verification**: `mage runPython` shows 10/10 passing

---

### Milestone 1.3: Fix Protobuf Baseline (Target: 10/10)
**Status**: 0/10 → 10/10  
**Issue**: Protobuf benchmarks are integrated but all failing (regression from working state)

- [ ] Task 1.3.1: Investigate protobuf benchmark failures
  - Run: `mage runProto` to see exact errors
  - Context: These were working previously, something broke them
  - Check: Recent changes that might have affected protobuf
  - Check: Proto file generation vs compilation vs runtime
  - Document: Root cause analysis (what changed?)
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

## Phase 2: Language Expansion (Priority: High)

**Goal**: Activate existing generators with benchmarks

### Milestone 2.1: C# Support (Target: 10/10)
**Status**: Generator exists, no benchmark

- [ ] Task 2.1.1: Audit C# generator for Message suffix compliance
  - Check: `pkg/generator/generator_csharp.go` for Message suffix
  - Check: Uses schema.Package not filename
  - Fix: Any violations found
  - Test: Generate test schema manually
  - Commit: "fix: C# generator Message suffix compliance"

- [ ] Task 2.1.2: Create C# benchmark harness
  - Create: `pkg/benchmark/benchmark_csharp.go`
  - Pattern: Follow Dart/Swift Tier B pattern (FFI to C ABI)
  - Generate: C# project with proper structure
  - Commit: "feat: C# benchmark harness"

- [ ] Task 2.1.3: Add C# to magefile
  - Add: `RunCSharp()` function to benchmarks/magefile.go
  - Add: C# to `bench` workflow
  - Test: `mage genAll` generates C# benchmarks
  - Commit: "feat: integrate C# into benchmark suite"

- [ ] Task 2.1.4: Validate and fix C# implementation
  - Run: `mage runCSharp`
  - Fix: Any errors found (likely DLL loading, P/Invoke)
  - Test: 10/10 passing
  - Commit: "fix: C# benchmark validation"

**Verification**: `mage runCSharp` shows 10/10 passing

---

### Milestone 2.2: Java Support (Target: 10/10)
**Status**: Generator exists, no benchmark

- [ ] Task 2.2.1: Audit Java generator for Message suffix compliance
  - Check: `pkg/generator/generator_java.go`
  - Fix: Message suffix violations
  - Test: Generate test schema manually
  - Commit: "fix: Java generator Message suffix compliance"

- [ ] Task 2.2.2: Create Java benchmark harness
  - Create: `pkg/benchmark/benchmark_java.go`
  - Pattern: JNI/JNA to C ABI (Tier B)
  - Generate: Maven/Gradle project structure
  - Commit: "feat: Java benchmark harness"

- [ ] Task 2.2.3: Add Java to magefile
  - Add: `RunJava()` function
  - Integrate: Into benchmark workflow
  - Test: `mage genAll` generates Java benchmarks
  - Commit: "feat: integrate Java into benchmark suite"

- [ ] Task 2.2.4: Validate and fix Java implementation
  - Run: `mage runJava`
  - Fix: JNI/JNA loading, native library path
  - Test: 10/10 passing
  - Commit: "fix: Java benchmark validation"

**Verification**: `mage runJava` shows 10/10 passing

---

### Milestone 2.3: PHP Support (Target: 10/10)
**Status**: Generator exists, no benchmark

- [ ] Task 2.3.1: Audit PHP generator for Message suffix compliance
  - Check: `pkg/generator/generator_php.go`
  - Fix: Message suffix violations
  - Test: Generate test schema manually
  - Commit: "fix: PHP generator Message suffix compliance"

- [ ] Task 2.3.2: Create PHP benchmark harness
  - Create: `pkg/benchmark/benchmark_php.go`
  - Pattern: FFI to C ABI (PHP 7.4+ FFI extension)
  - Generate: composer.json, proper PHP project
  - Commit: "feat: PHP benchmark harness"

- [ ] Task 2.3.3: Add PHP to magefile
  - Add: `RunPHP()` function
  - Integrate: Into benchmark workflow
  - Test: `mage genAll` generates PHP benchmarks
  - Commit: "feat: integrate PHP into benchmark suite"

- [ ] Task 2.3.4: Validate and fix PHP implementation
  - Run: `mage runPHP`
  - Fix: FFI loading, shared library path
  - Test: 10/10 passing
  - Commit: "fix: PHP benchmark validation"

**Verification**: `mage runPHP` shows 10/10 passing

---

## Phase 3: Rust Native Implementation (Priority: High)

**Goal**: Add Rust as Tier A (native, no FFI)

### Milestone 3.1: Rust Generator Foundation
**Status**: Does not exist

- [ ] Task 3.1.1: Design Rust code generation architecture
  - Research: Rust serialization patterns (serde reference)
  - Design: Message suffix (ConfigMessage struct)
  - Design: Encode/decode API (methods vs functions)
  - Document: Architecture in YAML
  - Commit: "docs: Rust generator architecture design"

- [ ] Task 3.1.2: Create Rust generator skeleton
  - Create: `pkg/generator/generator_rust.go`
  - Implement: Basic struct generation (empty structs)
  - Pattern: Follow Go/C++ Tier A pattern (self-contained)
  - Commit: "feat: Rust generator skeleton"

- [ ] Task 3.1.3: Implement Rust type system
  - Implement: Primitive types (bool, i8, i16, i32, i64, f32, f64, String)
  - Implement: Optional types (Option<T>)
  - Implement: Array types (Vec<T>)
  - Implement: Nested structs
  - Test: Generate struct schema
  - Commit: "feat: Rust type system implementation"

- [ ] Task 3.1.4: Implement Rust encoder
  - Implement: Wire format encoding (match ffire spec)
  - Pattern: impl for each type
  - Test: Encode struct schema matches wire format
  - Commit: "feat: Rust encoder implementation"

- [ ] Task 3.1.5: Implement Rust decoder
  - Implement: Wire format decoding
  - Error handling: Result<T, Error> pattern
  - Test: Decode struct schema matches expected
  - Commit: "feat: Rust decoder implementation"

**Verification**: Rust generator can generate struct schema with working encode/decode

---

### Milestone 3.2: Rust Benchmark Integration
**Status**: Benchmark harness does not exist

- [ ] Task 3.2.1: Create Rust benchmark harness
  - Create: `pkg/benchmark/benchmark_rust.go`
  - Generate: Cargo.toml with dependencies
  - Generate: Rust benchmark driver (criterion or manual)
  - Commit: "feat: Rust benchmark harness"

- [ ] Task 3.2.2: Add Rust to magefile
  - Add: `RunRust()` function
  - Build: `cargo build --release` integration
  - Run: Benchmark execution
  - Commit: "feat: integrate Rust into benchmark suite"

- [ ] Task 3.2.3: Validate Rust across all schemas
  - Test: `mage runRust`
  - Fix: Any encoding/decoding bugs found
  - Iterate: Until 10/10 passing
  - Commit: "fix: Rust validation fixes"

**Verification**: `mage runRust` shows 10/10 passing

---

## Phase 4: Performance Optimization (Priority: Medium)

**Goal**: Investigate and optimize slow Tier B languages

### Milestone 4.1: Performance Baseline & Analysis
**Status**: Need data

- [ ] Task 4.1.1: Collect comprehensive benchmark data
  - Run: `mage bench` and save results
  - Document: Performance comparison table
  - Identify: Slowest operations per language
  - Commit: "docs: baseline performance analysis"

- [ ] Task 4.1.2: Analyze Tier B FFI overhead
  - Profile: Swift, Dart, Python, C#, Java, PHP benchmarks
  - Measure: FFI call overhead vs C++ native
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

**Verification**: All Tier B languages within 5-10x of C++ (acceptable FFI overhead)

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
1. **Phase 1**: Stabilization (JS, Python, Protobuf) → 3 languages, 24 tasks
2. **Phase 3**: Rust implementation → 8 tasks
3. **Phase 5**: Documentation → 12 tasks
4. **Phase 6**: Release prep → 3 tasks

**Total Critical Path**: ~47 tasks

### High Priority (Important for comprehensive release):
- **Phase 2**: C#, Java, PHP support → 12 tasks

### Medium Priority (Can be post-1.0):
- **Phase 4**: Performance optimization → 2 tasks + language-specific

---

## Recommended Execution Order

1. **Week 1-2**: Phase 1 (Stabilization) - Get existing languages to 100%
2. **Week 3-4**: Phase 3 (Rust) - Native implementation
3. **Week 5**: Phase 2 (C#, Java, PHP) - Expand language support
4. **Week 6-7**: Phase 5 (Documentation) - Complete docs & examples
5. **Week 8**: Phase 4 (Performance) - Optimize slow paths
6. **Week 9**: Phase 6 (Release) - Final polish & CI/CD

**Target Release**: ~9 weeks from start

---

## Success Metrics

- [ ] All languages: 10/10 benchmark pass rate
- [ ] Total languages: 11 (Go, C++, Rust, Swift, Dart, Python, JS, C#, Java, PHP, Ruby)
- [ ] Documentation coverage: 100% (all sections complete)
- [ ] Examples: Working examples for each language
- [ ] CI/CD: Green builds on all PRs
- [ ] Performance: Tier B within 10x of Tier A
