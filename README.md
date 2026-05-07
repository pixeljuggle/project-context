# Project Context CLI

A fast, zero-dependency CLI tool that generates a perfect `project-context.md` file for LLMs, code reviews, or documentation.

It includes:

- Beautiful recursive directory tree
- All file contents in properly tagged code blocks
- Full support for `.gitignore`
- Hard-coded ignore for `.git` (and everything inside it)
- JSON config for custom ignores + per-folder rules
- Extra ignore rules via `-I` flag

## Installation

```bash
# Install latest version globally
go install github.com/pixeljuggle/project-context/cmd/project-context@latest

# Or build from source
git clone https://github.com/pixeljuggle/project-context.git
cd project-context
make install
```

## Quick Start

```bash
# In any project
project-context

# Custom options
project-context -root ./my-app -config my-config.json -I "*.tmp" -I "build/"
```

## Configuration (`project-context.json`)

```json
{
  "ignores": ["node_modules/", "dist/", "*.log"],
  "rules": {
    "docs/": { "content": false },
    "internal/secret.txt": { "content": false }
  }
}
```

## Makefile Commands

```bash
make build          # Build for current OS/arch
make install        # Install to $GOPATH/bin
make run            # Run directly
make all            # Build all popular platforms
make clean
```

See `Makefile` for full cross-compilation targets (Linux, macOS, Windows, arm64, etc.).
