# Documentation

ffire documentation built with [MkDocs Material](https://squidfunk.github.io/mkdocs-material/).

## Building Locally

```bash
# Install (one-time)
pip install mkdocs mkdocs-material

# Serve with live reload
mkdocs serve

# Build static site
mkdocs build
```

Then open http://127.0.0.1:8000

## Structure

```
docs/
├── index.md                      # Home page
├── architecture/
│   ├── overview.md              # System architecture  
│   ├── schema-format.md         # Schema language
│   ├── wire-format.md           # Binary format spec
│   └── generators.md            # Adding languages
├── development/
│   ├── testing.md               # Test strategy
│   ├── benchmarks.md            # Performance testing
│   └── keywords.md              # Reserved word handling
├── api/
│   ├── cli.md                   # Command-line usage
│   └── go-api.md                # Programmatic API
└── internals/
    ├── encoder-internals.md     # Encoder design
    └── optimizations.md         # Performance techniques
```

## Contributing

Documentation improvements welcome. Keep tone **concise and sober** - facts over hype.

## Deployment

GitHub Pages via `.github/workflows/docs.yml` (TODO: add workflow).
