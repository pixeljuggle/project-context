# Project Context

A zero-dependency command-line tool that generates a `project-context.md` file containing a recursive directory tree and the contents of relevant source files. The output is suitable for LLMs, code reviews, or documentation.

The tool supports full `.gitignore` rules (including negation with `!`), hard-coded ignores for common directories, per-file content rules, file-size limits, truncation, and extension filtering.

---

## Installation

### For end users

```bash
# Homebrew
brew install pixeljuggle/project-context/project-context

# Go
go install github.com/pixeljuggle/project-context/cmd/project-context@latest
```

Binaries for Linux, macOS, and Windows are available on the [Releases page](https://github.com/pixeljuggle/project-context/releases).

---

## Usage

Run the tool in the root of any project:

```bash
project-context
```

### Common workflows for fullstack TypeScript / Next.js projects

```bash
# Generate and copy to clipboard (most common)
project-context --stdout | pbcopy

# Limit size and truncate large files
project-context --max-size 750 --truncate 150 --stdout | pbcopy

# Include only source files
project-context --ext .ts,.tsx,.js,.jsx,.json,.md --stdout | pbcopy

# Additional ignores
project-context -I "*.test.*" -I "*.spec.*" -I "coverage/" --stdout | pbcopy
```

---

## Configuration

Create `project-context.json` in the project root to set defaults:

```json
{
  "ignores": ["*.log", "*.tmp", "coverage/"],
  "rules": {
    "public/": { "content": false },
    "src/assets/": { "content": false }
  },
  "maxSizeKB": 500,
  "truncateLines": 150
}
```

- `ignores`: Additional `.gitignore`-style patterns (applied after `.gitignore`)
- `rules`: Per-path rules (`content: false` shows the file in the tree but skips its content)
- `maxSizeKB`: Default maximum file size in KB (0 = unlimited)
- `truncateLines`: Default number of lines to keep for large files (0 = skip)

Command-line flags override config values.

---

## Command-Line Flags

| Flag              | Default                | Description |
|-------------------|------------------------|-----------|
| `--stdout`        | false                  | Print output to stdout instead of writing a file |
| `--max-size`      | 1024                   | Maximum file size in KB (0 = unlimited) |
| `--truncate`      | 0                      | Truncate large files to this many lines (0 = skip) |
| `--verbose`       | false                  | Print skipped or truncated files to stderr |
| `--ext`           | (repeatable)           | Include only files with these extensions |
| `--root`          | `.`                    | Project root directory |
| `--config`        | `project-context.json` | Path to config file |
| `--output`        | `project-context.md`   | Output filename |
| `--no-gitignore`  | false                  | Do not load `.gitignore` |
| `-I`              | (repeatable)           | Additional ignore pattern (supports `!` negation) |
| `--version`       | —                      | Print version and exit |

---

## Development

### For Go developers

```bash
git clone https://github.com/pixeljuggle/project-context.git
cd project-context

make build          # builds to ./bin/project-context
make install        # installs to $GOPATH/bin
make run            # runs against the project itself
go run ./cmd/project-context --stdout
```

### Makefile targets

| Command                | Purpose |
|------------------------|---------|
| `make build`           | Build for current platform |
| `make all`             | Build all supported platforms |
| `make lint`            | Run gofmt, vet, and staticcheck |
| `make test`            | Run tests |
| `make release`         | Local snapshot release |
| `make release-dry-run` | Validate release configuration |
| `make clean`           | Remove build artifacts |

Tests are in `internal/generator/generator_test.go`.

### Releasing

```bash
git tag v1.2.0
git push origin v1.2.0
```

GitHub Actions and GoReleaser handle cross-compilation, archives, checksums, and Homebrew tap publication.

---

## Contributing

1. Fork and clone the repository.
2. Make changes.
3. Run `make lint && go test ./...`.
4. Submit a pull request.

---

[Releases](https://github.com/pixeljuggle/project-context/releases) • [Issues](https://github.com/pixeljuggle/project-context/issues)
