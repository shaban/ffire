# ffire Benchmarks

Cross-language benchmark suite using Magefile for orchestration.

## Quick Start

```bash
# Install mage (one-time)
go install github.com/magefile/mage@latest

# Build ffire CLI (from repo root)
cd .. && go build -o ffire ./cmd/ffire && cd benchmarks

# Run benchmarks
mage bench              # All languages
mage bench:go           # Go only
mage bench:cpp          # C++ only
mage bench:dart         # Dart only
```

## Results

View `results/comparison.md` for side-by-side performance comparison.

JSON results saved to `results/<language>_<format>.json` for automation.

## Documentation

**Complete guide:** [docs/development/benchmarks.md](../docs/development/benchmarks.md)

Topics covered:
- Adding new benchmarks
- Language-specific setup
- Performance analysis
- Optimization techniques
- Memory profiling
