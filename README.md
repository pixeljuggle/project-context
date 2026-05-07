# Project Context CLI

## A fast, zero-dependency CLI tool that generates a perfect `project-context.md` for LLMs, code reviews, or documentation.

## Features

- Recursive directory tree (Git-style)
- All file contents in properly syntax-highlighted code blocks
- **Safe Markdown handling** — uses 4-backtick fences for `.md` files so nested code blocks never break the output
- Hard-coded ignores for common junk: `.git/`, `node_modules/`, `dist/`, `build/`, `target/`, `venv/`, `.venv/`, `.next/`, `__pycache__`, etc.
- Full `.gitignore` support + custom `project-context.json` config
- Smart skipping of binary files and large files (`--max-size`)
- Token estimation (rough but useful for LLM context limits)
- `--stdout` mode — perfect for piping directly to clipboard or AI tools
- Zero external dependencies (pure Go stdlib)

## Installation

### From GitHub Releases (recommended)

Download the latest binary for your OS from the [Releases page](https://github.com/pixeljuggle/project-context/releases).

### With Go

```bash
go install github.com/pixeljuggle/project-context/cmd/project-context@latest
```

## Usage

```bash
project-context --version
project-context -I "node_modules/" -I "dist/"
project-context -config my-config.json
```

## Automatic Releases

This project uses **GoReleaser** + GitHub Actions.  
Just run:

```bash
git tag v1.0.0
git push origin v1.0.0
```

A new release with binaries for Linux, macOS, and Windows will be created automatically.

See `.goreleaser.yaml` and `.github/workflows/release.yml` for details.

## Quick Start

```bash
# 1. Go to any project
cd ~/my-awesome-app

# 2. Generate the context file
project-context

# 3. (Optional) Pipe directly to clipboard for instant LLM use
project-context --stdout | pbcopy
```

### Realistic everyday examples

**Example 1: Full context for a Next.js / TypeScript project**

```bash
project-context --max-size 500 --stdout | pbcopy
```

**Example 2: Ignore extra patterns on the fly**

```bash
project-context -I "*.test.ts" -I "e2e/" -I "coverage/"
```

**Example 3: Use a custom config and output to a different name**

```bash
project-context -config my-context-config.json -output context-for-claude.md
```

**Example 4: Skip .gitignore and only use your own rules**

```bash
project-context --no-gitignore -I "node_modules/" -I "dist/"
```

## Configuration (`project-context.json`)

Create this file in your project root for project-specific rules:

```json
{
  "ignores": ["*.log", "*.tmp", "coverage/", "storybook-static/"],
  "rules": {
    "docs/": { "content": false },
    "public/": { "content": false },
    "src/assets/": { "content": false },
    "internal/secret-config.ts": { "content": false }
  }
}
```

- `ignores`: `.gitignore`-style patterns (applied in addition to `.gitignore`)
- `rules`: Per-folder or per-file rules (`content: false` = show in tree but skip content)

## Command-Line Flags

| Flag             | Default                | Description                             |
| ---------------- | ---------------------- | --------------------------------------- |
| `--stdout`       | false                  | Print to stdout instead of writing file |
| `--max-size`     | 1024                   | Max file size in KB (0 = unlimited)     |
| `--root`         | `.`                    | Project root directory                  |
| `--config`       | `project-context.json` | Path to config file                     |
| `--output`       | `project-context.md`   | Output markdown filename                |
| `--no-gitignore` | false                  | Skip loading `.gitignore`               |
| `-I`             | (repeatable)           | Additional ignore pattern               |
| `--version`      | -                      | Show version and exit                   |
| `--help`         | -                      | Show help                               |

## Makefile Commands

```bash
make build             # Build for current OS/arch → ./bin/project-context
make install           # Install globally via go install
make run               # Run directly with go run
make all               # Build all platforms (Linux, macOS, Windows + amd64/arm64) → ./bin/
make release           # Full local release test (creates ./dist/ with .tar.gz + .zip)
make release-dry-run   # Only validate .goreleaser.yaml config (no build)
make lint              # Run gofmt + go vet + staticcheck (recommended before committing)
make clean             # Remove bin/ and dist/
make test-build        # Quick version check after build
```

**Note about releases & linting:**

- `make release` uses GoReleaser in snapshot mode for local testing.
- `make release-dry-run` validates the release config only.
- `make lint` is the recommended pre-commit check (it will fail CI-style if formatting or issues are found).
